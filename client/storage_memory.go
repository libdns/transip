package client

import (
	"sync"
)

func NewTokenMemoryStorage() Storage {
	return &storageMemory{items: make(map[string]Token)}
}

type storageMemory struct {
	mutex sync.RWMutex
	items map[string]Token
}

func (s *storageMemory) Set(key string, token Token) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.items[key] = token
	return nil
}

func (s *storageMemory) Get(key string) (Token, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if item, ok := s.items[key]; ok {
		return item, nil
	}
	return nil, nil
}
