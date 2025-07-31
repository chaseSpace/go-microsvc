package dao

import (
	"context"
	"microsvc/bizcomm/comminfra"
	"microsvc/model/svc/currency"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/currencypb"
	"microsvc/util/db"
)

type goldTxLogT struct{}

var GoldTxLogDao = goldTxLogT{}

func (*goldTxLogT) orderFieldMap() db.OrderFieldMap {
	return db.OrderFieldMap{
		"id":         &struct{}{},
		"created_at": &struct{}{},
		"amount":     &struct{}{},
	}
}

func (g *goldTxLogT) GetGoldTxLog(ctx context.Context, uid int64, req *currencypb.GetGoldTxLogReq) (list []*currency.GoldTxLog, total int64, err error) {
	m := new(currency.GoldTxLog)
	m.SetSuffix(req.YearMonth)

	if !comminfra.HasTable(currency.Q.DB, m.TableName()) {
		return
	}
	q := currency.Q.WithContext(ctx).Table(m.TableName()).Where("uid = ?", uid)
	if req.TxType > currencypb.GoldTxType_GSTT_Unknown {
		q = q.Where("tx_type = ?", req.TxType)
	}
	var orderBy string
	orderBy, err = db.GenSortClause([]db.Sort{&commonpb.Sort{
		OrderField: req.OrderField,
		OrderType:  req.OrderType,
	}, db.IdDescFn()}, g.orderFieldMap())
	if err != nil {
		return
	}
	err = db.PageQuery(q, req.Page, orderBy, &total, &list)
	return
}
