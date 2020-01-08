package activedirectory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/ldap.v3"
)

//
func TestGetGroup(t *testing.T) {
	numberOfObjects := 1
	attributes := []string{"cn", "name", "member", "sAMAccountName", "description"}
	result := createADResult(numberOfObjects, attributes)

	fmt.Println(result)
	t.Run("getGroup - should forward errors from api.getObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}

		ou, err := api.getGroup("", "", "", []string{}, true)

		assert.Error(t, err)
		assert.Nil(t, ou)
	})

	t.Run("getGroup - should return nil when no ou was found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, &ldap.Error{Err: fmt.Errorf("not found"), ResultCode: 32})

		api := &API{client: mockClient}

		ou, err := api.getGroup("", "", "", []string{}, true)

		assert.NoError(t, err)
		assert.Nil(t, ou)
	})

	t.Run("getGroup - should return group object", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(result, nil)

		api := &API{client: mockClient}

		group, err := api.getGroup("","","",[]string{},true)

		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.IsType(t, &Group{}, group)
	})

	//
	t.Run("getGroup - should error when more than one object is found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(2, attributes), nil)

		api := &API{client: mockClient}

		group, err := api.getGroup("","","",[]string{},true)

		assert.Error(t, err)
		assert.Nil(t, group)
	})

//
	t.Run("getGroup - should return nil when api.client.Search returns nil", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}

		group, err := api.getGroup("","","",[]string{},true)

		assert.NoError(t, err)
		assert.Nil(t, group)
	})

	t.Run("getGroup - should return nil when api.client.Search returns 0 objects", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(0, attributes), nil)

		api := &API{client: mockClient}

		group, err := api.getGroup("","","",[]string{},true)

		assert.NoError(t, err)
		assert.Nil(t, group)
	})
}

func TestCreateGroup(t *testing.T) {
	attributes := []string{"cn", "name", "member", "sAMAccountName", "description"}

	t.Run("createGroup - should forward errors from api.client.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}

		err := api.createGroup("", "", "","",[]string{},true)
		assert.Error(t, err)
	})

	t.Run("createGroup- should error when ou already exists in another place", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)

		api := &API{client: mockClient}

		err := api.createGroup("", "", "","",[]string{},true)
		assert.Error(t, err)
	})
}

func TestMoveGroup(t *testing.T) {
	attributes := []string{"name", "description", "member"}

	t.Run("moveGroup - should forward error from ldap.Client.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.moveGroup("", "", "")
		assert.Error(t, err)
	})

	t.Run("moveGroup - should error when group was not found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}
		err := api.moveGroup("", "", "")
		assert.Error(t, err)
	})

	t.Run("moveGroup - should forward error from ldap.Client.ModifyDN", func(t *testing.T) {
		res := createADResult(1, attributes)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(res, nil)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.moveGroup("", "", "")
		assert.Error(t, err)
	})

	t.Run("moveGroup - should return nil when ou was updated", func(t *testing.T) {
		res := createADResult(1, attributes)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(res, nil)
		mockClient.On("ModifyDN", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.moveGroup("", "", "")
		assert.NoError(t, err)
	})

	t.Run("moveGroup - should call ModifyDN with correct ModifyDNrequest", func(t *testing.T) {
		res := createADResult(1, attributes)

		cn := getRandomString(10)
		oldOU := getRandomOU(3, 2)
		newOU := fmt.Sprintf("ou=%s,%s", getRandomString(10), oldOU)

		matchFunc := func(sr *ldap.ModifyDNRequest) bool {
			return sr.DN == fmt.Sprintf("cn=%s,%s", cn, oldOU) &&
				sr.NewSuperior == newOU && sr.NewRDN == fmt.Sprintf("cn=%s", cn)
		}

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(res, nil)
		mockClient.On("ModifyDN", mock.MatchedBy(matchFunc)).Return(nil)

		api := &API{client: mockClient}
		err := api.moveGroup(cn, oldOU, newOU)
		assert.NoError(t, err)
	})
	//
	t.Run("moveGroup - should do nothing when group is already located under the target ou", func(t *testing.T) {
		res := createADResult(1, attributes)

		cn := res.Entries[0].GetAttributeValue("name")
		oldOU := getRandomOU(2, 3)
		newOU := fmt.Sprintf("ou=%s,%s", getRandomString(10), oldOU)
		res.Entries[0].DN = fmt.Sprintf("cn=%s,%s", cn, newOU)

		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(res, nil)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.moveGroup(cn, oldOU, newOU)
		assert.NoError(t, err)
	})
}

