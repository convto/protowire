package protowire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
)

var (
	OverflowErr    = errors.New("over flow")
	UnknownTypeErr = errors.New("unknown type")
)

func Unmarshal(b []byte, v interface{}) error {
	sts, err := parseStructTags(v)
	if err != nil {
		return fmt.Errorf("failed to parse structTag from input interface{}: %w", err)
	}

	for len(b) > 0 {
		// タグは可変長バイト列形式
		fn, wt, n, err := parseTag(b)
		if err != nil {
			return fmt.Errorf("failed to read tag: %w", err)
		}
		b = b[n:]

		st := sts[fn]
		rv := reflect.ValueOf(v).Elem().Field(st.structFieldNum)
		if rv.Kind() == reflect.Ptr {
			if rv.IsNil() {
				rv.Set(reflect.New(rv.Type().Elem()))
			}
			rv = reflect.Indirect(rv)
		}

		n, err = parseValue(st, rv, wt, b)
		if err != nil {
			return fmt.Errorf("failed to read value: %w", err)
		}
		b = b[n:]
	}
	return nil
}

func parseTag(b []byte) (fn uint32, wt wireType, n int, err error) {
	tag, n, err := readVarint(b)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read varint field: %w", err)
	}
	// 仕様でtype, field_number合わせて32bitまでなので超えてたらエラー
	if tag > math.MaxUint32 {
		return 0, 0, 0, fmt.Errorf("invalid structTag size: %w", OverflowErr)
	}
	// 下位3bitはtype, それ以外はfield_number
	fn = uint32(tag >> 3)
	wt = wireType(tag & 0x7)
	return fn, wt, n, nil
}

func parseValue(st structTag, rv reflect.Value, wt wireType, b []byte) (n int, err error) {
	if wt != st.wt {
		return 0, fmt.Errorf("wrong wire type, struct wire tag: %d, binary wire tag: %d", st.wt, wt)
	}
	switch wt {
	case wireVarint:
		return parseVarint(st, rv, b)
	case wireFixed64:
		return parseFixed64(st, rv, b)
	case wireLengthDelimited:
		return parseLengthDelimited(st, rv, b)
	case wireFixed32:
		return parseFixed32(st, rv, b)
	default:
		return 0, fmt.Errorf("unsupported type: %d, err: %w", wt, UnknownTypeErr)
	}
}

func parseVarint(st structTag, rv reflect.Value, b []byte) (n int, err error) {
	f, n, err := readVarint(b)
	if err != nil {
		return 0, fmt.Errorf("failed to read varint field: %w", err)
	}
	switch {
	case (st.pt == protoInt64 || st.pt == protoSint64) && rv.Kind() == reflect.Int64:
		i := int64(f)
		if st.pt.isZigzag() {
			i = int64((uint64(i) >> 1) ^ uint64(((i&1)<<63)>>63))
		}
		rv.SetInt(i)
	case (st.pt == protoInt32 || st.pt == protoSint32) && rv.Kind() == reflect.Int32:
		i := int32(f)
		if st.pt.isZigzag() {
			i = int32((uint32(i) >> 1) ^ uint32(((i&1)<<31)>>31))
		}
		rv.SetInt(int64(i))
	case st.pt == protoUint64 && rv.Kind() == reflect.Uint64, st.pt == protoUint32 && rv.Kind() == reflect.Uint32:
		rv.SetUint(f)
	case st.pt == protoBool && rv.Kind() == reflect.Bool:
		rv.SetBool(f&1 == 1)
	default:
		return 0, fmt.Errorf("unsupported type of varint, proto type: %s, struct field type: %s", st.pt, rv.Type().String())
	}
	return n, nil
}

func parseFixed64(st structTag, rv reflect.Value, b []byte) (n int, err error) {
	f := binary.LittleEndian.Uint64(b)
	n = 8
	switch {
	case st.pt == protoSfixed64 && rv.Kind() == reflect.Int64:
		rv.SetInt(int64(f))
	case st.pt == protoFixed64 && rv.Kind() == reflect.Uint64:
		rv.SetUint(f)
	case st.pt == protoDouble && rv.Kind() == reflect.Float64:
		rv.SetFloat(math.Float64frombits(f))
	default:
		return 0, fmt.Errorf("unsupported type of 64-bit, proto type: %s, struct field type: %s", st.pt, rv.Type().String())
	}
	return n, nil
}

func parseLengthDelimited(st structTag, rv reflect.Value, b []byte) (n int, err error) {
	byteLen, n, err := readVarint(b)
	if err != nil {
		return 0, fmt.Errorf("failed to read varint field: %w", err)
	}
	val := b[n : n+int(byteLen)]
	n += int(byteLen)

	switch {
	case st.pt == protoString && rv.Kind() == reflect.String:
		rv.SetString(string(val))
	case st.pt == protoBytes && rv.Type() == reflect.TypeOf([]byte(nil)):
		rv.SetBytes(val)
	case st.pt == protoEmbed && rv.Kind() == reflect.Struct:
		ptr := reflect.New(rv.Type())
		if err := Unmarshal(val, ptr.Interface()); err != nil {
			return 0, fmt.Errorf("failed to read enbeded field: %w", err)
		}
		rv.Set(reflect.Indirect(ptr))
	default:
		return 0, fmt.Errorf("unsupported type of length-delimited, proto type: %s, struct field type: %s", st.pt, rv.Type().String())
	}
	return n, nil
}

func parseFixed32(st structTag, rv reflect.Value, b []byte) (n int, err error) {
	f := binary.LittleEndian.Uint32(b)
	n = 4
	switch {
	case st.pt == protoSfixed32 && rv.Kind() == reflect.Int32:
		rv.SetInt(int64(int32(f)))
	case st.pt == protoFixed32 && rv.Kind() == reflect.Uint32:
		rv.SetUint(uint64(f))
	case st.pt == protoFloat && rv.Kind() == reflect.Float32:
		rv.SetFloat(float64(math.Float32frombits(f)))
	default:
		return 0, fmt.Errorf("unsupported type of 64-bit, proto type: %s, struct field type: %s", st.pt, rv.Type().String())
	}
	return n, nil
}

// readVarint は可変長バイト列の読み取り処理
func readVarint(b []byte) (v uint64, n int, err error) {
	// little endian で読み取っていく
	for shift := uint(0); ; shift += 7 {
		// 64bitこえたらoverflow
		if shift >= 64 {
			return 0, 0, fmt.Errorf("failed to read varint: %w", OverflowErr)
		}
		// 対象のbyteの下位7bitを読み取ってvにつめていく
		target := b[n]
		n++
		v |= uint64(target&0x7F) << shift
		// 最上位bitが0だったら終端なのでよみとり終了
		if target < 0x80 {
			break
		}
	}
	return v, n, nil
}
