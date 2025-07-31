package urand

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/rand"
)

func RandIntRange(left, right int, duplicate map[int]int) int {
	if left >= right {
		return left
	}
	delta := right - left
	if len(duplicate) == delta+1 { // full
		return 0
	}
	for {
		i := rand.Intn(delta+1) + left
		if duplicate != nil && duplicate[i] == 0 {
			duplicate[i] = 1
			return i
		}
	}
}

const asciiChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Strings(length int, lowercase ...bool) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = asciiChars[rand.Intn(len(asciiChars))]
	}
	if len(lowercase) > 0 && lowercase[0] {
		return strings.ToLower(string(result))
	}
	return string(result)
}

func Digits(length int) string {
	rand.Seed(uint64(time.Now().UnixNano())) // 确保每次运行时生成的随机数不同
	var digits string
	for i := 0; i < length; i++ {
		digits += fmt.Sprintf("%d", rand.Intn(10)) // 生成0到9之间的随机数字
	}
	return digits
}

func Int31n(base, n int32) int32 {
	rand.Seed(uint64(time.Now().Nanosecond()))
	return base + rand.Int31n(n)
}

func Int63n(base, n int64) int64 {
	rand.Seed(uint64(time.Now().Nanosecond()))
	return base + rand.Int63n(n)
}
