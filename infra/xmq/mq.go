package xmq

import (
	"context"
	"fmt"
	"github.com/blinkbean/dingtalk"
	"microsvc/consts"
	"microsvc/deploy"
	"microsvc/enums"
	"microsvc/infra/xmq/define"
	"microsvc/infra/xmq/mqkafka"
	"microsvc/infra/xmq/mqredis"
	"microsvc/model/svc/mqconsumer"
	"microsvc/pkg/xlog"
	"microsvc/pkg/xnotify"
	"microsvc/util"
	"microsvc/util/db"
	"microsvc/util/graceful"
	"microsvc/util/ujson"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var provider define.MqProviderAPI

func Init(must bool, impl ...string) func(*deploy.XConfig, func(must bool, err error)) {
	graceful.AddStopFunc(Stop)

	return func(cc *deploy.XConfig, finished func(must bool, err error)) {
		var err error

		_impl := "redis"
		if len(impl) > 0 {
			_impl = impl[0]
		}
		switch _impl {
		case "redis":
			provider = mqredis.New() // 这里指定 mq 实现
		case "kafka":
			provider = mqkafka.New()
		default:
			panic("mq impl not support")
		}

		err = provider.Init(&cc.MqConfig)
		if err == nil {
			fmt.Printf("#### infra.mq[%s] init success\n", _impl)
		}
		finished(must, err)
	}
}

func Stop() {
	if provider == nil {
		return
	}
	err := provider.Stop()
	if err != nil {
		xlog.Error(provider.Name()+" [Stop] failed", zap.Error(err))
	}
}

// Produce 几乎任何服务都可以生产消息（除了 mqconsumer 自身）
func Produce(topic consts.Topic, msg define.MqMsgAPI) {
	msg.SetMsTimestamp(time.Now().UnixMilli())
	if msg.GetUniqueID() == "" {
		msg.SetUniqueID(util.NewKsuid())
	}
	buf, _ := ujson.Marshal(msg)

	var (
		ctx = context.TODO()
		err error
	)

	// 大部分消息入列前需要入库存档（以便重新消费MQ宕机期间发布的消息）
	if msg.NeedArchive() {
		row := &mqconsumer.MqLog{
			TopicUniqueId: msg.GetUniqueID(),
			Topic:         topic.String(),
			Data:          string(buf), // 必须是JSON格式，否则DB报错
			CreatedAt:     time.Now(),
		}
		row.SetSuffix(row.CreatedAt.Format("200601"))
		err = db.NewTableHelper(mqconsumer.QLog.WithContext(ctx), row.DDLSql()).AutoCreateTable(func(tx *gorm.DB) error {
			return tx.Table(row.TableName()).Create(row).Error
		})
		if err != nil {
			xlog.Error(provider.Name()+" [Produce] Insert mq log err", zap.Error(err), zap.String("topic", topic.String()), zap.Any("msg", msg))
			return
		}
	}

	err = provider.Produce(ctx, topic, buf)
	if err != nil {
		xlog.Error(provider.Name()+" [Produce] msg failed", zap.Error(err), zap.Any("msg", msg))
		return
	}
	xlog.Debug(provider.Name()+" [Produce] msg success", zap.String("topic", topic.String()), zap.Any("msg", msg))
}

// Consume 只有 mqconsumer 服务才能消费消息
func Consume[T define.MqMsgAPI](topic consts.Topic, handler func(context.Context, T) error, arg ...define.ConsumeExtraArg) {
	if deploy.XConf.Svc != enums.SvcMqConsumer && deploy.XConf.Svc != enums.SvcGateway {
		panic("only service->[mqconsumer/gateway] can consume msg")
	}
	provider.Consume(topic, func(ctx context.Context, msg []byte) {
		var t T
		err := ujson.Unmarshal(msg, &t)
		if err != nil {
			xlog.Error(provider.Name()+" [Mq-Consume] Unmarshal failed...", zap.Error(err), zap.String("topic", topic.String()), zap.ByteString("msg", msg))
			return
		}
		safeHandle(topic, t, func() {
			err = handler(ctx, t)
			if err != nil {
				xlog.Error(provider.Name()+" [Mq-Consume] handler failed...", zap.Error(err), zap.String("topic", topic.String()), zap.Any("msg", msg))
				return
			}
			xlog.Debug(provider.Name()+" [Mq-Consume] handler success...", zap.String("topic", topic.String()), zap.String("msg", string(msg)))
		})
	}, arg...)
}

func safeHandle(topic consts.Topic, msg any, consume func()) (safe bool) {
	util.Protect(func() { consume() }, func(exception interface{}) {
		title := fmt.Sprintf("MQ消费异常: %s", topic.String())
		// 接入告警
		xnotify.NotifyDingtalkMD(context.TODO(), xnotify.SceneMqException,
			xnotify.NotifyArgs{
				Title: title,
				//Content:       "",
				MarkdownLines: []*xnotify.MdLine{
					{
						ContentWithPlaceHolder: fmt.Sprintf("Panic: %v", exception),
						MarkType:               dingtalk.N,
					},
				},
			},
		)
		xlog.Error("Consumer task PANIC!!!", zap.String("topic", topic.String()),
			zap.Any("exception", exception), zap.Any("msg", msg))
	})
	return
}
