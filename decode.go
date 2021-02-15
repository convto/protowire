package protowire

import (
	"errors"
	"fmt"
	"math"
)

var (
	OrverFlowErr = errors.New("over flow")
)

func Unmarshal(b []byte, v interface{}) error {
	l := len(b)
	for l > 0 {
		// タグは可変長バイト列形式
		tag, n, err := readVarint(b)
		if err != nil {
			return err
		}
		l -= n
		// 仕様でtype, field_number合わせて32bitまでなので超えてたらエラー
		if tag > math.MaxUint32 {
			return OrverFlowErr
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
			return 0, 0, OrverFlowErr
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