package ldap

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	log "github.com/sirupsen/logrus"
)

// Provider for terraform ldap provider
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ldap_host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_HOST", nil),
				Description: "The LDAP server to connect to.",
			},
			"ldap_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_PORT", 389),
				Description: "The LDAP protocol port (default: 389).",
			},
			"use_tls": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_USE_TLS", true),
				Description: "Use TLS to secure the connection (default: true).",
			},
			"bind_user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_BIND_USER", nil),
				Description: "Bind user to be used for authenticating on the LDAP server.",
			},
			"bind_password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LDAP_BIND_PASSWORD", nil),
				Description: "Password to authenticate the Bind user.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"ldap_computer": resourceLDAPComputerObject(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	api := &API{
		ldapHost:     d.Get("ldap_host").(string),
		ldapPort:     d.Get("ldap_port").(int),
		useTLS:       d.Get("use_tls").(bool),
		bindUser:     d.Get("bind_user").(string),
		bindPassword: d.Get("bind_password").(string),
	}

	log.Infof("Connecting to ldap server %s (%d) as user %s.", api.ldapHost, api.ldapPort, api.bindUser)

	if err := api.connect(); err != nil {
		return nil, err
	}

	return api, nil
}
