// Managed by GoX v0.3.2

//line node_fragments.gox:1
// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

//line node_fragments.gox:33
func (f *FragmentMany) sample() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:34
			__e = __c.Set("class", "sample"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("sample"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:35
}

//line node_fragments.gox:37
func (f *FragmentMany) manyDoors() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:38
		for i := range 20 {
			__e = __c.Init("span"); if __e != nil { return }
			{
//line node_fragments.gox:39
				__e = __c.Set("style", "display:none"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:39
				__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:40
			__e = __c.Any(&f.n); if __e != nil { return }
		}
	return })
//line node_fragments.gox:42
}

//line node_fragments.gox:44
func (f *FragmentMany) replaced() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:46
		f.n.Static(ctx, f.sample())

//line node_fragments.gox:48
		for i := range 100 {
			__e = __c.Init("span"); if __e != nil { return }
			{
//line node_fragments.gox:49
				__e = __c.Set("style", "display:none"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:49
				__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:50
			__e = __c.Any(&f.n); if __e != nil { return }
		}
	return })
//line node_fragments.gox:52
}

//line node_fragments.gox:54
func (f *FragmentMany) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:56
		f.n.Inner(ctx, f.sample())
	n := doors.Door{}

//line node_fragments.gox:59
		__e = (n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:60
				__e = __c.Any(f.manyDoors()); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:62
		__e = __c.Any(test.Button("replace", func(ctx context.Context) bool {
		n.Inner(ctx, f.replaced())
		return true
	})); if __e != nil { return }
	return })
//line node_fragments.gox:66
}

type FragmentProxyWrappedSiblings struct {
	n doors.Door
	test.NoBeam
}

