package logic_signin

import (
	"context"
	"fmt"
	"microsvc/infra/svccli/rpc"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/dao"
	"microsvc/service/user/logic_punish"
	"microsvc/service/user/logic_signin/implement"
	"microsvc/service/user/logic_signin/internal"
)

type ctrl struct {
}

var Ext ctrl

func (c ctrl) SignInAll(ctx context.Context, req *userpb.SignInAllReq) (*userpb.SignInAllRes, error) {
	err := __beforeSignIn(ctx, req)
	if err != nil {
		return nil, err
	}

	impl := implement.SignInRegistry[req.Type]
	if impl == nil {
		return nil, xerr.ErrUnSupportedSignInType
	}
	// 1. 检查请求
	ext, err := impl.CheckSignInReq(ctx, req)
	if err != nil {
		return nil, err
	}
	// 2. 查询用户是否存在
	mod, err := impl.QueryUser(ctx, req, ext)
	if err != nil {
		return nil, err
	}
	if mod == nil {
		// 将查到的三方用户信息返回给前端（OAuth方式登录才有值），以便下一步调用注册接口
		return &userpb.SignInAllRes{OauthCacheKey: ext.OauthCacheKey}, nil
	}
	// 3. 用户存在，惩罚/密码检查
	err = c.__signInCheck(ctx, mod, req)
	if err != nil {
		return nil, err
	}
	token, expiry, err := implement.Base{}.GenSignToken(ctx, mod)
	if err != nil {
		return nil, err
	}

	// 写mq消息
	go __produceSignInMsg(ctx, mod, req)

	return &userpb.SignInAllRes{Token: token, Expiry: int64(expiry.Seconds()), Registered: true}, nil
}

// 登陆：通用检查
func (ctrl) __signInCheck(ctx context.Context, mod *user.User, req *userpb.SignInAllReq) (err error) {
	err = internal.SignInRateCheck.CheckByUID(ctx, mod.Uid)
	if err != nil {
		return err
	}
	// 是否禁止登录
	res, err := logic_punish.Int.GetUserPunish(ctx, &userpb.GetUserPunishReq{Uid: mod.Uid})
	if err != nil {
		return err
	}
	if len(res.Pmap) > 0 {
		if res.Pmap[int32(commonpb.PunishType_PT_SignIn)] != nil {
			return xerr.ErrSignInBanned
		}
		if res.Pmap[int32(commonpb.PunishType_PT_Ban)] != nil {
			return xerr.ErrAccountBanned
		}
	}

	// check input [token]
	switch req.Type {
	case commonpb.SignInType_SIT_PASSWORD, commonpb.SignInType_SIT_EMAIL: // check password
		if err = __checkPassword(req.Password, mod.Password, mod.PasswdSalt); err != nil {
			return err
		}
	case commonpb.SignInType_SIT_PHONE:
	default:
		return xerr.ErrUnSupportedSignInType
	}
	return nil
}

func (ctrl) SignUpAll(ctx context.Context, req *userpb.SignUpAllReq) (*userpb.SignUpAllRes, error) {
	err := __beforeSignUp(ctx, req)
	if err != nil {
		return nil, err
	}
	impl := implement.SignInRegistry[req.Type]
	if impl == nil {
		return nil, xerr.ErrUnSupportedSignInType
	}
	ext, err := impl.CheckSignUpReq(ctx, req)
	if err != nil {
		return nil, err
	}
	mod, err := impl.SignUp(ctx, req, ext)
	if err != nil {
		return nil, err
	}
	token, expiry, err := implement.Base{}.GenSignToken(ctx, mod)
	if err != nil {
		return nil, err
	}

	// 写mq消息
	go __produceSignUpMsg(ctx, mod)
	return &userpb.SignUpAllRes{Token: token, Expiry: int64(expiry.Seconds())}, nil
}

func (c ctrl) SendSmsCodeOnUser(ctx context.Context, req *userpb.SendSmsCodeOnUserReq) (r *userpb.SendSmsCodeOnUserRes, err error) {
	var u user.User
	account := fmt.Sprintf("%s|%s", req.AreaCode, req.Phone)
	getUser := func() {
		_, u, err = dao.GetUserByPhone(ctx, account)
	}
	switch req.Scene {
	case commonpb.SmsCodeScene_SCS_SignIn:
		getUser()
		if err != nil {
			return
		}
		if u.Uid == 0 {
			return nil, xerr.ErrParams.New("The phone number has not been registered")
		}
	case commonpb.SmsCodeScene_SCS_SignUp:
		getUser()
		if err != nil {
			return
		}
		if u.Uid != 0 {
			return nil, xerr.ErrParams.New("The phone number has been registered")
		}
	default:
		return nil, xerr.ErrParams.New("Unsupported sms code scene")
	}

	// 发送sms验证码
	_, err = rpc.Thirdparty().SendSmsCodeInt(ctx, &thirdpartypb.SendSmsCodeIntReq{
		AreaCode: req.AreaCode,
		Phone:    req.Phone,
		Scene:    req.Scene,
		TestOnly: req.TestOnly,
	})
	return
}

func (c ctrl) SendEmailCodeOnUser(ctx context.Context, req *userpb.SendEmailCodeOnUserReq) (r *userpb.SendEmailCodeOnUserRes, err error) {
	var u user.UserRegisterTh
	getUser := func() {
		_, u, err = dao.GetUserFromTh(ctx, commonpb.SignInType_SIT_EMAIL, req.Email)
	}

	// 针对不同场景添加不同的前置验证
	switch req.Scene {
	case commonpb.EmailCodeScene_ECS_ResetPasswd:
		// 若未注册，则返回提示
		getUser()
		if err != nil {
			return
		}
		if u.Uid == 0 {
			return nil, xerr.ErrParams.New("The email has not been registered")
		}
	case commonpb.EmailCodeScene_ECS_SignUp:
		// 若已注册，则返回提示
		getUser()
		if err != nil {
			return
		}
		if u.Uid != 0 {
			return nil, xerr.ErrParams.New("The email has been registered")
		}
	default:
		return nil, xerr.ErrParams.New("Unsupported email code scene")
	}

	// 发送邮箱验证码
	_, err = rpc.Thirdparty().SendEmailCodeInt(ctx, &thirdpartypb.SendEmailCodeIntReq{
		Email:    req.Email,
		Scene:    req.Scene,
		TestOnly: req.TestOnly,
	})
	return
}
