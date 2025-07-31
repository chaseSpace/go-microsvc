package user

import (
	"microsvc/enums"
	"microsvc/infra/svccli/rpc"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/giftpb"
	"microsvc/service/gift/deploy"
	"microsvc/test/tbase"
	"testing"
)

func init() {
	tbase.TearUp(enums.SvcGift, deploy.GiftConf)
}

func TestSaveGiftItem_Add(t *testing.T) {
	defer tbase.TearDown()

	rpc.Gift().SaveGiftItem(tbase.TestCallCtx, &giftpb.SaveGiftItemReq{
		Meta: &giftpb.Gift{
			Id:              1,
			Name:            "G1",
			Price:           01,
			Type:            0,
			Icon:            "ICON",
			SupportedScenes: []giftpb.GiftScene{giftpb.GiftScene_GS_IM},
		},
		IsAdd: true,
		//Status: 0,
	})
}

func TestSaveGiftItem_Update(t *testing.T) {
	defer tbase.TearDown()

	rpc.Gift().SaveGiftItem(tbase.TestCallCtx, &giftpb.SaveGiftItemReq{
		Meta: &giftpb.Gift{
			Id:              1,
			Name:            "G1",
			Price:           1,
			Type:            0,
			Icon:            "ICON1",
			SupportedScenes: []giftpb.GiftScene{giftpb.GiftScene_GS_Room},
		},
		IsAdd: false,
		State: giftpb.GiftState_GS_On,
	})
}

func TestDelGiftItem(t *testing.T) {
	defer tbase.TearDown()

	rpc.Gift().DelGiftItem(tbase.TestCallCtx, &giftpb.DelGiftItemReq{Id: 1})
}

func TestGetUserGiftTxLog(t *testing.T) {
	defer tbase.TearDown()

	rpc.Gift().GetUserGiftTxLog(tbase.TestCallCtx, &giftpb.GetUserGiftTxLogReq{
		//SearchFromUid: 1,
		//SearchToUid:   2,
		//SearchScenes: []giftpb.GiftScene{giftpb.GiftScene_GS_IM},
		//SearchGiftName: "xx",
		//SearchAmount: 2,
		//SearchGiftTypes: []giftpb.GiftType{giftpb.GiftType_GT_Normal},
		//SearchTxTypes: []giftpb.TxType{giftpb.TxType_TT_Purchase},
		//SearchMinPrice: 1,
		//SearchMaxPrice: 2,
		//SearchMinTotalValue: 2,
		Sort: []*commonpb.Sort{{
			OrderField: "created_at",
			OrderType:  commonpb.OrderType_OT_Asc,
		}, {
			OrderField: "total_value",
			OrderType:  commonpb.OrderType_OT_Asc,
		},
		},
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 2,
		},
	})
}
