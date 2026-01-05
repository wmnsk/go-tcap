// Copyright go-tcap authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package tcap

import (
	"fmt"
	"io"
)

// Code definitions.
const (
	AARQ = iota
	AARE
	ABRT
	// AUDT = 0
)

// Application Context definitions.
const (
	_ uint8 = iota
	NetworkLocUpContext
	LocationCancellationContext
	RoamingNumberEnquiryContext
	IstAlertingContext
	LocationInfoRetrievalContext
	CallControlTransferContext
	ReportingContext
	CallCompletionContext
	ServiceTerminationContext
	ResetContext
	HandoverControlContext
	SIWFSAllocationContext
	EquipmentMngtContext
	InfoRetrievalContext
	InterVlrInfoRetrievalContext
	SubscriberDataMngtContext
	TracingContext
	NetworkFunctionalSsContext
	NetworkUnstructuredSsContext
	ShortMsgGatewayContext
	ShortMsgRelayContext
	SubscriberDataModificationNotificationContext
	ShortMsgAlertContext
	MwdMngtContext
	ShortMsgMTRelayContext
	ImsiRetrievalContext
	MsPurgingContext
	SubscriberInfoEnquiryContext
	AnyTimeInfoEnquiryContext
	_
	GroupCallControlContext
	GprsLocationUpdateContext
	GprsLocationInfoRetrievalContext
	FailureReportContext
	GprsNotifyContext
	SsInvocationNotificationContext
	LocationSvcGatewayContext
	LocationSvcEnquiryContext
	AuthenticationFailureReportContext
	_
	_
	MmEventReportingContext
	AnyTimeInfoHandlingContext
)

// Result Value defnitions.
const (
	Accepted uint8 = iota
	RejectPerm
)

// Dialogue Service Diagnostic Tag defnitions.
const (
	_ int = iota
	DialogueServiceUser
	DialogueServiceProvider
)

// Reason defnitions for Dialogue Service User Diagnostic in ResultSourceDiagnostic.
const (
	Null uint8 = iota
	NoReasonGiven
	ApplicationContextNameNotSupplied
	NoCommonDialoguePortion = 2 // same as above...
)

// Abort Source defnitions.
const (
	AbortDialogueServiceUser int = iota
	AbortDialogueServiceProvider
)

// DialoguePDU represents a DialoguePDU field in Dialogue.
type DialoguePDU struct {
	Type                   Tag
	Length                 uint8
	ProtocolVersion        *IE
	ApplicationContextName *IE
	Result                 *IE
	ResultSourceDiagnostic *IE
	AbortSource            *IE
	UserInformation        *IE
}

// NewDialoguePDU creates a new DialoguePDU.
func NewDialoguePDU(dtype, pver int, ctx, ctxver, result uint8, diagsrc int, diagreason, abortsrc uint8, userinfo ...*IE) *DialoguePDU {
	d := &DialoguePDU{
		Type: NewApplicationWideConstructorTag(dtype),
		ProtocolVersion: &IE{
			Tag:   NewContextSpecificPrimitiveTag(0),
			Value: []byte{uint8(pver << 7)},
		},
		ApplicationContextName: NewApplicationContextName(ctx, ctxver),
		Result:                 NewResult(result),
		ResultSourceDiagnostic: NewResultSourceDiagnostic(diagsrc, diagreason),
		AbortSource: &IE{
			Tag:    NewContextSpecificPrimitiveTag(0),
			Length: 1,
			Value:  []byte{abortsrc},
		},
	}
	if len(userinfo) > 0 {
		d.UserInformation = &IE{
			Tag:   NewContextSpecificConstructorTag(30),
			Value: userinfo[0].Value,
		}
		d.UserInformation.SetLength()
	}
	d.SetLength()
	return d
}

// NewApplicationContextName creates a new ApplicationContextName as an IE.
// Note: In this function, each length in fields are hard-coded.
func NewApplicationContextName(ctx, ver uint8) *IE {
	return &IE{
		Tag:    NewContextSpecificConstructorTag(1),
		Length: uint8(9),
		Value:  []byte{0x06, 0x07, 4, 0, 0, 1, 0, ctx, ver},
	}
}

// NewResult returns a new Result.
func NewResult(res uint8) *IE {
	return &IE{
		Tag:    NewContextSpecificConstructorTag(2),
		Length: 3,
		Value:  []byte{0x02, 0x01, res},
	}
}

// NewResultSourceDiagnostic returns a new ResultSourceDiagnostic as an IE.
func NewResultSourceDiagnostic(dtype int, reason uint8) *IE {
	return &IE{
		Tag:    NewContextSpecificConstructorTag(3),
		Length: 5,
		Value: []byte{
			uint8(NewContextSpecificConstructorTag(dtype)),
			0x03,               // MarshalLength
			0x02, 0x01, reason, // Integer Tag, Length and Reason value.
		},
	}
}

