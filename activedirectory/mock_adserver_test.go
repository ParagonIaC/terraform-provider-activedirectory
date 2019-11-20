package activedirectory

import (
	"fmt"
	"net"

	"github.com/nmcclain/ldap"
	log "github.com/sirupsen/logrus"
)

type adHandler struct {
}

func (h adHandler) Bind(bindDN, bindPw string, conn net.Conn) (ldap.LDAPResultCode, error) {
	log.Errorf(bindDN, bindPw)
	if bindDN == "Tester" && bindPw == "Password" {
		return ldap.LDAPResultSuccess, nil
	}
	return ldap.LDAPResultUnavailable, fmt.Errorf("authentication failed")
}

func getADServer(host string, port int) (f func()) {
	var server *ldap.Server

	f = func() {
		log.Errorf("Creating test AD Server Failed:")
		server = ldap.NewServer()
		handler := adHandler{}
		server.BindFunc("", handler)
		if err := server.ListenAndServe(fmt.Sprintf("%s:%d", host, port)); err != nil {
			log.Errorf("AD Server Failed: %s", err.Error())
		}
	}

	return
}
