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
	"errors"
	"io"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/gox"
)

// Door is a dynamic placeholder in the DOM tree that can be updated,
// replaced, or removed after render.
//
// Doors start inactive and become active when rendered. Operations on an
// inactive door are stored virtually and applied when the door becomes active.
// If a door is removed or replaced, it becomes inactive again, but operations
// continue to update that virtual content for future rendering.
//
// The context used while rendering a door's content follows the door's
// lifecycle, which makes `ctx.Done()` safe to use in background goroutines
// that depend on the door staying mounted.
//
// X-prefixed methods return a channel that reports completion. For inactive
// doors, that channel closes immediately without sending a value.
type Door = door.Door

// Parallel renders the following element on the instance goroutine pool.
//
// Use it for fragments with database queries or external API calls to improve
// render time.
func Parallel() gox.Proxy {
	return gox.ProxyFunc(func(cur gox.Cursor, elem gox.Elem) error {
		j := parallelJob{
			ctx: cur.Context(),
			el:  elem,
		}
		return cur.Printer().Send(j)
	})
}

type parallelJob struct {
	ctx context.Context
	el  gox.Elem
}

func (pj parallelJob) Render(pip door.Pipe) {
	pip.Submit(func(cur gox.Cursor) error {
		return cur.CompCtx(pj.ctx, pj.el)
	})
}

func (p parallelJob) Context() context.Context {
	return p.ctx
}

func (parallelJob) Output(io.Writer) error {
	return errors.New("Parallel can only be used during a Doors render")
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
		core := cur.Context().Value(common.KeyCore).(core.Core)
		ctx := Free(cur.Context())
		core.Instance().Runtime().Go(ctx, f)
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
		core := cur.Context().Value(common.KeyCore).(core.Core)
		core.Instance().SetStatus(statusCode)
		return nil
	})
}
