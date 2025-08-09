package common

type CallData struct {
	Name    string
	Arg     any
	Payload Writable
}

type Call interface {
	Data() *CallData
	Result(error)
}

