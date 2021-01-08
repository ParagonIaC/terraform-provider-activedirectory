package activedirectory

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createADResult(cntEntries int, attributes []string) *ldap.SearchResult {
	result := new(ldap.SearchResult)

	result.Entries = make([]*ldap.Entry, cntEntries)
	for i := 0; i < cntEntries; i++ {
		result.Entries[i] = new(ldap.Entry)
		result.Entries[i].DN = getRandomString(10)
		result.Entries[i].Attributes = make([]*ldap.EntryAttribute, len(attributes))
		for idx, elem := range attributes {
			result.Entries[i].Attributes[idx] = &ldap.EntryAttribute{
				Name:   elem,
				Values: []string{getRandomString(10)},
			}
		}
	}

	return result
}
func createADResultForUsers(names []string, userBase string) *ldap.SearchResult {
	result := new(ldap.SearchResult)
	result.Entries = make([]*ldap.Entry, len(names))
	for i := 0; i < len(names); i++ {
		result.Entries[i] = createLdapUserEntry(names[i], userBase)
	}
	return result
}

func createLdapUserEntry(userName, userBase string) *ldap.Entry {
	result := new(ldap.Entry)
	result.DN = fmt.Sprintf("cn=%s,%s", userName, userBase)
	result.Attributes = make([]*ldap.EntryAttribute, 1)
	result.Attributes[0] = &ldap.EntryAttribute{
		Name:   "sAMAccountName",
		Values: []string{userName},
	}
	return result
}
func createADResultForGroups(groupNames []string, groupBase string, members [][]string, userBase string) *ldap.SearchResult {
	result := new(ldap.SearchResult)
	result.Entries = make([]*ldap.Entry, len(groupNames))
	for i := 0; i < len(groupNames); i++ {
		result.Entries[i] = createLdapGroupEntry(groupNames[i], groupBase, members[i], userBase)
	}
	return result
}
func createLdapGroupEntry(groupName, groupBaseOU string, membersNames []string, userBaseOU string) *ldap.Entry {
	result := new(ldap.Entry)
	result.DN = fmt.Sprintf("cn=%s,%s", groupName, groupBaseOU)
	result.Attributes = make([]*ldap.EntryAttribute, 4)
	membersFullDN := make([]string, len(membersNames))
	for i, memberName := range membersNames {
		membersFullDN[i] = fmt.Sprintf("cn=%s,%s", memberName, userBaseOU)
	}
	result.Attributes[0] = &ldap.EntryAttribute{
		Name:   "sAMAccountName",
		Values: []string{groupName},
	}
	result.Attributes[1] = &ldap.EntryAttribute{
		Name:   "name",
		Values: []string{groupName},
	}
	result.Attributes[2] = &ldap.EntryAttribute{
		Name:   "description",
		Values: []string{getRandomString(10)},
	}
	result.Attributes[3] = &ldap.EntryAttribute{
		Name:   "member",
		Values: []string{groupName},
	}

	return result
}

func TestSearchObject(t *testing.T) {
	numberOfObjects := 2
	attributes := []string{"cn", "desc"}
	searchResult := createADResult(numberOfObjects, attributes)

	t.Run("searchObject - should forward errors from ldap.client.Search", func(t *testing.T) {
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

			for j := 0; j < len(attributes); j++ {
				assert.Equal(t, searchResult.Entries[i].Attributes[j].Values, objects[i].attributes[searchResult.Entries[i].Attributes[j].Name])
			}
		}
	})

	t.Run("searchObject - should forward the input values to ldap.Client.Search", func(t *testing.T) {
		_filter := getRandomString(10)
		_baseDN := getRandomString(10)
		_attributes := []string{getRandomString(10), getRandomString(10)}

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

	t.Run("searchObject - should return nil when error result equal 32 (nothing found)", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, &ldap.Error{Err: fmt.Errorf("not found"), ResultCode: 32})

		api := &API{client: mockClient}

		objects, err := api.searchObject("", "", nil)

		assert.NoError(t, err)
		assert.Nil(t, objects)
	})

	t.Run("searchObject - should return nil when api.client.Search return nil", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}

		objects, err := api.searchObject("", "", nil)

		assert.NoError(t, err)
		assert.Nil(t, objects)
	})
}

