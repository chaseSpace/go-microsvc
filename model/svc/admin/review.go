package admin

import (
	"microsvc/model"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
)

type ReviewText struct {
	model.TableBase
	model.FieldUID
	model.FieldUpdatedBy
	Text    string                `gorm:"column:text" json:"text"`
	BizType commonpb.BizType      `gorm:"column:biz_type" json:"biz_type"`
	Status  commonpb.ReviewStatus `gorm:"column:status" json:"status"`
}

func (*ReviewText) TableName() string {
	return "review_text"
}

func (t *ReviewText) ToPB() *adminpb.ReviewText {
	return &adminpb.ReviewText{
		Id:       t.Id,
		Text:     t.Text,
		Status:   t.Status,
		BizType:  t.BizType,
		AdminUid: t.UpdatedBy,
	}
}

type ReviewImage struct {
	model.TableBase
	model.FieldUID
	model.FieldUpdatedBy
	Text    string                `gorm:"column:text" json:"text"`
	Urls    []string              `gorm:"column:urls;serializer:json" json:"urls"`
	BizType commonpb.BizType      `gorm:"column:biz_type" json:"biz_type"`
	Status  commonpb.ReviewStatus `gorm:"column:status" json:"status"`
}

func (*ReviewImage) TableName() string {
	return "review_image"
}
func (t *ReviewImage) ToPB() *adminpb.ReviewImage {
	return &adminpb.ReviewImage{
		Id:       t.Id,
		Urls:     t.Urls,
		Status:   t.Status,
		BizType:  t.BizType,
		AdminUid: t.UpdatedBy,
	}
}

type ReviewVideo struct {
	model.TableBase
	model.FieldUID
	model.FieldUpdatedBy
	Text          string                `gorm:"column:text" json:"text"`
	Url           string                `gorm:"column:url" json:"url"`
	BizType       commonpb.BizType      `gorm:"column:biz_type" json:"biz_type"`
	Status        commonpb.ReviewStatus `gorm:"column:status" json:"status"`
	BizUniqId     int64                 `gorm:"column:biz_uniq_id" json:"biz_uniq_id"`
	ThTaskId      string                `gorm:"column:th_task_id" json:"th_task_id"`
	ThName        string                `gorm:"column:th_name" json:"th_name"`
	QueryRetFails int                   `gorm:"column:query_ret_fails" json:"query_ret_fails"`
}

func (*ReviewVideo) TableName() string {
	return "review_video"
}
func (t *ReviewVideo) ToPB() *adminpb.ReviewVideo {
	return &adminpb.ReviewVideo{
		Id:       t.Id,
		Url:      t.Url,
		Status:   t.Status,
		BizType:  t.BizType,
		AdminUid: t.UpdatedBy,
	}
}

type ReviewAudio struct {
	model.TableBase
	model.FieldUID
	model.FieldUpdatedBy
	Text          string                `gorm:"column:text" json:"text"`
	Url           string                `gorm:"column:url" json:"url"`
	BizType       commonpb.BizType      `gorm:"column:biz_type" json:"biz_type"`
	Status        commonpb.ReviewStatus `gorm:"column:status" json:"status"`
	BizUniqId     int64                 `gorm:"column:biz_uniq_id" json:"biz_uniq_id"`
	ThTaskId      string                `gorm:"column:th_task_id" json:"th_task_id"`
	ThName        string                `gorm:"column:th_name" json:"th_name"`
	QueryRetFails int                   `gorm:"column:query_ret_fails" json:"query_ret_fails"`
}

func (*ReviewAudio) TableName() string {
	return "review_audio"
}
func (t *ReviewAudio) ToPB() *adminpb.ReviewAudio {
	return &adminpb.ReviewAudio{
		Id:       t.Id,
		Url:      t.Url,
		Status:   t.Status,
		BizType:  t.BizType,
		AdminUid: t.UpdatedBy,
	}
}
