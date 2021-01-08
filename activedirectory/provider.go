package activedirectory

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	log "github.com/sirupsen/logrus"
)

var ProviderHost = "registry.terraform.io"
var ProviderNamespace = "hashicorp"
var ProviderNameAd = "activedirectory"
var ProviderSource = ProviderHost + "/" + ProviderNamespace + "/" + ProviderNameAd
var ResourcesNameOrganizationUnit = ProviderNameAd + "_ou"
var ResourcesNameGroup = ProviderNameAd + "_group"
var ResourcesNameComputer = ProviderNameAd + "_computer"

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"host": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("AD_HOST", nil),
					Description: "The AD server to connect to.",
				},
				"port": {
					Type:        schema.TypeInt,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AD_PORT", 389),
					Description: "The AD protocol port (default: 389).",
				},
				"domain": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("AD_DOMAIN", nil),
					Description: "The AD base domain.",
				},
				"use_tls": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AD_USE_TLS", true),
					Description: "Use TLS to secure the connection (default: true).",
				},
				"no_cert_verify": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("AD_NO_CERT_VERIFY", false),
					Description: "Do not verify TLS certificate (default: false).",
				},
				"user": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("AD_USER", nil),
					Description: "User to be used for authenticating on the AD server.",
				},
				"password": {
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("AD_PASSWORD", nil),
					Description: "Password to authenticate the user.",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
			},
			ResourcesMap: map[string]*schema.Resource{
				ResourcesNameComputer:         resourceADComputerObject(),
				ResourcesNameOrganizationUnit: resourceADOUObject(),
				ResourcesNameGroup:            resourceADGroupObject(),
			},
		}
		p.ConfigureContextFunc = providerConfigure

		return p
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	api := &API{
		host:     d.Get("host").(string),
		port:     d.Get("port").(int),
		domain:   d.Get("domain").(string),
		useTLS:   d.Get("use_tls").(bool),
		insecure: d.Get("no_cert_verify").(bool),
		user:     d.Get("user").(string),
		password: d.Get("password").(string),
	}
	var diags diag.Diagnostics
	log.Infof("Connecting to %s:%d as user %s@%s.", api.host, api.port, api.user, api.domain)

	if err := api.connect(); err != nil {
		return nil, diag.FromErr(err)
	}
	return api, diags
}
