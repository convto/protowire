package protowire

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type structTag struct {
	structFieldNum int
	wt             wireType
	pt             protoType
	ft             fieldType
}

// parseStructTags はstructに振ってあるprotowireタグを読み取ってmapに変換する
// mapのキーはfield_number
func parseStructTags(v interface{}) (map[uint32]structTag, error) {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return nil, errors.New("struct must be a pointer")
	}
	rt := reflect.TypeOf(v).Elem()
	fieldSize := rt.NumField()
	tags := make(map[uint32]structTag, fieldSize)
	for i := 0; i < fieldSize; i++ {
		f := rt.Field(i)
		t := strings.Split(f.Tag.Get("protowire"), ",")
		if len(t) < 3 {
			return nil, fmt.Errorf("invalid struct tag length, len: %d", len(t))
		}
		fieldNum, err := strconv.Atoi(t[0])
		if err != nil {
			return nil, err
		}
		if fieldNum > 1<<29-1 {
			return nil, errors.New("invalid protowire structTag, largest field_number is 536,870,911")
		}
		wt, err := strconv.Atoi(t[1])
		if wt > 7 {
			return nil, errors.New("invalid protowire structTag, largest type is 7")
		}
		pt := protoType(t[2])
		ft := fieldDefault
		if len(t) == 4 {
			ft = fieldType(t[3])
		}
		tags[uint32(fieldNum)] = structTag{
			structFieldNum: i,
			wt:             wireType(wt),
			pt:             pt,
			ft:             ft,
		}
	}
	return tags, nil
}
