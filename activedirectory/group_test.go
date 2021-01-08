package activedirectory

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		group, err := api.getGroup("", "", "", []string{}, true)

		assert.NoError(t, err)
		assert.NotNil(t, group)
		assert.IsType(t, &Group{}, group)
	})

	//
	t.Run("getGroup - should error when more than one object is found", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(2, attributes), nil)

		api := &API{client: mockClient}

		group, err := api.getGroup("", "", "", []string{}, true)

		assert.Error(t, err)
		assert.Nil(t, group)
	})

	//
	t.Run("getGroup - should return nil when api.client.Search returns nil", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, nil)

		api := &API{client: mockClient}

		group, err := api.getGroup("", "", "", []string{}, true)

		assert.NoError(t, err)
		assert.Nil(t, group)
	})

	t.Run("getGroup - should return nil when api.client.Search returns 0 objects", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(0, attributes), nil)

		api := &API{client: mockClient}

		group, err := api.getGroup("", "", "", []string{}, true)

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

		err := api.createGroup("", "", "", "", []string{}, true)
		assert.Error(t, err)
	})

	t.Run("createGroup- should error when ou already exists in another place", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(createADResult(1, attributes), nil)

		api := &API{client: mockClient}

		err := api.createGroup("", "", "", "", []string{}, true)
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
	attributes := []string{"name", "description", "sAMAccountName"}

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

	t.Run("updateGroupMembers - should forward error from api.searchObject", func(t *testing.T) {
		mockClient := new(MockClient)
		mockClient.On("Search", mock.Anything).Return(nil, fmt.Errorf("error"))

		api := &API{client: mockClient}
		err := api.updateGroupMembers("test1", "", "", []string{}, []string{}, false)

		assert.Error(t, err)
	})

	groupName := getRandomString(5)
	groupBase := getRandomOU(2, 1)
	groupDN := fmt.Sprintf("cn=%s,%s", groupName, groupBase)
	userName := getRandomString(5)
	userBase := getRandomOU(2, 1)
	userDN := fmt.Sprintf("cn=%s,%s", userName, userBase)
	userADResult := createADResultForUsers([]string{userName}, userBase)
	members := make([][]string, 1)
	groupADResult := createADResultForGroups([]string{groupName}, groupBase, members, userBase)

	t.Run("updateGroupMembers - should call Modify with add one member", func(t *testing.T) {
		mockClient := new(MockClient)

		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))", groupName)
		})).Return(groupADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=group)(objectclass=user))(memberOf=%s))", groupDN)
		})).Return(nil, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=user)(objectclass=group))(|(sAMAccountName=%s)))", userName)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == "(objectclass=*)" && sr.BaseDN == groupDN
		})).Return(groupADResult, nil)
		mockClient.On("Modify", mock.MatchedBy(func(sr *ldap.ModifyRequest) bool {
			return len(sr.Changes) == 1 &&
				sr.Changes[0].Operation == ldap.AddAttribute &&
				sr.Changes[0].Modification.Type == "member" &&
				len(sr.Changes[0].Modification.Vals) == 1 &&
				sr.Changes[0].Modification.Vals[0] == userDN
		})).Return(nil, nil)
		api := &API{client: mockClient}
		err := api.updateGroupMembers(groupName, groupBase, userBase, []string{}, []string{userName}, false)
		assert.NoError(t, err)
	})
	userName2 := getRandomString(5)
	userDN2 := fmt.Sprintf("cn=%s,%s", userName2, userBase)
	userADResult = createADResultForUsers([]string{userName, userName2}, userBase)
	t.Run("updateGroupMembers - should call Modify with add two member", func(t *testing.T) {
		mockClient := new(MockClient)

		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))", groupName)
		})).Return(groupADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=group)(objectclass=user))(memberOf=%s))", groupDN)
		})).Return(nil, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return strings.Contains(sr.Filter, "(&(|(objectclass=user)(objectclass=group))(|") &&
				strings.Contains(sr.Filter, fmt.Sprintf("(sAMAccountName=%s)", userName)) &&
				strings.Contains(sr.Filter, fmt.Sprintf("(sAMAccountName=%s)", userName2))
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == "(objectclass=*)" && sr.BaseDN == groupDN
		})).Return(groupADResult, nil)
		mockClient.On("Modify", mock.MatchedBy(func(sr *ldap.ModifyRequest) bool {
			assert.Len(t, sr.Changes, 1)
			assert.Equal(t, sr.Changes[0].Operation, uint(ldap.AddAttribute))
			assert.Equal(t, sr.Changes[0].Modification.Type, "member")
			assert.Len(t, sr.Changes[0].Modification.Vals, 2)
			assert.Contains(t, sr.Changes[0].Modification.Vals, userDN)
			assert.Contains(t, sr.Changes[0].Modification.Vals, userDN2)
			return true
		})).Return(nil, nil)
		api := &API{client: mockClient}

		err := api.updateGroupMembers(groupName, groupBase, userBase, []string{}, []string{userName, userName2}, false)
		assert.NoError(t, err)
	})

	members = make([][]string, 1)
	members[0] = []string{userName}
	groupADResult = createADResultForGroups([]string{groupName}, groupBase, members, userBase)
	userADResult = createADResultForUsers([]string{userName}, userBase)

	t.Run("updateGroupMembers - should not call Modify when user already in group", func(t *testing.T) {
		mockClient := new(MockClient)

		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))", groupName)
		})).Return(groupADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=group)(objectclass=user))(memberOf=%s))", groupDN)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=user)(objectclass=group))(|(sAMAccountName=%s)))", userName)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == "(objectclass=*)" && sr.BaseDN == groupDN
		})).Return(groupADResult, nil)
		api := &API{client: mockClient}
		err := api.updateGroupMembers(groupName, groupBase, userBase, []string{}, []string{userName}, false)
		assert.NoError(t, err)
	})
	t.Run("updateGroupMembers - should call Modify with Delete member", func(t *testing.T) {
		mockClient := new(MockClient)

		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))", groupName)
		})).Return(groupADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=group)(objectclass=user))(memberOf=%s))", groupDN)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=user)(objectclass=group))(|(sAMAccountName=%s)))", userName)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == "(objectclass=*)" && sr.BaseDN == groupDN
		})).Return(groupADResult, nil)
		mockClient.On("Modify", mock.MatchedBy(func(sr *ldap.ModifyRequest) bool {
			return len(sr.Changes) == 1 &&
				sr.Changes[0].Operation == ldap.DeleteAttribute &&
				sr.Changes[0].Modification.Type == "member" &&
				len(sr.Changes[0].Modification.Vals) == 1 &&
				sr.Changes[0].Modification.Vals[0] == userDN
		})).Return(nil, nil)
		api := &API{client: mockClient}
		err := api.updateGroupMembers(groupName, groupBase, userBase, []string{userName}, []string{}, false)
		assert.NoError(t, err)
	})

	userADResult2 := createADResultForUsers([]string{userName2}, userBase)
	t.Run("updateGroupMembers - should call Modify with Delete member and Add member", func(t *testing.T) {
		mockClient := new(MockClient)

		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))", groupName)
		})).Return(groupADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=group)(objectclass=user))(memberOf=%s))", groupDN)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=user)(objectclass=group))(|(sAMAccountName=%s)))", userName)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=user)(objectclass=group))(|(sAMAccountName=%s)))", userName2)
		})).Return(userADResult2, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == "(objectclass=*)" && sr.BaseDN == groupDN
		})).Return(groupADResult, nil)
		mockClient.On("Modify", mock.MatchedBy(func(sr *ldap.ModifyRequest) bool {
			assert.Len(t, sr.Changes, 2)
			assert.Equal(t, sr.Changes[0].Operation, uint(ldap.AddAttribute))
			assert.Equal(t, sr.Changes[1].Operation, uint(ldap.DeleteAttribute))
			assert.Equal(t, sr.Changes[0].Modification.Type, "member")
			assert.Equal(t, sr.Changes[1].Modification.Type, "member")
			assert.Equal(t, sr.Changes[0].Modification.Vals[0], userDN2)
			assert.Equal(t, sr.Changes[1].Modification.Vals[0], userDN)
			return true

		})).Return(nil, nil)
		api := &API{client: mockClient}
		err := api.updateGroupMembers(groupName, groupBase, userBase, []string{userName}, []string{userName2}, false)
		assert.NoError(t, err)
	})

	t.Run("updateGroupMembers - should not call Modify because ignoreMembersUnknownByTerraform flag was set", func(t *testing.T) {
		mockClient := new(MockClient)

		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))", groupName)
		})).Return(groupADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=group)(objectclass=user))(memberOf=%s))", groupDN)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=user)(objectclass=group))(|(sAMAccountName=%s)))", userName)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == "(objectclass=*)" && sr.BaseDN == groupDN
		})).Return(groupADResult, nil)
		api := &API{client: mockClient}
		err := api.updateGroupMembers(groupName, groupBase, userBase, []string{}, []string{}, true)
		assert.NoError(t, err)
	})
	t.Run("updateGroupMembers - should  call Modify with Delete member because ignoreMembersUnknownByTerraform flag was not set", func(t *testing.T) {
		mockClient := new(MockClient)

		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))", groupName)
		})).Return(groupADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=group)(objectclass=user))(memberOf=%s))", groupDN)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == fmt.Sprintf("(&(|(objectclass=user)(objectclass=group))(|(sAMAccountName=%s)))", userName)
		})).Return(userADResult, nil)
		mockClient.On("Search", mock.MatchedBy(func(sr *ldap.SearchRequest) bool {
			return sr.Filter == "(objectclass=*)" && sr.BaseDN == groupDN
		})).Return(groupADResult, nil)
		mockClient.On("Modify", mock.MatchedBy(func(sr *ldap.ModifyRequest) bool {
			return len(sr.Changes) == 1 &&
				sr.Changes[0].Operation == ldap.DeleteAttribute &&
				sr.Changes[0].Modification.Type == "member" &&
				len(sr.Changes[0].Modification.Vals) == 1 &&
				sr.Changes[0].Modification.Vals[0] == userDN
		})).Return(nil, nil)
		api := &API{client: mockClient}
		err := api.updateGroupMembers(groupName, groupBase, userBase, []string{}, []string{}, false)
		assert.NoError(t, err)
	})
}

func TestDeleteGroup(t *testing.T) {
	numberOfObjects := 1
	attributes := []string{"name", "sAMAccountName", "description"}
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
