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

	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/gox"
)

type SourceRoute[T any] interface {
	init(source Source[T]) route[T]
}

type BeamRoute[T any] interface {
	SourceRoute[T]
	initBeam(beam Beam[T]) route[T]
}

type route[T any] interface {
	clean()
	seek(ctx context.Context, value T) (gox.Comp, seekResult)
	keep(ctx context.Context, value T) (gox.Comp, keepResult)
}

type seekResult int

const (
	seekNone seekResult = iota
	seekOk
	seekDefault
)

type keepResult int

const (
	keepCheck keepResult = iota
	keepEnd
	keepUpdate
	keepSwitch
)

type sourceRouteFunc[T any] func(source Source[T]) route[T]

func (f sourceRouteFunc[T]) init(source Source[T]) route[T] {
	return f(source)
}

type beamRouteFunc[T any] func(beam Beam[T]) route[T]

func (f beamRouteFunc[T]) init(source Source[T]) route[T] {
	return f(source)
}

func (f beamRouteFunc[T]) initBeam(beam Beam[T]) route[T] {
	return f(beam)
}

func RouteComp[T any](render gox.Comp) BeamRoute[T] {
	return beamRouteFunc[T](func(Beam[T]) route[T] {
		return &compRoute[T]{
			comp: render,
		}
	})
}

func RouteBind[T any](render func(v T) gox.Elem) BeamRoute[T] {
	return beamRouteFunc[T](func(Beam[T]) route[T] {
		return &bindRoute[T]{
			render: render,
		}
	})
}

func RouteMatch[T any](match func(v T) bool, routes ...SourceRoute[T]) SourceRoute[T] {
	return sourceRouteFunc[T](func(source Source[T]) route[T] {
		return &matchRoute[T]{
			routes: newRoutesRoute(routes, source),
			match:  match,
		}
	})
}

func RouteBeamMatch[T any](match func(v T) bool, routes ...BeamRoute[T]) BeamRoute[T] {
	return beamRouteFunc[T](func(beam Beam[T]) route[T] {
		return &matchRoute[T]{
			routes: newBeamRoutesRoute(routes, beam),
			match:  match,
		}
	})
}

func RouteValue[T any](value T, routes ...SourceRoute[T]) SourceRoute[T] {
	return sourceRouteFunc[T](func(source Source[T]) route[T] {
		equal := source.innerBeam().Equal
		return &matchRoute[T]{
			routes: newRoutesRoute(routes, source),
			match: func(v T) bool {
				return equal(value, v)
			},
		}
	})
}

func RouteBeamValue[T any](value T, routes ...BeamRoute[T]) BeamRoute[T] {
	return beamRouteFunc[T](func(beam Beam[T]) route[T] {
		equal := beam.innerBeam().Equal
		return &matchRoute[T]{
			routes: newBeamRoutesRoute(routes, beam),
			match: func(v T) bool {
				return equal(value, v)
			},
		}
	})
}

func RouteDerive[T1 any, T2 any](derive func(source Source[T1]) Source[T2], routes ...SourceRoute[T2]) SourceRoute[T1] {
	return sourceRouteFunc[T1](func(source Source[T1]) route[T1] {
		derived := derive(source)
		return &deriveRoute[T1, T2]{
			routes:  newRoutesRoute(routes, derived),
			derived: derived,
		}
	})
}

func RouteBeamDerive[T1 any, T2 any](derive func(beam Beam[T1]) Beam[T2], routes ...BeamRoute[T2]) BeamRoute[T1] {
	return beamRouteFunc[T1](func(beam Beam[T1]) route[T1] {
		derived := derive(beam)
		return &deriveRoute[T1, T2]{
			routes:  newBeamRoutesRoute(routes, derived),
			derived: derived,
		}
	})
}

func RouteModel[M any](routes ...SourceRoute[M]) SourceRoute[Location] {
	match, get, set := modelFuncs[M]()
	return RouteMatch(match, RouteDerive(func(s Source[Location]) Source[M] {
		return DeriveSourceEqual(s, get, set, func(a, b M) bool {
			return reflect.DeepEqual(a, b)
		})
	}, routes...))
}

