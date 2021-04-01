package protowire

import (
	"fmt"
	"reflect"
	"unsafe"
)

//go:linkname typelinks reflect.typelinks
func typelinks() ([]unsafe.Pointer, [][]int32)

//go:linkname rtypeOff reflect.rtypeOff
func rtypeOff(unsafe.Pointer, int32) unsafe.Pointer

// getImplements はreflectパッケージの非公開処理を用いて、与えられたinterfaceの実装を取得します
// TODO: コード生成などに着手するタイミングで実装の一覧を取得する関数を生成するようにしたい
func getImplements(iface reflect.Type) ([]reflect.Value, error) {
	if iface.Kind() != reflect.Interface {
		return nil, fmt.Errorf("iface must be interface, but %s", iface.Kind().String())
	}
	sections, offsets := typelinks()
	if len(sections) != 1 {
		return nil, fmt.Errorf("failed to get sections")
	}
	if len(offsets) != 1 {
		return nil, fmt.Errorf("failed to get offsets")
	}
	implements := make([]reflect.Value, 0)
	for i, base := range sections {
		for _, offset := range offsets[i] {
			typeAddr := rtypeOff(base, offset)
			typ := reflect.TypeOf(*(*interface{})(unsafe.Pointer(&typeAddr)))
			if typ.Implements(iface) {
				implements = append(implements, reflect.New(typ.Elem()))
			}
		}
	}
	return implements, nil
}
