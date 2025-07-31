package xgrpc

import (
	"context"
	"errors"
	"fmt"
	"microsvc/deploy"
	"microsvc/enums"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util"
	"microsvc/util/ujson"
	"microsvc/util/uregex"
	"regexp"
	"strings"
	"time"

	"github.com/valyala/bytebufferpool"
	"google.golang.org/grpc/peer"

	"github.com/samber/lo"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// -------------------------- clientUtil ------------------------------

type clientUtil struct{}

var cutil = clientUtil{}

func (clientUtil) newCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 1,                //  maximum number of requests allowed to pass through when the breaker is half-open
		Interval:    time.Second * 30, // cyclic period of breaker to clear interval counter, defaults to 0 that indicates the breaker never clear interval counter
		Timeout:     time.Second * 10, // timeout for CircuitBreaker stay open, breaker switch to half-open after `timeout`, default 60s.
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// define the condition of the breaker gets to open state
			//fmt.Printf("ReadyToTrip_xx  %+v\n", counts)
			return counts.ConsecutiveFailures > 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			xlog.Warn(fmt.Sprintf("grpc call - circuit breaker state change: %s, %s -> %s", name, from, to))
		},
	})
}
func (clientUtil) fromGatewayCall(ctx context.Context) bool {
	return getOutgoingMdVal(ctx, MdKeyFromGatewayFlag) == MdKeyFlagExist
}

func (c clientUtil) beautifyReqAndResInClient(ctx context.Context, req, reply interface{}) (reqStr, replyStr string) {
	if cutil.fromGatewayCall(ctx) {
		reqStr = string(req.([]byte))
		replyStr = reply.(*bytebufferpool.ByteBuffer).String()
	} else {
		reqStr = string(ujson.MustMarshal(req))
		replyStr = string(ujson.MustMarshal(reply))
	}
	reqStr = util.TruncateUTF8(reqStr, deploy.XConf.GRPC.LogPrintReqMaxLen, "...省略")
	replyStr = util.TruncateUTF8(replyStr, deploy.XConf.GRPC.LogPrintRespMaxLen, "...省略")
	return
}

func (clientUtil) breakerTakeError(err error) bool {
	var xe *xerr.XErr
	if errors.As(err, &xe) && xe.IsInternal() {
		return true
	}
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	return lo.Contains([]codes.Code{
		codes.Unavailable,
		codes.DeadlineExceeded,
		codes.Aborted,
		codes.FailedPrecondition,
	}, s.Code())
}

// extractFuncName
// input: svc.thirdparty.ThirdpartyInt/VerifySmsCode
// output: ThirdpartyInt/VerifySmsCode
func (clientUtil) extractFuncName(grpcMethod string) string {
	ss := strings.Split(grpcMethod, ".")
	return ss[len(ss)-1]
}

func (clientUtil) extractSvcName(grpcMethod string) string {
	ss := strings.Split(grpcMethod, ".")
	return ss[1]
}

// -------------------------- serverUtil ------------------------------
type serverUtil struct{}

var sutil = serverUtil{}

//var Sutil = serverUtil{}

func (serverUtil) setupCtx(ctx context.Context, method string) (context.Context, error) {
	isExtMethod := false
	fromGateway := getIncomingMdVal(ctx, MdKeyFromGatewayFlag) == MdKeyFlagExist

	// method such as: /svc.userpb.UserExt/SignInAll
	ss := strings.Split(method, "/")
	if len(ss) == 3 {
		if ss[1] != "grpc.health.v1.Health" {
			if strings.HasSuffix(ss[1], "Ext") {
				isExtMethod = true
			} else if !strings.HasSuffix(ss[1], "Int") {
				return nil, fmt.Errorf("illegal grpc method: %s", method)
			}
		}
		bizClientIP := getIncomingMdVal(ctx, MdKeyBizRemoteAddr)
		if !isExtMethod { // 对于Int接口，使用节点ip作为客户端ip
			p, ok := peer.FromContext(ctx)
			if !ok {
				return nil, xerr.ErrInternal.New("no peer context found")
			}
			bizClientIP = p.Addr.String()
		}
		if bizClientIP == "" {
			return nil, xerr.ErrInternal.New("[biz-remote-addr] not found from incoming context")
		}
		ipSlice := strings.Split(bizClientIP, ":")
		// 在ctx中不要存指针类型的值
		ctx = context.WithValue(ctx, CtxServerSideKey{}, CtxServerSideVal{
			IsExtMethod: isExtMethod,
			FromGateway: fromGateway,
			ClientIP:    ipSlice[0],
		})
		return ctx, nil
	}
	return nil, xerr.ErrGateway.New("illegal grpc method: %s", method)
}

