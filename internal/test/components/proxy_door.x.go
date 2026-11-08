// Managed by GoX v0.3.3

//line proxy_door.gox:1
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

package components

import (
	"context"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

func proxyDoorMod(name string) gox.Modify {
	return gox.ModifyFunc(func(_ context.Context, tag string, attrs gox.Attrs) error {
		attrs.Get("data-pd").Set(tag + ":" + name)
		return nil
	})
}

func proxyDoorID(id string) gox.Modify {
	return gox.ModifyFunc(func(_ context.Context, _ string, attrs gox.Attrs) error {
		attrs.Get("id").Set(id)
		return nil
	})
}

type ProxyDoorFragment struct {
	test.NoBeam
	r       *test.Reporter
	wrap    doors.Door
	cont    doors.Door
	proxy   doors.Door
	static  doors.Door
	mounted doors.Door
	id      doors.Door
	hostA   doors.Door
	hostB   doors.Door
	move    doors.Door
	text    doors.Door
}

//line proxy_door.gox:54
func (f *ProxyDoorFragment) moveMod(name string) gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line proxy_door.gox:55
		__e = (doors.ProxyMod(proxyDoorMod(name))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line proxy_door.gox:55
			__e = (doors.AClick{
		On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
			f.r.Update(ctx, 3, "move-"+name)
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line proxy_door.gox:60
				__e = __c.Any(&f.move); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
	return })
//line proxy_door.gox:61
}

//line proxy_door.gox:63
func (f *ProxyDoorFragment) movePlain() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line proxy_door.gox:64
		__e = __c.Any(&f.move); if __e != nil { return }
	return })
//line proxy_door.gox:65
}

//line proxy_door.gox:67
func (f *ProxyDoorFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line proxy_door.gox:69
		f.wrap.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line proxy_door.gox:69
			__e = __c.Set("id", "pd-wrap-inner"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("wrap-0"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:69
	return }))
	f.cont.Outer(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("div"); if __e != nil { return }
		{
//line proxy_door.gox:70
			__e = __c.Set("id", "pd-elem"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Init("span"); if __e != nil { return }
			{
//line proxy_door.gox:70
				__e = __c.Set("id", "pd-elem-inner"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("elem-0"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:70
	return }))
	f.static.Static(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.InitContainer(); if __e != nil { return }
		{
			__e = __c.Init("p"); if __e != nil { return }
			{
//line proxy_door.gox:71
				__e = __c.Set("id", "pd-static"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("static"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
			__e = __c.Init("p"); if __e != nil { return }
			{
//line proxy_door.gox:71
				__e = __c.Set("id", "pd-static-sibling"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("sibling"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:71
	return }))
	f.mounted.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line proxy_door.gox:72
			__e = __c.Set("id", "pd-mounted-inner"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("mounted"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:72
	return }))
	f.id.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line proxy_door.gox:73
			__e = __c.Set("id", "pd-id-inner"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("id-0"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:73
	return }))
	f.move.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line proxy_door.gox:74
			__e = __c.Set("id", "pd-move-inner"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("move"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:74
	return }))
	f.hostA.Inner(ctx, f.moveMod("a"))
	f.text.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
		__e = __c.Init("span"); if __e != nil { return }
		{
//line proxy_door.gox:76
			__e = __c.Set("id", "pd-text-inner"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
			__e = __c.Text("text"); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:76
	return }))

		__e = __c.Init("div"); if __e != nil { return }
		{
//line proxy_door.gox:78
			__e = __c.Set("id", "pd-root"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line proxy_door.gox:79
			__e = (doors.Class("pd-class")).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line proxy_door.gox:79
				__e = (doors.ProxyMod(proxyDoorMod("wrap"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
//line proxy_door.gox:79
					__e = (doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.r.Update(ctx, 0, "wrap-click")
				return false
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
//line proxy_door.gox:84
						__e = __c.Any(&f.wrap); if __e != nil { return }
					return })); if __e != nil { return }
				return })); if __e != nil { return }
			return })); if __e != nil { return }
//line proxy_door.gox:85
			__e = (doors.ProxyMod(proxyDoorMod("elem"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line proxy_door.gox:85
				__e = (doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.r.Update(ctx, 1, "elem-click")
				return false
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
//line proxy_door.gox:90
					__e = __c.Any(&f.cont); if __e != nil { return }
				return })); if __e != nil { return }
			return })); if __e != nil { return }
//line proxy_door.gox:91
			__e = (doors.ProxyMod(proxyDoorMod("proxy"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line proxy_door.gox:91
				__e = (doors.AClick{
			On: func(ctx context.Context, _ doors.RequestEvent[doors.PointerEvent]) bool {
				f.r.Update(ctx, 2, "proxy-click")
				return false
			},
		}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
//line proxy_door.gox:96
					__e = (&f.proxy).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
						ctx := __c.Context(); _ = ctx
						__e = __c.Init("div"); if __e != nil { return }
						{
//line proxy_door.gox:96
							__e = __c.Set("id", "pd-proxy"); if __e != nil { return }
							__e = __c.Submit(); if __e != nil { return }
							__e = __c.Init("span"); if __e != nil { return }
							{
//line proxy_door.gox:96
								__e = __c.Set("id", "pd-proxy-inner"); if __e != nil { return }
								__e = __c.Submit(); if __e != nil { return }
								__e = __c.Text("proxy-0"); if __e != nil { return }
							}
							__e = __c.Close(); if __e != nil { return }
						}
						__e = __c.Close(); if __e != nil { return }
					return })); if __e != nil { return }
				return })); if __e != nil { return }
			return })); if __e != nil { return }
