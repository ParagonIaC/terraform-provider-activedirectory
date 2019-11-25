package activedirectory

import (
	"crypto/tls"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// APIInterface is the basic interface for AD API
type APIInterface interface {
	connect() error
	getDomainDN() string

	// generic objects
	searchObject(filter, baseDN string, attributes []string) ([]*Object, error)
	getObject(dn string, attributes []string) (*Object, error)
	createObject(dn string, class []string, attributes map[string][]string) error
	deleteObject(dn string) error
	updateObject(dn string, classes []string, added, changed, removed map[string][]string) error

	// comupter objects
	getComputer(name string) (*Computer, error)
	createComputer(cn, ou, description string) error
	updateComputerOU(cn, ou, newOU string) error
	updateComputerDescription(cn, ou, description string) error
	deleteComputer(cn, ou string) error

	// ou objects
	getOU(name, baseOU string) (*OU, error)
	createOU(dn, cn, description string) error
	moveOU(cn, ou, newOU string) error
	updateOUName(name, baseOU, newName string) error
	updateOUDescription(cn, ou, description string) error
	deleteOU(dn string) error
}

// API is the basic struct which should implement the interface
type API struct {
	host     string
	port     int
	domain   string
	useTLS   bool
	user     string
	password string
	client   ldap.Client
}

// connects to an Active Directory server
func (api *API) connect() error {
	log.Infof("Connecting to %s:%d.", api.host, api.port)

	if api.host == "" {
		return fmt.Errorf("connect - no ad host specified")
	}

	if api.domain == "" {
		return fmt.Errorf("connect - no ad domain specified")
	}

	if api.user == "" {
		return fmt.Errorf("connect - no bind user specified")
	}

	client, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", api.host, api.port))
	if err != nil {
		return fmt.Errorf("connect - failed to connect: %s", err)
	}

	if api.useTLS {
		log.Info("Configuring client to use secure connection.")
		if err = client.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
			return fmt.Errorf("connect - failed to use secure connection: %s", err)
		}
	}

	log.Infof("Authenticating user %s@%s.", api.user, api.domain)
	if err = client.Bind(fmt.Sprintf("%s@%s", api.user, api.domain), api.password); err != nil {
		client.Close()
		return fmt.Errorf("connect - authentication failed: %s", err)
	}

	api.client = client

	log.Infof("Connected successfully to %s:%d.", api.host, api.port)
	return nil
}

func (api *API) getDomainDN() string {
	tmp := strings.Split(api.domain, ".")
	return fmt.Sprintf("dc=%s", strings.Join(tmp, ",dc="))
}
