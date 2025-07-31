package dao

import (
	"microsvc/consts"
	"microsvc/model/svc/currency"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/currencypb"
	"microsvc/util"
	"microsvc/util/db"
	"unicode/utf8"

	"golang.org/x/net/context"
	"gorm.io/gorm"
)

// GoldTxParams 金币：单人交易参数（购买行为）
type GoldTxParams struct {
	UID    int64
	Delta  int64
	TxType currencypb.GoldTxType
	Remark string
}

func (g *goldTxLogT) ExecuteGoldTransaction(ctx context.Context, params *GoldTxParams) (txID string, err error) {
	if err := g.checkParamsSingleTx(params); err != nil {
		return "", err
	}
	err = currency.Q.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txID, err = g.txLogic(ctx, tx, params)
		return err
	})
	return
}

func (g *goldTxLogT) txLogic(ctx context.Context, tx *gorm.DB, params *GoldTxParams) (string, error) {
	var err error
	q := tx.Model(&currency.GoldAccount{}).Where("uid = ?", params.UID)
	if params.Delta < 0 { // --
		q = q.Where("balance >= ?", -params.Delta).Update("balance", gorm.Expr("balance - ?", -params.Delta))
		if q.Error != nil {
			return "", q.Error
		}
		if q.RowsAffected != 1 {
			return "", xerr.ErrTxBalanceNotEnough
		}
	} else { // ++
		err = GoldAccountDao.saveGoldAccount(ctx, tx, params.UID, params.Delta)
		if err != nil {
			return "", err
		}
	}

	acc, err := GoldAccountDao.GetGoldAccount(ctx, tx, params.UID)
	if err != nil {
		return "", err
	}

	ct := acc.UpdatedAt
	row := &currency.GoldTxLog{
		TxId:      util.NewKsuid(),
		UID:       params.UID,
		Delta:     params.Delta,
		Balance:   acc.Balance,
		TxType:    params.TxType,
		Remark:    params.Remark,
		CreatedAt: ct,
	}
	row.SetSuffix(ct.Format("200601"))

	err = db.NewTableHelper(tx, row.DDLSql()).AutoCreateTable(func(tx *gorm.DB) error {
		return tx.Table(row.TableName()).Create(row).Error
	})
	return row.TxId, err
}

func (*goldTxLogT) checkParamsSingleTx(params *GoldTxParams) error {
	if params.UID < 1 {
		return xerr.ErrUserNotFound.AppendMsg("交易UID无效")
	}
	if params.Delta == 0 {
		return xerr.ErrTxAmountShouldNotBeZero
	}
	if utf8.RuneCountInString(params.Remark) > consts.TxRemarkMaxLen {
		return xerr.ErrTxRemarkTooLong
	}
	if params.TxType < 1 || currencypb.GoldTxType_name[int32(params.TxType)] == "" {
		return xerr.ErrTxInvalidTxType
	}
	return nil
}
