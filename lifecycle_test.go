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
	queue   []linksQueued
	links   []linksCall
	held    bool
	busy    bool
	ended   bool
	seq     uint64
	applied []lifecycleAction
}

func newLifecycleClient() *lifecycleClient {
	c := &lifecycleClient{}
	c.cond.L = &c.mu
	go c.run()
	return c
}

func (c *lifecycleClient) call(call actions.Call, link linksCall) uint64 {
	c.mu.Lock()
	if c.ended {
		c.mu.Unlock()
		call.Cancel()
		return 0
	}
	c.seq++
	seq := c.seq
	link.seq = seq
	c.links = append(c.links, link)
	c.queue = append(c.queue, linksQueued{call, seq})
	c.mu.Unlock()
	c.cond.Broadcast()
	return seq
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
		queued := c.queue[0]
		c.queue[0] = linksQueued{}
		c.queue = c.queue[1:]
		c.busy = true
		c.mu.Unlock()
		c.apply(queued.call, queued.seq)
		c.mu.Lock()
		c.busy = false
	}
}

func (c *lifecycleClient) apply(call actions.Call, seq uint64) {
	action, free, ok := call.Action()
	if !ok {
		call.Cancel()
		return
	}
	applied := lifecycleAction{name: action.Log()}
	c.mu.Lock()
	c.links[seq-1].name = applied.name
	c.mu.Unlock()
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
	if _, ok := action.(actions.Time); ok {
		call.Result(json.RawMessage(`"2026-01-02T03:04:05Z"`), nil)
		return
	}
	call.Result(nil, nil)
}

func (c *lifecycleClient) end() {
	c.mu.Lock()
	c.ended = true
	queue := c.queue
	c.queue = nil
	c.mu.Unlock()
	c.cond.Broadcast()
	for _, queued := range queue {
		queued.call.Cancel()
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

type linksQueued struct {
	call actions.Call
	seq  uint64
}

type linksCall struct {
	seq     uint64
	name    string
	action  actions.Action
	group   uint64
	batch   uint64
	door    uint64
	depSeq  uint64
	depDoor uint64
}

func (c *lifecycleClient) issued() []linksCall {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.links)
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
	linksID string
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

func (l *lifecycleInstance) ID() string {
	if l.linksID != "" {
		return l.linksID
	}
	return l.helperInstance.ID()
}

func (l *lifecycleInstance) Call(c actions.Call, group uint64, batch uint64, door uint64, depSeq uint64, depDoor uint64) uint64 {
	return l.client.call(c, linksCall{group: group, batch: batch, door: door, depSeq: depSeq, depDoor: depDoor})
}

func (l *lifecycleInstance) UserCall(ctx context.Context, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams, door uint64, depSeq uint64) {
	frames := ctex.GetFrames(ctx)
	l.client.call(
		&lifecycleUserCall{ctx: ctx, action: action, onResult: onResult, onCancel: onCancel, params: params},
		linksCall{name: action.Log(), action: action, group: frames.Group(), batch: frames.Batch(), door: door, depSeq: depSeq},
	)
}

func (l *lifecycleInstance) UserCallCheck(group uint64, batch uint64, target uint64, check func() bool, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	l.client.call(
		&lifecycleUserCall{ctx: context.Background(), check: check, action: action, onResult: onResult, onCancel: onCancel, params: params},
		linksCall{name: action.Log(), action: action, group: group, batch: batch, door: target},
	)
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

// waitApplied waits until the client has applied at least n calls.
func (h *lifecycleHarness) waitApplied(n int) {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for len(h.client.actions()) < n {
		if time.Now().After(deadline) {
			h.t.Fatalf("timed out waiting for %d applied calls, have %d", n, len(h.client.actions()))
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

func freeCtx() context.Context {
	return DetachedContext(context.Background())
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
	d.Inner(freeCtx(), textElem("v0"))
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
	d.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
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
			if !ctex.GetFrames(ctx).Detached() {
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
	}, RaceSerial, false)
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
// with a live context.
func TestOnSettleRenderCtxWaitsForCycle(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	entered := make(chan struct{})
	release := make(chan struct{})
	d.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
		OnSettle(cur.Context(), func(ctx context.Context) {
			if ctex.GetFrames(ctx).Detached() {
				h.events <- "settle-ctx-detached"
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
	}, RaceSerial, false)
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
	d.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
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
	d.Inner(freeCtx(), textElem("v0"))
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
	d.Inner(freeCtx(), textElem("v0"))
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
	d.Inner(freeCtx(), textElem("v0"))
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
	d.Inner(freeCtx(), textElem("v0"))
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
	d.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
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
	child.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
		OnClean(cur.Context(), func() { h.events <- "clean-child" })
		return cur.Text("child")
	}))
	parent.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
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
	d.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
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
	d.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
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
	d.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
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
		}, race, false)
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

func appliedNames(h *lifecycleHarness, from int) []string {
	applied := h.client.actions()[from:]
	names := make([]string, len(applied))
	for i, a := range applied {
		names[i] = a.name
	}
	return names
}

// OnBatch without a batch opens one around ops and fires only when the client has applied its closer.
func TestOnBatchOpensBatch(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	pageCtx := h.renderPage(mountDoor(d))
	h.waitIdle()
	from := len(h.client.actions())
	h.client.hold()
	OnBatch(DetachedContext(pageCtx), func(ctx context.Context) {
		if !ctex.GetFrames(ctx).Detached() {
			h.events <- "batch-ctx-not-detached"
			return
		}
		h.events <- "batch"
	}, func(ctx context.Context) {
		d.Inner(ctx, gox.Elem(func(cur gox.Cursor) error { return cur.Text("x") }))
	})
	h.waitQueued(3)
	h.expectNoEvent(50 * time.Millisecond)
	h.client.release()
	h.waitEvent("batch")
	if got := appliedNames(h, from); !slices.Equal(got, []string{"batch_init", "door_update", "batch_flush"}) {
		t.Fatalf("unexpected call order %v", got)
	}
}

// OnBatch in a batch hook joins the hook batch: no second batch, and on fires when the hook report is applied.
func TestOnBatchJoinsHookBatch(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	pageCtx := h.renderPage(mountDoor(d))
	h.waitIdle()
	from := len(h.client.actions())
	c := pageCtx.Value(common.KeyCore).(core.Core)
	hook, ok := c.Door().RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		OnBatch(ctx, func(context.Context) { h.events <- "batch" })
		d.Inner(ctx, gox.Elem(func(cur gox.Cursor) error { return cur.Text("x") }))
		return true
	}, RaceSerial, true)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	h.client.hold()
	h.root.TriggerHook(hook.HookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), 7)
	h.waitQueued(3)
	h.expectNoEvent(50 * time.Millisecond)
	h.client.release()
	h.waitEvent("batch")
	if got := appliedNames(h, from); !slices.Equal(got, []string{"batch_init", "door_update", "report hook"}) {
		t.Fatalf("unexpected call order %v", got)
	}
}

