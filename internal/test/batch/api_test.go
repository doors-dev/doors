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

func apiOpen(t *testing.T) (*rod.Page, *apiFragment) {
	t.Helper()
	f := newAPIFragment()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return f
	})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	if _, err := page.Timeout(waitLimit).Element("html[data-api-ready]"); err != nil {
		t.Fatal("page script did not register its handlers")
	}
	page.MustEval(`() => {
		const state = () => document.querySelector("#a").textContent + "|" + document.querySelector("#b").textContent
		window.apiStates = [state()]
		new MutationObserver(() => {
			const s = state()
			if (window.apiStates[window.apiStates.length - 1] !== s) {
				window.apiStates.push(s)
			}
		}).observe(document.body, { childList: true, subtree: true, characterData: true })
	}`)
	return page, f
}

func apiStates(t *testing.T, page *rod.Page) string {
	t.Helper()
	return page.MustEval(`() => window.apiStates.join(" ")`).Str()
}

func apiNotFired(t *testing.T, f *apiFragment) {
	t.Helper()
	select {
	case <-f.fired:
		t.Fatal("expected on not to run while the batch is open")
	case <-time.After(quiet):
	}
}

func apiFired(t *testing.T, f *apiFragment) {
	t.Helper()
	select {
	case <-f.fired:
	case <-time.After(waitLimit):
		t.Fatal("on did not run")
	}
}

func apiSeen(t *testing.T, f *apiFragment) apiRead {
	t.Helper()
	select {
	case r := <-f.seen:
		return r
	case <-time.After(waitLimit):
		t.Fatal("the call did not report")
	}
	return apiRead{}
}

func apiNotSeen(t *testing.T, f *apiFragment) {
	t.Helper()
	select {
	case r := <-f.seen:
		t.Fatalf("expected the call to wait behind the batch, got %q, %v", r.value, r.err)
	case <-time.After(quiet):
	}
}

func apiHeld(t *testing.T, page *rod.Page, f *apiFragment) {
	t.Helper()
	apiNotFired(t, f)
	if got := content(t, page, "#a"); got != "a0" {
		t.Fatalf("expected the first update held while the batch is open, got %q", got)
	}
}

func apiApplied(t *testing.T, page *rod.Page, f *apiFragment) {
	t.Helper()
	apiFired(t, f)
	if a, b := content(t, page, "#a"), content(t, page, "#b"); a != "a1" || b != "b1" {
		t.Fatalf("expected both updates on the page when on runs, got %q and %q", a, b)
	}
	if r := apiSeen(t, f); r.err != nil || !r.ok || r.value != "a1|b1" {
		t.Fatalf("expected on to read both updates from the page, got %q, %v, %v", r.value, r.err, r.ok)
	}
	if got := apiStates(t, page); got != "a0|b0 a1|b1" {
		t.Fatalf("expected both doors to change together, got %q", got)
	}
	select {
	case <-f.fired:
		t.Fatal("expected on to run once")
	case <-time.After(quiet):
	}
}

// OnBatch in a plain hook holds the first op until the second is issued, lands both together and runs on after them.
func TestAPIOnBatchPlainHook(t *testing.T) {
	page, f := apiOpen(t)
	test.ClickNow(t, page, "#api-plain")
	waitEntered(t, f.g)
	apiHeld(t, page, f)
	close(f.g.proceed)
	apiApplied(t, page, f)
}

// OnBatch in a Batch hook joins the hook batch: on waits for the handler and sees its later update too.
func TestAPIOnBatchJoinsBatchHook(t *testing.T) {
	page, f := apiOpen(t)
	test.ClickNow(t, page, "#api-join")
	waitEntered(t, f.g)
	apiHeld(t, page, f)
	close(f.g.proceed)
	apiApplied(t, page, f)
}

