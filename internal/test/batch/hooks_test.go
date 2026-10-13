// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package batch

import (
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

const hooksProbeInstall = `(ind, selectors) => {
	const el = document.querySelector(ind)
	window.hooksProbe = window.hooksProbe || {}
	window.hooksProbe[ind] = ""
	let on = el.hasAttribute("data-busy")
	const observer = new MutationObserver(() => {
		if (el.hasAttribute("data-busy")) {
			on = true
			return
		}
		if (!on) {
			return
		}
		window.hooksProbe[ind] = selectors.map((s) => document.querySelector(s).textContent).join("|")
		observer.disconnect()
	})
	observer.observe(el, { attributes: true })
}`

const hooksProbeRead = `(ind) => window.hooksProbe[ind]`

func hooksOpen(t *testing.T) (*rod.Page, *gate, *gate) {
	t.Helper()
	ga, gb := newGate(), newGate()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &hooksFragment{
			ga: ga,
			gb: gb,
		}
	})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	return page, ga, gb
}

func hooksProbe(t *testing.T, page *rod.Page, indicator string) {
	t.Helper()
	if _, err := page.Eval(hooksProbeInstall, indicator, []string{"#x", "#y"}); err != nil {
		t.Fatal("probe install failed: ", err)
	}
}

func hooksProbed(t *testing.T, page *rod.Page, indicator string) string {
	t.Helper()
	res, err := page.Eval(hooksProbeRead, indicator)
	if err != nil {
		t.Fatal("probe read failed: ", err)
	}
	return res.Value.Str()
}

