package mqnats

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"microsvc/consts"
	"microsvc/deploy"
	"microsvc/infra/xmq/define"
)

var _ define.MqProviderAPI = (*MqNats)(nil)

// MqNats NATS 是一个轻量级、高性能的消息系统，核心设计理念是简单、可靠、可扩展。
// 它支持发布/订阅、请求/响应、队列等多种消息模式，适用于微服务、IoT、边缘计算等场景
// NATS VS Kafka，如果要持久化，NATS性能不如Kafka，否则远高于Kafka
// Github：https://github.com/nats-io/nats-server
// Docker启动：docker run -p 4222:4222 --name nats -tid nats:latest
type MqNats struct {
	cli *nats.Conn
}

func New() *MqNats {
	return &MqNats{}
}

func (m *MqNats) Name() string {
	return "nats"
}

func (m *MqNats) Init(cc *deploy.MqConfig) (err error) {
	m.cli, err = nats.Connect(cc.NATS.URL)
	return err
}

func (m *MqNats) Stop() error {
	if m.cli == nil {
		return nil
	}
	return m.cli.Drain()
}

func (m *MqNats) Produce(ctx context.Context, topic consts.Topic, msg []byte) error {
	return m.cli.Publish(topic.String(), msg)
}

func (m *MqNats) Consume(topic consts.Topic, handler func(ctx context.Context, msg define.MsgRaw), arg ...define.ConsumeExtraArg) {
	_, err := m.cli.Subscribe(topic.String(), func(m *nats.Msg) {
		handler(context.TODO(), &_msgRaw{msg: m})
	})
	if err != nil {
		panic(fmt.Sprintf("MqNats [Consume] Subscribe failed: %v", err))
	}
}

type _msgRaw struct {
	msg *nats.Msg
}

var _ define.MsgRaw = (*_msgRaw)(nil)

func (m _msgRaw) Bytes() []byte {
	return m.msg.Data
}

func (m _msgRaw) Ack() error {
	if m.msg.Reply == "" { // 一般只有在 JetStream 模式下这里才会有值，才支持ack
		return nil
	}
	return m.msg.Ack()
}
