package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/bizcomm/commgw"
	"microsvc/enums"
	"microsvc/infra/svccli"
	"microsvc/infra/xgrpc"
	"microsvc/infra/xgrpc/protocodec"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/gateway/deploy"
	"microsvc/service/gateway/logic/logic_ws"
	"microsvc/service/gateway/logic/logic_wsmanage"
	"microsvc/util"
	"microsvc/util/ufile"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cast"

	"github.com/hashicorp/go-uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/valyala/bytebufferpool"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type GatewayCtrl struct {
}

// 不宜过短
const forwardTimeout = time.Second * 10

var (
	interceptors = []UnaryInterceptor{
		interceptor{}.OuterMost,
		interceptor{}.AddTrace,
		interceptor{}.Logging,
	}
)

func (c GatewayCtrl) Handler(ctx *fasthttp.RequestCtx) {
	addInterceptor(c.__handle, interceptors...)(ctx)
}

// ----------------------------------------------------------------

var (
	routeRegexToSvc = regexp.MustCompile(`^(\w+)/(\w+)$`)
)

func (c GatewayCtrl) __handle(fctx *fasthttp.RequestCtx) ([]byte, error) {
	fctx.SetStatusCode(200)
	fctx.Response.Header.Set("x-gateway-forward", "true")

	fullPath := string(fctx.Path())
	if !strings.HasPrefix(fullPath, deploy.PathProxyPrefix) {
		switch fullPath {
		case deploy.PathPing: // for health check.
			return []byte("pong"), nil
		case deploy.PathWS:
			return c.__handleLongConn(fctx)
		default:
			if strings.HasPrefix(fullPath, deploy.PathUploads) {
				return c.__handleFile(fctx)
			}
		}
		return nil, xerr.ErrNotFound.New("path must start with %s", deploy.PathProxyPrefix)
	}

	return c.__handleShortConn(fctx)
}

func (c GatewayCtrl) __handleFile(fctx *fasthttp.RequestCtx) (buf []byte, err error) {
	// 支持本地图片访问
	path := string(fctx.Path())
	buf, err = os.ReadFile("." + path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, xerr.ErrNotFound.New("file not found")
		}
		return nil, xerr.ErrUnknown.Append(err)
	}
	fctx.Response.Header.Set("Content-Length", cast.ToString(len(buf)))

	// 图片在浏览器自动预览
	if is, _, err := ufile.IsImageFile(buf); err != nil {
		return nil, xerr.ErrInternal.Append(err)
	} else if is {
		fctx.Response.Header.Set("Content-Type", "image/jpeg")
		return buf, nil
	}
	// 非图片则下载
	fctx.Response.Header.Set("Content-Type", "application/octet-stream")
	return
}

// 长链接：ws
func (c GatewayCtrl) __handleLongConn(fctx *fasthttp.RequestCtx) ([]byte, error) {
	up := websocket.FastHTTPUpgrader{
		HandshakeTimeout: time.Second * 3,
		// buffer size 不会限制消息大小
		ReadBufferSize:  1024 * 4,
		WriteBufferSize: 1024 * 4,
	}

	err := up.Upgrade(fctx, func(conn *websocket.Conn) {
		var cid string
		var err error

		defer func() {
			if err != nil {
				// client 会收到这个 close msg，其中包含了明确的错误信息！
				replyMsg := util.TruncateBytes([]byte(err.Error()), 100) // ws控制帧的消息长度不能过长
				err2 := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, string(replyMsg)), time.Now().Add(logic_ws.WriteWait))
				xlog.Error("ws upgrade failed", zap.Error(err), zap.NamedError("err2", err2), zap.String("conn-id", cid))
			}
			_ = conn.Close() // 这里必须关闭
		}()

		cid, err = uuid.GenerateUUID()
		if err != nil {
			err = errors.Wrap(err, "conn-id gen failed")
			return
		}

		// 用户登录后拿到token，来建立长连接
		token := string(fctx.Request.Header.Peek("token"))
		if token == "" {
			err = errors.Wrap(err, "token not found in header")
			return
		}

		// 对token鉴权，获取UID
		var claims auth.AuthenticateUserAPI
		claims, err = xgrpc.ParseSvcAuthToken(token, false)
		if err != nil {
			return
		}

		detail, err := parseLoginDetailFromReqHeader(fctx)
		if err != nil {
			return
		}
		ws := logic_ws.New(claims.GetCredentialUID(), cid, conn, detail)
		err = logic_wsmanage.WsManager.AddConn(ws)
		if err != nil { // 一定是程序bug
			err = errors.Wrap(err, "ws add conn failed")
			return
		}

		defer logic_wsmanage.WsManager.DeleteConn(cid)

		go ws.Writer()
		go ws.ProcessReportMsg()
		ws.Reader()
	})

	return nil, xerr.ErrInternal.New("ws upgrade failed").AutoAppend(err)
}

