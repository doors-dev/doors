// Managed by GoX v0.1.28

//line node_fragments.gox:1
package door

import (
	"context"
	"errors"
	"fmt"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type FragmentMany struct {
	n doors.Door
	test.NoBeam
}

//line node_fragments.gox:18
func (f *FragmentMany) sample() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:19
			__e = __c.Set("class", "sample"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("sample"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:20
}

//line node_fragments.gox:22
func (f *FragmentMany) manyDoors() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:23
		for i := range 20 {
			__e = __c.Init("span"); if __e != nil { return }
			{
//line node_fragments.gox:24
				__e = __c.Set("style", "display:none"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:24
				__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:25
			__e = __c.Any(&f.n); if __e != nil { return }
		}
	return })
//line node_fragments.gox:27
}

//line node_fragments.gox:29
func (f *FragmentMany) replaced() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:31
		f.n.Static(ctx, f.sample())

//line node_fragments.gox:33
		for i := range 100 {
			__e = __c.Init("span"); if __e != nil { return }
			{
//line node_fragments.gox:34
				__e = __c.Set("style", "display:none"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:34
				__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:35
			__e = __c.Any(&f.n); if __e != nil { return }
		}
	return })
//line node_fragments.gox:37
}

//line node_fragments.gox:39
func (f *FragmentMany) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:41
		f.n.Inner(ctx, f.sample())
		n := doors.Door{}

//line node_fragments.gox:44
		__e = (n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:45
				__e = __c.Any(f.manyDoors()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:47
		__e = __c.Any(test.Button("replace", func(ctx context.Context) bool {
		n.Inner(ctx, f.replaced())
		return true
	})); if __e != nil { return }
	return })
//line node_fragments.gox:51
}

type FragmentProxyWrappedSiblings struct {
	n doors.Door
	test.NoBeam
}

//line node_fragments.gox:58
func (f *FragmentProxyWrappedSiblings) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:59
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
//line node_fragments.gox:92
}

type FragmentProxyWrappedLoop struct {
	n doors.Door
	test.NoBeam
}

//line node_fragments.gox:99
func (f *FragmentProxyWrappedLoop) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:100
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:100
			for i := range 2 {
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:101
					__e = __c.Set("id", fmt.Sprintf("proxy-loop-%d", i)); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:101
					__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:103
}

type FragmentX struct {
	report doors.Door
	n doors.Door
	test.NoBeam
}

func (f *FragmentX) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

