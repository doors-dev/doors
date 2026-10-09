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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/gox"
)

// heavyGate renders heavy content that waits while the gate is shut. Each
// render that has to wait reports on entered first.
type heavyGate struct {
	mu      sync.Mutex
	gate    chan struct{}
	entered chan struct{}
	renders atomic.Int32
}

func newHeavyGate(t *testing.T) *heavyGate {
	g := &heavyGate{gate: make(chan struct{}), entered: make(chan struct{}, 16)}
	close(g.gate)
	t.Cleanup(g.open)
	return g
}

func (g *heavyGate) shut() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.gate = make(chan struct{})
}

func (g *heavyGate) open() {
	g.mu.Lock()
	defer g.mu.Unlock()
	select {
	case <-g.gate:
	default:
		close(g.gate)
	}
}

func (g *heavyGate) elem(text string) gox.Elem {
	return func(cur gox.Cursor) error {
		g.renders.Add(1)
		g.mu.Lock()
		gate := g.gate
		g.mu.Unlock()
		select {
		case <-gate:
		default:
			g.entered <- struct{}{}
			<-gate
		}
		return cur.Text(text)
	}
}

func (g *heavyGate) waitEntered(t *testing.T) {
	t.Helper()
	select {
	case <-g.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("the heavy render never started")
	}
}

// expectNotEntered fails if a heavy render starts within d.
func (g *heavyGate) expectNotEntered(t *testing.T, d time.Duration) {
	t.Helper()
	select {
	case <-g.entered:
		t.Fatal("the heavy render started before the previous content was ready")
	case <-time.After(d):
	}
}

type deferFunc = func(ctx context.Context, content any) <-chan error

// deferredMode is a package deferred op, the element it wraps heavy in, and
// the call heavy lands as.
type deferredMode struct {
	name  string
	op    deferFunc
	wrap  func(el gox.Elem) gox.Elem
	final actionWant
}

var deferredModes = []deferredMode{
	{"inner", DeferredInner, func(el gox.Elem) gox.Elem { return el }, actionWant{"door_update", "HEAVY"}},
	{"outer", DeferredOuter, func(el gox.Elem) gox.Elem { return tagElem("section", "heavy-outer", el) }, actionWant{"door_replace", `id="heavy-outer"`}},
	{"static", DeferredStatic, func(el gox.Elem) gox.Elem { return tagElem("section", "heavy-static", el) }, actionWant{"door_replace", `id="heavy-static"`}},
}

// deferLoader defers heavy with op on its own render context, hands the op
// channel to chs and renders skeleton.
func deferLoader(op deferFunc, heavy any, skeleton gox.Elem, chs chan<- (<-chan error)) gox.Elem {
	return func(cur gox.Cursor) error {
		chs <- op(cur.Context(), heavy)
		return skeleton(cur)
	}
}

// documentLoader is the IsDocument loader idiom: heavy inline in the page
// response, otherwise a deferred heavy behind a skeleton.
func documentLoader(heavy gox.Elem, chs chan<- (<-chan error)) gox.Elem {
	return func(cur gox.Cursor) error {
		if IsDocument(cur.Context()) {
			return heavy(cur)
		}
		chs <- DeferredInner(cur.Context(), heavy)
		return cur.Text("SKELETON")
	}
}

func recvOp(t *testing.T, chs <-chan (<-chan error)) <-chan error {
	t.Helper()
	select {
	case ch := <-chs:
		return ch
	case <-time.After(5 * time.Second):
		t.Fatal("the loader never deferred")
		return nil
	}
}

// expectScheduled reads the scheduled nil of an operation channel.
func expectScheduled(t *testing.T, name string, ch <-chan error) {
	t.Helper()
	select {
	case err, ok := <-ch:
		if !ok || err != nil {
			t.Fatalf("%s: expected the scheduled nil, got %v (open %v)", name, err, ok)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: never scheduled", name)
	}
}

// expectOp drains an operation channel and matches the values against want.
func expectOp(t *testing.T, name string, ch <-chan error, want ...error) {
	t.Helper()
	got := drainOp(t, ch)
	ok := len(got) == len(want)
	for i := 0; ok && i < len(want); i++ {
		ok = want[i] == nil && got[i] == nil || want[i] != nil && errors.Is(got[i], want[i])
	}
	if !ok {
		t.Fatalf("%s: got %v, want %v", name, got, want)
	}
}

// actionWant matches an applied action by name and a body substring.
type actionWant struct {
	name string
	body string
}

func expectActions(t *testing.T, got []lifecycleAction, want ...actionWant) {
	t.Helper()
	ok := len(got) == len(want)
	for i := 0; ok && i < len(want); i++ {
		ok = got[i].name == want[i].name && strings.Contains(got[i].body, want[i].body)
	}
	if !ok {
		t.Fatalf("expected actions %+v, got %+v", want, got)
	}
}

func contentCtxElem(ctxs chan<- context.Context, text string) gox.Elem {
	return func(cur gox.Cursor) error {
		ctxs <- cur.Context()
		return cur.Text(text)
	}
}

// expectNoPendingEvent fails if an event has already happened.
func expectNoPendingEvent(t *testing.T, h *lifecycleHarness) {
	t.Helper()
	select {
	case got := <-h.events:
		t.Fatalf("unexpected event %q", got)
	default:
	}
}

// waitEventBeforeReport waits for the event want without reading ch and fails
// if ch reports first.
func waitEventBeforeReport(t *testing.T, h *lifecycleHarness, want string, ch <-chan error) {
	t.Helper()
	timeout := time.After(5 * time.Second)
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case got := <-h.events:
			if got != want {
				t.Fatalf("expected event %q, got %q", want, got)
			}
			return
		case <-tick.C:
			if len(ch) == 0 {
				continue
			}
			select {
			case got := <-h.events:
				if got != want {
					t.Fatalf("expected event %q, got %q", want, got)
				}
				return
			case <-time.After(time.Second):
				t.Fatalf("the op reported before event %q", want)
			}
		case <-timeout:
			t.Fatalf("timed out waiting for event %q", want)
		}
	}
}

// expectOpError drains an operation channel and expects a single err with
// the given message prefix.
func expectOpError(t *testing.T, name string, ch <-chan error, err error, prefix string) {
	t.Helper()
	got := drainOp(t, ch)
	if len(got) != 1 || !errors.Is(got[0], err) || !strings.HasPrefix(got[0].Error(), prefix) {
		t.Fatalf("%s: expected one %q error wrapping %v, got %v", name, prefix, err, got)
	}
}

// A fresh Door placed by a parent update defers from its loader: the parent
// update carries the skeleton and a door_update with heavy follows it, never
// before, even with a fast heavy and the client held.
func TestDeferredPlacedByParentUpdate(t *testing.T) {
	for i := range 10 {
		h := newLifecycleHarness(t, 8)
		chs := make(chan (<-chan error), 8)
		p := &Door{}
		p.Inner(context.Background(), textElem("P0"))
		d := &Door{}
		d.Inner(context.Background(), deferLoader(DeferredInner, textElem("HEAVY"), textElem("SKELETON"), chs))
		h.renderPage(mountDoor(p))
		h.client.hold()
		pch := p.Inner(context.Background(), mountDoor(d))
		ch := recvOp(t, chs)
		h.waitQueued(2)
		h.client.release()
		expectOp(t, "parent", pch, nil, nil)
		expectOp(t, "deferred", ch, nil, nil)
		got := h.client.actions()
		expectActions(t, got, actionWant{"door_update", "SKELETON"}, actionWant{"door_update", "HEAVY"})
		if strings.Contains(got[0].body, "HEAVY") || got[0].id == got[1].id {
			t.Fatalf("run %d: expected the parent update with the skeleton, then heavy into the child, got %+v", i, got)
		}
	}
}

