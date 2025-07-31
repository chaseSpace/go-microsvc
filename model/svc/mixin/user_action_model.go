package mixin

import (
	"microsvc/model"
	"microsvc/protocol/svc/commonpb"
)

type UserLikes struct {
	model.FieldID
	model.FieldCt
	model.FieldUID
	TargetId   int64                   `gorm:"column:target_id" json:"target_id"`
	TargetType commonpb.LikeTargetType `gorm:"column:target_type" json:"target_type"`
}

func (UserLikes) TableName() string {
	return "user_likes"
}
