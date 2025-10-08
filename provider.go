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
	// AuthLogin the login used for authentication
	AuthLogin string `json:"login"`
	// AuthReadOnly set to true to create readonly keys
	AuthReadOnly bool `json:"read_only"`
	// AuthNotGlobalKey can be set to true to generate keys that are
	// restricted to clients with IP addresses included in the whitelist.
	AuthNotGlobalKey bool `json:"not_global_key"`
	// AuthExpirationTime specifies the time-to-live for an authentication token.
	AuthExpirationTime client.ExpirationTime `json:"expiration_time"`

	// PrivateKey can be generated here:
	// https://www.transip.nl/cp/account/api
	//
	// It is used for authentication and can be provided either as a
	// string containing the key or as a filepath to a file containing
	// the private key.
	PrivateKey string `json:"private_key"`

	// Debug set to true to dump the API requests.
	Debug bool `json:"debug"`
	// DebugLevel sets the verbosity for logging API requests and responses.
	// 0 (default) prints only the request line.
	// 1 prints the full request and response.
	DebugLevel client.DebugLevel `json:"debug_level"`
	// DebugOut defines the output destination for debug logs.
	// Defaults to standard output (STDOUT).
	DebugOut io.Writer `json:"-"`

	// BaseURI is the base URI used for API calls.
	// Default: https://api.transip.nl/v6/
	BaseUri *ApiBaseUri `json:"base_uri"`

	// TokenStorage specifies where tokens are stored and can be reused
	// until they expire. It can be set to:
	// - "memory" for in-memory storage,
	// - a file path for storing keys on disk
	// - empty, in which case keys will be stored in a "transip" directory
	//   in the user's temp folder.
	TokenStorage string `json:"token_storage"`

	// ClientControl has two modes:
	// - RecordLevelControl (default): updates records individually.
	// - FullZoneControl: replaces the entire zone in a single call.
	//   While this is much faster, it can encounter race conditions if
	//   another program modifies the zone simultaneously, as updates
	//   may be overwritten.
	ClientControl client.ControleMode `json:"client_control_mode"`
	client        Client

	pLock sync.RWMutex
	cLock sync.Mutex
}

func (p *Provider) getClient() Client {
	p.cLock.Lock()
	defer p.cLock.Unlock()

	if nil == p.client {

		if nil == p.BaseUri {
			p.BaseUri = DefaultApiBaseUri()
		}

		if nil == p.DebugOut {
			p.DebugOut = os.Stdout
		}

		if p.TokenStorage == "memory" && "" == p.AuthExpirationTime {
			p.AuthExpirationTime = client.ExpirationTime1Hour
		}

		if "" == p.AuthExpirationTime {
			p.AuthExpirationTime = client.ExpirationTime1Day
		}

		p.client = client.NewClient(p, NewTokenStorage(p.TokenStorage), p.ClientControl)
	}

	return p.client
}

func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	return provider.GetRecords(ctx, &p.pLock, p.getClient(), zone)
}

func (p *Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	return provider.AppendRecords(ctx, &p.pLock, p.getClient(), zone, recs)
}

func (p *Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	return provider.SetRecords(ctx, &p.pLock, p.getClient(), zone, recs)
}

func (p *Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) ([]libdns.Record, error) {
	return provider.DeleteRecords(ctx, &p.pLock, p.getClient(), zone, recs)
}

func (p *Provider) ListZones(ctx context.Context) ([]libdns.Zone, error) {
	return provider.ListZones(ctx, &p.pLock, p.getClient())
}

func NewTokenStorage(location string) client.Storage {

	var storage client.Storage

	if location == "memory" {
		storage = client.NewTokenMemoryStorage()
	} else {

		if location == "" {
			location = filepath.Join(os.TempDir(), "transip")
		}

		var err error

		storage, err = client.NewTokenFileStorage(location)

		if err != nil {
			storage = client.NewTokenMemoryStorage()
		}
	}

	return storage
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
