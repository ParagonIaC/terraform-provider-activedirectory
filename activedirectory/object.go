package activedirectory

import (
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	log "github.com/sirupsen/logrus"
)

// Object is the base implementation of ad object
type Object struct {
	dn         string
	attributes map[string][]string
}

// Search returns all ad objects which match the filter
func (api *API) searchObject(filter, baseDN string, attributes []string) ([]*Object, error) {
	log.Infof("Searching for objects in %s with filter %s", baseDN, filter)

	if len(attributes) == 0 {
		attributes = []string{"*"}
	}

	request := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attributes,
		nil,
	)

	result, err := api.client.Search(request)
	if err != nil {
		if err, ok := err.(*ldap.Error); ok {
			if err.ResultCode == 32 {
				log.Infof("No object found with filter %s", filter)
				return nil, nil
			}
		}

		return nil, fmt.Errorf("searchObject - failed to search for object (%s): %s, %s, %s, %s",
			filter, err, request.BaseDN, request.Filter, request.Attributes)
	}

	// nothing returned
	if result == nil {
		return nil, nil
	}

	objects := make([]*Object, len(result.Entries))
	for i, entry := range result.Entries {
		objects[i] = &Object{
			dn:         entry.DN,
			attributes: decodeADAttributes(entry.Attributes),
		}
	}

	return objects, nil
}

// Get returns ad object with distinguished name dn
func (api *API) getObject(dn string, attributes []string) (*Object, error) {
	log.Infof("Trying to get object %s", dn)

	objects, err := api.searchObject("(objectclass=*)", dn, attributes)
	if err != nil {
		return nil, fmt.Errorf("getObject - failed to get object %s: %s", dn, err)
	}

	if len(objects) == 0 {
		return nil, nil
	}

	if len(objects) > 1 {
		return nil, fmt.Errorf("getObject - more than one object with the same dn found")
	}

	return objects[0], nil
}

// Create create a ad object
func (api *API) createObject(dn string, classes []string, attributes map[string][]string) error {
	log.Infof("Creating object %s (class: %s)", dn, strings.Join(classes, ","))

	tmp, err := api.getObject(dn, nil)
	if err != nil {
		return fmt.Errorf("createObject - talking to active directory failed: %s", err)
	}

	// there is already an object with the same dn
	if tmp != nil {
		return fmt.Errorf("createObject - object %s already exists", dn)
	}

	// create ad add request
	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", classes)

	for key, value := range attributes {
		req.Attribute(key, value)
	}

	// add to ad
	if err := api.client.Add(req); err != nil {
		return fmt.Errorf("createObject - failed to create object %s: %s", dn, err)
	}

	log.Info("Object created")
	return nil
}

// Delete deletes a ad object
func (api *API) deleteObject(dn string) error {
	log.Infof("Removing object %s", dn)

	tmp, err := api.getObject(dn, nil)
	if err != nil {
		return fmt.Errorf("deleteComputer - talking to active directory failed: %s", err)
	}

	if tmp == nil {
		log.Info("Object is already deleted")
		return nil
	}

	// create ad delete request
	req := ldap.NewDelRequest(dn, nil)

	// delete object from ad
	if err := api.client.Del(req); err != nil {
		return fmt.Errorf("deleteObject - failed to delete object %s: %s", dn, err)
	}

	log.Info("Object removed")
	return nil
}

// Update updates a ad object
func (api *API) updateObject(dn string, classes []string, added, changed, removed map[string][]string) error {
	log.Infof("Updating object %s", dn)

	tmp, err := api.getObject(dn, nil)
	if err != nil {
		return fmt.Errorf("updateObject - talking to active directory failed: %s", err)
	}

	if tmp == nil {
		return fmt.Errorf("updateObject - object %s does not exist", dn)
	}

	req := ldap.NewModifyRequest(dn, nil)

	if classes != nil {
		req.Replace("objectClass", classes)
	}

	for key, value := range added {
		req.Add(key, value)
	}

	for key, value := range changed {
		req.Replace(key, value)
	}

	for key, value := range removed {
		req.Delete(key, value)
	}

	if err := api.client.Modify(req); err != nil {
		return fmt.Errorf("updateObject - failed to update %s: %s", dn, err)
	}

	log.Info("Object updated")
	return nil
}
