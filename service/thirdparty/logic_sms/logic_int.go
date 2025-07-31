package logic_sms

import (
	"context"
	"microsvc/bizcomm/commuser"
	"microsvc/consts"
	"microsvc/deploy"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/thirdparty/cache"
	"microsvc/util/urand"

	"go.uber.org/zap"
)

type intCtrl struct{}

var Int = intCtrl{} // 暴露struct而不是interface，方便IDE跳转

func (intCtrl) SendSmsCodeInt(ctx context.Context, req *thirdpartypb.SendSmsCodeIntReq) (r *thirdpartypb.SendSmsCodeIntRes, err error) {
	if !commuser.PhoneTool.IsPhoneAreaCodeSupported(req.AreaCode) {
		return nil, xerr.ErrNotSupportedPhoneArea
	}
	aphone := commuser.PhoneTool.GetDBPhone(req.AreaCode, req.Phone)
	var code string
	if deploy.XConf.Env.IsProd() {
		code = urand.Digits(consts.PhoneSmsCodeLen)
		err := __limiterCheck(ctx, aphone, req.Scene)
		if err != nil {
			return nil, err
		}
	} else {
		code = "123456"
	}
	err = cache.SmsCache.SaveSmsCode(ctx, aphone, code, int(req.Scene))
	if err != nil {
		return nil, err
	}
	xlog.Info("send sms code", zap.String("aphone", aphone), zap.String("code", code))
	if !req.TestOnly {
		err = __sendSmsCode(ctx, req.AreaCode, req.Phone, code)
	}
	return new(thirdpartypb.SendSmsCodeIntRes), err
}

func (intCtrl) VerifySmsCodeInt(ctx context.Context, req *thirdpartypb.VerifySmsCodeIntReq) (*thirdpartypb.VerifySmsCodeIntRes, error) {
	if req.Code == "" {
		return nil, xerr.ErrParams.New("code is empty")
	}
	if commonpb.SmsCodeScene_name[int32(req.Scene)] == "" || req.Scene == commonpb.SmsCodeScene_SCS_None {
		return nil, xerr.ErrParams.New("scene is invalid")
	}

	aphone := commuser.PhoneTool.GetDBPhone(req.AreaCode, req.Phone)
	code, err := cache.SmsCache.QuerySmsCode(ctx, aphone, int(req.Scene))
	if err != nil {
		return nil, err
	}
	if code == "" {
		return nil, xerr.ErrParams.New("Send verification code firstly")
	}
	if req.UpdateExpireSec > 0 {
		err = cache.SmsCache.UpdateExpire(ctx, aphone, int(req.Scene), req.UpdateExpireSec)
		if err != nil {
			return nil, err
		}
	} else if code == req.Code {
		err = cache.SmsCache.Delete(ctx, aphone, int(req.Scene))
		if err != nil {
			return nil, err
		}
	}
	return &thirdpartypb.VerifySmsCodeIntRes{Pass: code == req.Code}, nil
}
