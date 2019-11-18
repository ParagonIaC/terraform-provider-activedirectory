package ldap

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestProvider(t *testing.T) {
	t.Run("Provider - Should return a valid 'schema.Provider'", func(t *testing.T) {
		response := Provider()

		assert.NotNil(t, response)
		assert.IsType(t, &schema.Provider{}, response)

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["ldap_host"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["ldap_host"].Required)

		assert.Equal(t, schema.TypeInt, response.(*schema.Provider).Schema["ldap_port"].Type)
		assert.Equal(t, false, response.(*schema.Provider).Schema["ldap_port"].Required)

		assert.Equal(t, schema.TypeBool, response.(*schema.Provider).Schema["use_tls"].Type)
		assert.Equal(t, false, response.(*schema.Provider).Schema["use_tls"].Required)

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["bind_user"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["bind_user"].Required)

		assert.Equal(t, schema.TypeString, response.(*schema.Provider).Schema["bind_password"].Type)
		assert.Equal(t, true, response.(*schema.Provider).Schema["bind_password"].Required)
	})
}
