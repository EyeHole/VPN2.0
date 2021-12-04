package storage

import (
	"net"
	"sync"
)

type Storage struct {
	Connections map[string]net.Conn
	Mu          *sync.Mutex
}


func SetStorage() *Storage {
	storage := Storage{
		Connections: map[string]net.Conn{},
		Mu:          &sync.Mutex{},
	}
	return &storage
}


func (s *Storage) AddTun(name string, conn net.Conn) {
	s.Mu.Lock()
	s.Connections[name] = conn
	s.Mu.Unlock()
}

func (s *Storage) GetTun(name string) (net.Conn, bool) {
	s.Mu.Lock()
	conn, err := s.Connections[name]
	s.Mu.Unlock()
	return conn, err
}
