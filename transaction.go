// Copyright 2019-2020 go-tcap authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package tcap

import (
	"encoding/binary"
	"fmt"
)

// Message Type definitions.
const (
	Unidirectional int = iota + 1
	Begin
	_
	End
	Continue
	_
	Abort
)

// Abort Cause definitions.
const (
	UnrecognizedMessageType uint8 = iota
	UnrecognizedTransactionID
	BadlyFormattedTransactionPortion
	IncorrectTransactionPortion
	ResourceLimitation
)

// Transaction represents a Transaction Portion of TCAP.
type Transaction struct {
	Type              Tag
	Length            uint8
	OrigTransactionID *IE
	DestTransactionID *IE
	PAbortCause       *IE
	Payload           []byte
}

// NewTransaction returns a new Transaction Portion.
func NewTransaction(mtype, otid, dtid int, cause uint8, payload []byte) *Transaction {
	t := &Transaction{
		Type: NewApplicationWideConstructorTag(mtype),
		OrigTransactionID: &IE{
			Tag:   NewApplicationWidePrimitiveTag(8),
			Value: make([]byte, 4),
		},
		DestTransactionID: &IE{
			Tag:   NewApplicationWidePrimitiveTag(9),
			Value: make([]byte, 4),
		},
		PAbortCause: &IE{
			Tag:   NewApplicationWidePrimitiveTag(10),
			Value: []byte{cause},
		},
		Payload: payload,
	}
	binary.BigEndian.PutUint32(t.OrigTransactionID.Value, uint32(otid))
	binary.BigEndian.PutUint32(t.DestTransactionID.Value, uint32(dtid))
	t.SetLength()

	return t
}

// NewUnidirectional returns Unidirectional type of Transacion Portion.
func NewUnidirectional(payload []byte) *Transaction {
	t := NewTransaction(
		Unidirectional, // Type: Unidirectional
		0,              // otid
		0,              // dtid
		0,              // cause
		payload,        // payload
	)
	t.OrigTransactionID = nil
	t.DestTransactionID = nil
	t.PAbortCause = nil
	return t
}

// NewBegin returns Begin type of Transacion Portion.
func NewBegin(otid uint32, payload []byte) *Transaction {
	t := &Transaction{
		Type: NewApplicationWideConstructorTag(Begin),
		OrigTransactionID: &IE{
			Tag:   NewApplicationWidePrimitiveTag(8),
			Value: make([]byte, 4),
		},
		Payload: payload,
	}
	binary.BigEndian.PutUint32(t.OrigTransactionID.Value, otid)
	t.SetLength()

	return t
}

// NewEnd returns End type of Transacion Portion.
func NewEnd(otid uint32, payload []byte) *Transaction {
	t := &Transaction{
		Type: NewApplicationWideConstructorTag(End),
		DestTransactionID: &IE{
			Tag:   NewApplicationWidePrimitiveTag(9),
			Value: make([]byte, 4),
		},
		Payload: payload,
	}
	binary.BigEndian.PutUint32(t.DestTransactionID.Value, otid)
	t.SetLength()

	return t
}

// NewContinue returns Continue type of Transacion Portion.
func NewContinue(otid, dtid int, payload []byte) *Transaction {
	t := NewTransaction(
		Continue, // Type: Continue
		otid,     // otid
		dtid,     // dtid
		0,        // cause
		payload,  // payload
	)
	t.PAbortCause = nil
	return t
}

// NewAbort returns Abort type of Transacion Portion.
func NewAbort(dtid int, cause uint8, payload []byte) *Transaction {
	t := NewTransaction(
		Abort,   // Type: Abort
		0,       // otid
		dtid,    // dtid
		cause,   // cause
		payload, // payload
	)
	t.OrigTransactionID = nil
	return t
}

// MarshalBinary returns the byte sequence generated from a Transaction instance.
func (t *Transaction) MarshalBinary() ([]byte, error) {
	b := make([]byte, t.MarshalLen())
	if err := t.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (t *Transaction) MarshalTo(b []byte) error {
	b[0] = uint8(t.Type)
	b[1] = t.Length

	var offset = 2
	switch t.Type.Code() {
	case Unidirectional:
		break
	case Begin:
		if field := t.OrigTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	case End:
		if field := t.DestTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	case Continue:
		if field := t.OrigTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}

		if field := t.DestTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	case Abort:
		if field := t.DestTransactionID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}

		if field := t.PAbortCause; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	}
	copy(b[offset:t.MarshalLen()], t.Payload)
	return nil
}

