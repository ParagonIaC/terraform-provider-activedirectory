package ldap

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// Computer is the base implementation of ldap computer object
type Computer struct {
	name       string
	dn         string
	attributes map[string][]string
}

// returns computer object
func (api *API) getComputer(dn string, attributes []string) (*Computer, error) {
	attributes = append(attributes, "cn")

	// trying to get computer object
	ret, err := api.getObject(dn, attributes)
	if err != nil {
		return nil, err
	}

	return &Computer{
		name:       strings.Join(ret.attributes["cn"], ""),
		dn:         ret.dn,
		attributes: ret.attributes,
	}, nil
}

// creates a new computer object
func (api *API) createComputer(dn, cn string, attributes map[string][]string) error {
	log.Infof("Creating computer object %s", dn)

	attributes["name"] = []string{cn}
	attributes["sAMAccountName"] = []string{cn + "$"}
	attributes["userAccountControl"] = []string{"4096"}

	return api.createObject(dn, []string{"computer"}, attributes)
}

// moves an existing computer object to a new ou
func (api *API) updateComputerOU(dn, cn, ou string) error {
	log.Infof("Moving computer object %s to ou %s", dn, ou)

	// specific uid of the computer
	computerUID := fmt.Sprintf("cn=%s", cn)

	// move computer object to new ou
	req := ldap.NewModifyDNRequest(dn, computerUID, true, ou)
	if err := api.client.ModifyDN(req); err != nil {
		log.Errorf("Moving object %s to %s failed: %s", dn, ou, err)
		return err
	}

	log.Infof("Object %s moved", dn)

	return nil
}

// updates the attributes of an existing computer object
func (api *API) updateComputerAttributes(dn string, added map[string][]string, changed map[string][]string, removed map[string][]string) error {
	log.Infof("updating attributes of computer object %s", dn)
	return api.updateObject(dn, []string{"computer"}, added, changed, removed)
}

// deletes an existing computer object.
func (api *API) deleteComputer(dn string) error {
	log.Infof("Deleting computer object %s", dn)
	return api.deleteObject(dn)
}
