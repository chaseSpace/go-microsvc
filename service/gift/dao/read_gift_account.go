package dao

import (
	"context"
	"microsvc/model/svc/gift"
	"microsvc/pkg/xerr"

	"gorm.io/gorm"
)

func GetAccountAllGifts(ctx context.Context, tx *gorm.DB, uid int64) (list []*gift.GiftAccount, gmap map[int64]gift.GiftAccount, err error) {
	if tx == nil {
		tx = gift.Q.DB
	}
	err = tx.WithContext(ctx).Find(&list, "uid = ? and amount > 0", uid).Error
	gmap = make(map[int64]gift.GiftAccount, len(list))
	for _, v := range list {
		gmap[v.GiftID] = *v
	}
	return list, gmap, xerr.WrapMySQL(err)
}

func GetAccountSingleGift(ctx context.Context, tx *gorm.DB, giftId int64, uid ...int64) (list []*gift.GiftAccount, gmap map[int64]gift.GiftAccount, err error) {
	if tx == nil {
		tx = gift.Q.DB
	}
	err = tx.WithContext(ctx).Find(&list, "uid IN (?) and gift_id = ?", uid, giftId).Error
	gmap = make(map[int64]gift.GiftAccount, len(list))
	for _, v := range list {
		gmap[v.UID] = *v
	}
	return list, gmap, xerr.WrapMySQL(err)
}
