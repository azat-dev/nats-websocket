package nats_websocket

type Config struct {
	ListenInterface string `json:"listenInterface"`
	UrlPattern      string `json:"urlPattern"`
	PacketFormat    string `json:"packetFormat"`
	Timeout         int64  `json:"timeout"`
	JwtSecret       string `json:"jwtSecret"`
	NumberOfWorkers int    `json:"numberOfWorkers"`

	NatsAddress  string `json:"natsAddress"`
	NatsPoolSize int    `json:"natsPoolSize"`
}
