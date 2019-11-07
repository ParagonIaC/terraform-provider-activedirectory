package activedirectory

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// Computer - main struct to hold ad computer object data
type Computer struct {
	Name           string
	SAMAccountName string
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
		SAMAccountName: sr.Entries[0].GetAttributeValue("distinguishedName"),
		Description:    sr.Entries[0].GetAttributeValue("description"),
	}, nil
}
