package protowire

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func Test_parseTags(t *testing.T) {
	type tagTest struct {
		Age  int    `protowire:"1,0"`
		Name string `protowire:"2,2"`
		Max  string `protowire:"536870911,7"`
	}
	type invalidFieldNumber struct {
		Age int `protowire:"536870912,0"`
	}
	type invalidType struct {
		Age int `protowire:"1,8"`
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
					tp:             0,
					structFieldNum: 0,
				},
				2: {
					tp:             2,
					structFieldNum: 1,
				},
				536870911: {
					tp:             7,
					structFieldNum: 2,
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
			got, err := parseTags(tt.v)
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

func TestUnmarshal(t *testing.T) {
	varintTestBin, _ := hex.DecodeString("08b96010b292041801")
	type varintTest struct {
		Int32   int32 `protowire:"1,0"`
		Int64   int64 `protowire:"2,0"`
		Boolean bool  `protowire:"3,0"`
	}

	type args struct {
		b []byte
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *varintTest
		wantErr bool
	}{
		{
			name: "Varintの検証バイナリ",
			args: args{
				b: varintTestBin,
				v: &varintTest{},
			},
			want: &varintTest{
				Int32:   12345,
				Int64:   67890,
				Boolean: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unmarshal(tt.args.b, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.v, tt.want) {
				t.Errorf("Unmarshal() got = %v, want %v", tt.args.b, tt.want)
			}
		})
	}
}
