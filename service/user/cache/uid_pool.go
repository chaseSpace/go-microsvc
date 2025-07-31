package cache

import (
	"context"
	"microsvc/infra/cache"
	"microsvc/util"
	"microsvc/xvendor/genuserid"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
)

type UidQueuedPool struct {
	redisKey string
	client   *redis.Client
}

var _ genuserid.QueuedPool = (*UidQueuedPool)(nil)

func NewUidQueuedPool(key string, client *redis.Client) *UidQueuedPool {
	return &UidQueuedPool{
		redisKey: key,
		client:   client,
	}
}

func (u *UidQueuedPool) Push(ids []uint64) error {
	return u.client.LPush(context.TODO(), u.redisKey, lo.ToAnySlice(ids)...).Err()
}

func (u *UidQueuedPool) Pop() (uid uint64, err error) {
	ret := u.client.RPop(util.Ctx, u.redisKey)
	if cache.IsRedisErr(ret.Err()) {
		return 0, err
	}
	uid2, _ := ret.Uint64()
	return uid2, nil
}

func (u *UidQueuedPool) Size() (size int, err error) {
	ret := u.client.LLen(util.Ctx, u.redisKey)
	if ret.Err() != nil {
		return 0, ret.Err()
	}
	return int(ret.Val()), nil
}

func (u *UidQueuedPool) MaxUnusedUID() (uid uint64, err error) {
	ret := u.client.LIndex(util.Ctx, u.redisKey, 0)
	if cache.IsRedisErr(ret.Err()) {
		return 0, ret.Err()
	}
	return ret.Uint64()
}

func (u *UidQueuedPool) Reset() error {
	return u.client.Del(util.Ctx, u.redisKey).Err()
}
