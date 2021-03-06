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
		Int32   int32 `protowire:"1,0,int32,optional"`
		Int64   int64 `protowire:"2,0,int64,optional"`
		Boolean bool  `protowire:"3,0,bool,optional"`
	}

	testVarintZigzagBin, _ := proto.Marshal(&testdata.TestVarintZigzag{
		Sint32: -12345,
		Sint64: -67890,
	})
	type testVarintZigzag struct {
		Sint32 int32 `protowire:"1,0,sint32,optional"`
		Sint64 int64 `protowire:"2,0,sint64,optional"`
	}

	testLengthDelimitedBin, _ := proto.Marshal(&testdata.TestLengthDelimited{
		Str:   "これはてすとだよ",
		Bytes: []byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA},
	})
	type testLengthDelimited struct {
		Str   string `protowire:"1,2,string,optional"`
		Bytes []byte `protowire:"2,2,bytes,optional"`
	}

	test64BitBin, _ := proto.Marshal(&testdata.Test64Bit{
		Fixed64:  12345,
		Sfixed64: -67890,
		Double:   1.23456789,
	})
	type test64Bit struct {
		Fixed64  uint64  `protowire:"1,1,fixed64,optional"`
		Sfixed64 int64   `protowire:"2,1,sfixed64,optional"`
		Double   float64 `protowire:"3,1,double,optional"`
	}

	test32BitBin, _ := proto.Marshal(&testdata.Test32Bit{
		Fixed32:  12345,
		Sfixed32: -67890,
		Float:    1.23456789,
	})
	type test32Bit struct {
		Fixed32  uint32  `protowire:"1,5,fixed32,optional"`
		Sfixed32 int32   `protowire:"2,5,sfixed32,optional"`
		Float    float32 `protowire:"3,5,float,optional"`
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
		TestVarint          *testVarint          `protowire:"1,2,embed,optional"`
		TestLengthDelimited *testLengthDelimited `protowire:"2,2,embed,optional"`
		Test64Bit           *test64Bit           `protowire:"3,2,embed,optional"`
	}

	testRepeatedBin, _ := proto.Marshal(&testdata.TestRepeated{
		Int64: []int64{
			12345,
			67890,
			-12345,
			-67890,
		},
		Fixed64: []uint64{
			12345,
			67890,
		},
		Fixed32: []uint32{
			12345,
			67890,
		},
		Str: []string{
			"これはてすとです",
			"this is test",
			"🐛",
		},
		Bytes: [][]byte{
			{0x00, 0x11, 0x22, 0x33},
			{0x44, 0x55, 0x66, 0x77},
			{0x88, 0x99, 0xAA, 0xBB},
			{0xCC, 0xDD, 0xEE, 0xFF},
		},
		TestLengthDelimited: []*testdata.TestLengthDelimited{
			{
				Str:   "これはてすとだよ🐛",
				Bytes: []byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA},
			},
			{
				Str:   "this is test",
				Bytes: []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
			},
		},
	})
	type testRepeated struct {
		Int64               []int64                `protowire:"1,2,int64,packed,repeated"`
		Fixed64             []uint64               `protowire:"2,2,fixed64,packed,repeated"`
		Fixed32             []uint32               `protowire:"3,2,fixed32,packed,repeated"`
		Str                 []string               `protowire:"4,2,string,repeated"`
		Bytes               [][]byte               `protowire:"5,2,bytes,repeated"`
		TestLengthDelimited []*testLengthDelimited `protowire:"6,2,embed,repeated"`
	}

	testOneOfBin, _ := proto.Marshal(&testdata.TestOneOf{
		Name: "test oneof",
		TestIdentifier: &testdata.TestOneOf_Id{
			Id: "identifier string",
		},
		TestMessage: &testdata.TestOneOf_TextMessage{
			TextMessage: "message string",
		},
	})

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
		{
			name: "Repeatedの検証バイナリ",
			args: args{
				b: testRepeatedBin,
				v: &testRepeated{},
			},
			want: &testRepeated{
				Int64: []int64{
					12345,
					67890,
					-12345,
					-67890,
				},
				Fixed64: []uint64{
					12345,
					67890,
				},
				Fixed32: []uint32{
					12345,
					67890,
				},
				Str: []string{
					"これはてすとです",
					"this is test",
					"🐛",
				},
				Bytes: [][]byte{
					{0x00, 0x11, 0x22, 0x33},
					{0x44, 0x55, 0x66, 0x77},
					{0x88, 0x99, 0xAA, 0xBB},
					{0xCC, 0xDD, 0xEE, 0xFF},
				},
				TestLengthDelimited: []*testLengthDelimited{
					{
						Str:   "これはてすとだよ🐛",
						Bytes: []byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA},
					},
					{
						Str:   "this is test",
						Bytes: []byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
					},
				},
			},
		},
		{
			name: "oneofの検証バイナリ",
			args: args{
				b: testOneOfBin,
				v: &testOneOf{},
			},
			want: &testOneOf{
				Name: "test oneof",
				TestIdentifier: &TestOneOf_Id{
					Id: "identifier string",
				},
				TestMessage: &TestOneOf_TextMessage{
					TextMessage: "message string",
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
