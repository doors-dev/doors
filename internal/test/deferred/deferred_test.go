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

package deferred

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

const waitLimit = 5 * time.Second

// quiet is how long a test waits before asserting that something did not
// happen.
const quiet = 300 * time.Millisecond

type run struct {
	t     *testing.T
	s     *Shared
	bro   *test.Bro
	page  *rod.Page
	got   map[string][]string
	pings int
}

func open(t *testing.T, s *Shared, f func() test.Fragment) *run {
	t.Helper()
	bro := test.NewFragmentBro(browser, f)
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	t.Cleanup(s.ReleaseSkeleton)
	t.Cleanup(s.Release)
	return &run{t: t, s: s, bro: bro, page: page, got: map[string][]string{}}
}

func (r *run) click(id string) {
	r.t.Helper()
	test.ClickNow(r.t, r.page, "#"+id)
}

func (r *run) text(selector string) string {
	r.t.Helper()
	res, err := r.page.Eval(`(sel) => {
		const el = document.querySelector(sel)
		return el ? el.textContent.replace(/\s+/g, " ").trim() : "<missing>"
	}`, selector)
	if err != nil {
		r.t.Fatal("text eval: ", err)
	}
	return res.Value.Str()
}

func (r *run) expectText(selector string, want string) {
	r.t.Helper()
	if got := r.text(selector); got != want {
		r.t.Fatalf("%s: expected text %q, fact %q", selector, want, got)
	}
}

func (r *run) waitText(selector string, want string) {
	r.t.Helper()
	err := r.page.Timeout(waitLimit).Wait(rod.Eval(`(sel, want) => {
		const el = document.querySelector(sel)
		return !!el && el.textContent.replace(/\s+/g, " ").trim() === want
	}`, selector, want))
	if err != nil {
		r.t.Fatalf("%s: expected text %q, fact %q (%v)", selector, want, r.text(selector), err)
	}
}

func (r *run) count(selector string) int {
	r.t.Helper()
	res, err := r.page.Eval(`(sel) => document.querySelectorAll(sel).length`, selector)
	if err != nil {
		r.t.Fatal("count eval: ", err)
	}
	return res.Value.Int()
}

// flush updates the sentinel Door and waits for it on the page. Calls reach
// the page in order, so every call queued before is applied by then.
func (r *run) flush() {
	r.t.Helper()
	r.pings++
	r.click("sentinel-ping")
	r.waitText("#sentinel", fmt.Sprint("ping ", r.pings))
}

func (r *run) observe(selector string) {
	r.t.Helper()
	_, err := r.page.Eval(`(sel) => {
		if (window.__deferredObserver) {
			window.__deferredObserver.disconnect()
		}
		window.__deferredRecords = []
		const target = document.querySelector(sel)
		if (!target) {
			throw new Error("observe target not found: " + sel)
		}
		window.__deferredObserver = new MutationObserver((records) => {
			for (const record of records) {
				for (const node of record.addedNodes) {
					const text = (node.textContent || "").replace(/\s+/g, " ").trim()
					if (text !== "") {
						window.__deferredRecords.push(text)
					}
				}
			}
		})
		window.__deferredObserver.observe(target, { childList: true, subtree: true })
	}`, selector)
	if err != nil {
		r.t.Fatal("observe: ", err)
	}
}

// records returns the text of every node inserted into the observed area so
// far, after a flush.
func (r *run) records() []string {
	r.t.Helper()
	r.flush()
	res, err := r.page.Eval(`() => window.__deferredRecords || []`)
	if err != nil {
		r.t.Fatal("records eval: ", err)
	}
	var out []string
	for _, v := range res.Value.Arr() {
		out = append(out, v.Str())
	}
	return out
}

func (r *run) expectRecords(want ...string) {
	r.t.Helper()
	got := r.records()
	if !slices.Equal(got, want) {
		r.t.Fatalf("inserted content: expected %q, fact %q", want, got)
	}
}

func maybe(s string) string {
	return "?" + s
}

// expectRecordsMaybe is expectRecords where entries wrapped in maybe may be
// absent.
func (r *run) expectRecordsMaybe(want ...string) {
	r.t.Helper()
	got := r.records()
	i := 0
	for _, w := range want {
		text, optional := strings.CutPrefix(w, "?")
		if i < len(got) && got[i] == text {
			i++
			continue
		}
		if !optional {
			r.t.Fatalf("inserted content: expected %q, fact %q", want, got)
		}
	}
	if i != len(got) {
		r.t.Fatalf("inserted content: expected %q, fact %q", want, got)
	}
}

