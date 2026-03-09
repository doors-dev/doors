package pipe

import "github.com/doors-dev/doors/internal/front/action"

type Payload interface {
	Payload() action.Payload
	Release()
}

func EmptyPayload() Payload {
	return emptyPayload{}
}

type emptyPayload struct{}

func (e emptyPayload) Payload() action.Payload {
	return action.NewText("")
}

func (e emptyPayload) Release() {}