//line node_fragments.gox:73
func (f *FragmentProxyWrappedSiblings) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:74
		__e = __c.Any(gox.Elem(func(cur gox.Cursor) error {
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
//line node_fragments.gox:107
}

type FragmentProxyWrappedLoop struct {
	n doors.Door
	test.NoBeam
}

//line node_fragments.gox:114
func (f *FragmentProxyWrappedLoop) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:115
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:115
			for i := range 2 {
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:116
					__e = __c.Set("id", fmt.Sprintf("proxy-loop-%d", i)); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:116
					__e = __c.Any(fmt.Sprint(i)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:118
}

type FragmentX struct {
	report doors.Door
	n doors.Door
	test.NoBeam
}

func (f *FragmentX) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

//line node_fragments.gox:130
func (f *FragmentX) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:131
		__e = (f.n).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:132
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:134
		__e = (f.report).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:134
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
//line node_fragments.gox:152
					__e = __c.Set("id", "updatex"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("C"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:154
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
//line node_fragments.gox:172
				__e = __c.Set("id", "removex"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("R"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:173
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

//line node_fragments.gox:203
func (f *FragmentXDoor) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:204
		__e = __c.Any(&f.n); if __e != nil { return }
	return })
//line node_fragments.gox:205
}

//line node_fragments.gox:207
func (f *FragmentXDoor) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:209
		f.n.Inner(ctx, test.Marker("x-init"))
	f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:212
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:213
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:214
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Reload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:218
				__e = __c.Set("id", "xreload"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreload"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:219
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Outer(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("section"); if __e != nil { return }
				{
//line node_fragments.gox:221
					__e = __c.Set("id", "x-rebased-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:222
					__e = __c.Any(test.Marker("x-rebased")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:223
			return })), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:225
				__e = __c.Set("id", "xrebase"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xrebase"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:226
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Inner(ctx, nil), "ok clear")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:230
				__e = __c.Set("id", "xclear"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xclear"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:231
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Inner(ctx, test.Marker("x-updated")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:235
				__e = __c.Set("id", "xupdate"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xupdate"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:236
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Unmount(ctx), "ok unmount")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:240
				__e = __c.Set("id", "xunmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xunmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:241
		__e = __c.Any(test.Button("xremount", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:245
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.n.Static(ctx, test.Marker("x-replaced")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:249
				__e = __c.Set("id", "xreplace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("xreplace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:250
}

type EmbeddedFragment struct {
	n1 doors.Door
	n2 doors.Door
	n3 doors.Door
	test.NoBeam
}

//line node_fragments.gox:259
func (f *EmbeddedFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:260
		__e = (f.n1).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
//line node_fragments.gox:261
				__e = (f.n2).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:262
						__e = __c.Any(test.Marker("init")); if __e != nil { return }
					}
					__e = __c.Close(); if __e != nil { return }
				return })); if __e != nil { return }
//line node_fragments.gox:264
				__e = __c.Any(test.Marker("static")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:266
		__e = __c.Any(&f.n3); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:268
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:269
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
//line node_fragments.gox:278
			__e = __c.Set("id", "clear"); if __e != nil { return }
//line node_fragments.gox:279
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
//line node_fragments.gox:288
			__e = __c.Set("id", "replace"); if __e != nil { return }
//line node_fragments.gox:289
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
//line node_fragments.gox:299
}

type DynamicFragment struct {
	n1 doors.Door
	n2 doors.Door
	test.NoBeam
}

//line node_fragments.gox:307
func (f *DynamicFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:309
		f.n1.Inner(ctx, test.Marker("init"))

//line node_fragments.gox:312
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
//line node_fragments.gox:315
			__e = __c.Set("id", "update"); if __e != nil { return }
//line node_fragments.gox:316
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
//line node_fragments.gox:325
			__e = __c.Set("id", "replace"); if __e != nil { return }
//line node_fragments.gox:326
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
//line node_fragments.gox:336
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:337
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
//line node_fragments.gox:345
}

type BeforeFragment struct {
	doorInit doors.Door
	doorUpdate doors.Door
	doorRemoved doors.Door
	doorReplaced doors.Door
	test.NoBeam
}

//line node_fragments.gox:355
func (f *BeforeFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:356
		__e = (f.doorInit).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:357
				__e = __c.Any(test.Marker("init")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:361
		f.doorUpdate.Inner(ctx, test.Marker("updated"))

//line node_fragments.gox:363
		__e = (f.doorUpdate).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:367
		f.doorRemoved.Inner(ctx, test.Marker("removed"))

//line node_fragments.gox:370
		f.doorRemoved.Static(ctx, nil)

//line node_fragments.gox:372
		__e = __c.Any(&f.doorRemoved); if __e != nil { return }
//line node_fragments.gox:375
		f.doorReplaced.Static(ctx, test.Marker("replaced"))

//line node_fragments.gox:378
		__e = (f.doorReplaced).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:379
				__e = __c.Any(test.Marker("initReplaced")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:381
}

type LifeCycleFragment struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

//line node_fragments.gox:389
func (f *LifeCycleFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:390
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:390
			__e = __c.Any(f.initial()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:392
			__e = __c.Set("id", "reload"); if __e != nil { return }
//line node_fragments.gox:393
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
//line node_fragments.gox:402
			__e = __c.Set("id", "updateEmpty"); if __e != nil { return }
//line node_fragments.gox:403
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
//line node_fragments.gox:412
			__e = __c.Set("id", "updateEmptyAlt"); if __e != nil { return }
//line node_fragments.gox:413
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
//line node_fragments.gox:422
			__e = __c.Set("id", "updateContent"); if __e != nil { return }
//line node_fragments.gox:423
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
//line node_fragments.gox:432
			__e = __c.Set("id", "updateInner"); if __e != nil { return }
//line node_fragments.gox:433
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
//line node_fragments.gox:442
			__e = __c.Set("id", "updateOuter"); if __e != nil { return }
//line node_fragments.gox:443
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
//line node_fragments.gox:452
			__e = __c.Set("id", "replaceStatic"); if __e != nil { return }
//line node_fragments.gox:453
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
//line node_fragments.gox:462
			__e = __c.Set("id", "updateEditor"); if __e != nil { return }
//line node_fragments.gox:463
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
//line node_fragments.gox:472
			__e = __c.Set("id", "clear"); if __e != nil { return }
//line node_fragments.gox:473
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
//line node_fragments.gox:482
			__e = __c.Set("id", "unmount"); if __e != nil { return }
//line node_fragments.gox:483
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
//line node_fragments.gox:492
			__e = __c.Set("id", "remove"); if __e != nil { return }
//line node_fragments.gox:493
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
//line node_fragments.gox:501
}

//line node_fragments.gox:503
func (f *LifeCycleFragment) initial() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:505
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:506
					__e = __c.Any(test.Marker("presist")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:509
}
//line node_fragments.gox:510
func (f *LifeCycleFragment) newEmpty() gox.Elem {
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
					__e = __c.Set("id", "new"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:515
}

//line node_fragments.gox:517
func (f *LifeCycleFragment) newEmptyAlt() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:519
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("section"); if __e != nil { return }
				{
//line node_fragments.gox:519
					__e = __c.Set("id", "new-alt"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:522
}

//line node_fragments.gox:524
func (f *LifeCycleFragment) newContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:526
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:526
					__e = __c.Set("id", "new2"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:527
					__e = __c.Any(test.Marker("presist2")); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:530
}

//line node_fragments.gox:532
func (f *LifeCycleFragment) newOuter() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:533
			__e = __c.Set("id", "outer-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:534
			__e = __c.Any(test.Marker("outer-presist")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:536
}

//line node_fragments.gox:538
func (f *LifeCycleFragment) newEditor() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:540
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:542
}

type FragmentProxyReloadContent struct {
	frame doors.Door
	node doors.Door
	test.NoBeam
}

//line node_fragments.gox:550
func (f *FragmentProxyReloadContent) mountEmpty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:551
			__e = __c.Set("id", "proxy-redraw-frame"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:552
			__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:552
					__e = __c.Set("id", "proxy-redraw-root"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:555
}

//line node_fragments.gox:557
func (f *FragmentProxyReloadContent) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:558
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line node_fragments.gox:558
			__e = __c.Any(f.mountEmpty()); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:560
			__e = __c.Set("id", "proxy-redraw-update"); if __e != nil { return }
//line node_fragments.gox:561
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
//line node_fragments.gox:570
			__e = __c.Set("id", "proxy-redraw-remount"); if __e != nil { return }
//line node_fragments.gox:571
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
//line node_fragments.gox:580
			__e = __c.Set("id", "proxy-redraw-reload"); if __e != nil { return }
//line node_fragments.gox:581
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
//line node_fragments.gox:589
}

type FragmentClosestReload struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:599
func (f *FragmentClosestReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:601
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:603
			__e = __c.Set("id", "inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:603
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:605
			__e = __c.Set("id", "reload-nearest"); if __e != nil { return }
//line node_fragments.gox:606
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
//line node_fragments.gox:614
}

//line node_fragments.gox:616
func (f *FragmentClosestReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:618
		f.outerRenders++
	f.node.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:621
			__e = __c.Set("id", "outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:621
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:622
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:623
}

//line node_fragments.gox:625
func (f *FragmentClosestReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:627
		f.frame.Inner(ctx, f.outerContent())

//line node_fragments.gox:629
		__e = __c.Any(&f.frame); if __e != nil { return }
	return })
//line node_fragments.gox:630
}

type FragmentClosestReloadProxy struct {
	frame doors.Door
	node doors.Door
	outerRenders int
	innerRenders int
	test.NoBeam
}

//line node_fragments.gox:640
func (f *FragmentClosestReloadProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:641
		__e = (f.frame).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:641
				__e = __c.Set("id", "outer-proxy-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:643
				f.outerRenders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:645
					__e = __c.Set("id", "proxy-outer-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:645
					__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:646
				__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.Init("div"); if __e != nil { return }
					{
//line node_fragments.gox:646
						__e = __c.Set("id", "inner-proxy-root"); if __e != nil { return }
						__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:648
						f.innerRenders++

						__e = __c.Init("div"); if __e != nil { return }
						{
//line node_fragments.gox:650
							__e = __c.Set("id", "proxy-inner-count"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:650
							__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
						__e = __c.Init("button"); if __e != nil { return }
						{
//line node_fragments.gox:652
							__e = __c.Set("id", "reload-nearest-proxy"); if __e != nil { return }
//line node_fragments.gox:653
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
//line node_fragments.gox:663
}

type FragmentInlineDoorPointerProxy struct {
	renders int
	test.NoBeam
}

//line node_fragments.gox:670
func (f *FragmentInlineDoorPointerProxy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:671
		__e = (&doors.Door{}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:671
				__e = __c.Set("id", "inline-door-root"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:673
				f.renders++

				__e = __c.Init("div"); if __e != nil { return }
				{
//line node_fragments.gox:675
					__e = __c.Set("id", "inline-door-count"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:675
					__e = __c.Any(fmt.Sprintf("inline-%d", f.renders)); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
				__e = __c.Init("button"); if __e != nil { return }
				{
//line node_fragments.gox:677
					__e = __c.Set("id", "inline-door-reload"); if __e != nil { return }
//line node_fragments.gox:678
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
//line node_fragments.gox:687
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

//line node_fragments.gox:738
func (f *FragmentContainerInnerLifecycle) content() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:739
			__e = __c.Set("id", "container-inner-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:739
			__e = __c.Any(fmt.Sprintf("click-%d", f.clicks)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:740
}

//line node_fragments.gox:742
func (f *FragmentContainerInnerLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:743
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:744
				__e = __c.Set("id", "container-inner-root"); if __e != nil { return }
//line node_fragments.gox:745
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
//line node_fragments.gox:752
					__e = __c.Set("id", "container-inner-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("initial"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:754
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

//line node_fragments.gox:768
func (f *FragmentContainerEffectLifecycle) content(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:769
			__e = __c.Set("id", "container-effect-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:769
			__e = __c.Any(fmt.Sprintf("%s-%d", label, f.source.Get())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:770
}

//line node_fragments.gox:772
func (f *FragmentContainerEffectLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:774
		f.init()

//line node_fragments.gox:776
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:777
				__e = __c.Set("id", "container-effect-root"); if __e != nil { return }
//line node_fragments.gox:778
				__e = __c.Modify(containerEffectAttr{source: f.source}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:779
				__e = __c.Any(f.content("initial")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:781
		__e = __c.Any(test.Button("container-effect-inner", func(ctx context.Context) bool {
		f.node.Inner(ctx, f.content("inner"))
		return false
	})); if __e != nil { return }
//line node_fragments.gox:785
		__e = __c.Any(test.Button("container-effect-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:789
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

//line node_fragments.gox:811
func (f *FragmentContainerOuterLifecycle) outerInner() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:812
			__e = __c.Set("id", "container-outer-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:812
			__e = __c.Any(fmt.Sprintf("outer-click-%d", f.outerClicks)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:813
}

//line node_fragments.gox:815
func (f *FragmentContainerOuterLifecycle) outer() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:817
			__e = __c.Set("id", "container-outer-new-root"); if __e != nil { return }
//line node_fragments.gox:818
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
//line node_fragments.gox:825
				__e = __c.Set("id", "container-outer-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("outer"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:827
}

//line node_fragments.gox:829
func (f *FragmentContainerOuterLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:831
		f.init()

//line node_fragments.gox:833
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:834
				__e = __c.Set("id", "container-outer-root"); if __e != nil { return }
//line node_fragments.gox:835
				__e = __c.Modify(containerWatchAttr{source: f.source, cancels: &f.cancels, watches: &f.watches}); if __e != nil { return }
//line node_fragments.gox:836
				__e = __c.Modify(doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.node.Outer(ctx, f.outer())
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line node_fragments.gox:842
					__e = __c.Set("id", "container-outer-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("initial"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:844
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:845
		__e = __c.Any(test.Button("container-outer-report", func(ctx context.Context) bool {
		f.reportState(ctx)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:849
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

//line node_fragments.gox:878
func (f *FragmentContainerReloadLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:880
		f.init()

//line node_fragments.gox:882
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:883
				__e = __c.Set("id", "container-reload-root"); if __e != nil { return }
//line node_fragments.gox:884
				__e = __c.Modify(containerWatchAttr{source: f.source, cancels: &f.cancels, watches: &f.watches}); if __e != nil { return }
//line node_fragments.gox:885
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
//line node_fragments.gox:892
					__e = __c.Set("id", "container-reload-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:892
					__e = __c.Any(f.value()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:894
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:895
		__e = __c.Any(test.Button("container-reload-report", func(ctx context.Context) bool {
		f.reportState(ctx)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:899
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

//line node_fragments.gox:968
func (f *FragmentContainerHookStateLifecycle) content(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:969
			__e = __c.Set("id", "container-hook-state-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:969
			__e = __c.Any(fmt.Sprintf("%s-%d", label, f.source.Get())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:970
}

//line node_fragments.gox:972
func (f *FragmentContainerHookStateLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:974
		f.init()

//line node_fragments.gox:976
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:977
				__e = __c.Set("id", "container-hook-state-root"); if __e != nil { return }
//line node_fragments.gox:978
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
//line node_fragments.gox:993
				__e = __c.Any(f.content("initial")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:995
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:996
		__e = __c.Any(test.Button("container-hook-state-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1000
		__e = __c.Any(test.Button("container-hook-state-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1004
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

//line node_fragments.gox:1073
func (f *FragmentContainerHookStateOuterLifecycle) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:1074
			__e = __c.Set("id", "container-hook-outer-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1074
			__e = __c.Any(fmt.Sprintf("outer-click-%d", f.outerClicks)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1075
}

//line node_fragments.gox:1077
func (f *FragmentContainerHookStateOuterLifecycle) outer() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1079
			__e = __c.Set("id", "container-hook-outer-new-root"); if __e != nil { return }
//line node_fragments.gox:1080
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
//line node_fragments.gox:1091
				__e = __c.Set("id", "container-hook-outer-value"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("outer"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1093
}

//line node_fragments.gox:1095
func (f *FragmentContainerHookStateOuterLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1097
		f.init()

//line node_fragments.gox:1099
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:1100
				__e = __c.Set("id", "container-hook-outer-root"); if __e != nil { return }
//line node_fragments.gox:1101
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
//line node_fragments.gox:1108
					__e = __c.Set("id", "container-hook-outer-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("initial"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1110
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1111
		__e = __c.Any(test.Button("container-hook-outer-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1115
		__e = __c.Any(test.Button("container-hook-outer-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1119
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

//line node_fragments.gox:1156
func (f *FragmentContainerHookEffectLifecycle) content(label string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1158
		f.renders.Add(1)

		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:1160
			__e = __c.Set("id", "container-hook-effect-renders"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1160
			__e = __c.Any(fmt.Sprint(f.renders.Load())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("span"); if __e != nil { return }
		{
//line node_fragments.gox:1161
			__e = __c.Set("id", "container-hook-effect-value"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1161
			__e = __c.Any(fmt.Sprintf("%s-%d", label, f.source.Get())); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1162
}

//line node_fragments.gox:1164
func (f *FragmentContainerHookEffectLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1166
		f.init()

//line node_fragments.gox:1168
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:1169
				__e = __c.Set("id", "container-hook-effect-root"); if __e != nil { return }
//line node_fragments.gox:1170
				__e = __c.Modify(doors.AKeyDown{
			Keys: doors.Key{Key: "ContainerEffect"},
			On: func(ctx context.Context, _ doors.RequestEvent[doors.KeyboardEvent]) bool {
				value, ok := f.source.Effect(ctx)
				f.registrations.Add(1)
				f.last.Store(fmt.Sprintf("effect-%d-%t", value, ok))
				f.node.Inner(ctx, f.content("registered"))
				return false
			},
		}); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1180
				__e = __c.Any(f.content("outer")); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1182
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1183
		__e = __c.Any(test.Button("container-hook-effect-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1187
		__e = __c.Any(test.Button("container-hook-effect-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1191
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

//line node_fragments.gox:1267
func (f *FragmentContainerHookStateReloadLifecycle) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1269
		f.init()

//line node_fragments.gox:1271
		__e = (f.node).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line node_fragments.gox:1272
				__e = __c.Set("id", "container-hook-reload-root"); if __e != nil { return }
//line node_fragments.gox:1273
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
//line node_fragments.gox:1281
					__e = __c.Set("id", "container-hook-reload-value"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1281
					__e = __c.Any(f.value()); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1283
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1284
		__e = __c.Any(test.Button("container-hook-reload-update", func(ctx context.Context) bool {
		f.source.Update(ctx, f.source.Get() + 1)
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1288
		__e = __c.Any(test.Button("container-hook-reload-report", func(ctx context.Context) bool {
		f.reportState(ctx, "state")
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1292
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

//line node_fragments.gox:1324
func (f *FragmentClosestTrackedReload) innerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1326
		f.innerRenders++

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:1328
			__e = __c.Set("id", "x-inner-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1328
			__e = __c.Any(fmt.Sprintf("inner-%d", f.innerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:1330
			__e = __c.Set("id", "xreload-nearest"); if __e != nil { return }
//line node_fragments.gox:1331
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
//line node_fragments.gox:1338
}

//line node_fragments.gox:1340
func (f *FragmentClosestTrackedReload) outerContent() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1342
		f.outerRenders++
	f.node.Inner(ctx, f.innerContent())

		__e = __c.Init("div"); if __e != nil { return }
		{
//line node_fragments.gox:1345
			__e = __c.Set("id", "x-outer-count"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1345
			__e = __c.Any(fmt.Sprintf("outer-%d", f.outerRenders)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line node_fragments.gox:1346
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1347
}

//line node_fragments.gox:1349
func (f *FragmentClosestTrackedReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1351
		f.frame.Inner(ctx, f.outerContent())

//line node_fragments.gox:1353
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1354
		__e = __c.Any(&f.report); if __e != nil { return }
	return })
//line node_fragments.gox:1355
}

type FragmentRootTrackedReload struct {
	report doors.Door
	test.NoBeam
}

func (f *FragmentRootTrackedReload) rep(ctx context.Context, s string) {
	f.report.Inner(ctx, test.Report(s))
}

//line node_fragments.gox:1366
func (f *FragmentRootTrackedReload) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1367
		__e = __c.Any(&f.report); if __e != nil { return }
		__e = __c.Init("button"); if __e != nil { return }
		{
//line node_fragments.gox:1369
			__e = __c.Set("id", "root-xreload"); if __e != nil { return }
//line node_fragments.gox:1370
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
//line node_fragments.gox:1391
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

//line node_fragments.gox:1421
func (f *FragmentDetachedReplace) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1422
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1423
}

//line node_fragments.gox:1425
func (f *FragmentDetachedReplace) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1427
		f.node.Inner(ctx, test.Marker("replace-base"))
	f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1430
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1431
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1432
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Static(ctx, test.Marker("replace-detached")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1436
				__e = __c.Set("id", "replace-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1437
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Reload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1441
				__e = __c.Set("id", "reload-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1442
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Inner(ctx, test.Marker("replace-updated")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1446
				__e = __c.Set("id", "update-after-replace"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-after-replace"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1447
		__e = __c.Any(test.Button("remount-after-replace", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1451
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

//line node_fragments.gox:1481
func (f *FragmentDetachedRebase) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1482
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1483
}

//line node_fragments.gox:1485
func (f *FragmentDetachedRebase) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1486
			__e = __c.Set("id", "rebased-detached-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1487
			__e = __c.Any(test.Marker("rebased-detached")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1489
}

//line node_fragments.gox:1491
func (f *FragmentDetachedRebase) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1493
		f.node.Inner(ctx, test.Marker("rebase-base"))
	f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1496
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1497
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1498
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Unmount(ctx), "ok unmount")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1502
				__e = __c.Set("id", "unmount-detached"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("unmount-detached"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1503
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Reload(ctx), "ok reload")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1507
				__e = __c.Set("id", "reload-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("reload-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1508
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Outer(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1512
				__e = __c.Set("id", "rebase-after-unmount"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-after-unmount"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1513
		__e = __c.Any(test.Button("remount-after-rebase", func(ctx context.Context) bool {
		f.frame.Inner(ctx, f.mount())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1517
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

//line node_fragments.gox:1548
func (f *FragmentProxyMove) mountFrame1() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1549
			__e = __c.Set("id", "frame1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1550
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1552
}

//line node_fragments.gox:1554
func (f *FragmentProxyMove) mountFrame2() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1555
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1556
			__e = __c.Any(&f.node); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1558
}

//line node_fragments.gox:1560
func (f *FragmentProxyMove) frame2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1561
			__e = __c.Set("id", "frame2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1562
}

//line node_fragments.gox:1564
func (f *FragmentProxyMove) rebased() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1565
			__e = __c.Set("id", "proxy-moved-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1566
			__e = __c.Any(test.Marker("proxy-moved")); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1568
}

//line node_fragments.gox:1570
func (f *FragmentProxyMove) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1572
		f.node.Inner(ctx, test.Marker("proxy-base"))
	f.frame1.Inner(ctx, f.mountFrame1())
	f.frame2.Inner(ctx, f.frame2Empty())

//line node_fragments.gox:1576
		__e = __c.Any(&f.frame1); if __e != nil { return }
//line node_fragments.gox:1577
		__e = __c.Any(&f.frame2); if __e != nil { return }
//line node_fragments.gox:1578
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1579
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Outer(ctx, f.rebased()), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1583
				__e = __c.Set("id", "rebase-proxy-move"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-proxy-move"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1584
		__e = __c.Any(test.Button("move-proxy", func(ctx context.Context) bool {
		f.frame2.Inner(ctx, f.mountFrame2())
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1588
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

//line node_fragments.gox:1620
func (f *FragmentHierarchy) childBody() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("article"); if __e != nil { return }
		{
//line node_fragments.gox:1621
			__e = __c.Set("id", "child-body"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1622
			__e = __c.Any(&f.grand); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1624
}

//line node_fragments.gox:1626
func (f *FragmentHierarchy) host1Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1627
			__e = __c.Set("id", "host1"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1628
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1630
}

//line node_fragments.gox:1632
func (f *FragmentHierarchy) host2Body() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1633
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line node_fragments.gox:1634
			__e = __c.Any(&f.child); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1636
}

//line node_fragments.gox:1638
func (f *FragmentHierarchy) host2Empty() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("section"); if __e != nil { return }
		{
//line node_fragments.gox:1639
			__e = __c.Set("id", "host2"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
	return })
//line node_fragments.gox:1640
}

//line node_fragments.gox:1642
func (f *FragmentHierarchy) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1644
		f.grand.Inner(ctx, test.Marker("grand-init"))
	f.child.Inner(ctx, f.childBody())
	f.host1.Inner(ctx, f.host1Body())
	f.host2.Inner(ctx, f.host2Empty())

//line node_fragments.gox:1649
		__e = __c.Any(&f.host1); if __e != nil { return }
//line node_fragments.gox:1650
		__e = __c.Any(&f.host2); if __e != nil { return }
//line node_fragments.gox:1651
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1652
		__e = __c.Any(test.Button("move-child", func(ctx context.Context) bool {
		f.host2.Inner(ctx, f.host2Body())
		return false
	})); if __e != nil { return }
//line node_fragments.gox:1656
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.grand.Inner(ctx, test.Marker("grand-updated")), "ok grand")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1660
				__e = __c.Set("id", "grand-update"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("grand-update"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1661
		__e = __c.Any(test.Button("remove-host2", func(ctx context.Context) bool {
		f.host2.Static(ctx, nil)
		return false
	})); if __e != nil { return }
	return })
//line node_fragments.gox:1665
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

//line node_fragments.gox:1701
func (f *FragmentErrorTransitions) mount() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1702
		__e = __c.Any(&f.node); if __e != nil { return }
	return })
//line node_fragments.gox:1703
}

//line node_fragments.gox:1705
func (f *FragmentErrorTransitions) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line node_fragments.gox:1707
		f.node.Inner(ctx, test.Marker("error-base"))
	f.frame.Inner(ctx, f.mount())

//line node_fragments.gox:1710
		__e = __c.Any(&f.frame); if __e != nil { return }
//line node_fragments.gox:1711
		__e = __c.Any(&f.report); if __e != nil { return }
//line node_fragments.gox:1712
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Inner(ctx, f.errElem("update boom")), "ok update")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1716
				__e = __c.Set("id", "update-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("update-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1717
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Static(ctx, f.errElem("replace boom")), "ok replace")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1721
				__e = __c.Set("id", "replace-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("replace-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
//line node_fragments.gox:1722
		__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			return f.wait(ctx, f.node.Outer(ctx, f.errElem("rebase boom")), "ok rebase")
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("button"); if __e != nil { return }
			{
//line node_fragments.gox:1726
				__e = __c.Set("id", "rebase-error"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("rebase-error"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line node_fragments.gox:1727
}
