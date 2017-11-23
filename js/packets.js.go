package js

import "encoding/json"

type MessageType int32

const (
	TEXT   MessageType = 0
	BINARY MessageType = 1
)

type InputMessage struct {
	Type       MessageType `json:"type"`
	InputTime  int64       `json:"inputTime"`
	UserId     string      `json:"userId"`
	DeviceId   string      `json:"deviceId"`
	Host       string      `json:"host"`
	RemoteAddr string      `json:"remoteAddr"`
	Body       []byte      `json:"data"`
}

type Method int32

const (
	SEND_MESSAGE_TO_ALL_USER_DEVICES Method = 0
	SEND_MESSAGE_TO_DEVICE           Method = 1
)

type Command struct {
	Method Method          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type SendMessageToDeviceParams struct {
	MessageType MessageType `json:"messageType"`
	DeviceId    string      `json:"deviceId"`
	Message     []byte      `json:"message"`
}

type SendMessageToAllUserDevicesParams struct {
	MessageType   MessageType `json:"messageType"`
	UserId        string      `json:"userId"`
	ExcludeDevice string      `json:"excludeDevice"`
	Message       []byte      `json:"message"`
}
