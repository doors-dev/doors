// Managed by GoX v0.1.25

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
		__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
		__e = f.n.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
		f.n.Update(ctx, f.content())
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
		__e = n1.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
				n3.Update(ctx, test.ReportId(4, fmt.Sprint(s)))
				return false
			})

//line beam_fragments.gox:104
				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:106
		__e = n2.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
		__e = n1.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
				n3.Update(ctx, test.ReportId(4, fmt.Sprint(s.Int)))
				return false
			})

//line beam_fragments.gox:157
				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line beam_fragments.gox:159
		__e = n2.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
		__e = f.n.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
		__e = __c.Any(doors.Sub(f.p, func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line beam_fragments.gox:244
				__e = __c.AttrSet("id", "parity"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "watcher-i"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "watcher-newi"); if __e != nil { return }
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
//line beam_fragments.gox:297
		f.b.ReadAndSub(ctx, func(ctx context.Context, i int) bool {
			f.n.Update(ctx, f.content(i))
			return true
		})
		f.b.Mutate(ctx, func(i int) int {
			return i + 1
		})

//line beam_fragments.gox:305
		__e = __c.Any(&f.n); if __e != nil { return }
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
			__e = __c.AttrSet("id", "watcher-i"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "watcher-newi"); if __e != nil { return }
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
				f.n.Update(ctx, f.content(i))
				return true
			})
		}()

	return })
//line beam_fragments.gox:335
}

type BeamRenderUpdateWarningFragment struct {
	b    doors.Source[int]
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
			n3.Update(ctx, test.ReportId(4, fmt.Sprint(i)))
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
		__e = f.host.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
