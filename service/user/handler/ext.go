package handler

import (
	"context"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/logic_profile"
	"microsvc/service/user/logic_recommend"
	"microsvc/service/user/logic_signin"
)

var Ctrl userpb.UserExtServer = new(ctrl)

type ctrl struct {
}

func (ctrl) SignInAll(ctx context.Context, req *userpb.SignInAllReq) (*userpb.SignInAllRes, error) {
	return logic_signin.Ext.SignInAll(ctx, req)
}

func (ctrl) SignUpAll(ctx context.Context, req *userpb.SignUpAllReq) (*userpb.SignUpAllRes, error) {
	return logic_signin.Ext.SignUpAll(ctx, req)
}

func (c ctrl) SendSmsCodeOnUser(ctx context.Context, req *userpb.SendSmsCodeOnUserReq) (*userpb.SendSmsCodeOnUserRes, error) {
	return logic_signin.Ext.SendSmsCodeOnUser(ctx, req)
}

func (c ctrl) SendEmailCodeOnUser(ctx context.Context, req *userpb.SendEmailCodeOnUserReq) (*userpb.SendEmailCodeOnUserRes, error) {
	return logic_signin.Ext.SendEmailCodeOnUser(ctx, req)
}

func (ctrl) GetUserInfo(ctx context.Context, req *userpb.GetUserInfoReq) (*userpb.GetUserInfoRes, error) {
	return logic_profile.Ext.GetUserInfo(ctx, req)
}

func (c ctrl) UpdateUserInfo(ctx context.Context, req *userpb.UpdateUserInfoReq) (res *userpb.UpdateUserInfoRes, err error) {
	return logic_profile.Ext.UpdateUserInfo(ctx, req)
}

func (c ctrl) ResetPassword(ctx context.Context, req *userpb.ResetPasswordReq) (*userpb.ResetPasswordRes, error) {
	return logic_profile.Ext.ResetPassword(ctx, req)
}

func (c ctrl) SameCityUsers(ctx context.Context, req *userpb.SameCityUsersReq) (*userpb.SameCityUsersRes, error) {
	return logic_recommend.Ext.SameCityUsers(ctx, req)
}

func (c ctrl) NearbyUsers(ctx context.Context, req *userpb.NearbyUsersReq) (*userpb.NearbyUsersRes, error) {
	return logic_recommend.Ext.NearbyUsers(ctx, req)
}

func (c ctrl) GetRecommendUserDetail(ctx context.Context, req *userpb.GetRecommendUserDetailReq) (*userpb.GetRecommendUserDetailRes, error) {
	return logic_recommend.Ext.GetRecommendUserDetail(ctx, req)
}

func (c ctrl) DoGreet(ctx context.Context, req *userpb.DoGreetReq) (*userpb.DoGreetRes, error) {
	//TODO implement me
	panic("implement me")
}
