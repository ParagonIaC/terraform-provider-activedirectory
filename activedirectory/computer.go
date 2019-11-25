package activedirectory

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// Computer is the base implementation of ad computer object
type Computer struct {
	name       string
	dn         string
	attributes map[string][]string
}

// returns computer object
func (api *API) getComputer(name, baseOU string, attributes []string) (*Computer, error) {
	dn := fmt.Sprintf("ou=%s,%s", name, baseOU)
	log.Infof("Trying to get ad ou: %s", dn)

	attributes = append(attributes, "cn")

	// filter
	filter := fmt.Sprintf("(&(objectclass=computer)(name=%s))", name)

	// trying to get ou object
	ret, err := api.searchObject(filter, baseOU, attributes)
	if err != nil {
		if err, ok := err.(*ldap.Error); ok {
			if err.ResultCode == 32 {
				log.Info("AD computer object could not be found", dn)
				return nil, nil
			}
		}
		log.Errorf("Error will searching for ad computer object %s: %s:", dn, err)
		return nil, err
	}

	if len(ret) != 1 {
		return nil, nil
	}

	if ret == nil {
		return nil, nil
	}

	return &Computer{
		name:       strings.Join(ret[0].attributes["cn"], ""),
		dn:         strings.ToLower(ret[0].dn),
		attributes: ret[0].attributes,
	}, nil
}

// creates a new computer object
func (api *API) createComputer(dn, cn string, attributes map[string][]string) error {
	log.Infof("Creating computer object %s", dn)

	if attributes == nil {
		attributes = make(map[string][]string)
	}

	attributes["name"] = []string{cn}
	attributes["sAMAccountName"] = []string{cn + "$"}
	attributes["userAccountControl"] = []string{"4096"}

	return api.createObject(dn, []string{"computer"}, attributes)
}

// moves an existing computer object to a new ou
func (api *API) updateComputerOU(dn, cn, ou string) error {
	log.Infof("Moving computer object %s to ou %s", dn, ou)

	base := dn[(strings.Index(strings.ToLower(dn), "dc=")):]

	tmp, err := api.getComputer(cn, base, nil)
	if err != nil {
		return err
	}

	if tmp != nil {
		if tmp.dn != fmt.Sprintf("cn=%s,%s", cn, ou) {
			log.Infof("Computer object is already in the target ou")
			return nil
		}
	}

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
func (api *API) updateComputerAttributes(dn string, added, changed, removed map[string][]string) error {
	log.Infof("updating attributes of computer object %s", dn)
	return api.updateObject(dn, nil, added, changed, removed)
}

// deletes an existing computer object.
func (api *API) deleteComputer(dn string) error {
	log.Infof("Deleting computer object %s", dn)
	return api.deleteObject(dn)
}
