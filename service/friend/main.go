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
	"microsvc/protocol/svc/friendpb"
	"microsvc/service/friend/base"
	deploy2 "microsvc/service/friend/deploy"
	"microsvc/service/friend/handler"
	"microsvc/util/graceful"

	"google.golang.org/grpc"
)

func main() {
	deploy.InitCfg(enums.SvcFriend, deploy2.FriendConf)

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

	base.MustInit()

	x := xgrpc.New()
	x.Apply(func(s *grpc.Server) {
		friendpb.RegisterFriendExtServer(s, handler.Ctrl)
		friendpb.RegisterFriendIntServer(s, handler.IntCtrl)
	})

	x.Start(deploy.XConf)
	sd.MustRegister(deploy.XConf)

	graceful.Run()
}
