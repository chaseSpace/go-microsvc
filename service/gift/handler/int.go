package handler

import (
	"context"
	"microsvc/protocol/svc/giftpb"
	"microsvc/service/gift/logic"
)

var IntCtrl giftpb.GiftIntServer = new(intCtrl)

type intCtrl struct {
}

func (g intCtrl) GetGiftListInt(ctx context.Context, req *giftpb.GetGiftListIntReq) (*giftpb.GetGiftListIntRes, error) {
	return logic.Int.GetGiftListInt(ctx, req)
}

func (g intCtrl) SaveGiftItem(ctx context.Context, req *giftpb.SaveGiftItemReq) (*giftpb.SaveGiftItemRes, error) {
	return logic.Int.SaveGiftItem(ctx, req)
}

func (g intCtrl) DelGiftItem(ctx context.Context, req *giftpb.DelGiftItemReq) (*giftpb.DelGiftItemRes, error) {
	return logic.Int.DelGiftItem(ctx, req)
}

func (g intCtrl) GetUserGiftTxLog(ctx context.Context, req *giftpb.GetUserGiftTxLogReq) (*giftpb.GetUserGiftTxLogRes, error) {
	return logic.Int.GetUserGiftTxLog(ctx, req)
}
