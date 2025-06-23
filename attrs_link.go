package doors

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/node"
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

type ActiveQueryMatcher interface {
	queryMatch() queryMatch
}

type ActivePathMatcher interface {
	pathMatch() pathMatch
}

type activeMatchers struct{}

func ActivePathFull() ActivePathMatcher {
	return pathMatch([]any{"full"})
}

func ActivePathStarts() ActivePathMatcher {
	return pathMatch([]any{"starts"})
}

func ActivePathParts(n int) ActivePathMatcher {
	return pathMatch([]any{"parts", n})
}
func ActiveQueryAll() ActiveQueryMatcher {
	return queryMatch([]any{"all"})
}

func ActiveQueryIgnore() ActiveQueryMatcher {
	return queryMatch([]any{"ignore"})
}

func ActiveQuerySome(params ...string) ActiveQueryMatcher {
	return queryMatch([]any{"some", params})
}

type Active struct {
	PathMatcher  ActivePathMatcher
	QueryMatcher ActiveQueryMatcher
	Indicate   []Indicate
}


type AHref struct {
	Active          Active
	Indicate       []Indicate
	StopPropagation bool
	Model           any
}

func (h *AHref) active() []any {
    if h.Active.Indicate == nil || len(h.Active.Indicate) == 0 {
        return nil
    }
	if common.IsNill(h.Active.QueryMatcher) {
		h.Active.QueryMatcher = ActiveQueryAll()
	}
	if common.IsNill(h.Active.PathMatcher) {
		h.Active.PathMatcher = ActivePathFull()
	}
	return []any{h.Active.PathMatcher, h.Active.QueryMatcher, h.Active.Indicate}
}

func (h AHref) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	link, err := inst.NewLink(ctx, h.Model)
	if err != nil {
		slog.Error("href creation  error", slog.String("link_error", err.Error()))
		return
	}
	on, ok := link.ClickHandler()
	if ok {
		entry, ok := n.RegisterAttrHook(ctx, &node.AttrHook{
			Trigger: func(_ context.Context, _ http.ResponseWriter, _ *http.Request) bool {
				on()
				return false
			},
		})
		if !ok {
			return
		}
		attrs.AppendCapture(&front.LinkCapture{
			StopPropagation: h.StopPropagation,
		}, &front.Hook{
			Mode:      ModeBlock(),
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

func (s ARawSrc) init(ctx context.Context, n node.Core, inst instance.Core) (string, bool) {
	entry, ok := n.RegisterAttrHook(ctx, &node.AttrHook{
		Trigger: s.handle,
	})
	if !ok {
		return "", false
	}
	src := fmt.Sprintf("/d00r/%s/%d/%d", inst.Id(), entry.NodeId, entry.HookId)
	if s.Name != "" {
		src = src + "/" + s.Name
	}
	return src, true
}

func (s ARawSrc) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
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

// attribute to securely serve file
type ASrc struct {
	Path string
	Once bool
	Name string
}

func (s ASrc) init(ctx context.Context, n node.Core, inst instance.Core) (string, bool) {
	entry, ok := n.RegisterAttrHook(ctx, &node.AttrHook{
		Trigger: s.handle,
	})
	if !ok {
		return "", false
	}
	if s.Name == "" {
		s.Name = filepath.Base(s.Path)
	}
	link := fmt.Sprintf("/d00r/%s/%d/%d/%s", inst.Id(), entry.NodeId, entry.HookId, s.Name)
	return link, true
}

func (s ASrc) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
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

type AFileHref ASrc

func (s AFileHref) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	link, ok := (*ASrc)(&s).init(ctx, n, inst)
	if !ok {
		return
	}
	attrs.Set("href", link)
}

type ARawFileHref ARawSrc

func (s ARawFileHref) Init(ctx context.Context, n node.Core, inst instance.Core, attrs *front.Attrs) {
	link, ok := (*ARawSrc)(&s).init(ctx, n, inst)
	if !ok {
		return
	}
	attrs.Set("href", link)
}