// NewAbortSource returns a new AbortSource as an IE.
func NewAbortSource(src uint8) *IE {
	return &IE{
		Tag:    NewContextSpecificPrimitiveTag(4),
		Length: 1,
		Value:  []byte{src},
	}
}

// NewAARQ returns a new AARQ(Dialogue Request).
func NewAARQ(protover int, context, contextver uint8, userinfo ...*IE) *DialoguePDU {
	d := &DialoguePDU{
		Type: NewApplicationWideConstructorTag(AARQ),
		ProtocolVersion: &IE{
			Tag:   NewContextSpecificPrimitiveTag(0),
			Value: []byte{0x07, uint8(protover << 7)}, // I don't actually know what the 0x07(padding) means...
		},
		ApplicationContextName: NewApplicationContextName(context, contextver),
	}
	if len(userinfo) > 0 {
		d.UserInformation = &IE{
			Tag:   NewContextSpecificConstructorTag(30),
			Value: userinfo[0].Value,
		}
		d.UserInformation.SetLength()
	}
	d.SetLength()
	return d
}

// NewAARE returns a new AARE(Dialogue Response).
func NewAARE(protover int, context, contextver, result uint8, diagsrc int, reason uint8, userinfo ...*IE) *DialoguePDU {
	d := &DialoguePDU{
		Type: NewApplicationWideConstructorTag(AARE),
		ProtocolVersion: &IE{
			Tag:   NewContextSpecificPrimitiveTag(0),
			Value: []byte{0x07, uint8(protover << 7)}, // I don't actually know what the 0x07(padding) means...
		},
		ApplicationContextName: NewApplicationContextName(context, contextver),
		Result:                 NewResult(result),
		ResultSourceDiagnostic: NewResultSourceDiagnostic(diagsrc, reason),
	}
	if len(userinfo) > 0 {
		d.UserInformation = &IE{
			Tag:   NewContextSpecificConstructorTag(30),
			Value: userinfo[0].Value,
		}
		d.UserInformation.SetLength()
	}
	d.SetLength()
	return d
}

// NewABRT returns a new ABRT(Dialogue Abort).
func NewABRT(abortsrc uint8, userinfo ...*IE) *DialoguePDU {
	d := &DialoguePDU{
		Type: NewApplicationWideConstructorTag(ABRT),
		AbortSource: &IE{
			Tag:    NewContextSpecificPrimitiveTag(0),
			Length: 1,
			Value:  []byte{abortsrc},
		},
	}
	if len(userinfo) > 0 {
		d.UserInformation = &IE{
			Tag:   NewContextSpecificConstructorTag(30),
			Value: userinfo[0].Value,
		}
		d.UserInformation.SetLength()
	}
	d.SetLength()
	return d
}

/*
// NewAUDT returns a new AUDT(Unidirectional Dialogue).
func NewAUDT(protover int, context, contextver uint8, userinfo ...*IE) *DialoguePDU {
	d := NewDialoguePDU(
		AUDT,
		protover,
		context,
		contextver,
		0,
		0,
		0,
		0,
	)
	if len(userinfo) > 0 {
		d.UserInformation = userinfo[0]
	}
	d.ProtocolVersion.Clear()
	d.ApplicationContextName.Clear()
	d.Result.Clear()
	d.ResultSourceDiagnostic.Clear()
	d.SetLength()

	return d
}
*/

// MarshalBinary returns the byte sequence generated from a DialoguePDU.
func (d *DialoguePDU) MarshalBinary() ([]byte, error) {
	b := make([]byte, d.MarshalLen())
	if err := d.MarshalTo(b); err != nil {
		return nil, fmt.Errorf("failed to marshal DialoguePDU: %w", err)
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
func (d *DialoguePDU) MarshalTo(b []byte) error {
	if len(b) < 2 {
		return io.ErrUnexpectedEOF
	}

	b[0] = uint8(d.Type)
	b[1] = d.Length

	switch d.Type.Code() {
	case AARQ:
		return d.marshalAARQTo(b)
	case AARE:
		return d.marshalAARETo(b)
	case ABRT:
		return d.marshalABRTTo(b)
	default:
		return &InvalidCodeError{Code: d.Type.Code()}
	}
}

func (d *DialoguePDU) marshalAARQTo(b []byte) error {
	var offset = 2
	if field := d.ProtocolVersion; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	if field := d.ApplicationContextName; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	if field := d.UserInformation; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
	}

	return nil
}

func (d *DialoguePDU) marshalAARETo(b []byte) error {
	var offset = 2
	if field := d.ProtocolVersion; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	if field := d.ApplicationContextName; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	if field := d.Result; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	if field := d.ResultSourceDiagnostic; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	if field := d.UserInformation; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
	}

	return nil
}

