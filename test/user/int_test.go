package user

import (
	"context"
	"microsvc/enums"
	"microsvc/infra/svccli/rpc"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	deploy2 "microsvc/service/user/deploy"
	"microsvc/test/tbase"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserInfoInt(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rsp, err := rpc.User().GetUserInfoInt(context.TODO(), &userpb.GetUserInfoIntReq{Uids: []int64{1}})
	assert.Nil(t, err)
	if err == nil {
		assert.Equal(t, 1, len(rsp.Umap))
	}

	rsp, err = rpc.User().GetUserInfoInt(context.TODO(), &userpb.GetUserInfoIntReq{})
	assert.Equal(t, xerr.ErrParams.TraceSvc(enums.SvcUser).New("切片字段`uids`长度小于1"), err)
}

func TestAllocateUserNid(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	invalidUID := -1
	rsp, err := rpc.User().AllocateUserNid(context.TODO(), &userpb.AllocateUserNidReq{Uid: int64(invalidUID), Nid: 1111})
	assert.Equal(t, xerr.ErrParams.TraceSvc(enums.SvcUser).New("参数`required_uid`不能小于0，得到%v", invalidUID), err)

	rsp, err = rpc.User().AllocateUserNid(context.TODO(), &userpb.AllocateUserNidReq{Uid: 100010, Nid: 1111})
	assert.Nil(t, err)
	if err == nil {
		assert.NotNil(t, rsp)
	}
}

func TestNewPunish(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpc.User().NewPunish(context.TODO(), &userpb.NewPunishReq{
		Uid:      1,
		Type:     commonpb.PunishType_PT_Chat,
		Reason:   "Reason",
		Duration: 1000,
		AdminUid: 12,
	})
}

func TestIncrPunishDuration(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpc.User().IncrPunishDuration(context.TODO(), &userpb.IncrPunishDurationReq{
		Id:       1,
		Duration: 1000,
		Reason:   "Reason",
		AdminUid: 12,
	})
}

func TestDismissPunish(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpc.User().DismissPunish(context.TODO(), &userpb.DismissPunishReq{
		Id:       1,
		Reason:   "Reason222",
		AdminUid: 12,
	})
}

func TestPunishList(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpc.User().PunishList(context.TODO(), &userpb.PunishListReq{
		SearchUid:      []int64{1},
		SearchType:     []commonpb.PunishType{},
		SearchState:    commonpb.PunishState_PS_InProgress,
		SearchAdminUid: 0,
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 10,
		},
	})
}

func TestUserPunishLog(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpc.User().UserPunishLog(context.TODO(), &userpb.UserPunishLogReq{
		Uid: 1,
	})
}

func TestGetUserPunish(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpc.User().GetUserPunish(context.TODO(), &userpb.GetUserPunishReq{
		Uid:  1,
		Type: 0,
	})
}

func TestReviewProfile(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	rpc.User().ReviewProfile(context.TODO(), &userpb.ReviewProfileReq{
		Uid:     1,
		IsPass:  false,
		Reason:  "xx",
		BizType: commonpb.BizType_RBT_Avatar,
	})
}
