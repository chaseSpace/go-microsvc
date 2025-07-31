package user

import (
	"microsvc/model"
	"microsvc/protocol/svc/commonpb"
	"time"
)

type SignInLog struct {
	model.FieldID
	model.FieldUID
	Platform  commonpb.SignInPlatform `gorm:"column:platform" json:"platform"`
	System    commonpb.SignInSystem   `gorm:"column:system" json:"system"`
	Type      commonpb.SignInType     `gorm:"column:type" json:"type"`
	SignInAt  time.Time               `gorm:"column:sign_in_at" json:"sign_in_at"`
	IP        string                  `gorm:"column:ip" json:"ip"`
	CreatedAt time.Time               `gorm:"column:created_at" json:"created_at"`
}

func (*SignInLog) TableName() string {
	return "sign_in_log"
}
func (s *SignInLog) ToPB() *commonpb.UserSignInLog {
	return &commonpb.UserSignInLog{
		SignInAt: s.SignInAt.Unix(),
		Ip:       s.IP,
		Type:     s.Type,
		Platform: s.Platform,
		System:   s.System,
	}
}
