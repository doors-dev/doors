// Managed by GoX v0.2.3

//line dispatch.gox:1
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

// Served as a managed inline TypeScript script, so the $sys TS annotation in
// internal/resources/providers.go must stay syntactically valid for the page
// (and therefore every test below) to work. The "keysErr" listener is the
// sink for the #keys-err OnError action: it must NEVER fire for a declined
// (non-matching key) event, so #onerr staying empty pins the onErr half of
// the declined-event swallow.
const dispatchScript = `
window.__sysDispatch = (selector: string, event: Event) => {
	return $sys.dispatch(document.querySelector(selector)!, event)
}
$on("keysErr", (arg: any) => {
	document.querySelector("#onerr")!.textContent += String(arg)
})
`

type dispatchFragment struct {
	r *test.Reporter
	test.NoBeam
	clicks atomic.Int32
	debounce atomic.Int32
}

//line dispatch.gox:36
func (f *dispatchFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line dispatch.gox:38
		f.r.Update(ctx, 0, "0")
	f.r.Update(ctx, 1, "")
	f.r.Update(ctx, 2, "0")
	f.r.Update(ctx, 3, "")
	f.r.Update(ctx, 4, "")
	debounce := &doors.ScopeDebounce{Duration: 150 * time.Millisecond, Limit: 0}

//line dispatch.gox:45
		__e = __c.Any(f.r); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line dispatch.gox:47
			__e = __c.Set("id", "btn"); if __e != nil { return }
//line dispatch.gox:48
			__e = __c.Modify(doors.A(ctx, doors.AClick{
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				f.r.Update(ctx, 0, fmt.Sprint(f.clicks.Add(1)))
				return false
			},
		})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("btn"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line dispatch.gox:57
			__e = __c.Set("type", "text"); if __e != nil { return }
//line dispatch.gox:58
			__e = __c.Set("id", "keys"); if __e != nil { return }
//line dispatch.gox:59
			__e = __c.Modify(doors.A(ctx, doors.AKeyDown{
			Keys: []doors.Key{{Key: "Enter"}},
			On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
				f.r.Update(ctx, 1, "enter")
				return false
			},
		})); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.InitVoid("input"); if __e != nil { return }
		{
//line dispatch.gox:70
			__e = __c.Set("type", "text"); if __e != nil { return }
//line dispatch.gox:71
			__e = __c.Set("id", "keys-err"); if __e != nil { return }
//line dispatch.gox:72
			__e = __c.Modify(doors.A(ctx, doors.AKeyDown{
			Keys: []doors.Key{{Key: "Enter"}},
			OnError: doors.ActionEmit{Name: "keysErr", Arg: "onerr"},
			On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
				f.r.Update(ctx, 4, "enter")
				return false
			},
		})); if __e != nil { return }
		}
		__e = __c.Submit(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line dispatch.gox:80
			__e = __c.Set("id", "onerr"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line dispatch.gox:82
			__e = __c.Set("id", "deb"); if __e != nil { return }
//line dispatch.gox:83
			__e = __c.Modify(doors.A(ctx, doors.AClick{
			Scope: debounce,
			On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
				f.r.Update(ctx, 2, fmt.Sprint(f.debounce.Add(1)))
				return false
			},
		})); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("deb"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line dispatch.gox:92
		__e = (doors.AClick{
		On: func(ctx context.Context, r doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 3, "parent")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line dispatch.gox:97
				__e = __c.Set("id", "parent"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line dispatch.gox:98
					__e = __c.Set("id", "child"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("child"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("script"); if __e != nil { return }
		{
//line dispatch.gox:100
			__e = __c.Set("src", doors.ResourceString(dispatchScript)); if __e != nil { return }
			__e = __c.Set("inline", true); if __e != nil { return }
//line dispatch.gox:100
			__e = __c.Set("type", "typescript"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Raw(""); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line dispatch.gox:101
}
