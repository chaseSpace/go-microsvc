package mqconsumer

import (
	"microsvc/infra/cache"
	"microsvc/infra/orm"
	"microsvc/model"
)

var (
	QLog = orm.NewMysqlObj(model.MysqlDBLog)
)

var (
	R = cache.NewRedisObj(model.RedisDB)
)

func init() {
	// 此函数会在main函数执行前向orm注入服务需要使用的DB对象
	orm.Setup(QLog)
	cache.Setup(R)
}
