package cache

import (
	"context"
	"fmt"
	"microsvc/model/svc/thirdparty"
	"microsvc/pkg/xerr"
	"microsvc/xvendor/redistool"
	"time"
)

type EmailCaptchaCacheT struct {
	captchaExpire           time.Duration
	emailCounter, ipCounter *redistool.Counter
}

var EmailCaptchaCache = &EmailCaptchaCacheT{
	captchaExpire: time.Minute * 10,
	emailCounter:  redistool.NewCounter(thirdparty.R, "email_captcha_counter_email", time.Minute),
	ipCounter:     redistool.NewCounter(thirdparty.R, "email_captcha_counter_ip", time.Hour),
}

func (EmailCaptchaCacheT) key(account string, scene int) string {
	return fmt.Sprintf(CKeyEmailCode, account, scene)
}

func (EmailCaptchaCacheT) counterKeySuffix(ip string, scene int) string {
	return fmt.Sprintf(`%s:%d`, ip, scene)
}

func (s EmailCaptchaCacheT) CheckByEmail(ctx context.Context, email string, scene int) error {
	suffix := fmt.Sprintf(`%s:%d`, email, scene)
	if gt, err := s.emailCounter.GreatThanOrEqual(ctx, suffix, 1); err != nil {
		return err
	} else if gt {
		return xerr.ErrTooManyRequests.New("speed limit(email)")
	}
	return nil
}

func (s EmailCaptchaCacheT) CheckByIP(ctx context.Context, ip string, scene int) error {
	suffix := s.counterKeySuffix(ip, scene)
	if gt, err := s.ipCounter.GreatThanOrEqual(ctx, suffix, 5); err != nil {
		return err
	} else if gt {
		return xerr.ErrTooManyRequests.New("speed limit(ip)")
	}
	return nil
}

func (s EmailCaptchaCacheT) Counting(ctx context.Context, email, ip string, scene int) error {
	_, err1 := s.emailCounter.Incr(ctx, s.counterKeySuffix(email, scene))
	_, err2 := s.ipCounter.Incr(ctx, s.counterKeySuffix(ip, scene))
	return xerr.JoinErrors(err1, err2)
}

func (s EmailCaptchaCacheT) SaveCode(ctx context.Context, account, code string, scene int) error {
	err := thirdparty.R.Set(ctx, s.key(account, scene), code, s.captchaExpire).Err()
	return xerr.WrapRedis(err)
}

func (s EmailCaptchaCacheT) QueryCode(ctx context.Context, account string, scene int) (string, error) {
	r := thirdparty.R.Get(ctx, s.key(account, scene))
	return r.Val(), xerr.WrapRedis(r.Err())
}

func (s EmailCaptchaCacheT) Delete(ctx context.Context, account string, scene int) error {
	err := thirdparty.R.Del(ctx, s.key(account, scene)).Err()
	return xerr.WrapRedis(err)
}

func (s EmailCaptchaCacheT) CodeExpiry() time.Duration {
	return s.captchaExpire
}
