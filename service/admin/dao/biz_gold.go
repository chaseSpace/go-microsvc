package dao

import (
	"context"
	"microsvc/model/svc/currency"
	"microsvc/pkg/xerr"
)

type bizGold struct {
}

var BizGold bizGold

func (*bizGold) GetMultiUserAccount(ctx context.Context, uids []int64) (list []*currency.GoldAccount, err error) {
	err = currency.Q.WithContext(ctx).Take(&list, "uid in (?)", uids).Error
	return list, xerr.WrapMySQL(err)
}
