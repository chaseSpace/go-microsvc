package dao

import (
	"context"
	"microsvc/model/svc/gift"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util/db"
)

type giftConfT struct{}

var GiftConfDao = giftConfT{}

type ListGMParams struct {
	Sort *commonpb.Sort
	*commonpb.PageArgs
}

func (*giftConfT) orderFieldMap() db.OrderFieldMap {
	return db.OrderFieldMap{
		"created_at": &struct{}{},
		"price":      &struct{}{},
	}
}

// ListAllGiftConf 列出所有礼物配置（含下架状态）
func (g *giftConfT) ListAllGiftConf(ctx context.Context) (list []*gift.GiftConf, err error) {
	err = gift.Q.WithContext(ctx).Order("price, id").Find(&list).Error
	return
}

func (g *giftConfT) ListGiftConf(ctx context.Context, params *ListGMParams) (list []*gift.GiftConf, total int64, err error) {
	q := gift.Q.WithContext(ctx).Model(&gift.GiftConf{})

	var orderBy string
	orderBy, err = db.GenSortClause([]db.Sort{&commonpb.Sort{
		OrderField: params.Sort.OrderField,
		OrderType:  params.Sort.OrderType,
	}}, g.orderFieldMap())
	if err != nil {
		return
	}
	err = db.PageQuery(q, params.PageArgs, orderBy, &total, &list)
	return
}
