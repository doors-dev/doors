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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
	"github.com/go-rod/rod"
)

const limitsPending = 8

const limitsRoll = 400 * time.Millisecond

func limitsOpen(t *testing.T, conf doors.Conf) (*rod.Page, *limitsState) {
	t.Helper()
	s := limitsNewState()
	bro := test.NewPathBroOptions(browser, func(p test.PathLens) gox.Comp {
		return &test.Page{
			Source: p,
			F:      limitsNewFragment(s),
		}
	}, []doors.With{doors.WithConf(conf)})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	return page, s
}

func limitsEntered(t *testing.T, ch chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(waitLimit):
		t.Fatalf("%s did not happen", what)
	}
}

func limitsProceed(t *testing.T, ch chan struct{}, what string) {
	t.Helper()
	select {
	case ch <- struct{}{}:
	case <-time.After(waitLimit):
		t.Fatalf("%s is not waiting", what)
	}
}

func limitsHeld(t *testing.T, page *rod.Page, selector string, want string) {
	t.Helper()
	if got := content(t, page, selector); got != want {
		t.Fatalf("%s: expected %q while the batch is open, got %q", selector, want, got)
	}
}

func limitsWatch(t *testing.T, page *rod.Page, selectors ...string) {
	t.Helper()
	_, err := page.Eval(`(selectors) => {
		const read = () => selectors.map((s) => document.querySelector(s)?.textContent ?? "").join("|")
		const log = [read()]
		window.limitsLog = log
		new MutationObserver(() => {
			const now = read()
			if (log[log.length - 1] !== now) {
				log.push(now)
			}
		}).observe(document.body, { childList: true, characterData: true, subtree: true })
	}`, selectors)
	if err != nil {
		t.Fatal("watch failed: ", err)
	}
}

func limitsSeen(t *testing.T, page *rod.Page) string {
	t.Helper()
	res, err := page.Eval(`() => window.limitsLog.join(" ")`)
	if err != nil {
		t.Fatal("watch read failed: ", err)
	}
	return res.Value.Str()
}

func limitsCountRolls(t *testing.T, page *rod.Page) {
	t.Helper()
	_, err := page.Eval(`() => {
		const origin = window.fetch
		window.limitsRolls = 0
		window.fetch = (input, init) => {
			if (init?.method === "GET" && /\/s\/[^?]+\?t=/.test(String(input))) {
				window.limitsRolls++
			}
			return origin(input, init)
		}
	}`)
	if err != nil {
		t.Fatal("roll count failed: ", err)
	}
}

func limitsRolls(t *testing.T, page *rod.Page) int {
	t.Helper()
	res, err := page.Eval(`() => window.limitsRolls`)
	if err != nil {
		t.Fatal("roll count read failed: ", err)
	}
	return res.Value.Int()
}

func limitsVolume(t *testing.T, conf doors.Conf, forced bool) {
	t.Helper()
	page, s := limitsOpen(t, conf)
	test.ClickNow(t, page, "#volume")
	limitsEntered(t, s.entered, "the first pause")
	time.Sleep(quiet)
	for i := range limitsManyCount {
		limitsHeld(t, page, fmt.Sprintf("#m%d", i), "m0")
	}
	limitsProceed(t, s.proceed, "the handler")
	limitsEntered(t, s.entered, "the second pause")
	if forced {
		for i := range limitsManyCount {
			waitContent(t, page, fmt.Sprintf("#m%d", i), "v1")
		}
	} else {
		time.Sleep(quiet)
		for i := range limitsManyCount {
			limitsHeld(t, page, fmt.Sprintf("#m%d", i), "m0")
		}
	}
	limitsProceed(t, s.proceed, "the handler")
	for i := range limitsManyCount {
		waitContent(t, page, fmt.Sprintf("#m%d", i), "v2")
	}
}

// A batch that outgrows SolitairePending is force-applied while its handler still runs, and the page ends in the final state.
func TestLimitsVolumeCapForcesApply(t *testing.T) {
	limitsVolume(t, doors.Conf{SolitairePending: limitsPending}, true)
}

// The same batch under the default SolitairePending stays held until its handler finishes.
func TestLimitsVolumeBelowCapHolds(t *testing.T) {
	limitsVolume(t, doors.Conf{}, false)
}

// A batch left open across several stream rolls stays held and applies once when its handler finishes.
func TestLimitsBatchSurvivesRoll(t *testing.T) {
	page, s := limitsOpen(t, doors.Conf{SolitaireRollTime: limitsRoll})
	limitsWatch(t, page, "#a", "#b")
	limitsCountRolls(t, page)
	test.ClickNow(t, page, "#roll")
	limitsEntered(t, s.entered, "the pause")
	start := limitsRolls(t, page)
	deadline := time.Now().Add(waitLimit)
	for limitsRolls(t, page) < start+3 {
		if time.Now().After(deadline) {
			t.Fatal("the stream did not roll")
		}
		time.Sleep(50 * time.Millisecond)
	}
	limitsHeld(t, page, "#a", "a0")
	limitsProceed(t, s.proceed, "the handler")
	waitContent(t, page, "#a", "r1")
	waitContent(t, page, "#b", "r1")
	if got := limitsSeen(t, page); got != "a0|b0 r1|r1" {
		t.Fatalf("expected one joint apply, saw %q", got)
	}
}

