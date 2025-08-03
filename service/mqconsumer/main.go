package main

import (
	"google.golang.org/grpc"
	"microsvc/bizcomm/mq"
	"microsvc/consts"
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
	"microsvc/protocol/svc/mqconsumerpb"
	"microsvc/service/mqconsumer/consumer"
	deploy2 "microsvc/service/mqconsumer/deploy"
	"microsvc/service/mqconsumer/handler"
	"microsvc/util/graceful"
	"time"
)

func main() {
	deploy.InitCfg(enums.SvcMqConsumer, deploy2.MqConsumerConf)

	pkg.Setup(
		xlog.Init,
	)

	graceful.SetupSignal()
	defer graceful.OnExit()

	infra.Setup(
		cache.InitRedis(true),
		orm.Init(true),
		xmq.Init(true, "nats"),
		sd.Init(true),
		svccli.Init(true),
	)

	// 启动所有消费线程
	consumer.Init()

	x := xgrpc.New()
	x.Apply(func(s *grpc.Server) {
		mqconsumerpb.RegisterMqConsumerExtServer(s, handler.Ctrl)
		mqconsumerpb.RegisterMqConsumerIntServer(s, handler.IntCtrl)
	})

	x.Start(deploy.XConf)
	sd.MustRegister(deploy.XConf)

	go func() {
		time.Sleep(time.Second)

		for i := 0; i < 2; i++ {
			xmq.Produce(consts.TopicSignIn, mq.NewMsgSignIn(
				&mq.SignInBody{
					UID: int64(i),
				}),
			)
		}
	}()
	graceful.Run()
}