// OnBatch in a plain hook opens a batch on the hook's settle: the flush goes out before the hook report.
func TestOnBatchInPlainHook(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	pageCtx := h.renderPage(mountDoor(d))
	h.waitIdle()
	from := len(h.client.actions())
	c := pageCtx.Value(common.KeyCore).(core.Core)
	hook, ok := c.Door().RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		OnBatch(ctx, func(context.Context) { h.events <- "batch" }, func(ctx context.Context) {
			d.Inner(ctx, gox.Elem(func(cur gox.Cursor) error { return cur.Text("x") }))
		})
		return true
	}, RaceSerial, false)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	h.root.TriggerHook(hook.HookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), 7)
	h.waitEvent("batch")
	h.waitIdle()
	if got := appliedNames(h, from); !slices.Equal(got, []string{"batch_init", "door_update", "batch_flush", "report hook"}) {
		t.Fatalf("unexpected call order %v", got)
	}
}

// OnBatch in an OnSettle callback of a tracked hook, plain or Batch, opens its own batch in the hook's group: its ops carry the new batch and on fires only when its flush is applied.
func TestOnBatchInSettleOfHook(t *testing.T) {
	for _, batched := range []bool{false, true} {
		t.Run(fmt.Sprint("batch=", batched), func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			a, b := &Door{}, &Door{}
			_, page, from := linksPage(h, a, b)
			applied := make(chan []string, 1)
			h.client.hold()
			linksHook(h, page, batched, 7, func(ctx context.Context) {
				OnSettle(ctx, func(ctx context.Context) {
					OnBatch(ctx, func(context.Context) {
						applied <- appliedNames(h, 0)
						h.events <- "batch"
					}, func(ctx context.Context) {
						a.Inner(ctx, textElem("a"))
						b.Inner(ctx, textElem("b"))
					})
				})
			})
			want := []string{"batch_init", "report hook", "door_update", "door_update", "batch_flush"}
			if batched {
				want = append(want, "batch_init")
			}
			h.waitQueued(len(want))
			h.expectNoEvent(50 * time.Millisecond)
			h.client.release()
			h.waitEvent("batch")
			if got := <-applied; !slices.Contains(got, "batch_flush") {
				t.Fatalf("expected on after the flush is applied, got %v", got)
			}
			links := linksExpect(h, from, want...)
			report := linksNamed(t, links, "report hook", 1)[0]
			flush := linksNamed(t, links, "batch_flush", 1)[0]
			inits := linksNamed(t, links, "batch_init", len(want)-4)
			own := inits[len(inits)-1]
			if report.group == 0 || own.batch == 0 || own.batch == report.batch || flush.batch != own.batch {
				t.Fatalf("expected the report to close the hook and the flush a batch of its own, got %+v, %+v, %+v", report, own, flush)
			}
			if batched && inits[0].batch != report.batch {
				t.Fatalf("expected the report to close the hook's batch %d, got %+v", inits[0].batch, report)
			}
			for _, l := range linksNamed(t, links, "door_update", 2) {
				if l.batch != own.batch || l.seq < own.seq || l.seq > flush.seq {
					t.Fatalf("expected the op between batch_init and batch_flush of batch %d, got %+v", own.batch, l)
				}
			}
			for _, l := range links {
				if l.group != report.group {
					t.Fatalf("expected %s in group %d, got %d", l.name, report.group, l.group)
				}
			}
		})
	}
}

