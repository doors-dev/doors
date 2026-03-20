// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/gox"
)

// Door represents a dynamic placeholder in the DOM tree that can be updated,
// replaced, or removed at runtime.
//
// It is a fundamental building block of the framework, used to manage dynamic HTML content.
// All changes made to a Door are automatically synchronized with the frontend DOM.
//
// A Door is itself a templ.Component and can be used directly in templates:
//
//		~(door)
//		// or
//		~>(door)
//		<div>
//	     Initial Content
//	 </div>
//
// Doors start inactive and become active when rendered. Operations on inactive doors
// are stored virtually and applied when the door becomes active. If a door is removed
// or replaced, it becomes inactive again, but operations continue to update its virtual
// state for potential future rendering.
//
// The context used when rendering a Door's content follows the Door's lifecycle.
// This allows you to safely use `ctx.Done()` inside background goroutines
// that depend on the Door's presence in the DOM.
//
// Extended methods (prefixed with X) return a channel that can be used to track
// when operations complete. The channel receives nil on success or an error on failure,
// then closes. For inactive doors, the channel closes immediately without sending a value.
//
// During a single render cycle, Doors and their children are guaranteed to observe
// consistent state (Beam), ensuring stable and predictable rendering.
type Door = door.Door

// Sub creates a reactive component that automatically updates when a Beam value changes.
//
// It subscribes to the Beam and re-renders the inner content whenever the value changes.
// The render function is called with the current Beam value and must return a templ.Component.
//
// This is the preferred way to bind Beam values into the DOM in a declarative and reactive manner.
//
// Example:
//
//
//	elem demo(beam Beam[int]) {
//	    ~(doors.Sub(beam, elem(v int) {
//	        <span>~(value)</span>
//	    }))
//	}
//
// Parameters:
//   - beam: the reactive Beam to observe
//   - el: a function that maps the current Beam value to a gox.Elem
//
// Returns:
//   - A templ.Component that updates reactively as the Beam value changes

func Sub[T any](beam Beam[T], el func(T) gox.Elem) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		door := &Door{}
		ok := beam.Sub(cur.Context(), func(ctx context.Context, v T) bool {
			door.Update(ctx, gox.Elem(func(cur gox.Cursor) error {
				el := el(v)
				if el == nil {
					door.Clear(ctx)
					return nil
				}
				return el(cur)
			}))
			return false
		})
		if !ok {
			return nil
		}
		return cur.Editor(door)
	})
}

// Inject creates a reactive component that injects Beam values into the context for child components.
//
// It subscribes to the Beam and re-renders its children whenever the value changes,
// making the current value available to child components.
//
// This enables passing reactive values down the component tree without explicit prop drilling.
//
// Example:
//
//	~>Inject("user", userBeam) <span> ~(ctx.Value("user").(User).Name) </span> // Can use ctx.Value("user").(User) to get current user
func Inject[T any](key any, beam Beam[T]) gox.Proxy {
	return gox.ProxyFunc(func(cur gox.Cursor, el gox.Elem) error {
		door := &Door{}
		ok := beam.Sub(cur.Context(), func(ctx context.Context, v T) bool {
			door.Rebase(ctx, func(cur gox.Cursor) error {
				ctx := context.WithValue(cur.Context(), key, v)
				cur = gox.NewCursor(ctx, cur)
				return el(cur)
			})
			return false
		})
		if !ok {
			return nil
		}
		return cur.Editor(door)
	})
}


// Go starts a goroutine at render time using a blocking-safe context tied to the component's lifecycle.
//
// The goroutine runs only if the component is rendered. The context is canceled when the component
// is unmounted, allowing for proper cleanup. You must explicitly listen to ctx.Done() to stop work.
//
// The context allows safe blocking operations, making it safe to use with X* operations (e.g., XUpdate, XRemove).
//
// Example:
//
//	@doors.Go(func(ctx context.Context) {
//	    for {
//	        select {
//	        case <-time.After(time.Second):
//	            door.Update(ctx, currentTime())
//	        case <-ctx.Done():
//	            return
//	        }
//	    }
//	})
//
// Parameters:
//   - f: a function to run in a goroutine, scoped to the component's render lifecycle
//
// Returns:
//   - A non-visual templ.Component that starts the goroutine when rendered
func Go(f func(context.Context)) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		core := cur.Context().Value(ctex.KeyCore).(core.Core)
		ctx := ctex.SetBlockingCtx(cur.Context())
		core.Runtime().Go(ctx, f)
		return nil
	})
}

