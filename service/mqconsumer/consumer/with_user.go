package consumer

import (
	"context"
	"microsvc/bizcomm/mq"
	"microsvc/consts"
	"microsvc/infra/xmq"
	"microsvc/infra/xmq/define"
	"microsvc/model"
	"microsvc/model/svc/user"
	"microsvc/pkg/xlog"
	"time"

	"go.uber.org/zap"
)

// 用户相关的主题消费

type withUserT struct {
}

var consumerWithUser = &withUserT{}

func (w *withUserT) ConsumerName() string {
	return "User"
}

func (w *withUserT) Init() {
	go w.ConsumeSignIn()
	go w.ConsumeSignUp()
	go w.ConsumeUserInfoUpdate()
}

// ConsumeSignIn 消费登录消息
func (*withUserT) ConsumeSignIn() {
	topic := consts.TopicSignIn
	xmq.Consume[*mq.MsgSignIn](topic, func(ctx context.Context, _msg *mq.MsgSignIn) error {
		// 受限于GO泛型规则，这里似乎没有更好的写法
		// - 注意，这里分发给协程的msg要使用值类型，避免潜在的并发修改问题
		msg := *_msg

		return consumeSignInV.WriteSignInLog(ctx, msg)

	}, define.ConsumeExtraArg{ConsumeGroupId: consts.CGDefault})
}

func (*withUserT) ConsumeSignUp() {
	topic := consts.TopicSignUp
	xmq.Consume[*mq.MsgSignUp](topic, func(ctx context.Context, _msg *mq.MsgSignUp) error {
		msg := *_msg

		xlog.Debug("ConsumeSignUp success", zap.Any("msg", msg))
		return nil
	}, define.ConsumeExtraArg{ConsumeGroupId: consts.CGDefault})
}

func (*withUserT) ConsumeUserInfoUpdate() {
	topic := consts.TopicUserInfoUpdate
	xmq.Consume[*mq.MsgUserInfoUpdate](topic, func(ctx context.Context, _msg *mq.MsgUserInfoUpdate) error {
		msg := *_msg

		ctx = context.TODO()
		return consumerUserInfoUpdateV.PanicTest(ctx, msg)
	}, define.ConsumeExtraArg{ConsumeGroupId: consts.CGDefault})
}

// ----------------- 优雅的分割线（下面定义不同topic的多个消费方法） -----------------

// 消费登录消息
type consumeSignIn struct {
}

var consumeSignInV consumeSignIn

func (consumeSignIn) WriteSignInLog(ctx context.Context, msg mq.MsgSignIn) error {
	err := user.QLog.WithContext(ctx).Create(&user.SignInLog{
		FieldUID: model.FieldUID{UID: msg.UID},
		Platform: msg.Platform,
		System:   msg.System,
		Type:     msg.SignInType,
		IP:       msg.IP,
		SignInAt: time.UnixMilli(msg.MsTimestamp),
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// 消费用户信息更新消息
type consumerUserInfoUpdate struct {
}

var consumerUserInfoUpdateV consumerUserInfoUpdate

func (consumerUserInfoUpdate) PanicTest(ctx context.Context, msg mq.MsgUserInfoUpdate) error {
	panic(111)
}
