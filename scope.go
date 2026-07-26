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

// Scope is one client-side scheduling policy applied to a hook request.
//
// Values come from the [Scopes] types, such as [ScopeBlocking], [ScopeDebounce],
// or [ScopeLatest].
type Scope = front.Scope

// Scopes is a composable list of client-side scheduling policies for hook
// requests.
type Scopes interface {
	Scopes(core core.Core) []Scope
	Joiner[Scopes]
}

func scopesOrNil(core core.Core, scopes Scopes) []Scope {
	if scopes == nil {
		return nil
	}
	return scopes.Scopes(core)
}

type joinedScopes []Scopes

func (ss joinedScopes) And(s Scopes) Scopes {
	c := make(joinedScopes, len(ss), len(ss)+1)
	copy(c, ss)
	c = append(c, s)
	return c
}

func (ss joinedScopes) Scopes(core core.Core) []Scope {
	output := make([]Scope, 0)
	for _, ind := range ss {
		if ind == nil {
			continue
		}
		output = append(output, ind.Scopes(core)...)
	}
	return output
}

// ScopeBlocking rejects a hook request while an earlier request in the same
// scope has not finished.
//
// Each value is one scope: share one value between attrs to make them
// coordinate, or use separate values to keep them independent.
//
// A rejected request is never sent, so its handler does not run and its
// indication does not start.
type ScopeBlocking struct {
	id front.AutoID
}

func (sb *ScopeBlocking) And(s Scopes) Scopes {
	return joinedScopes([]Scopes{sb, s})
}

func (sb *ScopeBlocking) Scopes(core core.Core) []Scope {
	return []Scope{front.BlockingScope(sb.id.String(core))}
}

var _ Scopes = (*ScopeBlocking)(nil)

// ScopeSerial queues hook requests in the same scope and runs them one at a
// time, in arrival order.
//
// Unlike [ScopeBlocking], a later request waits instead of being rejected.
// One value is one scope; see [ScopeBlocking] for sharing.
type ScopeSerial struct {
	id front.AutoID
}

func (ss *ScopeSerial) And(s Scopes) Scopes {
	return joinedScopes([]Scopes{ss, s})
}

func (ss *ScopeSerial) Scopes(core core.Core) []Scope {
	return []Scope{front.SerialScope(ss.id.String(core))}
}

var _ Scopes = (*ScopeSerial)(nil)

// ScopeRate enforces a minimum interval between hook requests in the same
// scope.
//
// Unlike [ScopeDebounce], the first request is not delayed. While the interval
// is open only the newest pending request is kept, and the ones it replaces are
// rejected.
//
// One value is one scope; see [ScopeBlocking] for sharing.
type ScopeRate struct {
	// Tick is the minimum interval between two requests. Required.
	Tick time.Duration
	id   front.AutoID
}

func (sd *ScopeRate) And(s Scopes) Scopes {
	return joinedScopes([]Scopes{sd, s})
}

func (s *ScopeRate) Scopes(core core.Core) []Scope {
	return []Scope{front.RateScope(s.id.String(core), s.Tick)}
}

var _ Scopes = (*ScopeRate)(nil)

// ScopeDebounce delays a hook request until events stop arriving.
//
// A new event in the scope replaces the pending request, and the replaced one
// is rejected without being sent.
//
// One value is one scope; see [ScopeBlocking] for sharing.
type ScopeDebounce struct {
	// Duration is the quiet period before the pending request is sent. Every
	// new event restarts it. Required.
	Duration time.Duration
	// Limit caps the wait, measured from the first pending event. Optional;
	// zero means no cap.
	Limit time.Duration
	id    front.AutoID
}

func (sd *ScopeDebounce) And(s Scopes) Scopes {
	return joinedScopes([]Scopes{sd, s})
}

func (s *ScopeDebounce) Scopes(core core.Core) []Scope {
	return []Scope{front.DebounceScope(s.id.String(core), s.Duration, s.Limit)}
}

var _ Scopes = (*ScopeDebounce)(nil)

// ScopeFrame separates ordinary hook requests in a scope from a barrier
// request that runs alone.
//
// It does not satisfy [Scopes]; call Scope to obtain a value.
//
// One value is one scope; see [ScopeBlocking] for sharing.
type ScopeFrame struct {
	id front.AutoID
}

// Scope selects how a request takes part in this scope. With frame true the
// request is the barrier: it waits for the requests already running in the
// scope, then runs while every further request in the scope is rejected. With
// frame false it is an ordinary member, rejected while a barrier is pending or
// running.
func (d *ScopeFrame) Scope(frame bool) Scopes {
	return scopeFunc(func(core core.Core) []Scope {
		return []Scope{front.FrameScope(d.id.String(core), frame)}
	})
}

// ScopeConcurrent lets hook requests in the same scope overlap only when they
// share a group.
//
// It does not satisfy [Scopes]; call Scope to obtain a value.
//
// One value is one scope; see [ScopeBlocking] for sharing.
type ScopeConcurrent struct {
	id front.AutoID
}

// Scope assigns a request to the group identified by groupID. While the scope
// holds requests of one group, a request of any other group is rejected.
func (d *ScopeConcurrent) Scope(groupID int) Scopes {
	return scopeFunc(func(core core.Core) []Scope {
		return []Scope{front.ConcurrentScope(d.id.String(core), groupID)}
	})
}

// ScopeLatest accepts a hook request immediately and cancels the previous one
// in the same scope.
//
// Cancellation is client-side and best-effort: a canceled request may already
// have reached the server and run its handler. Prefer [ScopeDebounce] when the
// request must not be sent at all.
//
// One value is one scope; see [ScopeBlocking] for sharing.
type ScopeLatest struct {
	id front.AutoID
}

func (sl *ScopeLatest) And(s Scopes) Scopes {
	return joinedScopes([]Scopes{sl, s})
}

func (s *ScopeLatest) Scopes(core core.Core) []Scope {
	return []Scope{front.LatestScope(s.id.String(core))}
}

var _ Scopes = (*ScopeLatest)(nil)

type scopeFunc func(core core.Core) []Scope

func (sf scopeFunc) And(s Scopes) Scopes {
	return joinedScopes([]Scopes{sf, s})
}

func (sf scopeFunc) Scopes(core core.Core) []Scope {
	return sf(core)
}

var _ Scopes = scopeFunc(nil)

type linkScope struct{}

func (ls linkScope) And(s Scopes) Scopes {
	return joinedScopes([]Scopes{ls, s})
}

func (ls linkScope) Scopes(core core.Core) []Scope {
	return []Scope{front.LatestScope("link")}
}

var _ Scopes = linkScope{}