// Status sets the HTTP status code
// when rendered in a template.
// Makes effect only at initial page render.
// Example: ~(doors.Status(404))
func Status(statusCode int) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		core := cur.Context().Value(ctex.KeyCore).(core.Core)
		core.SetStatus(statusCode)
		return nil
	})
}

const headScript = `
let tags = new Set(await $data("tags"))
$on(await $data("event"), (data) => {
    document.title = data.title;
    const removeTags = tags;
    tags = new Set();
    for(const [name, attrs] of Object.entries(data.meta)) {
        removeTags.delete(name);
        tags.add(name);
        let meta = document.querySelector('meta[name="'+name+'"]');
        if (!meta) {
			meta = document.createElement('meta');
			meta.setAttribute('name', name);
			document.head.appendChild(meta);
        } 
		for (const name of meta.getAttributeNames()) {
			if(name == "name") {
				continue
			}
			const expected = attrs[name];
			if(!expected) {
				meta.removeAttribute(name);
				continue
			}
			delete attrs[name];
			if(expected === true) {
				continue
			}
			if(expected !== meta.getAttribute(name)) {
				meta.setAttribute(name, expected)
			}
		}
		for(const [name, value] of Object.entries(attrs)) {
			if(value === true) {
				meta.toggleAttribute(name, true);
				continue
			}
			meta.setAttribute(name, value)
		}
    }
    for(const name of removeTags) {
        const meta = document.querySelector('meta[name="'+name+'"]');
        meta.remove();
    }
});
`

// TitleMeta renders both <title> and <meta> elements that update dynamically based on a Beam value.
//
// It outputs HTML <title> and <meta> tags, and includes the necessary script bindings
// to ensure all metadata updates reactively when the Beam changes on the server.
//
// Example:
//
//	~(doors.TitleMeta(beam, elem(p Path) {
//      <title>Product ~(p.Name)</title>
//      <meta name="description" content=("Buy "+p.Name+"at the best price")>
//      <meta name="keywords" content=(p.Name+", product, buy")>
//      <meta name="og:title" content=(p.Name)>
//      <meta name="og:description" content="Check out this amazing product">
//	}))
//
// Parameters:
//   - b: a Beam providing the input value (usually page path Beam)
//   - el: a function that rendered the title and meta elements
//
// Returns:
//   - A gox.Editor that renders title and meta elements with remote call scripts.

func TitleMeta[M any](b Beam[M], el func(M) gox.Elem) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		ctx := cur.Context()
		core := ctx.Value(ctex.KeyCore).(core.Core)
		eventName := fmt.Sprintf("head~%d", core.NewID())
		currentSeq := &atomic.Uint32{}
		gz := !core.Conf().ServerDisableGzip
		m, ok := b.ReadAndSub(cur.Context(), func(ctx context.Context, m M) bool {
			seq := currentSeq.Add(1)
			report := ctex.WgAdd(ctx)
			core.Runtime().Submit(ctx, func(ok bool) {
				defer report()
				if !ok {
					return
				}
				p := &titleMetaPrinter{
					headData: headData{
						Meta: make(map[string]map[string]any),
					},
				}
				if err := el(m).Print(ctx, p); err != nil {
					slog.Error("TitleMeta rendering error", "error", err.Error())
					return
				}
				if seq != currentSeq.Load() {
					return
				}
				builder := ActionEmit{
					Name: eventName,
					Arg:  p.headData,
				}
				action, params, err := builder.action(ctx, core, gz)
				if err != nil {
					slog.Error("head data update action building error", "error", err.Error())
					return
				}
				core.CallCheck(
					func() bool {
						return seq == currentSeq.Load()
					},
					action,
					nil,
					nil,
					params,
				)
			}, nil)
			return false
		})
		if !ok {
			return nil
		}
		p := &metaLister{
			cur: cur,
		}
		if err := el(m).Print(cur.Context(), p); err != nil {
			return err
		}
		tags := p.list
		if err := cur.Init("script"); err != nil {
			return err
		}
		if err := cur.AttrMod(AData{
			Name:  "tags",
			Value: tags,
		}); err != nil {
			return err
		}
		if err := cur.AttrMod(AData{
			Name:  "event",
			Value: eventName,
		}); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		if err := cur.Raw(headScript); err != nil {
			return err
		}
		if err := cur.Close(); err != nil {
			return err
		}
		return nil
	})
}

