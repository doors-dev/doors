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

func Sub[T any](beam Beam[T], el func(T) gox.Elem) gox.EditorComp {
	return gox.EditorCompFunc(func(cur gox.Cursor) error {
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
