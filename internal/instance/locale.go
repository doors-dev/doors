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

package instance

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	_ "time/tzdata"

	"github.com/doors-dev/doors/internal/common"
	"golang.org/x/text/language"
)

const timeZoneChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_+-/"

var timeZones sync.Map

func (sess *session) initLocale(r *http.Request) {
	if sess.locale.Get().Language != language.Und {
		return
	}
	tag := acceptLanguage(strings.Join(r.Header.Values("Accept-Language"), ","))
	if tag == language.Und {
		return
	}
	sess.locale.Mutate(sess.ctx, func(l common.Locale) common.Locale {
		if l.Language == language.Und {
			l.Language = tag
		}
		return l
	})
}

func (sess *session) ReportLocale(lang string, zone string) {
	tag := parseLanguage(lang)
	loc := sess.loadTimeZone(zone)
	sess.locale.Mutate(sess.ctx, func(l common.Locale) common.Locale {
		l.Synced = true
		if tag != language.Und {
			l.Language = tag
		}
		if loc != nil {
			l.TimeZone = loc
		}
		return l
	})
}

func (sess *session) loadTimeZone(name string) *time.Location {
	if !validTimeZone(name) {
		return nil
	}
	if loc, ok := timeZones.Load(name); ok {
		return loc.(*time.Location)
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		sess.Logger().Debug("Unknown time zone", "time_zone", name, "error", err)
		return nil
	}
	actual, _ := timeZones.LoadOrStore(name, loc)
	return actual.(*time.Location)
}

func acceptLanguage(header string) language.Tag {
	best := language.Und
	bestQ := 0.0
	for entry := range strings.SplitSeq(header, ",") {
		value, params, _ := strings.Cut(entry, ";")
		tag := parseLanguage(strings.TrimSpace(value))
		if tag == language.Und {
			continue
		}
		q, ok := quality(params)
		if !ok || q <= bestQ {
			continue
		}
		best, bestQ = tag, q
	}
	return best
}

func quality(params string) (float64, bool) {
	for param := range strings.SplitSeq(params, ";") {
		name, value, _ := strings.Cut(param, "=")
		if !strings.EqualFold(strings.TrimSpace(name), "q") {
			continue
		}
		q, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return q, err == nil && q >= 0 && q <= 1
	}
	return 1, true
}

func parseLanguage(s string) language.Tag {
	if s == "" || len(s) > 64 {
		return language.Und
	}
	tag, err := language.Parse(s)
	if err != nil {
		return language.Und
	}
	return tag
}

func validTimeZone(s string) bool {
	return s != "" && s != "Local" && len(s) <= 64 && strings.Trim(s, timeZoneChars) == ""
}
