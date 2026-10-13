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
package batch

import (
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

const waitLimit = 5 * time.Second

const quiet = 300 * time.Millisecond

func open(t *testing.T) (*rod.Page, *gate) {
	t.Helper()
	g := newGate()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &batchFragment{
			r: test.NewReporter(1),
			g: g,
		}
	})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	return page, g
}

func content(t *testing.T, page *rod.Page, selector string) string {
	t.Helper()
	el, err := page.Timeout(waitLimit).Element(selector)
	if err != nil {
		t.Fatalf("element %s not found", selector)
	}
	s, err := el.Text()
	if err != nil {
		t.Fatalf("element %s has no text", selector)
	}
	return s
}

func waitContent(t *testing.T, page *rod.Page, selector string, want string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		got := content(t, page, selector)
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s: expected %q, got %q", selector, want, got)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitEntered(t *testing.T, g *gate) {
	t.Helper()
	select {
	case <-g.entered:
	case <-time.After(waitLimit):
		t.Fatal("handler did not start")
	}
}

// Without Batch the first update of a handler shows while the handler is still running.
func TestPlainHookStreams(t *testing.T) {
	page, g := open(t)
	test.ClickNow(t, page, "#plain")
	waitEntered(t, g)
	waitContent(t, page, "#a", "a2")
	if got := content(t, page, "#b"); got != "b0" {
		t.Fatalf("expected the second update not sent yet, got %q", got)
	}
	close(g.proceed)
	waitContent(t, page, "#b", "b2")
}

// With Batch the updates of a handler are held and land together when it finishes.
func TestBatchHookHoldsUpdates(t *testing.T) {
	page, g := open(t)
	test.ClickNow(t, page, "#batched")
	waitEntered(t, g)
	time.Sleep(quiet)
	if got := content(t, page, "#a"); got != "a0" {
		t.Fatalf("expected the first update held until the handler finishes, got %q", got)
	}
	close(g.proceed)
	waitContent(t, page, "#a", "a1")
	waitContent(t, page, "#b", "b1")
}

// While a batch is open an update of an untouched door applies at once,
// and an update of a door the batch holds waits and lands after it.
func TestBatchOvertakeAndWait(t *testing.T) {
	page, g := open(t)
	test.ClickNow(t, page, "#batched")
	waitEntered(t, g)
	test.ClickNow(t, page, "#other")
	waitContent(t, page, "#c", "c1")
	test.ClickNow(t, page, "#same")
	time.Sleep(quiet)
	if got := content(t, page, "#a"); got != "a0" {
		t.Fatalf("expected the update of a held door to wait, got %q", got)
	}
	close(g.proceed)
	waitContent(t, page, "#a", "a-late")
	waitContent(t, page, "#b", "b1")
}

// OnBatch applies its ops together and reports once the browser has applied them.
func TestOnBatchReportsApplied(t *testing.T) {
	page, _ := open(t)
	test.Click(t, page, "#onbatch")
	waitContent(t, page, "#report-0", "applied")
	waitContent(t, page, "#a", "a3")
	waitContent(t, page, "#b", "b3")
}
