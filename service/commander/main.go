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
	"microsvc/infra/xmq"
	"microsvc/pkg"
	"microsvc/pkg/xlog"
	deploy2 "microsvc/service/commander/deploy"
	"microsvc/service/commander/handler"
	"microsvc/util/graceful"

	"github.com/spf13/cobra"
)

func main() {
	deploy.InitCfg(enums.SvcCommander, deploy2.CommanderConf)

	pkg.Setup(
		xlog.Init,
	)

	defer graceful.OnExit()

	infra.Setup(
		cache.InitRedis(true),
		orm.Init(true),
		sd.Init(true),
		svccli.Init(true),
		xmq.Init(true),
	)

	rootCmd := &cobra.Command{Use: enums.SvcCommander.Name()}
	handler.MustInit(rootCmd)

	// THIS IS NOT A REGULAR SERVICE
	// 不是一个常规的微服务，不需要启动grpc服务
	// 启动示例：go run .
}
