package handler

import (
	"microsvc/infra/xgrpc"
	"microsvc/infra/xgrpc/protoutil"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/gateway/deploy"
	"microsvc/util"
	"microsvc/util/utime"
	"net/http"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/spf13/cast"

	"github.com/valyala/bytebufferpool"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type Handler func(ctx *fasthttp.RequestCtx) ([]byte, error)
type UnaryInterceptor func(ctx *fasthttp.RequestCtx, handle Handler) ([]byte, error)

// this interceptor model is same as grpc.UnaryClientInterceptor (onion model)
func addInterceptor(handle func(ctx *fasthttp.RequestCtx) ([]byte, error), interceptor ...UnaryInterceptor) fasthttp.RequestHandler {
	return func(fctx *fasthttp.RequestCtx) {
		resBytes, err := interceptor[0](fctx, getChainUnaryHandler(interceptor, 0, handle))
		if err == nil {
			fctx.SetBody(resBytes) // 网关调用服务只能得到bytes，也只能透传
		} else {
			fromGateway := cast.ToBool(fctx.Value(deploy.CtxKeyFromGateway))
			httpRes, _ := protoutil.WrapResponseOnGateway(nil, err, fromGateway, isProtobufReq(fctx))
			fctx.SetBody(httpRes)
		}
	}
}

// Merge interceptors from many to one
func getChainUnaryHandler(interceptors []UnaryInterceptor, curr int, finalInvoker Handler) Handler {
	if curr == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx *fasthttp.RequestCtx) ([]byte, error) {
		return interceptors[curr+1](ctx, getChainUnaryHandler(interceptors, curr+1, finalInvoker))
	}
}

func isProtobufReq(fctx *fasthttp.RequestCtx) bool {
	return string(fctx.Request.Header.ContentType()) == "application/x-protobuf"
}

// ------------ interceptor define ----------------

type interceptor struct {
}

// OuterMost 最外层拦截器(顺序不可调整)
func (interceptor) OuterMost(fctx *fasthttp.RequestCtx, handler Handler) (res []byte, err error) {
	fctx.SetUserValue(deploy.CtxKeyFromGateway, true)

	{ // CORS
		addCORS(fctx) // 必须在最外层设置
		if string(fctx.Method()) == fasthttp.MethodOptions {
			fctx.Response.SetStatusCode(http.StatusNoContent)
			return
		}
	}

	var buf = bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	return handler(fctx)
}

func (interceptor) AddTrace(fctx *fasthttp.RequestCtx, handler Handler) (res []byte, err error) {
	fctx.SetUserValue(xgrpc.MdKeyTraceId, util.NewKsuid())

	res, err = handler(fctx)
	return res, err
}

func (interceptor) Logging(fctx *fasthttp.RequestCtx, handler Handler) (res []byte, err error) {
	tid := fctx.Value(xgrpc.MdKeyTraceId).(string)
	start := time.Now()

	xlog.Info("logInterceptor_start", zap.ByteString("path", fctx.Path()), zap.String("trace-id", tid))
	//println(222)
	defer func() {
		elapsed := utime.DurationStr(time.Since(start))
		if xerr.IsNil(err) {
			//println(11111, len(res), string(res))
			xlog.Info("handle_ok", zap.ByteString("path", fctx.Path()), zap.String("dur", elapsed), zap.String("trace-id", tid))
		} else {
			xlog.Info("handle_fail", zap.ByteString("path", fctx.Path()), zap.Error(err),
				zap.String("dur", elapsed),
				zap.ByteString("body", fctx.PostBody()),
				zap.String("trace-id", tid))
		}
	}()

	res, err = handler(fctx)
	return res, err
}

// -------------------------------------
var allowedHeader = []string{
	"content-type",
	"accept",
	"referer",
	"user-agent",
	"authorization",
	"x-app-name",
	"x-app-version",
	"x-platform",
	"x-system",
	"x-language",
}

func addCORS(fctx *fasthttp.RequestCtx) {
	//app_name := string(fctx.Request.Header.Peek("app_name"))
	//app_version := string(fctx.Request.Header.Peek("app_version"))
	//platform := string(fctx.Request.Header.Peek("platform"))
	platform := commonpb.SignInPlatform(cast.ToInt8(fctx.Request.Header.Peek("platform")))
	if platform == commonpb.SignInPlatform_SIP_None {
		xlog.Warn("invalid platform", zap.Any("platform", platform))
	} else if commonpb.SignInPlatform_SIP_APP == platform { // app 不需要考虑cors
		return
	}
	allowOrigins := deploy.GatewayConf.Cors.AllowOrigins
	origin := string(fctx.Request.Header.Peek("origin"))
	if lo.Contains(allowOrigins, origin) || lo.Contains(allowOrigins, "*") {
		fctx.Response.Header.Add("Access-Control-Allow-Origin", origin)
	} else {
		xlog.Warn("CORS NOT ALLOWED", zap.String("ORIGIN", origin), zap.Strings("allowed", allowOrigins))
	}
	fctx.Response.Header.Add("Access-Control-Expose-Headers", "x-gateway-forward, authorization")
	fctx.Response.Header.Add("Access-Control-Allow-Headers", strings.Join(allowedHeader, ","))
	fctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	fctx.Response.Header.Add("Access-Control-Allow-Credentials", "true")
}
