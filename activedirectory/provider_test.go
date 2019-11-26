package activedirectory

import (
	"testing"
	"time"

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

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["host"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["host"].Required)

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["domain"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["domain"].Required)

		assert.Equal(t, schema.TypeInt, response.(*schema.Provider).Schema["port"].Type)
		assert.Equal(t, false, response.(*schema.Provider).Schema["port"].Required)

		assert.Equal(t, schema.TypeBool, response.(*schema.Provider).Schema["use_tls"].Type)
		assert.Equal(t, false, response.(*schema.Provider).Schema["use_tls"].Required)

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["user"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["user"].Required)

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["password"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["password"].Required)
	})
}

func TestProviderConfigure(t *testing.T) {
	host := "127.0.0.1"
	port := 11389
	domain := "domain.org"

	go getADServer(host, port)()
	// give ad server time to start
	time.Sleep(1000 * time.Millisecond)

	t.Run("providerConfigure - Should return an api object when connection to AD was successful", func(t *testing.T) {
		resourceSchema := Provider().(*schema.Provider).Schema
		resourceDataMap := map[string]interface{}{
			"user":     "Tester",
			"password": "Password",
			"use_tls":  false,
			"host":     host,
			"port":     port,
			"domain":   domain,
		}
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

		api, err := providerConfigure(resourceLocalData)

		assert.NoError(t, err)
		assert.IsType(t, &API{}, api)

		assert.IsType(t, &ldap.Conn{}, api.(*API).client)
		assert.Equal(t, host, api.(*API).host)
		assert.Equal(t, port, api.(*API).port)
		assert.Equal(t, domain, api.(*API).domain)
		assert.Equal(t, false, api.(*API).useTLS)
	})

	t.Run("providerConfigure - Should return a error when connection to AD failed", func(t *testing.T) {
		resourceSchema := Provider().(*schema.Provider).Schema
		resourceDataMap := map[string]interface{}{
			"user":     "Tester",
			"password": "wrong",
			"use_tls":  false,
			"host":     host,
			"port":     port,
			"domain":   domain,
		}
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

		api, err := providerConfigure(resourceLocalData)

		assert.Error(t, err)
		assert.Nil(t, api)
	})
}
