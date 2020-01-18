package activedirectory

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	log "github.com/sirupsen/logrus"
)

// Provider for terraform ad provider
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
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
				DefaultFunc: schema.EnvDefaultFunc("AD_NO_CERT_VERIFY", true),
				Description: "Do not verify TLS certificate (default: true).",
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

		ResourcesMap: map[string]*schema.Resource{
			"activedirectory_computer": resourceADComputerObject(),
			"activedirectory_ou":       resourceADOUObject(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	api := &API{
		host:     d.Get("host").(string),
		port:     d.Get("port").(int),
		domain:   d.Get("domain").(string),
		useTLS:   d.Get("use_tls").(bool),
		insecure: d.Get("no_cert_verify").(bool),
		user:     d.Get("user").(string),
		password: d.Get("password").(string),
	}

	log.Infof("Connecting to %s:%d as user %s@%s.", api.host, api.port, api.user, api.domain)

	if err := api.connect(); err != nil {
		return nil, fmt.Errorf("providerConfigure - connection to active directory failed: %s", err)
	}

	return api, nil
}