// A Door placed by a parent update defers from its loader while the parent
// still renders after it: the deferred render starts only once the parent
// update is scheduled.
func TestDeferredWaitsForParentCall(t *testing.T) {
	for _, m := range deferredModes {
		t.Run(m.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			chs := make(chan (<-chan error), 1)
			p := &Door{}
			d := &Door{}
			p.Inner(context.Background(), textElem("P0"))
			d.Inner(context.Background(), deferLoader(m.op, m.wrap(func(cur gox.Cursor) error {
				h.events <- "heavy"
				return cur.Text("HEAVY")
			}), textElem("SKELETON"), chs))
			h.renderPage(mountDoor(p))
			g.shut()
			pch := p.Inner(context.Background(), func(cur gox.Cursor) error {
				if err := cur.Comp(d); err != nil {
					return err
				}
				return g.elem("TAIL")(cur)
			})
			ch := recvOp(t, chs)
			g.waitEntered(t)
			h.expectNoEvent(50 * time.Millisecond)
			g.open()
			h.waitEvent("heavy")
			expectOp(t, "parent", pch, nil, nil)
			expectOp(t, "deferred", ch, nil, nil)
			h.waitIdle()
			got := h.client.actions()
			expectActions(t, got, actionWant{"door_update", "SKELETON"}, m.final)
			if !strings.Contains(got[0].body, "TAIL") || got[0].id == got[1].id {
				t.Fatalf("expected the parent update, then heavy into the child, got %+v", got)
			}
		})
	}
}

// A loader that defers during the page render leaves the skeleton in the page
// HTML; heavy follows as a door_update.
func TestDeferredFromPageRender(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	chs := make(chan (<-chan error), 8)
	d := &Door{}
	d.Inner(context.Background(), deferLoader(DeferredInner, textElem("HEAVY"), textElem("SKELETON"), chs))
	_, html, err := h.renderPageHTML(mountDoor(d))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "SKELETON") || strings.Contains(html, "HEAVY") {
		t.Fatalf("expected the skeleton in the page, got %q", html)
	}
	expectOp(t, "deferred", recvOp(t, chs), nil, nil)
	h.waitIdle()
	expectActions(t, h.client.actions(), actionWant{"door_update", "HEAVY"})
}

// d.Inner with a loader on a mounted inner Door, per deferred mode. Heavy
// starts only once the skeleton call is queued, even while the loader still
// renders after deferring. The skeleton update is delivered when it was sent
// before heavy is rendered, and dropped at send time when it was only queued
// (the loader op reports context.Canceled after its scheduled nil).
func TestDeferredSkeletonSentOrQueued(t *testing.T) {
	for _, m := range deferredModes {
		t.Run(m.name+"/sent", func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			chs := make(chan (<-chan error), 8)
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			h.renderPage(mountDoor(d))
			g.shut()
			ch1 := d.Inner(context.Background(), deferLoader(m.op, m.wrap(g.elem("HEAVY")), textElem("SKELETON"), chs))
			expectOp(t, "loader", ch1, nil, nil)
			ch2 := recvOp(t, chs)
			g.waitEntered(t)
			g.open()
			expectOp(t, "deferred", ch2, nil, nil)
			h.waitIdle()
			got := h.client.actions()
			expectActions(t, got, actionWant{"door_update", "SKELETON"}, m.final)
			if got[0].id != got[1].id {
				t.Fatalf("expected both calls on the same Door, got %+v", got)
			}
		})
		t.Run(m.name+"/queued", func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			chs := make(chan (<-chan error), 8)
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			h.renderPage(mountDoor(d))
			h.client.hold()
			g.shut()
			proceed := make(chan struct{})
			ch1 := d.Inner(context.Background(), func(cur gox.Cursor) error {
				chs <- m.op(cur.Context(), m.wrap(g.elem("HEAVY")))
				<-proceed
				return cur.Text("SKELETON")
			})
			ch2 := recvOp(t, chs)
			g.expectNotEntered(t, 50*time.Millisecond)
			close(proceed)
			expectScheduled(t, "loader", ch1)
			h.waitQueued(1)
			g.waitEntered(t)
			g.open()
			expectScheduled(t, "deferred", ch2)
			h.waitQueued(2)
			h.client.release()
			expectOp(t, "loader", ch1, context.Canceled)
			expectOp(t, "deferred", ch2, nil)
			h.waitIdle()
			expectActions(t, h.client.actions(), m.final)
		})
	}
}

// DeferredInner over a Door whose current op is Outer-mode or proxy-mode
// always delivers the skeleton, with the client held and a fast heavy: the
// skeleton replace is ordered first and its call context is the container's.
func TestDeferredInnerOverOuterKeepsSkeleton(t *testing.T) {
	t.Run("outer", func(t *testing.T) {
		for range 5 {
			h := newLifecycleHarness(t, 8)
			chs := make(chan (<-chan error), 8)
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			h.renderPage(mountDoor(d))
			h.client.hold()
			ch1 := d.Outer(context.Background(), deferLoader(DeferredInner, textElem("HEAVY"), tagElem("section", "box", textElem("SKELETON")), chs))
			ch2 := recvOp(t, chs)
			h.waitQueued(2)
			h.client.release()
			expectOp(t, "outer", ch1, nil, nil)
			expectOp(t, "deferred", ch2, nil, nil)
			got := h.client.actions()
			expectActions(t, got, actionWant{"door_replace", "SKELETON"}, actionWant{"door_update", "HEAVY"})
			if !strings.Contains(got[0].body, `id="box"`) || got[0].id != got[1].id {
				t.Fatalf("expected heavy into the skeleton container, got %+v", got)
			}
		}
	})
	t.Run("bind", func(t *testing.T) {
		for range 5 {
			h := newLifecycleHarness(t, 8)
			chs := make(chan (<-chan error), 8)
			src := NewSource(0)
			h.renderPage(src.Bind(func(v int) gox.Elem {
				if v == 0 {
					return textElem("BIND0")
				}
				return deferLoader(DeferredInner, textElem("HEAVY"), textElem("SKELETON"), chs)
			}))
			h.client.hold()
			uch := src.Update(context.Background(), 1)
			ch2 := recvOp(t, chs)
			h.waitQueued(2)
			h.client.release()
			expectOp(t, "update", uch, nil)
			expectOp(t, "deferred", ch2, nil, nil)
			h.waitIdle()
			got := h.client.actions()
			expectActions(t, got, actionWant{"door_replace", "SKELETON"}, actionWant{"door_update", "HEAVY"})
			if got[0].id != got[1].id {
				t.Fatalf("expected heavy into the bind container, got %+v", got)
			}
		}
	})
	t.Run("proxy", func(t *testing.T) {
		for range 5 {
			h := newLifecycleHarness(t, 8)
			chs := make(chan (<-chan error), 8)
			d := &Door{}
			_, html, err := h.renderPageHTML(func(cur gox.Cursor) error {
				return d.Proxy(cur, tagElem("div", "px", documentLoader(textElem("HEAVY"), chs)))
			})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(html, "HEAVY") || len(chs) != 0 {
				t.Fatalf("expected heavy inline in the page, got %q", html)
			}
			h.client.hold()
			rch := d.Reload(context.Background())
			ch2 := recvOp(t, chs)
			h.waitQueued(2)
			h.client.release()
			expectOp(t, "reload", rch, nil, nil)
			expectOp(t, "deferred", ch2, nil, nil)
			got := h.client.actions()
			expectActions(t, got, actionWant{"door_replace", "SKELETON"}, actionWant{"door_update", "HEAVY"})
			if !strings.Contains(got[0].body, `id="px"`) || got[0].id != got[1].id {
				t.Fatalf("expected heavy into the proxy element, got %+v", got)
			}
		}
	})
}

// DeferredOuter replaces the container and leaves a live Door that takes
// further updates in the new container.
func TestDeferredOuterLiveContainer(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	chs := make(chan (<-chan error), 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("V0"))
	h.renderPage(mountDoor(d))
	ch1 := d.Inner(context.Background(), deferLoader(DeferredOuter, tagElem("section", "heavy-outer", textElem("HEAVY")), textElem("SKELETON"), chs))
	expectScheduled(t, "loader", ch1)
	drainOp(t, ch1)
	expectOp(t, "deferred", recvOp(t, chs), nil, nil)
	expectOp(t, "after", d.Inner(context.Background(), textElem("AFTER")), nil, nil)
	h.waitIdle()
	got := h.client.actions()
	if len(got) < 2 {
		t.Fatalf("expected the new container and the update, got %+v", got)
	}
	last := got[len(got)-1]
	prev := got[len(got)-2]
	if prev.name != "door_replace" || !strings.Contains(prev.body, `id="heavy-outer"`) || last.name != "door_update" || last.body != "AFTER" || last.id != prev.id {
		t.Fatalf("expected the new container, then an update into it, got %+v", got)
	}
}

