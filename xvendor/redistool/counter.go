package redistool

import (
	"context"
	"fmt"
	"microsvc/consts"
	"microsvc/infra/cache"
	"microsvc/pkg/xerr"
	"time"
)

// Counter 模拟令牌限速器
type Counter struct {
	bizScene string // 业务场景
	rdb      *cache.RedisObj
	exp      time.Duration
}

// NewCounter 创建简单限速器
// biz: 业务场景完整描述，非唯一标识
func NewCounter(rdb *cache.RedisObj, biz string, exp time.Duration) *Counter {
	if biz == "" {
		panic("bizScene can not be empty")
	}
	return &Counter{
		bizScene: biz,
		rdb:      rdb,
		exp:      exp,
	}
}
func (c *Counter) key(biz string, suffix interface{}) string {
	return fmt.Sprintf("%s:%s:%v", consts.CounterKeyPrefix, biz, suffix)
}

func (c *Counter) GreatThanOrEqual(ctx context.Context, suffix interface{}, threshold int64) (bool, error) {
	ct, err := c.rdb.Get(ctx, c.key(c.bizScene, suffix)).Int64()
	return ct >= threshold, xerr.WrapRedis(err)
}

func (c *Counter) Incr(ctx context.Context, suffix interface{}, number ...int64) (ret int64, err error) {
	if len(number) == 0 {
		number = []int64{1}
	}
	ret, err = c.rdb.IncrBy(ctx, c.key(c.bizScene, suffix), number[0]).Result()
	if err != nil {
		return 0, xerr.WrapRedis(err)
	}
	if ret > number[0] {
		return
	}
	// 仅在首次设置过期时间，避免无限续期
	return ret, xerr.WrapRedis(c.rdb.Expire(ctx, c.key(c.bizScene, suffix), c.exp).Err())
}

func (c *Counter) Drop(ctx context.Context, suffix interface{}) error {
	return c.rdb.Del(ctx, c.key(c.bizScene, suffix)).Err()
}
