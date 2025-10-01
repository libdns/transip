package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pbergman/provider"
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

func NewClient(config Config, storage Storage) ApiClient {
	object := new(client)
	object.client = &http.Client{
		Transport: &transport{
			RoundTripper: http.DefaultTransport,
			refresh:      object.Authorize,
			config:       config,
			storage:      storage,
		},
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
	client *http.Client
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