// DeferredStatic lands heavy as static content after the skeleton, holding the
// parent's calls until it is placed. Later ops on the Door only change its
// stored state, which the parent renders the next time.
func TestDeferredStaticLaterOpsStoreState(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	g := newHeavyGate(t)
	chs := make(chan (<-chan error), 8)
	pctxs := make(chan context.Context, 8)
	p := &Door{}
	d := &Door{}
	d.Inner(context.Background(), textElem("V0"))
	p.Inner(context.Background(), func(cur gox.Cursor) error {
		pctxs <- cur.Context()
		return cur.Comp(d)
	})
	h.renderPage(mountDoor(p))
	pctx := <-pctxs
	g.shut()
	ch1 := d.Inner(context.Background(), deferLoader(DeferredStatic, tagElem("section", "heavy-static", g.elem("HEAVY")), textElem("SKELETON"), chs))
	expectOp(t, "loader", ch1, nil, nil)
	ch2 := recvOp(t, chs)
	g.waitEntered(t)
	callCh := Call(DetachedContext(pctx), ActionScroll{Selector: "#parent-call"})
	g.open()
	expectOp(t, "deferred", ch2, nil, nil)
	expectOp(t, "parent call", callCh, nil)
	h.waitIdle()
	got := h.client.actions()
	expectActions(t, got,
		actionWant{"door_update", "SKELETON"},
		actionWant{"door_replace", `id="heavy-static"`},
		actionWant{"scroll", "#parent-call"},
	)
	if got[1].id != got[0].id {
		t.Fatalf("expected the static replace on the Door, got %+v", got)
	}
	expectOp(t, "inner on static", d.Inner(context.Background(), textElem("STORED")))
	expectOp(t, "reload on static", d.Reload(context.Background()))
	expectOp(t, "unmount on static", d.Unmount(context.Background()))
	h.waitIdle()
	if n := len(h.client.actions()); n != len(got) {
		t.Fatalf("expected no calls from ops on a static Door, got %+v", h.client.actions())
	}
	expectOp(t, "parent reload", p.Reload(context.Background()), nil, nil)
	h.waitIdle()
	all := h.client.actions()
	last := all[len(all)-1]
	if last.name != "door_update" || !strings.Contains(last.body, "STORED") || strings.Contains(last.body, "HEAVY") {
		t.Fatalf("expected the parent to render the stored content, got %+v", last)
	}
}

// A newer op issued while the deferred heavy render is still running. A
// non-deferred op that replaces the content wins: the skeleton is cleaned, the
// deferred op reports context.Canceled, and neither the skeleton nor heavy is
// sent after the newer content. A newer deferred op, and Inner over a pending
// DeferredOuter, waits until heavy is scheduled: heavy is kept when the newer
// content goes into its new container, otherwise dropped if not yet sent. Over
// a pending DeferredStatic the Door is static: newer ops only store state and
// heavy lands.
func TestDeferredSupersededWhileRendering(t *testing.T) {
	newer := []struct {
		name     string
		op       func(d *Door, ctx context.Context) <-chan error
		final    actionWant
		keeps    bool
		deferred bool
	}{
		{"inner", func(d *Door, ctx context.Context) <-chan error { return d.Inner(ctx, textElem("NEWER")) }, actionWant{"door_update", "NEWER"}, true, false},
		{"outer", func(d *Door, ctx context.Context) <-chan error {
			return d.Outer(ctx, tagElem("section", "newer-outer", textElem("NEWER")))
		}, actionWant{"door_replace", `id="newer-outer"`}, false, false},
		{"reload", func(d *Door, ctx context.Context) <-chan error { return d.Reload(ctx) }, actionWant{}, false, false},
		{"unmount", func(d *Door, ctx context.Context) <-chan error { return d.Unmount(ctx) }, actionWant{"door_replace", ""}, false, false},
		{"static", func(d *Door, ctx context.Context) <-chan error { return d.Static(ctx, textElem("NEWER")) }, actionWant{"door_replace", "NEWER"}, false, false},
		{"deferred inner", func(d *Door, ctx context.Context) <-chan error { return d.DeferredInner(ctx, textElem("NEWER")) }, actionWant{"door_update", "NEWER"}, true, true},
		{"deferred outer", func(d *Door, ctx context.Context) <-chan error {
			return d.DeferredOuter(ctx, tagElem("section", "newer-outer", textElem("NEWER")))
		}, actionWant{"door_replace", `id="newer-outer"`}, false, true},
		{"deferred static", func(d *Door, ctx context.Context) <-chan error {
			return d.DeferredStatic(ctx, tagElem("section", "newer-static", textElem("NEWER")))
		}, actionWant{"door_replace", `id="newer-static"`}, false, true},
	}
	for _, m := range deferredModes {
		for _, c := range newer {
			t.Run(m.name+"/"+c.name, func(t *testing.T) {
				h := newLifecycleHarness(t, 8)
				g := newHeavyGate(t)
				chs := make(chan (<-chan error), 8)
				d := &Door{}
				d.Inner(context.Background(), textElem("V0"))
				h.renderPage(mountDoor(d))
				g.shut()
				ch1 := d.Inner(context.Background(), func(cur gox.Cursor) error {
					OnClean(cur.Context(), func() { h.events <- "clean-skeleton" })
					chs <- m.op(cur.Context(), m.wrap(g.elem("HEAVY")))
					return cur.Text("SKELETON")
				})
				expectOp(t, "loader", ch1, nil, nil)
				ch2 := recvOp(t, chs)
				g.waitEntered(t)
				nch := c.op(d, context.Background())
				want := []actionWant{{"door_update", "SKELETON"}}
				kept := m.name == "outer" && c.keeps
				switch {
				case m.name == "static":
					expectOp(t, c.name, nch)
					h.expectNoEvent(50 * time.Millisecond)
					g.open()
					h.waitEvent("clean-skeleton")
					expectOp(t, "deferred", ch2, nil, nil)
					want = append(want, m.final)
				case c.deferred || kept:
					h.expectNoEvent(50 * time.Millisecond)
					h.client.hold()
					g.open()
					h.waitEvent("clean-skeleton")
					h.waitQueued(2)
					h.client.release()
					expectOp(t, c.name, nch, nil, nil)
					if kept {
						expectOp(t, "deferred", ch2, nil, nil)
						want = append(want, m.final, c.final)
					} else {
						expectOp(t, "deferred", ch2, nil, context.Canceled)
						want = append(want, c.final)
					}
				default:
					h.waitEvent("clean-skeleton")
					g.open()
					expectOp(t, "deferred", ch2, context.Canceled)
					expectOp(t, c.name, nch, nil, nil)
					if c.final.name == "" {
						want = append(want, m.final)
					} else {
						want = append(want, c.final)
					}
				}
				h.waitIdle()
				got := h.client.actions()
				expectActions(t, got, want...)
				for i, w := range want {
					if w.body == "" && got[i].body != "" || i > 0 && got[i].id != got[0].id {
						t.Fatalf("expected the calls on the Door, an unmount empty, got %+v", got)
					}
				}
				h.expectNoEvent(50 * time.Millisecond)
			})
		}
	}
}

