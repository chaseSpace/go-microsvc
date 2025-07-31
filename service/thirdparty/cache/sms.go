package cache

import (
	"context"
	"fmt"
	"microsvc/bizcomm"
	"microsvc/model/svc/thirdparty"
	"microsvc/pkg/xerr"
	"time"
)

type SmsCacheT struct {
	captchaExpire          time.Duration
	captchaLimiterInterval time.Duration
	ipLimiterInterval      time.Duration
}

var SmsCache = &SmsCacheT{
	captchaExpire:          time.Minute,
	captchaLimiterInterval: time.Minute,
	ipLimiterInterval:      time.Second * 30,
}

func (SmsCacheT) key(account string, scene int) string {
	return fmt.Sprintf(CKeySmsCode, account, scene)
}

func (SmsCacheT) limiterKeyUidScene(account string, scene int) string {
	return fmt.Sprintf(CKeySmsCodeLimiterKeyAccountScene, account, scene)
}

func (SmsCacheT) limiterKeyIP(ip string) string {
	return fmt.Sprintf(CKeySmsCodeLimiterKeyIP, ip)
}

func (s SmsCacheT) AllowIP(ctx context.Context, ip string) (bool, error) {
	allow, err := bizcomm.Limiter(thirdparty.R.Client).Allow(ctx, s.limiterKeyIP(ip), s.ipLimiterInterval)
	return allow, err
}

func (s SmsCacheT) AllowAccountScene(ctx context.Context, account string, scene int) (bool, error) {
	allow, err := bizcomm.Limiter(thirdparty.R.Client).Allow(ctx, s.limiterKeyUidScene(account, scene), s.captchaLimiterInterval)
	return allow, err
}

func (s SmsCacheT) SaveSmsCode(ctx context.Context, account, code string, scene int) error {
	err := thirdparty.R.Set(ctx, s.key(account, scene), code, s.captchaExpire).Err()
	return xerr.WrapRedis(err)
}

func (s SmsCacheT) QuerySmsCode(ctx context.Context, account string, scene int) (string, error) {
	r := thirdparty.R.Get(ctx, s.key(account, scene))
	return r.Val(), xerr.WrapRedis(r.Err())
}

func (s SmsCacheT) Delete(ctx context.Context, account string, scene int) error {
	err := thirdparty.R.Del(ctx, s.key(account, scene)).Err()
	return xerr.WrapRedis(err)
}

func (s SmsCacheT) UpdateExpire(ctx context.Context, account string, scene int, seconds int64) error {
	err := thirdparty.R.Expire(ctx, s.key(account, scene), time.Duration(seconds)*time.Second).Err()
	return xerr.WrapRedis(err)
}
