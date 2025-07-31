package implement

import (
	"context"
	"fmt"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/cache"
	"microsvc/service/user/dao"
	deploy2 "microsvc/service/user/deploy"
	"microsvc/util/db"
	"microsvc/util/ucrypto"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

type __SignInGithub struct{}

type __GithubResp struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-the-authenticated-user
type __GithubUserInfo struct {
	ID        int64     `json:"id"`         // 唯一ID
	Login     string    `json:"login"`      // Github id
	AvatarURL string    `json:"avatar_url"` // 头像
	HtmlURL   string    `json:"html_url"`   // 用户主页
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CheckSignInReq
// https://docs.github.com/zh/apps/oauth-apps/building-oauth-apps/creating-an-oauth-app
func (a __SignInGithub) CheckSignInReq(ctx context.Context, req *userpb.SignInAllReq) (*SignInExt, error) {
	if req.Code == "" {
		return nil, xerr.ErrParams.New("code is empty")
	}
	data := map[string]string{
		"client_id":     deploy2.UserConf.OauthSupport.Github.ClientId,
		"client_secret": deploy2.UserConf.OauthSupport.Github.ClientSecret,
		"code":          req.Code,
	}
	r := new(__GithubResp)
	_, buf, errs := gorequest.New().SendMap(data).Post("https://github.com/login/oauth/access_token").EndStruct(r)
	if len(errs) > 0 {
		return nil, xerr.ErrThirdParty.Append(errs[0])
	}
	if r.AccessToken == "" {
		xlog.Error("github oauth token is empty", zap.String("response", string(buf)))
		return nil, xerr.ErrThirdParty.New("no access_token returned")
	}
	if !lo.Contains(strings.Split(r.Scope, ","), "user:email") {
		return nil, xerr.ErrThirdParty.New("no permission to request user email")
	}
	// 请求用户信息
	info, err := a.__getGithubUserInfo(ctx, r.AccessToken)
	if err != nil {
		return nil, err
	}
	req.AnyAccount = cast.ToString(info.ID)
	return &SignInExt{OauthCacheKey: ucrypto.MustSha1(fmt.Sprintf("github:%s", req.AnyAccount+info.Login))}, nil
}

// See also: https://blog.csdn.net/weixin_53510183/article/details/126150594
func (a __SignInGithub) __getGithubUserInfo(ctx context.Context, token string) (*__GithubUserInfo, error) {
	r := new(__GithubUserInfo)
	_, buf, errs := gorequest.New().Get("https://api.github.com/user").
		Set("Authorization", "token "+token).
		EndStruct(r)
	if len(errs) > 0 {
		return nil, xerr.ErrThirdParty.Append(errs[0])
	}
	xlog.Info("__getGithubUserInfo", zap.String("TOKEN", token), zap.String("RESPONSE", string(buf)))
	if r.ID < 1 {
		return nil, xerr.ErrThirdParty.New("__getGithubUserInfo: no id returned")
	}
	return r, nil
}

func (__SignInGithub) QueryUser(ctx context.Context, req *userpb.SignInAllReq, ext *SignInExt) (*user.User, error) {
	_, row, err := dao.GetUserFromTh(ctx, commonpb.SignInType_SIT_THIRD_GITHUB, req.AnyAccount)
	if err != nil {
		return nil, err
	}
	if row.Uid == 0 {
		// Oauth登陆需要暂存三方信息
		err = cache.SigninCtrlT.SaveOauthUserInfo(ctx, ext.OauthCacheKey, &ext.OauthUserProfile)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	_, umodel, err := dao.GetUser(ctx, row.Uid)
	return umodel, err
}

func (a __SignInGithub) CheckSignUpReq(ctx context.Context, req *userpb.SignUpAllReq) (*SignInExt, error) {
	ext := new(SignInExt)
	err := cache.SigninCtrlT.GetOauthUserInfo(ctx, req.AnyAccount, &ext.OauthUserProfile)
	if err != nil {
		return nil, err
	}
	// 替换为三方唯一ID，用于下一步注册
	req.AnyAccount = ext.OauthUserProfile.UniqId

	// 复用三方基础信息
	req.Body.Avatar = ext.OauthUserProfile.Avatar
	req.Body.Nickname = ext.OauthUserProfile.AccountId
	return ext, err
}

func (a __SignInGithub) SignUp(ctx context.Context, req *userpb.SignUpAllReq, ext *SignInExt) (umodel *user.User, err error) {
	// 创建三方注册表记录
	row := &user.UserRegisterTh{
		Account: req.AnyAccount,
		ThType:  req.Type,
	}

	// Oauth三方注册时，不需要填密码
	err = user.Q.Transaction(func(tx *gorm.DB) error {
		umodel, err = Base{}.commonSignUp(ctx, tx, req.AnyAccount, "", "", req.Type, req.Body)
		if err != nil {
			return err
		}
		row.Uid = umodel.Uid
		err = dao.CreateUserTh(tx, row)
		if err == nil {
			return nil
		}
		conflictIdx := ""
		if db.IsMysqlDuplicateErr(err, &conflictIdx) {
			if conflictIdx == user.TableUserThUKAccountThType {
				// 仅在客户端并发调用接口时可能触发，已经注册成功，调用登录接口即可
				return xerr.ErrAccountAlreadyExists
			}
		}
		return xerr.WrapMySQL(err)
	})
	return
}