func TestGetObject(t *testing.T) {
	attributes := []string{"cn", "desc"}

	t.Run("getObject - should forward errors from api.client.Search", func(t *testing.T) {
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
		searchResult := createADResult(numberOfObjects, nil)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}

		objects, err := api.getObject("", nil)

		assert.NoError(t, err)
		assert.Nil(t, objects)
	})

	t.Run("getObject - should return one object", func(t *testing.T) {
		numberOfObjects := 1
		searchResult := createADResult(numberOfObjects, attributes)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}

		object, err := api.getObject("", nil)

		assert.NoError(t, err)
		assert.NotNil(t, object)

		// check all values
		assert.Equal(t, searchResult.Entries[0].DN, object.dn)
		for j := 0; j < len(attributes); j++ {
			assert.Equal(t, searchResult.Entries[0].Attributes[j].Values, object.attributes[searchResult.Entries[0].Attributes[j].Name])
		}
	})

	t.Run("getObject - should forward the input values to api.client.Search", func(t *testing.T) {
		numberOfObjects := 1
		searchResult := createADResult(numberOfObjects, attributes)

		_baseDN := getRandomString(10)
		_attributes := []string{getRandomString(10), getRandomString(10)}

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

	t.Run("getObject - should error when more than one result", func(t *testing.T) {
		numberOfObjects := 2
		searchResult := createADResult(numberOfObjects, attributes)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}

		objects, err := api.getObject("", nil)

		assert.Error(t, err)
		assert.Nil(t, objects)
	})
}

func TestCreateObject(t *testing.T) {
	t.Run("createObject - should forward error from ldap.Client.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.createObject("", nil, nil)

		assert.Error(t, err)
	})

	t.Run("createObject - should forward error from ldap.Client.Add", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)
		mockClient.On("Add", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.createObject("", nil, nil)

		assert.Error(t, err)
	})

	t.Run("createObject - should error when object already exists", func(t *testing.T) {
		numberOfObjects := 1
		searchResult := createADResult(numberOfObjects, []string{})

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)

		api := &API{client: mockClient}
		err := api.createObject("", nil, nil)

		assert.Error(t, err)
	})

	t.Run("createObject - should return nil when object is created successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)
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
		mockClient.On("Search", mock.Anything).Return(nil, nil)
		mockClient.On("Add", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.createObject(_baseDN, _classes, _attributes)

		assert.NoError(t, err)
	})
}

func TestDeleteObject(t *testing.T) {
	numberOfObjects := 1
	searchResult := createADResult(numberOfObjects, []string{})

	t.Run("deleteObject - should forward error from ldap.client.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteObject("")

		assert.Error(t, err)
	})

	t.Run("deleteObject - should return nil when object is already deleted", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}
		err := api.deleteObject("")

		assert.NoError(t, err)
	})

	t.Run("deleteObject - should forward error from ldap.client.Del", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)
		mockClient.On("Del", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteObject("")

		assert.Error(t, err)
	})

	t.Run("deleteObject - should return nil when object is deleted successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)
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
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)
		mockClient.On("Del", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.deleteObject(_baseDN)

		assert.NoError(t, err)
	})
}

func TestUpdateObject(t *testing.T) {
	numberOfObjects := 1
	searchResult := createADResult(numberOfObjects, []string{})

	t.Run("updateObject - should forward error from ldap.client.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateObject("", nil, nil, nil, nil)

		assert.Error(t, err)
	})

	t.Run("updateObject - should error when object does not exists", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}
		err := api.updateObject("", nil, nil, nil, nil)

		assert.Error(t, err)
	})

	t.Run("updateObject - should forward error from ldap.client.Modify", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)
		mockClient.On("Modify", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateObject("", nil, nil, nil, nil)

		assert.Error(t, err)
	})

	t.Run("updateObject - should return nil when object is updated successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)
		mockClient.On("Modify", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateObject("", nil, nil, nil, nil)

		assert.NoError(t, err)
	})

	t.Run("updateObject - should forward the input values to ldap.Client.Modify", func(t *testing.T) {
		_baseDN := getRandomString(10)
		_classes := []string{getRandomString(10), getRandomString(10)}
		_added := map[string][]string{
			getRandomString(10): {getRandomString(10)},
		}
		_changed := map[string][]string{
			getRandomString(10): {getRandomString(10)},
		}
		_removed := map[string][]string{
			getRandomString(10): {getRandomString(10)},
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
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)
		mockClient.On("Modify", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}

		err := api.updateObject(_baseDN, _classes, _added, _changed, _removed)

		assert.NoError(t, err)
	})
}
