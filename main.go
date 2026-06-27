package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/dmalch/terraform-provider-familio/internal"
)

func main() {
	err := providerserver.Serve(context.Background(), internal.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/dmalch/familio",
	})
	if err != nil {
		log.Fatal(err)
	}
}
