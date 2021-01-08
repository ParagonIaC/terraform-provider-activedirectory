package activedirectory

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

func TerraformProviderRequestSection() string {
	return fmt.Sprintf(`
terraform {
  required_providers {
    %s = {
      source = "%s"
    }
  }
}
`, ProviderNameAd, ProviderSource)
}

func getTestConnection() (*API, error) {
	host := os.Getenv("AD_HOST")
	port, err := strconv.Atoi(os.Getenv("AD_PORT"))
	if err != nil {
		return nil, err
	}
	domain := os.Getenv("AD_DOMAIN")

	useTls, err := strconv.ParseBool(os.Getenv("AD_USE_TLS"))
	if err != nil {
		return nil, err
	}
	insecure, err := strconv.ParseBool(os.Getenv("AD_NO_CERT_VERIFY"))
	if err != nil {
		return nil, err
	}
	user := os.Getenv("AD_USER")
	password := os.Getenv("AD_PASSWORD")
	api := &API{
		host:     host,
		port:     port,
		domain:   domain,
		useTLS:   useTls,
		insecure: insecure,
		user:     user,
		password: password,
	}
	if err := api.connect(); err != nil {
		return nil, err
	}
	return api, nil
}

func init() {
	testAccProvider = New("dev")()
	testAccProviders = map[string]*schema.Provider{
		ProviderNameAd: testAccProvider,
	}
	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		ProviderNameAd: func() (*schema.Provider, error) { return New("dev")(), nil },
	}
}

// unit tests
func TestProvider(t *testing.T) {
	t.Run("Provider - Should return a valid 'schema.Provider'", func(t *testing.T) {
		response := New("dev")()

		assert.NotNil(t, response)
		assert.IsType(t, &schema.Provider{}, response)

		assert.Equal(t, schema.TypeString, response.Schema["host"].Type)
		assert.Equal(t, true, response.Schema["host"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["domain"].Type)
		assert.Equal(t, true, response.Schema["domain"].Required)

		assert.Equal(t, schema.TypeInt, response.Schema["port"].Type)
		assert.Equal(t, false, response.Schema["port"].Required)

		assert.Equal(t, schema.TypeBool, response.Schema["use_tls"].Type)
		assert.Equal(t, false, response.Schema["use_tls"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["user"].Type)
		assert.Equal(t, true, response.Schema["user"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["password"].Type)
		assert.Equal(t, true, response.Schema["password"].Required)
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
		resourceSchema := New("dev")().Schema
		resourceDataMap := map[string]interface{}{
			"user":     "Tester",
			"password": "Password",
			"use_tls":  false,
			"host":     host,
			"port":     port,
			"domain":   domain,
		}
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

		api, err := providerConfigure(nil, resourceLocalData)

		assert.False(t, err.HasError())
		assert.IsType(t, &API{}, api)

		assert.IsType(t, &ldap.Conn{}, api.(*API).client)
		assert.Equal(t, host, api.(*API).host)
		assert.Equal(t, port, api.(*API).port)
		assert.Equal(t, domain, api.(*API).domain)
		assert.Equal(t, false, api.(*API).useTLS)
	})

	t.Run("providerConfigure - Should return a error when connection to AD failed", func(t *testing.T) {
		resourceSchema := New("dev")().Schema
		resourceDataMap := map[string]interface{}{
			"user":     "Tester",
			"password": "wrong",
			"use_tls":  false,
			"host":     host,
			"port":     port,
			"domain":   domain,
		}
		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)

		api, err := providerConfigure(nil, resourceLocalData)

		assert.True(t, err.HasError())
		assert.Nil(t, api)
	})
}
