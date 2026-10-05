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
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
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
	"github.com/doors-dev/doors/internal/instance/utils"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/printer"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

// lifecycleLog captures the instance log.
type lifecycleLog struct {
	mu   sync.Mutex
	msgs []string
}

func (l *lifecycleLog) Enabled(context.Context, slog.Level) bool { return true }

func (l *lifecycleLog) Handle(_ context.Context, r slog.Record) error {
	msg := r.Message
	r.Attrs(func(a slog.Attr) bool {
		msg += " " + a.Key + "=" + a.Value.String()
		return true
	})
	l.mu.Lock()
	l.msgs = append(l.msgs, msg)
	l.mu.Unlock()
	return nil
}

func (l *lifecycleLog) WithAttrs([]slog.Attr) slog.Handler { return l }

func (l *lifecycleLog) WithGroup(string) slog.Handler { return l }

func (l *lifecycleLog) all() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.msgs)
}

func (l *lifecycleLog) count(prefix string) int {
	n := 0
	for _, msg := range l.all() {
		if strings.HasPrefix(msg, prefix) {
			n++
		}
	}
	return n
}

// lifecycleAction is an action the client applied.
type lifecycleAction struct {
	name string
	id   uint64
	body string
}

// lifecycleClient plays the browser side of the solitaire: calls queue in
// the order the instance issues them and are applied asynchronously, one at a
// time. Holding the client keeps calls queued, like a client that has not
// answered yet. On instance end, queued and later calls are canceled.
type lifecycleClient struct {
	mu      sync.Mutex
	cond    sync.Cond
	queue   []actions.Call
	held    bool
	busy    bool
	ended   bool
	applied []lifecycleAction
}

func newLifecycleClient() *lifecycleClient {
	c := &lifecycleClient{}
	c.cond.L = &c.mu
	go c.run()
	return c
}

func (c *lifecycleClient) call(call actions.Call) {
	c.mu.Lock()
	if c.ended {
		c.mu.Unlock()
		call.Cancel()
		return
	}
	c.queue = append(c.queue, call)
	c.mu.Unlock()
	c.cond.Broadcast()
}

func (c *lifecycleClient) run() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for {
		for !c.ended && (c.held || len(c.queue) == 0) {
			c.cond.Wait()
		}
		if c.ended {
			return
		}
		call := c.queue[0]
		c.queue[0] = nil
		c.queue = c.queue[1:]
		c.busy = true
		c.mu.Unlock()
		c.apply(call)
		c.mu.Lock()
		c.busy = false
	}
}

func (c *lifecycleClient) apply(call actions.Call) {
	action, free, ok := call.Action()
	if !ok {
		call.Cancel()
		return
	}
	applied := lifecycleAction{name: action.Log()}
	switch action := action.(type) {
	case actions.DoorReplace:
		applied.id, applied.body = action.ID, payloadText(action.Payload)
	case actions.DoorUpdate:
		applied.id, applied.body = action.ID, payloadText(action.Payload)
	default:
		applied.body = fmt.Sprint(action)
	}
	free()
	c.mu.Lock()
	c.applied = append(c.applied, applied)
	c.mu.Unlock()
	call.Result(nil, nil)
}

func (c *lifecycleClient) end() {
	c.mu.Lock()
	c.ended = true
	queue := c.queue
	c.queue = nil
	c.mu.Unlock()
	c.cond.Broadcast()
	for _, call := range queue {
		call.Cancel()
	}
}

func (c *lifecycleClient) hold() {
	c.mu.Lock()
	c.held = true
	c.mu.Unlock()
}

func (c *lifecycleClient) release() {
	c.mu.Lock()
	c.held = false
	c.mu.Unlock()
	c.cond.Broadcast()
}

func (c *lifecycleClient) queued() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.queue)
}

func (c *lifecycleClient) idle() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.queue) == 0 && !c.busy
}

func (c *lifecycleClient) actions() []lifecycleAction {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.applied)
}

func payloadText(p actions.Payload) string {
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		return "payload error: " + err.Error()
	}
	if p.Type() != actions.PayloadTextGZ {
		return buf.String()
	}
	r, err := gzip.NewReader(&buf)
	if err != nil {
		return "payload error: " + err.Error()
	}
	text, err := io.ReadAll(r)
	if err != nil {
		return "payload error: " + err.Error()
	}
	return string(text)
}

// lifecycleUserCall mirrors the instance's user call: the action is dropped
// once its context is canceled or its check fails.
type lifecycleUserCall struct {
	ctx      context.Context
	check    func() bool
	action   actions.Action
	onResult func(json.RawMessage, error)
	onCancel func()
	params   actions.CallParams
}

