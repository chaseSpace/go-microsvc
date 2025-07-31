package utype

import (
	"microsvc/util/ujson"

	"github.com/spf13/cast"
)

type Map map[string]interface{}

// ToStrAnyMap 将结构体转换为map[string]any
func ToStrAnyMap(v interface{}) map[string]any {
	r, e := cast.ToStringMapE(v)
	if e == nil {
		return r
	}
	r = make(map[string]any)
	if v == nil {
		return r
	}
	if buf, e := ujson.Marshal(v); e == nil {
		_ = ujson.Unmarshal(buf, &r)
		return r
	}
	return r
}

func Str2Ptr(src string) *string {
	if src != "" {
		return &src
	}
	return nil
}

func PtrOf[T any](v T) *T {
	return &v
}
