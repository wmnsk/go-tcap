# go-tcap

Simple TCAP implementation in Golang

[![CircleCI](https://circleci.com/gh/wmnsk/go-tcap.svg?style=shield)](https://circleci.com/gh/wmnsk/go-tcap)
[![GoDoc](https://godoc.org/github.com/wmnsk/go-tcap?status.svg)](https://godoc.org/github.com/wmnsk/go-tcap)
[![GitHub](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/wmnsk/go-tcap/blob/master/LICENSE)

Package tcap provides simple and painless handling of TCAP(Transaction Capabilities Application Part) in SS7/SIGTRAN protocol stack, implemented in the Go Programming Language.

Though TCAP is ASN.1-based protocol, this implementation does not use any ASN.1 parser. That makes this implementation flexible enough to create arbitrary payload with any combinations, which is useful for testing.

## Disclaimer

This is still an experimental project, and currently in its very early stage of development. Any part of implementations(including exported APIs) may be changed before released as v1.0.0.

## Getting started

### Prerequisites

go-gtp supports Go Modules. Just run go mod tidy in your project's directory to collect the required packages automatically.

```
go mod tidy
```

_This project follows [the Release Policy of Go](https://golang.org/doc/devel/release.html#policy)._

### Running examples

A sample client is available in [examples/client/](./examples/client/), which, by default, establishes SCTP/M3UA connection with a server sends a MAP cancelLocation. 

```
Transaction Capabilities Application Part
    begin
        [Transaction Id: 11111111]
        Source Transaction ID
        oid: 0.0.17.773.1.1.1 (id-as-dialogue)
        dialogueRequest
        components: 1 item
            Component: invoke (1)
                invoke
                    invokeID: 0
                    opCode: localValue (0)
                    CONSTRUCTOR
                        CONSTRUCTOR Tag
                        Tag: 0x00
                        Length: 10
                        Parameter (0x04)
                            Tag: 0x04
                            Length: 8
                        Data: 00
GSM Mobile Application
    Component: invoke (1)
        invoke
            invokeID: 0
            opCode: localValue (0)
                localValue: cancelLocation (3)
            identity: imsi-WithLMSI (1)
                imsi-WithLMSI
                    IMSI: 001010123456789
                    [Association IMSI: 001010123456789]
```

Some parameters can be speficied from command-line arguments. Other parameters including the ones in lower layers (such as Point Code in M3UA, Global Title in SCCP, etc.) should be updated by modifying the source code.

```
$ ./client -h
Usage of client:
  -addr string
        Remote IP and Port to connect to. (default "127.0.0.2:2905")
  -opcode int
        Operation Code in int. (default 3)
  -otid int
        Originating Transaction ID in uint32. (default 286331153)
  -payload string
        Hex representation of the payload (default "040800010121436587f9")
```

_If you are looking for a server that just can accept a SCTP/M3UA connection, [server example in go-m3ua project](https://github.com/wmnsk/go-m3ua/blob/master/examples/server/m3ua-server.go) would be a nice choice for you._

## Supported Features

### Transaction Portion

#### Message Types

| Message type   | Supported? |
|----------------|------------|
| Unidirectional |            |
| Begin          | Yes        |
| End            | Yes        |
| Continue       | Yes        |
| Abort          | Yes        |

#### Fields

| Tag                        | Supported? |
|----------------------------|------------|
| Originating Transaction ID | Yes        |
| Destination Transaction ID | Yes        |
| P-Abort Cause              | Yes        |

### Component Portion

#### Component types

| Component type           | Supported? |
|--------------------------|------------|
| Invoke                   | Yes        |
| Return Result (Last)     | Yes        |
| Return Result (Not Last) | Yes        |
| Return Error             | Yes        |
| Reject                   | Yes        |


### Dialogue Portion

#### Dialogue types

| Dialogue type                       | Supported? |
|-------------------------------------|------------|
| Dialogue Request (AARQ-apdu)        | Yes        |
| Dialogue Response (AARE-apdu)       | Yes        |
| Dialogue Abort (ABRT-apdu)          | Yes        |
| Unidirectional Dialogue (AUDT-apdu) |            |

#### Elements 

| Tag                         | Type         | Supported? |
|-----------------------------|--------------|------------|
| Object Identifier           | Structured   | Yes        |
| Single-ASN.1-type           | Structured   | Yes        |
| Dialogue PDU                | Structured   | Yes        |
| Object Identifier           | Unstructured |            |
| Single-ASN.1-type           | Unstructured |            |
| Unidirectional Dialogue PDU | Unstructured |            |


## Author(s)

Yoshiyuki Kurauchi ([Website](https://wmnsk.com/) / [Twitter](https://twitter.com/wmnskdmms))

I'm always open to welcome co-authors! Please feel free to talk to me.

## LICENSE

[MIT](https://github.com/wmnsk/go-tcap/blob/master/LICENSE)
