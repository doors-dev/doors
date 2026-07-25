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
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type lifecycleKiller struct {
	killed chan struct{}
	once   sync.Once
}

func (k *lifecycleKiller) Kill() {
	k.once.Do(func() {
		close(k.killed)
	})
}

func (k *lifecycleKiller) Logger() *slog.Logger { return slog.Default() }

type lifecycleInstance struct {
	*helperInstance
	ids atomic.Uint64
}

func (l *lifecycleInstance) NewID() uint64 {
	return l.ids.Add(1)
}

func (l *lifecycleInstance) Call(c actions.Call) {
	c.Result(nil, nil)
}

var _ door.Instance = &lifecycleInstance{}

type lifecycleHarness struct {
	t      *testing.T
	inst   *lifecycleInstance
	root   door.Root
	killer *lifecycleKiller
	events chan string
}

func newLifecycleHarness(t *testing.T, workers int) *lifecycleHarness {
	t.Helper()
	conf := common.Conf{}
	common.InitDefaults(&conf)
	inst := &lifecycleInstance{helperInstance: &helperInstance{
		conf:     conf,
		location: beam.NewSource(path.Location{}, path.EqualLocation, false),
	}}
	sessCtx, cancel := context.WithCancel(context.Background())
	inst.session = &helperSession{
		inst:   inst.helperInstance,
		app:    &helperApp{conf: &inst.helperInstance.conf},
		ctx:    sessCtx,
		cancel: cancel,
	}
	killer := &lifecycleKiller{killed: make(chan struct{})}
	rt := shredder.NewRuntime(context.Background(), workers, killer)
	inst.runtime = rt
	root := door.NewRoot(inst)
	t.Cleanup(func() {
		root.Kill()
		rt.Cancel()
		cancel()
	})
	return &lifecycleHarness{
		t:      t,
		inst:   inst,
		root:   root,
		killer: killer,
		events: make(chan string, 64),
	}
}

// renderPageErr renders a full page containing content and returns the
// page-level render context (root tracker content ctx). It is safe to call
// from a non-test goroutine.
func (h *lifecycleHarness) renderPageErr(content gox.Elem) (context.Context, error) {
	ctxCh := make(chan context.Context, 1)
	page := gox.Elem(func(cur gox.Cursor) error {
		ctxCh <- cur.Context()
		if content == nil {
			return nil
		}
		return content(cur)
	})
	stack, err := h.root.Render(context.Background(), page)
	if err != nil {
		return nil, err
	}
	if err := stack.Print(gox.NewPrinter(io.Discard)); err != nil {
		return nil, err
	}
	return <-ctxCh, nil
}

// renderPage renders a full page containing content and returns the page-level
// render context (root tracker content ctx).
func (h *lifecycleHarness) renderPage(content gox.Elem) context.Context {
	h.t.Helper()
	ctx, err := h.renderPageErr(content)
	if err != nil {
		h.t.Fatal(err)
	}
	return ctx
}

func (h *lifecycleHarness) waitEvent(want string) {
	h.t.Helper()
	select {
	case got := <-h.events:
		if got != want {
			h.t.Fatalf("expected event %q, got %q", want, got)
		}
	case <-time.After(5 * time.Second):
		h.t.Fatalf("timed out waiting for event %q", want)
	}
}

func (h *lifecycleHarness) waitEvents(want ...string) {
	h.t.Helper()
	pending := map[string]int{}
	for _, w := range want {
		pending[w]++
	}
	for n := len(want); n > 0; n-- {
		select {
		case got := <-h.events:
			if pending[got] == 0 {
				h.t.Fatalf("unexpected event %q while waiting for %v", got, want)
			}
			pending[got]--
		case <-time.After(5 * time.Second):
			h.t.Fatalf("timed out waiting for events %v", want)
		}
	}
}

func (h *lifecycleHarness) expectNoEvent(d time.Duration) {
	h.t.Helper()
	select {
	case got := <-h.events:
		h.t.Fatalf("unexpected event %q", got)
	case <-time.After(d):
	}
}

func mountDoor(d *Door) gox.Elem {
	return func(cur gox.Cursor) error {
		return d.Edit(cur)
	}
}

func textElem(text string) gox.Elem {
	return func(cur gox.Cursor) error {
		return cur.Text(text)
	}
}

