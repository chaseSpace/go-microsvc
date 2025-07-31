package genuserid

import (
	"context"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

/*
递增 userid 生成模块（号池版本，支持高并发调用）
*/

type IncrementalPoolUIDGenerator struct {
	poolSizeThreshold int // 池剩余id数量<=这个值就会扩容
	maxPoolSize       int // 池容量
	pushIdsBuffer     []uint64

	// 分布式锁
	locker DistributeLock

	getCurrentMaxUID func() (uint64, error)
	skipFn           func(uint64) (bool, error)
	pool             QueuedPool
}

type DistributeLock interface {
	Lock(ctx, taskCtx context.Context) error
	Unlock(ctx context.Context) error
}

type QueuedPool interface {
	Push(ids []uint64) error
	Pop() (uid uint64, err error)
	Size() (size int, err error)
	MaxUnusedUID() (uid uint64, err error)
}

func (u *IncrementalPoolUIDGenerator) GenUid(ctx context.Context) (uid uint64, err error) {
	var cc = make(chan struct{})

	var size int
	var insufficient bool
	//st := time.Now()
	//defer func() {
	//	println(333, time.Since(st).String())
	//}()

	// 当池容量 远小于 并发请求数时，循环次数会>2（所以要根据预估的业务的并发请求数 来配置 池容量，避免此操作阻塞过久）
	for i := 0; ; i++ {
		go func() {
			defer func() {
				cc <- struct{}{}
			}()
			uid, err = u.pool.Pop()
			if err != nil {
				err = xerr.ErrInternal.New("pool.Pop").AutoAppend(err)
				return
			}
			if uid == 0 {
				err = u.fillPool(ctx, size)
				err = xerr.ErrInternal.New("fillPool-1").AutoAppend(err)
			} else {
				if insufficient, err = u.isPoolInsufficient(ctx); err == nil && insufficient {
					err = u.fillPool(ctx, size)
					err = xerr.ErrInternal.New("fillPool-2").AutoAppend(err)
				}
			}
			//println(777, i, uid, size, err)
		}()

		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-cc:
			if err != nil || uid > 0 {
				return
			}
		}
	}
}

// isPoolInsufficient 是否号池id不足
func (u *IncrementalPoolUIDGenerator) isPoolInsufficient(ctx context.Context) (insufficient bool, err error) {
	size, err := u.pool.Size()
	if err != nil {
		return
	}
	return size <= u.poolSizeThreshold, nil
}

// 填充号池
func (u *IncrementalPoolUIDGenerator) fillPool(ctx context.Context, currPoolSize int) (err error) {
	// 填充号池使用分布式锁，以保证并发安全
	err = u.locker.Lock(ctx, ctx)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			err2 := u.locker.Unlock(ctx)
			if err2 != nil {
				xlog.Error("fillPool unlock", zap.Error(err2))
			}
		} else {
			println("fillPool defer")

			err = u.locker.Unlock(ctx)
			if err != nil {
				println("fillPool 333", err.Error())
			}
		}
	}()

	// 临界区内再次判断是否需要填充号池
	if insufficient, err := u.isPoolInsufficient(ctx); err != nil {
		return errors.Wrap(err, "isPoolInsufficient")
	} else if !insufficient {
		return nil
	}

	// 下面开始填充号池
	maxUnusedUID := uint64(0)
	if currPoolSize > 0 {
		maxUnusedUID, err = u.pool.MaxUnusedUID()
		if err != nil {
			return errors.Wrapf(err, "MaxUnusedUID")
		}
	}

	if maxUnusedUID == 0 {
		// 获取业务中当前已使用的最大uid
		if currMaxUID, err := u.getCurrentMaxUID(); err != nil {
			return errors.Wrap(err, "getCurrentMaxUID")
		} else {
			maxUnusedUID = currMaxUID
		}
	}

	defer func() {
		//size, _ := u.pool.Size()
		//fmt.Printf("555 %v %v %d\n", currPoolSize, u.pushIdsBuffer, size)

		u.pushIdsBuffer = u.pushIdsBuffer[:0] // reset
	}()

	var newId = maxUnusedUID
	// 确保每次把号池填满
	for i := 0; i < u.maxPoolSize-currPoolSize; i++ {
		for {
			newId++
			if skip, err := u.skipFn(newId); err != nil {
				return errors.Wrap(err, "skipFn")
			} else if skip {
				continue
			}
			break
		}
		u.pushIdsBuffer = append(u.pushIdsBuffer, newId)
	}
	return u.pool.Push(u.pushIdsBuffer)
}

type Option func(generator *IncrementalPoolUIDGenerator)

func WithPoolConfig(maxPoolSize int) Option {
	return func(generator *IncrementalPoolUIDGenerator) {
		generator.maxPoolSize = maxPoolSize
	}
}

func WithSkipFunc(skipFn func(uint64) (bool, error)) Option {
	return func(generator *IncrementalPoolUIDGenerator) {
		generator.skipFn = skipFn
	}
}

func NewUidGenerator(locker DistributeLock, pool QueuedPool, getCurrMaxUID func() (uint64, error), opts ...Option) UIDGenerator {
	g := &IncrementalPoolUIDGenerator{locker: locker, pool: pool, getCurrentMaxUID: getCurrMaxUID}
	for _, opt := range opts {
		opt(g)
	}
	if g.maxPoolSize == 0 {
		g.maxPoolSize = 100
	}
	// 最小容量可调，必须
	if g.maxPoolSize < 10 {
		panic("maxPoolSize must >= 10")
	}
	g.poolSizeThreshold = g.maxPoolSize / 3
	return g
}
