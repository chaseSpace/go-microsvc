package xgrpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"microsvc/bizcomm"
	"microsvc/bizcomm/auth"
	"microsvc/bizcomm/mq"
	"microsvc/consts"
	"microsvc/deploy"
	"microsvc/enums"
	"microsvc/infra/xgrpc/protoutil"
	"microsvc/infra/xmq"
	"microsvc/model/svc/micro_svc"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util"
	"microsvc/util/graceful"
	"microsvc/util/ujson"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bufbuild/protovalidate-go"

	"github.com/spf13/cast"

	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/k0kubun/pp/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type grpcHTTPRegister func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error

// server动态使用grpc端口范围
const grpcPortMin = 60000
const grpcPortMax = 60999

const httpPortMin = 61000
const httpPortMax = 61999

type XgRPC struct {
	srv                              *grpc.Server
	healthSrv                        *health.Server
	extHttpRegister, intHttpRegister grpcHTTPRegister
}

func New(interceptors ...grpc.UnaryServerInterceptor) *XgRPC {
	srv := newGRPCServer(deploy.XConf.Svc, interceptors...)

	// 注册健康检查接口
	healthSrv := health.NewServer()
	healthgrpc.RegisterHealthServer(srv, healthSrv)

	return &XgRPC{
		srv:             srv,
		healthSrv:       healthSrv,
		extHttpRegister: nil,
	}
}

func (x *XgRPC) Apply(regFunc func(s *grpc.Server)) {
	regFunc(x.srv)
}

func (x *XgRPC) SetHTTPExtRegister(register grpcHTTPRegister) {
	x.extHttpRegister = register
}

func (x *XgRPC) SetHTTPIntRegister(register grpcHTTPRegister) {
	x.intHttpRegister = register
}

func (x *XgRPC) Start(portSetter deploy.SvcListenPortSetter) {
	var (
		lis  net.Listener
		err  error
		port int
	)

	lis, port, err = getListener(deploy.XConf.GRPCPort, portSetter)
	if err != nil {
		xlog.Panic(err.Error())
	}
	grpcAddr := fmt.Sprintf("localhost:%d", port)

	fmt.Printf("\nCongratulations! ^_^\n")
	_, _ = pp.Printf("Your service [%s] is serving gRPC on %s\n", portSetter.GetSvc(), grpcAddr)

	// 手动设置微服务的健康状态
	x.healthSrv.SetServingStatus(deploy.XConf.Svc.Name(), healthgrpc.HealthCheckResponse_SERVING)

	defer graceful.AddStopFunc(func() { // grpc server should stop before http
		x.srv.GracefulStop()
		xlog.Info("xgrpc: gRPC server shutdown completed")
	})

	graceful.Register(func() {
		err = x.srv.Serve(lis)
		if err != nil {
			xlog.Error("xgrpc: failed to serve GRPC", zap.String("grpcAddr", grpcAddr), zap.Error(err))
		}
	})
	fmt.Println() // 手动换行
}

func serveHTTP(grpcAddr string, httpListener net.Listener, extHandlerRegister, intHandlerRegister grpcHTTPRegister) {
	conn, err := grpc.Dial(grpcAddr, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		xlog.Panic("xgrpc: grpc.Dial failed", zap.String("grpcAddr", grpcAddr), zap.Error(err))
	}
	defer conn.Close()

	muxOpt := newHTTPMuxOpts()
	mux := runtime.NewServeMux(muxOpt...) // create http gateway router for grpc service

	if extHandlerRegister != nil {
		err = extHandlerRegister(context.TODO(), mux, conn)
		if err != nil {
			xlog.Panic("xgrpc: register ext handler failed", zap.String("grpcAddr", grpcAddr), zap.Error(err))
		}
	}
	if intHandlerRegister != nil {
		err = intHandlerRegister(context.TODO(), mux, conn)
		if err != nil {
			xlog.Panic("xgrpc: register int handler failed", zap.String("grpcAddr", grpcAddr), zap.Error(err))
		}
	}
	svr := http.Server{Handler: mux}
	graceful.AddStopFunc(func() {
		util.RunTaskWithCtxTimeout(time.Second*3, func(ctx context.Context) {
			err = svr.Shutdown(ctx)
			xlog.Info("xgrpc: HTTP server shutdown completed", zap.Error(err))
		})
	})

	err = svr.Serve(httpListener)
	if err != nil && err != http.ErrServerClosed {
		xlog.Panic("xgrpc: failed to serve HTTP", zap.String("grpcAddr", grpcAddr), zap.Error(err))
	}
}

