package protowire

type testOneOf struct {
	Name           string                     `protowire:"1,2,string"`
	TestIdentifier isTestOneOf_TestIdentifier `protowire_oneof:"true"`
	TestMessage    isTestOneOf_TestMessage    `protowire_oneof:"true"`
}

type isTestOneOf_TestIdentifier interface {
	isTestOneOfIdentifier()
}

type TestOneOf_Id struct {
	Id string `protowire:"2,2,string,oneof"`
}

func (*TestOneOf_Id) isTestOneOfIdentifier() {}

type TestOneOf_Email struct {
	Email string `protowire:"3,2,string,oneof"`
}

func (*TestOneOf_Email) isTestOneOfIdentifier() {}

type isTestOneOf_TestMessage interface {
	isTestOneOfMessage()
}

type TestOneOf_TextMessage struct {
	TextMessage string `protowire:"4,2,string,oneof"`
}

func (*TestOneOf_TextMessage) isTestOneOfMessage() {}

type TestOneOf_BinaryMessage struct {
	BinaryMessage []byte `protowire:"5,2,bytes,oneof"`
}

func (*TestOneOf_BinaryMessage) isTestOneOfMessage() {}
