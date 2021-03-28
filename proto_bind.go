package protowire

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type protoBind struct {
	fieldsByNumber map[uint32]structField
	oneOfsByNumber map[uint32]oneOfField
}

// structField は `protowire` タグの内容やそのフィールドの reflect.Value を持ちます
type structField struct {
	wt wireType
	pt protoType
	ft fieldType
	rv reflect.Value
}

type oneOfField struct {
	iface       reflect.Value
	implement   reflect.Value
	structField structField
}

const (
	protoTag      = "protowire"
	protoOneOfTag = "protowire_oneof"
)

// parseProtoBind はstructの情報を読み取り、protobufのbindに必要な情報を生成する
func parseProtoBind(v interface{}) (protoBind, error) {
	rt := reflect.TypeOf(v)
	if rt.Kind() != reflect.Ptr {
		return protoBind{}, errors.New("target value must be a pointer")
	}
	rt = reflect.TypeOf(v).Elem()
	if rt.Kind() != reflect.Struct {
		return protoBind{}, errors.New("target value must be a struct")
	}
	pb := protoBind{
		fieldsByNumber: make(map[uint32]structField),
		oneOfsByNumber: make(map[uint32]oneOfField),
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		// protobuf_oneof タグには該当フィールドがoneofかどうかの情報が入る
		if t := f.Tag.Get(protoOneOfTag); t == "true" {
			iface := reflect.ValueOf(v).Elem().Field(i)
			ifaceTyp := iface.Type()
			if ifaceTyp.Kind() != reflect.Interface {
				return protoBind{}, fmt.Errorf("oneof field type must be interface, but %s", ifaceTyp.Kind().String())
			}
			rvs, err := getImplements(iface.Type())
			if err != nil {
				return protoBind{}, fmt.Errorf("failed to get %s implements: %w", ifaceTyp.String(), err)
			}
			for _, rv := range rvs {
				rt := rv.Type()
				if rt.Kind() == reflect.Ptr {
					rt = rt.Elem()
				}
				if rt.Kind() != reflect.Struct {
					return protoBind{}, errors.New("target value must be a struct")
				}
				if rt.NumField() != 1 {
					return protoBind{}, fmt.Errorf("oneof implement field size must be 1, but %d", rt.NumField())
				}
				fieldNum, sf, err := parseStructField(rt.Field(0), rv.Elem().Field(0))
				if err != nil {
					return protoBind{}, fmt.Errorf("failed to parse oneof struct field: %w", err)
				}
				if sf.ft != fieldOneOf {
					return protoBind{}, fmt.Errorf("oneof field type must be fieldOneOf, but %s", sf.ft)
				}
				pb.oneOfsByNumber[fieldNum] = oneOfField{
					iface:       iface,
					implement:   rv,
					structField: sf,
				}
			}
			continue
		}
		fieldNum, sf, err := parseStructField(f, reflect.ValueOf(v).Elem().Field(i))
		if err != nil {
			return protoBind{}, fmt.Errorf("failed to parse struct field: %w", err)
		}
		pb.fieldsByNumber[fieldNum] = sf
	}
	return pb, nil
}

func parseStructField(f reflect.StructField, rv reflect.Value) (uint32, structField, error) {
	t := strings.Split(f.Tag.Get(protoTag), ",")
	if len(t) < 3 {
		return 0, structField{}, fmt.Errorf("invalid struct tag length, len: %d", len(t))
	}
	fieldNum, err := strconv.Atoi(t[0])
	if err != nil {
		return 0, structField{}, err
	}
	if fieldNum > 1<<29-1 {
		return 0, structField{}, errors.New("invalid protowire structField, largest field_number is 536,870,911")
	}
	wt, err := strconv.Atoi(t[1])
	if wt > 7 {
		return 0, structField{}, errors.New("invalid protowire structField, largest type is 7")
	}
	pt := protoType(t[2])
	ft := fieldOptional
	if len(t) == 4 {
		ft = fieldType(t[3])
	}

	sf := structField{
		wt: wireType(wt),
		pt: pt,
		ft: ft,
		rv: rv,
	}
	return uint32(fieldNum), sf, nil
}
