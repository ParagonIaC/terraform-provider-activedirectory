package ldap

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// LDAPInterface is the basic interface for LDAP API
type LDAPInterface interface {
	Connect(string, string) error

	// generel ldap object
	GetObjectBy(dn string) (*LDAPObject, error)
	SearchObject(query string) ([]*LDAPObject, error)

	CreateObject(dn string, objType LDAPType, attributes map[string]string) error
	DeleteObject(dn string) error

	UpdateObject(dn string, attributes map[string]string) error

	// comupter object part
	GetComputersByLDAPFilter(string, string, []string) ([]*Computer, error)
	GetComputerByDN(string, string, []string) (*Computer, error)

	CreateComputer(*Computer, string) error

	UpdateComputerOU(*Computer, string) error
	UpdateComputerAttributes(*Computer, []*ldap.EntryAttribute) error

	DeleteComputer(dn string) error
}

// LDAP is the basic struct which should implement the LDAPInterface
type LDAP struct {
	ldapHost     string
	ldapPort     int
	useTLS       bool
	bindUser     string
	bindPassword string
	client       *ldap.Conn
}

// NewAPI create a AD API object
func NewLDAPConnection(host string, port int, useTLS bool) *LDAP {
	return &LDAP{ldapHost: host, ldapPort port, useTLS: useTLS}
}

// Connect connects to an Active Directory server
//	username - string
// 	password - string
func (api *LDAP) Connect(username, password string) (err error) {
	api.bindUser = username
	api.bindPassword = password

	api.client, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", api.ldapHost, api.ldapPort))
	if err != nil {
		log.Errorf("Connection to %s:%d failed: %s", api.ldapHost, api.ldapPort, err)
		return err
	}

	if c.UseTLS {
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

	log.Debugf("AD connection successful for user: %s", api.bindUser)
	return nil
}
