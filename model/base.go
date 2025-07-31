package model

import (
	"time"
)

const (
	MysqlDB         = "biz_core"
	MysqlDBLog      = "biz_core_log"
	MysqlDBAdmin    = "biz_admin"
	MysqlDBMicroSvc = "micro_svc"
	MysqlDBGateway  = "micro_gateway"

	RedisDB         = "biz_core"
	RedisDBAdmin    = "admin"
	RedisDBMicroSvc = "micro_svc"
	RedisDBGateway  = "micro_gateway"
)

type TableBase struct {
	FieldID
	FieldAt
}

type FieldID struct {
	Id int64 `gorm:"column:id" json:"id"`
}

func NewID(id int64) FieldID {
	return FieldID{Id: id}
}

type FieldUID struct {
	UID int64 `gorm:"column:uid" json:"uid"`
}

func NewUID(id int64) FieldUID {
	return FieldUID{UID: id}
}

type FieldAt struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func NewAt(ct, ut time.Time) FieldAt {
	return FieldAt{
		CreatedAt: ct,
		UpdatedAt: ut,
	}
}

type FieldCt struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func NewCt(ct time.Time) FieldCt {
	return FieldCt{CreatedAt: ct}
}

type FieldCreatedBy struct {
	CreatedBy int64 `gorm:"column:created_by" json:"created_by"`
}

func NewCb(id int64) FieldCreatedBy {
	return FieldCreatedBy{CreatedBy: id}
}

type FieldUpdatedBy struct {
	UpdatedBy int64 `gorm:"column:updated_by" json:"updated_by"`
}

func NewUb(id int64) FieldUpdatedBy {
	return FieldUpdatedBy{UpdatedBy: id}
}

type FieldDeletedAt struct {
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func NewDt(at time.Time) FieldDeletedAt {
	return FieldDeletedAt{DeletedAt: &at}
}

// FieldDeletedTs 使用 `not null`类型的int时间戳代替 `nullable`的datetime 类型
// - 以便设置复合唯一键
type FieldDeletedTs struct {
	DeletedTs int64 `gorm:"column:deleted_ts" json:"deleted_ts"`
}

func NewDtTs(ts int64) FieldDeletedTs {
	return FieldDeletedTs{DeletedTs: ts}
}

func (t *FieldUID) GetUID() int64 {
	return t.UID
}
func (t *FieldUpdatedBy) GetUpdateByAdminUID() int64 {
	return t.UpdatedBy
}

func NewTableBaseFieldID(id int64) TableBase {
	return TableBase{FieldID: FieldID{Id: id}}
}
