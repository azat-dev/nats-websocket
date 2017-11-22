package nats_websocket

import (
	"github.com/gorilla/websocket"
	"github.com/mailru/easygo/netpoll"
	"sync"
	"time"
)

type ConnectionId int64
type UserId string
type DeviceId string

// Connection wraps user connection.
type Connection struct {
	ws            *websocket.Conn
	desc          *netpoll.Desc
	id            ConnectionId
	userId        UserId
	deviceId      DeviceId
	startTime     time.Time
	lastMessageAt time.Time
	mutex         sync.RWMutex
}

func NewConnection(id ConnectionId, ws *websocket.Conn, desc *netpoll.Desc) *Connection {
	c := &Connection{
		ws:        ws,
		desc:      desc,
		id:        id,
		userId:    "",
		deviceId:  "",
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

	return c.userId != ""
}

func (c *Connection) IsClosed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.IsClosed()
}

func (c *Connection) GetInfo() (ConnectionId, UserId, DeviceId) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.id, c.userId, c.deviceId
}

func (c *Connection) Login(userId UserId, deviceId DeviceId) {
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

func (c *Connection) SetClosed() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.id = -1
	c.userId = ""
	c.deviceId = ""
}
