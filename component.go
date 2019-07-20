// Copyright 2018 shingo authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package tcap

import (
	"fmt"
	"io"
)

// Component Type definitions.
const (
	Invoke int = iota + 1
	ReturnResultLast
	ReturnError
	Reject
	_
	_
	ReturnResultNotLast
)

// Problem Type definitions.
const (
	GeneralProblem int = iota
	InvokeProblem
	ReturnResultProblem
	ReturnErrorProblem
)

// General Problem Code definitions.
const (
	UnrecognizedComponent uint8 = iota
	MistypedComponent
	BadlyStructuredComponent
)

// Invoke Problem Code definitions.
const (
	InvokeProblemDuplicateInvokeID uint8 = iota
	InvokeProblemUnrecognizedOperation
	InvokeProblemMistypedParameter
	InvokeProblemResourceLimitation
	InvokeProblemInitiatingRelease
	InvokeProblemUnrecognizedLinkedID
	InvokeProblemLinkedResponseUnexpected
	InvokeProblemUnexpectedLinkedOperation
)

// ReturnResult Problem Code definitions.
const (
	ResultProblemUnrecognizedInvokeID uint8 = iota
	ResultProblemReturnResultUnexpected
	ResultProblemMistypedParameter
)

// ReturnError Problem Code definitions.
const (
	ErrorProblemUnrecognizedInvokeID uint8 = iota
	ErrorProblemReturnErrorUnexpected
	ErrorProblemUnrecognizedError
	ErrorProblemUnexpectedError
	ErrorProblemMistypedParameter
)

// Components represents a TCAP Components(Header).
//
// This is a TCAP Components' Header part. Contents are in Component field.
type Components struct {
	Tag       Tag
	Length    uint8
	Component []*Component
}

// Component represents a TCAP Component.
type Component struct {
	Type          Tag
	Length        uint8
	InvokeID      *IE
	LinkedID      *IE
	ResultRetres  *IE
	SequenceTag   *IE
	OperationCode *IE
	ErrorCode     *IE
	ProblemCode   *IE
	Parameter     *IE
}

// NewComponents creates a new Components.
func NewComponents(comps ...*Component) *Components {
	c := &Components{
		Tag:       NewApplicationWideConstructorTag(12),
		Component: comps,
	}
	c.SetLength()

	return c
}

// NewInvoke returns a new single Invoke Component.
func NewInvoke(invID, lkID, opCode int, isLocal bool, param []byte) *Component {
	c := &Component{
		Type: NewContextSpecificConstructorTag(Invoke),
		InvokeID: &IE{
			Tag:    NewUniversalPrimitiveTag(2),
			Length: 1,
			Value:  []byte{uint8(invID)},
		},
		OperationCode: NewOperationCode(opCode, isLocal),
		Parameter: &IE{
			Tag:   NewUniversalConstructorTag(0x10),
			Value: param,
		},
	}
	if lkID > 0 {
		c.LinkedID = &IE{
			Tag:    NewContextSpecificPrimitiveTag(0),
			Length: 1,
			Value:  []byte{uint8(lkID)},
		}
	}

	c.SetLength()
	return c
}

// NewReturnResult returns a new single ReturnResultLast or ReturnResultNotLast Component.
func NewReturnResult(invID, opCode int, isLocal, isLast bool, param []byte) *Component {
	tag := ReturnResultNotLast
	if isLast {
		tag = ReturnResultLast
	}
	c := &Component{
		Type: NewContextSpecificConstructorTag(tag),
		ResultRetres: &IE{
			Tag: NewUniversalConstructorTag(0x10),
		},
		InvokeID: &IE{
			Tag:    NewUniversalPrimitiveTag(2),
			Length: 1,
			Value:  []byte{uint8(invID)},
		},
		OperationCode: NewOperationCode(opCode, isLocal),
		Parameter: &IE{
			Tag:   NewUniversalConstructorTag(0x10),
			Value: param,
		},
	}

	c.SetLength()
	return c
}

