package activedirectory

import (
	"gopkg.in/ldap.v3"
)

func decodeADAttributes(attributes []*ldap.EntryAttribute) map[string][]string {
	attr := make(map[string][]string)

	for _, e := range attributes {
		attr[e.Name] = e.Values
	}

	return attr
}
