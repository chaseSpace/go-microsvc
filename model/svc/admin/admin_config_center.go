package admin

import (
	"microsvc/model"
	"microsvc/protocol/svc/commonpb"
)

type AdminConfigCenter struct {
	model.FieldID
	model.FieldAt
	model.FieldCreatedBy
	model.FieldUpdatedBy
	model.FieldDeletedTs
	Key                string `gorm:"column:key" json:"key"`
	Name               string `gorm:"column:name" json:"name"`
	Value              string `gorm:"column:value" json:"value"`
	IsLock             bool   `gorm:"column:is_lock" json:"is_lock"`
	AllowProgramUpdate bool   `gorm:"column:allow_program_update" json:"allow_program_update"`
}

func (AdminConfigCenter) TableName() string {
	return "admin_config_center"
}
func (t AdminConfigCenter) ToPB() *commonpb.ConfigItem {
	return &commonpb.ConfigItem{
		Core: &commonpb.ConfigItemCore{
			Key:                t.Key,
			Name:               t.Name,
			Value:              t.Value,
			IsLock:             t.IsLock,
			AllowProgramUpdate: t.AllowProgramUpdate,
		},
		CreatedBy: t.CreatedBy,
		UpdatedBy: t.UpdatedBy,
	}
}
