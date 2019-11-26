package activedirectory

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// acceptance tests
func TestAccADOU_basic(t *testing.T) {
	ou := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	name := strings.ToLower(fmt.Sprintf("testacc_%s", getRandomString(3)))
	description := getRandomString(10)

	var ouObject OU

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckOU(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckADOUDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADOUTestData(ou, name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADOUExists("activedirectory_ou.test", &ouObject),
					testAccCheckADOUAttributes(&ouObject, ou, name, description),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "base_ou", ou),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "name", name),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "description", description),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "id", fmt.Sprintf("ou=%s,%s", name, ou)),
				),
			},
		},
	})
}

func TestAccADOU_update(t *testing.T) {
	ou := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	name := strings.ToLower(fmt.Sprintf("testacc_%s", getRandomString(3)))
	description := getRandomString(10)

	updatedName := fmt.Sprintf("update_%s", name)
	updatedOU := "ou=update," + ou
	updatedDescription := description + "_updated"

	var ouObject OU

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckOU(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckADOUDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADOUTestData(ou, name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADOUExists("activedirectory_ou.test", &ouObject),
					testAccCheckADOUAttributes(&ouObject, ou, name, description),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "base_ou", ou),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "name", name),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "description", description),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "id", fmt.Sprintf("ou=%s,%s", name, ou)),
				),
			},
			{
				Config: testAccResourceADOUTestData(updatedOU, updatedName, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADOUExists("activedirectory_ou.test", &ouObject),
					testAccCheckADOUAttributes(&ouObject, updatedOU, updatedName, updatedDescription),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "base_ou", updatedOU),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "name", updatedName),
					resource.TestCheckResourceAttr("activedirectory_ou.test", "description", updatedDescription),
				),
			},
		},
	})
}

// acceptance test helpers
func testAccCheckADOUDestroy(s *terraform.State) error {
	api := testAccProvider.Meta().(*API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "activedirectory_ou" {
			continue
		}

		ou, err := api.getOU(rs.Primary.Attributes["name"], rs.Primary.Attributes["base_ou"])
		if err != nil {
			return err
		}

		if ou != nil {
			return fmt.Errorf("ad ou (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckADOUExists(resourceName string, ou *OU) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AD ou ID is not set")
		}

		api := testAccProvider.Meta().(*API)
		tmpOU, err := api.getOU(rs.Primary.Attributes["name"], rs.Primary.Attributes["base_ou"])

		if err != nil {
			return err
		}

		*ou = *tmpOU
		return nil
	}
}

func testAccCheckADOUAttributes(ouObject *OU, ou, name, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.EqualFold(ouObject.name, name) {
			return fmt.Errorf("ou name not set correctly: %s, %s", ouObject.name, name)
		}

		if !strings.EqualFold(ouObject.description, description) {
			return fmt.Errorf("ou description not set correctly: %s, %s", ouObject.description, description)
		}

		if !strings.EqualFold(ouObject.dn, fmt.Sprintf("ou=%s,%s", name, ou)) {
			return fmt.Errorf("ou dn not set correctly: %s, %s, %s", ouObject.dn, name, ou)
		}

		return nil
	}
}

func testAccPreCheckOU(t *testing.T) {
	if v := os.Getenv("AD_HOST"); v == "" {
		t.Fatal("AD_HOST must be set for acceptance tests")
	}
	if v := os.Getenv("AD_PORT"); v == "" {
		t.Fatal("AD_PORT must be set for acceptance tests")
	}
	if v := os.Getenv("AD_DOMAIN"); v == "" {
		t.Fatal("AD_DOMAIN must be set for acceptance tests")
	}
	if v := os.Getenv("AD_USE_TLS"); v == "" {
		t.Fatal("AD_USE_TLS must be set for acceptance tests")
	}
	if v := os.Getenv("AD_USER"); v == "" {
		t.Fatal("AD_USER must be set for acceptance tests")
	}
	if v := os.Getenv("AD_PASSWORD"); v == "" {
		t.Fatal("AD_PASSWORD must be set for acceptance tests")
	}
	if v := os.Getenv("AD_TEST_BASE_OU"); v == "" {
		t.Fatal("AD_TEST_BASE_OU must be set for acceptance tests")
	}
}

