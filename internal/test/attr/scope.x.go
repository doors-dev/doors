// Managed by GoX v0.1.36

//line scope.gox:1
package attr

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type scopeFragment struct {
	r *test.Reporter
	test.NoBeam
	counter atomic.Int32
}

func (f *scopeFragment) update(ctx context.Context, marker string) {
	i := f.counter.Add(1)
	f.r.Update(ctx, 0, fmt.Sprint(i - 1))
	f.r.Update(ctx, 1, marker)
}

//line scope.gox:26
func (f *scopeFragment) scopePipeline() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line scope.gox:28
		ds := &doors.ScopeDebounce{Duration: 250 * time.Millisecond, Limit: 0}
		ds2 := &doors.ScopeDebounce{Duration: 250 * time.Millisecond, Limit: 0}
		ss := &doors.ScopeSerial{}
		fs := &doors.ScopeFrame{}

//line scope.gox:33
		__e = __c.Any(f.button("p1", ds, "1", false)); if __e != nil { return }
//line scope.gox:34
		__e = __c.Any(f.button("p2", fs.Scope(false).And(ds).And(ss), "2", true)); if __e != nil { return }
//line scope.gox:35
		__e = __c.Any(f.button("p3", fs.Scope(false).And(ds2).And(ss), "3", true)); if __e != nil { return }
//line scope.gox:36
		__e = __c.Any(f.button("p4", fs.Scope(false).And(ds2).And(ss), "4", false)); if __e != nil { return }
//line scope.gox:37
		__e = __c.Any(f.button("p5", fs.Scope(true), "5", true)); if __e != nil { return }
	return })
//line scope.gox:38
}

//line scope.gox:40
func (f *scopeFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line scope.gox:42
		f.update(ctx, "0")

//line scope.gox:44
		__e = __c.Any(f.r); if __e != nil { return }
//line scope.gox:46
		blocking := &doors.ScopeBlocking{}

//line scope.gox:48
		__e = __c.Any(f.button("b1", blocking, "1", true)); if __e != nil { return }
//line scope.gox:49
		__e = __c.Any(f.button("b2", blocking, "2", true)); if __e != nil { return }
//line scope.gox:50
		__e = __c.Any(f.button("b3", blocking, "3", true)); if __e != nil { return }
//line scope.gox:52
		serial := &doors.ScopeSerial{}

//line scope.gox:54
		__e = __c.Any(f.button("s1", serial, "1", true)); if __e != nil { return }
//line scope.gox:55
		__e = __c.Any(f.button("s2", serial, "2", true)); if __e != nil { return }
//line scope.gox:56
		__e = __c.Any(f.button("s3", serial, "3", true)); if __e != nil { return }
//line scope.gox:58
		debouce := &doors.ScopeDebounce{Duration: 300 * time.Millisecond, Limit: 0}

//line scope.gox:60
		__e = __c.Any(f.button("d1", debouce, "1", false)); if __e != nil { return }
//line scope.gox:61
		__e = __c.Any(f.button("d2", debouce, "2", false)); if __e != nil { return }
//line scope.gox:62
		__e = __c.Any(f.button("d3", debouce, "3", false)); if __e != nil { return }
//line scope.gox:64
		debouce = &doors.ScopeDebounce{Duration: 300 * time.Millisecond, Limit: 700 * time.Millisecond}

//line scope.gox:66
		__e = __c.Any(f.button("dl1", debouce, "1", false)); if __e != nil { return }
//line scope.gox:67
		__e = __c.Any(f.button("dl2", debouce, "2", false)); if __e != nil { return }
//line scope.gox:68
		__e = __c.Any(f.button("dl3", debouce, "3", false)); if __e != nil { return }
//line scope.gox:70
		frame := &doors.ScopeFrame{}

//line scope.gox:72
		__e = __c.Any(f.button("f1", frame.Scope(false), "1", true)); if __e != nil { return }
//line scope.gox:73
		__e = __c.Any(f.button("f2", frame.Scope(false), "2", false)); if __e != nil { return }
//line scope.gox:74
		__e = __c.Any(f.button("f3", frame.Scope(true), "3", true)); if __e != nil { return }
//line scope.gox:75
		__e = __c.Any(f.button("f4", frame.Scope(true), "4", false)); if __e != nil { return }
//line scope.gox:78
		latest := &doors.ScopeLatest{}

//line scope.gox:80
		__e = __c.Any(f.buttonLatest("l1", latest, "1")); if __e != nil { return }
//line scope.gox:81
		__e = __c.Any(f.buttonLatest("l2", latest, "2")); if __e != nil { return }
//line scope.gox:82
		__e = __c.Any(f.buttonLatest("l3", latest, "3")); if __e != nil { return }
//line scope.gox:85
		concurrent := &doors.ScopeConcurrent{}

//line scope.gox:87
		__e = __c.Any(f.button("c1", concurrent.Scope(1), "1", true)); if __e != nil { return }
//line scope.gox:88
		__e = __c.Any(f.button("c2", concurrent.Scope(1), "2", true)); if __e != nil { return }
//line scope.gox:89
		__e = __c.Any(f.button("c3", concurrent.Scope(2), "3", false)); if __e != nil { return }
//line scope.gox:92
		rate := &doors.ScopeRate{Tick: 300 * time.Millisecond}

//line scope.gox:94
		__e = __c.Any(f.button("r1", rate, "1", false)); if __e != nil { return }
//line scope.gox:95
		__e = __c.Any(f.button("r2", rate, "2", false)); if __e != nil { return }
//line scope.gox:96
		__e = __c.Any(f.button("r3", rate, "3", false)); if __e != nil { return }
//line scope.gox:97
		__e = __c.Any(f.scopePipeline()); if __e != nil { return }
	return })
//line scope.gox:98
}

//line scope.gox:100
func (f *scopeFragment) button(id string, scope doors.Scopes, marker string, delay bool) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line scope.gox:101
			__e = __c.Set("id", id); if __e != nil { return }
//line scope.gox:101
			__e = __c.Modify(doors.A(ctx, f.handler(scope, marker, delay))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line scope.gox:101
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line scope.gox:102
}

//line scope.gox:104
func (f *scopeFragment) buttonLatest(id string, scope doors.Scopes, marker string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("button"); if __e != nil { return }
		{
//line scope.gox:105
			__e = __c.Set("id", id); if __e != nil { return }
//line scope.gox:105
			__e = __c.Modify(doors.A(ctx, f.handlerLatest(scope, marker))); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line scope.gox:105
			__e = __c.Any(id); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line scope.gox:106
}

func (f *scopeFragment) handler(scope doors.Scopes, marker string, delay bool) doors.Attr {
	return doors.AClick{
		Scope: scope,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			if delay {
				<-time.After(300 * time.Millisecond)
			}
			f.update(ctx, marker)
			return false
		},
	}
}

func (f *scopeFragment) handlerLatest(scope doors.Scopes, marker string) doors.Attr {
	return doors.AClick{
		Scope: scope,
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			select {
			case <-time.After(300 * time.Millisecond):
			case <-ctx.Done():
				return false
			}
			f.update(ctx, marker)
			return false
		},
	}
}
