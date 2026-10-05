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
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/gox"
)

func foreignCanceledError() error {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return fmt.Errorf("db query: %w", ctx.Err())
}

func failingElem(err error) gox.Elem {
	return func(gox.Cursor) error {
		return err
	}
}

// A Door that fails while its parent renders it is left out of the parent's
// output, its error is logged once, and its operations report a placement
// error. An error wrapping a canceled context from outside Doors is a real
// error and is logged too.
func TestDoorPlacementErrorLogged(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		inUpdate bool
	}{
		{"page", errors.New("render failed"), false},
		{"page-foreign-canceled", foreignCanceledError(), false},
		{"update", errors.New("render failed"), true},
		{"update-foreign-canceled", foreignCanceledError(), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			child := &Door{}
			child.Inner(context.Background(), failingElem(c.err))
			host := &Door{}
			withChild := tagElem("div", "child-host", mountDoor(child))
			if c.inUpdate {
				host.Inner(context.Background(), textElem("h0"))
			} else {
				host.Inner(context.Background(), withChild)
			}
			pageCtx, html, err := h.renderPageHTML(mountDoor(host))
			if err != nil {
				t.Fatal(err)
			}
			pageCtx = DetachedContext(pageCtx)
			if c.inUpdate {
				if vals := drainOp(t, host.Inner(pageCtx, withChild)); len(vals) != 2 || !onlyNils(vals) {
					t.Fatalf("expected the parent update to succeed, got %v", vals)
				}
				html = h.client.actions()[0].body
			}
			if !strings.Contains(html, `id="child-host"></div>`) {
				t.Fatalf("expected the failed Door to be left out, got %q", html)
			}
			if n := h.logs.count("door rendering error"); n != 1 {
				t.Fatalf("expected one logged render error, got %d: %q", n, h.logs.all())
			}
			vals := drainOp(t, child.Inner(pageCtx, textElem("x")))
			if len(vals) != 1 || vals[0] == nil || !strings.Contains(vals[0].Error(), "placement error") {
				t.Fatalf("expected a placement error, got %v", vals)
			}
		})
	}
}

func badDataElem(cur gox.Cursor) error {
	if err := cur.Init("div"); err != nil {
		return err
	}
	if err := cur.Modify(AData{Name: "bad", Value: func() {}}); err != nil {
		return err
	}
	if err := cur.Submit(); err != nil {
		return err
	}
	return cur.Close()
}

func badScriptDataElem(cur gox.Cursor) error {
	if err := cur.Init("script"); err != nil {
		return err
	}
	if err := cur.Set("data:bad", make(chan int)); err != nil {
		return err
	}
	if err := cur.Submit(); err != nil {
		return err
	}
	if err := cur.Raw("void 0"); err != nil {
		return err
	}
	return cur.Close()
}

// Data that cannot be encoded fails the render that sets it. A page answers
// 500 with one logged render error.
func TestDataEncodeErrorFailsPage(t *testing.T) {
	for name, elem := range map[string]gox.Elem{"adata": badDataElem, "script-data": badScriptDataElem} {
		t.Run(name, func(t *testing.T) {
			logs := &lifecycleLog{}
			app := NewApp(func(context.Context, Request) gox.Elem {
				return elem
			}, WithLogger(slog.New(logs)))
			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusInternalServerError {
				t.Fatalf("expected status 500, got %d: %q", rec.Code, rec.Body.String())
			}
			if n := logs.count("door rendering error"); n != 1 {
				t.Fatalf("expected one logged render error, got %d: %q", n, logs.all())
			}
		})
	}
}

// A Door update with data that cannot be encoded reports the error on its
// channel, logs it once and sends nothing.
func TestDataEncodeErrorFailsDoorUpdate(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	d.Inner(context.Background(), textElem("v0"))
	pageCtx := DetachedContext(h.renderPage(mountDoor(d)))
	vals := drainOp(t, d.Inner(pageCtx, badDataElem))
	if len(vals) != 1 || vals[0] == nil {
		t.Fatalf("expected the render error, got %v", vals)
	}
	if n := h.waitLogged("door rendering error", 1); n != 1 {
		t.Fatalf("expected one logged render error, got %d: %q", n, h.logs.all())
	}
	h.waitIdle()
	if got := h.client.actions(); len(got) != 0 {
		t.Fatalf("expected no call for the failed update, got %+v", got)
	}
}

func gatedElem(err error, rendering chan<- context.Context, gate <-chan struct{}) gox.Elem {
	return func(cur gox.Cursor) error {
		rendering <- cur.Context()
		<-gate
		if err == nil {
			return cur.Context().Err()
		}
		return err
	}
}

