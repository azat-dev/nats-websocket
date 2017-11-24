package nats_websocket

import (
	"bytes"
	"encoding/json"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/akaumov/nats-pool"
	"github.com/akaumov/nats-websocket/js"
	"github.com/akaumov/nats-websocket/pb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/nats-io/go-nats"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"sync"
	. "testing"
	"time"
)

const (
	LISTEN_INTERFACE    = "localhost:8080"
	JWT_SECRET          = "123456"
	NATS_ADDRESS        = "nats://localhost:4222"
	NATS_POOL_SIZE      = 200
	NATS_OUTPUT_SUBJECT = "nats-websocket-received"
	NATS_LISTEN_SUBJECT = "nats-websocket-send"
)

var dialer = websocket.Dialer{}

func TestConnect(t *T) {

	server := startWsServer("json")
	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go startTestConnectConnection(t, i%2 == 0, &wg)
	}

	wg.Wait()
	server.Stop()
}

func startWsServer(packetFormat string) *NatsWebSocket {
	natsWebsocket := New(&Config{
		ListenInterface:   LISTEN_INTERFACE,
		JwtSecret:         JWT_SECRET,
		Timeout:           30000,
		UrlPattern:        "/",
		NatsAddress:       NATS_ADDRESS,
		NatsPoolSize:      NATS_POOL_SIZE,
		PacketFormat:      packetFormat,
		NumberOfWorkers:   20,
		NatsOutputSubject: NATS_OUTPUT_SUBJECT,
		NatsListenSubject: NATS_LISTEN_SUBJECT,
	})

	go func() {
		natsWebsocket.Start()
	}()

	time.Sleep(2 * time.Second)

	return natsWebsocket
}

func getToken(userId string, deviceId string) []byte {
	claims := jws.Claims{}

	claims.Set("userId", userId)
	claims.Set("deviceId", deviceId)

	token := jws.NewJWT(claims, crypto.SigningMethodHS256)
	serialized, _ := token.Serialize([]byte(JWT_SECRET))

	return serialized
}

func startTestConnectConnection(t *T, sendWrongToken bool, wg *sync.WaitGroup) {

	defer wg.Done()

	conn, _, err := dialer.Dial("ws://"+LISTEN_INTERFACE+"/", nil)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	var token []byte

	if sendWrongToken {
		token = uuid.NewV4().Bytes()
	} else {
		token = getToken(uuid.NewV4().String(), uuid.NewV4().String())
	}

	//test login
	err = conn.WriteMessage(websocket.TextMessage, []byte("login>:"+string(token)))
	assert.Nil(t, err)

	messageType, message, err := conn.ReadMessage()
	assert.Nil(t, err)
	assert.Equal(t, websocket.TextMessage, messageType)

	if sendWrongToken {
		assert.Equal(t, "login>:"+"error", string(message))
	} else {
		assert.Equal(t, "login>:"+"ok", string(message))
	}

	//test ping
	err = conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	assert.Nil(t, err)

	messageType, message, err = conn.ReadMessage()

	if sendWrongToken {
		assert.NotNil(t, err)
	} else {

		assert.Nil(t, err)
		assert.Equal(t, websocket.TextMessage, messageType)
		assert.Equal(t, "pong", string(message))
	}

	conn.Close()
}

func TestReceiveMessagesJSON(t *T) {
	testReceiveMessages(t, "json")
}

func TestReceiveMessagesProtobuf(t *T) {
	testReceiveMessages(t, "protobuf")
}

func testReceiveMessages(t *T, packetFormat string) {

	server := startWsServer(packetFormat)

	const NUMBER_OF_CONNECTIONS = 10

	natsPool, err := nats_pool.New(NATS_ADDRESS, 2)
	assert.Nil(t, err)

	subscribeClient, err := natsPool.Get()
	assert.Nil(t, err)

	usersMutex := sync.Mutex{}
	users := make([]string, 0)

	for i := 0; i < NUMBER_OF_CONNECTIONS; i++ {
		userId := uuid.NewV4().String()
		users = append(users, userId)
	}

	receivedChan := make(chan bool, NUMBER_OF_CONNECTIONS)

	subWaitGroup := sync.WaitGroup{}
	subWaitGroup.Add(NUMBER_OF_CONNECTIONS)

	subscribeClient.Subscribe(NATS_OUTPUT_SUBJECT, func(msg *nats.Msg) {

		usersMutex.Lock()
		defer usersMutex.Unlock()

		handleReceivedMessage(t, packetFormat, users, msg, receivedChan)

		subWaitGroup.Done()
	})

	connectionsWaitGroup := sync.WaitGroup{}

	for _, userId := range users {
		connectionsWaitGroup.Add(1)
		go startTestReceiveConnection(t, &connectionsWaitGroup, userId, userId)
	}

	connectionsWaitGroup.Wait()
	subWaitGroup.Wait()
	assert.Equal(t, NUMBER_OF_CONNECTIONS, len(receivedChan))

	natsPool.Put(subscribeClient)
	natsPool.Empty()

	server.Stop()
}

func handleReceivedMessage(t *T, packetFormat string, users []string, msg *nats.Msg, receivedChan chan bool) {

	var userId, deviceId string
	var body []byte

	if packetFormat == "json" {

		var input js.InputMessage
		err := json.Unmarshal(msg.Data, &input)
		assert.Nil(t, err)

		userId = input.UserId
		deviceId = input.DeviceId
		body = input.Body

		assert.Equal(t, js.TEXT, input.Type)

	} else {

		var input pb.InputMessage
		err := proto.Unmarshal(msg.Data, &input)
		assert.Nil(t, err)

		userId = input.UserId
		deviceId = input.DeviceId
		body = input.Body

		assert.Equal(t, pb.MessageType_TEXT, input.Type)
	}

	isUserFound := false
	for _, existingUserId := range users {
		if existingUserId == userId && existingUserId == deviceId {
			isUserFound = true
			break
		}
	}

	assert.Equal(t, true, isUserFound)
	assert.Equal(t, []byte("HELLO"), body)

	if bytes.Compare(body, []byte("HELLO")) == 0 {
		receivedChan <- true
	}
}

func startTestReceiveConnection(t *T, wg *sync.WaitGroup, userId string, deviceId string) {

	defer wg.Done()

	conn, _, err := dialer.Dial("ws://"+LISTEN_INTERFACE+"/", nil)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	var token []byte

	token = getToken(userId, deviceId)

	//test login
	err = conn.WriteMessage(websocket.TextMessage, []byte("login>:"+string(token)))
	assert.Nil(t, err)

	messageType, message, err := conn.ReadMessage()
	assert.Nil(t, err)
	assert.Equal(t, websocket.TextMessage, messageType)
	assert.Equal(t, "login>:"+"ok", string(message))

	//test ping
	err = conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	assert.Nil(t, err)

	messageType, message, err = conn.ReadMessage()

	assert.Nil(t, err)
	assert.Equal(t, websocket.TextMessage, messageType)
	assert.Equal(t, "pong", string(message))

	err = conn.WriteMessage(websocket.TextMessage, []byte("HELLO"))
	assert.Nil(t, err)

	conn.Close()
}
