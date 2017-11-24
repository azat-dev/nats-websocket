package nats_websocket

import (
	"sync"
)

//const NOT_LOGGED_LIFE_TIME = 5 * time.Second
//const PING_TIMEOUT = 10 * time.Minute

type ConnectionsStats struct {
	NumberOfUsers                int
	NumberOfDevices              int
	NumberOfNotLoggedConnections int
}

type ConnectionsStorage struct {
	mutex                        sync.RWMutex
	connectionsById              map[ConnectionId]*Connection
	connectionsByUserId          map[UserId]map[DeviceId]*Connection
	connectionsByDeviceId        map[DeviceId]*Connection // one connection per device
	numberOfNotLoggedConnections int
}

func NewConnectionsStorage() *ConnectionsStorage {
	return &ConnectionsStorage{
		mutex:                        sync.RWMutex{},
		connectionsById:              make(map[ConnectionId]*Connection),
		connectionsByUserId:          make(map[UserId]map[DeviceId]*Connection),
		connectionsByDeviceId:        make(map[DeviceId]*Connection),
		numberOfNotLoggedConnections: 0,
	}
}

func (s *ConnectionsStorage) AddNewConnection(connection *Connection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.numberOfNotLoggedConnections++
	s.connectionsById[connection.id] = connection
}

func (s *ConnectionsStorage) OnLogin(connection *Connection) *Connection {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, userId, deviceId := connection.GetInfo()
	if userId == "" {
		return nil
	}

	s.numberOfNotLoggedConnections--

	deviceConnectionBefore := s.connectionsByDeviceId[connection.deviceId]
	if deviceConnectionBefore != nil {
		s.removeConnection(deviceConnectionBefore)
	}
	s.connectionsByDeviceId[deviceId] = connection

	userConnections := s.connectionsByUserId[userId]
	if userConnections == nil {
		userConnections = make(map[DeviceId]*Connection)
		s.connectionsByUserId[userId] = userConnections
	}
	userConnections[deviceId] = connection

	return deviceConnectionBefore
}

func (s *ConnectionsStorage) RemoveConnection(connection *Connection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.removeConnection(connection)
}

func (s *ConnectionsStorage) removeConnection(connection *Connection) {

	connectionId, userId, deviceId := connection.GetInfo()

	connectionBefore := s.connectionsById[connectionId]
	if connectionBefore == nil {
		return
	}

	delete(s.connectionsById, connectionId)

	if userId == "" {
		s.numberOfNotLoggedConnections--
		return
	}

	userConnections := s.connectionsByUserId[userId]
	if userConnections != nil {
		delete(userConnections, deviceId)
		if len(userConnections) == 0 {
			delete(s.connectionsByUserId, userId)
		}
	}

	deviceConnection := s.connectionsByDeviceId[deviceId]
	if deviceConnection != nil {
		delete(s.connectionsByDeviceId, deviceId)
	}
}

func (s *ConnectionsStorage) GetUserConnections(userId UserId) map[DeviceId]*Connection {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.connectionsByUserId[userId]
}

func (s *ConnectionsStorage) GetDeviceConnection(deviceId DeviceId) *Connection {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.connectionsByDeviceId[deviceId]
}

func (s *ConnectionsStorage) GetConnectionById(connectionId ConnectionId) *Connection {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.connectionsById[connectionId]
}

func (s *ConnectionsStorage) GetStats() ConnectionsStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := ConnectionsStats{
		NumberOfDevices:              len(s.connectionsByDeviceId),
		NumberOfUsers:                len(s.connectionsByUserId),
		NumberOfNotLoggedConnections: s.numberOfNotLoggedConnections,
	}

	return stats
}

func (s *ConnectionsStorage) RemoveIf(condition func(con *Connection) bool, afterRemove func(con *Connection)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for id, connection := range s.connectionsById {
		if condition(connection) {

			_, userId, deviceId := connection.GetInfo()

			delete(s.connectionsById, id)

			if deviceId != "" {
				delete(s.connectionsByDeviceId, deviceId)
			}

			if userId != "" {
				userConnections := s.connectionsByUserId[userId]
				if userConnections != nil {
					if len(userConnections) == 1 {
						delete(s.connectionsByUserId, userId)
					} else {
						delete(userConnections, deviceId)
					}
				}
			}

			afterRemove(connection)
		}
	}
}
