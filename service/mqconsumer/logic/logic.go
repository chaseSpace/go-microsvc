package logic

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/mqconsumerpb"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) Test(ctx context.Context, caller *auth.SvcCaller, req *mqconsumerpb.TestReq) (*mqconsumerpb.TestRes, error) {
	panic(1)
}
