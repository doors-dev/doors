package door

import "context"

type replaceState struct {
	ctx     context.Context
	content any
	id      uint64
	stateInit
}

func (s *replaceState) whenTakoverReady(f func(bool)) {
	f(true)
}

func (s *replaceState) takeover(prev state) {
	switch prev := prev.(type) {
	case *replaceState:
		return
	case *updateState:
		if !prev.isMounted() {
			return
		}
	}

}
