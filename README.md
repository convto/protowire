# protowire

This is a binary unmarshaler with wire encoding that I made for learning purposes. DO NOT USE production！
The specification of the wire is as follows

- https://developers.google.com/protocol-buffers/docs/encoding
- https://developers.google.com/protocol-buffers/docs/proto3#scalar

## Usage

```go
bin, _ := hex.DecodeString("08b96010b292041801")

type wireMessage struct {
	Int32   int32 `protowire:"1,0,int32,optional"`
	Int64   int64 `protowire:"2,0,int64,optional"`
	Boolean bool  `protowire:"3,0,bool,optional"`
}

var wm wireMessage
protowire.Unmarshal(bin, &wm)

fmt.Printf("%+v", wm)
// -> {Int32:12345 Int64:67890 Boolean:true}
```

playground: https://play.golang.org/p/tdJvZhdYpcx

## Supported type

| Type | Meaning | Implemented |
| :---: | :--- | :--- |
|0|Varint|int32, int64, uint32, uint64, sint32, sint64, bool, ~~enum~~|
|1|64-bit|fixed64, sfixed64, double|
|2|Length-delimited|string, bytes, embedded messages, packed repeated fields|
|5|32-bit|fixed32, sfixed32, float|

## Supported field pattern
- oneof
- optional([The optional in proto3 is passed as an oneof value from the protoc compiler](https://github.com/protocolbuffers/protobuf/blob/master/docs/implementing_proto3_presence.md#background), so if We have already implemented oneof, We have actually implemented the optional.)
- embedded
- packed repeated
- unpacked repeated
