package logic_punish

import (
	"context"
	"microsvc/model"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/cache"
	"microsvc/service/user/dao"
	"unicode/utf8"

	"github.com/samber/lo"
)

type intCtrl struct{}

var Int = intCtrl{}

func (intCtrl) NewPunish(ctx context.Context, req *userpb.NewPunishReq) (*userpb.NewPunishRes, error) {
	_, err := cache.GetOneUser(ctx, req.Uid)
	if err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(req.Reason) > 100 {
		return nil, xerr.ErrParams.New("理由超过100个字")
	}
	if req.AdminUid < 1 {
		return nil, xerr.ErrParams.New("管理员UID无效")
	}
	params := &user.Punish{
		FieldUID:       model.FieldUID{UID: req.Uid},
		Type:           req.Type,
		Reason:         req.Reason,
		Duration:       req.Duration,
		State:          commonpb.PunishState_PS_InProgress,
		FieldCreatedBy: model.FieldCreatedBy{CreatedBy: req.AdminUid},
		FieldUpdatedBy: model.FieldUpdatedBy{UpdatedBy: req.AdminUid},
	}
	err = dao.Punish.New(ctx, params)
	return new(userpb.NewPunishRes), err
}

func (intCtrl) IncrPunishDuration(ctx context.Context, req *userpb.IncrPunishDurationReq) (*userpb.IncrPunishDurationRes, error) {
	err := dao.Punish.IncrDuration(ctx, req.Id, req.Duration, req.AdminUid, req.Reason)
	return new(userpb.IncrPunishDurationRes), err
}

func (intCtrl) DismissPunish(ctx context.Context, req *userpb.DismissPunishReq) (*userpb.DismissPunishRes, error) {
	err := dao.Punish.Dismiss(ctx, req.Id, req.AdminUid, req.Reason)
	return new(userpb.DismissPunishRes), err
}

func (intCtrl) PunishList(ctx context.Context, req *userpb.PunishListReq) (*userpb.PunishListRes, error) {
	list, total, err := dao.Punish.List(ctx, req)
	return &userpb.PunishListRes{
		List: lo.Map(list, func(item *user.PunishRPC, _ int) *userpb.Punish {
			return item.ToPB()
		}),
		Total: total,
	}, err
}

func (intCtrl) UserPunishLog(ctx context.Context, req *userpb.UserPunishLogReq) (*userpb.UserPunishLogRes, error) {
	list, err := dao.Punish.ListPunishLog(ctx, req.Uid)
	return &userpb.UserPunishLogRes{
		List: lo.Map(list, func(item *user.PunishLogRPC, _ int) *userpb.PunishLog {
			return item.ToPB()
		}),
	}, err
}

func (intCtrl) GetUserPunish(ctx context.Context, req *userpb.GetUserPunishReq) (*userpb.GetUserPunishRes, error) {
	u, err := cache.GetOneUser(ctx, req.Uid)
	if err != nil {
		return nil, err
	}
	rows, err := dao.Punish.GetUserPunish(ctx, req)
	if err != nil {
		return nil, err
	}
	return &userpb.GetUserPunishRes{
		Pmap: lo.SliceToMap(rows, func(item *user.PunishRPC) (int32, *userpb.Punish) {
			item.Nickname = u.Nickname
			return int32(item.Type), item.ToPB()
		}),
	}, nil
}