func (c *lifecycleUserCall) Params() actions.CallParams { return c.params }

func (c *lifecycleUserCall) Action() (actions.Action, func(), bool) {
	if c.check != nil && !c.check() {
		return nil, nil, false
	}
	if c.check == nil && c.ctx.Err() != nil {
		return nil, nil, false
	}
	return c.action, func() {}, true
}

func (c *lifecycleUserCall) Cancel() {
	if c.onCancel != nil {
		c.onCancel()
	}
}

func (c *lifecycleUserCall) Result(r json.RawMessage, err error) {
	if c.onResult != nil {
		c.onResult(r, err)
	}
}

type lifecycleApp struct {
	*helperApp
	logger *slog.Logger
}

func (a *lifecycleApp) Logger() *slog.Logger { return a.logger }

type lifecycleSession struct {
	*helperSession
	app *lifecycleApp
}

func (s *lifecycleSession) App() core.App { return s.app }

func (s *lifecycleSession) Logger() *slog.Logger { return s.app.logger }

// lifecycleInstance mirrors the production instance around a real Root: it
// owns the runtime, and ending it cancels the runtime, cancels pending calls
// and kills the root, in production order.
type lifecycleInstance struct {
	*helperInstance
	ids     atomic.Uint64
	sess    *lifecycleSession
	client  *lifecycleClient
	meta    core.TitleMeta
	root    door.Root
	endOnce sync.Once
	ended   chan struct{}
}

func (l *lifecycleInstance) NewID() uint64 {
	return l.ids.Add(1)
}

func (l *lifecycleInstance) Call(c actions.Call) {
	l.client.call(c)
}

func (l *lifecycleInstance) UserCall(ctx context.Context, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	l.client.call(&lifecycleUserCall{ctx: ctx, action: action, onResult: onResult, onCancel: onCancel, params: params})
}

func (l *lifecycleInstance) UserCallCheck(check func() bool, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	l.client.call(&lifecycleUserCall{ctx: context.Background(), check: check, action: action, onResult: onResult, onCancel: onCancel, params: params})
}

func (l *lifecycleInstance) Session() core.Session { return l.sess }

func (l *lifecycleInstance) Logger() *slog.Logger { return l.sess.app.logger }

func (l *lifecycleInstance) TitleMeta() core.TitleMeta { return l.meta }

func (l *lifecycleInstance) Kill() {
	l.end()
}

func (l *lifecycleInstance) end() {
	l.endOnce.Do(func() {
		l.runtime.Cancel()
		l.client.end()
		l.root.Kill()
		close(l.ended)
	})
}

var _ door.Instance = &lifecycleInstance{}

type lifecycleHarness struct {
	t      *testing.T
	inst   *lifecycleInstance
	root   door.Root
	client *lifecycleClient
	logs   *lifecycleLog
	events chan string
}

func newLifecycleHarness(t *testing.T, workers int) *lifecycleHarness {
	t.Helper()
	conf := common.Conf{}
	common.InitDefaults(&conf)
	logs := &lifecycleLog{}
	inst := &lifecycleInstance{
		helperInstance: &helperInstance{
			conf:     conf,
			location: beam.NewSource(path.Location{}, path.EqualLocation, false),
		},
		client: newLifecycleClient(),
		ended:  make(chan struct{}),
	}
	sessCtx, cancel := context.WithCancel(context.Background())
	inst.session = &helperSession{
		inst:   inst.helperInstance,
		app:    &helperApp{conf: &inst.helperInstance.conf},
		cancel: cancel,
	}
	inst.sess = &lifecycleSession{
		helperSession: inst.session,
		app:           &lifecycleApp{helperApp: inst.session.app, logger: slog.New(logs)},
	}
	inst.session.ctx = context.WithValue(sessCtx, common.KeySession, inst.sess)
	inst.meta = utils.NewTitleMeta(inst)
	inst.runtime = shredder.NewRuntime(inst.session.ctx, workers, inst)
	inst.root = door.NewRoot(inst)
	t.Cleanup(func() {
		inst.end()
		cancel()
	})
	return &lifecycleHarness{
		t:      t,
		inst:   inst,
		root:   inst.root,
		client: inst.client,
		logs:   logs,
		events: make(chan string, 64),
	}
}

func lifecycleInclude(gox.Cursor) error { return nil }