// A render that fails with a real error after a newer operation superseded it
// is logged once, and its channel reports context.Canceled. A render that only
// ends because it was superseded is not logged.
func TestDoorSupersededRenderErrorLogged(t *testing.T) {
	next := map[string]func(d *Door, ctx context.Context) <-chan error{
		"inner":   func(d *Door, ctx context.Context) <-chan error { return d.Inner(ctx, textElem("v2")) },
		"outer":   func(d *Door, ctx context.Context) <-chan error { return d.Outer(ctx, textElem("v2")) },
		"unmount": func(d *Door, ctx context.Context) <-chan error { return d.Unmount(ctx) },
	}
	for name, op := range next {
		for _, err := range []error{errors.New("render failed"), nil} {
			label := name + "/canceled"
			if err != nil {
				label = name + "/real-error"
			}
			t.Run(label, func(t *testing.T) {
				h := newLifecycleHarness(t, 8)
				d := &Door{}
				d.Inner(context.Background(), textElem("v0"))
				pageCtx := DetachedContext(h.renderPage(mountDoor(d)))
				rendering := make(chan context.Context, 1)
				gate := make(chan struct{})
				first := d.Inner(pageCtx, gatedElem(err, rendering, gate))
				renderCtx := <-rendering
				second := op(d, pageCtx)
				<-renderCtx.Done()
				close(gate)
				if vals := drainOp(t, first); len(vals) != 1 || !errors.Is(vals[0], context.Canceled) {
					t.Fatalf("expected the superseded operation to report context.Canceled, got %v", vals)
				}
				if vals := drainOp(t, second); len(vals) != 2 || !onlyNils(vals) {
					t.Fatalf("expected the newer operation to succeed, got %v", vals)
				}
				want := 0
				if err != nil {
					want = 1
				}
				if n := h.logs.count("door rendering error"); n != want {
					t.Fatalf("expected %d logged render errors, got %d: %q", want, n, h.logs.all())
				}
			})
		}
	}
}

// The same for a Door whose placement render fails after its parent's render
// was superseded.
func TestDoorSupersededPlacementErrorLogged(t *testing.T) {
	for _, err := range []error{errors.New("render failed"), nil} {
		label := "canceled"
		if err != nil {
			label = "real-error"
		}
		t.Run(label, func(t *testing.T) {
			h := newLifecycleHarness(t, 8)
			host := &Door{}
			host.Inner(context.Background(), textElem("h0"))
			pageCtx := DetachedContext(h.renderPage(mountDoor(host)))
			rendering := make(chan context.Context, 1)
			gate := make(chan struct{})
			child := &Door{}
			child.Inner(context.Background(), gatedElem(err, rendering, gate))
			first := host.Inner(pageCtx, mountDoor(child))
			renderCtx := <-rendering
			second := host.Inner(pageCtx, textElem("h2"))
			<-renderCtx.Done()
			close(gate)
			if vals := drainOp(t, first); len(vals) != 1 || !errors.Is(vals[0], context.Canceled) {
				t.Fatalf("expected the superseded parent update to report context.Canceled, got %v", vals)
			}
			if vals := drainOp(t, second); len(vals) != 2 || !onlyNils(vals) {
				t.Fatalf("expected the newer parent update to succeed, got %v", vals)
			}
			want := 0
			if err != nil {
				want = 1
			}
			if n := h.logs.count("door rendering error"); n != want {
				t.Fatalf("expected %d logged render errors, got %d: %q", want, n, h.logs.all())
			}
		})
	}
}

type panicPage string

func (p panicPage) Main() gox.Elem {
	return func(cur gox.Cursor) error {
		switch p {
		case "page":
			panic("page render panic")
		case "door":
			d := &Door{}
			d.Inner(cur.Context(), gox.Elem(func(cur gox.Cursor) error { panic("door render panic") }))
			return cur.Any(d)
		}
		return cur.Text("ok")
	}
}

// A panic during the initial page render answers 500, whether it happens in
// the page itself or in a Door's content.
func TestInitialRenderPanicAnswers500(t *testing.T) {
	for _, kind := range []string{"page", "door", "control"} {
		t.Run(kind, func(t *testing.T) {
			app := NewApp(func(context.Context, Request) gox.Comp {
				return panicPage(kind)
			}, WithConf(Conf{ServerSessionCookieNoSecure: true, InstanceConnectTimeout: time.Minute}), WithLogger(slog.New(&lifecycleLog{})))
			want := http.StatusInternalServerError
			if kind == "control" {
				want = http.StatusOK
			}
			for range 10 {
				rec := httptest.NewRecorder()
				app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
				if rec.Code != want {
					t.Fatalf("expected status %d, got %d: %q", want, rec.Code, rec.Body.String())
				}
			}
		})
	}
}
