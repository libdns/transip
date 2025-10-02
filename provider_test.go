package transip

import (
	"os"
	"testing"

	"github.com/pbergman/provider/test"
)

func TestProvider(t *testing.T) {

	var provider = &Provider{
		AuthLogin:  os.Getenv("LOGIN"),
		PrivateKey: os.Getenv("KEY"),
	}

	if _, ok := os.LookupEnv("DEBUG"); ok {
		provider.Debug = true
	}

	test.RunProviderTests(t, provider)
}
