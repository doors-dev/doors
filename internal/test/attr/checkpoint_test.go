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
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func newCheckpointPage(t *testing.T) (*test.Bro, *rod.Page, *checkpointLog) {
	t.Helper()
	log := &checkpointLog{}
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &checkpointFragment{log: log}
	})
	page := bro.Page(t, "/")
	return bro, page, log
}

func waitCheckpointLog(t *testing.T, log *checkpointLog, timeout time.Duration, entry string) []string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		entries := log.snapshot()
		if slices.Contains(entries, entry) {
			return entries
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected %q before timeout, got %v", entry, entries)
		}
		<-time.After(25 * time.Millisecond)
	}
}

func checkpointBefore(t *testing.T, entries []string, first string, second string) {
	t.Helper()
	i := slices.Index(entries, first)
	j := slices.Index(entries, second)
	if i < 0 || j < 0 || i > j {
		t.Fatalf("expected %q before %q, got %v", first, second, entries)
	}
}

func TestScopeCheckpointImmediate(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#c2")
	entries := waitCheckpointLog(t, log, time.Second, "end:c2")
	if !slices.Equal(entries, []string{"start:c2", "end:c2"}) {
		t.Fatalf("checkpoint with nothing running should run at once, got %v", entries)
	}
}

func TestScopeCheckpointWaitsForRunning(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#o1")
	waitCheckpointLog(t, log, time.Second, "start:o1")
	test.ClickNow(t, page, "#c2")
	<-time.After(100 * time.Millisecond)
	test.ClickNow(t, page, "#o2")
	entries := waitCheckpointLog(t, log, 3*time.Second, "end:c2")
	checkpointBefore(t, entries, "end:o1", "start:c2")
	checkpointBefore(t, entries, "end:o2", "start:c2")
}

func TestScopeCheckpointWaitsForStream(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#o1")
	waitCheckpointLog(t, log, time.Second, "start:o1")
	test.ClickNow(t, page, "#c2")
	for range 8 {
		<-time.After(150 * time.Millisecond)
		test.ClickNow(t, page, "#os")
	}
	entries := waitCheckpointLog(t, log, 3*time.Second, "end:c2")
	if n := checkpointCount(entries[:slices.Index(entries, "start:c2")], "end:os"); n != 8 {
		t.Fatalf("checkpoint should wait for the whole stream, ran after %d of 8: %v", n, entries)
	}
}

func TestScopeCheckpointOrdinaryDuringRunning(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#c1")
	waitCheckpointLog(t, log, time.Second, "start:c1")
	test.ClickNow(t, page, "#of")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:c1")
	checkpointBefore(t, entries, "end:of", "end:c1")
}

func TestScopeCheckpointQueueRunning(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#c1")
	waitCheckpointLog(t, log, time.Second, "start:c1")
	test.ClickNow(t, page, "#c2")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	checkpointBefore(t, entries, "end:c1", "start:c2")
}

func TestScopeCheckpointQueuePending(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#o1")
	waitCheckpointLog(t, log, time.Second, "start:o1")
	test.ClickNow(t, page, "#c1")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#c2")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	checkpointBefore(t, entries, "end:o1", "start:c1")
	checkpointBefore(t, entries, "end:c1", "start:c2")
}

func TestScopeCheckpointOrdinaryBetween(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#o1")
	waitCheckpointLog(t, log, time.Second, "start:o1")
	test.ClickNow(t, page, "#c2")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#o2")
	waitCheckpointLog(t, log, time.Second, "start:o2")
	test.ClickNow(t, page, "#c3")
	entries := waitCheckpointLog(t, log, 3*time.Second, "end:c3")
	checkpointBefore(t, entries, "end:o1", "start:c2")
	checkpointBefore(t, entries, "end:o2", "start:c2")
	checkpointBefore(t, entries, "end:c2", "start:c3")
}

func TestScopeCheckpointBlocking(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#o1")
	waitCheckpointLog(t, log, time.Second, "start:o1")
	test.ClickBurst(t, page, "#cb", "#cb")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:cb")
	checkpointBefore(t, entries, "end:o1", "start:cb")
	<-time.After(200 * time.Millisecond)
	if n := checkpointCount(log.snapshot(), "start:cb"); n != 1 {
		t.Fatalf("blocking checkpoint should drop the double click, ran %d times", n)
	}
	test.ClickNow(t, page, "#cb")
	waitSnapshotLog(t, log, time.Second, func(entries []string) bool {
		return checkpointCount(entries, "end:cb") == 2
	})
}

func TestScopeCheckpointRejectedMember(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#od")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#c2")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#od")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	checkpointBefore(t, entries, "end:od", "start:c2")
	if n := checkpointCount(entries, "start:od"); n != 1 {
		t.Fatalf("debounce should reject the first member, ran %d times", n)
	}
}

func TestScopeCheckpointFailedMember(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#he")
	waitCheckpointLog(t, log, time.Second, "start:he")
	test.ClickNow(t, page, "#c2")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	checkpointBefore(t, entries, "fail:he", "start:c2")
}

