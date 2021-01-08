package activedirectory

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Group is the base implementation of ad Group object
type Group struct {
	name        string
	dn          string
	description string
	member      []string
}

func (api *API) getGroup(name, baseOU, userBase string, member []string, ignoreMembersUnknownByTerraform bool) (*Group, error) {
	log.Infof("Getting group %s in %s", name, baseOU)
	attributes := []string{"name", "sAMAccaountName", "description"}

	// filter
	filter := fmt.Sprintf("(&(objectclass=group)(sAMAccountName=%s))", name)

	// trying to get ou object
	ret, err := api.searchObject(filter, baseOU, attributes)

	if err != nil {
		return nil, fmt.Errorf("getGroup - failed to search %s in %s: %s", name, baseOU, err)
	}

	if len(ret) == 0 {
		return nil, nil
	}

	if len(ret) > 1 {
		return nil, fmt.Errorf("getGroup - more than one ou object with the same name under the same base ou found")
	}
	if userBase == "" {
		userBase = api.getDomainDN()
	}
	membersFromLdap, err := api.getGroupMemberNames(ret[0].dn, userBase)
	if err != nil {
		return nil, fmt.Errorf("getGroup - failed to get group members %s in %s: %s", ret[0].dn, userBase, err)
	}
	membersMangedByTerraform := api.getMembersManagedByTerraform(membersFromLdap, member, ignoreMembersUnknownByTerraform)
	description := getAttribute("description", ret[0])
	return &Group{
		name:        ret[0].attributes["name"][0],
		dn:          ret[0].dn,
		description: description,
		member:      membersMangedByTerraform,
	}, nil
}

func getAttribute(attrName string, ldapObject *Object) string {
	for key := range ldapObject.attributes {
		if key == attrName {
			return ldapObject.attributes[key][0]
		}
	}
	return ""
}

func (api *API) getMembersManagedByTerraform(membersFromLdap []string, membersFromTerraform []string, ignoreMembersUnknownByTerraform bool) []string {
	membersMangedByTerraform := append([]string(nil), membersFromTerraform...)
	if ignoreMembersUnknownByTerraform {
		for _, m := range membersFromLdap {
			if stringInSlice(m, membersFromTerraform) {
				membersMangedByTerraform = append(membersMangedByTerraform, m)
			}
		}
	} else {
		for _, m := range membersFromLdap {
			if !stringInSlice(m, membersMangedByTerraform) {
				membersMangedByTerraform = append(membersMangedByTerraform, m)
			}
		}
	}
	return membersMangedByTerraform

}

func (api *API) getGroupMemberNames(groupDN, userBase string) ([]string, error) {
	log.Infof("Getting group members %s", groupDN)
	attributes := []string{"sAMAccountName"}
	filter := fmt.Sprintf("(&(|(objectclass=group)(objectclass=user))(memberOf=%s))", groupDN)
	ret, err := api.searchObject(filter, userBase, attributes)

	if err != nil {
		return nil, fmt.Errorf("getGroupMember - failed to search %s: %s", groupDN, err)
	}
	if ret == nil {
		return []string{}, nil
	}
	if len(ret) == 0 {
		return []string{}, nil
	}
	members := make([]string, len(ret))

	for i, member_row := range ret {
		log.Infof("Found group member: %s", member_row)
		members[i] = member_row.attributes["sAMAccountName"][0]
	}
	log.Infof("Group member names: %s", members)
	return members, nil
}

func (api *API) createGroup(name, baseOU, description, userBase string, member []string, ignoreMembersUnknownByTerraform bool) error {
	log.Infof("Creating group %s in %s", name, baseOU)
	log.Infof("Creating group with members %s", member)
	if userBase == "" {
		userBase = api.getDomainDN()
	}
	tmp, err := api.getGroup(name, baseOU, userBase, member, ignoreMembersUnknownByTerraform)
	if err != nil {
		return fmt.Errorf("createGroup - talking to active directory failed: %s", err)
	}

	// there is already an group object with the same name
	if tmp != nil {
		if tmp.name == name && tmp.dn == fmt.Sprintf("cn=%s,%s", name, baseOU) {
			log.Infof("Group object %s already exists, updating description", name)
			return api.updateGroupDescription(name, baseOU, description)
		}

		return fmt.Errorf("createGroup - group object %s already exists under this base group %s", name, baseOU)
	}
	memberDN, err := api.getGroupMemberDNByName(member, userBase)
	if err != nil {
		return fmt.Errorf("createGroup - getting group member full dn failed: %s", err)
	}

	for i, m := range memberDN {
		log.Infof("Full dn members %d %s", i, m)
	}
	log.Infof("Members list %s", memberDN)
	log.Infof("Members list count %d", len(memberDN))
	attributes := make(map[string][]string)
	attributes["cn"] = []string{name}
	attributes["name"] = []string{name}
	attributes["sAMAccountName"] = []string{name}
	attributes["groupType"] = []string{"-2147483646"}
	attributes["description"] = []string{description}
	if len(memberDN) > 0 {
		attributes["member"] = memberDN
	}

	return api.createObject(fmt.Sprintf("cn=%s,%s", name, baseOU), []string{"group", "top"}, attributes)
}

