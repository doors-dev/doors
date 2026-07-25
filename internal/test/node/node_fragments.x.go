// Managed by GoX v0.2.3

//line node_fragments.gox:1
package door

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type FragmentMany struct {
	n doors.Door
	test.NoBeam
}

//line node_fragments.gox:19
func (f *FragmentMany) sample() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:20
			__e = __c.Set("class", "sample"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("sample"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:21
}

//line node_fragments.gox:23
func (f *FragmentMany) manyDoors() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:24
		for i := range 20 {
			__e = __c.Init("span"); if __e != nil { return }
			{
//line node_fragments.gox:25
				__e = __c.Set("style", "display:none"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:25
				__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:26
			__e = __c.Any(&f.n); if __e != nil { return }
		}
	return })
//line node_fragments.gox:28
}

//line node_fragments.gox:30
func (f *FragmentMany) replaced() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:32
		f.n.Static(ctx, f.sample())

//line node_fragments.gox:34
		for i := range 100 {
			__e = __c.Init("span"); if __e != nil { return }
			{
//line node_fragments.gox:35
				__e = __c.Set("style", "display:none"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:35
				__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:36
			__e = __c.Any(&f.n); if __e != nil { return }
		}
	return })
//line node_fragments.gox:38
}

//line node_fragments.gox:40
func (f *FragmentMany) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:42
		f.n.Inner(ctx, f.sample())
	n := doors.Door{}

//line node_fragments.gox:45
		__e = (n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:46
				__e = __c.Any(f.manyDoors()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:48
		__e = __c.Any(test.Button("replace", func(ctx context.Context) bool {
		n.Inner(ctx, f.replaced())
		return true
	})); if __e != nil { return }
	return })
//line node_fragments.gox:52
}

type FragmentProxyWrappedSiblings struct {
	n doors.Door
	test.NoBeam
}

//line node_fragments.gox:59
func (f *FragmentProxyWrappedSiblings) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:60
		__e = __c.Any(gox.EditorFunc(func(cur gox.Cursor) error {
		return f.n.Proxy(cur, gox.Elem(func(cur gox.Cursor) error {
			if err := cur.Init("div"); err != nil {
				return err
			}
			if err := cur.Set("id", "proxy-wrap-first"); err != nil {
				return err
			}
			if err := cur.Submit(); err != nil {
				return err
			}
			if err := cur.Text("first"); err != nil {
				return err
			}
			if err := cur.Close(); err != nil {
				return err
			}
			
			if err := cur.Init("div"); err != nil {
				return err
			}
			if err := cur.Set("id", "proxy-wrap-second"); err != nil {
				return err
			}
			if err := cur.Submit(); err != nil {
				return err
			}
			if err := cur.Text("second"); err != nil {
				return err
			}
			return cur.Close()
		}))
	})); if __e != nil { return }
	return })
//line node_fragments.gox:93
}

type FragmentProxyWrappedLoop struct {
	n doors.Door
	test.NoBeam
}

//line node_fragments.gox:100
func (f *FragmentProxyWrappedLoop) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:101
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:101
			for i := range 2 {
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:102
					__e = __c.Set("id", fmt.Sprintf("proxy-loop-%d", i)); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:102
					__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:104
}

type FragmentX struct {
	report doors.Door
	n doors.Door
	test.NoBeam
}

func (f *FragmentX) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

