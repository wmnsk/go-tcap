// Copyright 2019-2020 go-tcap authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

/*
Package tcap provides simple and painless handling of TCAP(Transaction Capabilities Application Part) in SS7/SIGTRAN protocol stack.

Though TCAP is ASN.1-based protocol, this implementation does not use any ASN.1 parser.
That makes this implementation flexible enough to create arbitrary payload with any combinations, which is useful for testing.
*/
package tcap

import (
	"encoding/binary"
	"fmt"
)

// TCAP represents a General Structure of TCAP Information Elements.
type TCAP struct {
	Transaction *Transaction
	Dialogue    *Dialogue
	Components  *Components
}

// NewBeginInvoke creates a new NewBeginInvoke.
func NewBeginInvoke(otid uint32, dlgType, ctx, ctxver uint8, invID, opCode int, payload []byte) *TCAP {
	t := &TCAP{
		Transaction: NewBegin(otid, []byte{}),
		Dialogue:    NewDialogue(dlgType, 1, NewAARQ(1, ctx, ctxver), []byte{}),
		Components:  NewComponents(NewInvoke(invID, -1, opCode, true, payload)),
	}
	t.SetLength()

	return t
}

// NewEndReturnResult creates a new NewEndReturnResult.
func NewEndReturnResult(dtid uint32, dlgType, ctx, ctxver uint8, invID, opCode int, isLast bool, payload []byte) *TCAP {
	t := &TCAP{
		Transaction: NewEnd(dtid, []byte{}),
		Dialogue:    NewDialogue(dlgType, 1, NewAARE(1, ctx, ctxver, Accepted, DialogueServiceUser, Null), []byte{}),
		Components:  NewComponents(NewReturnResult(invID, opCode, true, isLast, payload)),
	}
	t.SetLength()

	return t
}

