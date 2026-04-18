// Managed by GoX v0.1.25

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
			__e = __c.AttrSet("class", "sample"); if __e != nil { return }
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
				__e = __c.AttrSet("style", "display:none"); if __e != nil { return }
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
				__e = __c.AttrSet("style", "display:none"); if __e != nil { return }
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
		__e = n.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
			if err := cur.AttrSet("id", "proxy-wrap-first"); err != nil {
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
			if err := cur.AttrSet("id", "proxy-wrap-second"); err != nil {
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
		__e = f.n.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:100
			for i := range 2 {
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:101
					__e = __c.AttrSet("id", fmt.Sprintf("proxy-loop-%d", i)); if __e != nil { return }
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
		__e = f.n.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:117
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:119
		__e = f.report.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:119
			__e = doors.AClick{
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
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("button"); if __e != nil { return }
				{
//line node_fragments.gox:134
					__e = __c.AttrSet("id", "updatex"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("C"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:136
		__e = doors.AClick{
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
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:151
				__e = __c.AttrSet("id", "removex"); if __e != nil { return }
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
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XReload(ctx), "ok reload")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:194
				__e = __c.AttrSet("id", "xreload"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreload"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:195
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XRebase(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("section"); if __e != nil { return }
				{
//line node_fragments.gox:197
					__e = __c.AttrSet("id", "x-rebased-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:198
					__e = __c.Any(test.Marker("x-rebased")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:199
			return })), "ok rebase")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:201
				__e = __c.AttrSet("id", "xrebase"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xrebase"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:202
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XClear(ctx), "ok clear")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:206
				__e = __c.AttrSet("id", "xclear"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xclear"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:207
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XUpdate(ctx, test.Marker("x-updated")), "ok update")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:211
				__e = __c.AttrSet("id", "xupdate"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xupdate"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:212
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XUnmount(ctx), "ok unmount")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:216
				__e = __c.AttrSet("id", "xunmount"); if __e != nil { return }
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
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XReplace(ctx, test.Marker("x-replaced")), "ok replace")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:225
				__e = __c.AttrSet("id", "xreplace"); if __e != nil { return }
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
		__e = f.n1.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:237
				__e = f.n2.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
			__e = __c.AttrSet("id", "remove"); if __e != nil { return }
//line node_fragments.gox:245
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "clear"); if __e != nil { return }
//line node_fragments.gox:255
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "replace"); if __e != nil { return }
//line node_fragments.gox:265
			__e = __c.AttrMod(doors.AClick{
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
		__e = f.n1.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
			__e = __c.AttrSet("id", "update"); if __e != nil { return }
//line node_fragments.gox:292
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "replace"); if __e != nil { return }
//line node_fragments.gox:302
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "remove"); if __e != nil { return }
//line node_fragments.gox:313
			__e = __c.AttrMod(doors.AClick{
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
		__e = f.doorInit.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
		__e = f.doorUpdate.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
		__e = f.doorReplaced.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
		__e = f.frame.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:366
			__e = __c.Any(f.initial()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:368
			__e = __c.AttrSet("id", "reload"); if __e != nil { return }
//line node_fragments.gox:369
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "updateEmpty"); if __e != nil { return }
//line node_fragments.gox:379
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "updateContent"); if __e != nil { return }
//line node_fragments.gox:389
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "updateEditor"); if __e != nil { return }
//line node_fragments.gox:399
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "clear"); if __e != nil { return }
//line node_fragments.gox:409
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "unmount"); if __e != nil { return }
//line node_fragments.gox:419
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "remove"); if __e != nil { return }
//line node_fragments.gox:429
			__e = __c.AttrMod(doors.AClick{
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
			__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
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
			__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:448
					__e = __c.AttrSet("id", "new"); if __e != nil { return }
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
			__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:455
					__e = __c.AttrSet("id", "new2"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "proxy-redraw-frame"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:475
			__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:475
					__e = __c.AttrSet("id", "proxy-redraw-root"); if __e != nil { return }
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
		__e = f.frame.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:481
			__e = __c.Any(f.mountEmpty()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:483
			__e = __c.AttrSet("id", "proxy-redraw-update"); if __e != nil { return }
//line node_fragments.gox:484
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "proxy-redraw-remount"); if __e != nil { return }
//line node_fragments.gox:494
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "proxy-redraw-reload"); if __e != nil { return }
//line node_fragments.gox:504
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:526
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:528
			__e = __c.AttrSet("id", "reload-nearest"); if __e != nil { return }
//line node_fragments.gox:529
			__e = __c.AttrMod(doors.AClick{
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
			__e = __c.AttrSet("id", "outer-count"); if __e != nil { return }
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
		__e = f.frame.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:564
				__e = __c.AttrSet("id", "outer-proxy-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:566
				f.outerRenders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:568
					__e = __c.AttrSet("id", "proxy-outer-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:568
					__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:569
				__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line node_fragments.gox:569
						__e = __c.AttrSet("id", "inner-proxy-root"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:571
						f.innerRenders++

						__e = __c.Init("div"); if __e != nil { return }
						{
//line node_fragments.gox:573
							__e = __c.AttrSet("id", "proxy-inner-count"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:573
							__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("button"); if __e != nil { return }
						{
//line node_fragments.gox:575
							__e = __c.AttrSet("id", "reload-nearest-proxy"); if __e != nil { return }
//line node_fragments.gox:576
							__e = __c.AttrMod(doors.AClick{
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

//line node_fragments.gox:615
func (f *FragmentClosestXReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:617
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:619
			__e = __c.AttrSet("id", "x-inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:619
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:621
			__e = __c.AttrSet("id", "xreload-nearest"); if __e != nil { return }
//line node_fragments.gox:622
			__e = __c.AttrMod(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				return f.wait(ctx, doors.XReload(ctx), "ok xreload")
			},
		}); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("xreload-nearest"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:629
}

//line node_fragments.gox:631
func (f *FragmentClosestXReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:633
		f.outerRenders++
		f.node.Update(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:636
			__e = __c.AttrSet("id", "x-outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:636
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:637
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:638
}

//line node_fragments.gox:640
func (f *FragmentClosestXReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:642
		f.frame.Update(ctx, f.outerContent())

//line node_fragments.gox:644
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:645
		__e = __c.Any(&f.report); if __e != nil { return }
	return })
//line node_fragments.gox:646
}

type FragmentRootXReload struct {
	report doors.Door
	test.NoBeam
}

func (f *FragmentRootXReload) rep(ctx context.Context, s string) {
	f.report.Update(ctx, test.Report(s))
}

//line node_fragments.gox:657
func (f *FragmentRootXReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:658
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:660
			__e = __c.AttrSet("id", "root-xreload"); if __e != nil { return }
//line node_fragments.gox:661
			__e = __c.AttrMod(doors.AClick{
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
//line node_fragments.gox:678
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

//line node_fragments.gox:705
func (f *FragmentDetachedReplace) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:706
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:707
}

//line node_fragments.gox:709
func (f *FragmentDetachedReplace) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:711
		f.node.Update(ctx, test.Marker("replace-base"))
		f.frame.Update(ctx, f.mount())

//line node_fragments.gox:714
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:715
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:716
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReplace(ctx, test.Marker("replace-detached")), "ok replace")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:720
				__e = __c.AttrSet("id", "replace-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:721
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:725
				__e = __c.AttrSet("id", "reload-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:726
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUpdate(ctx, test.Marker("replace-updated")), "ok update")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:730
				__e = __c.AttrSet("id", "update-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:731
		__e = __c.Any(test.Button("remount-after-replace", func(ctx context.Context) bool {
		f.frame.Update(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:735
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

//line node_fragments.gox:762
func (f *FragmentDetachedRebase) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:763
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:764
}

//line node_fragments.gox:766
func (f *FragmentDetachedRebase) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:767
			__e = __c.AttrSet("id", "rebased-detached-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:768
			__e = __c.Any(test.Marker("rebased-detached")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:770
}

//line node_fragments.gox:772
func (f *FragmentDetachedRebase) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:774
		f.node.Update(ctx, test.Marker("rebase-base"))
		f.frame.Update(ctx, f.mount())

//line node_fragments.gox:777
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:778
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:779
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUnmount(ctx), "ok unmount")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:783
				__e = __c.AttrSet("id", "unmount-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("unmount-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:784
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:788
				__e = __c.AttrSet("id", "reload-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:789
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XRebase(ctx, f.rebased()), "ok rebase")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:793
				__e = __c.AttrSet("id", "rebase-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:794
		__e = __c.Any(test.Button("remount-after-rebase", func(ctx context.Context) bool {
		f.frame.Update(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:798
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

//line node_fragments.gox:826
func (f *FragmentProxyMove) mountFrame1() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:827
			__e = __c.AttrSet("id", "frame1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:828
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:830
}

//line node_fragments.gox:832
func (f *FragmentProxyMove) mountFrame2() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:833
			__e = __c.AttrSet("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:834
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:836
}

//line node_fragments.gox:838
func (f *FragmentProxyMove) frame2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:839
			__e = __c.AttrSet("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:840
}

//line node_fragments.gox:842
func (f *FragmentProxyMove) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:843
			__e = __c.AttrSet("id", "proxy-moved-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:844
			__e = __c.Any(test.Marker("proxy-moved")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:846
}

//line node_fragments.gox:848
func (f *FragmentProxyMove) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:850
		f.node.Update(ctx, test.Marker("proxy-base"))
		f.frame1.Update(ctx, f.mountFrame1())
		f.frame2.Update(ctx, f.frame2Empty())

//line node_fragments.gox:854
		__e = __c.Any(&f.frame1); if __e != nil { return }
//line node_fragments.gox:855
		__e = __c.Any(&f.frame2); if __e != nil { return }
//line node_fragments.gox:856
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:857
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XRebase(ctx, f.rebased()), "ok rebase")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:861
				__e = __c.AttrSet("id", "rebase-proxy-move"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-proxy-move"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:862
		__e = __c.Any(test.Button("move-proxy", func(ctx context.Context) bool {
		f.frame2.Update(ctx, f.mountFrame2())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:866
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

//line node_fragments.gox:895
func (f *FragmentHierarchy) childBody() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("article"); if __e != nil { return }
		{
//line node_fragments.gox:896
			__e = __c.AttrSet("id", "child-body"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:897
			__e = __c.Any(&f.grand); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:899
}

//line node_fragments.gox:901
func (f *FragmentHierarchy) host1Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:902
			__e = __c.AttrSet("id", "host1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:903
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:905
}

//line node_fragments.gox:907
func (f *FragmentHierarchy) host2Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:908
			__e = __c.AttrSet("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:909
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:911
}

//line node_fragments.gox:913
func (f *FragmentHierarchy) host2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:914
			__e = __c.AttrSet("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:915
}

//line node_fragments.gox:917
func (f *FragmentHierarchy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:919
		f.grand.Update(ctx, test.Marker("grand-init"))
		f.child.Update(ctx, f.childBody())
		f.host1.Update(ctx, f.host1Body())
		f.host2.Update(ctx, f.host2Empty())

//line node_fragments.gox:924
		__e = __c.Any(&f.host1); if __e != nil { return }
//line node_fragments.gox:925
		__e = __c.Any(&f.host2); if __e != nil { return }
//line node_fragments.gox:926
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:927
		__e = __c.Any(test.Button("move-child", func(ctx context.Context) bool {
		f.host2.Update(ctx, f.host2Body())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:931
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.grand.XUpdate(ctx, test.Marker("grand-updated")), "ok grand")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:935
				__e = __c.AttrSet("id", "grand-update"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("grand-update"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:936
		__e = __c.Any(test.Button("remove-host2", func(ctx context.Context) bool {
		f.host2.Delete(ctx)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:940
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

//line node_fragments.gox:973
func (f *FragmentErrorTransitions) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:974
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:975
}

//line node_fragments.gox:977
func (f *FragmentErrorTransitions) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:979
		f.node.Update(ctx, test.Marker("error-base"))
		f.frame.Update(ctx, f.mount())

//line node_fragments.gox:982
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:983
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:984
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUpdate(ctx, f.errElem("update boom")), "ok update")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:988
				__e = __c.AttrSet("id", "update-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:989
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReplace(ctx, f.errElem("replace boom")), "ok replace")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:993
				__e = __c.AttrSet("id", "replace-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:994
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XRebase(ctx, f.errElem("rebase boom")), "ok rebase")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:998
				__e = __c.AttrSet("id", "rebase-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:999
}
