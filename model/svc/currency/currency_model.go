package currency

import (
	"fmt"
	"microsvc/consts"
	"microsvc/model"
	"microsvc/model/modelsql"
	"microsvc/protocol/svc/currencypb"
	"strings"
	"time"
)

type GoldAccount struct {
	model.TableBase
	UID           int64 `gorm:"column:uid" json:"uid"`
	Balance       int64 `gorm:"column:balance" json:"balance"`
	RechargeTotal int64 `gorm:"column:recharge_total" json:"recharge_total"`
}

func (*GoldAccount) TableName() string {
	return "gold_account"
}

type GoldTxLog struct {
	model.FieldID
	suffix    string
	TxId      string                `gorm:"column:tx_id" json:"tx_id"`
	UID       int64                 `gorm:"column:uid" json:"uid"`
	Delta     int64                 `gorm:"column:delta" json:"delta"`
	Balance   int64                 `gorm:"column:balance" json:"balance"` // 变更后
	Remark    string                `gorm:"column:remark" json:"remark"`
	TxType    currencypb.GoldTxType `gorm:"column:tx_type" json:"tx_type"`
	CreatedAt time.Time             `gorm:"column:created_at" json:"created_at"`
}

// TableName 注意：这个方法读取了内部变量
// gorm 不能通过 Model() 读取到 TableName（成员变量永远为空），只能通过 gorm.DB.Table(v.TableName()) 设置
func (t *GoldTxLog) TableName() string {
	return "gold_tx_log_" + t.suffix
}

func (t *GoldTxLog) SetSuffix(suffix string) {
	t.suffix = suffix
}

func (t *GoldTxLog) ToPB() *currencypb.GoldTxLog {
	return &currencypb.GoldTxLog{
		Uid:       t.UID,
		Delta:     t.Delta,
		TxType:    t.TxType,
		CreatedAt: t.CreatedAt.Unix(),
	}
}

func (t *GoldTxLog) DDLSql() string {
	return fmt.Sprintf(strings.Replace(modelsql.GoldSingleTxLogMonthTable, consts.YearMonth, t.suffix, 1))
}
