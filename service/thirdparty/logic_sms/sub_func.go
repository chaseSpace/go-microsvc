package logic_sms

import (
	"context"
	"microsvc/infra/xgrpc"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/thirdparty/cache"
	"microsvc/service/thirdparty/deploy"
	"microsvc/service/thirdparty/logic_sms/thirdparty"
	tencentSms "microsvc/service/thirdparty/logic_sms/thirdparty/tencent"
)

var provider thirdparty.SmsAPI

func MustInit(cc *deploy.SvcConfig) {
	provider = &tencentSms.SmsImpl{}
	provider.MustInit(cc.Sms.Tencent)
}

func __sendSmsCode(ctx context.Context, areaCode, phone, code string) (err error) {
	switch areaCode {
	case "86":
		err = provider.SendDomesticSmsCode(ctx, phone, code)
	default:
		err = provider.SendOverseasSmsCode(ctx, areaCode, phone, code)
	}
	return
}

func __limiterCheck(ctx context.Context, aphone string, scene commonpb.SmsCodeScene) error {
	ip := xgrpc.GetReqClientIP(ctx)

	// 以下频率限制策略需要按照 可访问频率 从低到高的顺序执行
	allow, err := cache.SmsCache.AllowIP(ctx, ip)
	if err != nil {
		return err
	}
	if !allow {
		return xerr.ErrTooManyRequests.AppendMsg("ip limit")
	}
	allow, err = cache.SmsCache.AllowAccountScene(ctx, aphone, int(scene))
	if err != nil {
		return err
	}
	if !allow {
		return xerr.ErrTooManyRequests.AppendMsg("uid+scene limit")
	}
	return nil
}
