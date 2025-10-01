# TransIP for `libdns`

This package implements the libdns interfaces for the [TransIP API](https://api.transip.nl/rest/docs.html)

## Authenticating

To authenticate, you need to generate a key pair key [here](https://www.transip.nl/cp/account/api).

## Example

Here's a minimal example of how to get all your DNS records using this `libdns` provider

```go
package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/pbergman/provider"
	"github.com/libdns/transip"
)

func main() {
	var x = &transip.Provider{
		AuthLogin:  "user",
		PrivateKey: "private.key",
	}

	zones, err := x.ListZones(context.Background())

	if err != nil {
		panic(err)
	}

	var writer = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)

	for _, zone := range zones {
		records, err := x.GetRecords(context.Background(), zone.Name)

		if err != nil {
			panic(err)
		}

		for _, record := range provider.RecordIterator(&records) {
			_, _ = fmt.Fprintf(writer, "%s\t%v\t%s\t%s\n", record.Name, record.TTL.Seconds(), record.Type, record.Data)
		}

	}

	_ = writer.Flush()
}
```

## Debugging

This library provides the ability to debug the request/response communication with the API server.

To enable debugging, simply set the `debugging` property to `true`:
```go
	var x = &transip.Provider{
        AuthLogin:  "user",
        PrivateKey: "private.key",
    }

	zones, err := provider.ListZones(context.Background())

	if err != nil {
		panic(err)
	}

	records, err := provider.GetRecords(context.Background(), "example.nl")
```

```shell
........................
[c] GET /v6/domains HTTP/1.1
[c] Host: api.transip.nl
[c] Accept: application/json
[c] Authorization: Bearer ********************
[c] Content-Type: application/json
[c] 
[s] HTTP/2.0 200 OK
[s] Connection: close
[s] Content-Type: application/json
......
[s] 
[s] {"domains":[{"name":"...
```

This will by default write to stdout but can set to any `io.Writer` by also setting the `DebugOut` property. 

```go
    var provider = &transip.Provider{
        AuthLogin:  "user",
        PrivateKey: "private.key",
        Debug: true,
        DebugOut: log.Writer(),
    }
```

## Testing

This library comes with a test suite that verifies the interface by creating a few test records, validating them, and then removing those records. To run the tests, you can use:

```shell
KEY=<KEY_FILE> LOGIN=<USER> go test
```

Or run more verbose test to dump all api requests and responses: 

```shell
KEY=<KEY_FILE> LOGIN=<USER> DEBUG=1 go test -v 
```