func (serverUtil) isExtMethod(ctx context.Context) bool {
	val := ctx.Value(CtxServerSideKey{}).(CtxServerSideVal)
	return val.IsExtMethod
}

func (serverUtil) fromGatewayCall(ctx context.Context) bool {
	return isIncomingMdKeyExist(ctx, MdKeyFromGatewayFlag)
}

func (serverUtil) isTestCall(ctx context.Context) bool {
	return isIncomingMdKeyExist(ctx, MdKeyTestCall)
}

func (serverUtil) checkAdminBaseExtReq(req interface{}) (commonpb.Lang, error) {
	lang := commonpb.Lang_CL_EN
	if r, ok := req.(AdminBaseExtReq); !ok || r.GetBase() == nil {
		return lang, xerr.ErrBaseReq.AppendMsg("missing field:`base`")
	} else {
		base := r.GetBase()
		if base.Platform == commonpb.SignInPlatform_SIP_None || commonpb.SignInPlatform_name[int32(base.Platform)] == "" {
			return lang, xerr.ErrBaseReq.AppendMsg("invalid platform")
		}
		if base.System == commonpb.SignInSystem_SIS_None || commonpb.SignInSystem_name[int32(base.System)] == "" {
			return lang, xerr.ErrBaseReq.AppendMsg("invalid system")
		}
		if base.Language == commonpb.Lang_CL_None || commonpb.Lang_name[int32(base.Language)] == "" {
			return lang, xerr.ErrBaseReq.AppendMsg("invalid language")
		}
		return base.Language, nil
	}
}
func (serverUtil) checkBaseExtReq(req interface{}) (commonpb.Lang, error) {
	lang := commonpb.Lang_CL_EN
	if r, ok := req.(BaseExtReq); !ok || r.GetBase() == nil {
		return lang, xerr.ErrBaseReq.AppendMsg("missing field:`base`")
	} else {
		base := r.GetBase()
		if !uregex.IsVersion(base.AppVersion) {
			return lang, xerr.ErrBaseReq.AppendMsg("invalid app version")
		}
		if base.Platform == commonpb.SignInPlatform_SIP_None || commonpb.SignInPlatform_name[int32(base.Platform)] == "" {
			return lang, xerr.ErrBaseReq.AppendMsg("invalid platform")
		}
		if base.System == commonpb.SignInSystem_SIS_None || commonpb.SignInSystem_name[int32(base.System)] == "" {
			return lang, xerr.ErrBaseReq.AppendMsg("invalid system")
		}
		if base.Language == commonpb.Lang_CL_None || commonpb.Lang_name[int32(base.Language)] == "" {
			return lang, xerr.ErrBaseReq.AppendMsg("invalid language")
		}
		return base.Language, nil
	}
}

type fullMethod struct {
	Svc  enums.Svc
	Ctrl string
	API  string
}

var methodRegex = regexp.MustCompile(`^/svc\.(.*)\.(.*)/(.*)$`)

func (serverUtil) ParseFullMethod(method string) fullMethod {
	res := methodRegex.FindAllStringSubmatch(method, -1)
	return fullMethod{
		Svc:  enums.Svc(res[0][1]),
		Ctrl: res[0][2],
		API:  res[0][3],
	}
}
