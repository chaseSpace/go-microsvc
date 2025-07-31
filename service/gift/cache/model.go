package cache

import "microsvc/model/svc/gift"

type giftListT struct {
	List []*gift.GiftConf
	// 以后可能添加其他礼物元数据
}
