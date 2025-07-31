package ucache

import (
	"github.com/patrickmn/go-cache"
	"time"
)

/*
 xcache包 统一管理 进程缓存
*/

var defaultCache = cache.New(time.Minute*5, time.Minute*10)

func New() *cache.Cache {
	return cache.New(time.Minute*5, time.Minute*10)
}

func NewWithArgs(defaultExpiration, cleanupInterval time.Duration) *cache.Cache {
	return cache.New(defaultExpiration, cleanupInterval)
}

func Set(key string, val interface{}, exp time.Duration) {
	defaultCache.Set(key, val, exp)
}

func SetDefault(key string, val interface{}) {
	defaultCache.SetDefault(key, val)
}

func Del(key string) {
	defaultCache.Delete(key)
}

func Get(key string) (interface{}, bool) {
	return defaultCache.Get(key)
}
