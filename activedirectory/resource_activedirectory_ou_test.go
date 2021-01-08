package activedirectory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"strings"
	"testing"
	"time"
)

// acceptance tests
func TestAccADOU_basic(t *testing.T) {
	baseOu := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	name := strings.ToLower(fmt.Sprintf("testacc_%s", getRandomString(3)))
	dn := "ou=" + name + "," + baseOu
	description := getRandomString(10)

	var ouObject OU

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckOU(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckADOUDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADOUTestData(baseOu, name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADOUExists(ResourcesNameOrganizationUnit+".test", &ouObject),
					testAccCheckADOUAttributes(&ouObject, baseOu, name, description),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "base_ou", baseOu),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "name", name),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "description", description),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "id", dn),
				),
			},
		},
	})
}

func TestAccADOU_update(t *testing.T) {
	baseOu := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	name := strings.ToLower(fmt.Sprintf("testacc_%s", getRandomString(3)))
	description := getRandomString(10)

	updatedName := fmt.Sprintf("update_%s", name)
	updatedDescription := description + "_updated"

	var ouObject OU

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckOU(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckADOUDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADOUTestData(baseOu, name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADOUExists(ResourcesNameOrganizationUnit+".test", &ouObject),
					testAccCheckADOUAttributes(&ouObject, baseOu, name, description),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "base_ou", baseOu),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "name", name),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "description", description),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "id", fmt.Sprintf("ou=%s,%s", name, baseOu)),
				),
			},
			{
				Config: testAccResourceADOUTestData(baseOu, updatedName, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADOUExists(ResourcesNameOrganizationUnit+".test", &ouObject),
					testAccCheckADOUAttributes(&ouObject, baseOu, updatedName, updatedDescription),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "base_ou", baseOu),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "name", updatedName),
					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "description", updatedDescription),
				),
			},
		},
	})
}

// acceptance test helpers
func testAccCheckADOUDestroy(s *terraform.State) error {
	api, err := getTestConnection()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != ResourcesNameOrganizationUnit {
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

		api, err := getTestConnection()
		defer api.client.Close()
		if err != nil {
			return err
		}

		tmpOU, err := api.getOU(rs.Primary.Attributes["name"], rs.Primary.Attributes["base_ou"])
		timout := 5
		for tmpOU.dn == "" && timout > 0 && err == nil {
			tmpOU, err = api.getOU(rs.Primary.Attributes["name"], rs.Primary.Attributes["base_ou"])
			time.Sleep(1 * time.Second)
			timout--
		}
		if timout < 0 {
			return fmt.Errorf("OU was not created in AD")
		}

		if err != nil {
			return err
		}
		*ou = *tmpOU
		return nil
	}
}

//
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
	testConfig := fmt.Sprintf(`
%s

resource "%s" "test" {
	base_ou      = "%s"
	name         = "%s"
	description  = "%s"
}
	`,
		TerraformProviderRequestSection(), ResourcesNameOrganizationUnit, ou, name, description,
	)
	return testConfig
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
		err := resourceADOUObjectCreate(nil, resourceLocalData, api)

		assert.False(t, err.HasError())
	})

	t.Run("resourceADOUObjectCreate - should return error when creating failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createOU", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectCreate(nil, resourceLocalData, api)

		assert.True(t, err.HasError())
	})

	t.Run("resourceADOUObjectCreate - id should be set to dn", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createOU", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("updateOUName", mock.Anything, mock.Anything).Return(nil)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectCreate(nil, resourceLocalData, api)

		assert.False(t, err.HasError())
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
		err := resourceADOUObjectRead(nil, resourceLocalData, api)

		assert.False(t, err.HasError())
	})

	t.Run("resourceADOUObjectRead - should return error when reading failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectRead(nil, resourceLocalData, api)

		assert.True(t, err.HasError())
	})

	t.Run("resourceADOUObjectRead - should return nil and id set to nil when not found", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(nil, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectRead(nil, resourceLocalData, api)

		assert.False(t, err.HasError())
		assert.Equal(t, resourceLocalData.Id(), "")
	})

	t.Run("resourceADOUObjectRead - should set 'description' accordingly", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectRead(nil, resourceLocalData, api)

		assert.False(t, err.HasError())
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
		err := resourceADOUObjectUpdate(nil, resourceLocalData, api)

		assert.False(t, err.HasError())
	})

	t.Run("resourceADOUObjectUpdate - should return error when updateOUDescription fails", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)
		api.On("updateOUDescription", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectUpdate(nil, resourceLocalData, api)

		assert.True(t, err.HasError())
	})

	t.Run("resourceADOUObjectUpdate - should return error when moveOU fails", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)
		api.On("updateOUName", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("updateOUDescription", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("moveOU", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectUpdate(nil, resourceLocalData, api)

		assert.True(t, err.HasError())
	})

	t.Run("resourceADOUObjectUpdate - should return error when updateOUName fails", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getOU", mock.Anything, mock.Anything).Return(testOU, nil)
		api.On("updateOUName", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
		api.On("updateOUDescription", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectUpdate(nil, resourceLocalData, api)

		assert.True(t, err.HasError())
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
		err := resourceADOUObjectDelete(nil, resourceLocalData, api)

		assert.True(t, err.HasError())
	})

	t.Run("resourceADOUObjectDelete - should return nil if deleting was successful", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("deleteOU", mock.Anything).Return(nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADOUObjectDelete(nil, resourceLocalData, api)

		assert.False(t, err.HasError())
	})
}
