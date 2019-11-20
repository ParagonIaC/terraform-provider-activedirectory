package activedirectory

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
	"gopkg.in/ldap.v3"
)

// acceptance tests
var testAccProviders map[string]terraform.ResourceProvider // nolint:gochecknoglobals
var testAccProvider *schema.Provider                       // nolint:gochecknoglobals

func init() { // nolint:gochecknoinits
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"activedirectory": testAccProvider,
	}
}

// unit tests
func TestProvider(t *testing.T) {
	t.Run("Provider - Should return a valid 'schema.Provider'", func(t *testing.T) {
		response := Provider()

		assert.NotNil(t, response)
		assert.IsType(t, &schema.Provider{}, response)

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["ad_host"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["ad_host"].Required)

		assert.Equal(t, schema.TypeInt, response.(*schema.Provider).Schema["ad_port"].Type)
		assert.Equal(t, false, response.(*schema.Provider).Schema["ad_port"].Required)

		assert.Equal(t, schema.TypeBool, response.(*schema.Provider).Schema["use_tls"].Type)
		assert.Equal(t, false, response.(*schema.Provider).Schema["use_tls"].Required)

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["bind_user"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["bind_user"].Required)

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["bind_password"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["bind_password"].Required)
	})
}

func TestProviderConfigure(t *testing.T) {
	host := "127.0.0.1"
	port := 10389

	go getADServer(host, port)()

	t.Run("providerConfigure - Should return a api when connection to AD was successful", func(t *testing.T) {
		resourceSchema := Provider().(*schema.Provider).Schema
		resourceDataMap := map[string]interface{}{
			"bind_user":     "Tester",
			"bind_password": "Password",
			"use_tls":       false,
			"ad_host":       host,
			"ad_port":       port,
		}
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

		api, err := providerConfigure(resourceLocalData)

		assert.NoError(t, err)
		assert.IsType(t, &API{}, api)

		assert.IsType(t, &ldap.Conn{}, api.(*API).client)
		assert.Equal(t, host, api.(*API).adHost)
		assert.Equal(t, port, api.(*API).adPort)
		assert.Equal(t, false, api.(*API).useTLS)
	})

	t.Run("providerConfigure - Should return a error when connection to AD failed", func(t *testing.T) {
		resourceSchema := Provider().(*schema.Provider).Schema
		resourceDataMap := map[string]interface{}{
			"bind_user":     "Tester",
			"bind_password": "wrong",
			"use_tls":       false,
			"ad_host":       host,
			"ad_port":       port,
		}
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

		api, err := providerConfigure(resourceLocalData)

		assert.Error(t, err)
		assert.Nil(t, api)
	})
}
