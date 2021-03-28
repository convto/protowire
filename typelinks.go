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

func getImplements(iface reflect.Type) ([]reflect.Value, error) {
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
