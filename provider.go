package transip

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/libdns/libdns"
	"github.com/libdns/transip/client"
	"github.com/pbergman/provider"
)

type Client interface {
	provider.Client
	provider.ZoneAwareClient
}

type Provider struct {
	AuthLogin          string                `json:"login"`
	AuthReadOnly       bool                  `json:"read_only"`
	AuthNotGlobalKey   bool                  `json:"not_global_key"`
	AuthExpirationTime client.ExpirationTime `json:"expiration_time"`
	PrivateKey         string                `json:"private_key"`
	Debug              bool                  `json:"debug"`
	DebugOut           io.Writer             `json:"-"`
	BaseUri            *ApiBaseUri           `json:"base_uri"`
	TokenStorage       string                `json:"token_storage"`

	client Client
	mutex  sync.RWMutex
}

func (p *Provider) getClient() Client {
	if nil == p.client {

		if nil == p.BaseUri {
			p.BaseUri = DefaultApiBaseUri()
		}

		if nil == p.DebugOut {
			p.DebugOut = os.Stdout
		}

		var storage client.Storage

		if p.TokenStorage == "memory" {
			storage = client.NewTokenMemoryStorage()

			if "" == p.AuthExpirationTime {
				p.AuthExpirationTime = client.ExpirationTime1Hour
			}
		} else {

			var root = p.TokenStorage

			if root == "" {
				root = filepath.Join(os.TempDir(), "transip")
			}

			var err error

			storage, err = client.NewTokenFileStorage(root)

			if err != nil {
				storage = client.NewTokenMemoryStorage()
			}
		}

		if "" == p.AuthExpirationTime {
			p.AuthExpirationTime = client.ExpirationTime1Day
		}

		p.client = client.NewClient(p, storage)
	}

	return p.client
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	return provider.GetRecords(ctx, &p.mutex, p.getClient(), zone)
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	return provider.AppendRecords(ctx, &p.mutex, p.getClient(), zone, recs)
}

func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	return provider.SetRecords(ctx, &p.mutex, p.getClient(), zone, recs)
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	return provider.DeleteRecords(ctx, &p.mutex, p.getClient(), zone, recs)
}

func (p *Provider) ListZones(ctx context.Context) ([]libdns.Zone, error) {
	return provider.ListZones(ctx, &p.mutex, p.getClient())
}

// Interface guards
var (
	_ client.Config         = (*Provider)(nil)
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
	_ libdns.ZoneLister     = (*Provider)(nil)
)
