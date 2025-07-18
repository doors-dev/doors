package common


type CallData struct {
	Name    string
	Arg     JsonWritable
	Payload Writable
}

type Call interface {
	Data() *CallData
	Result(error)
}

