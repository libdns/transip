package client

import (
	"context"
	"fmt"
	"net/http"
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

	response, err := t.RoundTripper.RoundTrip(req)

	// perhaps the token expired of revoked? let`s try once more
	if nil != response && response.StatusCode == http.StatusUnauthorized && false == ContextValue(req.Context(), "token_refresh_force", false) {
		return t.RoundTrip(req.WithContext(context.WithValue(req.Context(), "token_refresh_force", true)))
	}

	return response, err
}
