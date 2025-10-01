package client

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

type Storage interface {
	Set(key string, token Token) error
	Get(key string) (Token, error)
}

// StorageKey will generate a unique key within the context a key is requested
func StorageKey(login string, globalKey, readOnly bool) (string, error) {
	var hasher = sha1.New()

	if _, err := fmt.Fprintf(hasher, "login:%s|gk:%t|ro:%t", login, globalKey, readOnly); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
