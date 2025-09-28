package implement

import (
	"context"
	"microsvc/deploy"
	"microsvc/infra/svccli/rpc"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/dao"
	"microsvc/util/urand"
	"microsvc/util/uregex"

	"gorm.io/gorm"
)

type __SignInEmail struct{}

func (a __SignInEmail) CheckSignInReq(ctx context.Context, req *userpb.SignInAllReq) (*SignInExt, error) {
	if uregex.IsInvalidEmail(req.AnyAccount) {
		return nil, xerr.ErrParams.New("invalid email: <%s>", req.AnyAccount)
	}
	return new(SignInExt), nil
}

func (a __SignInEmail) CheckSignUpReq(ctx context.Context, req *userpb.SignUpAllReq) (*SignInExt, error) {
	if !deploy.XConf.IsDevEnv() {
		// check email code
		res, err := rpc.Thirdparty().VerifyEmailCodeInt(ctx, &thirdpartypb.VerifyEmailCodeIntReq{
			Email: req.AnyAccount,
			Code:  req.VerifyCode,
			Scene: commonpb.EmailCodeScene_ECS_SignUp,
		})
		if err != nil {
			return nil, err
		}
		if !res.IsMatch {
			return nil, xerr.ErrInvalidVerifyCode.New("wrong verification code")
		}
	}

	if uregex.IsInvalidEmail(req.AnyAccount) {
		return nil, xerr.ErrParams.New("invalid email: <%s>", req.AnyAccount)
	}
	return new(SignInExt), nil
}

func (__SignInEmail) QueryUser(ctx context.Context, req *userpb.SignInAllReq, ext *SignInExt) (*user.User, error) {
	_, row, err := dao.GetUserByEmail(ctx, req.AnyAccount)
	if err != nil {
		return nil, err
	}
	if row.Uid == 0 {
		return nil, nil
	}
	_, umodel, err := dao.GetUser(ctx, row.Uid)
	return umodel, err
}

func (a __SignInEmail) SignUp(ctx context.Context, req *userpb.SignUpAllReq, ext *SignInExt) (mod *user.User, err error) {
	// 注意此时 password 一定非空
	salt := urand.Strings(4)
	err = user.Q.Transaction(func(tx *gorm.DB) error {
		mod, err = Base{}.commonSignUp(ctx, tx, req.AnyAccount, req.Password, salt, req.Type, req.Body)
		return err
	})
	return
}
