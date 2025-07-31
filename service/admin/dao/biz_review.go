package dao

import (
	"context"
	"microsvc/model"
	"microsvc/model/svc/admin"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util/db"
)

type reviewDao struct {
}

var ReviewDao reviewDao

func (reviewDao) AddText(ctx context.Context, req *adminpb.AddReviewReq) (err error) {
	ent := &admin.ReviewText{
		FieldUID: model.FieldUID{UID: req.Uid},
		//FieldUpdatedBy:
		Text:    req.Text,
		BizType: req.BizType,
		Status:  req.Status,
	}
	err = admin.QAdmin.WithContext(ctx).Create(ent).Error
	err = xerr.WrapMySQL(err)
	return
}
func (reviewDao) AddImage(ctx context.Context, req *adminpb.AddReviewReq) (err error) {
	ent := &admin.ReviewImage{
		FieldUID: model.FieldUID{UID: req.Uid},
		//FieldUpdatedBy:
		Text:    req.Text,
		BizType: req.BizType,
		Status:  req.Status,
		Urls:    req.MediaUrls,
	}
	err = admin.QAdmin.WithContext(ctx).Create(ent).Error
	err = xerr.WrapMySQL(err)
	return
}
func (reviewDao) AddVideo(ctx context.Context, req *adminpb.AddReviewReq) (err error) {
	ent := &admin.ReviewVideo{
		FieldUID: model.FieldUID{UID: req.Uid},
		//FieldUpdatedBy:
		Text:      req.Text,
		BizType:   req.BizType,
		Status:    req.Status,
		Url:       req.MediaUrls[0],
		BizUniqId: req.BizUniqId,
		ThTaskId:  req.ThTaskId,
	}
	err = admin.QAdmin.WithContext(ctx).Create(ent).Error
	err = xerr.WrapMySQL(err)
	return
}
func (reviewDao) AddAudio(ctx context.Context, req *adminpb.AddReviewReq) (err error) {
	ent := &admin.ReviewAudio{
		FieldUID: model.FieldUID{UID: req.Uid},
		//FieldUpdatedBy:
		Text:      req.Text,
		BizType:   req.BizType,
		Status:    req.Status,
		Url:       req.MediaUrls[0],
		BizUniqId: req.BizUniqId,
		ThTaskId:  req.ThTaskId,
	}
	err = admin.QAdmin.WithContext(ctx).Create(ent).Error
	err = xerr.WrapMySQL(err)
	return
}

func (reviewDao) UpdateText(ctx context.Context, id, uid, updateBy int64, oldS, newS commonpb.ReviewStatus) (hits bool, err error) {
	q := admin.QAdmin.WithContext(ctx).Model(admin.ReviewText{}).Where("id = ? and uid = ?", id, uid).
		Where("status = ?", oldS).
		Updates(map[string]interface{}{
			"status":     newS,
			"updated_by": updateBy,
		})
	return q.RowsAffected > 0, xerr.WrapMySQL(q.Error)
}

func (reviewDao) UpdateImage(ctx context.Context, id, uid, updateBy int64, oldS, newS commonpb.ReviewStatus) (hits bool, err error) {
	q := admin.QAdmin.WithContext(ctx).Model(admin.ReviewImage{}).Where("id = ? and uid = ?", id, uid).
		Where("status = ?", oldS).
		Updates(map[string]interface{}{
			"status":     newS,
			"updated_by": updateBy,
		})
	return q.RowsAffected > 0, xerr.WrapMySQL(q.Error)
}

func (reviewDao) UpdateVideo(ctx context.Context, id, uid, updateBy int64, oldS, newS commonpb.ReviewStatus) (hits bool, err error) {
	q := admin.QAdmin.WithContext(ctx).Model(admin.ReviewVideo{}).Where("id = ? and uid = ?", id, uid).
		Where("status = ?", oldS).
		Updates(map[string]interface{}{
			"status":     newS,
			"updated_by": updateBy,
		})
	return q.RowsAffected > 0, xerr.WrapMySQL(q.Error)
}

