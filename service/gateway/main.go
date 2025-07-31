package main

import (
	"microsvc/deploy"
	"microsvc/enums"
	"microsvc/infra"
	"microsvc/infra/cache"
	"microsvc/infra/orm"
	"microsvc/infra/sd"
	"microsvc/infra/svccli"
	_ "microsvc/infra/xgrpc/protocodec"
	"microsvc/infra/xhttp"
	"microsvc/infra/xmq"
	"microsvc/pkg"
	"microsvc/pkg/xlog"
	deploy2 "microsvc/service/gateway/deploy"
	"microsvc/service/gateway/handler"
	"microsvc/service/gateway/logic/logic_wsmanage"
	"microsvc/util/graceful"
)

func main() {
	deploy.InitCfg(enums.SvcGateway, deploy2.GatewayConf)

	pkg.Setup(
		xlog.Init,
	)

	graceful.SetupSignal()
	defer graceful.OnExit()

	infra.Setup(
		sd.InitSimpleSdServer(), // It always starts sd-server on gateway!
		cache.InitRedis(true),
		orm.Init(true),
		svccli.Init(true),
		xmq.Init(true),
	)

	// ws 长连接，用于【服务器推送、客户端上报】场景
	logic_wsmanage.Init()

	ctrl := new(handler.GatewayCtrl)
	server := xhttp.New(deploy2.GatewayConf.HttpPort, ctrl.Handler)

	graceful.AddStopFunc(logic_wsmanage.WsManager.OnClose)
	graceful.Register(server.Start)
	graceful.Run()
}
