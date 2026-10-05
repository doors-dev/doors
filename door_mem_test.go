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
	"runtime"
	"testing"
	"time"
	"weak"

	"github.com/doors-dev/gox"
)

type bigContent struct {
	data []byte
}

func newBigContent() *bigContent {
	return &bigContent{data: make([]byte, 1<<20)}
}

func (b *bigContent) Main() gox.Elem {
	return func(cur gox.Cursor) error {
		return cur.Text("big")
	}
}

func collected(p weak.Pointer[bigContent]) bool {
	for range 20 {
		runtime.GC()
		if p.Value() == nil {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

func applied(t *testing.T, ch <-chan error) {
	t.Helper()
	if vals := drainOp(t, ch); !onlyNils(vals) {
		t.Fatalf("operation failed: %v", vals)
	}
}

// A Door keeps its content only until it is replaced.
func TestDoorReplacedContentCollectable(t *testing.T) {
	cases := map[string]func(t *testing.T, d *Door, ctx context.Context){
		"inner": func(t *testing.T, d *Door, ctx context.Context) {
			applied(t, d.Inner(ctx, textElem("small")))
		},
		"outer": func(t *testing.T, d *Door, ctx context.Context) {
			applied(t, d.Outer(ctx, textElem("small")))
		},
		"outer-outer": func(t *testing.T, d *Door, ctx context.Context) {
			applied(t, d.Outer(ctx, textElem("small")))
			applied(t, d.Outer(ctx, textElem("small")))
		},
		"outer-reload": func(t *testing.T, d *Door, ctx context.Context) {
			applied(t, d.Outer(ctx, textElem("small")))
			applied(t, d.Reload(ctx))
		},
		"static": func(t *testing.T, d *Door, ctx context.Context) {
			applied(t, d.Static(ctx, textElem("small")))
		},
		"unmount": func(t *testing.T, d *Door, ctx context.Context) {
			applied(t, d.Unmount(ctx))
			applied(t, d.Inner(ctx, textElem("small")))
		},
	}
	for name, replace := range cases {
		t.Run(name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			d := &Door{}
			pageCtx := DetachedContext(h.renderPage(mountDoor(d)))
			big := newBigContent()
			p := weak.Make(big)
			applied(t, d.Inner(pageCtx, big))
			big = nil
			replace(t, d, pageCtx)
			h.waitIdle()
			if !collected(p) {
				t.Fatal("replaced content is still reachable")
			}
			runtime.KeepAlive(d)
		})
	}
}

// Items appended with the Door.Static and next Door recipe are not kept: the
// server holds only the growth edge.
func TestDoorFeedItemsCollectable(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	tail := &Door{}
	pageCtx := DetachedContext(h.renderPage(mountDoor(tail)))
	var items []weak.Pointer[bigContent]
	for range 5 {
		big := newBigContent()
		items = append(items, weak.Make(big))
		next := &Door{}
		applied(t, tail.Static(pageCtx, func(cur gox.Cursor) error {
			if err := cur.Comp(big); err != nil {
				return err
			}
			return cur.Comp(next)
		}))
		tail = next
	}
	h.waitIdle()
	for i, item := range items {
		if !collected(item) {
			t.Fatalf("feed item %d is still reachable", i)
		}
	}
	runtime.KeepAlive(tail)
}

// A Bind does not keep the values it rendered before.
func TestBindPreviousValuesCollectable(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	v0 := newBigContent()
	v1 := newBigContent()
	p0 := weak.Make(v0)
	p1 := weak.Make(v1)
	src := NewSource(v0)
	v0 = nil
	pageCtx := DetachedContext(h.renderPage(src.Bind(func(v *bigContent) gox.Elem { return v.Main() })))
	applied(t, src.Update(pageCtx, v1))
	v1 = nil
	v2 := &bigContent{}
	applied(t, src.Update(pageCtx, v2))
	h.waitIdle()
	if !collected(p0) {
		t.Fatal("the initial value is still reachable")
	}
	if !collected(p1) {
		t.Fatal("the replaced value is still reachable")
	}
	runtime.KeepAlive(src)
	runtime.KeepAlive(v2)
}
