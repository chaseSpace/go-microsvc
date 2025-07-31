package main

import (
	"microsvc/deploy"
	"microsvc/enums"
	"microsvc/infra"
	"microsvc/infra/cache"
	"microsvc/infra/orm"
	"microsvc/infra/sd"
	"microsvc/infra/svccli"
	"microsvc/infra/xgrpc"
	_ "microsvc/infra/xgrpc/protocodec"
	"microsvc/infra/xmq"
	"microsvc/pkg"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/user/base"
	deploy2 "microsvc/service/user/deploy"
	"microsvc/service/user/handler"
	"microsvc/util/graceful"

	"google.golang.org/grpc"
)

func main() {
	// 初始化部署需要config
	deploy.InitCfg(enums.SvcUser, deploy2.UserConf)

	// 初始化服务用到的基础组件（封装于pkg目录下），如log, kafka等
	pkg.Setup(
		xlog.Init,
	)

	graceful.SetupSignal()
	defer graceful.OnExit()

	// 初始化几乎每个服务都需要的infra组件，must参数指定是否必须初始化成功，若must=true且err非空则panic
	// - 注意顺序
	infra.Setup(
		cache.InitRedis(true),
		orm.Init(true),
		sd.Init(true),
		svccli.Init(true),
		xmq.Init(false),
	)

	base.MustInit()

	x := xgrpc.New() // New一个封装好的grpc对象
	x.Apply(func(s *grpc.Server) {
		// 注册外部和内部的rpc接口对象
		userpb.RegisterUserExtServer(s, handler.Ctrl)
		userpb.RegisterUserIntServer(s, handler.IntCtrl)
	})

	x.Start(deploy.XConf)

	sd.MustRegister(deploy.XConf)

	graceful.Run()
}
