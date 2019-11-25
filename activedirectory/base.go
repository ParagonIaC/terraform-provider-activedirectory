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
	getComputer(name, baseOU string, attributes []string) (*Computer, error)
	createComputer(dn, cn string, attributes map[string][]string) error
	updateComputerOU(dn, cn, ou string) error
	updateComputerAttributes(dn string, added map[string][]string, changed map[string][]string, removed map[string][]string) error
	deleteComputer(dn string) error

	// ou objects
	getOU(name, baseOU string) (*OU, error)
	createOU(dn, cn, description string) error
	moveOU(dn, cn, ou string) error
	updateOUName(dn, name string) error
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
func (api *API) connect() error {
	log.Infof("Connecting to %s:%d with bind user %s.", api.adHost, api.adPort, api.bindUser)

	if api.adHost == "" {
		return fmt.Errorf("connect - no ad host specified")
	}

	client, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", api.adHost, api.adPort))
	if err != nil {
		return fmt.Errorf("connect - failed to connect: %s", err)
	}

	if api.useTLS {
		log.Info("Configuring client to use secure connection.")
		if err = client.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
			return fmt.Errorf("connect - failed to use secure connection: %s", err)
		}
	}

	if api.bindUser == "" {
		return fmt.Errorf("connect - no bind user specified")
	}

	log.Infof("Authenticating user %s.", api.bindUser)
	if err = client.Bind(api.bindUser, api.bindPassword); err != nil {
		client.Close()
		return fmt.Errorf("connect - authentication failed: %s", err)
	}

	api.client = client

	log.Infof("Connected successfully to %s:%d as user %s", api.adHost, api.adPort, api.bindUser)
	return nil
}
