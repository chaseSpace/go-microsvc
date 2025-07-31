package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/crontaskpb"
	"microsvc/service/crontask/logic"
)

var Ctrl crontaskpb.CrontaskExtServer = new(ctrl)

type ctrl struct {
}

func (t ctrl) Test(ctx context.Context, req *crontaskpb.TestReq) (*crontaskpb.TestRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.Test(ctx, caller, req)
}
