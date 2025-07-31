package implement

import (
	"context"
	"microsvc/enums"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/base"
	"microsvc/service/user/cache"
	"microsvc/service/user/dao"
	"microsvc/util/db"

	"gorm.io/gorm"
)

// __SignInWxApp 微信App原生登录
// https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Login/Development_Guide.html
type __SignInWxApp struct{}

func (a __SignInWxApp) CheckSignInReq(ctx context.Context, req *userpb.SignInAllReq) (*SignInExt, error) {
	return new(SignInExt), nil
}

func (a __SignInWxApp) CheckSignUpReq(ctx context.Context, req *userpb.SignUpAllReq) (*SignInExt, error) {
	return new(SignInExt), nil
}

func (__SignInWxApp) QueryUser(ctx context.Context, req *userpb.SignInAllReq, ext *SignInExt) (*user.User, error) {
	wxModel, err := dao.GetUserWeixin(ctx, req.AnyAccount, enums.UserWxTypeApp)
	if err != nil {
		return nil, err
	}
	if wxModel.Uid == 0 {
		return nil, nil
	}
	_, umodel, err := dao.GetUser(ctx, wxModel.Uid)
	return umodel, err
}

func (a __SignInWxApp) SignUp(ctx context.Context, req *userpb.SignUpAllReq, ext *SignInExt) (umodel *user.User, err error) {
	info, err := cache.WxAppCtrl.GetWxAppUserInfo(base.WxAppCli, ctx, req.Code)
	if err != nil {
		return nil, err
	}

	// 创建微信注册表记录
	row := &user.UserRegisterWeixin{
		Account:  info.OpenID,
		UnionId:  info.Unionid,
		Nickname: info.Nickname,
		Type:     enums.UserWxTypeApp,
	}

	// 事务内在两个表中插入记录
	err = user.Q.Transaction(func(tx *gorm.DB) error {
		umodel, err = Base{}.commonSignUp(ctx, tx, info.OpenID, "", "", req.Type, req.Body)
		if err != nil {
			return err
		}
		row.Uid = umodel.Uid
		err = dao.CreateUserWeixin(tx, row)
		if err == nil {
			return nil
		}
		conflictIdx := ""
		if db.IsMysqlDuplicateErr(err, &conflictIdx) {
			if conflictIdx == user.TableUserWxAppUKAccount {
				// 仅在客户端并发调用接口时可能触发，已经注册成功，调用登录接口即可
				return xerr.ErrParams.New("Wechat account already exists")
			}
		}
		return err
	})
	return
}
