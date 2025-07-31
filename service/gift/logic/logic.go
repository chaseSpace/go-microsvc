package logic

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/model/svc/gift"
	"microsvc/protocol/svc/giftpb"
	"microsvc/service/gift/cache"
	"microsvc/service/gift/dao"

	"github.com/samber/lo"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) GetGiftList(ctx context.Context, caller *auth.SvcCaller, req *giftpb.GetGiftListReq) (*giftpb.GetGiftListRes, error) {
	list, err := cache.GiftMgCtrl.List(ctx, req.Scene)
	if err != nil {
		return nil, err
	}
	// 礼物账户很长时间内可以不用缓存
	_, gmap, err := dao.GetAccountAllGifts(ctx, nil, caller.Uid)
	if err != nil {
		return nil, err
	}
	return &giftpb.GetGiftListRes{
		List: lo.Map(list, func(v *gift.GiftConf, _ int) *giftpb.Gift {
			return v.ToPB(gmap[v.Id].Amount)
		}),
	}, err
}

func (ctrl) SendGiftToOne(ctx context.Context, caller *auth.SvcCaller, req *giftpb.SendGiftToOneReq) (*giftpb.SendGiftToOneRes, error) {
	g, err := checkSendGiftReq(ctx, caller.Uid, req)
	if err != nil {
		return nil, err
	}
	err = dao.GiftTxCtrl.GiftTransaction(ctx, &dao.GiftTxParams{
		FromUID:   caller.Uid,
		ToUID:     req.ToUid,
		GiftID:    req.GiftId,
		Delta:     req.Amount,
		TxType:    req.TxType,
		GiftScene: req.TxScene,
		Remark:    "",
		G:         g,
	})
	return &giftpb.SendGiftToOneRes{}, err
}

func (ctrl) GetMyGiftTxLog(ctx context.Context, caller *auth.SvcCaller, req *giftpb.GetMyGiftTxLogReq) (*giftpb.GetMyGiftTxLogRes, error) {
	list, total, err := dao.GiftTxLogDao.GetPersonalTxLog(ctx, caller.Uid, req)
	if err != nil {
		return nil, err
	}
	return &giftpb.GetMyGiftTxLogRes{
		List:  lo.Map(list, func(v *gift.GiftTxLogPersonal, _ int) *giftpb.GiftPersonalTxLog { return v.ToPB() }),
		Total: total,
	}, err
}
