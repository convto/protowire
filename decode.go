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
	pb, err := parseProtoBind(v)
	if err != nil {
		return fmt.Errorf("failed to parse protoBind from input interface{}: %w", err)
	}

	for len(b) > 0 {
		fn, wt, n, err := parseTag(b)
		if err != nil {
			return fmt.Errorf("failed to read tag: %w", err)
		}
		b = b[n:]

		sf, ok := pb.fieldsByNumber[fn]
		if ok {
			if !sf.rv.CanSet() {
				return fmt.Errorf("cant't set field, field type: %s", sf.rv.Type().String())
			}
			n, err = parseValue(sf, wt, b)
			if err != nil {
				return fmt.Errorf("failed to read value: %w", err)
			}
			b = b[n:]
		}
		osf, ok := pb.oneOfsByNumber[fn]
		if ok {
			if !osf.protoFieldMetadata.rv.CanSet() || !osf.iface.CanSet() {
				return fmt.Errorf("cant't set oneof field, field type: %s", osf.protoFieldMetadata.rv.Type().String())
			}
			n, err = parseValue(osf.protoFieldMetadata, wt, b)
			if err != nil {
				return fmt.Errorf("failed to read oneof value: %w", err)
			}
			osf.iface.Set(osf.implement)
			b = b[n:]
		}
	}
	return nil
}

// parseTag はfield numberとwire typeを読み取ります
func parseTag(b []byte) (fn uint32, wt wireType, n int, err error) {
	tag, n, err := readVarint(b)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read varint field: %w", err)
	}
	// 仕様でtype, field_number合わせて32bitまでなので超えてたらエラー
	if tag > math.MaxUint32 {
		return 0, 0, 0, fmt.Errorf("invalid tag size: %w", OverflowErr)
	}
	// 下位3bitはtype, それ以外はfield_number
	fn = uint32(tag >> 3)
	wt = wireType(tag & 0x7)
	return fn, wt, n, nil
}

// parseValue は与えられたタグ情報やwire typeをもとにバイト列をパースします
func parseValue(sf protoFieldMetadata, wt wireType, b []byte) (n int, err error) {
	// バイナリから読み取ったwire typeは基本的にstruct tagのwire typeと一致する
	// structのフィールド定義がpacked repeated fieldsの場合はwire typeはlength delimitedもありうるので一致していなくても許容する
	if wt != sf.wt && !sf.fts.Has(fieldPacked) {
		return 0, fmt.Errorf("wrong wire type, struct wire tag: %d, binary wire tag: %d", sf.wt, wt)
	}

	ptwt, err := sf.pt.toWireType()
	if err != nil {
		return 0, fmt.Errorf("failed to convert proto type to wire type: %w", err)
	}
	switch wt {
	case wireVarint:
		// structのフィールド定義がpacked repeated fieldsだった場合は互換のためlength delimitedでなくともsliceとしてパースする
		if sf.fts.Has(fieldPacked) && sf.wt == wireLengthDelimited {

			//if ptwt == wireVarint && sf.fts.Has(fieldPacked) && sf.rv.Kind() == reflect.Slice {
			elem := reflect.New(sf.rv.Type().Elem()).Elem()
			n, err := parseVarint(sf.pt, elem, b)
			if err != nil {
				return 0, fmt.Errorf("failed to read packed varint field: %w", err)
			}
			sf.rv.Set(reflect.Append(sf.rv, elem))
			return n, nil
		}
		return parseVarint(sf.pt, sf.rv, b)
	case wireFixed64:
		// structのフィールド定義がpacked repeated fieldsだった場合は互換のためlength delimitedでなくともsliceとしてパースする
		if ptwt == wireFixed64 && sf.fts.Has(fieldPacked) && sf.rv.Kind() == reflect.Slice {
			elem := reflect.New(sf.rv.Type().Elem()).Elem()
			n, err := parseFixed64(sf.pt, elem, b)
			if err != nil {
				return 0, fmt.Errorf("failed to read packed 64-bit field: %w", err)
			}
			sf.rv.Set(reflect.Append(sf.rv, elem))
			return n, nil
		}
		return parseFixed64(sf.pt, sf.rv, b)
	case wireLengthDelimited:
		// 該当フィールドがsliceとして宣言されていれば、複数回パースできるようにする
		// LengthDelimitedはpackedとして宣言できないので、packed形式のことは考慮しなくてよい
		// https://developers.google.com/protocol-buffers/docs/encoding#optional
		// >Only repeated fields of primitive numeric types (types which use the varint, 32-bit, or 64-bit wire types) can be declared "packed".
		if sf.rv.Kind() == reflect.Slice && sf.rv.Type() != reflect.TypeOf([]byte(nil)) && !ptwt.Packable() {
			elem := reflect.New(sf.rv.Type().Elem()).Elem()
			n, err := parseLengthDelimited(sf.pt, sf.fts, elem, b)
			if err != nil {
				return 0, fmt.Errorf("failed to read repeatable length-delimited field: %w", err)
			}
			sf.rv.Set(reflect.Append(sf.rv, elem))
			return n, nil
		}
		return parseLengthDelimited(sf.pt, sf.fts, sf.rv, b)
	case wireFixed32:
		// structのフィールド定義がpacked repeated fieldsだった場合は互換のためlength delimitedでなくともsliceとしてパースする
		if ptwt == wireFixed32 && sf.fts.Has(fieldPacked) && sf.rv.Kind() == reflect.Slice {
			elem := reflect.New(sf.rv.Type().Elem()).Elem()
			n, err := parseFixed32(sf.pt, elem, b)
			if err != nil {
				return 0, fmt.Errorf("failed to read packed 32-bit field: %w", err)
			}
			sf.rv.Set(reflect.Append(sf.rv, elem))
			return n, nil
		}
		return parseFixed32(sf.pt, sf.rv, b)
	default:
		return 0, fmt.Errorf("unsupported type: %d, err: %w", wt, UnknownTypeErr)
	}
}

