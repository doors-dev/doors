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
	"log/slog"
	"reflect"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/gox"
)

// RouteModel creates a URL route for path model M and renders a writable
// source for the decoded model.
//
// The route matches when the current [Location] decodes into M. Updating the
// source re-encodes M back into the current location.
func RouteModel[M any, C gox.Comp](render func(Source[M]) C) RouteSource[Location] {
	a, err := path.GetModelAdapter[M]()
	if err != nil {
		slog.Error("Model adapter error", "error", err)
	}
	return modelSource[M, C]{
		err:     err,
		adapter: a,
		render:  render,
	}
}

// Deprecated: Use RouteModel
func RouteModelSource[M any, C gox.Comp](render func(Source[M]) C) RouteSource[Location] {
	return RouteModel(render)
}

type modelSource[M any, C gox.Comp] struct {
	err     error
	adapter path.ModelAdapter[M]
	render  func(Source[M]) C
}

func (ml modelSource[M, C]) sourceRender(l Source[Location]) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		nl := DeriveSourceEqual(l, func(l Location) M {
			m, ok := ml.adapter.Decode(l)
			if !ok {
				return *new(M)
			}
			return *m
		}, func(prev Location, m M) Location {
			loc, err := ml.adapter.Encode(&m)
			if err != nil {
				slog.Error("Model adapter encoding error", "error", err)
				return prev
			}
			return loc
		}, func(v1 M, v2 M) bool {
			return reflect.DeepEqual(v1, v2)
		})
		return cur.Comp(ml.render(nl))
	})
}

func (ml modelSource[M, C]) match(l Location) routeMatch {
	if ml.err != nil {
		return routeMatchFalse
	}
	if _, ok := ml.adapter.Decode(l); ok {
		return routeMatchTrue
	}
	return routeMatchFalse
}

var _ RouteSource[Location] = modelSource[any, gox.Comp]{}

// RouteModelBeam creates a URL route for path model M and renders a read-only
// beam for the decoded model.
//
// The route matches when the current [Location] decodes into M.
func RouteModelBeam[M any, C gox.Comp](render func(Beam[M]) C) RouteBeam[Location] {
	a, err := path.GetModelAdapter[M]()
	if err != nil {
		slog.Error("Model adapter error", "error", err)
	}
	return modelBeam[M, C]{
		err:     err,
		adapter: a,
		render:  render,
	}
}

type modelBeam[M any, C gox.Comp] struct {
	err     error
	adapter path.ModelAdapter[M]
	render  func(Beam[M]) C
}

func (ml modelBeam[M, C]) sourceRender(l Source[Location]) gox.Editor {
	return ml.beamRender(l)
}

func (ml modelBeam[M, C]) beamRender(l Beam[Location]) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		nl := DeriveBeamEqual(l, func(l Location) M {
			m, ok := ml.adapter.Decode(l)
			if !ok {
				return *new(M)
			}
			return *m
		}, func(v1 M, v2 M) bool {
			return reflect.DeepEqual(v1, v2)
		})
		return cur.Comp(ml.render(nl))
	})
}

func (ml modelBeam[M, C]) match(l Location) routeMatch {
	if ml.err != nil {
		return routeMatchFalse
	}
	if _, ok := ml.adapter.Decode(l); ok {
		return routeMatchTrue
	}
	return routeMatchFalse
}

var _ RouteBeam[Location] = modelBeam[any, gox.Comp]{}

// RouteDerive starts a route builder that matches when derive returns ok and
// exposes the derived value to the route render function.
func RouteDerive[T1 any, T2 comparable](derive func(T1) (T2, bool)) DeriveRoute[T1, T2] {
	return DeriveRoute[T1, T2]{
		derive: derive,
		equal:  beam.DefaultEqual[T2],
	}
}

// RouteDeriveEqual is like [RouteDerive] with custom equality for the derived
// value. If equal is nil, every matched update propagates.
func RouteDeriveEqual[T1 any, T2 any](derive func(T1) (T2, bool), equal func(T2, T2) bool) DeriveRoute[T1, T2] {
	return DeriveRoute[T1, T2]{
		derive: derive,
		equal:  equal,
	}
}

// DeriveRoute builds routes that render a derived value from the routed value.
type DeriveRoute[T1, T2 any] struct {
	derive func(T1) (T2, bool)
	equal  func(T2, T2) bool
}

// Source creates a writable route for the derived value.
func (r DeriveRoute[T1, T2]) Source(set func(T1, T2) T1, render func(Source[T2]) gox.Elem) RouteSource[T1] {
	return deriveRouteSource[T1, T2]{
		derive: r.derive,
		set:    set,
		equal:  r.equal,
		render: render,
	}
}

