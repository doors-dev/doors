// Managed by GoX v0.1.17+dirty

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

func (f *FragmentMany) sample() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.AttrSet("class", "sample"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("sample"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentMany) manyDoors() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		for i := range 20 {
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.AttrSet("style", "display:none"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Any(&f.n); if __e != nil { return }
		}
	return })
}

func (f *FragmentMany) replaced() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.n.Replace(ctx, f.sample())

		for i := range 100 {
			__e = __c.Init("span"); if __e != nil { return }
			{
				__e = __c.AttrSet("style", "display:none"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Any(&f.n); if __e != nil { return }
		}
	return })
}

func (f *FragmentMany) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.n.Update(ctx, f.sample())
		n := doors.Door{}

		__e = n.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				__e = __c.Any(f.manyDoors()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(test.Button("replace", func(ctx context.Context) bool {
		n.Update(ctx, f.replaced())
		return true
	})); if __e != nil { return }
	return })
}

type FragmentX struct {
	report doors.Door
	n doors.Door
	test.NoBeam
}

func (f *FragmentX) rep(ctx context.Context, s string) {
	f.report.Update(ctx, test.Report(s))
}

func (f *FragmentX) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = f.n.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = f.report.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
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
				ctx := __c.Context(); gox.Noop(ctx)
				__e = __c.Init("button"); if __e != nil { return }
				{
					__e = __c.AttrSet("id", "updatex"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("C"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
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
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "removex"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("R"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
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
		f.rep(ctx, "channel err: "+err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

func (f *FragmentXDoor) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(&f.n); if __e != nil { return }
	return })
}

func (f *FragmentXDoor) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.n.Update(ctx, test.Marker("x-init"))
		f.frame.Update(ctx, f.mount())

		__e = __c.Any(&f.frame); if __e != nil { return }
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XReload(ctx), "ok reload")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "xreload"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreload"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XRebase(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); gox.Noop(ctx)
				__e = __c.Init("section"); if __e != nil { return }
				{
					__e = __c.AttrSet("id", "x-rebased-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Any(test.Marker("x-rebased")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })), "ok rebase")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "xrebase"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xrebase"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XClear(ctx), "ok clear")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "xclear"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xclear"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XUpdate(ctx, test.Marker("x-updated")), "ok update")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "xupdate"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xupdate"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XUnmount(ctx), "ok unmount")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "xunmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xunmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(test.Button("xremount", func(ctx context.Context) bool {
		f.frame.Update(ctx, f.mount())
		return false
	})); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.XReplace(ctx, test.Marker("x-replaced")), "ok replace")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "xreplace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreplace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
}

type EmbeddedFragment struct {
	n1 doors.Door
	n2 doors.Door
	n3 doors.Door
	test.NoBeam
}

func (f *EmbeddedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = f.n1.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.InitContainer(); if __e != nil { return }
			{
				__e = f.n2.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); gox.Noop(ctx)
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
						__e = __c.Any(test.Marker("init")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
				__e = __c.Any(test.Marker("static")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(&f.n3); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "remove"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "clear"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "replace"); if __e != nil { return }
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
}

type DynamicFragment struct {
	n1 doors.Door
	n2 doors.Door
	test.NoBeam
}

func (f *DynamicFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.n1.Update(ctx, test.Marker("init"))

		__e = f.n1.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "update"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "replace"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "remove"); if __e != nil { return }
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
}

type BeforeFragment struct {
	doorInit doors.Door
	doorUpdate doors.Door
	doorRemoved doors.Door
	doorReplaced doors.Door
	test.NoBeam
}

