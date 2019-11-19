// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// File contains Modify functionality
//
// https://tools.ietf.org/html/rfc4511
//
// ModifyRequest ::= [APPLICATION 6] SEQUENCE {
//      object          LDAPDN,
//      changes         SEQUENCE OF change SEQUENCE {
//           operation       ENUMERATED {
//                add     (0),
//                delete  (1),
//                replace (2),
//                ...  },
//           modification    PartialAttribute } }
//
// PartialAttribute ::= SEQUENCE {
//      type       AttributeDescription,
//      vals       SET OF value AttributeValue }
//
// AttributeDescription ::= LDAPString
//                         -- Constrained to <attributedescription>
//                         -- [RFC4512]
//
// AttributeValue ::= OCTET STRING
//

package ldap

import (
	"errors"
	"log"

	"github.com/nmcclain/asn1-ber"
)

const (
	AddAttribute     = 0
	DeleteAttribute  = 1
	ReplaceAttribute = 2
)

var LDAPModifyAttributeMap = map[uint64]string{
	AddAttribute:     "Add",
	DeleteAttribute:  "Delete",
	ReplaceAttribute: "Replace",
}

type PartialAttribute struct {
	AttrType string
	AttrVals []string
}

func (p *PartialAttribute) encode() *ber.Packet {
	seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "PartialAttribute")
	seq.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, p.AttrType, "Type"))
	set := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "AttributeValue")
	for _, value := range p.AttrVals {
		set.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, value, "Vals"))
	}
	seq.AppendChild(set)
	return seq
}

type ModifyRequest struct {
	Dn                string
	AddAttributes     []PartialAttribute
	DeleteAttributes  []PartialAttribute
	ReplaceAttributes []PartialAttribute
}

func (m *ModifyRequest) Add(attrType string, attrVals []string) {
	m.AddAttributes = append(m.AddAttributes, PartialAttribute{AttrType: attrType, AttrVals: attrVals})
}

func (m *ModifyRequest) Delete(attrType string, attrVals []string) {
	m.DeleteAttributes = append(m.DeleteAttributes, PartialAttribute{AttrType: attrType, AttrVals: attrVals})
}

func (m *ModifyRequest) Replace(attrType string, attrVals []string) {
	m.ReplaceAttributes = append(m.ReplaceAttributes, PartialAttribute{AttrType: attrType, AttrVals: attrVals})
}

func (m ModifyRequest) encode() *ber.Packet {
	request := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationModifyRequest, nil, "Modify Request")
	request.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, m.Dn, "DN"))
	changes := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Changes")
	for _, attribute := range m.AddAttributes {
		change := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Change")
		change.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(AddAttribute), "Operation"))
		change.AppendChild(attribute.encode())
		changes.AppendChild(change)
	}
	for _, attribute := range m.DeleteAttributes {
		change := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Change")
		change.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(DeleteAttribute), "Operation"))
		change.AppendChild(attribute.encode())
		changes.AppendChild(change)
	}
	for _, attribute := range m.ReplaceAttributes {
		change := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "Change")
		change.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, uint64(ReplaceAttribute), "Operation"))
		change.AppendChild(attribute.encode())
		changes.AppendChild(change)
	}
	request.AppendChild(changes)
	return request
}

func NewModifyRequest(
	dn string,
) *ModifyRequest {
	return &ModifyRequest{
		Dn: dn,
	}
}

func (l *Conn) Modify(modifyRequest *ModifyRequest) error {
	messageID := l.nextMessageID()
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Request")
	packet.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "MessageID"))
	packet.AppendChild(modifyRequest.encode())

	l.Debug.PrintPacket(packet)

	channel, err := l.sendMessage(packet)
	if err != nil {
		return err
	}
	if channel == nil {
		return NewError(ErrorNetwork, errors.New("ldap: could not send message"))
	}
	defer l.finishMessage(messageID)

	l.Debug.Printf("%d: waiting for response", messageID)
	packet = <-channel
	l.Debug.Printf("%d: got response %p", messageID, packet)
	if packet == nil {
		return NewError(ErrorNetwork, errors.New("ldap: could not retrieve message"))
	}

	if l.Debug {
		if err := addLDAPDescriptions(packet); err != nil {
			return err
		}
		ber.PrintPacket(packet)
	}

	if packet.Children[1].Tag == ApplicationModifyResponse {
		resultCode, resultDescription := getLDAPResultCode(packet)
		if resultCode != 0 {
			return NewError(resultCode, errors.New(resultDescription))
		}
	} else {
		log.Printf("Unexpected Response: %d", packet.Children[1].Tag)
	}

	l.Debug.Printf("%d: returning", messageID)
	return nil
}