// renderPageHTML serves a page the way the instance does: the root renders
// content after the navigator subscribes, the stack is printed with the page
// printer, and a render or print error ends the instance. It returns the
// page-level render context (root tracker content ctx) and the printed page.
// It is safe to call from a non-test goroutine.
func (h *lifecycleHarness) renderPageHTML(content gox.Elem) (context.Context, string, error) {
	ctxCh := make(chan context.Context, 1)
	page := gox.Elem(func(cur gox.Cursor) error {
		ctxCh <- cur.Context()
		utils.NewNavigator(h.inst, cur.Context()).Sync()
		if content == nil {
			return nil
		}
		return content(cur)
	})
	stack, err := h.root.Render(context.Background(), page)
	if err != nil {
		h.inst.end()
		return nil, "", err
	}
	var buf bytes.Buffer
	pr := printer.NewPagePrinter(&buf, lifecycleInclude, nil, h.inst.TitleMeta())
	if err := stack.Print(h.inst.Session().App().PrinterMiddleware()(pr)); err != nil {
		h.inst.end()
		return nil, "", err
	}
	return <-ctxCh, buf.String(), nil
}

func (h *lifecycleHarness) renderPageErr(content gox.Elem) (context.Context, error) {
	ctx, _, err := h.renderPageHTML(content)
	return ctx, err
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

func (h *lifecycleHarness) waitQueued(n int) {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for h.client.queued() < n {
		if time.Now().After(deadline) {
			h.t.Fatalf("timed out waiting for %d queued calls, have %d", n, h.client.queued())
		}
		time.Sleep(time.Millisecond)
	}
}

func (h *lifecycleHarness) waitIdle() {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !h.client.idle() {
		if time.Now().After(deadline) {
			h.t.Fatalf("timed out waiting for the client to apply %d queued calls", h.client.queued())
		}
		time.Sleep(time.Millisecond)
	}
}

// waitLogged waits until at least n messages with prefix are logged and
// returns their count.
func (h *lifecycleHarness) waitLogged(prefix string, n int) int {
	deadline := time.Now().Add(5 * time.Second)
	for h.logs.count(prefix) < n && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	return h.logs.count(prefix)
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
		return cur.Comp(d)
	}
}

func textElem(text string) gox.Elem {
	return func(cur gox.Cursor) error {
		return cur.Text(text)
	}
}

func tagElem(tag string, id string, content gox.Elem) gox.Elem {
	return func(cur gox.Cursor) error {
		if err := cur.Init(tag); err != nil {
			return err
		}
		if err := cur.Set("id", id); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		if content != nil {
			if err := content(cur); err != nil {
				return err
			}
		}
		return cur.Close()
	}
}

// drainOp collects the values of an operation channel until it closes.
func drainOp(t *testing.T, ch <-chan error) []error {
	t.Helper()
	var vals []error
	timeout := time.After(5 * time.Second)
	for {
		select {
		case err, ok := <-ch:
			if !ok {
				return vals
			}
			vals = append(vals, err)
		case <-timeout:
			t.Fatalf("operation channel did not close, got %v", vals)
		}
	}
}

// A page render that updates a Door it just rendered still prints, as the
// production page printer does not check job contexts. The page shows the
// rendered content and the update reaches the client as a call.
func TestLifecyclePageUpdatedDuringRender(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	var ch <-chan error
	_, html, err := h.renderPageHTML(func(cur gox.Cursor) error {
		if err := cur.Comp(d); err != nil {
			return err
		}
		ch = d.Inner(DetachedContext(cur.Context()), textElem("v1"))
		return nil
	})
	if err != nil {
		t.Fatalf("page failed: %v", err)
	}
	if !strings.Contains(html, "v0") {
		t.Fatalf("expected the rendered content in the page, got %q", html)
	}
	if vals := drainOp(t, ch); len(vals) != 2 || vals[0] != nil || vals[1] != nil {
		t.Fatalf("expected the update to succeed, got %v", vals)
	}
	got := h.client.actions()
	if len(got) != 1 || got[0].name != "door_update" || got[0].body != "v1" {
		t.Fatalf("expected one door_update with v1, got %+v", got)
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

// OnReady from a real handler context (valve already open) runs inline on the
// calling goroutine before OnReady returns, independent of pool availability.
func TestOnReadyFromHandlerContextFiresInline(t *testing.T) {
	h := newLifecycleHarness(t, 1)
	pageCtx := h.renderPage(nil)

	// Occupy the single pool worker: inline execution must not need it.
	release := make(chan struct{})
	started := make(chan struct{})
	h.inst.Runtime().Submit(context.Background(), func(b bool) {
		if !b {
			return
		}
		close(started)
		<-release
	}, nil)
	defer close(release)
	<-started

	c := pageCtx.Value(common.KeyCore).(core.Core)
	hook, ok := c.Door().RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		ran := false
		OnReady(ctx, func(context.Context) { ran = true })
		if ran {
			h.events <- "ready-inline"
		} else {
			h.events <- "ready-async"
		}
		return true
	}, RaceSerial)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if !h.root.TriggerHook(hook.HookID, rec, req, 0) {
		t.Fatal("expected hook trigger to succeed")
	}
	h.waitEvent("ready-inline")
}

