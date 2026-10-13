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
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
	"github.com/go-rod/rod"
)

type staticLog struct {
	mu   sync.Mutex
	errs []string
}

func (l *staticLog) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelError
}

func (l *staticLog) Handle(_ context.Context, r slog.Record) error {
	entry := r.Message
	r.Attrs(func(a slog.Attr) bool {
		entry += " " + a.String()
		return true
	})
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errs = append(l.errs, entry)
	return nil
}

func (l *staticLog) WithAttrs([]slog.Attr) slog.Handler {
	return l
}

func (l *staticLog) WithGroup(string) slog.Handler {
	return l
}

func (l *staticLog) staticErrors() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return slices.Clone(l.errs)
}

type staticRig struct {
	page *rod.Page
	f    *staticFragment
	log  *staticLog
}

func staticOpen(t *testing.T) *staticRig {
	t.Helper()
	rig := &staticRig{
		f:   newStaticFragment(),
		log: &staticLog{},
	}
	bro := test.NewPathBroOptions(browser, func(s test.PathLens) gox.Comp {
		return &test.Page{
			Source: s,
			F:      rig.f,
		}
	}, []doors.With{doors.WithLogger(slog.New(rig.log))})
	rig.page = bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { rig.page.Close() })
	return rig
}

func (r *staticRig) staticRendered(t *testing.T, want string) {
	t.Helper()
	select {
	case got := <-r.f.rendered:
		if got != want {
			t.Fatalf("expected %q rendered, got %q", want, got)
		}
	case <-time.After(waitLimit):
		t.Fatalf("%q was not rendered", want)
	}
}

func (r *staticRig) staticStep(t *testing.T) {
	t.Helper()
	select {
	case r.f.step <- struct{}{}:
	case <-time.After(waitLimit):
		t.Fatal("handler is not waiting")
	}
}

func (r *staticRig) staticApplied(t *testing.T) {
	t.Helper()
	select {
	case <-r.f.applied:
	case <-time.After(waitLimit):
		t.Fatal("batch was not applied")
	}
}

func (r *staticRig) staticResult(t *testing.T) {
	t.Helper()
	select {
	case err := <-r.f.results:
		if err != nil {
			t.Fatalf("expected the call applied, got %v", err)
		}
	case <-time.After(waitLimit):
		t.Fatal("call did not complete")
	}
}

func (r *staticRig) staticPending(t *testing.T) {
	t.Helper()
	select {
	case err := <-r.f.results:
		t.Fatalf("expected the call to wait behind the batch, completed with %v", err)
	default:
	}
}

func (r *staticRig) staticAbsent(t *testing.T, selector string) {
	t.Helper()
	has, _, err := r.page.Has(selector)
	if err != nil {
		t.Fatalf("lookup of %s failed: %v", selector, err)
	}
	if has {
		t.Fatalf("expected %s not on the page", selector)
	}
}

func (r *staticRig) staticItems(t *testing.T) string {
	t.Helper()
	res, err := r.page.Eval(`() => Array.from(document.querySelectorAll("#feed .item")).map((el) => el.textContent).join(",")`)
	if err != nil {
		t.Fatalf("feed read failed: %v", err)
	}
	return res.Value.Str()
}

