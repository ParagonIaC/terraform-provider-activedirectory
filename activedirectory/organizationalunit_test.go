package activedirectory

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"gopkg.in/ldap.v3"
)

func TestGetOU(t *testing.T) {
	numberOfObjects := 1
	numberOfAttributes := 1
	searchResult := createADResult(numberOfObjects, numberOfAttributes)

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
}

func TestCreateOU(t *testing.T) {
	t.Run("createOU - should set standard attributes for ou objects", func(t *testing.T) {
		_name := "Test"
		_desc := "Desc"

		matchFunc := func(sr *ldap.AddRequest) bool {
			ret := sr.DN == _name

			stdAttributes := map[string][]string{
				"name":        {_name},
				"ou":          {_name},
				"description": {_desc},
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
		mockClient.On("Add", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.createOU(_name, _name, _desc)
		assert.NoError(t, err)
	})
}

func TestMoveOU(t *testing.T) {
	t.Run("moveOU - should forward error from ldap.Client.ModifyDN", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.moveOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("moveOU - should return nil when ou was updated", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("ModifyDN", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.moveOU("", "", "")
		assert.NoError(t, err)
	})

	t.Run("moveOU - should call ModifyDN with correct ModifyDNrequest", func(t *testing.T) {
		_cn := "test"
		_dn := fmt.Sprintf("ou=%s,ou=server,ou=org", _cn)
		_ou := "ou=sub,ou=server,ou=org"

		matchFunc := func(sr *ldap.ModifyDNRequest) bool {
			return sr.DN == _dn && sr.NewSuperior == _ou && sr.NewRDN == fmt.Sprintf("ou=%s", _cn)
		}

		mockClient := new(MockClient)
		mockClient.On("ModifyDN", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}
		err := api.moveOU(_dn, _cn, _ou)
		assert.NoError(t, err)
	})
}

func TestUpdateOUDescription(t *testing.T) {
	t.Run("updateOUDescription - should forward error from ldap.client.Modify", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Modify", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateOUDescription("", "")

		assert.Error(t, err)
	})

	t.Run("updateOUDescription - should return nil when ou was updated successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Modify", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateOUDescription("", "")

		assert.NoError(t, err)
	})
}

func TestDeleteOU(t *testing.T) {
	numberOfObjects := 1
	numberOfAttributes := 1
	searchResult := createADResult(numberOfObjects, numberOfAttributes)

	t.Run("deleteOU - should forward error from api.deleteObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(&ldap.SearchResult{}, nil)
		mockClient.On("Del", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteOU("")

		assert.Error(t, err)
	})

	t.Run("deleteOU - should forward error from api.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteOU("")

		assert.Error(t, err)
	})

	t.Run("deleteOU - should return nil when object is deleted successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(&ldap.SearchResult{}, nil)
		mockClient.On("Del", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.deleteOU("")

		assert.NoError(t, err)
	})

	t.Run("deleteOU - should return error when ou has child items", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}
		err := api.deleteOU("")

		assert.Error(t, err)
	})
}
