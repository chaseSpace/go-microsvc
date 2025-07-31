package commcache

const cKeyPrefix = "MICROSVC:"

const (
	CacheKeyRateLimitByIP  = cKeyPrefix + "rate_limit:by_ip:%s:%v"
	CacheKeyRateLimitByUID = cKeyPrefix + "rate_limit:by_uid:%d:%v"
)
