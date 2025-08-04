package mqredis

import (
	"context"
	"errors"
	"microsvc/consts"
	"microsvc/deploy"
	"microsvc/infra/xmq/define"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var _ define.MqProviderAPI = (*MqRedis)(nil)
var gCtx, gCancel = context.WithCancel(context.Background())

type MqRedis struct {
	name       string
	cli        *redis.Client
	listKeyPre string
}

func New() *MqRedis {
	return &MqRedis{listKeyPre: "msg_queue:"}
}

func (m *MqRedis) Name() string {
	return "redis"
}
func (m *MqRedis) Init(cc *deploy.MqConfig) (err error) {
	v := cc.Redis.Meta
	m.cli = redis.NewClient(&redis.Options{
		Addr:       v.Addr,
		Password:   v.Password,
		DB:         v.DB,
		MaxRetries: 2,
	})
	util.RunTaskWithCtxTimeout(time.Second, func(ctx context.Context) {
		err = m.cli.Ping(ctx).Err()
	})
	return
}

func (m *MqRedis) Stop() error {
	return m.cli.Close()
}

func (m *MqRedis) key(topic consts.Topic) string {
	return strings.Join([]string{m.listKeyPre, topic.String()}, ":")
}

func (m *MqRedis) Produce(ctx context.Context, topic consts.Topic, msg []byte) error {
	err := m.cli.LPush(ctx, m.key(topic), msg).Err()
	if err != nil {
		xlog.Error("MqRedis [Produce] LPush err", zap.Error(err),
			zap.String("topic", topic.String()), zap.String("msg", string(msg)))
	}
	return err
}

func (m *MqRedis) Consume(topic consts.Topic, handler func(ctx context.Context, msg define.MsgRaw), _arg ...define.ConsumeExtraArg) {
	delay := time.NewTimer(0)
	for {
		select {
		case <-gCtx.Done():
			return
		case <-delay.C:
			val, err := m.cli.BRPop(gCtx, time.Second, m.key(topic)).Result()
			if errors.Is(err, redis.Nil) {
				delay.Reset(time.Second * 3)
				continue
			}
			if err != nil {
				delay.Reset(time.Second * 5)
				xlog.Error("MqRedis [Consume] BRPop failed", zap.Error(err), zap.String("topic", topic.String()))
				continue
			}

			delay.Reset(0)
			isTimeout := util.RunTaskWithCtxTimeout(
				time.Second*10, func(ctx context.Context) {
					handler(ctx, &_msgRaw{msg: []byte(val[1])})
				},
			)
			if isTimeout {
				xlog.Error("MqRedis [Consume] handler timeout", zap.Error(err), zap.String("topic", topic.String()), zap.String("msg", val[1]))
			}
		}
	}
}

type _msgRaw struct {
	msg []byte
}

var _ define.MsgRaw = (*_msgRaw)(nil)

func (m _msgRaw) Bytes() []byte {
	return m.msg
}

func (m _msgRaw) Ack() error {
	// todo redis 暂无 ack机制，可自行实现
	return nil
}
