package protowire

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

var (
	OrverFlowErr = errors.New("over flow")
	UnknownType  = errors.New("unknown type")
)

type structTag struct {
	tp             uint8
	structFieldNum int
	zigzag         bool
}

// parseTags はstructに振ってあるprotowireタグを読み取ってmapに変換する
// mapのキーはfield_number
func parseTags(v interface{}) (map[uint32]structTag, error) {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return nil, errors.New("struct must be a pointer")
	}
	rt := reflect.Indirect(reflect.ValueOf(v)).Type()
	fieldSize := rt.NumField()
	tags := make(map[uint32]structTag, fieldSize)
	for i := 0; i < fieldSize; i++ {
		f := rt.Field(i)
		t := strings.Split(f.Tag.Get("protowire"), ",")
		fieldNum, err := strconv.Atoi(t[0])
		if err != nil {
			return nil, err
		}
		if fieldNum > 1<<29-1 {
			return nil, errors.New("invalid protowire structTag, largest field_number is 536,870,911")
		}
		tp, err := strconv.Atoi(t[1])
		if tp > 7 {
			return nil, errors.New("invalid protowire structTag, largest type is 7")
		}
		zigzag := false
		if len(t) == 3 && t[2] == "zigzag" {
			zigzag = true
		}
		tags[uint32(fieldNum)] = structTag{tp: uint8(tp), structFieldNum: i, zigzag: zigzag}
	}
	return tags, nil
}

func Unmarshal(b []byte, v interface{}) error {
	sts, err := parseTags(v)
	if err != nil {
		return fmt.Errorf("failed to parse structTag from input interface{}: %w", err)
	}

	for len(b) > 0 {
		// タグは可変長バイト列形式
		tag, n, err := readVarint(b)
		if err != nil {
			return fmt.Errorf("failed to unmarshal: %w", err)
		}
		// 仕様でtype, field_number合わせて32bitまでなので超えてたらエラー
		if tag > math.MaxUint32 {
			return fmt.Errorf("invalid structTag size: %w", OrverFlowErr)
		}
		// 下位3bitはtype, それ以外はfield_number
		fieldNum := uint32(tag >> 3)
		tp := uint8(tag & 0x7)

		st := sts[fieldNum]
		if st.tp != tp {
			return fmt.Errorf("wrong type, structTag type: %d, wire type: %d", st.tp, tp)
		}
		b = b[n:]
		switch tp {
		case 0:
			f, n, err := readVarint(b)
			if err != nil {
				return fmt.Errorf("failed to read varint field: %w", err)
			}
			target := reflect.ValueOf(v).Elem().Field(st.structFieldNum)
			switch target.Interface().(type) {
			case int64:
				i := int64(f)
				if st.zigzag {
					i = int64((uint64(i) >> 1) ^ uint64(((i&1)<<63)>>63))
				}
				target.SetInt(i)
			case int32:
				i := int32(f)
				if st.zigzag {
					i = int32((uint32(i) >> 1) ^ uint32(((i&1)<<31)>>31))
				}
				target.SetInt(int64(i))
			case int16, int8, int:
				target.SetInt(int64(f))
			case uint64, uint32, uint16, uint8, uint:
				target.SetUint(f)
			case bool:
				target.SetBool(f&1 == 1)
			default:
				return fmt.Errorf("unsupported type of varint: %s", target.Type().String())
			}
			b = b[n:]
		case 2:
			byteLen, n, err := readVarint(b)
			if err != nil {
				return fmt.Errorf("failed to read varint field: %w", err)
			}
			b = b[n:]
			val := b[:byteLen]
			target := reflect.ValueOf(v).Elem().Field(st.structFieldNum)
			switch target.Interface().(type) {
			case string:
				target.SetString(string(val))
			case []byte:
				target.SetBytes(val)
			default:
				return fmt.Errorf("unsupported type of length-delimited: %s", target.Type().String())
			}
			b = b[int(byteLen):]
		default:
			return fmt.Errorf("unsupported type: %d, err: %w", tp, UnknownType)
		}
	}
	return nil
}

// readVarint は可変長バイト列の読み取り処理
func readVarint(b []byte) (v uint64, n int, err error) {
	// little endian で読み取っていく
	for shift := uint(0); ; shift += 7 {
		// 64bitこえたらoverflow
		if shift >= 64 {
			return 0, 0, fmt.Errorf("failed to read varint: %w", OrverFlowErr)
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
