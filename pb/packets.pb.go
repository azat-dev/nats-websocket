// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pb/packets.proto

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	pb/packets.proto

It has these top-level messages:
	InputMessage
	Command
	SendMessageToDeviceParams
	SendMessageToAllUserDevicesParams
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type MessageType int32

const (
	MessageType_TEXT   MessageType = 0
	MessageType_BINARY MessageType = 1
)

var MessageType_name = map[int32]string{
	0: "TEXT",
	1: "BINARY",
}
var MessageType_value = map[string]int32{
	"TEXT":   0,
	"BINARY": 1,
}

func (x MessageType) String() string {
	return proto.EnumName(MessageType_name, int32(x))
}
func (MessageType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Command_Method int32

const (
	Command_SEND_MESSAGE_TO_ALL_USER_DEVICES Command_Method = 0
	Command_SEND_MESSAGE_TO_DEVICE           Command_Method = 1
)

var Command_Method_name = map[int32]string{
	0: "SEND_MESSAGE_TO_ALL_USER_DEVICES",
	1: "SEND_MESSAGE_TO_DEVICE",
}
var Command_Method_value = map[string]int32{
	"SEND_MESSAGE_TO_ALL_USER_DEVICES": 0,
	"SEND_MESSAGE_TO_DEVICE":           1,
}

func (x Command_Method) String() string {
	return proto.EnumName(Command_Method_name, int32(x))
}
func (Command_Method) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

type InputMessage struct {
	Type      MessageType `protobuf:"varint,1,opt,name=type,enum=pb.MessageType" json:"type,omitempty"`
	InputTime int64       `protobuf:"varint,2,opt,name=inputTime" json:"inputTime,omitempty"`
	UserId    string      `protobuf:"bytes,3,opt,name=userId" json:"userId,omitempty"`
	DeviceId  string      `protobuf:"bytes,4,opt,name=deviceId" json:"deviceId,omitempty"`
	Body      []byte      `protobuf:"bytes,5,opt,name=body,proto3" json:"body,omitempty"`
}

func (m *InputMessage) Reset()                    { *m = InputMessage{} }
func (m *InputMessage) String() string            { return proto.CompactTextString(m) }
func (*InputMessage) ProtoMessage()               {}
func (*InputMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *InputMessage) GetType() MessageType {
	if m != nil {
		return m.Type
	}
	return MessageType_TEXT
}

func (m *InputMessage) GetInputTime() int64 {
	if m != nil {
		return m.InputTime
	}
	return 0
}

func (m *InputMessage) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *InputMessage) GetDeviceId() string {
	if m != nil {
		return m.DeviceId
	}
	return ""
}

func (m *InputMessage) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

type Command struct {
	Method Command_Method `protobuf:"varint,1,opt,name=method,enum=pb.Command_Method" json:"method,omitempty"`
	// Types that are valid to be assigned to Params:
	//	*Command_SendToAllUserDevices
	//	*Command_SendToDevice
	Params isCommand_Params `protobuf_oneof:"params"`
}

func (m *Command) Reset()                    { *m = Command{} }
func (m *Command) String() string            { return proto.CompactTextString(m) }
func (*Command) ProtoMessage()               {}
func (*Command) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type isCommand_Params interface {
	isCommand_Params()
}

type Command_SendToAllUserDevices struct {
	SendToAllUserDevices *SendMessageToAllUserDevicesParams `protobuf:"bytes,2,opt,name=sendToAllUserDevices,oneof"`
}
type Command_SendToDevice struct {
	SendToDevice *SendMessageToDeviceParams `protobuf:"bytes,3,opt,name=sendToDevice,oneof"`
}

func (*Command_SendToAllUserDevices) isCommand_Params() {}
func (*Command_SendToDevice) isCommand_Params()         {}

func (m *Command) GetParams() isCommand_Params {
	if m != nil {
		return m.Params
	}
	return nil
}

func (m *Command) GetMethod() Command_Method {
	if m != nil {
		return m.Method
	}
	return Command_SEND_MESSAGE_TO_ALL_USER_DEVICES
}

func (m *Command) GetSendToAllUserDevices() *SendMessageToAllUserDevicesParams {
	if x, ok := m.GetParams().(*Command_SendToAllUserDevices); ok {
		return x.SendToAllUserDevices
	}
	return nil
}

func (m *Command) GetSendToDevice() *SendMessageToDeviceParams {
	if x, ok := m.GetParams().(*Command_SendToDevice); ok {
		return x.SendToDevice
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Command) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Command_OneofMarshaler, _Command_OneofUnmarshaler, _Command_OneofSizer, []interface{}{
		(*Command_SendToAllUserDevices)(nil),
		(*Command_SendToDevice)(nil),
	}
}

