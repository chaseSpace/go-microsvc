package utime

import (
	"testing"
	"time"
)

func TestUSANow(t *testing.T) {
	t.Log(USANow())
}

func TestIsNowBetweenStringHourMinute(t *testing.T) {
	// 设置测试用例
	testCases := []struct {
		start              string
		end                string
		alsoMatchYesterday bool
		expected           bool
		matchYesterday     bool
		err                bool
	}{
		// 测试用例数据，格式为开始时间、结束时间和预期结果
		{"09:00", "17:00", false, false, false, false},
		{"18:00", "02:00", false, false, false, false},
		{"02:00", "09:00", false, false, false, false},
		{"00:30", "05:00", false, true, false, false},
		{"00:00", "02:00", false, true, false, false},

		// alsoMatchYesterday-false
		{"19:00", "00:00", true, false, false, false},
		{"19:00", "00:29", true, false, false, false},

		// alsoMatchYesterday-true
		{"19:00", "02:00", true, true, true, false},
		{"23:00", "00:30", true, true, true, false},

		{"09:00", "17:00:00", false, false, false, true},
		{"09:00", "17", false, false, false, true},
	}

	input := time.Date(2024, 1, 1, 0, 30, 0, 0, time.UTC)
	for _, tc := range testCases {
		result, matchYesterday, err := IsBetweenStringHourMinute(tc.start, tc.end, input, tc.alsoMatchYesterday)
		if (err != nil) != tc.err {
			t.Errorf("(%s, %s) returned unexpected error state: %v", tc.start, tc.end, err)
		} else if result != tc.expected {
			t.Errorf("(%s, %s) = %v, expected %v", tc.start, tc.end, result, tc.expected)
		} else if matchYesterday != tc.matchYesterday {
			t.Errorf("(%s, %s) returned unexpected matchYesterday state: %v", tc.start, tc.end, matchYesterday)
		}
	}
}

func TestIsNowBetweenStringHourMinute111(t *testing.T) {
	result, matchYesterday, err := IsBetweenStringHourMinute("19:00", "11:00", time.Now(), true)
	if err != nil {
		t.Errorf("error: %v", err)
	} else {
		t.Logf("result: %v %v", result, matchYesterday)
	}
}
