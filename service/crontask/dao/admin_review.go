package dao

import (
	"context"
	"microsvc/model/svc/admin"
	"microsvc/protocol/svc/commonpb"
	"time"

	"gorm.io/gorm"
)

type adminReviewDao struct {
}

var AdminReviewDao adminReviewDao

func (adminReviewDao) ListReviewVideoPendingRecords(ctx context.Context, limit int) (list []*admin.ReviewVideo, err error) {
	err = admin.QAdmin.WithContext(ctx).
		Where("status = ?", commonpb.ReviewStatus_RS_Pending).
		Where("created_at <= ?", time.Now().Add(-time.Second*5).Unix()).
		Where("query_ret_fails < 3").
		Order("created_at").Limit(limit).Find(&list).Error
	return
}

func (adminReviewDao) UpdateReviewVideoFails(ctx context.Context, id int64) error {
	return admin.QAdmin.WithContext(ctx).
		Model(&admin.ReviewVideo{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"query_ret_fails": gorm.Expr("query_ret_fails + 1"),
			"updated_by":      0, // 0 means system
		}).Error
}

func (adminReviewDao) UpdateReviewVideoStatus(ctx context.Context, id int64, status commonpb.ReviewStatus) error {
	return admin.QAdmin.WithContext(ctx).
		Model(&admin.ReviewVideo{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_by": 0, // 0 means system
		}).Error
}
