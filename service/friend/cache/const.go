package cache

const CKeyPrefix = "SVC_FRIEND:"

const (
	// 仅缓存第一页
	friendListCacheKey = CKeyPrefix + "FriendList:uid_%v"
	followListCacheKey = CKeyPrefix + "OnewayList:uid_%v:follow"
	fansListCacheKey   = CKeyPrefix + "OnewayList:uid_%v:fans"

	// 访客功能
	saveVisitorCacheKey = CKeyPrefix + "SaveVisitorAt:uid_%v:targetId_%v"
	visitorListCacheKey = CKeyPrefix + "VisitorList:uid_%v"
)
