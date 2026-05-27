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
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/gox"
)

// Watcher receives low-level Beam update callbacks.
type Watcher[T any] = beam.Watcher[T]

// Beam is a read-only reactive value.
//
// Use a [Beam] to read, subscribe to, or derive a smaller view of a value.
// During one render/update cycle, a Door subtree observes one consistent value
// for the same beam.
type Beam[T any] interface {
	// Effect returns the current value and rerenders the closest dynamic parent
	// when the value changes.
	Effect(ctx context.Context) (T, bool)

	// Bind renders f with the current value and rerenders that fragment when
	// the value changes.
	Bind(func(T) gox.Elem) gox.EditorComp

	// RouteBeam returns a renderable component that renders the first matching
	// read-only route for this beam.
	RouteBeam(routes ...RouteBeam[T]) gox.EditorComp

	// Sub subscribes to the value stream. onValue is called immediately with the
	// current value in the same goroutine, and again on every update.
	//
	// The subscription continues until:
	//   - onValue returns true, or
	//   - a dynamic parent is unmounted.
	//
	// It returns false if the context was already canceled or does not belong to
	// an instance runtime.
	Sub(ctx context.Context, onValue func(context.Context, T) bool) bool

	// Read returns the current value without creating a subscription.
	//
	// It returns false if the context was canceled or does not belong to an
	// instance runtime.
	Read(ctx context.Context) (T, bool)

	// ReadAndSub returns the current value and then subscribes to future
	// updates. onValue is called only for subsequent updates.
	//
	// It returns false if the context was canceled or does not belong to an
	// instance runtime.
	ReadAndSub(ctx context.Context, onValue func(context.Context, T) bool) (T, bool)

	// Watch attaches a low-level watcher for separate init, update, and
	// cancellation callbacks.
	Watch(ctx context.Context, w Watcher[T]) (context.CancelFunc, bool)

	// Get returns the most recently stored value without requiring a runtime
	// context.
	//
	// Unlike [Beam.Read], Get does not participate in render-cycle consistency
	// guarantees. Use Read when consistency across the component tree matters.
	Get() T

	innerBeam() beam.Beamer[T]
}

// Source is a writable reactive value.
//
// A Source may be an original value owner created with [NewSource] or a
// derived view created with [DeriveSource]. Both forms can be read,
// subscribed to, routed, and updated. For reference types such as slices, maps,
// pointers, or mutable structs, replace the stored value instead of mutating it
// in place.
type Source[T any] interface {
	Beam[T]

	// Route returns a renderable component that renders the first matching
	// writable route for this source.
	Route(routes ...RouteSource[T]) gox.EditorComp

	// Deprecated: Use Route
	RouteSource(routes ...RouteSource[T]) gox.EditorComp

	// Update sets a new value and propagates it to subscribers and derived
	// beams through the underlying source. Any context is allowed.
	Update(context.Context, T)

	// XUpdate behaves like [Source.Update] and returns a channel that reports
	// when propagation has finished.
	//
	// The channel receives nil on successful propagation or an error if the
	// context is invalid or the instance ends before propagation finishes. Do
	// not wait on it during rendering. If you need to wait, use [Go] or your
	// own goroutine with [Free].
	XUpdate(context.Context, T) <-chan error

	// Mutate computes the next value from the current value and propagates it
	// through the underlying source. The function receives a copy of the
	// current value and must return the next value. Returning an unchanged copy
	// is a no-op when the underlying source treats the resulting parent value as
	// equal. Any context is allowed.
	Mutate(context.Context, func(T) T)

	// XMutate behaves like [Source.Mutate] and returns a channel that reports
	// when propagation has finished.
	//
	// The channel receives nil on successful propagation or an error if the
	// context is invalid or the instance ends before propagation finishes. Do
	// not wait on it during rendering. If you need to wait, use [Go] or your
	// own goroutine with [Free].
	XMutate(context.Context, func(T) T) <-chan error

	innerLens() beam.Lenser[T]
}

// NewSource creates a [Source] that uses `==` to suppress equal updates.
//
// Example:
//
//	count := doors.NewSource(0)
func NewSource[T comparable](init T) Source[T] {
	return source[T]{
		beam.NewSource(init, beam.DefaultEqual, false),
	}
}

// NewSourceEqual creates a [Source] with a custom equality function.
//
// equal should report whether new and old should be treated as equal and
// therefore not propagated. If equal is nil, every update propagates.
func NewSourceEqual[T any](init T, equal func(new T, old T) bool) Source[T] {
	return source[T]{
		beam.NewSource(init, equal, false),
	}
}

// NewSourceNoSkip creates a [Source] that uses `==` to suppress equal updates
// and lets in-progress propagation finish even if a newer update arrives.
func NewSourceNoSkip[T comparable](init T) Source[T] {
	return source[T]{
		beam.NewSource(init, beam.DefaultEqual, true),
	}
}

// NewSourceEqualNoSkip creates a [Source] with a custom equality function and
// lets in-progress propagation finish even if a newer update arrives.
//
// equal should report whether new and old should be treated as equal and
// therefore not propagated. If equal is nil, every update propagates.
func NewSourceEqualNoSkip[T any](init T, equal func(new T, old T) bool) Source[T] {
	return source[T]{
		beam.NewSource(init, equal, true),
	}
}

