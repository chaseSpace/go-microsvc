package friend

import (
	"microsvc/model"
)

const (
	TableFriendUKMain = "'uk_main'"
	TableBlockUKMain  = "'uk_main'"
)

// Friend 好友表（含单边关系）
type Friend struct {
	model.TableBase
	UID      int64 `gorm:"column:uid" json:"uid"`
	FID      int64 `gorm:"column:fid" json:"fid"`
	Intimacy int64 `gorm:"column:intimacy" json:"intimacy"` // 亲密度，单边维护
}

func (Friend) TableName() string {
	return "friend"
}

// Block 拉黑表
type Block struct {
	model.TableBase
	UID int64 `gorm:"column:uid" json:"uid"`
	BID int64 `gorm:"column:bid" json:"bid"` // 被拉黑者
}

func (Block) TableName() string {
	return "block"
}

// Visitor 访客表
type Visitor struct {
	model.TableBase
	UID           int64 `gorm:"column:uid" json:"uid"`
	VID           int64 `gorm:"column:vid" json:"vid"`
	DayVisitTimes int64 `gorm:"column:day_visit_times" json:"day_visit_times"`
	Date          int64 `gorm:"column:date" json:"date"` // yyyyMMdd
}

func (Visitor) TableName() string {
	return "visitor"
}