// OnReady registered during a render fires only after the cycle's scheduled
// signal (the first nil on the X channel), never while the producing render is
// still in flight, with a detached, still-live context. Both phases hold the
// render open to deterministically rule out any earlier latch (for example the
// operation guard, which activates at apply time before the render).
func TestOnReadyFiresAfterRenderCycleScheduled(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}

	// Initial page render, held open at the door content.
	initialEntered := make(chan struct{})
	initialRelease := make(chan struct{})
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnReady(cur.Context(), func(context.Context) { h.events <- "ready-initial" })
		close(initialEntered)
		<-initialRelease
		return cur.Text("v0")
	}))
	type pageResult struct {
		ctx context.Context
		err error
	}
	pageCh := make(chan pageResult, 1)
	go func() {
		ctx, err := h.renderPageErr(mountDoor(d))
		pageCh <- pageResult{ctx: ctx, err: err}
	}()
	<-initialEntered
	// The render is held open, so the page cycle cannot be complete yet and
	// the registration must not have fired.
	h.expectNoEvent(50 * time.Millisecond)
	close(initialRelease)
	page := <-pageCh
	if page.err != nil {
		t.Fatal(page.err)
	}
	pageCtx := page.ctx
	h.waitEvent("ready-initial")

	// Update cycle, held open at the replacing content.
	updateEntered := make(chan struct{})
	updateRelease := make(chan struct{})
	xch := make(chan (<-chan error), 1)
	content := gox.Elem(func(cur gox.Cursor) error {
		OnReady(cur.Context(), func(ctx context.Context) {
			ch := <-xch
			select {
			case err := <-ch:
				if err != nil {
					h.events <- "scheduled-error"
					return
				}
			default:
				h.events <- "ready-before-scheduled"
				return
			}
			if !ctex.IsFreeCtx(ctx) {
				h.events <- "ctx-not-detached"
				return
			}
			if ctx.Err() != nil {
				h.events <- "ctx-canceled"
				return
			}
			h.events <- "ready-after-scheduled"
		})
		close(updateEntered)
		<-updateRelease
		return cur.Text("v1")
	})
	ch := d.Inner(DetachedContext(pageCtx), content)
	xch <- ch
	<-updateEntered
	// While the render is held open, neither the ready callback nor the
	// scheduled signal may appear.
	h.expectNoEvent(50 * time.Millisecond)
	select {
	case err := <-ch:
		t.Fatalf("unexpected scheduled signal while the render was held open: %v", err)
	default:
	}
	close(updateRelease)
	h.waitEvent("ready-after-scheduled")
}

// OnReady from a real handler context (valve already open) dispatches to the
// instance pool promptly and never runs synchronously on the calling
// goroutine.
func TestOnReadyFromHandlerContextFiresPromptly(t *testing.T) {
	h := newLifecycleHarness(t, 1)
	pageCtx := h.renderPage(nil)

	// Occupy the single pool worker so an async dispatch cannot start yet.
	release := make(chan struct{})
	started := make(chan struct{})
	h.inst.Runtime().Submit(context.Background(), func(b bool) {
		if !b {
			return
		}
		close(started)
		<-release
	}, nil)
	<-started

	c := pageCtx.Value(common.KeyCore).(core.Core)
	hook, ok := c.Door().RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		OnReady(ctx, func(context.Context) { h.events <- "ready-handler" })
		return true
	}, nil)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if !h.root.TriggerHook(hook.HookID, rec, req, 0) {
		t.Fatal("expected hook trigger to succeed")
	}
	// The worker is still blocked: if OnReady had run inline on the handler
	// goroutine, the event would already be here.
	h.expectNoEvent(50 * time.Millisecond)
	close(release)
	h.waitEvent("ready-handler")
}

// OnFlush registered during a render with a trivial batch must not fire while
// the producing cycle is in flight; it fires only after the cycle is enqueued,
// with a detached live context.
func TestOnFlushRenderCtxWaitsForCycle(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	entered := make(chan struct{})
	release := make(chan struct{})
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnFlush(cur.Context(), func(ctx context.Context) {
			if !ctex.IsFreeCtx(ctx) {
				h.events <- "flush-ctx-not-detached"
				return
			}
			if ctx.Err() != nil {
				h.events <- "flush-ctx-canceled"
				return
			}
			h.events <- "flush-render"
		})
		close(entered)
		<-release
		return cur.Text("v0")
	}))
	rendered := make(chan struct{})
	go func() {
		h.renderPageErr(mountDoor(d))
		close(rendered)
	}()
	<-entered
	h.expectNoEvent(50 * time.Millisecond)
	close(release)
	h.waitEvent("flush-render")
	<-rendered
}