type metaLister struct {
	cur  gox.Cursor
	list []string
}

func (h *metaLister) Send(job gox.Job) (err error) {
	if open, ok := job.(*gox.JobHeadOpen); ok && strings.EqualFold(open.Tag, "meta") {
		if attr, ok := open.Attrs.Find("name"); ok && attr.IsSet() {
			if name, ok := attr.Value().(string); ok {
				h.list = append(h.list, name)
			}
		}
	}
	return h.cur.Send(job)
}

type headData struct {
	Title string                    `json:"title"`
	Meta  map[string]map[string]any `json:"meta"`
}

func (h *headData) setTitle(content string) {
	h.Title = content
}

func (h *headData) intoMap(a gox.Attrs) map[string]any {
	list := a.List()
	m := make(map[string]any, len(list))
	buf := &bytes.Buffer{}
	for _, attr := range list {
		if attr.Value() == nil {
			continue
		}
		if err := attr.OutputName(buf); err != nil {
			buf.Reset()
			slog.Error("attribute name render error", "error", err.Error())
			continue
		}
		name := buf.String()
		buf.Reset()
		if b, ok := attr.Value().(bool); ok {
			if !b {
				continue
			}
			m[name] = true
			continue
		}
		if err := attr.OutputValue(buf); err != nil {
			buf.Reset()
			slog.Error("attribute value render error", "error", err.Error())
			continue
		}
		m[name] = buf.String()
		buf.Reset()
	}
	return m
}

func (h *headData) addMeta(a gox.Attrs) {
	attr := a.Get("name")
	if !attr.IsSet() {
		return
	}
	name, ok := attr.Value().(string)
	if !ok {
		return
	}
	attr.Unset()
	h.Meta[name] = h.intoMap(a)
}

type titleMetaPrinter struct {
	titleID      uint64
	titleContent *bytes.Buffer
	headData     headData
	buf          *bytes.Buffer
}

func (h *titleMetaPrinter) Send(job gox.Job) (err error) {
	open, isOpen := job.(*gox.JobHeadOpen)
	if isOpen {
		defer gox.Release(open)
	}
	if h.titleID == 0 {
		switch true {
		case !isOpen:
			slog.Warn("TitleMeta contains unexpected content, it won't be updated")
		case strings.EqualFold(open.Tag, "title"):
			h.titleID = open.ID
			h.buf = &bytes.Buffer{}
		case strings.EqualFold(open.Tag, "meta"):
			h.headData.addMeta(open.Attrs)
		default:
			slog.Warn("TitleMeta contains unexpected tag, it won't be updated", "name", open.Tag)
		}
		return nil
	}
	close, isClose := job.(*gox.JobHeadClose)
	if isClose {
		defer gox.Release(close)
	}
	switch true {
	case isOpen:
		return errors.New("TitleMeta title contains unexpected content")
	case !isClose:
		return job.Output(h.buf)
	case close.ID == h.titleID:
		h.headData.setTitle(h.buf.String())
		h.titleID = 0
		return nil
	default:
		return errors.New("TitleMeta title contains unexpected content")
	}
}
