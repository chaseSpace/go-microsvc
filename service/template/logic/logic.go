package logic

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/templatepb"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) Test(ctx context.Context, caller *auth.SvcCaller, req *templatepb.TestReq) (*templatepb.TestRes, error) {
	panic(1)
}
