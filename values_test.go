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
	"net/http"
	"sync"
	"testing"
	"time"

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
			hook, ok := NewHook(ctx, ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
				seen.set("hook", r.Context())
				w.Write([]byte("hook"))
			}))
			if !ok {
				return nil
			}
			plain := Ctx(context.WithValue(context.Background(), valueOther{}, "other"))
			if err := plain.Proxy(cur, gox.Elem(func(cur gox.Cursor) error {
				seen.set("ctx", cur.Context())
				return nil
			})); err != nil {
				return err
			}
			own := Ctx(context.WithValue(context.Background(), valueKey{}, "own"))
			if err := own.Proxy(cur, gox.Elem(func(cur gox.Cursor) error {
				seen.set("own", cur.Context())
				return nil
			})); err != nil {
				return err
			}
			return headTag(cur, "html", func() error {
				return headTag(cur, "body", func() error {
					return cur.Text("hook:" + hook)
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
	for _, name := range []string{"page", "request", "render", "session", "instance", "detached", "hook", "ctx"} {
		if got := seen.get(name); got != "app" {
			t.Errorf("%s: expected the app value, got %v", name, got)
		}
	}
	if got := seen.get("own"); got != "own" {
		t.Errorf("expected the Ctx value to win, got %v", got)
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
