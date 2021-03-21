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
	testVarintBin, _ := hex.DecodeString("08b96010b292041801")
	type testVarint struct {
		Int32   int32 `protowire:"1,0"`
		Int64   int64 `protowire:"2,0"`
		Boolean bool  `protowire:"3,0"`
	}
	testVarintZigzagBin, _ := hex.DecodeString("08f1c00110e3a408")
	type testVarintZigzag struct {
		Sint32 int32 `protowire:"1,0,zigzag"`
		Sint64 int64 `protowire:"2,0,zigzag"`
	}
	testLengthDelimitedBin, _ := hex.DecodeString("0a18e38193e3828ce381afe381a6e38199e381a8e381a0e382881206ffeeddccbbaa")
	type testLengthDelimited struct {
		Str   string `protowire:"1,2"`
		Bytes []byte `protowire:"2,2"`
	}
	test64BitBin, _ := hex.DecodeString("09393000000000000011cef6feffffffffff191bde8342cac0f33f")
	type test64Bit struct {
		Fixed64  uint64  `protowire:"1,1"`
		Sfixed64 int64   `protowire:"2,1,zigzag"`
		Double   float64 `protowire:"3,1"`
	}
	test32BitBin, _ := hex.DecodeString("0d3930000015cef6feff1d52069e3f")
	type test32Bit struct {
		Fixed32  uint32  `protowire:"1,5"`
		Sfixed32 int32   `protowire:"2,5,zigzag"`
		Float    float32 `protowire:"3,5"`
	}

	type args struct {
		b []byte
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "Varintの検証バイナリ",
			args: args{
				b: testVarintBin,
				v: &testVarint{},
			},
			want: &testVarint{
				Int32:   12345,
				Int64:   67890,
				Boolean: true,
			},
		},
		{
			name: "Varintでzigzagの検証バイナリ",
			args: args{
				b: testVarintZigzagBin,
				v: &testVarintZigzag{},
			},
			want: &testVarintZigzag{
				Sint32: -12345,
				Sint64: -67890,
			},
		},
		{
			name: "Length-delimitedの検証バイナリ",
			args: args{
				b: testLengthDelimitedBin,
				v: &testLengthDelimited{},
			},
			want: &testLengthDelimited{
				Str:   "これはてすとだよ",
				Bytes: []byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA},
			},
		},
		{
			name: "64-bitの検証バイナリ",
			args: args{
				b: test64BitBin,
				v: &test64Bit{},
			},
			want: &test64Bit{
				Fixed64:  12345,
				Sfixed64: -67890,
				Double:   1.23456789,
			},
		},
		{
			name: "32-bitの検証バイナリ",
			args: args{
				b: test32BitBin,
				v: &test32Bit{},
			},
			want: &test32Bit{
				Fixed32:  12345,
				Sfixed32: -67890,
				Float:    1.23456789,
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