type source[T any] struct {
	beam.Source[T]
}

func (s source[T]) Route(routes ...RouteSource[T]) gox.EditorComp {
	return routeSource(s, routes)
}

func (s source[T]) RouteSource(routes ...RouteSource[T]) gox.EditorComp {
	return s.Route(routes...)
}

func (s source[T]) RouteBeam(routes ...RouteBeam[T]) gox.EditorComp {
	return routeBeam(s, routes)
}

func (s source[T]) Effect(ctx context.Context) (T, bool) {
	return effect(s, ctx)
}

func (s source[T]) Bind(f func(v T) gox.Elem) gox.EditorComp {
	return bind(s, f)
}

func (d source[T]) innerBeam() beam.Beamer[T] {
	return d.Source
}

func (d source[T]) innerLens() beam.Lenser[T] {
	return d.Source
}

// DeriveSource derives a writable [Source] from source and uses `==` to
// suppress equal derived values.
//
// get extracts the derived value from the source value. set receives the
// current source value and the new derived value and must return the next
// original source value.
//
// Example:
//
//	settings := doors.NewSource(Settings{Units: "metric"})
//	units := doors.DeriveSource(settings,
//		func(s Settings) string { return s.Units },
//		func(s Settings, units string) Settings {
//			s.Units = units
//			return s
//		},
//	)
func DeriveSource[T1 any, T2 comparable](source Source[T1], get func(v T1) T2, set func(v1 T1, v2 T2) T1) Source[T2] {
	return derivedSource[T1, T2]{
		beam.NewLens(source.innerLens(), get, set, beam.DefaultEqual),
	}
}

// DeriveSourceEqual derives a writable [Source] from source with a custom
// equality function for the derived value.
//
// equal should report whether new and old derived values should be treated as
// equal and therefore not propagated to this source's subscribers or derived
// beams. If equal is nil, every derived value propagates when the underlying
// source propagates.
func DeriveSourceEqual[T1 any, T2 any](source Source[T1], get func(v T1) T2, set func(v1 T1, v2 T2) T1, equal func(new T2, old T2) bool) Source[T2] {
	return derivedSource[T1, T2]{
		beam.NewLens(source.innerLens(), get, set, equal),
	}
}

type derivedSource[T1, T2 any] struct {
	beam.Lens[T1, T2]
}

func (l derivedSource[T1, T2]) Route(routes ...RouteSource[T2]) gox.EditorComp {
	return routeSource(l, routes)
}

func (l derivedSource[T1, T2]) RouteSource(routes ...RouteSource[T2]) gox.EditorComp {
	return l.Route(routes...)
}

func (l derivedSource[T1, T2]) RouteBeam(routes ...RouteBeam[T2]) gox.EditorComp {
	return routeBeam(l, routes)
}

func (l derivedSource[T1, T2]) Bind(f func(v T2) gox.Elem) gox.EditorComp {
	return bind(l, f)
}

func (l derivedSource[T1, T2]) Effect(ctx context.Context) (T2, bool) {
	return effect(l, ctx)
}

func (l derivedSource[T1, T2]) innerBeam() beam.Beamer[T2] {
	return l.Lens
}

func (l derivedSource[T1, T2]) innerLens() beam.Lenser[T2] {
	return l.Lens
}

// DeriveBeam derives a [Beam] from source and uses `==` to suppress equal
// derived values.
//
// Example:
//
//	fullName := doors.DeriveBeam(user, func(u User) string {
//		return u.FirstName + " " + u.LastName
//	})
func DeriveBeam[T1 any, T2 comparable](source Beam[T1], get func(v T1) T2) Beam[T2] {
	return derivedBeam[T1, T2]{
		beam.NewBeam(source.innerBeam(), get, beam.DefaultEqual),
	}
}

// DeriveBeamEqual derives a [Beam] from source with a custom equality function.
//
// equal should report whether new and old should be treated as equal and
// therefore not propagated. If equal is nil, every derived value propagates.
func DeriveBeamEqual[T1 any, T2 any](source Beam[T1], get func(v T1) T2, equal func(new T2, old T2) bool) Beam[T2] {
	return derivedBeam[T1, T2]{
		beam.NewBeam(source.innerBeam(), get, equal),
	}
}

type derivedBeam[T1, T2 any] struct {
	beam.Beam[T1, T2]
}

func (b derivedBeam[T1, T2]) RouteBeam(routes ...RouteBeam[T2]) gox.EditorComp {
	return routeBeam(b, routes)
}

func (b derivedBeam[T1, T2]) Bind(f func(v T2) gox.Elem) gox.EditorComp {
	return bind(b, f)
}

func (b derivedBeam[T1, T2]) Effect(ctx context.Context) (T2, bool) {
	return effect(b, ctx)
}

func (d derivedBeam[T1, T2]) innerBeam() beam.Beamer[T2] {
	return d.Beam
}

func effect[T any](b Beam[T], ctx context.Context) (T, bool) {
	return b.ReadAndSub(ctx, func(ctx context.Context, _ T) bool {
		ctx.Value(common.KeyCore).(core.Core).Door().Reload(ctx)
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
