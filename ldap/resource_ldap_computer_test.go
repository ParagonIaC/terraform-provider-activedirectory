package ldap

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// acceptance tests
func TestAccLDAPComputer_basic(t *testing.T) {
	ou := "ou=computer,dc=company,dc=org"
	name := "test-acc-computer"
	description := "terraform"

	var computer Computer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLDAPComputerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLDAPComputerTestData(ou, name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLDAPComputerExists("ldap_computer.test", &computer),
					testAccCheckLDAPComputerAttributes(&computer, ou, name, description),
					resource.TestCheckResourceAttr("ldap_computer.test", "ou", ou),
					resource.TestCheckResourceAttr("ldap_computer.test", "name", name),
					resource.TestCheckResourceAttr("ldap_computer.test", "description", description),
					resource.TestCheckResourceAttr("ldap_computer.test", "id", fmt.Sprintf("cn=%s,%s", ou, name)),
				),
			},
		},
	})
}

func TestAccLDAPComputer_update(t *testing.T) {
	ou := "ou=computer,dc=company,dc=org"
	name := "test-acc-computer"
	description := "terraform"

	ouUpdated := "ou=update,ou=computer,dc=company,dc=org"
	descriptionUpdated := "terraform_updated"

	var computer Computer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLDAPComputerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceLDAPComputerTestData(ou, name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLDAPComputerExists("ldap_computer.test", &computer),
					testAccCheckLDAPComputerAttributes(&computer, ou, name, description),
					resource.TestCheckResourceAttr("ldap_computer.test", "ou", ou),
					resource.TestCheckResourceAttr("ldap_computer.test", "name", name),
					resource.TestCheckResourceAttr("ldap_computer.test", "description", description),
					resource.TestCheckResourceAttr("ldap_computer.test", "id", fmt.Sprintf("cn=%s,%s", ou, name)),
				),
			},
			{
				Config: testAccResourceLDAPComputerTestData(ouUpdated, name, descriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					// testAccCheckLDAPComputerExists("ldap_computer.test", &computer),
					// testAccCheckLDAPComputerAttributes(&computer, ou, name, description),
					resource.TestCheckResourceAttr("ldap_computer.test", "ou", ouUpdated),
					// resource.TestCheckResourceAttr("ldap_computer.test", "name", name),
					resource.TestCheckResourceAttr("ldap_computer.test", "description", descriptionUpdated),
					// resource.TestCheckResourceAttr("ldap_computer.test", "id", fmt.Sprintf("cn=%s,%s", ou, name)),
				),
			},
		},
	})
}

// acceptance test helpers
func testAccCheckLDAPComputerDestroy(s *terraform.State) error {
	api := testAccProvider.Meta().(*API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ldap_computer" {
			continue
		}

		computer, err := api.getComputer(rs.Primary.ID, nil)
		if err != nil {
			return err
		}

		if computer != nil {
			return fmt.Errorf("ldap computer (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckLDAPComputerExists(resourceName string, computer *Computer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("LDAP computer ID is not set")
		}

		api := testAccProvider.Meta().(*API)
		_computer, err := api.getComputer(rs.Primary.ID, []string{"description"})

		if err != nil {
			return err
		}

		*computer = *_computer
		return nil
	}
}

