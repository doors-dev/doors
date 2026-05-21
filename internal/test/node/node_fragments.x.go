// Managed by GoX v0.1.32

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
			ch := f.n.XInner(ctx, test.Marker("updated"))
			err, ok := <-ch
			if !ok {
				f.rep(ctx, "channel closed")
				return false
			}
			if err != nil {
				f.rep(ctx, "channel err: " + err.Error())
				return false
			}
			f.rep(ctx, "ok upd")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("button"); if __e != nil { return }
				{
//line node_fragments.gox:135
					__e = __c.Set("id", "updatex"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("C"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:137
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			ch := f.n.XStatic(ctx, nil)
			err, ok := <-ch
			if !ok {
				f.rep(ctx, "channel closed")
				return false
			}
			if err != nil {
				f.rep(ctx, "channel err: " + err.Error())
				return false
			}
			f.rep(ctx, "ok del")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:152
				__e = __c.Set("id", "removex"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("R"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:153
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
	err, ok := <-ch
	if !ok {
		f.rep(ctx, "channel closed")
		return false
	}
	if err != nil {
		f.rep(ctx, "channel err: " + err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:180
func (f *FragmentXDoor) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:181
		__e = __c.Any(&f.n); if __e != nil { return }
	return })
//line node_fragments.gox:182
}

//line node_fragments.gox:184
func (f *FragmentXDoor) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:186
		f.n.Inner(ctx, test.Marker("x-init"))
		f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:189
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:190
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:191
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XReload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:195
				__e = __c.Set("id", "xreload"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreload"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:196
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XOuter(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("section"); if __e != nil { return }
				{
//line node_fragments.gox:198
					__e = __c.Set("id", "x-rebased-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:199
					__e = __c.Any(test.Marker("x-rebased")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:200
			return })), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:202
				__e = __c.Set("id", "xrebase"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xrebase"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:203
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XInner(ctx, nil), "ok clear")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:207
				__e = __c.Set("id", "xclear"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xclear"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:208
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XInner(ctx, test.Marker("x-updated")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:212
				__e = __c.Set("id", "xupdate"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xupdate"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:213
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XUnmount(ctx), "ok unmount")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:217
				__e = __c.Set("id", "xunmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xunmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:218
		__e = __c.Any(test.Button("xremount", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:222
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XStatic(ctx, test.Marker("x-replaced")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:226
				__e = __c.Set("id", "xreplace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreplace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:227
}

type EmbeddedFragment struct {
	n1 doors.Door
	n2 doors.Door
	n3 doors.Door
	test.NoBeam
}

//line node_fragments.gox:236
func (f *EmbeddedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:237
		__e = (f.n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:238
				__e = (f.n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:239
						__e = __c.Any(test.Marker("init")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line node_fragments.gox:241
				__e = __c.Any(test.Marker("static")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:243
		__e = __c.Any(&f.n3); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:245
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:246
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
//line node_fragments.gox:255
			__e = __c.Set("id", "clear"); if __e != nil { return }
//line node_fragments.gox:256
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
//line node_fragments.gox:265
			__e = __c.Set("id", "replace"); if __e != nil { return }
//line node_fragments.gox:266
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
//line node_fragments.gox:276
}

type DynamicFragment struct {
	n1 doors.Door
	n2 doors.Door
	test.NoBeam
}

//line node_fragments.gox:284
func (f *DynamicFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:286
		f.n1.Inner(ctx, test.Marker("init"))

//line node_fragments.gox:289
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
//line node_fragments.gox:292
			__e = __c.Set("id", "update"); if __e != nil { return }
//line node_fragments.gox:293
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
//line node_fragments.gox:302
			__e = __c.Set("id", "replace"); if __e != nil { return }
//line node_fragments.gox:303
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
//line node_fragments.gox:313
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:314
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
//line node_fragments.gox:322
}

type BeforeFragment struct {
	doorInit doors.Door
	doorUpdate doors.Door
	doorRemoved doors.Door
	doorReplaced doors.Door
	test.NoBeam
}

//line node_fragments.gox:332
func (f *BeforeFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:333
		__e = (f.doorInit).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:334
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:338
		f.doorUpdate.Inner(ctx, test.Marker("updated"))

//line node_fragments.gox:340
		__e = (f.doorUpdate).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:344
		f.doorRemoved.Inner(ctx, test.Marker("removed"))

//line node_fragments.gox:347
		f.doorRemoved.Static(ctx, nil)

//line node_fragments.gox:349
		__e = __c.Any(&f.doorRemoved); if __e != nil { return }
//line node_fragments.gox:352
		f.doorReplaced.Static(ctx, test.Marker("replaced"))

//line node_fragments.gox:355
		__e = (f.doorReplaced).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:356
				__e = __c.Any(test.Marker("initReplaced")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:358
}

type LifeCycleFragment struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

//line node_fragments.gox:366
func (f *LifeCycleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:367
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:367
			__e = __c.Any(f.initial()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:369
			__e = __c.Set("id", "reload"); if __e != nil { return }
//line node_fragments.gox:370
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
//line node_fragments.gox:379
			__e = __c.Set("id", "updateEmpty"); if __e != nil { return }
//line node_fragments.gox:380
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
//line node_fragments.gox:389
			__e = __c.Set("id", "updateEmptyAlt"); if __e != nil { return }
//line node_fragments.gox:390
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
//line node_fragments.gox:399
			__e = __c.Set("id", "updateContent"); if __e != nil { return }
//line node_fragments.gox:400
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
//line node_fragments.gox:409
			__e = __c.Set("id", "updateInner"); if __e != nil { return }
//line node_fragments.gox:410
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
//line node_fragments.gox:419
			__e = __c.Set("id", "updateOuter"); if __e != nil { return }
//line node_fragments.gox:420
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
//line node_fragments.gox:429
			__e = __c.Set("id", "replaceStatic"); if __e != nil { return }
//line node_fragments.gox:430
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
//line node_fragments.gox:439
			__e = __c.Set("id", "updateEditor"); if __e != nil { return }
//line node_fragments.gox:440
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
//line node_fragments.gox:449
			__e = __c.Set("id", "clear"); if __e != nil { return }
//line node_fragments.gox:450
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
//line node_fragments.gox:459
			__e = __c.Set("id", "unmount"); if __e != nil { return }
//line node_fragments.gox:460
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
//line node_fragments.gox:469
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:470
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
//line node_fragments.gox:478
}

//line node_fragments.gox:480
func (f *LifeCycleFragment) initial() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:482
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:483
					__e = __c.Any(test.Marker("presist")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:486
}
//line node_fragments.gox:487
func (f *LifeCycleFragment) newEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:489
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:489
					__e = __c.Set("id", "new"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:492
}

//line node_fragments.gox:494
func (f *LifeCycleFragment) newEmptyAlt() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:496
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("section"); if __e != nil { return }
				{
//line node_fragments.gox:496
					__e = __c.Set("id", "new-alt"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:499
}

//line node_fragments.gox:501
func (f *LifeCycleFragment) newContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:503
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:503
					__e = __c.Set("id", "new2"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:504
					__e = __c.Any(test.Marker("presist2")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:507
}

//line node_fragments.gox:509
func (f *LifeCycleFragment) newOuter() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:510
			__e = __c.Set("id", "outer-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:511
			__e = __c.Any(test.Marker("outer-presist")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:513
}

//line node_fragments.gox:515
func (f *LifeCycleFragment) newEditor() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:517
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:519
}

type FragmentProxyReloadContent struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

//line node_fragments.gox:527
func (f *FragmentProxyReloadContent) mountEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:528
			__e = __c.Set("id", "proxy-redraw-frame"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:529
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:529
					__e = __c.Set("id", "proxy-redraw-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:532
}

//line node_fragments.gox:534
func (f *FragmentProxyReloadContent) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:535
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:535
			__e = __c.Any(f.mountEmpty()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:537
			__e = __c.Set("id", "proxy-redraw-update"); if __e != nil { return }
//line node_fragments.gox:538
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
//line node_fragments.gox:547
			__e = __c.Set("id", "proxy-redraw-remount"); if __e != nil { return }
//line node_fragments.gox:548
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
//line node_fragments.gox:557
			__e = __c.Set("id", "proxy-redraw-reload"); if __e != nil { return }
//line node_fragments.gox:558
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
//line node_fragments.gox:566
}

type FragmentClosestReload struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:576
func (f *FragmentClosestReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:578
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:580
			__e = __c.Set("id", "inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:580
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:582
			__e = __c.Set("id", "reload-nearest"); if __e != nil { return }
//line node_fragments.gox:583
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
//line node_fragments.gox:591
}

//line node_fragments.gox:593
func (f *FragmentClosestReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:595
		f.outerRenders++
		f.node.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:598
			__e = __c.Set("id", "outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:598
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:599
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:600
}

//line node_fragments.gox:602
func (f *FragmentClosestReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:604
		f.frame.Inner(ctx, f.outerContent())

//line node_fragments.gox:606
		__e = __c.Any(&f.frame); if __e != nil { return }
	return })
//line node_fragments.gox:607
}

type FragmentClosestReloadProxy struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:617
func (f *FragmentClosestReloadProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:618
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:618
				__e = __c.Set("id", "outer-proxy-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:620
				f.outerRenders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:622
					__e = __c.Set("id", "proxy-outer-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:622
					__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:623
				__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line node_fragments.gox:623
						__e = __c.Set("id", "inner-proxy-root"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:625
						f.innerRenders++

						__e = __c.Init("div"); if __e != nil { return }
						{
//line node_fragments.gox:627
							__e = __c.Set("id", "proxy-inner-count"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:627
							__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("button"); if __e != nil { return }
						{
//line node_fragments.gox:629
							__e = __c.Set("id", "reload-nearest-proxy"); if __e != nil { return }
//line node_fragments.gox:630
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
//line node_fragments.gox:640
}

type FragmentInlineDoorPointerProxy struct {
	renders int
	test.NoBeam
}

//line node_fragments.gox:647
func (f *FragmentInlineDoorPointerProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:648
		__e = (&doors.Door{}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:648
				__e = __c.Set("id", "inline-door-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:650
				f.renders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:652
					__e = __c.Set("id", "inline-door-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:652
					__e = __c.Any(fmt.Sprintf("inline-%d", f.renders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line node_fragments.gox:654
					__e = __c.Set("id", "inline-door-reload"); if __e != nil { return }
//line node_fragments.gox:655
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
//line node_fragments.gox:664
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

//line node_fragments.gox:715
func (f *FragmentContainerInnerLifecycle) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:716
			__e = __c.Set("id", "container-inner-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:716
			__e = __c.Any(fmt.Sprintf("click-%d", f.clicks)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:717
}

//line node_fragments.gox:719
func (f *FragmentContainerInnerLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:720
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:721
				__e = __c.Set("id", "container-inner-root"); if __e != nil { return }
//line node_fragments.gox:722
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
//line node_fragments.gox:729
					__e = __c.Set("id", "container-inner-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("initial"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:731
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

//line node_fragments.gox:745
func (f *FragmentContainerEffectLifecycle) content(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:746
			__e = __c.Set("id", "container-effect-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:746
			__e = __c.Any(fmt.Sprintf("%s-%d", label, f.source.Get())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:747
}

//line node_fragments.gox:749
func (f *FragmentContainerEffectLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:751
		f.init()

//line node_fragments.gox:753
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:754
				__e = __c.Set("id", "container-effect-root"); if __e != nil { return }
//line node_fragments.gox:755
				__e = __c.Modify(containerEffectAttr{source: f.source}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:756
				__e = __c.Any(f.content("initial")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:758
		__e = __c.Any(test.Button("container-effect-inner", func(ctx context.Context) bool {
		f.node.Inner(ctx, f.content("inner"))
		return false
	})); if __e != nil { return }
//line node_fragments.gox:762
		__e = __c.Any(test.Button("container-effect-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:766
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

//line node_fragments.gox:788
func (f *FragmentContainerOuterLifecycle) outerInner() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:789
			__e = __c.Set("id", "container-outer-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:789
			__e = __c.Any(fmt.Sprintf("outer-click-%d", f.outerClicks)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:790
}

//line node_fragments.gox:792
func (f *FragmentContainerOuterLifecycle) outer() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:794
			__e = __c.Set("id", "container-outer-new-root"); if __e != nil { return }
//line node_fragments.gox:795
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
//line node_fragments.gox:802
				__e = __c.Set("id", "container-outer-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("outer"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:804
}

//line node_fragments.gox:806
func (f *FragmentContainerOuterLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:808
		f.init()

//line node_fragments.gox:810
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:811
				__e = __c.Set("id", "container-outer-root"); if __e != nil { return }
//line node_fragments.gox:812
				__e = __c.Modify(containerWatchAttr{source: f.source, cancels: &f.cancels, watches: &f.watches}); if __e != nil { return }
//line node_fragments.gox:813
				__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Outer(ctx, f.outer())
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line node_fragments.gox:819
					__e = __c.Set("id", "container-outer-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("initial"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:821
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:822
		__e = __c.Any(test.Button("container-outer-report", func(ctx context.Context) bool {
		f.reportState(ctx)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:826
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

//line node_fragments.gox:855
func (f *FragmentContainerReloadLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:857
		f.init()

//line node_fragments.gox:859
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:860
				__e = __c.Set("id", "container-reload-root"); if __e != nil { return }
//line node_fragments.gox:861
				__e = __c.Modify(containerWatchAttr{source: f.source, cancels: &f.cancels, watches: &f.watches}); if __e != nil { return }
//line node_fragments.gox:862
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
//line node_fragments.gox:869
					__e = __c.Set("id", "container-reload-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:869
					__e = __c.Any(f.value()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:871
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:872
		__e = __c.Any(test.Button("container-reload-report", func(ctx context.Context) bool {
		f.reportState(ctx)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:876
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

//line node_fragments.gox:945
func (f *FragmentContainerHookStateLifecycle) content(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:946
			__e = __c.Set("id", "container-hook-state-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:946
			__e = __c.Any(fmt.Sprintf("%s-%d", label, f.source.Get())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:947
}

//line node_fragments.gox:949
func (f *FragmentContainerHookStateLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:951
		f.init()

//line node_fragments.gox:953
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:954
				__e = __c.Set("id", "container-hook-state-root"); if __e != nil { return }
//line node_fragments.gox:955
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
//line node_fragments.gox:970
				__e = __c.Any(f.content("initial")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:972
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:973
		__e = __c.Any(test.Button("container-hook-state-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:977
		__e = __c.Any(test.Button("container-hook-state-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:981
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

//line node_fragments.gox:1050
func (f *FragmentContainerHookStateOuterLifecycle) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:1051
			__e = __c.Set("id", "container-hook-outer-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1051
			__e = __c.Any(fmt.Sprintf("outer-click-%d", f.outerClicks)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1052
}

//line node_fragments.gox:1054
func (f *FragmentContainerHookStateOuterLifecycle) outer() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1056
			__e = __c.Set("id", "container-hook-outer-new-root"); if __e != nil { return }
//line node_fragments.gox:1057
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
//line node_fragments.gox:1068
				__e = __c.Set("id", "container-hook-outer-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("outer"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1070
}

//line node_fragments.gox:1072
func (f *FragmentContainerHookStateOuterLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1074
		f.init()

//line node_fragments.gox:1076
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:1077
				__e = __c.Set("id", "container-hook-outer-root"); if __e != nil { return }
//line node_fragments.gox:1078
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
//line node_fragments.gox:1085
					__e = __c.Set("id", "container-hook-outer-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("initial"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1087
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1088
		__e = __c.Any(test.Button("container-hook-outer-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1092
		__e = __c.Any(test.Button("container-hook-outer-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1096
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

//line node_fragments.gox:1133
func (f *FragmentContainerHookEffectLifecycle) content(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1135
		f.renders.Add(1)

		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:1137
			__e = __c.Set("id", "container-hook-effect-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1137
			__e = __c.Any(fmt.Sprint(f.renders.Load())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:1138
			__e = __c.Set("id", "container-hook-effect-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1138
			__e = __c.Any(fmt.Sprintf("%s-%d", label, f.source.Get())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1139
}

//line node_fragments.gox:1141
func (f *FragmentContainerHookEffectLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1143
		f.init()

//line node_fragments.gox:1145
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:1146
				__e = __c.Set("id", "container-hook-effect-root"); if __e != nil { return }
//line node_fragments.gox:1147
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
//line node_fragments.gox:1157
				__e = __c.Any(f.content("outer")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1159
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1160
		__e = __c.Any(test.Button("container-hook-effect-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1164
		__e = __c.Any(test.Button("container-hook-effect-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1168
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

//line node_fragments.gox:1244
func (f *FragmentContainerHookStateReloadLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1246
		f.init()

//line node_fragments.gox:1248
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:1249
				__e = __c.Set("id", "container-hook-reload-root"); if __e != nil { return }
//line node_fragments.gox:1250
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
//line node_fragments.gox:1258
					__e = __c.Set("id", "container-hook-reload-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1258
					__e = __c.Any(f.value()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1260
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1261
		__e = __c.Any(test.Button("container-hook-reload-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1265
		__e = __c.Any(test.Button("container-hook-reload-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1269
}

type FragmentClosestXReload struct {
	frame doors.Door
	node doors.Door
	report doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

func (f *FragmentClosestXReload) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

func (f *FragmentClosestXReload) wait(ctx context.Context, ch <-chan error, okMsg string) bool {
	err, ok := <-ch
	if !ok {
		f.rep(ctx, "channel closed")
		return false
	}
	if err != nil {
		f.rep(ctx, "channel err: " + err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1298
func (f *FragmentClosestXReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1300
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:1302
			__e = __c.Set("id", "x-inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1302
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:1304
			__e = __c.Set("id", "xreload-nearest"); if __e != nil { return }
//line node_fragments.gox:1305
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				return f.wait(ctx, doors.XReload(ctx), "ok xreload")
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("xreload-nearest"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1312
}

//line node_fragments.gox:1314
func (f *FragmentClosestXReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1316
		f.outerRenders++
		f.node.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:1319
			__e = __c.Set("id", "x-outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1319
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:1320
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1321
}

//line node_fragments.gox:1323
func (f *FragmentClosestXReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1325
		f.frame.Inner(ctx, f.outerContent())

//line node_fragments.gox:1327
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1328
		__e = __c.Any(&f.report); if __e != nil { return }
	return })
//line node_fragments.gox:1329
}

type FragmentRootXReload struct {
	report doors.Door
	test.NoBeam
}

func (f *FragmentRootXReload) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

//line node_fragments.gox:1340
func (f *FragmentRootXReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1341
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:1343
			__e = __c.Set("id", "root-xreload"); if __e != nil { return }
//line node_fragments.gox:1344
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				err, ok := <-doors.XReload(ctx)
				if !ok {
					f.rep(ctx, "channel closed")
					return false
				}
				if err != nil {
					f.rep(ctx, "channel err: " + err.Error())
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
//line node_fragments.gox:1361
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
	err, ok := <-ch
	if !ok {
		f.rep(ctx, "channel closed")
		return false
	}
	if err != nil {
		f.rep(ctx, "channel err: " + err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1388
func (f *FragmentDetachedReplace) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1389
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1390
}

//line node_fragments.gox:1392
func (f *FragmentDetachedReplace) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1394
		f.node.Inner(ctx, test.Marker("replace-base"))
		f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1397
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1398
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1399
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XStatic(ctx, test.Marker("replace-detached")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1403
				__e = __c.Set("id", "replace-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1404
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1408
				__e = __c.Set("id", "reload-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1409
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XInner(ctx, test.Marker("replace-updated")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1413
				__e = __c.Set("id", "update-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1414
		__e = __c.Any(test.Button("remount-after-replace", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1418
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
	err, ok := <-ch
	if !ok {
		f.rep(ctx, "channel closed")
		return false
	}
	if err != nil {
		f.rep(ctx, "channel err: " + err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1445
func (f *FragmentDetachedRebase) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1446
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1447
}

//line node_fragments.gox:1449
func (f *FragmentDetachedRebase) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1450
			__e = __c.Set("id", "rebased-detached-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1451
			__e = __c.Any(test.Marker("rebased-detached")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1453
}

//line node_fragments.gox:1455
func (f *FragmentDetachedRebase) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1457
		f.node.Inner(ctx, test.Marker("rebase-base"))
		f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1460
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1461
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1462
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUnmount(ctx), "ok unmount")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1466
				__e = __c.Set("id", "unmount-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("unmount-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1467
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1471
				__e = __c.Set("id", "reload-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1472
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XOuter(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1476
				__e = __c.Set("id", "rebase-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1477
		__e = __c.Any(test.Button("remount-after-rebase", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1481
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
	err, ok := <-ch
	if !ok {
		f.rep(ctx, "channel closed")
		return false
	}
	if err != nil {
		f.rep(ctx, "channel err: " + err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1509
func (f *FragmentProxyMove) mountFrame1() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1510
			__e = __c.Set("id", "frame1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1511
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1513
}

//line node_fragments.gox:1515
func (f *FragmentProxyMove) mountFrame2() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1516
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1517
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1519
}

//line node_fragments.gox:1521
func (f *FragmentProxyMove) frame2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1522
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1523
}

//line node_fragments.gox:1525
func (f *FragmentProxyMove) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1526
			__e = __c.Set("id", "proxy-moved-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1527
			__e = __c.Any(test.Marker("proxy-moved")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1529
}

//line node_fragments.gox:1531
func (f *FragmentProxyMove) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1533
		f.node.Inner(ctx, test.Marker("proxy-base"))
		f.frame1.Inner(ctx, f.mountFrame1())
		f.frame2.Inner(ctx, f.frame2Empty())

//line node_fragments.gox:1537
		__e = __c.Any(&f.frame1); if __e != nil { return }
//line node_fragments.gox:1538
		__e = __c.Any(&f.frame2); if __e != nil { return }
//line node_fragments.gox:1539
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1540
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XOuter(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1544
				__e = __c.Set("id", "rebase-proxy-move"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-proxy-move"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1545
		__e = __c.Any(test.Button("move-proxy", func(ctx context.Context) bool {
		f.frame2.Inner(ctx, f.mountFrame2())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1549
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
	err, ok := <-ch
	if !ok {
		f.rep(ctx, "channel closed")
		return false
	}
	if err != nil {
		f.rep(ctx, "channel err: " + err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

//line node_fragments.gox:1578
func (f *FragmentHierarchy) childBody() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("article"); if __e != nil { return }
		{
//line node_fragments.gox:1579
			__e = __c.Set("id", "child-body"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1580
			__e = __c.Any(&f.grand); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1582
}

//line node_fragments.gox:1584
func (f *FragmentHierarchy) host1Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1585
			__e = __c.Set("id", "host1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1586
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1588
}

//line node_fragments.gox:1590
func (f *FragmentHierarchy) host2Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1591
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1592
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1594
}

//line node_fragments.gox:1596
func (f *FragmentHierarchy) host2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1597
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1598
}

//line node_fragments.gox:1600
func (f *FragmentHierarchy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1602
		f.grand.Inner(ctx, test.Marker("grand-init"))
		f.child.Inner(ctx, f.childBody())
		f.host1.Inner(ctx, f.host1Body())
		f.host2.Inner(ctx, f.host2Empty())

//line node_fragments.gox:1607
		__e = __c.Any(&f.host1); if __e != nil { return }
//line node_fragments.gox:1608
		__e = __c.Any(&f.host2); if __e != nil { return }
//line node_fragments.gox:1609
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1610
		__e = __c.Any(test.Button("move-child", func(ctx context.Context) bool {
		f.host2.Inner(ctx, f.host2Body())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1614
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.grand.XInner(ctx, test.Marker("grand-updated")), "ok grand")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1618
				__e = __c.Set("id", "grand-update"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("grand-update"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1619
		__e = __c.Any(test.Button("remove-host2", func(ctx context.Context) bool {
		f.host2.Static(ctx, nil)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1623
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
	err, ok := <-ch
	if !ok {
		f.rep(ctx, "channel closed")
		return false
	}
	if err != nil {
		f.rep(ctx, "channel err: " + err.Error())
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

//line node_fragments.gox:1656
func (f *FragmentErrorTransitions) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1657
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1658
}

//line node_fragments.gox:1660
func (f *FragmentErrorTransitions) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1662
		f.node.Inner(ctx, test.Marker("error-base"))
		f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1665
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1666
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1667
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XInner(ctx, f.errElem("update boom")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1671
				__e = __c.Set("id", "update-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1672
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XStatic(ctx, f.errElem("replace boom")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1676
				__e = __c.Set("id", "replace-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1677
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XOuter(ctx, f.errElem("rebase boom")), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1681
				__e = __c.Set("id", "rebase-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:1682
}
