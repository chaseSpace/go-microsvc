package xvendor

import (
	"context"
	"github.com/redis/go-redis/v9"
	"microsvc/pkg/xerr"
	"time"
)

// 一个简单的redis实现的限速器（只能以interval形式使用）

type Limiter struct {
	rdb *redis.Client
}

func NewLimiter(rdb *redis.Client) *Limiter {
	return &Limiter{rdb: rdb}
}

func (l *Limiter) Allow(ctx context.Context, identity string, interval time.Duration) (bool, error) {
	r := l.rdb.SetNX(ctx, identity, time.Now().UnixMilli(), interval)
	return r.Val(), xerr.WrapRedis(r.Err())
}

func (l *Limiter) Reset(ctx context.Context, identity string) error {
	return xerr.WrapRedis(l.rdb.Del(ctx, identity).Err())
}
