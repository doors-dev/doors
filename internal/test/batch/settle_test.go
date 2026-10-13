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

func settleOpen(t *testing.T) (*rod.Page, *settleFragment) {
	t.Helper()
	f := newSettleFragment()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return f
	})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	if _, err := page.Timeout(waitLimit).Element("html[data-settle-ready]"); err != nil {
		t.Fatal("page script did not register its handlers")
	}
	page.MustEval(`() => {
		const state = () => document.querySelector("#x").textContent + "|" + document.querySelector("#y").textContent
		window.settleStates = [state()]
		new MutationObserver(() => {
			const s = state()
			if (window.settleStates[window.settleStates.length - 1] !== s) {
				window.settleStates.push(s)
			}
		}).observe(document.body, { childList: true, subtree: true, characterData: true })
	}`)
	return page, f
}

func settleStates(t *testing.T, page *rod.Page) string {
	t.Helper()
	return page.MustEval(`() => window.settleStates.join(" ")`).Str()
}

func settleReturned(t *testing.T, f *settleFragment) {
	t.Helper()
	select {
	case <-f.returned:
	case <-time.After(waitLimit):
		t.Fatal("handler did not return")
	}
}

func settleFired(t *testing.T, f *settleFragment) {
	t.Helper()
	select {
	case <-f.fired:
	case <-time.After(waitLimit):
		t.Fatal("settle callback did not run")
	}
}

func settleSeen(t *testing.T, f *settleFragment) settleRead {
	t.Helper()
	select {
	case r := <-f.seen:
		return r
	case <-time.After(waitLimit):
		t.Fatal("the call did not report")
	}
	return settleRead{}
}

func settleHeld(t *testing.T, page *rod.Page, f *settleFragment, selector string, want string) {
	t.Helper()
	select {
	case <-f.fired:
		t.Fatal("expected the settle callback not to run while the batch is open")
	case <-time.After(quiet):
	}
	if got := content(t, page, selector); got != want {
		t.Fatalf("expected the update of %s held while the batch is open, got %q", selector, got)
	}
	if !hooksBusy(t, page, "#ind-a") {
		t.Fatal("expected the indicator to stay while the batch is open")
	}
}

func settleFollowUp(t *testing.T, page *rod.Page) {
	t.Helper()
	test.ClickNow(t, page, "#settle-after")
	waitContent(t, page, "#x", "x-after")
	waitContent(t, page, "#y", "y-after")
}

// A Door op issued from OnSettle in a Batch hook keeps the two-step channel: scheduled, applied, closed, with the content on the page by then.
func TestSettleDoorOpTwoStepChannel(t *testing.T) {
	page, f := settleOpen(t)
	test.ClickNow(t, page, "#settle-op")
	waitEntered(t, f.g)
	settleHeld(t, page, f, "#x", "x0")
	if got := content(t, page, "#o"); got != "o0" {
		t.Fatalf("expected the settle op not issued while the handler runs, got %q", got)
	}
	close(f.g.proceed)
	settleFired(t, f)
	var op settleOp
	select {
	case op = <-f.ops:
	case <-time.After(waitLimit):
		t.Fatal("the settle op channel did not close")
	}
	if len(op.values) != 2 || op.values[0] != nil || op.values[1] != nil {
		t.Fatalf("expected scheduled and applied nils from the settle op, got %v", op.values)
	}
	if op.page.err != nil || !op.page.ok || op.page.value != "o1" {
		t.Fatalf("expected the op content on the page when its channel closes, got %q, %v, %v", op.page.value, op.page.err, op.page.ok)
	}
	waitContent(t, page, "#settle-res", "2:<nil>,<nil>:o1")
	waitContent(t, page, "#x", "x1")
	waitContent(t, page, "#y", "y1")
	hooksWaitIdle(t, page, "#ind-a")
}

// A Call with Into issued from OnSettle in a Batch hook returns the real page value, read after the batch landed.
func TestSettleCallIntoReturnsValue(t *testing.T) {
	page, f := settleOpen(t)
	test.ClickNow(t, page, "#settle-call")
	waitEntered(t, f.g)
	settleHeld(t, page, f, "#x", "x0")
	select {
	case r := <-f.seen:
		t.Fatalf("expected no call before the handler settles, got %q, %v", r.value, r.err)
	default:
	}
	close(f.g.proceed)
	settleFired(t, f)
	if r := settleSeen(t, f); r.err != nil || !r.ok || r.value != "x1|y1" {
		t.Fatalf("expected the page value after the batch, got %q, %v, %v", r.value, r.err, r.ok)
	}
	if got := settleStates(t, page); got != "x0|y0 x1|y1" {
		t.Fatalf("expected both doors to change together, got %q", got)
	}
	hooksWaitIdle(t, page, "#ind-a")
}

// Calls issued synchronously in OnSettle of a Batch hook reach the element the batch delivers and report success.
func TestSettleSyncCallReachesBatchContent(t *testing.T) {
	page, f := settleOpen(t)
	test.ClickNow(t, page, "#settle-sync")
	waitEntered(t, f.g)
	settleHeld(t, page, f, "#d", "d0")
	close(f.g.proceed)
	settleFired(t, f)
	if r := settleSeen(t, f); r.err != nil || !r.ok || r.value != "set" {
		t.Fatalf("expected the attribute set on the delivered element, got %q, %v, %v", r.value, r.err, r.ok)
	}
	if _, err := page.Timeout(waitLimit).Element(`#settle-item[data-settle="set"]`); err != nil {
		t.Fatal("expected the delivered element with the attribute on the page")
	}
	hooksWaitIdle(t, page, "#ind-a")
}

// HoldSettle in a Batch hook keeps the batch open after the handler returns: updates issued later land together on release, and only then the indicator is released.
func TestSettleHoldKeepsBatchOpen(t *testing.T) {
	page, f := settleOpen(t)
	hooksProbe(t, page, "#ind-a")
	test.ClickNow(t, page, "#settle-hold")
	settleReturned(t, f)
	waitEntered(t, f.g)
	settleHeld(t, page, f, "#x", "x0")
	if got := hooksProbed(t, page, "#ind-a"); got != "" {
		t.Fatalf("expected the indicator to stay until release, released with %q on the page", got)
	}
	close(f.g.proceed)
	if got := hooksWaitProbed(t, page, "#ind-a"); got != "x-hold|y-hold" {
		t.Fatalf("expected both updates applied when the indicator is released, got %q", got)
	}
	settleFired(t, f)
	if got := settleStates(t, page); got != "x0|y0 x-hold|y-hold" {
		t.Fatalf("expected both doors to change together, got %q", got)
	}
	settleFollowUp(t, page)
}

// HoldSettle released without further updates still closes the batch: the held update lands and the indicator is released.
func TestSettleHoldReleaseClosesBatch(t *testing.T) {
	page, f := settleOpen(t)
	hooksProbe(t, page, "#ind-a")
	test.ClickNow(t, page, "#settle-idle")
	settleReturned(t, f)
	waitEntered(t, f.g)
	settleHeld(t, page, f, "#x", "x0")
	close(f.g.proceed)
	if got := hooksWaitProbed(t, page, "#ind-a"); got != "x-idle|y0" {
		t.Fatalf("expected the held update applied when the indicator is released, got %q", got)
	}
	settleFired(t, f)
	waitContent(t, page, "#x", "x-idle")
	settleFollowUp(t, page)
}
