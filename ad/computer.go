package activedirectory

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// Computer - main struct to hold ad computer object data
type Computer struct {
	Name       string
	DN         string
	Attributes []*ldap.EntryAttribute
}

// GetComputersByName queries AD for a sepcific computer account
func (api *API) GetComputersByName(name string, ou string, attributes []string) (computer []*Computer, err error) {
	if name == "" {
		return nil, fmt.Errorf("no computer name specified")
	}

	// if no ou is specified, sear whole domain
	if ou == "" {
		tmp := strings.Split(api.domain, ".")
		ou = fmt.Sprintf("dc=%s", strings.Join(tmp, ",dc="))
	}

	// prepare for search request
	ldapFilter := "(&(objectClass=Computer)(cn=" + name + "))"
	attributes = append(attributes, "cn", "distinguishedName")

	// create search request
	searchRequest := ldap.NewSearchRequest(ou,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		ldapFilter, attributes, nil,
	)

	// performing ldap search
	result, err := api.client.Search(searchRequest)
	if err != nil {
		log.Error("Error will searching for computer %s: %s:", name, err)
		return nil, err
	}

	computer = make([]*Computer, len(result.Entries))
	for i, entry := range result.Entries {
		computer[i] = &Computer{
			Name:       entry.GetAttributeValue("cn"),
			DN:         entry.GetAttributeValue("distinguishedName"),
			Attributes: entry.Attributes,
		}
	}

	return computer, nil
}

// CreateComputer create a new ad computer object.
func (api *API) CreateComputer(computer *Computer, ou string) error {
	log.Infof("Creating ad computer object %s in ou %s", computer.Name, ou)

	// create dn name of computer object
	dnName := fmt.Sprintf("cn=%s,%s", computer.Name, ou)

	// create ldap add request
	req := ldap.NewAddRequest(dnName, nil)
	req.Attribute("objectClass", []string{"computer"})
	req.Attribute("name", []string{computer.Name})
	req.Attribute("sAMAccountName", []string{computer.Name + "$"})
	req.Attribute("userAccountControl", []string{"4096"})

	// add all requested attributes
	for _, value := range computer.Attributes {
		req.Attribute(value.Name, value.Values)
	}

	// add to ldap
	if err := api.client.Add(req); err != nil {
		return err
	}

	// update dn parameter in Computer object
	computer.DN = dnName

	return nil
}

// UpdateComputerOU moves an existing AD computer object to a new OU.
func (api *API) UpdateComputerOU(computer *Computer, ou string) error {
	log.Infof("moving ad computer object %s to ou %s", computer.Name, ou)

	// specific uid of the computer
	computerUID := fmt.Sprintf("uid=%", computer.Name)

	// move computer object to new ou
	req := ldap.NewModifyDNRequest(computer.DN, computerUID, true, ou)
	if err := api.client.ModifyDN(req); err != nil {
		return err
	}

	// update DN to reflect ou change
	computer.DN = fmt.Sprintf("cn=%s,%s", computer.Name, ou)

	return nil
}

// UpdateComputerAttributes updates the attributes of an existing AD computer.
func (api *API) UpdateComputerAttributes(computer *Computer, attributes []*ldap.EntryAttribute) error {
	log.Infof("updaing attributes for ad computer objects %s", computer.Name)

	req := ldap.NewModifyRequest(computer.DN, nil)

	// loop through all attributes
	for _, value := range attributes {
		if len(value.Values) == 0 {
			req.Delete(value.Name, []string{})
		} else {
			req.Replace(value.Name, value.Values)
		}
	}

	// modify ldap object
	if err := api.client.Modify(req); err != nil {
		return err
	}

	// loop through all attributes to update computer object
	for _, value := range attributes {
		for _, tmpValue := range computer.Attributes {
			if value.Name == tmpValue.Name {
				tmpValue.Values = value.Values
			}
		}
	}

	return nil
}

// DeleteComputer delete an existing computer object.
func (api *API) DeleteComputer(computer Computer) error {
	log.Infof("Deleting AD computer object %s", computer.DN)

	// create ldap delete request
	req := ldap.NewDelRequest(computer.DN, nil)

	// delete object from ldap
	if err := api.client.Del(req); err != nil {
		return err
	}
	return nil
}
