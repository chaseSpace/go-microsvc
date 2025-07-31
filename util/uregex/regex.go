package uregex

import (
	"encoding/json"
	"regexp"

	"github.com/pkg/errors"
)

var jsonRegex = regexp.MustCompile(`\{.*}`)

func ExtractJson(s string, v interface{}) error {
	res := jsonRegex.FindString(s)
	if res == "" {
		return errors.New("json not found")
	}
	return json.Unmarshal([]byte(res), v)
}

var emailRegex = regexp.MustCompile(`^\w+(\w+-\w+)*@\w+(-\w+)*(\.\w+)+$`)

func IsInvalidEmail(email string) bool {
	return !emailRegex.MatchString(email)
}

// 更复杂：必须要包含大小写、数字、特殊字符
//var passwdRegex2 = regexp.MustCompile(`(?=.*[0-9])(?=.*[A-Z])(?=.*[a-z])(?=.*[^a-zA-Z0-9]).{8,15}`)

func AssertLength(src string, min, max int) bool {
	return len(src) >= min && len(src) <= max
}

var (
	AlphaUpperRegex  = regexp.MustCompile(`[A-Z]+`)
	AlphaLowerRegex  = regexp.MustCompile(`[a-z]+`)
	DigitRegex       = regexp.MustCompile(`[0-9]+`)
	SpecialCharRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)
)

// IsLegalPasswd
// lv1: 必须包含大小写、数字、长度8-16
// lv2: 必须包含大小写、数字、特殊字符、长度8-16
// -- (没有找到一个能通过用例的正则)
func IsLegalPasswd(in string, lv int8) (r bool) {
	switch lv {
	case 1:
		r = AlphaUpperRegex.MatchString(in) && AlphaLowerRegex.MatchString(in) && DigitRegex.MatchString(in) && AssertLength(in, 8, 16)
	case 2:
		r = AlphaUpperRegex.MatchString(in) && AlphaLowerRegex.MatchString(in) && DigitRegex.MatchString(in) && SpecialCharRegex.MatchString(in) && AssertLength(in, 8, 16)
	default:
		panic("unknown level for IsLegalPasswd")
	}
	return
}

var (
	NonDigitAlphaCharRegex  = regexp.MustCompile(`[^\p{L}\p{N}]+`)                                  // 非unicode数字字母属性的特殊字符
	UnusualSpecialCharRegex = regexp.MustCompile(`[^\p{L}\p{N}a-zA-Z0-9()./\\+\-_=~·!！;:"|' ,<>?]`) // 非数字字母和常用符号
	OnlySpecialCharRegex    = regexp.MustCompile(`^[^\p{L}\p{N}]+$`)                                // 匹配全都是【非unicode数字字母属性的特殊字符】
)

func ReplaceNonDigitAlphaChar(in, to string) string {
	return NonDigitAlphaCharRegex.ReplaceAllLiteralString(in, to)
}
func ReplaceUnusualSpecialChar(in, to string) string {
	return UnusualSpecialCharRegex.ReplaceAllLiteralString(in, to)
}

var versionRegex = regexp.MustCompile(`^\d\.\d\.\d+?$`)

func IsVersion(ver string) bool {
	return versionRegex.MatchString(ver)
}

// EnglishWordsRegex 至少一个2字母的单词
var EnglishWordsRegex = regexp.MustCompile(`(\b| )[a-zA-Z]{2,}(\b| )`)
