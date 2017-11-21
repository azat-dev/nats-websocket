package main

import (
	"github.com/akaumov/nats-websocket"
)

func main() {

	server := nats_websocket.New(&nats_websocket.Config{
		NatsPoolSize:    10,
		NatsAddress:     "nats://localhost:32770",
		UrlPattern:      "/",
		ListenInterface: "localhost:8080",
		PacketFormat:    "json",
		Timeout:         30000,
		JwtSecret:       "123456",
		NumberOfWorkers: 200,
	})

	server.Start()
}
