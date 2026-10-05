// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
