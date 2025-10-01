package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

type transport struct {
	http.RoundTripper

	config  Config
	storage Storage
	refresh TokenFetcher
}

func (t *transport) getToken(ctx context.Context) (Token, error) {
	token, err := t.storage.Get(t.config.StorageKey())

	if err != nil {
		return nil, err
	}

	if token == nil || token.IsExpired() || ContextValue(ctx, "token_refresh_force", false) {
		value, err := t.refresh(ctx, t.config)

		if err != nil {
			return nil, err
		}

		token, err = NewToken(value)

		if err != nil {
			return nil, err
		}

		if err := t.storage.Set(t.config.StorageKey(), token); err != nil {
			return nil, err
		}
	}

	return token, nil
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {

	if ContextValue(req.Context(), "authorize", true) {
		jwt, err := t.getToken(req.Context())

		if err != nil {
			return nil, err
		}

		req.Header.Set("authorization", fmt.Sprintf("Bearer %s", jwt))
	}

	if uri := t.config.GetBaseUri(); uri != nil {
		req.URL = uri.ResolveReference(req.URL)
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")

	var writer = t.config.GetDebug()

	if writer != nil {
		dump(req, httputil.DumpRequest, "c", writer)
	}

	response, err := t.RoundTripper.RoundTrip(req)

	if writer != nil && nil != response {
		dump(response, httputil.DumpResponse, "s", writer)
	}

	// perhaps the token expired of revoked? let`s try once more
	if nil != response && response.StatusCode == http.StatusUnauthorized && false == ContextValue(req.Context(), "token_refresh_force", false) {
		return t.RoundTrip(req.WithContext(context.WithValue(req.Context(), "token_refresh_force", true)))
	}

	return response, err
}

func dump[T *http.Request | *http.Response](x T, d func(T, bool) ([]byte, error), p string, o io.Writer) {
	if out, err := d(x, true); err == nil {
		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			_, _ = fmt.Fprintf(o, "[%s] %s\n", p, scanner.Text())
		}
	}
}