func (r *run) collect() {
	for {
		select {
		case res := <-r.s.results:
			key, value, _ := strings.Cut(res, ":")
			r.got[key] = append(r.got[key], value)
		default:
			return
		}
	}
}

func (r *run) result(label string) string {
	r.t.Helper()
	if list := r.got[label]; len(list) > 0 {
		r.got[label] = list[1:]
		return list[0]
	}
	deadline := time.After(waitLimit)
	for {
		select {
		case res := <-r.s.results:
			key, value, _ := strings.Cut(res, ":")
			if key == label {
				return value
			}
			r.got[key] = append(r.got[key], value)
		case <-deadline:
			r.t.Fatalf("result %q not reported", label)
		}
	}
}

func (r *run) expectResult(label string, want string) {
	r.t.Helper()
	if got := r.result(label); got != want {
		r.t.Fatalf("result %q: expected %q, fact %q", label, want, got)
	}
}

func (r *run) expectNoResult(label string) {
	r.t.Helper()
	r.collect()
	if list := r.got[label]; len(list) > 0 {
		r.t.Fatalf("result %q: expected none yet, fact %q", label, list)
	}
}

// probe waits for the op channel a handler stored under label.
func (r *run) probe(label string) <-chan error {
	r.t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		if ch, ok := r.s.Probed(label); ok {
			return ch
		}
		if time.Now().After(deadline) {
			r.t.Fatalf("op %q was not issued", label)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// hold drops stale start signals, then holds heavy renders.
func (r *run) hold() {
	for {
		select {
		case <-r.s.entered:
		default:
			r.s.Hold()
			return
		}
	}
}

func (r *run) entered(label string) {
	r.t.Helper()
	deadline := time.After(waitLimit)
	for {
		select {
		case got := <-r.s.entered:
			if got == label {
				return
			}
		case <-deadline:
			r.t.Fatalf("%q did not start rendering", label)
		}
	}
}

func (r *run) expectNotEntered(label string, d time.Duration) {
	r.t.Helper()
	deadline := time.After(d)
	for {
		select {
		case got := <-r.s.entered:
			if got == label {
				r.t.Fatalf("%q started rendering too early", label)
			}
		case <-deadline:
			return
		}
	}
}

// concurrent skips a test that needs the instance to make progress while a
// render is held: with one instance goroutine the held render blocks it.
func concurrent(t *testing.T) {
	if test.LimitMode() {
		t.Skip("a held render blocks the only instance goroutine")
	}
}

func (r *run) settle(area string, label string) {
	r.t.Helper()
	r.waitText(area, "heavy "+label)
	r.expectResult("deferred "+label, "nil,nil")
}

// A fresh blend door placed by a parent update shows its skeleton, then heavy.
func TestDeferredFreshBlendDoor(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &FreshFragment{S: s} })
	r.observe("#area")
	r.click("place-blend")
	r.settle("#area", "blend")
	r.expectRecords("skel blend", "heavy blend")
}

// A fresh inner door placed by a parent update shows its skeleton, then heavy.
func TestDeferredFreshInnerDoor(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &FreshFragment{S: s} })
	r.observe("#area")
	r.click("place-inner")
	r.settle("#area", "inner")
	r.expectRecords("skel inner", "heavy inner")
}

// The skeleton of a fresh door stays on the page while heavy is rendering.
func TestDeferredFreshDoorHeld(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &FreshFragment{S: s} })
	r.observe("#area")
	r.hold()
	r.click("place-blend")
	r.entered("blend")
	r.waitText("#area", "skel blend")
	s.Release()
	r.settle("#area", "blend")
	r.expectRecords("skel blend", "heavy blend")
}

// A door placed by an update is not part of the document, so the IsDocument
// loader defers.
func TestDeferredFreshDocumentLoader(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &FreshFragment{S: s} })
	r.observe("#area")
	r.click("place-document")
	r.settle("#area", "document")
	r.expectRecords("skel document", "heavy document")
}

