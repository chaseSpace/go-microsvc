package ulock

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func initRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:       "127.0.0.1:6379",
		Password:   "123",
		DB:         0,
		MaxRetries: 2,
	})
	err := rdb.Ping(context.TODO()).Err()
	if err != nil {
		panic(err)
	}
	return rdb
}

func timeoutCtx(to time.Duration) context.Context {
	c, _ := context.WithTimeout(context.Background(), to)
	return c
}

func TestNewDLockConcurrency(t *testing.T) {
	cli := initRedis()
	defer cli.Close()

	k := NewDLock("hello_lock", cli)

	var imap = make(map[int]bool)

	count := 5 // 不要设置过高，不然任务会因为冲突太多运行过久，打开下面的print进行观察
	v := 0

	var x sync.WaitGroup
	for i := 0; i < count; i++ {

		x.Add(1)
		go func() {
			taskCtx, cancel := context.WithCancel(context.Background())

			defer x.Done()
			defer cancel()

			//println(111111)
			err := k.Lock(context.TODO(), taskCtx)
			if !assert.Nil(t, err) {
				t.FailNow()
			}
			//println(222222)

			v++
			imap[v] = true
			err = k.Unlock(context.TODO())
			if !assert.Nil(t, err) {
				t.FailNow()
			}
		}()
	}

	x.Wait()
	assert.Equal(t, count, len(imap))
}

func TestNewDLockWithLockAgainAndNotUnlock(t *testing.T) {
	cli := initRedis()
	defer cli.Close()

	k := NewDLock("hello_lock", cli)
	taskCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := k.Lock(context.TODO(), taskCtx)
	assert.Nil(t, err)

	// try to lock again
	err = k.Lock(timeoutCtx(time.Millisecond*100), taskCtx)
	assert.Equal(t, context.DeadlineExceeded, err)

	err = k.Unlock(context.TODO())
	assert.Nil(t, err)

	// lock successfully after unlock
	err = k.Lock(context.TODO(), taskCtx)
	assert.Nil(t, err)
	err = k.Unlock(context.TODO())
	assert.Nil(t, err)
}

func TestNewDLockWithUnlockWithoutLock(t *testing.T) {
	cli := initRedis()
	defer cli.Close()

	k := NewDLock("hello_lock", cli)

	err := k.Unlock(context.TODO())
	assert.Equal(t, ErrUnlockFailedOnNotLocked, err)
}

func TestNewDLockWithIsLocked(t *testing.T) {
	cli := initRedis()
	defer cli.Close()

	k := NewDLock("hello_lock", cli)
	taskCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// IsLocked returns false if not locked
	locked, err := k.IsLocked(context.TODO(), false)
	assert.Nil(t, err)
	assert.False(t, locked)

	// then we get lock
	err = k.Lock(context.TODO(), taskCtx)
	assert.Nil(t, err)

	// now, IsLocked returns true
	locked, err = k.IsLocked(context.TODO(), true)
	assert.Nil(t, err)
	assert.True(t, locked)

	// 被别人锁住的情况
	k = NewDLock("hello_lock", cli)
	locked, err = k.IsLocked(context.TODO(), true)
	assert.Nil(t, err)
	assert.False(t, locked)
}

func TestNewDLockLockNoPreempt(t *testing.T) {
	cli := initRedis()
	defer cli.Close()

	var err error
	k := NewDLock("hello_lock", cli)
	err = k.LockNoRetry(context.TODO(), context.TODO())
	assert.Nil(t, err)

	err = k.LockNoRetry(context.TODO(), context.TODO())
	assert.Equal(t, ErrLockFailed, err)

	_ = k.Unlock(context.TODO())
}
