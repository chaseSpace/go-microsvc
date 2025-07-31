package logic_signin

import (
	"context"
	"microsvc/bizcomm/mq"
	"microsvc/consts"
	"microsvc/infra/xgrpc"
	"microsvc/infra/xmq"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/logic_signin/internal"
	"microsvc/util/ucrypto"
)

func __beforeSignIn(ctx context.Context, req *userpb.SignInAllReq) (err error) {
	// 简单的频率检查
	err = internal.SignInRateCheck.CheckByIP(ctx, xgrpc.GetReqClientIP(ctx))
	if err != nil {
		return err
	}
	return
}

func __checkPassword(inputHash, existHashWithSalt, salt string) error {
	if existHashWithSalt == "" {
		return xerr.ErrPasswordNotSetOnLogin
	}
	if inputHash == "" {
		return xerr.ErrPasswordNotGiven
	}
	inputHashWithSalt, _ := ucrypto.Sha1(inputHash, salt)
	if inputHashWithSalt != existHashWithSalt {
		return xerr.ErrSignInFailed
	}
	return nil
}

func __produceSignInMsg(ctx context.Context, mod *user.User, req *userpb.SignInAllReq) {
	xmq.Produce(consts.TopicSignIn, mq.NewMsgSignIn(
		&mq.SignInBody{
			UID:        mod.Uid,
			Nickname:   mod.Nickname,
			Firstname:  mod.Firstname,
			Lastname:   mod.Lastname,
			AppName:    req.Base.AppName,
			AppVersion: req.Base.AppVersion,
			IP:         xgrpc.GetReqClientIP(ctx),
			Platform:   req.Base.Platform,
			System:     req.Base.System,
		}),
	)
}
func __produceSignUpMsg(ctx context.Context, mod *user.User) {
	xmq.Produce(consts.TopicSignUp, mq.NewMsgSignUp(
		&mq.SignUpBody{
			UID:          mod.Uid,
			Nickname:     mod.Nickname,
			Firstname:    mod.Firstname,
			Lastname:     mod.Lastname,
			RegisteredAt: mod.CreatedAt.Unix(),
			RegChan:      mod.RegChannel,
		}))
}

// 注册：统一检查
func __beforeSignUp(ctx context.Context, req *userpb.SignUpAllReq) (err error) {
	// 简单的频率检查
	err = internal.SignUpRateCheck.Check(ctx, xgrpc.GetReqClientIP(ctx))
	if err != nil {
		return err
	}
	switch req.Type {
	case commonpb.SignInType_SIT_PHONE:
		if req.VerifyCode == "" {
			return xerr.ErrInvalidVerifyCode.New("please input verify code")
		}
	case commonpb.SignInType_SIT_PASSWORD, commonpb.SignInType_SIT_EMAIL:
		if req.Password == "" {
			return xerr.ErrParams.New("The password is required")
		}
		// 输入密码必须是合法密文
		err := user.InfoStaticCheckCtrl.CheckPassword(req.Password)
		if err != nil {
			return err
		}
	default:
		return xerr.ErrParams.New("not supported signup type:" + req.Type.String())
	}
	return nil
}
