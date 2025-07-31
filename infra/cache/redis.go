package cache

import (
	"context"
	"errors"
	"fmt"
	"microsvc/deploy"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"microsvc/util/graceful"
	"time"

	errors2 "github.com/pkg/errors"

	"github.com/k0kubun/pp/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// var instMap = make(map[deploy.DBname]*redis.Client)
var instMap = make(map[deploy.DBname]RedisObj)

func InitRedis(must bool) func(*deploy.XConfig, func(must bool, err error)) {
	graceful.AddStopFunc(Stop)

	return func(cc *deploy.XConfig, finished func(must bool, err error)) {
		var err error
		for alias, v := range cc.Redis {
			rdb := redis.NewClient(&redis.Options{
				Addr:       v.Addr,
				Password:   v.Password,
				DB:         v.DB,
				MaxRetries: 2,
			})
			util.RunTaskWithCtxTimeout(time.Second, func(ctx context.Context) {
				err = rdb.Ping(ctx).Err()
			})
			if err != nil {
				_, _ = pp.Printf("#### infra.redis init failed: alias:%s err:%s --config:%+v\n", alias, err.Error(), v)
				break
			}
			instMap[v.DBname] = RedisObj{
				name:   v.DBname,
				Client: rdb,
				config: v,
			}
		}

		if err == nil {
			fmt.Println("#### infra.redis init success")
			if err = setupSvcDB(); err != nil {
				panic(err)
			}
		}

		finished(must, errors2.Wrap(err, "cache.InitRedis"))
	}
}

type RedisObj struct {
	name deploy.DBname
	*redis.Client
	config *deploy.Redis
	// 这里可以添加一些其他自定义成员
}

func (m *RedisObj) IsInvalid() bool {
	return m.Client == nil
}

func (m *RedisObj) Stop() {
	err := m.Client.Close()
	if err != nil {
		xlog.Error("orm.Stop() failed", zap.Error(err))
	}
}

func (m *RedisObj) String() string {
	return fmt.Sprintf("RedisObj{name:%s, instExists:%v}", m.name, m.Client != nil)
}

func (m *RedisObj) Config() *deploy.Redis {
	return m.config
}

var servicesDB []*RedisObj

func setupSvcDB() error {
	for _, obj := range servicesDB {
		name := obj.name
		*obj = instMap[name]
		if obj.IsInvalid() {
			return fmt.Errorf("cache.RedisObj is invalid, name: [%s]", name)
		}
	}
	return nil
}

func Stop() {
	for _, db := range instMap {
		_ = db.Close()
	}
	if len(instMap) > 0 {
		xlog.Debug("cache-redis: resource released...")
	}
}

func NewRedisObj(dbname deploy.DBname) *RedisObj {
	o := &RedisObj{name: dbname}
	return o
}

func Setup(obj ...*RedisObj) {
	for _, o := range obj {
		if o.name == "" {
			panic("cache.Setup: need name")
		}
	}
	servicesDB = append(servicesDB, obj...)
}

func IsRedisErr(err error) bool {
	return err != nil && !errors.Is(err, redis.Nil)
}
