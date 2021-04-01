package protowire

import (
	"errors"
	"fmt"
	"reflect"
)

type protoBind struct {
	fieldsByNumber map[uint32]structField
	oneOfsByNumber map[uint32]oneOfField
}

const protoOneOfTag = "protowire_oneof"

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
			oneOfFieldByNumber, err := newOneOfFields(reflect.ValueOf(v).Elem().Field(i))
			if err != nil {
				return protoBind{}, fmt.Errorf("failed to parse oneof fields: %w", err)
			}
			for fn, of := range oneOfFieldByNumber {
				pb.oneOfsByNumber[fn] = of
			}
			continue
		}
		fieldNum, sf, err := newStructField(f, reflect.ValueOf(v).Elem().Field(i))
		if err != nil {
			return protoBind{}, fmt.Errorf("failed to parse struct field: %w", err)
		}
		pb.fieldsByNumber[fieldNum] = sf
	}
	return pb, nil
}
