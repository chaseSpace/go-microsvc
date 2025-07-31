package user

import (
	"microsvc/bizcomm/commadmin"
	"microsvc/enums"
	"microsvc/infra/svccli/rpcext"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	deploy2 "microsvc/service/user/deploy"
	"microsvc/test/tbase"
	"microsvc/util/ucrypto"
	"microsvc/util/ujson"
	"testing"
	"time"

	"github.com/spf13/cast"

	"github.com/stretchr/testify/assert"
)

func __setNicknameUpdateRate() {
	_json := `{
		"duration_limit": "2s",
		"date_range_limit": [],
		"banned": false,
		"max_history_len": 1
	}`
	_, err := rpcext.Admin().ConfigCenterAdd(tbase.TestCallCtx, &adminpb.ConfigCenterAddReq{
		Base: tbase.TestBaseAdminReq,
		Item: &commonpb.ConfigItemCore{
			Key:                commadmin.ConfigKeyUpdateRateNickname,
			Name:               "none",
			Value:              _json,
			IsLock:             false,
			AllowProgramUpdate: false,
		},
		IsOverride: true,
	})
	if err != nil {
		panic(err)
	}
}

func TestUpdateUserInfo_Nickname(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	uid := int64(100010)
	oldNick := "user1" // 先手动修改数据库用户的昵称
	adminUpdateNickname(t, uid, oldNick)

	ctx := tbase.NewTestCallCtx(true, uid)

	var err error
	// case-1 uid=0
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base:      tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{},
	})

	assert.True(t, xerr.ErrParams.New("请指定更新字段").TraceSvc(enums.SvcUser).Equal(err))

	// case-2 未指定更新字段
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
	})
	assert.Equal(t, xerr.ErrParams.TraceSvc(enums.SvcUser).New("请指定更新字段"), err)

	// case-3 字段未变化
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Nickname,
				AnyValue:  oldNick,
			},
		},
	})
	assert.Equal(t, xerr.ErrParams.TraceSvc(enums.SvcUser).New("昵称未变化"), err)

	// case-4 更新成功
	newNick := "user2"
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Nickname,
				AnyValue:  newNick,
			},
		},
	})
	assert.Equal(t, nil, err)
	assertNicknameEqual(t, uid, newNick)

	__setNicknameUpdateRate()
	// case-5 更新频繁(2s内)
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Nickname,
				AnyValue:  oldNick + "xxx",
			},
		},
	})
	assert.Equal(t, xerr.ErrForbidden.TraceSvc(enums.SvcUser).New("更新资料过于频繁"), err)

	time.Sleep(time.Second * 2)
	// case-5 暂停一会儿就可以了
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Nickname,
				AnyValue:  "user3",
			},
		},
	})
	assertNicknameEqual(t, uid, "user3")
}

func TestUpdateUserInfo_Passwd(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	uid := int64(100010)

	ctx := tbase.NewTestCallCtx(true, uid)
	// 先重置密码
	adminResetPassword(t, uid)

	var err error

	// case-1 新密码空
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
			},
		},
	})
	assert.Equal(t, xerr.ErrPasswdFormat.TraceSvc(enums.SvcUser), err)

	// case-2 新密码不是哈希（长度不是40位）
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  "|123",
			},
		},
	})
	assert.Equal(t, xerr.ErrPasswdFormat.TraceSvc(enums.SvcUser), err)

	// case-3 新密码合格（数据库记录无旧密码）
	newPass := "123"
	passHash1, _ := ucrypto.Sha1(newPass)
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  "|" + passHash1,
			},
		},
	})
	assert.Nil(t, err)

	// case-4 旧密码参数空（此时数据库有旧密码）
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  "|" + passHash1,
			},
		},
	})
	assert.Equal(t, xerr.ErrParams.TraceSvc(enums.SvcUser).New("旧密码错误"), err)

	// case-5 更换新密码
	newPass = "456"
	passHash2, _ := ucrypto.Sha1(newPass)
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  passHash1 + "|" + passHash2,
			},
		},
	})
	println(11111, passHash1, passHash2)
	assert.Nil(t, err)

	// case-6 相同密码
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Password,
				AnyValue:  passHash2 + "|" + passHash2,
			},
		},
	})
	assert.Equal(t, xerr.ErrPasswdNoChange.TraceSvc(enums.SvcUser), err)
}

func TestUpdateUserInfo_Sex(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	var err error
	uid := int64(100010)
	ctx := tbase.NewTestCallCtx(true, uid)

	sexChange := []commonpb.Sex{commonpb.Sex_Male, commonpb.Sex_Female}
	assertSexEqual(t, uid, sexChange[0]) // 先将确定db性别

	// case-1 性别非法
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Sex,
				AnyValue:  cast.ToString(commonpb.Sex_Unknown),
			},
		},
	})
	assert.Equal(t, err, xerr.ErrParams.New("性别非法"))

	// case-2 性别未变化
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Sex,
				AnyValue:  cast.ToString(sexChange[0]),
			},
		},
	})
	assert.Equal(t, err, xerr.ErrParams.New("性别未变化"))

	// case-3 性别更新
	_, err = rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Sex,
				AnyValue:  cast.ToString(sexChange[1]),
			},
		},
	})
	assert.Nil(t, err)
}

func TestUpdateUserInfo_Tags(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	uid := int64(100025)
	ctx := tbase.NewTestCallCtx(true, uid)

	rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Tags,
				AnyValue:  ujson.MustMarshal2Str([]string{"tag1", "tag2"}),
			},
		},
	})
}

func TestUpdateUserInfo_Phone(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	uid := int64(100025)
	ctx := tbase.NewTestCallCtx(true, uid)

	_, err := rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Phone,
				AnyValue:  "86|18587901111",
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateUserInfo_Email(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	uid := int64(100025)
	ctx := tbase.NewTestCallCtx(true, uid)

	_, err := rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Email,
				AnyValue:  "123@qq.com",
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestUpdateUserInfo_FirstnameLastname(t *testing.T) {
	tbase.TearUp(enums.SvcUser, deploy2.UserConf)
	defer tbase.TearDown()

	uid := int64(100013)
	ctx := tbase.NewTestCallCtx(true, uid)

	_, err := rpcext.User().UpdateUserInfo(ctx, &userpb.UpdateUserInfoReq{
		Base: tbase.TestBaseExtReq,
		BodyArray: []*userpb.UpdateBody{
			{
				FieldType: userpb.UserInfoType_UUIT_Firstname,
				AnyValue:  "f1",
			},
			{
				FieldType: userpb.UserInfoType_UUIT_Lastname,
				AnyValue:  "l1",
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestX(t *testing.T) {
	{
		defer println(1111)
	}
	println(2222)
}
