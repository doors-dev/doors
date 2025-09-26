// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package front

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/instance"
)

func IntoScopeSet(inst instance.Core, scope []Scope) []*ScopeSet {
	a := make([]*ScopeSet, len(scope))
	for i, s := range scope {
		a[i] = s.Scope(inst)
	}
	return a
}

type ScopeSet struct {
	Type string
	Id   string
	Opt  any
}

func (s *ScopeSet) MarshalJSON() ([]byte, error) {
	a := []any{s.Type, s.Id, s.Opt}
	return json.Marshal(a)
}

type Scope interface {
	Scope(inst instance.Core) *ScopeSet
}

type ScopeAutoId struct {
	once sync.Once
	id   string
}

func (s *ScopeAutoId) Id(inst instance.Core) string {
	s.once.Do(func() {
		id := inst.NewId()
		s.id = fmt.Sprint(id)
	})
	return s.id
}

func DebounceScope(id string, duration time.Duration, limit time.Duration) *ScopeSet {
	return &ScopeSet{
		Id:   id,
		Type: "debounce",
		Opt:  []any{duration.Milliseconds(), limit.Milliseconds()},
	}
}

func BlockingScope(id string) *ScopeSet {
	return &ScopeSet{
		Id:   id,
		Type: "blocking",
	}
}
func SerialScope(id string) *ScopeSet {
	return &ScopeSet{
		Id:   id,
		Type: "serial",
	}
}
func FrameScope(id string, frame bool) *ScopeSet {
	return &ScopeSet{
		Id:   id,
		Type: "frame",
		Opt:  frame,
	}
}

func ConcurrentScope(id string, groupId int) *ScopeSet {
	return &ScopeSet{
		Id:   id,
		Type: "concurrent",
		Opt:  groupId,
	}
}

func FreeScope(id string) *ScopeSet {
	return &ScopeSet{
		Id:   id,
		Type: "free",
	}
}