func (reviewDao) ListText(ctx context.Context, req *adminpb.ListReviewTextReq) (list []*admin.ReviewText, total int64, err error) {
	q := admin.QAdmin.WithContext(ctx).Model(admin.ReviewText{})
	if req.Id > 0 {
		q = q.Where("id = ?", req.Id)
	} else {
		if req.SearchUid > 0 {
			q = q.Where("uid = ?", req.SearchUid)
		}
		if len(req.StatusArray) > 0 {
			q = q.Where("status in (?)", req.StatusArray)
		}
		if len(req.BizTypeArray) > 0 {
			q = q.Where("biz_type in (?)", req.BizTypeArray)
		}
		if req.SearchAdminUid > 0 {
			q = q.Where("update_by_admin_uid = ?", req.SearchAdminUid)
		}
	}
	err = db.PageQuery(q, req.Page, "created_at, id", &total, &list)
	err = xerr.WrapMySQL(err)
	return
}

func (reviewDao) ListImage(ctx context.Context, req *adminpb.ListReviewImageReq) (list []*admin.ReviewImage, total int64, err error) {
	q := admin.QAdmin.WithContext(ctx).Model(admin.ReviewImage{})
	if req.Id > 0 {
		q = q.Where("id = ?", req.Id)
	} else {
		if req.SearchUid > 0 {
			q = q.Where("uid = ?", req.SearchUid)
		}
		if len(req.StatusArray) > 0 {
			q = q.Where("status in (?)", req.StatusArray)
		}
		if len(req.BizTypeArray) > 0 {
			q = q.Where("biz_type in (?)", req.BizTypeArray)
		}
		if req.SearchAdminUid > 0 {
			q = q.Where("update_by_admin_uid = ?", req.SearchAdminUid)
		}
	}
	err = db.PageQuery(q, req.Page, "created_at, id", &total, &list)
	err = xerr.WrapMySQL(err)
	return
}

func (reviewDao) ListVideo(ctx context.Context, req *adminpb.ListReviewVideoReq) (list []*admin.ReviewVideo, total int64, err error) {
	q := admin.QAdmin.WithContext(ctx).Model(admin.ReviewVideo{})
	if req.Id > 0 {
		q = q.Where("id = ?", req.Id)
	} else {
		if req.SearchUid > 0 {
			q = q.Where("uid = ?", req.SearchUid)
		}
		if len(req.StatusArray) > 0 {
			q = q.Where("status in (?)", req.StatusArray)
		}
		if len(req.BizTypeArray) > 0 {
			q = q.Where("biz_type in (?)", req.BizTypeArray)
		}
		if req.SearchAdminUid > 0 {
			q = q.Where("update_by_admin_uid = ?", req.SearchAdminUid)
		}
		if req.ThTaskId != "" {
			q = q.Where("th_task_id = ?", req.ThTaskId)
		}
	}
	err = db.PageQuery(q, req.Page, "created_at, id", &total, &list)
	err = xerr.WrapMySQL(err)
	return
}

func (reviewDao) ListAudio(ctx context.Context, req *adminpb.ListReviewAudioReq) (list []*admin.ReviewAudio, total int64, err error) {
	q := admin.QAdmin.WithContext(ctx).Model(admin.ReviewAudio{})
	if req.SearchUid > 0 {
		q = q.Where("uid = ?", req.SearchUid)
	}
	if len(req.StatusArray) > 0 {
		q = q.Where("status in (?)", req.StatusArray)
	}
	if len(req.BizTypeArray) > 0 {
		q = q.Where("biz_type in (?)", req.BizTypeArray)
	}
	if req.SearchAdminUid > 0 {
		q = q.Where("update_by_admin_uid = ?", req.SearchAdminUid)
	}
	if req.ThTaskId != "" {
		q = q.Where("th_task_id = ?", req.ThTaskId)
	}
	err = db.PageQuery(q, req.Page, "created_at, id", &total, &list)
	err = xerr.WrapMySQL(err)
	return
}
