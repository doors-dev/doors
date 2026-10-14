// Managed by GoX v0.3.3-0.20261011164703-508309b766d1

//line static_fragments.gox:1
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
package batch

import (
	"context"
	"sync"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type staticFragment struct {
	test.NoBeam
	rendered chan string
	step     chan struct{}
	applied  chan struct{}
	results  chan error
	x        doors.Door
	y        doors.Door
	mu       sync.Mutex
	tail     *doors.Door
}

func newStaticFragment() *staticFragment {
	return &staticFragment{
		rendered: make(chan string, 16),
		step:     make(chan struct{}),
		applied:  make(chan struct{}, 16),
		results:  make(chan error, 16),
		tail:     new(doors.Door),
	}
}

func (f *staticFragment) staticPause(ctx context.Context) {
	select {
	case <-f.step:
	case <-ctx.Done():
	}
}

func (f *staticFragment) staticWatch(ch <-chan error) {
	go func() {
		var err error
		for e := range ch {
			if e != nil {
				err = e
			}
		}
		f.results <- err
	}()
}

func (f *staticFragment) staticOpen(ctx context.Context) {
	doors.OnBatch(ctx, func(context.Context) {
		f.applied <- struct{}{}
	})
}

func (f *staticFragment) staticAppend(ctx context.Context, label string) <-chan error {
	f.mu.Lock()
	defer f.mu.Unlock()
	next := new(doors.Door)
	ch := f.tail.Static(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitContainer(); if __e != nil { return }
		{
			__e = __c.Init("p"); if __e != nil { return }
			{
//line static_fragments.gox:77
				__e = __c.Set("class", "item"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line static_fragments.gox:77
				__e = __c.Any(label); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line static_fragments.gox:78
			__e = __c.Any(next); if __e != nil { return }
//line static_fragments.gox:79
			__e = __c.Any(f.staticMark(label)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line static_fragments.gox:80
	return }))
	f.tail = next
	return ch
}

//line static_fragments.gox:85
func (f *staticFragment) staticMark(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line static_fragments.gox:87
		f.rendered <- label

	return })
//line static_fragments.gox:89
}

//line static_fragments.gox:91
func (f *staticFragment) staticText(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line static_fragments.gox:92
		__e = __c.Any(f.staticMark(label)); if __e != nil { return }
//line static_fragments.gox:93
		__e = __c.Any(label); if __e != nil { return }
	return })
//line static_fragments.gox:94
}

//line static_fragments.gox:96
func (f *staticFragment) staticNested() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("p"); if __e != nil { return }
		{
//line static_fragments.gox:97
			__e = __c.Set("id", "b-item"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("item"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line static_fragments.gox:98
		__e = (f.y).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line static_fragments.gox:98
				__e = __c.Set("id", "y"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line static_fragments.gox:98
				__e = __c.Any(f.staticText("y0")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line static_fragments.gox:99
}

//line static_fragments.gox:101
func (f *staticFragment) staticScroll(detached bool) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line static_fragments.gox:103
		action := doors.ActionScroll{Selector: "#d-target"}
	if detached {
		f.staticWatch(doors.Call(doors.DetachedContext(ctx), action))
	} else {
		doors.Call(ctx, action)
	}

		__e = __c.Init("div"); if __e != nil { return }
		{
//line static_fragments.gox:110
			__e = __c.Set("style", "height:300vh"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("p"); if __e != nil { return }
		{
//line static_fragments.gox:111
			__e = __c.Set("id", "d-target"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("target"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line static_fragments.gox:112
		__e = __c.Any(f.staticMark("d")); if __e != nil { return }
	return })
//line static_fragments.gox:113
}

//line static_fragments.gox:115
func (f *staticFragment) staticBatched(id string, on func(ctx context.Context)) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line static_fragments.gox:116
		__e = (doors.AClick{
		Batch: true,
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.staticOpen(ctx)
			on(ctx)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line static_fragments.gox:123
				__e = __c.Set("id", id); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line static_fragments.gox:123
				__e = __c.Any(id); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line static_fragments.gox:124
}

//line static_fragments.gox:126
func (f *staticFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line static_fragments.gox:127
		__e = (f.x).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line static_fragments.gox:127
				__e = __c.Set("id", "x"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("x0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("section"); if __e != nil { return }
		{
//line static_fragments.gox:128
			__e = __c.Set("id", "feed"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line static_fragments.gox:128
			__e = __c.Any(f.tail); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line static_fragments.gox:129
		__e = __c.Any(f.staticBatched("a-batched", func(ctx context.Context) {
		f.x.Inner(ctx, f.staticText("x1"))
		f.staticPause(ctx)
	})); if __e != nil { return }
//line static_fragments.gox:133
		__e = __c.Any(test.Button("a-plain", func(ctx context.Context) bool {
		f.staticWatch(f.x.Static(doors.DetachedContext(ctx), gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("p"); if __e != nil { return }
			{
//line static_fragments.gox:134
				__e = __c.Set("id", "x-static"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line static_fragments.gox:134
				__e = __c.Any(f.staticText("static")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line static_fragments.gox:134
		return })))
		return false
	})); if __e != nil { return }
//line static_fragments.gox:137
		__e = __c.Any(f.staticBatched("b-batched", func(ctx context.Context) {
		f.x.Static(ctx, f.staticNested())
		f.staticPause(ctx)
	})); if __e != nil { return }
//line static_fragments.gox:141
		__e = __c.Any(test.Button("b-plain", func(ctx context.Context) bool {
		f.staticWatch(f.y.Inner(doors.DetachedContext(ctx), f.staticText("y1")))
		return false
	})); if __e != nil { return }
//line static_fragments.gox:145
		__e = __c.Any(f.staticBatched("c-batched", func(ctx context.Context) {
		for _, label := range []string{"i1", "i2", "i3"} {
			f.staticAppend(ctx, label)
			f.staticPause(ctx)
		}
	})); if __e != nil { return }
//line static_fragments.gox:151
		__e = __c.Any(test.Button("c-plain", func(ctx context.Context) bool {
		f.staticWatch(f.staticAppend(doors.DetachedContext(ctx), "p"))
		return false
	})); if __e != nil { return }
//line static_fragments.gox:155
		__e = __c.Any(f.staticBatched("d-batched", func(ctx context.Context) {
		f.x.Static(ctx, f.staticScroll(false))
		f.staticPause(ctx)
	})); if __e != nil { return }
//line static_fragments.gox:159
		__e = __c.Any(f.staticBatched("d-detached", func(ctx context.Context) {
		f.x.Static(ctx, f.staticScroll(true))
		f.staticPause(ctx)
	})); if __e != nil { return }
	return })
//line static_fragments.gox:163
}

var _ gox.Comp = (*staticFragment)(nil)