// OnBatch from a goroutine with a detached context opens its own batch, lands the ops together and reports.
func TestAPIOnBatchDetachedGoroutine(t *testing.T) {
	page, f := apiOpen(t)
	close(f.start)
	waitEntered(t, f.g)
	apiHeld(t, page, f)
	close(f.g.proceed)
	apiApplied(t, page, f)
}

// A call issued from OnSettle in a Batch hook reaches the element the batch delivers.
func TestAPIOnSettleCallSeesBatchContent(t *testing.T) {
	page, f := apiOpen(t)
	test.ClickNow(t, page, "#api-settle")
	waitEntered(t, f.g)
	time.Sleep(quiet)
	if got := content(t, page, "#d"); got != "d0" {
		t.Fatalf("expected the delivered content held while the handler runs, got %q", got)
	}
	if got := content(t, page, "#api-seen"); got != "none" {
		t.Fatalf("expected the settle call not to run before the batch, got %q", got)
	}
	close(f.g.proceed)
	waitContent(t, page, "#api-seen", "delivered")
	if _, err := page.Timeout(waitLimit).Element(`#api-item[data-api="set"]`); err != nil {
		t.Fatal("expected the settle call to set the attribute on the delivered element")
	}
}

func apiScheduledOnly(t *testing.T, page *rod.Page, f *apiFragment, button string) {
	t.Helper()
	test.ClickNow(t, page, button)
	waitContent(t, page, "#b", "b1")
	waitContent(t, page, "#a", "a1")
	res := <-f.guard
	if len(res.call) != 1 || res.call[0] != nil || res.value != "" {
		t.Fatalf("expected only the scheduled nil from a call with Into and dst untouched, got %v, %q", res.call, res.value)
	}
	if len(res.plain) != 1 || res.plain[0] != nil {
		t.Fatalf("expected only the scheduled nil from a plain call, got %v", res.plain)
	}
	if len(res.inner) != 1 || res.inner[0] != nil {
		t.Fatalf("expected only the scheduled nil from a door op, got %v", res.inner)
	}
}

// In a Batch hook a Call with Into, a plain Call and a Door op each send the scheduled nil and close; dst stays unwritten.
func TestAPIBatchHookScheduledOnly(t *testing.T) {
	page, f := apiOpen(t)
	apiScheduledOnly(t, page, f, "#api-guard")
}

// In a plain hook a Call with Into, a plain Call and a Door op each send the scheduled nil and close; dst stays unwritten.
func TestAPIPlainHookScheduledOnly(t *testing.T) {
	page, f := apiOpen(t)
	apiScheduledOnly(t, page, f, "#api-guard-plain")
}

// A detached Call from an untouched door returns the real page value while the batch is still open.
func TestAPIDetachedCallPassesOpenBatch(t *testing.T) {
	page, f := apiOpen(t)
	test.ClickNow(t, page, "#api-detached")
	if r := apiSeen(t, f); r.err != nil || !r.ok || r.value != "a0|b0" {
		t.Fatalf("expected the page value from before the batch, got %q, %v, %v", r.value, r.err, r.ok)
	}
	waitEntered(t, f.g)
	if got := content(t, page, "#a"); got != "a0" {
		t.Fatalf("expected the batch still open after the detached call, got %q", got)
	}
	close(f.g.proceed)
	waitContent(t, page, "#a", "a1")
	waitContent(t, page, "#b", "b1")
}

// A detached Call on a door the batch holds waits for the batch to close and then runs in issue order, ahead of the later updates.
func TestAPIDetachedCallWaitsOnHeldDoor(t *testing.T) {
	page, f := apiOpen(t)
	test.ClickNow(t, page, "#api-held")
	waitEntered(t, f.g)
	apiNotSeen(t, f)
	close(f.g.proceed)
	if r := apiSeen(t, f); r.err != nil || !r.ok || r.value != "a0|b0" {
		t.Fatalf("expected the page value from before the later updates, got %q, %v, %v", r.value, r.err, r.ok)
	}
	waitContent(t, page, "#a", "a1")
	waitContent(t, page, "#b", "b1")
}
