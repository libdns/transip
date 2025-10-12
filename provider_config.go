package transip

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/libdns/transip/client"
)

func (p *Provider) Login() string {
	return p.AuthLogin
}

func (p *Provider) ReadOnly() bool {
	return p.AuthReadOnly
}

func (p *Provider) GlobalKey() bool {
	return false == p.AuthNotGlobalKey
}

func (p *Provider) GetBaseUri() *url.URL {

	if nil == p.BaseUri {
		return nil
	}

	return (*url.URL)(p.BaseUri)
}

func (p *Provider) GetPrivateKey() (*rsa.PrivateKey, error) {
	var block *pem.Block

	if false == bytes.HasPrefix([]byte(p.PrivateKey), []byte("-----BEGIN PRIVATE KEY-----")) {
		out, err := os.ReadFile(p.PrivateKey)

		if err != nil {
			return nil, err
		}

		block, _ = pem.Decode(out)
	} else {
		block, _ = pem.Decode([]byte(p.PrivateKey))
	}

	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block of private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	if v, ok := key.(*rsa.PrivateKey); ok {
		return v, nil
	}

	return nil, fmt.Errorf("failed to parse private key")
}

func (p *Provider) ExpirationTime() client.ExpirationTime {
	return p.AuthExpirationTime
}

func (p *Provider) StorageKey() string {
	var hasher = sha1.New()

	if _, err := fmt.Fprintf(hasher, "login:%s|gk:%t|ro:%t", p.Login(), p.GlobalKey(), p.ReadOnly()); err != nil {
		return hex.EncodeToString(strconv.AppendBool(strconv.AppendBool([]byte(p.Login()), p.GlobalKey()), p.ReadOnly()))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}
