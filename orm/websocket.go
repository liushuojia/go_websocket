package orm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

const (
	heartbeatTotal = 10 // 心跳最多监控次数
	heartbeatWait  = 10 // 心跳等待时间
)

//升级长连接
var WS = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 60 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Connection struct {
	wsConnect    *websocket.Conn
	inChan       chan []byte
	outChan      chan []byte
	closeChan    chan byte
	mutex        sync.Mutex // 对closeChan关闭上锁
	isClosed     bool       // 防止closeChan被关闭多次
	heartbeatNum int64      // 心跳监控次数
}

type OnelineMessage struct {
	Action   string `json:"action"`             // action:	heartbeat	publish		subscribe	unsubscribe
	ClientId string `json:"clientId,omitempty"` //
	Topic    string `json:"topic,omitempty"`    //
	Message  string `json:"message,omitempty"`  //
}

func InitConnection(c *gin.Context) (*Connection, error) {
	wsConn, err := WS.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		wsConnect:    wsConn,
		inChan:       make(chan []byte, 1000),
		outChan:      make(chan []byte, 1000),
		closeChan:    make(chan byte, 1),
		heartbeatNum: 0,
	}

	// 启动读写协程
	go conn.readLoop()
	go conn.writeLoop()
	go conn.heartbeatLoop()
	return conn, nil
}
func (conn *Connection) Close() {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	// 线程安全，可多次调用
	conn.wsConnect.Close()

	// 利用标记，让closeChan只关闭一次
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}

}
func (conn *Connection) IsClosed() bool {
	return conn.isClosed
}
func (conn *Connection) heartbeatLoop() {
	index := 1
	//心跳监控
	for {
		time.Sleep(heartbeatWait * time.Second)
		if conn.heartbeatNum > heartbeatTotal {
			goto ERR
		}
		j, _ := json.Marshal(OnelineMessage{
			Action:  "heartbeat",
			Message: fmt.Sprintf("%d times", index),
		})
		if err := conn.WriteMessage(j); err != nil {
			goto ERR
		}
		conn.heartbeatNum++
	}
ERR:
	conn.Close()
}

func (conn *Connection) ReadMessage() ([]byte, error) {
	select {
	case data := <-conn.inChan:
		conn.heartbeatNum = 0
		return data, nil
	case <-conn.closeChan:
		return nil, errors.New("connection is closeed")
	}
}
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConnect.ReadMessage(); err != nil {
			goto ERR
		}

		//阻塞在这里，等待inChan有空闲位置
		select {
		case conn.inChan <- data:
		case <-conn.closeChan: // closeChan 感知 conn断开
			goto ERR
		}
	}
ERR:
	conn.Close()
}

func (conn *Connection) WriteMessage(data []byte) error {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		return errors.New("connection is closeed")
	}
	return nil
}
func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)
	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}
		if err = conn.wsConnect.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	conn.Close()
}
