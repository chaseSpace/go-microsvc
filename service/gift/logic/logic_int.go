package logic

import (
	"context"
	"microsvc/model"
	"microsvc/model/svc/gift"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/giftpb"
	"microsvc/service/gift/dao"

	"github.com/samber/lo"
)

var Int intCtrl

type intCtrl struct {
}

func (intCtrl) GetGiftListInt(ctx context.Context, req *giftpb.GetGiftListIntReq) (*giftpb.GetGiftListIntRes, error) {
	list, total, err := dao.GiftConfDao.ListGiftConf(ctx, &dao.ListGMParams{
		Sort:     req.Sort,
		PageArgs: req.Page,
	})
	if err != nil {
		return nil, err
	}
	return &giftpb.GetGiftListIntRes{
		List: lo.Map(list, func(item *gift.GiftConf, _ int) *giftpb.GiftItem {
			return item.ToGiftItem()
		}),
		Total: total,
	}, nil
}

func (intCtrl) SaveGiftItem(ctx context.Context, req *giftpb.SaveGiftItemReq) (*giftpb.SaveGiftItemRes, error) {
	if req.IsAdd && req.Meta.Id > 0 {
		return nil, xerr.ErrIDShouldBeZeroOnAdd
	}
	if !req.IsAdd && req.Meta.Id == 0 {
		return nil, xerr.ErrIDShouldNotBeZeroOnUpdate
	}

	res := &giftpb.SaveGiftItemRes{}
	if req.IsAdd {
		err := dao.AddGiftItem(ctx, &gift.GiftConf{
			Icon:            req.Meta.Icon,
			Name:            req.Meta.Name,
			Price:           req.Meta.Price,
			State:           giftpb.GiftState_GS_Off, // 新增时 默认下架状态
			SupportedScenes: req.Meta.SupportedScenes,
			Type:            req.Meta.Type,
		})
		return res, err
	}

	changed, err := dao.UpdateGiftItem(ctx, &gift.GiftConf{
		TableBase:       model.NewTableBaseFieldID(req.Meta.Id),
		Icon:            req.Meta.Icon,
		Name:            req.Meta.Name,
		Price:           req.Meta.Price,
		State:           req.State,
		SupportedScenes: req.Meta.SupportedScenes,
		Type:            req.Meta.Type,
	})
	if err != nil {
		return nil, err
	}
	if !changed {
		return nil, xerr.ErrNoRowAffectedOnUpdate
	}
	return res, nil
}

func (intCtrl) DelGiftItem(ctx context.Context, req *giftpb.DelGiftItemReq) (*giftpb.DelGiftItemRes, error) {
	if req.Id < 1 {
		return nil, xerr.ErrInvalidID
	}
	deleted, err := dao.DelGiftItem(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if !deleted {
		return nil, xerr.ErrDataNotExist
	}
	return &giftpb.DelGiftItemRes{}, nil
}

func (intCtrl) GetUserGiftTxLog(ctx context.Context, req *giftpb.GetUserGiftTxLogReq) (*giftpb.GetUserGiftTxLogRes, error) {
	list, total, err := dao.GiftTxLogDao.GetTxLog(ctx, req)
	if err != nil {
		return nil, err
	}
	return &giftpb.GetUserGiftTxLogRes{
		List:  lo.Map(list, func(item *gift.GiftTxLog, _ int) *giftpb.GiftTxLogInt { return item.ToIntPB() }),
		Total: total,
	}, err
}
