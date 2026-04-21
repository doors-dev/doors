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
		f.n.Replace(ctx, f.sample())

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
		f.n.Update(ctx, f.sample())
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
		n.Update(ctx, f.replaced())
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
	f.report.Update(ctx, test.Report(s))
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
			ch := f.n.XUpdate(ctx, test.Marker("updated"))
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
			ch := f.n.XDelete(ctx)
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
	f.report.Update(ctx, test.Report(s))
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
		f.n.Update(ctx, test.Marker("x-init"))
		f.frame.Update(ctx, f.mount())

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
			return f.wait(ctx, f.n.XRebase(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
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
			return f.wait(ctx, f.n.XClear(ctx), "ok clear")
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
			return f.wait(ctx, f.n.XUpdate(ctx, test.Marker("x-updated")), "ok update")
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
		f.frame.Update(ctx, f.mount())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:221
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XReplace(ctx, test.Marker("x-replaced")), "ok replace")
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
				f.n2.Delete(ctx)
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
				f.n1.Clear(ctx)
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
				f.n2.Update(ctx, test.Marker("replaced"))
				f.n3.Update(ctx, test.Marker("temp"))
				f.n3.Replace(ctx, &f.n2)
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
		f.n1.Update(ctx, test.Marker("init"))

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
				f.n1.Update(ctx, test.Marker("updated"))
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
				f.n2.Update(ctx, test.Marker("replaced"))
				f.n1.Replace(ctx, &f.n2)
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
				f.n2.Delete(ctx)
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
		f.doorUpdate.Update(ctx, test.Marker("updated"))

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
		f.doorRemoved.Update(ctx, test.Marker("removed"))

//line node_fragments.gox:346
		f.doorRemoved.Delete(ctx)

//line node_fragments.gox:348
		__e = __c.Any(&f.doorRemoved); if __e != nil { return }
//line node_fragments.gox:351
		f.doorReplaced.Replace(ctx, test.Marker("replaced"))

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
				f.frame.Update(ctx, f.newEmpty())
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
			__e = __c.Set("id", "updateContent"); if __e != nil { return }
//line node_fragments.gox:389
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.frame.Update(ctx, f.newContent())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Update2"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:398
			__e = __c.Set("id", "updateEditor"); if __e != nil { return }
//line node_fragments.gox:399
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.frame.Update(ctx, f.newEditor())
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
			__e = __c.Set("id", "clear"); if __e != nil { return }
//line node_fragments.gox:409
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Clear(ctx)
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Clear"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:418
			__e = __c.Set("id", "unmount"); if __e != nil { return }
//line node_fragments.gox:419
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
//line node_fragments.gox:428
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:429
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Delete(ctx)
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("Remove"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:437
}

//line node_fragments.gox:439
func (f *LifeCycleFragment) initial() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:441
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:442
					__e = __c.Any(test.Marker("presist")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:445
}
//line node_fragments.gox:446
func (f *LifeCycleFragment) newEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:448
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:448
					__e = __c.Set("id", "new"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:451
}

//line node_fragments.gox:453
func (f *LifeCycleFragment) newContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:455
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:455
					__e = __c.Set("id", "new2"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:456
					__e = __c.Any(test.Marker("presist2")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:459
}

//line node_fragments.gox:461
func (f *LifeCycleFragment) newEditor() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:463
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:465
}

type FragmentProxyReloadContent struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

//line node_fragments.gox:473
func (f *FragmentProxyReloadContent) mountEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:474
			__e = __c.Set("id", "proxy-redraw-frame"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:475
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:475
					__e = __c.Set("id", "proxy-redraw-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:478
}

//line node_fragments.gox:480
func (f *FragmentProxyReloadContent) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:481
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:481
			__e = __c.Any(f.mountEmpty()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:483
			__e = __c.Set("id", "proxy-redraw-update"); if __e != nil { return }
//line node_fragments.gox:484
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Update(ctx, test.Marker("proxy-redraw-content"))
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-redraw-update"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:493
			__e = __c.Set("id", "proxy-redraw-remount"); if __e != nil { return }
//line node_fragments.gox:494
			__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.frame.Update(ctx, f.mountEmpty())
				return false
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("proxy-redraw-remount"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:503
			__e = __c.Set("id", "proxy-redraw-reload"); if __e != nil { return }
//line node_fragments.gox:504
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
//line node_fragments.gox:512
}

type FragmentClosestReload struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:522
func (f *FragmentClosestReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:524
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:526
			__e = __c.Set("id", "inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:526
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:528
			__e = __c.Set("id", "reload-nearest"); if __e != nil { return }
//line node_fragments.gox:529
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
//line node_fragments.gox:537
}

//line node_fragments.gox:539
func (f *FragmentClosestReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:541
		f.outerRenders++
		f.node.Update(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:544
			__e = __c.Set("id", "outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:544
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:545
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:546
}

//line node_fragments.gox:548
func (f *FragmentClosestReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:550
		f.frame.Update(ctx, f.outerContent())

//line node_fragments.gox:552
		__e = __c.Any(&f.frame); if __e != nil { return }
	return })
//line node_fragments.gox:553
}

type FragmentClosestReloadProxy struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:563
func (f *FragmentClosestReloadProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:564
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:564
				__e = __c.Set("id", "outer-proxy-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:566
				f.outerRenders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:568
					__e = __c.Set("id", "proxy-outer-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:568
					__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:569
				__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line node_fragments.gox:569
						__e = __c.Set("id", "inner-proxy-root"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:571
						f.innerRenders++

						__e = __c.Init("div"); if __e != nil { return }
						{
//line node_fragments.gox:573
							__e = __c.Set("id", "proxy-inner-count"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:573
							__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("button"); if __e != nil { return }
						{
//line node_fragments.gox:575
							__e = __c.Set("id", "reload-nearest-proxy"); if __e != nil { return }
//line node_fragments.gox:576
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
//line node_fragments.gox:586
}

type FragmentInlineDoorPointerProxy struct {
	renders int
	test.NoBeam
}

//line node_fragments.gox:593
func (f *FragmentInlineDoorPointerProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:594
		__e = (&doors.Door{}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:594
				__e = __c.Set("id", "inline-door-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:596
				f.renders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:598
					__e = __c.Set("id", "inline-door-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:598
					__e = __c.Any(fmt.Sprintf("inline-%d", f.renders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line node_fragments.gox:600
					__e = __c.Set("id", "inline-door-reload"); if __e != nil { return }
//line node_fragments.gox:601
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
//line node_fragments.gox:610
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
	f.report.Update(ctx, test.Report(s))
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

//line node_fragments.gox:639
func (f *FragmentClosestXReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:641
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:643
			__e = __c.Set("id", "x-inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:643
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:645
			__e = __c.Set("id", "xreload-nearest"); if __e != nil { return }
//line node_fragments.gox:646
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
//line node_fragments.gox:653
}

//line node_fragments.gox:655
func (f *FragmentClosestXReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:657
		f.outerRenders++
		f.node.Update(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:660
			__e = __c.Set("id", "x-outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:660
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:661
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:662
}

//line node_fragments.gox:664
func (f *FragmentClosestXReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:666
		f.frame.Update(ctx, f.outerContent())

//line node_fragments.gox:668
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:669
		__e = __c.Any(&f.report); if __e != nil { return }
	return })
//line node_fragments.gox:670
}

type FragmentRootXReload struct {
	report doors.Door
	test.NoBeam
}

func (f *FragmentRootXReload) rep(ctx context.Context, s string) {
	f.report.Update(ctx, test.Report(s))
}

//line node_fragments.gox:681
func (f *FragmentRootXReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:682
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:684
			__e = __c.Set("id", "root-xreload"); if __e != nil { return }
//line node_fragments.gox:685
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
//line node_fragments.gox:702
}

type FragmentDetachedReplace struct {
	report doors.Door
	frame doors.Door
	node doors.Door
	test.NoBeam
}

func (f *FragmentDetachedReplace) rep(ctx context.Context, s string) {
	f.report.Update(ctx, test.Report(s))
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

//line node_fragments.gox:729
func (f *FragmentDetachedReplace) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:730
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:731
}

//line node_fragments.gox:733
func (f *FragmentDetachedReplace) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:735
		f.node.Update(ctx, test.Marker("replace-base"))
		f.frame.Update(ctx, f.mount())

//line node_fragments.gox:738
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:739
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:740
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReplace(ctx, test.Marker("replace-detached")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:744
				__e = __c.Set("id", "replace-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:745
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:749
				__e = __c.Set("id", "reload-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:750
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUpdate(ctx, test.Marker("replace-updated")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:754
				__e = __c.Set("id", "update-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:755
		__e = __c.Any(test.Button("remount-after-replace", func(ctx context.Context) bool {
		f.frame.Update(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:759
}

type FragmentDetachedRebase struct {
	report doors.Door
	frame doors.Door
	node doors.Door
	test.NoBeam
}

func (f *FragmentDetachedRebase) rep(ctx context.Context, s string) {
	f.report.Update(ctx, test.Report(s))
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

//line node_fragments.gox:786
func (f *FragmentDetachedRebase) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:787
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:788
}

//line node_fragments.gox:790
func (f *FragmentDetachedRebase) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:791
			__e = __c.Set("id", "rebased-detached-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:792
			__e = __c.Any(test.Marker("rebased-detached")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:794
}

//line node_fragments.gox:796
func (f *FragmentDetachedRebase) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:798
		f.node.Update(ctx, test.Marker("rebase-base"))
		f.frame.Update(ctx, f.mount())

//line node_fragments.gox:801
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:802
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:803
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUnmount(ctx), "ok unmount")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:807
				__e = __c.Set("id", "unmount-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("unmount-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:808
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:812
				__e = __c.Set("id", "reload-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:813
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XRebase(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:817
				__e = __c.Set("id", "rebase-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:818
		__e = __c.Any(test.Button("remount-after-rebase", func(ctx context.Context) bool {
		f.frame.Update(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:822
}

type FragmentProxyMove struct {
	report doors.Door
	frame1 doors.Door
	frame2 doors.Door
	node doors.Door
	test.NoBeam
}

func (f *FragmentProxyMove) rep(ctx context.Context, s string) {
	f.report.Update(ctx, test.Report(s))
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

//line node_fragments.gox:850
func (f *FragmentProxyMove) mountFrame1() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:851
			__e = __c.Set("id", "frame1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:852
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:854
}

//line node_fragments.gox:856
func (f *FragmentProxyMove) mountFrame2() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:857
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:858
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:860
}

//line node_fragments.gox:862
func (f *FragmentProxyMove) frame2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:863
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:864
}

//line node_fragments.gox:866
func (f *FragmentProxyMove) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:867
			__e = __c.Set("id", "proxy-moved-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:868
			__e = __c.Any(test.Marker("proxy-moved")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:870
}

//line node_fragments.gox:872
func (f *FragmentProxyMove) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:874
		f.node.Update(ctx, test.Marker("proxy-base"))
		f.frame1.Update(ctx, f.mountFrame1())
		f.frame2.Update(ctx, f.frame2Empty())

//line node_fragments.gox:878
		__e = __c.Any(&f.frame1); if __e != nil { return }
//line node_fragments.gox:879
		__e = __c.Any(&f.frame2); if __e != nil { return }
//line node_fragments.gox:880
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:881
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XRebase(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:885
				__e = __c.Set("id", "rebase-proxy-move"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-proxy-move"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:886
		__e = __c.Any(test.Button("move-proxy", func(ctx context.Context) bool {
		f.frame2.Update(ctx, f.mountFrame2())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:890
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
	f.report.Update(ctx, test.Report(s))
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

//line node_fragments.gox:919
func (f *FragmentHierarchy) childBody() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("article"); if __e != nil { return }
		{
//line node_fragments.gox:920
			__e = __c.Set("id", "child-body"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:921
			__e = __c.Any(&f.grand); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:923
}

//line node_fragments.gox:925
func (f *FragmentHierarchy) host1Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:926
			__e = __c.Set("id", "host1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:927
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:929
}

//line node_fragments.gox:931
func (f *FragmentHierarchy) host2Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:932
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:933
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:935
}

//line node_fragments.gox:937
func (f *FragmentHierarchy) host2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:938
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:939
}

//line node_fragments.gox:941
func (f *FragmentHierarchy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:943
		f.grand.Update(ctx, test.Marker("grand-init"))
		f.child.Update(ctx, f.childBody())
		f.host1.Update(ctx, f.host1Body())
		f.host2.Update(ctx, f.host2Empty())

//line node_fragments.gox:948
		__e = __c.Any(&f.host1); if __e != nil { return }
//line node_fragments.gox:949
		__e = __c.Any(&f.host2); if __e != nil { return }
//line node_fragments.gox:950
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:951
		__e = __c.Any(test.Button("move-child", func(ctx context.Context) bool {
		f.host2.Update(ctx, f.host2Body())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:955
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.grand.XUpdate(ctx, test.Marker("grand-updated")), "ok grand")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:959
				__e = __c.Set("id", "grand-update"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("grand-update"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:960
		__e = __c.Any(test.Button("remove-host2", func(ctx context.Context) bool {
		f.host2.Delete(ctx)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:964
}

type FragmentErrorTransitions struct {
	report doors.Door
	frame doors.Door
	node doors.Door
	test.NoBeam
}

func (f *FragmentErrorTransitions) rep(ctx context.Context, s string) {
	f.report.Update(ctx, test.Report(s))
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

//line node_fragments.gox:997
func (f *FragmentErrorTransitions) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:998
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:999
}

//line node_fragments.gox:1001
func (f *FragmentErrorTransitions) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1003
		f.node.Update(ctx, test.Marker("error-base"))
		f.frame.Update(ctx, f.mount())

//line node_fragments.gox:1006
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1007
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1008
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUpdate(ctx, f.errElem("update boom")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1012
				__e = __c.Set("id", "update-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1013
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReplace(ctx, f.errElem("replace boom")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1017
				__e = __c.Set("id", "replace-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1018
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XRebase(ctx, f.errElem("rebase boom")), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1022
				__e = __c.Set("id", "rebase-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:1023
}
