package protowire

import "fmt"

type wireType uint8

const (
	wireVarint wireType = iota
	wireFixed64
	wireLengthDelimited
	wireStartGroup
	wireEndGroup
	wireFixed32
)

func (wt wireType) Packable() bool {
	if wt == wireVarint || wt == wireFixed32 || wt == wireFixed64 {
		return true
	}
	return false
}

type protoType string

const (
	// varint proto type
	protoInt32  protoType = "int32"
	protoInt64  protoType = "int64"
	protoUint32 protoType = "uint32"
	protoUint64 protoType = "uint64"
	protoSint32 protoType = "sint32"
	protoSint64 protoType = "sint64"
	protoBool   protoType = "bool"
	protoEnum   protoType = "enum"
	// 64bit proto type
	protoFixed64  protoType = "fixed64"
	protoSfixed64 protoType = "sfixed64"
	protoDouble   protoType = "double"
	// length-Delimited proto type
	protoString protoType = "string"
	protoBytes  protoType = "bytes"
	protoEmbed  protoType = "embed"
	// 32bit proto type
	protoFixed32  protoType = "fixed32"
	protoSfixed32 protoType = "sfixed32"
	protoFloat    protoType = "float"
)

func (pt protoType) isZigzag() bool {
	if pt == protoSint32 || pt == protoSint64 {
		return true
	}
	return false
}

func (pt protoType) toWireType() (wireType, error) {
	switch pt {
	case protoInt32, protoInt64, protoUint32, protoUint64, protoSint32, protoSint64, protoBool, protoEnum:
		return wireVarint, nil
	case protoFixed64, protoSfixed64, protoDouble:
		return wireFixed64, nil
	case protoString, protoBytes, protoEmbed:
		return wireLengthDelimited, nil
	case protoFixed32, protoSfixed32, protoFloat:
		return wireFixed32, nil
	default:
		return 0, fmt.Errorf("unknown proto type: %s", pt)
	}
}

type fieldType string

const (
	fieldDefault  fieldType = "default"
	fieldPacked   fieldType = "packed"
	fieldRequired fieldType = "required"
	fieldOneOf    fieldType = "oneof"
)