// Deferring an unmounted door stores the content; placing it renders heavy.
func TestDeferredUnmountedDoor(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &FreshFragment{S: s} })
	r.click("defer-loose")
	r.expectResult("loose", "")
	if c := s.Count("heavy loose"); c != 0 {
		t.Fatal("unmounted door rendered heavy ", c, " times")
	}
	r.click("place-loose")
	r.waitText("#area", "heavy loose")
}

// A Bind update whose content defers its inner shows the skeleton, then heavy.
func TestDeferredBindInner(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &BindFragment{S: s} })
	r.settle("#area", "0")
	r.observe("#area")
	r.click("next")
	r.settle("#area", "1")
	r.expectRecords("skel 1", "heavy 1")
	r.click("next")
	r.settle("#area", "2")
	r.expectRecords("skel 1", "heavy 1", "skel 2", "heavy 2")
}

// DeferredOuter puts heavy in place of the Bind door container and leaves a
// live door that the next Bind update replaces again. A skeleton that was not
// sent before heavy was rendered is dropped.
func TestDeferredBindOuter(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &BindFragment{S: s, Outer: true} })
	r.settle("#area", "0")
	r.observe("#area")
	for _, label := range []string{"1", "2"} {
		r.click("next")
		r.settle("#area", label)
		if c := r.count("#area .skel"); c != 0 {
			t.Fatal("skeleton left on the page after heavy ", label)
		}
		if c := r.count("#area > b.heavy"); c != 1 {
			t.Fatal("expected heavy in place of the container, fact ", c, " heavy elements")
		}
	}
	r.expectRecordsMaybe(maybe("skel 1"), "heavy 1", maybe("skel 2"), "heavy 2")
	for _, label := range []string{"0", "1", "2"} {
		if c := s.Count("heavy " + label); c != 1 {
			t.Fatal("heavy ", label, " expected to render once, fact ", c)
		}
	}
}

// A Bind DeferredOuter skeleton stays on the page while heavy is rendering.
func TestDeferredBindOuterHeld(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &BindFragment{S: s, Outer: true} })
	r.settle("#area", "0")
	r.observe("#area")
	r.hold()
	r.click("next")
	r.entered("1")
	r.waitText("#area", "skel 1")
	s.Release()
	r.settle("#area", "1")
	r.expectRecords("skel 1", "heavy 1")
}

// A skeleton that was only queued when heavy finished is dropped at send and
// its op reports context.Canceled after its scheduled nil; a skeleton that was
// sent lands and its op reports success. Either way heavy lands last.
func TestDeferredMountedInner(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &MountedFragment{S: s} })
	r.waitText("#area", "initial")
	r.observe("#area")
	r.click("load")
	r.settle("#area", "mounted")
	switch load := r.result("load"); load {
	case "nil,nil":
		r.expectRecords("skel mounted", "heavy mounted")
	case "nil,context canceled":
		r.expectRecords("heavy mounted")
	default:
		t.Fatalf("result %q: expected nil,nil or nil,context canceled, fact %q", "load", load)
	}
}

// Inner with a loader on a mounted door keeps the skeleton while heavy is
// rendering.
func TestDeferredMountedInnerHeld(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &MountedFragment{S: s} })
	r.waitText("#area", "initial")
	r.observe("#area")
	r.hold()
	r.click("load")
	r.entered("mounted")
	r.waitText("#area", "skel mounted")
	s.Release()
	r.settle("#area", "mounted")
	r.expectRecords("skel mounted", "heavy mounted")
	r.expectResult("load", "nil,nil")
}

// Heavy starts rendering only once the loader content is rendered and
// scheduled: it does not start while the loader is still rendering after
// deferring, and when it starts the loader op has already reported scheduled.
// With heavy held, the skeleton reaches the page.
func TestDeferredStartGate(t *testing.T) {
	concurrent(t)
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &MountedFragment{S: s} })
	r.waitText("#area", "initial")
	r.observe("#area")
	r.hold()
	s.HoldSkeleton()
	r.click("load-gated")
	r.entered("skel gated")
	load := r.probe("load gated")
	r.expectNotEntered("gated", quiet)
	if c := s.Count("heavy gated"); c != 0 {
		t.Fatal("heavy started while the loader was rendering, ", c, " renders")
	}
	if n := len(load); n != 0 {
		t.Fatal("loader op reported ", n, " values while rendering")
	}
	r.flush()
	r.expectText("#area", "initial")
	s.ReleaseSkeleton()
	r.entered("gated")
	if n, ok := s.Queued("gated"); !ok || n == 0 {
		t.Fatal("heavy started before the loader op reported scheduled")
	}
	r.waitText("#area", "skel gated")
	s.track("load gated", load)
	r.expectResult("load gated", "nil,nil")
	r.expectNoResult("deferred gated")
	s.Release()
	r.settle("#area", "gated")
	r.expectRecords("skel gated", "heavy gated")
	if c := s.Count("heavy gated"); c != 1 {
		t.Fatal("heavy expected to render once, fact ", c)
	}
}

