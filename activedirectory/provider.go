package activedirectory

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	log "github.com/sirupsen/logrus"
)

// Provider for terraform ad provider
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ad_host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AD_HOST", nil),
				Description: "The AD server to connect to.",
			},
			"ad_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AD_PORT", 389),
				Description: "The AD protocol port (default: 389).",
			},
			"use_tls": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AD_USE_TLS", true),
				Description: "Use TLS to secure the connection (default: true).",
			},
			"bind_user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AD_BIND_USER", nil),
				Description: "Bind user to be used for authenticating on the AD server.",
			},
			"bind_password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AD_BIND_PASSWORD", nil),
				Description: "Password to authenticate the Bind user.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"activedirectory_computer": resourceADComputerObject(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	api := &API{
		adHost:       d.Get("ad_host").(string),
		adPort:       d.Get("ad_port").(int),
		useTLS:       d.Get("use_tls").(bool),
		bindUser:     d.Get("bind_user").(string),
		bindPassword: d.Get("bind_password").(string),
	}

	log.Infof("Connecting to ad server %s (%d) as user %s.", api.adHost, api.adPort, api.bindUser)

	if err := api.connect(); err != nil {
		return nil, err
	}

	return api, nil
}
