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
	"context"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"golang.org/x/text/language"
)

func TestAcceptLanguage(t *testing.T) {
	cases := []struct {
		header string
		want   string
	}{
		{"", "und"},
		{"de-DE", "de-DE"},
		{"de-DE,en;q=0.9", "de-DE"},
		{"en;q=0.5, fr-CH, de;q=0.9", "fr-CH"},
		{"fr-CH, it", "fr-CH"},
		{"en;q=0.8, de;q=0.8", "en"},
		{"*, en;q=0.5", "en"},
		{"*", "und"},
		{"de;q=0, en;q=0.1", "en"},
		{"de;q=0", "und"},
		{"de;q=abc, en;q=0.2", "en"},
		{"de;q=2, en;q=0.2", "en"},
		{"de;q=NaN, en;q=0.2", "en"},
		{" zh-Hant-TW ; Q=0.7 , en;q=0.6", "zh-Hant-TW"},
		{"de-DE;level=1;q=0.4, en;q=0.3", "de-DE"},
		{"1de, en;q=0.1", "en"},
		{"garbage!!, en;q=0.1", "en"},
		{"en-us", "en-US"},
		{"zh-hant-tw", "zh-Hant-TW"},
		{"en_US, fr;q=0.3", "en-US"},
		{"iw, en;q=0.5", "he"},
	}
	for _, c := range cases {
		if got := acceptLanguage(c.header).String(); got != c.want {
			t.Errorf("acceptLanguage(%q) = %q, want %q", c.header, got, c.want)
		}
	}
}

func TestParseLanguage(t *testing.T) {
	cases := []struct {
		value string
		want  string
	}{
		{"en", "en"},
		{"de-DE", "de-DE"},
		{"en-us", "en-US"},
		{"EN-GB", "en-GB"},
		{"sr-latn-rs", "sr-Latn-RS"},
		{"es-419", "es-419"},
		{"iw", "he"},
		{"", "und"},
		{"und", "und"},
		{"*", "und"},
		{"-", "und"},
		{"en--US", "und"},
		{"1en", "und"},
		{"en-toolongsub", "und"},
		{"dé", "und"},
		{"en US", "und"},
		{"xx", "und"},
		{"en-" + strings.Repeat("a1b2c3d4-", 7) + "x", "und"},
	}
	for _, c := range cases {
		if got := parseLanguage(c.value).String(); got != c.want {
			t.Errorf("parseLanguage(%q) = %q, want %q", c.value, got, c.want)
		}
	}
}

func TestValidTimeZone(t *testing.T) {
	valid := []string{"UTC", "Europe/Berlin", "America/Argentina/Buenos_Aires", "Etc/GMT+5", "Etc/GMT-14", "America/Port-au-Prince"}
	invalid := []string{"", "Local", "Europe/Berlin ", "Europe.Berlin", "Europe\\Berlin", "<script>", strings.Repeat("a", 65)}
	for _, s := range valid {
		if !validTimeZone(s) {
			t.Errorf("validTimeZone(%q) = false, want true", s)
		}
	}
	for _, s := range invalid {
		if validTimeZone(s) {
			t.Errorf("validTimeZone(%q) = true, want false", s)
		}
	}
}

func TestLoadTimeZone(t *testing.T) {
	sess := NewSession(newSessionTestApp(), "time-zone-session")
	defer sess.Kill()
	first := sess.loadTimeZone("Europe/Berlin")
	if first == nil || first.String() != "Europe/Berlin" {
		t.Fatalf("Europe/Berlin: got %v", first)
	}
	if second := sess.loadTimeZone("Europe/Berlin"); second != first {
		t.Fatal("expected the same location for the same name")
	}
	for _, name := range []string{"", "Local", "Mars/Olympus_Mons", "../etc/passwd", "/etc/localtime"} {
		if loc := sess.loadTimeZone(name); loc != nil {
			t.Errorf("loadTimeZone(%q) = %v, want nil", name, loc)
		}
	}
}

