package main

import (
	"microsvc/bizcomm/bigmodel/tongyi"
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
	"microsvc/protocol/svc/crontaskpb"
	"microsvc/service/crontask/crontask"
	deploy2 "microsvc/service/crontask/deploy"
	"microsvc/service/crontask/handler"
	"microsvc/util/graceful"

	"github.com/stripe/stripe-go/v81"

	"google.golang.org/grpc"
)

func main() {
	deploy.InitCfg(enums.SvcCrontask, deploy2.CrontaskConf)

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

	// 大模型client初始化
	//kimi.Init(deploy2.CrontaskConf.Kimi.APIKey)
	tongyi.Init(deploy2.CrontaskConf.Tongyi.APIKey)

	stripe.Key = deploy.XConf.Stripe.Key

	crontask.MustInit()

	x := xgrpc.New()
	x.Apply(func(s *grpc.Server) {
		crontaskpb.RegisterCrontaskExtServer(s, handler.Ctrl)
		crontaskpb.RegisterCrontaskIntServer(s, handler.IntCtrl)
	})

	x.Start(deploy.XConf)
	sd.MustRegister(deploy.XConf)

	graceful.Run()
}
