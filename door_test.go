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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/gox"
)

func countedElem(n *atomic.Int32, content gox.Elem) gox.Elem {
	return func(cur gox.Cursor) error {
		n.Add(1)
		return content(cur)
	}
}

func onlyNils(vals []error) bool {
	for _, v := range vals {
		if v != nil {
			return false
		}
	}
	return true
}

// On success an operation channel sends a nil once the call is scheduled and
// another once the client has applied it, then closes.
func TestDoorOpChannelScheduledThenApplied(t *testing.T) {
	ops := []struct {
		name string
		op   func(d *Door, ctx context.Context) <-chan error
	}{
		{"inner", func(d *Door, ctx context.Context) <-chan error { return d.Inner(ctx, textElem("v1")) }},
		{"outer", func(d *Door, ctx context.Context) <-chan error { return d.Outer(ctx, textElem("v1")) }},
		{"static", func(d *Door, ctx context.Context) <-chan error { return d.Static(ctx, textElem("v1")) }},
		{"reload", func(d *Door, ctx context.Context) <-chan error { return d.Reload(ctx) }},
		{"unmount", func(d *Door, ctx context.Context) <-chan error { return d.Unmount(ctx) }},
	}
	for _, c := range ops {
		t.Run(c.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			d := &Door{}
			d.Inner(context.Background(), textElem("v0"))
			pageCtx := DetachedContext(h.renderPage(mountDoor(d)))
			h.client.hold()
			ch := c.op(d, pageCtx)
			h.waitQueued(1)
			select {
			case err, ok := <-ch:
				if !ok || err != nil {
					t.Fatalf("expected the scheduled nil, got %v (open %v)", err, ok)
				}
			default:
				t.Fatal("no value once the call was scheduled")
			}
			select {
			case err, ok := <-ch:
				t.Fatalf("unexpected value %v (open %v) before the client applied the call", err, ok)
			default:
			}
			h.client.release()
			if vals := drainOp(t, ch); len(vals) != 1 || vals[0] != nil {
				t.Fatalf("expected the applied nil, then close, got %v", vals)
			}
		})
	}
}

// Static before mount stores content that is not a live container. A later
// Inner stores content for the Door's own default container.
func TestDoorStaticThenInnerStored(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	var staticRenders atomic.Int32
	d.Static(context.Background(), countedElem(&staticRenders, tagElem("section", "static-root", textElem("STATIC-BODY"))))
	d.Inner(context.Background(), textElem("INNER-BODY"))
	_, html, err := h.renderPageHTML(mountDoor(d))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "<d0-r") || !strings.Contains(html, "INNER-BODY</d0-r>") {
		t.Fatalf("expected the inner content in the default container, got %q", html)
	}
	if strings.Contains(html, "static-root") || staticRenders.Load() != 0 {
		t.Fatalf("static content was reused as the container (renders %d): %q", staticRenders.Load(), html)
	}
}

// The same after a Static on a mounted Door: when the parent renders the Door
// again, the Inner content gets the default container and the static content
// does not run again.
func TestDoorStaticMountedThenInnerParentRender(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	p := &Door{}
	d := &Door{}
	var staticRenders atomic.Int32
	d.Inner(context.Background(), textElem("d0"))
	p.Inner(context.Background(), mountDoor(d))
	pageCtx := DetachedContext(h.renderPage(mountDoor(p)))
	if vals := drainOp(t, d.Static(pageCtx, countedElem(&staticRenders, tagElem("section", "static-root", textElem("STATIC-BODY"))))); !onlyNils(vals) {
		t.Fatalf("static failed: %v", vals)
	}
	if vals := drainOp(t, d.Inner(pageCtx, textElem("INNER-BODY"))); len(vals) != 0 {
		t.Fatalf("expected the Inner on a static Door to only store content, got %v", vals)
	}
	if vals := drainOp(t, p.Reload(pageCtx)); len(vals) != 2 || !onlyNils(vals) {
		t.Fatalf("parent reload failed: %v", vals)
	}
	got := h.client.actions()
	last := got[len(got)-1]
	if last.name != "door_update" || !strings.Contains(last.body, "<d0-r") || !strings.Contains(last.body, "INNER-BODY</d0-r>") {
		t.Fatalf("expected the parent update to carry the inner content in the default container, got %+v", last)
	}
	if strings.Contains(last.body, "static-root") || staticRenders.Load() != 1 {
		t.Fatalf("static content was reused as the container (renders %d): %q", staticRenders.Load(), last.body)
	}
}

