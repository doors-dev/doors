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
	"time"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
)

// Scope controls event processing concurrency by managing how multiple events
// are queued, debounced, blocked, or serialized. Scopes provide fine-grained
// control over event handling behavior to prevent race conditions and improve UX.
type Scope = front.Scope

// ScopeSet represents the configuration data for a specific scope instance.
// It contains the scope type, ID, and any additional parameters needed for execution.
type ScopeSet = front.ScopeSet

type scopeFunc func(inst instance.Core) *ScopeSet

func (s scopeFunc) Scope(inst instance.Core) *ScopeSet {
	return s(inst)
}

// BlockingScope prevents concurrent event processing within the same scope.
// When an event is already being processed, subsequent events are cancelled
// until the current event completes. This is useful for preventing double-clicks
// or rapid form submissions.
type BlockingScope struct {
	id front.ScopeAutoId
}

// ScopeBlocking creates a blocking scope that cancels subsequent events while
// one is already processing. Use this to prevent duplicate operations like
// double form submissions or multiple API calls from rapid clicking.
func ScopeBlocking() []Scope {
	return []Scope{&BlockingScope{}}
}

func (b *BlockingScope) Scope(inst instance.Core) *ScopeSet {
	return front.BlockingScope(b.id.Id(inst))
}

// SerialScope processes events one at a time in the order they were received.
// Events are queued and processed sequentially.
type SerialScope struct {
	id front.ScopeAutoId
}

func (b *SerialScope) Scope(inst instance.Core) *ScopeSet {
	return front.SerialScope(b.id.Id(inst))
}

// ScopeSerial creates a serial scope that processes events one at a time in order.
// Events are queued and executed sequentially.
func ScopeSerial() []Scope {
	return []Scope{&SerialScope{}}
}

// LatestScope cancels previous events and only processes the most recent one.
// When a new event arrives, any currently processing event is cancelled
// and the new event takes priority. This is useful for search-as-you-type
// or real-time filtering scenarios.
type LatestScope struct {
	id front.ScopeAutoId
}

func (b *LatestScope) Scope(inst instance.Core) *ScopeSet {
	return front.LatestScope(b.id.Id(inst))
}

// ScopeLatest creates a scope that only processes the most recent event,
// cancelling any previous events that are still processing. This ensures
// only the latest user action is processed.
func ScopeLatest() []Scope {
	return []Scope{&LatestScope{}}
}

// DebounceScope delays event processing using a debounce mechanism with both
// duration and limit parameters. Events are delayed by the duration, but
// will always execute within the limit timeframe regardless of new events.
type DebounceScope struct {
	id front.ScopeAutoId
}

// Scope creates a debounced scope with the specified timing parameters.
// The duration parameter sets the debounce delay - events are delayed by this amount
// and reset if new events arrive. The limit parameter sets the maximum time an event
// can be delayed - events will execute after this time regardless of new events.
//
// Parameters:
//   - duration: Debounce delay time (resets on new events)
//   - limit: Maximum delay time (executes regardless of new events)
func (d *DebounceScope) Scope(duration time.Duration, limit time.Duration) Scope {
	return scopeFunc(func(inst instance.Core) *ScopeSet {
		return front.DebounceScope(d.id.Id(inst), duration, limit)
	})
}

// ScopeDebounce creates a debounced scope that delays event execution.
// Events are delayed by duration, but will always execute within limit time.
// This is useful for preventing excessive API calls during rapid user input.
//
// Parameters:
//   - duration: How long to wait after the last event before executing
//   - limit: Maximum time to wait before forcing execution
func ScopeDebounce(duration time.Duration, limit time.Duration) []Scope {
	return []Scope{(&DebounceScope{}).Scope(duration, limit)}
}



// FrameScope manages two types of events: immediate events and frame events.
// Immediate events (frame=false) executed normaly. 
// Frame events (frame=true) wait until all previous events in the scope complete,
// while blocking new events, and then execute normaly.
type FrameScope struct {
	id front.ScopeAutoId
}

// Scope creates a frame-based scope with the specified event type.
// Immediate events (frame=false) executed normaly. 
// Frame events (frame=true) wait until all previous events in the scope complete,
// while blocking new events (frame=true and frame=false), and then execute normaly.
//
// Parameters:
//   - frame: false for immediate execution, true to wait for other events to complete
func (d *FrameScope) Scope(frame bool) Scope {
	return scopeFunc(func(inst instance.Core) *ScopeSet {
		return front.FrameScope(d.id.Id(inst), frame)
	})
}
