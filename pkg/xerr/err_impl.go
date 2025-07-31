package xerr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"microsvc/enums"
	"microsvc/pkg/i18n"
	"microsvc/pkg/i18n/langimpl"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util"
	"microsvc/util/utilcommon"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/cast"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// XErr 所有成员字段必须首字母大写，否则无法跨服务传递
type XErr struct {
	Code           int32
	Msg            string
	NMsg           string         `json:",omitempty"` // new msg with New()
	isTranslated   bool           // internal state
	API            string         `json:",omitempty"`
	SubId          langimpl.SubId `json:",omitempty"` // 用于多语言，JSON转换时保留，用于DEBUG
	PassedServices []string       `json:",omitempty"` // 服务之间透传时追加
	Chains         []XErr         `json:",omitempty"`

	callers *stack

	ctxValue interface{} // interface类型数据不宜序列化，亦不跨进程传输
}

func New(msg string, code ...int32) XErr {
	cd := ErrInternal.Code
	if len(code) > 0 {
		cd = code[0]
	}
	return XErr{Code: cd, Msg: msg}
}

// FromErr 将任何错误类型转换为XErr，并返回是否XErr类型
func FromErr(err error) (t XErr, ok bool) {
	if err == nil {
		return ErrNil, true
	}
	if ok = errors.As(err, &t); ok {
		return
	}
	// cross service transfer
	return FromErrStr(err.Error())
}

// FromErrStr 将任何错误字符串转换为XErr，并返回是否XErr类型
func FromErrStr(s string) (t XErr, ok bool) {
	err := json.Unmarshal([]byte(s), &t)
	if err != nil {
		return New(s), false
	}
	return t, t.Code > 0
}

// ToXErr 将任何错误类型转换为XErr，返回值永远非空
func ToXErr(err error) XErr {
	if t, ok := FromErr(err); ok {
		return t
	}
	return ErrUnknown.New(err.Error())
}

func (t XErr) joinChains() string {
	if len(t.Chains) == 0 {
		return ""
	}
	errChains := lo.Map(t.Chains, func(item XErr, index int) string {
		return "[" + item.FlatMsg() + "]"
	})
	return " - chains: " + strings.Join(errChains, " ➜ ")
}

func (t XErr) FormatMsg() string {
	var msg = t.Msg + t.NMsg
	if len(t.Chains) == 0 {
		return msg
	}
	errChains := lo.Map(t.Chains, func(item XErr, index int) string {
		return "[" + item.FormatMsg() + "]"
	})
	return msg + " - chains: " + strings.Join(errChains, " ➜ ")
}

// FlatMsgSimple 展平错误信息，含code，不含错误链和堆栈
func (t XErr) FlatMsgSimple() string {
	var msg = t.Msg
	if t.SubId != "" {
		msg += " ➜ " + string(t.SubId)
	}
	msg += t.NMsg
	msg = fmt.Sprintf("code:%d - msg:%s", t.Code, msg)
	return msg
}

// FlatMsg 将所有错误明细展平为单行字符串，含全部明细，不含堆栈
func (t XErr) FlatMsg() string {
	var msg = t.Msg + t.NMsg
	msg = fmt.Sprintf("code:%d - msg:%s", t.Code, msg)
	if len(t.PassedServices) > 1 {
		msg += " - rpcChains:" + strings.Join(lo.Reverse(t.PassedServices), "➜")
	}
	return msg + t.joinChains()
}

func (t XErr) Format(s fmt.State, verb rune) {
	for i, e := range append([]XErr{t}, t.Chains...) {
		switch verb {
		case 'v':
			if s.Flag('+') {
				io.WriteString(s, e.FlatMsg())
				if e.callers != nil {
					e.callers.Format(s, verb)
				}
			}
		case 's':
			fmt.Fprintf(s, "%s", e.FlatMsgSimple())
		}
		end := ""
		if i == len(t.Chains) {
			end = "END"
		}
		io.WriteString(s, fmt.Sprintf("\n\n------------- chain of err[%d] ↑↑↑ %s -------------\n\n", i, end))
	}
}

func (t XErr) Error() string {
	return util.ToJsonStr(&t)
}

// New 一旦使用它，说明你放弃了错误的国际化，尽量使用English
func (t XErr) New(msg string, args ...any) XErr {
	t.Msg = fmt.Sprintf(msg, args...)
	t.callers = callers()
	t.isTranslated = true // 不再国际化
	return t
}

func (t XErr) NewWithSubId(subId langimpl.SubId) XErr {
	t.SubId = subId
	return t
}

// Append 追加err msg
// 注意：caller 必须自行对err参数进行国际化后再传入
func (t XErr) Append(err error) XErr {
	t.callers = callers()
	t.Chains = append(t.Chains, ToXErr(err))
	return t
}

func (t XErr) AppendMsg(msg string, args ...any) XErr {
	if t.API != "" {
		t.NMsg += fmt.Sprintf("(%s)", t.API)
	}
	t.NMsg += " ➜ " + fmt.Sprintf(msg, args...)
	t.callers = callers()
	return t
}

