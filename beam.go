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
	"context"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/gox"
)

// Watcher receives low-level Beam update callbacks.
type Watcher[T any] = beam.Watcher[T]

// Beam is a read-only reactive value.
//
// Use a [Beam] to read, subscribe to, or derive a smaller view of state. During
// one render/update cycle, a Door subtree observes one consistent value for the
// same beam.
type Beam[T any] interface {
	// Effect returns the current value and rerenders the closest dynamic parent
	// when the value changes.
	Effect(ctx context.Context) (T, bool)

	Bind(func(T) gox.Elem) gox.EditorComp

	// Sub subscribes to the value stream. onValue is called immediately with the
	// current value in the same goroutine, and again on every update.
	// Only instance runtime context is allowed.
	//
	// The subscription continues until:
	//   - onValue returns true, or
	//   - a dynamic parent is unmounted.
	//
	// It returns false if the context was already canceled or does not belong to
	// an instance runtime.
	Sub(ctx context.Context, onValue func(context.Context, T) bool) bool

	// Read returns the current value without creating a subscription.
	// Only instance runtime context is allowed.
	//
	// It returns false if the context was canceled or does not belong to an
	// instance runtime.
	Read(ctx context.Context) (T, bool)

	// ReadAndSub returns the current value and then subscribes to future
	// updates. onValue is called only for subsequent updates.
	// Only instance runtime context is allowed.
	//
	// It returns false if the context was canceled or does not belong to an
	// instance runtime.
	ReadAndSub(ctx context.Context, onValue func(context.Context, T) bool) (T, bool)

	// Watch attaches a low-level watcher for separate init, update, and
	// cancellation callbacks. Only instance runtime context is allowed.
	Watch(ctx context.Context, w Watcher[T]) (context.CancelFunc, bool)

	inner() beam.Beam[T]
}

// Source is a writable [Beam].
//
// Create a source with [NewSource] or [NewSourceEqual], then derive smaller
// beams with [NewBeam] or [NewBeamEqual]. For reference types such as slices,
// maps, pointers, or mutable structs, replace the stored value instead of
// mutating it in place.
type Source[T any] interface {
	Beam[T]

	// Update sets a new value and propagates it to subscribers and derived
	// beams. The update is applied only if it passes the source's distinct
	// function. Any context is allowed.
	Update(context.Context, T)

	// XUpdate behaves like [Source.Update] and returns a channel that reports
	// when propagation has finished.
	//
	// The channel receives nil on successful propagation or an error if the
	// context is invalid or the instance ends before propagation finishes. Do
	// not wait on it during rendering. If you need to wait, do it in a hook,
	// inside `doors.Go(...)`, or in your own goroutine with `doors.Free(ctx)`.
	XUpdate(context.Context, T) <-chan error

	// Mutate computes the next value from the current value and propagates it
	// if it passes the source's distinct function. The function receives a copy
	// of the current value and must return the next value. Returning an
	// unchanged copy is a no-op when a distinct function is in use.
	// Any context is allowed.
	Mutate(context.Context, func(T) T)

	// XMutate behaves like [Source.Mutate] and returns a channel that reports
	// when propagation has finished.
	//
	// The channel receives nil on successful propagation or an error if the
	// context is invalid or the instance ends before propagation finishes. Do
	// not wait on it during rendering. If you need to wait, do it in a hook,
	// inside `doors.Go(...)`, or in your own goroutine with `doors.Free(ctx)`.
	XMutate(context.Context, func(T) T) <-chan error

	// Get returns the most recently stored value without requiring a runtime
	// context.
	//
	// Unlike [Beam.Read], Get does not participate in render-cycle consistency
	// guarantees. Use Read when consistency across the component tree matters.
	Get() T

	// DisableSkipping forces every committed value to propagate, even if newer
	// values arrive before earlier updates finish syncing. This is useful when a
	// source is used as a communication channel and every message matters.
	DisableSkipping()
}

// NewSource creates a [Source] that uses `==` to suppress equal updates.
//
// Example:
//
//	count := doors.NewSource(0)
func NewSource[T comparable](init T) Source[T] {
	return sourceBeam[T]{
		beam.NewSource(init),
	}
}

// NewSourceEqual creates a [Source] with a custom equality function.
//
// equal should report whether new and old should be treated as equal and
// therefore not propagated. If equal is nil, every update propagates.
func NewSourceEqual[T any](init T, equal func(new T, old T) bool) Source[T] {
	return sourceBeam[T]{
		beam.NewSourceEqual(init, equal),
	}
}

type sourceBeam[T any] struct {
	*beam.SourceBeam[T]
}

func (s sourceBeam[T]) Effect(ctx context.Context) (T, bool) {
	return effect(s, ctx)
}

func (s sourceBeam[T]) Bind(f func(T) gox.Elem) gox.EditorComp {
	return bind(s, f)
}

func (d sourceBeam[T]) inner() beam.Beam[T] {
	return d.SourceBeam
}

// NewBeam derives a [Beam] from source and uses `==` to suppress equal
// derived values.
//
// Example:
//
//	fullName := doors.NewBeam(user, func(u User) string {
//		return u.FirstName + " " + u.LastName
//	})
func NewBeam[T1 any, T2 comparable](source Beam[T1], cast func(T1) T2) Beam[T2] {
	return derivedBeam[T1, T2]{
		beam.NewBeam(source.inner(), cast),
	}
}

// NewBeamEqual derives a [Beam] from source with a custom equality function.
//
// equal should report whether new and old should be treated as equal and
// therefore not propagated. If equal is nil, every derived value propagates.
func NewBeamEqual[T1 any, T2 any](source Beam[T1], cast func(T1) T2, equal func(new T2, old T2) bool) Beam[T2] {
	return derivedBeam[T1, T2]{
		beam.NewBeamEqual(source.inner(), cast, equal),
	}
}

type derivedBeam[T1, T2 any] struct {
	*beam.DerivedBeam[T1, T2]
}

func (b derivedBeam[T1, T2]) Bind(f func(T2) gox.Elem) gox.EditorComp {
	return bind(b, f)
}

func (b derivedBeam[T1, T2]) Effect(ctx context.Context) (T2, bool) {
	return effect(b, ctx)
}

func (d derivedBeam[T1, T2]) inner() beam.Beam[T2] {
	return d.DerivedBeam
}

func effect[T any](b Beam[T], ctx context.Context) (T, bool) {
	return b.ReadAndSub(ctx, func(ctx context.Context, _ T) bool {
		ctx.Value(ctex.KeyCore).(core.Core).Reload(ctx)
		return true
	})
}

func bind[T any](b Beam[T], f func(T) gox.Elem) gox.EditorComp {
	return gox.EditorCompFunc(func(cur gox.Cursor) error {
		door := &Door{}
		ok := b.Sub(cur.Context(), func(ctx context.Context, v T) bool {
			door.Outer(ctx, gox.Elem(func(cur gox.Cursor) error {
				el := f(v)
				if el == nil {
					return nil
				}
				return el(cur)
			}))
			return false
		})
		if !ok {
			return nil
		}
		return cur.Editor(door)
	})
}
