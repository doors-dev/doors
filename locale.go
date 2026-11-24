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

package doors

import (
	"context"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
)

// Locale is the language and time zone of the browser, as read from
// [BrowserLocale].
type Locale = common.Locale

// BrowserLocale returns the language and time zone of the browser that opened
// the session.
//
// A common pattern is to read it with [Beam.Effect] where it is needed: the
// value rarely changes, and when it does, the closest dynamic parent
// rerenders.
//
// On the first page of a session, the initial render has only the language
// from the request's Accept-Language header and [time.Local] as the time zone:
// the browser reports its own language and time zone after the page has
// loaded, which sets Synced. Later pages of the session initially render with
// the previously reported values.
//
// ctx must carry a Doors session; otherwise BrowserLocale panics.
//
// Example:
//
//	locale, _ := doors.BrowserLocale(ctx).Effect(ctx)
//	return t.In(locale.TimeZone).Format(time.Kitchen)
func BrowserLocale(ctx context.Context) Beam[Locale] {
	sess := ctx.Value(common.KeySession).(core.Session)
	return source[Locale]{sess.Locale()}
}
