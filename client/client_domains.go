package client

import (
	"context"
	"net/http"

	"github.com/pbergman/provider"
)

type Domain struct {
	Name string `json:"name"`
}

func (d *Domain) String() string {
	return d.Name
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
		domains[i] = domain
	}

	return domains, nil
}
