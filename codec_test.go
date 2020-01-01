// Copyright 2019-2020 go-tcap authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package tcap_test

import (
	"encoding"
	"testing"

	"github.com/pascaldekloe/goe/verify"
	"github.com/wmnsk/go-tcap"
)

type serializable interface {
	encoding.BinaryMarshaler
	MarshalLen() int
}

var testcases = []struct {
	description string
	structured  serializable
	serialized  []byte
	parseFunc   func(b []byte) (serializable, error)
}{
	// TCAP (All)
	// TODO: Add more patterns
	{
		description: "TCAP/Begin - AARQ - Invoke",
		structured: tcap.NewBeginInvoke(
			0x11111111,                     // OTID
			tcap.DialogueAsID,              // DialogueType
			tcap.AnyTimeInfoEnquiryContext, // ACN
			3,                              // ACN Version
			0,                              // Invoke Id
			71,                             // OpCode
			[]byte{0xde, 0xad, 0xbe, 0xef}, // Payload
		),
		serialized: []byte{
			// Transaction Portion
			0x62, 0x36, 0x48, 0x04, 0x11, 0x11, 0x11, 0x11,
			// Dialogue Portion
			0x6b, 0x1e, 0x28, 0x1c, 0x06, 0x07, 0x00, 0x11, 0x86, 0x05, 0x01, 0x01, 0x01, 0xa0, 0x11, 0x60,
			0x0f, 0x80, 0x02, 0x07, 0x80, 0xa1, 0x09, 0x06, 0x07, 0x04, 0x00, 0x00, 0x01, 0x00, 0x1d, 0x03,
			// Component Portion
			0x6c, 0x0e, 0xa1, 0x0c, 0x02, 0x01, 0x00, 0x02, 0x01, 0x47, 0x30, 0x04, 0xde, 0xad, 0xbe, 0xef,
		},
		parseFunc: func(b []byte) (serializable, error) {
			v, err := tcap.Parse(b)
			if err != nil {
				return nil, err
			}
			// clear unnecessary payload
			v.Transaction.Payload = nil
			v.Dialogue.SingleAsn1Type.Value = nil
			v.Dialogue.Payload = nil

			return v, nil
		},
	}, {
		description: "TCAP/End - AARE - ReturnResultLast",
		structured: tcap.NewEndReturnResult(
			0x11111111,                     // OTID
			tcap.DialogueAsID,              // DialogueType
			tcap.AnyTimeInfoEnquiryContext, // ACN
			3,                              // ACN Version
			1,                              // Invoke Id
			71,                             // OpCode
			true,                           // Last or not
			[]byte{0xde, 0xad, 0xbe, 0xef}, // Payload
		),
		serialized: []byte{
			// Transaction Portion
			0x64, 0x44, 0x49, 0x04, 0x11, 0x11, 0x11, 0x11,
			// Dialogue Portion
			0x6b, 0x2a, 0x28, 0x28, 0x06, 0x07, 0x00, 0x11, 0x86, 0x05, 0x01, 0x01, 0x01, 0xa0, 0x1d, 0x61,
			0x1b, 0x80, 0x02, 0x07, 0x80, 0xa1, 0x09, 0x06, 0x07, 0x04, 0x00, 0x00, 0x01, 0x00, 0x1d, 0x03,
			0xa2, 0x03, 0x02, 0x01, 0x00, 0xa3, 0x05, 0xa1,
			0x03, 0x02, 0x01, 0x00,
			// Component Portion
			0x6c, 0x10, 0xa2, 0x0e, 0x02, 0x01, 0x01, 0x30, 0x09, 0x02, 0x01, 0x47, 0x30, 0x04, 0xde, 0xad,
			0xbe, 0xef,
		},
		parseFunc: func(b []byte) (serializable, error) {
			v, err := tcap.Parse(b)
			if err != nil {
				return nil, err
			}
			// clear unnecessary payload
			v.Transaction.Payload = nil
			v.Dialogue.SingleAsn1Type.Value = nil
			v.Dialogue.Payload = nil
			v.Components.Component[0].ResultRetres.Value = nil

			return v, nil
		},
	},
	// Transaction Portion
	{
		description: "Transaction/Unidirectional",
		structured: tcap.NewUnidirectional(
			[]byte{0xca, 0xfe},
		),
		serialized: []byte{0x61, 0x02, 0xca, 0xfe},
		parseFunc:  func(b []byte) (serializable, error) { return tcap.ParseTransaction(b) },
	}, {
		description: "Transaction/Begin",
		structured:  tcap.NewBegin(0xdeadbeef, []byte{0xfa, 0xce}),
		serialized:  []byte{0x62, 0x08, 0x48, 0x04, 0xde, 0xad, 0xbe, 0xef, 0xfa, 0xce},
		parseFunc:   func(b []byte) (serializable, error) { return tcap.ParseTransaction(b) },
	}, {
		description: "Transaction/End",
		structured:  tcap.NewEnd(0xdeadbeef, []byte{0xfa, 0xce}),
		serialized:  []byte{0x64, 0x08, 0x49, 0x04, 0xde, 0xad, 0xbe, 0xef, 0xfa, 0xce},
		parseFunc:   func(b []byte) (serializable, error) { return tcap.ParseTransaction(b) },
	}, {
		description: "Transaction/Continue",
		structured:  tcap.NewContinue(0xdeadbeef, 0xdeadbeef, []byte{0xfa, 0xce}),
		serialized: []byte{
			0x65, 0x0e, 0x48, 0x04, 0xde, 0xad, 0xbe, 0xef, 0x49, 0x04, 0xde, 0xad, 0xbe, 0xef, 0xfa, 0xce,
		},
		parseFunc: func(b []byte) (serializable, error) { return tcap.ParseTransaction(b) },
	}, {
		description: "Transaction/Abort",
		structured:  tcap.NewAbort(0xdeadbeef, tcap.UnrecognizedMessageType, []byte{0xfa, 0xce}),
		serialized: []byte{
			0x67, 0x0b, 0x49, 0x04, 0xde, 0xad, 0xbe, 0xef, 0x4a, 0x01, 0x00, 0xfa, 0xce,
		},
		parseFunc: func(b []byte) (serializable, error) { return tcap.ParseTransaction(b) },
	},
	// Dialogue Portion
	{
		description: "Dialogue/AARQ",
		structured: tcap.NewDialogue(
			1, 1, // OID, Version
			tcap.NewAARQ(
				// Version, Context, ContextVersion
				1, tcap.AnyTimeInfoEnquiryContext, 3,
			),
			[]byte{0xde, 0xad, 0xbe, 0xef},
		),
		serialized: []byte{
			0x6b, 0x22, 0x28, 0x20, 0x06, 0x07, 0x00, 0x11, 0x86, 0x05, 0x01, 0x01, 0x01, 0xa0, 0x11, 0x60,
			0x0f, 0x80, 0x02, 0x07, 0x80, 0xa1, 0x09, 0x06, 0x07, 0x04, 0x00, 0x00, 0x01, 0x00, 0x1d, 0x03,
			0xde, 0xad, 0xbe, 0xef,
		},
		parseFunc: func(b []byte) (serializable, error) {
			v, err := tcap.ParseDialogue(b)
			if err != nil {
				return nil, err
			}
			// clear unnecessary payload
			v.SingleAsn1Type.Value = nil

			return v, nil
		},
	}, {
		description: "Dialogue/AARE",
		structured: tcap.NewDialogue(
			1, 1, // OID, Version
			tcap.NewAARE(
				// Version, Context, ContextVersion
				1, tcap.AnyTimeInfoEnquiryContext, 3,
				// Result, ResultSourceDiag, Reason
				0, 1, 0,
			),
			[]byte{0xde, 0xad, 0xbe, 0xef},
		),
		serialized: []byte{
			0x6b, 0x2e, 0x28, 0x2c, 0x06, 0x07, 0x00, 0x11, 0x86, 0x05, 0x01, 0x01, 0x01, 0xa0, 0x1d, 0x61,
			0x1b, 0x80, 0x02, 0x07, 0x80, 0xa1, 0x09, 0x06, 0x07, 0x04, 0x00, 0x00, 0x01, 0x00, 0x1d, 0x03,
			0xa2, 0x03, 0x02, 0x01, 0x00, 0xa3, 0x05, 0xa1, 0x03, 0x02, 0x01, 0x00, 0xde, 0xad, 0xbe, 0xef,
		},
		parseFunc: func(b []byte) (serializable, error) {
			v, err := tcap.ParseDialogue(b)
			if err != nil {
				return nil, err
			}
			// clear unnecessary payload
			v.SingleAsn1Type.Value = nil

			return v, nil
		},
	},
	// Component Portion
	{
		description: "Components/invoke",
		structured:  tcap.NewComponents(tcap.NewInvoke(0, 0, 71, true, []byte{0xde, 0xad, 0xbe, 0xef})),
		serialized: []byte{
			0x6c, 0x0e, 0xa1, 0x0c, 0x02, 0x01, 0x00, 0x02, 0x01, 0x47, 0x30, 0x04, 0xde, 0xad, 0xbe, 0xef,
		},
		parseFunc: func(b []byte) (serializable, error) { return tcap.ParseComponents(b) },
	}, {
		description: "Components/returnResultLast",
		structured:  tcap.NewComponents(tcap.NewReturnResult(0, 71, true, true, []byte{0xde, 0xad, 0xbe, 0xef})),
		serialized: []byte{
			0x6c, 0x10, 0xa2, 0x0e, 0x02, 0x01, 0x00, 0x30, 0x09, 0x02, 0x01, 0x47, 0x30, 0x04, 0xde, 0xad,
			0xbe, 0xef,
		},
		parseFunc: func(b []byte) (serializable, error) {
			v, err := tcap.ParseComponents(b)
			if err != nil {
				return nil, err
			}
			// clear unnecessary payload
			v.Component[0].ResultRetres.Value = nil

			return v, nil
		},
	}, {
		description: "Components/returnError",
		structured:  tcap.NewComponents(tcap.NewReturnError(0, 71, true, []byte{0xde, 0xad, 0xbe, 0xef})),
		serialized: []byte{
			0x6c, 0x0e, 0xa3, 0x0c, 0x02, 0x01, 0x00, 0x02, 0x01, 0x47, 0x30, 0x04, 0xde, 0xad, 0xbe, 0xef,
		},
		parseFunc: func(b []byte) (serializable, error) { return tcap.ParseComponents(b) },
	},
	// Generic IE
	{
		description: "IE/Single",
		structured:  tcap.NewIE(tcap.NewTag(01, 0, 0x08), []byte{0xde, 0xad, 0xbe, 0xef}),
		serialized:  []byte{0x48, 0x04, 0xde, 0xad, 0xbe, 0xef},
		parseFunc:   func(b []byte) (serializable, error) { return tcap.ParseIE(b) },
	},
}

func TestCodec(t *testing.T) {
	t.Helper()

	for _, c := range testcases {
		t.Run("Parse", func(t *testing.T) {
			msg, err := c.parseFunc(c.serialized)
			if err != nil {
				t.Fatal(err)
			}

			if got, want := msg, c.structured; !verify.Values(t, "", got, want) {
				t.Fail()
			}
		})

		t.Run("Marshal", func(t *testing.T) {
			b, err := c.structured.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			if got, want := b, c.serialized; !verify.Values(t, "", got, want) {
				t.Fail()
			}
		})

		t.Run("Len", func(t *testing.T) {
			if got, want := c.structured.MarshalLen(), len(c.serialized); got != want {
				t.Fatalf("got %v want %v", got, want)
			}
		})
	}
}
