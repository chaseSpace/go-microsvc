package mixin

// mixin 不对应某个服务，可以被任何服务使用，仅存放公共表

//var (
//	Q    = orm.NewMysqlObj(model.MysqlDB)
//	QLog = orm.NewMysqlObj(model.MysqlDBLog)
//)
//
//var (
//	R = cache.NewRedisObj(model.RedisDB)
//)
//
//func init() {
//	// 此函数会在main函数执行前向orm注入服务需要使用的DB对象
//	orm.Setup(Q, QLog)
//	cache.Setup(R)
//}