// OnSettle registered during a render with a trivial batch must not fire while
// the producing cycle is in flight; it fires only after the cycle is enqueued,
// with a detached live context.
func TestOnSettleRenderCtxWaitsForCycle(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	entered := make(chan struct{})
	release := make(chan struct{})
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnSettle(cur.Context(), func(ctx context.Context) {
			if !ctex.IsFreeCtx(ctx) {
				h.events <- "settle-ctx-not-detached"
				return
			}
			if ctx.Err() != nil {
				h.events <- "settle-ctx-canceled"
				return
			}
			h.events <- "settle-render"
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
	h.waitEvent("settle-render")
	<-rendered
}

// OnSettle in a handler joins the handler's whole batch: it must not fire while
// the handler is still open and fires once the batch is dispatched.
func TestOnSettleHandlerCtxWaitsForBatch(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	c := pageCtx.Value(common.KeyCore).(core.Core)
	entered := make(chan struct{})
	proceed := make(chan struct{})
	hook, ok := c.Door().RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		OnSettle(ctx, func(context.Context) { h.events <- "settle-handler" })
		close(entered)
		<-proceed
		return true
	}, RaceSerial)
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
	h.waitEvent("settle-handler")
	<-triggered
}

// OnSettle on a detached context opens a batch spanning the ops calls: ops run
// synchronously, a nested OnSettle inside an op reuses the same batch context,
// and on fires after the batch completes.
func TestOnSettleDetachedCtxOpsAndNestedReuse(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	var outer, inner context.Context
	var synced atomic.Bool
	OnSettle(DetachedContext(pageCtx), func(context.Context) {
		if inner == nil || inner != outer {
			h.events <- "settle-nested-ctx-mismatch"
			return
		}
		if !synced.Load() {
			h.events <- "settle-ops-not-sync"
			return
		}
		h.events <- "settle-detached"
	}, func(ctx context.Context) {
		outer = ctx
		OnSettle(ctx, func(context.Context) {}, func(ctx context.Context) {
			inner = ctx
		})
	}, func(ctx context.Context) {
		synced.Store(true)
	})
	h.waitEvent("settle-detached")
}

// A reload whose node CAS fails (door updated in the same batch) must not
// strand the batch counter: on still fires.
func TestOnSettleReloadCasFailureClosesBatch(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	ctxCh := make(chan context.Context, 1)
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		ctxCh <- cur.Context()
		return cur.Text("v0")
	}))
	h.renderPage(mountDoor(d))
	contentCtx := <-ctxCh

	OnSettle(DetachedContext(contentCtx), func(context.Context) {
		h.events <- "settle-after-stale-reload"
	}, func(ctx context.Context) {
		d.Inner(ctx, textElem("v1"))
		Reload(ctx)
	})
	h.waitEvent("settle-after-stale-reload")
}

// An unmount inside the batch holds it open: on must not fire before the
// unmount's client call is scheduled.
func TestOnSettleWaitsForUnmountDispatch(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	pageCtx := h.renderPage(mountDoor(d))

	xch := make(chan (<-chan error), 1)
	OnSettle(DetachedContext(pageCtx), func(ctx context.Context) {
		ch := <-xch
		select {
		case err := <-ch:
			if err != nil {
				h.events <- "unmount-error"
				return
			}
			h.events <- "settle-after-unmount-scheduled"
		default:
			h.events <- "settle-before-unmount-scheduled"
		}
	}, func(ctx context.Context) {
		xch <- d.Unmount(ctx)
	})
	h.waitEvent("settle-after-unmount-scheduled")
}

// A canceled owner context does not drop OnSettle: on still runs once the
// batch point is reached; only render failure or instance shutdown drop it.
func TestOnSettleRunsOnCanceledContext(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	cctx, cancel := context.WithCancel(DetachedContext(pageCtx))
	cancel()
	OnSettle(cctx, func(ctx context.Context) {
		if ctx.Err() == nil {
			h.events <- "settle-ctx-not-canceled"
			return
		}
		h.events <- "settle-canceled"
	})
	h.waitEvent("settle-canceled")
}

