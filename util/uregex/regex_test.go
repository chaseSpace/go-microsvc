package uregex

import (
	"microsvc/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractJson(t *testing.T) {
	type A struct {
		A int
	}
	var (
		a  A
		a1 A
	)
	util.AssertNilErr(ExtractJson(`{"a":1}`, &a))
	assert.Equal(t, 1, a.A)

	util.AssertNilErr(ExtractJson(`xcqec1=-.{"a":1}xxx,./`, &a1))
	assert.Equal(t, 1, a1.A)

	type B struct {
		B A
	}
	var (
		b B
	)
	util.AssertNilErr(ExtractJson(`xcqec1=-.{"b": {"a": 1}}xxx,./`, &b))
	assert.Equal(t, 1, b.B.A)
}

type EmailTest struct {
	Email string
	Valid bool
}

func Test(t *testing.T) {
	testCases := []EmailTest{
		{"1@example.com", true},
		{"test@example.com", true},
		{"test-email@example.co.uk", true},
		{"user-name@sub-domain.domain.co", true},
		{"user-name@sub-domain.domain.com", true},

		{"user.name@domain.com", false},
		{"user-name+tag@sub.domain.com", false},
		{"user@.com", false},
		{"@example.com", false},
		{"user@com", false},
		{"user@domain..com", false},
		{"user@domain..com", false},
		{"user@domain.c-om", false},
		{"user@domain.com-", false},
		{"user@domain-.com", false},
		{"user@domain..com", false},
	}

	for _, test := range testCases {
		if r := emailRegex.MatchString(test.Email); r != test.Valid {
			t.Fatalf(`For <%s>, ret should be %v, not %v`, test.Email, test.Valid, r)
		}
	}
}

// TestIsLegalPasswd is a test function for IsLegalPasswd.
func TestIsLegalPasswdLv1(t *testing.T) {
	testCases := []struct {
		passwd   string
		expected bool
	}{
		// level1: 包含大小写和数字，8~15位
		{"short", false},
		{"short1", false},
		{"shorT1", false},
		{"enoughab", false},
		{"enough12", false},
		{"enougH12", true},
		{"enougH1.", true},
		{"enougH1.long", true},
	}

	for _, tc := range testCases {
		t.Run(tc.passwd, func(t *testing.T) {
			result := IsLegalPasswd(tc.passwd, 1)
			if result != tc.expected {
				t.Errorf("IsLegalPasswd(%s) = %v; expected %v", tc.passwd, result, tc.expected)
			}
		})
	}
}

func TestReplaceNonDigitAlphaChar(t *testing.T) {
	// 定义一个字符串，其中包含一些特殊字符（但不能排除正常语音）
	str := `Hello, World! This is a test string with special chars|字符串|自然復仇|བཀྲ་ཤིས་བ: #$%^&*()_+-=[]{}|;':",./<>?`
	after := ReplaceNonDigitAlphaChar(str, "")
	t.Log(after)
}

func TestReplaceUnusualSpecialChar(t *testing.T) {
	// 定义一个字符串，其中包含一些特殊字符
	str := `Hello, World! This is a test string with special chars|字符串|自然復仇|བཀྲ་ཤིས་བ: #$%^&*()_+-=[]{}|;':",./<>?--`
	after := ReplaceUnusualSpecialChar(str, "")
	t.Log(after)
}

func TestOnly(t *testing.T) {
	str := " -"
	t.Log(OnlySpecialCharRegex.MatchString(str))
}

func TestIsVersion_InvalidVersion_ReturnsFalse(t *testing.T) {
	validVersions := []string{
		"1.0.01",
		"1.1.0",
	}
	for _, version := range validVersions {
		if !IsVersion(version) {
			t.Errorf("Expected %s to be a valid version", version)
		}
	}
}

func TestEnglishWordsRegex(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", true},
		{"a b", false},
		{"hi w", true},
		{"hi wr", true},
	}

	for _, test := range tests {
		actual := EnglishWordsRegex.MatchString(test.input)
		if actual != test.expected {
			t.Errorf("EnglishWordsRegex.MatchString(%q) = %v; want %v", test.input, actual, test.expected)
		}
	}
}
