# go-tcap

Simple TCAP implementation in Golang

[![CircleCI](https://circleci.com/gh/wmnsk/go-tcap.svg?style=shield)](https://circleci.com/gh/wmnsk/go-tcap)
[![GolangCI](https://golangci.com/badges/github.com/wmnsk/go-tcap.svg)](https://golangci.com/r/github.com/wmnsk/go-tcap)
[![GoDoc](https://godoc.org/github.com/wmnsk/go-tcap?status.svg)](https://godoc.org/github.com/wmnsk/go-tcap)
[![GitHub](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/wmnsk/go-tcap/blob/master/LICENSE)

Package tcap provides simple and painless handling of TCAP(Transaction Capabilities Application Part) in SS7/SIGTRAN protocol stack, implemented in the Go Programming Language.

Though TCAP is ASN.1-based protocol, this implementation does not use any ASN.1 parser. That makes this implementation flexible enough to create arbitrary payload with any combinations, which is useful for testing.

## Disclaimer

This is still an experimental project, and currently in its very early stage of development. Any part of implementations(including exported APIs) may be changed before released as v1.0.0.

## Getting started

The following package should be installed before getting started.

```shell-session
go get -u github.com/pascaldekloe/goe
go get -u github.com/pkg/errors
```

If you use Go 1.11+, you can also use Go Modules.


```shell-session
GO111MODULE=on go [test | build | run | etc...]
```

## Supported Features

### Transaction Portion

#### Message Types

| Message type   | Supported? |
| -------------- | ---------- |
| Unidirectional |            |
| Begin          | Yes        |
| End            | Yes        |
| Continue       | Yes        |
| Abort          | Yes        |

#### Fields

| Tag                        | Supported? |
| -------------------------- | ---------- |
| Originating Transaction ID | Yes        |
| Destination Transaction ID | Yes        |
| P-Abort Cause              | Yes        |

### Component Portion

#### Component types

| Component type           | Supported? |
| ------------------------ | ---------- |
| Invoke                   | Yes        |
| Return Result (Last)     | Yes        |
| Return Result (Not Last) | Yes        |
| Return Error             | Yes        |
| Reject                   | Yes        |


### Dialogue Portion

#### Dialogue types

| Dialogue type                       | Supported? |
| ----------------------------------- | ---------- |
| Dialogue Request (AARQ-apdu)        | Yes        |
| Dialogue Response (AARE-apdu)       | Yes        |
| Dialogue Abort (ABRT-apdu)          | Yes        |
| Unidirectional Dialogue (AUDT-apdu) |            |

#### Elements 

| Tag                         | Type         | Supported? |
| --------------------------- | ------------ | ---------- |
| Object Identifier           | Structured   | Yes        |
| Single-ASN.1-type           | Structured   | Yes        |
| Dialogue PDU                | Structured   | Yes        |
| Object Identifier           | Unstructured |            |
| Single-ASN.1-type           | Unstructured |            |
| Unidirectional Dialogue PDU | Unstructured |            |


## Author(s)

Yoshiyuki Kurauchi ([My Website](https://wmnsk.com/) / [Twitter](https://twitter.com/wmnskdmms))

I'm always open to welcome co-authors! Please feel free to talk to me.

## LICENSE

[MIT](https://github.com/wmnsk/go-tcap/blob/master/LICENSE)
