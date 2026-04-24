// Managed by GoX v0.1.28

//line beam_fragments.gox:1
package beam

import (
	"context"
	"fmt"
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

//line beam_fragments.gox:25
func (f *BeamSkipFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:27
		f.r.Update(ctx, 0, "init")
		f.b.ReadAndSub(ctx, func(ctx context.Context, s state) bool {
			<-time.After(300 * time.Millisecond)
			return false
		})

//line beam_fragments.gox:33
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:35
				f.b.Sub(ctx, func(ctx context.Context, s state) bool {
				if s.Str == "1" {
					f.r.Update(ctx, 0, "propagated")
				}
				return false
			})

			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:43
		__e = __c.Any(test.Button("update1", func(ctx context.Context) bool {
		f.b.Update(ctx, state{Str: "1"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:47
		__e = __c.Any(test.Button("update2", func(ctx context.Context) bool {
		f.b.Update(ctx, state{Str: "2"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:51
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:52
}

type BeamDeriveFragment struct {
	r *test.Reporter
	b doors.Source[state]
	n doors.Door
	test.NoBeam
}
//line beam_fragments.gox:60
func (f *BeamDeriveFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:61
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:62
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:64
		__e = __c.Any(test.Button("reload", func(ctx context.Context) bool {
		f.n.Inner(ctx, f.content())
		return true
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:68
}

//line beam_fragments.gox:70
func (f *BeamDeriveFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:72
		d := doors.NewBeam(f.b, func(s state) int {
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

//line beam_fragments.gox:87
		__e = __c.Any(test.ReportId(1, fmt.Sprint(r))); if __e != nil { return }
//line beam_fragments.gox:88
		__e = (n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:90
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := d.Read(ctx)

//line beam_fragments.gox:96
				__e = __c.Any(test.ReportId(2, fmt.Sprint(r))); if __e != nil { return }
//line beam_fragments.gox:98
				n3 := doors.Door{}
			d.Sub(ctx, func(ctx context.Context, s int) bool {
				n3.Inner(ctx, test.ReportId(4, fmt.Sprint(s)))
				return false
			})

//line beam_fragments.gox:104
				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:106
		__e = (n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:108
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := f.b.Read(ctx)

//line beam_fragments.gox:114
				__e = __c.Any(test.ReportId(3, fmt.Sprint(r.Int))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:116
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:117
}

type BeamConsistentFragment struct {
	r *test.Reporter
	b doors.Source[state]
	n doors.Door
	test.NoBeam
}

//line beam_fragments.gox:126
func (f *BeamConsistentFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:128
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

//line beam_fragments.gox:140
		__e = __c.Any(test.ReportId(1, fmt.Sprint(r.Int))); if __e != nil { return }
//line beam_fragments.gox:141
		__e = (n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:143
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := f.b.Read(ctx)

//line beam_fragments.gox:149
				__e = __c.Any(test.ReportId(2, fmt.Sprint(r.Int))); if __e != nil { return }
//line beam_fragments.gox:151
				n3 := doors.Door{}
			f.b.Sub(ctx, func(ctx context.Context, s state) bool {
				n3.Inner(ctx, test.ReportId(4, fmt.Sprint(s.Int)))
				return false
			})

//line beam_fragments.gox:157
				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:159
		__e = (n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:161
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := f.b.Read(ctx)

//line beam_fragments.gox:167
				__e = __c.Any(test.ReportId(3, fmt.Sprint(r.Int))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:169
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:170
}

//line beam_fragments.gox:172
func (f *BeamConsistentFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:173
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:174
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:176
		__e = __c.Any(test.Button("reload", func(ctx context.Context) bool {
		f.n.Reload(ctx)
		return true
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:180
}

type BeamUpdateFragment struct {
	r *test.Reporter
	b doors.Source[state]
	test.NoBeam
}

//line beam_fragments.gox:188
func (f *BeamUpdateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:190
		f.b.Sub(ctx, func(ctx context.Context, s state) bool {
			f.r.Update(ctx, 0, fmt.Sprint(s.Int))
			return false
		})

//line beam_fragments.gox:196
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
//line beam_fragments.gox:217
}

type BeamEqualFragment struct {
	r *test.Reporter
	b doors.Source[state]
	p doors.Beam[string]
	test.NoBeam
}

//line beam_fragments.gox:226
func (f *BeamEqualFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:228
		if f.p == nil {
			f.p = doors.NewBeamEqual(f.b, func(s state) string {
				if s.Int % 2 == 0 {
					return "even"
				}
				return "odd"
			}, func(new string, old string) bool {
				return new == old
			})
		}
		f.b.Sub(ctx, func(ctx context.Context, s state) bool {
			f.r.Update(ctx, 0, fmt.Sprint(s.Int))
			return false
		})

//line beam_fragments.gox:243
		__e = __c.Any(f.p.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:244
				__e = __c.Set("id", "parity"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:244
				__e = __c.Any(v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:245
	})); if __e != nil { return }
//line beam_fragments.gox:246
		__e = __c.Any(doors.Go(func(ctx context.Context) {
		<-time.After(100 * time.Millisecond)
		f.r.Update(ctx, 2, "go")
	})); if __e != nil { return }
//line beam_fragments.gox:251
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
//line beam_fragments.gox:276
}

type BeamRenderBranchUpdateFrameFragment struct {
	b doors.Source[int]
	n doors.Door
	test.NoBeam
}

//line beam_fragments.gox:284
func (f *BeamRenderBranchUpdateFrameFragment) content(i int) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:285
			__e = __c.Set("id", "watcher-i"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:285
			__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:287
		f.b.Mutate(ctx, func(i int) int {
			return i + 1
		})
		newI, _ := f.b.Read(ctx)

		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:292
			__e = __c.Set("id", "watcher-newi"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:292
			__e = __c.Any(fmt.Sprint(newI)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:293
}

//line beam_fragments.gox:295
func (f *BeamRenderBranchUpdateFrameFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:296
		__e = __c.Any(&f.n); if __e != nil { return }
//line beam_fragments.gox:298
		f.b.ReadAndSub(ctx, func(ctx context.Context, i int) bool {
			f.n.Inner(ctx, f.content(i))
			return true
		})
		f.b.Mutate(ctx, func(i int) int {
			return i + 1
		})

	return })
//line beam_fragments.gox:306
}

type BeamRenderBranchInitFrameFragment struct {
	b doors.Source[int]
	n doors.Door
	test.NoBeam
}

//line beam_fragments.gox:314
func (f *BeamRenderBranchInitFrameFragment) content(i int) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:315
			__e = __c.Set("id", "watcher-i"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:315
			__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:317
		f.b.Mutate(ctx, func(i int) int {
			return i + 1
		})
		newI, _ := f.b.Read(ctx)

		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:322
			__e = __c.Set("id", "watcher-newi"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:322
			__e = __c.Any(fmt.Sprint(newI)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:323
}

//line beam_fragments.gox:325
func (f *BeamRenderBranchInitFrameFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:326
		__e = __c.Any(&f.n); if __e != nil { return }
//line beam_fragments.gox:328
		go func() {
			f.b.Sub(ctx, func(ctx context.Context, i int) bool {
				f.n.Inner(ctx, f.content(i))
				return true
			})
		}()

	return })
//line beam_fragments.gox:335
}

type BeamRenderUpdateWarningFragment struct {
	b doors.Source[int]
	host doors.Door
	test.NoBeam
}

//line beam_fragments.gox:343
func (f *BeamRenderUpdateWarningFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:345
		n3 := doors.Door{}
		_, _ = f.b.Read(ctx)
		f.b.Sub(ctx, func(ctx context.Context, i int) bool {
			n3.Inner(ctx, test.ReportId(4, fmt.Sprint(i)))
			return false
		})

//line beam_fragments.gox:352
		__e = __c.Any(&n3); if __e != nil { return }
	return })
//line beam_fragments.gox:353
}

//line beam_fragments.gox:355
func (f *BeamRenderUpdateWarningFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:356
		__e = (f.host).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:357
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:359
		__e = __c.Any(test.Button("warning-reload", func(ctx context.Context) bool {
		f.host.Reload(ctx)
		return true
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:363
}

type BeamEffectSourceFragment struct {
	b doors.Source[int]
	frame doors.Door
	host doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line beam_fragments.gox:374
func (f *BeamEffectSourceFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:376
		f.innerRenders++
		value, _ := f.b.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:379
			__e = __c.Set("id", "effect-source-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:379
			__e = __c.Any(fmt.Sprint(value)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:380
			__e = __c.Set("id", "effect-source-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:380
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:381
}

//line beam_fragments.gox:383
func (f *BeamEffectSourceFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:385
		f.outerRenders++
		f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:388
			__e = __c.Set("id", "effect-source-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:388
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:389
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:390
}

//line beam_fragments.gox:392
func (f *BeamEffectSourceFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:394
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:396
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:397
		__e = __c.Any(test.Button("effect-source-update-1", func(ctx context.Context) bool {
		f.b.Update(ctx, 1)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:401
		__e = __c.Any(test.Button("effect-source-update-2", func(ctx context.Context) bool {
		f.b.Update(ctx, 2)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:405
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

//line beam_fragments.gox:417
func (f *BeamEffectDerivedFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:419
		f.innerRenders++
		value, _ := f.d.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:422
			__e = __c.Set("id", "effect-derived-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:422
			__e = __c.Any(value); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:423
			__e = __c.Set("id", "effect-derived-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:423
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:424
}

//line beam_fragments.gox:426
func (f *BeamEffectDerivedFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:428
		f.outerRenders++
		f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:431
			__e = __c.Set("id", "effect-derived-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:431
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:432
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:433
}

//line beam_fragments.gox:435
func (f *BeamEffectDerivedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:437
		if f.d == nil {
			f.d = doors.NewBeam(f.b, func(v int) string {
				return fmt.Sprintf("v:%d", v)
			})
		}
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:444
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:445
		__e = __c.Any(test.Button("effect-derived-update-1", func(ctx context.Context) bool {
		f.b.Update(ctx, 1)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:449
		__e = __c.Any(test.Button("effect-derived-update-2", func(ctx context.Context) bool {
		f.b.Update(ctx, 2)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:453
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

//line beam_fragments.gox:465
func (f *BeamEffectMultiFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:467
		f.innerRenders++
		left, _ := f.left.Effect(ctx)
		right, _ := f.right.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:471
			__e = __c.Set("id", "effect-multi-left"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:471
			__e = __c.Any(fmt.Sprint(left)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:472
			__e = __c.Set("id", "effect-multi-right"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:472
			__e = __c.Any(fmt.Sprint(right)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:473
			__e = __c.Set("id", "effect-multi-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:473
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:474
}

//line beam_fragments.gox:476
func (f *BeamEffectMultiFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:478
		f.outerRenders++
		f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:481
			__e = __c.Set("id", "effect-multi-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:481
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:482
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:483
}

//line beam_fragments.gox:485
func (f *BeamEffectMultiFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:487
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:489
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:490
		__e = __c.Any(test.Button("effect-multi-left-update", func(ctx context.Context) bool {
		f.left.Update(ctx, 1)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:494
		__e = __c.Any(test.Button("effect-multi-right-update", func(ctx context.Context) bool {
		f.right.Update(ctx, 1)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:498
}

type BeamEffectDuplicateFragment struct {
	b doors.Source[int]
	frame doors.Door
	host doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line beam_fragments.gox:509
func (f *BeamEffectDuplicateFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:511
		f.innerRenders++
		first, _ := f.b.Effect(ctx)
		second, _ := f.b.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:515
			__e = __c.Set("id", "effect-dup-first"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:515
			__e = __c.Any(fmt.Sprint(first)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:516
			__e = __c.Set("id", "effect-dup-second"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:516
			__e = __c.Any(fmt.Sprint(second)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:517
			__e = __c.Set("id", "effect-dup-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:517
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:518
}

//line beam_fragments.gox:520
func (f *BeamEffectDuplicateFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:522
		f.outerRenders++
		f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:525
			__e = __c.Set("id", "effect-dup-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:525
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:526
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:527
}

//line beam_fragments.gox:529
func (f *BeamEffectDuplicateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:531
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:533
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:534
		__e = __c.Any(test.Button("effect-dup-update", func(ctx context.Context) bool {
		f.b.Update(ctx, 1)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:538
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

//line beam_fragments.gox:550
func (f *BeamReadAndSubFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:552
		if f.derived == nil {
			f.derived = doors.NewBeam(f.source, func(v int) string {
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

//line beam_fragments.gox:568
		__e = __c.Any(test.Button("beam-read-sub-update-2", func(ctx context.Context) bool {
		f.source.Update(ctx, 2)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:572
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
//line beam_fragments.gox:586
		__e = __c.Any(test.Button("beam-read-sub-update-3", func(ctx context.Context) bool {
		f.source.Update(ctx, 3)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:590
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
//line beam_fragments.gox:604
		__e = __c.Any(test.Button("beam-read-sub-update-4", func(ctx context.Context) bool {
		f.source.Update(ctx, 4)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:608
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:609
}