// OnFlush in a handler joins the handler's whole batch: it must not fire while
// the handler is still open and fires once the batch is dispatched.
func TestOnFlushHandlerCtxWaitsForBatch(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	c := pageCtx.Value(common.KeyCore).(core.Core)
	entered := make(chan struct{})
	proceed := make(chan struct{})
	hook, ok := c.Door().RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		OnFlush(ctx, func(context.Context) { h.events <- "flush-handler" })
		close(entered)
		<-proceed
		return true
	}, nil)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	triggered := make(chan struct{})
	go func() {
		h.root.TriggerHook(hook.HookID, rec, req, 0)
		close(triggered)
	}()
	<-entered
	h.expectNoEvent(50 * time.Millisecond)
	close(proceed)
	h.waitEvent("flush-handler")
	<-triggered
}

// OnFlush on a detached context opens a batch spanning the ops calls: ops run
// synchronously, a nested OnFlush inside an op reuses the same batch context,
// and on fires after the batch completes.
func TestOnFlushDetachedCtxOpsAndNestedReuse(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	var outer, inner context.Context
	var synced atomic.Bool
	OnFlush(DetachedContext(pageCtx), func(context.Context) {
		if inner == nil || inner != outer {
			h.events <- "flush-nested-ctx-mismatch"
			return
		}
		if !synced.Load() {
			h.events <- "flush-ops-not-sync"
			return
		}
		h.events <- "flush-detached"
	}, func(ctx context.Context) {
		outer = ctx
		OnFlush(ctx, func(context.Context) {}, func(ctx context.Context) {
			inner = ctx
		})
	}, func(ctx context.Context) {
		synced.Store(true)
	})
	h.waitEvent("flush-detached")
}

// A reload whose node CAS fails (door updated in the same batch) must not
// strand the batch counter: on still fires.
func TestOnFlushReloadCasFailureClosesBatch(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	ctxCh := make(chan context.Context, 1)
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		ctxCh <- cur.Context()
		return cur.Text("v0")
	}))
	h.renderPage(mountDoor(d))
	contentCtx := <-ctxCh

	OnFlush(DetachedContext(contentCtx), func(context.Context) {
		h.events <- "flush-after-stale-reload"
	}, func(ctx context.Context) {
		d.Inner(ctx, textElem("v1"))
		Reload(ctx)
	})
	h.waitEvent("flush-after-stale-reload")
}

// An unmount inside the batch holds it open: on must not fire before the
// unmount's client call is scheduled.
func TestOnFlushWaitsForUnmountDispatch(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	pageCtx := h.renderPage(mountDoor(d))

	xch := make(chan (<-chan error), 1)
	OnFlush(DetachedContext(pageCtx), func(ctx context.Context) {
		ch := <-xch
		select {
		case err := <-ch:
			if err != nil {
				h.events <- "unmount-error"
				return
			}
			h.events <- "flush-after-unmount-scheduled"
		default:
			h.events <- "flush-before-unmount-scheduled"
		}
	}, func(ctx context.Context) {
		xch <- d.Unmount(ctx)
	})
	h.waitEvent("flush-after-unmount-scheduled")
}

// A canceled owner context does not drop OnFlush: on still runs once the
// batch point is reached; only render failure or instance shutdown drop it.
func TestOnFlushRunsOnCanceledContext(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	cctx, cancel := context.WithCancel(DetachedContext(pageCtx))
	cancel()
	OnFlush(cctx, func(ctx context.Context) {
		if ctx.Err() == nil {
			h.events <- "flush-ctx-not-canceled"
			return
		}
		h.events <- "flush-canceled"
	})
	h.waitEvent("flush-canceled")
}

// OnFlush registered after runtime shutdown still fires, with a canceled
// context.
func TestOnFlushRunsOnRuntimeShutdown(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	h.inst.runtime.Cancel()
	OnFlush(DetachedContext(pageCtx), func(ctx context.Context) {
		if ctx.Err() == nil {
			h.events <- "flush-shutdown-ctx-live"
			return
		}
		h.events <- "flush-shutdown"
	})
	h.waitEvent("flush-shutdown")
}