func (d *DialoguePDU) marshalABRTTo(b []byte) error {
	var offset = 2
	if field := d.AbortSource; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
		offset += field.MarshalLen()
	}

	if field := d.UserInformation; field != nil {
		if err := field.MarshalTo(b[offset : offset+field.MarshalLen()]); err != nil {
			return err
		}
	}

	return nil
}

// ParseDialoguePDU parses given byte sequence as an DialoguePDU.
func ParseDialoguePDU(b []byte) (*DialoguePDU, error) {
	d := &DialoguePDU{}
	if err := d.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return d, nil
}

// UnmarshalBinary sets the values retrieved from byte sequence in an DialoguePDU.
func (d *DialoguePDU) UnmarshalBinary(b []byte) error {
	if len(b) < 4 {
		return io.ErrUnexpectedEOF
	}

	d.Type = Tag(b[0])
	d.Length = b[1]

	switch d.Type.Code() {
	case AARQ:
		return d.parseAARQFromBytes(b)
	case AARE:
		return d.parseAAREFromBytes(b)
	case ABRT:
		return d.parseABRTFromBytes(b)
	default:
		return &InvalidCodeError{Code: d.Type.Code()}
	}
}

func (d *DialoguePDU) parseAARQFromBytes(b []byte) error {
	var err error
	var offset = 2
	d.ProtocolVersion, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += d.ProtocolVersion.MarshalLen()

	d.ApplicationContextName, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += d.ApplicationContextName.MarshalLen()

	if offset < len(b)-1 {
		if b[offset] == uint8(NewContextSpecificConstructorTag(30)) {
			d.UserInformation, err = ParseIE(b[offset:])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DialoguePDU) parseAAREFromBytes(b []byte) error {
	var err error
	var offset = 2
	d.ProtocolVersion, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += d.ProtocolVersion.MarshalLen()

	d.ApplicationContextName, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += d.ApplicationContextName.MarshalLen()

	d.Result, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += d.Result.MarshalLen()

	d.ResultSourceDiagnostic, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += d.ResultSourceDiagnostic.MarshalLen()

	if offset < len(b)-1 {
		if b[offset] == uint8(NewContextSpecificConstructorTag(30)) {
			d.UserInformation, err = ParseIE(b[offset:])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *DialoguePDU) parseABRTFromBytes(b []byte) error {
	var err error
	var offset = 2
	d.AbortSource, err = ParseIE(b[offset:])
	if err != nil {
		return err
	}
	offset += d.AbortSource.MarshalLen()
	if offset < len(b)-1 {
		if b[offset] == uint8(NewContextSpecificConstructorTag(30)) {
			d.UserInformation, err = ParseIE(b[offset:])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// MarshalLen returns the serial length of DialoguePDU.
func (d *DialoguePDU) MarshalLen() int {
	l := 2
	switch d.Type.Code() {
	case AARQ:
		if field := d.ProtocolVersion; field != nil {
			l += field.MarshalLen()
		}
		if field := d.ApplicationContextName; field != nil {
			l += field.MarshalLen()
		}
	case AARE:
		if field := d.ProtocolVersion; field != nil {
			l += field.MarshalLen()
		}
		if field := d.ApplicationContextName; field != nil {
			l += field.MarshalLen()
		}
		if field := d.Result; field != nil {
			l += field.MarshalLen()
		}
		if field := d.ResultSourceDiagnostic; field != nil {
			l += field.MarshalLen()
		}
	case ABRT:
		if field := d.AbortSource; field != nil {
			l += field.MarshalLen()
		}
	}

	if field := d.UserInformation; field != nil {
		l += field.MarshalLen()
	}
	return l
}

// SetLength sets the length in Length field.
func (d *DialoguePDU) SetLength() {
	switch d.Type.Code() {
	case AARQ:
		if field := d.ProtocolVersion; field != nil {
			field.SetLength()
		}
		if field := d.ApplicationContextName; field != nil {
			field.SetLength()
		}
	case AARE:
		if field := d.ProtocolVersion; field != nil {
			field.SetLength()
		}
		if field := d.ApplicationContextName; field != nil {
			field.SetLength()
		}
		if field := d.Result; field != nil {
			field.SetLength()
		}
		if field := d.ResultSourceDiagnostic; field != nil {
			field.SetLength()
		}
	case ABRT:
		if field := d.AbortSource; field != nil {
			field.SetLength()
		}
	}
	if field := d.UserInformation; field != nil {
		field.SetLength()
	}
	d.Length = uint8(d.MarshalLen() - 2)
}

// DialogueType returns the name of Dialogue Type in string.
func (d *DialoguePDU) DialogueType() string {
	switch d.Type.Code() {
	case AARQ:
		return "AARQ"
	case AARE:
		return "AARE"
	case ABRT:
		return "ABRT"
	default:
		return ""
	}
}

// Version returns Protocol Version in string.
func (d *DialoguePDU) Version() string {
	if d.Type.Code() == AARQ || d.Type.Code() == AARE {
		return fmt.Sprintf("%d", d.ProtocolVersion.Value[len(d.ProtocolVersion.Value)-1]>>7)
	}
	return ""
}

// Context returns the Context part of ApplicationContextName in string.
func (d *DialoguePDU) Context() string {
	appCtx := d.ApplicationContextName
	if appCtx == nil {
		return ""
	}
	if len(appCtx.Value) < 8 {
		return ""
	}

	if d.Type.Code() == AARQ || d.Type.Code() == AARE {
		switch appCtx.Value[7] {
		case NetworkLocUpContext:
			return "networkLocUpContext"
		case LocationCancellationContext:
			return "locationCancellationContext"
		case RoamingNumberEnquiryContext:
			return "roamingNumberEnquiryContext"
		case IstAlertingContext:
			return "istAlertingContext"
		case LocationInfoRetrievalContext:
			return "locationInfoRetrievalContext"
		case CallControlTransferContext:
			return "callControlTransferContext"
		case ReportingContext:
			return "reportingContext"
		case CallCompletionContext:
			return "callCompletionContext"
		case ServiceTerminationContext:
			return "serviceTerminationContext"
		case ResetContext:
			return "resetContext"
		case HandoverControlContext:
			return "handoverControlContext"
		case SIWFSAllocationContext:
			return "sIWFSAllocationContext"
		case EquipmentMngtContext:
			return "equipmentMngtContext"
		case InfoRetrievalContext:
			return "infoRetrievalContext"
		case InterVlrInfoRetrievalContext:
			return "interVlrInfoRetrievalContext"
		case SubscriberDataMngtContext:
			return "SubscriberDataMngtContext"
		case TracingContext:
			return "tracingContext"
		case NetworkFunctionalSsContext:
			return "networkFunctionalSsContext"
		case NetworkUnstructuredSsContext:
			return "networkUnstructuredSsContext"
		case ShortMsgGatewayContext:
			return "shortMsgGatewayContext"
		case ShortMsgRelayContext:
			return "shortMsgRelayContext"
		case SubscriberDataModificationNotificationContext:
			return "subscriberDataModificationNotificationContext"
		case ShortMsgAlertContext:
			return "shortMsgAlertContext"
		case MwdMngtContext:
			return "mwdMngtContext"
		case ShortMsgMTRelayContext:
			return "shortMsgMTRelayContext"
		case ImsiRetrievalContext:
			return "imsiRetrievalContext"
		case MsPurgingContext:
			return "msPurgingContext"
		case SubscriberInfoEnquiryContext:
			return "subscriberInfoEnquiryContext"
		case AnyTimeInfoEnquiryContext:
			return "anyTimeInfoEnquiryContext"
		case GroupCallControlContext:
			return "groupCallControlContext"
		case GprsLocationUpdateContext:
			return "gprsLocationUpdateContext"
		case GprsLocationInfoRetrievalContext:
			return "gprsLocationInfoRetrievalContext"
		case FailureReportContext:
			return "failureReportContext"
		case GprsNotifyContext:
			return "gprsNotifyContext"
		case SsInvocationNotificationContext:
			return "ssInvocationNotificationContext"
		case LocationSvcGatewayContext:
			return "locationSvcGatewayContext"
		case LocationSvcEnquiryContext:
			return "locationSvcEnquiryContext"
		case AuthenticationFailureReportContext:
			return "authenticationFailureReportContext"
		case MmEventReportingContext:
			return "mmEventReportingContext"
		case AnyTimeInfoHandlingContext:
			return "anyTimeInfoHandlingContext"
		}
	}

	return ""
}

// ContextVersion returns the Version part of ApplicationContextName in string.
func (d *DialoguePDU) ContextVersion() string {
	appCtx := d.ApplicationContextName
	if appCtx == nil {
		return ""
	}
	if len(appCtx.Value) < 8 {
		return ""
	}

	if d.Type.Code() == AARQ || d.Type.Code() == AARE {
		return fmt.Sprintf("%d", appCtx.Value[8])
	}
	return ""
}

// String returns DialoguePDU in human readable string.
func (d *DialoguePDU) String() string {
	return fmt.Sprintf("{Type: %#x, Length: %d, ProtocolVersion: %v, ApplicationContextName: %v, Result: %v, ResultSourceDiagnostic: %v, AbortSource: %v, UserInformation: %v}",
		d.Type,
		d.Length,
		d.ProtocolVersion,
		d.ApplicationContextName,
		d.Result,
		d.ResultSourceDiagnostic,
		d.AbortSource,
		d.UserInformation,
	)
}
