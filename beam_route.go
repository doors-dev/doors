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

// RouteModel returns a URL route that matches when the current [Location]
// decodes into path model M and renders with a writable [Source] of M.
//
// Updating that source re-encodes M into the URL.
func RouteModel[M any, C gox.Comp](render func(s Source[M]) C) RouteSource[Location] {
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

// RouteModelBeam returns a URL route that matches when the current [Location]
// decodes into path model M and renders with a read-only [Beam] of M.
//
// Unlike [RouteModel], the model cannot be written back to the URL.
func RouteModelBeam[M any, C gox.Comp](render func(b Beam[M]) C) RouteBeam[Location] {
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

// RouteDerive returns a route builder that matches when derive reports true
// and exposes the derived value to the branch. It uses == to suppress
// propagation of equal derived values.
func RouteDerive[T1 any, T2 comparable](derive func(v T1) (T2, bool)) DeriveRoute[T1, T2] {
	return DeriveRoute[T1, T2]{
		derive: derive,
		equal:  beam.DefaultEqual[T2],
	}
}

// RouteDeriveEqual returns a route builder that matches when derive reports
// true and exposes the derived value to the branch. It uses equal to suppress
// propagation; if equal is nil, every update propagates.
func RouteDeriveEqual[T1 any, T2 any](derive func(v T1) (T2, bool), equal func(v1 T2, v2 T2) bool) DeriveRoute[T1, T2] {
	return DeriveRoute[T1, T2]{
		derive: derive,
		equal:  equal,
	}
}

// DeriveRoute builds route branches that receive a value derived from the
// routed value.
type DeriveRoute[T1, T2 any] struct {
	derive func(T1) (T2, bool)
	equal  func(T2, T2) bool
}

// Source returns a writable branch for the derived value. set maps an updated
// derived value back onto the routed value; updates made while the branch does
// not match leave the routed value unchanged.
func (r DeriveRoute[T1, T2]) Source(set func(v1 T1, v2 T2) T1, render func(s Source[T2]) gox.Elem) RouteSource[T1] {
	return deriveRouteSource[T1, T2]{
		derive: r.derive,
		set:    set,
		equal:  r.equal,
		render: render,
	}
}

// Beam returns a read-only branch that renders with a [Beam] of the derived
// value. render runs when the branch becomes active; later value changes reach
// it through the beam.
func (r DeriveRoute[T1, T2]) Beam(render func(b Beam[T2]) gox.Elem) RouteBeam[T1] {
	return deriveRouteBeam[T1, T2]{
		derive: r.derive,
		equal:  r.equal,
		render: render,
	}
}

// Bind returns a read-only branch that renders with the derived value itself.
// Unlike [DeriveRoute.Beam], render runs again on every change of that value.
func (r DeriveRoute[T1, T2]) Bind(render func(v T2) gox.Elem) RouteBeam[T1] {
	return deriveRouteBeam[T1, T2]{
		derive: r.derive,
		equal:  r.equal,
		render: func(b Beam[T2]) gox.Elem {
			return b.Bind(render).Main()
		},
	}
}

// RouteValue returns a route builder that matches routed values equal to v.
func RouteValue[T comparable](v T) MatchRoute[T] {
	return MatchRoute[T]{
		pred: func(t T) bool {
			return t == v
		},
	}
}

// RouteMatch returns a route builder that matches when match reports true and
// exposes the whole routed value to the branch.
func RouteMatch[T any](match func(v T) bool) MatchRoute[T] {
	return MatchRoute[T]{
		pred: match,
	}
}

// MatchRoute builds route branches that receive the whole routed value.
type MatchRoute[T any] struct {
	pred func(T) bool
}

// Comp returns a read-only branch that renders comp, which receives no value.
func (m MatchRoute[T]) Comp(comp gox.Comp) RouteBeam[T] {
	return matchRouteBeam[T]{
		pred: m.pred,
		render: func(Beam[T]) gox.Elem {
			return comp.Main()
		},
	}
}

// Beam returns a read-only branch that renders with a [Beam] of the routed
// value. render runs when the branch becomes active; later value changes reach
// it through the beam.
func (m MatchRoute[T]) Beam(render func(b Beam[T]) gox.Elem) RouteBeam[T] {
	return matchRouteBeam[T]{
		pred:   m.pred,
		render: render,
	}
}

// Bind returns a read-only branch that renders with the routed value itself.
// Unlike [MatchRoute.Beam], render runs again on every change of that value.
func (m MatchRoute[T]) Bind(render func(v T) gox.Elem) RouteBeam[T] {
	return matchRouteBeam[T]{
		pred: m.pred,
		render: func(b Beam[T]) gox.Elem {
			return b.Bind(render).Main()
		},
	}
}

// Source returns a writable branch for the routed value. Unlike
// [MatchRoute.Beam], it is not accepted by [Beam.RouteBeam].
func (m MatchRoute[T]) Source(render func(s Source[T]) gox.Elem) RouteSource[T] {
	return matchRouteSource[T]{
		pred:   m.pred,
		render: render,
	}
}

// RouteDefault returns a fallback route that always matches and renders with a
// writable [Source] of the routed value. Place it last; routes after it are
// never tried.
func RouteDefault[T any, C gox.Comp](render func(s Source[T]) C) RouteSource[T] {
	return defaultRouteSource[T](func(l Source[T]) gox.Elem {
		var comp gox.Comp = render(l)
		if comp == nil {
			return nil
		}
		return comp.Main()
	})
}

// RouteDefaultBeam returns a fallback route that always matches and renders
// with a read-only [Beam] of the routed value. Place it last; routes after it
// are never tried.
func RouteDefaultBeam[T any, C gox.Comp](render func(b Beam[T]) C) RouteBeam[T] {
	return defaultRouteBeam[T](func(l Beam[T]) gox.Elem {
		var comp gox.Comp = render(l)
		if comp == nil {
			return nil
		}
		return comp.Main()
	})
}

// RouteDefaultBind returns a fallback route that always matches and renders
// with the routed value itself, rerendering on every change. Place it last;
// routes after it are never tried.
func RouteDefaultBind[T any, C gox.Comp](render func(v T) C) RouteBeam[T] {
	return defaultRouteBeam[T](func(b Beam[T]) gox.Elem {
		return b.Bind(func(t T) gox.Elem {
			var comp gox.Comp = render(t)
			if comp == nil {
				return nil
			}
			el := comp.Main()
			if el == nil {
				return nil
			}
			return el
		}).Main()
	})
}

// RouteDefaultComp returns a fallback route that always matches and renders
// comp, which receives no value. Place it last; routes after it are never
// tried.
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

// RouteBeam is a read-only route branch for values of type T1. It is accepted
// wherever [RouteSource] is.
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
			door.Outer(ctx, gox.Elem(func(cur gox.Cursor) error {
				return cur.Editor(render(r))
			}))
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