// A failing render still fires OnFlush, with a canceled context reporting the
// drop of the batch; OnClean fires as well.
func TestOnFlushRunsOnRenderError(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	pageCtx := h.renderPage(mountDoor(d))

	renderErr := errors.New("render failed")
	failing := gox.Elem(func(cur gox.Cursor) error {
		OnFlush(cur.Context(), func(ctx context.Context) {
			if ctx.Err() == nil {
				h.events <- "flush-error-ctx-live"
				return
			}
			h.events <- "flush-error-canceled"
		})
		OnClean(cur.Context(), func() { h.events <- "clean-error" })
		return renderErr
	})
	ch := d.Inner(DetachedContext(pageCtx), failing)
	if err := <-ch; !errors.Is(err, renderErr) {
		t.Fatalf("expected render error, got %v", err)
	}
	h.waitEvents("clean-error", "flush-error-canceled")
	h.expectNoEvent(100 * time.Millisecond)
}

// A superseded render cycle drops its OnReady while its OnClean still fires,
// deferred until the replacing cycle is enqueued.
func TestOnReadySupersededDroppedCleanStillFires(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	pageCtx := h.renderPage(mountDoor(d))

	blocked := make(chan struct{})
	first := gox.Elem(func(cur gox.Cursor) error {
		ctx := cur.Context()
		OnReady(ctx, func(context.Context) { h.events <- "ready-first" })
		OnClean(ctx, func() { h.events <- "clean-first" })
		close(blocked)
		// Hold the first cycle open until the second operation supersedes it.
		<-ctx.Done()
		return cur.Text("v1")
	})
	ch1 := d.Inner(DetachedContext(pageCtx), first)
	<-blocked

	second := gox.Elem(func(cur gox.Cursor) error {
		OnReady(cur.Context(), func(context.Context) { h.events <- "ready-second" })
		return cur.Text("v2")
	})
	d.Inner(DetachedContext(pageCtx), second)

	h.waitEvents("clean-first", "ready-second")
	if err := <-ch1; !errors.Is(err, context.Canceled) {
		t.Fatalf("expected superseded operation to report context.Canceled, got %v", err)
	}
	// ready-first must never fire.
	h.expectNoEvent(100 * time.Millisecond)
}

// A failing render drops its OnReady while OnClean fires immediately; the X
// channel reports the render error.
func TestOnReadyDroppedOnRenderError(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	pageCtx := h.renderPage(mountDoor(d))

	renderErr := errors.New("render failed")
	failing := gox.Elem(func(cur gox.Cursor) error {
		OnReady(cur.Context(), func(context.Context) { h.events <- "ready-error" })
		OnClean(cur.Context(), func() { h.events <- "clean-error" })
		return renderErr
	})
	ch := d.Inner(DetachedContext(pageCtx), failing)
	if err := <-ch; !errors.Is(err, renderErr) {
		t.Fatalf("expected render error, got %v", err)
	}
	h.waitEvent("clean-error")
	// ready-error must never fire.
	h.expectNoEvent(100 * time.Millisecond)
}

// On replacement, the replaced content's OnClean runs exactly once, deferred
// until the replacing cycle is enqueued (not at apply time): it must still be
// pending while the replacing content renders.
func TestOnCleanReplaceDeferred(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	var cleanDispatched atomic.Bool
	var cleanCount atomic.Int32
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnReady(cur.Context(), func(context.Context) { h.events <- "ready-old" })
		OnClean(cur.Context(), func() {
			cleanDispatched.Store(true)
			cleanCount.Add(1)
			h.events <- "clean-old"
		})
		return cur.Text("v0")
	}))
	pageCtx := h.renderPage(mountDoor(d))
	h.waitEvent("ready-old")

	next := gox.Elem(func(cur gox.Cursor) error {
		if cleanDispatched.Load() {
			// The replaced OnClean must still be pending while the replacing
			// content renders: it is deferred until this cycle is enqueued.
			h.events <- "clean-before-replacing-render"
			return nil
		}
		OnReady(cur.Context(), func(context.Context) { h.events <- "ready-new" })
		return cur.Text("v1")
	})
	ch := d.Inner(DetachedContext(pageCtx), next)
	if err := <-ch; err != nil {
		t.Fatalf("expected replacing update to schedule, got %v", err)
	}
	h.waitEvents("clean-old", "ready-new")
	if got := cleanCount.Load(); got != 1 {
		t.Fatalf("expected the replaced OnClean to run exactly once, ran %d times", got)
	}
	h.expectNoEvent(100 * time.Millisecond)
}

