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
	entries := waitCheckpointLog(t, log, 300*time.Millisecond, "end:c2")
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
	test.ClickNow(t, page, "#of")
	entries := waitCheckpointLog(t, log, 2*time.Second, "end:c2")
	checkpointBefore(t, entries, "end:of", "start:c2")
	checkpointBefore(t, entries, "end:o1", "start:c2")
}

func TestScopeCheckpointNoStarvation(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#o1")
	waitCheckpointLog(t, log, time.Second, "start:o1")
	test.ClickNow(t, page, "#c2")
	for range 10 {
		<-time.After(150 * time.Millisecond)
		test.ClickNow(t, page, "#o2")
	}
	entries := waitCheckpointLog(t, log, 3*time.Second, "end:c2")
	checkpointBefore(t, entries, "end:o1", "start:c2")
	checkpointBefore(t, entries, "start:c2", "end:o2")
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
	checkpointBefore(t, entries, "start:c2", "end:o2")
	checkpointBefore(t, entries, "end:o2", "start:c3")
}

func TestScopeCheckpointBlocking(t *testing.T) {
	bro, page, log := newCheckpointPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#o1")
	waitCheckpointLog(t, log, time.Second, "start:o1")
	test.ClickBurst(t, page, "#cb", "#cb")
	waitCheckpointLog(t, log, 2*time.Second, "end:cb")
	<-time.After(200 * time.Millisecond)
	if n := checkpointCount(log.snapshot(), "start:cb"); n != 1 {
		t.Fatalf("blocking checkpoint should drop the double click, ran %d times", n)
	}
	test.ClickNow(t, page, "#cb")
	waitSnapshotLog(t, log, time.Second, func(entries []string) bool {
		return checkpointCount(entries, "end:cb") == 2
	})
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
