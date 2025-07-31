package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/giftpb"
	"microsvc/service/gift/logic"
)

var Ctrl giftpb.GiftExtServer = new(ctrl)

type ctrl struct{}

func (ctrl) GetGiftList(ctx context.Context, req *giftpb.GetGiftListReq) (*giftpb.GetGiftListRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.GetGiftList(ctx, caller, req)
}

func (ctrl) SendGiftToOne(ctx context.Context, req *giftpb.SendGiftToOneReq) (*giftpb.SendGiftToOneRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.SendGiftToOne(ctx, caller, req)
}

func (ctrl) GetMyGiftTxLog(ctx context.Context, req *giftpb.GetMyGiftTxLogReq) (*giftpb.GetMyGiftTxLogRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.GetMyGiftTxLog(ctx, caller, req)
}
