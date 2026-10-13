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

func pathOpen(t *testing.T) (*rod.Page, *gate) {
	t.Helper()
	page, g, _ := pathOpenDone(t)
	return page, g
}

func pathOpenDone(t *testing.T) (*rod.Page, *gate, chan struct{}) {
	t.Helper()
	g := newGate()
	done := make(chan struct{}, 1)
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &pathFragment{g: g, done: done}
	})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	return page, g, done
}

func pathWaitDone(t *testing.T, done chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(waitLimit):
		t.Fatal("handler did not finish")
	}
}

func pathOf(t *testing.T, page *rod.Page) string {
	t.Helper()
	return page.MustEval("() => location.pathname").String()
}

func pathWait(t *testing.T, page *rod.Page, want string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		got := pathOf(t, page)
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("path: expected %q, got %q", want, got)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func pathMetaOf(t *testing.T, page *rod.Page) string {
	t.Helper()
	return page.MustEval(`() => {
		const all = document.querySelectorAll('meta[name="pathdesc"]')
		return Array.from(all).map((m) => m.getAttribute("content")).join(",")
	}`).String()
}

func pathMetaWait(t *testing.T, page *rod.Page, want string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		got := pathMetaOf(t, page)
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("meta: expected %q, got %q", want, got)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// A location change made by a Batch hook is held with its door update and both land when the handler finishes.
func TestPathBatchHoldsLocation(t *testing.T) {
	page, g := pathOpen(t)
	pathWait(t, page, "/")
	test.ClickNow(t, page, "#path-batched")
	waitEntered(t, g)
	time.Sleep(quiet)
	if got := pathOf(t, page); got != "/" {
		t.Fatalf("expected the location held until the handler finishes, got %q", got)
	}
	if got := content(t, page, "#a"); got != "a0" {
		t.Fatalf("expected the content held until the handler finishes, got %q", got)
	}
	close(g.proceed)
	pathWait(t, page, "/s/batched")
	waitContent(t, page, "#a", "a1")
	waitContent(t, page, "#b", "b1")
}

// Without Batch a location change made by a handler is still sent only when the handler finishes, while its door update streams.
func TestPathPlainHookDefersLocation(t *testing.T) {
	page, g := pathOpen(t)
	pathWait(t, page, "/")
	test.ClickNow(t, page, "#path-plain")
	waitEntered(t, g)
	waitContent(t, page, "#a", "a2")
	time.Sleep(quiet)
	if got := pathOf(t, page); got != "/" {
		t.Fatalf("expected the location not sent while the handler runs, got %q", got)
	}
	close(g.proceed)
	pathWait(t, page, "/s/plain")
	waitContent(t, page, "#b", "b2")
}

// A Batch hook that only changes the location holds it while open and ends with the new URL.
func TestPathBatchLocationOnly(t *testing.T) {
	page, g := pathOpen(t)
	pathWait(t, page, "/")
	test.ClickNow(t, page, "#path-only")
	waitEntered(t, g)
	time.Sleep(quiet)
	if got := pathOf(t, page); got != "/" {
		t.Fatalf("expected the location held until the handler finishes, got %q", got)
	}
	close(g.proceed)
	pathWait(t, page, "/s/only")
}

// A location change is a batch member: it waits with its batch behind another open batch holding the same door.
func TestPathBatchLocationWaitsWithMembers(t *testing.T) {
	page, g, done := pathOpenDone(t)
	pathWait(t, page, "/")
	test.ClickNow(t, page, "#path-holder")
	waitEntered(t, g)
	test.ClickNow(t, page, "#path-follow")
	pathWaitDone(t, done)
	time.Sleep(quiet)
	if got := pathOf(t, page); got != "/" {
		t.Fatalf("expected the location held with its batch, got %q", got)
	}
	if got := content(t, page, "#a"); got != "a0" {
		t.Fatalf("expected the content held, got %q", got)
	}
	close(g.proceed)
	pathWait(t, page, "/s/follow")
	waitContent(t, page, "#a", "a-follow")
}

// Without Batch a location change applies at once while the door update of the same handler waits behind an open batch.
func TestPathPlainLocationOvertakes(t *testing.T) {
	page, g, done := pathOpenDone(t)
	pathWait(t, page, "/")
	test.ClickNow(t, page, "#path-holder")
	waitEntered(t, g)
	test.ClickNow(t, page, "#path-follow-plain")
	pathWaitDone(t, done)
	pathWait(t, page, "/s/follow-plain")
	if got := content(t, page, "#a"); got != "a0" {
		t.Fatalf("expected the update of a held door to wait, got %q", got)
	}
	close(g.proceed)
	waitContent(t, page, "#a", "a-follow-plain")
}

// Meta updates are not batched: the meta tag of content re-rendered by a Batch hook changes while the content is still held.
func TestPathMetaNotBatched(t *testing.T) {
	page, g := pathOpen(t)
	pathMetaWait(t, page, "first")
	test.ClickNow(t, page, "#path-meta")
	waitEntered(t, g)
	pathMetaWait(t, page, "second")
	if got := content(t, page, "#a"); got != "a0" {
		t.Fatalf("expected the content held, got %q", got)
	}
	close(g.proceed)
	waitContent(t, page, "#a", "a3")
	waitContent(t, page, "#b", "b3")
	pathMetaWait(t, page, "second")
}
