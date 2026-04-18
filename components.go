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
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/gox"
)

// Door is a dynamic placeholder in the DOM tree that can be updated,
// replaced, or removed after render.
//
// Doors start inactive and become active when rendered. Operations on an
// inactive door are stored virtually and applied when the door becomes active.
// If a door is removed or replaced, it becomes inactive again, but operations
// continue to update that virtual state for future rendering.
//
// The context used while rendering a door's content follows the door's
// lifecycle, which makes `ctx.Done()` safe to use in background goroutines
// that depend on the door staying mounted.
//
// X-prefixed methods return a channel that reports completion. For inactive
// doors, that channel closes immediately without sending a value.
type Door = door.Door

// Frame returns a fresh Door pointer for one-off inline dynamic fragments.
//
// It is a shorthand for `&doors.Door{}` and is useful with proxy syntax when
// the Door does not need to be stored on a struct field.
func Frame() *Door {
	return &Door{}
}

// Sub renders a dynamic fragment driven by beam.
//
// It subscribes to beam and re-renders the inner content whenever the value
// changes. Returning nil from el clears the fragment.
//
// Example:
//
//	elem demo(beam Beam[int]) {
//		~(doors.Sub(beam, elem(v int) {
//			<span>~(v)</span>
//		}))
//	}
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

// Inject renders el with the latest beam value stored in the child context
// under key.
//
// Example:
//
//	~>doors.Inject("user", userBeam) <section>
//		~{
//			user := ctx.Value("user").(User)
//		}
//		<span>~(user.Name)</span>
//	</section>
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

// Go starts f when the surrounding component is rendered.
//
// The passed context is canceled when the dynamic owner is unmounted, which
// makes [Go] a good fit for background loops that should stop with the page.
// The context is also equivalent to calling [Free] on the surrounding context,
// so it is safe to use with X* operations that should keep the current dynamic
// ownership.
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
func Go(f func(context.Context)) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		core := cur.Context().Value(ctex.KeyCore).(core.Core)
		ctx := Free(cur.Context())
		core.Runtime().Go(ctx, f)
		return nil
	})
}

// Status sets the initial HTTP status code for the current page render.
//
// Example:
//
//	~(doors.Status(http.StatusNotFound))
func Status(statusCode int) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		core := cur.Context().Value(ctex.KeyCore).(core.Core)
		core.SetStatus(statusCode)
		return nil
	})
}