// NewReturnError returns a new single ReturnError Component.
func NewReturnError(invID, errCode int, isLocal bool, param []byte) *Component {
	c := &Component{
		Type: NewContextSpecificConstructorTag(ReturnError),
		InvokeID: &IE{
			Tag:    NewUniversalPrimitiveTag(2),
			Length: 1,
			Value:  []byte{uint8(invID)},
		},
		ErrorCode: NewErrorCode(errCode, isLocal),
		Parameter: &IE{
			Tag:   NewUniversalConstructorTag(0x10),
			Value: param,
		},
	}

	c.SetLength()
	return c
}

// NewReject returns a new single Reject Component.
func NewReject(invID, problemType int, problemCode uint8, param []byte) *Component {
	c := &Component{
		Type: NewContextSpecificConstructorTag(Invoke),
		InvokeID: &IE{
			Tag:    NewUniversalPrimitiveTag(2),
			Length: 1,
			Value:  []byte{uint8(invID)},
		},
		ProblemCode: &IE{
			Tag:    NewContextSpecificPrimitiveTag(problemType),
			Length: 1,
			Value:  []byte{problemCode},
		},
	}

	c.SetLength()
	return c
}

// NewOperationCode returns a Operation Code.
func NewOperationCode(code int, isLocal bool) *IE {
	var tag = 4
	if isLocal {
		tag = 2
	}
	return &IE{
		Tag:    NewUniversalPrimitiveTag(tag),
		Length: 1,
		Value:  []byte{uint8(code)},
	}
}

// NewErrorCode returns a Error Code.
func NewErrorCode(code int, isLocal bool) *IE {
	return NewOperationCode(code, isLocal)
}

// MarshalBinary returns the byte sequence generated from a Components instance.
func (c *Components) MarshalBinary() ([]byte, error) {
	b := make([]byte, c.MarshalLen())
	if err := c.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (c *Components) MarshalTo(b []byte) error {
	b[0] = uint8(c.Tag)
	b[1] = c.Length

	var offset = 2
	for _, comp := range c.Component {
		if err := comp.MarshalTo(b[offset : offset+comp.MarshalLen()]); err != nil {
			return err
		}
		offset += comp.MarshalLen()
	}
	return nil
}

// MarshalBinary returns the byte sequence generated from a Components instance.
func (c *Component) MarshalBinary() ([]byte, error) {
	b := make([]byte, c.MarshalLen())
	if err := c.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (c *Component) MarshalTo(b []byte) error {
	b[0] = uint8(c.Type)
	b[1] = c.Length

	var offset = 2
	if err := c.InvokeID.MarshalTo(b[offset : offset+c.InvokeID.MarshalLen()]); err != nil {
		return err
	}
	offset += c.InvokeID.MarshalLen()

	switch c.Type.Code() {
	case Invoke:
		if field := c.LinkedID; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}

		if field := c.OperationCode; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}

		if field := c.Parameter; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	case ReturnResultLast, ReturnResultNotLast:
		if field := c.ResultRetres; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}

		if field := c.OperationCode; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}

		if field := c.Parameter; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	case ReturnError:
		if field := c.ErrorCode; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}

		if field := c.Parameter; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	case Reject:
		if field := c.ProblemCode; field != nil {
			if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
				return err
			}
			offset += field.MarshalLen()
		}
	}
	return nil
}

