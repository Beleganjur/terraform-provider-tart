package main

import (
	"github.com/beleganjur/terraform-provider-tart/tart"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: tart.Provider,
	})
}
