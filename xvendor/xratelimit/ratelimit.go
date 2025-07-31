package xratelimit

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

var (
	ErrRateLimit       = errors.New("RateLimiter")
	ErrRateLimitStore  = errors.Wrap(ErrRateLimit, "store error")
	ErrRateLimitParams = errors.Wrap(ErrRateLimit, "params error")
)

// IncrByRespCode provide to Store implementation used only.
type IncrByRespCode int8

const (
	IncrByRespCodeInvalidArgs IncrByRespCode = -1
	IncrByRespCodeQPSFulled   IncrByRespCode = 0
	// if qps is not full yet and qps+count would not full, return qps+count (must be greater than 0)
)

type RateLimiter struct {
	store Store
}

type Store interface {
	AtomicIncrBy(ctx context.Context, key, unixTime string, count, maxQPS int64) (qps IncrByRespCode, err error)
}

func New(store Store) *RateLimiter {
	return &RateLimiter{
		store: store,
	}
}

func (r *RateLimiter) Allow(ctx context.Context, key string, maxQPS int64) (bool, error) {
	return r.AllowN(ctx, key, 1, maxQPS)
}

func (r *RateLimiter) AllowN(ctx context.Context, key string, count, maxQPS int64) (bool, error) {
	if count > maxQPS || count < 1 || maxQPS < 1 {
		return false, errors.Wrap(ErrRateLimitParams, "count > maxQPS or count < 1 or maxQPS < 1")
	}

	unixTime := fmt.Sprintf("%d", time.Now().Unix())
	code, err := r.store.AtomicIncrBy(ctx, key, unixTime, count, maxQPS)
	if err != nil {
		return false, errors.Wrap(ErrRateLimitStore, err.Error())
	}
	if code > IncrByRespCodeQPSFulled { // pass
		return true, nil
	} else if code == IncrByRespCodeInvalidArgs {
		return false, ErrRateLimitParams
	} else if code != IncrByRespCodeQPSFulled {
		return false, errors.Wrap(ErrRateLimitStore, fmt.Sprintf("unknown code %v", code))
	}
	return false, nil
}

func (r *RateLimiter) Wait(ctx context.Context, key string, maxQPS int64) (bool, error) {
	return r.WaitN(ctx, key, 1, maxQPS)
}

func (r *RateLimiter) WaitN(ctx context.Context, key string, count, maxQPS int64) (bool, error) {
	if count > maxQPS || count < 1 || maxQPS < 1 {
		return false, errors.Wrap(ErrRateLimitParams, "count > maxQPS or count < 1 or maxQPS < 1")
	}
	for {
		select {
		case <-ctx.Done():
			return false, context.DeadlineExceeded
		default:
			unixTime := fmt.Sprintf("%d", time.Now().Unix())
			code, err := r.store.AtomicIncrBy(ctx, key, unixTime, count, maxQPS)
			if err != nil {
				return false, errors.Wrap(ErrRateLimitStore, "store.AtomicIncrBy")
			}
			if code > IncrByRespCodeQPSFulled { // pass
				return true, nil
			} else if code == IncrByRespCodeInvalidArgs {
				return false, ErrRateLimitParams
			} else if code != IncrByRespCodeQPSFulled {
				return false, errors.Wrap(ErrRateLimitStore, fmt.Sprintf("unknown code %v", code))
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
}
