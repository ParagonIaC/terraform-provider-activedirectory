package main

import (
	"github.com/adlerrobert/terraform-provider-ad/ad"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ad.Provider,
	})
}