func (api *API) getGroupMemberDNByName(names []string, userBase string) ([]string, error) {

	log.Infof("Searching group member %s in %s", names, userBase)
	filter := fmt.Sprintf("(&(|(objectclass=user)(objectclass=group))(|")
	for _, m := range names {
		filter = fmt.Sprintf("%s(sAMAccountName=%s)", filter, m)
	}
	filter = fmt.Sprintf("%s))", filter)
	log.Infof("Filter for search group members: %s", filter)
	ret, err := api.searchObject(filter, userBase, []string{"sAMAccountName"})
	if len(ret) != len(names) {
		log.Errorf("Not all members found in ldap: %s", names)
		memberNamesFromLdap := make([]string, len(ret))
		for i, m := range ret {
			memberNamesFromLdap[i] = m.attributes["sAMAccountName"][0]
		}
		memberNotFound := make([]string, 0)
		for _, m := range names {
			if !stringInSlice(m, memberNamesFromLdap) {
				memberNotFound = append(memberNotFound, m)
			}
		}
		return nil, fmt.Errorf("searchMemberDNByName - not found members with sAMAccountName=%s", memberNotFound)

	}
	if err != nil {
		return nil, fmt.Errorf("searchMemberDNByName - failed to search filter: %s in %s: %s", filter, userBase, err)
	}

	members_dn := make([]string, len(names))
	for i, m := range ret {
		members_dn[i] = m.dn
	}

	return members_dn, nil
}

func (api *API) updateGroupMembers(cn, baseOU, userBase string, oldMembers, newMembers []string, ignoreMembersUnknownByTerraform bool) error {
	if userBase == "" {
		userBase = api.getDomainDN()
	}
	group, err := api.getGroup(cn, baseOU, userBase, oldMembers, ignoreMembersUnknownByTerraform)
	if err != nil {
		return fmt.Errorf("updateGroupMembers - getting group  cn=%s%s: %s", cn, baseOU, err)
	}
	membersToRemove := make(map[string]bool)
	membersToAdd := make(map[string]bool)
	// users to remove
	for _, oldMember := range oldMembers {
		if stringInSlice(oldMember, group.member) {
			if !stringInSlice(oldMember, newMembers) {
				membersToRemove[oldMember] = true
			}
		}
	}
	log.Infof("members to remove: %v", membersToRemove)
	// users to add
	for _, newMember := range newMembers {
		if !stringInSlice(newMember, group.member) {
			if !stringInSlice(newMember, oldMembers) {
				membersToAdd[newMember] = true
			}
		}
	}
	log.Infof("members to add: %v", membersToAdd)
	membersToRemoveBecauseOutsideModifcation := make(map[string]bool)
	//membersToAddBecauseOutsideModifcation := make(map[string]bool)
	// users which was add to group outsite terrafrom, remove it!!
	for _, currentMember := range group.member {
		if !keyInMap(currentMember, membersToRemove) {
			if !stringInSlice(currentMember, newMembers) {
				membersToRemoveBecauseOutsideModifcation[currentMember] = true
			}
		}
	}
	log.Infof("members to remove because added outside terraform: %v", membersToRemoveBecauseOutsideModifcation)

	if !ignoreMembersUnknownByTerraform {
		for m := range membersToRemoveBecauseOutsideModifcation {
			membersToRemove[m] = true
		}
	}

	add := make(map[string][]string)
	remove := make(map[string][]string)
	if len(membersToAdd) > 0 {
		add["member"], err = api.getGroupMemberDNByName(mapKeys(membersToAdd), userBase)
		if err != nil {
			return fmt.Errorf("updateGroupMembers - convert names to full dn for add members for group  cn=%s,%s add(%s), %s", cn, baseOU, add, err)
		}
	}
	if len(membersToRemove) > 0 {
		remove["member"], err = api.getGroupMemberDNByName(mapKeys(membersToRemove), userBase)
		if err != nil {
			return fmt.Errorf("updateGroupMembers - convert names to full dn for remove members for group  cn=%s,%s remove(%s), %s", cn, baseOU, remove, err)
		}
	}
	log.Infof("Final members to add: %s", add["member"])
	log.Infof("Final members to remove: %s", remove["member"])
	if len(add) > 0 || len(remove) > 0 {
		err = api.updateObject(group.dn, nil, add, make(map[string][]string), remove)
		if err != nil {
			return fmt.Errorf("updateGroupMembers - updating members in group  cn=%s,%s add(%s), remove(%s)    %s", cn, baseOU, add, remove, err)
		}
	} else {
		log.Infof("Members group(name=%s) not change.", group.name)
	}

	return nil

}

