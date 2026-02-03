// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

type HrefActiveMatch int

const (
	MatchFull HrefActiveMatch = iota
	MatchPath
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

type QueryMatcher interface {
	queryMatch() queryMatch
}

type PathMatcher interface {
	pathMatch() pathMatch
}

func PathMatcherFull() PathMatcher {
	return pathMatch([]any{"full"})
}

func PathMatcherStarts() PathMatcher {
	return pathMatch([]any{"starts"})
}

// PathMatcherParts checks if path parts by specified indexes matche
func PathMatcherParts(i ...int) PathMatcher {
	return pathMatch([]any{"parts", i})
}

// QueryMatcherIgnoreSome excludes the given query parameters from comparison.
func QueryMatcherIgnoreSome(params ...string) QueryMatcher {
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
	return queryMatch([]any{"some"})
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

// Active configures active link
// indication
type Active struct {
	// Path match strategy
	PathMatcher PathMatcher
	// Query param match strategy, applied sequientially
	QueryMatcher []QueryMatcher
	// Indicators to apply when active
	Indicator []Indicator
}

// AHref prepares the href attribute for internal navigation
// and configures dynamic link behavior.
type AHref struct {
	// Target path model value. Required.
	Model any
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

func (h *AHref) active() []any {
	if len(h.Active.Indicator) == 0 {
		return nil
	}
	h.Active.QueryMatcher = append(h.Active.QueryMatcher, queryMatch([]any{"all"}))
	if common.IsNill(h.Active.PathMatcher) {
		h.Active.PathMatcher = PathMatcherFull()
	}
	return []any{h.Active.PathMatcher, h.Active.QueryMatcher, front.IntoIndicate(h.Active.Indicator)}
}

func (h AHref) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(h, cur, elem)
}

func (h AHref) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	link, err := core.NewLink(h.Model)
	if err != nil {
		slog.Error("href creation  error", slog.String("link_error", err.Error()))
		return nil
	}
	h.Scope = append([]Scope{&ScopeBlocking{}}, h.Scope...)
	on, ok := link.ClickHandler()
	if ok {
		handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			if len(h.After) != 0 {
				req := &request{w: w, r: r, ctx: ctx}
				req.After(h.After)
			}
			on(ctx)
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
			Indicate:  front.IntoIndicate(h.Indicator),
			Scope:     front.IntoScopeSet(core, h.Scope),
			Before:    intoActions(ctx, h.Before),
			OnError:   intoActions(ctx, h.OnError),
			Hook: hook,
		})
	}
	path, ok := link.Path()
	if ok {
		attrs.Get("href").Set(path)
	}
	active := h.active()
	if active != nil {
		front.AttrsSetActive(attrs, active)
	}
	return nil
}



// ARawSrc prepares the src attribute for a downloadable resource
// served directly and privately through a custom handler.
type ARawSrc struct {
	// If true, resource is available for download only once.
	Once bool
	// File name. Optional.
	Name string
	// Handler for serving the resource request.
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (s ARawSrc) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(s, cur, elem)
}

func (s ARawSrc) init(ctx context.Context) (string, bool) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	hook, ok := core.RegisterHook(s.handle, nil)
	if !ok {
		return "", false
	}
	src := fmt.Sprintf("/~0/%s/%d/%d", core.InstanceID(), hook.DoorID, hook.HookID)
	if s.Name != "" {
		src = src + "/" + s.Name
	}
	return src, true
}

func (s ARawSrc) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	src, ok := s.init(ctx)
	if !ok {
		return nil
	}
	attrs.Get("src").Set(src)
	return nil
}


func (s *ARawSrc) handle(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
	s.Handler(w, r)
	return s.Once
}

// ASrc prepares the src attribute for a downloadable resource
// served privately from a file system path.
type ASrc struct {
	// If true, resource is available for download only once.
	Once bool
	// File name. Optional.
	Name string
	// File system path to serve.
	Path string
}

func (s ASrc) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(s, cur, elem)
}

func (s ASrc) init(ctx context.Context) (string, bool) {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	hook, ok := core.RegisterHook(s.handle, nil)
	if !ok {
		return "", false
	}
	if s.Name == "" {
		s.Name = filepath.Base(s.Path)
	}
	return  fmt.Sprintf("/~0/%s/%d/%d/%s", core.InstanceID(), hook.DoorID, hook.HookID, s.Name), true
}

func (s ASrc) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	src, ok := s.init(ctx)
	if !ok {
		return nil
	}
	attrs.Get("src").Set(src)
	return nil
}

func (s *ASrc) handle(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
	http.ServeFile(w, r, s.Path)
	return s.Once
}

// AFileHref prepares the href attribute for a downloadable resource
// served privately from a file system path.
type AFileHref struct {
	// If true, resource is available for download only once.
	Once bool
	// File name. Optional.
	Name string
	// File system path to serve.
	Path string
}

func (s AFileHref) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(s, cur, elem)
}

func (s AFileHref) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	link, ok := (*ASrc)(&s).init(ctx)
	if !ok {
		return nil
	}
	attrs.Get("href").Set(link)
	return nil
}


// ARawFileHref prepares the href attribute for a downloadable resource
// served privately and directly through a custom handler.
type ARawFileHref struct {
	// If true, resource is available for download only once.
	Once bool
	// File name. Optional.
	Name string
	// Handler for serving the resource request.
	Handler func(w http.ResponseWriter, r *http.Request)
}

func (s ARawFileHref) Proxy(cur gox.Cursor, elem gox.Elem) error {
	return proxyAddAttrMod(s, cur, elem)
}

func (s ARawFileHref) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	link, ok := (*ARawSrc)(&s).init(ctx)
	if !ok {
		return nil
	}
	attrs.Get("href").Set(link)
	return nil
}
