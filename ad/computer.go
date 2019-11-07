package activedirectory

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// Computer - main struct to hold ad computer object data
type Computer struct {
	Name           string
	DN string
	Description    string
}

// GetComputerByName queries AD for a sepcific computer account
func (api *API) GetComputerByName(name string, ou string) (computer *Computer, err error) {
	if name == "" {
		return nil, fmt.Errorf("no computer name specified")
	}

	if ou == "" {
		// TODO
		// ou = api.domain
	}

	searchRequest := ldap.NewSearchRequest(
		ou, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=Computer)(cn="+name+"))",
		[]string{"cn", "distinguishedName", "description"},
		nil,
	)

	sr, err := api.client.Search(searchRequest)
	if err != nil {
		log.Error("Error will searching for computer %s: %s:", name, err)
		return nil, err
	}

	if len(sr.Entries) == 0 {
		return nil, fmt.Errorf("Computer with the name %s was not found under %s", name, ou)
	} else if len(sr.Entries) > 1 {
		return nil, fmt.Errorf("More than one computer with the name %s under %s", name, ou)
	}

	return &Computer{
		Name:           sr.Entries[0].GetAttributeValue("cn"),
		DN: sr.Entries[0].GetAttributeValue("distinguishedName"),
		Description:    sr.Entries[0].GetAttributeValue("description"),
	}, nil
}

// CreateComputer create a new ad computer object.
func (api *API) CreateComputer (computer Computer) (computer *Computer, err error) {
	addRequest := ldap.NewAddRequest(dnName, nil)
	addRequest.Attribute("objectClass", []string{"computer"})
	addRequest.Attribute("name", []string{computer.Name})
	addRequest.Attribute("sAMAccountName", []string{computer.Name + "$"})
	addRequest.Attribute("userAccountControl", []string{"4096"})
	if computer.Description != "" {
		addRequest.Attribute("description", []string{computer.Description})
	}
	err := adConn.Add(addRequest)
	if err != nil {
		return err
	}
	return nil
}

// UpdateComputer
func (api *API) UpdateComputer (computer Computer) (computer *Computer, err error) {

}

// DeleteComputer
func (api *API) DeleteComputer (dnName string) (computer *Computer, err error) {
	log.Info("Deleting AD computer object %s", dnName)
	req := ldap.NewDelRequest(dnName, nil)
	
	if err := api.client.Del(req); err != nil {
		return err
	}
	return nil
}
