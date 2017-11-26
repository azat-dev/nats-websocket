package nats_websocket

import (
	"github.com/stretchr/testify/assert"
	. "testing"
)

func TestCountOfNotLoggedConnections(t *T) {

	storage := NewConnectionsStorage()

	stats := storage.GetStats()
	assert.Equal(t, 0, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	con1 := NewConnection(ConnectionId(1), nil)
	con2 := NewConnection(ConnectionId(2), nil)

	//add two connects and remove them
	storage.AddNewConnection(con1)

	stats = storage.GetStats()
	assert.Equal(t, 1, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	storage.AddNewConnection(con2)

	stats = storage.GetStats()
	assert.Equal(t, 2, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	storage.RemoveConnection(con1)

	stats = storage.GetStats()
	assert.Equal(t, 1, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	storage.RemoveConnection(con2)

	stats = storage.GetStats()
	assert.Equal(t, 0, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	storage.RemoveConnection(con2)

	stats = storage.GetStats()
	assert.Equal(t, 0, stats.NumberOfNotLoggedConnections, "number of not logged connections shouldn't go lower then 0")
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)
}

func TestLogin(t *T) {

	storage := NewConnectionsStorage()

	stats := storage.GetStats()
	assert.Equal(t, 0, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	con1 := NewConnection(ConnectionId(1), nil)
	con2 := NewConnection(ConnectionId(2), nil)
	con3 := NewConnection(ConnectionId(3), nil)
	con4 := NewConnection(ConnectionId(4), nil)

	assert.Equal(t, 0, storage.numberOfNotLoggedConnections)

	storage.AddNewConnection(con1)

	stats = storage.GetStats()
	assert.Equal(t, 1, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	storage.AddNewConnection(con2)

	stats = storage.GetStats()
	assert.Equal(t, 2, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	storage.AddNewConnection(con3)

	stats = storage.GetStats()
	assert.Equal(t, 3, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	storage.AddNewConnection(con4)

	stats = storage.GetStats()
	assert.Equal(t, 4, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 0, stats.NumberOfUsers)
	assert.Equal(t, 0, stats.NumberOfDevices)

	con1.userId = UserId("user1")
	con1.deviceId = DeviceId("user1Device1")

	deviceConBefore := storage.OnLogin(con1)
	assert.Nil(t, deviceConBefore)

	stats = storage.GetStats()
	assert.Equal(t, 3, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 1, stats.NumberOfUsers)
	assert.Equal(t, 1, stats.NumberOfDevices)

	con2.userId = UserId("user2")
	con2.deviceId = DeviceId("user2Device1")

	deviceConBefore = storage.OnLogin(con2)
	assert.Nil(t, deviceConBefore)

	stats = storage.GetStats()
	assert.Equal(t, 2, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 2, stats.NumberOfUsers)
	assert.Equal(t, 2, stats.NumberOfDevices)

	con3.userId = UserId("user2")
	con3.deviceId = DeviceId("user2Device1")

	deviceConBefore = storage.OnLogin(con3)
	assert.NotNil(t, deviceConBefore)
	assert.Equal(t, con2.id, deviceConBefore.id)

	stats = storage.GetStats()
	assert.Equal(t, 1, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 2, stats.NumberOfUsers)
	assert.Equal(t, 2, stats.NumberOfDevices)

	con4.userId = UserId("user2")
	con4.deviceId = DeviceId("user2Device2")

	deviceConBefore = storage.OnLogin(con4)
	assert.Nil(t, deviceConBefore)

	stats = storage.GetStats()
	assert.Equal(t, 0, stats.NumberOfNotLoggedConnections)
	assert.Equal(t, 2, stats.NumberOfUsers)
	assert.Equal(t, 3, stats.NumberOfDevices)
}

func TestRetriveConnections(t *T) {

	storage := NewConnectionsStorage()

	con1 := NewConnection(ConnectionId(1), nil)
	con2 := NewConnection(ConnectionId(2), nil)
	con3 := NewConnection(ConnectionId(3), nil)

	storage.AddNewConnection(con1)
	storage.AddNewConnection(con2)
	storage.AddNewConnection(con3)

	con1.userId = UserId("user1")
	con1.deviceId = DeviceId("user1Device1")
	storage.OnLogin(con1)

	con2.userId = UserId("user2")
	con2.deviceId = DeviceId("user2Device1")
	storage.OnLogin(con2)

	con3.userId = UserId("user2")
	con3.deviceId = DeviceId("user2Device2")
	storage.OnLogin(con3)

	assert.Equal(t, ConnectionId(1), storage.GetConnectionById(1).id)
	assert.Equal(t, ConnectionId(2), storage.GetConnectionById(2).id)
	assert.Equal(t, ConnectionId(3), storage.GetConnectionById(3).id)

	assert.Equal(t, DeviceId("user1Device1"), storage.GetDeviceConnection("user1Device1").deviceId)
	assert.Equal(t, DeviceId("user2Device1"), storage.GetDeviceConnection("user2Device1").deviceId)
	assert.Equal(t, DeviceId("user2Device2"), storage.GetDeviceConnection("user2Device2").deviceId)

	user1Connections := storage.GetUserConnections(UserId("user1"))
	assert.NotNil(t, user1Connections)
	assert.Equal(t, 1, len(user1Connections))
	assert.Equal(t, con1, user1Connections[DeviceId("user1Device1")])

	user2Connections := storage.GetUserConnections(UserId("user2"))
	assert.NotNil(t, user2Connections)
	assert.Equal(t, 2, len(user2Connections))
	assert.Equal(t, con2, user2Connections[DeviceId("user2Device1")])
	assert.Equal(t, con3, user2Connections[DeviceId("user2Device2")])
}
