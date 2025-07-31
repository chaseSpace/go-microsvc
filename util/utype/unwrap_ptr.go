package utype

import "time"

func StringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func RuneVal(s *rune) rune {
	if s == nil {
		return 0
	}
	return *s
}

func IntVal(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func Float32Val(f *float32) float32 {
	if f == nil {
		return 0.0
	}
	return *f
}

func Float64Val(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}

func BoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func TimeVal(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
