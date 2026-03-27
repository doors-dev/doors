// Managed by GoX v0.1.17+dirty

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

func (f *BeamSkipFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.r.Update(ctx, 0, "init")
		f.b.ReadAndSub(ctx, func(ctx context.Context, s state) bool {
			<-time.After(300 * time.Millisecond)
			return false
		})

		__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				f.b.Sub(ctx, func(ctx context.Context, s state) bool {
				if s.Str == "1" {
					f.r.Update(ctx, 0, "propagated")
				}
				return false
			})

			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(test.Button("update1", func(ctx context.Context) bool {
		f.b.Update(ctx, state{Str: "1"})
		return false
	})); if __e != nil { return }
		__e = __c.Any(test.Button("update2", func(ctx context.Context) bool {
		f.b.Update(ctx, state{Str: "2"})
		return false
	})); if __e != nil { return }
		__e = __c.Any(f.r); if __e != nil { return }
	return })
}

type BeamDeriveFragment struct {
	r *test.Reporter
	b doors.Source[state]
	n doors.Door
	test.NoBeam
}
func (f *BeamDeriveFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = f.n.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(test.Button("reload", func(ctx context.Context) bool {
		f.n.Update(ctx, f.content())
		return true
	})); if __e != nil { return }
	return })
}

func (f *BeamDeriveFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
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

		__e = __c.Any(test.ReportId(1, fmt.Sprint(r))); if __e != nil { return }
		__e = n1.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := d.Read(ctx)

				__e = __c.Any(test.ReportId(2, fmt.Sprint(r))); if __e != nil { return }
				n3 := doors.Door{}
			d.Sub(ctx, func(ctx context.Context, s int) bool {
				n3.Update(ctx, test.ReportId(4, fmt.Sprint(s)))
				return false
			})

				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = n2.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := f.b.Read(ctx)

				__e = __c.Any(test.ReportId(3, fmt.Sprint(r.Int))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(f.r); if __e != nil { return }
	return })
}

type BeamConsistentFragment struct {
	r *test.Reporter
	b doors.Source[state]
	n doors.Door
	test.NoBeam
}

func (f *BeamConsistentFragment) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
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

		__e = __c.Any(test.ReportId(1, fmt.Sprint(r.Int))); if __e != nil { return }
		__e = n1.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := f.b.Read(ctx)

				__e = __c.Any(test.ReportId(2, fmt.Sprint(r.Int))); if __e != nil { return }
				n3 := doors.Door{}
			f.b.Sub(ctx, func(ctx context.Context, s state) bool {
				n3.Update(ctx, test.ReportId(4, fmt.Sprint(s.Int)))
				return false
			})

				__e = __c.Any(&n3); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = n2.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				f.b.Mutate(ctx, func(s state) state {
				s.Int = s.Int + 1
				return s
			})
			r, _ := f.b.Read(ctx)

				__e = __c.Any(test.ReportId(3, fmt.Sprint(r.Int))); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(f.r); if __e != nil { return }
	return })
}

func (f *BeamConsistentFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = f.n.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				__e = __c.Any(f.content()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(test.Button("reload", func(ctx context.Context) bool {
		f.n.Reload(ctx)
		return true
	})); if __e != nil { return }
	return })
}

type BeamUpdateFragment struct {
	r *test.Reporter
	b doors.Source[state]
	test.NoBeam
}

func (f *BeamUpdateFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.b.Sub(ctx, func(ctx context.Context, s state) bool {
			f.r.Update(ctx, 0, fmt.Sprint(s.Int))
			return false
		})

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
}

type BeamEqualFragment struct {
	r *test.Reporter
	b doors.Source[state]
	p doors.Beam[string]
	test.NoBeam
}

func (f *BeamEqualFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		if f.p == nil {
			f.p = doors.NewBeamEqual(f.b, func(s state) string {
				if s.Int%2 == 0 {
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

		__e = __c.Any(doors.Sub(f.p, func(v string) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "parity"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Any(v); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
	})); if __e != nil { return }
		__e = __c.Any(doors.Go(func(ctx context.Context) {
		<-time.After(100 * time.Millisecond)
		f.r.Update(ctx, 2, "go")
	})); if __e != nil { return }
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
}
