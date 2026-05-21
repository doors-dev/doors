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

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/gox"
)

// Location is a parsed or generated URL path plus query string.
type Location = path.Location

// NewLocation encodes model into a [Location].
//
// model may be a [Location], a path-model struct, a pointer to a path-model
// struct, or a value implementing [LocationEncoder].
func NewLocation(model any) (Location, error) {
	return path.Encode(model)
}

// LocationEncoder is implemented by custom navigation models that can encode
// themselves as a [Location].
type LocationEncoder = path.Encoder

// Router returns the current instance URL as a writable reactive source.
//
// Updating the returned source changes the browser location and reroutes the
// current page.
func Router(ctx context.Context) Source[Location] {
	core := ctx.Value(common.KeyCore).(core.Core)
	return source[Location]{
		Source: core.Instance().Location(),
	}
}

// Deprecated: Use Route(...)
func RouterBeam(routes ...RouteBeam[Location]) gox.EditorComp {
	return gox.EditorCompFunc(func(cur gox.Cursor) error {
		path := Router(cur.Context())
		return path.RouteBeam(routes...).Edit(cur)
	})
}

// Deprecated: Use Route(...)
func RouterSource(routes ...RouteSource[Location]) gox.EditorComp {
	return Route(routes...)
}

// Route returns a renderable component that routes the current URL through
// writable route branches.
func Route(routes ...RouteSource[Location]) gox.EditorComp {
	return gox.EditorCompFunc(func(cur gox.Cursor) error {
		path := Router(cur.Context())
		return path.Route(routes...).Edit(cur)
	})
}

// RouteLocationDefault creates a fallback URL route that renders a
// writable source for the current [Location].
func RouteLocationDefault[C gox.Comp](render func(Source[Location]) C) RouteSource[Location] {
	return defaultRouteSource[Location](func(l Source[Location]) gox.Elem {
		return render(l).Main()
	})
}

// RouteLocationDefaultBeam creates a fallback URL route that renders a
// read-only beam for the current [Location].
func RouteLocationDefaultBeam[C gox.Comp](render func(Beam[Location]) C) RouteBeam[Location] {
	return defaultRouteBeam[Location](func(l Beam[Location]) gox.Elem {
		return render(l).Main()
	})
}

// RouteLocationDefaultComp creates a fallback URL route that renders comp.
func RouteLocationDefaultComp(comp gox.Comp) RouteBeam[Location] {
	return defaultRouteBeam[Location](func(b Beam[Location]) gox.Elem {
		return comp.Main()
	})
}
