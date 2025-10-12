package transip

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/libdns/transip/client"
	"github.com/pbergman/provider"
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

	var handler = &Provider{
		AuthLogin:  os.Getenv("LOGIN"),
		PrivateKey: os.Getenv("KEY"),
	}

	if _, ok := os.LookupEnv("DEBUG"); ok {
		if x, ok := strconv.Atoi(os.Getenv("DEBUG")); ok == nil {
			handler.DebugLevel = provider.OutputLevel(x)
		}
	}

	if _, ok := os.LookupEnv("FULL_ZONE"); ok {
		handler.ClientControl = client.FullZoneControl
	}

	test.RunProviderTests(t, handler, test.TestAll)
}
