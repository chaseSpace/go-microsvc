package user

import (
	"microsvc/enums"
	"microsvc/infra/svccli/rpcext"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/momentpb"
	"microsvc/service/friend/deploy"
	"microsvc/test/tbase"
	"testing"
)

func init() {
	tbase.TearUp(enums.SvcMoment, deploy.FriendConf)
}

func TestListFollowMoment(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Moment().ListFollowMoment(tbase.TestCallCtx, &momentpb.ListFollowMomentReq{
		Base:      tbase.TestBaseExtReq,
		LastIndex: 0,
		PageSize:  20,
	})
}

func TestLikeMoment(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Moment().LikeMoment(tbase.TestCallCtx, &momentpb.LikeMomentReq{
		Base:   tbase.TestBaseExtReq,
		Mid:    1,
		IsLike: true,
	})
}

func TestCommentMoment(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Moment().CommentMoment(tbase.TestCallCtx, &momentpb.CommentMomentReq{
		Base:     tbase.TestBaseExtReq,
		Mid:      1,
		ReplyUid: 12,
		Content:  "hello",
	})
}

func TestForwardMoment(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Moment().ForwardMoment(tbase.TestCallCtx, &momentpb.ForwardMomentReq{
		Base: tbase.TestBaseExtReq,
		Mid:  1,
	})
}

func TestGetComment(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Moment().GetComment(tbase.TestCallCtx, &momentpb.GetCommentReq{
		Base: tbase.TestBaseExtReq,
		Mid:  1,
		Sort: []*commonpb.Sort{
			{
				OrderField: "created_at",
				OrderType:  commonpb.OrderType_OT_Desc,
			},
		},
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 5,
		},
	})
}
func TestListLatestMoment(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Moment().ListLatestMoment(tbase.TestCallCtx, &momentpb.ListLatestMomentReq{
		Base:      tbase.TestBaseExtReq,
		LastIndex: 0,
		PageSize:  20,
	})
}
