package handler

import (
	"context"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/logic_profile"
	"microsvc/service/user/logic_punish"
)

type intCtrl struct{}

var IntCtrl userpb.UserIntServer = new(intCtrl)

func (intCtrl) GetUserInfoInt(ctx context.Context, req *userpb.GetUserInfoIntReq) (*userpb.GetUserInfoIntRes, error) {
	return logic_profile.Int.GetUserInfoInt(ctx, req)
}

func (intCtrl) AllocateUserNid(ctx context.Context, req *userpb.AllocateUserNidReq) (*userpb.AllocateUserNidRes, error) {
	return logic_profile.Int.AllocateUserNid(ctx, req)
}

func (intCtrl) AdminUpdateUserInfo(ctx context.Context, req *userpb.AdminUpdateUserInfoReq) (res *userpb.AdminUpdateUserInfoRes, err error) {
	return logic_profile.Int.AdminUpdateUserInfo(ctx, req)
}

func (intCtrl) NewPunish(ctx context.Context, req *userpb.NewPunishReq) (*userpb.NewPunishRes, error) {
	return logic_punish.Int.NewPunish(ctx, req)
}

func (intCtrl) IncrPunishDuration(ctx context.Context, req *userpb.IncrPunishDurationReq) (*userpb.IncrPunishDurationRes, error) {
	return logic_punish.Int.IncrPunishDuration(ctx, req)
}

func (intCtrl) DismissPunish(ctx context.Context, req *userpb.DismissPunishReq) (*userpb.DismissPunishRes, error) {
	return logic_punish.Int.DismissPunish(ctx, req)
}

func (intCtrl) PunishList(ctx context.Context, req *userpb.PunishListReq) (*userpb.PunishListRes, error) {
	return logic_punish.Int.PunishList(ctx, req)
}

func (intCtrl) UserPunishLog(ctx context.Context, req *userpb.UserPunishLogReq) (*userpb.UserPunishLogRes, error) {
	return logic_punish.Int.UserPunishLog(ctx, req)
}

func (intCtrl) GetUserPunish(ctx context.Context, req *userpb.GetUserPunishReq) (*userpb.GetUserPunishRes, error) {
	return logic_punish.Int.GetUserPunish(ctx, req)
}

func (c intCtrl) ReviewProfile(ctx context.Context, req *userpb.ReviewProfileReq) (*userpb.ReviewProfileRes, error) {
	return logic_profile.Int.ReviewProfile(ctx, req)
}