// OnBatch under OnSettle with ops inside a settle callback of a Batch hook opens a new batch instead of joining the closed one.
func TestOnBatchInNestedSettleOfBatchHook(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	a, b := &Door{}, &Door{}
	_, page, from := linksPage(h, a, b)
	applied := make(chan []string, 1)
	h.client.hold()
	linksHook(h, page, true, 7, func(ctx context.Context) {
		OnSettle(ctx, func(ctx context.Context) {
			OnSettle(ctx, func(context.Context) {}, func(ctx context.Context) {
				OnBatch(ctx, func(context.Context) {
					applied <- appliedNames(h, 0)
					h.events <- "batch"
				}, func(ctx context.Context) {
					a.Inner(ctx, textElem("a"))
					b.Inner(ctx, textElem("b"))
				})
			})
		})
	})
	h.waitQueued(6)
	h.expectNoEvent(50 * time.Millisecond)
	h.client.release()
	h.waitEvent("batch")
	if got := <-applied; !slices.Contains(got, "batch_flush") {
		t.Fatalf("expected on after the flush is applied, got %v", got)
	}
	links := linksExpect(h, from, "batch_init", "batch_init", "report hook", "door_update", "door_update", "batch_flush")
	inits := linksNamed(t, links, "batch_init", 2)
	report := linksNamed(t, links, "report hook", 1)[0]
	flush := linksNamed(t, links, "batch_flush", 1)[0]
	hook, own := inits[0], inits[1]
	if report.batch != hook.batch || own.batch == 0 || own.batch == hook.batch || flush.batch != own.batch {
		t.Fatalf("expected a second batch closed by its own flush, got %+v, %+v, %+v", hook, own, flush)
	}
	for _, l := range linksNamed(t, links, "door_update", 2) {
		if l.batch != own.batch || l.seq < own.seq || l.seq > flush.seq {
			t.Fatalf("expected the op between batch_init and batch_flush of batch %d, got %+v", own.batch, l)
		}
	}
}

// OnSettle with ops inside a settle callback of a tracked Batch hook puts neither the hook's group nor its batch on the ops' calls or on calls from its own callback.
func TestOnSettleOpsInSettleCallbackUntagged(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	a := &Door{}
	_, page, from := linksPage(h, a)
	linksHook(h, page, true, 7, func(ctx context.Context) {
		OnSettle(ctx, func(ctx context.Context) {
			OnSettle(ctx, func(ctx context.Context) {
				Call(ctx, ActionScroll{Selector: "#b"})
				h.events <- "settled"
			}, func(ctx context.Context) {
				a.Inner(ctx, textElem("a"))
				Call(ctx, ActionScroll{Selector: "#a"})
			})
		})
	})
	h.waitEvent("settled")
	links := linksExpect(h, from, "batch_init", "report hook", "door_update", "scroll", "scroll")
	if report := linksNamed(t, links, "report hook", 1)[0]; report.group == 0 || report.batch == 0 {
		t.Fatalf("expected the hook in a group and a batch, got %+v", report)
	}
	untagged := []linksCall{linksNamed(t, links, "door_update", 1)[0], linksScroll(t, links, "#a"), linksScroll(t, links, "#b")}
	for _, l := range untagged {
		if l.group != 0 || l.batch != 0 {
			t.Fatalf("expected %s without group and batch, got %+v", l.name, l)
		}
	}
}

