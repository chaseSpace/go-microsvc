package handler

import (
	"context"
	"microsvc/protocol/svc/adminpb"
	"microsvc/service/admin/logic_gift"
	"microsvc/service/admin/logic_gold"
	"microsvc/service/admin/logic_misc"
	"microsvc/service/admin/logic_review"
	"microsvc/service/admin/logic_user"
)

var _ adminpb.AdminExtServer = new(ctrl)
var Ctrl = new(ctrl)

type ctrl struct{}

func (c ctrl) UpdateUserGold(ctx context.Context, req *adminpb.UpdateUserGoldReq) (*adminpb.UpdateUserGoldRes, error) {
	return logic_gold.Ext.UpdateUserGold(ctx, req)
}

func (c ctrl) GetGiftList(ctx context.Context, req *adminpb.GetGiftListReq) (*adminpb.GetGiftListRes, error) {
	return logic_gift.Ext.GetGiftList(ctx, req)
}

func (c ctrl) SaveGiftItem(ctx context.Context, req *adminpb.SaveGiftItemReq) (*adminpb.SaveGiftItemRes, error) {
	return logic_gift.Ext.SaveGiftItem(ctx, req)
}

func (c ctrl) DelGiftItem(ctx context.Context, req *adminpb.DelGiftItemReq) (*adminpb.DelGiftItemRes, error) {
	return logic_gift.Ext.DelGiftItem(ctx, req)
}

func (c ctrl) GetUserGiftTxLog(ctx context.Context, req *adminpb.GetUserGiftTxLogReq) (*adminpb.GetUserGiftTxLogRes, error) {
	return logic_gift.Ext.GetUserGiftTxLog(ctx, req)
}

func (c ctrl) NewPunish(ctx context.Context, req *adminpb.NewPunishReq) (*adminpb.NewPunishRes, error) {
	return logic_user.Ext.NewPunish(ctx, req)
}

func (c ctrl) IncrPunishDuration(ctx context.Context, req *adminpb.IncrPunishDurationReq) (*adminpb.IncrPunishDurationRes, error) {
	return logic_user.Ext.IncrPunishDuration(ctx, req)
}

func (c ctrl) DismissPunish(ctx context.Context, req *adminpb.DismissPunishReq) (*adminpb.DismissPunishRes, error) {
	return logic_user.Ext.DismissPunish(ctx, req)
}

func (c ctrl) PunishList(ctx context.Context, req *adminpb.PunishListReq) (*adminpb.PunishListRes, error) {
	return logic_user.Ext.PunishList(ctx, req)
}

func (c ctrl) UserPunishLog(ctx context.Context, req *adminpb.UserPunishLogReq) (*adminpb.UserPunishLogRes, error) {
	return logic_user.Ext.UserPunishLog(ctx, req)
}

func (c ctrl) ListUser(ctx context.Context, req *adminpb.ListUserReq) (*adminpb.ListUserRes, error) {
	return logic_user.Ext.ListBizUser(ctx, req)
}

func (c ctrl) ListUserAPICallLog(ctx context.Context, req *adminpb.ListUserAPICallLogReq) (*adminpb.ListUserAPICallLogRes, error) {
	return logic_user.Ext.ListUserAPICallLog(ctx, req)
}

func (c ctrl) ListUserLastSignInLogs(ctx context.Context, req *adminpb.ListUserLastSignInLogsReq) (*adminpb.ListUserLastSignInLogsRes, error) {
	return logic_user.Ext.ListUserLastSignInLogs(ctx, req)
}

func (c ctrl) ListReviewText(ctx context.Context, req *adminpb.ListReviewTextReq) (*adminpb.ListReviewTextRes, error) {
	return logic_review.Ext.ListReviewText(ctx, req)
}

func (c ctrl) ListReviewImage(ctx context.Context, req *adminpb.ListReviewImageReq) (*adminpb.ListReviewImageRes, error) {
	return logic_review.Ext.ListReviewImage(ctx, req)
}

func (c ctrl) ListReviewVideo(ctx context.Context, req *adminpb.ListReviewVideoReq) (*adminpb.ListReviewVideoRes, error) {
	return logic_review.Ext.ListReviewVideo(ctx, req)
}

func (c ctrl) ListReviewAudio(ctx context.Context, req *adminpb.ListReviewAudioReq) (*adminpb.ListReviewAudioRes, error) {
	return logic_review.Ext.ListReviewAudio(ctx, req)
}

func (c ctrl) UpdateReviewStatus(ctx context.Context, req *adminpb.UpdateReviewStatusReq) (*adminpb.UpdateReviewStatusRes, error) {
	return logic_review.Ext.UpdateReviewStatus(ctx, req)
}

func (c ctrl) ConfigCenterAdd(ctx context.Context, req *adminpb.ConfigCenterAddReq) (*adminpb.ConfigCenterAddRes, error) {
	return logic_misc.Ext.ConfigCenterAdd(ctx, req)
}

func (c ctrl) ConfigCenterList(ctx context.Context, req *adminpb.ConfigCenterListReq) (*adminpb.ConfigCenterListRes, error) {
	return logic_misc.Ext.ConfigCenterList(ctx, req)
}

func (c ctrl) ConfigCenterUpdate(ctx context.Context, req *adminpb.ConfigCenterUpdateReq) (*adminpb.ConfigCenterUpdateRes, error) {
	return logic_misc.Ext.ConfigCenterUpdate(ctx, req)
}

func (c ctrl) ConfigCenterDelete(ctx context.Context, req *adminpb.ConfigCenterDeleteReq) (*adminpb.ConfigCenterDeleteRes, error) {
	return logic_misc.Ext.ConfigCenterDelete(ctx, req)
}

func (c ctrl) SwitchCenterAdd(ctx context.Context, req *adminpb.SwitchCenterAddReq) (*adminpb.SwitchCenterAddRes, error) {
	return logic_misc.Ext.SwitchCenterAdd(ctx, req)
}

func (c ctrl) SwitchCenterList(ctx context.Context, req *adminpb.SwitchCenterListReq) (*adminpb.SwitchCenterListRes, error) {
	return logic_misc.Ext.SwitchCenterList(ctx, req)
}

func (c ctrl) SwitchCenterUpdate(ctx context.Context, req *adminpb.SwitchCenterUpdateReq) (*adminpb.SwitchCenterUpdateRes, error) {
	return logic_misc.Ext.SwitchCenterUpdate(ctx, req)
}

func (c ctrl) SwitchCenterDelete(ctx context.Context, req *adminpb.SwitchCenterDeleteReq) (*adminpb.SwitchCenterDeleteRes, error) {
	return logic_misc.Ext.SwitchCenterDelete(ctx, req)
}