// The session source compares Locale with ==, so TimeZone is compared by
// pointer. Equal reports are suppressed only because loadTimeZone hands out
// one shared *time.Location per zone name, across sessions and concurrent
// loads.
func TestLocaleEqualityReliesOnTimeZoneCache(t *testing.T) {
	const zone = "Pacific/Chatham"
	a := NewSession(newSessionTestApp(), "locale-cache-a")
	defer a.Kill()
	b := NewSession(newSessionTestApp(), "locale-cache-b")
	defer b.Kill()
	locs := make([]*time.Location, 16)
	var wg sync.WaitGroup
	for i := range locs {
		sess := a
		if i%2 == 1 {
			sess = b
		}
		wg.Go(func() {
			locs[i] = sess.loadTimeZone(zone)
		})
	}
	wg.Wait()
	for i, loc := range locs {
		if loc == nil || loc != locs[0] {
			t.Fatalf("load %d returned a different location: %p, want %p", i, loc, locs[0])
		}
	}
	ctx := context.Background()
	if !a.locale.Sub(ctx, func(context.Context, common.Locale) bool { return false }) {
		t.Fatal("sub refused")
	}
	a.ReportLocale("fr-FR", zone)
	same := common.Locale{Language: language.MustParse("fr-FR"), TimeZone: b.loadTimeZone(zone), Synced: true}
	if _, ok := <-a.locale.Update(ctx, same); ok {
		t.Fatal("an equal locale propagated")
	}
	fresh, err := time.LoadLocation(zone)
	if err != nil {
		t.Fatal(err)
	}
	if err, ok := <-a.locale.Update(ctx, common.Locale{Language: same.Language, TimeZone: fresh, Synced: true}); !ok || err != nil {
		t.Fatalf("an uncached location of the same zone was suppressed (%v, %v): equality is no longer by pointer", err, ok)
	}
}

func TestSessionLocale(t *testing.T) {
	sess := NewSession(newSessionTestApp(), "locale-session")
	defer sess.Kill()
	if got := sess.Locale().Get(); got != (common.Locale{Language: language.Und, TimeZone: time.Local}) {
		t.Fatalf("initial: %+v", got)
	}
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Language", "de-de,en;q=0.9")
	sess.initLocale(r)
	if got := sess.Locale().Get(); got != (common.Locale{Language: language.MustParse("de-DE"), TimeZone: time.Local}) {
		t.Fatalf("after header: %+v", got)
	}
	sess.ReportLocale("fr-fr", "Asia/Tokyo")
	got := sess.Locale().Get()
	if got.Language != language.MustParse("fr-FR") || got.TimeZone.String() != "Asia/Tokyo" || !got.Synced {
		t.Fatalf("after report: %+v", got)
	}
	want := got
	sess.ReportLocale("fr-FR", "Asia/Tokyo")
	if got := sess.Locale().Get(); got != want {
		t.Fatalf("equal report changed the locale: %+v", got)
	}
	sess.initLocale(r)
	if got := sess.Locale().Get(); got != want {
		t.Fatalf("header overrode the report: %+v", got)
	}
	sess.ReportLocale("garbage!!", "Local")
	if got := sess.Locale().Get(); got != want {
		t.Fatalf("invalid report changed the locale: %+v", got)
	}
	sess.ReportLocale("", "Mars/Olympus_Mons")
	if got := sess.Locale().Get(); got != want {
		t.Fatalf("unknown time zone changed the locale: %+v", got)
	}
	sess.ReportLocale("", "Europe/Berlin")
	got = sess.Locale().Get()
	if got.Language != want.Language || got.TimeZone.String() != "Europe/Berlin" || !got.Synced {
		t.Fatalf("after time zone report: %+v", got)
	}
}

func TestSessionLocaleSynced(t *testing.T) {
	sess := NewSession(newSessionTestApp(), "locale-synced")
	defer sess.Kill()
	sess.ReportLocale("garbage!!", "Mars/Olympus_Mons")
	if got := sess.Locale().Get(); got != (common.Locale{Language: language.Und, TimeZone: time.Local, Synced: true}) {
		t.Fatalf("unusable first report: %+v, want und, time.Local, synced", got)
	}
	sess.ReportLocale("de-DE", "Asia/Tokyo")
	got := sess.Locale().Get()
	if got.Language != language.MustParse("de-DE") || got.TimeZone.String() != "Asia/Tokyo" || !got.Synced {
		t.Fatalf("known zone did not replace time.Local: %+v", got)
	}
	want := got
	sess.ReportLocale("", "")
	if got := sess.Locale().Get(); got != want {
		t.Fatalf("empty report changed the locale: %+v", got)
	}
}