// MarshalBinary returns the byte sequence generated from a TCAP instance.
func (t *TCAP) MarshalBinary() ([]byte, error) {
	b := make([]byte, t.MarshalLen())
	if err := t.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (t *TCAP) MarshalTo(b []byte) error {
	var offset = 0
	if portion := t.Transaction; portion != nil {
		if err := portion.MarshalTo(b[offset : offset+portion.MarshalLen()]); err != nil {
			return err
		}
		offset += portion.MarshalLen()
	}

	if portion := t.Dialogue; portion != nil {
		if err := portion.MarshalTo(b[offset : offset+portion.MarshalLen()]); err != nil {
			return err
		}
		offset += portion.MarshalLen()
	}

	if portion := t.Components; portion != nil {
		if err := portion.MarshalTo(b[offset : offset+portion.MarshalLen()]); err != nil {
			return err
		}
	}

	return nil
}

// Parse parses given byte sequence as an TCAP.
func Parse(b []byte) (*TCAP, error) {
	t := &TCAP{}
	if err := t.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return t, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in an TCAP.
func (t *TCAP) UnmarshalBinary(b []byte) error {
	var err error
	var offset = 0

	t.Transaction, err = ParseTransaction(b[offset:])
	if err != nil {
		return err
	}
	if len(t.Transaction.Payload) == 0 {
		return nil
	}

	switch t.Transaction.Payload[0] {
	case 0x6b:
		t.Dialogue, err = ParseDialogue(t.Transaction.Payload)
		if err != nil {
			return err
		}
		if len(t.Dialogue.Payload) == 0 {
			return nil
		}

		t.Components, err = ParseComponents(t.Dialogue.Payload)
		if err != nil {
			return err
		}
	case 0x6c:
		t.Components, err = ParseComponents(t.Transaction.Payload)
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseBer parses given byte sequence as a TCAP.
func ParseBer(b []byte) ([]*TCAP, error) {
	parsed, err := ParseAsBer(b)
	if err != nil {
		return nil, err
	}

	tcaps := make([]*TCAP, 0)
	for _, tx := range parsed {
		t := &TCAP{
			Transaction: &Transaction{},
			Dialogue:    &Dialogue{},
			Components:  &Components{},
		}

		if err := t.Transaction.SetValsFrom(tx); err != nil {
			return nil, err
		}

		for _, dx := range tx.IE {
			switch dx.Tag {
			case 0x6b:
				if err := t.Dialogue.SetValsFrom(dx); err != nil {
					return nil, err
				}
			case 0x6c:
				if err := t.Components.SetValsFrom(dx); err != nil {
					return nil, err
				}
			}
		}

		tcaps = append(tcaps, t)
	}

	return tcaps, nil
}

// MarshalLen returns the serial length of TCAP.
func (t *TCAP) MarshalLen() int {
	l := 0
	if portion := t.Components; portion != nil {
		l += portion.MarshalLen()
	}
	if portion := t.Dialogue; portion != nil {
		l += portion.MarshalLen()
	}
	if portion := t.Transaction; portion != nil {
		l += portion.MarshalLen()
	}
	return l
}

// SetLength sets the length in Length field.
func (t *TCAP) SetLength() {
	if portion := t.Components; portion != nil {
		portion.SetLength()
	}
	if portion := t.Dialogue; portion != nil {
		portion.SetLength()
	}
	if portion := t.Transaction; portion != nil {
		portion.SetLength()
		if c := t.Components; c != nil {
			portion.Length += uint8(c.MarshalLen())
		}
		if d := t.Dialogue; d != nil {
			portion.Length += uint8(d.MarshalLen())
		}
	}
}

// OTID returns the TCAP Originating Transaction ID in Transaction Portion in uint32.
func (t *TCAP) OTID() uint32 {
	if ts := t.Transaction; ts != nil {
		if otid := ts.OrigTransactionID; otid != nil {
			return binary.BigEndian.Uint32(otid.Value)
		}
	}

	return 0
}

// DTID returns the TCAP Originating Transaction ID in Transaction Portion in uint32.
func (t *TCAP) DTID() uint32 {
	if ts := t.Transaction; ts != nil {
		if dtid := ts.DestTransactionID; dtid != nil {
			return binary.BigEndian.Uint32(dtid.Value)
		}
	}

	return 0
}

// AppContextName returns the ACN in string.
func (t *TCAP) AppContextName() string {
	if d := t.Dialogue; d != nil {
		return d.Context()
	}

	return ""
}

// AppContextNameWithVersion returns the ACN with ACN Version in string.
// XXX - Looking for better way to return the value in the same format...
func (t *TCAP) AppContextNameWithVersion() string {
	if d := t.Dialogue; d != nil {
		return d.Context() + "-v" + d.ContextVersion()
	}

	return ""
}

// AppContextNameOid returns the ACN with ACN Version in OID formatted string.
// XXX - Looking for better way to return the value in the same format...
func (t *TCAP) AppContextNameOid() string {
	if r := t.Dialogue; r != nil {
		if rp := r.DialoguePDU; rp != nil {
			var oid = "0."
			for i, x := range rp.ApplicationContextName.Value[2:] {
				oid += fmt.Sprint(x)
				if i <= 6 {
					break
				}
				oid += "."
			}
			return oid
		}
	}

	return ""
}

// ComponentType returns the ComponentType in Component Portion in the list of string.
//
// The returned value is of type []string, as it may have multiple Components.
func (t *TCAP) ComponentType() []string {
	if c := t.Components; c != nil {
		var iids []string
		for _, cm := range c.Component {
			iids = append(iids, cm.ComponentTypeString())
		}
		return iids
	}

	return nil
}

// InvokeID returns the InvokeID in Component Portion in the list of string.
//
// The returned value is of type []string, as it may have multiple Components.
func (t *TCAP) InvokeID() []uint8 {
	if c := t.Components; c != nil {
		var iids []uint8
		for _, cm := range c.Component {
			iids = append(iids, cm.InvID())
		}

		return iids
	}

	return nil
}

// OpCode returns the OpCode in Component Portion in the list of string.
//
// The returned value is of type []string, as it may have multiple Components.
func (t *TCAP) OpCode() []uint8 {
	if c := t.Components; c != nil {
		var ops []uint8
		for _, cm := range c.Component {
			ops = append(ops, cm.OpCode())
		}

		return ops
	}

	return nil
}

// LayerPayload returns the upper layer as byte slice.
//
// The returned value is of type [][]byte, as it may have multiple Components.
func (t *TCAP) LayerPayload() [][]byte {
	if c := t.Components; c != nil {
		var ret [][]byte
		for _, cm := range c.Component {
			ret = append(ret, cm.Parameter.Value)
		}

		return ret
	}

	return nil
}

// String returns the SCCP common header values in human readable format.
func (t *TCAP) String() string {
	return fmt.Sprintf("{Transaction: %v, Dialogue: %v, Components: %v}",
		t.Transaction,
		t.Dialogue,
		t.Components,
	)
}
