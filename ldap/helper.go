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

func encodeLDAPAttributes(attributes map[string][]string) []*ldap.EntryAttribute {
	attr := make([]*ldap.EntryAttribute, len(attributes))

	i := 0
	for key, value := range attributes {
		attr[i] = &ldap.EntryAttribute{
			Name:   key,
			Values: value,
		}
		i++
	}

	return attr
}
