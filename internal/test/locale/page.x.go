// Managed by GoX v0.4.0

//line page.gox:1
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

package locale

import (
	"fmt"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

type localeFragment struct {
	test.NoBeam
}

func show(l doors.Locale) string {
	zone := l.TimeZone.String()
	if l.TimeZone == time.Local {
		zone = "local"
	}
	return fmt.Sprintf("%s|%s|%t", l.Language, zone, l.Synced)
}

//line page.gox:38
func (f *localeFragment) Main() gox.Elem {
	return gox.Elem(func(__c gox.Cursor) (__e error) {
		ctx := __c.Context(); _ = ctx
//line page.gox:40
		locale := doors.BrowserLocale(ctx)
	initial, _ := locale.Read(ctx)

		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:43
			__e = __c.Set("id", "instance-id"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:43
			__e = __c.Any(doors.InstanceID(ctx)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
		__e = __c.Init("div"); if __e != nil { return }
		{
//line page.gox:44
			__e = __c.Set("id", "initial"); if __e != nil { return }
			__e = __c.Submit(); if __e != nil { return }
//line page.gox:44
			__e = __c.Any(show(initial)); if __e != nil { return }
		}
		__e = __c.Close(); if __e != nil { return }
//line page.gox:45
		__e = __c.Any(locale.Bind(func(l doors.Locale) gox.Elem {
		return gox.Elem(func(__c gox.Cursor) (__e error) {
			ctx := __c.Context(); _ = ctx
			__e = __c.Init("div"); if __e != nil { return }
			{
//line page.gox:46
				__e = __c.Set("id", "live"); if __e != nil { return }
				__e = __c.Submit(); if __e != nil { return }
//line page.gox:46
				__e = __c.Any(show(l)); if __e != nil { return }
			}
			__e = __c.Close(); if __e != nil { return }
		return })
//line page.gox:47
	})); if __e != nil { return }
	return })
//line page.gox:48
}
