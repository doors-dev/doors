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

package doors

import (
	"time"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
)

// Scope controls how overlapping client-side events coordinate before the
// backend request starts.
type Scope = front.Scope

// ScopeSet is the serialized form of a [Scope].
type ScopeSet = front.ScopeSet

type scopeFunc func(core core.Core) ScopeSet

func (s scopeFunc) Scope(core core.Core) ScopeSet {
	return s(core)
}

// ScopeBlocking rejects a new event while another event in the same shared
// scope is still running.
type ScopeBlocking struct {
	id front.AutoId
}

func (b *ScopeBlocking) Scope(core core.Core) ScopeSet {
	return front.BlockingScope(b.id.Id(core))
}

// ScopeOnlyBlocking returns a single [ScopeBlocking].
func ScopeOnlyBlocking() []Scope {
	return []Scope{&ScopeBlocking{}}
}

// ScopeSerial queues accepted events and runs them in arrival order.
type ScopeSerial struct {
	id front.AutoId
}

func (b *ScopeSerial) Scope(core core.Core) ScopeSet {
	return front.SerialScope(b.id.Id(core))
}

// ScopeOnlySerial returns a single [ScopeSerial].
func ScopeOnlySerial() []Scope {
	return []Scope{&ScopeSerial{}}
}

// ScopeDebounce delays a burst of events and keeps the latest pending one.
type ScopeDebounce struct {
	id front.AutoId
}

// Scope returns a debounced [Scope].
//
// duration is the resettable delay. limit is the maximum total wait; 0 means
// no maximum.
func (d *ScopeDebounce) Scope(duration, limit time.Duration) Scope {
	return scopeFunc(func(core core.Core) ScopeSet {
		return front.DebounceScope(d.id.Id(core), duration, limit)
	})
}

// ScopeOnlyDebounce returns one debounced [Scope].
func ScopeOnlyDebounce(duration, limit time.Duration) []Scope {
	return []Scope{(&ScopeDebounce{}).Scope(duration, limit)}
}

// ScopeFrame coordinates normal events with a barrier event.
//
// Normal members use Scope(false). The barrier member uses Scope(true), waits
// for earlier members to finish, and then runs exclusively.
type ScopeFrame struct {
	id front.AutoId
}

// Scope returns a frame member or a frame barrier depending on frame.
func (d *ScopeFrame) Scope(frame bool) Scope {
	return scopeFunc(func(core core.Core) ScopeSet {
		return front.FrameScope(d.id.Id(core), frame)
	})
}

// ScopeConcurrent allows overlap only for events that use the same group id.
type ScopeConcurrent struct {
	id front.AutoId
}

// Scope returns a concurrent [Scope] for groupId.
func (d *ScopeConcurrent) Scope(groupId int) Scope {
	return scopeFunc(func(core core.Core) ScopeSet {
		return front.ConcurrentScope(d.id.Id(core), groupId)
	})
}

// ScopeLatest keeps only the newest event in a shared scope.
// When a new event arrives, any currently processing event is canceled and the
// new event takes priority.
type ScopeLatest struct {
	id front.AutoId
}

func (b *ScopeLatest) Scope(core core.Core) ScopeSet {
	return front.LatestScope(b.id.Id(core))
}

// ScopeOnlyLatest returns a single [ScopeLatest].
func ScopeOnlyLatest() []Scope {
	return []Scope{&ScopeLatest{}}
}

type linkScope struct{}

func (b linkScope) Scope(core core.Core) ScopeSet {
	return front.LatestScope("link")
}
