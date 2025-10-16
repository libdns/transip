package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/pbergman/provider"
)

type ControleMode uint8

func (c *ControleMode) UnmarshalJSON(b []byte) error {
	var x int

	if err := json.Unmarshal(b, &x); err == nil {
		*c = ControleMode(x)
		return nil
	}

	var z string

	if err := json.Unmarshal(b, &z); err != nil {
		return errors.New("invalid controle mode")
	}

	if regexp.MustCompile(`full(_|\s)?zone((_|\s)?control)?`).Match([]byte(z)) {
		*c = FullZoneControl
	}

	return nil
}

const (
	RecordLevelControl ControleMode = iota
	FullZoneControl
)

type ApiClient interface {
	provider.Client
	provider.ZoneAwareClient
}

type ErrorResponse struct {
	Message string `json:"error"`
	Code    int    `json:"-"`
}

func (e ErrorResponse) Error() string {
	return e.Message
}

type Links []*Link

type Link struct {
	Rel  string `json:"rel"`
	Link string `json:"link"`
}

func NewClient(config Config, storage Storage, mode ControleMode) ApiClient {

	var object = &client{
		control: mode,
		buf:     NewBufPool(),
	}

	var transporter http.RoundTripper = &transport{
		RoundTripper: http.DefaultTransport,
		refresh:      object.Authorize,
		config:       config,
		storage:      storage,
	}

	if v, ok := config.(provider.DebugConfig); ok {
		transporter = &provider.DebugTransport{
			RoundTripper: transporter,
			Config:       v,
		}
	}

	object.client = &http.Client{
		Transport: transporter,
	}

	return object
}

func ContextValue[T any](ctx context.Context, name string, defaultOnNil T) T {

	if nil == ctx.Value(name) {
		return defaultOnNil
	}

	return ctx.Value(name).(T)
}

type client struct {
	client  *http.Client
	buf     *sync.Pool
	control ControleMode
}

func (a *client) toDnsPath(domain string) string {
	return fmt.Sprintf("domains/%s/dns", url.PathEscape(strings.TrimSuffix(domain, ".")))
}

func (a *client) fetch(ctx context.Context, path string, method string, body io.Reader, object any) error {

	request, err := http.NewRequestWithContext(ctx, method, path, body)

	if err != nil {
		return err
	}

	response, err := a.client.Do(request)

	if err != nil {
		return err
	}

	defer response.Body.Close()

	if !strings.HasPrefix(response.Header.Get("content-type"), "application/json") {
		return fmt.Errorf("unexpected response type: %s", response.Header.Get("content-type"))
	}

	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		if nil != object {
			if err := json.NewDecoder(response.Body).Decode(object); err != nil {
				return err
			}
		}
	} else {
		var message ErrorResponse

		if err := json.NewDecoder(response.Body).Decode(&message); err != nil {
			return err
		}

		message.Code = response.StatusCode

		return message
	}

	return nil
}

var (
	_ provider.Client          = (*client)(nil)
	_ provider.ZoneAwareClient = (*client)(nil)
)
