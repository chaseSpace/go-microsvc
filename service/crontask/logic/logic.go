package logic

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/crontaskpb"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) Test(ctx context.Context, caller *auth.SvcCaller, req *crontaskpb.TestReq) (*crontaskpb.TestRes, error) {
	panic(1)
}
