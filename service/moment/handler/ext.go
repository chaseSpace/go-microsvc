package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/momentpb"
	"microsvc/service/moment/logic"
)

var Ctrl momentpb.MomentExtServer = new(ctrl)

type ctrl struct {
}

func (c ctrl) CreateMoment(ctx context.Context, req *momentpb.CreateMomentReq) (*momentpb.CreateMomentRes, error) {
	caller := auth.GetAuthUID(ctx)
	return logic.Ext.CreateMoment(ctx, caller, req)
}

func (c ctrl) DeleteMoment(ctx context.Context, req *momentpb.DeleteMomentReq) (*momentpb.DeleteMomentRes, error) {
	caller := auth.GetAuthUID(ctx)
	return logic.Ext.DeleteMoment(ctx, caller, req)
}

func (c ctrl) LikeMoment(ctx context.Context, req *momentpb.LikeMomentReq) (*momentpb.LikeMomentRes, error) {
	caller := auth.GetAuthUID(ctx)
	return logic.Ext.LikeMoment(ctx, caller, req)
}

func (c ctrl) CommentMoment(ctx context.Context, req *momentpb.CommentMomentReq) (*momentpb.CommentMomentRes, error) {
	caller := auth.GetAuthUID(ctx)
	return logic.Ext.CommentMoment(ctx, caller, req)
}

func (c ctrl) ForwardMoment(ctx context.Context, req *momentpb.ForwardMomentReq) (*momentpb.ForwardMomentRes, error) {
	caller := auth.GetAuthUID(ctx)
	return logic.Ext.ForwardMoment(ctx, caller, req)
}

func (c ctrl) GetComment(ctx context.Context, req *momentpb.GetCommentReq) (*momentpb.GetCommentRes, error) {
	return logic.Ext.GetComment(ctx, req)
}

func (c ctrl) ListFollowMoment(ctx context.Context, req *momentpb.ListFollowMomentReq) (*momentpb.ListFollowMomentRes, error) {
	caller := auth.GetAuthUID(ctx)
	return logic.Ext.ListFollowMoment(ctx, caller, req)
}

func (c ctrl) ListRecommendMoment(ctx context.Context, req *momentpb.ListRecommendMomentReq) (*momentpb.ListRecommendMomentRes, error) {
	panic("implement me")
}

func (c ctrl) ListLatestMoment(ctx context.Context, req *momentpb.ListLatestMomentReq) (*momentpb.ListLatestMomentRes, error) {
	return logic.Ext.ListLatestMoment(ctx, auth.ExtractSvcUser(ctx), req)
}
