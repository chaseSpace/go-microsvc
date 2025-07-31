package logic_ws

import (
	"context"
	"log"
	"microsvc/bizcomm/commgw"
	"microsvc/infra/svccli/rpc"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/mqconsumerpb"
	"microsvc/util"
	"microsvc/util/ujson"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type ServeWs struct {
	liftTimeStart time.Time
	conn          *websocket.Conn
	send          chan []byte
	recv          chan []byte
	err           []error
	UID           int64
	connID        string

	isWriterExited bool
	detail         *commgw.LoginDetail
}

const (
	maxMessageSize = 1024 * 2

	WriteWait = 5 * time.Second

	PingPeriod = time.Second * 30 // 客户端ping间隔
	//PingPeriod = time.Second * 6 // for test

	PongWait = PingPeriod * 6 / 5 // 略大于ping间隔
)

func (c *ServeWs) ConnId() string {
	return c.connID
}

func New(uid int64, cid string, conn *websocket.Conn, detail *commgw.LoginDetail) *ServeWs {
	return &ServeWs{
		liftTimeStart: time.Now(),
		conn:          conn,
		send:          make(chan []byte, 20),
		recv:          make(chan []byte, 10),
		UID:           uid,
		connID:        cid,
		detail:        detail,
	}
}

func (c *ServeWs) LoginDetail() commgw.LoginDetail {
	return *c.detail
}

func (c *ServeWs) onClosed() {
	_ = c.conn.Close()
	close(c.send)
	close(c.recv)
	if len(c.err) > 0 {
		xlog.Error("ws conn closed", zap.String("connID", c.connID), zap.Error(xerr.JoinErrors(c.err...)))
	} else {
		xlog.Debug("ws conn closed(normally)", zap.String("connID", c.connID))
	}
}

func (c *ServeWs) SendMsg(msg []byte) {
	if len(c.send) == cap(c.send) {
		xlog.Warn("send chan full", zap.String("connID", c.connID))
		return
	}
	xlog.Debug("ws send msg", zap.String("connID", c.connID), zap.Int("len", len(msg)))
	c.send <- msg
}

// Reader 负责读消息、回复ping和释放资源
// 关闭逻辑如下：
// - client关闭，触发read err
// - writer主动关闭，触发read err
func (c *ServeWs) Reader() {
	defer func() {
		println("ws conn closed-r", c.connID)
		c.onClosed() // Triggering writer to exit
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPingHandler(func(string) error {
		log.Printf("recv ping, response pong\n")
		_ = c.conn.WriteControl(websocket.PongMessage, nil, time.Now().Add(WriteWait))
		return nil
	})

	var msg []byte
	var err error
	for {
		_, msg, err = c.conn.ReadMessage()
		if err != nil {
			// 注意这里的函数语义，错误码是反向填写
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) && !c.isWriterExited {
				c.addError(err)
			}
			break
		}
		c.recv <- msg
	}
}

// Writer 只负责发消息
func (c *ServeWs) Writer() {
	ticker := time.NewTicker(PongWait)
	defer func() {
		ticker.Stop()
		c.isWriterExited = true
		// Don't close conn here, let the reader do it
		println("ws conn closed-w", c.connID)
	}()
	for {
		select {
		case message, ok := <-c.send:
			//c.writeCloseMsg()
			//return
			c.setWriteDeadline()
			if !ok {
				// The channel has been closed, and a reader exit will be triggered here.
				c.writeCloseMsg()
				return
			}
			//log.Printf("send msg:%s\n", message)

			// JSON -> text
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				c.addError(errors.Wrap(err, "write message"))
				return
			}
		}
	}
}

func (c *ServeWs) addError(err error) {
	c.err = append(c.err, err)
}

// ProcessReportMsg 异步处理客户端上报消息：提高处理效率
func (c *ServeWs) ProcessReportMsg() {
	defer func() {
		println("ws conn closed-p", c.connID)
	}()
	for msg := range c.recv {
		log.Printf("recv msg:%s\n", msg)
		//c.send <- msg // for test

		// 解析消息，然后转发到微服务处理
		var rmsg commonpb.ReportMsg
		err := ujson.Unmarshal(msg, &rmsg)
		if err != nil {
			xlog.Error("unmarshal report msg failed", zap.Error(err), zap.ByteString("msg", msg))
			pmsg := &commonpb.PushMsg{
				Type: commonpb.PushMsgType_PMT_ErrorMsg,
				Buf:  ujson.MustMarshal(&commonpb.MsgErrorMsg{Text: err.Error()}),
			}
			c.send <- ujson.MustMarshal(pmsg)
			continue
		}

		// 异步
		go util.RunTaskWithCtxTimeout(time.Second*5, func(ctx context.Context) {
			_, err = rpc.MqConsumer().ReportMsg(ctx, &mqconsumerpb.ReportMsgReq{Msg: &rmsg})
			if err != nil {
				xlog.Error("report msg failed", zap.Error(err), zap.Any("msg", &rmsg))
			}
		})

		//commgw.SendMsgToUser(context.TODO(), 100015, mq.NewMqMsgPushMsg(&mq.PushMsgBody{
		//	UID: 100015,
		//	Msg: commonpb.PushMsg{Type: commonpb.PushMsgType_PMT_KickOffline, Buf: ujson.MustMarshal(&commonpb.MsgKickOffline{Reason: "kick"})},
		//}))
	}
}

func (c *ServeWs) setWriteDeadline() {
	_ = c.conn.SetWriteDeadline(time.Now().Add(WriteWait))
}

func (c *ServeWs) writeCloseMsg() {
	err := c.conn.WriteMessage(websocket.CloseMessage, []byte("conn was closed by gateway"))
	if err != nil {
		c.addError(errors.Wrap(err, "on closed"))
	}
}
