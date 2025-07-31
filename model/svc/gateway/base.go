package gateway

import (
	"microsvc/infra/cache"
	"microsvc/infra/orm"
	"microsvc/model"
)

var (
	Q = orm.NewMysqlObj(model.MysqlDBGateway)
	R = cache.NewRedisObj(model.RedisDBGateway)
)

func init() {
	// 此函数会在main函数执行前向orm注入服务需要使用的DB对象
	orm.Setup(Q)
	cache.Setup(R)
}
