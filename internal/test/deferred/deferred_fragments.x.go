// Managed by GoX v0.3.2

//line deferred_fragments.gox:1
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

package deferred

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type Shared struct {
	mu sync.Mutex
	block chan struct{}
	skelBlock chan struct{}
	entered chan string
	results chan string
	counts map[string]int
	probes map[string]<-chan error
	queued map[string]int
}

func NewShared() *Shared {
	return &Shared{
		entered: make(chan string, 64),
		results: make(chan string, 64),
		counts: map[string]int{},
		probes: map[string]<-chan error{},
		queued: map[string]int{},
	}
}

func (s *Shared) count(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts[key]++
}

func (s *Shared) Count(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.counts[key]
}

func (s *Shared) hold(block *chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if *block == nil {
		*block = make(chan struct{})
	}
}

func (s *Shared) release(block *chan struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if *block != nil {
		close(*block)
		*block = nil
	}
}

func (s *Shared) Hold() {
	s.hold(&s.block)
}

func (s *Shared) Release() {
	s.release(&s.block)
}

func (s *Shared) HoldSkeleton() {
	s.hold(&s.skelBlock)
}

func (s *Shared) ReleaseSkeleton() {
	s.release(&s.skelBlock)
}

func (s *Shared) Probe(label string, ch <-chan error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.probes[label] = ch
}

func (s *Shared) Probed(label string) (<-chan error, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.probes[label]
	return ch, ok
}

func (s *Shared) Queued(label string) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.queued[label]
	return n, ok
}

func (s *Shared) pass(label string) {
	s.mu.Lock()
	s.counts["heavy " + label]++
	if probe, ok := s.probes["load " + label]; ok {
		s.queued[label] = len(probe)
	}
	block := s.block
	s.mu.Unlock()
	select {
	case s.entered <- label:
	default:
	}
	if block != nil {
		<-block
	}
}

func (s *Shared) passSkeleton(label string) {
	s.mu.Lock()
	block := s.skelBlock
	s.mu.Unlock()
	select {
	case s.entered <- "skel " + label:
	default:
	}
	if block != nil {
		<-block
	}
}

func (s *Shared) track(label string, ch <-chan error) {
	go func() {
		var out []string
		for err := range ch {
			if err == nil {
				out = append(out, "nil")
				continue
			}
			out = append(out, err.Error())
		}
		s.results <- label + ":" + strings.Join(out, ",")
	}()
}

//line deferred_fragments.gox:159
func (s *Shared) heavy(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:161
		s.pass(label)

		__e = __c.Init("b"); if __e != nil { return }
		{
//line deferred_fragments.gox:163
			__e = __c.Set("class", "heavy"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:163
			__e = __c.Any("heavy " + label); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line deferred_fragments.gox:164
}

//line deferred_fragments.gox:166
func (s *Shared) skeleton(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("i"); if __e != nil { return }
		{
//line deferred_fragments.gox:167
			__e = __c.Set("class", "skel"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:167
			__e = __c.Any("skel " + label); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line deferred_fragments.gox:168
}

//line deferred_fragments.gox:170
func (s *Shared) innerLoader(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:172
		s.count("deferred " + label)
	s.track("deferred " + label, doors.DeferredInner(ctx, s.heavy(label)))

//line deferred_fragments.gox:175
		__e = __c.Any(s.skeleton(label)); if __e != nil { return }
	return })
//line deferred_fragments.gox:176
}

//line deferred_fragments.gox:178
func (s *Shared) gatedLoader(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:180
		s.count("deferred " + label)
	s.track("deferred " + label, doors.DeferredInner(ctx, s.heavy(label)))
	s.passSkeleton(label)

//line deferred_fragments.gox:184
		__e = __c.Any(s.skeleton(label)); if __e != nil { return }
	return })
//line deferred_fragments.gox:185
}

//line deferred_fragments.gox:187
func (s *Shared) outerLoader(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:189
		s.count("deferred " + label)
	s.track("deferred " + label, doors.DeferredOuter(ctx, s.heavy(label)))

//line deferred_fragments.gox:192
		__e = __c.Any(s.skeleton(label)); if __e != nil { return }
	return })