// Fifty updates of one door in a Batch hook show only the last one, when the handler finishes.
func TestLimitsSingleDoorShowsFinal(t *testing.T) {
	page, s := limitsOpen(t, doors.Conf{})
	limitsWatch(t, page, "#a")
	test.ClickNow(t, page, "#single")
	limitsEntered(t, s.entered, "the pause")
	time.Sleep(quiet)
	limitsHeld(t, page, "#a", "a0")
	limitsProceed(t, s.proceed, "the handler")
	final := fmt.Sprintf("s%d", limitsSingleCount)
	waitContent(t, page, "#a", final)
	if got := limitsSeen(t, page); got != "a0 "+final {
		t.Fatalf("expected only the final content, saw %q", got)
	}
}

func limitsLeave(t *testing.T, leave func(page *rod.Page)) {
	t.Helper()
	page, s := limitsOpen(t, doors.Conf{
		RequestTimeout:    time.Second,
		InstanceTTL:       2 * time.Second,
		SessionTTL:        time.Minute,
		SolitaireRollTime: limitsRoll,
	})
	test.ClickNow(t, page, "#leave")
	limitsEntered(t, s.entered, "the pause")
	leave(page)
	select {
	case <-s.released:
	case <-time.After(2 * waitLimit):
		t.Fatal("the handler was not released")
	}
	waitContent(t, page, "#a", "a0")
	test.ClickNow(t, page, "#leave")
	limitsEntered(t, s.entered, "the pause")
	time.Sleep(quiet)
	limitsHeld(t, page, "#a", "a0")
	limitsProceed(t, s.proceed, "the handler")
	waitContent(t, page, "#a", "l1")
	waitContent(t, page, "#b", "l1")
}

// Reloading the page while a batch is open cancels the handler once the instance expires, and the fresh page batches as usual.
func TestLimitsReloadReleasesHandler(t *testing.T) {
	limitsLeave(t, func(page *rod.Page) {
		page.MustReload().MustWaitLoad()
	})
}

// Navigating away while a batch is open cancels the handler once the instance expires, and a fresh page batches as usual.
func TestLimitsNavigateAwayReleasesHandler(t *testing.T) {
	limitsLeave(t, func(page *rod.Page) {
		url := page.MustInfo().URL
		page.MustNavigate("about:blank").MustWaitLoad()
		page.MustNavigate(url).MustWaitLoad()
	})
}

// Deferred content is not a batch member: the batch lands with the placeholder and the slow content replaces it later.
func TestLimitsDeferredAfterBatch(t *testing.T) {
	page, s := limitsOpen(t, doors.Conf{})
	limitsWatch(t, page, "#a", "#d", "#b")
	test.ClickNow(t, page, "#deferred")
	limitsEntered(t, s.entered, "the pause")
	limitsEntered(t, s.heavyEntered, "the deferred render")
	time.Sleep(quiet)
	limitsHeld(t, page, "#d", "d0")
	limitsProceed(t, s.proceed, "the handler")
	waitContent(t, page, "#d", "skeleton")
	waitContent(t, page, "#b", "e1")
	time.Sleep(quiet)
	if got := content(t, page, "#d"); got != "skeleton" {
		t.Fatalf("expected the placeholder until the deferred render finishes, got %q", got)
	}
	limitsProceed(t, s.heavyProceed, "the deferred render")
	waitContent(t, page, "#d", "heavy")
	if got := limitsSeen(t, page); got != "a0|d0|b0 e1|skeleton|e1 e1|heavy|e1" {
		t.Fatalf("expected the batch then the deferred content, saw %q", got)
	}
}

// Deferred content rendered while the batch is still open waits for the batch that holds its door.
func TestLimitsDeferredBehindBatch(t *testing.T) {
	page, s := limitsOpen(t, doors.Conf{})
	limitsWatch(t, page, "#a", "#d", "#b")
	test.ClickNow(t, page, "#deferred")
	limitsEntered(t, s.entered, "the pause")
	limitsEntered(t, s.heavyEntered, "the deferred render")
	time.Sleep(quiet)
	limitsProceed(t, s.heavyProceed, "the deferred render")
	time.Sleep(quiet)
	limitsHeld(t, page, "#d", "d0")
	limitsHeld(t, page, "#a", "a0")
	limitsProceed(t, s.proceed, "the handler")
	waitContent(t, page, "#d", "heavy")
	waitContent(t, page, "#b", "e1")
	seen := strings.Fields(limitsSeen(t, page))
	for i, state := range seen {
		first := i == 0 && state == "a0|d0|b0"
		last := i == len(seen)-1 && state == "e1|heavy|e1"
		if !first && !last && state != "e1|skeleton|e1" {
			t.Fatalf("expected the deferred content no earlier than the batch, saw %q", seen)
		}
	}
}

// Twenty queued clicks of a Batch hook each land as one joint update and leave the last value.
func TestLimitsRapidClicks(t *testing.T) {
	page, s := limitsOpen(t, doors.Conf{})
	limitsWatch(t, page, "#a", "#b")
	selectors := make([]string, 20)
	want := []string{"a0|b0"}
	for i := range selectors {
		selectors[i] = "#rapid"
		want = append(want, fmt.Sprintf("%d|%d", i+1, i+1))
	}
	test.ClickBurst(t, page, selectors...)
	for range selectors {
		limitsEntered(t, s.entered, "the pause")
		time.Sleep(30 * time.Millisecond)
		limitsProceed(t, s.proceed, "the handler")
	}
	waitContent(t, page, "#a", "20")
	waitContent(t, page, "#b", "20")
	if got := limitsSeen(t, page); got != strings.Join(want, " ") {
		t.Fatalf("expected joint updates in order, saw %q", got)
	}
}
