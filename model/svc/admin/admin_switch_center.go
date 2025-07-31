package admin

import (
	"microsvc/model"
	"microsvc/protocol/svc/commonpb"

	"github.com/pkg/errors"
)

type AdminSwitchCenter struct {
	model.FieldID
	model.FieldAt
	model.FieldCreatedBy
	model.FieldUpdatedBy
	model.FieldDeletedTs
	Key      string               `gorm:"column:key" json:"key"`                              // 开关键，不含空格，唯一
	Name     string               `gorm:"column:name" json:"name"`                            // 开关中文名
	Value    commonpb.SwitchValue `gorm:"column:value" json:"value"`                          // 开关值
	ValueExt SwitchValExt         `gorm:"column:value_ext; serializer:json" json:"value_ext"` // 扩展值，json格式
	IsLock   bool                 `gorm:"column:is_lock" json:"is_lock"`                      // 是否锁定，true表示仅创建人可改！
}

// TableName specifies the table name for AdminSwitchCenter
func (AdminSwitchCenter) TableName() string {
	return "admin_switch_center"
}

func (t AdminSwitchCenter) ToPB() *commonpb.SwitchItem {
	return &commonpb.SwitchItem{
		Core: &commonpb.SwitchItemCore{
			Key:      t.Key,
			Name:     t.Name,
			Value:    t.Value,
			ValueExt: t.ValueExt,
			IsLock:   t.IsLock,
		},
		CreatedBy: t.CreatedBy,
		UpdatedBy: t.UpdatedBy,
	}
}

// SwitchValExt 扩展 SValue，key: int - value: description
type SwitchValExt map[int32]string

func (s SwitchValExt) Check(svalue commonpb.SwitchValue) error {
	if commonpb.SwitchValue_name[int32(svalue)] == "" && (s == nil || s[int32(svalue)] == "") {
		return errors.New("无效的开关值")
	}
	return nil
}
