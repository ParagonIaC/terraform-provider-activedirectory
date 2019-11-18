package ldap

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"gopkg.in/ldap.v3"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestGetComputer(t *testing.T) {
	numberOfObjects := 1
	numberOfAttributes := 3
	searchResult := createLDAPResult(numberOfObjects, numberOfAttributes)

	t.Run("getComputer - should forward errors from api.getObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}

		computer, err := api.getComputer("", nil)

		assert.Error(t, err)
		assert.Nil(t, computer)
	})

	t.Run("getComputer - should return nil when no computer was found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, &ldap.Error{Err: fmt.Errorf("not found"), ResultCode: 32})

		api := &API{client: mockClient}

		computer, err := api.getComputer("", nil)

		assert.NoError(t, err)
		assert.Nil(t, computer)
	})

	t.Run("getComputer - should return computer object", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}

		computer, err := api.getComputer("", []string{"cn"})

		assert.NoError(t, err)
		assert.NotNil(t, computer)
		assert.IsType(t, &Computer{}, computer)
	})

	t.Run("getComputer - should add 'cn' to the list of attributes", func(t *testing.T) {
		matchFunc := func(sr *ldap.SearchRequest) bool {
			return contains(sr.Attributes, "cn")
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.MatchedBy(matchFunc)).Return(searchResult, nil)

		api := &API{client: mockClient}

		computer, err := api.getComputer("DN0", nil)
		assert.NoError(t, err)
		assert.NotNil(t, computer)

		assert.Equal(t, computer.dn, "DN0")
	})
}

func TestCreateComputer(t *testing.T) {
	t.Run("createComputer - should set standard attributes for computer objects", func(t *testing.T) {
		_name := "Test"
		matchFunc := func(sr *ldap.AddRequest) bool {
			ret := sr.DN == _name

			stdAttributes := map[string][]string{
				"name":               {_name},
				"sAMAccountName":     {_name + "$"},
				"userAccountControl": {"4096"},
			}

			found := 0
			for _, e := range sr.Attributes {
				if _, ok := stdAttributes[e.Type]; ok {
					found++
					ret = ret && reflect.DeepEqual(stdAttributes[e.Type], e.Vals)
				}

				if e.Type == "objectClass" {
					ret = ret && contains(e.Vals, "computer")
				}
			}

			ret = ret && (found == len(stdAttributes))

			return ret
		}

		mockClient := new(MockClient)
		mockClient.On("Add", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.createComputer(_name, _name, nil)
		assert.NoError(t, err)
	})
}

func TestUpdateComputerOU(t *testing.T) {
	t.Run("updateComputerOU - should forward error from ldap.Client.ModifyDN", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateComputerOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("updateComputerOU - should return nil when computer ou was updated", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("ModifyDN", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateComputerOU("", "", "")
		assert.NoError(t, err)
	})

	t.Run("updateComputerOU - should call ModifyDN with correct ModifyDNrequest", func(t *testing.T) {
		_dn := "dn=\\test,ou=computer,ou=org"
		_cn := "test"
		_ou := "ou=sub,ou=computer,ou=org"

		matchFunc := func(sr *ldap.ModifyDNRequest) bool {
			return sr.DN == _dn && sr.NewSuperior == _ou && sr.NewRDN == fmt.Sprintf("cn=%s", _cn)
		}

		mockClient := new(MockClient)
		mockClient.On("ModifyDN", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}
		err := api.updateComputerOU(_dn, _cn, _ou)
		assert.NoError(t, err)
	})
}

func TestUpdateComputerAttributes(t *testing.T) {
	t.Run("updateComputerAttributes - should forward error from ldap.client.Modify", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Modify", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateComputerAttributes("", nil, nil, nil)

		assert.Error(t, err)
	})

	t.Run("updateComputerAttributes - should return nil when object is updated successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Modify", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateComputerAttributes("", nil, nil, nil)

		assert.NoError(t, err)
	})

	t.Run("updateComputerAttributes - should not modify objectClass", func(t *testing.T) {
		_baseDN := "baseDN"

		matchFunc := func(req *ldap.ModifyRequest) bool {
			ret := req.DN == _baseDN
			for i := 0; i < len(req.Changes); i++ {
				if req.Changes[i].Modification.Type == "objectClass" {
					ret = false
				}
			}

			return ret
		}

		mockClient := new(MockClient)
		mockClient.On("Modify", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.updateComputerAttributes(_baseDN, nil, nil, nil)

		assert.NoError(t, err)
	})
}

func TestDeleteComputer(t *testing.T) {
	t.Run("deleteComputer - should forward error from api.deleteObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Del", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteComputer("")

		assert.Error(t, err)
	})

	t.Run("deleteComputer - should return nil when object is deleted successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Del", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.deleteComputer("")

		assert.NoError(t, err)
	})
}
