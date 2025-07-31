package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"microsvc/util/db"
	"microsvc/util/ujson"
	"time"

	"github.com/silenceper/wechat/v2/miniprogram"
	"github.com/silenceper/wechat/v2/miniprogram/auth"
	"github.com/silenceper/wechat/v2/officialaccount"
	"github.com/silenceper/wechat/v2/officialaccount/oauth"
	"go.uber.org/zap"
)

type wxAppCtrlT struct {
}

var WxAppCtrl = wxAppCtrlT{}

func (wxAppCtrlT) userInfoKey(openid string) (string, time.Duration) {
	return fmt.Sprintf(WxAppUserInfoKey, openid), time.Minute * 5
}

func (w wxAppCtrlT) GetWxAppUserInfo(cli *officialaccount.OfficialAccount, ctx context.Context, code string) (*oauth.UserInfo, error) {
	res, err := cli.GetOauth().GetUserAccessTokenContext(ctx, code)
	if err != nil {
		return nil, xerr.ErrThirdParty.New("Fetch weixin user access_token failed-1: %s", err)
	}

	xlog.Info("wxAppCtrlT.GetWxAppAccessToken OK", zap.Any("res", res), zap.String("code", code))
	if res.AccessToken == "" {
		return nil, xerr.ErrThirdParty.New("Fetch weixin user access_token failed-2: %+v", res)
	}

	return w.getUserInfo(cli, ctx, res.AccessToken, res.OpenID)
}

func (w wxAppCtrlT) getUserInfo(cli *officialaccount.OfficialAccount, ctx context.Context, accessToken, openid string) (*oauth.UserInfo, error) {
	key, exp := w.userInfoKey(openid)
	r := user.R.Get(ctx, key)
	if db.IgnoreNilErr(r.Err()) != nil {
		return nil, xerr.WrapRedis(r.Err())
	}
	if r.Val() != "" {
		val := new(oauth.UserInfo)
		_ = json.Unmarshal([]byte(r.Val()), val)
		return val, nil
	}
	// lang: 国家地区语言版本，zh_CN 简体，zh_TW 繁体，en 英语，默认为 en
	res2, err := cli.GetOauth().GetUserInfoContext(ctx, accessToken, openid, "en")
	if err != nil {
		return nil, xerr.ErrThirdParty.New("Fetch weixin user info failed-1: %s", err)
	}

	xlog.Info("wxAppCtrlT.GetUserInfoContext OK", zap.Any("res2", res2), zap.String("openid", openid))
	if res2.OpenID == "" {
		return nil, xerr.ErrThirdParty.New("Fetch weixin user info failed-2: %+v", res2)
	}
	// 写入缓存（因为获取用户信息接口有频率限制）
	err = user.R.Set(ctx, key, util.ToJsonStr(res2), exp).Err()
	return &res2, err
}

type wxMiniCtrlT struct {
}

var WxMiniCtrl = wxMiniCtrlT{}

func (w wxMiniCtrlT) GetWxMiniUserAccount(cli *miniprogram.MiniProgram, ctx context.Context, code string) (*auth.ResCode2Session, error) {
	res, err := cli.GetAuth().Code2SessionContext(ctx, code)
	if err != nil {
		return nil, xerr.ErrThirdParty.New("Fetch weixin-miniprogram user openid failed-1: %s", err)
	}

	xlog.Info("GetWxMiniUserInfo.Code2SessionContext OK", zap.Any("res", res), zap.String("code", code))
	if res.OpenID == "" {
		return nil, xerr.ErrThirdParty.New("Fetch weixin-miniprogram user openid failed-2: %+v", res)
	}

	return &res, nil
}

type signinCtrlT struct {
}

var SigninCtrlT = signinCtrlT{}

func (s signinCtrlT) OauthInfoCacheKey(suffix string) (string, time.Duration) {
	return fmt.Sprintf(OauthUserInfoKey, suffix), time.Minute * 5
}

func (s signinCtrlT) SaveOauthUserInfo(ctx context.Context, suffix string, v any) error {
	key, exp := s.OauthInfoCacheKey(suffix)
	return user.R.Set(ctx, key, util.ToJsonStr(v), exp).Err()
}
func (s signinCtrlT) GetOauthUserInfo(ctx context.Context, suffix string, v any) error {
	key, _ := s.OauthInfoCacheKey(suffix)
	r, err := user.R.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return ujson.Unmarshal(r, v)
}
