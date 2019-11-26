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
func TestAccADComputer_basic(t *testing.T) {
	ou := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	name := strings.ToLower(fmt.Sprintf("testacc-%s", getRandomString(3)))
	description := getRandomString(10)

	var computer Computer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckADComputerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADComputerTestData(ou, name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADComputerExists("activedirectory_computer.test", &computer),
					testAccCheckADComputerAttributes(&computer, ou, name, description),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "ou", ou),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "name", name),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "description", description),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "id", fmt.Sprintf("cn=%s,%s", name, ou)),
				),
			},
		},
	})
}

func TestAccADComputer_update(t *testing.T) {
	ou := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	name := strings.ToLower(fmt.Sprintf("testacc-%s", getRandomString(3)))
	description := getRandomString(10)

	updatedOU := "ou=update," + ou
	updatedDescription := description + "_updated"

	var computer Computer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckADComputerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADComputerTestData(ou, name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADComputerExists("activedirectory_computer.test", &computer),
					testAccCheckADComputerAttributes(&computer, ou, name, description),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "ou", ou),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "name", name),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "description", description),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "id", fmt.Sprintf("cn=%s,%s", name, ou)),
				),
			},
			{
				Config: testAccResourceADComputerTestData(updatedOU, name, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADComputerExists("activedirectory_computer.test", &computer),
					testAccCheckADComputerAttributes(&computer, updatedOU, name, updatedDescription),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "ou", updatedOU),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "description", updatedDescription),
				),
			},
		},
	})
}

// acceptance test helpers
func testAccCheckADComputerDestroy(s *terraform.State) error {
	api := testAccProvider.Meta().(*API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "activedirectory_computer" {
			continue
		}

		computer, err := api.getComputer(rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		if computer != nil {
			return fmt.Errorf("ad computer (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckADComputerExists(resourceName string, computer *Computer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AD computer ID is not set")
		}

		api := testAccProvider.Meta().(*API)
		_computer, err := api.getComputer(rs.Primary.Attributes["name"])

		if err != nil {
			return err
		}

		*computer = *_computer
		return nil
	}
}

func testAccCheckADComputerAttributes(computer *Computer, ou, name, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.EqualFold(computer.name, name) {
			return fmt.Errorf("computer name not set correctly: %s, %s", computer.name, name)
		}

		if !strings.EqualFold(computer.description, description) {
			return fmt.Errorf("computer description not set correctly: %s, %s", computer.description, description)
		}

		if !strings.EqualFold(computer.dn, fmt.Sprintf("cn=%s,%s", name, ou)) {
			return fmt.Errorf("computer dn not set correctly: %s, %s, %s", computer.dn, name, ou)
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
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
func testAccResourceADComputerTestData(ou, name, description string) string {
	return fmt.Sprintf(`
resource "activedirectory_computer" "test" {
	ou           = "%s"
	name         = "%s"
	description  = "%s"
}
	`,
		ou, name, description,
	)
}

// unit tests
func TestResourceADComputerObject(t *testing.T) {
	t.Run("resourceADComputerObject - should return *schema.Resource", func(t *testing.T) {
		response := resourceADComputerObject()
		assert.IsType(t, &schema.Resource{}, response)

		assert.Equal(t, schema.TypeString, response.Schema["name"].Type)
		assert.Equal(t, true, response.Schema["name"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["ou"].Type)
		assert.Equal(t, true, response.Schema["ou"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["description"].Type)
		assert.Equal(t, false, response.Schema["description"].Required)
	})
}

func TestResourceADComputerObjectCreate(t *testing.T) {
	name := getRandomString(10)
	ou := getRandomOU(2, 2)
	description := getRandomString(10)

	testComputer := &Computer{
		name:        name,
		dn:          fmt.Sprintf("cn=%s,%s", name, ou),
		description: description,
	}

	resourceSchema := resourceADComputerObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"ou":          ou,
		"description": description,
	}

	t.Run("resourceADComputerObjectCreate - should return nil when everything is good", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createComputer", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("getComputer", mock.Anything, mock.Anything, mock.Anything).Return(testComputer, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectCreate(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceADComputerObjectCreate - should return error when creating failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createComputer", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectCreate(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceADComputerObjectCreate - id should be set to dn", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createComputer", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("getComputer", mock.Anything, mock.Anything, mock.Anything).Return(testComputer, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectCreate(resourceLocalData, api)

		assert.NoError(t, err)
		assert.True(t, strings.EqualFold(resourceLocalData.Id(), testComputer.dn))
	})
}

func TestResourceADComputerObjectRead(t *testing.T) {
	name := getRandomString(10)
	ou := getRandomOU(2, 2)
	description := getRandomString(10)

	testComputer := &Computer{
		name:        name,
		dn:          fmt.Sprintf("cn=%s,%s", name, ou),
		description: description,
	}

	resourceSchema := resourceADComputerObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        "",
		"ou":          "",
		"description": "",
	}

	t.Run("resourceADComputerObjectRead - should return nil when everything is good", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything, mock.Anything).Return(testComputer, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectRead(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceADComputerObjectRead - should return error when reading failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectRead(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceADComputerObjectRead - should return nil and id set to nil when not found", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectRead(resourceLocalData, api)

		assert.NoError(t, err)
		assert.Equal(t, resourceLocalData.Id(), "")
	})

	t.Run("resourceADComputerObjectRead - should set 'description' accordingly", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything, mock.Anything).Return(testComputer, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectRead(resourceLocalData, api)

		assert.NoError(t, err)
		assert.Equal(t, resourceLocalData.Get("description").(string), testComputer.description)
	})
}

func TestResourceADComputerObjectUpdate(t *testing.T) {
	name := getRandomString(10)
	ou := getRandomOU(3, 2)
	description := getRandomString(10)

	testComputer := &Computer{
		name:        name,
		dn:          fmt.Sprintf("cn=%s,%s", name, ou),
		description: getRandomString(10),
	}

	resourceSchema := resourceADComputerObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"ou":          ou,
		"description": description,
	}

	t.Run("resourceADComputerObjectUpdate - should return nil when everything is okay", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything, mock.Anything).Return(testComputer, nil)
		api.On("updateComputerDescription", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("updateComputerOU", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectUpdate(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceADComputerObjectUpdate - should return error when updateComputerDescription fails", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything, mock.Anything).Return(testComputer, nil)
		api.On("updateComputerDescription", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectUpdate(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceADComputerObjectUpdate - should return error when updateComputerOU fails", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("getComputer", mock.Anything, mock.Anything, mock.Anything).Return(testComputer, nil)
		api.On("updateComputerDescription", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("updateComputerOU", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectUpdate(resourceLocalData, api)

		assert.Error(t, err)
	})
}

func TestResourceADComputerObjectDelete(t *testing.T) {
	name := getRandomString(10)
	ou := getRandomOU(2, 3)
	description := getRandomString(10)

	resourceSchema := resourceADComputerObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"ou":          ou,
		"description": description,
	}

	t.Run("resourceADComputerObjectDelete - should forward errors from api.deleteComputer", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("deleteComputer", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectDelete(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceADComputerObjectDelete - should return nil if deleting was successful", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("deleteComputer", mock.Anything, mock.Anything).Return(nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADComputerObjectDelete(resourceLocalData, api)

		assert.NoError(t, err)
	})
}
