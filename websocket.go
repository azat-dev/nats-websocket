package nats_websocket

import (
	"bytes"
	"encoding/json"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/akaumov/nats-websocket/js"
	"github.com/akaumov/nats-websocket/pb"
	"github.com/akaumov/natspool"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/mailru/easygo/netpoll"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type MessageType int32

const (
	TEXT   MessageType = 0
	BINARY MessageType = 1
)

const (
	LOGIN_PREFIX = "login>:"
)

type PackMessageFunc func(packetFormat string, messageType MessageType, userId UserId, deviceId DeviceId, data []byte) []byte

type NatsWebSocket struct {
	config   *Config
	natsPool *natspool.Pool

	httpServer           *http.Server
	upgrader             websocket.Upgrader
	poller               netpoll.Poller
	connections          *ConnectionsStorage
	lastConnectionNumber int64
	inputWorkersPool     *Pool
	packMessage          PackMessageFunc
}

func New(config *Config) *NatsWebSocket {
	return NewCustom(config, nil)
}

func NewCustom(config *Config, packMessageHook PackMessageFunc) *NatsWebSocket {

	if packMessageHook == nil {
		packMessageHook = DefaultPackMessage
	}

	return &NatsWebSocket{
		config:           config,
		upgrader:         websocket.Upgrader{},
		connections:      NewConnectionsStorage(),
		inputWorkersPool: NewPool(config.NumberOfWorkers),
		packMessage:      packMessageHook,
	}
}

func DefaultPackMessage(packetFormat string, messageType MessageType, userId UserId, deviceId DeviceId, data []byte) []byte {

	switch packetFormat {
	case "protobuf":
		var pbMsgType pb.InputMessage_Type

		switch messageType {
		case TEXT:
			pbMsgType = pb.InputMessage_TEXT
		case BINARY:
			pbMsgType = pb.InputMessage_BINARY
		}

		inputMessage := pb.InputMessage{
			Type:      pbMsgType,
			InputTime: time.Now().UnixNano() / 1000000,
			UserId:    string(userId),
			DeviceId:  string(deviceId),
			Body:      data,
		}

		inputMessageBytes, _ := proto.Marshal(&inputMessage)
		return inputMessageBytes

	case "json":
		var jsMsgType js.InputMessageType

		switch messageType {
		case TEXT:
			jsMsgType = js.TEXT
		case BINARY:
			jsMsgType = js.BINARY
		}

		inputMessage := js.InputMessage{
			Type:      jsMsgType,
			InputTime: time.Now().UnixNano() / 1000000,
			UserId:    string(userId),
			DeviceId:  string(deviceId),
			Body:      data,
		}

		inputMessageBytes, _ := json.Marshal(&inputMessage)
		return inputMessageBytes
	}

	log.Panicf("unsuported packet format: %v", packetFormat)
	return nil
}

func (w *NatsWebSocket) getNewConnectionId() ConnectionId {
	return ConnectionId(atomic.AddInt64(&w.lastConnectionNumber, 1))
}

func (w *NatsWebSocket) registerConnectionInNetPoll(connection *websocket.Conn) {

	desc := netpoll.Must(netpoll.HandleRead(connection.UnderlyingConn()))

	wsConnection := NewConnection(w.getNewConnectionId(), connection, desc)
	w.connections.AddNewConnection(wsConnection)

	connection.SetCloseHandler(func(code int, text string) error {
		w.onClose(wsConnection)
		return nil
	})

	w.poller.Start(desc, func(event netpoll.Event) {
		w.inputWorkersPool.Schedule(func() { w.onMessage(wsConnection) })
	})
}

func (w *NatsWebSocket) unregisterConnectionFromNetPoll(connection *Connection) {
	w.poller.Stop(connection.desc)
}

func (w *NatsWebSocket) onConnection(writer http.ResponseWriter, request *http.Request) {

	connection, err := w.upgrader.Upgrade(writer, request, nil)
	if err != nil {
		return
	}

	connection.SetReadLimit(1000)
	w.registerConnectionInNetPoll(connection)
}

func (w *NatsWebSocket) onMessage(netConnection *Connection) {

	messageType, message, err := netConnection.ReadMessage()
	if err != nil {
		w.poller.Stop(netConnection.desc)
		netConnection.Close(websocket.CloseInternalServerErr, "ServerError")
		w.onClose(netConnection)
		return
	}

	netConnection.UpdateLastPingTime()

	switch messageType {
	case websocket.TextMessage:
		w.onTextMessage(netConnection, message)
	case websocket.BinaryMessage:
		w.onBinaryMessage(netConnection, message)
	case websocket.CloseMessage:
		w.onClose(netConnection)
		return
	}
}

func (w *NatsWebSocket) onTextMessage(connection *Connection, message []byte) {

	isLoginMessage := bytes.HasPrefix(message, []byte(LOGIN_PREFIX))
	if isLoginMessage {
		w.login(connection, message[len(LOGIN_PREFIX):])
		return
	}

	if !connection.IsLoggedIn() {
		return
	}

	if bytes.Compare(message, []byte("ping")) == 0 {
		connection.Send([]byte("pong"))
		return
	}

	_, userId, deviceId := connection.GetInfo()

	busClient, err := w.natsPool.Get()
	if err != nil {
		return
	}

	packedMessage := w.packMessage(w.config.PacketFormat, TEXT, userId, deviceId, message)
	busClient.Publish("nats-websocket", packedMessage)
}

func (w *NatsWebSocket) onBinaryMessage(netConnection *Connection, message []byte) {

	if !netConnection.IsLoggedIn() {
		return
	}
	_, userId, deviceId := netConnection.GetInfo()

	busClient, err := w.natsPool.Get()
	if err != nil {
		return
	}

	packedMessage := w.packMessage(w.config.PacketFormat, BINARY, userId, deviceId, message)
	busClient.Publish("nats-websocket", packedMessage)
}

func (w *NatsWebSocket) onClose(connection *Connection) {

	connectionId, _, _ := connection.GetInfo()
	if connectionId == -1 {
		return
	}

	w.unregisterConnectionFromNetPoll(connection)
	w.connections.RemoveConnection(connection)
}

func (w *NatsWebSocket) login(connection *Connection, tokenString []byte) {

	newToken, err := jws.ParseJWT([]byte(tokenString))
	if err != nil {
		connection.Send([]byte(LOGIN_PREFIX + "error"))
		return
	}

	err = newToken.Validate([]byte(w.config.JwtSecret), crypto.SigningMethodHS256)
	if err != nil {
		connection.Send([]byte(LOGIN_PREFIX + "error"))
		return
	}

	claims := newToken.Claims()
	userId := UserId(claims.Get("userId").(string))
	deviceId := DeviceId(claims.Get("deviceId").(string))

	_, conUserId, conDeviceId := connection.GetInfo()

	if conUserId != "" {

		if conUserId != userId || conDeviceId != deviceId {
			connection.Send([]byte(LOGIN_PREFIX + "error"))
			return
		}

		connection.Send([]byte(LOGIN_PREFIX + "ok"))
		return
	}

	connection.Login(userId, deviceId)

	deviceConnectionBefore := w.connections.OnLogin(connection)
	if deviceConnectionBefore != nil {
		w.unregisterConnectionFromNetPoll(deviceConnectionBefore)
		deviceConnectionBefore.Close(websocket.CloseGoingAway, "OneConnectionPerDevice")
	}

	connection.Send([]byte(LOGIN_PREFIX + "ok"))
}

func (w *NatsWebSocket) startHttpServer() {

	http.HandleFunc(w.config.UrlPattern, w.onConnection)

	srv := http.Server{
		Addr: w.config.ListenInterface,
	}

	w.httpServer = &srv

	log.Println("Start nats-http on: " + w.config.ListenInterface)
	log.Fatal(srv.ListenAndServe())
}

func getOsSignalWatcher() chan os.Signal {

	stopChannel := make(chan os.Signal)
	signal.Notify(stopChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)

	return stopChannel
}

func (w *NatsWebSocket) Start() {

	stopSignal := getOsSignalWatcher()

	poller, err := netpoll.New(nil)
	if err != nil {
		log.Panicf("Can't start poller")
		return
	}

	w.poller = poller

	natsPool, err := natspool.New(w.config.NatsAddress, w.config.NatsPoolSize)
	if err != nil {
		log.Panicf("can't connect to nats: %v", err)
	}

	w.natsPool = natsPool
	defer func() { natsPool.Empty() }()

	go func() {
		<-stopSignal
		w.Stop()
	}()

	w.startHttpServer()
}

func (w *NatsWebSocket) Stop() {

	if w.httpServer != nil {
		w.httpServer.Shutdown(nil)
		log.Println("http: shutdown")
	}

	w.natsPool.Empty()
	log.Println("natspool: empty")
}