//line deferred_fragments.gox:193
}

//line deferred_fragments.gox:195
func (s *Shared) staticLoader(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:197
		s.count("deferred " + label)
	s.track("deferred " + label, doors.DeferredStatic(ctx, s.heavy(label)))

//line deferred_fragments.gox:200
		__e = __c.Any(s.skeleton(label)); if __e != nil { return }
	return })
//line deferred_fragments.gox:201
}

//line deferred_fragments.gox:203
func (s *Shared) documentLoader(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:204
		__e = __c.Any(func() any {
		if doors.IsDocument(ctx) {
			return s.heavy(label)
		}
		s.count("deferred " + label)
		s.track("deferred " + label, doors.DeferredInner(ctx, s.heavy(label)))
		return s.skeleton(label)
	}()); if __e != nil { return }
	return })
//line deferred_fragments.gox:212
}

type Sentinel struct {
	d doors.Door
	n atomic.Int32
}

//line deferred_fragments.gox:219
func (s *Sentinel) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line deferred_fragments.gox:220
			__e = __c.Set("id", "sentinel"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:220
			__e = __c.Any(&s.d); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:221
		__e = __c.Any(test.Button("sentinel-ping", func(ctx context.Context) bool {
		s.d.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:222
				__e = __c.Any(fmt.Sprint("ping ", s.n.Add(1))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:222
		return }))
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:225
}

type FreshFragment struct {
	S *Shared
	parent doors.Door
	loose doors.Door
	sentinel Sentinel
	test.NoBeam
}

//line deferred_fragments.gox:235
func (f *FreshFragment) blendHost(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:237
		d := &doors.Door{}

//line deferred_fragments.gox:239
		__e = (d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line deferred_fragments.gox:239
				__e = __c.Set("class", "door"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:239
				__e = __c.Any(f.S.innerLoader(label)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line deferred_fragments.gox:240
}

//line deferred_fragments.gox:242
func (f *FreshFragment) innerHost(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:244
		d := &doors.Door{}
	d.Inner(ctx, f.S.innerLoader(label))

//line deferred_fragments.gox:247
		__e = __c.Any(d); if __e != nil { return }
	return })
//line deferred_fragments.gox:248
}

//line deferred_fragments.gox:250
func (f *FreshFragment) documentHost(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:252
		d := &doors.Door{}

//line deferred_fragments.gox:254
		__e = (d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line deferred_fragments.gox:254
				__e = __c.Set("class", "door"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:254
				__e = __c.Any(f.S.documentLoader(label)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line deferred_fragments.gox:255
}

//line deferred_fragments.gox:257
func (f *FreshFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:258
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:258
			__e = __c.Any(&f.parent); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:259
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:260
		__e = __c.Any(test.Button("place-blend", func(ctx context.Context) bool {
		f.parent.Inner(ctx, f.blendHost("blend"))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:264
		__e = __c.Any(test.Button("place-inner", func(ctx context.Context) bool {
		f.parent.Inner(ctx, f.innerHost("inner"))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:268
		__e = __c.Any(test.Button("place-document", func(ctx context.Context) bool {
		f.parent.Inner(ctx, f.documentHost("document"))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:272
		__e = __c.Any(test.Button("defer-loose", func(ctx context.Context) bool {
		f.S.track("loose", f.loose.DeferredInner(ctx, f.S.heavy("loose")))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:276
		__e = __c.Any(test.Button("place-loose", func(ctx context.Context) bool {
		f.parent.Inner(ctx, &f.loose)
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:280
}

type BindFragment struct {
	S *Shared
	Outer bool
	src doors.Source[int]
	sentinel Sentinel
	test.NoBeam
}

//line deferred_fragments.gox:290
func (f *BindFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:292
		f.src = doors.NewSource(0)

		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:294
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:295
			__e = __c.Any(f.src.Bind(func(v int) gox.Elem {
			if f.Outer {
				return f.S.outerLoader(fmt.Sprint(v))
			}
			return f.S.innerLoader(fmt.Sprint(v))
		})); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:302
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:303
		__e = __c.Any(test.Button("next", func(ctx context.Context) bool {
		f.src.Mutate(ctx, func(v int) int {
			return v + 1
		})
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:309
}

type MountedFragment struct {
	S *Shared
	d doors.Door
	sentinel Sentinel
	test.NoBeam
}

//line deferred_fragments.gox:318
func (f *MountedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:320
		f.d.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("initial"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:320
	return }))

		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:322
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:322
			__e = __c.Any(&f.d); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:323
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:324
		__e = __c.Any(test.Button("load", func(ctx context.Context) bool {
		f.S.track("load", f.d.Inner(ctx, f.S.innerLoader("mounted")))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:328
		__e = __c.Any(test.Button("load-gated", func(ctx context.Context) bool {
		f.S.Probe("load gated", f.d.Inner(ctx, f.S.gatedLoader("gated")))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:332
		__e = __c.Any(test.Button("newer-deferred", func(ctx context.Context) bool {
		f.S.Probe("newer", f.d.DeferredInner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("newer"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:333
		return })))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:336
		__e = __c.Any(test.Button("plain-inner", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("plain"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:337
		return })))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:340
		__e = __c.Any(test.Button("plain-outer", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Outer(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("plain"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:341
		return })))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:344
		__e = __c.Any(test.Button("plain-static", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Static(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("plain"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:345
		return })))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:348
		__e = __c.Any(test.Button("plain-reload", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Reload(ctx))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:352
		__e = __c.Any(test.Button("plain-unmount", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Unmount(ctx))
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:356
}

type ReloadFragment struct {
	S *Shared
	blend doors.Door
	outer doors.Door
	inner doors.Door
	plain doors.Door
	sentinel Sentinel
	test.NoBeam
}

//line deferred_fragments.gox:368
func (f *ReloadFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:370
		f.outer.Outer(ctx, f.S.documentLoader("outer"))
	f.inner.Inner(ctx, f.S.documentLoader("inner"))

		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:373
			__e = __c.Set("id", "area-blend"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:374
			__e = (f.blend).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line deferred_fragments.gox:374
					__e = __c.Set("class", "door"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:374
					__e = __c.Any(f.S.documentLoader("blend")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:376
			__e = __c.Set("id", "area-outer"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:376
			__e = __c.Any(&f.outer); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:377
			__e = __c.Set("id", "area-inner"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:377
			__e = __c.Any(&f.inner); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:378
			__e = __c.Set("id", "area-plain"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:379
			__e = (f.plain).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line deferred_fragments.gox:379
					__e = __c.Set("class", "door"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:379
					__e = __c.Any(f.S.innerLoader("plain")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:381
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:382
		__e = __c.Any(test.Button("reload-blend", func(ctx context.Context) bool {
		f.S.track("reload blend", f.blend.Reload(ctx))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:386
		__e = __c.Any(test.Button("reload-outer", func(ctx context.Context) bool {
		f.S.track("reload outer", f.outer.Reload(ctx))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:390
		__e = __c.Any(test.Button("reload-inner", func(ctx context.Context) bool {
		f.S.track("reload inner", f.inner.Reload(ctx))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:394
		__e = __c.Any(test.Button("reload-plain", func(ctx context.Context) bool {
		f.S.track("reload plain", f.plain.Reload(ctx))
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:398
}

type StaticFragment struct {
	S *Shared
	d doors.Door
	sentinel Sentinel
	test.NoBeam
}

//line deferred_fragments.gox:407
func (f *StaticFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:408
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:409
			__e = (f.d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line deferred_fragments.gox:409
					__e = __c.Set("class", "door"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:409
					__e = __c.Any(f.S.staticLoader("static")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:411
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:412
		__e = __c.Any(test.Button("after", func(ctx context.Context) bool {
		f.S.track("after", f.d.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("after"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:413
		return })))
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:416
}

type RootFragment struct {
	S *Shared
	test.NoBeam
}

//line deferred_fragments.gox:423
func (f *RootFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:424
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:424
			__e = __c.Any(f.S.innerLoader("root")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line deferred_fragments.gox:425
}
