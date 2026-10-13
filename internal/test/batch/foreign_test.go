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

func foreignOpen(t *testing.T) (*rod.Page, *rod.Page, *foreignShared) {
	t.Helper()
	s := foreignNewShared()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &foreignFragment{s: s}
	})
	t.Cleanup(bro.Close)
	t.Cleanup(s.ga.foreignRelease)
	t.Cleanup(s.gb.foreignRelease)
	pageA := bro.Page(t, "/")
	t.Cleanup(func() { pageA.Close() })
	pageB := bro.Page(t, "/s")
	t.Cleanup(func() { pageB.Close() })
	if sessions, instances := bro.App().SessionCount(), bro.App().InstanceCount(); sessions != 1 || instances != 2 {
		t.Fatalf("expected two instances of one session, got %d sessions and %d instances", sessions, instances)
	}
	waitContent(t, pageA, "#foreign-shared", "s0")
	waitContent(t, pageB, "#foreign-shared", "s0")
	limitsWatch(t, pageA, "#foreign-shared", "#foreign-own")
	limitsWatch(t, pageB, "#foreign-x", "#foreign-y")
	return pageA, pageB, s
}

func foreignBusy(t *testing.T, page *rod.Page) bool {
	t.Helper()
	el, err := page.Timeout(waitLimit).Element("#foreign-ind")
	if err != nil {
		t.Fatal("indicator element not found")
	}
	attr, err := el.Attribute("data-busy")
	if err != nil {
		t.Fatal("indicator attribute read failed: ", err)
	}
	return attr != nil && *attr == "true"
}

func foreignWaitBusy(t *testing.T, page *rod.Page, want bool) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for foreignBusy(t, page) != want {
		if time.Now().After(deadline) {
			t.Fatalf("expected the indicator busy state %v", want)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func foreignFired(t *testing.T, s *foreignShared, want string) {
	t.Helper()
	select {
	case got := <-s.fired:
		if got != want {
			t.Fatalf("expected on for %q, got %q", want, got)
		}
	case <-time.After(waitLimit):
		t.Fatalf("on for %q did not run", want)
	}
}

func foreignNotFired(t *testing.T, s *foreignShared) {
	t.Helper()
	select {
	case got := <-s.fired:
		t.Fatalf("expected on not to run yet, got %q", got)
	case <-time.After(quiet):
	}
}

func foreignHeld(t *testing.T, page *rod.Page) {
	t.Helper()
	if got := limitsSeen(t, page); got != "s0|own0" {
		t.Fatalf("expected nothing new on the batching page while its handler runs, got %q", got)
	}
	if !foreignBusy(t, page) {
		t.Fatal("expected the indicator to stay while the handler runs")
	}
}

// A Batch hook that updates a session-shared Source holds its own page only: the other tab shows the new value while the hook is still open.
func TestForeignBatchDoesNotHoldPeer(t *testing.T) {
	pageA, pageB, s := foreignOpen(t)
	test.ClickBurst(t, pageA, "#foreign-batch")
	waitEntered(t, s.ga.gate)
	waitContent(t, pageB, "#foreign-shared", "s1")
	time.Sleep(quiet)
	foreignHeld(t, pageA)
	s.ga.foreignRelease()
	waitContent(t, pageA, "#foreign-shared", "s1")
	waitContent(t, pageA, "#foreign-own", "own-s1")
	foreignWaitBusy(t, pageA, false)
	if got := limitsSeen(t, pageA); got != "s0|own0 s1|own-s1" {
		t.Fatalf("expected the shared value and the own door to land together, got %q", got)
	}
}

// OnBatch in the other tab's subscription opens that tab's own batch: its ops land together and on runs while the first tab's batch is still open.
func TestForeignPeerOnBatchIndependent(t *testing.T) {
	pageA, pageB, s := foreignOpen(t)
	test.ClickBurst(t, pageA, "#foreign-batch")
	waitEntered(t, s.ga.gate)
	foreignFired(t, s, "s1")
	if x, y := content(t, pageB, "#foreign-x"), content(t, pageB, "#foreign-y"); x != "x-s1" || y != "y-s1" {
		t.Fatalf("expected both updates on the other tab when on runs, got %q and %q", x, y)
	}
	if got := limitsSeen(t, pageB); got != "x0|y0 x-s1|y-s1" {
		t.Fatalf("expected both doors of the other tab to change together, got %q", got)
	}
	time.Sleep(quiet)
	foreignHeld(t, pageA)
	s.ga.foreignRelease()
	waitContent(t, pageA, "#foreign-own", "own-s1")
	foreignWaitBusy(t, pageA, false)
	select {
	case got := <-s.fired:
		t.Fatalf("expected on to run once, got another for %q", got)
	case <-time.After(quiet):
	}
}

func foreignPeerPinned(t *testing.T, value string) {
	t.Helper()
	pageA, pageB, s := foreignOpen(t)
	test.ClickBurst(t, pageA, "#foreign-"+value)
	waitEntered(t, s.ga.gate)
	waitEntered(t, s.gb.gate)
	waitContent(t, pageB, "#foreign-shared", value)
	foreignWaitBusy(t, pageA, true)
	s.ga.foreignRelease()
	foreignWaitBusy(t, pageA, false)
	waitContent(t, pageA, "#foreign-shared", value)
	waitContent(t, pageA, "#foreign-own", "own-"+value)
	foreignNotFired(t, s)
	if got := limitsSeen(t, pageB); got != "x0|y0" {
		t.Fatalf("expected the other tab's batch still held, got %q", got)
	}
	s.gb.foreignRelease()
	foreignFired(t, s, value)
	if got := limitsSeen(t, pageB); got != "x0|y0 x-"+value+"|y-"+value {
		t.Fatalf("expected both doors of the other tab to change together, got %q", got)
	}
}

// A Batch hook finishes and releases its indicator while the other tab's batch, started by the shared update, is kept open with HoldSettle.
func TestForeignHookFinishesWhilePeerBatchOpen(t *testing.T) {
	foreignPeerPinned(t, "slow")
}

// A Batch hook finishes and releases its indicator after the other tab was closed mid-handler, also when it updates the shared Source again.
func TestForeignHookFinishesAfterPeerClosed(t *testing.T) {
	pageA, pageB, s := foreignOpen(t)
	test.ClickBurst(t, pageA, "#foreign-twice")
	waitEntered(t, s.ga.gate)
	waitContent(t, pageB, "#foreign-shared", "s1")
	foreignFired(t, s, "s1")
	if err := pageB.Close(); err != nil {
		t.Fatal("closing the other tab failed: ", err)
	}
	time.Sleep(quiet)
	foreignHeld(t, pageA)
	s.ga.foreignRelease()
	foreignWaitBusy(t, pageA, false)
	waitContent(t, pageA, "#foreign-shared", "s2")
	waitContent(t, pageA, "#foreign-own", "own-s1")
	if got := limitsSeen(t, pageA); got != "s0|own0 s2|own-s1" {
		t.Fatalf("expected the handler's updates to land together, got %q", got)
	}
}
