package utime

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFormatTimeRangeToEngHumanReadable(t *testing.T) {
	// same days, this year
	st := time.Date(2024, 1, 1, 19, 0, 0, 0, time.UTC)
	et := time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC)
	s, _ := FormatTimeRangeToEngHumanReadable(st, et, time.Now())
	t.Log("f1", s)

	// same days, diff year
	st = time.Date(2023, 1, 1, 19, 0, 0, 0, time.UTC)
	et = time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC)
	s, _ = FormatTimeRangeToEngHumanReadable(st, et, time.Now())
	t.Log("f1.1", s)

	// same days, not this year
	st = time.Date(2023, 1, 1, 19, 0, 0, 0, time.UTC)
	et = time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC)
	s, _ = FormatTimeRangeToEngHumanReadable(st, et, time.Now())
	t.Log("f2", s)

	// different days, this year
	st = time.Date(2024, 1, 1, 19, 0, 0, 0, time.UTC)
	et = time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)
	s, _ = FormatTimeRangeToEngHumanReadable(st, et, time.Now())
	t.Log("f3", s)
	// different days, diff year
	st = time.Date(2023, 1, 1, 19, 0, 0, 0, time.UTC)
	et = time.Date(2024, 1, 2, 3, 0, 0, 0, time.UTC)
	s, _ = FormatTimeRangeToEngHumanReadable(st, et, time.Now())
	t.Log("f3.1", s)

	// different days, not this year
	st = time.Date(2023, 1, 1, 19, 0, 0, 0, time.UTC)
	et = time.Date(2023, 1, 2, 3, 0, 0, 0, time.UTC)
	s, _ = FormatTimeRangeToEngHumanReadable(st, et, time.Now())
	t.Log("f4", s)
}

func TestFormatHourMinuteStrToUSAStyle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"13:00", "1pm"},
		{"12:00", "12am"},
		{"09:30", "09:30"},
		{"23:59", "23:59"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatHourMinuteStrToUSAStyle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
