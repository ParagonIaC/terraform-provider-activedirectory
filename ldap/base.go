package ldap

import (
	"crypto/tls"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// APIInterface is the basic interface for LDAP API
type APIInterface interface {
	connect() error

	// generic objects
	searchObject(filter, baseDN string, attributes []string) ([]*Object, error)
	getObject(dn string, attributes []string) (*Object, error)
	createObject(dn string, class []string, attributes map[string][]string) error
	deleteObject(dn string) error
	updateObject(dn string, classes []string, added map[string][]string, changed map[string][]string, removed map[string][]string) error

	// comupter objects
	getComputer(dn string, attributes []string) (*Computer, error)
	createComputer(dn, cn string, attributes map[string][]string) error
	updateComputerOU(dn, cn, ou string) error
	updateComputerAttributes(dn string, added map[string][]string, changed map[string][]string, removed map[string][]string) error
	deleteComputer(dn string) error
}

// API is the basic struct which should implement the interface
type API struct {
	ldapHost     string
	ldapPort     int
	useTLS       bool
	bindUser     string
	bindPassword string
	client       ldap.Client
}

// connects to an Active Directory server
func (api *API) connect() (err error) {
	log.Debugf("Trying LDAP connection with user %s to server %s", api.bindUser, api.ldapHost)

	api.client, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", api.ldapHost, api.ldapPort))
	if err != nil {
		log.Errorf("Connection to %s:%d failed: %s", api.ldapHost, api.ldapPort, err)
		return err
	}

	if api.useTLS {
		if err := api.client.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
			api.client = nil
			return err
		}
	}

	if err = api.client.Bind(api.bindUser, api.bindPassword); err != nil {
		log.Errorf("Authentication failed: %s", err)
		api.client.Close()
		return err
	}

	log.Debugf("LDAP connection successful for user: %s", api.bindUser)
	return nil
}
