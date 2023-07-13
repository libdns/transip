package transip

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/libdns/libdns"
	transipdomain "github.com/transip/gotransip/v6/domain"
)

// unFQDN trims any trailing "." from fqdn. TransIP's API does not use FQDNs.
func (p *Provider) unFQDN(fqdn string) string {
	return strings.TrimSuffix(fqdn, ".")
}

// Provider implements the libdns interfaces for Route53
type Provider struct {
	AccountName    string `json:"account_name"`
	PrivateKeyPath string `json:"private_key_path"`
	repository     *transipdomain.Repository
	mutex          sync.Mutex
}

// GetRecords lists all the records in the zone.
func (p *Provider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	records, err := p.getDNSEntries(ctx, p.unFQDN(zone))
	if err != nil {
		return nil, err
	}

	return records, nil
}

// AppendRecords adds records to the zone. It returns the records that were added.
func (p *Provider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var appendedRecords []libdns.Record

	for _, record := range records {

		if record.TTL < time.Duration(300)*time.Second {
			record.TTL = time.Duration(300) * time.Second
		}

		newRecord, err := p.addDNSEntry(ctx, p.unFQDN(zone), record)
		if err != nil {
			return nil, err
		}
		
		appendedRecords = append(appendedRecords, newRecord)
	}

	return appendedRecords, nil
}

// DeleteRecords deletes the records from the zone.
func (p *Provider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var deletedRecords []libdns.Record

	for _, record := range records {
		deletedRecord, err := p.removeDNSEntry(ctx, p.unFQDN(zone), record)
		if err != nil {
			return nil, err
		}
		deletedRecord.TTL = time.Duration(deletedRecord.TTL) * time.Second
		deletedRecords = append(deletedRecords, deletedRecord)
	}

	return deletedRecords, nil
}

// SetRecords sets the records in the zone, either by updating existing records
// or creating new ones. It returns the updated records.
func (p *Provider) SetRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var setRecords []libdns.Record

	for _, record := range records {
		setRecord, err := p.updateDNSEntry(ctx, p.unFQDN(zone), record)
		if err != nil {
			return nil, err
		}
		setRecord.TTL = time.Duration(setRecord.TTL) * time.Second
		setRecords = append(setRecords, setRecord)
	}

	return setRecords, nil
}

// Interface guards
var (
	_ libdns.RecordGetter   = (*Provider)(nil)
	_ libdns.RecordAppender = (*Provider)(nil)
	_ libdns.RecordSetter   = (*Provider)(nil)
	_ libdns.RecordDeleter  = (*Provider)(nil)
)
