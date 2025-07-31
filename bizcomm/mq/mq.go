package mq

import (
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	"time"
)

type defaultMsgAPI struct {
	MsTimestamp int64
	UniqueID    string
}

func (*defaultMsgAPI) NeedArchive() bool {
	return true
}

func (v *defaultMsgAPI) SetMsTimestamp(ts int64) {
	v.MsTimestamp = ts
}
func (v *defaultMsgAPI) GetMsgTime() time.Time {
	return time.UnixMilli(v.MsTimestamp)
}

func (v *defaultMsgAPI) GetUniqueID() string {
	return v.UniqueID
}

func (v *defaultMsgAPI) SetUniqueID(uid string) {
	v.UniqueID = uid
}

// ----------------- 优雅的分割线 -----------------

// MsgUserInfoUpdate 更新用户信息消息（仅成功才有）
type MsgUserInfoUpdate struct {
	*defaultMsgAPI
	Body *UserInfoUpdateBody
}

type UserInfoUpdateBody struct {
	UID  int64
	Body *userpb.UpdateBody
}

func NewMsgUserInfoUpdate(body *UserInfoUpdateBody) *MsgUserInfoUpdate {
	return &MsgUserInfoUpdate{
		defaultMsgAPI: &defaultMsgAPI{},
		Body:          body,
	}
}

// MsgSignIn 登录消息
type MsgSignIn struct {
	*defaultMsgAPI
	*SignInBody
}

type SignInBody struct {
	UID        int64
	Nickname   string
	Firstname  string
	Lastname   string
	AppName    string
	AppVersion string
	IP         string
	Platform   commonpb.SignInPlatform
	System     commonpb.SignInSystem
	SignInType commonpb.SignInType
}

func NewMsgSignIn(body *SignInBody) *MsgSignIn {
	return &MsgSignIn{
		defaultMsgAPI: &defaultMsgAPI{},
		SignInBody:    body,
	}
}

// MsgSignUp 注册消息
type MsgSignUp struct {
	*defaultMsgAPI
	Body *SignUpBody
}

type SignUpBody struct {
	UID          int64
	Nickname     string
	Firstname    string
	Lastname     string
	RegisteredAt int64
	RegChan      string
}

func NewMsgSignUp(body *SignUpBody) *MsgSignUp {
	return &MsgSignUp{
		defaultMsgAPI: &defaultMsgAPI{},
		Body:          body,
	}
}

func (m *MsgSignUp) GetUniqueID() string {
	return m.UniqueID
}

type MsgAPICallLog struct {
	*defaultMsgAPI
	*APICallLogBody
}

func NewMsgAPICallLog(body *APICallLogBody) *MsgAPICallLog {
	return &MsgAPICallLog{
		defaultMsgAPI:  &defaultMsgAPI{},
		APICallLogBody: body,
	}
}

type APICallLogBody struct {
	UID         int64
	APIName     string
	APICtrl     string
	ReqIP       string
	DurMs       int64
	Success     bool
	Svc         string
	FromGateway bool
	Panic       bool
	ErrMsg      string
}

func (m *MsgAPICallLog) NeedArchive() bool {
	return false
}

type MsgPushMsg struct {
	*defaultMsgAPI
	*PushMsgBody
}

func (MsgPushMsg) NeedArchive() bool {
	return false
}

func NewMsgPushMsg(body *PushMsgBody) *MsgPushMsg {
	return &MsgPushMsg{
		defaultMsgAPI: &defaultMsgAPI{},
		PushMsgBody:   body,
	}
}

type PushMsgBody struct {
	UID int64
	Msg commonpb.PushMsg
}