// Deferred ops queued behind a blocked deferred render run one at a time: each
// starts only once the previous content is rendered and its call scheduled,
// then replaces it. A replaced call still unsent is dropped, unless the newer
// content goes into its new container.
func TestDeferredChain(t *testing.T) {
	same := func(el gox.Elem) gox.Elem { return el }
	section := func(id string) func(gox.Elem) gox.Elem {
		return func(el gox.Elem) gox.Elem { return tagElem("section", id, el) }
	}
	chains := []struct {
		name   string
		second func(d *Door, ctx context.Context, content any) <-chan error
		wrap2  func(gox.Elem) gox.Elem
		third  func(d *Door, ctx context.Context, content any) <-chan error
		wrap3  func(gox.Elem) gox.Elem
		want2  []error
		final  []actionWant
	}{
		{"inner inner", (*Door).DeferredInner, same, (*Door).DeferredInner, same, []error{nil, context.Canceled}, []actionWant{{"door_update", "H3"}}},
		{"outer inner", (*Door).DeferredOuter, section("h2"), (*Door).DeferredInner, same, []error{nil, nil}, []actionWant{{"door_replace", `id="h2">H2`}, {"door_update", "H3"}}},
		{"outer outer", (*Door).DeferredOuter, section("h2"), (*Door).DeferredOuter, section("h3"), []error{nil, context.Canceled}, []actionWant{{"door_replace", `id="h3">H3`}}},
	}
	for _, c := range chains {
		t.Run(c.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g1 := newHeavyGate(t)
			g2 := newHeavyGate(t)
			d := &Door{}
			d.Inner(context.Background(), func(cur gox.Cursor) error {
				OnClean(cur.Context(), func() { h.events <- "clean-old" })
				return cur.Text("OLD")
			})
			h.renderPage(mountDoor(d))
			g1.shut()
			g2.shut()
			ch1 := d.DeferredInner(context.Background(), g1.elem("H1"))
			g1.waitEntered(t)
			ch2 := c.second(d, context.Background(), c.wrap2(g2.elem("H2")))
			ch3 := c.third(d, context.Background(), c.wrap3(func(cur gox.Cursor) error {
				h.events <- "h3-start"
				return cur.Text("H3")
			}))
			g2.expectNotEntered(t, 50*time.Millisecond)
			expectNoPendingEvent(t, h)
			h.client.hold()
			g1.open()
			h.waitEvent("clean-old")
			select {
			case <-g2.entered:
			case got := <-h.events:
				t.Fatalf("expected the second render to start first, got event %q", got)
			case <-time.After(5 * time.Second):
				t.Fatal("the second render never started")
			}
			h.expectNoEvent(50 * time.Millisecond)
			g2.open()
			h.waitEvent("h3-start")
			h.waitQueued(3)
			h.client.release()
			expectOp(t, "first", ch1, nil, context.Canceled)
			expectOp(t, "second", ch2, c.want2...)
			expectOp(t, "third", ch3, nil, nil)
			h.waitIdle()
			expectActions(t, h.client.actions(), c.final...)
			h.expectNoEvent(50 * time.Millisecond)
		})
	}
}

// A deferred op issued while the plain op before it has not started rendering
// yet, an Inner waiting for its container or an Outer waiting for the Door's
// placement, starts only once that op is rendered and its call scheduled, so
// the deferred content lands last.
func TestDeferredWaitsForPendingPlainOp(t *testing.T) {
	heavy := func(h *lifecycleHarness) gox.Elem {
		return func(cur gox.Cursor) error {
			h.events <- "heavy"
			return cur.Text("HEAVY")
		}
	}
	t.Run("inner waiting for outer", func(t *testing.T) {
		h := newLifecycleHarness(t, 8)
		g := newHeavyGate(t)
		d := &Door{}
		d.Inner(context.Background(), textElem("V0"))
		h.renderPage(mountDoor(d))
		g.shut()
		och := d.Outer(context.Background(), tagElem("section", "o1", g.elem("OUTER")))
		g.waitEntered(t)
		ich := d.Inner(context.Background(), textElem("SKEL"))
		dch := d.DeferredInner(context.Background(), heavy(h))
		h.expectNoEvent(50 * time.Millisecond)
		h.client.hold()
		g.open()
		h.waitEvent("heavy")
		h.waitQueued(3)
		h.client.release()
		expectOp(t, "outer", och, nil, nil)
		expectOp(t, "inner", ich, nil, context.Canceled)
		expectOp(t, "deferred", dch, nil, nil)
		h.waitIdle()
		expectActions(t, h.client.actions(), actionWant{"door_replace", `id="o1"`}, actionWant{"door_update", "HEAVY"})
	})
	t.Run("outer waiting for placement", func(t *testing.T) {
		h := newLifecycleHarness(t, 8)
		place := newHeavyGate(t)
		g := newHeavyGate(t)
		p := &Door{}
		d := &Door{}
		p.Inner(context.Background(), textElem("P0"))
		h.renderPage(mountDoor(p))
		h.client.hold()
		place.shut()
		g.shut()
		d.Inner(context.Background(), place.elem("PLACED"))
		pch := p.Inner(context.Background(), mountDoor(d))
		place.waitEntered(t)
		och := d.Outer(context.Background(), tagElem("section", "o1", g.elem("OUTER")))
		dch := d.DeferredInner(context.Background(), heavy(h))
		place.open()
		g.waitEntered(t)
		h.expectNoEvent(50 * time.Millisecond)
		g.open()
		h.waitEvent("heavy")
		h.waitQueued(3)
		h.client.release()
		expectOp(t, "parent", pch, nil, nil)
		expectOp(t, "outer", och, nil, nil)
		expectOp(t, "deferred", dch, nil, nil)
		h.waitIdle()
		expectActions(t, h.client.actions(), actionWant{"door_update", "PLACED"}, actionWant{"door_replace", `id="o1"`}, actionWant{"door_update", "HEAVY"})
	})
}

// A child Door placed by both the replaced content and the deferred content
// stays alive and on the page while the deferred content renders: its old
// placement is cleaned with the replaced content and no removal is sent.
func TestDeferredKeepsChildPlacedByBoth(t *testing.T) {
	for _, m := range deferredModes {
		t.Run(m.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			chs := make(chan (<-chan error), 1)
			var renders atomic.Int32
			c := &Door{}
			c.Inner(context.Background(), func(cur gox.Cursor) error {
				if renders.Add(1) == 1 {
					OnClean(cur.Context(), func() { h.events <- "clean-child" })
				}
				return cur.Text("CHILD")
			})
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			h.renderPage(mountDoor(d))
			g.shut()
			heavy := m.wrap(func(cur gox.Cursor) error {
				if err := cur.Comp(c); err != nil {
					return err
				}
				return g.elem("HEAVY")(cur)
			})
			ch1 := d.Inner(context.Background(), func(cur gox.Cursor) error {
				chs <- m.op(cur.Context(), heavy)
				if err := cur.Text("SKELETON"); err != nil {
					return err
				}
				return cur.Comp(c)
			})
			expectOp(t, "loader", ch1, nil, nil)
			ch2 := recvOp(t, chs)
			g.waitEntered(t)
			h.expectNoEvent(100 * time.Millisecond)
			expectActions(t, h.client.actions(), actionWant{"door_update", "SKELETON"})
			g.open()
			h.waitEvent("clean-child")
			expectOp(t, "deferred", ch2, nil, nil)
			h.waitIdle()
			expectActions(t, h.client.actions(), actionWant{"door_update", "SKELETON"}, m.final)
		})
	}
}

// A DeferredInner or a plain Inner queued on a pending DeferredOuter, both
// waiting for the content they replace to render, is superseded by an op that
// replaces the container: the replaced content is canceled at once, not when
// its render ends, and only the newer op lands.
func TestDeferredInnerOnPendingOuterSuperseded(t *testing.T) {
	middle := []struct {
		name string
		op   func(d *Door, ctx context.Context, content any) <-chan error
	}{
		{"deferred inner", (*Door).DeferredInner},
		{"inner", (*Door).Inner},
	}
	newer := []struct {
		name  string
		op    func(d *Door, ctx context.Context) <-chan error
		final actionWant
	}{
		{"outer", func(d *Door, ctx context.Context) <-chan error {
			return d.Outer(ctx, tagElem("section", "newer-outer", textElem("NEWER")))
		}, actionWant{"door_replace", `id="newer-outer"`}},
		{"static", func(d *Door, ctx context.Context) <-chan error { return d.Static(ctx, textElem("NEWER")) }, actionWant{"door_replace", "NEWER"}},
		{"unmount", func(d *Door, ctx context.Context) <-chan error { return d.Unmount(ctx) }, actionWant{"door_replace", ""}},
	}
	for _, mid := range middle {
		for _, c := range newer {
			t.Run(mid.name+"/"+c.name, func(t *testing.T) {
				h := newLifecycleHarness(t, 8)
				g := newHeavyGate(t)
				d := &Door{}
				d.Inner(context.Background(), textElem("V0"))
				h.renderPage(mountDoor(d))
				g.shut()
				ctxs := make(chan context.Context, 1)
				ch0 := d.Inner(context.Background(), func(cur gox.Cursor) error {
					ctxs <- cur.Context()
					return g.elem("OLD")(cur)
				})
				g.waitEntered(t)
				old := <-ctxs
				ch1 := d.DeferredOuter(context.Background(), tagElem("section", "h1", textElem("H1")))
				ch2 := mid.op(d, context.Background(), textElem("H2"))
				nch := c.op(d, context.Background())
				select {
				case <-old.Done():
				case <-time.After(5 * time.Second):
					t.Fatal("the replaced content was not canceled")
				}
				g.open()
				expectOp(t, "old", ch0, context.Canceled)
				expectOp(t, "deferred outer", ch1, context.Canceled)
				expectOp(t, mid.name, ch2, context.Canceled)
				expectOp(t, c.name, nch, nil, nil)
				h.waitIdle()
				expectActions(t, h.client.actions(), c.final)
			})
		}
	}
}

