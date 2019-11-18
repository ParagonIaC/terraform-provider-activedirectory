package ldap

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
