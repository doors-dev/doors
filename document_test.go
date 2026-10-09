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
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/doors-dev/gox"
)

type documentProbes struct {
	mu   sync.Mutex
	seen map[string][]bool
}

func (p *documentProbes) record(name string, ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.seen == nil {
		p.seen = map[string][]bool{}
	}
	p.seen[name] = append(p.seen[name], IsDocument(ctx))
}

func (p *documentProbes) elem(name string) gox.Elem {
	return func(cur gox.Cursor) error {
		p.record(name, cur.Context())
		return cur.Text(name)
	}
}

func (p *documentProbes) wait(t *testing.T, name string, want ...bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		p.mu.Lock()
		got := slices.Clone(p.seen[name])
		p.mu.Unlock()
		if len(got) >= len(want) {
			if !slices.Equal(got, want) {
				t.Errorf("%s: IsDocument = %v, want %v", name, got, want)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Errorf("%s: timed out with %v, want %v", name, got, want)
			return
		}
		time.Sleep(time.Millisecond)
	}
}

type documentKey struct{}

func TestIsDocument(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	p := &documentProbes{}
	bg := context.Background()

	placed := &Door{}
	nested := &Door{}
	nested.Inner(bg, p.elem("nested"))
	placed.Inner(bg, func(cur gox.Cursor) error {
		if err := p.elem("placed")(cur); err != nil {
			return err
		}
		return cur.Comp(nested)
	})
	proxy := &Door{}
	staticPlaced := &Door{}
	staticPlaced.Static(bg, p.elem("static placed"))
	staticLater := &Door{}
	staticLater.Inner(bg, textElem("before static"))
	updated := &Door{}
	updated.Inner(bg, textElem("before update"))
	reloaded := &Door{}
	reloaded.Inner(bg, p.elem("reload"))
	deferred := &Door{}
	deferred.Inner(bg, func(cur gox.Cursor) error {
		DeferredInner(cur.Context(), p.elem("deferred"))
		return cur.Text("skeleton")
	})
	duringRender := &Door{}
	duringRender.Inner(bg, textElem("before"))
	parent := &Door{}
	child := &Door{}
	child.Inner(bg, p.elem("child"))
	parentContent := gox.Elem(func(cur gox.Cursor) error {
		return cur.Comp(child)
	})
	parent.Inner(bg, parentContent)
	src := NewSource(0)
	renderSrc := NewSource(0)

	h.renderPage(func(cur gox.Cursor) error {
		ctx := cur.Context()
		p.record("root", ctx)
		p.record("derived", context.WithValue(ctx, documentKey{}, true))
		p.record("detached", DetachedContext(ctx))
		p.record("instance", InstanceContext(ctx))
		OnReady(ctx, func(ctx context.Context) {
			p.record("on ready", ctx)
		})
		OnSettle(ctx, func(ctx context.Context) {
			p.record("on settle", ctx)
		}, func(ctx context.Context) {
			p.record("on settle op", ctx)
		})
		src.Sub(ctx, func(ctx context.Context, _ int) bool {
			p.record("sub", ctx)
			return false
		})
		for _, el := range []gox.Elem{
			mountDoor(placed),
			func(cur gox.Cursor) error {
				return proxy.Proxy(cur, tagElem("div", "proxy", p.elem("proxy")))
			},
			mountDoor(staticPlaced),
			mountDoor(staticLater),
			mountDoor(updated),
			mountDoor(reloaded),
			mountDoor(deferred),
			mountDoor(duringRender),
			mountDoor(parent),
			func(cur gox.Cursor) error {
				return Parallel().Proxy(cur, p.elem("parallel"))
			},
			src.Bind(func(v int) gox.Elem {
				return p.elem("bind")
			}),
			Go(func(ctx context.Context) {
				p.record("go", ctx)
			}),
		} {
			if err := el(cur); err != nil {
				return err
			}
		}
		renderSrc.Sub(ctx, func(ctx context.Context, v int) bool {
			if v != 0 {
				p.record("sub update from render", ctx)
			}
			return false
		})
		renderSrc.Update(ctx, 1)
		duringRender.Inner(ctx, p.elem("update during render"))
		return nil
	})

	staticLater.Static(bg, p.elem("static later"))
	updated.Inner(bg, p.elem("update"))
	reloaded.Reload(bg)
	parent.Inner(bg, parentContent)
	src.Update(bg, 1)

	p.wait(t, "root", true)
	p.wait(t, "derived", true)
	p.wait(t, "detached", false)
	p.wait(t, "instance", false)
	p.wait(t, "on settle op", true)
	p.wait(t, "placed", true)
	p.wait(t, "nested", true)
	p.wait(t, "proxy", true)
	p.wait(t, "static placed", true)
	p.wait(t, "parallel", true)
	p.wait(t, "on ready", false)
	p.wait(t, "on settle", false)
	p.wait(t, "go", false)
	p.wait(t, "deferred", false)
	p.wait(t, "update during render", false)
	p.wait(t, "static later", false)
	p.wait(t, "update", false)
	p.wait(t, "reload", true, false)
	p.wait(t, "child", true, false)
	p.wait(t, "bind", true, false)
	p.wait(t, "sub", false, false)
	p.wait(t, "sub update from render", false)
}
