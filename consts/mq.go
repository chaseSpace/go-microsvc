package consts

import (
	"fmt"
	"strings"
)

type Topic string

const (
	TopicUserInfoUpdate Topic = "t_user_info_update"
	TopicSignIn         Topic = "t_sign_in"
	TopicSignUp         Topic = "t_sign_up"
	TopicAPICallLog     Topic = "t_api_call_log"
	TopicPushMsg        Topic = "t_push_msg_%s" // 服务器主动推送消息
)

func (t Topic) String() string {
	return string(t)
}

func (t Topic) Format(args ...interface{}) Topic {
	return Topic(fmt.Sprintf(t.String(), args...))
}

// NATSStreamSubject 针对NATS服务，获取其Subject
func (t Topic) NATSStreamSubject() (string, string) {
	// 对应JetStream模式，topic命名规则: STREAM|SUBJECT
	var stream, subject string
	ss := strings.Split(t.String(), "|")
	if len(ss) == 1 {
		stream = "GO_MICROSVC"
		return stream, ss[0]
	}
	stream = ss[0]
	subject = ss[1]
	return stream, subject
}

type ConsumerGroup string // 如 Kafka 使用
type ConsumerName string  // 如 NATS 使用

const (
	CGDefault ConsumerGroup = "cg_default"
)

const (
	ConsumerNameDefault = "consumer_default"
)

func (t ConsumerGroup) String() string {
	return string(t)
}
