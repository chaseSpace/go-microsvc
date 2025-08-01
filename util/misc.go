package util

import (
	"fmt"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util/urand"
	"net"
	"net/url"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/k0kubun/pp/v3"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

var WaringPP = pp.New()

func init() {
	WaringPP.SetColorScheme(colorScheme)
	WaringPP.SetExportedOnly(true)
}

type TcpListenerFetcher struct {
	portMin, portMax int
	mem              map[int]int
}

func NewTcpListenerFetcher(portMin, portMax int) *TcpListenerFetcher {
	return &TcpListenerFetcher{portMin: portMin, portMax: portMax, mem: make(map[int]int)}
}

func (t *TcpListenerFetcher) Get() (lis net.Listener, port int, err error) {
	if t.portMin >= t.portMax {
		return nil, 0, errors.New("portMin must less than portMax")
	}
	loops := t.portMax - t.portMin + 1
	for i := 0; i < loops; i++ {
		port = urand.RandIntRange(t.portMin, t.portMax, t.mem)
		lis, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			if strings.Contains(err.Error(), "already") {
				//println("continue", port)
				continue
			}
			return
		}
		//println(111, port)
		return
	}
	return nil, 0, fmt.Errorf("failed, tried %d times", loops)
}

var (
	// unique-id, copy-friendly, sortable by gen time
	// see https://github.com/segmentio/ksuid
	__ksuid      = ksuid.New()
	__ksuidMutex = sync.Mutex{}
)

// NewKsuid 生成时钟相关的、递增性的唯一id（大小写敏感）
// - 注意：仅保证单进程唯一
func NewKsuid() string {
	__ksuidMutex.Lock()
	defer __ksuidMutex.Unlock()
	__ksuid = __ksuid.Next()
	return __ksuid.String()
}

func GetOptArg[T any](a []T, def T) T {
	if len(a) == 0 {
		return def
	}
	return a[0]
}

type FuzzyCharTyp int8

const (
	FuzzyCharTypNone FuzzyCharTyp = iota
	FuzzyCharTypPhone
	FuzzyCharTypCitizenId // 身份证
)

// FuzzyChars 对字符串进行模糊处理
// example：
//
//		-- 123456 => 12**56
//		-- 15983882334 => 159****2334
//	    -- 440308198612183456 => 440308********3456
func FuzzyChars(src string, typ ...FuzzyCharTyp) string {
	tp := GetOptArg[FuzzyCharTyp](typ, FuzzyCharTypNone)
	_lenDiv3 := len(src) / 3
	if _lenDiv3 == 0 {
		return ""
	}
	tmp := []rune(src)
	start := 0
	end := 0
	switch tp {
	case FuzzyCharTypPhone:
		ss := strings.Split(src, "-")
		if len(ss) == 2 {
			if len(ss[1])/3 > 0 {
				_lenDiv3 = len(ss[1]) / 3
				start = len(ss[0]) + 1 + _lenDiv3
				if _lenDiv3*3 != len(src) {
					end = len(src) - _lenDiv3
				} else {
					end = len(ss[0]) + 1 + _lenDiv3*2
				}
			} else {
				tp = FuzzyCharTypNone
			}

			goto OUTOF_SWITCH
		}

		if len(src) == 11 {
			start = 3
			end = 7
		}
	case FuzzyCharTypCitizenId:
		if len(src) == 18 {
			start = 6
			end = 14
		}
	default:
		if tp != FuzzyCharTypNone {
			panic(fmt.Sprintf("unknown FuzzyCharTyp:%v", tp))
		}
	}

OUTOF_SWITCH:
	if tp == FuzzyCharTypNone {
		start = _lenDiv3
		if _lenDiv3*3 != len(src) {
			end = len(src) - _lenDiv3
		} else {
			end = _lenDiv3 * 2
		}
	}

	for i := start; i < end; i++ {
		tmp[i] = '*'
	}
	return string(tmp)
}

func SortInt64Asc(ids ...int64) (slice []int64) {
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	return ids
}

func AnotherUID(uid int64, array [2]int64) int64 {
	if uid == array[0] {
		return array[1]
	}
	return array[0]
}

func ReadEnvBool(key string) bool {
	val := os.Getenv(key)
	return val == "true" || val == "1"
}

func FormatStack(skip int) (errStack string) {
	pc := make([]uintptr, 6)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		errStack += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
	}
	return
}

func TruncateUTF8(s string, length int, appendTail ...string) string {
	if length == 0 {
		return ""
	}

	runes := []rune(s) // 将字符串转换为rune切片
	if len(runes) > length {
		if len(appendTail) > 0 {
			return string(runes[:length]) + appendTail[0]
		}
		return string(runes[:length]) // 截断到指定长度
	}

	return s
}

func TruncateBytes(s []byte, length int) []byte {
	if length <= 0 {
		return nil
	}
	if len(s) > length {
		return s[:length] // 截断到指定长度
	}
	return s
}

func GetElemFromSlice[T *any](slice []T, idx int) T {
	if len(slice) > idx {
		return slice[idx]
	}
	return nil
}

var colorScheme = pp.ColorScheme{
	Integer: pp.Green | pp.Bold,
	Float:   pp.Black | pp.BackgroundWhite | pp.Bold,
	String:  pp.Yellow,
}

var currencyTypeMap = map[commonpb.CurrencyType]string{
	commonpb.CurrencyType_CT_CNY: "¥",
	commonpb.CurrencyType_CT_USD: "$",
}

func GetPriceDesc(price float64, currencyType commonpb.CurrencyType) string {
	unit := currencyTypeMap[currencyType]
	if unit == "" {
		return fmt.Sprintf("?%.2f", price)
	}
	return fmt.Sprintf("%s%.2f", unit, price)
}

func GenDigitOrderNo(uid int64) string {
	ss := strings.Split(time.Now().Format(`060102.150405.000`), ".")
	return fmt.Sprintf(`%s-%d%s%s`, ss[0], uid*2, ss[1], ss[2])
}

func MustJoinUrl(base string, path ...string) string {
	if ret, err := url.JoinPath(base, path...); err != nil {
		panic(err)
	} else {
		return ret
	}
}
