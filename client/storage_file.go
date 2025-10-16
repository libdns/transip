package client

import (
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// NewTokenFileStorage will use given directory for storing token sessions
// and try to create when calling this function
//
// storage, err := NewTokenFileStorage(filepath.Join(os.TempDir(), "transip"))
func NewTokenFileStorage(root string) (Storage, error) {

	err := os.MkdirAll(root, 0700)

	if err != nil {
		return nil, err
	}

	return &storageFile{root: root, items: make(map[string]Token)}, nil
}

type storageFile struct {
	root  string
	mutex sync.RWMutex
	items map[string]Token
}

func (s *storageFile) Set(key string, token Token) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items[key] = token

	file, err := os.OpenFile(filepath.Join(s.root, key), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return err
	}

	if err := gob.NewEncoder(file).Encode(token); err != nil {

		defer func() {
			_ = os.Remove(file.Name())
		}()

		return err
	}

	defer file.Close()

	return nil
}

func (s *storageFile) Get(key string) (Token, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if item, ok := s.items[key]; ok {
		return item, nil
	}

	file, err := os.Open(filepath.Join(s.root, key))

	if err != nil {

		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	defer file.Close()

	var token Token = new(token)

	if err := gob.NewDecoder(file).Decode(&token); err != nil {
		return nil, err
	}

	s.items[key] = token

	return token, nil
}
