package activedirectory

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/ldap.v3"
)

func createADResult(cntEntries, cntAttributes int) *ldap.SearchResult {
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
	searchResult := createADResult(numberOfObjects, numberOfAttributes)

	t.Run("searchObject - should forward errors from ldap.client.search", func(t *testing.T) {
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

	t.Run("searchObject - should forward the input values to ldap.Client.Search", func(t *testing.T) {
		_filter := "filter"
		_baseDN := "baseDN"
		_attributes := []string{"Attribute1", "Atribute2"}

		matchFunc := func(sr *ldap.SearchRequest) bool {
			return sr.BaseDN == _baseDN && sr.Filter == _filter && reflect.DeepEqual(sr.Attributes, _attributes)
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.MatchedBy(matchFunc)).Return(searchResult, nil)

		api := &API{client: mockClient}

		objects, err := api.searchObject(_filter, _baseDN, _attributes)

		assert.NoError(t, err)
		assert.NotNil(t, objects)
	})

	t.Run("searchObject - if not attribute specified search with '*'", func(t *testing.T) {
		matchFunc := func(sr *ldap.SearchRequest) bool {
			return reflect.DeepEqual(sr.Attributes, []string{"*"})
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.MatchedBy(matchFunc)).Return(searchResult, nil)

		api := &API{client: mockClient}

		objects, err := api.searchObject("", "", nil)

		assert.NoError(t, err)
		assert.NotNil(t, objects)
	})
}

func TestGetObject(t *testing.T) {
	t.Run("getObject - should forward errors from getObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}

		objects, err := api.getObject("", nil)

		assert.Error(t, err)
		assert.Nil(t, objects)
	})

	t.Run("getObject - should return nil when object not found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, &ldap.Error{Err: fmt.Errorf("not found"), ResultCode: 32})

		api := &API{client: mockClient}

		objects, err := api.getObject("", nil)

		assert.NoError(t, err)
		assert.Nil(t, objects)
	})

	t.Run("getObject - should return nil when nothing was found", func(t *testing.T) {
		numberOfObjects := 0
		numberOfAttributes := 3
		searchResult := createADResult(numberOfObjects, numberOfAttributes)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}

		objects, err := api.getObject("", nil)

		assert.NoError(t, err)
		assert.Nil(t, objects)
	})

	t.Run("getObject - should return one object", func(t *testing.T) {
		numberOfObjects := 1
		numberOfAttributes := 3
		searchResult := createADResult(numberOfObjects, numberOfAttributes)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}

		object, err := api.getObject("", nil)

		assert.NoError(t, err)
		assert.NotNil(t, object)

		// check all values
		assert.Equal(t, searchResult.Entries[0].DN, object.dn)
		for j := 0; j < numberOfAttributes; j++ {
			assert.Equal(t, searchResult.Entries[0].Attributes[j].Values, object.attributes[searchResult.Entries[0].Attributes[j].Name])
		}
	})

	t.Run("getObject - should forward the input values to API.Search", func(t *testing.T) {
		numberOfObjects := 1
		numberOfAttributes := 3
		searchResult := createADResult(numberOfObjects, numberOfAttributes)

		_baseDN := "baseDN"
		_attributes := []string{"Attribute1", "Atribute2"}

		matchFunc := func(sr *ldap.SearchRequest) bool {
			return sr.BaseDN == _baseDN && reflect.DeepEqual(sr.Attributes, _attributes)
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.MatchedBy(matchFunc)).Return(searchResult, nil)

		api := &API{client: mockClient}

		objects, err := api.getObject(_baseDN, _attributes)

		assert.NoError(t, err)
		assert.NotNil(t, objects)
	})
}

func TestCreateObject(t *testing.T) {
	t.Run("createObject - should forward error from ldap.Client.Add", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Add", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.createObject("", nil, nil)

		assert.Error(t, err)
	})

	t.Run("createObject - should return nil when object is created successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Add", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.createObject("", nil, nil)

		assert.NoError(t, err)
	})

	t.Run("createObject - should forward the input values to ldap.Client.Add", func(t *testing.T) {
		_baseDN := "baseDN"
		_classes := []string{"Class1", "Class2"}
		_attributes := map[string][]string{
			"Attribute1": {"Value1"},
			"Attribute2": {"Value1", "Value2"},
		}

		matchFunc := func(req *ldap.AddRequest) bool {
			ret := req.DN == _baseDN
			ret = ret && (len(req.Attributes) == 3)

			for i := 0; i < len(req.Attributes); i++ {
				if req.Attributes[i].Type == "objectClass" {
					ret = ret && reflect.DeepEqual(req.Attributes[i].Vals, _classes)
				} else {
					ret = ret && reflect.DeepEqual(req.Attributes[i].Vals, _attributes[req.Attributes[i].Type])
				}
			}

			return ret
		}

		mockClient := new(MockClient)
		mockClient.On("Add", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.createObject(_baseDN, _classes, _attributes)

		assert.NoError(t, err)
	})
}

func TestDeleteObject(t *testing.T) {
	t.Run("deleteObject - should forward error from ldap.client.Del", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Del", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteObject("")

		assert.Error(t, err)
	})

	t.Run("deleteObject - should return nil when object is deleted successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Del", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.deleteObject("")

		assert.NoError(t, err)
	})

	t.Run("deleteObject - should forward the input values to ldap.Client.Del", func(t *testing.T) {
		_baseDN := "baseDN"

		matchFunc := func(sr *ldap.DelRequest) bool {
			return sr.DN == _baseDN
		}

		mockClient := new(MockClient)
		mockClient.On("Del", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.deleteObject(_baseDN)

		assert.NoError(t, err)
	})
}

func TestUpdateObject(t *testing.T) {
	t.Run("updateObject - should forward error from ldap.client.Modify", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Modify", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateObject("", nil, nil, nil, nil)

		assert.Error(t, err)
	})

	t.Run("updateObject - should return nil when object is updated successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Modify", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateObject("", nil, nil, nil, nil)

		assert.NoError(t, err)
	})

	t.Run("updateObject - should forward the input values to ldap.Client.Modify", func(t *testing.T) {
		_baseDN := "baseDN"
		_classes := []string{"Class1", "Class2"}
		_added := map[string][]string{
			"Attribute1": {"Value1"},
		}
		_changed := map[string][]string{
			"Attribute2": {"Value2"},
		}
		_removed := map[string][]string{
			"Attribute3": {"Value3"},
		}

		matchFunc := func(req *ldap.ModifyRequest) bool {
			ret := req.DN == _baseDN
			for i := 0; i < len(req.Changes); i++ {
				if req.Changes[i].Modification.Type == "objectClass" {
					ret = ret && reflect.DeepEqual(req.Changes[i].Modification.Vals, _classes)
				} else {
					if req.Changes[i].Operation == ldap.AddAttribute {
						ret = ret && reflect.DeepEqual(req.Changes[i].Modification.Vals, _added[req.Changes[i].Modification.Type])
					} else if req.Changes[i].Operation == ldap.ReplaceAttribute {
						ret = ret && reflect.DeepEqual(req.Changes[i].Modification.Vals, _changed[req.Changes[i].Modification.Type])
					} else {
						ret = ret && reflect.DeepEqual(req.Changes[i].Modification.Vals, _removed[req.Changes[i].Modification.Type])
					}
				}
			}

			return ret
		}

		mockClient := new(MockClient)
		mockClient.On("Modify", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.updateObject(_baseDN, _classes, _added, _changed, _removed)

		assert.NoError(t, err)
	})
}
