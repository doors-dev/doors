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

// Scope is the low-level client scheduling scope used by event hooks.
type Scope = front.Scope

// Scopes configures how hook requests are scheduled and deduplicated.
type Scopes interface {
	// Scopes returns the low-level client scopes for this app core.
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

// ScopeBlocking allows only one request in this scope to run at a time.
//
// Later requests are rejected while an earlier request is active.
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

// ScopeSerial queues requests in this scope and runs them one after another.
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

// ScopeRate enforces a minimum interval between requests in this scope.
//
// The first event fires immediately. Subsequent events are throttled so
// that no two requests start closer together than Tick.
type ScopeRate struct {
	// Tick is the minimum interval between events.
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

// ScopeDebounce delays requests until input settles.
type ScopeDebounce struct {
	// Duration is the quiet period before a request is sent.
	Duration time.Duration
	// Limit is the maximum delay before a pending request must be sent.
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

// ScopeFrame groups requests by frame lifecycle.
type ScopeFrame struct {
	id front.AutoID
}

// Scope returns a scope for either the frame or non-frame group.
func (d *ScopeFrame) Scope(frame bool) Scopes {
	return scopeFunc(func(core core.Core) []Scope {
		return []Scope{front.FrameScope(d.id.String(core), frame)}
	})
}

// ScopeConcurrent allows one request per group to run concurrently.
type ScopeConcurrent struct {
	id front.AutoID
}

// Scope returns a concurrent scheduling group.
func (d *ScopeConcurrent) Scope(groupID int) Scopes {
	return scopeFunc(func(core core.Core) []Scope {
		return []Scope{front.ConcurrentScope(d.id.String(core), groupID)}
	})
}

// ScopeLatest keeps only the latest request in this scope.
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
