package client

import (
	"context"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type token struct {
	value   string
	payload *tokenPayload
}

func (t *token) GobEncode() ([]byte, error) {
	return []byte(t.value), nil
}

func (t *token) GobDecode(data []byte) error {
	t.value = string(data)

	payload, err := getPayload(t.value)

	if err != nil {
		return err
	}

	t.payload = payload

	return nil
}

func (t *token) String() string {
	return t.value
}

func (t *token) IsExpired() bool {
	if nil == t.payload {
		return true
	}
	return time.Unix(t.payload.Expires, 0).Before(time.Now())
}

func (t *token) ReadOnly() bool {
	if nil == t.payload {
		return true
	}
	return t.payload.ReadOnly
}

func (t *token) GlobalKey() bool {
	if nil == t.payload {
		return true
	}
	return t.payload.GlobalKey
}

type Token interface {
	String() string
	IsExpired() bool
	ReadOnly() bool
	GlobalKey() bool

	gob.GobDecoder
}

type tokenPayload struct {
	NotBefore int64 `json:"nbf"`
	IssuedAt  int64 `json:"iat"`
	Expires   int64 `json:"exp"`
	ReadOnly  bool  `json:"ro"`
	GlobalKey bool  `json:"gk"`
}

func NewToken(x string) (Token, error) {

	payload, err := getPayload(x)

	if nil != err {
		return nil, err
	}

	return &token{value: x, payload: payload}, nil
}

func getPayload(x string) (*tokenPayload, error) {

	var parts = strings.SplitN(x, ".", 3)

	if len(parts) != 3 {
		return nil, errors.New("invalid jwt format")
	}

	buf, err := base64.RawURLEncoding.DecodeString(parts[1])

	if err != nil {
		return nil, err
	}

	var payload *tokenPayload

	if err := json.Unmarshal(buf, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

type TokenFetcher func(ctx context.Context, config Config) (string, error)