// OnSettle registered after runtime shutdown still fires, with a canceled
// context.
func TestOnSettleRunsOnRuntimeShutdown(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	h.inst.runtime.Cancel()
	OnSettle(DetachedContext(pageCtx), func(ctx context.Context) {
		if ctx.Err() == nil {
			h.events <- "settle-shutdown-ctx-live"
			return
		}
		h.events <- "settle-shutdown"
	})
	h.waitEvent("settle-shutdown")
}

// A failing render still fires OnSettle, with a canceled context reporting the
// drop of the batch; OnClean fires as well.
func TestOnSettleRunsOnRenderError(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	pageCtx := h.renderPage(mountDoor(d))

	renderErr := errors.New("render failed")
	failing := gox.Elem(func(cur gox.Cursor) error {
		OnSettle(cur.Context(), func(ctx context.Context) {
			if ctx.Err() == nil {
				h.events <- "settle-error-ctx-live"
				return
			}
			h.events <- "settle-error-canceled"
		})
		OnClean(cur.Context(), func() { h.events <- "clean-error" })
		return renderErr
	})
	ch := d.Inner(DetachedContext(pageCtx), failing)
	if err := <-ch; !errors.Is(err, renderErr) {
		t.Fatalf("expected render error, got %v", err)
	}
	h.waitEvents("clean-error", "settle-error-canceled")
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
		return cur.Comp(child)
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

	h.inst.Kill()
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
	case <-h.inst.ended:
	case <-time.After(5 * time.Second):
		t.Fatal("expected a panic in OnClean to kill the instance")
	}
}

// A panic inside OnReady is recovered by the runtime: the instance is killed
// and the test process survives. The inline callback panics within the
// render-completing task, so the page render itself may report an error.
func TestOnReadyPanicRecovered(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		OnReady(cur.Context(), func(context.Context) { panic("boom-ready") })
		return cur.Text("v0")
	}))
	h.renderPageErr(mountDoor(d))
	select {
	case <-h.inst.ended:
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

// Disabled: Door.Static content has no lifecycle of its own. It registers on
// the parent door, and a failed static sync sends no call and leaves the parent
// intact, so OnReady from the failed content still fires once the parent is
// ready. Accepted as the cost of dropping the per-render static tracker.
//
// OnReady inside Door.Static content is dropped when the static render errors.
// func TestOnReadyStaticDroppedOnError(t *testing.T) {
// 	h := newLifecycleHarness(t, 8)
// 	d := &Door{}
// 	pageCtx := h.renderPage(mountDoor(d))
//
// 	ch := d.Static(DetachedContext(pageCtx), gox.Elem(func(cur gox.Cursor) error {
// 		OnReady(cur.Context(), func(context.Context) { h.events <- "ready-static-error" })
// 		return errors.New("static boom")
// 	}))
// 	if err := <-ch; err == nil {
// 		t.Fatal("expected static render error")
// 	}
// 	h.expectNoEvent(100 * time.Millisecond)
// }

// RaceStrict answers 412 to a call whose track is older than the last one it
// ran and skips the handler; RaceSerial runs every call.
func TestHookRaceStrictCancelsOutOfOrder(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	c := pageCtx.Value(common.KeyCore).(core.Core)
	tracks := []uint64{5, 7, 6}
	run := func(race Race) ([]uint64, []int) {
		t.Helper()
		var track uint64
		var ran []uint64
		hook, ok := c.Door().RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			ran = append(ran, track)
			return false
		}, race)
		if !ok {
			t.Fatal("expected hook registration to succeed")
		}
		codes := make([]int, 0, len(tracks))
		for _, track = range tracks {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if !h.root.TriggerHook(hook.HookID, rec, req, track) {
				t.Fatalf("expected hook trigger with track %d to be found", track)
			}
			codes = append(codes, rec.Code)
		}
		return ran, codes
	}

	ran, codes := run(RaceStrict)
	if !slices.Equal(ran, []uint64{5, 7}) {
		t.Fatalf("expected strict hook to run tracks [5 7], ran %v", ran)
	}
	if !slices.Equal(codes, []int{http.StatusOK, http.StatusOK, http.StatusPreconditionFailed}) {
		t.Fatalf("expected strict hook statuses [200 200 412], got %v", codes)
	}

	ran, codes = run(RaceSerial)
	if !slices.Equal(ran, tracks) {
		t.Fatalf("expected serial hook to run tracks %v, ran %v", tracks, ran)
	}
	if !slices.Equal(codes, []int{http.StatusOK, http.StatusOK, http.StatusOK}) {
		t.Fatalf("expected serial hook statuses [200 200 200], got %v", codes)
	}
}
