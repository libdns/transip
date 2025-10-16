package client

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/libdns/libdns"
	"github.com/pbergman/provider"
)

type DNSEntries struct {
	Entries []*DNSRecord `json:"dnsEntries"`
}

type DNSRecord struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Expire  int    `json:"expire"`
}

func MarshallRRRecord(data *DNSRecord, zone string) *libdns.RR {
	return &libdns.RR{
		Name: libdns.RelativeName(data.Name, zone),
		Type: data.Type,
		Data: data.Content,
		TTL:  time.Duration(data.Expire) * time.Second,
	}
}

func MarshallDNSRecords(data *libdns.RR, zone string) *DNSRecord {
	var record = &DNSRecord{
		Type:    data.Type,
		Content: data.Data,
		Name:    libdns.RelativeName(data.Name, zone),
	}

	switch data.TTL.Seconds() {
	case 60, 300, 3600, 14400, 28800, 86400:
		record.Expire = int(data.TTL.Seconds())
	default:

		record.Expire = 3600

	}

	return record
}

func (c *client) SetDNSList(ctx context.Context, domain string, change provider.ChangeList) ([]libdns.Record, error) {

	switch c.control {
	case FullZoneControl:

		var data = &DNSEntries{
			Entries: make([]*DNSRecord, 0),
		}

		for record := range change.Iterate(provider.NoChange | provider.Create) {
			data.Entries = append(data.Entries, MarshallDNSRecords(record, domain))
		}

		var buffer = c.buf.Get().(*buf)

		defer buffer.Close()

		if err := json.NewEncoder(buffer).Encode(data); err != nil {
			return nil, err
		}

		if err := c.fetch(ctx, c.toDnsPath(domain), http.MethodPut, buffer, nil); err != nil {
			return nil, err
		}

	default:

		if err := c.mutate(ctx, domain, change, provider.Delete); err != nil {
			return nil, err
		}

		if err := c.mutate(ctx, domain, change, provider.Create); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (c *client) GetDNSList(ctx context.Context, domain string) ([]libdns.Record, error) {
	var data DNSEntries

	if err := c.fetch(ctx, c.toDnsPath(domain), http.MethodGet, nil, &data); err != nil {
		return nil, err
	}

	var records = make([]libdns.Record, len(data.Entries))

	for i, c := 0, len(data.Entries); i < c; i++ {
		records[i] = MarshallRRRecord(data.Entries[i], domain)
	}

	return records, nil
}
