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

func placementOpen(t *testing.T) (*rod.Page, *gate, *placementLog) {
	t.Helper()
	g := newGate()
	l := newPlacementLog()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &placementFragment{g: g, l: l}
	})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	waitContent(t, page, "#pd-ready", "ready")
	return page, g, l
}

func placementCount(t *testing.T, page *rod.Page, selector string) int {
	t.Helper()
	els, err := page.Elements(selector)
	if err != nil {
		t.Fatalf("query %s failed: %v", selector, err)
	}
	return len(els)
}

func placementPending(t *testing.T, l *placementLog, label string, want string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		got, closed, _ := l.state(label)
		if closed {
			t.Fatalf("%s: expected the call still waiting after %q, it finished with %q", label, want, got)
		}
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s: expected %q while waiting, got %q", label, want, got)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func placementDone(t *testing.T, l *placementLog, label string, want string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		got, closed, _ := l.state(label)
		if closed {
			if got != want {
				t.Fatalf("%s: expected %q, got %q", label, want, got)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s: expected %q, still open with %q", label, want, got)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func placementHeld(t *testing.T, page *rod.Page, selector string, want string) {
	t.Helper()
	time.Sleep(quiet)
	if got := content(t, page, selector); got != want {
		t.Fatalf("%s: expected %q while the batch is open, got %q", selector, want, got)
	}
}

// An update of a door that a held update delivers waits for it and lands on top once the batch closes.
func TestPlacementChildUpdateWaitsForPlacing(t *testing.T) {
	page, g, l := placementOpen(t)
	test.ClickNow(t, page, "#pa-open")
	waitEntered(t, g)
	placementDone(t, l, "box", "nil")
	test.ClickNow(t, page, "#pa-poke")
	placementPending(t, l, "poke", "nil")
	placementHeld(t, page, "#box", "box0")
	placementPending(t, l, "poke", "nil")
	close(g.proceed)
	waitContent(t, page, "#kid", "kid1")
	placementDone(t, l, "poke", "nil,nil")
	if n := placementCount(t, page, "#pa-wrap > #kid"); n != 1 {
		t.Fatalf("expected the delivered door in its place once, got %d", n)
	}
}

// A detached update issued by the Batch handler for a door its held update delivers waits for the batch.
func TestPlacementDetachedChildUpdateWaitsForPlacing(t *testing.T) {
	page, g, l := placementOpen(t)
	test.ClickNow(t, page, "#pa-open-go")
	waitEntered(t, g)
	placementPending(t, l, "poke", "nil")
	placementHeld(t, page, "#box", "box0")
	placementPending(t, l, "poke", "nil")
	close(g.proceed)
	waitContent(t, page, "#kid", "kid1")
	placementDone(t, l, "poke", "nil,nil")
}

// An update of the innermost door of a chain delivered by one held update waits for the batch.
func TestPlacementNestedChainUpdate(t *testing.T) {
	page, g, l := placementOpen(t)
	test.ClickNow(t, page, "#pb-open")
	waitEntered(t, g)
	placementDone(t, l, "box", "nil")
	test.ClickNow(t, page, "#pb-poke")
	placementPending(t, l, "poke", "nil")
	placementHeld(t, page, "#box", "box0")
	placementPending(t, l, "poke", "nil")
	close(g.proceed)
	waitContent(t, page, "#x", "x1")
	placementDone(t, l, "poke", "nil,nil")
	if n := placementCount(t, page, "#box > #p > #q > #x"); n != 1 {
		t.Fatalf("expected the chain nested as delivered, got %d matches", n)
	}
}

// Outside updates that place a door and then update it both wait behind the held update delivering their parent.
func TestPlacementNestedChainOfWaitingCalls(t *testing.T) {
	page, g, l := placementOpen(t)
	test.ClickNow(t, page, "#pb-open")
	waitEntered(t, g)
	placementDone(t, l, "box", "nil")
	test.ClickNow(t, page, "#pb-swap")
	placementPending(t, l, "swap", "nil")
	test.ClickNow(t, page, "#pb-leaf")
	placementPending(t, l, "leaf", "nil")
	placementHeld(t, page, "#box", "box0")
	placementPending(t, l, "swap", "nil")
	placementPending(t, l, "leaf", "nil")
	close(g.proceed)
	waitContent(t, page, "#y", "y1")
	placementDone(t, l, "swap", "nil,nil")
	placementDone(t, l, "leaf", "nil,nil")
	if n := placementCount(t, page, "#box > #p > #q > #y"); n != 1 {
		t.Fatalf("expected the later door inside the delivered chain, got %d matches", n)
	}
	if n := placementCount(t, page, "#x"); n != 0 {
		t.Fatalf("expected the replaced door gone, got %d", n)
	}
}

// A door moved by a held update shows up in its new parent only when the batch closes, with a later update on top.
func TestPlacementMoveLandsWithBatch(t *testing.T) {
	page, g, l := placementOpen(t)
	waitContent(t, page, "#left > #mv", "mv0")
	test.ClickNow(t, page, "#pc-move")
	waitEntered(t, g)
	placementDone(t, l, "right", "nil")
	test.ClickNow(t, page, "#pc-poke")
	placementPending(t, l, "poke", "nil")
	placementHeld(t, page, "#right", "right0")
	placementPending(t, l, "poke", "nil")
	t.Logf("while the move is held: old copy count %d, new copy count %d", placementCount(t, page, "#left > #mv"), placementCount(t, page, "#right > #mv"))
	close(g.proceed)
	waitContent(t, page, "#right > #mv", "mv1")
	placementDone(t, l, "poke", "nil,nil")
	if n := placementCount(t, page, "#mv"); n != 1 {
		t.Fatalf("expected the moved door on the page once, got %d", n)
	}
	if got := content(t, page, "#left"); got != "" {
		t.Fatalf("expected the old parent empty, got %q", got)
	}
}

// When the same batch also holds an update of the old parent, the old copy of a moved door stays until the batch closes.
func TestPlacementMoveWithHeldSource(t *testing.T) {
	page, g, l := placementOpen(t)
	waitContent(t, page, "#left > #mv", "mv0")
	test.ClickNow(t, page, "#pc-swap")
	waitEntered(t, g)
	placementDone(t, l, "left", "nil")
	placementDone(t, l, "right", "nil")
	test.ClickNow(t, page, "#pc-poke")
	placementPending(t, l, "poke", "nil")
	placementHeld(t, page, "#right", "right0")
	if got := content(t, page, "#left > #mv"); got != "mv0" {
		t.Fatalf("expected the old copy untouched while the batch is open, got %q", got)
	}
	placementPending(t, l, "poke", "nil")
	close(g.proceed)
	waitContent(t, page, "#right > #mv", "mv1")
	waitContent(t, page, "#left", "left1")
	placementDone(t, l, "poke", "nil,nil")
	if n := placementCount(t, page, "#mv"); n != 1 {
		t.Fatalf("expected the moved door on the page once, got %d", n)
	}
}

// User calls issued while rendering content that a held update delivers run after that content is on the page.
func TestPlacementCallsFromHeldContent(t *testing.T) {
	page, g, l := placementOpen(t)
	test.ClickNow(t, page, "#pd-open")
	waitEntered(t, g)
	placementDone(t, l, "box", "nil")
	placementPending(t, l, "emit-detached", "nil")
	placementPending(t, l, "set", "nil")
	placementHeld(t, page, "#pd-log", "")
	placementHeld(t, page, "#box", "box0")
	placementPending(t, l, "emit-detached", "nil")
	placementPending(t, l, "set", "nil")
	close(g.proceed)
	waitContent(t, page, "#pd-log", "hit:in;hit:out;")
	waitContent(t, page, "#kid > #pd-target", "out")
	placementDone(t, l, "emit-detached", "nil,nil")
	placementDone(t, l, "set", "nil,nil")
	if pong, reached := l.values(); pong != "pong:out" || reached != 1 {
		t.Fatalf("expected the handler result and one element reached, got %q and %d", pong, reached)
	}
	mark, err := page.MustElement("#pd-set").Attribute("data-mark")
	if err != nil || mark == nil || *mark != "on" {
		t.Fatalf("expected the setter applied to the delivered element, got %v", mark)
	}
	if got, closed, _ := l.state("emit"); !closed || (got != "" && got != "nil") {
		t.Fatalf("expected the render call finished without an error, got %q closed %v", got, closed)
	}
}