// OnClean fires on unmount; registration on an already-cleaned owner still
// fires, and OnReady on a cleaned owner is dropped.
func TestOnCleanUnmountAndCleanedOwner(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	ctxCh := make(chan context.Context, 1)
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		ctxCh <- cur.Context()
		OnClean(cur.Context(), func() { h.events <- "clean-unmount" })
		return cur.Text("v0")
	}))
	pageCtx := h.renderPage(mountDoor(d))
	contentCtx := <-ctxCh

	d.Unmount(pageCtx)
	h.waitEvent("clean-unmount")

	late := make(chan struct{})
	OnClean(contentCtx, func() { close(late) })
	select {
	case <-late:
	case <-time.After(5 * time.Second):
		t.Fatal("expected OnClean on a cleaned owner to still fire")
	}

	OnReady(contentCtx, func(context.Context) { h.events <- "ready-late" })
	h.expectNoEvent(100 * time.Millisecond)
}

// In a nested-door cascade, both children's and parents' OnClean callbacks
// fire.
func TestOnCleanCascadeFires(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	parent := &Door{}
	child := &Door{}
	child.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnClean(cur.Context(), func() { h.events <- "clean-child" })
		return cur.Text("child")
	}))
	parent.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnClean(cur.Context(), func() { h.events <- "clean-parent" })
		return child.Edit(cur)
	}))
	pageCtx := h.renderPage(mountDoor(parent))

	parent.Unmount(pageCtx)
	h.waitEvents("clean-child", "clean-parent")
}

// OnClean fires on instance end for both nested doors and root registrations.
func TestOnCleanOnInstanceEnd(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnClean(cur.Context(), func() { h.events <- "clean-door" })
		return cur.Text("v0")
	}))
	pageCtx := h.renderPage(mountDoor(d))
	OnClean(pageCtx, func() { h.events <- "clean-root" })

	h.root.Kill()
	h.waitEvents("clean-door", "clean-root")
}

// A panic inside OnClean is recovered: the remaining callbacks in the batch
// still run, the instance is killed, and the test process survives.
func TestOnCleanPanicRecovered(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnClean(cur.Context(), func() { panic("boom-clean") })
		OnClean(cur.Context(), func() { h.events <- "clean-after-panic" })
		return cur.Text("v0")
	}))
	pageCtx := h.renderPage(mountDoor(d))

	d.Unmount(pageCtx)
	h.waitEvent("clean-after-panic")
	select {
	case <-h.killer.killed:
	case <-time.After(5 * time.Second):
		t.Fatal("expected a panic in OnClean to kill the instance")
	}
}

// A panic inside OnReady is recovered on the pool: the instance is killed and
// the test process survives.
func TestOnReadyPanicRecovered(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnReady(cur.Context(), func(context.Context) { panic("boom-ready") })
		return cur.Text("v0")
	}))
	h.renderPage(mountDoor(d))
	select {
	case <-h.killer.killed:
	case <-time.After(5 * time.Second):
		t.Fatal("expected a panic in OnReady to kill the instance")
	}
}

// Both functions panic on a context that does not belong to a Doors render or
// handler.
func TestLifecycleForeignContextPanics(t *testing.T) {
	recovered := func(f func()) (r any) {
		defer func() { r = recover() }()
		f()
		return nil
	}
	if recovered(func() { OnReady(context.Background(), func(context.Context) {}) }) == nil {
		t.Fatal("expected OnReady to panic on a non-doors context")
	}
	if recovered(func() { OnClean(context.Background(), func() {}) }) == nil {
		t.Fatal("expected OnClean to panic on a non-doors context")
	}
}

// OnReady inside Door.Static content binds to the static render cycle: it must
// not fire while the static render is held open and fires once the cycle is
// enqueued.
func TestOnReadyStaticFiresAfterCycle(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	pageCtx := h.renderPage(mountDoor(d))

	entered := make(chan struct{})
	release := make(chan struct{})
	d.Static(DetachedContext(pageCtx), gox.Elem(func(cur gox.Cursor) error {
		OnReady(cur.Context(), func(context.Context) { h.events <- "ready-static" })
		close(entered)
		<-release
		return cur.Text("static")
	}))
	<-entered
	h.expectNoEvent(50 * time.Millisecond)
	close(release)
	h.waitEvent("ready-static")
}

// OnReady inside Door.Static content is dropped when the static render errors.
func TestOnReadyStaticDroppedOnError(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	pageCtx := h.renderPage(mountDoor(d))

	ch := d.Static(DetachedContext(pageCtx), gox.Elem(func(cur gox.Cursor) error {
		OnReady(cur.Context(), func(context.Context) { h.events <- "ready-static-error" })
		return errors.New("static boom")
	}))
	if err := <-ch; err == nil {
		t.Fatal("expected static render error")
	}
	h.expectNoEvent(100 * time.Millisecond)
}
