package cache

import (
	"context"
	"fmt"
	"microsvc/model/svc/moment"
	"time"
)

type momentCacheT struct {
	forwardsExpire time.Duration
}

var MomentCache = &momentCacheT{
	forwardsExpire: time.Hour * 24 * 3,
}

func (v momentCacheT) forwardsKey(mid, uid int64) string {
	return fmt.Sprintf(CKeyMomentForwards, mid, uid)
}

func (v momentCacheT) NeverForward(ctx context.Context, mid, uid int64) (bool, error) {
	b, err := moment.R.SetNX(ctx, v.forwardsKey(mid, uid), time.Now().Unix(), v.forwardsExpire).Result()
	return b, err
}
