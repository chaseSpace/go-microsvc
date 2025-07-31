package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/currencypb"
	"microsvc/service/currency/logic_gold"
)

var Ctrl currencypb.CurrencyExtServer = new(ctrl)

type ctrl struct{}

func (ctrl) GetGoldAccount(ctx context.Context, req *currencypb.GetGoldAccountReq) (*currencypb.GetGoldAccountRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic_gold.Ext.GetGoldAccount(ctx, caller, req)
}

func (ctrl) GetGoldTxLog(ctx context.Context, req *currencypb.GetGoldTxLogReq) (*currencypb.GetGoldTxLogRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic_gold.Ext.GetGoldTxLog(ctx, caller, req)
}

func (ctrl) TestGoldTx(ctx context.Context, req *currencypb.TestGoldTxReq) (*currencypb.TestGoldTxRes, error) {
	return logic_gold.Ext.TestGoldTx(ctx, auth.ExtractSvcUser(ctx), req)
}
