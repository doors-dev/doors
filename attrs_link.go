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
	"net/http"

	"github.com/doors-dev/doors/internal/app"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/gox"
)

type queryMatch [][]any

func (qm queryMatch) And(q QueryMatcher) QueryMatcher {
	qm2 := q.queryMatch()
	out := make(queryMatch, len(qm)+len(qm2))
	copy(out, qm)
	copy(out[len(qm):], qm2)
	return out
}

func (q queryMatch) queryMatch() queryMatch {
	return q
}

var _ QueryMatcher = queryMatch(nil)

type pathMatch []any

func (q pathMatch) pathMatch() pathMatch {
	return q
}

// QueryMatcher customizes how query parameters participate in active-link
// matching. Matchers can be chained with [QueryMatcher.And].
type QueryMatcher interface {
	queryMatch() queryMatch
	Joiner[QueryMatcher]
}

// PathMatcher customizes how the path participates in active-link matching.
type PathMatcher interface {
	pathMatch() pathMatch
}

// PathMatcherFull matches the full generated path.
func PathMatcherFull() PathMatcher {
	return pathMatch([]any{"full"})
}

// PathMatcherStarts matches when the current path starts with the link path.
func PathMatcherStarts() PathMatcher {
	return pathMatch([]any{"starts"})
}

// PathMatcherSegments matches only the listed path segment indexes (zero-based).
func PathMatcherSegments(i ...int) PathMatcher {
	if i == nil {
		i = []int{}
	}
	return pathMatch([]any{"parts", i})
}

// QueryMatcherIgnoreSome excludes the given query parameters from comparison.
func QueryMatcherIgnoreSome(params ...string) QueryMatcher {
	if params == nil {
		params = []string{}
	}
	return queryMatch([][]any{{"ignore_some", params}})
}

// QueryMatcherIgnoreAll excludes all remaining query parameters from comparison.
func QueryMatcherIgnoreAll() QueryMatcher {
	return queryMatch([][]any{{"ignore_all"}})
}

// QueryMatcherSome compares only the provided query parameters at this step.
func QueryMatcherSome(params ...string) QueryMatcher {
	if params == nil {
		params = []string{}
	}
	return queryMatch([][]any{{"some", params}})
}

// QueryMatcherIfPresent matches the given parameters only if they are present.
func QueryMatcherIfPresent(params ...string) QueryMatcher {
	return queryMatch([][]any{{"if", params}})
}

// Active configures how [ALink] marks itself as active.
type Active struct {
	// PathMatcher controls path matching. Defaults to [PathMatcherFull].
	PathMatcher PathMatcher
	// QueryMatcher controls query matching. Matchers are applied sequentially;
	// any remaining parameters are compared after the configured matcher chain.
	QueryMatcher QueryMatcher
	// FragmentMatch includes the URL fragment in active matching.
	FragmentMatch bool
	// Indicator is applied to the link while it matches the current location.
	Indicator Indicators
}

// ALink builds a real href from Model and adds Doors navigation behavior.
//
// A normal click updates the current instance location source and reroutes the
// page dynamically. The href remains valid for browser features such as opening
// in a new tab, copying the link, or navigation without client-side runtime.
//
// Example:
//
//	attrs := doors.A(ctx, doors.ALink{
//		Model: Path{Home: true},
//	})
type ALink struct {
	// Target path model value. Required.
	Model any
	// Fragment identifier. Optional.
	Fragment string
	// Active link indicator configuration. Optional.
	Active Active
	// Stop event propagation (for dynamic links). Optional.
	StopPropagation bool
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope Scopes
	// Visual indicators while the hook is running. Optional.
	Indicator Indicators
	// Actions to run before the hook request. Optional.
	Before Actions
	// Actions to run after the hook request. Optional.
	After Actions
	// Actions to run on error.
	// Default (nil) triggers a location reload.
	OnError Actions
}

func (h *ALink) active() []any {
	indicators := indicatorsOrNil(h.Active.Indicator)
	if len(indicators) == 0 {
		return nil
	}
	if h.Active.QueryMatcher == nil {
		h.Active.QueryMatcher = make(queryMatch, 0)
	}
	h.Active.QueryMatcher = h.Active.QueryMatcher.And(queryMatch([][]any{{"all"}}))
	if h.Active.PathMatcher == nil {
		h.Active.PathMatcher = PathMatcherFull()
	}
	return []any{h.Active.PathMatcher, h.Active.QueryMatcher.queryMatch(), h.Active.FragmentMatch, indicators}
}

func (h ALink) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(h, cur, elem)
}

func (h ALink) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	loc, err := path.Encode(h.Model)
	if err != nil {
		slog.Error("href creation error", "error", err)
		return nil
	}
	h.Scope = linkScope{}.And(h.Scope)
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		if r.Header.Get(app.ZombieHeader) != "" {
			req := &request{w: w, r: r, ctx: ctx}
			req.After(ActionLocationReload{})
			InstanceEnd(ctx)
			return false
		}
		if h.Fragment != "" {
			if h.After == nil {
				h.After = ActionScroll{Selector: "#" + h.Fragment}
			} else {
				h.After = h.After.And(ActionScroll{Selector: "#" + h.Fragment})
			}
		}
		if h.After != nil {
			req := request{w: w, r: r, ctx: ctx}
			req.After(h.After)
		}
		core.Location().Update(ctx, loc)
		return false
	}
	hook, ok := core.RegisterHook(handler, nil)
	if !ok {
		return nil
	}
	if h.OnError == nil {
		h.OnError = ActionLocationReload{}
	}
	front.AttrsAppendCapture(attrs, front.LinkCapture{
		StopPropagation: h.StopPropagation,
	}, front.Hook{
		Indicate: indicatorsOrNil(h.Indicator),
		Scope:    scopesOrNil(core, h.Scope),
		Before:   intoActions(ctx, actionsOrNil(h.Before)),
		OnError:  intoActions(ctx, actionsOrNil(h.OnError)),
		Hook:     hook,
	})
	href := loc.String()
	if h.Fragment != "" {
		href += "#" + h.Fragment
	}
	attrs.Get("href").Set(href)
	active := h.active()
	if active != nil {
		front.AttrsSetParent(attrs, core.DoorID())
		front.AttrsSetActive(attrs, active)
	}
	return nil
}
