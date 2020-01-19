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
func TestAccADGroup_basic(t *testing.T) {
	ou := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	userBase := strings.ToLower(os.Getenv("AD_TEST_USER_BASE"))
	name := strings.ToLower(fmt.Sprintf("testacc_%s", getRandomString(3)))
	description := getRandomString(10)

	var ouGroup Group

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckOU(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckADOUDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADGroupTestData(ou, name, description, userBase, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADGroupExists("activedirectory_group.test", &ouGroup),
					testAccCheckADGroupAttributes(&ouGroup, ou, name, description),
					resource.TestCheckResourceAttr("activedirectory_group.test", "base_ou", ou),
					resource.TestCheckResourceAttr("activedirectory_group.test", "name", name),
					resource.TestCheckResourceAttr("activedirectory_group.test", "description", description),
					resource.TestCheckResourceAttr("activedirectory_group.test", "id", fmt.Sprintf("cn=%s,%s", name, ou)),
				),
			},
		},
	})
}

func TestAccADGroup_update(t *testing.T) {
	ou := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	name := strings.ToLower(fmt.Sprintf("testacc_%s", getRandomString(3)))
	description := getRandomString(10)

	updatedName := fmt.Sprintf("update_%s", name)
	updatedOU := "cn=update," + ou
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

func testAccCheckADGroupExists(resourceName string, group *Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AD group ID is not set")
		}

		api := testAccProvider.Meta().(*API)
		tmpOU, err := api.getGroup(rs.Primary.Attributes["name"],
			rs.Primary.Attributes["base_ou"],
			rs.Primary.Attributes["user_base"],
			[]string{},
			false)

		if err != nil {
			return err
		}

		*group = *tmpOU
		return nil
	}
}

func testAccCheckADGroupAttributes(groupObject *Group, ou, name, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.EqualFold(groupObject.name, name) {
			return fmt.Errorf("group name not set correctly: %s, %s", groupObject.name, name)
		}

		if !strings.EqualFold(groupObject.description, description) {
			return fmt.Errorf("group description not set correctly: %s, %s", groupObject.description, description)
		}

		if !strings.EqualFold(groupObject.dn, fmt.Sprintf("cn=%s,%s", name, ou)) {
			return fmt.Errorf("group dn not set correctly: %s, %s, %s", groupObject.dn, name, ou)
		}

		return nil
	}
}

// acceptance test data
func testAccResourceADGroupTestData(ou, name, description, userBase string, ignoreMembersUnknownByTerraform bool) string {
	return fmt.Sprintf(`
resource "activedirectory_group" "test" {
	base_ou      					= "%s"
	name         					= "%s"
	description  					= "%s"
    user_base    					= "%s"
	member 		 					= []
    ignoreMembersUnknownByTerraform = %t
}
	`,
		ou, name, description, userBase, ignoreMembersUnknownByTerraform,
	)
}

// unit tests
func TestResourceADGroupObject(t *testing.T) {
	t.Run("resourceADGroupObject - should return *schema.Resource", func(t *testing.T) {
		response := resourceADGroupObject()
		assert.IsType(t, &schema.Resource{}, response)

		assert.Equal(t, schema.TypeString, response.Schema["name"].Type)
		assert.Equal(t, true, response.Schema["name"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["base_ou"].Type)
		assert.Equal(t, true, response.Schema["base_ou"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["description"].Type)
		assert.Equal(t, false, response.Schema["description"].Required)

		assert.Equal(t, schema.TypeString, response.Schema["user_base"].Type)
		assert.Equal(t, true, response.Schema["user_base"].Required)

		assert.Equal(t, schema.TypeSet, response.Schema["member"].Type)
		assert.Equal(t, false, response.Schema["member"].Required)

		assert.Equal(t, schema.TypeBool, response.Schema["ignore_members_unknown_by_terraform"].Type)
		assert.Equal(t, false, response.Schema["ignore_members_unknown_by_terraform"].Required)

	})
}

func TestResourceADGroupObjectCreate(t *testing.T) {
	name := getRandomString(10)
	ou := getRandomOU(2, 2)
	description := getRandomString(10)
	member := []string{"somebody"}

	testGroup := &Group{
		name:        name,
		dn:          fmt.Sprintf("cn=%s,%s", name, ou),
		description: description,
		member:      member,
	}

	resourceSchema := resourceADGroupObject().Schema
	resourceDataMap := map[string]interface{}{
		"name":        name,
		"base_ou":     ou,
		"description": description,
		"member":      make([]interface{}, 0),
	}

	t.Run("resourceADGroupObjectCreate - should return nil when everything is good", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		api.On("getGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(testGroup, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADGroupObjectCreate(resourceLocalData, api)

		assert.NoError(t, err)
	})

	t.Run("resourceADGroupObjectCreate - should return error when creating failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADGroupObjectCreate(resourceLocalData, api)

		assert.Error(t, err)
	})

	t.Run("resourceADGroupObjectCreate - id should be set to dn", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		api.On("renameGroup", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("getGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(testGroup, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		err := resourceADGroupObjectCreate(resourceLocalData, api)

		assert.NoError(t, err)
		assert.True(t, strings.EqualFold(resourceLocalData.Id(), testGroup.dn))
	})
}
