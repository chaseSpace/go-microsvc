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
	"microsvc/protocol/svc/templatepb"
	deploy2 "microsvc/service/template/deploy"
	"microsvc/service/template/handler"
	"microsvc/util/graceful"

	"google.golang.org/grpc"
)

func main() {
	deploy.InitCfg(enums.SvcTemplate, deploy2.TemplateConf)

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

	x := xgrpc.New()
	x.Apply(func(s *grpc.Server) {
		templatepb.RegisterTemplateExtServer(s, handler.Ctrl)
		templatepb.RegisterTemplateIntServer(s, handler.IntCtrl)
	})

	x.Start(deploy.XConf)
	sd.MustRegister(deploy.XConf)

	graceful.Run()
}