func (f *BeforeFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = f.doorInit.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		f.doorUpdate.Update(ctx, test.Marker("updated"))

		__e = f.doorUpdate.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		f.doorRemoved.Update(ctx, test.Marker("removed"))

		f.doorRemoved.Delete(ctx)

		__e = __c.Any(&f.doorRemoved); if __e != nil { return }
		f.doorReplaced.Replace(ctx, test.Marker("replaced"))

		__e = f.doorReplaced.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Any(test.Marker("initReplaced")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
}

type LifeCycleFragment struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

func (f *LifeCycleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = f.frame.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Any(f.initial()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "reload"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "updateEmpty"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "updateContent"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "updateEditor"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "clear"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "unmount"); if __e != nil { return }
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
			__e = __c.AttrSet("id", "remove"); if __e != nil { return }
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
}

func (f *LifeCycleFragment) initial() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); gox.Noop(ctx)
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Any(test.Marker("presist")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}
func (f *LifeCycleFragment) newEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); gox.Noop(ctx)
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.AttrSet("id", "new"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *LifeCycleFragment) newContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = f.node.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); gox.Noop(ctx)
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.AttrSet("id", "new2"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Any(test.Marker("presist2")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *LifeCycleFragment) newEditor() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
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
		f.rep(ctx, "channel err: "+err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

func (f *FragmentDetachedReplace) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
}

func (f *FragmentDetachedReplace) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.node.Update(ctx, test.Marker("replace-base"))
		f.frame.Update(ctx, f.mount())

		__e = __c.Any(&f.frame); if __e != nil { return }
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReplace(ctx, test.Marker("replace-detached")), "ok replace")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "replace-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "reload-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUpdate(ctx, test.Marker("replace-updated")), "ok update")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "update-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(test.Button("remount-after-replace", func(ctx context.Context) bool {
		f.frame.Update(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
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
		f.rep(ctx, "channel err: "+err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

func (f *FragmentDetachedRebase) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
}

func (f *FragmentDetachedRebase) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("section"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "rebased-detached-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Any(test.Marker("rebased-detached")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentDetachedRebase) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.node.Update(ctx, test.Marker("rebase-base"))
		f.frame.Update(ctx, f.mount())

		__e = __c.Any(&f.frame); if __e != nil { return }
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUnmount(ctx), "ok unmount")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "unmount-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("unmount-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReload(ctx), "ok reload")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "reload-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XRebase(ctx, f.rebased()), "ok rebase")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "rebase-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(test.Button("remount-after-rebase", func(ctx context.Context) bool {
		f.frame.Update(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
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
		f.rep(ctx, "channel err: "+err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

func (f *FragmentProxyMove) mountFrame1() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("section"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "frame1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentProxyMove) mountFrame2() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("section"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentProxyMove) frame2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("section"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentProxyMove) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("section"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "proxy-moved-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Any(test.Marker("proxy-moved")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentProxyMove) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.node.Update(ctx, test.Marker("proxy-base"))
		f.frame1.Update(ctx, f.mountFrame1())
		f.frame2.Update(ctx, f.frame2Empty())

		__e = __c.Any(&f.frame1); if __e != nil { return }
		__e = __c.Any(&f.frame2); if __e != nil { return }
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XRebase(ctx, f.rebased()), "ok rebase")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "rebase-proxy-move"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-proxy-move"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(test.Button("move-proxy", func(ctx context.Context) bool {
		f.frame2.Update(ctx, f.mountFrame2())
		return false
	})); if __e != nil { return }
	return })
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
		f.rep(ctx, "channel err: "+err.Error())
		return false
	}
	f.rep(ctx, okMsg)
	return false
}

func (f *FragmentHierarchy) childBody() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("article"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "child-body"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Any(&f.grand); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentHierarchy) host1Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("section"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "host1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentHierarchy) host2Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("section"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentHierarchy) host2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Init("section"); if __e != nil { return }
		{
			__e = __c.AttrSet("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
}

func (f *FragmentHierarchy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.grand.Update(ctx, test.Marker("grand-init"))
		f.child.Update(ctx, f.childBody())
		f.host1.Update(ctx, f.host1Body())
		f.host2.Update(ctx, f.host2Empty())

		__e = __c.Any(&f.host1); if __e != nil { return }
		__e = __c.Any(&f.host2); if __e != nil { return }
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = __c.Any(test.Button("move-child", func(ctx context.Context) bool {
		f.host2.Update(ctx, f.host2Body())
		return false
	})); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.grand.XUpdate(ctx, test.Marker("grand-updated")), "ok grand")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "grand-update"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("grand-update"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Any(test.Button("remove-host2", func(ctx context.Context) bool {
		f.host2.Delete(ctx)
		return false
	})); if __e != nil { return }
	return })
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
		f.rep(ctx, "channel err: "+err.Error())
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

func (f *FragmentErrorTransitions) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
}

func (f *FragmentErrorTransitions) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); gox.Noop(ctx)
		f.node.Update(ctx, test.Marker("error-base"))
		f.frame.Update(ctx, f.mount())

		__e = __c.Any(&f.frame); if __e != nil { return }
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XUpdate(ctx, f.errElem("update boom")), "ok update")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "update-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XReplace(ctx, f.errElem("replace boom")), "ok replace")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "replace-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
		__e = doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.XRebase(ctx, f.errElem("rebase boom")), "ok rebase")
		},
	}.Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); gox.Noop(ctx)
			__e = __c.Init("button"); if __e != nil { return }
			{
				__e = __c.AttrSet("id", "rebase-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
}