func RouteBeamModel[M any](routes ...BeamRoute[M]) BeamRoute[Location] {
	match, get, _ := modelFuncs[M]()
	return RouteBeamMatch(match, RouteBeamDerive(func(b Beam[Location]) Beam[M] {
		return DeriveBeamEqual(b, get, func(a, b M) bool {
			return reflect.DeepEqual(a, b)
		})
	}, routes...))
}

func modelFuncs[M any]() (func(Location) bool, func(Location) M, func(Location, M) Location) {
	a, err := path.GetModelAdapter[M]()
	if err != nil {
		slog.Error("Model adapter error", "error", err)
	}
	match := func(l Location) bool {
		if err != nil {
			return false
		}
		_, ok := a.Decode(l)
		return ok
	}
	get := func(l Location) M {
		m, ok := a.Decode(l)
		if !ok {
			return *new(M)
		}
		return *m
	}
	set := func(prev Location, m M) Location {
		loc, err := a.Encode(&m)
		if err != nil {
			slog.Error("Model adapter encoding error", "error", err)
			return prev
		}
		return loc
	}
	return match, get, set
}

func RouteBeam[T any](render func(b Beam[T]) gox.Elem) BeamRoute[T] {
	return beamRouteFunc[T](func(beam Beam[T]) route[T] {
		return &compRoute[T]{
			comp: gox.Elem(func(cur gox.Cursor) error {
				return cur.Comp(render(beam))
			}),
		}
	})
}

func RouteSource[T any](render func(s Source[T]) gox.Elem) SourceRoute[T] {
	return sourceRouteFunc[T](func(source Source[T]) route[T] {
		return &compRoute[T]{
			comp: gox.Elem(func(cur gox.Cursor) error {
				return cur.Comp(render(source))
			}),
		}
	})
}

func renderSourceRoutes[T any](source Source[T], routes []SourceRoute[T]) gox.Elem {
	return renderRoutes(source, func() routesRoute[T] {
		return newRoutesRoute(routes, source)
	})
}

func renderBeamRoutes[T any](beam Beam[T], routes []BeamRoute[T]) gox.Elem {
	return renderRoutes(beam, func() routesRoute[T] {
		return newBeamRoutesRoute(routes, beam)
	})
}

func renderRoutes[T any](beam Beam[T], newRoutes func() routesRoute[T]) gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		door := new(Door)
		rr := newRoutes()
		matched := false
		blank := true
		show := func(ctx context.Context, el gox.Comp) {
			if el == nil && blank {
				return
			}
			blank = el == nil
			door.Outer(ctx, el)
		}
		ok := beam.Sub(cur.Context(), func(ctx context.Context, t T) bool {
			if !matched {
				el, res := rr.seek(ctx, t)
				if res == seekNone {
					return false
				}
				matched = true
				show(ctx, el)
				return false
			}
			el, res := rr.keep(ctx, t)
			switch res {
			case keepCheck:
				return false
			case keepEnd:
				rr.clean()
				el, res := rr.seek(ctx, t)
				matched = res != seekNone
				show(ctx, el)
				return false
			default:
				show(ctx, el)
				return false
			}

		})
		if !ok {
			return nil
		}
		return cur.Comp(door)
	})
}

type routesRoute[T any] struct {
	routes        []route[T]
	active        int
	activeDefault bool
}

func newRoutesRoute[T any](routes []SourceRoute[T], source Source[T]) routesRoute[T] {
	inited := make([]route[T], len(routes))
	for i, r := range routes {
		inited[i] = r.init(source)
	}
	return routesRoute[T]{
		routes: inited,
		active: -1,
	}
}

func newBeamRoutesRoute[T any](routes []BeamRoute[T], beam Beam[T]) routesRoute[T] {
	inited := make([]route[T], len(routes))
	for i, r := range routes {
		inited[i] = r.initBeam(beam)
	}
	return routesRoute[T]{
		routes: inited,
		active: -1,
	}
}

