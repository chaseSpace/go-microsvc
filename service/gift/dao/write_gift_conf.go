package dao

import (
	"context"
	"microsvc/model/svc/gift"
	"microsvc/pkg/xerr"
	"microsvc/util/db"
)

func checkDuplicateErr(err error) error {
	if err != nil {
		idx := ""
		if db.IsMysqlDuplicateErr(err, &idx) && idx == gift.TableUKGiftConfName {
			return xerr.ErrParams.New("Gift name already exists")
		}
	}
	return xerr.WrapMySQL(err)
}

func AddGiftItem(ctx context.Context, row *gift.GiftConf) (err error) {
	if err = row.Check(); err != nil {
		return
	}
	err = gift.Q.WithContext(ctx).Create(row).Error
	return checkDuplicateErr(err)
}

func UpdateGiftItem(ctx context.Context, row *gift.GiftConf) (changed bool, err error) {
	if err = row.Check(); err != nil {
		return
	}
	do := gift.Q.WithContext(ctx).Where("id = ?", row.Id).Select("*").Omit("created_at").Updates(row)
	return do.RowsAffected == 1, checkDuplicateErr(do.Error)
}

func DelGiftItem(ctx context.Context, id int64) (deleted bool, err error) {
	do := gift.Q.WithContext(ctx).Delete(&gift.GiftConf{}, "id = ?", id)
	return do.RowsAffected == 1, xerr.WrapMySQL(do.Error)
}
