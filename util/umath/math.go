package umath

import "golang.org/x/exp/constraints"

func Abs[T constraints.Integer | constraints.Float](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// EnsurePositive 确保数字为正整数
func EnsurePositive[T constraints.Integer](num T) T {
	if num < 0 {
		return 0
	}
	return num
}
