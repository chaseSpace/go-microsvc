package dao

import (
	"context"
	"microsvc/model/svc/friend"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/friendpb"
	"microsvc/util"
	"microsvc/util/db"

	"gorm.io/gorm"
)

func friendsFieldMap() db.OrderFieldMap {
	return map[string]*struct{}{
		"id":         {},
		"intimacy":   {},
		"created_at": {},
	}
}

func FriendList(ctx context.Context, uid int64, req *friendpb.FriendListReq) (list []*friend.Friend, total int64, err error) {
	q := friend.Q.WithContext(ctx).Model(&friend.Friend{})

	// 默认返回 friend.* （应是读取list）
	q = q.Joins("LEFT JOIN friend AS b on friend.fid = b.uid").
		Where("friend.uid = ? and b.fid = ?", uid, uid)

	var orderBy string
	orderBy, err = db.GenSortClause([]db.Sort{&commonpb.Sort{
		OrderField: req.OrderField,
		OrderType:  req.OrderType,
	}, db.IdDescFn()}, friendsFieldMap())
	if err != nil {
		return
	}
	err = db.PageQuery(q, req.Page, orderBy, &total, &list)
	return
}

func OnewayList(ctx context.Context, uid int64, req *friendpb.FriendOnewayListReq) (list []*friend.Friend, total int64, err error) {
	q := friend.Q.WithContext(ctx).Model(friend.Friend{})

	if req.IsFollow {
		q = q.Where("uid = ?", uid)
	} else {
		// 粉丝
		q = q.Where("fid = ?", uid)
	}

	var orderBy string
	orderBy, err = db.GenSortClause([]db.Sort{&commonpb.Sort{
		OrderField: req.OrderField,
		OrderType:  req.OrderType,
	}, db.IdDescFn()}, friendsFieldMap())
	if err != nil {
		return
	}
	err = db.PageQuery(q, req.Page, orderBy, &total, &list)
	return
}

func GetFriendCnt(ctx context.Context, tx *gorm.DB, uid int64) (total int64, err error) {
	err = friend.Q.WithContext(ctx).Model(&friend.Friend{}).
		Joins("LEFT JOIN friend AS b on friend.fid = b.uid").
		Where("friend.uid = ? and b.fid = ?", uid, uid).
		Count(&total).Error
	return total, xerr.WrapMySQL(err)
}

// GetOnewayData 获取单边关系
func GetOnewayData(ctx context.Context, tx *gorm.DB, uid, beFollowUid int64) (rows []*friend.Friend, err error) {
	q := friend.Q.WithContext(ctx)
	if tx != nil {
		q = tx.WithContext(ctx)
	}
	ids := []int64{uid, beFollowUid}
	err = q.Find(&rows, "uid in (?) and fid in (?)", ids, ids).Error
	return rows, xerr.WrapMySQL(err)
}

// GetFriendData 获取好友数据
func GetFriendData(ctx context.Context, tx *gorm.DB, uid, fid int64) (row friend.Friend, err error) {
	q := friend.Q.WithContext(ctx)
	if tx != nil {
		q = tx.WithContext(ctx)
	}
	err = q.Model(&friend.Friend{}).
		Joins("LEFT JOIN friend AS b on friend.fid = b.uid").
		Where("friend.uid = ? and b.fid = ?", uid, uid).
		Where("friend.fid = ?", fid).
		Select("friend.*").Scan(&row).Error
	return row, xerr.WrapMySQL(err)
}

