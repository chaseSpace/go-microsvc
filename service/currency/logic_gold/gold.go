package logic_gold

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/model/svc/currency"
	"microsvc/protocol/svc/currencypb"
	"microsvc/service/currency/cache"
	"microsvc/service/currency/dao"

	"github.com/samber/lo"
)

type ctrlExt struct {
}

var Ext ctrlExt

func (ctrlExt) GetGoldAccount(ctx context.Context, caller *auth.SvcCaller, _ *currencypb.GetGoldAccountReq) (*currencypb.GetGoldAccountRes, error) {
	row, err := cache.GoldCtrl.GetGoldAccount(ctx, caller.Uid)
	if err != nil {
		return nil, err
	}
	return &currencypb.GetGoldAccountRes{
		Balance:       row.Balance,
		RechargeTotal: row.RechargeTotal,
	}, nil
}

func (ctrlExt) GetGoldTxLog(ctx context.Context, caller *auth.SvcCaller, req *currencypb.GetGoldTxLogReq) (*currencypb.GetGoldTxLogRes, error) {
	list, total, err := dao.GoldTxLogDao.GetGoldTxLog(ctx, caller.Uid, req)
	if err != nil {
		return nil, err
	}
	return &currencypb.GetGoldTxLogRes{
		List: lo.Map(list, func(item *currency.GoldTxLog, index int) *currencypb.GoldTxLog {
			return item.ToPB()
		}),
		Total: total,
	}, nil
}

func (ctrlExt) TestGoldTx(ctx context.Context, caller *auth.SvcCaller, req *currencypb.TestGoldTxReq) (*currencypb.TestGoldTxRes, error) {
	params := dao.GoldTxParams{
		UID:    req.Uid,
		Delta:  req.Delta,
		TxType: req.TxType,
		Remark: req.Remark,
	}
	txID, err := dao.GoldTxLogDao.ExecuteGoldTransaction(ctx, &params)
	if err != nil {
		return nil, err
	}
	_ = cache.GoldCtrl.ClearGoldAccount(ctx, req.Uid)
	return &currencypb.TestGoldTxRes{TxId: txID}, err
}
