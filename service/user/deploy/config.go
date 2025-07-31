package deploy

import (
	"errors"
	"microsvc/deploy"
	"microsvc/util/utime"
	"time"
)

// SvcConfig 每个服务特有的配置结构
type SvcConfig struct {
	deploy.CommConfig   `mapstructure:"root"`
	Auth                `mapstructure:"auth"`
	WxApp               `mapstructure:"wx_app"`
	WxMini              `mapstructure:"wx_mini"`
	OpenSignInRateLimit bool `mapstructure:"open_sign_in_rate_limit"` // 开启登录频率限制
	OpenSignUpRateLimit bool `mapstructure:"open_sign_up_rate_limit"` // 开启注册频率限制
	DefaultAssets       `mapstructure:"default_assets"`
	OauthSupport        `mapstructure:"oauth_support"`
}

type Auth struct {
	TokenExpiry string `mapstructure:"token_expiry"`
}

func (s SvcConfig) GetTokenExpiry() (d time.Duration, e error) {
	d, e = utime.ParseDuration(s.TokenExpiry)
	if e != nil {
		return
	}
	if d == 0 && !deploy.XConf.IsDevEnv() { // 非dev环境不允许0
		return 0, errors.New("config: token_expiry cannot be 0 on non-dev environment")
	}
	return
}

func (s SvcConfig) SelfCheck() error {
	_, err := s.GetTokenExpiry()
	// add other check...
	return err
}

type WxApp struct {
	Appid     string `mapstructure:"appid"`
	AppSecret string `mapstructure:"app_secret"`
}

type WxMini struct {
	Appid     string `mapstructure:"appid"`
	AppSecret string `mapstructure:"app_secret"`
}

type DefaultAssets struct {
	Avatar string `mapstructure:"avatar"`
}

type OauthSupport struct {
	Github struct {
		ClientId     string `mapstructure:"client_id"`
		ClientSecret string `mapstructure:"client_secret"`
		RedirectUri  string `mapstructure:"redirect_uri"`
	}
}

var _ deploy.SvcConfImpl = new(SvcConfig)

// UserConf 变量命名建议使用服务名作为前缀，避免main文件引用到其他svc的配置变量
var UserConf = new(SvcConfig)

/* 服务内使用的非配置model */

type UpdateInfoRate struct {
	DurationLimit  string   `json:"duration_limit"`   // “2s”
	DateRangeLimit []string `json:"date_range_limit"` // [ "2024-07-01 00:00:00", "2025-07-11 00:00:00" ]
	Banned         bool     `json:"banned"`           // 开关，true表示禁止更新
	MaxHistoryLen  int64    `json:"max_history_len"`  // 最大历史记录长度（redis list）
}