// parseVarint はwire typeがvarintな場合の値の読み取りをします
// sint64, sint32 が指定された場合はバイト数の削減のためzigzag encodingを利用する
func parseVarint(pt protoType, rv reflect.Value, b []byte) (n int, err error) {
	val, n, err := readVarint(b)
	if err != nil {
		return 0, fmt.Errorf("failed to read varint field: %w", err)
	}
	switch {
	case (pt == protoInt64 || pt == protoSint64) && rv.Kind() == reflect.Int64:
		i := int64(val)
		if pt.isZigzag() {
			i = int64((uint64(i) >> 1) ^ uint64(((i&1)<<63)>>63))
		}
		rv.SetInt(i)
	case (pt == protoInt32 || pt == protoSint32) && rv.Kind() == reflect.Int32:
		i := int32(val)
		if pt.isZigzag() {
			i = int32((uint32(i) >> 1) ^ uint32(((i&1)<<31)>>31))
		}
		rv.SetInt(int64(i))
	case pt == protoUint64 && rv.Kind() == reflect.Uint64, pt == protoUint32 && rv.Kind() == reflect.Uint32:
		rv.SetUint(val)
	case pt == protoBool && rv.Kind() == reflect.Bool:
		rv.SetBool(val&1 == 1)
	default:
		return 0, fmt.Errorf("unsupported type of varint, proto type: %s, struct field type: %s", pt, rv.Type().String())
	}
	return n, nil
}

// parseFixed64 はwire typeが64-bitな場合の値の読み取りをします
func parseFixed64(pt protoType, rv reflect.Value, b []byte) (n int, err error) {
	val := binary.LittleEndian.Uint64(b)
	n = 8
	switch {
	case pt == protoSfixed64 && rv.Kind() == reflect.Int64:
		rv.SetInt(int64(val))
	case pt == protoFixed64 && rv.Kind() == reflect.Uint64:
		rv.SetUint(val)
	case pt == protoDouble && rv.Kind() == reflect.Float64:
		rv.SetFloat(math.Float64frombits(val))
	default:
		return 0, fmt.Errorf("unsupported type of 64-bit, proto type: %s, struct field type: %s", pt, rv.Type().String())
	}
	return n, nil
}

// parseLengthDelimited はwire typeがlength delimitedな場合の値の読み取りをします
// 先頭に可変長バイト列としてバイト長がエンコードされており、そのあとにデータが格納されています
func parseLengthDelimited(pt protoType, fts fieldTypes, rv reflect.Value, b []byte) (n int, err error) {
	byteLen, n, err := readVarint(b)
	if err != nil {
		return 0, fmt.Errorf("failed to read varint field: %w", err)
	}
	val := b[n : n+int(byteLen)]
	n += int(byteLen)

	switch {
	case pt == protoString && rv.Kind() == reflect.String:
		rv.SetString(string(val))
	case pt == protoBytes && rv.Type() == reflect.TypeOf([]byte(nil)):
		rv.SetBytes(val)
	case pt == protoEmbed && rv.Kind() == reflect.Ptr:
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		if err := Unmarshal(val, rv.Interface()); err != nil {
			return 0, fmt.Errorf("failed to read enbed field: %w", err)
		}
	case fts.Has(fieldPacked) && fts.Has(fieldRepeated) && rv.Kind() == reflect.Slice:
		// packed repeated fieldsの場合は該当フィールドのproto定義上の型情報を元にどのwire typeとしてパースすればよいか判断する
		ptwt, err := pt.toWireType()
		if err != nil {
			return 0, fmt.Errorf("failed to convert prototype to wiretype: %w", err)
		}
		switch ptwt {
		case wireVarint:
			for len(val) > 0 {
				elem := reflect.New(rv.Type().Elem()).Elem()
				m, err := parseVarint(pt, elem, val)
				if err != nil {
					return 0, fmt.Errorf("failed to read packed varint field: %w", err)
				}
				val = val[m:]
				rv.Set(reflect.Append(rv, elem))
			}
		case wireFixed64:
			for len(val) > 0 {
				elem := reflect.New(rv.Type().Elem()).Elem()
				m, err := parseFixed64(pt, elem, val)
				if err != nil {
					return 0, errors.New("failed to read packed 64-bit field")
				}
				val = val[m:]
				rv.Set(reflect.Append(rv, elem))
			}
		case wireFixed32:
			for len(val) > 0 {
				elem := reflect.New(rv.Type().Elem()).Elem()
				m, err := parseFixed32(pt, elem, val)
				if err != nil {
					return 0, errors.New("failed to read packed 32-bit field")
				}
				val = val[m:]
				rv.Set(reflect.Append(rv, elem))
			}
		}
	default:
		return 0, fmt.Errorf("unsupported type of length-delimited, proto type: %s, struct field type: %s", pt, rv.Type().String())
	}
	return n, nil
}

func parseFixed32(pt protoType, rv reflect.Value, b []byte) (n int, err error) {
	val := binary.LittleEndian.Uint32(b)
	n = 4
	switch {
	case pt == protoSfixed32 && rv.Kind() == reflect.Int32:
		rv.SetInt(int64(int32(val)))
	case pt == protoFixed32 && rv.Kind() == reflect.Uint32:
		rv.SetUint(uint64(val))
	case pt == protoFloat && rv.Kind() == reflect.Float32:
		rv.SetFloat(float64(math.Float32frombits(val)))
	default:
		return 0, fmt.Errorf("unsupported type of 64-bit, proto type: %s, struct field type: %s", pt, rv.Type().String())
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
