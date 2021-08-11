package protowire

import (
	"reflect"
	"testing"
)

func Test_parseProtoBind(t *testing.T) {
	type tagTest struct {
		Age  int32  `protowire:"1,0,int32,optional"`
		Name string `protowire:"2,2,string,optional"`
		Max  string `protowire:"536870911,2,string,optional"`
	}
	type multipleFieldTypeTest struct {
		Age []int32 `protowire:"1,0,int32,repeated,packed"`
	}
	type invalidFieldNumber struct {
		Age int32 `protowire:"536870912,0,int32,optional"`
	}
	type invalidType struct {
		Age int32 `protowire:"1,8,xxx,optional"`
	}

	tests := []struct {
		name    string
		v       interface{}
		want    protoBind
		wantErr bool
	}{
		{
			name: "タグの値を読み取れる",
			v:    &tagTest{},
			want: protoBind{
				fieldsByNumber: map[uint32]protoFieldMetadata{
					1: {
						wt:  wireVarint,
						pt:  protoInt32,
						fts: fieldTypes{fieldOptional},
						rv:  reflect.ValueOf(int32(0)),
					},
					2: {
						wt:  wireLengthDelimited,
						pt:  protoString,
						fts: fieldTypes{fieldOptional},
						rv:  reflect.ValueOf(""),
					},
					536870911: {
						wt:  wireLengthDelimited,
						pt:  protoString,
						fts: fieldTypes{fieldOptional},
						rv:  reflect.ValueOf(""),
					},
				},
				oneOfsByNumber: nil,
			},
		},
		{
			name: "fieldTypeが複数の場合も読み取れる",
			v:    &multipleFieldTypeTest{},
			want: protoBind{
				fieldsByNumber: map[uint32]protoFieldMetadata{
					1: {
						wt:  wireVarint,
						pt:  protoInt32,
						fts: fieldTypes{fieldRepeated, fieldPacked},
						rv:  reflect.ValueOf([]int32(nil)),
					},
				},
				oneOfsByNumber: nil,
			},
		},
		{
			name: "タグにoneofが指定されていた場合はその実装なども読み取る",
			v:    &testOneOf{},
			want: protoBind{
				fieldsByNumber: map[uint32]protoFieldMetadata{
					1: {
						wt:  wireLengthDelimited,
						pt:  protoString,
						fts: fieldTypes{fieldOptional},
						rv:  reflect.ValueOf(""),
					},
				},
				oneOfsByNumber: map[uint32]oneOfField{
					2: {
						iface:     reflect.New(reflect.TypeOf((*isTestOneOf_TestIdentifier)(nil)).Elem()).Elem(),
						implement: reflect.ValueOf(&TestOneOf_Id{}),
						structField: protoFieldMetadata{
							wt:  wireLengthDelimited,
							pt:  protoString,
							fts: fieldTypes{fieldOneOf},
							rv:  reflect.ValueOf(""),
						},
					},
					3: {
						iface:     reflect.New(reflect.TypeOf((*isTestOneOf_TestIdentifier)(nil)).Elem()).Elem(),
						implement: reflect.ValueOf(&TestOneOf_Email{}),
						structField: protoFieldMetadata{
							wt:  wireLengthDelimited,
							pt:  protoString,
							fts: fieldTypes{fieldOneOf},
							rv:  reflect.ValueOf(""),
						},
					},
					4: {
						iface:     reflect.New(reflect.TypeOf((*isTestOneOf_TestMessage)(nil)).Elem()).Elem(),
						implement: reflect.ValueOf(&TestOneOf_TextMessage{}),
						structField: protoFieldMetadata{
							wt:  wireLengthDelimited,
							pt:  protoString,
							fts: fieldTypes{fieldOneOf},
							rv:  reflect.ValueOf(""),
						},
					},
					5: {
						iface:     reflect.New(reflect.TypeOf((*isTestOneOf_TestMessage)(nil)).Elem()).Elem(),
						implement: reflect.ValueOf(&TestOneOf_BinaryMessage{}),
						structField: protoFieldMetadata{
							wt:  wireLengthDelimited,
							pt:  protoBytes,
							fts: fieldTypes{fieldOneOf},
							rv:  reflect.ValueOf([]byte{}),
						},
					},
				},
			},
		},
		{
			name:    "vがポインタじゃないとエラー",
			v:       tagTest{},
			want:    protoBind{},
			wantErr: true,
		},
		{
			name:    "field numberが上限より大きいとエラー",
			v:       &invalidFieldNumber{},
			want:    protoBind{},
			wantErr: true,
		},
		{
			name:    "field numberが上限より大きいとエラー",
			v:       &invalidType{},
			want:    protoBind{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseProtoBind(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseProtoBind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.fieldsByNumber != nil {
				for k, v := range got.fieldsByNumber {
					want, ok := tt.want.fieldsByNumber[k]
					if !ok {
						t.Errorf("parseBindInfo() got = %v, want %v", got, tt.want)
					}
					if v.wt != want.wt ||
						!reflect.DeepEqual(v.fts, want.fts) ||
						v.pt != want.pt ||
						v.rv.Type().String() != want.rv.Type().String() {
						t.Errorf("parseBindInfo() got = \n%v\n, want \n%v", got, tt.want)
					}
				}
			}
			if got.oneOfsByNumber != nil {
				for k, v := range got.oneOfsByNumber {
					want, ok := tt.want.oneOfsByNumber[k]
					if !ok {
						t.Errorf("parseBindInfo() got = %v, want %v", got, tt.want)
					}
					if v.iface.String() != want.iface.String() ||
						v.implement.Type().String() != want.implement.Type().String() ||
						v.structField.wt != want.structField.wt ||
						!reflect.DeepEqual(v.structField.fts, want.structField.fts) ||
						v.structField.pt != want.structField.pt ||
						v.structField.rv.Type().String() != want.structField.rv.Type().String() {
						t.Errorf("parseBindInfo() got = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}
