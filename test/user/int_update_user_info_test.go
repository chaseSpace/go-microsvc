package user

import (
	"microsvc/enums"
	"microsvc/infra/svccli/rpc"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/userpb"
	deploy2 "microsvc/service/user/deploy"
	"microsvc/test/tbase"
	"microsvc/util/ucrypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateUserInfoInt_Nickname(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	uid := int64(100010)
	oldNick := "user1" // 先手动修改数据库用户的昵称
	defer adminUpdateNickname(t, uid, oldNick)

	assertNicknameEqual(t, uid, oldNick)

	var err error
	// case-1 uid=0
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid:       0,
		BodyArray: []*userpb.UpdateBody{},
	})
	assert.True(t, xerr.ErrUserNotFound.Equal(err))

	// case-2 未指定更新字段
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid:       uid,
		BodyArray: []*userpb.UpdateBody{},
	})
	assert.Equal(t, xerr.ErrParams.New("请指定更新字段"), err)

	// case-3 字段未变化
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Nickname,
				AnyValue:  oldNick,
			},
		},
	})
	assert.Equal(t, xerr.ErrParams.New("昵称未变化"), err)

	// case-4 更新成功
	newNick := "user2"
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Nickname,
				AnyValue:  newNick,
			},
		},
	})
	assert.Equal(t, nil, err)
	assertNicknameEqual(t, uid, newNick)

	// case-5 允许更新频繁(2s内)
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Nickname,
				AnyValue:  newNick + "xxx",
			},
		},
	})
	assert.Equal(t, nil, err)
	assertNicknameEqual(t, uid, newNick+"xxx")
}

func TestUpdateUserInfoInt_Passwd(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	var err error
	uid := int64(100010) // 不需要提前设置数据表中此用户的密码（admin不验证旧密码）

	// case-1 新密码空
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
			},
		},
	})
	assert.Equal(t, xerr.ErrPasswdFormat, err)

	// case-2 新密码不是哈希（长度不是40位）
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  "|123",
			},
		},
	})
	assert.Equal(t, xerr.ErrPasswdFormat, err)

	// case-3 新密码合格
	newPass := "123"
	s, _ := ucrypto.Sha1(newPass, "")
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  "|" + s,
			},
		},
	})
	assert.Nil(t, err)

	// case-4 更换新密码（数据库有旧密码，但旧密码参数随意设置，admin不验证）
	newPass = "456"
	s2, _ := ucrypto.Sha1(newPass, "")
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  ".|" + s2,
			},
		},
	})
	assert.Nil(t, err)

	// case-5 admin允许设置相同密码
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  s2 + "|" + s2,
			},
		},
	})
	assert.Nil(t, err)
}

func TestUpdateUserInfoInt_ClearPasswd(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	var err error
	uid := int64(100010) // 提前清空数据表中此用户的密码

	// case-1 无密码进行清空
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_ClearPassword,
			},
		},
	})
	assert.Equal(t, xerr.ErrParams.New("密码未设置，无需重置"), err)

	// -- 设置密码
	s2, _ := ucrypto.Sha1("newPass", "")
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_ClearPassword,
				AnyValue:  s2,
			},
		},
	})

	// case-2 有密码进行清空
	_, err = rpc.User().AdminUpdateUserInfo(tbase.TestCallCtx, &userpb.AdminUpdateUserInfoReq{
		Uid: uid,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_ClearPassword,
			},
		},
	})
	assert.Nil(t, err)
}
