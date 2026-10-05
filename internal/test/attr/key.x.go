// Managed by GoX v0.3.2

//line key.gox:1
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

package attr

import (
	"context"
	"fmt"
	
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type keyFragment struct {
	r *test.Reporter
	test.NoBeam
}

//line key.gox:31
func (f *keyFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line key.gox:32
		__e = (doors.AKeyDown{
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.r.Update(ctx, 0, r.Event().Key)
			f.r.Update(ctx, 1, "down")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
//line key.gox:38
			__e = (doors.AKeyUp{
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.r.Update(ctx, 2, r.Event().Key)
			f.r.Update(ctx, 3, "up")
			if r.Event().ShiftKey {
				f.r.Update(ctx, 4, fmt.Sprint(r.Event().ShiftKey))
			}
			if r.Event().CtrlKey {
				f.r.Update(ctx, 5, fmt.Sprint(r.Event().CtrlKey))
			}
			if r.Event().AltKey {
				f.r.Update(ctx, 6, fmt.Sprint(r.Event().AltKey))
			}
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
				ctx := __c.Context(); _ = ctx
//line key.gox:53
				__e = (doors.AKeyUp{
		Keys: doors.Key{Key: "e", CtrlMod: doors.ModOn},
		On: func(ctx context.Context, r doors.RequestEvent[doors.KeyboardEvent]) bool {
			f.r.Update(ctx, 7, "ctrl-e")
			return false
		},
	}).Proxy(__c, gox.Elem(func(__c gox.Cursor) (__e error) {
					ctx := __c.Context(); _ = ctx
					__e = __c.InitVoid("input"); if __e != nil { return }
					{
//line key.gox:59
						__e = __c.Set("type", "text"); if __e != nil { return }
//line key.gox:59
						__e = __c.Set("id", "input"); if __e != nil { return }
//line key.gox:59
						__e = __c.Set("placeholder", ""); if __e != nil { return }
					}
					__e = __c.Submit(); if __e != nil { return }
				return })); if __e != nil { return }
			return })); if __e != nil { return }
		return })); if __e != nil { return }
		__e = __c.InitVoid("hr"); if __e != nil { return }
		{
		}
		__e = __c.Submit(); if __e != nil { return }
//line key.gox:62
		f.r.Update(ctx, 0, "")
	f.r.Update(ctx, 1, "")
	f.r.Update(ctx, 2, "")
	f.r.Update(ctx, 3, "")
	f.r.Update(ctx, 4, "")
	f.r.Update(ctx, 5, "")
	f.r.Update(ctx, 6, "")
	f.r.Update(ctx, 7, "")

//line key.gox:71
		__e = __c.Any(f.r); if __e != nil { return }
	return })
//line key.gox:72
}
