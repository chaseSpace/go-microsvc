package xgrpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	MdKeyAuthToken       = "authorization" // store token for authentication
	MdKeyTraceId         = "trace-id"      // store trace id
	MdKeyFromGatewayFlag = "from-gateway"
	MdKeyFakeAuth        = "fake-auth"
	MdKeyTestCall        = "test-call"
	MdKeyBizRemoteAddr   = "biz-remote-addr" // client-ip  | e.g. 127.0.0.1:2222
	MdKeyUseProtobuf     = "use-protobuf"
)

const MdKeyFlagExist = "1"

func transferMDInsideOfCtx(ctx context.Context, method string, key ...string) (context.Context, error) {
	ctx, err := sutil.setupCtx(ctx, method)
	if err != nil {
		return nil, err
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		md2 := metadata.New(nil)
		for _, k := range key {
			if len(md[k]) > 0 {
				md2[k] = md[k]
			}
		}
		return metadata.NewOutgoingContext(ctx, md2), nil
	}
	return ctx, nil
}

// getOutgoingMdVal should be used in client side
func getOutgoingMdVal(ctx context.Context, key string) string {
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		if ss := md.Get(key); len(ss) > 0 {
			return ss[0]
		}
	}
	return ""
}

// getIncomingMdVal should be used in server side
func getIncomingMdVal(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if ss := md.Get(key); len(ss) > 0 {
			return ss[0]
		}
	}
	return ""
}

func isIncomingMdKeyExist(ctx context.Context, key string) bool {
	return getIncomingMdVal(ctx, key) == MdKeyFlagExist
}

// GetReqClientIP e.g. 127.0.0.1:1166
// 不可能为空
func GetReqClientIP(ctx context.Context) string {
	ip := ctx.Value(CtxServerSideKey{}).(CtxServerSideVal).ClientIP
	if ip == "" {
		panic("no client ip, plz check")
	}
	return ip
}

func IsProtobufData(ctx context.Context) bool {
	return getIncomingMdVal(ctx, MdKeyUseProtobuf) == MdKeyFlagExist
}
