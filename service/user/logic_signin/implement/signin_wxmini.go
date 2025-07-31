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

// __SignInWxMini 微信小程序登录
// https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/login.html
type __SignInWxMini struct{}

func (a __SignInWxMini) CheckSignInReq(ctx context.Context, req *userpb.SignInAllReq) (*SignInExt, error) {
	return new(SignInExt), nil
}

func (a __SignInWxMini) CheckSignUpReq(ctx context.Context, req *userpb.SignUpAllReq) (*SignInExt, error) {
	return new(SignInExt), nil
}
func (__SignInWxMini) QueryUser(ctx context.Context, req *userpb.SignInAllReq, ext *SignInExt) (*user.User, error) {
	info, err := cache.WxMiniCtrl.GetWxMiniUserAccount(base.WxMiniCli, ctx, req.Code)
	if err != nil {
		return nil, err
	}
	wxModel, err := dao.GetUserWeixin(ctx, info.OpenID, enums.UserWxTypeMini)
	if err != nil {
		return nil, err
	}
	if wxModel.Uid == 0 {
		return nil, nil
	}
	_, umodel, err := dao.GetUser(ctx, wxModel.Uid)
	return umodel, err
}

func (m __SignInWxMini) SignUp(ctx context.Context, req *userpb.SignUpAllReq, ext *SignInExt) (umodel *user.User, err error) {
	info, err := cache.WxMiniCtrl.GetWxMiniUserAccount(base.WxMiniCli, ctx, req.Code)
	if err != nil {
		return nil, err
	}

	//info := &auth.ResCode2Session{OpenID: "oUg6H6omNletG_RYxVkiNOouB42w"}
	// 创建WX小程序注册表记录
	row := &user.UserRegisterWeixin{
		Account: info.OpenID,
		UnionId: info.UnionID,
		Type:    enums.UserWxTypeMini,
	}

	if req.Body == nil {
		req.Body = &userpb.SignUpBody{} // 小程序注册时拿不到用户信息
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
				return xerr.ErrInternal.New("Wechat mini-program account already exists")
			}
		}
		return err
	})
	return
}
