package client

import (
	"context"
	"errors"
	"net/http"
)

func (c *client) Ping(ctx context.Context) error {

	var data struct {
		Ping string `json:"ping"`
	}

	if err := c.fetch(ctx, "api-test", http.MethodGet, nil, &data); err != nil {
		return err
	}

	if data.Ping != "pong" {
		return errors.New("invalid response")
	}

	return nil
}
