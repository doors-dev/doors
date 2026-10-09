// Managed by GoX v0.3.2

//line beam_fragments.gox:1
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

package beam

import (
	"context"
	"fmt"
	"strings"
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type state struct {
	Int int
	Str string
}

type BeamSkipFragment struct {
	r *test.Reporter
	b doors.Source[state]
	node doors.Door
	test.NoBeam
}

//line beam_fragments.gox:40
func (f *BeamSkipFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:42
		f.r.Update(ctx, 0, "init")
	f.b.ReadAndSub(ctx, func(ctx context.Context, s state) bool {
		<-time.After(300 * time.Millisecond)
		return false
	})

//line beam_fragments.gox:48
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:50
				f.b.Sub(ctx, func(ctx context.Context, s state) bool {
			if s.Str == "1" {
				f.r.Update(ctx, 0, "propagated")
			}
			return false
		})

			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:58
		__e = __c.Any(test.Button("update1", func(ctx context.Context) bool {
		f.b.Update(ctx, state{Str: "1"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:62
		__e = __c.Any(test.Button("update2", func(ctx context.Context) bool {
		f.b.Update(ctx, state{Str: "2"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:66
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:67
}

type BeamDeriveFragment struct {
	r *test.Reporter
	b doors.Source[state]
	n doors.Door
	test.NoBeam
}
//line beam_fragments.gox:75
func (f *BeamDeriveFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:76
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:77
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:79
		__e = __c.Any(test.Button("reload", func(ctx context.Context) bool {
		f.n.Inner(ctx, f.content())
		return true
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:83
}

//line beam_fragments.gox:85
func (f *BeamDeriveFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:87
		d := doors.DeriveBeam(f.b, func(s state) int {
		return s.Int
	})
	f.b.Sub(ctx, func(ctx context.Context, s state) bool {
		f.r.Update(ctx, 0, fmt.Sprint(s.Int))
		return false
	})
	n1 := doors.Door{}
	n2 := doors.Door{}
	f.b.Mutate(ctx, func(s state) state {
		s.Int = s.Int + 1
		return s
	})
	r, _ := d.Read(ctx)

//line beam_fragments.gox:102
		__e = __c.Any(test.ReportId(1, fmt.Sprint(r))); if __e != nil { return }
//line beam_fragments.gox:103
		__e = (n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:105
				f.b.Mutate(ctx, func(s state) state {
			s.Int = s.Int + 1
			return s
		})
		r, _ := d.Read(ctx)

//line beam_fragments.gox:111
				__e = __c.Any(test.ReportId(2, fmt.Sprint(r))); if __e != nil { return }
//line beam_fragments.gox:113
				n3 := doors.Door{}
		d.Sub(ctx, func(ctx context.Context, s int) bool {
			n3.Inner(ctx, test.ReportId(4, fmt.Sprint(s)))
			return false
		})

//line beam_fragments.gox:119
				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:121
		__e = (n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:123
				f.b.Mutate(ctx, func(s state) state {
			s.Int = s.Int + 1
			return s
		})
		r, _ := f.b.Read(ctx)

//line beam_fragments.gox:129
				__e = __c.Any(test.ReportId(3, fmt.Sprint(r.Int))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:131
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:132
}

type BeamConsistentFragment struct {
	r *test.Reporter
	b doors.Source[state]
	n doors.Door
	test.NoBeam
}

//line beam_fragments.gox:141
func (f *BeamConsistentFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:143
		f.b.Sub(ctx, func(ctx context.Context, s state) bool {
		f.r.Update(ctx, 0, fmt.Sprint(s.Int))
		return false
	})
	n1 := doors.Door{}
	n2 := doors.Door{}
	f.b.Mutate(ctx, func(s state) state {
		s.Int = s.Int + 1
		return s
	})
	r, _ := f.b.Read(ctx)

//line beam_fragments.gox:155
		__e = __c.Any(test.ReportId(1, fmt.Sprint(r.Int))); if __e != nil { return }
//line beam_fragments.gox:156
		__e = (n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:158
				f.b.Mutate(ctx, func(s state) state {
			s.Int = s.Int + 1
			return s
		})
		r, _ := f.b.Read(ctx)

//line beam_fragments.gox:164
				__e = __c.Any(test.ReportId(2, fmt.Sprint(r.Int))); if __e != nil { return }
//line beam_fragments.gox:166
				n3 := doors.Door{}
		f.b.Sub(ctx, func(ctx context.Context, s state) bool {
			n3.Inner(ctx, test.ReportId(4, fmt.Sprint(s.Int)))
			return false
		})

//line beam_fragments.gox:172
				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:174
		__e = (n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:176
				f.b.Mutate(ctx, func(s state) state {
			s.Int = s.Int + 1
			return s
		})
		r, _ := f.b.Read(ctx)

//line beam_fragments.gox:182
				__e = __c.Any(test.ReportId(3, fmt.Sprint(r.Int))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:184
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:185
}

//line beam_fragments.gox:187
func (f *BeamConsistentFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:188
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:189
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:191
		__e = __c.Any(test.Button("reload", func(ctx context.Context) bool {
		f.n.Reload(ctx)
		return true
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:195
}

type BeamUpdateFragment struct {
	r *test.Reporter
	b doors.Source[state]
	test.NoBeam
}

//line beam_fragments.gox:203
func (f *BeamUpdateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:205
		f.b.Sub(ctx, func(ctx context.Context, s state) bool {
		f.r.Update(ctx, 0, fmt.Sprint(s.Int))
		return false
	})

//line beam_fragments.gox:211
		__e = __c.Many(test.Button("update", func(ctx context.Context) bool {
			f.b.Update(ctx, state{
				Int: 1,
			})
			return true
		}),
		test.Button("mutate", func(ctx context.Context) bool {
			f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			return true
		}),
		test.Button("mutate-cancel", func(ctx context.Context) bool {
			f.b.Mutate(ctx, func(s state) state {
				return s
			})
			return true
		}),
		f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:232
}

type BeamEqualFragment struct {
	r *test.Reporter
	b doors.Source[state]
	p doors.Beam[string]
	test.NoBeam
}

//line beam_fragments.gox:241
func (f *BeamEqualFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:243
		if f.p == nil {
		f.p = doors.DeriveBeamEqual(f.b, func(s state) string {
			if s.Int % 2 == 0 {
				return "even"
			}
			return "odd"
		}, func(old string, new string) bool {
			return old == new
		})
	}
	f.b.Sub(ctx, func(ctx context.Context, s state) bool {
		f.r.Update(ctx, 0, fmt.Sprint(s.Int))
		return false
	})

//line beam_fragments.gox:258
		__e = __c.Any(f.p.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:259
				__e = __c.Set("id", "parity"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:259
				__e = __c.Any(v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:260
	})); if __e != nil { return }
//line beam_fragments.gox:261
		__e = __c.Any(doors.Go(func(ctx context.Context) {
		<-time.After(100 * time.Millisecond)
		f.r.Update(ctx, 2, "go")
	})); if __e != nil { return }
//line beam_fragments.gox:266
		__e = __c.Many(test.Button("same", func(ctx context.Context) bool {
			f.b.Update(ctx, state{
				Int: 0,
				Str: "same",
			})
			return false
		}),
		test.Button("one", func(ctx context.Context) bool {
			f.b.Update(ctx, state{
				Int: 1,
			})
			return false
		}),
		test.Button("three", func(ctx context.Context) bool {
			f.b.Update(ctx, state{
				Int: 3,
			})
			return false
		}),
		test.Button("get", func(ctx context.Context) bool {
			f.r.Update(ctx, 1, fmt.Sprint(f.b.Get().Int))
			return false
		}),
		f.r,); if __e != nil { return }
	return })
//line beam_fragments.gox:291
}

type BeamRenderBranchUpdateFrameFragment struct {
	b doors.Source[int]
	n doors.Door
	test.NoBeam
}

//line beam_fragments.gox:299
func (f *BeamRenderBranchUpdateFrameFragment) content(i int) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:300
			__e = __c.Set("id", "watcher-i"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:300
			__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:302
		f.b.Mutate(ctx, func(i int) int {
		return i + 1
	})
	newI, _ := f.b.Read(ctx)

		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:307
			__e = __c.Set("id", "watcher-newi"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:307
			__e = __c.Any(fmt.Sprint(newI)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:308
}

//line beam_fragments.gox:310
func (f *BeamRenderBranchUpdateFrameFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:311
		__e = __c.Any(&f.n); if __e != nil { return }
//line beam_fragments.gox:313
		f.b.ReadAndSub(ctx, func(ctx context.Context, i int) bool {
		f.n.Inner(ctx, f.content(i))
		return true
	})
	f.b.Mutate(ctx, func(i int) int {
		return i + 1
	})

	return })
//line beam_fragments.gox:321
}

type BeamRenderBranchInitFrameFragment struct {
	b doors.Source[int]
	n doors.Door
	test.NoBeam
}

//line beam_fragments.gox:329
func (f *BeamRenderBranchInitFrameFragment) content(i int) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:330
			__e = __c.Set("id", "watcher-i"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:330
			__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:332
		f.b.Mutate(ctx, func(i int) int {
		return i + 1
	})
	newI, _ := f.b.Read(ctx)

		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:337
			__e = __c.Set("id", "watcher-newi"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:337
			__e = __c.Any(fmt.Sprint(newI)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:338
}

//line beam_fragments.gox:340
func (f *BeamRenderBranchInitFrameFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:341
		__e = __c.Any(&f.n); if __e != nil { return }
//line beam_fragments.gox:343
		go func() {
		f.b.Sub(ctx, func(ctx context.Context, i int) bool {
			f.n.Inner(ctx, f.content(i))
			return true
		})
	}()

	return })
//line beam_fragments.gox:350
}

type BeamRenderUpdateWarningFragment struct {
	b doors.Source[int]
	host doors.Door
	test.NoBeam
}

//line beam_fragments.gox:358
func (f *BeamRenderUpdateWarningFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:360
		n3 := doors.Door{}
	_, _ = f.b.Read(ctx)
	f.b.Sub(ctx, func(ctx context.Context, i int) bool {
		n3.Inner(ctx, test.ReportId(4, fmt.Sprint(i)))
		return false
	})

//line beam_fragments.gox:367
		__e = __c.Any(&n3); if __e != nil { return }
	return })
//line beam_fragments.gox:368
}

//line beam_fragments.gox:370
func (f *BeamRenderUpdateWarningFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:371
		__e = (f.host).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:372
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:374
		__e = __c.Any(test.Button("warning-reload", func(ctx context.Context) bool {
		f.host.Reload(ctx)
		return true
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:378
}

type BeamEffectSourceFragment struct {
	b doors.Source[int]
	frame doors.Door
	host doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line beam_fragments.gox:389
func (f *BeamEffectSourceFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:391
		f.innerRenders++
	value, _ := f.b.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:394
			__e = __c.Set("id", "effect-source-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:394
			__e = __c.Any(fmt.Sprint(value)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:395
			__e = __c.Set("id", "effect-source-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:395
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:396
}

//line beam_fragments.gox:398
func (f *BeamEffectSourceFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:400
		f.outerRenders++
	f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:403
			__e = __c.Set("id", "effect-source-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:403
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:404
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:405
}

//line beam_fragments.gox:407
func (f *BeamEffectSourceFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:409
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:411
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:412
		__e = __c.Any(test.Button("effect-source-update-1", func(ctx context.Context) bool {
		f.b.Update(ctx, 1)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:416
		__e = __c.Any(test.Button("effect-source-update-2", func(ctx context.Context) bool {
		f.b.Update(ctx, 2)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:420
}

type BeamEffectDerivedFragment struct {
	b doors.Source[int]
	d doors.Beam[string]
	frame doors.Door
	host doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line beam_fragments.gox:432
func (f *BeamEffectDerivedFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:434
		f.innerRenders++
	value, _ := f.d.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:437
			__e = __c.Set("id", "effect-derived-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:437
			__e = __c.Any(value); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:438
			__e = __c.Set("id", "effect-derived-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:438
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:439
}

//line beam_fragments.gox:441
func (f *BeamEffectDerivedFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:443
		f.outerRenders++
	f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:446
			__e = __c.Set("id", "effect-derived-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:446
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:447
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:448
}

//line beam_fragments.gox:450
func (f *BeamEffectDerivedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:452
		if f.d == nil {
		f.d = doors.DeriveBeam(f.b, func(v int) string {
			return fmt.Sprintf("v:%d", v)
		})
	}
	f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:459
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:460
		__e = __c.Any(test.Button("effect-derived-update-1", func(ctx context.Context) bool {
		f.b.Update(ctx, 1)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:464
		__e = __c.Any(test.Button("effect-derived-update-2", func(ctx context.Context) bool {
		f.b.Update(ctx, 2)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:468
}

type BeamEffectMultiFragment struct {
	left doors.Source[int]
	right doors.Source[int]
	frame doors.Door
	host doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line beam_fragments.gox:480
func (f *BeamEffectMultiFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:482
		f.innerRenders++
	left, _ := f.left.Effect(ctx)
	right, _ := f.right.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:486
			__e = __c.Set("id", "effect-multi-left"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:486
			__e = __c.Any(fmt.Sprint(left)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:487
			__e = __c.Set("id", "effect-multi-right"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:487
			__e = __c.Any(fmt.Sprint(right)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:488
			__e = __c.Set("id", "effect-multi-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:488
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:489
}

//line beam_fragments.gox:491
func (f *BeamEffectMultiFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:493
		f.outerRenders++
	f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:496
			__e = __c.Set("id", "effect-multi-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:496
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:497
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:498
}

//line beam_fragments.gox:500
func (f *BeamEffectMultiFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:502
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:504
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:505
		__e = __c.Any(test.Button("effect-multi-left-update", func(ctx context.Context) bool {
		f.left.Update(ctx, 1)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:509
		__e = __c.Any(test.Button("effect-multi-right-update", func(ctx context.Context) bool {
		f.right.Update(ctx, 1)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:513
}

type BeamEffectDuplicateFragment struct {
	b doors.Source[int]
	frame doors.Door
	host doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line beam_fragments.gox:524
func (f *BeamEffectDuplicateFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:526
		f.innerRenders++
	first, _ := f.b.Effect(ctx)
	second, _ := f.b.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:530
			__e = __c.Set("id", "effect-dup-first"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:530
			__e = __c.Any(fmt.Sprint(first)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:531
			__e = __c.Set("id", "effect-dup-second"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:531
			__e = __c.Any(fmt.Sprint(second)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:532
			__e = __c.Set("id", "effect-dup-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:532
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:533
}

//line beam_fragments.gox:535
func (f *BeamEffectDuplicateFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:537
		f.outerRenders++
	f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:540
			__e = __c.Set("id", "effect-dup-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:540
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:541
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:542
}

//line beam_fragments.gox:544
func (f *BeamEffectDuplicateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:546
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:548
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:549
		__e = __c.Any(test.Button("effect-dup-update", func(ctx context.Context) bool {
		f.b.Update(ctx, 1)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:553
}

type BeamReadAndSubFragment struct {
	source doors.Source[int]
	derived doors.Beam[string]
	r *test.Reporter
	derivedRegistered bool
	sourceRegistered bool
	derived2Registered bool
	test.NoBeam
}

//line beam_fragments.gox:565
func (f *BeamReadAndSubFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:567
		if f.derived == nil {
		f.derived = doors.DeriveBeam(f.source, func(v int) string {
			return fmt.Sprintf("v:%d", v)
		})
	}
	if !f.derivedRegistered {
		initial, ok := f.derived.ReadAndSub(ctx, func(ctx context.Context, value string) bool {
			f.r.Update(ctx, 1, value)
			return true
		})
		if ok {
			f.r.Update(ctx, 0, initial)
			f.derivedRegistered = true
		}
	}

//line beam_fragments.gox:583
		__e = __c.Any(test.Button("beam-read-sub-update-2", func(ctx context.Context) bool {
		f.source.Update(ctx, 2)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:587
		__e = __c.Any(test.Button("beam-read-sub-register-source", func(ctx context.Context) bool {
		if f.sourceRegistered {
			return false
		}
		initial, ok := f.source.ReadAndSub(ctx, func(ctx context.Context, value int) bool {
			f.r.Update(ctx, 3, fmt.Sprint(value))
			return true
		})
		if ok {
			f.r.Update(ctx, 2, fmt.Sprint(initial))
			f.sourceRegistered = true
		}
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:601
		__e = __c.Any(test.Button("beam-read-sub-update-3", func(ctx context.Context) bool {
		f.source.Update(ctx, 3)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:605
		__e = __c.Any(test.Button("beam-read-sub-register-derived-2", func(ctx context.Context) bool {
		if f.derived2Registered {
			return false
		}
		initial, ok := f.derived.ReadAndSub(ctx, func(ctx context.Context, value string) bool {
			f.r.Update(ctx, 5, value)
			return true
		})
		if ok {
			f.r.Update(ctx, 4, initial)
			f.derived2Registered = true
		}
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:619
		__e = __c.Any(test.Button("beam-read-sub-update-4", func(ctx context.Context) bool {
		f.source.Update(ctx, 4)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:623
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:624
}

type BeamLensRoundTripFragment struct {
	source doors.Source[state]
	intLens doors.Source[int]
	strLens doors.Source[string]
	evenLens doors.Source[bool]
	parityLens doors.Source[string]
	label doors.Beam[string]
	sourceRenders int
	intRenders int
	strRenders int
	evenRenders int
	labelRenders int
	parityRenders int
	r *test.Reporter
	test.NoBeam
}

func (f *BeamLensRoundTripFragment) init() {
	if f.source != nil {
		return
	}
	f.source = doors.NewSource(state{
		Int: 1,
		Str: "a",
	})
	f.intLens = doors.DeriveSource(f.source, func(s state) int {
		return s.Int
	}, func(s state, v int) state {
		s.Int = v
		return s
	})
	f.strLens = doors.DeriveSource(f.source, func(s state) string {
		return s.Str
	}, func(s state, v string) state {
		s.Str = v
		return s
	})
	f.evenLens = doors.DeriveSource(f.intLens, func(v int) bool {
		return v % 2 == 0
	}, func(v int, even bool) int {
		if even == (v % 2 == 0) {
			return v
		}
		return v + 1
	})
	f.parityLens = doors.DeriveSourceEqual(f.source, func(s state) string {
		if s.Int % 2 == 0 {
			return "even"
		}
		return "odd"
	}, func(s state, _ string) state {
		return s
	}, func(old string, new string) bool {
		return old == new
	})
	f.label = doors.DeriveBeam(f.intLens, func(v int) string {
		return fmt.Sprintf("label:%d", v)
	})
}

func (f *BeamLensRoundTripFragment) reportState(ctx context.Context, prefix string) {
	source, sourceOK := f.source.Read(ctx)
	intValue, intOK := f.intLens.Read(ctx)
	strValue, strOK := f.strLens.Read(ctx)
	evenValue, evenOK := f.evenLens.Read(ctx)
	label, labelOK := f.label.Read(ctx)
	f.r.Update(ctx, 0, fmt.Sprintf(
		"%s source-%d-%s-%t int-%d-%t str-%s-%t even-%t-%t label-%s-%t",
		prefix,
		source.Int,
		source.Str,
		sourceOK,
		intValue,
		intOK,
		strValue,
		strOK,
		evenValue,
		evenOK,
		label,
		labelOK,
	))
}

//line beam_fragments.gox:709
func (f *BeamLensRoundTripFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:711
		f.init()

//line beam_fragments.gox:713
		__e = __c.Any(f.source.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:715
			f.sourceRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:717
				__e = __c.Set("id", "lens-source"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:717
				__e = __c.Any(fmt.Sprintf("%d:%s:%d", v.Int, v.Str, f.sourceRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:718
	})); if __e != nil { return }
//line beam_fragments.gox:719
		__e = __c.Any(f.intLens.Bind(func(v int) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:721
			f.intRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:723
				__e = __c.Set("id", "lens-int"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:723
				__e = __c.Any(fmt.Sprintf("%d:%d", v, f.intRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:724
	})); if __e != nil { return }
//line beam_fragments.gox:725
		__e = __c.Any(f.strLens.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:727
			f.strRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:729
				__e = __c.Set("id", "lens-str"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:729
				__e = __c.Any(fmt.Sprintf("%s:%d", v, f.strRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:730
	})); if __e != nil { return }
//line beam_fragments.gox:731
		__e = __c.Any(f.evenLens.Bind(func(v bool) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:733
			f.evenRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:735
				__e = __c.Set("id", "lens-even"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:735
				__e = __c.Any(fmt.Sprintf("%t:%d", v, f.evenRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:736
	})); if __e != nil { return }
//line beam_fragments.gox:737
		__e = __c.Any(f.label.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:739
			f.labelRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:741
				__e = __c.Set("id", "lens-label"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:741
				__e = __c.Any(fmt.Sprintf("%s:%d", v, f.labelRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:742
	})); if __e != nil { return }
//line beam_fragments.gox:743
		__e = __c.Any(f.parityLens.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:745
			f.parityRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:747
				__e = __c.Set("id", "lens-parity"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:747
				__e = __c.Any(fmt.Sprintf("%s:%d", v, f.parityRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:748
	})); if __e != nil { return }
//line beam_fragments.gox:749
		__e = __c.Any(test.Button("lens-source-update", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 2, Str: "b"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:753
		__e = __c.Any(test.Button("lens-int-update", func(ctx context.Context) bool {
		f.intLens.Update(ctx, 5)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:757
		__e = __c.Any(test.Button("lens-int-mutate", func(ctx context.Context) bool {
		f.intLens.Mutate(ctx, func(v int) int {
			return v + 1
		})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:763
		__e = __c.Any(test.Button("lens-str-update", func(ctx context.Context) bool {
		f.strLens.Update(ctx, "lens")
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:767
		__e = __c.Any(test.Button("lens-even-false", func(ctx context.Context) bool {
		f.evenLens.Update(ctx, false)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:771
		__e = __c.Any(test.Button("lens-even-true", func(ctx context.Context) bool {
		f.evenLens.Update(ctx, true)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:775
		__e = __c.Any(test.Button("lens-xupdate", func(ctx context.Context) bool {
		err, ok := <-f.intLens.Update(doors.DetachedContext(ctx), 9)
		if !ok {
			f.r.Update(ctx, 1, "x-closed")
			return false
		}
		if err != nil {
			f.r.Update(ctx, 1, "x-err:" + err.Error())
			return false
		}
		f.r.Update(ctx, 1, "x-ok")
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:788
		__e = __c.Any(test.Button("lens-report", func(ctx context.Context) bool {
		f.reportState(ctx, "report")
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:792
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:793
}

type BeamRouteStateFragment struct {
	source doors.Source[state]
	sourceRenders int
	lensRouteRenders int
	lensListRenders int
	defaultLensRenders int
	beamRouteRenders int
	beamListRenders int
	defaultBeamRenders int
	test.NoBeam
}

func (f *BeamRouteStateFragment) init() {
	if f.source != nil {
		return
	}
	f.source = doors.NewSource(state{
		Int: 1,
	})
}

func (f *BeamRouteStateFragment) matchString(s state) (string, bool) {
	return s.Str, s.Str != ""
}

func (f *BeamRouteStateFragment) setString(s state, v string) state {
	s.Str = v
	return s
}

func (f *BeamRouteStateFragment) matchList(s state) ([]string, bool) {
	if s.Str == "" {
		return nil, false
	}
	return []string{s.Str, fmt.Sprint(s.Int)}, true
}

func (f *BeamRouteStateFragment) setList(s state, v []string) state {
	if len(v) == 0 {
		s.Str = ""
		return s
	}
	s.Str = v[0]
	return s
}

func (f *BeamRouteStateFragment) equalList(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

//line beam_fragments.gox:853
func (f *BeamRouteStateFragment) lensRoute(l doors.Source[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:855
		f.lensRouteRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:857
			__e = __c.Set("id", "route-lens-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:857
			__e = __c.Any(fmt.Sprint(f.lensRouteRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:858
		__e = __c.Any(l.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:859
				__e = __c.Set("id", "route-lens-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:859
				__e = __c.Any("lens:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:860
	})); if __e != nil { return }
//line beam_fragments.gox:861
		__e = __c.Any(test.Button("route-lens-update", func(ctx context.Context) bool {
		l.Update(ctx, "child")
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:865
		__e = __c.Any(test.Button("route-lens-clear", func(ctx context.Context) bool {
		l.Update(ctx, "")
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:869
}

//line beam_fragments.gox:871
func (f *BeamRouteStateFragment) lensListRoute(l doors.Source[[]string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:873
		f.lensListRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:875
			__e = __c.Set("id", "route-lens-list-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:875
			__e = __c.Any(fmt.Sprint(f.lensListRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:876
		__e = __c.Any(l.Bind(func(v []string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:877
				__e = __c.Set("id", "route-lens-list-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:877
				__e = __c.Any(strings.Join(v, ",")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:878
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:879
}

//line beam_fragments.gox:881
func (f *BeamRouteStateFragment) lensDefault(b doors.Beam[state]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:883
		f.defaultLensRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:885
			__e = __c.Set("id", "route-lens-default-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:885
			__e = __c.Any(fmt.Sprint(f.defaultLensRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:886
		__e = __c.Any(b.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:887
				__e = __c.Set("id", "route-lens-default-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:887
				__e = __c.Any(fmt.Sprintf("default:%d:%s", v.Int, v.Str)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:888
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:889
}

//line beam_fragments.gox:891
func (f *BeamRouteStateFragment) beamRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:893
		f.beamRouteRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:895
			__e = __c.Set("id", "route-beam-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:895
			__e = __c.Any(fmt.Sprint(f.beamRouteRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:896
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:897
				__e = __c.Set("id", "route-beam-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:897
				__e = __c.Any("beam:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:898
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:899
}

//line beam_fragments.gox:901
func (f *BeamRouteStateFragment) beamListRoute(b doors.Beam[[]string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:903
		f.beamListRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:905
			__e = __c.Set("id", "route-beam-list-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:905
			__e = __c.Any(fmt.Sprint(f.beamListRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:906
		__e = __c.Any(b.Bind(func(v []string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:907
				__e = __c.Set("id", "route-beam-list-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:907
				__e = __c.Any(strings.Join(v, ",")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:908
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:909
}

//line beam_fragments.gox:911
func (f *BeamRouteStateFragment) beamDefault(b doors.Beam[state]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:913
		f.defaultBeamRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:915
			__e = __c.Set("id", "route-beam-default-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:915
			__e = __c.Any(fmt.Sprint(f.defaultBeamRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:916
		__e = __c.Any(b.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:917
				__e = __c.Set("id", "route-beam-default-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:917
				__e = __c.Any(fmt.Sprintf("beam-default:%d:%s", v.Int, v.Str)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:918
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:919
}

//line beam_fragments.gox:921
func (f *BeamRouteStateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:923
		f.init()

//line beam_fragments.gox:925
		__e = __c.Any(f.source.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:927
			f.sourceRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:929
				__e = __c.Set("id", "route-source"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:929
				__e = __c.Any(fmt.Sprintf("%d:%s:%d", v.Int, v.Str, f.sourceRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:930
	})); if __e != nil { return }
//line beam_fragments.gox:931
		__e = __c.Any(f.source.Route(
		doors.RouteDerive(f.matchString).Source(f.setString, f.lensRoute),
		doors.RouteDefaultBeam(f.lensDefault),
	)); if __e != nil { return }
//line beam_fragments.gox:935
		__e = __c.Any(f.source.RouteBeam(
		doors.RouteDerive(f.matchString).Beam(f.beamRoute),
		doors.RouteDefaultBeam(f.beamDefault),
	)); if __e != nil { return }
//line beam_fragments.gox:939
		__e = __c.Any(f.source.Route(
		doors.RouteDeriveEqual(f.matchList, f.equalList).Source(f.setList, f.lensListRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:942
		__e = __c.Any(f.source.RouteBeam(
		doors.RouteDeriveEqual(f.matchList, f.equalList).Beam(f.beamListRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:945
		__e = __c.Any(test.Button("route-source-lens", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 2, Str: "doc"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:949
		__e = __c.Any(test.Button("route-source-next", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 2, Str: "next"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:953
		__e = __c.Any(test.Button("route-source-default-int", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 4})
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:957
}

type BeamRouteNoDefaultBurstFragment struct {
	source doors.Source[state]
	test.NoBeam
}

func (f *BeamRouteNoDefaultBurstFragment) init() {
	if f.source != nil {
		return
	}
	f.source = doors.NewSource(state{})
}

func (f *BeamRouteNoDefaultBurstFragment) matchString(s state) (string, bool) {
	return s.Str, s.Str != ""
}

func (f *BeamRouteNoDefaultBurstFragment) setString(s state, v string) state {
	s.Str = v
	return s
}

//line beam_fragments.gox:980
func (f *BeamRouteNoDefaultBurstFragment) lensRoute(l doors.Source[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:981
		__e = __c.Any(l.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:982
				__e = __c.Set("id", "route-burst-lens"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:982
				__e = __c.Any("lens:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:983
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:984
}

//line beam_fragments.gox:986
func (f *BeamRouteNoDefaultBurstFragment) beamRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:987
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:988
				__e = __c.Set("id", "route-burst-beam"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:988
				__e = __c.Any("beam:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:989
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:990
}

//line beam_fragments.gox:992
func (f *BeamRouteNoDefaultBurstFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:994
		f.init()

//line beam_fragments.gox:996
		__e = __c.Any(f.source.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:997
				__e = __c.Set("id", "route-burst-source"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:997
				__e = __c.Any(fmt.Sprintf("%d:%s", v.Int, v.Str)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:998
	})); if __e != nil { return }
//line beam_fragments.gox:999
		__e = __c.Any(f.source.Route(
		doors.RouteDerive(f.matchString).Source(f.setString, f.lensRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1002
		__e = __c.Any(f.source.RouteBeam(
		doors.RouteDerive(f.matchString).Beam(f.beamRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1005
		__e = __c.Any(test.Button("route-burst-none", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 1, Str: "queued"})
		f.source.Update(ctx, state{Int: 2})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:1010
		__e = __c.Any(test.Button("route-burst-hit", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 3, Str: "after"})
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1014
}

type BeamRouteEntrypointsFragment struct {
	source doors.Source[state]
	str doors.Source[string]
	parity doors.Beam[string]
	test.NoBeam
}

func (f *BeamRouteEntrypointsFragment) init() {
	if f.source != nil {
		return
	}
	f.source = doors.NewSource(state{
		Int: 1,
		Str: "a",
	})
	f.str = doors.DeriveSource(f.source, func(s state) string {
		return s.Str
	}, func(s state, v string) state {
		s.Str = v
		return s
	})
	f.parity = doors.DeriveBeam(f.source, func(s state) string {
		if s.Int % 2 == 0 {
			return "even"
		}
		return "odd"
	})
}

func (f *BeamRouteEntrypointsFragment) matchNonEmpty(v string) (string, bool) {
	return v, v != ""
}

func (f *BeamRouteEntrypointsFragment) matchOdd(v string) (string, bool) {
	return v, v == "odd"
}

func (f *BeamRouteEntrypointsFragment) setString(_ string, v string) string {
	return v
}

//line beam_fragments.gox:1057
func (f *BeamRouteEntrypointsFragment) lensRoute(l doors.Source[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1058
		__e = __c.Any(l.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1059
				__e = __c.Set("id", "route-entry-lens"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1059
				__e = __c.Any("lens:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1060
	})); if __e != nil { return }
//line beam_fragments.gox:1061
		__e = __c.Any(test.Button("route-entry-lens-update", func(ctx context.Context) bool {
		l.Update(ctx, "b")
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1065
}

//line beam_fragments.gox:1067
func (f *BeamRouteEntrypointsFragment) lensBeamRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1068
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1069
				__e = __c.Set("id", "route-entry-lens-beam"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1069
				__e = __c.Any("beam:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1070
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1071
}

//line beam_fragments.gox:1073
func (f *BeamRouteEntrypointsFragment) derivedRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1074
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1075
				__e = __c.Set("id", "route-entry-derived"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1075
				__e = __c.Any("derived:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1076
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1077
}

//line beam_fragments.gox:1079
func (f *BeamRouteEntrypointsFragment) derivedDefault(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1080
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1081
				__e = __c.Set("id", "route-entry-derived-default"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1081
				__e = __c.Any("derived-default:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1082
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1083
}

//line beam_fragments.gox:1085
func (f *BeamRouteEntrypointsFragment) simpleLensRoute(l doors.Source[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1086
		__e = __c.Any(l.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1087
				__e = __c.Set("id", "route-entry-simple-lens"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1087
				__e = __c.Any("simple-lens:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1088
	})); if __e != nil { return }
//line beam_fragments.gox:1089
		__e = __c.Any(test.Button("route-entry-simple-lens-update", func(ctx context.Context) bool {
		l.Update(ctx, "simple-lens")
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1093
}

//line beam_fragments.gox:1095
func (f *BeamRouteEntrypointsFragment) simpleBeamRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1096
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1097
				__e = __c.Set("id", "route-entry-simple-beam"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1097
				__e = __c.Any("simple-beam:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1098
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1099
}

//line beam_fragments.gox:1101
func (f *BeamRouteEntrypointsFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1103
		f.init()

//line beam_fragments.gox:1105
		__e = __c.Any(f.source.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1106
				__e = __c.Set("id", "route-entry-source"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1106
				__e = __c.Any(fmt.Sprintf("%d:%s", v.Int, v.Str)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1107
	})); if __e != nil { return }
//line beam_fragments.gox:1108
		__e = __c.Any(f.str.Route(
		doors.RouteDerive(f.matchNonEmpty).Source(f.setString, f.lensRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1111
		__e = __c.Any(f.str.RouteBeam(
		doors.RouteDerive(f.matchNonEmpty).Beam(f.lensBeamRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1114
		__e = __c.Any(f.parity.RouteBeam(
		doors.RouteDerive(f.matchOdd).Beam(f.derivedRoute),
		doors.RouteDefaultBeam(f.derivedDefault),
	)); if __e != nil { return }
//line beam_fragments.gox:1118
		__e = __c.Any(f.str.Route(
		doors.RouteMatch(func(v string) bool {
			return strings.HasPrefix(v, "simple")
		}).Source(f.simpleLensRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1123
		__e = __c.Any(f.str.RouteBeam(
		doors.RouteMatch(func(v string) bool {
			return strings.HasPrefix(v, "simple")
		}).Beam(f.simpleBeamRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1128
		__e = __c.Any(test.Button("route-entry-set-simple", func(ctx context.Context) bool {
		f.str.Update(ctx, "simple")
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:1132
		__e = __c.Any(test.Button("route-entry-even", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 2, Str: f.str.Get()})
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1136
}
