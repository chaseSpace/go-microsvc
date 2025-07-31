package cache

import "time"

const CKeyPrefix = "SVC_USER:"

const (
	UserInfoCacheKey      = CKeyPrefix + "UserInfo:%v"
	UserInfoUpdateHistory = CKeyPrefix + "UserInfoUpdateHistory:%d:%s"
	WxAppUserInfoKey      = CKeyPrefix + "WxAppUserInfo:openid:%v"
	OauthUserInfoKey      = CKeyPrefix + "OauthUserInfo:%v"
)

const (
	UserInfoExpiry = time.Minute * 10
)

// 在一个Lua脚本中操作redis list
const updateUserInfoHistoryLuaScript = `
	local key = KEYS[1]
	local element = ARGV[1]
	
	redis.call('LPUSH', key, element)
	redis.call('LTRIM', key, 0, %d-1) -- 保留最新的N条记录
`
