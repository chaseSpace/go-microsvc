package utime

import (
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"regexp"
	"strings"
	"time"
)

// FormatHourMinuteStrToUSAStyle
// 将 09:00 格式化为 9am，或者 19:00 格式化为 7pm
// 若分钟数非0，则不转换
func FormatHourMinuteStrToUSAStyle(t string) string {
	ss := regexp.MustCompile(`^(\d\d):00$`).FindStringSubmatch(t)
	if len(ss) == 2 {
		h := cast.ToInt64(ss[1])
		if h > 12 {
			h = h - 12
			return cast.ToString(h) + "pm"
		} else {
			return cast.ToString(h) + "am"
		}
	}
	return t
}

// FormatHourMinuteRangeStrToUSAStyle 将 09:00-13:00 格式化为 9am-1pm
func FormatHourMinuteRangeStrToUSAStyle(hourRange string) string {
	ss := strings.Split(hourRange, "-")
	if len(ss) != 2 {
		return hourRange
	}
	return strings.Join([]string{FormatHourMinuteStrToUSAStyle(ss[0]), FormatHourMinuteStrToUSAStyle(ss[1])}, "-")
}

// FormatTimeRangeToEngHumanReadable 格式化时间区间为英语格式、人类可读
// Format(cross days <=1): <hour range> <day of week>, <month> <day>, <year>
// Format(cross days >1):
// f1: 09:00-13:00 July 8, Friday[, 2022]
// f1.1: 09:00-late July 8, Friday[, 2022] (if end-time is early morning)
// f2: 09:00 July 8, Friday[, 2022] | 13:00 July 8, Monday[, 2022] (| is given outer to split)
func FormatTimeRangeToEngHumanReadable(st, et, now time.Time) (string, error) {
	// 时间范围验证
	if st.After(et) {
		return "", errors.New("start time is after end time")
	}
	// 格式化时间
	startHour := FormatHourMinuteStrToUSAStyle(st.Format("15:04")) // 1 AM
	endStr := ""
	if et.Hour() <= 5 { // early morning
		endStr = "LATE"
	} else {
		endStr = FormatHourMinuteStrToUSAStyle(et.Format("15:04"))
	}
	dayOfWeek := st.Format("Monday")
	monthDay := st.Format("January 2")

	if et.Sub(st).Hours() <= 24 && st.Year() == et.Year() {
		if st.Year() == now.Year() {
			return fmt.Sprintf("%s-%s %s, %s", startHour, endStr, monthDay, dayOfWeek), nil
		}
		return fmt.Sprintf("%s-%s %s, %s, %d", startHour, endStr, monthDay, dayOfWeek, st.Year()), nil
	}
	stYear := st.Year()
	endHour := FormatHourMinuteStrToUSAStyle(et.Format("15:04"))
	endWeekday := et.Format("Monday")
	endMonthDay := et.Format("January 2")
	etYear := et.Year()

	var left, right string
	if stYear == etYear {
		left = fmt.Sprintf("%s %s, %s", startHour, monthDay, dayOfWeek)
		right = fmt.Sprintf("%s %s, %s", endHour, endMonthDay, endWeekday)
	} else {
		left = fmt.Sprintf("%s %s, %s, %d", startHour, monthDay, dayOfWeek, stYear)
		right = fmt.Sprintf("%s %s, %s, %d", endHour, endMonthDay, endWeekday, etYear)
	}
	return left + " | " + right, nil
}
