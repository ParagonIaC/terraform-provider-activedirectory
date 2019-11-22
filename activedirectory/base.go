package activedirectory

import (
	"crypto/tls"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// APIInterface is the basic interface for AD API
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

	// ou objects
	getOU(dn string) (*OU, error)
	createOU(dn, cn, description string) error
	moveOU(dn, cn, ou string) error
	updateOUDescription(dn, description string) error
	deleteOU(dn string) error
}

// API is the basic struct which should implement the interface
type API struct {
	adHost       string
	adPort       int
	useTLS       bool
	bindUser     string
	bindPassword string
	client       ldap.Client
}

// connects to an Active Directory server
func (api *API) connect() (err error) {
	log.Debugf("Trying AD connection with user %s to server %s", api.bindUser, api.adHost)

	api.client, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", api.adHost, api.adPort))
	if err != nil {
		log.Errorf("Connection to %s:%d failed: %s", api.adHost, api.adPort, err)
		return err
	}

	if api.useTLS {
		if err = api.client.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
			api.client = nil
			return err
		}
	}

	if err = api.client.Bind(api.bindUser, api.bindPassword); err != nil {
		log.Errorf("Authentication failed: %s", err)
		api.client.Close()
		return err
	}

	log.Debugf("AD connection successful for user: %s", api.bindUser)
	return nil
}
