package client

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pbergman/provider"
)

type DNSEntry struct {
	Entry *DNSRecord `json:"dnsEntry"`
}

func (c *client) mutate(ctx context.Context, domain string, change provider.ChangeList, state provider.ChangeState) error {
	var method string

	switch state {
	case provider.Delete:
		method = http.MethodDelete
	case provider.Create:
		method = http.MethodPost
	default:
		method = http.MethodPatch
	}

	var buf = c.buf.Get().(*buf)

	defer buf.Close()

	for record := range change.Iterate(state) {

		if err := json.NewEncoder(buf).Encode(&DNSEntry{Entry: MarshallDNSRecords(record, domain)}); err != nil {
			return err
		}

		if err := c.fetch(ctx, c.toDnsPath(domain), method, buf, nil); err != nil {
			return err
		}

		buf.Reset()
	}

	return nil
}
