package logic_wsmanage

import (
	"context"
	"errors"
	"microsvc/bizcomm/commgw"
	"microsvc/bizcomm/mq"
	"microsvc/consts"
	"microsvc/infra/xmq"
	"microsvc/infra/xmq/define"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/service/gateway/cache"
	"microsvc/service/gateway/deploy"
	"microsvc/service/gateway/logic/logic_ws"
	"microsvc/util/ujson"
	"sync"

	"go.uber.org/zap"
)

type WsManagerT struct {
	wsClients map[string]*logic_ws.ServeWs
	mu        sync.RWMutex
}

var WsManager = &WsManagerT{
	wsClients: make(map[string]*logic_ws.ServeWs),
}

func Init() {
	go WsManager.ConsumeIMMsg()
}

func (m *WsManagerT) ProtectWrite(f func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f()
}
func (m *WsManagerT) ProtectRead(f func()) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f()
}

// ConsumeIMMsg 持续消费来自mq的推送消息
func (m *WsManagerT) ConsumeIMMsg() {

	topic := consts.TopicPushMsg.Format(deploy.GatewayConf.GwID())
	// 这个消费者是个特殊，不归于 mqconsumer 管理
	xmq.Consume[mq.MsgPushMsg](topic, func(ctx context.Context, t mq.MsgPushMsg) error {
		var err error
		var cacheConn *commgw.HashValueConn

		defer func() {
			if err != nil {
				xlog.Error("ConsumeIMMsg defer", zap.Error(err), zap.Any("connObj", cacheConn), zap.Int64("uid", t.UID),
					zap.String("dropped.msgType", t.Msg.Type.String()))
			} else {
				xlog.Debug("ConsumeIMMsg", zap.Int64("uid", t.UID), zap.String("msgType", t.Msg.Type.String()))
			}
		}()

		cacheConn, err = cache.WsCache.GetConnByUID(ctx, t.UID)
		if err != nil {
			return err
		}
		if cacheConn == nil {
			err = errors.New("user is offline")
			return err
		}
		var conn *logic_ws.ServeWs
		var ok bool
		m.ProtectRead(func() {
			if conn, ok = m.wsClients[cacheConn.ConnId]; !ok {
				err = xerr.New("conn not found")
			}
		})
		if ok {
			// 使用 proto-json 方法而不是常规json，是为了利用前者的 EmitUnpopulated:false 特性
			// - 能够减少传输字节
			conn.SendMsg(ujson.MustProtoJsonMarshal(&t.Msg, false))
		}
		return err
	}, define.ConsumeExtraArg{ConsumeGroupId: consts.CGDefault})
}

func (m *WsManagerT) AddConn(ws *logic_ws.ServeWs) error {
	if ws.ConnId() == "" {
		return errors.New("ConnID is empty")
	}
	if ws.UID == 0 {
		return errors.New("UID is 0")
	}
	var err error
	m.ProtectWrite(func() {
		if m.wsClients[ws.ConnId()] != nil {
			err = errors.New("ConnID already exists") // should not reach! (bug)
		} else {
			connSave := cache.WsCache.HashValueConn(ws.ConnId(), deploy.GatewayConf.GwID(), ws.LoginDetail())
			err = cache.WsCache.SaveUserConnMap(context.TODO(), ws.UID, connSave)
			if err == nil {
				m.wsClients[ws.ConnId()] = ws
			}
		}
	})
	xlog.Info("AddConn --DEBUG--", zap.Int64("UID", ws.UID), zap.String("connID", ws.ConnId()), zap.Error(err))
	return err
}

func (m *WsManagerT) DeleteConn(connId string) {
	var conn *logic_ws.ServeWs
	m.ProtectWrite(func() {
		conn = m.wsClients[connId]
		if conn != nil {
			// 关闭channel后，会自动关闭goroutine
			delete(m.wsClients, connId)
		}
	})
	if conn != nil {
		_ = cache.WsCache.DeleteConnByUID(context.TODO(), conn.UID)
	}
}

func (m *WsManagerT) ConnectCount() (ct int) {
	m.ProtectRead(func() {
		ct = len(m.wsClients)
	})
	return
}

func (m *WsManagerT) OnClose() {
	var uids []int64
	m.ProtectRead(func() {
		for _, conn := range m.wsClients {
			uids = append(uids, conn.UID)
		}
	})
	if len(uids) > 0 {
		_ = cache.WsCache.DeleteConnByUID(context.TODO(), uids...)
	}
}
