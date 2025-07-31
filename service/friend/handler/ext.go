package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/friendpb"
	"microsvc/service/friend/logic"
)

var Ctrl friendpb.FriendExtServer = new(ctrl)

type ctrl struct{}

func (ctrl) FriendList(ctx context.Context, req *friendpb.FriendListReq) (*friendpb.FriendListRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.FriendList(ctx, caller, req)
}

func (ctrl) FriendOnewayList(ctx context.Context, req *friendpb.FriendOnewayListReq) (*friendpb.FriendOnewayListRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.FriendOnewayList(ctx, caller, req)
}

func (ctrl) FollowOne(ctx context.Context, req *friendpb.FollowOneReq) (*friendpb.FollowOneRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	if req.IsFollow {
		return logic.Ext.FollowOne(ctx, caller, req)
	}
	return logic.Ext.UnFollowOne(ctx, caller, req)
}

func (ctrl) SearchFriendList(ctx context.Context, req *friendpb.SearchFriendListReq) (*friendpb.SearchFriendListRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.SearchFriendList(ctx, caller, req)
}

func (ctrl) SearchFriendOnewayList(ctx context.Context, req *friendpb.SearchFriendOnewayListReq) (*friendpb.SearchFriendOnewayListRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.SearchFriendOnewayList(ctx, caller, req)
}

func (ctrl) BlockOne(ctx context.Context, req *friendpb.BlockOneReq) (*friendpb.BlockOneRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.BlockOne(ctx, caller, req)
}

func (ctrl) BlockList(ctx context.Context, req *friendpb.BlockListReq) (*friendpb.BlockListRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.BlockList(ctx, caller, req)
}

func (ctrl) RelationWithOne(ctx context.Context, req *friendpb.RelationWithOneReq) (*friendpb.RelationWithOneRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.RelationWithOne(ctx, caller, req)
}

func (ctrl) SaveVisitor(ctx context.Context, req *friendpb.SaveVisitorReq) (*friendpb.SaveVisitorRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.SaveVisitor(ctx, caller, req)
}

func (ctrl) VisitorList(ctx context.Context, req *friendpb.VisitorListReq) (*friendpb.VisitorListRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic.Ext.VisitorList(ctx, caller, req)
}
