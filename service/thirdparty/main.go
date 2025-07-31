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
	"microsvc/protocol/svc/thirdpartypb"
	deploy2 "microsvc/service/thirdparty/deploy"
	"microsvc/service/thirdparty/handler"
	"microsvc/service/thirdparty/logic_oss"
	"microsvc/service/thirdparty/logic_review"
	"microsvc/service/thirdparty/logic_sms"
	"microsvc/util/graceful"

	"google.golang.org/grpc"
)

// thirdparty 服务
// 主要对接并提供各种第三方服务，如短信、oss、审核等功能

func main() {
	deploy.InitCfg(enums.SvcThirdparty, deploy2.ThirdpartyConf)

	pkg.Setup(
		xlog.Init,
	)

	graceful.SetupSignal()
	defer graceful.OnExit()

	infra.Setup(
		cache.InitRedis(true),
		orm.Init(true),
		sd.Init(true),
		svccli.Init(true),
		xmq.Init(true),
	)

	// 第三方组件初始化
	logic_review.MustInit(deploy2.ThirdpartyConf)
	logic_sms.MustInit(deploy2.ThirdpartyConf)
	logic_oss.MustInit(deploy2.ThirdpartyConf)

	x := xgrpc.New()
	x.Apply(func(s *grpc.Server) {
		thirdpartypb.RegisterThirdpartyExtServer(s, handler.Ctrl)
		thirdpartypb.RegisterThirdpartyIntServer(s, handler.IntCtrl)
	})

	x.Start(deploy.XConf)
	sd.MustRegister(deploy.XConf)

	graceful.Run()
}