func (sr *routesRoute[T]) clean() {
	if sr.active != -1 {
		sr.routes[sr.active].clean()
	}
	sr.active = -1
	sr.activeDefault = false
}

func (sr *routesRoute[T]) seek(ctx context.Context, value T) (gox.Comp, seekResult) {
	for i, r := range sr.routes {
		el, res := r.seek(ctx, value)
		if res == seekNone {
			continue
		}
		sr.active = i
		sr.activeDefault = res == seekDefault
		return el, res
	}
	return nil, seekNone
}

func (sr *routesRoute[T]) keep(ctx context.Context, value T) (gox.Comp, keepResult) {
	if sr.activeDefault {
		for i := 0; i < sr.active; i++ {
			el, res := sr.routes[i].seek(ctx, value)
			if res == seekNone {
				continue
			}
			sr.routes[sr.active].clean()
			sr.active = i
			sr.activeDefault = false
			return el, keepSwitch
		}
	}
	el, res := sr.routes[sr.active].keep(ctx, value)
	if res == keepSwitch {
		sr.activeDefault = false
	}
	return el, res
}

type matchRoute[T any] struct {
	routes routesRoute[T]
	match  func(T) bool
}

func (sr *matchRoute[T]) clean() {
	sr.routes.clean()
}

func (sr *matchRoute[T]) seek(ctx context.Context, value T) (gox.Comp, seekResult) {
	if !sr.match(value) {
		return nil, seekNone
	}
	el, res := sr.routes.seek(ctx, value)
	if res == seekNone {
		return nil, seekNone
	}
	return el, seekOk
}

func (sr *matchRoute[T]) keep(ctx context.Context, value T) (gox.Comp, keepResult) {
	if !sr.match(value) {
		return nil, keepEnd
	}
	el, res := sr.routes.keep(ctx, value)
	if res == keepSwitch {
		return el, keepUpdate
	}
	return el, res
}

type deriveRoute[T1 any, T2 any] struct {
	routes  routesRoute[T2]
	derived Beam[T2]
	last    T2
	none    bool
}

func (sr *deriveRoute[T1, T2]) clean() {
	sr.routes.clean()
	var v T2
	sr.last = v
	sr.none = false
}

func (sr *deriveRoute[T1, T2]) seek(ctx context.Context, _ T1) (gox.Comp, seekResult) {
	value, ok := sr.derived.Read(ctx)
	if !ok {
		return nil, seekOk
	}
	if sr.none && sr.derived.innerBeam().Equal(sr.last, value) {
		return nil, seekNone
	}
	el, res := sr.routes.seek(ctx, value)
	sr.last = value
	sr.none = res == seekNone
	return el, res
}

func (sr *deriveRoute[T1, T2]) keep(ctx context.Context, _ T1) (gox.Comp, keepResult) {
	value, ok := sr.derived.Read(ctx)
	if !ok {
		return nil, keepCheck
	}
	if sr.derived.innerBeam().Equal(sr.last, value) {
		return nil, keepCheck
	}
	sr.last = value
	return sr.routes.keep(ctx, value)
}

type compRoute[T any] struct {
	comp gox.Comp
}

func (sr *compRoute[T]) clean() {
}

func (sr *compRoute[T]) seek(ctx context.Context, _ T) (gox.Comp, seekResult) {
	return sr.comp, seekDefault
}

func (sr *compRoute[T]) keep(ctx context.Context, _ T) (gox.Comp, keepResult) {
	return nil, keepCheck
}

type bindRoute[T any] struct {
	render func(v T) gox.Elem
}

func (sr *bindRoute[T]) clean() {
}

func (sr *bindRoute[T]) comp(value T) gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		return cur.Comp(sr.render(value))
	})
}

func (sr *bindRoute[T]) seek(ctx context.Context, value T) (gox.Comp, seekResult) {
	return sr.comp(value), seekDefault
}

func (sr *bindRoute[T]) keep(ctx context.Context, value T) (gox.Comp, keepResult) {
	return sr.comp(value), keepUpdate
}
