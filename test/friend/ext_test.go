package user

import (
	"microsvc/enums"
	"microsvc/infra/svccli/rpcext"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/friendpb"
	"microsvc/service/friend/deploy"
	"microsvc/test/tbase"
	"testing"
)

func init() {
	tbase.TearUp(enums.SvcFriend, deploy.FriendConf)
}

func TestFriendList(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Friend().FriendList(tbase.TestCallCtx, &friendpb.FriendListReq{
		Base:       tbase.TestBaseExtReq,
		OrderField: "intimacy",
		OrderType:  commonpb.OrderType_OT_Desc,
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 2,
		},
	})
}

func TestFriendFollowList(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Friend().FriendOnewayList(tbase.TestCallCtx, &friendpb.FriendOnewayListReq{
		Base:       tbase.TestBaseExtReq,
		OrderField: "created_at",
		OrderType:  commonpb.OrderType_OT_Desc,
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 2,
		},
		IsFollow: true,
	})
}

func TestFriendFansList(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Friend().FriendOnewayList(tbase.TestCallCtx, &friendpb.FriendOnewayListReq{
		Base:       tbase.TestBaseExtReq,
		OrderField: "created_at",
		OrderType:  commonpb.OrderType_OT_Desc,
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 2,
		},
	})
}

func TestFollowOne(t *testing.T) {
	defer tbase.TearDown()

	_, _ = rpcext.Friend().FollowOne(tbase.TestCallCtx, &friendpb.FollowOneReq{
		Base:      tbase.TestBaseExtReq,
		TargetUid: 2,
		IsFollow:  true,
	})
}

func TestUnFollowOne(t *testing.T) {
	defer tbase.TearDown()

	_, _ = rpcext.Friend().FollowOne(tbase.TestCallCtx, &friendpb.FollowOneReq{
		Base:      tbase.TestBaseExtReq,
		TargetUid: 2,
		IsFollow:  false,
	})
}

func TestSearchFriendList(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Friend().SearchFriendList(tbase.TestCallCtx, &friendpb.SearchFriendListReq{
		Base:       tbase.TestBaseExtReq,
		Keyword:    "u", // 也可搜ID
		OrderType:  commonpb.OrderType_OT_Asc,
		OrderField: "intimacy",
	})
}

func TestSearchFollowList(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Friend().SearchFriendOnewayList(tbase.TestCallCtx, &friendpb.SearchFriendOnewayListReq{
		Base:       tbase.TestBaseExtReq,
		Keyword:    "u",
		IsFollow:   true,
		OrderType:  commonpb.OrderType_OT_Asc,
		OrderField: "intimacy",
	})
}

func TestSearchFansList(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Friend().SearchFriendOnewayList(tbase.TestCallCtx, &friendpb.SearchFriendOnewayListReq{
		Base:       tbase.TestBaseExtReq,
		Keyword:    "22",
		IsFollow:   false,
		OrderType:  commonpb.OrderType_OT_Asc,
		OrderField: "intimacy",
	})
}

func TestBlockOne(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Friend().BlockOne(tbase.TestCallCtx, &friendpb.BlockOneReq{
		Base:      tbase.TestBaseExtReq,
		TargetUid: 2,
		IsBlock:   true,
	})
}

func TestBlockList(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Friend().BlockList(tbase.TestCallCtx, &friendpb.BlockListReq{
		Base: tbase.TestBaseExtReq,
		Page: &commonpb.PageArgs{
			Pn: 2,
			Ps: 2,
		},
	})
}

func TestRelationWithOne(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Friend().RelationWithOne(tbase.TestCallCtx, &friendpb.RelationWithOneReq{
		Base:      tbase.TestBaseExtReq,
		TargetUid: 2,
	})
}

func TestSaveVisitor(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Friend().SaveVisitor(tbase.TestCallCtx, &friendpb.SaveVisitorReq{
		Base:      tbase.TestBaseExtReq,
		TargetUid: 2,
		Seconds:   181,
	})
}

func TestVisitorList(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Friend().VisitorList(tbase.TestCallCtx, &friendpb.VisitorListReq{
		Base: tbase.TestBaseExtReq,
		Page: &commonpb.PageArgs{
			Pn: 2,
			Ps: 2,
		},
	})
}
