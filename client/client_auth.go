package client

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
)

type AuthRequest struct {
	Login          string         `json:"login"`
	Nonce          string         `json:"nonce"`
	Label          string         `json:"label,omitempty"`
	ReadOnly       bool           `json:"read_only"`
	ExpirationTime ExpirationTime `json:"expiration_time"`
	GlobalKey      bool           `json:"global_key"`
}

func (c *client) Authorize(ctx context.Context, config Config) (string, error) {

	request, err := c.makeAuthorizeRequest(ctx, config)

	resp, err := c.client.Do(request)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var data struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Token, nil
}

func (c *client) makeAuthorizeRequest(ctx context.Context, config Config) (*http.Request, error) {

	var payload = NewAuthRequest(config)
	var hasher = sha512.New()
	var body = new(bytes.Buffer)
	var writer = io.MultiWriter(hasher, body)

	key, err := config.GetPrivateKey()

	if err != nil {
		return nil, err
	}

	if err := json.NewEncoder(writer).Encode(payload); err != nil {
		return nil, err
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA512, hasher.Sum(nil))

	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(context.WithValue(ctx, "authorize", false), http.MethodPost, "auth", body)

	if err != nil {
		return nil, err
	}

	request.Header.Set("signature", base64.StdEncoding.EncodeToString(signature))

	return request, nil
}

func NewAuthRequest(config Config) *AuthRequest {

	var payload = &AuthRequest{
		Login:     config.Login(),
		ReadOnly:  config.ReadOnly(),
		GlobalKey: config.GlobalKey(),
	}

	if v, o := config.(ConfigLabel); o {
		payload.Label = v.Label()
	} else {
		payload.Label = "libdns client - " + random(4)
	}

	if v, o := config.(ConfigExpirationTime); o {
		payload.ExpirationTime = v.ExpirationTime()
	} else {
		payload.ExpirationTime = ExpirationTime1Hour
	}

	if v, o := config.(ConfigNonce); o {
		payload.Nonce = v.Nonce()
	} else {
		payload.Nonce = random(8)
	}

	return payload
}
