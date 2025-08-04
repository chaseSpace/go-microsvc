package mqnats

import (
	"context"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
	"microsvc/consts"
	"microsvc/deploy"
	"microsvc/infra/xmq/define"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"sync"
	"time"
)

var _ define.MqProviderAPI = (*MqNats)(nil)
var gCtx, gCancel = context.WithCancel(context.Background())

// MqNats NATS 是一个轻量级、高性能的消息系统，核心设计理念是简单、可靠、可扩展。
// 它支持发布/订阅、请求/响应、队列等多种消息模式，适用于微服务、IoT、边缘计算等场景
// NATS VS Kafka，如果要持久化，NATS性能不如Kafka，否则远高于Kafka
// Github：https://github.com/nats-io/nats-server
// Docker启动：docker run -p 4222:4222 --name nats -tid nats:latest
// **************************
// # 是否持久化
// # 1. 不持久化：Core NATS（默认）
// # --- 每条消息最多投递一次，不持久化、不重试、无 ACK
// # 2. 持久化：JetStream（需 Server 启用）
// # --- 三种 ACK 策略：
// # --- No.1. Explicit; 默认。每条消息必须显式调用 msg.Ack() / Nak() / Term() 等，否则到期重投。
// # --- No.2. All；批确认。收到一批消息后一次性 ACK 整批。
// # --- No.3. None; 无需 ACK。服务器认为只要发出去就算成功，不保证重投。
type MqNats struct {
	//cli       *nats.Conn
	js        jetstream.JetStream
	streamMgr map[string][]string // key: stream, val: subjects
	mu        sync.Mutex
}

func New() *MqNats {
	return &MqNats{streamMgr: map[string][]string{}}
}

func (m *MqNats) Name() string {
	return "nats"
}

func (m *MqNats) Init(cc *deploy.MqConfig) (err error) {
	cli, err := nats.Connect(cc.NATS.URL)
	if err != nil {
		return err
	}
	m.js, err = jetstream.New(cli)
	return err
}

func (m *MqNats) Stop() error {
	if m.js == nil {
		return nil
	}
	gCancel()
	return m.js.Conn().Drain()
}

func (m *MqNats) getStream(topic consts.Topic) (jetstream.Stream, error) {
	streamName, subjectName := topic.NATSStreamSubject()
	stream, err := m.js.CreateOrUpdateStream(context.TODO(), jetstream.StreamConfig{
		Name:     streamName,
		Subjects: m.getSubjectsByStream(streamName, subjectName),
		Storage:  jetstream.MemoryStorage, // 实际上线时建议采用 jetstream.FileStorage
	})
	if err != nil {
		return nil, fmt.Errorf("MqNats [getStream] CreateStream failed: %v", err)
	}
	return stream, nil
}

func (m *MqNats) getConsumer(topic consts.Topic, consumerName string) (jetstream.Consumer, error) {
	stream, err := m.getStream(topic)
	if err != nil {
		return nil, err
	}
	_, subjectName := topic.NATSStreamSubject()
	cons, err := stream.CreateOrUpdateConsumer(context.TODO(), jetstream.ConsumerConfig{
		Durable:       consumerName,                // Durable 是持久化消费者，重启会恢复进度
		FilterSubject: subjectName,                 // 指定订阅的subject
		AckPolicy:     jetstream.AckExplicitPolicy, // 需手动Ack
		AckWait:       time.Second * 30,            // 若这个时间内还未处理完一条消息，则mq server会自动重投，消费者需要自行保证幂等性
		MaxWaiting:    1,                           // 最多等待的未确认消息数
	})
	return cons, err
}

func (m *MqNats) Produce(ctx context.Context, topic consts.Topic, msg []byte) error {
	_, subjectName := topic.NATSStreamSubject()
	_, err := m.js.Publish(ctx, subjectName, msg)
	return err
}

func (m *MqNats) Consume(topic consts.Topic, handler func(ctx context.Context, msg define.MsgRaw), arg ...define.ConsumeExtraArg) {
	_arg := define.ConsumeExtraArg{}
	if len(arg) > 0 {
		_arg = arg[0]
	}
	if _arg.NATSConsumerName == "" {
		panic("NATS ConsumerName is empty, this will lead to repeated consumption")
	}
	cons, err := m.getConsumer(topic, _arg.NATSConsumerName)
	if err != nil {
		panic(err)
	}

	delay := time.NewTimer(0)
	var batch jetstream.MessageBatch

	for {
		select {
		case <-gCtx.Done():
			return
		case <-delay.C:
			batch, err = cons.FetchNoWait(5)
			if err != nil {
				delay.Reset(time.Second * 5)
				xlog.Error("MqNats [Consume] Fetch failed", zap.Error(err), zap.String("topic", topic.String()))
				continue
			}

			ct := 0
			for msg := range batch.Messages() {
				ct++
				isTimeout := util.RunTaskWithCtxTimeout(
					time.Second*10, func(ctx context.Context) {
						handler(ctx, &_msgRaw{msg: msg})
					},
				)
				if isTimeout {
					xlog.Error("MqNats [Consume] handler timeout", zap.Error(err), zap.String("topic", topic.String()),
						zap.ByteString("msg", msg.Data()))
				}
			}

			if ct == 0 {
				delay.Reset(time.Second * 3)
			} else {
				delay.Reset(0)
			}
		}
	}
}

func (m *MqNats) getSubjectsByStream(streamName, appendSubject string) (subjects []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	subjects = m.streamMgr[streamName]
	if appendSubject != "" {
		subjects = append(subjects, appendSubject)
	}
	m.streamMgr[streamName] = subjects
	return
}

type _msgRaw struct {
	msg jetstream.Msg
}

var _ define.MsgRaw = (*_msgRaw)(nil)

func (m _msgRaw) Bytes() []byte {
	return m.msg.Data()
}

func (m _msgRaw) Ack() error {
	if m.msg.Reply() == "" { // 一般只有在 JetStream 模式下这里才会有值，才支持ack
		return nil
	}
	return m.msg.Ack()
}
