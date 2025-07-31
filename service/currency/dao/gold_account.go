package dao

import (
	"context"
	"microsvc/model/svc/currency"
	"microsvc/pkg/xerr"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type goldAccountT struct{}

var GoldAccountDao = goldAccountT{}

func (*goldAccountT) GetGoldAccount(ctx context.Context, tx *gorm.DB, uid int64) (row currency.GoldAccount, err error) {
	if tx == nil {
		tx = currency.Q.DB
	}
	err = tx.WithContext(ctx).Take(&row, "uid = ?", uid).Error
	return row, xerr.WrapMySQL(err)
}

func (*goldAccountT) saveGoldAccount(ctx context.Context, tx *gorm.DB, uid, incr int64) (err error) {
	// insert into ... on duplicate key update id=id
	err = tx.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"balance": gorm.Expr("balance + ?", incr)}),
		}).Create(&currency.GoldAccount{UID: uid, Balance: incr}).Error
	return xerr.WrapMySQL(err)
}
