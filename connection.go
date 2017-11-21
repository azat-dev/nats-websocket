package nats_websocket

import (
	"github.com/gorilla/websocket"
	"github.com/mailru/easygo/netpoll"
	"sync"
	"time"
)

// Connection wraps user connection.
type Connection struct {
	ws            *websocket.Conn
	desc          *netpoll.Desc
	id            int64
	userId        *string
	deviceId      *string
	startTime     time.Time
	lastMessageAt time.Time
	mutex         sync.RWMutex
}

func NewConnection(id int64, ws *websocket.Conn, desc *netpoll.Desc) *Connection {
	c := &Connection{
		ws:        ws,
		desc:      desc,
		id:        id,
		userId:    nil,
		deviceId:  nil,
		startTime: time.Now(),
		mutex:     sync.RWMutex{},
	}
	return c
}

func (c *Connection) ReadMessage() (messageType int, p []byte, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.ws.ReadMessage()
}

func (c *Connection) Send(message []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ws.WriteMessage(websocket.TextMessage, message)
}

func (c *Connection) Close(reason string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason))
	c.ws.Close()
}

func (c *Connection) IsLoggedIn() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.userId != nil
}

func (c *Connection) IsClosed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.IsClosed()
}

func (c *Connection) GetInfo() (int64, *string, *string) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.id, c.deviceId, c.userId
}

func (c *Connection) Login(userId *string, deviceId *string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.userId = userId
	c.deviceId = deviceId
	c.ws.SetReadLimit(0)
}

func (c *Connection) UpdateLastPingTime() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lastMessageAt = time.Now()
}