// 短连接：转发到微服务
func (c GatewayCtrl) __handleShortConn(fctx *fasthttp.RequestCtx) (respBuf []byte, err error) {
	if string(fctx.Method()) != fasthttp.MethodPost {
		return nil, xerr.ErrParams.New("only POST method is allowed")
	}

	// 解析path
	fullPath := string(fctx.Path())
	dstPath := fullPath[len(deploy.PathProxyPrefix):]
	items := routeRegexToSvc.FindStringSubmatch(dstPath)
	if len(items) != 3 {
		return nil, xerr.ErrNotFound.New("unknown request path: %s", fullPath)
	}

	dst := &deploy.ForwardDestination{
		Svc:  enums.Svc(items[1]),
		Path: items[2],
	}

	fctx.SetUserValue(deploy.CtxKeyFromGateway, false) // 标志着请求即将转发到子服务

	var reply = bytebufferpool.Get()
	defer bytebufferpool.Put(reply)

	var requestDone = make(chan int)
	ctx, cancel := newRPCCtx(fctx)
	defer cancel()

	// 异步是为了在网关这管理超时！
	go func() {
		conn := svccli.GetConn(dst.Svc)
		subType := protocodec.JSONByteBuffer // web
		fctx.SetContentType("application/json")
		if isProtobufReq(fctx) {
			subType = protocodec.PBByteBuffer // mobile
			fctx.SetContentType(`application/octet-stream`)
		}
		err = conn.Invoke(ctx, dst.GetGRPCPath(), fctx.PostBody(), reply, grpc.CallContentSubtype(subType))
		close(requestDone)
	}()

	select {
	case <-fctx.Done():
		return nil, xerr.ErrGateway.Append(fctx.Err())
	case <-ctx.Done():
		return nil, xerr.ErrBizTimeout
	case <-requestDone:
		return reply.Bytes(), err
	}
}

func newRPCCtx(fctx *fasthttp.RequestCtx) (context.Context, context.CancelFunc) {

	traceId, _ := fctx.Value(xgrpc.MdKeyTraceId).(string)

	md := metadata.Pairs(
		xgrpc.MdKeyAuthToken, string(fctx.Request.Header.Peek(auth.HeaderKey)),
		xgrpc.MdKeyTraceId, traceId,
		xgrpc.MdKeyFromGatewayFlag, xgrpc.MdKeyFlagExist,
		xgrpc.MdKeyBizRemoteAddr, fctx.RemoteAddr().String(),
	)

	if isProtobufReq(fctx) {
		md.Append(xgrpc.MdKeyUseProtobuf, xgrpc.MdKeyFlagExist)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), forwardTimeout)

	return metadata.NewOutgoingContext(ctx, md), cancel
}

func parseLoginDetailFromReqHeader(fctx *fasthttp.RequestCtx) (r *commgw.LoginDetail, err error) {
	h := &fctx.Request.Header

	ip := fctx.RemoteAddr().String()
	if ip, _, err = net.SplitHostPort(ip); err != nil {
		return
	}
	r = &commgw.LoginDetail{
		Platform: commonpb.SignInPlatform(cast.ToInt8(string(h.Peek("platform")))),
		System:   commonpb.SignInSystem(cast.ToInt8(string(h.Peek("system")))),
		IP:       ip,
	}
	if r.IsInvalid() {
		xlog.Error("invalid login detail alongside websocket request", zap.Any("detail", r))
		return nil, xerr.ErrParams.New("invalid login detail on HTTP header")
	}
	return
}
