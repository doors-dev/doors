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
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func groupOpen(t *testing.T) (*rod.Page, *groupFragment) {
	t.Helper()
	f := newGroupFragment()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return f
	})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	page.MustEval(`() => {
		const state = () => document.querySelector("#x").textContent + "|" + document.querySelector("#y").textContent
		window.groupStates = [state()]
		new MutationObserver(() => {
			const s = state()
			if (window.groupStates[window.groupStates.length - 1] !== s) {
				window.groupStates.push(s)
			}
		}).observe(document.body, { childList: true, subtree: true, characterData: true })
	}`)
	return page, f
}

func groupNotFired(t *testing.T, f *groupFragment) {
	t.Helper()
	select {
	case <-f.fired:
		t.Fatal("expected on not to run before its updates are on the page")
	default:
	}
}

func groupFired(t *testing.T, page *rod.Page, f *groupFragment, x string, y string) {
	t.Helper()
	select {
	case <-f.fired:
	case <-time.After(waitLimit):
		t.Fatal("on did not run")
	}
	if gotX, gotY := content(t, page, "#x"), content(t, page, "#y"); gotX != x || gotY != y {
		t.Fatalf("expected both updates on the page when on runs, got %q and %q", gotX, gotY)
	}
}

// The indicator of a plain hook stays until the updates of its OnBatch, held behind another open batch, are on the page.
func TestGroupIndicatorWaitsForBlockedOnBatch(t *testing.T) {
	page, f := groupOpen(t)
	test.ClickNow(t, page, "#group-holder")
	waitEntered(t, f.g)
	time.Sleep(quiet)
	hooksProbe(t, page, "#ind")
	test.ClickNow(t, page, "#group-plain")
	time.Sleep(quiet)
	if x, y := content(t, page, "#x"), content(t, page, "#y"); x != "x0" || y != "y0" {
		t.Fatalf("expected the updates held behind the open batch, got %q and %q", x, y)
	}
	if got := hooksProbed(t, page, "#ind"); got != "" {
		t.Fatalf("expected the indicator to stay while the updates are held, released at %q", got)
	}
	if !hooksBusy(t, page, "#ind") {
		t.Fatal("expected the indicator to be on while the updates are held")
	}
	groupNotFired(t, f)
	close(f.g.proceed)
	if got := hooksWaitProbed(t, page, "#ind"); got != "x-plain|y-plain" {
		t.Fatalf("expected both updates applied when the indicator is released, got %q", got)
	}
	groupFired(t, page, f, "x-plain", "y-plain")
}

// OnBatch in an OnSettle callback of a Batch hook lands its updates together and runs on after them.
func TestGroupOnBatchInSettle(t *testing.T) {
	page, f := groupOpen(t)
	test.ClickNow(t, page, "#group-settle")
	waitEntered(t, f.g)
	time.Sleep(quiet)
	if got := content(t, page, "#x"); got != "x0" {
		t.Fatalf("expected the first update held until the second is ready, got %q", got)
	}
	groupNotFired(t, f)
	close(f.g.proceed)
	groupFired(t, page, f, "x-settle", "y-settle")
	if got := page.MustEval(`() => window.groupStates.join(" ")`).Str(); got != "x0|y0 x-settle|y-settle" {
		t.Fatalf("expected both doors to change together, got %q", got)
	}
}

// The indicator of a plain hook stays until the updates of an OnBatch in its OnSettle callback are on the page.
func TestGroupIndicatorWaitsForOnBatchInSettle(t *testing.T) {
	page, f := groupOpen(t)
	hooksProbe(t, page, "#ind")
	test.ClickNow(t, page, "#group-settle-plain")
	waitEntered(t, f.g)
	waitEntered(t, f.g)
	time.Sleep(quiet)
	if got := hooksProbed(t, page, "#ind"); got != "" {
		t.Fatalf("expected the indicator to stay while the updates render, released at %q", got)
	}
	if !hooksBusy(t, page, "#ind") {
		t.Fatal("expected the indicator to be on while the updates render")
	}
	groupNotFired(t, f)
	close(f.g.proceed)
	if got := hooksWaitProbed(t, page, "#ind"); got != "x-late|y-late" {
		t.Fatalf("expected both updates applied when the indicator is released, got %q", got)
	}
	groupFired(t, page, f, "x-late", "y-late")
}

// The indicator of a plain hook stays until a second OnBatch, opened from the settle callback of a first one after the report arrived, is on the page.
func TestGroupIndicatorWaitsForNestedOnBatch(t *testing.T) {
	page, f := groupOpen(t)
	hooksProbe(t, page, "#ind")
	test.ClickNow(t, page, "#group-nested")
	waitEntered(t, f.g)
	time.Sleep(quiet)
	if !hooksBusy(t, page, "#ind") {
		t.Fatal("expected the indicator to be on while the first update renders")
	}
	close(f.g.proceed)
	waitEntered(t, f.g2)
	waitContent(t, page, "#x", "x-nested")
	time.Sleep(quiet)
	if got := hooksProbed(t, page, "#ind"); got != "" {
		t.Fatalf("expected the indicator to stay while the second update renders, released at %q", got)
	}
	if !hooksBusy(t, page, "#ind") {
		t.Fatal("expected the indicator to be on while the second update renders")
	}
	if got := content(t, page, "#y"); got != "y0" {
		t.Fatalf("expected the second update not applied yet, got %q", got)
	}
	close(f.g2.proceed)
	if got := hooksWaitProbed(t, page, "#ind"); got != "x-nested|y-nested" {
		t.Fatalf("expected both updates applied when the indicator is released, got %q", got)
	}
	groupFired(t, page, f, "x-nested", "y-nested")
	groupFired(t, page, f, "x-nested", "y-nested")
}
