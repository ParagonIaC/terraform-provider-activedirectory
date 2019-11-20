package activedirectory

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/ldap.v3"
)

func TestDecodeADAttributes(t *testing.T) {
	t.Run("decodeADAttributes - should return map[string][]string", func(t *testing.T) {
		ret := decodeADAttributes(nil)
		assert.IsType(t, ret, map[string][]string{})
	})

	t.Run("decodeADAttributes - should map ldap.EntryAttribute to map[string]*", func(t *testing.T) {
		attributes := make([]*ldap.EntryAttribute, 10)
		for i := 0; i < len(attributes); i++ {
			attributes[i] = &ldap.EntryAttribute{
				Name:   fmt.Sprintf("Attr%d", i),
				Values: []string{fmt.Sprintf("Value%d", i)},
			}
		}

		ret := decodeADAttributes(attributes)

		assert.Equal(t, len(attributes), len(ret))
		for _, e := range attributes {
			assert.True(t, reflect.DeepEqual(e.Values, ret[e.Name]))
		}
	})
}