// Detached contexts, the callbacks run on them, and wrappers of them are detached; handler, render, subscription, OnSettle callback and ops contexts are not.
func TestDetachedPerContextKind(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	marks := make(chan string, 32)
	record := func(name string, ctx context.Context) {
		marks <- fmt.Sprintf("%s=%v", name, ctex.GetFrames(ctx).Detached())
	}
	src := NewSource(0)
	pageCtx := h.renderPage(func(cur gox.Cursor) error {
		record("render", cur.Context())
		OnReady(cur.Context(), func(ctx context.Context) { record("ready", ctx) })
		src.Sub(cur.Context(), func(ctx context.Context, v int) bool {
			switch v {
			case 1:
				record("sub", ctx)
			case 2:
				record("sub-detached", ctx)
			}
			return false
		})
		return Go(func(ctx context.Context) { record("go", ctx) })(cur)
	})
	h.waitIdle()
	page := pageCtx.Value(common.KeyCore).(core.Core).Door()
	detached := DetachedContext(pageCtx)
	record("detached", detached)
	record("instance", InstanceContext(pageCtx))
	timed, cancel := context.WithTimeout(detached, time.Minute)
	defer cancel()
	record("timeout", timed)
	linksHook(h, page, false, 8, func(ctx context.Context) {
		record("handler", ctx)
		OnSettle(ctx, func(ctx context.Context) { record("settle", ctx) })
		src.Update(ctx, 1)
	})
	OnSettle(detached, func(context.Context) {}, func(ctx context.Context) { record("settle-ops", ctx) })
	OnBatch(detached, func(ctx context.Context) { record("batch", ctx) }, func(ctx context.Context) { record("batch-ops", ctx) })
	src.Update(detached, 2)
	want := []string{
		"batch-ops=false", "batch=true", "detached=true", "go=true", "handler=false", "instance=true",
		"ready=true", "render=false", "settle-ops=false", "settle=false", "sub-detached=true", "sub=false", "timeout=true",
	}
	var got []string
	for range want {
		select {
		case g := <-marks:
			got = append(got, g)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out, got %v", got)
		}
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

// OnBatch still fires when the instance ends before the client reports the batch.
func TestOnBatchFiresOnInstanceEnd(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(nil)
	h.waitIdle()
	h.client.hold()
	OnBatch(DetachedContext(pageCtx), func(context.Context) { h.events <- "batch" })
	h.expectNoEvent(50 * time.Millisecond)
	h.inst.end()
	h.waitEvent("batch")
}

// Call sends only the scheduled value and leaves dst alone in a batch or plain handler; on a detached ctx it also sends the result after writing dst.
func TestCallChannelPerContext(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx, page, _ := linksPage(h)
	var batchAt, plainAt, detachedAt time.Time
	var batchChs, plainChs []<-chan error
	linksHook(h, page, true, 7, func(ctx context.Context) {
		batchChs = append(batchChs, Call(ctx, ActionScroll{Selector: "#x"}), Call(ctx, ActionTime{}.Into(&batchAt)))
	})
	linksHook(h, page, false, 8, func(ctx context.Context) {
		plainChs = append(plainChs, Call(ctx, ActionScroll{Selector: "#x"}), Call(ctx, ActionTime{}.Into(&plainAt)))
	})
	for _, ch := range append(batchChs, plainChs...) {
		if got := drainOp(t, ch); len(got) != 1 || got[0] != nil {
			t.Fatalf("expected only the scheduled nil, got %v", got)
		}
	}
	detachedCh := Call(DetachedContext(pageCtx), ActionTime{}.Into(&detachedAt))
	timed, cancel := context.WithTimeout(DetachedContext(pageCtx), time.Minute)
	defer cancel()
	timedCh := Call(timed, ActionScroll{Selector: "#x"})
	for _, ch := range []<-chan error{detachedCh, timedCh} {
		if got := drainOp(t, ch); len(got) != 2 || got[0] != nil || got[1] != nil {
			t.Fatalf("expected the scheduled and the result nil, got %v", got)
		}
	}
	h.waitIdle()
	if !batchAt.IsZero() || !plainAt.IsZero() {
		t.Fatalf("expected dst untouched outside a detached ctx, got %v and %v", batchAt, plainAt)
	}
	if want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC); !detachedAt.Equal(want) {
		t.Fatalf("expected dst written on a detached ctx, got %v", detachedAt)
	}
}

// A detached Call canceled after it was scheduled reports context.Canceled as its second value.
func TestCallCanceledAfterScheduling(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(textElem("page"))
	h.waitIdle()
	h.client.hold()
	ctx, cancel := context.WithCancel(DetachedContext(pageCtx))
	ch := Call(ctx, ActionScroll{Selector: "#x"})
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("expected the scheduled nil, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("call was never scheduled")
	}
	cancel()
	h.client.release()
	if got := drainOp(t, ch); len(got) != 1 || !errors.Is(got[0], context.Canceled) {
		t.Fatalf("expected context.Canceled after the scheduled nil, got %v", got)
	}
}

// Before scheduling Call has one shape on every context: a build error then close, or close without a value when canceled.
func TestCallChannelBeforeScheduling(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	pageCtx := h.renderPage(textElem("page"))
	canceledDetached, cancel := context.WithCancel(DetachedContext(pageCtx))
	cancel()
	canceledPlain, cancel := context.WithCancel(pageCtx)
	cancel()
	for name, ctx := range map[string]context.Context{"detached": DetachedContext(pageCtx), "plain": pageCtx} {
		if got := drainOp(t, Call(ctx, ActionEmit[any]{Name: "bad", Arg: make(chan int)})); len(got) != 1 || got[0] == nil {
			t.Fatalf("%s: expected the build error only, got %v", name, got)
		}
	}
	for name, ctx := range map[string]context.Context{"detached": canceledDetached, "plain": canceledPlain} {
		if got := drainOp(t, Call(ctx, ActionScroll{Selector: "#x"})); len(got) != 0 {
			t.Fatalf("%s: expected a close without a value, got %v", name, got)
		}
	}
}

// OnSettle with ops in a render opens work that settles only after the render is ready, also for an OnSettle joined in its ops.
func TestOnSettleRenderOpsWaitForRender(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	var done atomic.Bool
	release := make(chan struct{})
	entered := make(chan struct{})
	d.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
		OnSettle(cur.Context(), func(context.Context) {
			if !done.Load() {
				h.events <- "outer-before-render"
				return
			}
			h.events <- "outer"
		}, func(ctx context.Context) {
			OnSettle(ctx, func(context.Context) {
				if !done.Load() {
					h.events <- "inner-before-render"
					return
				}
				h.events <- "inner"
			})
		})
		close(entered)
		<-release
		done.Store(true)
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
	h.waitEvents("outer", "inner")
	<-rendered
}

// OnSettle in a handler does not wait for foreign work holding the door.
func TestOnSettleHandlerSkipsForeignWork(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	a := &Door{}
	pageCtx, page, _ := linksPage(h, a)
	gate := make(chan struct{})
	open := sync.OnceFunc(func() { close(gate) })
	defer open()
	entered := make(chan struct{})
	static := a.Static(DetachedContext(pageCtx), gox.Elem(func(cur gox.Cursor) error {
		close(entered)
		<-gate
		return cur.Text("static")
	}))
	<-entered
	linksHook(h, page, false, 7, func(ctx context.Context) {
		OnSettle(ctx, func(context.Context) { h.events <- "settle" })
	})
	h.waitEvent("settle")
	open()
	<-static
}

// OnBatch in a render opens a batch that closes only after the render is ready.
func TestOnBatchRenderWaitsForRender(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	var done atomic.Bool
	release := make(chan struct{})
	entered := make(chan struct{})
	d.Inner(freeCtx(), gox.Elem(func(cur gox.Cursor) error {
		OnBatch(cur.Context(), func(context.Context) {
			if !done.Load() {
				h.events <- "batch-before-render"
				return
			}
			h.events <- "batch"
		})
		close(entered)
		<-release
		done.Store(true)
		return cur.Text("v0")
	}))
	rendered := make(chan struct{})
	go func() {
		h.renderPageErr(mountDoor(d))
		close(rendered)
	}()
	<-entered
	h.expectNoEvent(50 * time.Millisecond)
	if slices.Contains(linksNames(h.client.issued()), "batch_flush") {
		t.Fatalf("expected no batch_flush before the render is ready, got %v", linksNames(h.client.issued()))
	}
	close(release)
	h.waitEvent("batch")
	<-rendered
}