func TestScopeCheckpointTimedOutMember(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	clicked := time.Now()
	test.ClickNow(t, page, "#ht")
	waitCheckpointLog(t, log, time.Second, "start:ht")
	test.ClickNow(t, page, "#c2")
	waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:ht")
	checkpointBefore(t, entries, "start:ht", "start:c2")
	checkpointBefore(t, entries, "end:c2", "end:ht")
	if d := log.at("start:c2").Sub(clicked); d < 250*time.Millisecond {
		t.Fatalf("timed-out request should hold the checkpoint until its timeout, checkpoint started %v after its click", d)
	}
}

func TestScopeCheckpointCanceledCheckpoint(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#o1")
	waitCheckpointLog(t, log, time.Second, "start:o1")
	test.ClickNow(t, page, "#cl")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#c2")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#cl")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:cl")
	checkpointBefore(t, entries, "end:o1", "start:c2")
	checkpointBefore(t, entries, "end:c2", "start:cl")
	if n := checkpointCount(entries, "start:cl"); n != 1 {
		t.Fatalf("latest should cancel the waiting checkpoint, ran %d times", n)
	}
}

func TestScopeCheckpointHeldMember(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#od")
	<-time.After(100 * time.Millisecond)
	test.ClickNow(t, page, "#c2")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	checkpointBefore(t, entries, "end:od", "start:c2")
}

func TestScopeCheckpointOrdinaryHoldsNext(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#c1")
	waitCheckpointLog(t, log, time.Second, "start:c1")
	test.ClickNow(t, page, "#c2")
	test.ClickNow(t, page, "#o1")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	checkpointBefore(t, entries, "start:o1", "end:c1")
	checkpointBefore(t, entries, "end:c1", "end:o1")
	checkpointBefore(t, entries, "end:o1", "start:c2")
}

func TestScopeCheckpointDebouncedField(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	typeField := func(value string) {
		page.MustEval(`(value) => {
			const el = document.getElementById("field")
			el.value = value
			el.dispatchEvent(new Event("input", { bubbles: true }))
		}`, value)
	}
	typeField("a")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#sub")
	<-time.After(100 * time.Millisecond)
	typeField("ab")
	waitSnapshotLog(t, log, 2*time.Second, func(entries []string) bool {
		return slices.ContainsFunc(entries, func(e string) bool { return strings.HasPrefix(e, "submit:") })
	})
	entries := log.snapshot()
	checkpointBefore(t, entries, "field:ab", "submit:ab")
}

func TestScopeCheckpointCanceledRunning(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#lc")
	waitCheckpointLog(t, log, time.Second, "start:lc")
	test.ClickNow(t, page, "#c2")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#lc")
	waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	waitSnapshotLog(t, log, 3*time.Second, func(entries []string) bool {
		return checkpointCount(entries, "end:lc") == 2
	})
	checkpointBefore(t, log.snapshot(), "start:c2", "end:lc")
}

func TestScopeCheckpointCancelDuringAfter(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#ls")
	waitCheckpointLog(t, log, time.Second, "start:ls")
	<-time.After(200 * time.Millisecond)
	test.ClickNow(t, page, "#ls")
	waitSnapshotLog(t, log, 2*time.Second, func(entries []string) bool {
		return checkpointCount(entries, "start:ls") == 2
	})
	test.ClickNow(t, page, "#c2")
	waitCheckpointLog(t, log, 3*time.Second, "end:c2")
}

func TestScopeCheckpointBlockingAfter(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#ck")
	waitCheckpointLog(t, log, time.Second, "start:ck")
	test.ClickNow(t, page, "#ck")
	waitSnapshotLog(t, log, 2*time.Second, func(entries []string) bool {
		return checkpointCount(entries, "end:ck") == 2
	})
	if entries := log.snapshot(); !slices.Equal(entries, []string{"start:ck", "end:ck", "start:ck", "end:ck"}) {
		t.Fatalf("second checkpoint should be queued behind the first, got %v", entries)
	}
}

func TestScopeCheckpointRejectedDownstream(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#o1")
	waitCheckpointLog(t, log, time.Second, "start:o1")
	test.ClickNow(t, page, "#bo")
	waitCheckpointLog(t, log, time.Second, "start:bo")
	test.ClickNow(t, page, "#ck")
	test.ClickNow(t, page, "#c2")
	waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:bo")
	checkpointBefore(t, entries, "end:o1", "start:c2")
	checkpointBefore(t, entries, "end:c2", "end:bo")
	if n := checkpointCount(entries, "start:ck"); n != 0 {
		t.Fatalf("blocking should reject the promoted checkpoint, ran %d times: %v", n, entries)
	}
}

func TestScopeCheckpointLatestRate(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickBurst(t, page, "#lr", "#lr", "#lr")
	test.ClickNow(t, page, "#c2")
	entries := waitCheckpointLog(t, log, 3*time.Second, "end:c2")
	checkpointBefore(t, entries, "end:lr", "start:c2")
	if n := checkpointCount(entries, "start:lr"); n != 1 {
		t.Fatalf("latest should leave one request, ran %d times: %v", n, entries)
	}
}

func checkpointCount(entries []string, entry string) int {
	n := 0
	for _, e := range entries {
		if e == entry {
			n++
		}
	}
	return n
}

func waitSnapshotLog(t *testing.T, log *checkpointLog, timeout time.Duration, cond func([]string) bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		entries := log.snapshot()
		if cond(entries) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("checkpoint log condition not met before timeout, got %v", entries)
		}
		<-time.After(25 * time.Millisecond)
	}
}
