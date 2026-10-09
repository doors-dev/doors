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

//line deferred_fragments.gox:214
func (s *Shared) sectionLoader(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:215
			__e = __c.Set("class", "wrapper"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:216
			if doors.IsDocument(ctx) {
//line deferred_fragments.gox:217
				__e = __c.Any(s.heavy(label)); if __e != nil { return }
			} else  {
//line deferred_fragments.gox:220
				s.count("deferred " + label)
			s.track("deferred " + label, doors.DeferredInner(ctx, s.heavy(label)))

//line deferred_fragments.gox:223
				__e = __c.Any(s.skeleton(label)); if __e != nil { return }
			}
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line deferred_fragments.gox:226
}

//line deferred_fragments.gox:228
func (s *Shared) swapLoader(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:229
		if doors.IsDocument(ctx) {
//line deferred_fragments.gox:230
			__e = __c.Any(s.heavy(label)); if __e != nil { return }
		} else  {
//line deferred_fragments.gox:233
			s.count("deferred " + label)
		s.track("deferred " + label, doors.DeferredOuter(ctx, s.heavy(label)))

//line deferred_fragments.gox:236
			__e = __c.Any(s.skeleton(label)); if __e != nil { return }
		}
	return })
//line deferred_fragments.gox:238
}

type Sentinel struct {
	d doors.Door
	n atomic.Int32
}

