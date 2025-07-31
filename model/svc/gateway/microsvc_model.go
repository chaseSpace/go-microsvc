package gateway

import (
	"microsvc/bizcomm/mq"
	"microsvc/consts"
	"microsvc/enums"
	"microsvc/model"
	"microsvc/model/modelsql"
	"microsvc/protocol/svc/adminpb"
	"microsvc/util"
	"microsvc/util/utime"
	"strings"
	"time"
)

type APIRateLimitConf struct {
	model.TableBase
	Svc         enums.Svc             `gorm:"column:svc" json:"svc"`           // user
	APIPath     string                `gorm:"column:api_path" json:"api_path"` // UserExt/GetUserInfo (还未支持Int方法)
	MaxQPSByIP  int64                 `gorm:"column:max_qps_by_ip" json:"max_qps_by_ip"`
	MaxQPSByUID int64                 `gorm:"column:max_qps_by_uid" json:"max_qps_by_uid"`
	State       APIRateLimitConfState `gorm:"column:state" json:"state"`
}

func (*APIRateLimitConf) TableName() string {
	return "api_rate_limit_conf"
}

type APIRateLimitConfState int8

const (
	APIRateLimitConfStateDisabled APIRateLimitConfState = 0
	APIRateLimitConfStateEnabled  APIRateLimitConfState = 1
)

type APICallLog struct {
	model.FieldID
	suffix    string
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	APICallLogInner
}

type APICallLogInner struct {
	model.FieldUID
	APIName     string `gorm:"column:api_name" json:"api_name"`
	APICtrl     string `gorm:"column:api_ctrl" json:"api_ctrl"`
	ReqIP       string `gorm:"column:req_ip" json:"req_ip"`
	DurMs       int64  `gorm:"column:dur_ms" json:"dur_ms"`
	Success     bool   `gorm:"column:success" json:"success"`
	Svc         string `gorm:"column:svc" json:"svc"`
	FromGateway bool   `gorm:"column:from_gateway" json:"from_gateway"`
	Panic       bool   `gorm:"column:panic" json:"panic"`
	ErrMsg      string `gorm:"column:err_msg" json:"err_msg"` // max-len=200
}

func (t *APICallLog) SetInner(v *mq.APICallLogBody) {
	t.UID = v.UID
	t.APIName = v.APIName
	t.APICtrl = v.APICtrl
	t.ReqIP = v.ReqIP
	t.DurMs = v.DurMs
	t.Success = v.Success
	t.Svc = v.Svc
	t.FromGateway = v.FromGateway
	t.Panic = v.Panic
	t.ErrMsg = util.TruncateUTF8(v.ErrMsg, 200)
}

func (t *APICallLog) TableName() string {
	return "api_call_log_" + t.suffix
}
func (t *APICallLog) SetSuffix(suffix string) {
	t.suffix = suffix
}
func (t *APICallLog) DLLSql() string {
	return strings.Replace(modelsql.APICallLogMonthTable, consts.YearMonth, t.suffix, 1)
}

func (t *APICallLog) ToPB() *adminpb.APICallLog {
	return &adminpb.APICallLog{
		Uid:       t.UID,
		ApiCtrl:   t.APICtrl,
		ApiName:   t.APIName,
		ReqIp:     t.ReqIP,
		Success:   t.Success,
		ErrMsg:    t.ErrMsg,
		Duration:  utime.DurationStr(time.Millisecond * time.Duration(t.DurMs)),
		CreatedAt: t.CreatedAt.Format(time.DateTime),
	}
}
