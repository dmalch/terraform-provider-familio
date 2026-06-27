package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/dmalch/terraform-provider-familio/internal"
)

// Set by goreleaser via -ldflags "-X main.version=… -X main.commit=…".
var (
	version = "dev"
	commit  = ""
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	_ = commit
	err := providerserver.Serve(context.Background(), internal.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/dmalch/familio",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}
