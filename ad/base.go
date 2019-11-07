package activedirectory

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ldap.v3"
)

// APIInterface is the basic interface for AD API
type APIInterface interface {
	connect() error
}

// API is the basic struct which should implement the APIInterface
type API struct {
	domain   string
	ip       string
	username string
	password string
	client   *ldap.Conn
}

// NewAPI create a AD API object
func NewAPI(ip string, domain string) (api *API) {
	return &API{ip: ip, domain: domain}
}

// Connect connects to an Active Directory server
//	username - string
// 	password - string
func (api *API) Connect(username string, password string) (err error) {
	api.username = fmt.Sprintf("%s@%s", username, api.domain)
	api.password = password

	dial, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", api.ip, 389))
	if err != nil {
		log.Error("Connection to %s:%d failed: %s", api.ip, 389, err)
		return err
	}

	if err = dial.Bind(username, password); err != nil {
		log.Error("Authentication failed: %s", err)
		return err
	}

	log.Debug("AD connection successful for user: %s", api.username)
	return nil
}
