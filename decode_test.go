package protowire

import (
	"reflect"
	"testing"

	"github.com/convto/protowire/testdata"
	"github.com/golang/protobuf/proto"
)

func TestUnmarshal(t *testing.T) {
	testVarintBin, _ := proto.Marshal(&testdata.TestVarint{
		Int32:   12345,
		Int64:   67890,
		Boolean: true,
	})
	type testVarint struct {
		Int32   int32 `protowire:"1,0,int32"`
		Int64   int64 `protowire:"2,0,int64"`
		Boolean bool  `protowire:"3,0,bool"`
	}

	testVarintZigzagBin, _ := proto.Marshal(&testdata.TestVarintZigzag{
		Sint32: -12345,
		Sint64: -67890,
	})
	type testVarintZigzag struct {
		Sint32 int32 `protowire:"1,0,sint32"`
		Sint64 int64 `protowire:"2,0,sint64"`
	}

	testLengthDelimitedBin, _ := proto.Marshal(&testdata.TestLengthDelimited{
		Str:   "これはてすとだよ",
		Bytes: []byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA},
	})
	type testLengthDelimited struct {
		Str   string `protowire:"1,2,string"`
		Bytes []byte `protowire:"2,2,bytes"`
	}

	test64BitBin, _ := proto.Marshal(&testdata.Test64Bit{
		Fixed64:  12345,
		Sfixed64: -67890,
		Double:   1.23456789,
	})
	type test64Bit struct {
		Fixed64  uint64  `protowire:"1,1,fixed64"`
		Sfixed64 int64   `protowire:"2,1,sfixed64"`
		Double   float64 `protowire:"3,1,double"`
	}

	test32BitBin, _ := proto.Marshal(&testdata.Test32Bit{
		Fixed32:  12345,
		Sfixed32: -67890,
		Float:    1.23456789,
	})
	type test32Bit struct {
		Fixed32  uint32  `protowire:"1,5,fixed32"`
		Sfixed32 int32   `protowire:"2,5,sfixed32"`
		Float    float32 `protowire:"3,5,float"`
	}

	testEmbedBin, _ := proto.Marshal(&testdata.TestEmbed{
		EmbedVarint: &testdata.TestVarint{
			Int32:   -12345,
			Int64:   -67890,
			Boolean: true,
		},
		EmbedLengthDelimited: &testdata.TestLengthDelimited{
			Str:   "これはてすとだよ🐛",
			Bytes: []byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA},
		},
		Embed64Bit: &testdata.Test64Bit{
			Fixed64:  12345,
			Sfixed64: -67890,
			Double:   1.23456789,
		},
	})
	type testEmbed struct {
		TestVarint          *testVarint          `protowire:"1,2,embed"`
		TestLengthDelimited *testLengthDelimited `protowire:"2,2,embed"`
		Test64Bit           *test64Bit           `protowire:"3,2,embed"`
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
		{
			name: "Embedの検証バイナリ",
			args: args{
				b: testEmbedBin,
				v: &testEmbed{},
			},
			want: &testEmbed{
				TestVarint: &testVarint{
					Int32:   -12345,
					Int64:   -67890,
					Boolean: true,
				},
				TestLengthDelimited: &testLengthDelimited{
					Str:   "これはてすとだよ🐛",
					Bytes: []byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA},
				},
				Test64Bit: &test64Bit{
					Fixed64:  12345,
					Sfixed64: -67890,
					Double:   1.23456789,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Unmarshal(tt.args.b, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.v, tt.want) {
				t.Errorf("Unmarshal() got = %v, want %v", tt.args.v, tt.want)
			}
		})
	}
}
