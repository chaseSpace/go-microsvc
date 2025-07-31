package orm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"microsvc/deploy"
	"microsvc/infra/cache"
	"microsvc/pkg/xlog"
	"microsvc/util/graceful"
	"os"
	"time"

	"github.com/go-gorm/caches/v4"
	"github.com/redis/go-redis/v9"

	errors2 "github.com/pkg/errors"

	"github.com/k0kubun/pp/v3"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var instMap = make(map[deploy.DBname]*gorm.DB)

func Init(must bool) func(*deploy.XConfig, func(must bool, err error)) {
	graceful.AddStopFunc(Stop)

	return func(cc *deploy.XConfig, finished func(must bool, err error)) {
		gconf := &gorm.Config{
			Logger:          ormLogger(),
			CreateBatchSize: 100, // 批量插入时，分批进行
			//TranslateError:  true, // 将数据库特有的错误转换为GORM的通用错误
		}
		var db *gorm.DB
		var err error
		if len(cc.Mysql) == 0 {
			fmt.Println("### there is no mysql config found")
		} else {
			for alias, v := range cc.Mysql {
				db, err = gorm.Open(mysql.Open(v.Dsn()), gconf)
				if err != nil {
					_, _ = pp.Printf("#### infra.mysql init failed: alias:%s err:%s --dsn: %s\n", alias, err.Error(), v.Dsn())
					break
				}
				instMap[v.DBname] = db
			}
		}

		if err != nil {
			finished(must, err)
			return
		}

		fmt.Println("#### infra.mysql init success")
		sqlDB, _ := db.DB()

		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(20)

		// 检查业务需要的db是否在配置中存在
		err = setupSvcDB()

		finished(must, errors2.Wrap(err, "cache.Init"))
	}
}

func ormLogger() logger.Interface {
	return logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
		SlowThreshold:             100 * time.Millisecond,
		LogLevel:                  logger.Info,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})
}

type MysqlObj struct {
	name deploy.DBname
	*gorm.DB
	// 你可能希望在对象中包含一些其他自定义成员，在这里添加
}

func (m *MysqlObj) IsInvalid() bool {
	return m.DB == nil
}

func (m *MysqlObj) Stop() {
	db, _ := m.DB.DB()
	err := db.Close()
	if err != nil {
		xlog.Error("orm.Stop() failed", zap.Error(err))
	}
}

func (m *MysqlObj) String() string {
	return fmt.Sprintf("mysqlObj{name:%s, instExists:%v}", m.name, m.DB != nil)
}

var servicesDB []*MysqlObj

func setupSvcDB() error {
	for _, obj := range servicesDB {
		obj.DB = instMap[obj.name]
		if obj.IsInvalid() {
			return fmt.Errorf("orm.MysqlObj is invalid, %s", obj)
		}
	}
	return nil
}

func Stop() {
	for _, gdb := range instMap {
		db, _ := gdb.DB()
		_ = db.Close()
	}
	if len(instMap) > 0 {
		xlog.Debug("orm-mysql: resource released...")
	}
}

func NewMysqlObj(dbname deploy.DBname) *MysqlObj {
	return &MysqlObj{name: dbname}
}

func Setup(obj ...*MysqlObj) {
	for _, o := range obj {
		if o.name == "" {
			panic("orm.Setup: need name")
		}
	}
	servicesDB = append(servicesDB, obj...)
}

func UseCachePlugin(obj *MysqlObj, redisObj *cache.RedisObj) {
	cachesPlugin := &caches.Caches{Conf: &caches.Config{
		Cacher: &redisCacheT{rdb: redisObj.Client},
	}}
	err := obj.Use(cachesPlugin)
	if err != nil {
		panic(err)
	}
}

// ----------------- Cache -----------------------

type redisCacheT struct {
	rdb *redis.Client
}

func (c *redisCacheT) Get(ctx context.Context, key string, q *caches.Query[any]) (*caches.Query[any], error) {
	res, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if err := q.Unmarshal([]byte(res)); err != nil {
		return nil, err
	}

	return q, nil
}

func (c *redisCacheT) Store(ctx context.Context, key string, val *caches.Query[any]) error {
	res, err := val.Marshal()
	if err != nil {
		return err
	}

	err = c.rdb.Set(ctx, key, res, time.Minute*10).Err() // Set proper cache time
	return err
}

func (c *redisCacheT) Invalidate(ctx context.Context) error {
	var (
		cursor uint64
		keys   []string
	)
	for {
		var (
			k   []string
			err error
		)
		k, cursor, err = c.rdb.Scan(ctx, cursor, fmt.Sprintf("%s*", caches.IdentifierPrefix), 0).Result()
		if err != nil {
			return err
		}
		keys = append(keys, k...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		if _, err := c.rdb.Del(ctx, keys...).Result(); err != nil {
			return err
		}
	}
	return nil
}