// An Inner or a Reload over a DeferredInner that replaces Outer content
// cleans that content once the Outer call is scheduled, before the newer
// content renders: an Outer still rendering keeps its container call, and a
// dropped deferred render still running does not keep the content alive.
func TestDeferredInnerOverOuterSuperseded(t *testing.T) {
	for _, reload := range []bool{false, true} {
		name := "inner"
		if reload {
			name = "reload"
		}
		newer := func(d *Door, content gox.Elem) <-chan error {
			if reload {
				return d.Reload(context.Background())
			}
			return d.Inner(context.Background(), content)
		}
		t.Run(name+"/outer rendering", func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			h.renderPage(mountDoor(d))
			g.shut()
			och := d.Outer(context.Background(), tagElem("section", "box", g.elem("OUTER")))
			g.waitEntered(t)
			dch := d.DeferredInner(context.Background(), textElem("DEFERRED"))
			nch := newer(d, textElem("NEWER"))
			g.open()
			expectOp(t, "outer", och, nil, nil)
			expectOp(t, "deferred", dch, context.Canceled)
			expectOp(t, name, nch, nil, nil)
			final := "NEWER"
			if reload {
				final = "DEFERRED"
			}
			h.waitIdle()
			expectActions(t, h.client.actions(), actionWant{"door_replace", `id="box">OUTER`}, actionWant{"door_update", final})
		})
		t.Run(name+"/deferred rendering", func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			var renders atomic.Int32
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			h.renderPage(mountDoor(d))
			och := d.Outer(context.Background(), tagElem("section", "box", func(cur gox.Cursor) error {
				OnClean(cur.Context(), func() { h.events <- "clean-outer" })
				return cur.Text("OUTER")
			}))
			expectOp(t, "outer", och, nil, nil)
			g.shut()
			dch := d.DeferredInner(context.Background(), func(cur gox.Cursor) error {
				if renders.Add(1) == 1 {
					return g.elem("HEAVY")(cur)
				}
				h.events <- "newer-render"
				return cur.Text("HEAVY")
			})
			g.waitEntered(t)
			nch := newer(d, func(cur gox.Cursor) error {
				h.events <- "newer-render"
				return cur.Text("NEWER")
			})
			h.waitEvent("clean-outer")
			h.waitEvent("newer-render")
			g.open()
			expectOp(t, "deferred", dch, context.Canceled)
			expectOp(t, name, nch, nil, nil)
			final := "NEWER"
			if reload {
				final = "HEAVY"
			}
			h.waitIdle()
			expectActions(t, h.client.actions(), actionWant{"door_replace", `id="box">OUTER`}, actionWant{"door_update", final})
		})
	}
}

// A plain op over a chain of pending deferred ops of mixed kinds, with or
// without a plain op queued on them, cancels the content the chain waits for
// at once, not when its render ends: every op in the chain reports
// context.Canceled and only the newer op lands.
func TestDeferredPendingChainSuperseded(t *testing.T) {
	type op = func(d *Door) <-chan error
	chain := func(ops ...op) []op { return ops }
	inner := func(text string) op {
		return func(d *Door) <-chan error { return d.DeferredInner(context.Background(), textElem(text)) }
	}
	outer := func(id string) op {
		return func(d *Door) <-chan error {
			return d.DeferredOuter(context.Background(), tagElem("section", id, textElem(id)))
		}
	}
	reload := func(d *Door) <-chan error { return d.Reload(context.Background()) }
	newInner := func(d *Door) <-chan error { return d.Inner(context.Background(), textElem("NEWER")) }
	newOuter := func(d *Door) <-chan error {
		return d.Outer(context.Background(), tagElem("section", "newer-outer", textElem("NEWER")))
	}
	newStatic := func(d *Door) <-chan error { return d.Static(context.Background(), textElem("NEWER")) }
	cases := []struct {
		name  string
		chain []op
		newer op
		final actionWant
	}{
		{"inner inner/inner", chain(inner("H1"), inner("H2")), newInner, actionWant{"door_update", "NEWER"}},
		{"inner inner/reload", chain(inner("H1"), inner("H2")), reload, actionWant{"door_update", "H2"}},
		{"outer outer/outer", chain(outer("h1"), outer("h2")), newOuter, actionWant{"door_replace", `id="newer-outer"`}},
		{"outer outer/reload", chain(outer("h1"), outer("h2")), reload, actionWant{"door_replace", `id="h2"`}},
		{"inner outer/static", chain(inner("H1"), outer("h2")), newStatic, actionWant{"door_replace", "NEWER"}},
		{"outer inner inner/outer", chain(outer("h1"), inner("H2"), inner("H3")), newOuter, actionWant{"door_replace", `id="newer-outer"`}},
		{"outer inner reload/static", chain(outer("h1"), inner("H2"), reload), newStatic, actionWant{"door_replace", "NEWER"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			h.renderPage(mountDoor(d))
			g.shut()
			ctxs := make(chan context.Context, 1)
			ch0 := d.Inner(context.Background(), func(cur gox.Cursor) error {
				ctxs <- cur.Context()
				return g.elem("OLD")(cur)
			})
			g.waitEntered(t)
			old := <-ctxs
			var chs []<-chan error
			for _, issue := range c.chain {
				chs = append(chs, issue(d))
			}
			nch := c.newer(d)
			select {
			case <-old.Done():
			case <-time.After(5 * time.Second):
				t.Fatal("the replaced content was not canceled")
			}
			g.open()
			expectOp(t, "old", ch0, context.Canceled)
			for i, ch := range chs {
				expectOp(t, fmt.Sprintf("chain %d", i), ch, context.Canceled)
			}
			expectOp(t, "newer", nch, nil, nil)
			h.waitIdle()
			expectActions(t, h.client.actions(), c.final)
		})
	}
}

// A parent that removes or rerenders the Door while the deferred heavy render
// is running wins: the skeleton is cleaned, the deferred op reports
// context.Canceled, and the rerender places the stored heavy content inline.
func TestDeferredTargetReplacedByParent(t *testing.T) {
	for _, m := range deferredModes {
		for _, reload := range []bool{false, true} {
			name := m.name + "/remove"
			if reload {
				name = m.name + "/reload"
			}
			t.Run(name, func(t *testing.T) {
				h := newLifecycleHarness(t, 8)
				g := newHeavyGate(t)
				chs := make(chan (<-chan error), 8)
				p := &Door{}
				d := &Door{}
				d.Inner(context.Background(), textElem("V0"))
				p.Inner(context.Background(), mountDoor(d))
				h.renderPage(mountDoor(p))
				g.shut()
				ch1 := d.Inner(context.Background(), func(cur gox.Cursor) error {
					OnClean(cur.Context(), func() { h.events <- "clean-skeleton" })
					chs <- m.op(cur.Context(), m.wrap(g.elem("HEAVY")))
					return cur.Text("SKELETON")
				})
				expectOp(t, "loader", ch1, nil, nil)
				ch2 := recvOp(t, chs)
				g.waitEntered(t)
				var pch <-chan error
				if reload {
					pch = p.Reload(context.Background())
				} else {
					pch = p.Inner(context.Background(), textElem("P1"))
				}
				h.waitEvent("clean-skeleton")
				g.open()
				expectOp(t, "deferred", ch2, context.Canceled)
				expectOp(t, "parent", pch, nil, nil)
				h.waitIdle()
				got := h.client.actions()
				expectActions(t, got, actionWant{"door_update", "SKELETON"}, actionWant{"door_update", ""})
				body := got[1].body
				if got[1].id == got[0].id || !reload && body != "P1" || reload && (!strings.Contains(body, m.final.body) || strings.Contains(body, "SKELETON")) {
					t.Fatalf("expected the parent update without the skeleton, got %+v", got)
				}
				h.expectNoEvent(50 * time.Millisecond)
			})
		}
	}
}

