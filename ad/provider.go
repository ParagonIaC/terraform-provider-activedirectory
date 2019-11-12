package activedirectory

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	log "github.com/sirupsen/logrus"
)

// Provider for terraform active directory
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{

			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Domain of the AD Server",
				DefaultFunc: schema.EnvDefaultFunc("AD_DOMAIN", nil),
			},

			"ip": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The IP of the AD Server",
				DefaultFunc: schema.EnvDefaultFunc("AD_IP", nil),
			},

			"user": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The user name to connect to the AD Server",
				DefaultFunc: schema.EnvDefaultFunc("AD_USER", nil),
			},

			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The user password to connect to the AD Server",
				DefaultFunc: schema.EnvDefaultFunc("AD_PASSWORD", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"activedirectory_computer": resourceADComputer(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	api := NewAPI(d.Get("ip").(string), d.Get("domain").(string))

	log.Infof("Connecting to AD %s (%s) as user %s.", d.Get("domain").(string), d.Get("ip").(string), d.Get("user").(string))

	if err := api.Connect(d.Get("user").(string), d.Get("password").(string)); err != nil {
		return nil, err
	}

	return api, nil
}
