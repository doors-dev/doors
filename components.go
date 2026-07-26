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

// Door is a dynamic part of the page that can be updated, replaced, or removed
// after render.
//
// A Door is not mounted until it is rendered, and [Door.Unmount] or removal by
// an ancestor returns it to that state. While it is not mounted, operations
// change its stored state, which the next render puts on the page.
//
// The context of content rendered inside a Door is canceled when that content
// leaves the page, which makes ctx.Done() usable in background work.
type Door = door.Door

// Parallel renders the following element on the instance goroutine pool.
//
// Use it for fragments that wait on a database query or an external API call.
func Parallel() gox.Proxy {
	return gox.ProxyFunc(func(cur gox.Cursor, elem gox.Elem) error {
		j := parallelJob{
			ctx: cur.Context(),
			el:  elem,
		}
		return cur.Printer().Send(j)
	})
}

// Ctx renders the following element with the values of ctx.
//
// Only value lookups are served from ctx; cancellation and Doors ownership stay
// with the enclosing render.
func Ctx(ctx context.Context) gox.Proxy {
	return gox.ProxyFunc(func(cur gox.Cursor, elem gox.Elem) error {
		ctx := common.NewRenderCtx(cur.Context(), ctx)
		cur = gox.NewCursor(ctx, cur.Printer())
		return elem(cur)
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
// The context passed to f is the surrounding render context through
// [DetachedContext]: it is canceled when the surrounding content leaves the
// page and keeps the current dynamic ownership. Waiting on completion channels
// inside f is safe.
//
// Example:
//
//	~(doors.Go(func(ctx context.Context) {
//	    <-time.After(time.Second)
//	    d.Inner(ctx, currentTime())
//	}))
func Go(f func(ctx context.Context)) gox.Editor {
	return gox.EditorFunc(func(cur gox.Cursor) error {
		core := cur.Context().Value(common.KeyCore).(core.Core)
		ctx := DetachedContext(cur.Context())
		core.Instance().Runtime().Go(ctx, f)
		return nil
	})
}

// Status sets the HTTP status code of the initial page response, replacing the
// default 200. Calling it after that response is sent has no effect.
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
