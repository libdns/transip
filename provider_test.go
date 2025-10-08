package transip

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/libdns/transip/client"
	"github.com/pbergman/provider/test"
)

func TestProvider_Unmarshall(t *testing.T) {
	var provider *Provider
	var buf = `{
"client_control_mode": "full zone"
}`

	if err := json.Unmarshal([]byte(buf), &provider); err != nil {
		t.Fatal(err)
	}

	if provider.ClientControl != client.FullZoneControl {
		t.Fatalf("invalid client control mode, expecting %d got %d", client.FullZoneControl, provider.ClientControl)
	}

}

func TestProvider(t *testing.T) {

	var provider = &Provider{
		AuthLogin:  os.Getenv("LOGIN"),
		PrivateKey: os.Getenv("KEY"),
	}

	if _, ok := os.LookupEnv("DEBUG"); ok {
		provider.Debug = true
	}

	if _, ok := os.LookupEnv("FULL_ZONE"); ok {
		provider.ClientControl = client.FullZoneControl
	}

	test.RunProviderTests(t, provider, test.TestAll)
}
