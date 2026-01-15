package door

import (
	"context"
)

type updateState struct {
	ctx     context.Context
	content any
	head    *doorHead
	tracker *tracker
	stateInit
	*stateTakover
}

func (s *updateState) isMounted() bool {
	return s.head != nil
}

func (s *updateState) whenTakoverReady(f func(bool)) {
	if s.stateTakover == nil {
		f(true)
		return
	}
	s.stateTakover.whenTakoverReady(f)
}

func (s *updateState) takeover(prev state) {
	switch prev := prev.(type) {
	case *replaceState:
		return
	case *updateState:
		if !prev.isMounted() {
			return
		}
		s.head = prev.head
		s.stateTakover = prev.stateTakover
	case *proxyState:
		s.head = prev.head
		s.stateTakover = &prev.stateTakover
	case *jobState:
		s.head = prev.head
		s.stateTakover = &prev.stateTakover
	}
}

