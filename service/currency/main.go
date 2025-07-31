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
	"microsvc/pkg"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/currencypb"
	deploy2 "microsvc/service/currency/deploy"
	"microsvc/service/currency/handler"
	"microsvc/util/graceful"

	"google.golang.org/grpc"
)

func main() {
	deploy.InitCfg(enums.SvcCurrency, deploy2.CurrencyConf)

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
	)

	x := xgrpc.New()
	x.Apply(func(s *grpc.Server) {
		currencypb.RegisterCurrencyExtServer(s, handler.Ctrl)
		currencypb.RegisterCurrencyIntServer(s, handler.IntCtrl)
	})

	x.Start(deploy.XConf)
	sd.MustRegister(deploy.XConf)

	graceful.Run()
}
