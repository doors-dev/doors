// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package doors

import (
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"time"
)

// Scope controls concurrency for event processing.
// It defines how events are queued, blocked, debounced, or serialized.
type Scope = front.Scope

// ScopeSet holds the configuration of a scope instance.
type ScopeSet = front.ScopeSet

type scopeFunc func(inst instance.Core) *ScopeSet

func (s scopeFunc) Scope(inst instance.Core) *ScopeSet {
	return s(inst)
}

// ScopeBlocking cancels new events while one is processing.
// Useful for preventing double-clicks or duplicate submissions.
type ScopeBlocking struct {
	id front.ScopeAutoId
}

func (b *ScopeBlocking) Scope(inst instance.Core) *ScopeSet {
	return front.BlockingScope(b.id.Id(inst))
}

// ScopeOnlyBlocking creates a blocking scope that cancels concurrent events.
func ScopeOnlyBlocking() []Scope {
	return []Scope{&ScopeBlocking{}}
}

// ScopeSerial queues events and processes them in order.
type ScopeSerial struct {
	id front.ScopeAutoId
}

func (b *ScopeSerial) Scope(inst instance.Core) *ScopeSet {
	return front.SerialScope(b.id.Id(inst))
}

// ScopeOnlySerial creates a serial scope that executes events sequentially.
func ScopeOnlySerial() []Scope {
	return []Scope{&ScopeSerial{}}
}


// ScopeDebounce delays events by duration but guarantees execution
// within the specified limit. New events reset the delay.
type ScopeDebounce struct {
	id front.ScopeAutoId
}

// Scope creates a debounced scope.
//   - duration: debounce delay, reset by new events
//   - limit: maximum wait before execution regardless of new events
func (d *ScopeDebounce) Scope(duration, limit time.Duration) Scope {
	return scopeFunc(func(inst instance.Core) *ScopeSet {
		return front.DebounceScope(d.id.Id(inst), duration, limit)
	})
}

// ScopeOnlyDebounce creates a debounced scope with duration and limit.
func ScopeOnlyDebounce(duration, limit time.Duration) []Scope {
	return []Scope{(&ScopeDebounce{}).Scope(duration, limit)}
}

// ScopeFrame distinguishes immediate and frame events.
// Immediate events run normally. Frame events wait for all prior
// events to finish, block new ones, then run exclusively.
type ScopeFrame struct {
	id front.ScopeAutoId
}

// Scope creates a frame-based scope.
//   - frame=false: execute immediately
//   - frame=true: wait for completion of all events, then execute exclusively
func (d *ScopeFrame) Scope(frame bool) Scope {
	return scopeFunc(func(inst instance.Core) *ScopeSet {
		return front.FrameScope(d.id.Id(inst), frame)
	})
}

// ScopeConcurrent can be occupied by events with the same
// groupId, other - blocked
type ScopeConcurrent struct {
	id front.ScopeAutoId
}

func (d *ScopeConcurrent) Scope(groupId int) Scope {
	return scopeFunc(func(inst instance.Core) *ScopeSet {
		return front.ConcurrentScope(d.id.Id(inst), groupId)
	})
}

