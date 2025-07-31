package dao

import (
	"context"
	"microsvc/model/svc/gift"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/giftpb"
	"microsvc/util/db"
	"strings"

	"gorm.io/gorm"
)

type giftTxLogT struct{}

var GiftTxLogDao = giftTxLogT{}

func (*giftTxLogT) orderFieldMap() db.OrderFieldMap {
	return db.OrderFieldMap{
		"id":          &struct{}{},
		"created_at":  &struct{}{},
		"price":       &struct{}{},
		"total_value": &struct{}{},
	}
}

// GetPersonalTxLog 获取个人礼物流水
func (g *giftTxLogT) GetPersonalTxLog(ctx context.Context, uid int64, req *giftpb.GetMyGiftTxLogReq) (list []*gift.GiftTxLogPersonal, total int64, err error) {
	t := gift.GiftTxLogPersonal{}
	t.SetSuffix(req.YearMonth)

	q := gift.Q.WithContext(ctx).Table(t.TableName()).Where("uid = ?", uid)
	if req.Scene > giftpb.GiftScene_GS_Unknown {
		q = q.Where("gift_scene = ?", req.Scene)
	}
	var orderBy string
	orderBy, err = db.GenSortClause([]db.Sort{&commonpb.Sort{
		OrderField: req.OrderField,
		OrderType:  req.OrderType,
	}}, g.orderFieldMap())
	if err != nil {
		return
	}

	err = db.PageQuery(q, req.Page, orderBy, &total, &list)
	err = xerr.WrapMySQL(db.IgnoreTableNotExist(err))
	return
}

// GetTxLog 获取全部礼物流水（含个人、双人交易）
func (g *giftTxLogT) GetTxLog(ctx context.Context, req *giftpb.GetUserGiftTxLogReq) (list []*gift.GiftTxLog, total int64, err error) {
	q := g.parseSearchParams(gift.Q.WithContext(ctx).Model(&gift.GiftTxLog{}), req)

	req.Sort = append(req.Sort, db.IdAscFn()) // 默认按id升序
	var orderBy string
	orderBy, err = db.GenSortClause(req.Sort, g.orderFieldMap())
	if err != nil {
		return
	}
	err = db.PageQuery(q, req.Page, orderBy, &total, &list)
	err = xerr.WrapMySQL(err)
	return
}

func (*giftTxLogT) parseSearchParams(q *gorm.DB, req *giftpb.GetUserGiftTxLogReq) *gorm.DB {
	if req.SearchFromUid > 0 {
		q = q.Where("from_uid = ?", req.SearchFromUid)
	}
	if req.SearchToUid > 0 {
		q = q.Where("to_uid = ?", req.SearchToUid)
	}
	if req.SearchGiftId > 0 {
		q = q.Where("gift_id = ?", req.SearchGiftId)
	}
	if len(req.SearchScenes) > 0 {
		q = q.Where("gift_scene in (?)", req.SearchScenes)
	}
	req.SearchGiftName = strings.TrimSpace(req.SearchGiftName)
	if req.SearchGiftName != "" {
		q = q.Where("gift_name like ?", "%"+req.SearchGiftName+"%")
	}
	if req.SearchAmount > 0 {
		q = q.Where("amount = ?", req.SearchAmount)
	}
	if len(req.SearchTxTypes) > 0 {
		q = q.Where("tx_type in (?)", req.SearchTxTypes)
	}
	if len(req.SearchGiftTypes) > 0 {
		q = q.Where("gift_type in (?)", req.SearchGiftTypes)
	}
	if req.SearchMinPrice > 0 {
		q = q.Where("price >= ?", req.SearchMinPrice)
	}
	if req.SearchMaxPrice > 0 {
		q = q.Where("price <= ?", req.SearchMaxPrice)
	}
	if req.SearchMinTotalValue > 0 {
		q = q.Where("total_value >= ?", req.SearchMinTotalValue)
	}
	return q
}
