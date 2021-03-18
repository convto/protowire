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
		tags[uint32(fieldNum)] = structTag{tp: uint8(tp), structFieldNum: i}
	}
	return tags, nil
}

func Unmarshal(b []byte, v interface{}) error {
	sts, err := parseTags(v)
	if err != nil {
		return fmt.Errorf("failed to parse structTag from input interface{}: %w", err)
	}

	l := len(b)
	for l > 0 {
		// タグは可変長バイト列形式
		tag, n, err := readVarint(b)
		if err != nil {
			return fmt.Errorf("failed to unmarshal: %w", err)
		}
		l -= n
		// 仕様でtype, field_number合わせて32bitまでなので超えてたらエラー
		if tag > math.MaxUint32 {
			return fmt.Errorf("invalid structTag size: %w", OrverFlowErr)
		}
		// 下位3bitはtype, それ以外はfield_number
		fieldNum := tag >> 3
		tp := tag & 0x7
		fmt.Printf("readed byte size: %d, field_number: %d, type: %d\n", n, fieldNum, tp)
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
