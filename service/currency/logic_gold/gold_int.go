package logic_gold

import (
	"context"
	"microsvc/protocol/svc/currencypb"
	"microsvc/service/currency/dao"
)

type intCtrl struct {
}

var Int intCtrl

func (intCtrl) GetGoldAccount(ctx context.Context, req *currencypb.GetGoldAccountIntReq) (*currencypb.GetGoldAccountIntRes, error) {
	row, err := dao.GoldAccountDao.GetGoldAccount(ctx, nil, req.Uid)
	if err != nil {
		return nil, err
	}
	return &currencypb.GetGoldAccountIntRes{
		Balance:       row.Balance,
		RechargeTotal: row.RechargeTotal,
	}, nil
}

func (intCtrl) UpdateUserGold(ctx context.Context, req *currencypb.UpdateUserGoldReq) (*currencypb.UpdateUserGoldRes, error) {
	params := &dao.GoldTxParams{
		UID:    req.Uid,
		Delta:  req.Delta,
		TxType: req.TxType,
		Remark: req.Remark,
	}
	_, err := dao.GoldTxLogDao.ExecuteGoldTransaction(ctx, params)
	return new(currencypb.UpdateUserGoldRes), err
}
