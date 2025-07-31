package user

import "microsvc/protocol/svc/userpb"

type PunishRPC struct {
	Punish
	Nickname string `gorm:"column:nickname" json:"nickname"`
}

func (*PunishRPC) TableName() string {
	return new(Punish).TableName()
}

func (p *PunishRPC) ToPB() *userpb.Punish {
	return &userpb.Punish{
		Id:        p.Id,
		Uid:       p.UID,
		Type:      p.Type,
		Duration:  p.Duration,
		Reason:    p.Reason,
		State:     p.State,
		CreatedAt: p.CreatedAt.Unix(),
		UpdatedAt: p.UpdatedAt.Unix(),
		CreatedBy: p.CreatedBy,
		UpdatedBy: p.UpdatedBy,
		Nickname:  p.Nickname,
	}
}

type PunishLogRPC struct {
	PunishLog
	Nickname string `gorm:"column:nickname" json:"nickname"`
}

func (*PunishLogRPC) TableName() string {
	return new(PunishLog).TableName()
}
func (p *PunishLogRPC) ToPB() *userpb.PunishLog {
	return &userpb.PunishLog{
		Id:        p.Id,
		Uid:       p.UID,
		Type:      p.PunishType,
		OpType:    p.OpType,
		Duration:  p.Duration,
		Reason:    p.Reason,
		CreatedAt: p.CreatedAt.Unix(),
		CreatedBy: p.CreatedBy,
		Nickname:  p.Nickname,
	}
}