// A Door moved to another parent while a DeferredOuter on it waits for the
// content it replaces or renders: that content is canceled once the new
// parent's update is scheduled, not when the pending render ends, and the old
// container is removed.
func TestDeferredOuterTargetMoved(t *testing.T) {
	for _, rendering := range []bool{false, true} {
		name := "pending"
		if rendering {
			name = "rendering"
		}
		t.Run(name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			ctxs := make(chan context.Context, 1)
			var renders atomic.Int32
			p := &Door{}
			q := &Door{}
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			p.Inner(context.Background(), mountDoor(d))
			q.Inner(context.Background(), textElem("Q0"))
			h.renderPage(func(cur gox.Cursor) error {
				if err := cur.Comp(p); err != nil {
					return err
				}
				return cur.Comp(q)
			})
			g.shut()
			ch0 := d.Inner(context.Background(), func(cur gox.Cursor) error {
				ctxs <- cur.Context()
				if rendering {
					return cur.Text("OLD")
				}
				return g.elem("OLD")(cur)
			})
			old := <-ctxs
			var want []actionWant
			if rendering {
				expectOp(t, "old", ch0, nil, nil)
				want = append(want, actionWant{"door_update", "OLD"})
			} else {
				g.waitEntered(t)
			}
			ch1 := d.DeferredOuter(context.Background(), tagElem("section", "moved", func(cur gox.Cursor) error {
				if rendering && renders.Add(1) == 1 {
					return g.elem("HEAVY")(cur)
				}
				return cur.Text("HEAVY")
			}))
			if rendering {
				g.waitEntered(t)
			}
			h.client.hold()
			qch := q.Inner(context.Background(), mountDoor(d))
			select {
			case <-old.Done():
			case <-time.After(5 * time.Second):
				t.Fatal("the replaced content was not canceled")
			}
			h.waitQueued(2)
			h.client.release()
			g.open()
			if !rendering {
				expectOp(t, "old", ch0, context.Canceled)
			}
			expectOp(t, "deferred outer", ch1, context.Canceled)
			expectOp(t, "move", qch, nil, nil)
			h.waitIdle()
			want = append(want, actionWant{"door_update", `id="moved">HEAVY`}, actionWant{"door_replace", ""})
			got := h.client.actions()
			expectActions(t, got, want...)
			if last := got[len(got)-1]; last.body != "" || last.id == got[len(got)-2].id {
				t.Fatalf("expected the move into the new parent, then the old container removed, got %+v", got)
			}
		})
	}
}

// The package funcs target the Door that rendered ctx: after a newer op on
// that Door they report context.Canceled and leave it alone.
func TestDeferredPackageFuncLostCAS(t *testing.T) {
	ops := map[string]deferFunc{"inner": DeferredInner, "outer": DeferredOuter, "static": DeferredStatic}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			ctxs := make(chan context.Context, 8)
			d := &Door{}
			d.Inner(context.Background(), contentCtxElem(ctxs, "V0"))
			h.renderPage(mountDoor(d))
			dctx := <-ctxs
			ich := d.Inner(context.Background(), textElem("V1"))
			expectOp(t, "deferred", op(dctx, textElem("LATE")), context.Canceled)
			expectOp(t, "inner", ich, nil, nil)
			h.waitIdle()
			expectActions(t, h.client.actions(), actionWant{"door_update", "V1"})
		})
	}
}

// The package funcs on the page-level root context send one error and close.
func TestDeferredPackageFuncRootContext(t *testing.T) {
	ops := map[string]deferFunc{"inner": DeferredInner, "outer": DeferredOuter, "static": DeferredStatic}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			pageCtx := h.renderPage(nil)
			vals := drainOp(t, op(pageCtx, textElem("X")))
			if len(vals) != 1 || vals[0] == nil || vals[0].Error() != "root door cannot be deferred" {
				t.Fatalf("expected one root error, got %v", vals)
			}
			h.waitIdle()
			if got := h.client.actions(); len(got) != 0 {
				t.Fatalf("expected no calls, got %+v", got)
			}
		})
	}
}

// The package funcs on the context of a Door its parent removed close without
// a value and store the content in their mode, which shows when the Door is
// mounted again: inner children in a d0-r container, an outer container that
// is the Door itself, static content inline with no Door.
func TestDeferredPackageFuncUnmountedTarget(t *testing.T) {
	ops := []struct {
		name    string
		op      deferFunc
		content gox.Elem
		want    func(body string) bool
	}{
		{"inner", DeferredInner, textElem("STORED"), func(body string) bool {
			return strings.HasPrefix(body, "<d0-r ") && strings.HasSuffix(body, ">STORED</d0-r>") && !strings.Contains(body, "<section")
		}},
		{"outer", DeferredOuter, tagElem("section", "stored-outer", textElem("STORED")), func(body string) bool {
			return strings.HasPrefix(body, "<section ") && strings.Contains(body, ` data-d0r="`) && strings.Contains(body, ` id="stored-outer"`) &&
				strings.HasSuffix(body, ">STORED</section>") && !strings.Contains(body, "<d0-r")
		}},
		{"static", DeferredStatic, tagElem("section", "stored-static", textElem("STORED")), func(body string) bool {
			return body == `<section id="stored-static">STORED</section>`
		}},
	}
	for _, c := range ops {
		t.Run(c.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			ctxs := make(chan context.Context, 8)
			p := &Door{}
			d := &Door{}
			d.Inner(context.Background(), contentCtxElem(ctxs, "V0"))
			p.Inner(context.Background(), mountDoor(d))
			h.renderPage(mountDoor(p))
			dctx := <-ctxs
			expectOp(t, "parent away", p.Inner(context.Background(), textElem("P1")), nil, nil)
			expectOp(t, "deferred", c.op(dctx, c.content))
			expectOp(t, "parent back", p.Inner(context.Background(), mountDoor(d)), nil, nil)
			h.waitIdle()
			got := h.client.actions()
			expectActions(t, got, actionWant{"door_update", "P1"}, actionWant{"door_update", "STORED"})
			if !c.want(got[1].body) {
				t.Fatalf("expected the parent to mount the stored %s content, got %q", c.name, got[1].body)
			}
		})
	}
}

// The package funcs on a context of the Door's container, here from a
// modifier of its Proxy element, target the Door like its content context:
// it stays live after DeferredInner and DeferredOuter, and only stores later
// ops after DeferredStatic.
func TestDeferredPackageFuncContainerContext(t *testing.T) {
	for _, m := range deferredModes {
		t.Run(m.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			ctxs := make(chan context.Context, 8)
			d := &Door{}
			h.renderPage(func(cur gox.Cursor) error {
				return d.Proxy(cur, func(cur gox.Cursor) error {
					if err := cur.Init("div"); err != nil {
						return err
					}
					if err := cur.Modify(gox.ModifyFunc(func(ctx context.Context, _ string, _ gox.Attrs) error {
						ctxs <- ctx
						return nil
					})); err != nil {
						return err
					}
					if err := cur.Submit(); err != nil {
						return err
					}
					if err := cur.Text("V0"); err != nil {
						return err
					}
					return cur.Close()
				})
			})
			expectOp(t, "deferred", m.op(<-ctxs, m.wrap(textElem("HEAVY"))), nil, nil)
			want := []actionWant{m.final}
			after := d.Inner(context.Background(), textElem("AFTER"))
			if m.name == "static" {
				expectOp(t, "after", after)
			} else {
				expectOp(t, "after", after, nil, nil)
				want = append(want, actionWant{"door_update", "AFTER"})
			}
			h.waitIdle()
			expectActions(t, h.client.actions(), want...)
		})
	}
}

