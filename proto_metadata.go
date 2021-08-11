package protowire

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const protoTag = "protowire"

// protoFieldMetadata は `protowire` タグの内容やそのフィールドの reflect.Value などの、wireのパースに必要なメタデータを表します
type protoFieldMetadata struct {
	wt  wireType
	pt  protoType
	fts fieldTypes
	rv  reflect.Value
}

// newProtoFieldMetadata はstructに振られた `protowire` タグ情報や、
// そのフィールドに値をSetするための reflect.Value 値などからmetadataを生成する
func newProtoFieldMetadata(f reflect.StructField, rv reflect.Value) (uint32, protoFieldMetadata, error) {
	t := strings.Split(f.Tag.Get(protoTag), ",")
	if len(t) < 4 {
		return 0, protoFieldMetadata{}, fmt.Errorf("invalid struct tag length, len: %d", len(t))
	}
	fieldNum, err := strconv.Atoi(t[0])
	if err != nil {
		return 0, protoFieldMetadata{}, err
	}
	if fieldNum > 1<<29-1 {
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

	sf := protoFieldMetadata{
		wt:  wireType(wt),
		pt:  pt,
		fts: fts,
		rv:  rv,
	}
	return uint32(fieldNum), sf, nil
}

// oneOfField はoneofをパースするためにinterfaceやその実装の情報とstructのフィールド定義を持ちます
// interfaceを実装するimplementの型はフィールド数1のstructである必要があります
// implementは実装のポインタであり、structFieldは実装のstructのフィールド情報です
// ifaceはstructから読み取ったoneofフィールドの値であり、ここに値をSetすれば元の構造の値が更新されます。手順は以下です
// 1: protoFieldMetadata.rv にセット
// 2: 1で implement が更新されるので iface に implement をセット
type oneOfField struct {
	iface       reflect.Value
	implement   reflect.Value
	structField protoFieldMetadata
}

// newOneOfFields はあるoneofフィールドに代入される可能性のあるすべての構造の情報を読み取ります
// 実装上oneofのフィールドはinterfaceとなっており、その実装としていくつかのstructが存在することを想定しています
// あるoneofフィールドを実装しているstructをすべて読み取り、そのstructのタグ情報や、値の代入のためのreflect.Valueの取得などを行います
func newOneOfFields(iface reflect.Value) (map[uint32]oneOfField, error) {
	ifaceTyp := iface.Type()
	if ifaceTyp.Kind() != reflect.Interface {
		return nil, fmt.Errorf("oneof field type must be interface, but %s", ifaceTyp.Kind().String())
	}
	rvs, err := getImplements(iface.Type())
	if err != nil {
		return nil, fmt.Errorf("failed to get %s implements: %w", ifaceTyp.String(), err)
	}
	oneOfsByNumber := make(map[uint32]oneOfField, len(rvs))
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
		fieldNum, sf, err := newProtoFieldMetadata(rt.Field(0), rv.Elem().Field(0))
		if err != nil {
			return nil, fmt.Errorf("failed to parse oneof struct field: %w", err)
		}
		if !sf.fts.Has(fieldOneOf) {
			return nil, fmt.Errorf("oneof field type must be fieldOneOf, but %s", sf.fts)
		}
		oneOfsByNumber[fieldNum] = oneOfField{
			iface:       iface,
			implement:   rv,
			structField: sf,
		}
	}
	return oneOfsByNumber, nil
}
