package ldap

import (
	"gopkg.in/ldap.v3"
)

func decodeLDAPAttributes(attributes []*ldap.EntryAttribute) map[string][]string {
	attr := make(map[string][]string)

	for _, e := range attributes {
		attr[e.Name] = e.Values
	}

	return attr
}