func (t XErr) Equal(err error) bool {
	if e, ok := FromErr(err); ok {
		return e.Code == t.Code && e.SubId == t.SubId
	}
	return false
}

func (t XErr) DeepEqual(err error) bool {
	if e, ok := FromErr(err); ok {
		return e.Code == t.Code && e.SubId == t.SubId && e.Msg == t.Msg
	}
	return false
}

func (t XErr) Is(err error) bool {
	if err == nil {
		return false
	}
	if e, ok := FromErr(err); ok {
		for _, ce := range e.Chains {
			if t.Equal(ce) {
				return true
			}
		}
		return e.Code == t.Code && strings.HasPrefix(e.Msg, t.Msg)
	}
	return false
}

// TraceSvc 操作的字段是切片，只能是一个指针接收者
func (t XErr) TraceSvc(svc enums.Svc) XErr {
	t.PassedServices = append(t.PassedServices, svc.Name())
	return t
}

func (t XErr) IsNil() bool {
	return t.Code == ErrNil.Code
}

func (t XErr) IsInternal() bool {
	return t.Code >= 500 && t.Code < 600
}

func (t XErr) AutoAppend(err error, appendErrStr ...bool) error {
	if err == nil {
		return nil
	}
	if len(appendErrStr) > 0 && appendErrStr[0] {
		return t.AppendMsg(err.Error())
	}
	return t.Append(err)
}

func (t XErr) toI18nMsg(lang commonpb.Lang) string {
	var msg = t.Msg
	var tmsg = i18n.TranslateError(lang, langimpl.Code(t.Code), t.SubId)
	if tmsg != "" {
		msg = tmsg
	} else if t.SubId != "" {
		msg = strings.Join([]string{t.Msg, string(t.SubId)}, " ➜ ")
	}
	if t.NMsg != "" {
		msg += t.NMsg
	}
	return msg
}

// Translate 错误信息国际化（仅可调用一次，否则Msg会重复）
func (t XErr) Translate(lang commonpb.Lang) XErr {
	//fmt.Printf("1111, %+v %s\n", t, lang)
	if t.isTranslated {
		return t
	}
	t.isTranslated = true
	t.Msg = t.toI18nMsg(lang)
	t.SubId = ""
	t.NMsg = ""
	for i, e := range t.Chains {
		t.Chains[i] = e.Translate(lang)
	}
	return t
}

func (t XErr) WithAPI(api string) XErr {
	t.API = api
	return t
}

func (t XErr) WithCtxVal(val interface{}) XErr {
	t.ctxValue = val
	return t
}

func (t XErr) GetCtxVal() interface{} {
	return t.ctxValue
}

func (t XErr) GetStringVal() string {
	return cast.ToString(t.ctxValue)
}

func (t XErr) MatchCtxVal(val interface{}) bool {
	return t.ctxValue == val
}

// ---------------- 分割线 -------------------

// IsNil helper function to XErr.IsNil()
func IsNil(err error) bool {
	xErr := ToXErr(err)
	return xErr.IsNil()
}

func WrapRedis(err error) error {
	if err == nil || errors.Is(err, redis.Nil) {
		return nil
	}
	flag := utilcommon.CurrFuncName(2) // 获取再上层的函数名
	return ErrRedis.New(flag + ": " + err.Error())
}

func WrapMySQL(err error) error {
	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if strings.Contains(err.Error(), "Duplicate entry") {
		var err2 *mysql.MySQLError
		var conflictIdx string
		if errors.As(err, &err2) && err2.Number == 1062 {
			ss := strings.Split(err.Error(), "for key ")
			if len(ss) == 2 {
				conflictIdx = ss[1]
			}
		}
		return ErrDataDuplicate.WithCtxVal(conflictIdx)
	}
	//if errors.Is(err, gorm.ErrDuplicatedKey) {
	//	return ErrDataDuplicate
	//}
	flag := utilcommon.CurrFuncName(2) // 获取再上层的函数名
	return ErrMySQL.New(flag + ": " + err.Error())
}

func WrapDBUpdate(db *gorm.DB) error {
	if err := WrapMySQL(db.Error); err != nil {
		return err
	}
	if db.RowsAffected == 0 {
		return ErrNoRowAffectedOnUpdate
	}
	return nil
}

func WrapDBDelete(db *gorm.DB) error {
	if err := WrapMySQL(db.Error); err != nil {
		return err
	}
	if db.RowsAffected == 0 {
		return ErrDataNotExist
	}
	return nil
}

func IgnoreDupErr(err error) error {
	err = WrapMySQL(err)
	if ErrDataDuplicate.Equal(err) {
		return nil
	}
	return err
}

func JoinErrors(err ...error) error {
	var v XErr
	for i, e := range err {
		if e != nil {
			if i == 0 {
				v = ToXErr(e)
			} else {
				v.Chains = append(v.Chains, ToXErr(e))
			}
		}
	}
	if v.Code == 0 {
		return nil
	}
	return v
}