// ParseComponents parses given byte sequence as an Components.
func ParseComponents(b []byte) (*Components, error) {
	c := &Components{}
	if err := c.UnmarshalBinary(b); err != nil {
		if err == io.ErrUnexpectedEOF {
			return c, nil
		}
		return nil, err
	}
	return c, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in an Components.
func (c *Components) UnmarshalBinary(b []byte) error {
	if len(b) < 2 {
		return io.ErrUnexpectedEOF
	}

	c.Tag = Tag(b[0])
	c.Length = b[1]

	var offset = 2
	for {
		if len(b) < 2 {
			break
		}

		comp, err := ParseComponent(b[offset:])
		if err != nil {
			return err
		}
		c.Component = append(c.Component, comp)

		if len(b[offset:]) == int(comp.Length)+2 {
			break
		}
		b = b[offset+comp.MarshalLen()-2:]
	}
	return nil
}

// ParseComponent parses given byte sequence as an Component.
func ParseComponent(b []byte) (*Component, error) {
	c := &Component{}
	if err := c.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return c, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in an Component.
func (c *Component) UnmarshalBinary(b []byte) error {
	if len(b) < 2 {
		return io.ErrUnexpectedEOF
	}
	c.Type = Tag(b[0])
	c.Length = b[1]

	var err error
	var offset = 2
	c.InvokeID, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += c.InvokeID.MarshalLen()

	switch c.Type.Code() {
	case Invoke:
		/* TODO: Implement LinkedID Parser.
		c.LinkedID, err = ParseIE(b[offset:])
		if err != nil {
			return err
		}
		offset += c.LinkedID.MarshalLen()
		*/
		c.OperationCode, err = ParseIE(b[offset:])
		if err != nil {
			return err
		}
		offset += c.OperationCode.MarshalLen()

		c.Parameter, err = ParseIE(b[offset:])
		if err != nil {
			return err
		}
		offset += c.Parameter.MarshalLen()
	case ReturnResultLast, ReturnResultNotLast:
		c.ResultRetres, err = ParseIE(b[offset:])
		if err != nil {
			return err
		}
		offset = 0
		b = c.ResultRetres.Value[offset:]

		c.OperationCode, err = ParseIE(b[offset:])
		if err != nil {
			return err
		}
		offset += c.OperationCode.MarshalLen()

		c.Parameter, err = ParseIE(b[offset:])
		if err != nil {
			return err
		}
		offset += c.Parameter.MarshalLen()
	case ReturnError:
		c.ErrorCode, err = ParseIE(b[offset:])
		if err != nil {
			return err
		}
		offset += c.ErrorCode.MarshalLen()

		c.Parameter, err = ParseIE(b[offset:])
		if err != nil {
			return err
		}
		offset += c.Parameter.MarshalLen()
	case Reject:
		c.ProblemCode, err = ParseIE(b[offset:])
		if err != nil {
			return err
		}
		offset += c.ProblemCode.MarshalLen()
	}
	return nil
}

// SetValsFrom sets the values from IE parsed by ParseBer
func (c *Components) SetValsFrom(berParsed *IE) error {
	c.Tag = berParsed.Tag
	c.Length = berParsed.Length
	for _, ie := range berParsed.IE {
		comp := &Component{
			Type:   ie.Tag,
			Length: ie.Length,
		}
		switch ie.Tag {
		case 0xa1: // Invoke
			for i, iex := range ie.IE {
				switch iex.Tag {
				case 0x02:
					if i == 0 {
						comp.InvokeID = iex
					} else {
						comp.OperationCode = iex
					}
				case 0x30:
					comp.Parameter = iex
				}
			}
		case 0xa2, 0xa7: // ReturnResult(Not)Last
			for i, iex := range ie.IE {
				switch iex.Tag {
				case 0x02:
					if i == 0 {
						comp.InvokeID = iex
					}
				case 0x30:
					comp.ResultRetres = iex
					for _, riex := range iex.IE {
						switch riex.Tag {
						case 0x02:
							comp.OperationCode = riex
						case 0x30:
							comp.Parameter = riex
						}
					}
				}
			}
		case 0xa3: // ReturnError
			for i, iex := range ie.IE {
				switch iex.Tag {
				case 0x02:
					if i == 0 {
						comp.InvokeID = iex
					} else {
						comp.ErrorCode = iex
					}
				case 0x30:
					comp.Parameter = iex
				}
			}
		}

		c.Component = append(c.Component, comp)
	}

	return nil
}

// MarshalLen returns the serial length of Components.
func (c *Components) MarshalLen() int {
	var l = 2
	for _, comp := range c.Component {
		l += comp.MarshalLen()
	}
	return l
}

// MarshalLen returns the serial length of Component.
func (c *Component) MarshalLen() int {
	var l = 2 + c.InvokeID.MarshalLen()
	switch c.Type.Code() {
	case Invoke:
		if field := c.LinkedID; field != nil {
			l += field.MarshalLen()
		}
		if field := c.OperationCode; field != nil {
			l += field.MarshalLen()
		}
		if field := c.Parameter; field != nil {
			l += field.MarshalLen()
		}
	case ReturnResultLast, ReturnResultNotLast:
		if field := c.ResultRetres; field != nil {
			l += field.MarshalLen()
		}
		if field := c.OperationCode; field != nil {
			l += field.MarshalLen()
		}
		if field := c.Parameter; field != nil {
			l += field.MarshalLen()
		}
	case ReturnError:
		if field := c.ErrorCode; field != nil {
			l += field.MarshalLen()
		}
		if field := c.Parameter; field != nil {
			l += field.MarshalLen()
		}
	case Reject:
		if field := c.ProblemCode; field != nil {
			l += field.MarshalLen()
		}
	}
	return l
}

// SetLength sets the length in Length field.
func (c *Components) SetLength() {
	c.Length = 0
	for _, comp := range c.Component {
		comp.SetLength()
		c.Length += uint8(comp.MarshalLen())
	}
}

// SetLength sets the length in Length field.
func (c *Component) SetLength() {
	l := 0
	if field := c.InvokeID; field != nil {
		field.SetLength()
	}
	if field := c.LinkedID; field != nil {
		field.SetLength()
	}
	if field := c.OperationCode; field != nil {
		field.SetLength()
		l += c.OperationCode.MarshalLen()
	}
	if field := c.ErrorCode; field != nil {
		field.SetLength()
		l += c.ErrorCode.MarshalLen()
	}
	if field := c.Parameter; field != nil {
		field.SetLength()
		l += c.Parameter.MarshalLen()
	}
	if field := c.ProblemCode; field != nil {
		field.SetLength()
		l += c.ProblemCode.MarshalLen()
	}
	if field := c.SequenceTag; field != nil {
		field.SetLength()
		l += c.SequenceTag.MarshalLen()
	}
	if field := c.ResultRetres; field != nil {
		field.Length = uint8(l)
	}
	c.Length = uint8(c.MarshalLen() - 2)
}

// ComponentTypeString returns the Component Type in string.
func (c *Component) ComponentTypeString() string {
	switch c.Type.Code() {
	case Invoke:
		return "invoke"
	case ReturnResultLast:
		return "returnResultLast"
	case ReturnError:
		return "returnError"
	case Reject:
		return "reject"
	case ReturnResultNotLast:
		return "returnResultNotLast"
	}
	return ""
}

// InvID returns the InvID in string.
func (c *Component) InvID() uint8 {
	if c.InvokeID != nil {
		return c.InvokeID.Value[0]
	}
	return 0
}

// OpCode returns the OpCode in string.
func (c *Component) OpCode() uint8 {
	if c.Type.Code() == ReturnError {
		return c.ErrorCode.Value[0]
	} else if c.Type.Code() != Reject {
		return c.OperationCode.Value[0]
	}
	return 0
}

// String returns the Components values in human readable format.
func (c *Components) String() string {
	return fmt.Sprintf("{Tag: %#x, Length: %d, Component: %v}",
		c.Tag,
		c.Length,
		c.Component,
	)
}

// String returns the Component values in human readable format.
func (c *Component) String() string {
	return fmt.Sprintf("{Type: %#x, Length: %d, ResultRetres: %v, InvokeID: %v, LinkedID: %v, OperationCode: %v, ErrorCode: %v, ProblemCode: %v, Parameter: %v}",
		c.Type,
		c.Length,
		c.ResultRetres,
		c.InvokeID,
		c.LinkedID,
		c.OperationCode,
		c.ErrorCode,
		c.ProblemCode,
		c.Parameter,
	)
}
