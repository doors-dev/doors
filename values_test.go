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
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/gox"
)

type valueKey struct{}

type valueOther struct{}

type valueSeen struct {
	mu   sync.Mutex
	seen map[string]any
}

func (s *valueSeen) set(name string, ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen[name] = ctx.Value(valueKey{})
}

func (s *valueSeen) get(name string) any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.seen[name]
}

func valuePage(seen *valueSeen) func(ctx context.Context, r Request) gox.Comp {
	return func(ctx context.Context, r Request) gox.Comp {
		seen.set("page", ctx)
		seen.set("request", r.Context())
		return gox.Elem(func(cur gox.Cursor) error {
			ctx := cur.Context()
			seen.set("render", ctx)
			seen.set("session", SessionContext(ctx))
			seen.set("instance", InstanceContext(ctx))
			seen.set("detached", DetachedContext(ctx))
			hook, err := ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
				seen.set("hook", r.Context())
				w.Write([]byte("hook"))
			}).URL(ctx)
			if err != nil {
				return nil
			}
			var ctxHook, ownHook string
			plain := Ctx(context.WithValue(context.Background(), valueOther{}, "other"))
			if err := plain.Proxy(cur, gox.Elem(func(cur gox.Cursor) error {
				seen.set("ctx", cur.Context())
				var err error
				ctxHook, err = ResourceHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
					seen.set("ctx-hook", ctx)
					return false
				}).URL(cur.Context())
				return err
			})); err != nil {
				return err
			}
			own := Ctx(context.WithValue(context.Background(), valueKey{}, "own"))
			if err := own.Proxy(cur, gox.Elem(func(cur gox.Cursor) error {
				seen.set("own", cur.Context())
				var err error
				ownHook, err = ResourceHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
					seen.set("own-hook", ctx)
					return false
				}).URL(cur.Context())
				return err
			})); err != nil {
				return err
			}
			return headTag(cur, "html", func() error {
				return headTag(cur, "body", func() error {
					return cur.Text("hook:" + hook + " ctx=" + ctxHook + " own=" + ownHook)
				})
			})
		})
	}
}

func newValueApp(seen *valueSeen, with ...With) App {
	conf := Conf{ServerSessionCookieNoSecure: true, InstanceConnectTimeout: time.Minute}
	return NewApp(valuePage(seen), append([]With{WithConf(conf)}, with...)...)
}

// A value from WithValue reaches every Doors context, and a value set with Ctx under the same key wins.
func TestWithValue(t *testing.T) {
	seen := &valueSeen{seen: map[string]any{}}
	app := newValueApp(seen, WithValue(valueKey{}, "first"), WithValue(valueKey{}, "app"))
	srv := headServer(t, app, false)
	c := headClient(t, srv, true)
	_, body := headDo(t, c, http.MethodGet, srv.URL+"/", false)
	hookURL, _ := headHook(t, body)
	headDo(t, c, http.MethodGet, srv.URL+hookURL, false)
	for _, m := range valueHookRegexp.FindAllSubmatch(body, -1) {
		headDo(t, c, http.MethodGet, srv.URL+string(m[1]), false)
	}
	for _, name := range []string{"page", "request", "render", "session", "instance", "detached", "hook", "ctx", "ctx-hook"} {
		if got := seen.get(name); got != "app" {
			t.Errorf("%s: expected the app value, got %v", name, got)
		}
	}
	for _, name := range []string{"own", "own-hook"} {
		if got := seen.get(name); got != "own" {
			t.Errorf("%s: expected the Ctx value to win, got %v", name, got)
		}
	}
}

// A watcher reads the values of its Sub ctx on init and on updates, not the values of the updater, while the Doors marks of the updater still reach it.
func TestCtxWatcherValues(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	src := NewSource(0)
	h.renderPage(func(cur gox.Cursor) error {
		return Ctx(context.WithValue(context.Background(), valueKey{}, "sub")).Proxy(cur, gox.Elem(func(cur gox.Cursor) error {
			src.Sub(cur.Context(), func(ctx context.Context, v int) bool {
				replace, _ := ctx.Value(common.KeyHistoryReplace).(bool)
				h.events <- fmt.Sprintf("%d %v %v", v, ctx.Value(valueKey{}), replace)
				return false
			})
			return nil
		}))
	})
	h.waitEvent("0 sub false")
	src.Update(HistoryReplaceContext(context.WithValue(freeCtx(), valueKey{}, "updater")), 1)
	h.waitEvent("1 sub true")
}

