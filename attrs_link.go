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

func PathMatcherParts(n int) PathMatcher {
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

type Active struct {
	PathMatcher  PathMatcher
	QueryMatcher QueryMatcher
	Indicator    []Indicator
}

type AHref struct {
	// Target path model value
	Model any
	// Active link indicator configuration
	Active Active
	// Stops event propagation (for dynamic link)
	StopPropagation bool
	// Scrolls into selector (for dynamic link)
	ScrollInto string
	// Loading indications (for dynamic link)
	Indicator []Indicator
	// Action on error (for dynamic link)
	OnError []OnError
	// For analytics purposes
	Callback func()
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
				if h.Callback != nil {
					defer h.Callback()
				}
				if h.ScrollInto != "" {
					r := request{
						w:   w,
						r:   r,
						ctx: ctx,
					}
					r.After(AfterScrollInto(h.ScrollInto, false))
				}
				on(ctx)
				return false
			},
		})
		if !ok {
			return
		}
		attrs.AppendCapture(&front.LinkCapture{
			StopPropagation: h.StopPropagation,
		}, &front.Hook{
			Scope:     []*ScopeSet{front.LatestScope("link")},
			Indicate:  front.IntoIndicate(h.Indicator),
			Error:     front.IntoErrorAction(h.OnError),
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

type ARawSrc struct {
	Once    bool
	Name    string
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

// attribute to securely serve a file
type ASrc struct {
	Path string
	Once bool
	Name string
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

type AFileHref struct {
	Path string
	Once bool
	Name string
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

type ARawFileHref struct {
	Once    bool
	Name    string
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