type proxyRespMarshaler struct {
	runtime.JSONPb
}

func (c *proxyRespMarshaler) Marshal(grpcRsp interface{}) (b []byte, err error) {
	lastResp := &protoutil.HTTPResp{
		Code: xerr.ErrNil.Code,
		Msg:  xerr.ErrNil.Msg,
		Data: nil,
	}
	defer func() {
		b, err = c.JSONPb.Marshal(lastResp)
	}()
	if grpcRsp == nil {
		lastResp.Code = xerr.ErrInternal.Code
		lastResp.Msg = "http-proxy: no error, but grpc response is empty"
		return
	}
	//data, err := anypb.New(grpcRsp.(proto.Message))
	//if err != nil {
	//	lastResp.Code = xerr.ErrInternal.Code
	//	lastResp.Msg = fmt.Sprintf("http-proxy: call anypb.New() failed: %v, rsp:%+v", err, grpcRsp)
	//	return
	//}
	//lastResp.Data = data
	return
}

func gatewayMarshaler() *proxyRespMarshaler {
	jpb := runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			AllowPartial:    true,
			UseProtoNames:   true,
			UseEnumNumbers:  true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			AllowPartial:   true,
			DiscardUnknown: true,
		},
	}
	return &proxyRespMarshaler{JSONPb: jpb}
}

func newHTTPMuxOpts() []runtime.ServeMuxOption {
	marshaler := gatewayMarshaler()
	return []runtime.ServeMuxOption{
		runtime.WithMarshalerOption(marshaler.ContentType(nil), marshaler),
		runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
			var header = map[string]bool{
				"x-token": true,
			}
			s = strings.ToLower(s)
			return s, header[s]
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
			rsp := &protoutil.HTTPResp{
				Code: xerr.ErrInternal.Code,
				Msg:  err.Error(),
			}
			s, ok := status.FromError(err)
			if ok {
				if e, ok := xerr.FromErrStr(s.Message()); ok {
					rsp.Code = e.Code
					rsp.Msg = e.Msg
				} else {
					rsp.Msg = s.Message()
				}
			}
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(util.ToJson(rsp))
		}),
	}
}

func newGRPCServer(svc enums.Svc, interceptors ...grpc.UnaryServerInterceptor) *grpc.Server {
	certDir := filepath.Join(deploy.XConf.GetConfDir(), "cert")

	certPath := filepath.Join(certDir, "server-cert.pem")
	keyPath := filepath.Join(certDir, "server-key.pem")

	// 加载服务器证书和私钥
	serverCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		panic(err)
	}

	// 加载根证书池，用于验证客户端证书
	rootCA, err := os.ReadFile(filepath.Join(certDir, "ca-cert.pem"))
	if err != nil {
		panic(err)
	}
	rootCAPool := x509.NewCertPool()
	ok := rootCAPool.AppendCertsFromPEM(rootCA)
	if !ok {
		panic("newGRPCServer: rootCAPool.AppendCertsFromPEM failed")
	}

	// 创建服务器 TLS 配置
	// 使用根证书验证client证书
	serverTLSConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    rootCAPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,

		// 在自定义验证逻辑里面，添加证书过期时告警的逻辑，而不是返回error
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			fmt.Printf("\n")
			defer func() {
				fmt.Printf("\n")
			}()

			// 验证证书链中的每个证书（一般是 客户端证书、根证书的顺序）
			for _, chain := range verifiedChains {
				for _, cert := range chain {
					switch cert.Subject.CommonName {
					case certClientCN:
						//pp.Printf("验证通过--Client证书信息: CN:%s before:%s  after:%s \n",
						//	cert.Subject.CommonName, cert.NotBefore, cert.NotAfter)
					case certRootCN:
						//pp.Printf("验证通过--根证书信息: CN:%s before:%s  after:%s \n",
						//	cert.Subject.CommonName, cert.NotBefore, cert.NotAfter)
					default:
						// 授权特定client
						if specialClientAuth(svc.Name(), cert.DNSNames) {
							//pp.Printf("验证通过--特定client CN：%s  DNSNames: %+v\n", cert.Subject.CommonName, cert.DNSNames)
						} else {
							return fmt.Errorf("grpc: handshake faield, invalid client certificate with CN(%s)", cert.Subject.CommonName)
						}
					}
					// 获取证书的有效期
					now := time.Now()
					if now.Before(cert.NotBefore) {
						return fmt.Errorf("grpc: handshake faield, client certificate is invalid before %s", cert.NotBefore)
					}
					if now.After(cert.NotAfter) {
						// 这一步可以不做强验证，因为一旦证书过期（忘记及时更新），这里返回err会导致服务间通信失败
						// 这里可以加上告警
						//return fmt.Errorf("client certificate is expired at %s", cert.NotAfter)

						pp.Printf("client certificate is expired at %s", cert.NotAfter)
					}
				}
			}
			return nil
		},
	}

	inter := newServerInterceptor(svc)

	// 拦截器列表，洋葱模式执行
	base := []grpc.UnaryServerInterceptor{
		inter.OuterMost,
		inter.StandardizationGRPCErr,
		inter.TraceGRPC,
		inter.RateLimitByIP,
		inter.ValidateParams,
		inter.RecoverBizLogic,
		inter.Authentication,
		inter.RateLimitByUID,
		inter.Innermost,
	}

	// 创建 gRPC 服务器
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.Creds(credentials.NewTLS(serverTLSConfig)),
		grpc.ChainUnaryInterceptor(
			append(base, interceptors...)...,
		))
	return server
}