// updates the description of an existing ou object
func (api *API) updateGroupDescription(cn, baseOU, description string) error {
	log.Infof("Updating description of group %s under %s", cn, baseOU)
	return api.updateObject(fmt.Sprintf("cn=%s,%s", cn, baseOU), nil, nil, map[string][]string{
		"description": {description},
	}, nil)
}
func (api *API) renameGroup(oldName, baseOu, newName string) error {
	log.Infof("Renaming group %s to %s.", oldName, newName)
	// specific uid of the group
	UID := fmt.Sprintf("cn=%s", newName)
	object, err := api.getObject(fmt.Sprintf("cn=%s,%s", oldName, baseOu), []string{})
	if err != nil {
		return fmt.Errorf("renameGroup - failed to move group: %s", err)
	}
	if object == nil {
		return fmt.Errorf("renameGroup - group not found: %s", err)
	}
	req := ldap.NewModifyDNRequest(fmt.Sprintf("cn=%s,%s", oldName, baseOu), UID, true, "")
	if err = api.client.ModifyDN(req); err != nil {
		return fmt.Errorf("renameGroup - failed to rename group: %s", err)
	}
	changeAttr := map[string][]string{}
	changeAttr["sAMAccountName"] = []string{newName}
	err = api.updateObject(fmt.Sprintf("cn=%s,%s", newName, baseOu), nil, nil, changeAttr, nil)
	if err != nil {
		return fmt.Errorf("renameGroup - failed to rename sAMAccountName: %s", err)
	}

	log.Infof("Group renamed.")
	return nil
}

func (api *API) moveGroup(name, oldOU, newOU string) error {
	log.Infof("Moving group from %s to %s.", oldOU, newOU)
	// specific uid of the group
	UID := fmt.Sprintf("cn=%s", name)
	oldDN := fmt.Sprintf("cn=%s,%s", name, oldOU)
	newDN := fmt.Sprintf("cn=%s,%s", name, newOU)
	object, err := api.getObject(oldDN, []string{})
	if err != nil {
		return fmt.Errorf("moveGroup - failed to move group: %s", err)
	}
	if object == nil {
		return fmt.Errorf("moveGroup - group not found: %s", err)
	}
	if object.dn == newDN {
		log.Infof("Group already under right organization unit: %s", newOU)
		return nil
	}
	req := ldap.NewModifyDNRequest(oldDN, UID, true, newOU)
	if err := api.client.ModifyDN(req); err != nil {
		return fmt.Errorf("moveGroup - failed to move group: %s", err)
	}

	log.Infof("Group moved.")
	return nil
}

func (api *API) deleteGroup(dn string) error {
	log.Infof("Deleting group %s.", dn)

	objects, err := api.searchObject("(objectclass=*)", dn, nil)
	if err != nil {
		return fmt.Errorf("deleteGroup - failed remove group %s: %s", dn, err)
	}

	if len(objects) > 0 {
		if len(objects) > 1 || !strings.EqualFold(objects[0].dn, dn) {
			return fmt.Errorf("deleteGroup - failed to delete group %s because it has child items: %s", dn, objects[0].dn)
		}
	}

	return api.deleteObject(dn)
}

func mapKeys(mapp map[string]bool) []string {
	keys := make([]string, len(mapp))
	i := 0
	for m := range mapp {
		keys[i] = m
		i++
	}
	return keys

}

func keyInMap(a string, mapp map[string]bool) bool {
	for key := range mapp {
		if a == key {
			return true
		}
	}
	return false
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
