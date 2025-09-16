// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package doors

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
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

func PathMatcherParts(n uint8) PathMatcher {
	return pathMatch([]any{"parts", n})
}
func QueryMatcherAll() QueryMatcher {
	return queryMatch([]any{"all"})
}

func QueryMatcherIgnore() QueryMatcher {
	return queryMatch([]any{"ignore"})
}

func QueryMatcherSome(params ...string) QueryMatcher {
	return queryMatch([]any{"some", params})
}

// Active configures active link
// indication
type Active struct {
	// Path match strategy
	PathMatcher PathMatcher
	// Query param match strategy
	QueryMatcher QueryMatcher
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
	if common.IsNill(h.Active.QueryMatcher) {
		h.Active.QueryMatcher = QueryMatcherAll()
	}
	if common.IsNill(h.Active.PathMatcher) {
		h.Active.PathMatcher = PathMatcherFull()
	}
	return []any{h.Active.PathMatcher, h.Active.QueryMatcher, front.IntoIndicate(h.Active.Indicator)}
}

func (h AHref) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, h)
}

func (h AHref) Attr() AttrInit {
	return h
}

func (h AHref) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	link, err := inst.NewLink(h.Model)
	if err != nil {
		slog.Error("href creation  error", slog.String("link_error", err.Error()))
		return
	}
	on, ok := link.ClickHandler()
	if ok {
		entry, ok := n.RegisterAttrHook(ctx, &door.AttrHook{
			Trigger: func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
				if len(h.After) != 0 {
					req := &request{w: w, r: r, ctx: ctx}
					req.After(h.After)
				}
				on(ctx)
				return false
			},
		})
		if !ok {
			return
		}
		if h.OnError == nil {
			h.OnError = ActionOnlyLocationReload()
		}
		attrs.AppendCapture(&front.LinkCapture{
			StopPropagation: h.StopPropagation,
		}, &front.Hook{
			Scope:     []*ScopeSet{front.LatestScope("link")},
			Indicate:  front.IntoIndicate(h.Indicator),
			Before:    intoActions(ctx, h.Before),
			OnError:   intoActions(ctx, h.OnError),
			HookEntry: entry,
		})
	}
	path, ok := link.Path()
	if ok {
		attrs.Set("href", path)
	}
	active := h.active()
	if active != nil {
		attrs.SetObject("data-d00r-active", active)
	}
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

func (s ARawSrc) init(ctx context.Context, n door.Core, inst instance.Core) (string, bool) {
	entry, ok := n.RegisterAttrHook(ctx, &door.AttrHook{
		Trigger: s.handle,
	})
	if !ok {
		return "", false
	}
	src := fmt.Sprintf("/d00r/%s/%d/%d", inst.Id(), entry.DoorId, entry.HookId)
	if s.Name != "" {
		src = src + "/" + s.Name
	}
	return src, true
}

func (s ARawSrc) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, s)
}

func (s ARawSrc) Attr() AttrInit {
	return s
}

func (s ARawSrc) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	src, ok := s.init(ctx, n, inst)
	if !ok {
		return
	}
	attrs.Set("src", src)
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

func (s ASrc) init(ctx context.Context, n door.Core, inst instance.Core) (string, bool) {
	entry, ok := n.RegisterAttrHook(ctx, &door.AttrHook{
		Trigger: s.handle,
	})
	if !ok {
		return "", false
	}
	if s.Name == "" {
		s.Name = filepath.Base(s.Path)
	}
	link := fmt.Sprintf("/d00r/%s/%d/%d/%s", inst.Id(), entry.DoorId, entry.HookId, s.Name)
	return link, true
}

func (s ASrc) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, s)
}

func (s ASrc) Attr() AttrInit {
	return s
}

func (s ASrc) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	src, ok := s.init(ctx, n, inst)
	if !ok {
		return
	}
	attrs.Set("src", src)
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

func (s AFileHref) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, s)
}

func (s AFileHref) Attr() AttrInit {
	return s
}

func (s AFileHref) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	link, ok := (*ASrc)(&s).init(ctx, n, inst)
	if !ok {
		return
	}
	attrs.Set("href", link)
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

func (s ARawFileHref) Render(ctx context.Context, w io.Writer) error {
	return front.AttrRender(ctx, w, s)
}

func (s ARawFileHref) Attr() AttrInit {
	return s
}

func (s ARawFileHref) Init(ctx context.Context, n door.Core, inst instance.Core, attrs *front.Attrs) {
	link, ok := (*ARawSrc)(&s).init(ctx, n, inst)
	if !ok {
		return
	}
	attrs.Set("href", link)
}