// -------- 下面是grpc中间件 -----------

var emptyHandler = func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil }

type ServerInterceptor struct {
	svc       enums.Svc
	validator *protovalidate.Validator
}

func newServerInterceptor(svc enums.Svc) ServerInterceptor {
	validator, _ := protovalidate.New()
	return ServerInterceptor{svc: svc, validator: validator}
}

// OuterMost 最外层的拦截器
func (s ServerInterceptor) OuterMost(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	defer func() { // 最外层的panic 就是拦截器 panic了，不再是handler问题
		if hints := recover(); hints != nil {
			tip := fmt.Sprintf("panic recovered: %v", hints)
			if err == nil {
				// 此处的panic属于拦截器异常
				err = xerr.ErrInternal.TraceSvc(s.svc).New(tip)
			}
			xlog.DPanic("gRPC ServerInterceptor Recovered [OuterMost]", zap.Any("err", hints))
		}
		if err != nil {
			err = xerr.ToXErr(err).TraceSvc(s.svc)
		}
		elapsed := time.Since(start)
		reqS := util.TruncateUTF8(string(ujson.MustMarshal(req)), deploy.XConf.GRPC.LogPrintReqMaxLen, "...省略")
		resS := util.TruncateUTF8(string(ujson.MustMarshal(resp)), deploy.XConf.GRPC.LogPrintRespMaxLen, "...省略")
		zapFields := []zap.Field{
			zap.String("method", info.FullMethod), zap.String("dur", elapsed.String()),
			zap.String("req", reqS), zap.String("trace-id", getIncomingMdVal(ctx, MdKeyTraceId)),
		}

		if err != nil {
			println(fmt.Sprintf("%+v\n", err)) // print origin stack to stderr
			zapFields = append(zapFields, zap.String("err", xerr.ToXErr(err).FlatMsg()))
			xlog.ErrorNoStack("grpc reply_err log", zapFields...)
		} else {
			zapFields = append(zapFields, zap.String("resp", resS))
			xlog.Debug("grpc reply_ok log", zapFields...)
		}
		// 如果是网关调用，则转换成网关标准的Resp
		resp, err = s.__convertToGatewayResp(ctx, resp, err)
	}()

	// 请求处理前，透传来自网关的部分属性，以便请求经过的每个子服务都能读取
	// see: https://golang2.eddycjy.com/posts/ch3/09-grpc-metadata-creds/
	ctx, err = transferMDInsideOfCtx(ctx, info.FullMethod, MdKeyTraceId, MdKeyBizRemoteAddr)
	if err != nil {
		return
	}

	resp, err = handler(ctx, req)
	return
}

func (s ServerInterceptor) StandardizationGRPCErr(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		e, ok := status.FromError(err)
		if ok {
			err = xerr.ToXErr(errors.New(e.Message()))
		} else {
			err = xerr.ToXErr(err)
		}
		//if !sutil.fromGatewayCall(ctx) {
		//	return nil, xerr.ErrRPC.WithAPI(info.FullMethod).Append(err)
		//}
		return nil, err
	}
	return
}

