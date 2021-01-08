package activedirectory

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func createADComputerResult() *ldap.SearchResult {
	attributes := []string{"cn", "description"}
	return createADResult(1, attributes)
}

func TestGetComputer(t *testing.T) { // nolint:funlen // Test function
	name := getRandomString(10)
	attributes := []string{"description", "cn"}

	t.Run("getComputer - should forward errors from api.getObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}

		computer, err := api.getComputer("")

		assert.Error(t, err)
		assert.Nil(t, computer)
	})

	t.Run("getComputer - should return nil when no computer was found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, &ldap.Error{Err: fmt.Errorf("not found"), ResultCode: 32})

		api := &API{client: mockClient}

		computer, err := api.getComputer("")

		assert.NoError(t, err)
		assert.Nil(t, computer)
	})

	t.Run("getComputer - should return a computer object", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)

		api := &API{client: mockClient}

		computer, err := api.getComputer("")

		assert.NoError(t, err)
		assert.NotNil(t, computer)
		assert.IsType(t, &Computer{}, computer)
	})

	t.Run("getComputer - should include 'cn' and 'description'", func(t *testing.T) {
		matchFunc := func(sr *ldap.SearchRequest) bool {
			return contains(sr.Attributes, "cn") &&
				contains(sr.Attributes, "description")
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.MatchedBy(matchFunc)).Return(createADResult(1, attributes), nil)

		api := &API{client: mockClient}

		computer, err := api.getComputer(name)
		assert.NoError(t, err)
		assert.NotNil(t, computer)
	})

	t.Run("getComputer - should search for computer objects", func(t *testing.T) {
		matchFunc := func(sr *ldap.SearchRequest) bool {
			return strings.Contains(sr.Filter, "objectclass=computer")
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.MatchedBy(matchFunc)).Return(createADResult(1, attributes), nil)

		api := &API{client: mockClient}

		computer, err := api.getComputer(name)
		assert.NoError(t, err)
		assert.NotNil(t, computer)
	})

	t.Run("getComputer - should error when more than one object were found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(2, attributes), nil)

		api := &API{client: mockClient}

		computer, err := api.getComputer("")

		assert.Error(t, err)
		assert.Nil(t, computer)
	})

	t.Run("getComputer - should return computer object filled with search result data", func(t *testing.T) {
		sr := createADComputerResult()
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)

		api := &API{client: mockClient}

		computer, err := api.getComputer(sr.Entries[0].GetAttributeValue("cn"))

		assert.NoError(t, err)
		assert.NotNil(t, computer)

		for _, elem := range sr.Entries[0].Attributes {
			if elem.Name == "description" {
				assert.Equal(t, elem.Values[0], computer.description)
			} else if elem.Name == "cn" {
				assert.Equal(t, elem.Values[0], computer.name)
			}
		}

		assert.Equal(t, sr.Entries[0].DN, computer.dn)
	})
}

func TestCreateComputer(t *testing.T) {
	name := getRandomString(10)
	baseOU := getRandomOU(2, 2)
	description := getRandomString(10)

	t.Run("createComputer - should set standard attributes for computer objects", func(t *testing.T) {
		matchFunc := func(sr *ldap.AddRequest) bool {
			ret := sr.DN == fmt.Sprintf("cn=%s,%s", name, baseOU)

			stdAttributes := map[string][]string{
				"name":               {name},
				"sAMAccountName":     {name + "$"},
				"userAccountControl": {"4096"},
				"description":        {description},
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
		mockClient.On("Search", mock.Anything).Return(nil, nil)
		mockClient.On("Add", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.createComputer(name, baseOU, description)
		assert.NoError(t, err)
	})

	t.Run("createComputer - should forward error from api.getComputer", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}

		err := api.createComputer("", "", "")

		assert.Error(t, err)
	})

	t.Run("createComputer - should error when a computer with the same name in another OU exists", func(t *testing.T) {
		sr := createADComputerResult()

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)

		api := &API{client: mockClient}

		err := api.createComputer(sr.Entries[0].GetAttributeValue("cn"), baseOU, "")

		assert.Error(t, err)
	})

	t.Run("createComputer - should update description when the exact object is found", func(t *testing.T) {
		sr := createADComputerResult()
		sr.Entries[0].DN = fmt.Sprintf("cn=%s,%s", sr.Entries[0].GetAttributeValue("cn"), baseOU)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)
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

		err := api.createComputer(sr.Entries[0].GetAttributeValue("cn"), baseOU, "")

		assert.NoError(t, err)
	})
}