// Beam creates a read-only route for the derived value.
func (r DeriveRoute[T1, T2]) Beam(render func(Beam[T2]) gox.Elem) RouteBeam[T1] {
	return deriveRouteBeam[T1, T2]{
		derive: r.derive,
		equal:  r.equal,
		render: render,
	}
}

// RouteValue starts a route builder that matches values equal to v.
func RouteValue[T comparable](v T) MatchRoute[T] {
	return MatchRoute[T]{
		pred: func(t T) bool {
			return t == v
		},
	}
}

// RouteMatch starts a route builder that matches when pred returns true and
// renders the full routed value.
func RouteMatch[T any](match func(T) bool) MatchRoute[T] {
	return MatchRoute[T]{
		pred: match,
	}
}

// MatchRoute builds routes that render the full routed value when a predicate
// matches.
type MatchRoute[T any] struct {
	pred func(T) bool
}

// Comp creates a route that renders comp when this matcher is active.
func (m MatchRoute[T]) Comp(comp gox.Comp) RouteBeam[T] {
	return matchRouteBeam[T]{
		pred: m.pred,
		render: func(Beam[T]) gox.Elem {
			return comp.Main()
		},
	}
}

// Beam creates a read-only route for the full routed value.
func (m MatchRoute[T]) Beam(render func(Beam[T]) gox.Elem) RouteBeam[T] {
	return matchRouteBeam[T]{
		pred:   m.pred,
		render: render,
	}
}

// Source creates a writable route for the full routed value.
func (m MatchRoute[T]) Source(render func(Source[T]) gox.Elem) RouteSource[T] {
	return matchRouteSource[T]{
		pred:   m.pred,
		render: render,
	}
}

// RouteDefaultBeam creates a fallback route that always matches and renders a
// read-only beam for the full routed value.
func RouteDefaultBeam[T any, C gox.Comp](render func(Beam[T]) C) RouteBeam[T] {
	return defaultRouteBeam[T](func(l Beam[T]) gox.Elem {
		return render(l).Main()
	})
}

// RouteDefaultSource creates a fallback route that always matches and renders a
// writable source for the full routed value.
func RouteDefaultSource[T any, C gox.Comp](render func(Source[T]) C) RouteSource[T] {
	return defaultRouteSource[T](func(l Source[T]) gox.Elem {
		return render(l).Main()
	})
}

// RouteDefaultComp creates a fallback route that always matches and renders
// comp.
func RouteDefaultComp[T any](comp gox.Comp) RouteBeam[T] {
	return defaultRouteBeam[T](func(b Beam[T]) gox.Elem {
		return comp.Main()
	})
}

type routeMatch int

const (
	routeMatchFalse routeMatch = iota
	routeMatchTrue
	routeMatchDefault
)

// RouteSource is a writable route branch for values of type T1.
type RouteSource[T1 any] interface {
	match(v T1) routeMatch
	sourceRender(l Source[T1]) gox.Editor
}

// RouteBeam is a read-only route branch for values of type T1.
type RouteBeam[T1 any] interface {
	RouteSource[T1]
	beamRender(l Beam[T1]) gox.Editor
}

func routeSource[T any](l Source[T], routes []RouteSource[T]) gox.EditorComp {
	return routeRender(l, routes, func(r RouteSource[T]) gox.Editor {
		return r.sourceRender(l)
	})
}

func routeBeam[T any](b Beam[T], routes []RouteBeam[T]) gox.EditorComp {
	return routeRender(b, routes, func(r RouteBeam[T]) gox.Editor {
		return r.beamRender(b)
	})
}

func routeRender[T any, R RouteSource[T]](
	source Beam[T],
	routes []R,
	render func(R) gox.Editor,
) gox.EditorComp {
	return gox.EditorCompFunc(func(cur gox.Cursor) error {
		index := -1
		defaultActive := false
		door := &Door{}
		source.Sub(cur.Context(), func(ctx context.Context, v T) bool {
			prevIndex := index
			checked := -1
			if index != -1 && !defaultActive {
				checked = index
				m := routes[index].match(v)
				if m == routeMatchDefault {
					defaultActive = true
					return false
				}
				if m == routeMatchTrue {
					defaultActive = false
					return false
				}
			}
			index = -1
			for i := range routes {
				if checked == i {
					continue
				}
				m := routes[i].match(v)
				if m == routeMatchDefault {
					index = i
					defaultActive = true
					break
				}
				if m == routeMatchTrue {
					index = i
					defaultActive = false
					break
				}
			}
			if prevIndex == index {
				return false
			}
			if index == -1 {
				door.Outer(ctx, nil)
				return false
			}
			r := routes[index]
			door.Outer(ctx, func(cur gox.Cursor) error {
				return cur.Editor(render(r))
			})
			return false
		})
		return cur.Editor(door)
	})
}