// A plain op issued while heavy is rendering supersedes the deferred op at
// once: the area shows the plain op's content and heavy reports
// context.Canceled.
func TestDeferredPlainOpWins(t *testing.T) {
	concurrent(t)
	cases := []struct {
		op      string
		text    string
		records []string
		heavy   int
	}{
		{op: "inner", text: "plain", records: []string{"skel mounted", "plain"}, heavy: 1},
		{op: "outer", text: "plain", records: []string{"skel mounted", "plain"}, heavy: 1},
		{op: "static", text: "plain", records: []string{"skel mounted", "plain"}, heavy: 1},
		{op: "unmount", text: "", records: []string{"skel mounted"}, heavy: 1},
		{op: "reload", text: "heavy mounted", records: []string{"skel mounted", "heavy mounted"}, heavy: 2},
	}
	for _, c := range cases {
		t.Run(c.op, func(t *testing.T) {
			s := NewShared()
			r := open(t, s, func() test.Fragment { return &MountedFragment{S: s} })
			r.waitText("#area", "initial")
			r.observe("#area")
			r.hold()
			r.click("load")
			r.entered("mounted")
			r.waitText("#area", "skel mounted")
			r.click("plain-" + c.op)
			if c.heavy > 1 {
				r.entered("mounted")
				s.Release()
			}
			r.waitText("#area", c.text)
			r.expectResult("plain", "nil,nil")
			s.Release()
			r.expectResult("deferred mounted", "context canceled")
			r.expectRecords(c.records...)
			r.expectText("#area", c.text)
			if n := s.Count("heavy mounted"); n != c.heavy {
				t.Fatal("heavy expected to render ", c.heavy, " times, fact ", n)
			}
		})
	}
}

// A newer DeferredInner issued while heavy is rendering waits for it: the area
// keeps the skeleton and the newer op reports nothing until heavy is done,
// then the newer content lands. Heavy lands or is dropped at send.
func TestDeferredNewerDeferredWaits(t *testing.T) {
	concurrent(t)
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &MountedFragment{S: s} })
	r.waitText("#area", "initial")
	r.observe("#area")
	r.hold()
	r.click("load")
	r.entered("mounted")
	r.waitText("#area", "skel mounted")
	r.click("newer-deferred")
	newer := r.probe("newer")
	time.Sleep(quiet)
	r.flush()
	r.expectText("#area", "skel mounted")
	if n := len(newer); n != 0 {
		t.Fatal("newer op reported ", n, " values while heavy was rendering")
	}
	r.expectNoResult("deferred mounted")
	s.Release()
	s.track("newer", newer)
	r.waitText("#area", "newer")
	r.expectResult("newer", "nil,nil")
	if got := r.result("deferred mounted"); got != "nil,nil" && got != "nil,context canceled" {
		t.Fatalf("result %q: expected nil,nil or nil,context canceled, fact %q", "deferred mounted", got)
	}
	r.expectRecordsMaybe("skel mounted", maybe("heavy mounted"), "newer")
	r.expectText("#area", "newer")
	if c := s.Count("heavy mounted"); c != 1 {
		t.Fatal("heavy expected to render once, fact ", c)
	}
}

