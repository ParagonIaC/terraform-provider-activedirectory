package activedirectory

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// OU is the base implementation of ad organizational unit object
type OU struct {
	name        string
	dn          string
	description string
}

// returns ou object
func (api *API) getOU(name, baseOU string) (*OU, error) {
	dn := fmt.Sprintf("ou=%s,%s", name, baseOU)
	log.Infof("Trying to get ad ou: %s", dn)

	attributes := []string{"name", "ou", "description"}

	// filter
	filter := fmt.Sprintf("(&(objectclass=organizationalUnit)(name=%s))", name)

	// trying to get ou object
	ret, err := api.searchObject(filter, baseOU, attributes)
	if err != nil {
		if err, ok := err.(*ldap.Error); ok {
			if err.ResultCode == 32 {
				log.Info("AD ou object could not be found", dn)
				return nil, nil
			}
		}
		log.Errorf("Error will searching for ad ou object %s: %s:", dn, err)
		return nil, err
	}

	if len(ret) != 1 {
		return nil, nil
	}

	if ret == nil {
		return nil, nil
	}

	return &OU{
		name:        strings.Join(ret[0].attributes["name"], ""),
		dn:          ret[0].dn,
		description: strings.Join(ret[0].attributes["description"], ""),
	}, nil
}

// creates a new ou object
func (api *API) createOU(dn, name, description string) error {
	log.Infof("Creating ou object %s", dn)

	attributes := make(map[string][]string)
	attributes["name"] = []string{name}
	attributes["ou"] = []string{name}
	attributes["description"] = []string{description}

	return api.createObject(dn, []string{"organizationalUnit", "top"}, attributes)
}

// moves an existing ou object to a new ou
func (api *API) moveOU(dn, cn, ou string) error {
	log.Infof("Moving ou object %s to ou %s", dn, ou)

	// specific uid of the ou
	UID := fmt.Sprintf("ou=%s", cn)

	// move ou object to new ou
	req := ldap.NewModifyDNRequest(dn, UID, true, ou)
	if err := api.client.ModifyDN(req); err != nil {
		log.Errorf("Moving object %s to %s failed: %s", dn, ou, err)
		return err
	}

	log.Infof("Object %s moved", dn)

	return nil
}

// updates the description of an existing ou object
func (api *API) updateOUDescription(dn, description string) error {
	log.Infof("updating description of ou object %s", dn)
	return api.updateObject(dn, nil, nil, map[string][]string{
		"description": {description},
	}, nil)
}

// updates the name of an existing ou object
func (api *API) updateOUName(dn, name string) error {
	log.Infof("updating name of ou object %s", dn)

	ou := strings.ToLower(dn[(len(name) + 3):]) // remove 'ou=' and ','

	return api.moveOU(dn, name, ou)
}

// deletes an existing ou object.
func (api *API) deleteOU(dn string) error {
	objects, err := api.searchObject("(objectclass=*)", dn, nil)
	if err != nil {
		return err
	}

	if len(objects) > 0 {
		return fmt.Errorf("deleting of OU %s not possible because it has child items", dn)
	}

	log.Infof("Deleting ou object %s", dn)
	return api.deleteObject(dn)
}
