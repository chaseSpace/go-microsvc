package logic

import (
	"context"
	"microsvc/infra/svccli/rpc"
	"microsvc/model/svc/gift"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/giftpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/gift/cache"
)

func checkSendGiftReq(ctx context.Context, sender int64, req *giftpb.SendGiftToOneReq) (*gift.GiftConf, error) {
	// 检查静态参数
	if sender == req.ToUid {
		return nil, xerr.ErrGiftTxFromAndToCannotBeSame
	}
	if req.Amount < 1 {
		return nil, xerr.ErrGiftTxAmountMustBePositive
	}
	if req.TxType == giftpb.TxType_TT_Unknown || giftpb.TxType_name[int32(req.TxType)] == "" {
		return nil, xerr.ErrGiftTxInvalidTxType
	}
	if req.TxScene == giftpb.GiftScene_GS_Unknown || giftpb.GiftScene_name[int32(req.TxScene)] == "" {
		return nil, xerr.ErrGiftTxInvalidTxScene
	}

	// 检查动态参数
	res, err := rpc.User().GetUserInfoInt(ctx, &userpb.GetUserInfoIntReq{Uids: []int64{req.ToUid}})
	if err != nil {
		return nil, err
	} else if len(res.Umap) == 0 {
		return nil, xerr.ErrUserNotFound.New("Gift receiver(uid:%d) not found", req.ToUid)
	}
	// check giftID if is valid
	if g, err := cache.GiftMgCtrl.GetAvailableOne(ctx, req.GiftId, req.TxScene); err != nil {
		return nil, err
	} else {
		return g, nil
	}
}
