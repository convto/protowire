package protowire

import (
	"reflect"
	"testing"
)

func Test_newProtoMetadata(t *testing.T) {
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
		want    protoMetadata
		wantErr bool
	}{
		{
			name: "タグの値を読み取れる",
			v:    &tagTest{},
			want: protoMetadata{
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
			want: protoMetadata{
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
			want: protoMetadata{
				fieldsByNumber: map[uint32]protoFieldMetadata{
					1: {
						wt:  wireLengthDelimited,
						pt:  protoString,
						fts: fieldTypes{fieldOptional},
						rv:  reflect.ValueOf(""),
					},
				},
				oneOfsByNumber: map[uint32]oneOfFieldMetadata{
					2: {
						iface:     reflect.New(reflect.TypeOf((*isTestOneOf_TestIdentifier)(nil)).Elem()).Elem(),
						implement: reflect.ValueOf(&TestOneOf_Id{}),
						protoFieldMetadata: protoFieldMetadata{
							wt:  wireLengthDelimited,
							pt:  protoString,
							fts: fieldTypes{fieldOneOf},
							rv:  reflect.ValueOf(""),
						},
					},
					3: {
						iface:     reflect.New(reflect.TypeOf((*isTestOneOf_TestIdentifier)(nil)).Elem()).Elem(),
						implement: reflect.ValueOf(&TestOneOf_Email{}),
						protoFieldMetadata: protoFieldMetadata{
							wt:  wireLengthDelimited,
							pt:  protoString,
							fts: fieldTypes{fieldOneOf},
							rv:  reflect.ValueOf(""),
						},
					},
					4: {
						iface:     reflect.New(reflect.TypeOf((*isTestOneOf_TestMessage)(nil)).Elem()).Elem(),
						implement: reflect.ValueOf(&TestOneOf_TextMessage{}),
						protoFieldMetadata: protoFieldMetadata{
							wt:  wireLengthDelimited,
							pt:  protoString,
							fts: fieldTypes{fieldOneOf},
							rv:  reflect.ValueOf(""),
						},
					},
					5: {
						iface:     reflect.New(reflect.TypeOf((*isTestOneOf_TestMessage)(nil)).Elem()).Elem(),
						implement: reflect.ValueOf(&TestOneOf_BinaryMessage{}),
						protoFieldMetadata: protoFieldMetadata{
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
			want:    protoMetadata{},
			wantErr: true,
		},
		{
			name:    "field numberが上限より大きいとエラー",
			v:       &invalidFieldNumber{},
			want:    protoMetadata{},
			wantErr: true,
		},
		{
			name:    "field numberが上限より大きいとエラー",
			v:       &invalidType{},
			want:    protoMetadata{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newProtoMetadata(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("newProtoMetadata() error = %v, wantErr %v", err, tt.wantErr)
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
						v.protoFieldMetadata.wt != want.protoFieldMetadata.wt ||
						!reflect.DeepEqual(v.protoFieldMetadata.fts, want.protoFieldMetadata.fts) ||
						v.protoFieldMetadata.pt != want.protoFieldMetadata.pt ||
						v.protoFieldMetadata.rv.Type().String() != want.protoFieldMetadata.rv.Type().String() {
						t.Errorf("parseBindInfo() got = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}