//line node_fragments.gox:115
func (f *FragmentX) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:116
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:117
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:119
		__e = (f.report).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:119
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
//line node_fragments.gox:134
					__e = __c.Set("id", "updatex"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("C"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:136
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
//line node_fragments.gox:151
				__e = __c.Set("id", "removex"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("R"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:152
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

//line node_fragments.gox:179
func (f *FragmentXDoor) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:180
		__e = __c.Any(&f.n); if __e != nil { return }
	return })
//line node_fragments.gox:181
}

//line node_fragments.gox:183
func (f *FragmentXDoor) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:185
		f.n.Inner(ctx, test.Marker("x-init"))
		f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:188
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:189
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:190
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XReload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:194
				__e = __c.Set("id", "xreload"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreload"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:195
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XOuter(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("section"); if __e != nil { return }
				{
//line node_fragments.gox:197
					__e = __c.Set("id", "x-rebased-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:198
					__e = __c.Any(test.Marker("x-rebased")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:199
			return })), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:201
				__e = __c.Set("id", "xrebase"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xrebase"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:202
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XInner(ctx, nil), "ok clear")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:206
				__e = __c.Set("id", "xclear"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xclear"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:207
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XInner(ctx, test.Marker("x-updated")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:211
				__e = __c.Set("id", "xupdate"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xupdate"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:212
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XUnmount(ctx), "ok unmount")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:216
				__e = __c.Set("id", "xunmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xunmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:217
		__e = __c.Any(test.Button("xremount", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:221
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XStatic(ctx, test.Marker("x-replaced")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:225
				__e = __c.Set("id", "xreplace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreplace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:226
}

type EmbeddedFragment struct {
	n1 doors.Door
	n2 doors.Door
	n3 doors.Door
	test.NoBeam
}

//line node_fragments.gox:235
func (f *EmbeddedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:236
		__e = (f.n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:237
				__e = (f.n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:238
						__e = __c.Any(test.Marker("init")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line node_fragments.gox:240
				__e = __c.Any(test.Marker("static")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:242
		__e = __c.Any(&f.n3); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:244
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:245
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
//line node_fragments.gox:254
			__e = __c.Set("id", "clear"); if __e != nil { return }
//line node_fragments.gox:255
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
//line node_fragments.gox:264
			__e = __c.Set("id", "replace"); if __e != nil { return }
//line node_fragments.gox:265
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
//line node_fragments.gox:275
}

type DynamicFragment struct {
	n1 doors.Door
	n2 doors.Door
	test.NoBeam
}

//line node_fragments.gox:283
func (f *DynamicFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:285
		f.n1.Inner(ctx, test.Marker("init"))

//line node_fragments.gox:288
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
//line node_fragments.gox:291
			__e = __c.Set("id", "update"); if __e != nil { return }
//line node_fragments.gox:292
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
//line node_fragments.gox:301
			__e = __c.Set("id", "replace"); if __e != nil { return }
//line node_fragments.gox:302
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
//line node_fragments.gox:312
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:313
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
//line node_fragments.gox:321
}

type BeforeFragment struct {
	doorInit doors.Door
	doorUpdate doors.Door
	doorRemoved doors.Door
	doorReplaced doors.Door
	test.NoBeam
}

//line node_fragments.gox:331
func (f *BeforeFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:332
		__e = (f.doorInit).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:333
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:337
		f.doorUpdate.Inner(ctx, test.Marker("updated"))

//line node_fragments.gox:339
		__e = (f.doorUpdate).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:343
		f.doorRemoved.Inner(ctx, test.Marker("removed"))

//line node_fragments.gox:346
		f.doorRemoved.Static(ctx, nil)

//line node_fragments.gox:348
		__e = __c.Any(&f.doorRemoved); if __e != nil { return }
//line node_fragments.gox:351
		f.doorReplaced.Static(ctx, test.Marker("replaced"))

//line node_fragments.gox:354
		__e = (f.doorReplaced).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:355
				__e = __c.Any(test.Marker("initReplaced")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:357
}

type LifeCycleFragment struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

//line node_fragments.gox:365
func (f *LifeCycleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:366
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:366
			__e = __c.Any(f.initial()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:368
			__e = __c.Set("id", "reload"); if __e != nil { return }
//line node_fragments.gox:369
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
//line node_fragments.gox:378
			__e = __c.Set("id", "updateEmpty"); if __e != nil { return }
//line node_fragments.gox:379
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
//line node_fragments.gox:388
			__e = __c.Set("id", "updateEmptyAlt"); if __e != nil { return }
//line node_fragments.gox:389
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
//line node_fragments.gox:398
			__e = __c.Set("id", "updateContent"); if __e != nil { return }
//line node_fragments.gox:399
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
//line node_fragments.gox:408
			__e = __c.Set("id", "updateInner"); if __e != nil { return }
//line node_fragments.gox:409
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
//line node_fragments.gox:418
			__e = __c.Set("id", "updateOuter"); if __e != nil { return }
//line node_fragments.gox:419
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
//line node_fragments.gox:428
			__e = __c.Set("id", "replaceStatic"); if __e != nil { return }
//line node_fragments.gox:429
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
//line node_fragments.gox:438
			__e = __c.Set("id", "updateEditor"); if __e != nil { return }
//line node_fragments.gox:439
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
//line node_fragments.gox:448
			__e = __c.Set("id", "clear"); if __e != nil { return }
//line node_fragments.gox:449
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
//line node_fragments.gox:458
			__e = __c.Set("id", "unmount"); if __e != nil { return }
//line node_fragments.gox:459
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
//line node_fragments.gox:468
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:469
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
//line node_fragments.gox:477
}

//line node_fragments.gox:479
func (f *LifeCycleFragment) initial() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:481
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:482
					__e = __c.Any(test.Marker("presist")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:485
}
//line node_fragments.gox:486
func (f *LifeCycleFragment) newEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:488
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:488
					__e = __c.Set("id", "new"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:491
}

//line node_fragments.gox:493
func (f *LifeCycleFragment) newEmptyAlt() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:495
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("section"); if __e != nil { return }
				{
//line node_fragments.gox:495
					__e = __c.Set("id", "new-alt"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:498
}

//line node_fragments.gox:500
func (f *LifeCycleFragment) newContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:502
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:502
					__e = __c.Set("id", "new2"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:503
					__e = __c.Any(test.Marker("presist2")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:506
}

//line node_fragments.gox:508
func (f *LifeCycleFragment) newOuter() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:509
			__e = __c.Set("id", "outer-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:510
			__e = __c.Any(test.Marker("outer-presist")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:512
}

//line node_fragments.gox:514
func (f *LifeCycleFragment) newEditor() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:516
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:518
}

type FragmentProxyReloadContent struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

//line node_fragments.gox:526
func (f *FragmentProxyReloadContent) mountEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:527
			__e = __c.Set("id", "proxy-redraw-frame"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:528
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:528
					__e = __c.Set("id", "proxy-redraw-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:531
}

//line node_fragments.gox:533
func (f *FragmentProxyReloadContent) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:534
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:534
			__e = __c.Any(f.mountEmpty()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:536
			__e = __c.Set("id", "proxy-redraw-update"); if __e != nil { return }
//line node_fragments.gox:537
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
//line node_fragments.gox:546
			__e = __c.Set("id", "proxy-redraw-remount"); if __e != nil { return }
//line node_fragments.gox:547
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
//line node_fragments.gox:556
			__e = __c.Set("id", "proxy-redraw-reload"); if __e != nil { return }
//line node_fragments.gox:557
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
//line node_fragments.gox:565
}

type FragmentClosestReload struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:575
func (f *FragmentClosestReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:577
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:579
			__e = __c.Set("id", "inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:579
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:581
			__e = __c.Set("id", "reload-nearest"); if __e != nil { return }
//line node_fragments.gox:582
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
//line node_fragments.gox:590
}

//line node_fragments.gox:592
func (f *FragmentClosestReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:594
		f.outerRenders++
		f.node.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:597
			__e = __c.Set("id", "outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:597
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:598
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:599
}

//line node_fragments.gox:601
func (f *FragmentClosestReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:603
		f.frame.Inner(ctx, f.outerContent())

//line node_fragments.gox:605
		__e = __c.Any(&f.frame); if __e != nil { return }
	return })
//line node_fragments.gox:606
}

type FragmentClosestReloadProxy struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:616
func (f *FragmentClosestReloadProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:617
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:617
				__e = __c.Set("id", "outer-proxy-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:619
				f.outerRenders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:621
					__e = __c.Set("id", "proxy-outer-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:621
					__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:622
				__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line node_fragments.gox:622
						__e = __c.Set("id", "inner-proxy-root"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:624
						f.innerRenders++

						__e = __c.Init("div"); if __e != nil { return }
						{
//line node_fragments.gox:626
							__e = __c.Set("id", "proxy-inner-count"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:626
							__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("button"); if __e != nil { return }
						{
//line node_fragments.gox:628
							__e = __c.Set("id", "reload-nearest-proxy"); if __e != nil { return }
//line node_fragments.gox:629
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
//line node_fragments.gox:639
}

type FragmentInlineDoorPointerProxy struct {
	renders int
	test.NoBeam
}

//line node_fragments.gox:646
func (f *FragmentInlineDoorPointerProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:647
		__e = (&doors.Door{}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:647
				__e = __c.Set("id", "inline-door-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:649
				f.renders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:651
					__e = __c.Set("id", "inline-door-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:651
					__e = __c.Any(fmt.Sprintf("inline-%d", f.renders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line node_fragments.gox:653
					__e = __c.Set("id", "inline-door-reload"); if __e != nil { return }
//line node_fragments.gox:654
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
//line node_fragments.gox:663
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

//line node_fragments.gox:692
func (f *FragmentClosestXReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:694
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:696
			__e = __c.Set("id", "x-inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:696
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:698
			__e = __c.Set("id", "xreload-nearest"); if __e != nil { return }
//line node_fragments.gox:699
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
//line node_fragments.gox:706
}

//line node_fragments.gox:708
func (f *FragmentClosestXReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:710
		f.outerRenders++
		f.node.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:713
			__e = __c.Set("id", "x-outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:713
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:714
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:715
}

//line node_fragments.gox:717
func (f *FragmentClosestXReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:719
		f.frame.Inner(ctx, f.outerContent())

//line node_fragments.gox:721
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:722
		__e = __c.Any(&f.report); if __e != nil { return }
	return })
//line node_fragments.gox:723
}

type FragmentRootXReload struct {
	report doors.Door
	test.NoBeam
}

func (f *FragmentRootXReload) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

//line node_fragments.gox:734
func (f *FragmentRootXReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:735
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:737
			__e = __c.Set("id", "root-xreload"); if __e != nil { return }
//line node_fragments.gox:738
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
//line node_fragments.gox:755
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

//line node_fragments.gox:782
func (f *FragmentDetachedReplace) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:783
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:784
}

//line node_fragments.gox:786
func (f *FragmentDetachedReplace) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:788
		f.node.Inner(ctx, test.Marker("replace-base"))
		f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:791
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:792
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:793
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XStatic(ctx, test.Marker("replace-detached")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:797
				__e = __c.Set("id", "replace-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:798
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:802
				__e = __c.Set("id", "reload-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:803
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XInner(ctx, test.Marker("replace-updated")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:807
				__e = __c.Set("id", "update-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:808
		__e = __c.Any(test.Button("remount-after-replace", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:812
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

//line node_fragments.gox:839
func (f *FragmentDetachedRebase) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:840
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:841
}

//line node_fragments.gox:843
func (f *FragmentDetachedRebase) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:844
			__e = __c.Set("id", "rebased-detached-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:845
			__e = __c.Any(test.Marker("rebased-detached")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:847
}

//line node_fragments.gox:849
func (f *FragmentDetachedRebase) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:851
		f.node.Inner(ctx, test.Marker("rebase-base"))
		f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:854
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:855
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:856
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUnmount(ctx), "ok unmount")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:860
				__e = __c.Set("id", "unmount-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("unmount-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:861
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:865
				__e = __c.Set("id", "reload-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:866
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XOuter(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:870
				__e = __c.Set("id", "rebase-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:871
		__e = __c.Any(test.Button("remount-after-rebase", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:875
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

//line node_fragments.gox:903
func (f *FragmentProxyMove) mountFrame1() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:904
			__e = __c.Set("id", "frame1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:905
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:907
}

//line node_fragments.gox:909
func (f *FragmentProxyMove) mountFrame2() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:910
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:911
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:913
}

//line node_fragments.gox:915
func (f *FragmentProxyMove) frame2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:916
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:917
}

//line node_fragments.gox:919
func (f *FragmentProxyMove) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:920
			__e = __c.Set("id", "proxy-moved-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:921
			__e = __c.Any(test.Marker("proxy-moved")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:923
}

//line node_fragments.gox:925
func (f *FragmentProxyMove) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:927
		f.node.Inner(ctx, test.Marker("proxy-base"))
		f.frame1.Inner(ctx, f.mountFrame1())
		f.frame2.Inner(ctx, f.frame2Empty())

//line node_fragments.gox:931
		__e = __c.Any(&f.frame1); if __e != nil { return }
//line node_fragments.gox:932
		__e = __c.Any(&f.frame2); if __e != nil { return }
//line node_fragments.gox:933
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:934
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XOuter(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:938
				__e = __c.Set("id", "rebase-proxy-move"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-proxy-move"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:939
		__e = __c.Any(test.Button("move-proxy", func(ctx context.Context) bool {
		f.frame2.Inner(ctx, f.mountFrame2())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:943
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

//line node_fragments.gox:972
func (f *FragmentHierarchy) childBody() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("article"); if __e != nil { return }
		{
//line node_fragments.gox:973
			__e = __c.Set("id", "child-body"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:974
			__e = __c.Any(&f.grand); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:976
}

//line node_fragments.gox:978
func (f *FragmentHierarchy) host1Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:979
			__e = __c.Set("id", "host1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:980
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:982
}

//line node_fragments.gox:984
func (f *FragmentHierarchy) host2Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:985
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:986
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:988
}

//line node_fragments.gox:990
func (f *FragmentHierarchy) host2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:991
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:992
}

//line node_fragments.gox:994
func (f *FragmentHierarchy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:996
		f.grand.Inner(ctx, test.Marker("grand-init"))
		f.child.Inner(ctx, f.childBody())
		f.host1.Inner(ctx, f.host1Body())
		f.host2.Inner(ctx, f.host2Empty())

//line node_fragments.gox:1001
		__e = __c.Any(&f.host1); if __e != nil { return }
//line node_fragments.gox:1002
		__e = __c.Any(&f.host2); if __e != nil { return }
//line node_fragments.gox:1003
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1004
		__e = __c.Any(test.Button("move-child", func(ctx context.Context) bool {
		f.host2.Inner(ctx, f.host2Body())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1008
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.grand.XInner(ctx, test.Marker("grand-updated")), "ok grand")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1012
				__e = __c.Set("id", "grand-update"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("grand-update"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1013
		__e = __c.Any(test.Button("remove-host2", func(ctx context.Context) bool {
		f.host2.Static(ctx, nil)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1017
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

//line node_fragments.gox:1050
func (f *FragmentErrorTransitions) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1051
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1052
}

//line node_fragments.gox:1054
func (f *FragmentErrorTransitions) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1056
		f.node.Inner(ctx, test.Marker("error-base"))
		f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1059
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1060
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1061
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XInner(ctx, f.errElem("update boom")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1065
				__e = __c.Set("id", "update-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1066
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XStatic(ctx, f.errElem("replace boom")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1070
				__e = __c.Set("id", "replace-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1071
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XOuter(ctx, f.errElem("rebase boom")), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1075
				__e = __c.Set("id", "rebase-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:1076
}
