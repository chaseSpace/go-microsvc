package xerr

import (
	"errors"
	"fmt"
	"microsvc/protocol/svc/commonpb"
	"testing"
)

func TestWrapRedis(t *testing.T) {
	println(WrapRedis(fmt.Errorf("xxx")).Error())
}

func TestAppend(t *testing.T) {
	var list = []struct {
		name     string
		actual   string
		expected string
	}{
		{
			name:     "append",
			actual:   ErrInternal.Append(errors.New("xxx")).Translate(commonpb.Lang_CL_CN).FlatMsg(),
			expected: "code:500 - msg:内部服务器错误 - chains: [code:520 - msg:xxx]",
		},
		{
			name:     "append 2",
			actual:   ErrRPC.Append(ErrInternal.Append(errors.New("xxx")).Translate(commonpb.Lang_CL_CN)).FlatMsg(),
			expected: "code:1016 - msg:ErrRPC - chains: [code:500 - msg:内部服务器错误 - chains: [code:520 - msg:xxx]]",
		},
		//{
		//	actual:   ErrInternal.Append(ErrAccountAlreadyExists).FlatMsgSimple(),
		//	expected: "code:500 - msg:ErrInternal - chains: [code:400 - msg:ErrParams ➜ AccountAlreadyExists]",
		//},
		//{
		//	actual:   ErrInternal.Append(ErrAccountAlreadyExists.Translate(commonpb.Lang_CL_CN)).Translate(commonpb.Lang_CL_CN).FlatMsgSimple(),
		//	expected: "code:500 - msg:内部服务器错误 - chains: [code:500 - msg:Duplicated error translate operation]",
		//},
		//{
		//	actual:   ErrInternal.Append(ErrAccountAlreadyExists).Translate(commonpb.Lang_CL_CN).FlatMsgSimple(),
		//	expected: "code:500 - msg:内部服务器错误 - chains: [code:400 - msg:账户已存在]",
		//},
		//{
		//	actual:   ErrInternal.Append(ErrAccountAlreadyExists).Append(ErrRedis).FlatMsgSimple(),
		//	expected: "code:500 - msg:ErrInternal - chains: [code:400 - msg:ErrParams ➜ AccountAlreadyExists] ➜ [code:512 - msg:ErrRedis]",
		//},
	}
	for _, item := range list {
		t.Run(item.name, func(t *testing.T) {
			if item.actual != item.expected {
				t.Errorf("actual: %s |||  expected: %s", item.actual, item.expected)
			}
		})
	}
}