func hooksWaitProbed(t *testing.T, page *rod.Page, indicator string) string {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		if got := hooksProbed(t, page, indicator); got != "" {
			return got
		}
		if time.Now().After(deadline) {
			t.Fatalf("indicator on %s was not released", indicator)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func hooksBusy(t *testing.T, page *rod.Page, indicator string) bool {
	t.Helper()
	el, err := page.Timeout(waitLimit).Element(indicator)
	if err != nil {
		t.Fatalf("element %s not found", indicator)
	}
	attr, err := el.Attribute("data-busy")
	if err != nil {
		t.Fatalf("element %s attribute read failed", indicator)
	}
	return attr != nil
}

func hooksWaitIdle(t *testing.T, page *rod.Page, indicator string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for hooksBusy(t, page, indicator) {
		if time.Now().After(deadline) {
			t.Fatalf("indicator on %s was not released", indicator)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func hooksHeld(t *testing.T, page *rod.Page, indicator string) {
	t.Helper()
	time.Sleep(quiet)
	if !hooksBusy(t, page, indicator) {
		t.Fatalf("expected the indicator on %s to stay while the handler runs", indicator)
	}
	if got := content(t, page, "#x"); got != "x0" {
		t.Fatalf("expected the first update held until the handler finishes, got %q", got)
	}
}

func hooksSplit(t *testing.T, page *rod.Page, g *gate, tag string) {
	t.Helper()
	waitEntered(t, g)
	hooksHeld(t, page, "#ind-a")
	close(g.proceed)
	if got := hooksWaitProbed(t, page, "#ind-a"); got != "x-"+tag+"|y-"+tag {
		t.Fatalf("expected both updates applied when the indicator is released, got %q", got)
	}
	waitContent(t, page, "#x", "x-"+tag)
	waitContent(t, page, "#y", "y-"+tag)
}

// The indicator of a Batch hook is released only once the held updates are on the page.
func TestHooksIndicatorAfterBatch(t *testing.T) {
	page, ga, _ := hooksOpen(t)
	hooksProbe(t, page, "#ind-a")
	test.ClickNow(t, page, "#batch-a")
	hooksSplit(t, page, ga, "a")
}

// Without an open batch the indicator of a plain hook is released once its update is on the page.
func TestHooksPlainIndicatorAfterUpdate(t *testing.T) {
	page, _, _ := hooksOpen(t)
	hooksProbe(t, page, "#ind-b")
	test.ClickNow(t, page, "#plain-b")
	if got := hooksWaitProbed(t, page, "#ind-b"); got != "x-plain|y0" {
		t.Fatalf("expected the update applied when the indicator is released, got %q", got)
	}
}

// The indicator of a plain hook that updates a door held by an open batch stays until that update is on the page.
func TestHooksPlainIndicatorBehindBatch(t *testing.T) {
	page, ga, _ := hooksOpen(t)
	test.ClickNow(t, page, "#batch-a")
	waitEntered(t, ga)
	hooksProbe(t, page, "#ind-b")
	test.ClickNow(t, page, "#plain-b")
	time.Sleep(quiet)
	if got := content(t, page, "#x"); got != "x0" {
		t.Fatalf("expected the update of a held door to wait, got %q", got)
	}
	if got := hooksProbed(t, page, "#ind-b"); got != "" {
		t.Fatalf("expected the indicator to stay until the update is applied, released with %q on the page", got)
	}
	close(ga.proceed)
	if got := hooksWaitProbed(t, page, "#ind-b"); got != "x-plain|y-a" {
		t.Fatalf("expected the update applied when the indicator is released, got %q", got)
	}
	waitContent(t, page, "#x", "x-plain")
	waitContent(t, page, "#y", "y-a")
	hooksWaitIdle(t, page, "#ind-a")
}

// Two open batches touching the same doors in opposite order both apply, the later calls win and nothing hangs.
func TestHooksCrossedBatches(t *testing.T) {
	for _, first := range []string{"a", "b"} {
		t.Run(first+"-first", func(t *testing.T) {
			page, ga, gb := hooksOpen(t)
			test.ClickNow(t, page, "#batch-a")
			waitEntered(t, ga)
			test.ClickNow(t, page, "#batch-b")
			waitEntered(t, gb)
			time.Sleep(quiet)
			one, other, done := ga, gb, "#ind-a"
			if first == "b" {
				one, other, done = gb, ga, "#ind-b"
			}
			close(one.proceed)
			time.Sleep(quiet)
			if x, y := content(t, page, "#x"), content(t, page, "#y"); x != "x0" || y != "y0" {
				t.Fatalf("expected the finished batch to wait behind the open one, got %q and %q", x, y)
			}
			if !hooksBusy(t, page, done) {
				t.Fatalf("expected the indicator on %s to stay until its batch is applied", done)
			}
			close(other.proceed)
			waitContent(t, page, "#x", "x-b")
			waitContent(t, page, "#y", "y-a")
			hooksWaitIdle(t, page, "#ind-a")
			hooksWaitIdle(t, page, "#ind-b")
			test.ClickNow(t, page, "#after")
			waitContent(t, page, "#x", "x-after")
			waitContent(t, page, "#y", "y-after")
		})
	}
}

// A panic in a Batch hook ends the instance: the indicator is released, the held update is dropped and the page reloads clean.
func TestHooksPanicReloadsPage(t *testing.T) {
	page, ga, _ := hooksOpen(t)
	if _, err := page.Eval(`() => { window.hooksMark = true }`); err != nil {
		t.Fatal("mark failed: ", err)
	}
	test.ClickNow(t, page, "#panic")
	waitEntered(t, ga)
	hooksHeld(t, page, "#ind-a")
	close(ga.proceed)
	deadline := time.Now().Add(waitLimit)
	for {
		res, err := page.Eval(`() => window.hooksMark !== true && document.readyState === "complete"`)
		if err == nil && res.Value.Bool() {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("expected the page to reload after the panic")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if got := content(t, page, "#x"); got != "x0" {
		t.Fatalf("expected a clean page after the reload, got %q", got)
	}
	if hooksBusy(t, page, "#ind-a") {
		t.Fatal("expected no indicator after the reload")
	}
	test.ClickNow(t, page, "#after")
	waitContent(t, page, "#x", "x-after")
}

// A Batch hook that answers with an error status still closes its batch: the $hook promise rejects and the update made before it lands.
func TestHooksErrorStatusClosesBatch(t *testing.T) {
	page, ga, _ := hooksOpen(t)
	test.ClickNow(t, page, "#status")
	waitEntered(t, ga)
	hooksHeld(t, page, "#ind-a")
	close(ga.proceed)
	waitContent(t, page, "#x", "x-status")
	hooksWaitIdle(t, page, "#ind-a")
	if got := content(t, page, "#seen"); !strings.HasPrefix(got, "rejected:") {
		t.Fatalf("expected the hook promise rejected, got %q", got)
	}
	test.ClickNow(t, page, "#after")
	waitContent(t, page, "#x", "x-after")
}

// Batch on a form submit holds the updates of the handler until it finishes.
func TestHooksSubmitBatch(t *testing.T) {
	page, ga, _ := hooksOpen(t)
	hooksProbe(t, page, "#ind-a")
	test.ClickNow(t, page, "#submit")
	hooksSplit(t, page, ga, "form")
}

// Batch on a key event holds the updates of the handler until it finishes.
func TestHooksKeyBatch(t *testing.T) {
	page, ga, _ := hooksOpen(t)
	hooksProbe(t, page, "#ind-a")
	el, err := page.Timeout(waitLimit).Element("#key")
	if err != nil {
		t.Fatal("element #key not found")
	}
	if err := el.Type(input.KeyK); err != nil {
		t.Fatal("typing failed: ", err)
	}
	hooksSplit(t, page, ga, "k")
}

// Batch on a hook called from JavaScript holds the updates, and the $hook promise settles after they are applied.
func TestHooksJsBatch(t *testing.T) {
	page, ga, _ := hooksOpen(t)
	hooksProbe(t, page, "#ind-a")
	test.ClickNow(t, page, "#js")
	hooksSplit(t, page, ga, "js")
	waitContent(t, page, "#seen", "done:x-js:y-js")
}