func (ServerInterceptor) TraceGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// TODO: add tracing
	return handler(ctx, req)
}

// ValidateParams 验证请求参数
// Docs: https://github.com/bufbuild/protovalidate/tree/main/docs
// Examples: https://github.com/bufbuild/protovalidate/tree/main/examples
func (s ServerInterceptor) ValidateParams(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	err = s.validator.Validate(req.(proto.Message))
	if err != nil {
		return nil, xerr.ErrParams.New(err.Error())
	}
	return handler(ctx, req)
}

// RecoverBizLogic 捕获业务逻辑的panic
func (s ServerInterceptor) RecoverBizLogic(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()
	defer func() {
		hasPanic := false
		elapsed := time.Since(start)
		if r := recover(); r != nil {
			hasPanic = true
			// 此处的panic属于api异常
			err = xerr.ErrAPIUnavailable.New(fmt.Sprintf("panic recovered: %v", r))
			xlog.DPanic("gRPC ServerInterceptor recovered [RecoverBizLogic]", zap.String("method", info.FullMethod), zap.Any("err", r),
				zap.String("trace-id", getIncomingMdVal(ctx, MdKeyTraceId)))
		}
		// 写入RPC日志（不含常见的静态参数错误，已被 ValidateParams 拦截）
		go s.__writeRpcLog(ctx, info.FullMethod, elapsed, hasPanic, err, s.svc.Name(), sutil.fromGatewayCall(ctx))
	}()
	resp, err = handler(ctx, req)
	return
}

func (s ServerInterceptor) Authentication(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	var credential auth.AuthenticateUserAPI
	var lang commonpb.Lang
	lang, credential, err = s.__authentication(ctx, req, info)
	if err != nil {
		return
	}
	if credential != nil {
		ctx = context.WithValue(ctx, auth.CtxAuthenticated{}, credential)
	}
	resp, err = handler(ctx, req)
	if err != nil && sutil.isExtMethod(ctx) { // 在这里对错误信息进行国际化，至于更外层产生的错误，不再处理
		return nil, xerr.ToXErr(err).Translate(lang)
	}
	return
}

// RateLimitByIP IP限速 在身份认证前执行
func (s ServerInterceptor) RateLimitByIP(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if !sutil.isExtMethod(ctx) { // 仅针对外部请求限速
		return handler(ctx, req)
	}
	return bizcomm.APIRateLimiter(deploy.XConf.OpenAPIRateLimit, micro_svc.R.Client).
		CheckByIP(ctx, info.FullMethod, GetReqClientIP(ctx), func() (interface{}, error) {
			return handler(ctx, req)
		})
}

// RateLimitByUID UID限速 在身份认证后执行
func (s ServerInterceptor) RateLimitByUID(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return bizcomm.APIRateLimiter(deploy.XConf.OpenAPIRateLimit, micro_svc.R.Client).
		CheckByUID(ctx, info.FullMethod, auth.GetAuthUID(ctx), func() (interface{}, error) {
			return handler(ctx, req)
		})
}

// Innermost 最内层的拦截器
func (ServerInterceptor) Innermost(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if sutil.isExtMethod(ctx) && !sutil.fromGatewayCall(ctx) && !sutil.isTestCall(ctx) {
		return nil, xerr.ErrParams.New("call external gRPC method(%s) only allowed bypass gateway", info.FullMethod)
	}
	return handler(ctx, req)
}

func (ServerInterceptor) __convertToGatewayResp(ctx context.Context, resp interface{}, err error) (interface{}, error) {
	// 来自网关的请求直接转换为标准响应返出去
	if sutil.isExtMethod(ctx) && sutil.fromGatewayCall(ctx) {
		return protoutil.WrapResponseOnService(resp, err, IsProtobufData(ctx))
	}
	return resp, err
}

