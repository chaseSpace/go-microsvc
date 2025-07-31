package internal

import (
	"context"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/service/user/deploy"
	"microsvc/util/ulock"
	"microsvc/xvendor/redistool"
	"time"
)

// 登陆限速器（Counter实现）
type signInRateCheck struct {
	ipCounter  *redistool.Counter
	uidCounter *redistool.Counter
}

var SignInRateCheck = &signInRateCheck{
	ipCounter:  redistool.NewCounter(user.R, "sign_in", time.Minute*10),
	uidCounter: redistool.NewCounter(user.R, "sign_in", time.Minute*3),
}

func (s *signInRateCheck) CheckByIP(ctx context.Context, ip string) error {
	if !deploy.UserConf.OpenSignInRateLimit {
		return nil
	}
	err := ulock.QuickExec(ctx, user.R.Client, func(ctx context.Context) error {
		if gte, err := s.ipCounter.GreatThanOrEqual(ctx, ip, 10); err != nil {
			return err
		} else if gte {
			return xerr.ErrLoginFrequentlyLong
		}
		_, err := s.ipCounter.Incr(ctx, ip)
		return err
	})
	return err
}

func (s *signInRateCheck) CheckByUID(ctx context.Context, uid int64) error {
	if !deploy.UserConf.OpenSignInRateLimit {
		return nil
	}
	err := ulock.QuickExec(ctx, user.R.Client, func(ctx context.Context) error {
		if gte, err := s.uidCounter.GreatThanOrEqual(ctx, uid, 3); err != nil {
			return err
		} else if gte {
			return xerr.ErrLoginFrequently
		}
		_, err := s.uidCounter.Incr(ctx, uid)
		return err
	})
	return err
}

// 注册限速器（Counter实现）
type signUpRateCheck struct {
	ipCounter *redistool.Counter
}

var SignUpRateCheck = &signUpRateCheck{
	ipCounter: redistool.NewCounter(user.R, "sign_up", time.Hour*24),
}

func (s *signUpRateCheck) Check(ctx context.Context, ip string) error {
	if !deploy.UserConf.OpenSignUpRateLimit {
		return nil
	}
	if ct, err := s.ipCounter.Incr(ctx, ip); err != nil {
		return err
	} else if ct >= 5 { // 同一ip，24小时内注册5次，则限制
		return xerr.ErrTooManyRequests.New("Registration too frequent(1000)")
	}
	return nil
}