// The Door methods keep the current content alive until the new content is
// rendered; d.Inner, the control, cleans it before rendering.
func TestDeferredDoorMethodsKeepOldContent(t *testing.T) {
	cases := []struct {
		name     string
		op       func(d *Door, ctx context.Context, content any) <-chan error
		heavy    func(g *heavyGate) gox.Elem
		final    actionWant
		deferred bool
	}{
		{"deferred inner", (*Door).DeferredInner, func(g *heavyGate) gox.Elem { return g.elem("HEAVY") }, actionWant{"door_update", "HEAVY"}, true},
		{"deferred outer", (*Door).DeferredOuter, func(g *heavyGate) gox.Elem { return tagElem("section", "heavy-outer", g.elem("HEAVY")) }, actionWant{"door_replace", `id="heavy-outer"`}, true},
		{"deferred static", (*Door).DeferredStatic, func(g *heavyGate) gox.Elem { return tagElem("section", "heavy-static", g.elem("HEAVY")) }, actionWant{"door_replace", `id="heavy-static"`}, true},
		{"inner", (*Door).Inner, func(g *heavyGate) gox.Elem { return g.elem("HEAVY") }, actionWant{"door_update", "HEAVY"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			ctxs := make(chan context.Context, 8)
			var cleaned atomic.Bool
			d := &Door{}
			d.Inner(context.Background(), func(cur gox.Cursor) error {
				ctxs <- cur.Context()
				OnClean(cur.Context(), func() {
					cleaned.Store(true)
					h.events <- "clean-old"
				})
				return cur.Text("OLD")
			})
			h.renderPage(mountDoor(d))
			oldCtx := <-ctxs
			g.shut()
			ch := c.op(d, context.Background(), c.heavy(g))
			g.waitEntered(t)
			if c.deferred {
				if oldCtx.Err() != nil || cleaned.Load() {
					t.Fatal("the old content was cleaned before the new content was rendered")
				}
				select {
				case v, ok := <-ch:
					t.Fatalf("unexpected value %v (open %v) before the new content was rendered", v, ok)
				default:
				}
			} else if oldCtx.Err() == nil || !cleaned.Load() {
				t.Fatal("expected Inner to clean the old content before rendering")
			}
			g.open()
			h.waitEvent("clean-old")
			expectOp(t, c.name, ch, nil, nil)
			h.waitIdle()
			expectActions(t, h.client.actions(), c.final)
		})
	}
}

// The skeleton stays live while heavy renders, whatever op rendered it: its
// context, hooks and subscriptions work, and its OnClean fires only after
// heavy is rendered. Heavy starts only once the skeleton call is queued, even
// while the loader still renders after deferring.
func TestDeferredSkeletonAliveDuringHeavy(t *testing.T) {
	kinds := []struct {
		name  string
		first actionWant
		setup func(loader gox.Elem) (page gox.Elem, trigger func() <-chan error, want []error)
	}{
		{"inner", actionWant{"door_update", "SKELETON"}, func(loader gox.Elem) (gox.Elem, func() <-chan error, []error) {
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			return mountDoor(d), func() <-chan error { return d.Inner(context.Background(), loader) }, []error{nil, nil}
		}},
		{"outer", actionWant{"door_replace", `id="box">SKELETON`}, func(loader gox.Elem) (gox.Elem, func() <-chan error, []error) {
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			return mountDoor(d), func() <-chan error {
				return d.Outer(context.Background(), tagElem("section", "box", loader))
			}, []error{nil, nil}
		}},
		{"bind", actionWant{"door_replace", "SKELETON"}, func(loader gox.Elem) (gox.Elem, func() <-chan error, []error) {
			src := NewSource(0)
			return src.Bind(func(v int) gox.Elem {
				if v == 0 {
					return textElem("V0")
				}
				return loader
			}), func() <-chan error { return src.Update(context.Background(), 1) }, []error{nil}
		}},
		{"proxy", actionWant{"door_replace", `id="px">SKELETON`}, func(loader gox.Elem) (gox.Elem, func() <-chan error, []error) {
			d := &Door{}
			return func(cur gox.Cursor) error {
				return d.Proxy(cur, tagElem("div", "px", loader))
			}, func() <-chan error { return d.Reload(context.Background()) }, []error{nil, nil}
		}},
	}
	for _, k := range kinds {
		t.Run(k.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			chs := make(chan (<-chan error), 8)
			ctxs := make(chan context.Context, 8)
			hooks := make(chan uint64, 8)
			proceed := make(chan struct{})
			src := NewSource(0)
			heavy := gox.Elem(func(cur gox.Cursor) error {
				if h.client.queued() == 0 {
					h.events <- "heavy-before-skeleton"
				}
				if err := g.elem("HEAVY")(cur); err != nil {
					return err
				}
				h.events <- "heavy-rendered"
				return nil
			})
			loader := gox.Elem(func(cur gox.Cursor) error {
				ctx := cur.Context()
				if IsDocument(ctx) {
					return cur.Text("V0")
				}
				ctxs <- ctx
				OnClean(ctx, func() { h.events <- "clean-skeleton" })
				src.Sub(ctx, func(_ context.Context, v int) bool {
					if v != 0 {
						h.events <- fmt.Sprintf("sub-%d", v)
					}
					return false
				})
				hook, ok := ctx.Value(common.KeyCore).(core.Core).Door().RegisterHook(func(context.Context, http.ResponseWriter, *http.Request) bool {
					h.events <- "hook"
					return false
				}, RaceSerial)
				if !ok {
					return errors.New("hook registration failed")
				}
				hooks <- hook.HookID
				chs <- DeferredInner(ctx, heavy)
				<-proceed
				return cur.Text("SKELETON")
			})
			page, trigger, want := k.setup(loader)
			h.renderPage(page)
			h.client.hold()
			g.shut()
			tch := trigger()
			ch2 := recvOp(t, chs)
			skeletonCtx := <-ctxs
			hookID := <-hooks
			g.expectNotEntered(t, 50*time.Millisecond)
			close(proceed)
			h.waitQueued(1)
			g.waitEntered(t)
			h.client.release()
			h.waitIdle()
			if skeletonCtx.Err() != nil {
				t.Fatal("the skeleton context was canceled while heavy renders")
			}
			expectNoPendingEvent(t, h)
			expectOp(t, "update", src.Update(context.Background(), 1), nil)
			h.waitEvent("sub-1")
			if !h.root.TriggerHook(hookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), 0) {
				t.Fatal("the skeleton hook was gone while heavy renders")
			}
			h.waitEvent("hook")
			g.open()
			h.waitEvent("heavy-rendered")
			h.waitEvent("clean-skeleton")
			expectOp(t, "deferred", ch2, nil, nil)
			expectOp(t, k.name, tch, want...)
			if skeletonCtx.Err() == nil {
				t.Fatal("expected the skeleton context canceled once heavy was rendered")
			}
			if h.root.TriggerHook(hookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), 0) {
				t.Fatal("the skeleton hook outlived the skeleton")
			}
			expectOp(t, "update after", src.Update(context.Background(), 2))
			h.expectNoEvent(50 * time.Millisecond)
			h.waitIdle()
			got := h.client.actions()
			expectActions(t, got, k.first, actionWant{"door_update", "HEAVY"})
			if got[0].id != got[1].id || strings.Contains(got[0].body, "HEAVY") {
				t.Fatalf("expected heavy into the skeleton container, got %+v", got)
			}
		})
	}
}

// The IsDocument loader idiom puts heavy inline in the page response with no
// deferred op; a later update shows the skeleton, then heavy.
func TestDeferredIsDocumentLoader(t *testing.T) {
	t.Run("reload", func(t *testing.T) {
		h := newLifecycleHarness(t, 8)
		g := newHeavyGate(t)
		chs := make(chan (<-chan error), 8)
		d := &Door{}
		d.Inner(context.Background(), documentLoader(g.elem("HEAVY"), chs))
		_, html, err := h.renderPageHTML(mountDoor(d))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(html, "HEAVY") || strings.Contains(html, "SKELETON") {
			t.Fatalf("expected heavy inline in the page, got %q", html)
		}
		h.waitIdle()
		if got := h.client.actions(); len(got) != 0 || len(chs) != 0 {
			t.Fatalf("expected no deferred op from the page render, got %+v", got)
		}
		g.shut()
		expectOp(t, "reload", d.Reload(context.Background()), nil, nil)
		ch := recvOp(t, chs)
		g.open()
		expectOp(t, "deferred", ch, nil, nil)
		h.waitIdle()
		expectActions(t, h.client.actions(), actionWant{"door_update", "SKELETON"}, actionWant{"door_update", "HEAVY"})
	})
	t.Run("parent", func(t *testing.T) {
		h := newLifecycleHarness(t, 8)
		chs := make(chan (<-chan error), 8)
		p := &Door{}
		d := &Door{}
		d.Inner(context.Background(), documentLoader(textElem("HEAVY"), chs))
		p.Inner(context.Background(), mountDoor(d))
		_, html, err := h.renderPageHTML(mountDoor(p))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(html, "HEAVY") || strings.Contains(html, "SKELETON") {
			t.Fatalf("expected heavy inline in the page, got %q", html)
		}
		h.waitIdle()
		if got := h.client.actions(); len(got) != 0 || len(chs) != 0 {
			t.Fatalf("expected no deferred op from the page render, got %+v", got)
		}
		h.client.hold()
		pch := p.Reload(context.Background())
		ch := recvOp(t, chs)
		h.waitQueued(2)
		h.client.release()
		expectOp(t, "parent reload", pch, nil, nil)
		expectOp(t, "deferred", ch, nil, nil)
		got := h.client.actions()
		expectActions(t, got, actionWant{"door_update", "SKELETON"}, actionWant{"door_update", "HEAVY"})
		if strings.Contains(got[0].body, "HEAVY") || got[0].id == got[1].id {
			t.Fatalf("expected the parent update with the skeleton, then heavy into the child, got %+v", got)
		}
	})
}

