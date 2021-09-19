// Copyright 2019-2020 go-tcap authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package tcap

import (
	"fmt"
	"io"
)

// Tag is a Tag in TCAP IE
type Tag uint8

// Class definitions.
const (
	Universal int = iota
	ApplicationWide
	ContextSpecific
	Private
)

// Type definitions.
const (
	Primitive int = iota
	Constructor
)

// NewTag creates a new Tag.
func NewTag(cls, form, code int) Tag {
	return Tag((cls << 6) | (form << 5) | code)
}

// NewUniversalPrimitiveTag creates a new NewUniversalPrimitiveTag.
func NewUniversalPrimitiveTag(code int) Tag {
	return NewTag(Universal, Primitive, code)
}

// NewUniversalConstructorTag creates a new NewUniversalConstructorTag.
func NewUniversalConstructorTag(code int) Tag {
	return NewTag(Universal, Constructor, code)
}

// NewApplicationWidePrimitiveTag creates a new NewApplicationWidePrimitiveTag.
func NewApplicationWidePrimitiveTag(code int) Tag {
	return NewTag(ApplicationWide, Primitive, code)
}

// NewApplicationWideConstructorTag creates a new NewApplicationWideConstructorTag.
func NewApplicationWideConstructorTag(code int) Tag {
	return NewTag(ApplicationWide, Constructor, code)
}

// NewContextSpecificPrimitiveTag creates a new NewContextSpecificPrimitiveTag.
func NewContextSpecificPrimitiveTag(code int) Tag {
	return NewTag(ContextSpecific, Primitive, code)
}

// NewContextSpecificConstructorTag creates a new NewContextSpecificConstructorTag.
func NewContextSpecificConstructorTag(code int) Tag {
	return NewTag(ContextSpecific, Constructor, code)
}

// NewPrivatePrimitiveTag creates a new NewPrivatePrimitiveTag.
func NewPrivatePrimitiveTag(code int) Tag {
	return NewTag(Private, Primitive, code)
}

// NewPrivateConstructorTag creates a new NewPrivateConstructorTag.
func NewPrivateConstructorTag(code int) Tag {
	return NewTag(Private, Constructor, code)
}

// Class returns the Class retieved from a Tag.
func (t Tag) Class() int {
	return int(t) >> 6 & 0x3
}

// Form returns the Form retieved from a Tag.
func (t Tag) Form() int {
	return int(t) >> 5 & 0x1
}

// Code returns the Code retieved from a Tag.
func (t Tag) Code() int {
	return int(t) & 0x1f
}

// IE is a General Structure of TCAP Information Elements.
type IE struct {
	Tag
	Length uint8
	Value  []byte
	IE     []*IE
}

// NewIE creates a new IE.
func NewIE(tag Tag, value []byte) *IE {
	i := &IE{
		Tag:   tag,
		Value: value,
	}
	i.SetLength()

	return i
}

// MarshalBinary returns the byte sequence generated from a IE instance.
func (i *IE) MarshalBinary() ([]byte, error) {
	b := make([]byte, i.MarshalLen())
	if err := i.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (i *IE) MarshalTo(b []byte) error {
	if len(b) < 2 {
		return io.ErrUnexpectedEOF
	}

	b[0] = uint8(i.Tag)
	var offset = writeLength( b , i.Length)
	if(i.Value != nil){
		copy(b[offset:i.MarshalLen()], i.Value)
	}
	return nil
}

// ParseMultiIEs parses multiple (unspecified number of) IEs to []*IE at a time.
func ParseMultiIEs(b []byte) ([]*IE, error) {
	var ies []*IE
	for {
		if len(b) == 0 {
			break
		}

		i, err := ParseIE(b)
		if err != nil {
			return nil, err
		}
		ies = append(ies, i)
		b = b[i.MarshalLen():]
		continue
	}
	return ies, nil
}

// ParseIE parses given byte sequence as an IE.
func ParseIE(b []byte) (*IE, error) {
	i := &IE{}
	if err := i.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return i, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in an IE.
func (i *IE) UnmarshalBinary(b []byte) error {
	l := len(b)
	var offset int = 2
	if l < 3 {
		return io.ErrUnexpectedEOF
	}

	i.Tag = Tag(b[0])
	i.Length, offset = readLength(b)
	if l < offset+int(i.Length) {
		return io.ErrUnexpectedEOF
	}
	i.Value = b[offset : offset+int(i.Length)]
	return nil
}

// ParseAsBer parses given byte sequence as multiple IEs.
//
// Deprecated: use ParseAsBER instead.
func ParseAsBer(b []byte) ([]*IE, error) {
	return ParseAsBER(b)
}

// ParseAsBER parses given byte sequence as multiple IEs.
func ParseAsBER(b []byte) ([]*IE, error) {
	var ies []*IE
	for {
		if len(b) == 0 {
			break
		}

		i, err := ParseIERecursive(b)
		if err != nil {
			return nil, err
		}
		ies = append(ies, i)

		if len(i.IE) == 0 {
			b = b[i.MarshalLen():]
			continue
		}

		if i.IE[0].MarshalLen() < i.MarshalLen()-2 {
			var l = 2
			for _, ie := range i.IE {
				l += ie.MarshalLen()
			}
			b = b[l:]
			continue
		}
		b = b[i.MarshalLen():]
	}
	return ies, nil
}

// ParseIERecursive parses given byte sequence as an IE.
func ParseIERecursive(b []byte) (*IE, error) {
	i := &IE{}
	if err := i.ParseRecursive(b); err != nil {
		return nil, err
	}
	return i, nil
}

// ParseRecursive sets the values retrieved from byte sequence in an IE.
func (i *IE) ParseRecursive(b []byte) error {
	l := len(b)
	var offset int = 2
	if l < 2 {
		return io.ErrUnexpectedEOF
	}

	i.Tag = Tag(b[0])
	i.Length, offset = readLength(b)
	if int(i.Length)+offset > len(b) {
		return nil
	}
	i.Value = b[offset : offset+int(i.Length)]

	if i.Tag.Form() == 1 {
		x, err := ParseAsBER(i.Value)
		if err != nil {
			return nil
		}
		i.IE = append(i.IE, x...)
	}

	return nil
}

// MarshalLen returns the serial length of IE.
func (i *IE) MarshalLen() int {
	if(i.Value != nil && len(i.Value) > 0){
		return handleMarshalLen(uint8(len(i.Value)), len(i.Value))
	} else {
		return handleMarshalLen(i.Length, len(i.Value))
	}
}

// SetLength sets the length in Length field.
func (i *IE) SetLength() {
	i.Length = uint8(len(i.Value))
}

// String returns IE in human readable string.
func (i *IE) String() string {
	return fmt.Sprintf("{Tag: %#x, Length: %d, Value: %x, IE: %v}",
		i.Tag,
		i.Length,
		i.Value,
		i.IE,
	)
}
