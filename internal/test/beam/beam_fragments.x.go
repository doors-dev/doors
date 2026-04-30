// Managed by GoX v0.1.28

//line beam_fragments.gox:1
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

//line beam_fragments.gox:26
func (f *BeamSkipFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:28
		f.r.Update(ctx, 0, "init")
		f.b.ReadAndSub(ctx, func(ctx context.Context, s state) bool {
			<-time.After(300 * time.Millisecond)
			return false
		})

//line beam_fragments.gox:34
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:36
				f.b.Sub(ctx, func(ctx context.Context, s state) bool {
				if s.Str == "1" {
					f.r.Update(ctx, 0, "propagated")
				}
				return false
			})

			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:44
		__e = __c.Any(test.Button("update1", func(ctx context.Context) bool {
		f.b.Update(ctx, state{Str: "1"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:48
		__e = __c.Any(test.Button("update2", func(ctx context.Context) bool {
		f.b.Update(ctx, state{Str: "2"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:52
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:53
}

type BeamDeriveFragment struct {
	r *test.Reporter
	b doors.Source[state]
	n doors.Door
	test.NoBeam
}
//line beam_fragments.gox:61
func (f *BeamDeriveFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:62
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:63
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:65
		__e = __c.Any(test.Button("reload", func(ctx context.Context) bool {
		f.n.Inner(ctx, f.content())
		return true
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:69
}

//line beam_fragments.gox:71
func (f *BeamDeriveFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:73
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

//line beam_fragments.gox:88
		__e = __c.Any(test.ReportId(1, fmt.Sprint(r))); if __e != nil { return }
//line beam_fragments.gox:89
		__e = (n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:91
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := d.Read(ctx)

//line beam_fragments.gox:97
				__e = __c.Any(test.ReportId(2, fmt.Sprint(r))); if __e != nil { return }
//line beam_fragments.gox:99
				n3 := doors.Door{}
			d.Sub(ctx, func(ctx context.Context, s int) bool {
				n3.Inner(ctx, test.ReportId(4, fmt.Sprint(s)))
				return false
			})

//line beam_fragments.gox:105
				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:107
		__e = (n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:109
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := f.b.Read(ctx)

//line beam_fragments.gox:115
				__e = __c.Any(test.ReportId(3, fmt.Sprint(r.Int))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:117
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:118
}

type BeamConsistentFragment struct {
	r *test.Reporter
	b doors.Source[state]
	n doors.Door
	test.NoBeam
}

//line beam_fragments.gox:127
func (f *BeamConsistentFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:129
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

//line beam_fragments.gox:141
		__e = __c.Any(test.ReportId(1, fmt.Sprint(r.Int))); if __e != nil { return }
//line beam_fragments.gox:142
		__e = (n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:144
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := f.b.Read(ctx)

//line beam_fragments.gox:150
				__e = __c.Any(test.ReportId(2, fmt.Sprint(r.Int))); if __e != nil { return }
//line beam_fragments.gox:152
				n3 := doors.Door{}
			f.b.Sub(ctx, func(ctx context.Context, s state) bool {
				n3.Inner(ctx, test.ReportId(4, fmt.Sprint(s.Int)))
				return false
			})

//line beam_fragments.gox:158
				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:160
		__e = (n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:162
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := f.b.Read(ctx)

//line beam_fragments.gox:168
				__e = __c.Any(test.ReportId(3, fmt.Sprint(r.Int))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:170
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:171
}

//line beam_fragments.gox:173
func (f *BeamConsistentFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:174
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:175
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:177
		__e = __c.Any(test.Button("reload", func(ctx context.Context) bool {
		f.n.Reload(ctx)
		return true
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:181
}

type BeamUpdateFragment struct {
	r *test.Reporter
	b doors.Source[state]
	test.NoBeam
}

//line beam_fragments.gox:189
func (f *BeamUpdateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:191
		f.b.Sub(ctx, func(ctx context.Context, s state) bool {
			f.r.Update(ctx, 0, fmt.Sprint(s.Int))
			return false
		})

//line beam_fragments.gox:197
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
//line beam_fragments.gox:218
}

type BeamEqualFragment struct {
	r *test.Reporter
	b doors.Source[state]
	p doors.Beam[string]
	test.NoBeam
}

//line beam_fragments.gox:227
func (f *BeamEqualFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:229
		if f.p == nil {
			f.p = doors.DeriveBeamEqual(f.b, func(s state) string {
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

//line beam_fragments.gox:244
		__e = __c.Any(f.p.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:245
				__e = __c.Set("id", "parity"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:245
				__e = __c.Any(v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:246
	})); if __e != nil { return }
//line beam_fragments.gox:247
		__e = __c.Any(doors.Go(func(ctx context.Context) {
		<-time.After(100 * time.Millisecond)
		f.r.Update(ctx, 2, "go")
	})); if __e != nil { return }
//line beam_fragments.gox:252
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
//line beam_fragments.gox:277
}

type BeamRenderBranchUpdateFrameFragment struct {
	b doors.Source[int]
	n doors.Door
	test.NoBeam
}

//line beam_fragments.gox:285
func (f *BeamRenderBranchUpdateFrameFragment) content(i int) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:286
			__e = __c.Set("id", "watcher-i"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:286
			__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:288
		f.b.Mutate(ctx, func(i int) int {
			return i + 1
		})
		newI, _ := f.b.Read(ctx)

		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:293
			__e = __c.Set("id", "watcher-newi"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:293
			__e = __c.Any(fmt.Sprint(newI)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:294
}

//line beam_fragments.gox:296
func (f *BeamRenderBranchUpdateFrameFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:297
		__e = __c.Any(&f.n); if __e != nil { return }
//line beam_fragments.gox:299
		f.b.ReadAndSub(ctx, func(ctx context.Context, i int) bool {
			f.n.Inner(ctx, f.content(i))
			return true
		})
		f.b.Mutate(ctx, func(i int) int {
			return i + 1
		})

	return })
//line beam_fragments.gox:307
}

type BeamRenderBranchInitFrameFragment struct {
	b doors.Source[int]
	n doors.Door
	test.NoBeam
}

//line beam_fragments.gox:315
func (f *BeamRenderBranchInitFrameFragment) content(i int) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:316
			__e = __c.Set("id", "watcher-i"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:316
			__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:318
		f.b.Mutate(ctx, func(i int) int {
			return i + 1
		})
		newI, _ := f.b.Read(ctx)

		__e = __c.Init("span"); if __e != nil { return }
		{
//line beam_fragments.gox:323
			__e = __c.Set("id", "watcher-newi"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:323
			__e = __c.Any(fmt.Sprint(newI)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:324
}

//line beam_fragments.gox:326
func (f *BeamRenderBranchInitFrameFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:327
		__e = __c.Any(&f.n); if __e != nil { return }
//line beam_fragments.gox:329
		go func() {
			f.b.Sub(ctx, func(ctx context.Context, i int) bool {
				f.n.Inner(ctx, f.content(i))
				return true
			})
		}()

	return })
//line beam_fragments.gox:336
}

type BeamRenderUpdateWarningFragment struct {
	b doors.Source[int]
	host doors.Door
	test.NoBeam
}

//line beam_fragments.gox:344
func (f *BeamRenderUpdateWarningFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:346
		n3 := doors.Door{}
		_, _ = f.b.Read(ctx)
		f.b.Sub(ctx, func(ctx context.Context, i int) bool {
			n3.Inner(ctx, test.ReportId(4, fmt.Sprint(i)))
			return false
		})

//line beam_fragments.gox:353
		__e = __c.Any(&n3); if __e != nil { return }
	return })
//line beam_fragments.gox:354
}

//line beam_fragments.gox:356
func (f *BeamRenderUpdateWarningFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:357
		__e = (f.host).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line beam_fragments.gox:358
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:360
		__e = __c.Any(test.Button("warning-reload", func(ctx context.Context) bool {
		f.host.Reload(ctx)
		return true
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:364
}

type BeamEffectSourceFragment struct {
	b doors.Source[int]
	frame doors.Door
	host doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line beam_fragments.gox:375
func (f *BeamEffectSourceFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:377
		f.innerRenders++
		value, _ := f.b.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:380
			__e = __c.Set("id", "effect-source-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:380
			__e = __c.Any(fmt.Sprint(value)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:381
			__e = __c.Set("id", "effect-source-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:381
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:382
}

//line beam_fragments.gox:384
func (f *BeamEffectSourceFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:386
		f.outerRenders++
		f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:389
			__e = __c.Set("id", "effect-source-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:389
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:390
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:391
}

//line beam_fragments.gox:393
func (f *BeamEffectSourceFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:395
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:397
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:398
		__e = __c.Any(test.Button("effect-source-update-1", func(ctx context.Context) bool {
		f.b.Update(ctx, 1)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:402
		__e = __c.Any(test.Button("effect-source-update-2", func(ctx context.Context) bool {
		f.b.Update(ctx, 2)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:406
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

//line beam_fragments.gox:418
func (f *BeamEffectDerivedFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:420
		f.innerRenders++
		value, _ := f.d.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:423
			__e = __c.Set("id", "effect-derived-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:423
			__e = __c.Any(value); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:424
			__e = __c.Set("id", "effect-derived-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:424
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:425
}

//line beam_fragments.gox:427
func (f *BeamEffectDerivedFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:429
		f.outerRenders++
		f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:432
			__e = __c.Set("id", "effect-derived-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:432
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:433
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:434
}

//line beam_fragments.gox:436
func (f *BeamEffectDerivedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:438
		if f.d == nil {
			f.d = doors.DeriveBeam(f.b, func(v int) string {
				return fmt.Sprintf("v:%d", v)
			})
		}
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:445
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:446
		__e = __c.Any(test.Button("effect-derived-update-1", func(ctx context.Context) bool {
		f.b.Update(ctx, 1)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:450
		__e = __c.Any(test.Button("effect-derived-update-2", func(ctx context.Context) bool {
		f.b.Update(ctx, 2)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:454
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

//line beam_fragments.gox:466
func (f *BeamEffectMultiFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:468
		f.innerRenders++
		left, _ := f.left.Effect(ctx)
		right, _ := f.right.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:472
			__e = __c.Set("id", "effect-multi-left"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:472
			__e = __c.Any(fmt.Sprint(left)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:473
			__e = __c.Set("id", "effect-multi-right"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:473
			__e = __c.Any(fmt.Sprint(right)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:474
			__e = __c.Set("id", "effect-multi-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:474
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:475
}

//line beam_fragments.gox:477
func (f *BeamEffectMultiFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:479
		f.outerRenders++
		f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:482
			__e = __c.Set("id", "effect-multi-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:482
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:483
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:484
}

//line beam_fragments.gox:486
func (f *BeamEffectMultiFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:488
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:490
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:491
		__e = __c.Any(test.Button("effect-multi-left-update", func(ctx context.Context) bool {
		f.left.Update(ctx, 1)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:495
		__e = __c.Any(test.Button("effect-multi-right-update", func(ctx context.Context) bool {
		f.right.Update(ctx, 1)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:499
}

type BeamEffectDuplicateFragment struct {
	b doors.Source[int]
	frame doors.Door
	host doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line beam_fragments.gox:510
func (f *BeamEffectDuplicateFragment) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:512
		f.innerRenders++
		first, _ := f.b.Effect(ctx)
		second, _ := f.b.Effect(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:516
			__e = __c.Set("id", "effect-dup-first"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:516
			__e = __c.Any(fmt.Sprint(first)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:517
			__e = __c.Set("id", "effect-dup-second"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:517
			__e = __c.Any(fmt.Sprint(second)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:518
			__e = __c.Set("id", "effect-dup-inner-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:518
			__e = __c.Any(fmt.Sprint(f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line beam_fragments.gox:519
}

//line beam_fragments.gox:521
func (f *BeamEffectDuplicateFragment) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:523
		f.outerRenders++
		f.host.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:526
			__e = __c.Set("id", "effect-dup-outer-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:526
			__e = __c.Any(fmt.Sprint(f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:527
		__e = __c.Any(&f.host); if __e != nil { return }
	return })
//line beam_fragments.gox:528
}

//line beam_fragments.gox:530
func (f *BeamEffectDuplicateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:532
		f.frame.Inner(ctx, f.outerContent())

//line beam_fragments.gox:534
		__e = __c.Any(&f.frame); if __e != nil { return }
//line beam_fragments.gox:535
		__e = __c.Any(test.Button("effect-dup-update", func(ctx context.Context) bool {
		f.b.Update(ctx, 1)
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:539
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

//line beam_fragments.gox:551
func (f *BeamReadAndSubFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:553
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

//line beam_fragments.gox:569
		__e = __c.Any(test.Button("beam-read-sub-update-2", func(ctx context.Context) bool {
		f.source.Update(ctx, 2)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:573
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
//line beam_fragments.gox:587
		__e = __c.Any(test.Button("beam-read-sub-update-3", func(ctx context.Context) bool {
		f.source.Update(ctx, 3)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:591
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
//line beam_fragments.gox:605
		__e = __c.Any(test.Button("beam-read-sub-update-4", func(ctx context.Context) bool {
		f.source.Update(ctx, 4)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:609
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:610
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
	}, func(new string, old string) bool {
		return new == old
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

//line beam_fragments.gox:695
func (f *BeamLensRoundTripFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:697
		f.init()

//line beam_fragments.gox:699
		__e = __c.Any(f.source.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:701
			f.sourceRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:703
				__e = __c.Set("id", "lens-source"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:703
				__e = __c.Any(fmt.Sprintf("%d:%s:%d", v.Int, v.Str, f.sourceRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:704
	})); if __e != nil { return }
//line beam_fragments.gox:705
		__e = __c.Any(f.intLens.Bind(func(v int) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:707
			f.intRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:709
				__e = __c.Set("id", "lens-int"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:709
				__e = __c.Any(fmt.Sprintf("%d:%d", v, f.intRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:710
	})); if __e != nil { return }
//line beam_fragments.gox:711
		__e = __c.Any(f.strLens.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:713
			f.strRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:715
				__e = __c.Set("id", "lens-str"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:715
				__e = __c.Any(fmt.Sprintf("%s:%d", v, f.strRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:716
	})); if __e != nil { return }
//line beam_fragments.gox:717
		__e = __c.Any(f.evenLens.Bind(func(v bool) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:719
			f.evenRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:721
				__e = __c.Set("id", "lens-even"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:721
				__e = __c.Any(fmt.Sprintf("%t:%d", v, f.evenRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:722
	})); if __e != nil { return }
//line beam_fragments.gox:723
		__e = __c.Any(f.label.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:725
			f.labelRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:727
				__e = __c.Set("id", "lens-label"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:727
				__e = __c.Any(fmt.Sprintf("%s:%d", v, f.labelRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:728
	})); if __e != nil { return }
//line beam_fragments.gox:729
		__e = __c.Any(f.parityLens.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:731
			f.parityRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:733
				__e = __c.Set("id", "lens-parity"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:733
				__e = __c.Any(fmt.Sprintf("%s:%d", v, f.parityRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:734
	})); if __e != nil { return }
//line beam_fragments.gox:735
		__e = __c.Any(test.Button("lens-source-update", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 2, Str: "b"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:739
		__e = __c.Any(test.Button("lens-int-update", func(ctx context.Context) bool {
		f.intLens.Update(ctx, 5)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:743
		__e = __c.Any(test.Button("lens-int-mutate", func(ctx context.Context) bool {
		f.intLens.Mutate(ctx, func(v int) int {
			return v + 1
		})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:749
		__e = __c.Any(test.Button("lens-str-update", func(ctx context.Context) bool {
		f.strLens.Update(ctx, "lens")
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:753
		__e = __c.Any(test.Button("lens-even-false", func(ctx context.Context) bool {
		f.evenLens.Update(ctx, false)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:757
		__e = __c.Any(test.Button("lens-even-true", func(ctx context.Context) bool {
		f.evenLens.Update(ctx, true)
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:761
		__e = __c.Any(test.Button("lens-xupdate", func(ctx context.Context) bool {
		err, ok := <-f.intLens.XUpdate(doors.Free(ctx), 9)
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
//line beam_fragments.gox:774
		__e = __c.Any(test.Button("lens-report", func(ctx context.Context) bool {
		f.reportState(ctx, "report")
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:778
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line beam_fragments.gox:779
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

//line beam_fragments.gox:839
func (f *BeamRouteStateFragment) lensRoute(l doors.Source[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:841
		f.lensRouteRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:843
			__e = __c.Set("id", "route-lens-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:843
			__e = __c.Any(fmt.Sprint(f.lensRouteRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:844
		__e = __c.Any(l.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:845
				__e = __c.Set("id", "route-lens-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:845
				__e = __c.Any("lens:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:846
	})); if __e != nil { return }
//line beam_fragments.gox:847
		__e = __c.Any(test.Button("route-lens-update", func(ctx context.Context) bool {
		l.Update(ctx, "child")
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:851
		__e = __c.Any(test.Button("route-lens-clear", func(ctx context.Context) bool {
		l.Update(ctx, "")
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:855
}

//line beam_fragments.gox:857
func (f *BeamRouteStateFragment) lensListRoute(l doors.Source[[]string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:859
		f.lensListRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:861
			__e = __c.Set("id", "route-lens-list-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:861
			__e = __c.Any(fmt.Sprint(f.lensListRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:862
		__e = __c.Any(l.Bind(func(v []string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:863
				__e = __c.Set("id", "route-lens-list-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:863
				__e = __c.Any(strings.Join(v, ",")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:864
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:865
}

//line beam_fragments.gox:867
func (f *BeamRouteStateFragment) lensDefault(b doors.Beam[state]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:869
		f.defaultLensRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:871
			__e = __c.Set("id", "route-lens-default-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:871
			__e = __c.Any(fmt.Sprint(f.defaultLensRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:872
		__e = __c.Any(b.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:873
				__e = __c.Set("id", "route-lens-default-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:873
				__e = __c.Any(fmt.Sprintf("default:%d:%s", v.Int, v.Str)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:874
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:875
}

//line beam_fragments.gox:877
func (f *BeamRouteStateFragment) beamRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:879
		f.beamRouteRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:881
			__e = __c.Set("id", "route-beam-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:881
			__e = __c.Any(fmt.Sprint(f.beamRouteRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:882
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:883
				__e = __c.Set("id", "route-beam-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:883
				__e = __c.Any("beam:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:884
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:885
}

//line beam_fragments.gox:887
func (f *BeamRouteStateFragment) beamListRoute(b doors.Beam[[]string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:889
		f.beamListRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:891
			__e = __c.Set("id", "route-beam-list-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:891
			__e = __c.Any(fmt.Sprint(f.beamListRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:892
		__e = __c.Any(b.Bind(func(v []string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:893
				__e = __c.Set("id", "route-beam-list-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:893
				__e = __c.Any(strings.Join(v, ",")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:894
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:895
}

//line beam_fragments.gox:897
func (f *BeamRouteStateFragment) beamDefault(b doors.Beam[state]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:899
		f.defaultBeamRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line beam_fragments.gox:901
			__e = __c.Set("id", "route-beam-default-render"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:901
			__e = __c.Any(fmt.Sprint(f.defaultBeamRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line beam_fragments.gox:902
		__e = __c.Any(b.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:903
				__e = __c.Set("id", "route-beam-default-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:903
				__e = __c.Any(fmt.Sprintf("beam-default:%d:%s", v.Int, v.Str)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:904
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:905
}

//line beam_fragments.gox:907
func (f *BeamRouteStateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:909
		f.init()

//line beam_fragments.gox:911
		__e = __c.Any(f.source.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:913
			f.sourceRenders++

			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:915
				__e = __c.Set("id", "route-source"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:915
				__e = __c.Any(fmt.Sprintf("%d:%s:%d", v.Int, v.Str, f.sourceRenders)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:916
	})); if __e != nil { return }
//line beam_fragments.gox:917
		__e = __c.Any(f.source.RouteSource(
		doors.RouteDerive(f.matchString).Source(f.setString, f.lensRoute),
		doors.RouteDefaultBeam(f.lensDefault),
	)); if __e != nil { return }
//line beam_fragments.gox:921
		__e = __c.Any(f.source.RouteBeam(
		doors.RouteDerive(f.matchString).Beam(f.beamRoute),
		doors.RouteDefaultBeam(f.beamDefault),
	)); if __e != nil { return }
//line beam_fragments.gox:925
		__e = __c.Any(f.source.RouteSource(
		doors.RouteDeriveEqual(f.matchList, f.equalList).Source(f.setList, f.lensListRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:928
		__e = __c.Any(f.source.RouteBeam(
		doors.RouteDeriveEqual(f.matchList, f.equalList).Beam(f.beamListRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:931
		__e = __c.Any(test.Button("route-source-lens", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 2, Str: "doc"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:935
		__e = __c.Any(test.Button("route-source-next", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 2, Str: "next"})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:939
		__e = __c.Any(test.Button("route-source-default-int", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 4})
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:943
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

//line beam_fragments.gox:966
func (f *BeamRouteNoDefaultBurstFragment) lensRoute(l doors.Source[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:967
		__e = __c.Any(l.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:968
				__e = __c.Set("id", "route-burst-lens"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:968
				__e = __c.Any("lens:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:969
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:970
}

//line beam_fragments.gox:972
func (f *BeamRouteNoDefaultBurstFragment) beamRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:973
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:974
				__e = __c.Set("id", "route-burst-beam"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:974
				__e = __c.Any("beam:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:975
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:976
}

//line beam_fragments.gox:978
func (f *BeamRouteNoDefaultBurstFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:980
		f.init()

//line beam_fragments.gox:982
		__e = __c.Any(f.source.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:983
				__e = __c.Set("id", "route-burst-source"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:983
				__e = __c.Any(fmt.Sprintf("%d:%s", v.Int, v.Str)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:984
	})); if __e != nil { return }
//line beam_fragments.gox:985
		__e = __c.Any(f.source.RouteSource(
		doors.RouteDerive(f.matchString).Source(f.setString, f.lensRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:988
		__e = __c.Any(f.source.RouteBeam(
		doors.RouteDerive(f.matchString).Beam(f.beamRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:991
		__e = __c.Any(test.Button("route-burst-none", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 1, Str: "queued"})
		f.source.Update(ctx, state{Int: 2})
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:996
		__e = __c.Any(test.Button("route-burst-hit", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 3, Str: "after"})
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1000
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

//line beam_fragments.gox:1043
func (f *BeamRouteEntrypointsFragment) lensRoute(l doors.Source[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1044
		__e = __c.Any(l.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1045
				__e = __c.Set("id", "route-entry-lens"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1045
				__e = __c.Any("lens:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1046
	})); if __e != nil { return }
//line beam_fragments.gox:1047
		__e = __c.Any(test.Button("route-entry-lens-update", func(ctx context.Context) bool {
		l.Update(ctx, "b")
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1051
}

//line beam_fragments.gox:1053
func (f *BeamRouteEntrypointsFragment) lensBeamRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1054
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1055
				__e = __c.Set("id", "route-entry-lens-beam"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1055
				__e = __c.Any("beam:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1056
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1057
}

//line beam_fragments.gox:1059
func (f *BeamRouteEntrypointsFragment) derivedRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1060
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1061
				__e = __c.Set("id", "route-entry-derived"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1061
				__e = __c.Any("derived:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1062
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1063
}

//line beam_fragments.gox:1065
func (f *BeamRouteEntrypointsFragment) derivedDefault(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1066
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1067
				__e = __c.Set("id", "route-entry-derived-default"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1067
				__e = __c.Any("derived-default:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1068
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1069
}

//line beam_fragments.gox:1071
func (f *BeamRouteEntrypointsFragment) simpleLensRoute(l doors.Source[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1072
		__e = __c.Any(l.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1073
				__e = __c.Set("id", "route-entry-simple-lens"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1073
				__e = __c.Any("simple-lens:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1074
	})); if __e != nil { return }
//line beam_fragments.gox:1075
		__e = __c.Any(test.Button("route-entry-simple-lens-update", func(ctx context.Context) bool {
		l.Update(ctx, "simple-lens")
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1079
}

//line beam_fragments.gox:1081
func (f *BeamRouteEntrypointsFragment) simpleBeamRoute(b doors.Beam[string]) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1082
		__e = __c.Any(b.Bind(func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1083
				__e = __c.Set("id", "route-entry-simple-beam"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1083
				__e = __c.Any("simple-beam:" + v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1084
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1085
}

//line beam_fragments.gox:1087
func (f *BeamRouteEntrypointsFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line beam_fragments.gox:1089
		f.init()

//line beam_fragments.gox:1091
		__e = __c.Any(f.source.Bind(func(v state) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:1092
				__e = __c.Set("id", "route-entry-source"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line beam_fragments.gox:1092
				__e = __c.Any(fmt.Sprintf("%d:%s", v.Int, v.Str)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line beam_fragments.gox:1093
	})); if __e != nil { return }
//line beam_fragments.gox:1094
		__e = __c.Any(f.str.RouteSource(
		doors.RouteDerive(f.matchNonEmpty).Source(f.setString, f.lensRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1097
		__e = __c.Any(f.str.RouteBeam(
		doors.RouteDerive(f.matchNonEmpty).Beam(f.lensBeamRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1100
		__e = __c.Any(f.parity.RouteBeam(
		doors.RouteDerive(f.matchOdd).Beam(f.derivedRoute),
		doors.RouteDefaultBeam(f.derivedDefault),
	)); if __e != nil { return }
//line beam_fragments.gox:1104
		__e = __c.Any(f.str.RouteSource(
		doors.RouteMatch(func(v string) bool {
			return strings.HasPrefix(v, "simple")
		}).Source(f.simpleLensRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1109
		__e = __c.Any(f.str.RouteBeam(
		doors.RouteMatch(func(v string) bool {
			return strings.HasPrefix(v, "simple")
		}).Beam(f.simpleBeamRoute),
	)); if __e != nil { return }
//line beam_fragments.gox:1114
		__e = __c.Any(test.Button("route-entry-set-simple", func(ctx context.Context) bool {
		f.str.Update(ctx, "simple")
		return false
	})); if __e != nil { return }
//line beam_fragments.gox:1118
		__e = __c.Any(test.Button("route-entry-even", func(ctx context.Context) bool {
		f.source.Update(ctx, state{Int: 2, Str: f.str.Get()})
		return false
	})); if __e != nil { return }
	return })
//line beam_fragments.gox:1122
}
