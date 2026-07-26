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
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
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

// QueryMatcher is a query-string comparison step of [Active] matching. Chain
// steps with [Joiner.And].
type QueryMatcher interface {
	queryMatch() queryMatch
	Joiner[QueryMatcher]
}

// PathMatcher is the path comparison rule of [Active] matching.
type PathMatcher interface {
	pathMatch() pathMatch
}

// PathMatcherFull returns a [PathMatcher] that matches when the whole path is
// equal, ignoring leading and trailing slashes.
func PathMatcherFull() PathMatcher {
	return pathMatch([]any{"full"})
}

// PathMatcherStarts returns a [PathMatcher] that matches when the current path
// starts with the link path. The comparison is by string prefix, not by whole
// segments.
func PathMatcherStarts() PathMatcher {
	return pathMatch([]any{"starts"})
}

// PathMatcherSegments returns a [PathMatcher] that compares only the path
// segments at the given indexes, counted from zero. Without indexes it matches
// any path.
func PathMatcherSegments(i ...int) PathMatcher {
	if i == nil {
		i = []int{}
	}
	return pathMatch([]any{"parts", i})
}

// QueryMatcherIgnoreSome returns a [QueryMatcher] that drops params from the
// comparison and passes the remaining parameters to the next step.
func QueryMatcherIgnoreSome(params ...string) QueryMatcher {
	if params == nil {
		params = []string{}
	}
	return queryMatch([][]any{{"ignore_some", params}})
}

// QueryMatcherIgnoreAll returns a [QueryMatcher] that ends the chain, leaving
// every remaining parameter out of the comparison.
func QueryMatcherIgnoreAll() QueryMatcher {
	return queryMatch([][]any{{"ignore_all"}})
}

// QueryMatcherSome returns a [QueryMatcher] that compares params at this step
// and drops them from the comparison of later steps.
func QueryMatcherSome(params ...string) QueryMatcher {
	if params == nil {
		params = []string{}
	}
	return queryMatch([][]any{{"some", params}})
}

// QueryMatcherIfPresent returns a [QueryMatcher] that compares params only
// when the current location carries them, and drops them from the comparison
// of later steps.
func QueryMatcherIfPresent(params ...string) QueryMatcher {
	return queryMatch([][]any{{"if", params}})
}

// Active is the rule that decides when an [ALink] counts as the current
// location.
type Active struct {
	// PathMatcher compares the link path with the current path. Default:
	// [PathMatcherFull].
	PathMatcher PathMatcher
	// QueryMatcher compares query parameters step by step, then any parameter
	// left over must be equal. Optional; nil compares all of them.
	QueryMatcher QueryMatcher
	// FragmentMatch also requires the fragment to be equal. Optional; the
	// fragment is ignored by default.
	FragmentMatch bool
	// Indicator is applied to the link while it matches the current location.
	// Required; without it the link is never marked active.
	Indicator Indicators
}

// ALink turns the element it is attached to into an in-app navigation link.
//
// It sets href from Model and intercepts the click: the click updates the
// instance location and the page reroutes in place. The href stays a real URL,
// so it can be copied, opened in another tab, or loaded directly. A click that
// changes neither path nor query sends no request.
//
// Unlike [ActionLocationAssign], the page is not loaded again and the current
// instance stays alive.
//
// Example:
//
//	~>doors.ALink{
//		Model: Path{Section: SectionHome},
//	} <a>Home</a>
type ALink struct {
	// Model is the target path model value, a [Location], or a
	// [LocationEncoder]. Required; a value that cannot be encoded logs an
	// error and leaves the element without href or navigation.
	Model any
	// Fragment is appended to href and scrolled into view once the
	// navigation succeeds. Optional.
	Fragment string
	// Active marks the link while it matches the current location. Optional;
	// it does nothing unless its Indicator is set.
	Active Active
	// StopPropagation stops the click from bubbling up the DOM. Optional.
	StopPropagation bool
	// HistoryReplace replaces the current history entry instead of pushing a
	// new one, so Back skips this navigation. Optional.
	HistoryReplace bool
	// Scope controls how the request is scheduled. Optional; all links share a
	// built-in scope, so a new click cancels a pending navigation.
	Scope Scopes
	// Indicator lists temporary DOM changes applied while the request is
	// in flight. Optional.
	Indicator Indicators
	// Before lists client-side actions to run before the request is
	// sent. Optional.
	Before Actions
	// After lists client-side actions to run once the navigation
	// succeeds. Optional.
	After Actions
	// OnError lists client-side actions to run when the request fails.
	// Optional; a canceled navigation reverts the URL without running it.
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
	core := ctx.Value(common.KeyCore).(core.Core)
	loc, err := path.Encode(h.Model)
	if err != nil {
		core.App().Logger().Error("href creation error", "error", err)
		return nil
	}
	h.Scope = linkScope{}.And(h.Scope)
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		if core.Instance().Session().App().Draining() {
			w.WriteHeader(http.StatusGone)
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
		if h.HistoryReplace {
			ctx = HistoryReplaceContext(ctx)
		}
		core.Instance().Location().Update(ctx, loc)
		return false
	}
	hook, ok := core.Door().RegisterHook(handler, nil)
	if !ok {
		return nil
	}
	front.AttrsAppendCapture(attrs, front.LinkCapture{
		StopPropagation: h.StopPropagation,
		HistoryReplace:  h.HistoryReplace,
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
		front.AttrsSetParent(attrs, core.Door().ID())
		front.AttrsSetActive(attrs, active)
	}
	return nil
}
