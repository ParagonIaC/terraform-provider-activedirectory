package activedirectory

import (
	"fmt"

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
	log.Infof("Getting organizational unit %s in %s", name, baseOU)

	attributes := []string{"name", "ou", "description"}

	// filter
	filter := fmt.Sprintf("(&(objectclass=organizationalUnit)(name=%s))", name)

	// trying to get ou object
	ret, err := api.searchObject(filter, baseOU, attributes)
	if err != nil {
		return nil, fmt.Errorf("getOU - failed to search %s in %s: %s", name, baseOU, err)
	}

	if ret == nil || len(ret) == 0 {
		return nil, nil
	}

	if len(ret) > 1 {
		return nil, fmt.Errorf("getOU - more than one ou object with the same name under the same base ou found")
	}

	return &OU{
		name:        ret[0].attributes["name"][0],
		dn:          ret[0].dn,
		description: ret[0].attributes["description"][0],
	}, nil
}

// creates a new ou object
func (api *API) createOU(name, baseOU, description string) error {
	log.Infof("Creating ou %s in %s", name, baseOU)

	tmp, err := api.getOU(name, baseOU)
	if err != nil {
		return fmt.Errorf("createOU - talking to active directory failed: %s", err)
	}

	// there is already an ou object with the same name
	if tmp != nil {
		if tmp.name == name && tmp.dn == fmt.Sprintf("ou=%s,%s", name, baseOU) {
			log.Infof("OU object %s already exists, updating description", name)
			return api.updateOUDescription(name, baseOU, description)
		}

		return fmt.Errorf("createOU - ou object %s already exists under this base ou %s", name, baseOU)
	}

	attributes := make(map[string][]string)
	attributes["name"] = []string{name}
	attributes["ou"] = []string{name}
	attributes["description"] = []string{description}

	return api.createObject(fmt.Sprintf("ou=%s,%s", name, baseOU), []string{"organizationalUnit", "top"}, attributes)
}

// moves an existing ou object to a new ou
func (api *API) moveOU(cn, baseOU, newOU string) error {
	log.Infof("Moving ou object %s from %s to %s.", cn, baseOU, newOU)

	tmp, err := api.getOU(cn, baseOU)
	if err != nil {
		return fmt.Errorf("moveOU - talking to active directory failed: %s", err)
	}

	if tmp == nil {
		return fmt.Errorf("moveOU - ou object %s does not exists under %s: %s", cn, baseOU, err)
	}

	// ou object is already in the target OU, nothing to do
	if tmp.dn == fmt.Sprintf("ou=%s,%s", cn, newOU) {
		log.Infof("OU object is already under the target ou")
		return nil
	}

	// specific uid of the ou
	UID := fmt.Sprintf("ou=%s", cn)

	// move ou object to new ou
	req := ldap.NewModifyDNRequest(fmt.Sprintf("ou=%s,%s", cn, baseOU), UID, true, newOU)
	if err := api.client.ModifyDN(req); err != nil {
		return fmt.Errorf("moveOU - failed to move ou: %s", err)
	}

	log.Infof("OU moved.")
	return nil
}

// updates the description of an existing ou object
func (api *API) updateOUDescription(cn, baseOU, description string) error {
	log.Infof("Updating description of ou %s under %s", cn, baseOU)
	return api.updateObject(fmt.Sprintf("ou=%s,%s", cn, baseOU), nil, nil, map[string][]string{
		"description": {description},
	}, nil)
}

// updates the name of an existing ou object
func (api *API) updateOUName(name, baseOU, newName string) error {
	log.Infof("Updating name of ou %s under %s.", name, baseOU)
	return api.moveOU(name, newName, baseOU)
}

// deletes an existing ou object.
func (api *API) deleteOU(dn string) error {
	log.Infof("Deleting ou %s.", dn)

	objects, err := api.searchObject("(objectclass=*)", dn, nil)
	if err != nil {
		return fmt.Errorf("deleteOU - failed remove ou %s: %s", dn, err)
	}

	if len(objects) > 0 {
		return fmt.Errorf("deleteOU - failed to delete ou %s because it has child items", dn)
	}

	return api.deleteObject(dn)
}
