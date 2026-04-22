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

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/router"
	"github.com/doors-dev/gox"
)

// HrefActiveMatch describes a path matching strategy for active links.
type HrefActiveMatch int

const (
	// MatchFull requires the entire path to match.
	MatchFull HrefActiveMatch = iota
	// MatchPath requires the same path but ignores the query.
	MatchPath
	// MatchStarts requires the current path to start with the target path.
	MatchStarts
)

type queryMatch []any
type pathMatch []any

func (q queryMatch) queryMatch() queryMatch {
	return q
}

func (q pathMatch) pathMatch() pathMatch {
	return q
}

// QueryMatcher customizes how query parameters participate in active-link
// matching.
type QueryMatcher interface {
	queryMatch() queryMatch
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
	return queryMatch([]any{"ignore_some", params})
}

// QueryMatcherOnlyIgnoreSome ignores the given parameters and matches all remaining.
func QueryMatcherOnlyIgnoreSome(params ...string) []QueryMatcher {
	return []QueryMatcher{QueryMatcherIgnoreSome(params...)}
}

// QueryMatcherIgnoreAll excludes all remaining query parameters from comparison.
func QueryMatcherIgnoreAll() QueryMatcher {
	return queryMatch([]any{"ignore_all"})
}

// QueryMatcherOnlyIgnoreAll ignores all query parameters.
func QueryMatcherOnlyIgnoreAll() []QueryMatcher {
	return []QueryMatcher{QueryMatcherIgnoreAll()}
}

// QueryMatcherSome matches only the provided query parameters.
func QueryMatcherSome(params ...string) QueryMatcher {
	if params == nil {
		params = []string{}
	}
	return queryMatch([]any{"some", params})
}

// QueryMatcherOnlySome matches the provided query parameters and ignores all others.
func QueryMatcherOnlySome(params ...string) []QueryMatcher {
	return []QueryMatcher{QueryMatcherSome(params...), QueryMatcherIgnoreAll()}
}

// QueryMatcherIfPresent matches the given parameters only if they are present.
func QueryMatcherIfPresent(params ...string) QueryMatcher {
	return queryMatch([]any{"if", params})
}

// QueryMatcherOnlyIfPresent matches the given parameters if present and ignores all others.
func QueryMatcherOnlyIfPresent(params ...string) []QueryMatcher {
	return []QueryMatcher{QueryMatcherIfPresent(params...), QueryMatcherIgnoreAll()}
}

// Active configures how [ALink] marks itself as active.
type Active struct {
	// Path match strategy
	PathMatcher PathMatcher
	// Query param match strategy, applied sequentially.
	QueryMatcher []QueryMatcher
	// Match fragment, false by default
	FragmentMatch bool
	// Indicators to apply when active
	Indicator []Indicator
}

// ALink builds a real `href` from Model and, when possible, upgrades the link
// to same-model client navigation.
//
// Example:
//
//	attrs := doors.A(ctx, doors.ALink{
//		Model: Path{Home: true},
//	})
type ALink struct {
	// Target path model value. Required.
	Model any
	// Fragment identifier
	// Optional
	Fragment string
	// Active link indicator configuration. Optional.
	Active Active
	// Stop event propagation (for dynamic links). Optional.
	StopPropagation bool
	// Defines how the hook is scheduled (e.g. blocking, debounce).
	// Optional.
	Scope []Scope
	// Visual indicators while the hook is running
	// (for dynamic links). Optional.
	Indicator []Indicator
	// Actions to run before the hook request (for dynamic links). Optional.
	Before []Action
	// Actions to run after the hook request (for dynamic links). Optional.
	After []Action
	// Actions to run on error (for dynamic links).
	// Default (nil) triggers a location reload.
	OnError []Action
}

func (h *ALink) active() []any {
	if len(h.Active.Indicator) == 0 {
		return nil
	}
	h.Active.QueryMatcher = append(h.Active.QueryMatcher, queryMatch([]any{"all"}))
	if h.Active.PathMatcher == nil {
		h.Active.PathMatcher = PathMatcherFull()
	}
	return []any{h.Active.PathMatcher, h.Active.QueryMatcher, h.Active.FragmentMatch, front.IntoIndicate(h.Active.Indicator)}
}

func (h ALink) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyMod(h, cur, elem)
}

func (h ALink) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	link, err := core.NewLink(h.Model)
	if err != nil {
		slog.Error("href creation error", "error", err)
		return nil
	}
	h.Scope = append([]Scope{linkScope{}}, h.Scope...)
	if link.On != nil {
		handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			if r.Header.Get(router.ZombieHeader) != "" {
				req := &request{w: w, r: r, ctx: ctx}
				req.After(ActionOnlyLocationReload())
				InstanceEnd(ctx)
				return false
			}
			if h.Fragment != "" {
				h.After = append(h.After, ActionScroll{Selector: "#" + h.Fragment})
			}
			if len(h.After) != 0 {
				req := &request{w: w, r: r, ctx: ctx}
				req.After(h.After)
			}
			link.On(ctx)
			return false
		}
		hook, ok := core.RegisterHook(handler, nil)
		if !ok {
			return nil
		}
		if h.OnError == nil {
			h.OnError = ActionOnlyLocationReload()
		}
		front.AttrsAppendCapture(attrs, front.LinkCapture{
			StopPropagation: h.StopPropagation,
		}, front.Hook{
			Indicate: front.IntoIndicate(h.Indicator),
			Scope:    front.IntoScopeSet(core, h.Scope),
			Before:   intoActions(ctx, h.Before),
			OnError:  intoActions(ctx, h.OnError),
			Hook:     hook,
		})
	}
	fragment := ""
	if h.Fragment != "" {
		fragment = "#" + h.Fragment
	}
	attrs.Get("href").Set(link.Location.String() + fragment)
	active := h.active()
	if active != nil {
		front.AttrsSetParent(attrs, core.DoorID())
		front.AttrsSetActive(attrs, active)
	}
	return nil
}
