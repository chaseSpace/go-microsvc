package micro_svc

import (
	"microsvc/infra/cache"
	"microsvc/infra/orm"
	"microsvc/model"
)

var (
	Q = orm.NewMysqlObj(model.MysqlDBMicroSvc)
	R = cache.NewRedisObj(model.RedisDBMicroSvc)
)

func init() {
	// 此函数会在main函数执行前向orm注入服务需要使用的DB对象
	orm.Setup(Q)
	cache.Setup(R)
}
