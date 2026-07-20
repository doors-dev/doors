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
	"slices"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

func dispatchBro(t *testing.T) (*test.Bro, *rod.Page) {
	t.Helper()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &dispatchFragment{
			r: test.NewReporter(5),
		}
	})
	page := bro.Page(t, "/")
	return bro, page
}

func dispatchEval(t *testing.T, page *rod.Page, js string, expected []string) {
	t.Helper()
	value, err := page.Eval(js)
	if err != nil {
		t.Fatal("dispatch eval failed: ", err)
	}
	items := value.Value.Arr()
	got := make([]string, 0, len(items))
	for _, item := range items {
		got = append(got, item.Str())
	}
	if !slices.Equal(got, expected) {
		t.Fatalf("dispatch eval expected %q, got %q", expected, got)
	}
}

// The handler patches a door (report 0); when the dispatch promise resolves,
// the patched DOM must ALREADY be visible in the same eval.
func TestDispatchClickCompletes(t *testing.T) {
	bro, page := dispatchBro(t)
	defer bro.Close()
	defer page.Close()

	test.TestReportId(t, page, 0, "0")
	dispatchEval(t, page, `async () => {
		const n = await window.__sysDispatch("#btn", new PointerEvent("click", { bubbles: true, cancelable: true }))
		return [String(n), document.querySelector("#report-0").textContent]
	}`, []string{"1", "1"})
	test.TestReportId(t, page, 0, "1")
}

// AKeyDown with Keys: Enter completes; "x" is declined by the filter with no
// request made - dispatch rejects with a canceled HookErr.
func TestDispatchKeysFilter(t *testing.T) {
	bro, page := dispatchBro(t)
	defer bro.Close()
	defer page.Close()

	dispatchEval(t, page, `async () => {
		const enter = await window.__sysDispatch("#keys", new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }))
		const enterReport = document.querySelector("#report-1").textContent
		let caught
		try {
			await window.__sysDispatch("#keys", new KeyboardEvent("keydown", { key: "x", bubbles: true, cancelable: true }))
		} catch (err) {
			caught = err
		}
		return [String(enter), enterReport, String(caught !== undefined), String(caught.canceled())]
	}`, []string{"1", "enter", "true", "true"})
}

// Two rapid dispatches into a debounce scope: the first rejects canceled, the
// second completes and its patch is applied at resolution.
func TestDispatchDebounce(t *testing.T) {
	bro, page := dispatchBro(t)
	defer bro.Close()
	defer page.Close()

	dispatchEval(t, page, `async () => {
		const ev = () => new PointerEvent("click", { bubbles: true, cancelable: true })
		const p1 = window.__sysDispatch("#deb", ev())
		const p2 = window.__sysDispatch("#deb", ev())
		const [r1, r2] = await Promise.allSettled([p1, p2])
		return [
			r1.status,
			String(r1.reason.canceled()),
			r2.status,
			String(r2.value),
			document.querySelector("#report-2").textContent,
		]
	}`, []string{"rejected", "true", "fulfilled", "1", "1"})
}

// An event type with zero doors listeners resolves 0 immediately.
func TestDispatchNoListeners(t *testing.T) {
	bro, page := dispatchBro(t)
	defer bro.Close()
	defer page.Close()

	dispatchEval(t, page, `async () => {
		const n = await window.__sysDispatch("#btn", new Event("dblclick", { bubbles: true, cancelable: true }))
		return [String(n)]
	}`, []string{"0"})
}

// A bubbling dispatch on a child below the hooked ancestor still tracks the
// ancestor hook.
func TestDispatchBubbles(t *testing.T) {
	bro, page := dispatchBro(t)
	defer bro.Close()
	defer page.Close()

	dispatchEval(t, page, `async () => {
		const n = await window.__sysDispatch("#child", new PointerEvent("click", { bubbles: true, cancelable: true }))
		return [String(n), document.querySelector("#report-3").textContent]
	}`, []string{"1", "parent"})
}

// Native events keep working unchanged on the same page: a real click reaches
// the handler, and a native non-matching key is silently filtered (no handler
// run, no error surfaced) while a matching key still fires.
//
// The canceled-kind HookErr thrown at the capture filter sites MUST be
// swallowed by the attach() listener's catch (capture.ts,
// `error.canceled() || error.notFound()`). If the swallow regresses, a
// non-matching native keydown lands in console.error("capture execution
// error") when the hook has no error actions (#keys), or runs the
// server-declared OnError actions when it has them (#keys-err). Both sinks
// are asserted empty here; both matching-key follow-ups prove the hooks were
// live, so the empty sinks mean "swallowed", not "never attached".
func TestDispatchNativeSanity(t *testing.T) {
	bro, page := dispatchBro(t)
	defer bro.Close()
	defer page.Close()

	// Trap console.error before any native interaction.
	_, err := page.Eval(`() => {
		window.__errs = []
		const orig = console.error
		console.error = function(...args) {
			window.__errs.push(args.map(String).join(" "))
			return orig.apply(this, args)
		}
	}`)
	if err != nil {
		t.Fatal("console trap install failed: ", err)
	}
	// Zero console errors and no onErr action ran; on failure the trapped
	// messages are spliced into the diff output.
	silent := func() {
		t.Helper()
		dispatchEval(t, page, `() => [
			String(window.__errs.length),
			...window.__errs,
			document.querySelector("#onerr").textContent,
		]`, []string{"0", ""})
	}

	test.Click(t, page, "#btn")
	waitReport(t, page, 0, "1")

	// Non-matching native key on the hook WITHOUT error actions: swallowed,
	// not console.error'd.
	test.Click(t, page, "#keys")
	page.Keyboard.MustType(input.KeyX)
	<-time.After(200 * time.Millisecond)
	test.TestReportId(t, page, 1, "")
	silent()

	// Non-matching native key on the hook WITH a server-declared OnError
	// action: swallowed before the onErr branch, sink stays empty.
	test.Click(t, page, "#keys-err")
	page.Keyboard.MustType(input.KeyX)
	<-time.After(200 * time.Millisecond)
	test.TestReportId(t, page, 4, "")
	silent()

	// Matching keys still fire on both hooks (they were live all along).
	page.Keyboard.MustType(input.Enter)
	waitReport(t, page, 4, "enter")
	test.Click(t, page, "#keys")
	page.Keyboard.MustType(input.Enter)
	waitReport(t, page, 1, "enter")
	silent()
}
