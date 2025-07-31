package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/templatepb"
	"microsvc/service/template/logic"
)

var Ctrl templatepb.TemplateExtServer = new(ctrl)

type ctrl struct {
}

func (t ctrl) Test(ctx context.Context, req *templatepb.TestReq) (*templatepb.TestRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.Test(ctx, caller, req)
}
