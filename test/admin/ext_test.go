package user

import (
	"github.com/stretchr/testify/assert"
	"microsvc/bizcomm/auth"
	deploy2 "microsvc/deploy"
	"microsvc/enums"
	"microsvc/infra/svccli/rpcext"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/giftpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/admin/deploy"
	"microsvc/test/tbase"
	"microsvc/util"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
)

func init() {
	tbase.TearUp(enums.SvcAdmin, deploy.AdminConf)
}

func TestGenAdminToken(t *testing.T) {
	if deploy2.XConf.AdminTokenSignKey == "" {
		panic("admin-key is empty")
	}
	now := time.Now()
	uid := 1
	token, err := auth.GenerateJwT(
		&auth.SvcClaims{
			SvcCaller: auth.SvcCaller{
				Credential: auth.Credential{
					Uid:      1,
					Nickname: "admin",
					Sex:      enums.SexMale,
					LoginAt:  time.Now().Format(time.DateTime),
					RegAt:    time.Now().Format(time.DateTime),
				},
			},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: nil, //never expire
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
				Issuer:    auth.TokenIssuer,
				Subject:   cast.ToString(uid),
				ID:        util.NewKsuid(),
			},
		}, deploy2.XConf.AdminTokenSignKey)

	util.AssertNilErr(err)
	t.Logf(token)
}

func TestUpdateUserGold(t *testing.T) {
	defer tbase.TearDown()

	_, err := rpcext.Admin().UpdateUserGold(tbase.TestCallCtx, &adminpb.UpdateUserGoldReq{
		Base:   tbase.TestBaseAdminReq,
		Uid:    1,
		Delta:  100,
		Remark: "xxx",
	})
	assert.Nil(t, err)
}

func TestGetGiftList(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().GetGiftList(tbase.TestCallCtx, &adminpb.GetGiftListReq{
		Base: tbase.TestBaseAdminReq,
		Inner: &giftpb.GetGiftListIntReq{
			Sort: &commonpb.Sort{
				OrderField: "created_at",
				OrderType:  commonpb.OrderType_OT_Desc,
			},
			Page: &commonpb.PageArgs{
				Pn: 1,
				Ps: 5,
			},
		},
	})
}

func TestSaveGiftItem(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().SaveGiftItem(tbase.TestCallCtx, &adminpb.SaveGiftItemReq{
		Base: tbase.TestBaseAdminReq,
		Inner: &giftpb.SaveGiftItemReq{
			IsAdd: true,
			Meta: &giftpb.Gift{
				Name:            "xxx",
				Price:           100,
				Type:            giftpb.GiftType_GT_Normal,
				Icon:            "icon",
				SupportedScenes: []giftpb.GiftScene{giftpb.GiftScene_GS_IM},
			},
		},
	})
}

func TestDelGiftItem(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().DelGiftItem(tbase.TestCallCtx, &adminpb.DelGiftItemReq{
		Base: tbase.TestBaseAdminReq,
		Inner: &giftpb.DelGiftItemReq{
			Id: 3,
		},
	})
}

func TestGetUserGiftTxLog(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().GetUserGiftTxLog(tbase.TestCallCtx, &adminpb.GetUserGiftTxLogReq{
		Base: tbase.TestBaseAdminReq,
		Inner: &giftpb.GetUserGiftTxLogReq{
			Page: &commonpb.PageArgs{
				Pn: 1,
				Ps: 5,
			},
		},
	})
}

func TestNewPunish(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().NewPunish(tbase.TestCallCtx, &adminpb.NewPunishReq{
		Base: tbase.TestBaseAdminReq,
		Inner: &userpb.NewPunishReq{
			Uid:      1,
			Duration: 100,
			Reason:   "222",
			Type:     commonpb.PunishType_PT_Ban,
			AdminUid: 12,
		},
	})
}

func TestConfigCenterAdd(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().ConfigCenterAdd(tbase.TestCallCtx, &adminpb.ConfigCenterAddReq{
		Base: tbase.TestBaseAdminReq,
		Item: &commonpb.ConfigItemCore{
			Key:                "2",
			Name:               "1",
			Value:              "1",
			IsLock:             true,
			AllowProgramUpdate: true,
		},
	})
}

func TestConfigCenterList(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().ConfigCenterList(tbase.TestCallCtx, &adminpb.ConfigCenterListReq{
		Base: tbase.TestBaseAdminReq,
		Name: "1",
		//Key: "spider:ins.scrape_bar_activity.cookie",
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 2,
		},
	})
}

func TestConfigCenterUpdate(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().ConfigCenterUpdate(tbase.TestCallCtx, &adminpb.ConfigCenterUpdateReq{
		Base: tbase.TestBaseAdminReq,
		Item: &commonpb.ConfigItemCore{
			Key:                "2",
			Name:               "1",
			Value:              "123",
			IsLock:             false,
			AllowProgramUpdate: true,
		},
	})
}

func TestConfigCenterDelete(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().ConfigCenterDelete(tbase.TestCallCtx, &adminpb.ConfigCenterDeleteReq{
		Base: tbase.TestBaseAdminReq,
		Key:  "2",
	})
}

func TestSwitchCenterAdd(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().SwitchCenterAdd(tbase.TestCallCtx, &adminpb.SwitchCenterAddReq{
		Base: tbase.TestBaseAdminReq,
		Core: &commonpb.SwitchItemCore{
			Key:      "123",
			Name:     "1",
			Value:    12,
			ValueExt: map[int32]string{12: "x"},
			IsLock:   false,
		},
	})
}

func TestSwitchCenterList(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().SwitchCenterList(tbase.TestCallCtx, &adminpb.SwitchCenterListReq{
		Base: tbase.TestBaseAdminReq,
		//Name: "1",
		//Key: "1",
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 2,
		},
	})
}

func TestSwitchCenterUpdate(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().SwitchCenterUpdate(tbase.TestCallCtx, &adminpb.SwitchCenterUpdateReq{
		Base: tbase.TestBaseAdminReq,
		Core: &commonpb.SwitchItemCore{
			Key:    "1",
			Name:   "x",
			Value:  commonpb.SwitchValue_ST_On,
			IsLock: true,
		},
	})
}

func TestSwitchCenterDelete(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().SwitchCenterDelete(tbase.TestCallCtx, &adminpb.SwitchCenterDeleteReq{
		Base: tbase.TestBaseAdminReq,
		Key:  "123",
	})
}

func TestListUser(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Admin().ListUser(tbase.TestCallCtx, &adminpb.ListUserReq{
		Base:      tbase.TestBaseAdminReq,
		SearchUid: 0,
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 20,
		},
	})
}

func TestListUserLastSignInLogs(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Admin().ListUserLastSignInLogs(tbase.TestCallCtx, &adminpb.ListUserLastSignInLogsReq{
		Base:  tbase.TestBaseAdminReq,
		Uid:   100010,
		Limit: 10,
	})
}
