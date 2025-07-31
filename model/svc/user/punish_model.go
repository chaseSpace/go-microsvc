package user

import (
	"microsvc/model"
	"microsvc/protocol/svc/commonpb"
)

type Punish struct {
	model.TableBase
	model.FieldUID
	model.FieldCreatedBy
	model.FieldUpdatedBy
	Type     commonpb.PunishType  `gorm:"column:type" json:"type"`
	Reason   string               `gorm:"column:reason" json:"reason"` // 更新/解除时会追加原因，多个原因以;分隔
	Duration int64                `gorm:"column:duration" json:"duration"`
	State    commonpb.PunishState `gorm:"column:state" json:"state"`
}

func (*Punish) TableName() string {
	return "punish"
}

func (v *Punish) TableAs(alias string) string {
	return v.TableName() + " AS " + alias
}

type PunishLog struct {
	model.FieldID
	model.FieldUID
	model.FieldCt
	model.FieldCreatedBy
	OpType     commonpb.PunishOpType `gorm:"column:op_type" json:"op_type"`
	PunishType commonpb.PunishType   `gorm:"column:punish_type" json:"punish_type"`
	Reason     string                `gorm:"column:reason" json:"reason"` // 不会更新
	Duration   int64                 `gorm:"column:duration" json:"duration"`
}

func (*PunishLog) TableName() string {
	return "punish_log"
}

func (v *PunishLog) TableAs(alias string) string {
	return v.TableName() + " AS " + alias
}
