package activedirectory

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//thanks to https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano()) //nolint

func getRandomString(n int) string {
	b := make([]byte, n)

	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func getRandomDomain(n int) string {
	domain := make([]string, n)

	for i := 0; i < n; i++ {
		domain[i] = getRandomString(5)
	}

	return strings.Join(domain, ".")
}

func getRandomOU(n, m int) string {
	ou := make([]string, n+m)

	for i := 0; i < n; i++ {
		ou[i] = fmt.Sprintf("ou=%s", getRandomString(5))
	}

	for i := 0; i < m; i++ {
		ou[i+n] = fmt.Sprintf("dc=%s", getRandomString(3))
	}

	return strings.Join(ou, ",")
}

func TestConnect(t *testing.T) {
	host := "127.0.0.1"
	port := 11389

	t.Run("connect - should fail when host is not specified", func(t *testing.T) {
		api := &API{}
		err := api.connect()
		assert.Error(t, err)
	})

	t.Run("connect - should fail when domain is not specified", func(t *testing.T) {
		api := &API{host: host}
		err := api.connect()
		assert.Error(t, err)
	})

	t.Run("connect - should fail when user is not specified", func(t *testing.T) {
		api := &API{
			host:   host,
			domain: "domain",
		}
		err := api.connect()
		assert.Error(t, err)
	})

	t.Run("connect - should fail when no server is reachable", func(t *testing.T) {
		api := &API{
			host:     host,
			port:     port,
			domain:   "domain",
			user:     "Tester",
			password: "wrong",
		}
		err := api.connect()

		assert.Error(t, err)
	})

	go getADServer(host, port)()

	// give ad server time to start
	time.Sleep(1000 * time.Millisecond)

	t.Run("connect - should fail when useTLS is set and TLS is not working", func(t *testing.T) {
		api := &API{
			host:     host,
			port:     port,
			useTLS:   true, // mock AD server has no TLS
			user:     "Tester",
			password: "Password",
			domain:   "domain",
		}
		err := api.connect()

		assert.Error(t, err)
		assert.Equal(t, nil, api.client)
	})

	t.Run("connect - should fail when authentication fails", func(t *testing.T) {
		api := &API{
			host:     host,
			port:     port,
			user:     "Tester",
			password: "wrong",
			domain:   "domain",
		}
		err := api.connect()

		assert.Error(t, err)
	})

	t.Run("connect - should return nil when everything is okay", func(t *testing.T) {
		api := &API{
			host:     host,
			port:     port,
			user:     "Tester",
			password: "Password",
			domain:   "domain.org",
		}
		err := api.connect()

		assert.NoError(t, err)
	})
}

func TestGetDomainDN(t *testing.T) {
	t.Run("getDomainDN - should return dn encoded version of domain", func(t *testing.T) {
		api := &API{
			domain: "test.example.com",
		}

		assert.Equal(t, "dc=test,dc=example,dc=com", api.getDomainDN())
	})

	t.Run("getDomainDN - should ignore case", func(t *testing.T) {
		api := &API{
			domain: "TEST.Example.ORG",
		}

		assert.Equal(t, "dc=test,dc=example,dc=org", api.getDomainDN())
	})
}
