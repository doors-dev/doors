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

// HistoryReplaceContext marks ctx so that a routing update made with it replaces
// the current browser history entry (history.replaceState) instead of pushing a
// new one (history.pushState).
//
// It works with any update that changes the URL, because they all write through
// to the same underlying [Location] state and the mark travels with the write:
// the top-level [Router] source, a path-model route source (from [RouteModel]),
// or a derived route source (from RouteDerive(...).Source). Update or Mutate are
// both honored.
//
// The mark only influences the resulting history entry; passing it to an update
// of a source not backed by the URL has no effect.
//
//	// replace instead of push, on a path-model source
//	path.Update(HistoryReplaceContext(ctx), Path{Section: SectionDocs})
func HistoryReplaceContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, common.KeyHistoryReplace, true)
}

// Route returns a renderable component that picks one of several writable
// route branches based on the current URL.
//
// Routes are tried in order and the first match renders. Use [RouteModel] to
// match on path-model structs and [RouteDefault] as the fallback.
func Route(routes ...RouteSource[Location]) gox.EditorComp {
	return gox.EditorCompFunc(func(cur gox.Cursor) error {
		path := Router(cur.Context())
		return path.Route(routes...).Edit(cur)
	})
}

// RouteLocationDefault creates a fallback URL route that calls render with a
// writable [Source] of the current [Location].
//
// This is the URL-specific wrapper around [RouteDefault].
//
// Deprecated: use [RouteDefault] instead.
func RouteLocationDefault[C gox.Comp](render func(s Source[Location]) C) RouteSource[Location] {
	return RouteDefault(render)
}

// RouteLocationDefaultBeam creates a fallback URL route that calls render
// with a read-only [Beam] of the current [Location].
//
// This is the URL-specific wrapper around [RouteDefaultBeam].
//
// Deprecated: use [RouteDefaultBeam] instead.
func RouteLocationDefaultBeam[C gox.Comp](render func(b Beam[Location]) C) RouteBeam[Location] {
	return RouteDefaultBeam(render)
}

// RouteLocationDefaultBind creates a fallback URL route that calls render with
// the raw [Location] value directly — a shorthand for [RouteDefaultBeam]
// plus bind.
//
// Deprecated: use [RouteDefaultBind] instead.
func RouteLocationDefaultBind[C gox.Comp](render func(b Location) C) RouteBeam[Location] {
	return RouteDefaultBind(render)
}

// RouteLocationDefaultComp creates a fallback URL route that renders a
// fixed component. The component receives no reactive value.
//
// Deprecated: use [RouteDefaultComp] instead.
func RouteLocationDefaultComp(comp gox.Comp) RouteBeam[Location] {
	return RouteDefaultComp[Location](comp)
}
