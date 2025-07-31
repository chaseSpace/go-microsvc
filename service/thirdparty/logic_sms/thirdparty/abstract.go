package thirdparty

import (
	"context"
)

type SmsAPI interface {
	Name() string
	MustInit(config interface{})
	SendDomesticSmsCode(ctx context.Context, phone, code string) error
	SendOverseasSmsCode(ctx context.Context, areaCode, phone, code string) error
}
