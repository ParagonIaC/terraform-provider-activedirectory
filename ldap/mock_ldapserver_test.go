package ldap

import (
	"fmt"
	"net"

	"github.com/nmcclain/ldap"
	log "github.com/sirupsen/logrus"
)

type ldapHandler struct {
}

func (h ldapHandler) Bind(bindDN, bindPw string, conn net.Conn) (ldap.LDAPResultCode, error) {
	log.Errorf(bindDN, bindPw)
	if bindDN == "Tester" && bindPw == "Password" {
		return ldap.LDAPResultSuccess, nil
	}
	return ldap.LDAPResultUnavailable, fmt.Errorf("Authentication failed")
}

func getLDAPServer(host string, port int) *ldap.Server {
	s := ldap.NewServer()
	handler := ldapHandler{}
	s.BindFunc("", handler)
	if err := s.ListenAndServe(fmt.Sprintf("%s:%d", host, port)); err != nil {
		log.Errorf("LDAP Server Failed: %s", err.Error())
	}

	return s
}
