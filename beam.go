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

// Watcher is the value-and-cancellation callback pair passed to [Beam.Watch].
type Watcher[T any] = beam.Watcher[T]

// Beam is a read-only reactive value.
//
// Use it to read, subscribe to, or derive a narrower view of the value. Within
// one render/update cycle, a Door subtree observes one consistent value for
// the same beam.
type Beam[T any] interface {
	// Effect returns the current value and rerenders the closest dynamic parent
	// when the value changes. Prefer [Beam.Bind] when a smaller fragment should
	// rerender.
	//
	// It returns false if the context was canceled or does not belong to an
	// instance runtime.
	Effect(ctx context.Context) (T, bool)

	// Bind renders f with the current value and rerenders only that fragment
	// when the value changes. If f returns nil, the fragment renders nothing.
	Bind(func(T) gox.Elem) gox.EditorComp

	// RouteBeam renders the first matching read-only route for the current
	// value. The fragment swaps only when the matching route changes.
	RouteBeam(routes ...RouteBeam[T]) gox.EditorComp

	// Sub subscribes to the value stream. onValue is called immediately with the
	// current value on the calling goroutine, then again on every update.
	//
	// The subscription ends when onValue returns true or the content that
	// registered it is cleared. From a context outside an instance, for
	// example [SessionContext] or a background context, it ends when the
	// context is canceled, detected on the next propagation.
	//
	// It returns false if the context was canceled or the instance is shut
	// down.
	Sub(ctx context.Context, onValue func(context.Context, T) bool) bool

	// Read returns the current value without subscribing to updates. Any
	// context is allowed.
	//
	// It returns false if the context was canceled or the instance is shut
	// down.
	Read(ctx context.Context) (T, bool)

	// ReadAndSub returns the current value and subscribes to later updates.
	// Unlike [Beam.Sub], onValue is not called with the current value. If
	// onValue is nil, no subscription is created. The subscription lives as
	// described in [Beam.Sub].
	//
	// It returns false if the context was canceled or the instance is shut
	// down.
	ReadAndSub(ctx context.Context, onValue func(context.Context, T) bool) (T, bool)

	// Watch subscribes w to the value stream and returns a function that
	// cancels the subscription. The subscription lives as described in
	// [Beam.Sub]. It returns false if the context was canceled or the
	// instance is shut down.
	//
	// w receives the current value on the calling goroutine, then every update.
	// Prefer [Beam.Sub] or [Beam.ReadAndSub] unless you need to cancel the
	// subscription explicitly.
	Watch(ctx context.Context, w Watcher[T]) (context.CancelFunc, bool)

	// Get returns the latest stored value.
	//
	// Unlike [Beam.Read], Get needs no context and does not participate in
	// render cycle consistency. Use Read when the value must agree with the
	// rest of the render.
	Get() T

	innerBeam() beam.Beamer[T]
}

// Source is a writable reactive value.
//
// A Source is either an original source created with [NewSource] or a derived
// source created with [DeriveSource]; both can be read, subscribed to, routed,
// and updated. For slices, maps, pointers, and mutable structs, store a
// replacement instead of mutating the stored value in place.
type Source[T any] interface {
	Beam[T]

	// Route renders the first matching writable route for the current value.
	// The fragment swaps only when the matching route changes. Unlike
	// [Beam.RouteBeam], routes receive a writable [Source].
	Route(routes ...RouteSource[T]) gox.EditorComp

	// Update sets a new value and propagates it to subscribers and derived beams
	// through the underlying source. Any context is allowed.
	//
	// The returned channel is optional to use. It receives nil when propagation
	// completes, or context.Canceled when a newer update supersedes it, then
	// closes. It closes without a value when no propagation happens: the update
	// is suppressed as equal or there are no subscribers. Do not wait on it
	// during rendering. If you need to wait, use [Go] or your own goroutine with
	// [DetachedContext].
	Update(context.Context, T) <-chan error

	// Mutate computes the next value from the current value and propagates it
	// through the underlying source. The function receives a copy of the current
	// value and must return the next value. It may run more than once when a
	// concurrent update lands first, so keep it free of side effects. Any
	// context is allowed.
	//
	// The returned channel is optional to use; see [Source.Update] for the
	// contract.
	Mutate(context.Context, func(T) T) <-chan error

	innerLens() beam.Lenser[T]
}

// NewSource returns a writable source holding init. It uses == to suppress
// equal updates.
func NewSource[T comparable](init T) Source[T] {
	return source[T]{
		beam.NewSource(init, beam.DefaultEqual, false),
	}
}

// NewSourceEqual returns a writable source holding init. It calls equal to
// suppress equal updates.
//
// equal reports whether new and old should be treated as equal and therefore
// not propagated. If equal is nil, every update propagates. equal runs while
// the source's lock is held: it must not panic and must not call back into any
// source or beam.
func NewSourceEqual[T any](init T, equal func(new T, old T) bool) Source[T] {
	return source[T]{
		beam.NewSource(init, equal, false),
	}
}

// NewSourceNoSkip returns a writable source holding init. It uses == to
// suppress equal updates and lets in-progress propagation finish even if a
// newer update arrives.
func NewSourceNoSkip[T comparable](init T) Source[T] {
	return source[T]{
		beam.NewSource(init, beam.DefaultEqual, true),
	}
}

// NewSourceEqualNoSkip returns a writable source holding init. It calls equal
// to suppress equal updates and lets in-progress propagation finish even if a
// newer update arrives.
//
// The equal contract follows [NewSourceEqual].
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

// DeriveSource returns a writable derived source of source. It uses == to
// suppress equal derived values.
//
// get extracts the derived value from the source value. set receives the
// current source value and the new derived value and must return the next
// source value.
//
// Example:
//
//	settings := doors.NewSource(Settings{Units: "metric"})
//	units := doors.DeriveSource(settings,
//		func(s Settings) string { return s.Units },
//		func(s Settings, u string) Settings { s.Units = u; return s },
//	)
func DeriveSource[T1 any, T2 comparable](source Source[T1], get func(v T1) T2, set func(v1 T1, v2 T2) T1) Source[T2] {
	return derivedSource[T1, T2]{
		beam.NewLens(source.innerLens(), get, set, beam.DefaultEqual),
	}
}

// DeriveSourceEqual returns a writable derived source of source. It calls
// equal to suppress equal derived values. If equal is nil, every derived value
// propagates.
//
// get and set follow [DeriveSource].
//
// equal runs while the derived source's lock is held: it must not panic and
// must not call back into any source or beam.
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

// DeriveBeam returns a derived beam of source, computed with get. It uses ==
// to suppress equal derived values.
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

// DeriveBeamEqual returns a derived beam of source, computed with get. It calls
// equal to suppress equal derived values. If equal is nil, every derived value
// propagates.
//
// equal runs while the derived beam's lock is held: it must not panic and must
// not call back into any source or beam.
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
	if _, ok := ctx.Value(common.KeyCore).(core.Core); !ok {
		var zero T
		return zero, false
	}
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