func TestUpdateComputerOU(t *testing.T) {
	ou := getRandomOU(3, 2)
	newOU := fmt.Sprintf("ou=%s,%s", getRandomString(5), ou)

	t.Run("updateComputerOU - should forward error from api.getComputer", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateComputerOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("updateComputerOU - should error when computer object not exists", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}
		err := api.updateComputerOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("updateComputerOU - should forward error from ldap.Client.ModifyDN", func(t *testing.T) {
		sr := createADComputerResult()

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateComputerOU("", "", "")
		assert.Error(t, err)
	})

	t.Run("updateComputerOU - should do nothing when computer object is already in the correct ou", func(t *testing.T) {
		sr := createADComputerResult()
		cn := sr.Entries[0].GetAttributeValue("cn")
		sr.Entries[0].DN = fmt.Sprintf("CN=%s,%s", cn, newOU)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateComputerOU(cn, ou, newOU)
		assert.NoError(t, err)
	})

	t.Run("updateComputerOU - should return nil when computer ou was updated", func(t *testing.T) {
		sr := createADComputerResult()
		cn := sr.Entries[0].GetAttributeValue("cn")

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)
		mockClient.On("ModifyDN", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateComputerOU(cn, ou, newOU)
		assert.NoError(t, err)
	})

	t.Run("updateComputerOU - should call ModifyDN with correct ModifyDNrequest", func(t *testing.T) {
		sr := createADComputerResult()
		cn := sr.Entries[0].GetAttributeValue("cn")

		matchFunc := func(sr *ldap.ModifyDNRequest) bool {
			return sr.DN == fmt.Sprintf("cn=%s,%s", cn, ou) &&
				sr.NewSuperior == newOU && sr.NewRDN == fmt.Sprintf("cn=%s", cn)
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)
		mockClient.On("ModifyDN", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}
		err := api.updateComputerOU(cn, ou, newOU)
		assert.NoError(t, err)
	})
}

func TestUpdateComputerDescription(t *testing.T) {
	t.Run("updateComputerDescription - should forward error from ldap.client.Modify", func(t *testing.T) {
		sr := createADComputerResult()
		cn := sr.Entries[0].GetAttributeValue("cn")

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)
		mockClient.On("Modify", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateComputerDescription(cn, "", "")

		assert.Error(t, err)
	})

	t.Run("updateComputerDescription - should return nil when object is updated successfully", func(t *testing.T) {
		sr := createADComputerResult()
		cn := sr.Entries[0].GetAttributeValue("cn")

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)
		mockClient.On("Modify", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateComputerDescription(cn, "", "")

		assert.NoError(t, err)
	})

	t.Run("updateComputerDescription - should modify description", func(t *testing.T) {
		sr := createADComputerResult()

		matchFunc := func(req *ldap.ModifyRequest) bool {
			for i := 0; i < len(req.Changes); i++ {
				if req.Changes[i].Modification.Type == "description" {
					return true
				}
			}

			return false
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(sr, nil)
		mockClient.On("Modify", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.updateComputerDescription("", "", "")

		assert.NoError(t, err)
	})
}

func TestDeleteComputer(t *testing.T) {
	t.Run("deleteComputer - should forward error from api.deleteObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteComputer("", "")

		assert.Error(t, err)
	})

	t.Run("deleteComputer - should return nil when object is deleted successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)
		mockClient.On("Del", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.deleteComputer("", "")

		assert.NoError(t, err)
	})
}
