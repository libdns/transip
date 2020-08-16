# TransIP for `libdns`

This package implements the libdns interfaces for the [TransIP API](https://api.transip.nl/rest/docs.html#introduction) (which has a nice Go implementation here: https://github.com/transip/gotransip)

## Authenticating

To authenticate you need to supply our AccountName and the path to your private key file to the Provider.

## Example

Here's a minimal example of how to get all your DNS records using this `libdns` provider

```go
package main

import (
        "context"
        "fmt"
        "github.com/libdns/transip"
)

func main() {
        provider := transip.Provider{AccountName: "myaccountname", PrivateKeyPath: "./transip.key"}

        records, err  := provider.GetRecords(context.TODO(), "example.com")
        if err != nil {
                fmt.Println(err.Error())
        }

        for _, record := range records {
                fmt.Printf("%s %v %s %s\n", record.Name, record.TTL.Seconds(), record.Type, record.Value)
        }
}
```
