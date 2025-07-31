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
	"microsvc/protocol/svc/giftpb"
	deploy2 "microsvc/service/gift/deploy"
	"microsvc/service/gift/handler"
	"microsvc/util/graceful"

	"google.golang.org/grpc"
)

func main() {
	deploy.InitCfg(enums.SvcGift, deploy2.GiftConf)

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
		giftpb.RegisterGiftExtServer(s, handler.Ctrl)
		giftpb.RegisterGiftIntServer(s, handler.IntCtrl)
	})

	x.Start(deploy.XConf)
	sd.MustRegister(deploy.XConf)

	graceful.Run()
}
