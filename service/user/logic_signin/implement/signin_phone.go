package implement

import (
	"context"
	"microsvc/bizcomm/commuser"
	"microsvc/deploy"
	"microsvc/infra/svccli/rpc"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/dao"

	"gorm.io/gorm"
)

type __SignInPhone struct{}

func (s __SignInPhone) CheckSignInReq(ctx context.Context, req *userpb.SignInAllReq) (*SignInExt, error) {
	ext, err := s.__checkSmsCode(ctx, req.AnyAccount, req.VerifyCode)
	if err != nil {
		return nil, err
	}
	return ext, nil
}

func (s __SignInPhone) QueryUser(ctx context.Context, req *userpb.SignInAllReq, ext *SignInExt) (*user.User, error) {
	_, umodel, err := dao.GetUserByPhone(ctx, req.AnyAccount)
	if err != nil {
		return nil, err
	}
	if umodel.Uid == 0 {
		return nil, xerr.ErrUnRegisteredPhone
	}
	return &umodel, nil
}

func (s __SignInPhone) CheckSignUpReq(ctx context.Context, req *userpb.SignUpAllReq) (*SignInExt, error) {
	ext, err := s.__checkSmsCode(ctx, req.AnyAccount, req.VerifyCode)
	if err != nil {
		return nil, err
	}
	return ext, nil
}

func (s __SignInPhone) SignUp(ctx context.Context, req *userpb.SignUpAllReq, ext *SignInExt) (mod *user.User, err error) {
	err = user.Q.Transaction(func(tx *gorm.DB) error {
		mod, err = Base{}.commonSignUp(ctx, tx, req.AnyAccount, "", "", req.Type, req.Body)
		return err
	})
	return
}

func (s __SignInPhone) __checkSmsCode(ctx context.Context, anyAccount, code string) (*SignInExt, error) {
	if code == "" {
		return nil, xerr.ErrInvalidVerifyCode.New("please input verify code")
	}
	areaCode, phone, err := commuser.PhoneTool.ParsePhoneStr(anyAccount)
	if err != nil {
		return nil, err
	}
	_, err = commuser.PhoneTool.CheckPhone(areaCode, phone)
	if err != nil {
		return nil, err
	}
	if err = commuser.CheckPhoneSmsCode(code); err != nil {
		return nil, err
	}
	if deploy.XConf.IsDevEnv() {
		// TEMP TEST 开发环境不检查验证码
		return nil, nil
	}
	_, err = rpc.Thirdparty().VerifySmsCodeInt(ctx, &thirdpartypb.VerifySmsCodeIntReq{
		AreaCode: areaCode,
		Phone:    phone,
		Code:     code,
		Scene:    commonpb.SmsCodeScene_SCS_SignIn,
		// 更新验证码有效期，以便注册时重用（这个时间略大于前端用户填写资料的估计耗时）
		// 若前端填写资料时间过长，则验证码失效，用户需要重新获取验证码
		UpdateExpireSec: 60 * 3,
	})
	return &SignInExt{
		PhoneAreaCode: areaCode,
		Phone:         phone,
	}, err
}
