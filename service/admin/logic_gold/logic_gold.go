package logic_gold

import (
	"context"
	"fmt"
	"microsvc/bizcomm/auth"
	"microsvc/infra/svccli/rpc"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/currencypb"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) UpdateUserGold(ctx context.Context, req *adminpb.UpdateUserGoldReq) (*adminpb.UpdateUserGoldRes, error) {
	caller := auth.ExtractAdminUser(ctx)
	txType := currencypb.GoldTxType_GSTT_AdminIncr
	if req.Delta < 0 {
		txType = currencypb.GoldTxType_GSTT_AdminDecr
	}
	res, err := rpc.Currency().UpdateUserGold(ctx, &currencypb.UpdateUserGoldReq{
		Uid:    req.Uid,
		Delta:  req.Delta,
		TxType: txType,
		Remark: fmt.Sprintf("admin(%d-%s): %s", caller.Uid, caller.Nickname, req.Remark),
	})
	return &adminpb.UpdateUserGoldRes{Inner: res}, err
}

func (ctrl) GetGiftList(ctx context.Context, caller *auth.AdminCaller, req *adminpb.GetGiftListReq) (*adminpb.GetGiftListRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("Field `inner` is required")
	}
	res, err := rpc.Gift().GetGiftListInt(ctx, req.Inner)
	return &adminpb.GetGiftListRes{
		Inner: res,
	}, err
}
