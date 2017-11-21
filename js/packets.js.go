package js

type InputMessageType int32

const (
	TEXT   = 0
	BINARY = 1
)

type InputMessage struct {
	Type       InputMessageType `json:"type"`
	InputTime  int64            `json:"inputTime"`
	UserId     string           `json:"userId"`
	DeviceId   string           `json:"deviceId"`
	Host       string           `json:"host"`
	RemoteAddr string           `json:"remoteAddr"`
	Body       []byte           `json:"data"`
}
