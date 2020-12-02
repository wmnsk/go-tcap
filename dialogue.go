// Copyright 2019-2020 go-tcap authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package tcap

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// Dialogue OID: Dialogue-As-ID and Unidialogue-As-Id.
const (
	DialogueAsID uint8 = iota + 1
	UnidialogueAsID
)

// Dialogue represents a Dialogue Portion of TCAP.
type Dialogue struct {
	Tag              Tag
	Length           uint8
	ExternalTag      Tag
	ExternalLength   uint8
	ObjectIdentifier *IE
	SingleAsn1Type   *IE
	DialoguePDU      *DialoguePDU
	Payload          []byte
}

// NewDialogue creates a new Dialogue with the DialoguePDU given.
func NewDialogue(oid, ver uint8, pdu *DialoguePDU, payload []byte) *Dialogue {
	d := &Dialogue{
		Tag:         NewApplicationWideConstructorTag(11),
		ExternalTag: NewUniversalConstructorTag(8),
		ObjectIdentifier: &IE{
			Tag:    NewUniversalPrimitiveTag(6),
			Length: 7,
			Value:  []byte{0, 17, 134, 5, 1, oid, ver},
		},
		SingleAsn1Type: &IE{
			Tag:    NewContextSpecificConstructorTag(0),
			Length: uint8(pdu.MarshalLen()),
		},
		DialoguePDU: pdu,
		Payload:     payload,
	}
	d.SetLength()

	return d
}

// MarshalBinary returns the byte sequence generated from a Dialogue instance.
func (d *Dialogue) MarshalBinary() ([]byte, error) {
	b := make([]byte, d.MarshalLen())
	if err := d.MarshalTo(b); err != nil {
		return nil, errors.Wrap(err, "failed to serialize Dialogue:")
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (d *Dialogue) MarshalTo(b []byte) error {
	if len(b) < 4 {
		return io.ErrUnexpectedEOF
	}
	b[0] = uint8(d.Tag)
	b[1] = d.Length
	b[2] = uint8(d.ExternalTag)
	b[3] = d.ExternalLength

	var offset = 4
	if field := d.ObjectIdentifier; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	if field := d.SingleAsn1Type; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	if field := d.DialoguePDU; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	copy(b[offset:], d.Payload)
	return nil
}

// ParseDialogue parses given byte sequence as an Dialogue.
func ParseDialogue(b []byte) (*Dialogue, error) {
	d := &Dialogue{}
	if err := d.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return d, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in an Dialogue.
func (d *Dialogue) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l < 5 {
		return io.ErrUnexpectedEOF
	}

	d.Tag = Tag(b[0])
	d.Length = b[1]
	d.ExternalTag = Tag(b[2])
	d.ExternalLength = b[3]

	var err error
	var offset = 4
	d.ObjectIdentifier, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += d.ObjectIdentifier.MarshalLen()

	d.SingleAsn1Type, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += d.SingleAsn1Type.MarshalLen()

	d.DialoguePDU, err = ParseDialoguePDU(d.SingleAsn1Type.Value)
	if err != nil {
		return err
	}

	d.Payload = b[offset:]

	return nil
}

// SetValsFrom sets the values from IE parsed by ParseBER.
func (d *Dialogue) SetValsFrom(berParsed *IE) error {
	d.Tag = berParsed.Tag
	d.Length = berParsed.Length
	for _, ie := range berParsed.IE {
		var dpdu *IE
		if ie.Tag == 0x28 {
			d.ExternalTag = ie.Tag
			d.ExternalLength = ie.Length
			for _, iex := range ie.IE {
				switch iex.Tag {
				case 0x06:
					d.ObjectIdentifier = iex
				case 0xa0:
					d.SingleAsn1Type = iex
					dpdu = iex.IE[0]
				}
			}
		}

		switch dpdu.Tag.Code() {
		case AARQ, AARE, ABRT:
			d.DialoguePDU = &DialoguePDU{
				Type:   dpdu.Tag,
				Length: dpdu.Length,
			}
		}
		for _, iex := range dpdu.IE {
			switch iex.Tag {
			case 0x80:
				d.DialoguePDU.ProtocolVersion = iex
			case 0xa1:
				d.DialoguePDU.ApplicationContextName = iex
			case 0xa2:
				d.DialoguePDU.Result = iex
			case 0xa3:
				d.DialoguePDU.ResultSourceDiagnostic = iex
			}
		}
	}
	return nil
}

// MarshalLen returns the serial length of Dialogue.
func (d *Dialogue) MarshalLen() int {
	l := 4
	if field := d.ObjectIdentifier; field != nil {
		l += field.MarshalLen()
	}
	if field := d.SingleAsn1Type; field != nil {
		l += field.MarshalLen()
	}
	if field := d.DialoguePDU; field != nil {
		l += field.MarshalLen()
	}
	l += len(d.Payload)

	return l
}

// SetLength sets the length in Length field.
func (d *Dialogue) SetLength() {
	d.Length = uint8(d.MarshalLen() - 2)
	d.ExternalLength = uint8(d.MarshalLen() - 4)
}

// String returns the SCCP common header values in human readable format.
func (d *Dialogue) String() string {
	return fmt.Sprintf("{Tag: %#x, Length: %d, ExternalTag: %x, ExternalLength: %d, ObjectIdentifier: %v, SingleAsn1Type: %v, DialoguePDU: %v, Payload: %x}",
		d.Tag,
		d.Length,
		d.ExternalTag,
		d.ExternalLength,
		d.ObjectIdentifier,
		d.SingleAsn1Type,
		d.DialoguePDU,
		d.Payload,
	)
}

// Version returns Protocol Version in string.
func (d *Dialogue) Version() string {
	if d.DialoguePDU == nil {
		return ""
	}

	return d.DialoguePDU.Version()
}

// Context returns the Context part of ApplicationContextName in string.
func (d *Dialogue) Context() string {
	if d.DialoguePDU == nil {
		return ""
	}

	return d.DialoguePDU.Context()
}

// ContextVersion returns the Version part of ApplicationContextName in string.
func (d *Dialogue) ContextVersion() string {
	if d.DialoguePDU == nil {
		return ""
	}

	return d.DialoguePDU.ContextVersion()
}
