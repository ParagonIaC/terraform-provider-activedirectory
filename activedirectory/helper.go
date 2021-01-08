package activedirectory

import (
	"github.com/go-ldap/ldap/v3"
)

func decodeADAttributes(attributes []*ldap.EntryAttribute) map[string][]string {
	attr := make(map[string][]string)

	for _, e := range attributes {
		attr[e.Name] = e.Values
	}

	return attr
}
