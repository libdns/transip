package client

import (
	"context"
	"net/http"

	"github.com/pbergman/provider"
)

type Domain struct {
	Name string `json:"name"`
}

type DomainName string

func (d DomainName) Name() string {
	return string(d)
}

func (c *client) Domains(ctx context.Context) ([]provider.Domain, error) {

	var data struct {
		Domains []*Domain `json:"domains"`
		Links   *Links    `json:"_links"`
	}

	if err := c.fetch(ctx, "domains", http.MethodGet, nil, &data); err != nil {
		return nil, err
	}

	var domains = make([]provider.Domain, len(data.Domains))

	for i, domain := range data.Domains {
		domains[i] = DomainName(domain.Name)
	}

	return domains, nil
}
