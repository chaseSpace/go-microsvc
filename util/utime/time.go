package utime

import (
	"fmt"
	"microsvc/protocol/svc/commonpb"
	"time"

	"github.com/spf13/cast"
)

const (
	DateTimeMs = "2006-01-02 15:04:05.000"
	Date       = "2006-01-02"
	DateHour   = "2006-01-02 15"
)

func IsInTimeRange(t time.Time, start, end time.Time) bool {
	return t.Unix() >= start.Unix() && t.Unix() <= end.Unix()
}

func CheckTimeStr(layout string, str ...string) (list []time.Time, err error) {
	for _, s := range str {
		t, err := time.ParseInLocation(layout, s, time.Local)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return
}

func DateToday() int64 {
	date := time.Now().Format("20060102")
	return cast.ToInt64(date)
}
func DateYesterday() int64 {
	date := time.Now().AddDate(0, 0, -1).Format("20060102")
	return cast.ToInt64(date)
}

func DurationStr(t time.Duration) string {
	if int(t.Seconds()) > 0 {
		return fmt.Sprintf("%.1fs", t.Seconds())
	}
	if t.Milliseconds() > 0 {
		return fmt.Sprintf("%dms", t.Milliseconds())
	}
	if t.Microseconds() > 0 {
		return fmt.Sprintf("%dµs", t.Microseconds())
	}
	return t.String()
}

func GetDayZero(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// USANow 获取美国时间
func USANow() time.Time {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.Now().UTC().Add(-time.Hour * 4) // 美国通用时区 (UTC-4)
	}
	return time.Now().In(location) // 自动考虑夏令时
}

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func ParsePBTimeRange(in *commonpb.TimeRange) *TimeRange {
	st, _ := time.ParseInLocation(time.DateTime, in.StartDt, time.Local)
	et, _ := time.ParseInLocation(time.DateTime, in.EndDt, time.Local)
	return &TimeRange{
		Start: st,
		End:   et,
	}
}

func GetUnixFromTimePtr(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}

// GetUnixTwoTimePtr 获取两个时间戳（方便st，et）
func GetUnixTwoTimePtr(t, t2 *time.Time) (v1 int64, v2 int64) {
	if t != nil {
		v1 = t.Unix()
	}
	if t2 != nil {
		v2 = t2.Unix()
	}
	return
}

func GetTwoTimePtrFromUnix(t, t2 int64) (v1 *time.Time, v2 *time.Time) {
	// 注意，t只能是秒时间戳
	if t > 0 {
		v := time.Unix(t, 0)
		v1 = &v
	}
	if t2 > 0 {
		v := time.Unix(t2, 0)
		v2 = &v
	}
	return
}

// IsBetweenStringHourMinute 判断时间是否在指定的 HH:MM 时间段内
func IsBetweenStringHourMinute(start, end string, inputTime time.Time, matchAsYesterday ...bool) (hits bool, matchYesterday bool, err error) {
	st, e1 := time.Parse("15:04", start)
	et, e2 := time.Parse("15:04", end)
	if e1 != nil || e2 != nil {
		return false, false, fmt.Errorf("invalid format, start:[%s], end:[%s]", start, end)
	}

	stNum := st.Hour()*60 + st.Minute()
	etNum := et.Hour()*60 + et.Minute()

	if et.Hour() < st.Hour() {
		etNum = (et.Hour()+24)*60 + et.Minute()
	}
	//println(111, st.Hour(), et.Hour(), inputTime.Hour(), len(matchAsYesterday) > 0 && matchAsYesterday[0])
	inputNum := inputTime.Hour()*60 + inputTime.Minute()
	isBetween := func() bool { return inputNum >= stNum && inputNum <= etNum }

	// 仅作为昨天的时间段匹配
	if len(matchAsYesterday) > 0 && matchAsYesterday[0] {
		// match hrange as yesterday range
		inputNum = (inputTime.Hour()+24)*60 + inputTime.Minute()
		return isBetween(), true, nil
	}
	return isBetween(), false, nil
}
