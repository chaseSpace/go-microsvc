package ulock

import (
	"context"
	"microsvc/util/urand"
	"microsvc/util/utilcommon"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type DistributedLock interface {
	Lock(acquireCtx, taskCtx context.Context) error
	LockSimple(taskCtx context.Context) error
	LockNoRetry(acquireCtx, taskCtx context.Context) error
	Unlock(context.Context) error
	IsLocked(ctx context.Context, byMe bool) (bool, error)
}

var _ DistributedLock = (*DistributedLockInRedis)(nil)

type DistributedLockInRedis struct {
	client                  *redis.Client
	uniqueId, lockedRandVal string
	isRenewing              atomic.Bool
}

const dLockKeyPrefix = "dLockKeyPrefix:"
const defaultLockExpiry = time.Second * 3

// DEBUG Lua脚本：https://www.51cto.com/article/778877.html
const unlockLuaScript = `
local key = KEYS[1] -- 从KEYS数组获取key
local expectedValue = ARGV[1] -- 从ARGV数组获取预期的值

local value = redis.call("get", key)

-- key 不存在
if not value then 
    return 0
-- value匹配 则删除
elseif value == expectedValue then
	redis.call("del", key)
	return 2
else -- value不匹配 返回1
    return 1
end
`

var (
	ErrLockFailed                  = errors.New("redis lock failed")
	ErrUnlockFailed                = errors.New("redis unlock failed")
	ErrUnlockFailedOnNotLocked     = errors.Wrap(ErrUnlockFailed, "not locked")
	ErrUnlockFailedOnLockedByOther = errors.Wrap(ErrUnlockFailed, "locked by other")
)

func NewDLock(uniqueId string, cli *redis.Client) DistributedLock {
	if len(uniqueId) < 5 {
		panic("uniqueId length must >= 5")
	}
	return &DistributedLockInRedis{
		client:   cli,
		uniqueId: dLockKeyPrefix + uniqueId,
	}
}

// renewInLoop 自动续约锁
func (d *DistributedLockInRedis) renewInLoop(taskCtx context.Context) {
	if !d.isRenewing.Swap(true) {
		return
	}
	go func() {
		for {
			select {
			case <-taskCtx.Done():
				println("renewInLoop task done")
				ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
				err := d.Unlock(ctx) // 退出时需要再执行一次解锁操作，因为外部解锁后这里可能又多续了一次导致延迟释放
				if err != nil && !errors.Is(err, ErrUnlockFailedOnNotLocked) {
					utilcommon.PrintlnStackMsg("renewInLoop Unlock err:%v  key:%s", err, d.uniqueId)
				}
				cancel()
				return
			default:
				ret := d.client.SetNX(taskCtx, d.uniqueId, d.lockedRandVal, defaultLockExpiry)
				if err := ret.Err(); err != nil && !errors.Is(err, context.Canceled) {
					utilcommon.PrintlnStackMsg("renewInLoop SetNX err:%v\n key:%s", err, d.uniqueId)
					return
				}
				time.Sleep(time.Millisecond * 500) // 理论上，这个时间小于 defaultLockExpiry，所以续约永远有效
			}
		}
	}()
	return
}

// LockSimple 作用同Lock，使用简单，推荐
func (d *DistributedLockInRedis) LockSimple(taskCtx context.Context) error {
	lockCtx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
	return d.Lock(lockCtx, taskCtx)
}

// Lock 获取锁
// - acquireCtx: 获取锁的超时上下文
// - taskCtx: 续约锁的超时上下文（与任务的生命周期一致）
func (d *DistributedLockInRedis) Lock(acquireCtx, taskCtx context.Context) (err error) {
	randVal := urand.Strings(6)

	var asyncDone = make(chan struct{})
	var locked bool
	for {
		go func() {
			defer func() { asyncDone <- struct{}{} }()
			ret := d.client.SetNX(acquireCtx, d.uniqueId, randVal, defaultLockExpiry)
			if ret.Err() != nil {
				err = errors.Wrap(ErrLockFailed, ret.Err().Error())
				return
			}
			if !ret.Val() {
				return
			}
			d.lockedRandVal = randVal
			locked = true
		}()

		select {
		case <-acquireCtx.Done():
			return acquireCtx.Err()
		case <-asyncDone:
			if err != nil {
				return err
			}
			if locked {
				go d.renewInLoop(taskCtx)
				return
			}
			// 继续尝试
			time.Sleep(time.Millisecond * time.Duration(urand.Int31n(10, 10)))
		}
	}
}

// LockNoRetry 只尝试一次，不多次抢占
func (d *DistributedLockInRedis) LockNoRetry(acquireCtx, taskCtx context.Context) (err error) {
	randVal := urand.Strings(6)
	ret := d.client.SetNX(acquireCtx, d.uniqueId, randVal, defaultLockExpiry)
	if ret.Err() != nil {
		err = errors.Wrap(ErrLockFailed, ret.Err().Error())
		return
	}
	if !ret.Val() {
		return ErrLockFailed
	}
	d.lockedRandVal = randVal

	go d.renewInLoop(taskCtx)

	return
}

func (d *DistributedLockInRedis) Unlock(ctx context.Context) error {
	if d.lockedRandVal == "" {
		return ErrUnlockFailedOnNotLocked
	}
	ret := d.client.Eval(ctx, unlockLuaScript, []string{d.uniqueId}, d.lockedRandVal)
	i, err := ret.Int64()
	if err != nil {
		return errors.Wrap(ErrUnlockFailed, err.Error())
	}
	if i == 0 {
		return ErrUnlockFailedOnNotLocked
	}
	if i == 1 {
		return ErrUnlockFailedOnLockedByOther
	}
	return nil
}

// IsLocked 是否锁住（无论是不是自己）
// - byMe: 是否被调用者锁住
func (d *DistributedLockInRedis) IsLocked(ctx context.Context, byMe bool) (bool, error) {
	ret := d.client.Get(ctx, d.uniqueId)
	if ret.Err() != nil && !errors.Is(ret.Err(), redis.Nil) {
		return false, ret.Err()
	}
	if errors.Is(ret.Err(), redis.Nil) {
		return false, nil
	}
	if byMe {
		return ret.Val() == d.lockedRandVal, nil
	}
	return ret.Val() != "", nil
}

func QuickExec(ctx context.Context, cli *redis.Client, exec func(ctx context.Context) error) error {
	v := NewDLock("quick_exec_"+urand.Strings(6), cli)
	err := v.LockSimple(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = v.Unlock(ctx) }()
	return exec(ctx)
}
