package activedirectory

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestGetOU(t *testing.T) {
	numberOfObjects := 1
	attributes := []string{"ou", "description"}
	searchResult := createADResult(numberOfObjects, attributes)

	t.Run("getOU - should forward errors from api.getObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}

		ou, err := api.getOU("", "")

		assert.Error(t, err)
		assert.Nil(t, ou)
	})

	t.Run("getOU - should return nil when no ou was found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, &ldap.Error{Err: fmt.Errorf("not found"), ResultCode: 32})

		api := &API{client: mockClient}

		ou, err := api.getOU("", "")

		assert.NoError(t, err)
		assert.Nil(t, ou)
	})

	t.Run("getOU - should return ou object", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}

		ou, err := api.getOU("", "")

		assert.NoError(t, err)
		assert.NotNil(t, ou)
		assert.IsType(t, &OU{}, ou)
	})

	t.Run("getOU - should error when more than one object is found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(2, attributes), nil)

		api := &API{client: mockClient}

		ou, err := api.getOU("", "")

		assert.Error(t, err)
		assert.Nil(t, ou)
	})

	t.Run("getOU - should return nil when api.client.Search returns nil", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}

		ou, err := api.getOU("", "")

		assert.NoError(t, err)
		assert.Nil(t, ou)
	})

	t.Run("getOU - should return nil when api.client.Search returns 0 objects", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(0, attributes), nil)

		api := &API{client: mockClient}

		ou, err := api.getOU("", "")

		assert.NoError(t, err)
		assert.Nil(t, ou)
	})
}

func TestCreateOU(t *testing.T) {
	attributes := []string{"ou", "description"}

	t.Run("createOU - should forward errors from api.client.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}

		err := api.createOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("createOU - should error when ou already exists in another place", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)

		api := &API{client: mockClient}

		err := api.createOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("createOU - should update description when exact object was found", func(t *testing.T) {
		res := createADResult(1, attributes)
		name := res.Entries[0].GetAttributeValue("ou")
		description := res.Entries[0].GetAttributeValue("description")
		baseOU := getRandomOU(2, 2)
		res.Entries[0].DN = fmt.Sprintf("OU=%s,%s", name, baseOU)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(res, nil)
		matchFunc := func(sr *ldap.ModifyRequest) bool {
			for _, elem := range sr.Changes {
				if elem.Modification.Type == "description" {
					return true
				}
			}

			return false
		}
		mockClient.On("Modify", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.createOU(name, baseOU, description)
		assert.NoError(t, err)
	})

	t.Run("createOU - should set standard attributes for ou objects", func(t *testing.T) {
		name := getRandomString(10)
		ou := getRandomOU(2, 3)
		desc := getRandomString(10)

		matchFunc := func(sr *ldap.AddRequest) bool {
			ret := sr.DN == fmt.Sprintf("ou=%s,%s", name, ou)

			stdAttributes := map[string][]string{
				"ou":          {name},
				"description": {desc},
			}

			found := 0
			for _, e := range sr.Attributes {
				if _, ok := stdAttributes[e.Type]; ok {
					found++
					ret = ret && reflect.DeepEqual(stdAttributes[e.Type], e.Vals)
				}

				if e.Type == "objectClass" {
					ret = ret && contains(e.Vals, "organizationalUnit")
				}
			}

			ret = ret && (found == len(stdAttributes))

			return ret
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)
		mockClient.On("Add", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.createOU(name, ou, desc)
		assert.NoError(t, err)
	})
}

func TestMoveOU(t *testing.T) {
	attributes := []string{"ou", "description"}

	t.Run("moveOU - should forward error from ldap.Client.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.moveOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("moveOU - should error when ou was not found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}
		err := api.moveOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("moveOU - should forward error from ldap.Client.ModifyDN", func(t *testing.T) {
		res := createADResult(1, attributes)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(res, nil)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.moveOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("moveOU - should return nil when ou was updated", func(t *testing.T) {
		res := createADResult(1, attributes)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(res, nil)
		mockClient.On("ModifyDN", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.moveOU("", "", "")
		assert.NoError(t, err)
	})

	t.Run("moveOU - should call ModifyDN with correct ModifyDNrequest", func(t *testing.T) {
		res := createADResult(1, attributes)

		cn := getRandomString(10)
		ou := getRandomOU(3, 2)
		newOU := fmt.Sprintf("ou=%s,%s", getRandomString(10), ou)

		matchFunc := func(sr *ldap.ModifyDNRequest) bool {
			return sr.DN == fmt.Sprintf("ou=%s,%s", cn, ou) &&
				sr.NewSuperior == newOU && sr.NewRDN == fmt.Sprintf("ou=%s", cn)
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(res, nil)
		mockClient.On("ModifyDN", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}
		err := api.moveOU(cn, ou, newOU)
		assert.NoError(t, err)
	})

	t.Run("moveOU - should do nothing when ou is already located under the target ou", func(t *testing.T) {
		res := createADResult(1, attributes)

		cn := res.Entries[0].GetAttributeValue("ou")
		ou := getRandomOU(2, 3)
		newOU := fmt.Sprintf("ou=%s,%s", getRandomString(10), ou)
		res.Entries[0].DN = fmt.Sprintf("ou=%s,%s", cn, newOU)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(res, nil)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.moveOU(cn, ou, newOU)
		assert.NoError(t, err)
	})
}

func TestUpdateOUDescription(t *testing.T) {
	attributes := []string{"ou", "description"}

	t.Run("updateOUDescription - should forward error from ldap.client.Modify", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)
		mockClient.On("Modify", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateOUDescription("", "", "")

		assert.Error(t, err)
	})

	t.Run("updateOUDescription - should return nil when ou was updated successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)
		mockClient.On("Modify", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateOUDescription("", "", "")

		assert.NoError(t, err)
	})

	t.Run("updateOUDescription - should modify description", func(t *testing.T) {
		matchFunc := func(req *ldap.ModifyRequest) bool {
			for i := 0; i < len(req.Changes); i++ {
				if req.Changes[i].Modification.Type == "description" {
					return true
				}
			}

			return false
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)
		mockClient.On("Modify", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.updateOUDescription("", "", "")

		assert.NoError(t, err)
	})
}

func TestUpdateOUName(t *testing.T) {
	attributes := []string{"ou", "description"}

	t.Run("updateOUName - should forward error from ldap.Client.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateOUName("a", "b", "c")
		assert.Error(t, err)
	})

	t.Run("updateOUName - should error when ou was not found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}
		err := api.updateOUName("a", "b", "c")
		assert.Error(t, err)
	})

	t.Run("updateOUName - should forward error from ldap.client.ModifyDN", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateOUName("b", "c", "d")

		assert.Error(t, err)
	})

	t.Run("updateOUName - should return nil when ou was updated successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)
		mockClient.On("ModifyDN", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateOUName("c", "d", "e")

		assert.NoError(t, err)
	})
}

func TestDeleteOU(t *testing.T) {
	numberOfObjects := 1
	attributes := []string{"ou", "description"}
	searchResult := createADResult(numberOfObjects, attributes)

	t.Run("deleteOU - should forward error from api.searchObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteOU("test1")

		assert.Error(t, err)
	})

	t.Run("deleteOU - should forward error from api.deleteObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)
		mockClient.On("Del", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteOU("test2")

		assert.Error(t, err)
	})

	t.Run("deleteOU - should return nil when object is deleted successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)
		mockClient.On("Del", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.deleteOU("test3")

		assert.NoError(t, err)
	})

	t.Run("deleteOU - should return error when ou has child items", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}
		err := api.deleteOU("test4")

		assert.Error(t, err)
	})
}
