package bizcomm

import (
	"microsvc/bizcomm/comminfra"
	"microsvc/xvendor"
	"microsvc/xvendor/xratelimit"
	"sync"

	"github.com/redis/go-redis/v9"
)

// 这里存放微服务可能需要的全局对象，全部使用懒加载方式初始化

var (
	limiterOnce sync.Once
	limiter     *xvendor.Limiter

	rateLimitProviderOnce sync.Once
	rateLimitProvider     *comminfra.APIRateLimitConfProvider
	// 其他
)

func Limiter(rdb *redis.Client) (l *xvendor.Limiter) {
	limiterOnce.Do(func() {
		if limiter == nil {
			limiter = xvendor.NewLimiter(rdb)
		}
	})
	return limiter
}

func APIRateLimiter(isOpen bool, rdb *redis.Client) *comminfra.APIRateLimitConfProvider {
	rateLimitProviderOnce.Do(func() {
		if rateLimitProvider == nil {
			limiter := xratelimit.New(xratelimit.NewRedisStore(rdb))
			rateLimitProvider = comminfra.NewAPIRateLimitConfProvider(isOpen, limiter)
			rateLimitProvider.AutoRefresh()
		}
	})
	return rateLimitProvider
}
