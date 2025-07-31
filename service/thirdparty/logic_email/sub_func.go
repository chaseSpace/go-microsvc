package logic_email

import (
	"context"
	"fmt"
	"microsvc/infra/xgrpc"
	"microsvc/model/svc/thirdparty"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/thirdparty/cache"
	"microsvc/service/thirdparty/deploy"
	"microsvc/util/ulock"
	"strings"

	"gopkg.in/gomail.v2"
)

const emailNotifyHtml = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Email Verification Code Notification</title>
<style>
    body {
        font-family: Arial, sans-serif;
        background-color: #f4f4f4;
        padding: 20px;
    }
    .container {
        max-width: 600px;
        margin: 0 auto;
        background-color: #fff;
        padding: 20px;
        border-radius: 5px;
        box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
    }
    .verification-code {
        font-size: 24px;
        font-weight: bold;
        color: #0a59b0; /* Adding a color for emphasis */
    }
</style>
</head>
<body>
<div class="container">
    <p><strong>Email Verification Code From CocktailHack:</strong></p>
    <div class="verification-code">%v</div>
    <p>Please use this verification code to complete the verification process within %d minutes.</p>
    <p>If you did not initiate this action, please disregard this email.</p>
</div>
</body>
</html>`

func __sendEmailCode(ctx context.Context, email, code string, validMinute float64) (err error) {
	cc := deploy.ThirdpartyConf.Email

	m := gomail.NewMessage()
	m.SetHeader("From", cc.Username+" <"+cc.Sender+">")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "CocktailHack Register Verification Code")
	m.SetBody("text/html", fmt.Sprintf(emailNotifyHtml, code, int64(validMinute)))

	// 第4个参数是填授权码
	d := gomail.NewDialer(cc.SMTPHost, cc.SMTPPort, cc.Sender, cc.Password)
	err = d.DialAndSend(m)
	if err != nil {
		if strings.Contains(err.Error(), "contain a non-existent account") {
			return xerr.ErrThirdParty.AppendMsg("The email might be a non-existent account, please check it!")
		}
	}
	return xerr.ErrThirdParty.AutoAppend(err, true)
}

func __limiterCheck(ctx context.Context, email string, scene commonpb.EmailCodeScene) error {
	return ulock.QuickExec(ctx, thirdparty.R.Client, func(ctx context.Context) error {
		// check email/check ip
		err := cache.EmailCaptchaCache.CheckByEmail(ctx, email, int(scene))
		if err != nil {
			return err
		}
		ip := xgrpc.GetReqClientIP(ctx)
		err = cache.EmailCaptchaCache.CheckByIP(ctx, ip, int(scene))
		if err != nil {
			return err
		}
		//incr email/incr ip
		err = cache.EmailCaptchaCache.Counting(ctx, email, ip, int(scene))
		return err
	})
}

type __verifyEmailCodeReq struct {
	Email             string
	Code              string
	Scene             commonpb.EmailCodeScene
	DeleteAfterVerify bool
}

func __verifyEmailCode(ctx context.Context, req *__verifyEmailCodeReq) (isMatch bool, err error) {
	code, err := cache.EmailCaptchaCache.QueryCode(ctx, req.Email, int(req.Scene))
	if err != nil {
		return false, err
	}
	if code == "" {
		return false, xerr.ErrEmailCodeNeedSendFirst
	}
	isMatch = code == req.Code
	if isMatch && req.DeleteAfterVerify {
		err = cache.EmailCaptchaCache.Delete(ctx, req.Email, int(req.Scene))
	}
	return
}
