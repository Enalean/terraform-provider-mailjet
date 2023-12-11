//go:generate tfplugindocs
package main

import (
	"context"
	"flag"
	"log"

	"github.com/enalean/terraform-provider-mailjet/mailjet"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is filled by goreleaser during build
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/enalean/mailjet",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), mailjet.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