func _Command_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Command)
	// params
	switch x := m.Params.(type) {
	case *Command_SendToAllUserDevices:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.SendToAllUserDevices); err != nil {
			return err
		}
	case *Command_SendToDevice:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.SendToDevice); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Command.Params has unexpected type %T", x)
	}
	return nil
}

func _Command_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Command)
	switch tag {
	case 2: // params.sendToAllUserDevices
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(SendMessageToAllUserDevicesParams)
		err := b.DecodeMessage(msg)
		m.Params = &Command_SendToAllUserDevices{msg}
		return true, err
	case 3: // params.sendToDevice
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(SendMessageToDeviceParams)
		err := b.DecodeMessage(msg)
		m.Params = &Command_SendToDevice{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Command_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Command)
	// params
	switch x := m.Params.(type) {
	case *Command_SendToAllUserDevices:
		s := proto.Size(x.SendToAllUserDevices)
		n += proto.SizeVarint(2<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Command_SendToDevice:
		s := proto.Size(x.SendToDevice)
		n += proto.SizeVarint(3<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type SendMessageToDeviceParams struct {
	MessageType MessageType `protobuf:"varint,1,opt,name=messageType,enum=pb.MessageType" json:"messageType,omitempty"`
	DeviceId    string      `protobuf:"bytes,2,opt,name=deviceId" json:"deviceId,omitempty"`
	Message     []byte      `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (m *SendMessageToDeviceParams) Reset()                    { *m = SendMessageToDeviceParams{} }
func (m *SendMessageToDeviceParams) String() string            { return proto.CompactTextString(m) }
func (*SendMessageToDeviceParams) ProtoMessage()               {}
func (*SendMessageToDeviceParams) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *SendMessageToDeviceParams) GetMessageType() MessageType {
	if m != nil {
		return m.MessageType
	}
	return MessageType_TEXT
}

func (m *SendMessageToDeviceParams) GetDeviceId() string {
	if m != nil {
		return m.DeviceId
	}
	return ""
}

func (m *SendMessageToDeviceParams) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

type SendMessageToAllUserDevicesParams struct {
	MessageType   MessageType `protobuf:"varint,1,opt,name=messageType,enum=pb.MessageType" json:"messageType,omitempty"`
	UserId        string      `protobuf:"bytes,2,opt,name=userId" json:"userId,omitempty"`
	ExcludeDevice string      `protobuf:"bytes,3,opt,name=excludeDevice" json:"excludeDevice,omitempty"`
	Message       []byte      `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
}

func (m *SendMessageToAllUserDevicesParams) Reset()         { *m = SendMessageToAllUserDevicesParams{} }
func (m *SendMessageToAllUserDevicesParams) String() string { return proto.CompactTextString(m) }
func (*SendMessageToAllUserDevicesParams) ProtoMessage()    {}
func (*SendMessageToAllUserDevicesParams) Descriptor() ([]byte, []int) {
	return fileDescriptor0, []int{3}
}

func (m *SendMessageToAllUserDevicesParams) GetMessageType() MessageType {
	if m != nil {
		return m.MessageType
	}
	return MessageType_TEXT
}

func (m *SendMessageToAllUserDevicesParams) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *SendMessageToAllUserDevicesParams) GetExcludeDevice() string {
	if m != nil {
		return m.ExcludeDevice
	}
	return ""
}

func (m *SendMessageToAllUserDevicesParams) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

func init() {
	proto.RegisterType((*InputMessage)(nil), "pb.InputMessage")
	proto.RegisterType((*Command)(nil), "pb.Command")
	proto.RegisterType((*SendMessageToDeviceParams)(nil), "pb.SendMessageToDeviceParams")
	proto.RegisterType((*SendMessageToAllUserDevicesParams)(nil), "pb.SendMessageToAllUserDevicesParams")
	proto.RegisterEnum("pb.MessageType", MessageType_name, MessageType_value)
	proto.RegisterEnum("pb.Command_Method", Command_Method_name, Command_Method_value)
}

func init() { proto.RegisterFile("pb/packets.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 425 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x53, 0xd1, 0x8e, 0x93, 0x40,
	0x14, 0xed, 0xb0, 0xc8, 0xb6, 0xb7, 0x55, 0xc9, 0x8d, 0xd9, 0xe0, 0x46, 0x13, 0x64, 0xd7, 0xa4,
	0xd9, 0x07, 0x8c, 0xf5, 0x0b, 0xd8, 0x96, 0x28, 0x66, 0xbb, 0x9a, 0x81, 0x35, 0x1a, 0x1f, 0x08,
	0x74, 0x26, 0xda, 0x58, 0x60, 0xc2, 0x50, 0x63, 0xdf, 0xfd, 0x09, 0x3f, 0x42, 0xbf, 0xd1, 0x74,
	0x60, 0xb7, 0x50, 0xdd, 0x98, 0xf8, 0xc6, 0x3d, 0xe7, 0xdc, 0x7b, 0xcf, 0xcc, 0x19, 0xc0, 0x14,
	0xe9, 0x33, 0x91, 0x2c, 0xbe, 0xf0, 0x4a, 0xba, 0xa2, 0x2c, 0xaa, 0x02, 0x35, 0x91, 0x3a, 0x3f,
	0x08, 0x8c, 0x82, 0x5c, 0xac, 0xab, 0x39, 0x97, 0x32, 0xf9, 0xc4, 0xf1, 0x04, 0xf4, 0x6a, 0x23,
	0xb8, 0x45, 0x6c, 0x32, 0xbe, 0x37, 0xb9, 0xef, 0x8a, 0xd4, 0x6d, 0xa8, 0x68, 0x23, 0x38, 0x55,
	0x24, 0x3e, 0x82, 0xc1, 0x72, 0xdb, 0x14, 0x2d, 0x33, 0x6e, 0x69, 0x36, 0x19, 0x1f, 0xd0, 0x1d,
	0x80, 0x47, 0x60, 0xac, 0x25, 0x2f, 0x03, 0x66, 0x1d, 0xd8, 0x64, 0x3c, 0xa0, 0x4d, 0x85, 0xc7,
	0xd0, 0x67, 0xfc, 0xeb, 0x72, 0xc1, 0x03, 0x66, 0xe9, 0x8a, 0xb9, 0xa9, 0x11, 0x41, 0x4f, 0x0b,
	0xb6, 0xb1, 0xee, 0xd8, 0x64, 0x3c, 0xa2, 0xea, 0xdb, 0xf9, 0xa9, 0xc1, 0xe1, 0xb4, 0xc8, 0xb2,
	0x24, 0x67, 0x78, 0x06, 0x46, 0xc6, 0xab, 0xcf, 0x05, 0x6b, 0x8c, 0xe1, 0xd6, 0x58, 0x43, 0xba,
	0x73, 0xc5, 0xd0, 0x46, 0x81, 0x1f, 0xe1, 0x81, 0xe4, 0x39, 0x8b, 0x0a, 0x6f, 0xb5, 0xba, 0x92,
	0xbc, 0x9c, 0xa9, 0x25, 0x52, 0x19, 0x1d, 0x4e, 0x9e, 0x6e, 0x3b, 0x43, 0x9e, 0xb3, 0xeb, 0x63,
	0xed, 0xc9, 0xde, 0x26, 0x65, 0x92, 0xc9, 0x57, 0x3d, 0xfa, 0xd7, 0x21, 0x38, 0x85, 0x51, 0x8d,
	0xd7, 0x80, 0x3a, 0xe2, 0x70, 0xf2, 0xf8, 0x8f, 0xa1, 0x35, 0x7d, 0x33, 0xac, 0xd3, 0xe4, 0xbc,
	0x06, 0xa3, 0xf6, 0x8c, 0xa7, 0x60, 0x87, 0xfe, 0xe5, 0x2c, 0x9e, 0xfb, 0x61, 0xe8, 0xbd, 0xf4,
	0xe3, 0xe8, 0x4d, 0xec, 0x5d, 0x5c, 0xc4, 0x57, 0xa1, 0x4f, 0xe3, 0x99, 0xff, 0x2e, 0x98, 0xfa,
	0xa1, 0xd9, 0xc3, 0x63, 0x38, 0xda, 0x57, 0xd5, 0xa4, 0x49, 0xce, 0xfb, 0x60, 0x08, 0xb5, 0xc5,
	0xf9, 0x4e, 0xe0, 0xe1, 0xad, 0x1e, 0xf0, 0x39, 0x0c, 0xb3, 0x5d, 0x90, 0xb7, 0xe5, 0xdb, 0xd6,
	0x74, 0x02, 0xd3, 0xf6, 0x02, 0xb3, 0xe0, 0xb0, 0x91, 0xaa, 0x2b, 0x18, 0xd1, 0xeb, 0xd2, 0xf9,
	0x45, 0xe0, 0xc9, 0x3f, 0xef, 0xf7, 0x7f, 0xec, 0xec, 0xde, 0x95, 0xd6, 0x79, 0x57, 0xa7, 0x70,
	0x97, 0x7f, 0x5b, 0xac, 0xd6, 0x8c, 0xb7, 0x32, 0x19, 0xd0, 0x2e, 0xd8, 0x36, 0xac, 0x77, 0x0c,
	0x9f, 0x9d, 0xc0, 0xb0, 0xb5, 0x13, 0xfb, 0xa0, 0x47, 0xfe, 0xfb, 0xc8, 0xec, 0x21, 0x80, 0x71,
	0x1e, 0x5c, 0x7a, 0xf4, 0x83, 0x49, 0x52, 0x43, 0xfd, 0x33, 0x2f, 0x7e, 0x07, 0x00, 0x00, 0xff,
	0xff, 0xef, 0xbd, 0x2d, 0x8a, 0x47, 0x03, 0x00, 0x00,
}