// A failed heavy render behaves as a failed non-deferred op: the skeleton is
// cleaned before the deferred op reports the error, a queued skeleton call is
// dropped, and nothing is sent for heavy. Afterwards an inner Door takes
// updates, an outer Door reports its outer error, and a static Door only
// stores state.
func TestDeferredHeavyRenderError(t *testing.T) {
	errHeavy := errors.New("heavy failed")
	for _, m := range deferredModes {
		for _, held := range []bool{false, true} {
			name := m.name + "/sent"
			if held {
				name = m.name + "/queued"
			}
			t.Run(name, func(t *testing.T) {
				h := newLifecycleHarness(t, 8)
				g := newHeavyGate(t)
				chs := make(chan (<-chan error), 8)
				ctxs := make(chan context.Context, 8)
				d := &Door{}
				d.Inner(context.Background(), textElem("V0"))
				h.renderPage(mountDoor(d))
				heavy := m.wrap(func(cur gox.Cursor) error {
					if err := g.elem("HEAVY")(cur); err != nil {
						return err
					}
					return errHeavy
				})
				if held {
					h.client.hold()
				}
				g.shut()
				ch1 := d.Inner(context.Background(), func(cur gox.Cursor) error {
					ctx := cur.Context()
					ch := m.op(ctx, heavy)
					OnClean(ctx, func() { h.events <- fmt.Sprintf("clean-skeleton reported=%d", len(ch)) })
					ctxs <- ctx
					chs <- ch
					return cur.Text("SKELETON")
				})
				ch2 := recvOp(t, chs)
				skeletonCtx := <-ctxs
				var want []actionWant
				if held {
					expectScheduled(t, "loader", ch1)
					h.waitQueued(1)
				} else {
					expectOp(t, "loader", ch1, nil, nil)
					want = append(want, actionWant{"door_update", "SKELETON"})
				}
				g.waitEntered(t)
				g.open()
				waitEventBeforeReport(t, h, "clean-skeleton reported=0", ch2)
				expectOp(t, "deferred", ch2, errHeavy)
				if skeletonCtx.Err() == nil {
					t.Fatal("expected the skeleton context canceled")
				}
				if held {
					h.client.release()
					expectOp(t, "loader", ch1, context.Canceled)
				}
				h.waitIdle()
				expectActions(t, h.client.actions(), want...)
				after := d.Inner(context.Background(), textElem("AFTER"))
				switch m.name {
				case "inner":
					expectOp(t, "after", after, nil, nil)
					want = append(want, actionWant{"door_update", "AFTER"})
				case "outer":
					expectOpError(t, "after", after, errHeavy, "outer error: ")
				case "static":
					expectOp(t, "after", after)
				}
				h.waitIdle()
				expectActions(t, h.client.actions(), want...)
				h.expectNoEvent(50 * time.Millisecond)
			})
		}
	}
}

// DeferredInner over an Outer whose render fails reports the outer error,
// whether it waited for that render or came after it.
func TestDeferredInnerOverFailedOuter(t *testing.T) {
	errOuter := errors.New("outer failed")
	for _, pending := range []bool{true, false} {
		name := "after"
		if pending {
			name = "pending"
		}
		t.Run(name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			h.renderPage(mountDoor(d))
			if pending {
				g.shut()
			}
			och := d.Outer(context.Background(), tagElem("section", "box", func(cur gox.Cursor) error {
				if err := g.elem("OUTER")(cur); err != nil {
					return err
				}
				return errOuter
			}))
			var dch <-chan error
			if pending {
				g.waitEntered(t)
				dch = d.DeferredInner(context.Background(), textElem("HEAVY"))
				g.open()
			}
			expectOp(t, "outer", och, errOuter)
			if !pending {
				dch = d.DeferredInner(context.Background(), textElem("HEAVY"))
			}
			expectOpError(t, "deferred", dch, errOuter, "outer error: ")
			h.waitIdle()
			expectActions(t, h.client.actions())
		})
	}
}

// A loader that defers and then fails its placement: the deferred op reports
// the placement error and heavy is never rendered; the loader content is
// cleaned once, with the placement. A later Inner reports the placement error
// too, or only stores state after DeferredStatic. DeferredStatic on a Door
// whose placement failed reports the same.
func TestDeferredAfterPlacementError(t *testing.T) {
	errPlace := errors.New("placement failed")
	for _, m := range deferredModes {
		t.Run("loader "+m.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			chs := make(chan (<-chan error), 8)
			var renders atomic.Int32
			p := &Door{}
			d := &Door{}
			d.Inner(context.Background(), func(cur gox.Cursor) error {
				OnClean(cur.Context(), func() { h.events <- "clean-loader" })
				chs <- m.op(cur.Context(), m.wrap(func(cur gox.Cursor) error {
					renders.Add(1)
					return cur.Text("HEAVY")
				}))
				return errPlace
			})
			p.Inner(context.Background(), textElem("P0"))
			h.renderPage(mountDoor(p))
			expectOp(t, "parent", p.Inner(context.Background(), mountDoor(d)), nil, nil)
			expectOpError(t, "deferred", recvOp(t, chs), errPlace, "placement error: ")
			h.waitEvent("clean-loader")
			after := d.Inner(context.Background(), textElem("AFTER"))
			if m.name == "static" {
				expectOp(t, "after", after)
			} else {
				expectOpError(t, "after", after, errPlace, "placement error: ")
			}
			h.waitIdle()
			expectActions(t, h.client.actions(), actionWant{"door_update", ""})
			if got := h.client.actions(); got[0].body != "" || renders.Load() != 0 {
				t.Fatalf("expected nothing placed and heavy never rendered, got %+v", got)
			}
			h.expectNoEvent(50 * time.Millisecond)
		})
	}
	t.Run("door static", func(t *testing.T) {
		h := newLifecycleHarness(t, 8)
		p := &Door{}
		d := &Door{}
		d.Inner(context.Background(), func(cur gox.Cursor) error {
			OnClean(cur.Context(), func() { h.events <- "clean-failed" })
			return errPlace
		})
		p.Inner(context.Background(), textElem("P0"))
		h.renderPage(mountDoor(p))
		expectOp(t, "parent", p.Inner(context.Background(), mountDoor(d)), nil, nil)
		h.waitEvent("clean-failed")
		expectOpError(t, "deferred static", d.DeferredStatic(context.Background(), textElem("STATIC")), errPlace, "placement error: ")
		h.waitIdle()
		expectActions(t, h.client.actions(), actionWant{"door_update", ""})
		h.expectNoEvent(50 * time.Millisecond)
	})
}

// A deferred op started from a batch does not hold it: OnSettle fires while
// the deferred content is still rendering.
func TestDeferredNotAwaitedBySettle(t *testing.T) {
	ops := map[string]func(d *Door, ctx context.Context, content any) <-chan error{
		"inner":  (*Door).DeferredInner,
		"outer":  (*Door).DeferredOuter,
		"static": (*Door).DeferredStatic,
	}
	for _, m := range deferredModes {
		t.Run(m.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			g := newHeavyGate(t)
			d := &Door{}
			d.Inner(context.Background(), textElem("V0"))
			ctx := h.renderPage(mountDoor(d))
			g.shut()
			var ch <-chan error
			settled := make(chan struct{})
			OnSettle(ctx, func(context.Context) { close(settled) }, func(ctx context.Context) {
				ch = ops[m.name](d, ctx, m.wrap(g.elem("HEAVY")))
			})
			g.waitEntered(t)
			select {
			case <-settled:
			case <-time.After(5 * time.Second):
				t.Fatal("the batch waited for the deferred content")
			}
			g.open()
			expectOp(t, "deferred", ch, nil, nil)
			h.waitIdle()
			expectActions(t, h.client.actions(), m.final)
			if n := h.logs.count("attempted to schedule on completed frame"); n != 0 {
				t.Fatalf("logged %d completed-frame warnings", n)
			}
		})
	}
}
