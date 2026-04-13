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

package front

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/core"
)

func IntoScopeSet(inst core.Core, scope []Scope) []ScopeSet {
	a := make([]ScopeSet, len(scope))
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

func (s ScopeSet) MarshalJSON() ([]byte, error) {
	a := []any{s.Type, s.Id, s.Opt}
	return json.Marshal(a)
}

type Scope interface {
	Scope(core core.Core) ScopeSet
}

type AutoId struct {
	once sync.Once
	id   string
}

func (s *AutoId) Id(inst core.Core) string {
	s.once.Do(func() {
		id := inst.NewID()
		s.id = fmt.Sprint(id)
	})
	return s.id
}

func DebounceScope(id string, duration time.Duration, limit time.Duration) ScopeSet {
	return ScopeSet{
		Id:   id,
		Type: "debounce",
		Opt:  []any{duration.Milliseconds(), limit.Milliseconds()},
	}
}

func BlockingScope(id string) ScopeSet {
	return ScopeSet{
		Id:   id,
		Type: "blocking",
	}
}
func SerialScope(id string) ScopeSet {
	return ScopeSet{
		Id:   id,
		Type: "serial",
	}
}
func FrameScope(id string, frame bool) ScopeSet {
	return ScopeSet{
		Id:   id,
		Type: "frame",
		Opt:  frame,
	}
}

func ConcurrentScope(id string, groupId int) ScopeSet {
	return ScopeSet{
		Id:   id,
		Type: "concurrent",
		Opt:  groupId,
	}
}

func LatestScope(id string) ScopeSet {
	return ScopeSet{
		Id:   id,
		Type: "latest",
	}
}

func FreeScope(id string) ScopeSet {
	return ScopeSet{
		Id:   id,
		Type: "free",
	}
}