// The page HTML of IsDocument loaders carries heavy and no skeleton.
func TestDeferredDocumentHTML(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &ReloadFragment{S: s} })
	for _, label := range []string{"blend", "outer", "inner"} {
		r.waitText("#area-"+label, "heavy "+label)
	}
	r.settle("#area-plain", "plain")
	res, err := http.Get(test.Host + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	for _, label := range []string{"blend", "outer", "inner"} {
		if !strings.Contains(html, "heavy "+label) {
			t.Fatal("page html has no heavy ", label)
		}
		if strings.Contains(html, "skel "+label) {
			t.Fatal("page html has skeleton ", label)
		}
		if c := s.Count("deferred " + label); c != 0 {
			t.Fatal("document loader ", label, " deferred ", c, " times")
		}
	}
	if !strings.Contains(html, "skel plain") || strings.Contains(html, "heavy plain") {
		t.Fatal("page html of the plain loader must carry the skeleton only")
	}
}

// Reloading a blend door whose content is an IsDocument loader shows the
// skeleton, then heavy.
func TestDeferredReloadBlend(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &ReloadFragment{S: s} })
	r.waitText("#area-blend", "heavy blend")
	r.observe("#area-blend")
	r.click("reload-blend")
	r.settle("#area-blend", "blend")
	r.expectResult("reload blend", "nil,nil")
	r.expectRecords("skel blend", "heavy blend")
}

// Reloading an outer door whose content is an IsDocument loader shows the
// skeleton, then heavy.
func TestDeferredReloadOuter(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &ReloadFragment{S: s} })
	r.waitText("#area-outer", "heavy outer")
	r.observe("#area-outer")
	r.click("reload-outer")
	r.settle("#area-outer", "outer")
	r.expectResult("reload outer", "nil,nil")
	r.expectRecords("skel outer", "heavy outer")
}

// Reloading an inner door whose content is an IsDocument loader keeps the
// skeleton while heavy is rendering.
func TestDeferredReloadInnerHeld(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &ReloadFragment{S: s} })
	r.waitText("#area-inner", "heavy inner")
	r.observe("#area-inner")
	r.hold()
	r.click("reload-inner")
	r.entered("inner")
	if c := s.Count("heavy inner"); c != 2 {
		t.Fatal("expected the page render and the deferred render of heavy, fact ", c)
	}
	r.waitText("#area-inner", "skel inner")
	s.Release()
	r.settle("#area-inner", "inner")
	r.expectResult("reload inner", "nil,nil")
	r.expectRecords("skel inner", "heavy inner")
}

// Once the deferred content is applied, Reload rerenders it without the
// loader.
func TestDeferredReloadAfterApplied(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &ReloadFragment{S: s} })
	r.settle("#area-plain", "plain")
	r.waitText("#area-blend", "heavy blend")
	r.click("reload-blend")
	r.settle("#area-blend", "blend")
	r.expectResult("reload blend", "nil,nil")
	r.observe("#area-plain")
	before := s.Count("heavy plain")
	r.click("reload-plain")
	r.expectResult("reload plain", "nil,nil")
	r.expectRecords("heavy plain")
	if c := s.Count("deferred plain"); c != 1 {
		t.Fatal("plain loader expected to defer once, fact ", c)
	}
	if c := s.Count("heavy plain"); c != before+1 {
		t.Fatal("plain heavy expected one more render, fact ", c-before)
	}
	r.observe("#area-blend")
	before = s.Count("heavy blend")
	r.click("reload-blend")
	r.expectResult("reload blend", "nil,nil")
	r.expectRecords("heavy blend")
	if c := s.Count("deferred blend"); c != 1 {
		t.Fatal("blend loader expected to defer once, fact ", c)
	}
	if c := s.Count("heavy blend"); c != before+1 {
		t.Fatal("blend heavy expected one more render, fact ", c-before)
	}
}

// DeferredStatic replaces the door container with heavy and later operations
// only change the stored state.
func TestDeferredStatic(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &StaticFragment{S: s} })
	r.settle("#area", "static")
	test.TestMustNot(t, r.page, "#area .door")
	r.observe("#area")
	r.click("after")
	r.expectResult("after", "")
	r.expectRecords()
	r.expectText("#area", "heavy static")
}

// A loader rendered by the root page gets an error and keeps its skeleton.
func TestDeferredRoot(t *testing.T) {
	s := NewShared()
	r := open(t, s, func() test.Fragment { return &RootFragment{S: s} })
	r.expectResult("deferred root", "root door cannot be deferred")
	r.waitText("#area", "skel root")
	if c := s.Count("heavy root"); c != 0 {
		t.Fatal("root heavy rendered ", c, " times")
	}
}
