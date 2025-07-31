package define

import (
	"context"
	"microsvc/consts"
	"microsvc/deploy"
	"time"
)

type ConsumeExtraArg struct {
	ConsumeGroupId consts.ConsumerGroup // kafka 需要
}

type MqProviderAPI interface {
	Name() string
	Init(cc *deploy.MqConfig) error
	Stop() error
	Produce(ctx context.Context, topic consts.Topic, msg []byte) error
	Consume(topic consts.Topic, handler func(ctx context.Context, msg []byte), arg ...ConsumeExtraArg)
}

type MqMsgAPI interface {
	// GetUniqueID 任何消息都需要设置唯一ID
	// - 严谨性不高的消息使用 util.NewKsuid()
	// - 像订单号这种需要严格唯一的自定义生成规则
	GetUniqueID() string
	SetUniqueID(string)
	SetMsTimestamp(int64)
	GetMsgTime() time.Time
	NeedArchive() bool
}
