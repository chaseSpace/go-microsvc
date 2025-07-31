package handler

import (
	"context"
	"microsvc/protocol/svc/currencypb"
	"microsvc/service/currency/logic_gold"
)

var IntCtrl currencypb.CurrencyIntServer = new(intCtrl)

type intCtrl struct {
}

func (c intCtrl) GetGoldAccount(ctx context.Context, req *currencypb.GetGoldAccountIntReq) (*currencypb.GetGoldAccountIntRes, error) {
	return logic_gold.Int.GetGoldAccount(ctx, req)
}

func (intCtrl) UpdateUserGold(ctx context.Context, req *currencypb.UpdateUserGoldReq) (*currencypb.UpdateUserGoldRes, error) {
	return logic_gold.Int.UpdateUserGold(ctx, req)
}