// Static stores no inner content, so mounting the Door through an empty
// proxy renders the proxy element empty.
func TestDoorStaticThenEmptyProxy(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Static(context.Background(), textElem("STATIC-BODY"))
	_, html, err := h.renderPageHTML(func(cur gox.Cursor) error {
		return d.Proxy(cur, tagElem("div", "px", nil))
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, `id="px"></div>`) || strings.Contains(html, "STATIC-BODY") {
		t.Fatalf("expected an empty proxy element, got %q", html)
	}
}

// Outer followed by Inner before the Outer reaches the client delivers both:
// the replace, then the update into the new container.
func TestDoorOuterThenInnerKeepsBoth(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	pageCtx := DetachedContext(h.renderPage(mountDoor(d)))
	h.client.hold()
	outer := d.Outer(pageCtx, tagElem("section", "new-outer", textElem("outer-content")))
	inner := d.Inner(pageCtx, textElem("inner-content"))
	for name, ch := range map[string]<-chan error{"Outer": outer, "Inner": inner} {
		select {
		case err, ok := <-ch:
			if !ok || err != nil {
				t.Fatalf("expected the %s to be scheduled, got %v (open %v)", name, err, ok)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("the %s was never scheduled", name)
		}
	}
	h.client.release()
	if vals := drainOp(t, outer); len(vals) != 1 || vals[0] != nil {
		t.Fatalf("expected the Outer to be applied, got %v", vals)
	}
	if vals := drainOp(t, inner); len(vals) != 1 || vals[0] != nil {
		t.Fatalf("expected the Inner to be applied, got %v", vals)
	}
	got := h.client.actions()
	if len(got) != 2 {
		t.Fatalf("expected two applied calls, got %+v", got)
	}
	if got[0].name != "door_replace" || !strings.Contains(got[0].body, `id="new-outer"`) {
		t.Fatalf("expected the Outer replace first, got %+v", got[0])
	}
	if got[1].name != "door_update" || got[1].id != got[0].id || got[1].body != "inner-content" {
		t.Fatalf("expected the Inner update into the replaced container, got %+v", got[1])
	}
}

// Calls made while Door.Static content renders, and from its OnReady, reach
// the client after the static markup.
func TestDoorStaticCallsAfterMarkup(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	pageCtx := DetachedContext(h.renderPage(mountDoor(d)))
	h.client.hold()
	ch := d.Static(pageCtx, func(cur gox.Cursor) error {
		Call(cur.Context(), ActionScroll{Selector: "#render-call"})
		OnReady(cur.Context(), func(ctx context.Context) {
			Call(ctx, ActionScroll{Selector: "#ready-call"})
		})
		return tagElem("section", "static-root", textElem("static"))(cur)
	})
	h.waitQueued(3)
	h.client.release()
	if vals := drainOp(t, ch); len(vals) != 2 || !onlyNils(vals) {
		t.Fatalf("expected the static to succeed, got %v", vals)
	}
	h.waitIdle()
	got := h.client.actions()
	if len(got) != 3 || got[0].name != "door_replace" || !strings.Contains(got[0].body, "static-root") {
		t.Fatalf("expected the static markup before the calls it made, got %+v", got)
	}
	calls := got[1].body + got[2].body
	if !strings.Contains(calls, "#render-call") || !strings.Contains(calls, "#ready-call") {
		t.Fatalf("expected the render and OnReady calls after the markup, got %+v", got)
	}
}

// A Door.Static from a beam watcher on the parent's screen, while a newer
// update of the same source is already queued on that screen, settles.
func TestDoorStaticFromWatcherWithUpdateInFlight(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	x := NewSource(0)
	d := &Door{}
	d.Inner(context.Background(), textElem("d0"))
	secondIssued := make(chan struct{})
	staticCh := make(chan (<-chan error), 1)
	var fired atomic.Bool
	pageCtx := DetachedContext(h.renderPage(func(cur gox.Cursor) error {
		if err := cur.Comp(d); err != nil {
			return err
		}
		x.Sub(cur.Context(), func(ctx context.Context, v int) bool {
			if v == 0 || !fired.CompareAndSwap(false, true) {
				return false
			}
			<-secondIssued
			staticCh <- d.Static(ctx, textElem("static"))
			return false
		})
		return nil
	}))
	first := x.Update(pageCtx, 1)
	second := x.Update(pageCtx, 2)
	close(secondIssued)
	var static <-chan error
	select {
	case static = <-staticCh:
	case <-time.After(5 * time.Second):
		t.Fatal("watcher never issued the static")
	}
	drainOp(t, first)
	drainOp(t, second)
	if vals := drainOp(t, static); len(vals) != 2 || !onlyNils(vals) {
		t.Fatalf("expected the static to succeed, got %v", vals)
	}
}
