package user

import (
	"microsvc/enums"
	"microsvc/infra/svccli/rpc"
	"microsvc/infra/svccli/rpcext"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/test/tbase"
	"microsvc/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertNicknameEqual(t *testing.T, uid int64, nick string) {
	rsp, err := rpcext.User().GetUserInfo(tbase.TestCallCtx, &userpb.GetUserInfoReq{
		Base: tbase.TestBaseExtReq,
		Uids: []int64{uid},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rsp.Umap))

	oldNickname := rsp.Umap[uid].Nickname
	if nick != oldNickname {
		t.Fatalf("nickname not equal, nick:%v old:%v", nick, oldNickname)
	}
}

func adminUpdateNickname(t *testing.T, uid int64, nick string) {
	_, err := rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Nickname,
				AnyValue:  nick},
		},
	})
	util.AssertNilErr(err)
}

func adminResetPassword(t *testing.T, uid int64) {
	_, err := rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_ClearPassword,
				AnyValue:  ""},
		},
	})
	if err != nil {
		if !xerr.ErrParams.New("密码未设置，无需重置").TraceSvc(enums.SvcUser).Equal(err) {
			panic(err)
		}
	}
}

func assertSexEqual(t *testing.T, uid int64, sex commonpb.Sex) {
	rsp, err := rpcext.User().GetUserInfo(tbase.TestCallCtx, &userpb.GetUserInfoReq{
		Base: tbase.TestBaseExtReq,
		Uids: []int64{uid},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rsp.Umap))

	oldSex := rsp.Umap[uid].Sex
	assert.Equal(t, sex, oldSex)
}
