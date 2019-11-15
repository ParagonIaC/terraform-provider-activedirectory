package ldap

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/ldap.v3"
)

func createLDAPResult(cntEntries, cntAttributes int) *ldap.SearchResult {
	result := new(ldap.SearchResult)

	result.Entries = make([]*ldap.Entry, cntEntries)
	for i := 0; i < cntEntries; i++ {
		result.Entries[i] = new(ldap.Entry)
		result.Entries[i].DN = fmt.Sprintf("DN%d", i)
		result.Entries[i].Attributes = make([]*ldap.EntryAttribute, cntAttributes)
		for j := 0; j < cntAttributes; j++ {
			result.Entries[i].Attributes[j] = &ldap.EntryAttribute{
				Name:   fmt.Sprintf("Name%d", j),
				Values: []string{fmt.Sprintf("Values%d", j)},
			}
		}
	}

	return result
}

func TestSearchObject(t *testing.T) {
	numberOfObjects := 2
	numberOfAttributes := 3
	searchResult := createLDAPResult(numberOfObjects, numberOfAttributes)

	t.Run("searchObject - should forward errors from getObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}

		objects, err := api.searchObject("", "", nil)

		assert.Error(t, err)
		assert.Nil(t, objects)
	})

	t.Run("searchObject - should return a list of objects", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}

		objects, err := api.searchObject("", "", nil)

		assert.NoError(t, err)
		assert.NotNil(t, objects)
		assert.Len(t, objects, 2)

		// check all values
		for i := 0; i < numberOfObjects; i++ {
			assert.Equal(t, searchResult.Entries[i].DN, objects[i].dn)

			for j := 0; j < numberOfAttributes; j++ {
				assert.Equal(t, searchResult.Entries[i].Attributes[j].Values, objects[i].attributes[searchResult.Entries[i].Attributes[j].Name])
			}
		}
	})
}
