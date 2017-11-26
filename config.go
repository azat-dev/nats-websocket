package nats_websocket

type Config struct {
	ListenInterface string `json:"listenInterface"`
	UrlPattern      string `json:"urlPattern"`
	PacketFormat    string `json:"packetFormat"`
	JwtSecret       string `json:"jwtSecret"`

	NatsAddress       string `json:"natsAddress"`
	NatsPoolSize      int    `json:"natsPoolSize"`
	NatsListenSubject string `json:"natsListenSubject"`
	NatsOutputSubject string `json:"natsOutputSubject"`
}
