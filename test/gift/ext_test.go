package user

import (
	"microsvc/enums"
	"microsvc/infra/svccli/rpcext"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/giftpb"
	"microsvc/service/gift/deploy"
	"microsvc/test/tbase"
	"testing"
)

func init() {
	tbase.TearUp(enums.SvcGift, deploy.GiftConf)
}

func TestGetGiftList(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Gift().GetGiftList(tbase.TestCallCtx, &giftpb.GetGiftListReq{
		Base:  tbase.TestBaseExtReq,
		Scene: giftpb.GiftScene_GS_IM,
	})
}

func TestSendGiftToOne(t *testing.T) {
	defer tbase.TearDown()
	rpcext.Gift().SendGiftToOne(tbase.TestCallCtx, &giftpb.SendGiftToOneReq{
		Base:    tbase.TestBaseExtReq,
		ToUid:   100010,
		GiftId:  1,
		Amount:  1,
		TxType:  giftpb.TxType_TT_Send,
		TxScene: giftpb.GiftScene_GS_IM,
	})
}

func TestGetMyGiftTxLog(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Gift().GetMyGiftTxLog(tbase.TestCallCtx, &giftpb.GetMyGiftTxLogReq{
		Base:       tbase.TestBaseExtReq,
		OrderField: "created_at",
		OrderType:  commonpb.OrderType_OT_Desc,
		Scene:      giftpb.GiftScene_GS_Unknown,
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 5,
		},
		YearMonth: "202408",
	})
}
