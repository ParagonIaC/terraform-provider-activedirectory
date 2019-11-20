package activedirectory

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// Object is the base implementation of ad object
type Object struct {
	dn         string
	attributes map[string][]string
}

// Search returns all ad objects which match the filter
func (api *API) searchObject(filter, baseDN string, attributes []string) ([]*Object, error) {
	log.Infof("Searching for ad objects in %s: %s", baseDN, filter)

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
		log.Errorf("Error will searching with filter %s: %s:", filter, err)
		return nil, err
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
	log.Infof("Trying to get ad object: %s", dn)

	objects, err := api.searchObject("(objectclass=*)", dn, attributes)
	if err != nil {
		if err, ok := err.(*ldap.Error); ok {
			if err.ResultCode == 32 {
				log.Info("AD object could not be found", dn)
				return nil, nil
			}
		}
		log.Errorf("Error will searching for ad object %s: %s:", dn, err)
		return nil, err
	}

	if len(objects) != 1 {
		return nil, nil
	}

	return objects[0], nil
}

// Create create a ad object
func (api *API) createObject(dn string, classes []string, attributes map[string][]string) error {
	log.Infof("Creating object %s (%s)", dn, strings.Join(classes, ","))

	// create ad add request
	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", classes)

	for key, value := range attributes {
		req.Attribute(key, value)
	}

	// add to ad
	if err := api.client.Add(req); err != nil {
		log.Errorf("Creating of object %s failed: %s", dn, err)
		return err
	}

	log.Infof("Object %s created", dn)

	return nil
}

// Delete deletes a ad object
func (api *API) deleteObject(dn string) error {
	log.Infof("Removing object %s", dn)

	// create ad delete request
	req := ldap.NewDelRequest(dn, nil)

	// delete object from ad
	if err := api.client.Del(req); err != nil {
		log.Errorf("Removing of object %s failed: %s", dn, err)
		return err
	}

	log.Infof("Object %s removed", dn)

	return nil
}

// Update updates a ad object
func (api *API) updateObject(dn string, classes []string, added, changed, removed map[string][]string) error {
	log.Infof("Updating object %s", dn)

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
		log.Errorf("Updating object %s failed: %s", dn, err)
		return err
	}

	log.Infof("Object %s updated", dn)
	return nil
}