// 身份认证
// NOTE：没有任何DB操作
func (s ServerInterceptor) __authentication(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo) (lang commonpb.Lang, claims auth.AuthenticateUserAPI, err error) {
	lang = commonpb.Lang_CL_EN
	if sutil.isExtMethod(ctx) { // 外部请求必须携带base字段
		if s.svc == enums.SvcAdmin {
			lang, err = sutil.checkAdminBaseExtReq(req)
			if err != nil {
				return
			}
		} else {
			lang, err = sutil.checkBaseExtReq(req)
			if err != nil {
				return
			}
		}
	}

	tokenStr := strings.TrimPrefix(getIncomingMdVal(ctx, MdKeyAuthToken), "Bearer ")
	fakeUIDStr := getIncomingMdVal(ctx, MdKeyFakeAuth)
	// 若有token，则也会认证
	if (auth.NoAuthMethods[info.FullMethod] != nil && tokenStr == "" && fakeUIDStr == "") || !sutil.isExtMethod(ctx) {
		return
	}

	if tokenStr == "" {
		if fakeUIDStr != "" {
			fakeUID := cast.ToInt64(fakeUIDStr)
			if s.svc == enums.SvcAdmin {
				claims = auth.NewTestAdminUser(fakeUID, enums.SexMale)
			} else {
				claims = auth.NewTestSvcUser(fakeUID, enums.SexMale)
			}
			return
		}
		return lang, nil, xerr.ErrUnauthorized.AppendMsg("empty token")
	}

	var (
		jti string
	)

	claims, err = ParseSvcAuthToken(tokenStr, s.svc == enums.SvcAdmin)
	if err != nil {
		return
	}
	// TODO: Check the 'jti' field to prevent replay attacks.
	xlog.Warn("NEED CHECK `jti` FIELD -- server.interceptor.Authentication", zap.String("jti", jti))

	if !claims.IsValidCredential() {
		return lang, nil, xerr.ErrUnauthorized.AppendMsg("invalid claims")
	}
	return
}

func (ServerInterceptor) __writeRpcLog(ctx context.Context, fullMethod string, elapsed time.Duration, hasPanic bool,
	err error, svc string, fromGateway bool) {
	parsedMethod := sutil.ParseFullMethod(fullMethod)
	callLog := mq.NewMsgAPICallLog(
		&mq.APICallLogBody{
			APIName:     parsedMethod.API,
			APICtrl:     parsedMethod.Ctrl,
			ReqIP:       GetReqClientIP(ctx),
			DurMs:       elapsed.Milliseconds(),
			Success:     err == nil,
			Svc:         svc,
			FromGateway: fromGateway,
			Panic:       hasPanic,
			//ErrMsg:    "",
		})
	if err != nil {
		callLog.ErrMsg = xerr.ToXErr(err).FlatMsg()
	}

	// 这里调用频率非常高，消息入列让下游去慢慢入库
	xmq.Produce(consts.TopicAPICallLog, callLog)
}

// ---------------- 优雅的分割线 -----------------

func ParseSvcAuthToken(tokenStr string, isAdmin bool) (claims auth.AuthenticateUserAPI, err error) {
	var signKey string
	if isAdmin {
		claims = &auth.AdminClaims{}
		signKey = deploy.XConf.AdminTokenSignKey
	} else {
		claims = &auth.SvcClaims{}
		signKey = deploy.XConf.SvcTokenSignKey
	}
	if signKey == "" {
		return nil, xerr.ErrInternal.New("empty sign key from config!!!")
	}
	t, err := jwt.ParseWithClaims(tokenStr, jwt.Claims(claims), func(token *jwt.Token) (interface{}, error) {
		issuer, _ := token.Claims.GetIssuer()
		if issuer != auth.TokenIssuer && issuer != auth.TokenIssuerExtAdmin {
			return nil, fmt.Errorf("unknown issuer:%s", issuer)
		}
		subject, _ := token.Claims.GetSubject()
		if subject != cast.ToString(claims.GetCredentialUID()) {
			return nil, fmt.Errorf("unknown subject:%s", subject)
		}
		return []byte(signKey), nil
	})
	//if isAdmin {
	//	pp.Println(1111, claims.(*auth.AdminClaims))
	//} else {
	//	pp.Println(2222, claims.(*auth.SvcClaims))
	//}
	if err != nil {
		return nil, xerr.ErrUnauthorized.AppendMsg(err.Error())
	}
	if !t.Valid {
		return nil, xerr.ErrUnauthorized.AppendMsg("invalid token")
	}
	if !claims.IsValidCredential() {
		return nil, xerr.ErrUnauthorized.AppendMsg("invalid claims")
	}
	return claims, err
}
