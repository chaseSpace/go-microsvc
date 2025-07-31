package ws_test

import (
	"log"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/gateway/logic/logic_ws"
	"microsvc/util/ujson"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/cast"

	"github.com/fasthttp/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 3 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	writerExit chan bool // 仅供内存泄漏测试使用

	isWriterExited bool // writer也可能触发close，要标识一下
}

func (c *Client) readPump() {
	defer func() {
		close(c.send) // 触发writer关闭，免得它还在通过关闭的conn发消息
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(appData string) error {
		log.Printf("recv pong\n")
		return nil
	})
	for {
		typ, message, err := c.conn.ReadMessage()
		if err != nil {
			if !c.isWriterExited { // 若不是writer退出，则认为有错误
				log.Printf("readPump-break - error: %v", err)
			}
			break
		}
		log.Printf("client recv: typ:%d %s\n", typ, message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(logic_ws.PingPeriod)
	defer func() {
		ticker.Stop()
		c.isWriterExited = true // 必须在conn.close前执行
		c.conn.Close()
		c.writerExit <- true
	}()

	pingTimes := 3 // 维持3次心跳后退出
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Printf("client write break - close(send)\n")
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				log.Printf("client write break - Write-err:%v\n", err)
				return
			}
		case <-ticker.C: // client 发送心跳
			pingTimes--
			log.Printf("client write ping\n")
			if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				log.Printf("client write break - ping err:%v\n", err)
				return
			}
			if pingTimes == 0 {
				_ = c.conn.WriteControl(websocket.CloseMessage, nil, time.Now().Add(writeWait))
				log.Printf("client write break - ping times end, normal exit!\n")
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func TestServeWs(t *testing.T) {
	h := http.Header{}
	h.Add("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEwMDAxNSwibmlja25hbWUiOiJ4eHgiLCJzZXgiOjEsImxvZ2luX2F0IjoiMjAyNC0xMC0xMiAxNDoxNDozNiIsInJlZ19hdCI6IjIwMjQtMTAtMTIgMTQ6MTM6MTQiLCJpc3MiOiJ4Lm1pY3Jvc3ZjIiwic3ViIjoiMTAwMDE1IiwibmJmIjoxNzI4NzEzNjc2LCJpYXQiOjE3Mjg3MTM2NzYsImp0aSI6IjJuS0Y0TTdIbk8wcVFrWTZ0R3FuUHB1am5YeSJ9.yanf3UdbN_WEAQx7HcdZPccViwoZSvVBVgmMAROAOWo") // 用client登录token来建立websocket连接
	h.Add("platform", cast.ToString(int8(commonpb.SignInPlatform_SIP_PC)))
	h.Add("system", cast.ToString(int8(commonpb.SignInSystem_SIS_MacOS)))
	//println(string(ujson.MustMarshal(&h)))

	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8000/ws", h)
	if err != nil {
		t.Fatal("dial:", err)
	}
	client := &Client{conn: c, send: make(chan []byte, 256), writerExit: make(chan bool)}

	go client.writePump()
	go func() {
		for _, s := range []string{"a", "b", "c"} {
			msg := ujson.MustMarshal(commonpb.ReportMsg{
				Type:   commonpb.ReportMsgType_RMT_TEST,
				DtTest: &commonpb.MsgTest{Reason: s},
			})
			client.send <- msg
			t.Logf(`client sent [%s]`, msg)
		}
	}()
	client.readPump()
	<-client.writerExit // 确保无泄漏

	// 此用例在发送 {pingTimes} 次心跳后自动退出！
}
