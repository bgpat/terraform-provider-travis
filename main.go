package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/bgpat/terraform-provider-travis/travis"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: travis.Provider,
	})
}
