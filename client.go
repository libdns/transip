package transip

import (
	"context"
	"time"

	"github.com/libdns/libdns"
	"github.com/transip/gotransip/v6"
	transipdomain "github.com/transip/gotransip/v6/domain"
)

func (p *Provider) setupRepository() error {
	if p.repository == nil {
		client, err := gotransip.NewClient(gotransip.ClientConfiguration{
			AccountName:	p.AccountName,
			PrivateKeyPath:	p.PrivateKeyPath,
		})
		if err != nil {
			return err
		}
		p.repository = &transipdomain.Repository{Client: client}
	}

	return nil
}

func (p *Provider) getDNSEntries(ctx context.Context, domain string) ([]libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.setupRepository()
	if err != nil {
		return nil, err
	}
	
	var records []libdns.Record
	var dnsEntries []transipdomain.DNSEntry

	dnsEntries, err = p.repository.GetDNSEntries(domain)
	if err != nil {
		return nil, err
	}

	for _, entry := range dnsEntries {
		record := libdns.Record{
			Name:  entry.Name,
			Value: entry.Content,
			Type:  entry.Type,
			TTL:   time.Duration(entry.Expire) * time.Second,
		}
		records = append(records, record)
	}

	return records, nil
}

func (p *Provider) addDNSEntry(ctx context.Context, domain string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.setupRepository()
	if err != nil {
		return libdns.Record{}, err
	}

	entry := transipdomain.DNSEntry{
		Name:    record.Name,
		Content: record.Value,
		Type:    record.Type,
		Expire:  int(record.TTL.Seconds()),
	}

	err = p.repository.AddDNSEntry(domain, entry)
	if err != nil {
		return libdns.Record{}, err
	}

	return record, nil
}

func (p *Provider) removeDNSEntry(ctx context.Context, domain string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.setupRepository()
	if err != nil {
		return libdns.Record{}, err
	}

	entry := transipdomain.DNSEntry{
		Name:    record.Name,
		Content: record.Value,
		Type:    record.Type,
		Expire:  int(record.TTL.Seconds()),
	}

	err = p.repository.RemoveDNSEntry(domain, entry)
	if err != nil {
		return libdns.Record{}, err
	}

	return record, nil
}

func (p *Provider) updateDNSEntry(ctx context.Context, domain string, record libdns.Record) (libdns.Record, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	err := p.setupRepository()
	if err != nil {
		return libdns.Record{}, err
	}

	entry := transipdomain.DNSEntry{
		Name:    record.Name,
		Content: record.Value,
		Type:    record.Type,
		Expire:  int(record.TTL.Seconds()),
	}

	err = p.repository.UpdateDNSEntry(domain, entry)
	if err != nil {
		return libdns.Record{}, err
	}

	return record, nil
}
