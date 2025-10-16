package transip

import (
	"net/url"
	"strconv"
)

func DefaultApiBaseUri() *ApiBaseUri {
	return &ApiBaseUri{
		Scheme: "https",
		Host:   "api.transip.nl",
		Path:   "/v6/",
	}
}

type ApiBaseUri url.URL

func (a *ApiBaseUri) MarshalJSON() ([]byte, error) {
	return []byte((*url.URL)(a).String()), nil
}

func (a *ApiBaseUri) UnmarshalJSON(data []byte) error {

	if out, err := strconv.Unquote(string(data)); err == nil {
		data = []byte(out)
	}

	b, err := url.Parse(string(data))

	if err != nil {
		return err
	}

	*a = ApiBaseUri(*b)

	return nil
}
