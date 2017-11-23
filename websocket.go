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
	"github.com/nats-io/go-nats"
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
		var pbMsgType pb.MessageType

		switch messageType {
		case TEXT:
			pbMsgType = pb.MessageType_TEXT
		case BINARY:
			pbMsgType = pb.MessageType_BINARY
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
		var jsMsgType js.MessageType

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

func (w *NatsWebSocket) registerConnectionInNetPoll(connection *websocket.Conn) *Connection {

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

	return wsConnection
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
	con := w.registerConnectionInNetPoll(connection)
	w.cleanConnectionsIfNeed(con)
}

func (w *NatsWebSocket) cleanConnectionsIfNeed(netConnection *Connection) {

	now := time.Now().Unix()
	stats := w.connections.GetStats()

	if stats.NumberOfNotLoggedConnections > 200 {
		w.connections.RemoveIf(func(con *Connection) bool {

			return now-con.startTime.Unix() > 60

		}, func(con *Connection) {

			w.unregisterConnectionFromNetPoll(con)
			con.Close(websocket.ClosePolicyViolation, "Auth")

		})
	}
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
		connection.SendText([]byte("pong"))
		return
	}

	_, userId, deviceId := connection.GetInfo()

	busClient, err := w.natsPool.Get()
	if err != nil {
		return
	}

	packedMessage := w.packMessage(w.config.PacketFormat, TEXT, userId, deviceId, message)
	busClient.Publish(w.config.NatsOutputSubject, packedMessage)
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
	busClient.Publish(w.config.NatsOutputSubject, packedMessage)
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
		connection.SendText([]byte(LOGIN_PREFIX + "error"))
		return
	}

	err = newToken.Validate([]byte(w.config.JwtSecret), crypto.SigningMethodHS256)
	if err != nil {
		connection.SendText([]byte(LOGIN_PREFIX + "error"))
		return
	}

	claims := newToken.Claims()
	userId := UserId(claims.Get("userId").(string))
	deviceId := DeviceId(claims.Get("deviceId").(string))

	_, conUserId, conDeviceId := connection.GetInfo()

	if conUserId != "" {

		if conUserId != userId || conDeviceId != deviceId {
			connection.SendText([]byte(LOGIN_PREFIX + "error"))
			return
		}

		connection.SendText([]byte(LOGIN_PREFIX + "ok"))
		return
	}

	connection.Login(userId, deviceId)

	deviceConnectionBefore := w.connections.OnLogin(connection)
	if deviceConnectionBefore != nil {
		deviceConnectionBefore.Close(websocket.CloseGoingAway, "OneConnectionPerDevice")
		w.unregisterConnectionFromNetPoll(deviceConnectionBefore)
	}

	connection.SendText([]byte(LOGIN_PREFIX + "ok"))
}

func (w *NatsWebSocket) startListenPacketsFromBus() {

	busClient, err := w.natsPool.Get()
	if err != nil {
		log.Fatalf("Can't connect to nats: %v", err)
		return
	}

	_, err = busClient.Subscribe(w.config.NatsListenSubject, func(msg *nats.Msg) {
		w.handleOutputMsg(msg)
	})

	if err != nil {
		log.Fatalf("Can't connect to nats: %v", err)
		return
	}
}

func (w *NatsWebSocket) handleOutputMsg(msg *nats.Msg) {

	switch w.config.PacketFormat {
	case "protobuf":
		var command pb.Command

		err := proto.Unmarshal(msg.Data, &command)
		if err != nil {
			return
		}

		if command.Method == pb.Command_SEND_MESSAGE_TO_DEVICE {

			params := command.GetSendToDevice()
			deviceId := DeviceId(params.DeviceId)

			messageType := TEXT
			if params.MessageType == pb.MessageType_BINARY {
				messageType = BINARY
			}

			w.sendMessageToDevice(deviceId, messageType, params.Message)

		} else {

			params := command.GetSendToAllUserDevices()
			userId := UserId(params.UserId)
			excludeDevice := DeviceId(params.ExcludeDevice)

			messageType := TEXT
			if params.MessageType == pb.MessageType_BINARY {
				messageType = BINARY
			}

			w.sendMessageToAllUserDevices(userId, excludeDevice, messageType, params.Message)
		}

	case "json":
		var command js.Command

		err := json.Unmarshal(msg.Data, &command)
		if err != nil {
			return
		}

		if command.Method == js.SEND_MESSAGE_TO_DEVICE {

			var params js.SendMessageToDeviceParams
			err := json.Unmarshal(command.Params, &params)
			if err != nil {
				return
			}

			deviceId := DeviceId(params.DeviceId)

			messageType := TEXT
			if params.MessageType == js.BINARY {
				messageType = BINARY
			}

			w.sendMessageToDevice(deviceId, messageType, params.Message)

		} else {

			var params js.SendMessageToAllUserDevicesParams
			err := json.Unmarshal(command.Params, &params)
			if err != nil {
				return
			}

			userId := UserId(params.UserId)
			excludeDevice := DeviceId(params.ExcludeDevice)

			messageType := TEXT
			if params.MessageType == js.BINARY {
				messageType = BINARY
			}

			w.sendMessageToAllUserDevices(userId, excludeDevice, messageType, params.Message)
		}
	}
}

func (w *NatsWebSocket) sendMessageToDevice(deviceId DeviceId, messageType MessageType, message []byte) {

	connection := w.connections.GetDeviceConnection(deviceId)
	if connection == nil {
		return
	}

	switch messageType {
	case TEXT:
		connection.SendText(message)
	case BINARY:
		connection.SendBinary(message)
	}
}

func (w *NatsWebSocket) sendMessageToAllUserDevices(userId UserId, excludeDevice DeviceId, messageType MessageType, message []byte) {

	userConnections := w.connections.GetUserConnections(userId)
	if userConnections == nil {
		return
	}

	for deviceId, connection := range userConnections {
		if deviceId == excludeDevice {
			continue
		}

		switch messageType {
		case TEXT:
			connection.SendText(message)
		case BINARY:
			connection.SendBinary(message)
		}
	}
}

func (w *NatsWebSocket) startHttpServer() {

	mux := http.NewServeMux()
	mux.HandleFunc(w.config.UrlPattern, w.onConnection)
	srv := http.Server{
		Addr:    w.config.ListenInterface,
		Handler: mux,
	}

	w.httpServer = &srv

	log.Println("Start nats-http on: " + w.config.ListenInterface)
	srv.ListenAndServe()
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

	go w.startListenPacketsFromBus()

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