func (r *staticRig) staticWaitItems(t *testing.T, want string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		got := r.staticItems(t)
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("feed: expected %q, got %q", want, got)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (r *staticRig) staticScrollY(t *testing.T) float64 {
	t.Helper()
	res, err := r.page.Eval(`() => window.scrollY`)
	if err != nil {
		t.Fatalf("scroll read failed: %v", err)
	}
	return res.Value.Num()
}

func (r *staticRig) staticClean(t *testing.T) {
	t.Helper()
	if errs := r.log.staticErrors(); len(errs) != 0 {
		t.Fatalf("expected no errors logged, got %s", fmt.Sprint(errs))
	}
}

// A plain Static of a door whose Inner is held by an open batch waits for the batch and replaces the door without a failed call.
func TestStaticAfterHeldInner(t *testing.T) {
	r := staticOpen(t)
	test.ClickNow(t, r.page, "#a-batched")
	r.staticRendered(t, "x1")
	test.ClickNow(t, r.page, "#a-plain")
	r.staticRendered(t, "static")
	time.Sleep(quiet)
	if got := content(t, r.page, "#x"); got != "x0" {
		t.Fatalf("expected the door untouched while the batch is open, got %q", got)
	}
	r.staticAbsent(t, "#x-static")
	r.staticPending(t)
	r.staticStep(t)
	waitContent(t, r.page, "#x-static", "static")
	r.staticAbsent(t, "#x")
	r.staticApplied(t)
	r.staticResult(t)
	r.staticClean(t)
}

// A plain update of a door placed by a held Static waits for the batch and lands in the placed door.
func TestStaticHoldsPlacedDoorUpdate(t *testing.T) {
	r := staticOpen(t)
	test.ClickNow(t, r.page, "#b-batched")
	r.staticRendered(t, "y0")
	test.ClickNow(t, r.page, "#b-plain")
	r.staticRendered(t, "y1")
	time.Sleep(quiet)
	if got := content(t, r.page, "#x"); got != "x0" {
		t.Fatalf("expected the Static held while the batch is open, got %q", got)
	}
	r.staticAbsent(t, "#y")
	r.staticPending(t)
	r.staticStep(t)
	waitContent(t, r.page, "#y", "y1")
	waitContent(t, r.page, "#b-item", "item")
	r.staticAbsent(t, "#x")
	r.staticApplied(t)
	r.staticResult(t)
	r.staticClean(t)
}

// The Static plus next door append recipe run several times in one Batch hook shows all items together, in order.
func TestStaticAppendRecipeInBatch(t *testing.T) {
	r := staticOpen(t)
	test.ClickNow(t, r.page, "#c-batched")
	r.staticRendered(t, "i1")
	r.staticStep(t)
	r.staticRendered(t, "i2")
	r.staticStep(t)
	r.staticRendered(t, "i3")
	time.Sleep(quiet)
	if got := r.staticItems(t); got != "" {
		t.Fatalf("expected no items while the batch is open, got %q", got)
	}
	r.staticStep(t)
	r.staticWaitItems(t, "i1,i2,i3")
	r.staticApplied(t)
	r.staticClean(t)
}

// A plain append between the appends of an open Batch hook waits for the batch and keeps its place in the feed.
func TestStaticAppendRecipePlainInside(t *testing.T) {
	r := staticOpen(t)
	test.ClickNow(t, r.page, "#c-batched")
	r.staticRendered(t, "i1")
	test.ClickNow(t, r.page, "#c-plain")
	r.staticRendered(t, "p")
	r.staticStep(t)
	r.staticRendered(t, "i2")
	r.staticStep(t)
	r.staticRendered(t, "i3")
	time.Sleep(quiet)
	if got := r.staticItems(t); got != "" {
		t.Fatalf("expected no items while the batch is open, got %q", got)
	}
	r.staticPending(t)
	r.staticStep(t)
	r.staticWaitItems(t, "i1,p,i2,i3")
	r.staticApplied(t)
	r.staticResult(t)
	r.staticClean(t)
}

// A call made while rendering the content of a held Static runs after that content is on the page.
func TestStaticContentCallInBatch(t *testing.T) {
	r := staticOpen(t)
	test.ClickNow(t, r.page, "#d-batched")
	r.staticRendered(t, "d")
	time.Sleep(quiet)
	r.staticAbsent(t, "#d-target")
	r.staticStep(t)
	waitContent(t, r.page, "#d-target", "target")
	r.staticApplied(t)
	if got := r.staticScrollY(t); got == 0 {
		t.Fatal("expected the scroll to find the static content")
	}
	r.staticClean(t)
}

// A detached call made while rendering the content of a held Static waits for the batch and succeeds on that content.
func TestStaticContentCallDetached(t *testing.T) {
	r := staticOpen(t)
	test.ClickNow(t, r.page, "#d-detached")
	r.staticRendered(t, "d")
	time.Sleep(quiet)
	r.staticAbsent(t, "#d-target")
	r.staticPending(t)
	r.staticStep(t)
	waitContent(t, r.page, "#d-target", "target")
	r.staticResult(t)
	r.staticApplied(t)
	if got := r.staticScrollY(t); got == 0 {
		t.Fatal("expected the scroll to find the static content")
	}
	r.staticClean(t)
}
