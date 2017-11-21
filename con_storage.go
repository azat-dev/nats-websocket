package nats_websocket

import (
	"log"
	"time"
)

import (
	"sync"
)

const NOT_LOGGED_LIFE_TIME = 5 * time.Second
const PING_TIMEOUT = 10 * time.Minute

type ConnectionsStorage struct {
	mutex                     sync.RWMutex
	connectionsById           map[int64]*Connection
	connectionsByUserId       map[string]map[int64]*Connection
	connectionsByDeviceId     map[string]*Connection // one connection per device
	numberOfClosedConnections int
}

func NewConnectionsStorage() *ConnectionsStorage {
	return &ConnectionsStorage{
		mutex:                     sync.RWMutex{},
		connectionsById:           make(map[int64]*Connection),
		connectionsByUserId:       make(map[string]map[int64]*Connection),
		connectionsByDeviceId:     make(map[string]*Connection),
		numberOfClosedConnections: 0,
	}
}

func (s *ConnectionsStorage) AddNewConnection(connection *Connection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.connectionsById[connection.id] = connection
}

func (s *ConnectionsStorage) OnLogin(connection *Connection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	connectionId, deviceId, userId := connection.GetInfo()
	if userId == nil || *userId == "" {
		return
	}

	connectionBefore := s.connectionsByDeviceId[*connection.deviceId]
	if connectionBefore != nil {
		s.removeConnection(connectionBefore)
	}
	s.connectionsByDeviceId[*deviceId] = connection

	connections := s.connectionsByUserId[*userId]
	if connections == nil {
		connections = make(map[int64]*Connection)
		s.connectionsByUserId[*userId] = connections
	}
	connections[connectionId] = connection

	if connections == nil {
		connections = make(map[int64]*Connection)

	}
	connections[connectionId] = connection
}

func (s *ConnectionsStorage) RemoveConnection(connection *Connection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.removeConnection(connection)
}

func (s *ConnectionsStorage) removeConnection(connection *Connection) {
	defer func() {
		connection.SetClosed()
	}()

	connectionId, deviceId, userId := connection.GetInfo()

	delete(s.connectionsById, connectionId)

	if userId == nil {
		return
	}

	userChannels := s.connectionsByUserId[*userId]
	if userChannels != nil {
		delete(userChannels, connectionId)
		if len(userChannels) == 0 {
			delete(s.connectionsByUserId, *userId)
		}
	}

	connectionBefore := s.connectionsByDeviceId[*deviceId]
	if connectionBefore != nil {
		delete(s.connectionsByDeviceId, *deviceId)
	}

	s.numberOfClosedConnections++
}

func (s *ConnectionsStorage) GetUserConnections(userId string) map[int64]*Connection {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.connectionsByUserId[userId]
}

func (s *ConnectionsStorage) GetDeviceConnection(deviceId string) *Connection {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.connectionsByDeviceId[deviceId]
}

func (s *ConnectionsStorage) GetConnectionById(connectionId int64) *Connection {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.connectionsById[connectionId]
}

func (s *ConnectionsStorage) cleanConnections() {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Println("Start clean")

	for _, connection := range s.connectionsById {

		if connection.userId == nil && (time.Since(connection.startTime) > NOT_LOGGED_LIFE_TIME) {

			log.Println("Delete connection ", connection.id)
			connection.Close("NotLoggedIn")
			//TODO: is it call onClose?
			s.removeConnection(connection)

		} else if time.Since(connection.lastMessageAt) > PING_TIMEOUT {

			connection.Close("PingTimeout")
			s.removeConnection(connection)
		}
	}
}

func (s *ConnectionsStorage) startCleanNotLoggedChannels(stopSignal chan bool) {

	ticker := time.NewTicker(PING_TIMEOUT)

	for {
		select {
		case <-ticker.C:
			s.cleanConnections()
		case <-stopSignal:
			ticker.Stop()
			return
		}
	}
}