// acceptance test data
func testAccResourceADOUTestData(ou, name, description string) string {
	return fmt.Sprintf(`
resource "activedirectory_ou" "test" {
	base_ou      = "%s"
	name         = "%s"
	description  = "%s"
}
	`,
		ou, name, description,
	)
}

// unit tests
func TestResourceADOUObject(t *testing.T) {
	t.Run("resourceADOUObject - should return *schema.Resource", func(t *testing.T) {
		response := resourceADOUObject()
		assert.IsType(t, &schema.Resource{}, response)

		assert.Equal(t, schema.TypeString, response.Schema["name"].Type)
		assert.Equal(t, true, response.Schema["name"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["base_ou"].Type)
		assert.Equal(t, true, response.Schema["base_ou"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["description"].Type)
		assert.Equal(t, false, response.Schema["description"].Required)
	})
}

func TestResourceADOUObjectCreate(t *testing.T) {
	name := getRandomString(10)
	ou := getRandomOU(2, 2)
	description := getRandomString(10)

	testOU := &OU{
		name:        name,
		dn:          fmt.Sprintf("ou=%s,%s", name, ou),
		description: description,
	}

	resourceSchema := resourceADOUObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"base_ou":     ou,
		"description": description,
	}

	t.Run("resourceADOUObjectCreate - should return nil when everything is good", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createOU", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectCreate(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceADOUObjectCreate - should return error when creating failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createOU", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectCreate(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceADOUObjectCreate - id should be set to dn", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createOU", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("updateOUName", mock.Anything, mock.Anything).Return(nil)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectCreate(resourceLocalData, api)

		assert.NoError(t, err)
		assert.True(t, strings.EqualFold(resourceLocalData.Id(), testOU.dn))
	})
}

func TestResourceADOUObjectRead(t *testing.T) {
	name := getRandomString(10)
	ou := getRandomOU(2, 2)
	description := getRandomString(10)

	testOU := &OU{
		name:        name,
		dn:          fmt.Sprintf("ou=%s,%s", name, ou),
		description: description,
	}

	resourceSchema := resourceADOUObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        "",
		"base_ou":     "",
		"description": "",
	}

	t.Run("resourceADOUObjectRead - should return nil when everything is good", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectRead(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceADOUObjectRead - should return error when reading failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectRead(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceADOUObjectRead - should return nil and id set to nil when not found", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(nil, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectRead(resourceLocalData, api)

		assert.NoError(t, err)
		assert.Equal(t, resourceLocalData.Id(), "")
	})

	t.Run("resourceADOUObjectRead - should set 'description' accordingly", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectRead(resourceLocalData, api)

		assert.NoError(t, err)
		assert.Equal(t, resourceLocalData.Get("description").(string), testOU.description)
	})
}

func TestResourceADOUObjectUpdate(t *testing.T) {
	name := getRandomString(10)
	ou := getRandomOU(3, 2)
	description := getRandomString(10)

	testOU := &OU{
		name:        name,
		dn:          fmt.Sprintf("ou=%s,%s", name, ou),
		description: getRandomString(20),
	}

	resourceSchema := resourceADOUObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"base_ou":     ou,
		"description": description,
	}

	t.Run("resourceADOUObjectUpdate - should return nil when everything is okay", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)
		api.On("updateOUDescription", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("updateOUName", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("moveOU", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectUpdate(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceADOUObjectUpdate - should return error when updateOUDescription fails", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)
		api.On("updateOUDescription", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectUpdate(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceADOUObjectUpdate - should return error when moveOU fails", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)
		api.On("updateOUName", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("updateOUDescription", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("moveOU", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectUpdate(resourceLocalData, api)

		assert.Error(t, err)
	})
}

func TestResourceADOUObjectDelete(t *testing.T) {
	name := getRandomString(10)
	ou := getRandomOU(2, 2)
	description := getRandomString(10)

	resourceSchema := resourceADOUObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"base_ou":     ou,
		"description": description,
	}

	t.Run("resourceADOUObjectDelete - should forward errors from api.deleteOU", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("deleteOU", mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectDelete(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceADOUObjectDelete - should return nil if deleting was successful", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("deleteOU", mock.Anything).Return(nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectDelete(resourceLocalData, api)

		assert.NoError(t, err)
	})
}
