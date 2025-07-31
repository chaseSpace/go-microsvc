package commgw

import (
	"context"
	"microsvc/bizcomm/mq"
	"microsvc/consts"
	"microsvc/infra/cache"
	"microsvc/infra/xmq"
	"microsvc/model/svc/gateway"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util/ujson"
	"time"

	"github.com/spf13/cast"
	"go.uber.org/zap"
)

const CKeyPrefix = "GATEWAY:"

const (
	CKeyHashUID2ConnID = CKeyPrefix + "hash_uid_2_conn"
)

type HashValueConn struct {
	ConnId, GatewayID string
	CreatedAt         int64 // sec ts
	Detail            LoginDetail
}

func (v *HashValueConn) OnlineDur() time.Duration {
	return time.Since(time.Unix(v.CreatedAt, 0))
}

type LoginDetail struct {
	Platform commonpb.SignInPlatform
	System   commonpb.SignInSystem
	IP       string
}

func (v *LoginDetail) IsInvalid() bool {
	validPlatform := v.Platform > commonpb.SignInPlatform_SIP_None && commonpb.SignInPlatform_name[int32(v.Platform)] != ""
	validSystem := v.System > commonpb.SignInSystem_SIS_None && commonpb.SignInSystem_name[int32(v.System)] != ""
	return v.IP == "" || !validPlatform || !validSystem
}

func GetConnByUID(ctx context.Context, uid int64) (r *HashValueConn, err error) {
	ret, err := gateway.R.HGet(ctx, CKeyHashUID2ConnID, cast.ToString(uid)).Bytes()
	if cache.IsRedisErr(err) || len(ret) == 0 {
		return nil, xerr.WrapRedis(err)
	}
	return r, ujson.Unmarshal(ret, &r)
}

// SendMsgToUser 发消息给用户，由网关以外的子服务调用，经过mq流转给对应网关进行发送
func SendMsgToUser(ctx context.Context, uid int64, msg *mq.MsgPushMsg) {
	conn, err := GetConnByUID(ctx, uid)
	if err != nil {
		xlog.Error("SendMsgToUser failed", zap.Int64("UID", uid), zap.Error(err))
		return
	}
	if conn == nil || conn.ConnId == "" {
		xlog.DPanic("SendMsgToUser failed, user is OFFLINE", zap.Int64("UID", uid))
		return
	}
	go xmq.Produce(consts.TopicPushMsg.Format(conn.GatewayID), msg)
}

func IsUserOnline(ctx context.Context, uid int64) bool {
	v, err := GetConnByUID(ctx, uid)
	return err == nil && v != nil && v.ConnId != ""
}
