package protowire

type wireType uint8

const (
	wireVarint wireType = iota
	wireFixed64
	wireLengthDelimited
	wireStartGroup
	wireEndGroup
	wireFixed32
)

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

type fieldType string

const (
	fieldDefault  fieldType = "default"
	fieldPacked   fieldType = "packed"
	fieldRequired fieldType = "required"
)
