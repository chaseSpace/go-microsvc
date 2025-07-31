package thirdparty

import (
	"microsvc/enums"
	"microsvc/model"
)

type ReviewText struct {
	model.TableBase
	Id        int32                  `gorm:"column:id" json:"id"`
	UID       int                    `gorm:"column:uid" json:"uid"`
	Text      string                 `gorm:"column:text" json:"text"`
	TextType  enums.ReviewTextTyp    `gorm:"column:text_type" json:"text_type"`
	Status    enums.ReviewTextStatus `gorm:"column:status" json:"status"`
	AdminUID  *int                   `gorm:"column:admin_uid" json:"admin_uid"`
	AdminNote *string                `gorm:"column:admin_note" json:"admin_note"`
}

func (r *ReviewText) TableName() string {
	return "review_text"
}

type ReviewImage struct {
	model.TableBase
	Id        int32                 `gorm:"column:id" json:"id"`
	UID       int                   `gorm:"column:uid" json:"uid"`
	ImgUrl    string                `gorm:"column:img_url" json:"img_url"`
	ImgType   enums.ReviewImgTyp    `gorm:"column:img_type" json:"img_type"`
	Status    enums.ReviewImgStatus `gorm:"column:status" json:"status"`
	AdminUID  *int                  `gorm:"column:admin_uid" json:"admin_uid"`
	AdminNote *string               `gorm:"column:admin_note" json:"admin_note"`
}

func (r *ReviewImage) TableName() string {
	return "review_image"
}