type deriveRouteSource[T1, T2 any] struct {
	derive func(T1) (T2, bool)
	set    func(T1, T2) T1
	equal  func(T2, T2) bool
	render func(Source[T2]) gox.Elem
}

func (r deriveRouteSource[T1, T2]) sourceSet(sourceV T1, v T2) T1 {
	if r.match(sourceV) == routeMatchFalse {
		return sourceV
	}
	return r.set(sourceV, v)
}

func (r deriveRouteSource[T1, T2]) sourceGet(sourceV T1) T2 {
	v, ok := r.derive(sourceV)
	if !ok {
		return *new(T2)
	}
	return v
}

func (r deriveRouteSource[T1, T2]) sourceRender(l Source[T1]) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		l := DeriveSourceEqual(l, r.sourceGet, r.sourceSet, r.equal)
		el := r.render(l)
		if el == nil {
			return nil
		}
		return el(cur)
	})
}

func (r deriveRouteSource[T1, T2]) match(v T1) routeMatch {
	if _, ok := r.derive(v); ok {
		return routeMatchTrue
	}
	return routeMatchFalse
}

var _ RouteSource[any] = deriveRouteSource[any, any]{}

type defaultRouteSource[T1 any] func(l Source[T1]) gox.Elem

func (r defaultRouteSource[T1]) sourceRender(l Source[T1]) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		el := r(l)
		if el == nil {
			return nil
		}
		return el(cur)
	})
}

func (r defaultRouteSource[T1]) match(T1) routeMatch {
	return routeMatchDefault
}

var _ RouteSource[any] = defaultRouteSource[any](nil)

type deriveRouteBeam[T1, T2 any] struct {
	derive func(T1) (T2, bool)
	equal  func(T2, T2) bool
	render func(Beam[T2]) gox.Elem
}

func (r deriveRouteBeam[T1, T2]) get(sourceV T1) T2 {
	v, ok := r.derive(sourceV)
	if !ok {
		return *new(T2)
	}
	return v
}

func (r deriveRouteBeam[T1, T2]) sourceRender(l Source[T1]) gox.Editor {
	return r.beamRender(l)
}

func (r deriveRouteBeam[T1, T2]) beamRender(l Beam[T1]) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		l := DeriveBeamEqual(l, r.get, r.equal)
		el := r.render(l)
		if el == nil {
			return nil
		}
		return el(cur)
	})
}

func (r deriveRouteBeam[T1, T2]) match(v T1) routeMatch {
	if _, ok := r.derive(v); ok {
		return routeMatchTrue
	}
	return routeMatchFalse
}

var _ RouteBeam[any] = deriveRouteBeam[any, any]{}

type defaultRouteBeam[T1 any] func(l Beam[T1]) gox.Elem

func (r defaultRouteBeam[T1]) sourceRender(l Source[T1]) gox.Editor {
	return r.beamRender(l)
}

func (r defaultRouteBeam[T1]) beamRender(l Beam[T1]) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		el := r(l)
		if el == nil {
			return nil
		}
		return el(cur)
	})
}

func (r defaultRouteBeam[T1]) match(T1) routeMatch {
	return routeMatchDefault
}

var _ RouteBeam[any] = defaultRouteBeam[any](nil)

type matchRouteBeam[T any] struct {
	pred   func(T) bool
	render func(Beam[T]) gox.Elem
}

func (r matchRouteBeam[T]) match(v T) routeMatch {
	if r.pred(v) {
		return routeMatchTrue
	}
	return routeMatchFalse
}

func (r matchRouteBeam[T]) sourceRender(b Source[T]) gox.Editor {
	return r.beamRender(b)
}

func (r matchRouteBeam[T]) beamRender(b Beam[T]) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		el := r.render(b)
		if el == nil {
			return nil
		}
		return el(cur)
	})
}

var _ RouteBeam[any] = matchRouteBeam[any]{}

type matchRouteSource[T any] struct {
	pred   func(T) bool
	render func(Source[T]) gox.Elem
}

func (r matchRouteSource[T]) match(v T) routeMatch {
	if r.pred(v) {
		return routeMatchTrue
	}
	return routeMatchFalse
}

func (r matchRouteSource[T]) sourceRender(b Source[T]) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		el := r.render(b)
		if el == nil {
			return nil
		}
		return el(cur)
	})
}

var _ RouteSource[any] = matchRouteSource[any]{}
