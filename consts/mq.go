package consts

import "fmt"

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

type ConsumerGroup string

const (
	CGDefault ConsumerGroup = "cg_default"
)

func (t ConsumerGroup) String() string {
	return string(t)
}
