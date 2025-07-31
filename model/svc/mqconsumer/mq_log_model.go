package mqconsumer

import (
	"fmt"
	"microsvc/consts"
	"microsvc/model"
	"microsvc/model/modelsql"
	"strings"
	"time"
)

type MqLog struct {
	model.FieldID
	suffix        string
	TopicUniqueId string    `gorm:"column:topic_unique_id" json:"topic_unique_id"`
	Topic         string    `gorm:"column:topic" json:"topic"`
	Data          string    `gorm:"column:data" json:"data"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (t *MqLog) TableName() string {
	return "mq_log_" + t.suffix
}

func (t *MqLog) SetSuffix(suffix string) {
	t.suffix = suffix
}
func (t *MqLog) DDLSql() string {
	return fmt.Sprintf(strings.Replace(modelsql.MqLogMonthTable, consts.YearMonth, t.suffix, 1))
}
