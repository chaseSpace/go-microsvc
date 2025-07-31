package admin

import (
	"microsvc/infra/cache"
	"microsvc/infra/orm"
	"microsvc/model"
)

var (
	Q      = orm.NewMysqlObj(model.MysqlDB)
	QLog   = orm.NewMysqlObj(model.MysqlDBLog)
	QAdmin = orm.NewMysqlObj(model.MysqlDBAdmin)
)

var (
	RAdmin = cache.NewRedisObj(model.RedisDBAdmin)
)

func init() {
	// 此函数会在main函数执行前向orm注入服务需要使用的DB对象
	orm.Setup(Q, QLog, QAdmin)
	cache.Setup(RAdmin)
}
