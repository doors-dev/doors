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
	"io"
	"sync/atomic"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/door2"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/instance"
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
//	@door
//	// or
//	@door {
//	    // initial HTML content
//	}
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
type Door = door2.Door

type editorFunc func(cur gox.Cursor) error

func (e editorFunc) Use(cur gox.Cursor) error {
	return e(cur)
}

// Sub creates a reactive component that automatically updates when a Beam value changes.
//
// It subscribes to the Beam and re-renders the inner content whenever the value changes.
// The render function is called with the current Beam value and must return a templ.Component.
//
// This is the preferred way to bind Beam values into the DOM in a declarative and reactive manner.
//
// Example:
//
//	templ display(value int) {
//	    <span>{strconv.Itoa(value)}</span>
//	}
//
//	templ demo(beam Beam[int]) {
//	    @doors.Sub(beam, func(v int) templ.Component {
//	        return display(v)
//	    })
//	}
//
// Parameters:
//   - beam: the reactive Beam to observe
//   - render: a function that maps the current Beam value to a templ.Component
//
// Returns:
//   - A templ.Component that updates reactively as the Beam value changes

func Sub[T any](beam Beam[T], el func(T) gox.Elem) gox.Editor {
	return editorFunc(func(cur gox.Cursor) error {
		door := &Door{}
		ok := beam.Sub(cur.Context(), func(ctx context.Context, v T) bool {
			door.Update(ctx, gox.Elem(func(cur gox.Cursor) error {
				return el(v)(cur)
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
//	@Inject("user", userBeam) {
//	    @UserProfile() // Can use ctx.Value("user").(User) to get current user
//	}
func Inject[T any](key any, beam Beam[T], content gox.Comp) gox.Editor {
	return editorFunc(func(cur gox.Cursor) error {
		door := &Door{}
		ok := beam.Sub(cur.Context(), func(ctx context.Context, v T) bool {
			door.Update(ctx, gox.Elem(func(cur gox.Cursor) error {
				ctx := context.WithValue(cur.Context(), key, v)
				cur = gox.NewCursor(ctx, cur)
				return content.Main()(cur)
			}))
			return false
		})
		if !ok {
			return nil
		}
		return cur.Editor(door)
	})
}

// If shows children if the beam value is true
func If(beam Beam[bool]) templ.Component {
	return E(func(ctx context.Context) templ.Component {
		children := templ.GetChildren(ctx)
		ctx = templ.ClearChildren(ctx)
		door := &Door{}
		ok := beam.Sub(ctx, func(ctx context.Context, v bool) bool {
			if !v {
				door.Clear(ctx)
				return false
			}
			door.Update(ctx, children)
			return false
		})
		if !ok {
			return nil
		}
		return door
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
	return editorFunc(func(cur gox.Cursor) error {
		return nil
	})
}

// Status sets the HTTP status code
// when rendered in a template.
// Makes effect only at initial page render.
// Example: ~(doors.Status(404))
func Status(statusCode int) gox.Editor {
	return editorFunc(func(cur gox.Cursor) error {
		core := cur.Context().Value(ctex.KeyCore).(core.Core)
		core.SetStatus(statusCode)
		return nil
	})
}

// HeadData represents page metadata including title and meta tags
type HeadData struct {
	Title string
	Meta  map[string]string
}

type headUsed struct{}

const headScript = `
let tags = new Set($data("tags"))
$on("d00r_head", (data) => {
    document.title = data.title;
    const removeTags = tags
    tags = new Set()
    for(const [name, content] of Object.entries(data.meta)) {
        removeTags.delete(name)
        tags.add(name)
        let meta = document.querySelector('meta[name="'+name+'"]');
        if (meta) {
            meta.setAttribute('content', content);
            continue
        } 
        meta = document.createElement('meta');
        meta.setAttribute('name', name);
        meta.setAttribute('content', content);
        document.head.appendChild(meta);
    }
    for(const name of removeTags) {
        const meta = document.querySelector('meta[name="'+name+'"]');
        meta.remove();
    }
});
`

// Head renders both <title> and <meta> elements that update dynamically based on a Beam value.
//
// It outputs HTML <title> and <meta> tags, and includes the necessary script bindings
// to ensure all metadata updates reactively when the Beam changes on the server.
//
// Example:
//
//	@doors.Head(beam, func(p Path) HeadData {
//	    return HeadData{
//	        Title: "Product: " + p.Name,
//	        Meta: map[string]string{
//	            "description": "Buy " + p.Name + " at the best price",
//	            "keywords": p.Name + ", product, buy",
//	            "og:title": p.Name,
//	            "og:description": "Check out this amazing product",
//	        },
//	    }
//	})
//
// Parameters:
//   - b: a Beam providing the input value (usually page path Beam)
//   - cast: a function that maps the Beam value to a HeadData struct.
//
// Returns:
//   - A templ.Component that renders title and meta elements with remote call scripts.
func Head[M any](b Beam[M], cast func(M) HeadData) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, ok := InstanceSave(ctx, headUsed{}, headUsed{}).(headUsed)
		if ok {
			return nil
		}
		inst := ctx.Value(common.CtxKeyInstance).(instance.Core)
		door := ctx.Value(common.CtxKeyDoor).(door.Core)
		currentSeq := &atomic.Uint32{}
		m, ok := b.ReadAndSub(ctx, func(ctx context.Context, m M) bool {
			seq := currentSeq.Add(1)
			report := ctxwg.Add(ctx)
			ok := inst.Spawn(func() {
				defer report()
				newData := cast(m)
				if seq != currentSeq.Load() {
					return
				}
				inst.CallCheck(
					func() bool {
						return seq == currentSeq.Load()
					},
					&action.Emit{
						Name: "d00r_head",
						Arg: map[string]any{
							"title": newData.Title,
							"meta": func() map[string]string {
								escapedTags := make(map[string]string, len(newData.Meta))
								for k, v := range newData.Meta {
									escapedTags[k] = templ.EscapeString(v)
								}
								return escapedTags
							}(),
						},
						DoorID: door.Id(),
					},
					nil,
					nil,
					action.CallParams{},
				)
			})
			if !ok {
				report()
			}
			return false
		})
		if !ok {
			return nil
		}
		headData := cast(m)
		tags := make([]string, len(headData.Meta))
		i := 0
		for k := range headData.Meta {
			tags[i] = k
			i++
		}
		_, err := fmt.Fprintf(w, "<title>%s</title>", templ.EscapeString(headData.Title))
		if err != nil {
			return err
		}
		for name, content := range headData.Meta {
			_, err := fmt.Fprintf(w, "<meta name=\"%s\" content=\"%s\"/>", templ.EscapeString(name), templ.EscapeString(content))
			if err != nil {
				return err
			}
		}
		content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
			err := AData{
				Name:  "tags",
				Value: tags,
			}.Render(ctx, w)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "<script>%s</script>", headScript)
			return err

		})
		childenCtx := templ.WithChildren(ctx, content)
		return Script().Render(childenCtx, w)
	})
}

func inlineName(attr templ.Attributes, ext string) string {
	name := "inline"
	dataName, ok := attr["data-name"]
	if ok {
		dataNameStr, ok := dataName.(string)
		if ok {
			name = dataNameStr
		}
	}
	return name + "." + ext
}
