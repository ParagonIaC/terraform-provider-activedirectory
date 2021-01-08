package activedirectory

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// acceptance tests
//func TestAccADGroup_basic_v2(t *testing.T) {
//	baseOu := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
//	userBase := strings.ToLower(os.Getenv("AD_TEST_USER_BASE"))
//	testUsersAMAccountName := strings.ToLower(os.Getenv("AD_TEST_USER_1_sAMAccountName"))
//	name := strings.ToLower(fmt.Sprintf("testacc_%s", getRandomString(3)))
//	description := getRandomString(10)
//
//	var ouObject OU
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:          func() { testAccPreCheckOU(t) },
//		ProviderFactories: testAccProviderFactories,
//		CheckDestroy:      testAccCheckADOUDestroy,
//		Steps: []resource.TestStep{
//			{
//				Config: testAccResourceADOUTestData(baseOu, name, description),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckADOUExists(ResourcesNameOrganizationUnit+".test", &ouObject),
//					testAccCheckADOUAttributes(&ouObject, baseOu, name, description),
//					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "base_ou", baseOu),
//					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "name", name),
//					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "description", description),
//					resource.TestCheckResourceAttr(ResourcesNameOrganizationUnit+".test", "id", dn),
//				),
//			},
//		},
//	})
//}

func TestAccADGroup_basic(t *testing.T) {
	baseOU := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	userBase := strings.ToLower(os.Getenv("AD_TEST_USER_BASE"))
	name := strings.ToLower(fmt.Sprintf("testacc_%s", getRandomString(3)))
	description := getRandomString(10)

	var groupObject Group

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckGroup(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckADGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADGroupTestData(baseOU, name, description, userBase, []string{},false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADGroupExists(ResourcesNameGroup+".test", &groupObject),
					testAccCheckADGroupAttributes(&groupObject, baseOU, name, description),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "base_ou", baseOU),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "name", name),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "description", description),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "id", fmt.Sprintf("cn=%s,%s", name, baseOU)),
				),
			},
		},
	})
}

func TestAccADGroup_update(t *testing.T) {
	baseOU := strings.ToLower(os.Getenv("AD_TEST_BASE_OU"))
	userBase := strings.ToLower(os.Getenv("AD_TEST_USER_BASE"))
	name := strings.ToLower(fmt.Sprintf("testacc_%s", getRandomString(3)))
	description := getRandomString(10)

	updatedName := fmt.Sprintf("update_%s", name)
	updatedDescription := description + "_updated"

	var groupObject Group

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheckGroup(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckADOUDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADGroupTestData(baseOU, name, description, userBase, []string{},false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADGroupExists(ResourcesNameGroup+".test", &groupObject),
					testAccCheckADGroupAttributes(&groupObject, baseOU, name, description),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "base_ou", baseOU),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "name", name),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "description", description),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "id", fmt.Sprintf("cn=%s,%s", name, baseOU)),
				),
			},
			{
				Config: testAccResourceADGroupTestData(baseOU, updatedName, updatedDescription, userBase,[]string{}, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADGroupExists(ResourcesNameGroup+".test", &groupObject),
					testAccCheckADGroupAttributes(&groupObject, baseOU, updatedName, updatedDescription),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "base_ou", baseOU),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "name", updatedName),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "description", updatedDescription),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "id", fmt.Sprintf("cn=%s,%s", updatedName, baseOU)),

				),
			},
			{
				Config: testAccResourceADGroupTestData(baseOU, updatedName, updatedDescription, userBase, []string{},false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADGroupExists(ResourcesNameGroup+".test", &groupObject),
					testAccCheckADGroupAttributes(&groupObject, baseOU, updatedName, updatedDescription),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "base_ou", baseOU),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "name", updatedName),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "description", updatedDescription),
					resource.TestCheckResourceAttr(ResourcesNameGroup+".test", "id", fmt.Sprintf("cn=%s,%s", updatedName, baseOU)),

				),
			},
		},
	})
}

// acceptance test helpers
func testAccPreCheckGroup(t *testing.T) {
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
	if v := os.Getenv("AD_TEST_USER_BASE"); v == "" {
		t.Fatal("AD_TEST_USER_BASE must be set for acceptance tests")
	}
}

func testAccCheckADGroupDestroy(s *terraform.State) error {
	api, err := getTestConnection()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != ResourcesNameGroup {
			continue
		}

		ou, err := api.getGroup(rs.Primary.Attributes["name"], rs.Primary.Attributes["base_ou"], "", []string{}, false)
		if err != nil {
			return err
		}

		if ou != nil {
			return fmt.Errorf("ad group (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
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

		api, err := getTestConnection()
		if err != nil {
			return err
		}
		tmpGroup, err := api.getGroup(rs.Primary.Attributes["name"],
			rs.Primary.Attributes["base_ou"],
			rs.Primary.Attributes["user_base"],
			[]string{},
			false)

		if err != nil {
			return err
		}

		*group = *tmpGroup
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
func testAccResourceADGroupTestData(ou, name, description, userBase string, users []string, ignoreMembersUnknownByTerraform bool) string {
	usersString := fmt.Sprintf("[ %s ]", strings.Join(users, ","))
	return fmt.Sprintf(`
%s

resource "%s" "test" {
	base_ou      					= "%s"
	name         					= "%s"
	description  					= "%s"
    user_base    					= "%s"
	member 		 					=  %s
    ignore_members_unknown_by_terraform = %t
}
	`,
		TerraformProviderRequestSection(), ResourcesNameGroup, ou, name, description, userBase, usersString, ignoreMembersUnknownByTerraform,
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
		assert.Equal(t, false, response.Schema["user_base"].Required)

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
		diags := resourceADGroupObjectCreate(nil, resourceLocalData, api)

		assert.False(t, diags.HasError())
	})

	t.Run("resourceADGroupObjectCreate - should return error when creating failed", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(fmt.Errorf("error"))

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		diags := resourceADGroupObjectCreate(nil, resourceLocalData, api)

		assert.True(t, diags.HasError())
	})

	t.Run("resourceADGroupObjectCreate - id should be set to dn", func(t *testing.T) {
		api := new(MockAPIInterface)
		api.On("createGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil)
		api.On("renameGroup", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		api.On("getGroup", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(testGroup, nil)

		resourceLocalData := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
		diags := resourceADGroupObjectCreate(nil, resourceLocalData, api)

		assert.False(t, diags.HasError())
		assert.True(t, strings.EqualFold(resourceLocalData.Id(), testGroup.dn))
	})
}
