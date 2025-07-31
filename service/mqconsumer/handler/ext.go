package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/mqconsumerpb"
	"microsvc/service/mqconsumer/logic"
)

var Ctrl mqconsumerpb.MqConsumerExtServer = new(ctrl)

type ctrl struct {
}

func (t ctrl) Test(ctx context.Context, req *mqconsumerpb.TestReq) (*mqconsumerpb.TestRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.Test(ctx, caller, req)
}