var valueHookRegexp = regexp.MustCompile(`(?:ctx|own)=(/~/doors/h/[0-9a-zA-Z]+/\d+)`)

// A handler reads the values of the render it was registered in: Ctx above the Door, Ctx inside it, nested and replacing Ctx, and Static content of a Door placed under Ctx.
func TestCtxHandlerValues(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	var mu sync.Mutex
	hooks := map[string]uint64{}
	register := func(name string) gox.Elem {
		return func(cur gox.Cursor) error {
			c := cur.Context().Value(common.KeyCore).(core.Core)
			hook, ok := c.Door().RegisterHook(cur.Context(), func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
				h.events <- fmt.Sprintf("%s %v %v", name, ctx.Value(valueKey{}), ctx.Value(valueOther{}))
				return false
			}, RaceSerial, false, false)
			if !ok {
				return context.Canceled
			}
			mu.Lock()
			hooks[name] = hook.HookID
			mu.Unlock()
			return nil
		}
	}
	under := func(ctx context.Context, el gox.Elem) gox.Elem {
		return func(cur gox.Cursor) error {
			return Ctx(ctx).Proxy(cur, el)
		}
	}
	d, s := &Door{}, &Door{}
	content := gox.Elem(func(cur gox.Cursor) error {
		if err := register("door")(cur); err != nil {
			return err
		}
		if err := under(context.WithValue(cur.Context(), valueOther{}, "nested"), register("nested"))(cur); err != nil {
			return err
		}
		if err := under(context.WithValue(context.Background(), valueOther{}, "replaced"), register("replaced"))(cur); err != nil {
			return err
		}
		return under(context.WithValue(cur.Context(), valueKey{}, "inner"), register("inner"))(cur)
	})
	h.renderPage(func(cur gox.Cursor) error {
		if err := register("page")(cur); err != nil {
			return err
		}
		if err := under(context.WithValue(context.Background(), valueKey{}, "static"), mountDoor(s))(cur); err != nil {
			return err
		}
		d.Inner(freeCtx(), content)
		return under(context.WithValue(context.Background(), valueKey{}, "top"), mountDoor(d))(cur)
	})
	if err := <-s.Static(freeCtx(), register("static")); err != nil {
		t.Fatal(err)
	}
	h.waitIdle()
	for name, want := range map[string]string{
		"page":     "page <nil> <nil>",
		"door":     "door top <nil>",
		"nested":   "nested top nested",
		"replaced": "replaced <nil> replaced",
		"inner":    "inner inner <nil>",
		"static":   "static static <nil>",
	} {
		mu.Lock()
		id, ok := hooks[name]
		mu.Unlock()
		if !ok {
			t.Fatalf("%s: hook not registered", name)
		}
		if !h.root.TriggerHook(id, httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), 0) {
			t.Fatalf("%s: expected the hook to trigger", name)
		}
		h.waitEvent(want)
	}
}

// WithFunc builds its option once per app, so each app gets its own value.
func TestWithFunc(t *testing.T) {
	calls := 0
	with := WithFunc(func() With {
		calls++
		return WithValue(valueKey{}, calls)
	})
	first := &valueSeen{seen: map[string]any{}}
	second := &valueSeen{seen: map[string]any{}}
	for _, run := range []struct {
		seen *valueSeen
		want int
	}{{first, 1}, {second, 2}} {
		app := newValueApp(run.seen, with, WithFunc(nil), WithFunc(func() With { return nil }))
		srv := headServer(t, app, false)
		headDo(t, headClient(t, srv, true), http.MethodGet, srv.URL+"/", false)
		if got := run.seen.get("render"); got != run.want {
			t.Errorf("expected value %d, got %v", run.want, got)
		}
	}
}

// An app with a nil or uncomparable WithValue key panics.
func TestWithValueKey(t *testing.T) {
	for name, key := range map[string]any{"nil": nil, "slice": []int{1}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: expected a panic", name)
				}
			}()
			newValueApp(&valueSeen{seen: map[string]any{}}, WithValue(key, "v"))
		}()
	}
}
