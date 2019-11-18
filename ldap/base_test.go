package ldap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	host := "127.0.0.1"
	port := 3809

	t.Run("connect - should fail when no server is reachable", func(t *testing.T) {
		api := &API{
			ldapHost:     host,
			ldapPort:     port,
			bindUser:     "Tester",
			bindPassword: "wrong",
		}
		err := api.connect()

		assert.Error(t, err)
	})

	go getLDAPServer(host, port)

	// give ldap server time to start
	time.Sleep(500 * time.Millisecond)

	t.Run("connect - should fail when authentication fails", func(t *testing.T) {
		api := &API{
			ldapHost:     host,
			ldapPort:     port,
			bindUser:     "Tester",
			bindPassword: "wrong",
		}
		err := api.connect()

		assert.Error(t, err)
	})

	t.Run("connect - should return nil when everything is okay", func(t *testing.T) {
		api := &API{
			ldapHost:     host,
			ldapPort:     port,
			bindUser:     "Tester",
			bindPassword: "Password",
		}
		err := api.connect()

		assert.NoError(t, err)
	})

	t.Run("connect - should use fail when useTSL is set and TSL is not working", func(t *testing.T) {
		api := &API{
			ldapHost:     host,
			ldapPort:     port,
			useTLS:       true,
			bindUser:     "Tester",
			bindPassword: "Password",
		}
		err := api.connect()

		assert.Error(t, err)
		assert.Equal(t, nil, api.client)
	})
}
