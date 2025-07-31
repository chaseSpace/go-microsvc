package user

import (
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

	"github.com/k0kubun/pp/v3"

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

	_, _ = rpcext.Admin().UpdateUserGold(tbase.TestCallCtx, &adminpb.UpdateUserGoldReq{
		Base:   tbase.TestBaseAdminReq,
		Uid:    1,
		Delta:  100,
		Remark: "xxx",
	})
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

func TestListBar(t *testing.T) {
	defer tbase.TearDown()
	r, _ := rpcext.Admin().ListBar(tbase.TestCallCtx, &adminpb.ListBarReq{
		Base:     tbase.TestBaseAdminReq,
		SearchId: 2,
		//SearchName: "Amor",
		//SearchState:    0,
		//SearchInsId:    "",
		//SearchCity:     "",
		//SearchMinStars: 1.5,
		Page: &commonpb.PageArgs{
			Pn:         1,
			Ps:         2,
			IsDownload: false,
		},
		Sort: []*commonpb.Sort{
			{
				OrderField: "created_at",
				OrderType:  commonpb.OrderType_OT_Desc,
			},
		},
	})
	pp.Print(r.List, r.Total)
}

func TestAddBar(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Admin().AddBar(tbase.TestCallCtx, &adminpb.AddBarReq{
		Base: tbase.TestBaseAdminReq,
		Bar: &commonpb.Bar{
			Name:       "x31",
			State:      commonpb.BarState_BS_Running,
			WebsiteUrl: "url",
			BizHours: []*commonpb.BizHours{
				{
					WStart: 0,
					WEnd:   2,
					HRange: "18:00-23:00",
				},
			},
			InstagramId: "ins",
			LocState:    "NY",
			LocCity:     "York",
			LocAddr:     "addr",
			Desc:        "desc",
			ConsumeNote: "note",
			Stars:       1,
			CoverUrl:    "url",
			Photos:      []string{"url2"},
			Phone:       "011000",
			Popularity:  112,
			Geometry: &commonpb.Geometry{
				Lng: "1.1",
				Lat: "2",
			},
		},
	})
}

func TestUpdateBar(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Admin().UpdateBar(tbase.TestCallCtx, &adminpb.UpdateBarReq{
		Base: tbase.TestBaseAdminReq,
		Bar: &commonpb.Bar{
			Id:         2,
			Name:       "x1",
			State:      2,
			WebsiteUrl: "",
			BizHours: []*commonpb.BizHours{
				{
					WStart: 0,
					WEnd:   2,
					HRange: "18:00-23:00",
				},
				{
					WStart: 3,
					WEnd:   3,
					HRange: "19:00-02:00",
				},
			},
			InstagramId: "ins",
			LocState:    "NY",
			LocCity:     "York",
			LocAddr:     "addr",
			Desc:        "desc",
			ConsumeNote: "cn",
			Stars:       0,
			CoverUrl:    "curl",
			Photos:      []string{"x"},
			Phone:       "010",
			Geometry: &commonpb.Geometry{
				Lng: "1.1",
				Lat: "2.3",
			},
		},
	})
}

func TestDelBar(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Admin().DelBar(tbase.TestCallCtx, &adminpb.DelBarReq{
		Base:  tbase.TestBaseAdminReq,
		BarId: 1,
	})
}

func TestListWine(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Admin().ListWine(tbase.TestCallCtx, &adminpb.ListWineReq{
		Base:             tbase.TestBaseAdminReq,
		SearchId:         1,
		SearchName:       "",
		SearchBarId:      0,
		SearchType:       0,
		SearchType2:      0,
		SearchState:      0,
		SearchCustomCate: "",
		SearchPrice:      &commonpb.FloatRange{
			//Min: 0,
			//Max: 222,
		},
		//SearchCurrencyType: commonpb.CurrencyType_CT_USD,
		//SearchMinStars:     2,
		Page: &commonpb.PageArgs{
			Pn:         1,
			Ps:         2,
			IsDownload: false,
		},
		Sort: []*commonpb.Sort{
			{
				OrderField: "created_at",
				OrderType:  commonpb.OrderType_OT_Desc,
			},
		},
	})
}

func TestAddWine(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Admin().AddWine(tbase.TestCallCtx, &adminpb.AddWineReq{
		Base: tbase.TestBaseAdminReq,
		Wine: &commonpb.Wine{
			Id:           1,
			BarId:        2,
			Name:         "2",
			Type:         2,
			State:        1,
			Tag:          []string{"t1", "t2"},
			Desc:         "desc",
			CustomCate:   "cc",
			Price:        1.1,
			CurrencyType: 2,
			Photos:       []string{"u1"},
			Stars:        1.5,
			Popularity:   123,
		},
	})
}

func TestUpdateWine(t *testing.T) {
	defer tbase.TearDown()
	_, err := rpcext.Admin().UpdateWine(tbase.TestCallCtx, &adminpb.UpdateWineReq{
		Base: tbase.TestBaseAdminReq,
		Wine: &commonpb.Wine{
			Id:           1,
			BarId:        2,
			Name:         "221",
			Type:         2,
			State:        1,
			Tag:          []string{"t1", "t2"},
			Desc:         "desc",
			CustomCate:   "cc",
			Price:        11,
			CurrencyType: 2,
			Photos:       []string{"u1"},
			Stars:        1.5,
		},
	})
	pp.Print(err)
}

func TestDelWine(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Admin().DelWine(tbase.TestCallCtx, &adminpb.DelWineReq{
		Base:   tbase.TestBaseAdminReq,
		WineId: 3,
	})
}

func TestListEvent(t *testing.T) {
	defer tbase.TearDown()
	r, err := rpcext.Admin().ListEvent(tbase.TestCallCtx, &adminpb.ListEventReq{
		Base: tbase.TestBaseAdminReq,
		//SearchId: 1,
		//SearchName:       "n",
		//SearchBarId:      11,
		//SearchSrcAccount: "acc",
		SearchSrcPubTimeRange: &commonpb.TimeRange{
			//StartDt: "2023-11-01 11:11:12",
			EndDt: "",
		},
		SearchTimeStartRange: &commonpb.TimeRange{
			StartDt: "",
			//EndDt:   "2023-11-01 11:11:12",
		},
		SearchTimeEndRange: &commonpb.TimeRange{
			StartDt: "",
			//EndDt:   "2023-11-01 11:11:12",
		},
		SearchCreatedAtRange: &commonpb.TimeRange{
			//StartDt: "2023-11-01 11:11:12",
			//EndDt:   "2023-11-01 11:11:12",
		},
		//SearchPubState: commonpb.EventPubState_EPS_UnPublished,
		SearchCooperateType: commonpb.EventCooperateType_ECT_External,
		Page: &commonpb.PageArgs{
			Pn:         1,
			Ps:         2,
			IsDownload: false,
		},
		Sort: []*commonpb.Sort{
			{
				OrderField: "created_at",
				OrderType:  commonpb.OrderType_OT_Desc,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	pp.Print(r.List, r.Total)
}

func TestAddEvent(t *testing.T) {
	defer tbase.TearDown()
	_, err := rpcext.Admin().AddEvent(tbase.TestCallCtx, &adminpb.AddEventReq{
		Base: tbase.TestBaseAdminReq,
		Event: &commonpb.BarEvent{
			EditBarId:         1,
			EditName:          "E10",
			EditDesc:          "desc",
			EditLocation:      "loc",
			EditPhotos:        []string{"x1", "x2"},
			EditTimeStart:     time.Now().Unix(),
			EditTimeEnd:       time.Now().Add(time.Hour).Unix(),
			EditPubState:      commonpb.EventPubState_EPS_UnPublished,
			EditCooperateType: commonpb.EventCooperateType_ECT_WithUs,
			EditPriceCurrency: commonpb.CurrencyType_CT_CNY,
			EditTicketLimit:   1,
			EditIsRecommend:   true,
			RuntimeTicketSold: 1,
			Src:               nil,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateEvent(t *testing.T) {
	defer tbase.TearDown()
	_, err := rpcext.Admin().UpdateEvent(tbase.TestCallCtx, &adminpb.UpdateEventReq{
		Base: tbase.TestBaseAdminReq,
		Event: &commonpb.BarEvent{
			Id:                3,
			EditBarId:         222,
			EditName:          "Event Name 3",
			EditDesc:          "",
			EditLocation:      "",
			EditPhotos:        nil,
			EditTimeStart:     0,
			EditTimeEnd:       0,
			EditPubState:      commonpb.EventPubState_EPS_UnPublished,
			EditCooperateType: commonpb.EventCooperateType_ECT_External,
			EditTicketLimit:   2,
			EditIsRecommend:   true,
			RuntimeTicketSold: 1,
			Src:               nil,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDelEvent(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Admin().DelEvent(tbase.TestCallCtx, &adminpb.DelEventReq{
		Base:    tbase.TestBaseAdminReq,
		EventId: 3,
	})
}

func TestListEventOrder(t *testing.T) {
	defer tbase.TearDown()
	r, err := rpcext.Admin().ListEventOrder(tbase.TestCallCtx, &adminpb.ListEventOrderReq{
		Base: tbase.TestBaseAdminReq,
		SearchCtimeRange: &commonpb.TimeRange{
			StartDt: "",
			EndDt:   "",
		},
		Page: &commonpb.PageArgs{
			Pn:         1,
			Ps:         2,
			IsDownload: false,
		},
		Sort: []*commonpb.Sort{
			{
				OrderField: "state",
				OrderType:  0,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	pp.Println(r.List, r.Total, r.TotalPaidMoney, r.TotalPaidOrders, r.TotalUnpaidOrders)
}

func TestDeleteEventOrder(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Admin().DeleteEventOrder(tbase.TestCallCtx, &adminpb.DeleteEventOrderReq{
		Base:    tbase.TestBaseAdminReq,
		OrderNo: "250122-2110820507",
	})
}
