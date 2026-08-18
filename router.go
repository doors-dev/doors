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

// Location is a URL path plus query string, held in decoded form, as read from
// and written to [Router].
//
// Copies share Segments and Query, so use Clone before modifying a Location
// read from a source.
type Location = path.Location

// NewLocation encodes model into a [Location].
//
// model may be a [Location], a path model struct, a pointer to a path model
// struct, or a value implementing [LocationEncoder]. It returns an error if
// model is none of these or cannot be encoded.
func NewLocation(model any) (Location, error) {
	return path.Encode(model)
}

// LocationEncoder is implemented by models that build their own [Location]
// instead of describing their path with tags.
type LocationEncoder = path.Encoder

// Router returns the current URL of the page instance as a writable [Source].
//
// Updating it changes the browser URL without a page reload and reroutes the
// current page. Updates to a location with the same path and query are
// suppressed.
//
// ctx must belong to a Doors render or handler; otherwise Router panics.
func Router(ctx context.Context) Source[Location] {
	core := ctx.Value(common.KeyCore).(core.Core)
	return source[Location]{
		Source: core.Instance().Location(),
	}
}

// HistoryReplaceContext marks ctx so that a URL update made with it replaces
// the current browser history entry instead of pushing a new one.
//
// [Source.Update] and [Source.Mutate] honor the mark on the [Router] source
// and on any source derived from it, such as a [RouteModel] source. Passing it
// to an update of a source not backed by the URL has no effect.
func HistoryReplaceContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, common.KeyHistoryReplace, true)
}

// Route returns a component that renders the first matching route branch for
// the current URL.
//
// Routes are tried in order, and a branch is rerendered only when the matching
// route changes. If no route matches, nothing is rendered. Use [RouteModel] to
// match path models and [RouteDefault] as the fallback.
func Route(routes ...RouteSource[Location]) gox.EditorComp {
	return gox.EditorCompFunc(func(cur gox.Cursor) error {
		path := Router(cur.Context())
		return path.Route(routes...).Edit(cur)
	})
}