//line proxy_door.gox:97
			__e = (doors.ProxyMod(proxyDoorMod("static"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line proxy_door.gox:97
				__e = __c.Any(&f.static); if __e != nil { return }
			return })); if __e != nil { return }
//line proxy_door.gox:98
			__e = (doors.ProxyMod(proxyDoorMod("mounted"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line proxy_door.gox:98
				__e = __c.Any(&f.mounted); if __e != nil { return }
			return })); if __e != nil { return }
//line proxy_door.gox:99
			__e = (doors.ProxyMod(proxyDoorID("pd-id"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line proxy_door.gox:99
				__e = __c.Any(&f.id); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line proxy_door.gox:101
			__e = __c.Set("id", "pd-host-a"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line proxy_door.gox:101
			__e = __c.Any(&f.hostA); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line proxy_door.gox:102
			__e = __c.Set("id", "pd-host-b"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line proxy_door.gox:102
			__e = __c.Any(&f.hostB); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line proxy_door.gox:103
			__e = __c.Set("id", "pd-text-host"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line proxy_door.gox:103
			__e = (doors.ProxyMod(proxyDoorMod("text"))).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line proxy_door.gox:103
				__e = __c.Any(&f.text); if __e != nil { return }
			return })); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:104
		__e = __c.Any(test.Button("pd-inner", func(ctx context.Context) bool {
		f.wrap.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
//line proxy_door.gox:105
				__e = __c.Set("id", "pd-wrap-inner"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("wrap-1"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:105
		return }))
		f.cont.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
//line proxy_door.gox:106
				__e = __c.Set("id", "pd-elem-inner"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("elem-1"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:106
		return }))
		f.proxy.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
//line proxy_door.gox:107
				__e = __c.Set("id", "pd-proxy-inner"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("proxy-1"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:107
		return }))
		return false
	})); if __e != nil { return }
//line proxy_door.gox:110
		__e = __c.Any(test.Button("pd-outer", func(ctx context.Context) bool {
		f.wrap.Outer(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
				__e = __c.Init("span"); if __e != nil { return }
				{
//line proxy_door.gox:111
					__e = __c.Set("id", "pd-wrap-inner"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("wrap-2"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:111
		return }))
		f.cont.Outer(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line proxy_door.gox:112
				__e = __c.Set("id", "pd-elem"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line proxy_door.gox:112
					__e = __c.Set("id", "pd-elem-inner"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("elem-2"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:112
		return }))
		f.proxy.Outer(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line proxy_door.gox:113
				__e = __c.Set("id", "pd-proxy"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Init("span"); if __e != nil { return }
				{
//line proxy_door.gox:113
					__e = __c.Set("id", "pd-proxy-inner"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("proxy-2"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:113
		return }))
		return false
	})); if __e != nil { return }
//line proxy_door.gox:116
		__e = __c.Any(test.Button("pd-mounted-static", func(ctx context.Context) bool {
		f.mounted.Static(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("p"); if __e != nil { return }
			{
//line proxy_door.gox:117
				__e = __c.Set("id", "pd-mounted-static"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("mounted-static"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:117
		return }))
		return false
	})); if __e != nil { return }
//line proxy_door.gox:120
		__e = __c.Any(test.Button("pd-id-inner-update", func(ctx context.Context) bool {
		f.id.Inner(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("span"); if __e != nil { return }
			{
//line proxy_door.gox:121
				__e = __c.Set("id", "pd-id-inner"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
				__e = __c.Text("id-1"); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:121
		return }))
		return false
	})); if __e != nil { return }
//line proxy_door.gox:124
		__e = __c.Any(test.Button("pd-id-outer-update", func(ctx context.Context) bool {
		f.id.Outer(ctx, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.InitContainer(); if __e != nil { return }
			{
				__e = __c.Init("span"); if __e != nil { return }
				{
//line proxy_door.gox:125
					__e = __c.Set("id", "pd-id-inner"); if __e != nil { return }
					__e = __c.Submit(); if __e != nil { return }
					__e = __c.Text("id-2"); if __e != nil { return }
				}
				__e = __c.Close(); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
//line proxy_door.gox:125
		return }))
		return false
	})); if __e != nil { return }
//line proxy_door.gox:128
		__e = __c.Any(test.Button("pd-remod", func(ctx context.Context) bool {
		f.hostA.Inner(ctx, f.moveMod("b"))
		return false
	})); if __e != nil { return }
//line proxy_door.gox:132
		__e = __c.Any(test.Button("pd-move", func(ctx context.Context) bool {
		f.r.Update(ctx, 3, "moved")
		f.hostA.Inner(ctx, nil)
		f.hostB.Inner(ctx, f.movePlain())
		return false
	})); if __e != nil { return }
//line proxy_door.gox:138
		__e = __c.Any(test.Button("pd-text-static", func(ctx context.Context) bool {
		f.text.Static(ctx, "done")
		return false
	})); if __e != nil { return }
//line proxy_door.gox:142
		__e = __c.Any(test.Button("pd-ping", func(ctx context.Context) bool {
		f.r.Update(ctx, 4, "ping")
		return false
	})); if __e != nil { return }
//line proxy_door.gox:146
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line proxy_door.gox:147
}
