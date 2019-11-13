package main

import (
	ldap "github.com/adlerrobert/terraform-provider-ldap/ldap"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ldap.Provider,
	})
}
