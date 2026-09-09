package instance

import (
	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/instance/utils"
)

type TabState = utils.TabState

type TabStateDerive[T any] struct {
	Key    string
	Equal  func(T, T) bool
	tab    utils.TabStateManager
	source beam.Lens[TabState, *T]
}

func (s *TabStateDerive[T]) Derive() {
	s.source = utils.DeriveTabState(s.tab, s.Key, s.Equal)
}

func (s *TabStateDerive[T]) Get() beam.Lens[TabState, *T] {
	return s.source
}

func (s *TabStateDerive[T]) setTab(tab utils.TabStateManager) {
	s.tab = tab
}

func (i *instance) TabState(s core.TabStateDeriver) {
	ss, ok := s.(interface {
		setTab(utils.TabStateManager)
	})
	if !ok {
		panic("unknown state deriver")
	}
	ss.setTab(i.tabState)
	s.Derive()
}
