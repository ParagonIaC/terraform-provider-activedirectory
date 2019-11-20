package main

import (
	ad "github.com/adlerrobert/terraform-provider-activedirectory/activedirectory"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ad.Provider,
	})
}
