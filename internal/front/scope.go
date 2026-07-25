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
	"strconv"
	"sync"
	"time"

	"github.com/doors-dev/doors/internal/core"
)

type scopeKind string

const (
	rateScope       scopeKind = "rate"
	debounceScope   scopeKind = "debounce"
	blockingScope   scopeKind = "blocking"
	serialScope     scopeKind = "serial"
	concurrentScope scopeKind = "concurrent"
	latestScope     scopeKind = "latest"
	frameScope      scopeKind = "frame"
	freeScope       scopeKind = "free"
)

type Scoper interface {
	Scopes() []Scope
}

type Scope struct {
	Kind scopeKind
	Id   string
	Opt  any
}

func (s Scope) MarshalJSON() ([]byte, error) {
	a := []any{s.Kind, s.Id, s.Opt}
	return json.Marshal(a)
}

type AutoID struct {
	once sync.Once
	id   uint64
}

func (s *AutoID) ID(inst core.Core) uint64 {
	s.once.Do(func() {
		s.id = inst.Instance().NewID()
	})
	return s.id
}

func (s *AutoID) String(inst core.Core) string {
	return strconv.FormatUint(s.ID(inst), 10)
}

func RateScope(id string, tick time.Duration) Scope {
	return Scope{
		Id:   id,
		Kind: rateScope,
		Opt:  tick.Milliseconds(),
	}
}

func DebounceScope(id string, duration time.Duration, limit time.Duration) Scope {
	return Scope{
		Id:   id,
		Kind: debounceScope,
		Opt:  []any{duration.Milliseconds(), limit.Milliseconds()},
	}
}

func BlockingScope(id string) Scope {
	return Scope{
		Id:   id,
		Kind: blockingScope,
	}
}
func SerialScope(id string) Scope {
	return Scope{
		Id:   id,
		Kind: serialScope,
	}
}
func FrameScope(id string, frame bool) Scope {
	return Scope{
		Id:   id,
		Kind: frameScope,
		Opt:  frame,
	}
}

func ConcurrentScope(id string, groupId int) Scope {
	return Scope{
		Id:   id,
		Kind: concurrentScope,
		Opt:  groupId,
	}
}

func LatestScope(id string) Scope {
	return Scope{
		Id:   id,
		Kind: latestScope,
	}
}

func FreeScope(id string) Scope {
	return Scope{
		Id:   id,
		Kind: freeScope,
	}
}
