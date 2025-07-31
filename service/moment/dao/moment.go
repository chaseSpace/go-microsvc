package dao

import (
	"context"
	"microsvc/enums"
	"microsvc/model"
	"microsvc/model/svc/moment"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/momentpb"
	"microsvc/util/db"
	"time"

	"gorm.io/gorm"
)

type momentDao struct {
}

var MomentDao momentDao

func (momentDao) CreateMoment(ctx context.Context, uid int64, req *momentpb.CreateMomentReq) (mid int64, err error) {
	row := &moment.Moment{
		FieldUID:  model.FieldUID{UID: uid},
		Text:      req.Text,
		Type:      req.Type,
		MediaUrls: req.MediaUrls,
	}
	err = moment.Q.WithContext(ctx).Create(row).Error
	err = xerr.WrapMySQL(err)
	return row.Id, err
}

func (momentDao) GetMoment(ctx context.Context, mid int64, must bool) (ent *moment.Moment, err error) {
	err = moment.Q.WithContext(ctx).Where("deleted_at IS NULL").Take(&ent, "id = ?", mid).Error
	if must && !db.IsMysqlErr(err) && (ent == nil || ent.Id == 0) {
		return nil, xerr.ErrMomentNotFound
	}
	return ent, xerr.WrapMySQL(err)
}

func (momentDao) DeleteMoment(ctx context.Context, uid, mid int64) error {
	// 软删除
	err := moment.Q.WithContext(ctx).Model(&moment.Moment{}).Where("uid = ? and mid = ?", uid, mid).
		Update("deleted_at", time.Now()).Error
	return xerr.WrapMySQL(err)
}

func (momentDao) UpdateReviewStatus(ctx context.Context, uid, mid int64, status momentpb.ReviewStatus, passAt int64) error {
	// NOTE: 即使是审核不过，也不要删除数据（除非用户自己删除）
	err := moment.Q.WithContext(ctx).Model(moment.Moment{}).
		Where("uid = ? and id = ?", uid, mid).
		Update("review_status", status).
		Update("review_pass_at", passAt).Error
	return xerr.WrapMySQL(err)
}

func (momentDao) GetMomentCommentNum(ctx context.Context, midSlice []int64) (numMap map[int64]int64, err error) {
	type Temp struct {
		Mid, Ct int64
	}
	var list []*Temp
	err = moment.Q.WithContext(ctx).Model(&moment.MomentComment{}).Where("mid in (?)", midSlice).
		Where("deleted_at IS NULL").
		Group("mid").
		Select("mid, COUNT(1) AS ct").
		Scan(&list).Error
	if err != nil {
		return nil, xerr.WrapMySQL(err)
	}
	numMap = make(map[int64]int64)
	for _, v := range list {
		numMap[v.Mid] = v.Ct
	}
	return
}

func (momentDao) LikeMoment(ctx context.Context, caller int64, isLike bool, mid int64) (err error) {
	q := moment.Q.WithContext(ctx).Model(&moment.Moment{}).
		Where("id=?", mid).
		Where("deleted_at IS NULL")
	if db.Count(q) == 0 {
		return xerr.ErrMomentNotFound
	}
	if isLike {
		err = q.Update("likes", db.JSONSet("likes", caller, nil)).Error
	} else {
		err = q.Update("likes", db.JSONRemove("likes", caller)).Error
	}
	return xerr.WrapMySQL(err)
}

func (momentDao) CommentMoment(ctx context.Context, uid, mid, replyUID int64, content string) (err error) {
	err = moment.Q.WithContext(ctx).Create(&moment.MomentComment{
		FieldUID: model.FieldUID{UID: uid},
		Mid:      mid,
		ReplyUID: replyUID,
		Content:  content,
	}).Error
	err = xerr.WrapMySQL(err)
	return
}

func (momentDao) ForwardMoment(ctx context.Context, mid int64) (err error) {
	err = moment.Q.WithContext(ctx).Model(&moment.Moment{}).
		Where("id = ?", mid).
		Update("forwards", gorm.Expr("forwards+1")).Error
	return xerr.WrapMySQL(err)
}

func (momentDao) GetComment(ctx context.Context, req *momentpb.GetCommentReq) (list []*moment.MomentComment, total int64, err error) {
	q := moment.Q.WithContext(ctx).Model(&moment.MomentComment{}).Where("mid = ?", req.Mid)
	orderBy, err := db.GenSortClause(req.Sort, map[string]*struct{}{"created_at": {}})
	if err != nil {
		return
	}
	err = db.PageQuery(q, req.Page, orderBy, &total, &list)
	err = xerr.WrapMySQL(err)
	return
}

func (momentDao) ListFollowMoment(ctx context.Context, uid, lastIndex, ps int64) (list []*moment.Moment, err error) {
	// NOTE: 可以看见自己发布的审核中的动态
	q := moment.Q.WithContext(ctx).Model(moment.Moment{}).
		Joins("LEFT JOIN friend on moment.uid = friend.fid").
		Order("moment.review_pass_at desc, moment.id desc").
		Limit(int(ps)).
		Where("friend.uid = ?", uid).
		Where("moment.deleted_at IS NULL").
		Where("moment.review_status = ?", momentpb.ReviewStatus_RS_Pass).
		Or("moment.uid = ? and moment.review_status = ?", uid, momentpb.ReviewStatus_RS_Pending)
	if lastIndex > 0 {
		q = q.Where("moment.review_pass_at < ?", lastIndex) // 因为是降序
	}
	err = q.Select("DISTINCT moment.*").Find(&list).Error
	err = xerr.WrapMySQL(err)
	return
}

func (momentDao) ListLatestMoment(ctx context.Context, lastIndex, ps int64, sex enums.Sex) (list []*moment.Moment, err error) {
	q := moment.Q.WithContext(ctx).Model(moment.Moment{}).
		Joins("LEFT JOIN user on moment.uid = user.uid").
		Order("review_pass_at desc, id desc").
		Limit(int(ps))

	if lastIndex > 0 {
		// 此字段可以省略 review_status 字段的过滤
		q = q.Where("review_pass_at between ? and ?", 0, lastIndex) // 因为是降序
	}
	// 查异性动态
	if sex == enums.SexFemale {
		q = q.Where("user.sex = ?", enums.SexMale)
	} else {
		q = q.Where("user.sex = ?", enums.SexFemale)
	}
	err = q.Select("moment.*").Find(&list).Error
	err = xerr.WrapMySQL(err)
	return
}