//line node_fragments.gox:116
func (f *FragmentX) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:117
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:118
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:120
		__e = (f.report).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:120
			__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			ch := f.n.Inner(ctx, test.Marker("updated"))
			count := 0
			for err := range ch {
				count++
				if err != nil {
					f.rep(ctx, "channel err: " + err.Error())
					return false
				}
			}
			if count == 0 {
				f.rep(ctx, "channel closed")
				return false
			}
			f.rep(ctx, "ok upd")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("button"); if __e != nil { return }
				{
//line node_fragments.gox:138
					__e = __c.Set("id", "updatex"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("C"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:140
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			ch := f.n.Static(ctx, nil)
			count := 0
			for err := range ch {
				count++
				if err != nil {
					f.rep(ctx, "channel err: " + err.Error())
					return false
				}
			}
			if count == 0 {
				f.rep(ctx, "channel closed")
				return false
			}
			f.rep(ctx, "ok del")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:158
				__e = __c.Set("id", "removex"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("R"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:159
}

type FragmentXDoor struct {
	report doors.Door
	frame doors.Door
	n doors.Door
	test.NoBeam
}

func (f *FragmentXDoor) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

func (f *FragmentXDoor) wait(ctx context.Context, ch <-chan error, okMsg string) bool {
	count := 0
	for err := range ch {
		count++
		if err != nil {
			f.rep(ctx, "channel err: " + err.Error())
			return false
		}
	}
	if count == 0 {
		f.rep(ctx, "channel closed")
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:189
func (f *FragmentXDoor) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:190
		__e = __c.Any(&f.n); if __e != nil { return }
	return })
//line node_fragments.gox:191
}

//line node_fragments.gox:193
func (f *FragmentXDoor) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:195
		f.n.Inner(ctx, test.Marker("x-init"))
	f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:198
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:199
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:200
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Reload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:204
				__e = __c.Set("id", "xreload"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreload"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:205
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Outer(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("section"); if __e != nil { return }
				{
//line node_fragments.gox:207
					__e = __c.Set("id", "x-rebased-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:208
					__e = __c.Any(test.Marker("x-rebased")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:209
			return })), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:211
				__e = __c.Set("id", "xrebase"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xrebase"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:212
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Inner(ctx, nil), "ok clear")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:216
				__e = __c.Set("id", "xclear"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xclear"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:217
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Inner(ctx, test.Marker("x-updated")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:221
				__e = __c.Set("id", "xupdate"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xupdate"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:222
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Unmount(ctx), "ok unmount")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:226
				__e = __c.Set("id", "xunmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xunmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:227
		__e = __c.Any(test.Button("xremount", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:231
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Static(ctx, test.Marker("x-replaced")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:235
				__e = __c.Set("id", "xreplace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreplace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:236
}

type EmbeddedFragment struct {
	n1 doors.Door
	n2 doors.Door
	n3 doors.Door
	test.NoBeam
}

//line node_fragments.gox:245
func (f *EmbeddedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:246
		__e = (f.n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:247
				__e = (f.n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:248
						__e = __c.Any(test.Marker("init")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line node_fragments.gox:250
				__e = __c.Any(test.Marker("static")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:252
		__e = __c.Any(&f.n3); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:254
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:255
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.n2.Static(ctx, nil)
				return true
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("C"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:264
			__e = __c.Set("id", "clear"); if __e != nil { return }
//line node_fragments.gox:265
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.n1.Inner(ctx, nil)
				return true
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("C"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:274
			__e = __c.Set("id", "replace"); if __e != nil { return }
//line node_fragments.gox:275
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.n2.Inner(ctx, test.Marker("replaced"))
				f.n3.Inner(ctx, test.Marker("temp"))
				f.n3.Static(ctx, &f.n2)
				return true
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("C"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:285
}

type DynamicFragment struct {
	n1 doors.Door
	n2 doors.Door
	test.NoBeam
}

//line node_fragments.gox:293
func (f *DynamicFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:295
		f.n1.Inner(ctx, test.Marker("init"))

//line node_fragments.gox:298
		__e = (f.n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:301
			__e = __c.Set("id", "update"); if __e != nil { return }
//line node_fragments.gox:302
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.n1.Inner(ctx, test.Marker("updated"))
				return true
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("U"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:311
			__e = __c.Set("id", "replace"); if __e != nil { return }
//line node_fragments.gox:312
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.n2.Inner(ctx, test.Marker("replaced"))
				f.n1.Static(ctx, &f.n2)
				return true
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Rp"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:322
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:323
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.n2.Static(ctx, nil)
				return true
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Remove"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:331
}

type BeforeFragment struct {
	doorInit doors.Door
	doorUpdate doors.Door
	doorRemoved doors.Door
	doorReplaced doors.Door
	test.NoBeam
}

//line node_fragments.gox:341
func (f *BeforeFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:342
		__e = (f.doorInit).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:343
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:347
		f.doorUpdate.Inner(ctx, test.Marker("updated"))

//line node_fragments.gox:349
		__e = (f.doorUpdate).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:353
		f.doorRemoved.Inner(ctx, test.Marker("removed"))

//line node_fragments.gox:356
		f.doorRemoved.Static(ctx, nil)

//line node_fragments.gox:358
		__e = __c.Any(&f.doorRemoved); if __e != nil { return }
//line node_fragments.gox:361
		f.doorReplaced.Static(ctx, test.Marker("replaced"))

//line node_fragments.gox:364
		__e = (f.doorReplaced).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:365
				__e = __c.Any(test.Marker("initReplaced")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:367
}

type LifeCycleFragment struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

//line node_fragments.gox:375
func (f *LifeCycleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:376
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:376
			__e = __c.Any(f.initial()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:378
			__e = __c.Set("id", "reload"); if __e != nil { return }
//line node_fragments.gox:379
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Reload(ctx)
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Reload"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:388
			__e = __c.Set("id", "updateEmpty"); if __e != nil { return }
//line node_fragments.gox:389
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.frame.Inner(ctx, f.newEmpty())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Update1"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:398
			__e = __c.Set("id", "updateEmptyAlt"); if __e != nil { return }
//line node_fragments.gox:399
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.frame.Inner(ctx, f.newEmptyAlt())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Update1Alt"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:408
			__e = __c.Set("id", "updateContent"); if __e != nil { return }
//line node_fragments.gox:409
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.frame.Inner(ctx, f.newContent())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Update2"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:418
			__e = __c.Set("id", "updateInner"); if __e != nil { return }
//line node_fragments.gox:419
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Inner(ctx, test.Marker("inner-maintained"))
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("UpdateInner"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:428
			__e = __c.Set("id", "updateOuter"); if __e != nil { return }
//line node_fragments.gox:429
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Outer(ctx, f.newOuter())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("UpdateOuter"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:438
			__e = __c.Set("id", "replaceStatic"); if __e != nil { return }
//line node_fragments.gox:439
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Static(ctx, test.Marker("static-presist"))
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("ReplaceStatic"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:448
			__e = __c.Set("id", "updateEditor"); if __e != nil { return }
//line node_fragments.gox:449
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.frame.Inner(ctx, f.newEditor())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Update2"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:458
			__e = __c.Set("id", "clear"); if __e != nil { return }
//line node_fragments.gox:459
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Inner(ctx, nil)
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Clear"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:468
			__e = __c.Set("id", "unmount"); if __e != nil { return }
//line node_fragments.gox:469
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Unmount(ctx)
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Unmount"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:478
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:479
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Static(ctx, nil)
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Remove"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:487
}

//line node_fragments.gox:489
func (f *LifeCycleFragment) initial() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:491
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:492
					__e = __c.Any(test.Marker("presist")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:495
}
//line node_fragments.gox:496
func (f *LifeCycleFragment) newEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:498
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:498
					__e = __c.Set("id", "new"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:501
}

//line node_fragments.gox:503
func (f *LifeCycleFragment) newEmptyAlt() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:505
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("section"); if __e != nil { return }
				{
//line node_fragments.gox:505
					__e = __c.Set("id", "new-alt"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:508
}

//line node_fragments.gox:510
func (f *LifeCycleFragment) newContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:512
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:512
					__e = __c.Set("id", "new2"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:513
					__e = __c.Any(test.Marker("presist2")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:516
}

//line node_fragments.gox:518
func (f *LifeCycleFragment) newOuter() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:519
			__e = __c.Set("id", "outer-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:520
			__e = __c.Any(test.Marker("outer-presist")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:522
}

//line node_fragments.gox:524
func (f *LifeCycleFragment) newEditor() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:526
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:528
}

type FragmentProxyReloadContent struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

//line node_fragments.gox:536
func (f *FragmentProxyReloadContent) mountEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:537
			__e = __c.Set("id", "proxy-redraw-frame"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:538
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:538
					__e = __c.Set("id", "proxy-redraw-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:541
}

//line node_fragments.gox:543
func (f *FragmentProxyReloadContent) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:544
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:544
			__e = __c.Any(f.mountEmpty()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:546
			__e = __c.Set("id", "proxy-redraw-update"); if __e != nil { return }
//line node_fragments.gox:547
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Inner(ctx, test.Marker("proxy-redraw-content"))
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-redraw-update"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:556
			__e = __c.Set("id", "proxy-redraw-remount"); if __e != nil { return }
//line node_fragments.gox:557
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.frame.Inner(ctx, f.mountEmpty())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-redraw-remount"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:566
			__e = __c.Set("id", "proxy-redraw-reload"); if __e != nil { return }
//line node_fragments.gox:567
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Reload(ctx)
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-redraw-reload"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:575
}

type FragmentClosestReload struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:585
func (f *FragmentClosestReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:587
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:589
			__e = __c.Set("id", "inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:589
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:591
			__e = __c.Set("id", "reload-nearest"); if __e != nil { return }
//line node_fragments.gox:592
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				doors.Reload(ctx)
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("reload-nearest"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:600
}

//line node_fragments.gox:602
func (f *FragmentClosestReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:604
		f.outerRenders++
	f.node.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:607
			__e = __c.Set("id", "outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:607
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:608
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:609
}

//line node_fragments.gox:611
func (f *FragmentClosestReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:613
		f.frame.Inner(ctx, f.outerContent())

//line node_fragments.gox:615
		__e = __c.Any(&f.frame); if __e != nil { return }
	return })
//line node_fragments.gox:616
}

type FragmentClosestReloadProxy struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:626
func (f *FragmentClosestReloadProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:627
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:627
				__e = __c.Set("id", "outer-proxy-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:629
				f.outerRenders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:631
					__e = __c.Set("id", "proxy-outer-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:631
					__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:632
				__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line node_fragments.gox:632
						__e = __c.Set("id", "inner-proxy-root"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:634
						f.innerRenders++

						__e = __c.Init("div"); if __e != nil { return }
						{
//line node_fragments.gox:636
							__e = __c.Set("id", "proxy-inner-count"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:636
							__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("button"); if __e != nil { return }
						{
//line node_fragments.gox:638
							__e = __c.Set("id", "reload-nearest-proxy"); if __e != nil { return }
//line node_fragments.gox:639
							__e = __c.Modify(doors.AClick{
					On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
						doors.Reload(ctx)
						return false
					},
				}); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Text("reload-nearest-proxy"); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:649
}

type FragmentInlineDoorPointerProxy struct {
	renders int
	test.NoBeam
}

//line node_fragments.gox:656
func (f *FragmentInlineDoorPointerProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:657
		__e = (&doors.Door{}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:657
				__e = __c.Set("id", "inline-door-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:659
				f.renders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:661
					__e = __c.Set("id", "inline-door-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:661
					__e = __c.Any(fmt.Sprintf("inline-%d", f.renders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line node_fragments.gox:663
					__e = __c.Set("id", "inline-door-reload"); if __e != nil { return }
//line node_fragments.gox:664
					__e = __c.Modify(doors.AClick{
				On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
					doors.Reload(ctx)
					return false
				},
			}); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("inline-door-reload"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:673
}

type containerEffectAttr struct {
	source doors.Source[int]
}

func (a containerEffectAttr) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	value, ok := a.source.Effect(ctx)
	if !ok {
		attrs.Get("data-container-effect").Set("canceled")
		return nil
	}
	attrs.Get("data-container-effect").Set(fmt.Sprint(value))
	return nil
}

type containerWatchAttr struct {
	source doors.Source[int]
	cancels *atomic.Int64
	watches *atomic.Int64
}

func (a containerWatchAttr) Modify(ctx context.Context, _ string, attrs gox.Attrs) error {
	_, _ = a.source.Watch(ctx, &containerLifecycleWatcher{
		cancels: a.cancels,
		watches: a.watches,
	})
	attrs.Get("data-container-watch").Set("on")
	return nil
}

type containerLifecycleWatcher struct {
	cancels *atomic.Int64
	watches *atomic.Int64
}

func (w *containerLifecycleWatcher) Watch(context.Context, int) bool {
	w.watches.Add(1)
	return false
}

func (w *containerLifecycleWatcher) Cancel() {
	w.cancels.Add(1)
}

type FragmentContainerInnerLifecycle struct {
	node doors.Door
	clicks int
	test.NoBeam
}

//line node_fragments.gox:724
func (f *FragmentContainerInnerLifecycle) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:725
			__e = __c.Set("id", "container-inner-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:725
			__e = __c.Any(fmt.Sprintf("click-%d", f.clicks)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:726
}

//line node_fragments.gox:728
func (f *FragmentContainerInnerLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:729
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:730
				__e = __c.Set("id", "container-inner-root"); if __e != nil { return }
//line node_fragments.gox:731
				__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.clicks++
				f.node.Inner(ctx, f.content())
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line node_fragments.gox:738
					__e = __c.Set("id", "container-inner-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("initial"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:740
}

type FragmentContainerEffectLifecycle struct {
	node doors.Door
	source doors.Source[int]
	test.NoBeam
}

func (f *FragmentContainerEffectLifecycle) init() {
	if f.source == nil {
		f.source = doors.NewSource(0)
	}
}

//line node_fragments.gox:754
func (f *FragmentContainerEffectLifecycle) content(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:755
			__e = __c.Set("id", "container-effect-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:755
			__e = __c.Any(fmt.Sprintf("%s-%d", label, f.source.Get())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:756
}

//line node_fragments.gox:758
func (f *FragmentContainerEffectLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:760
		f.init()

//line node_fragments.gox:762
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:763
				__e = __c.Set("id", "container-effect-root"); if __e != nil { return }
//line node_fragments.gox:764
				__e = __c.Modify(containerEffectAttr{source: f.source}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:765
				__e = __c.Any(f.content("initial")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:767
		__e = __c.Any(test.Button("container-effect-inner", func(ctx context.Context) bool {
		f.node.Inner(ctx, f.content("inner"))
		return false
	})); if __e != nil { return }
//line node_fragments.gox:771
		__e = __c.Any(test.Button("container-effect-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:775
}

type FragmentContainerOuterLifecycle struct {
	node doors.Door
	report doors.Door
	source doors.Source[int]
	cancels atomic.Int64
	watches atomic.Int64
	outerClicks int
	test.NoBeam
}

func (f *FragmentContainerOuterLifecycle) init() {
	if f.source == nil {
		f.source = doors.NewSource(0)
	}
}

func (f *FragmentContainerOuterLifecycle) reportState(ctx context.Context) {
	f.report.Inner(ctx, test.Report(fmt.Sprintf("cancels-%d watches-%d", f.cancels.Load(), f.watches.Load())))
}

//line node_fragments.gox:797
func (f *FragmentContainerOuterLifecycle) outerInner() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:798
			__e = __c.Set("id", "container-outer-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:798
			__e = __c.Any(fmt.Sprintf("outer-click-%d", f.outerClicks)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:799
}

//line node_fragments.gox:801
func (f *FragmentContainerOuterLifecycle) outer() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:803
			__e = __c.Set("id", "container-outer-new-root"); if __e != nil { return }
//line node_fragments.gox:804
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.outerClicks++
				f.node.Inner(ctx, f.outerInner())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("span"); if __e != nil { return }
			{
//line node_fragments.gox:811
				__e = __c.Set("id", "container-outer-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("outer"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:813
}

//line node_fragments.gox:815
func (f *FragmentContainerOuterLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:817
		f.init()

//line node_fragments.gox:819
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:820
				__e = __c.Set("id", "container-outer-root"); if __e != nil { return }
//line node_fragments.gox:821
				__e = __c.Modify(containerWatchAttr{source: f.source, cancels: &f.cancels, watches: &f.watches}); if __e != nil { return }
//line node_fragments.gox:822
				__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Outer(ctx, f.outer())
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line node_fragments.gox:828
					__e = __c.Set("id", "container-outer-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("initial"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:830
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:831
		__e = __c.Any(test.Button("container-outer-report", func(ctx context.Context) bool {
		f.reportState(ctx)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:835
}

type FragmentContainerReloadLifecycle struct {
	node doors.Door
	report doors.Door
	source doors.Source[int]
	cancels atomic.Int64
	watches atomic.Int64
	clicks int
	test.NoBeam
}

func (f *FragmentContainerReloadLifecycle) init() {
	if f.source == nil {
		f.source = doors.NewSource(0)
	}
}

func (f *FragmentContainerReloadLifecycle) reportState(ctx context.Context) {
	f.report.Inner(ctx, test.Report(fmt.Sprintf("cancels-%d watches-%d", f.cancels.Load(), f.watches.Load())))
}

func (f *FragmentContainerReloadLifecycle) value() string {
	if f.clicks == 0 {
		return "initial"
	}
	return fmt.Sprintf("click-%d", f.clicks)
}

//line node_fragments.gox:864
func (f *FragmentContainerReloadLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:866
		f.init()

//line node_fragments.gox:868
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:869
				__e = __c.Set("id", "container-reload-root"); if __e != nil { return }
//line node_fragments.gox:870
				__e = __c.Modify(containerWatchAttr{source: f.source, cancels: &f.cancels, watches: &f.watches}); if __e != nil { return }
//line node_fragments.gox:871
				__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.clicks++
				f.node.Reload(ctx)
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line node_fragments.gox:878
					__e = __c.Set("id", "container-reload-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:878
					__e = __c.Any(f.value()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:880
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:881
		__e = __c.Any(test.Button("container-reload-report", func(ctx context.Context) bool {
		f.reportState(ctx)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:885
}

type FragmentContainerHookStateLifecycle struct {
	node doors.Door
	report doors.Door
	source doors.Source[int]
	derived doors.Beam[string]
	last atomic.Value
	registered bool
	subEvents atomic.Int64
	watches atomic.Int64
	cancels atomic.Int64
	test.NoBeam
}

func (f *FragmentContainerHookStateLifecycle) init() {
	if f.source != nil {
		return
	}
	f.source = doors.NewSource(0)
	f.derived = doors.DeriveBeam(f.source, func(v int) string {
		return fmt.Sprintf("derived-%d", v)
	})
}

func (f *FragmentContainerHookStateLifecycle) reportText(prefix string) string {
	last, _ := f.last.Load().(string)
	if last == "" {
		last = "none"
	}
	return fmt.Sprintf(
		"%s %s value-%d sub-%d watches-%d cancels-%d",
		prefix,
		last,
		f.source.Get(),
		f.subEvents.Load(),
		f.watches.Load(),
		f.cancels.Load(),
	)
}

func (f *FragmentContainerHookStateLifecycle) reportState(ctx context.Context, prefix string) {
	f.report.Inner(ctx, test.Report(f.reportText(prefix)))
}

func (f *FragmentContainerHookStateLifecycle) registerState(ctx context.Context) {
	current, readOK := f.source.Read(ctx)
	derived, derivedOK := f.derived.Read(ctx)
	initial, subOK := f.source.ReadAndSub(ctx, func(ctx context.Context, v int) bool {
		seq := f.subEvents.Add(1)
		f.report.Inner(ctx, test.Report(fmt.Sprintf("sub-%d-%d", seq, v)))
		return false
	})
	_, watchOK := f.source.Watch(ctx, &containerLifecycleWatcher{
		cancels: &f.cancels,
		watches: &f.watches,
	})
	f.last.Store(fmt.Sprintf(
		"read-%d-%t derived-%s-%t initial-%d-%t watch-%t",
		current,
		readOK,
		derived,
		derivedOK,
		initial,
		subOK,
		watchOK,
	))
}

//line node_fragments.gox:954
func (f *FragmentContainerHookStateLifecycle) content(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:955
			__e = __c.Set("id", "container-hook-state-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:955
			__e = __c.Any(fmt.Sprintf("%s-%d", label, f.source.Get())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:956
}

//line node_fragments.gox:958
func (f *FragmentContainerHookStateLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:960
		f.init()

//line node_fragments.gox:962
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:963
				__e = __c.Set("id", "container-hook-state-root"); if __e != nil { return }
//line node_fragments.gox:964
				__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				if !f.registered {
					f.registered = true
					f.registerState(ctx)
					f.node.Inner(ctx, f.content("registered"))
					return false
				}
				f.source.Mutate(ctx, func(v int) int {
					return v + 1
				})
				f.node.Inner(ctx, f.content("mutated"))
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:979
				__e = __c.Any(f.content("initial")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:981
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:982
		__e = __c.Any(test.Button("container-hook-state-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:986
		__e = __c.Any(test.Button("container-hook-state-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:990
}

type FragmentContainerHookStateOuterLifecycle struct {
	node doors.Door
	report doors.Door
	source doors.Source[int]
	derived doors.Beam[string]
	last atomic.Value
	subEvents atomic.Int64
	watches atomic.Int64
	cancels atomic.Int64
	outerClicks int
	test.NoBeam
}

func (f *FragmentContainerHookStateOuterLifecycle) init() {
	if f.source != nil {
		return
	}
	f.source = doors.NewSource(0)
	f.derived = doors.DeriveBeam(f.source, func(v int) string {
		return fmt.Sprintf("derived-%d", v)
	})
}

func (f *FragmentContainerHookStateOuterLifecycle) reportText(prefix string) string {
	last, _ := f.last.Load().(string)
	if last == "" {
		last = "none"
	}
	return fmt.Sprintf(
		"%s %s value-%d sub-%d watches-%d cancels-%d",
		prefix,
		last,
		f.source.Get(),
		f.subEvents.Load(),
		f.watches.Load(),
		f.cancels.Load(),
	)
}

func (f *FragmentContainerHookStateOuterLifecycle) reportState(ctx context.Context, prefix string) {
	f.report.Inner(ctx, test.Report(f.reportText(prefix)))
}

func (f *FragmentContainerHookStateOuterLifecycle) registerState(ctx context.Context) {
	current, readOK := f.source.Read(ctx)
	derived, derivedOK := f.derived.Read(ctx)
	initial, subOK := f.source.ReadAndSub(ctx, func(ctx context.Context, v int) bool {
		seq := f.subEvents.Add(1)
		f.report.Inner(ctx, test.Report(fmt.Sprintf("outer-sub-%d-%d", seq, v)))
		return false
	})
	_, watchOK := f.source.Watch(ctx, &containerLifecycleWatcher{
		cancels: &f.cancels,
		watches: &f.watches,
	})
	f.last.Store(fmt.Sprintf(
		"read-%d-%t derived-%s-%t initial-%d-%t watch-%t",
		current,
		readOK,
		derived,
		derivedOK,
		initial,
		subOK,
		watchOK,
	))
}

//line node_fragments.gox:1059
func (f *FragmentContainerHookStateOuterLifecycle) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:1060
			__e = __c.Set("id", "container-hook-outer-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1060
			__e = __c.Any(fmt.Sprintf("outer-click-%d", f.outerClicks)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1061
}

//line node_fragments.gox:1063
func (f *FragmentContainerHookStateOuterLifecycle) outer() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1065
			__e = __c.Set("id", "container-hook-outer-new-root"); if __e != nil { return }
//line node_fragments.gox:1066
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.registerState(ctx)
				f.outerClicks++
				f.source.Mutate(ctx, func(v int) int {
					return v + 1
				})
				f.node.Inner(ctx, f.outerContent())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("span"); if __e != nil { return }
			{
//line node_fragments.gox:1077
				__e = __c.Set("id", "container-hook-outer-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("outer"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1079
}

//line node_fragments.gox:1081
func (f *FragmentContainerHookStateOuterLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1083
		f.init()

//line node_fragments.gox:1085
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:1086
				__e = __c.Set("id", "container-hook-outer-root"); if __e != nil { return }
//line node_fragments.gox:1087
				__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.registerState(ctx)
				f.node.Outer(ctx, f.outer())
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line node_fragments.gox:1094
					__e = __c.Set("id", "container-hook-outer-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("initial"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1096
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1097
		__e = __c.Any(test.Button("container-hook-outer-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1101
		__e = __c.Any(test.Button("container-hook-outer-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1105
}

type FragmentContainerHookEffectLifecycle struct {
	node doors.Door
	report doors.Door
	source doors.Source[int]
	last atomic.Value
	renders atomic.Int64
	registrations atomic.Int64
	test.NoBeam
}

func (f *FragmentContainerHookEffectLifecycle) init() {
	if f.source == nil {
		f.source = doors.NewSource(0)
	}
}

func (f *FragmentContainerHookEffectLifecycle) reportText(prefix string) string {
	last, _ := f.last.Load().(string)
	if last == "" {
		last = "none"
	}
	return fmt.Sprintf(
		"%s %s value-%d renders-%d registrations-%d",
		prefix,
		last,
		f.source.Get(),
		f.renders.Load(),
		f.registrations.Load(),
	)
}

func (f *FragmentContainerHookEffectLifecycle) reportState(ctx context.Context, prefix string) {
	f.report.Inner(ctx, test.Report(f.reportText(prefix)))
}

//line node_fragments.gox:1142
func (f *FragmentContainerHookEffectLifecycle) content(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1144
		f.renders.Add(1)

		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:1146
			__e = __c.Set("id", "container-hook-effect-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1146
			__e = __c.Any(fmt.Sprint(f.renders.Load())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:1147
			__e = __c.Set("id", "container-hook-effect-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1147
			__e = __c.Any(fmt.Sprintf("%s-%d", label, f.source.Get())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1148
}

//line node_fragments.gox:1150
func (f *FragmentContainerHookEffectLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1152
		f.init()

//line node_fragments.gox:1154
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:1155
				__e = __c.Set("id", "container-hook-effect-root"); if __e != nil { return }
//line node_fragments.gox:1156
				__e = __c.Modify(doors.AKeyDown{
			Filter: []string{"ContainerEffect"},
			On: func(ctx context.Context, _ doors.RequestEvent[doors.KeyboardEvent]) bool {
				value, ok := f.source.Effect(ctx)
				f.registrations.Add(1)
				f.last.Store(fmt.Sprintf("effect-%d-%t", value, ok))
				f.node.Inner(ctx, f.content("registered"))
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1166
				__e = __c.Any(f.content("outer")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1168
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1169
		__e = __c.Any(test.Button("container-hook-effect-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1173
		__e = __c.Any(test.Button("container-hook-effect-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1177
}

type FragmentContainerHookStateReloadLifecycle struct {
	node doors.Door
	report doors.Door
	source doors.Source[int]
	derived doors.Beam[string]
	last atomic.Value
	reloads int
	subEvents atomic.Int64
	watches atomic.Int64
	cancels atomic.Int64
	test.NoBeam
}

func (f *FragmentContainerHookStateReloadLifecycle) init() {
	if f.source != nil {
		return
	}
	f.source = doors.NewSource(0)
	f.derived = doors.DeriveBeam(f.source, func(v int) string {
		return fmt.Sprintf("derived-%d", v)
	})
}

func (f *FragmentContainerHookStateReloadLifecycle) reportText(prefix string) string {
	last, _ := f.last.Load().(string)
	if last == "" {
		last = "none"
	}
	return fmt.Sprintf(
		"%s %s value-%d sub-%d watches-%d cancels-%d",
		prefix,
		last,
		f.source.Get(),
		f.subEvents.Load(),
		f.watches.Load(),
		f.cancels.Load(),
	)
}

func (f *FragmentContainerHookStateReloadLifecycle) reportState(ctx context.Context, prefix string) {
	f.report.Inner(ctx, test.Report(f.reportText(prefix)))
}

func (f *FragmentContainerHookStateReloadLifecycle) registerState(ctx context.Context) {
	current, readOK := f.source.Read(ctx)
	derived, derivedOK := f.derived.Read(ctx)
	initial, subOK := f.source.ReadAndSub(ctx, func(ctx context.Context, v int) bool {
		seq := f.subEvents.Add(1)
		f.report.Inner(ctx, test.Report(fmt.Sprintf("reload-sub-%d-%d", seq, v)))
		return false
	})
	_, watchOK := f.source.Watch(ctx, &containerLifecycleWatcher{
		cancels: &f.cancels,
		watches: &f.watches,
	})
	f.last.Store(fmt.Sprintf(
		"read-%d-%t derived-%s-%t initial-%d-%t watch-%t",
		current,
		readOK,
		derived,
		derivedOK,
		initial,
		subOK,
		watchOK,
	))
}

func (f *FragmentContainerHookStateReloadLifecycle) value() string {
	if f.reloads == 0 {
		return "initial"
	}
	return fmt.Sprintf("reload-%d", f.reloads)
}

//line node_fragments.gox:1253
func (f *FragmentContainerHookStateReloadLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1255
		f.init()

//line node_fragments.gox:1257
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:1258
				__e = __c.Set("id", "container-hook-reload-root"); if __e != nil { return }
//line node_fragments.gox:1259
				__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.registerState(ctx)
				f.reloads++
				f.node.Reload(ctx)
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line node_fragments.gox:1267
					__e = __c.Set("id", "container-hook-reload-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1267
					__e = __c.Any(f.value()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1269
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1270
		__e = __c.Any(test.Button("container-hook-reload-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1274
		__e = __c.Any(test.Button("container-hook-reload-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1278
}

type FragmentClosestTrackedReload struct {
	frame doors.Door
	node doors.Door
	report doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

func (f *FragmentClosestTrackedReload) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

func (f *FragmentClosestTrackedReload) wait(ctx context.Context, ch <-chan error, okMsg string) bool {
	count := 0
	for err := range ch {
		count++
		if err != nil {
			f.rep(ctx, "channel err: " + err.Error())
			return false
		}
	}
	if count == 0 {
		f.rep(ctx, "channel closed")
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1310
func (f *FragmentClosestTrackedReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1312
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:1314
			__e = __c.Set("id", "x-inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1314
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:1316
			__e = __c.Set("id", "xreload-nearest"); if __e != nil { return }
//line node_fragments.gox:1317
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				return f.wait(ctx, doors.Reload(ctx), "ok xreload")
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("xreload-nearest"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1324
}

//line node_fragments.gox:1326
func (f *FragmentClosestTrackedReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1328
		f.outerRenders++
	f.node.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:1331
			__e = __c.Set("id", "x-outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1331
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:1332
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1333
}

//line node_fragments.gox:1335
func (f *FragmentClosestTrackedReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1337
		f.frame.Inner(ctx, f.outerContent())

//line node_fragments.gox:1339
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1340
		__e = __c.Any(&f.report); if __e != nil { return }
	return })
//line node_fragments.gox:1341
}

type FragmentRootTrackedReload struct {
	report doors.Door
	test.NoBeam
}

func (f *FragmentRootTrackedReload) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

//line node_fragments.gox:1352
func (f *FragmentRootTrackedReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1353
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:1355
			__e = __c.Set("id", "root-xreload"); if __e != nil { return }
//line node_fragments.gox:1356
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				ch := doors.Reload(ctx)
				count := 0
				for err := range ch {
					count++
					if err != nil {
						f.rep(ctx, "channel err: " + err.Error())
						return false
					}
				}
				if count == 0 {
					f.rep(ctx, "channel closed")
					return false
				}
				f.rep(ctx, "ok xreload")
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("root-xreload"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1377
}

type FragmentDetachedReplace struct {
	report doors.Door
	frame doors.Door
	node doors.Door
	test.NoBeam
}

func (f *FragmentDetachedReplace) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

func (f *FragmentDetachedReplace) wait(ctx context.Context, ch <-chan error, okMsg string) bool {
	count := 0
	for err := range ch {
		count++
		if err != nil {
			f.rep(ctx, "channel err: " + err.Error())
			return false
		}
	}
	if count == 0 {
		f.rep(ctx, "channel closed")
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1407
func (f *FragmentDetachedReplace) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1408
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1409
}

//line node_fragments.gox:1411
func (f *FragmentDetachedReplace) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1413
		f.node.Inner(ctx, test.Marker("replace-base"))
	f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1416
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1417
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1418
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Static(ctx, test.Marker("replace-detached")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1422
				__e = __c.Set("id", "replace-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1423
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Reload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1427
				__e = __c.Set("id", "reload-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1428
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Inner(ctx, test.Marker("replace-updated")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1432
				__e = __c.Set("id", "update-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1433
		__e = __c.Any(test.Button("remount-after-replace", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1437
}

type FragmentDetachedRebase struct {
	report doors.Door
	frame doors.Door
	node doors.Door
	test.NoBeam
}

func (f *FragmentDetachedRebase) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

func (f *FragmentDetachedRebase) wait(ctx context.Context, ch <-chan error, okMsg string) bool {
	count := 0
	for err := range ch {
		count++
		if err != nil {
			f.rep(ctx, "channel err: " + err.Error())
			return false
		}
	}
	if count == 0 {
		f.rep(ctx, "channel closed")
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1467
func (f *FragmentDetachedRebase) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1468
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1469
}

//line node_fragments.gox:1471
func (f *FragmentDetachedRebase) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1472
			__e = __c.Set("id", "rebased-detached-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1473
			__e = __c.Any(test.Marker("rebased-detached")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1475
}

//line node_fragments.gox:1477
func (f *FragmentDetachedRebase) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1479
		f.node.Inner(ctx, test.Marker("rebase-base"))
	f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1482
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1483
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1484
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Unmount(ctx), "ok unmount")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1488
				__e = __c.Set("id", "unmount-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("unmount-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1489
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Reload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1493
				__e = __c.Set("id", "reload-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1494
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Outer(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1498
				__e = __c.Set("id", "rebase-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1499
		__e = __c.Any(test.Button("remount-after-rebase", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1503
}

type FragmentProxyMove struct {
	report doors.Door
	frame1 doors.Door
	frame2 doors.Door
	node doors.Door
	test.NoBeam
}

func (f *FragmentProxyMove) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

func (f *FragmentProxyMove) wait(ctx context.Context, ch <-chan error, okMsg string) bool {
	count := 0
	for err := range ch {
		count++
		if err != nil {
			f.rep(ctx, "channel err: " + err.Error())
			return false
		}
	}
	if count == 0 {
		f.rep(ctx, "channel closed")
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1534
func (f *FragmentProxyMove) mountFrame1() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1535
			__e = __c.Set("id", "frame1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1536
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1538
}

//line node_fragments.gox:1540
func (f *FragmentProxyMove) mountFrame2() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1541
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1542
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1544
}

//line node_fragments.gox:1546
func (f *FragmentProxyMove) frame2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1547
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1548
}

//line node_fragments.gox:1550
func (f *FragmentProxyMove) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1551
			__e = __c.Set("id", "proxy-moved-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1552
			__e = __c.Any(test.Marker("proxy-moved")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1554
}

//line node_fragments.gox:1556
func (f *FragmentProxyMove) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1558
		f.node.Inner(ctx, test.Marker("proxy-base"))
	f.frame1.Inner(ctx, f.mountFrame1())
	f.frame2.Inner(ctx, f.frame2Empty())

//line node_fragments.gox:1562
		__e = __c.Any(&f.frame1); if __e != nil { return }
//line node_fragments.gox:1563
		__e = __c.Any(&f.frame2); if __e != nil { return }
//line node_fragments.gox:1564
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1565
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Outer(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1569
				__e = __c.Set("id", "rebase-proxy-move"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-proxy-move"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1570
		__e = __c.Any(test.Button("move-proxy", func(ctx context.Context) bool {
		f.frame2.Inner(ctx, f.mountFrame2())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1574
}

type FragmentHierarchy struct {
	report doors.Door
	host1 doors.Door
	host2 doors.Door
	child doors.Door
	grand doors.Door
	test.NoBeam
}

func (f *FragmentHierarchy) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

func (f *FragmentHierarchy) wait(ctx context.Context, ch <-chan error, okMsg string) bool {
	count := 0
	for err := range ch {
		count++
		if err != nil {
			f.rep(ctx, "channel err: " + err.Error())
			return false
		}
	}
	if count == 0 {
		f.rep(ctx, "channel closed")
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1606
func (f *FragmentHierarchy) childBody() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("article"); if __e != nil { return }
		{
//line node_fragments.gox:1607
			__e = __c.Set("id", "child-body"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1608
			__e = __c.Any(&f.grand); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1610
}

//line node_fragments.gox:1612
func (f *FragmentHierarchy) host1Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1613
			__e = __c.Set("id", "host1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1614
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1616
}

//line node_fragments.gox:1618
func (f *FragmentHierarchy) host2Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1619
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1620
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1622
}

//line node_fragments.gox:1624
func (f *FragmentHierarchy) host2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1625
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1626
}

//line node_fragments.gox:1628
func (f *FragmentHierarchy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1630
		f.grand.Inner(ctx, test.Marker("grand-init"))
	f.child.Inner(ctx, f.childBody())
	f.host1.Inner(ctx, f.host1Body())
	f.host2.Inner(ctx, f.host2Empty())

//line node_fragments.gox:1635
		__e = __c.Any(&f.host1); if __e != nil { return }
//line node_fragments.gox:1636
		__e = __c.Any(&f.host2); if __e != nil { return }
//line node_fragments.gox:1637
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1638
		__e = __c.Any(test.Button("move-child", func(ctx context.Context) bool {
		f.host2.Inner(ctx, f.host2Body())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1642
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.grand.Inner(ctx, test.Marker("grand-updated")), "ok grand")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1646
				__e = __c.Set("id", "grand-update"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("grand-update"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1647
		__e = __c.Any(test.Button("remove-host2", func(ctx context.Context) bool {
		f.host2.Static(ctx, nil)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1651
}

type FragmentErrorTransitions struct {
	report doors.Door
	frame doors.Door
	node doors.Door
	test.NoBeam
}

func (f *FragmentErrorTransitions) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

func (f *FragmentErrorTransitions) wait(ctx context.Context, ch <-chan error, okMsg string) bool {
	count := 0
	for err := range ch {
		count++
		if err != nil {
			f.rep(ctx, "channel err: " + err.Error())
			return false
		}
	}
	if count == 0 {
		f.rep(ctx, "channel closed")
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

func (f *FragmentErrorTransitions) errElem(msg string) gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		return errors.New(msg)
	})
}

//line node_fragments.gox:1687
func (f *FragmentErrorTransitions) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1688
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1689
}

//line node_fragments.gox:1691
func (f *FragmentErrorTransitions) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1693
		f.node.Inner(ctx, test.Marker("error-base"))
	f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1696
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1697
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1698
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Inner(ctx, f.errElem("update boom")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1702
				__e = __c.Set("id", "update-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1703
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Static(ctx, f.errElem("replace boom")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1707
				__e = __c.Set("id", "replace-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1708
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Outer(ctx, f.errElem("rebase boom")), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1712
				__e = __c.Set("id", "rebase-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:1713
}
