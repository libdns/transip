package client

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"io"
	fallback "math/rand/v2"
	"net/url"
)

type ExpirationTime string

const (
	ExpirationTime1Hour  ExpirationTime = "1 hour"
	ExpirationTime2Hour  ExpirationTime = "120 minutes"
	ExpirationTime48Hour ExpirationTime = "48 hours"
	ExpirationTime1Day   ExpirationTime = "1 day"
	ExpirationTime15Day  ExpirationTime = "15 days"
	ExpirationTime1Week  ExpirationTime = "1 week"
	ExpirationTime4Week  ExpirationTime = "4 weeks"
)

type Config interface {
	Login() string
	ReadOnly() bool
	GlobalKey() bool

	StorageKey() string
	GetBaseUri() *url.URL
	GetPrivateKey() (*rsa.PrivateKey, error)
	GetDebug() io.Writer
}

type ConfigLabel interface {
	Label() string
}

type ConfigExpirationTime interface {
	ExpirationTime() ExpirationTime
}

type ConfigNonce interface {
	Nonce() string
}

func random(size int) (s string) {
	var buf = make([]byte, size)

	defer func() {
		if r := recover(); r != nil {
			var c = fallback.NewChaCha8([32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456")))
			_, _ = c.Read(buf)
			s = hex.EncodeToString(buf)
		}
	}()

	// will panic, which should almost never do but in rare
	// it does, we will fallback to math package
	_, _ = rand.Read(buf)

	return hex.EncodeToString(buf)
}
