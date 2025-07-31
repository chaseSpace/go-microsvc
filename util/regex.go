package util

import (
	"github.com/spf13/cast"
	"regexp"
)

var (
	digestRegexPattern = regexp.MustCompile(`^\d+$`)
)

func IsDigestStr(str string) (int64, bool) {
	if digestRegexPattern.Match([]byte(str)) {
		return cast.ToInt64(str), true
	}
	return 0, false
}
