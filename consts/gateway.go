package consts

type PushMsgID string

const (
	PushMsgID_KickOffline PushMsgID = "kick_offline" // 踢下线
)

func (s PushMsgID) Str() string {
	return string(s)
}
