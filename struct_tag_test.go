package protowire

import (
	"reflect"
	"testing"
)

func Test_parseTags(t *testing.T) {
	type tagTest struct {
		Age  int32  `protowire:"1,0,int32"`
		Name string `protowire:"2,2,string"`
		Max  string `protowire:"536870911,2,string"`
	}
	type invalidFieldNumber struct {
		Age int32 `protowire:"536870912,0,int32"`
	}
	type invalidType struct {
		Age int32 `protowire:"1,8,xxx"`
	}

	tests := []struct {
		name    string
		v       interface{}
		want    map[uint32]structTag
		wantErr bool
	}{
		{
			name: "タグの値を読み取れる",
			v:    &tagTest{},
			want: map[uint32]structTag{
				1: {
					structFieldNum: 0,
					wt:             wireVarint,
					pt:             protoInt32,
					ft:             fieldDefault,
				},
				2: {
					structFieldNum: 1,
					wt:             wireLengthDelimited,
					pt:             protoString,
					ft:             fieldDefault,
				},
				536870911: {
					structFieldNum: 2,
					wt:             wireLengthDelimited,
					pt:             protoString,
					ft:             fieldDefault,
				},
			},
		},
		{
			name:    "vがポインタじゃないとエラー",
			v:       tagTest{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "field numberが上限より大きいとエラー",
			v:       &invalidFieldNumber{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "field numberが上限より大きいとエラー",
			v:       &invalidType{},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStructTags(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTags() got = %v, want %v", got, tt.want)
			}
		})
	}
}