// ParseTransaction parses given byte sequence as an Transaction.
func ParseTransaction(b []byte) (*Transaction, error) {
	t := &Transaction{}
	if err := t.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return t, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in an Transaction.
func (t *Transaction) UnmarshalBinary(b []byte) error {
	t.Type = Tag(b[0])
	t.Length = b[1]

	var err error
	var offset = 2
	switch t.Type.Code() {
	case Unidirectional:
		break
	case Begin:
		t.OrigTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.OrigTransactionID.MarshalLen()
	case End:
		t.DestTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.DestTransactionID.MarshalLen()
	case Continue:
		t.OrigTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.OrigTransactionID.MarshalLen()
		t.DestTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.DestTransactionID.MarshalLen()
	case Abort:
		t.DestTransactionID, err = ParseIE(b[offset : offset+6])
		if err != nil {
			return err
		}
		offset += t.DestTransactionID.MarshalLen()
		t.PAbortCause, err = ParseIE(b[offset : offset+3])
		if err != nil {
			return err
		}
		offset += t.PAbortCause.MarshalLen()
	}
	t.Payload = b[offset:]
	return nil
}

// SetValsFrom sets the values from IE parsed by ParseBer
func (t *Transaction) SetValsFrom(berParsed *IE) error {
	t.Type = berParsed.Tag
	t.Length = berParsed.Length
	for _, ie := range berParsed.IE {
		switch ie.Tag {
		case 0x48:
			t.OrigTransactionID = ie
		case 0x49:
			t.DestTransactionID = ie
		case 0x4a:
			t.PAbortCause = ie
		}
	}
	return nil
}

// MarshalLen returns the serial length of Transaction.
func (t *Transaction) MarshalLen() int {
	l := 2
	switch t.Type.Code() {
	case Unidirectional:
		break
	case Begin:
		if field := t.OrigTransactionID; field != nil {
			l += field.MarshalLen()
		}
	case End:
		if field := t.DestTransactionID; field != nil {
			l += field.MarshalLen()
		}
	case Continue:
		if field := t.OrigTransactionID; field != nil {
			l += field.MarshalLen()
		}
		if field := t.DestTransactionID; field != nil {
			l += field.MarshalLen()
		}
	case Abort:
		if field := t.DestTransactionID; field != nil {
			l += field.MarshalLen()
		}
		if field := t.PAbortCause; field != nil {
			l += field.MarshalLen()
		}
	}
	return l + len(t.Payload)
}

// SetLength sets the length in Length field.
func (t *Transaction) SetLength() {
	if field := t.OrigTransactionID; field != nil {
		field.SetLength()
	}
	if field := t.DestTransactionID; field != nil {
		field.SetLength()
	}
	if field := t.PAbortCause; field != nil {
		field.SetLength()
	}
	t.Length = uint8(t.MarshalLen() - 2)
}

// MessageTypeString returns the name of Message Type in string.
func (t *Transaction) MessageTypeString() string {
	switch t.Type.Code() {
	case Unidirectional:
		return "Unidirectional"
	case Begin:
		return "Begin"
	case End:
		return "End"
	case Continue:
		return "Continue"
	case Abort:
		return "Abort"
	}
	return ""
}

// OTID returns the OrigTransactionID in string.
func (t *Transaction) OTID() string {
	switch t.Type.Code() {
	case Begin, Continue:
		if field := t.OrigTransactionID; field != nil {
			return fmt.Sprintf("%04x", field.Value)
		}
	}
	return ""
}

// DTID returns the DestTransactionID in string.
func (t *Transaction) DTID() string {
	switch t.Type.Code() {
	case End, Continue, Abort:
		if field := t.DestTransactionID; field != nil {
			return fmt.Sprintf("%04x", field.Value)
		}
	}
	return ""
}

// AbortCause returns the P-Abort Cause in string.
func (t *Transaction) AbortCause() string {
	cause := t.PAbortCause
	if cause == nil {
		return ""
	}

	if t.Type.Code() == Abort {
		switch t.PAbortCause.Value[0] {
		case UnrecognizedMessageType:
			return "UnrecognizedMessageType"
		case UnrecognizedTransactionID:
			return "UnrecognizedTransactionID"
		case BadlyFormattedTransactionPortion:
			return "BadlyFormattedTransactionPortion"
		case IncorrectTransactionPortion:
			return "IncorrectTransactionPortion"
		case ResourceLimitation:
			return "ResourceLimitation"
		}
	}
	return ""
}

// String returns the SCCP common header values in human readable format.
func (t *Transaction) String() string {
	return fmt.Sprintf("{Type: %#x, Length: %d, OrigTransactionID: %v, DestTransactionID: %v, PAbortCause: %v, Payload: %x}",
		t.Type,
		t.Length,
		t.OrigTransactionID,
		t.DestTransactionID,
		t.PAbortCause,
		t.Payload,
	)
}
