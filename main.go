package main

import (
	"context"
	"fmt"
	"."
)

func main() {
	provider := transip.Provider{AccountName: "mdbraber", PrivateKeyPath: "./transip.key"}
	provider.NewSession()

	records, err  := provider.GetRecords(context.TODO(), "mdbraber.com")
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err.Error())
	}

	for _, record := range records {
		fmt.Println(record.Name)
		fmt.Println(record.Value)
	}
}