func testAccCheckLDAPComputerAttributes(computer *Computer, ou, name, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if computer.name != name {
			return fmt.Errorf("computer name not set correctly")
		}

		if !reflect.DeepEqual(computer.attributes["description"], []string{description}) {
			return fmt.Errorf("computer description not set correctly")
		}

		if computer.dn != fmt.Sprintf("cn=%s,%s", name, ou) {
			return fmt.Errorf("computer dn not set correctly")
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("LDAP_HOST"); v == "" {
		t.Fatal("LDAP_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("LDAP_PORT"); v == "" {
		t.Fatal("LDAP_PORT must be set for acceptance tests")
	}
	if v := os.Getenv("LDAP_USE_TLS"); v == "" {
		t.Fatal("LDAP_USE_TLS must be set for acceptance tests")
	}
	if v := os.Getenv("LDAP_BIND_USER"); v == "" {
		t.Fatal("LDAP_BIND_USER must be set for acceptance tests")
	}
	if v := os.Getenv("LDAP_BIND_PASSWORD"); v == "" {
		t.Fatal("LDAP_BIND_PASSWORD must be set for acceptance tests")
	}
}

// acceptance test data
func testAccResourceLDAPComputerTestData(ou, name, description string) string {
	return fmt.Sprintf(`
provider "ldap" {
	ldap_host      = "%s"
	ldap_port      = %s
	use_tls		   = %s
	bind_user      = "%s"
	bind_password  = "%s"
}

resource "ldap_computer" "test" {
	ou           = "%s"
	name         = "%s"
	description  = "%s"
}
	`,
		os.Getenv("LDAP_HOST"),
		os.Getenv("LDAP_PORT"),
		os.Getenv("LDAP_USE_TLS"),
		os.Getenv("LDAP_BIND_USER"),
		os.Getenv("LDAP_BIND_PASSWORD"),
		ou, name, description,
	)
}

// unit tests
func TestResourceLDAPComputerObject(t *testing.T) {
	t.Run("resourceLDAPComputerObject - should return *schema.Resource", func(t *testing.T) {
		response := resourceLDAPComputerObject()
		assert.IsType(t, &schema.Resource{}, response)

		assert.Equal(t, schema.TypeString, response.Schema["name"].Type)
		assert.Equal(t, true, response.Schema["name"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["ou"].Type)
		assert.Equal(t, true, response.Schema["ou"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["description"].Type)
		assert.Equal(t, false, response.Schema["description"].Required)
	})
}

func TestResourceLDAPComputerObjectCreate(t *testing.T) {
	name := "Test1"
	ou := "ou=test1,ou=org"
	description := "terraform"

	testComputer := &Computer{
		name: name,
		dn:   fmt.Sprintf("cn=%s,%s", name, ou),
		attributes: map[string][]string{
			"description": {description},
		},
	}

	resourceSchema := resourceLDAPComputerObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"ou":          ou,
		"description": description,
	}

	t.Run("resourceLDAPComputerObjectCreate - should return nil when everything is good", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createComputer", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("getComputer", mock.Anything, mock.Anything).Return(testComputer, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectCreate(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceLDAPComputerObjectCreate - should return error when creating failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createComputer", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectCreate(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceLDAPComputerObjectCreate - id should be set to dn", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createComputer", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("getComputer", mock.Anything, mock.Anything).Return(testComputer, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectCreate(resourceLocalData, api)

		assert.NoError(t, err)
		assert.Equal(t, resourceLocalData.Id(), testComputer.dn)
	})
}

func TestResourceLDAPComputerObjectRead(t *testing.T) {
	name := "Test2"
	ou := "ou=test2,ou=org"
	description := "terraform"

	testComputer := &Computer{
		name: name,
		dn:   fmt.Sprintf("cn=%s,%s", name, ou),
		attributes: map[string][]string{
			"description": {description},
		},
	}

	resourceSchema := resourceLDAPComputerObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"ou":          ou,
		"description": "other desciption",
	}

	t.Run("resourceLDAPComputerObjectRead - should return nil when everything is good", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything).Return(testComputer, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectRead(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceLDAPComputerObjectRead - should return error when reading failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectRead(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceLDAPComputerObjectRead - should return nil and id set to nil when not found", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything).Return(nil, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectRead(resourceLocalData, api)

		assert.NoError(t, err)
		assert.Equal(t, resourceLocalData.Id(), "")
	})

	t.Run("resourceLDAPComputerObjectRead - should set 'description' accordingly", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything).Return(testComputer, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectRead(resourceLocalData, api)

		assert.NoError(t, err)
		assert.Equal(t, resourceLocalData.Get("description").(string), testComputer.attributes["description"][0])
	})
}

func TestResourceLDAPComputerObjectUpdate(t *testing.T) {
	name := "Test3"
	ou := "ou=test3,ou=org"
	description := "terraform"

	testComputer := &Computer{
		name: name,
		dn:   fmt.Sprintf("cn=%s,%s", name, ou),
		attributes: map[string][]string{
			"description": {"updated"},
		},
	}

	resourceSchema := resourceLDAPComputerObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"ou":          ou,
		"description": description,
	}

	t.Run("resourceLDAPComputerObjectUpdate - should return nil when everything is okay", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything).Return(testComputer, nil)
		api.On("updateComputerAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("updateComputerOU", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectUpdate(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceLDAPComputerObjectUpdate - should return error when updateComputerAttributes fails", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything).Return(testComputer, nil)
		api.On("updateComputerAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectUpdate(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceLDAPComputerObjectUpdate - should return error when updateComputerOU fails", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything).Return(testComputer, nil)
		api.On("updateComputerAttributes", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("updateComputerOU", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectUpdate(resourceLocalData, api)

		assert.Error(t, err)
	})
}

func TestResourceLDAPComputerObjectDelete(t *testing.T) {
	resourceSchema := resourceLDAPComputerObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        "test",
		"ou":          "ou",
		"description": "other desciption",
	}

	t.Run("resourceLDAPComputerObjectDelete - should forward errors from api.deleteComputer", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("deleteComputer", mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectDelete(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceLDAPComputerObjectDelete - should return nil if deleting was successful", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("deleteComputer", mock.Anything).Return(nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceLDAPComputerObjectDelete(resourceLocalData, api)

		assert.NoError(t, err)
	})
}
