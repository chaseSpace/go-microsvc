package utime

import "testing"

func TestParseDuration(t *testing.T) {
	t.Log(ParseDuration("1d"))
	t.Log(ParseDuration("1d1h"))
}
