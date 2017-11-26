package nats_websocket

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type ConnectionId int64
type UserId string
type DeviceId string

// Connection wraps user connection.
type Connection struct {
	ws            *websocket.Conn
	id            ConnectionId
	userId        UserId
	deviceId      DeviceId
	startTime     time.Time
	lastMessageAt time.Time
	dataMutex     sync.RWMutex
	writeMutex    sync.Mutex
}

func NewConnection(id ConnectionId, ws *websocket.Conn) *Connection {
	c := &Connection{
		ws:         ws,
		id:         id,
		userId:     "",
		deviceId:   "",
		startTime:  time.Now(),
		dataMutex:  sync.RWMutex{},
		writeMutex: sync.Mutex{},
	}
	return c
}

func (c *Connection) ReadMessage() (messageType int, p []byte, err error) {
	return c.ws.ReadMessage()
}

func (c *Connection) SendText(message []byte) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	c.ws.WriteMessage(websocket.TextMessage, message)
}

func (c *Connection) SendBinary(message []byte) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	c.ws.WriteMessage(websocket.BinaryMessage, message)
}

func (c *Connection) Close(code int, reason string) {
	c.dataMutex.Lock()
	c.dataMutex.Unlock()
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason))
	c.ws.Close()

	c.id = -1
	c.userId = ""
	c.deviceId = ""
}

func (c *Connection) IsLoggedIn() bool {
	c.dataMutex.RLock()
	defer c.dataMutex.RUnlock()

	return c.userId != ""
}

func (c *Connection) IsClosed() bool {
	c.dataMutex.RLock()
	defer c.dataMutex.RUnlock()

	return c.IsClosed()
}

func (c *Connection) GetInfo() (ConnectionId, UserId, DeviceId) {
	c.dataMutex.RLock()
	defer c.dataMutex.RUnlock()

	return c.id, c.userId, c.deviceId
}

func (c *Connection) GetStartTime() time.Time {
	c.dataMutex.RLock()
	defer c.dataMutex.RUnlock()

	return c.startTime
}

func (c *Connection) Login(userId UserId, deviceId DeviceId) {
	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()

	c.userId = userId
	c.deviceId = deviceId
	c.ws.SetReadLimit(0)
}

func (c *Connection) UpdateLastPingTime() {
	c.dataMutex.Lock()
	defer c.dataMutex.Unlock()

	c.lastMessageAt = time.Now()
}
