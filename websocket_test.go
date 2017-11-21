package nats_websocket

import (
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"sync"
	. "testing"
	"time"
)

var dialer = websocket.Dialer{}

func TestConnections(t *T) {

	natsWebsocket := New(&Config{
		ListenInterface: "localhost:8090",
		JwtSecret:       "123456",
		Timeout:         30000,
		UrlPattern:      "/",
		NatsAddress:     "nats://localhost:32770",
		NatsPoolSize:    10,
		PacketFormat:    "json",
		NumberOfWorkers: 20,
	})

	go func() {
		natsWebsocket.Start()
	}()

	time.Sleep(2 * time.Second)

	wg := sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go startTestConnection(t, i%2 == 0, &wg)
	}

	wg.Wait()
}

func getToken() []byte {
	claims := jws.Claims{}

	claims.Set("userId", uuid.NewV4().String())
	claims.Set("deviceId", uuid.NewV4().String())

	token := jws.NewJWT(claims, crypto.SigningMethodHS256)
	serialized, _ := token.Serialize([]byte("123456"))

	return serialized
}

func startTestConnection(t *T, sendWrongToken bool, wg *sync.WaitGroup) {

	defer wg.Done()

	conn, _, err := dialer.Dial("ws://localhost:8090/", nil)
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
		token = getToken()
	}

	//test login
	err = conn.WriteMessage(websocket.TextMessage, []byte("login>:"+string(token)))
	assert.Nil(t, err)

	messageType, message, err := conn.ReadMessage()
	assert.Nil(t, err)
	assert.Equal(t, messageType, websocket.TextMessage)

	if sendWrongToken {
		assert.Equal(t, string(message), "login>:"+"error")
	} else {
		assert.Equal(t, string(message), "login>:"+"ok")
	}

	//test ping
	err = conn.WriteMessage(websocket.TextMessage, []byte("ping"))
	assert.Nil(t, err)

	messageType, message, err = conn.ReadMessage()

	if sendWrongToken {
		assert.NotNil(t, err)
	} else {

		assert.Nil(t, err)
		assert.Equal(t, messageType, websocket.TextMessage)
		assert.Equal(t, string(message), "pong")
	}

	conn.Close()
}
