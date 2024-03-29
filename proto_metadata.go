package protowire

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	protoOneOfTag = "protowire_oneof"
	protoTag      = "protowire"
)

type protoMetadata struct {
	fields      map[fieldNumber]protoFieldMetadata
	oneOfFields map[fieldNumber]oneOfFieldMetadata
}

// newProtoMetadata はstructの情報を読み取り、wireのパースに必要な情報を生成します
func newProtoMetadata(v interface{}) (protoMetadata, error) {
	rt := reflect.TypeOf(v)
	if rt.Kind() != reflect.Ptr {
		return protoMetadata{}, errors.New("target value must be a pointer")
	}
	rt = reflect.TypeOf(v).Elem()
	if rt.Kind() != reflect.Struct {
		return protoMetadata{}, errors.New("target value must be a struct")
	}
	pm := protoMetadata{
		fields:      make(map[fieldNumber]protoFieldMetadata),
		oneOfFields: make(map[fieldNumber]oneOfFieldMetadata),
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		// protobuf_oneof タグには該当フィールドがoneofかどうかの情報が入ります
		if t := f.Tag.Get(protoOneOfTag); t == "true" {
			oneOfFields, err := getOneOfFieldMetadataByIface(reflect.ValueOf(v).Elem().Field(i))
			if err != nil {
				return protoMetadata{}, fmt.Errorf("failed to get oneof fields: %w", err)
			}
			for fn, ofm := range oneOfFields {
				pm.oneOfFields[fn] = ofm
			}
			continue
		}
		fn, fm, err := newProtoFieldMetadata(f, reflect.ValueOf(v).Elem().Field(i))
		if err != nil {
			return protoMetadata{}, fmt.Errorf("failed to create struct field: %w", err)
		}
		pm.fields[fn] = fm
	}
	return pm, nil
}

// protoFieldMetadata は `protowire` タグの内容やそのフィールドの reflect.Value などの、wireのパースに必要なメタデータを表します
type protoFieldMetadata struct {
	wt  wireType
	pt  protoType
	fts fieldTypes
	rv  reflect.Value
}

// newProtoFieldMetadata はstructに振られた `protowire` タグ情報や、
// そのフィールドに値をSetするための reflect.Value 値などからmetadataを生成します
func newProtoFieldMetadata(f reflect.StructField, rv reflect.Value) (fieldNumber, protoFieldMetadata, error) {
	t := strings.Split(f.Tag.Get(protoTag), ",")
	if len(t) < 4 {
		return 0, protoFieldMetadata{}, fmt.Errorf("invalid struct tag length, len: %d", len(t))
	}
	fn, err := strconv.Atoi(t[0])
	if err != nil {
		return 0, protoFieldMetadata{}, err
	}
	if fn > 1<<29-1 {
		return 0, protoFieldMetadata{}, errors.New("invalid protoFieldMetadata, largest field_number is 536,870,911")
	}
	wt, err := strconv.Atoi(t[1])
	if wt > 7 {
		return 0, protoFieldMetadata{}, errors.New("invalid protoFieldMetadata, largest type is 7")
	}
	pt := protoType(t[2])
	fts := make([]fieldType, len(t[3:]))
	for i, v := range t[3:] {
		ft, err := newFieldType(v)
		if err != nil {
			return 0, protoFieldMetadata{}, fmt.Errorf("invalid field type: %w", err)
		}
		fts[i] = ft
	}

	fm := protoFieldMetadata{
		wt:  wireType(wt),
		pt:  pt,
		fts: fts,
		rv:  rv,
	}
	return fieldNumber(fn), fm, nil
}

// oneOfFieldMetadata はoneofをパースするためにinterfaceやその実装の情報とstructのフィールド定義を持ちます
// interfaceを実装するimplementの型はフィールド数1のstructである必要があります
// implementは実装のポインタであり、structFieldは実装のstructのフィールド情報です
// ifaceはstructから読み取ったoneofフィールドの値であり、ここに値をSetすれば元の構造の値が更新されます。手順は以下です
// 1: protoFieldMetadata.rv にセット
// 2: 1で implement が更新されるので iface に implement をセット
type oneOfFieldMetadata struct {
	iface              reflect.Value
	implement          reflect.Value
	protoFieldMetadata protoFieldMetadata
}

// getOneOfFieldMetadataByIface はあるoneofフィールドに代入される可能性のあるすべての構造の情報を読み取ります
// 実装上oneofのフィールドはinterfaceとなっており、その実装としていくつかのstructが存在することを想定しています
// あるoneofフィールドを実装しているstructをすべて読み取り、そのstructのタグ情報や、値の代入のためのreflect.Valueの取得などを行います
func getOneOfFieldMetadataByIface(iface reflect.Value) (map[fieldNumber]oneOfFieldMetadata, error) {
	ifaceTyp := iface.Type()
	if ifaceTyp.Kind() != reflect.Interface {
		return nil, fmt.Errorf("oneof field type must be interface, but %s", ifaceTyp.Kind().String())
	}
	rvs, err := getImplements(iface.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to get %s implements: %w", ifaceTyp.String(), err)
	}
	oneOfFields := make(map[fieldNumber]oneOfFieldMetadata, len(rvs))
	for _, rv := range rvs {
		rt := rv.Type()
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		if rt.Kind() != reflect.Struct {
			return nil, errors.New("target value must be a struct")
		}
		if rt.NumField() != 1 {
			return nil, fmt.Errorf("oneof implement field size must be 1, but %d", rt.NumField())
		}
		fieldNum, fm, err := newProtoFieldMetadata(rt.Field(0), rv.Elem().Field(0))
		if err != nil {
			return nil, fmt.Errorf("failed to parse oneof struct field: %w", err)
		}
		if !fm.fts.Has(fieldOneOf) {
			return nil, fmt.Errorf("oneof field type must be fieldOneOf, but %s", fm.fts)
		}
		oneOfFields[fieldNum] = oneOfFieldMetadata{
			iface:              iface,
			implement:          rv,
			protoFieldMetadata: fm,
		}
	}
	return oneOfFields, nil
}
