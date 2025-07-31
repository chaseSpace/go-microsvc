package logic_email

import (
	"context"
	"microsvc/consts"
	"microsvc/deploy"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/thirdparty/cache"
	"microsvc/util/urand"

	"go.uber.org/zap"
)

type intCtrl struct{}

var Int = intCtrl{} // 暴露struct而不是interface，方便IDE跳转

func (intCtrl) SendEmailCodeInt(ctx context.Context, req *thirdpartypb.SendEmailCodeIntReq) (r *thirdpartypb.SendEmailCodeIntRes, err error) {
	var code string
	if deploy.XConf.Env.IsProd() {
		code = urand.Digits(consts.EmailCodeLen)
		err = __limiterCheck(ctx, req.Email, req.Scene)
		if err != nil {
			return nil, err
		}
	} else {
		code = "123456"
	}
	err = cache.EmailCaptchaCache.SaveCode(ctx, req.Email, code, int(req.Scene))
	if err != nil {
		return nil, err
	}
	xlog.Info("send email code", zap.String("to_email", req.Email), zap.String("code", code))
	if !req.TestOnly {
		err = __sendEmailCode(ctx, req.Email, code, cache.EmailCaptchaCache.CodeExpiry().Minutes())
	}
	return new(thirdpartypb.SendEmailCodeIntRes), err
}

func (c intCtrl) VerifyEmailCodeInt(ctx context.Context, req *thirdpartypb.VerifyEmailCodeIntReq) (*thirdpartypb.VerifyEmailCodeIntRes, error) {
	match, err := __verifyEmailCode(ctx, &__verifyEmailCodeReq{
		Email:             req.Email,
		Code:              req.Code,
		Scene:             req.Scene,
		DeleteAfterVerify: req.DeleteAfterVerify,
	})
	return &thirdpartypb.VerifyEmailCodeIntRes{IsMatch: match}, err
}
