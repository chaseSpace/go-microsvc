package mqkafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/k0kubun/pp/v3"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"microsvc/consts"
	"microsvc/deploy"
	"microsvc/infra/xmq/define"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"strings"
	"sync"
	"time"
)

var _ define.MqProviderAPI = (*MqKafka)(nil)

type MqKafka struct {
	rwMap            sync.Map
	mainThreadCtx    context.Context
	mainThreadCancel context.CancelFunc
	waitConsumers    sync.WaitGroup
}

// New MqKafka，建议使用kafka 4.x
func New() *MqKafka {
	return &MqKafka{}
}

func (m *MqKafka) Name() string {
	return "kafka"
}

func (m *MqKafka) Init(cc *deploy.MqConfig) (err error) {
	defer func() {
		if err != nil {
			err = xerr.Wrap(err, "failed to connect to kafka")
		}
	}()
	m.mainThreadCtx, m.mainThreadCancel = context.WithCancel(context.TODO())

	// 此处仅测试kafka是否可正常连接，启动后懒加载reader、writer
	conn, err := kafka.DialLeader(context.Background(), "tcp", cc.Kafka.Brokers[0], "topic_example", 0)
	if err != nil {
		return err
	}
	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	_, err = conn.WriteMessages(
		kafka.Message{Value: []byte("ping!")},
	)
	if err != nil {
		return err
	}
	_ = conn.Close() // 用完即抛

	return
}

func (m *MqKafka) Stop() error {
	m.mainThreadCancel()
	m.waitConsumers.Wait()
	println("Waiting all readers & writers quit...")
	m.rwMap.Range(func(key, value any) bool {
		// 这里每次close都耗时几秒，why？
		if strings.HasSuffix(key.(string), "|reader") {
			_ = value.(*kafka.Reader).Close()
		} else {
			_ = value.(*kafka.Writer).Close()
		}
		_, _ = pp.Printf("[%s] released.\n", key)
		return true
	})
	println() // 换行
	xlog.Debug("mq-kafka: resource released...")
	return nil
}

func (m *MqKafka) LazyGetWriter(topic consts.Topic) *kafka.Writer {
	k := fmt.Sprintf("%s|writer", topic)
	if w, ok := m.rwMap.Load(k); ok {
		return w.(*kafka.Writer)
	}
	w := &kafka.Writer{
		Addr:  kafka.TCP(deploy.XConf.Kafka.Brokers...),
		Topic: topic.String(),
		// 指定分区的balancer模式为最小字节分布，其他有RoundRobin、Hash、 ReferenceHash、 CRC32Balancer、 Murmur2Balancer
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll, // ack模式
		Async:        false,            // 异步?
	}
	_w, _ := m.rwMap.LoadOrStore(k, w)
	return _w.(*kafka.Writer)
}

func (m *MqKafka) LazyGetReader(topic consts.Topic, groupId consts.ConsumerGroup) *kafka.Reader {
	k := fmt.Sprintf("%s|%s|reader", topic, groupId)
	if r, ok := m.rwMap.Load(k); ok {
		return r.(*kafka.Reader)
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          deploy.XConf.Kafka.Brokers,
		Topic:            topic.String(),
		GroupID:          groupId.String(), // 必须指定，否则从offset=0开始消费
		ReadBatchTimeout: time.Second * 10,
		//Partition: 0, 有了GroupID就无需指定
	})
	_r, _ := m.rwMap.LoadOrStore(k, r)
	return _r.(*kafka.Reader)
}

func (m *MqKafka) Produce(ctx context.Context, topic consts.Topic, msg []byte) error {
	if m.mainThreadCtx.Err() != nil {
		return errors.New("main thread quited")
	}
	// 生产消息的过程不需要保护，没发出的消息可忽略（caller应判断发送结果）

	w := m.LazyGetWriter(topic)
	err := w.WriteMessages(ctx,
		kafka.Message{
			Key:   nil, // 分区策略使用，可置空
			Value: msg,
		},
	)
	if err != nil {
		xlog.Error("MqKafka [Produce] err", zap.Error(err),
			zap.String("topic", topic.String()), zap.String("msg", util.TruncateUTF8(string(msg), 20)))
	}
	return err
}

// Consume 消费消息
// - 消费过程受到保护，若主线程停止，需要等待此协程退出
func (m *MqKafka) Consume(topic consts.Topic, handler func(ctx context.Context, msg []byte), _arg ...define.ConsumeExtraArg) {
	arg := define.ConsumeExtraArg{}
	if len(_arg) > 0 {
		arg = _arg[0]
	}
	r := m.LazyGetReader(topic, arg.ConsumeGroupId)

	m.waitConsumers.Add(1)
	defer func() { m.waitConsumers.Done() }()

	var logTicker = time.NewTicker(time.Second * 3)
	var ct int
	var msg kafka.Message
	var err error

	read := func() (msg kafka.Message, err error) {
		var tempCtx, cancel = context.WithTimeout(m.mainThreadCtx, time.Second*3)
		defer cancel()
		return r.ReadMessage(tempCtx)
	}

	for {
		msg, err = read()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) { // timeout
				continue
			}
			if errors.Is(err, context.Canceled) { // 主线程退出
				xlog.Info("MqKafka [Consume] main thread quit", zap.String("topic", topic.String()), zap.String("groupId", arg.ConsumeGroupId.String()))
				return
			}
			xlog.Error("MqKafka [Consume] ReadMessage failed", zap.Error(err), zap.String("topic", topic.String()), zap.String("groupId", arg.ConsumeGroupId.String()))
			time.Sleep(time.Second * 3) // 故障退避
			continue
		}

		// successful
		handler(context.TODO(), msg.Value)
		ct++

		// 控制日志频率
		select {
		case <-logTicker.C:
			xlog.Info("MqKafka [Consume] Counting", zap.Int("COUNT", ct), zap.String("topic", topic.String()),
				zap.String("groupId", arg.ConsumeGroupId.String()))
		}
	}
}