//line deferred_fragments.gox:245
func (s *Sentinel) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line deferred_fragments.gox:246
			__e = __c.Set("id", "sentinel"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:246
			__e = __c.Any(&s.d); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:247
		__e = __c.Any(test.Button("sentinel-ping", func(ctx context.Context) bool {
		s.d.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:248
				__e = __c.Any(fmt.Sprint("ping ", s.n.Add(1))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:248
		return }))
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:251
}

type FreshFragment struct {
	S *Shared
	parent doors.Door
	loose doors.Door
	sentinel Sentinel
	test.NoBeam
}

//line deferred_fragments.gox:261
func (f *FreshFragment) blendHost(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:263
		d := &doors.Door{}

//line deferred_fragments.gox:265
		__e = (d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line deferred_fragments.gox:265
				__e = __c.Set("class", "door"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:265
				__e = __c.Any(f.S.innerLoader(label)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line deferred_fragments.gox:266
}

//line deferred_fragments.gox:268
func (f *FreshFragment) innerHost(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:270
		d := &doors.Door{}
	d.Inner(ctx, f.S.innerLoader(label))

//line deferred_fragments.gox:273
		__e = __c.Any(d); if __e != nil { return }
	return })
//line deferred_fragments.gox:274
}

//line deferred_fragments.gox:276
func (f *FreshFragment) documentHost(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:278
		d := &doors.Door{}

//line deferred_fragments.gox:280
		__e = (d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line deferred_fragments.gox:280
				__e = __c.Set("class", "door"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:280
				__e = __c.Any(f.S.documentLoader(label)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line deferred_fragments.gox:281
}

//line deferred_fragments.gox:283
func (f *FreshFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:284
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:284
			__e = __c.Any(&f.parent); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:285
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:286
		__e = __c.Any(test.Button("place-blend", func(ctx context.Context) bool {
		f.parent.Inner(ctx, f.blendHost("blend"))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:290
		__e = __c.Any(test.Button("place-inner", func(ctx context.Context) bool {
		f.parent.Inner(ctx, f.innerHost("inner"))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:294
		__e = __c.Any(test.Button("place-document", func(ctx context.Context) bool {
		f.parent.Inner(ctx, f.documentHost("document"))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:298
		__e = __c.Any(test.Button("defer-loose", func(ctx context.Context) bool {
		f.S.track("loose", f.loose.DeferredInner(ctx, f.S.heavy("loose")))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:302
		__e = __c.Any(test.Button("place-loose", func(ctx context.Context) bool {
		f.parent.Inner(ctx, &f.loose)
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:306
}

type BindFragment struct {
	S *Shared
	Outer bool
	Section bool
	src doors.Source[int]
	sentinel Sentinel
	test.NoBeam
}

//line deferred_fragments.gox:317
func (f *BindFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:319
		f.src = doors.NewSource(0)

		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:321
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:322
			__e = __c.Any(f.src.Bind(func(v int) gox.Elem {
			if f.Section {
				return f.S.sectionLoader(fmt.Sprint(v))
			}
			if f.Outer {
				return f.S.outerLoader(fmt.Sprint(v))
			}
			return f.S.innerLoader(fmt.Sprint(v))
		})); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:332
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:333
		__e = __c.Any(test.Button("next", func(ctx context.Context) bool {
		f.src.Mutate(ctx, func(v int) int {
			return v + 1
		})
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:339
}

type MountedFragment struct {
	S *Shared
	d doors.Door
	sentinel Sentinel
	test.NoBeam
}

//line deferred_fragments.gox:348
func (f *MountedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:350
		f.d.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("initial"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:350
	return }))

		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:352
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:352
			__e = __c.Any(&f.d); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:353
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:354
		__e = __c.Any(test.Button("load", func(ctx context.Context) bool {
		f.S.track("load", f.d.Inner(ctx, f.S.innerLoader("mounted")))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:358
		__e = __c.Any(test.Button("load-gated", func(ctx context.Context) bool {
		f.S.Probe("load gated", f.d.Inner(ctx, f.S.gatedLoader("gated")))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:362
		__e = __c.Any(test.Button("skeleton-heavy", func(ctx context.Context) bool {
		f.S.track("skeleton", f.d.Inner(ctx, f.S.skeleton("method")))
		f.S.track("deferred method", f.d.DeferredInner(ctx, f.S.heavy("method")))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:367
		__e = __c.Any(test.Button("newer-deferred", func(ctx context.Context) bool {
		f.S.Probe("newer", f.d.DeferredInner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("newer"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:368
		return })))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:371
		__e = __c.Any(test.Button("plain-inner", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("plain"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:372
		return })))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:375
		__e = __c.Any(test.Button("plain-outer", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Outer(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("plain"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:376
		return })))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:379
		__e = __c.Any(test.Button("plain-static", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Static(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("plain"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:380
		return })))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:383
		__e = __c.Any(test.Button("plain-reload", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Reload(ctx))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:387
		__e = __c.Any(test.Button("plain-unmount", func(ctx context.Context) bool {
		f.S.track("plain", f.d.Unmount(ctx))
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:391
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

//line deferred_fragments.gox:403
func (f *ReloadFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line deferred_fragments.gox:405
		f.outer.Outer(ctx, f.S.documentLoader("outer"))
	f.inner.Inner(ctx, f.S.documentLoader("inner"))

		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:408
			__e = __c.Set("id", "area-blend"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:409
			__e = (f.blend).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line deferred_fragments.gox:409
					__e = __c.Set("class", "door"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:409
					__e = __c.Any(f.S.documentLoader("blend")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:411
			__e = __c.Set("id", "area-outer"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:411
			__e = __c.Any(&f.outer); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:412
			__e = __c.Set("id", "area-inner"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:412
			__e = __c.Any(&f.inner); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:413
			__e = __c.Set("id", "area-plain"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:414
			__e = (f.plain).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line deferred_fragments.gox:414
					__e = __c.Set("class", "door"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:414
					__e = __c.Any(f.S.innerLoader("plain")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:416
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:417
		__e = __c.Any(test.Button("reload-blend", func(ctx context.Context) bool {
		f.S.track("reload blend", f.blend.Reload(ctx))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:421
		__e = __c.Any(test.Button("reload-outer", func(ctx context.Context) bool {
		f.S.track("reload outer", f.outer.Reload(ctx))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:425
		__e = __c.Any(test.Button("reload-inner", func(ctx context.Context) bool {
		f.S.track("reload inner", f.inner.Reload(ctx))
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:429
		__e = __c.Any(test.Button("reload-plain", func(ctx context.Context) bool {
		f.S.track("reload plain", f.plain.Reload(ctx))
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:433
}

type StaticFragment struct {
	S *Shared
	d doors.Door
	sentinel Sentinel
	test.NoBeam
}

//line deferred_fragments.gox:442
func (f *StaticFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:443
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:444
			__e = (f.d).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line deferred_fragments.gox:444
					__e = __c.Set("class", "door"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:444
					__e = __c.Any(f.S.staticLoader("static")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:446
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:447
		__e = __c.Any(test.Button("after", func(ctx context.Context) bool {
		f.S.track("after", f.d.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("after"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:448
		return })))
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:451
}

type RouteFragment struct {
	S *Shared
	sentinel Sentinel
	test.Beam
}

func routeTo(name string) func(test.Path) bool {
	return func(p test.Path) bool {
		return p.Vp && p.P == name
	}
}

//line deferred_fragments.gox:465
func (f *RouteFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line deferred_fragments.gox:466
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:467
			__e = __c.Any(f.B.Route(
			doors.RouteMatch(routeTo("inner")).Comp(f.S.sectionLoader("inner")),
			doors.RouteMatch(routeTo("outer")).Comp(f.S.swapLoader("outer")),
			doors.RouteDefaultComp[test.Path](gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("p"); if __e != nil { return }
				{
//line deferred_fragments.gox:470
					__e = __c.Set("class", "home"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("home"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:470
			return })),
		)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line deferred_fragments.gox:473
		__e = __c.Any(&f.sentinel); if __e != nil { return }
//line deferred_fragments.gox:474
		__e = __c.Any(test.Button("go-home", func(ctx context.Context) bool {
		f.B.Update(ctx, test.Path{Vh: true})
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:478
		__e = __c.Any(test.Button("go-inner", func(ctx context.Context) bool {
		f.B.Update(ctx, test.Path{Vp: true, P: "inner"})
		return false
	})); if __e != nil { return }
//line deferred_fragments.gox:482
		__e = __c.Any(test.Button("go-outer", func(ctx context.Context) bool {
		f.B.Update(ctx, test.Path{Vp: true, P: "outer"})
		return false
	})); if __e != nil { return }
	return })
//line deferred_fragments.gox:486
}

type RootFragment struct {
	S *Shared
	test.NoBeam
}

//line deferred_fragments.gox:493
func (f *RootFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line deferred_fragments.gox:494
			__e = __c.Set("id", "area"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line deferred_fragments.gox:494
			__e = __c.Any(f.S.innerLoader("root")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line deferred_fragments.gox:495
}