func SearchFriendList(ctx context.Context, uid int64, req *friendpb.SearchFriendListReq) (list []*friend.Friend, err error) {
	var orderBy string
	orderBy, err = db.GenSortClause([]db.Sort{&commonpb.Sort{
		OrderField: req.OrderField,
		OrderType:  req.OrderType,
	}, db.IdDescFn()}, friendsFieldMap())
	if err != nil {
		return
	}
	q := friend.Q.WithContext(ctx).Model(&friend.Friend{}).
		Joins("LEFT JOIN friend AS b on friend.fid = b.uid").
		Joins("LEFT JOIN user on friend.fid = user.uid").
		Where("friend.uid = ? and b.fid = ?", uid, uid)

	if id, ok := util.IsDigestStr(req.Keyword); ok {
		q = q.Where("friend.fid = ?", id)
	} else {
		q = q.Where("user.nickname like ?", "%"+req.Keyword+"%")
	}

	err = q.Order(orderBy).Limit(20).Scan(&list).Error
	return
}

func SearchFriendOnewayList(ctx context.Context, uid int64, req *friendpb.SearchFriendOnewayListReq, limit int) (list []*friend.Friend, err error) {
	var orderBy string
	orderBy, err = db.GenSortClause([]db.Sort{&commonpb.Sort{
		OrderField: req.OrderField,
		OrderType:  req.OrderType,
	}, db.IdDescFn()}, friendsFieldMap())
	if err != nil {
		return
	}
	q := friend.Q.WithContext(ctx).Model(&friend.Friend{})

	if req.IsFollow {
		// 关注列表
		q = q.Joins("LEFT JOIN user on friend.fid = user.uid").Where("friend.uid = ?", uid)
		if id, ok := util.IsDigestStr(req.Keyword); ok {
			q = q.Where("friend.fid = ?", id)
		} else {
			q = q.Where("user.nickname like ?", "%"+req.Keyword+"%")
		}
	} else {
		// 粉丝列表
		q = q.Joins("LEFT JOIN user on friend.uid = user.uid").Where("friend.fid = ?", uid)
		if id, ok := util.IsDigestStr(req.Keyword); ok {
			q = q.Where("friend.uid = ?", id)
		} else {
			q = q.Where("user.nickname like ?", "%"+req.Keyword+"%")
		}
	}

	err = q.Order(orderBy).Limit(limit).Scan(&list).Error
	return
}

func BlockList(ctx context.Context, uid int64, req *friendpb.BlockListReq) (list []*friend.Block, total int64, err error) {
	q := friend.Q.WithContext(ctx).Model(&friend.Block{}).Where("uid = ?", uid)
	err = db.PageQuery(q, req.Page, "created_at desc, id", &total, &list)
	return
}

func GetBlockData(ctx context.Context, tx *gorm.DB, uid int64, bid int64) (rows []*friend.Block, err error) {
	q := friend.Q.WithContext(ctx)
	if tx != nil {
		q = tx.WithContext(ctx)
	}
	ids := []int64{uid, bid}
	err = q.Find(&rows, "uid in (?) and bid in (?)", ids, ids).Error
	return rows, xerr.WrapMySQL(err)
}

func VisitorList(ctx context.Context, uid int64, page *commonpb.PageArgs) (list []*friend.Visitor, total int64, err error) {
	q := friend.Q.WithContext(ctx).Model(&friend.Visitor{}).Where("uid = ?", uid)
	err = db.PageQuery(q, page, "created_at desc, id", &total, &list)
	return list, total, xerr.WrapMySQL(err)
}

func VisitorsTotal(ctx context.Context, uid int64, maxDate int64) (total int64, err error) {
	err = friend.Q.WithContext(ctx).Model(&friend.Visitor{}).
		Where("uid = ? and date <= ?", uid, maxDate).
		Distinct("vid").
		Count(&total).Error
	return total, xerr.WrapMySQL(err)
}

func VisitorsRepeated(ctx context.Context, uid int64, maxDate int64) (total int64, err error) {
	err = friend.Q.WithContext(ctx).Model(&friend.Visitor{}).
		Where("uid = ? and date <= ?", uid, maxDate).
		Group("uid, vid").
		Having("SUM(day_visit_times) > 1").
		Count(&total).Error
	return total, xerr.WrapMySQL(err)
}
