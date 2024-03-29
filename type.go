package protowire

import (
	"fmt"
)

// wireType はwireバイナリの各fieldのtype
type wireType uint8

const (
	wireVarint          wireType = 0
	wireFixed64         wireType = 1
	wireLengthDelimited wireType = 2
	// wireStartGroup unsupported wire type
	// wireEndGroup unsupported wire type
	wireFixed32 wireType = 5
)

func (wt wireType) Packable() bool {
	if wt == wireVarint || wt == wireFixed32 || wt == wireFixed64 {
		return true
	}
	return false
}

// fieldNumber はwireバイナリのフィールド番号
type fieldNumber uint32

// protoType はwireバイナリをパースするときにどの型としてパースするのかの情報
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

// fieldType はwireバイナリの各フィールドの形式などについての情報
type fieldType string

func newFieldType(s string) (fieldType, error) {
	switch ft := fieldType(s); ft {
	case fieldOptional, fieldPacked, fieldRepeated, fieldOneOf:
		return ft, nil
	default:
		return "", fmt.Errorf("unsupported field type: %s", s)
	}
}

const (
	fieldOptional fieldType = "optional"
	fieldPacked   fieldType = "packed"
	fieldRepeated fieldType = "repeated"
	fieldOneOf    fieldType = "oneof"
)

type fieldTypes []fieldType

func (fs fieldTypes) Has(ft fieldType) bool {
	for _, f := range fs {
		if f == ft {
			return true
		}
	}
	return false
}

func (fs fieldTypes) validate() error {
	if fs.Has(fieldOneOf) && fs.Has(fieldRepeated) {
		return fmt.Errorf("if field types has oneof, field type repeated can not set: %s", fs)
	}
	if fs.Has(fieldPacked) && !fs.Has(fieldRepeated) {
		return fmt.Errorf("if field types has packed, field types must have repeated: %s", fs)
	}
	return nil
}
