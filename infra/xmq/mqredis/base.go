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

type MqRedis struct {
	name       string
	cli        *redis.Client
	listKeyPre string
}

func New() *MqRedis {
	return &MqRedis{listKeyPre: "mq_cache"}
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
	xlog.Debug("mq-redis: resource released...")
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

func (m *MqRedis) Consume(topic consts.Topic, handler func(ctx context.Context, msg []byte), _arg ...define.ConsumeExtraArg) {
	serverFailed := false
	for {
		if serverFailed {
			time.Sleep(time.Second * 5)
		}
		val, err := m.cli.BRPop(context.TODO(), time.Second*10, m.key(topic)).Result()
		if errors.Is(err, redis.Nil) {
			continue
		}
		if err != nil {
			serverFailed = true
			xlog.Error("MqRedis [Consume] BRPop failed", zap.Error(err), zap.String("topic", topic.String()))
			continue
		}
		serverFailed = false
		handler(context.TODO(), []byte(val[1]))
	}
}