func TestUpdateGroupDescription(t *testing.T) {
	attributes := []string{"name", "description"}

	t.Run("updateGroupDescription - should forward error from ldap.client.Modify", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)
		mockClient.On("Modify", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateGroupDescription("", "", "")

		assert.Error(t, err)
	})

	t.Run("updateGroupDescription - should return nil when ou was updated successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)
		mockClient.On("Modify", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.updateGroupDescription("", "", "")

		assert.NoError(t, err)
	})


	t.Run("updateGroupDescription - should modify description", func(t *testing.T) {
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

		err := api.updateGroupDescription("", "", "")

		assert.NoError(t, err)
	})
}
//
func TestUpdateGroupName(t *testing.T) {
	attributes := []string{"name", "description","sAMAccountName"}

	t.Run("renameGroup - should forward error from ldap.Client.Search", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.renameGroup("a", "b", "c")
		assert.Error(t, err)
	})

	t.Run("renameGroup - should error when ou was not found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}
		err := api.renameGroup("a", "b", "c")
		assert.Error(t, err)
	})

	t.Run("renameGroup - should forward error from ldap.client.ModifyDN", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)
		mockClient.On("ModifyDN", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.renameGroup("b", "c", "d")

		assert.Error(t, err)
	})


	t.Run("renameGroup - should return nil when ou was updated successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)
		mockClient.On("ModifyDN", mock.Anything).Return(nil)
		mockClient.On("Modify", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.renameGroup("c", "d", "e")

		assert.NoError(t, err)
	})
}

func TestUpdateMemberGroup(t *testing.T) {
	numberOfObjects := 1
	attributes := []string{"name", "description"}
	searchResult := createADResult(numberOfObjects, attributes)

	t.Run("updateGroupMembers - should forward error from api.searchObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateGroupMembers("test1","","",[]string{},[]string{},false)

		assert.Error(t, err)
	})

	searchResult.Entries[0].Attributes = append(searchResult.Entries[0].Attributes,&ldap.EntryAttribute{
		Name: "sAMAccountName",
		Values: []string{searchResult.Entries[0].Attributes[0].Values[0]},
	})
	name:= searchResult.Entries[0].Attributes[0].Values[0]
	dn :=searchResult.Entries[0].DN
	t.Run("updateGroupMembers - should call Modify with correct ModifyRequest ", func(t *testing.T) {
		mockClient := new(MockClient)
		matchFunc := func(sr *ldap.ModifyRequest) bool {
			if len(sr.Changes) == 1 &&
				sr.Changes[0].Operation == ldap.AddAttribute  &&
				sr.Changes[0].Modification.Type == "member" &&
				len(sr.Changes[0].Modification.Vals) == 1 &&
				sr.Changes[0].Modification.Vals[0] == "userName"  {
					return true
			}
			return false
		}

		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))",name)
		})).Return(searchResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=group)(objectclass=user))(memberOf=%s))",dn)
		})).Return(nil, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == "(&(|(objectclass=user)(objectclass=group))(|(sAMAccountName=userName)))"
		})).Return(createADResultForUsers([]string{"userName"},""), nil)

		mockClient.On("Modify",mock.MatchedBy(matchFunc)).Return(nil, nil)
		api := &API{client: mockClient}
		err := api.updateGroupMembers(name,"","",[]string{},[]string{"userName"},false)

		assert.NoError(t, err)
	})
}

func TestDeleteGroup(t *testing.T) {
	numberOfObjects := 1
	attributes := []string{"name","sAMAccountName", "description"}
	searchResult := createADResult(numberOfObjects, attributes)

	t.Run("deleteGroup - should forward error from api.searchObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))


		api := &API{client: mockClient}
		err := api.deleteGroup("test1")

		assert.Error(t, err)
	})

	t.Run("deleteGroup - should forward error from api.deleteObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(searchResult, nil)
		mockClient.On("Del", mock.Anything).Return(fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.deleteGroup("test2")

		assert.Error(t, err)
	})

	t.Run("deleteGroup - should return nil when object is deleted successfully", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)
		mockClient.On("Del", mock.Anything).Return(nil)

		api := &API{client: mockClient}
		err := api.deleteGroup("test3")

		assert.NoError(t, err)
	})
}
