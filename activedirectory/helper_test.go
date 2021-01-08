package activedirectory

import (
	"github.com/go-ldap/ldap/v3"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecodeADAttributes(t *testing.T) {
	t.Run("decodeADAttributes - should return map[string][]string", func(t *testing.T) {
		ret := decodeADAttributes(nil)
		assert.IsType(t, ret, map[string][]string{})
	})

	t.Run("decodeADAttributes - should map ldap.EntryAttribute to map[string]*", func(t *testing.T) {
		rand.Seed(time.Now().UnixNano())
		num := rand.Intn(10)

		attributes := make([]*ldap.EntryAttribute, num)
		for i := 0; i < len(attributes); i++ {
			attributes[i] = &ldap.EntryAttribute{
				Name:   getRandomString(10),
				Values: []string{getRandomString(10)},
			}
		}

		ret := decodeADAttributes(attributes)

		assert.Equal(t, len(attributes), len(ret))
		for _, e := range attributes {
			assert.True(t, reflect.DeepEqual(e.Values, ret[e.Name]))
		}
	})
}
