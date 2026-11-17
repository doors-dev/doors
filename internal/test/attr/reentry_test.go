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
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func reentryPage(t *testing.T) (*test.Bro, *rod.Page, *checkpointLog) {
	t.Helper()
	log := &checkpointLog{}
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &reentryFragment{log: log}
	})
	return bro, bro.Page(t, "/"), log
}

func rateReentry(t *testing.T, other string, chained string, queued string) {
	bro, page, log := reentryPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickBurst(t, page, "#"+other, "#"+chained, "#"+queued, "#"+other)
	<-time.After(2500 * time.Millisecond)
	before := checkpointCount(log.snapshot(), "start:"+chained)
	test.ClickNow(t, page, "#"+chained)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if checkpointCount(log.snapshot(), "start:"+chained) > before {
			return
		}
		<-time.After(25 * time.Millisecond)
	}
	t.Fatalf("%s is stuck after the rate scope replaced its candidate: %v", chained, log.snapshot())
}

func TestScopeRateReentrySerial(t *testing.T) {
	rateReentry(t, "rb", "ra", "ra")
}

func TestScopeRateReentryCheckpoint(t *testing.T) {
	rateReentry(t, "cb", "ca", "ca")
}

func TestScopeRateReentryFrame(t *testing.T) {
	rateReentry(t, "fb", "fo", "ff")
}

func TestScopeLatestReentry(t *testing.T) {
	bro, page, log := reentryPage(t)
	defer bro.Close()
	defer page.Close()

	test.ClickBurst(t, page, "#la", "#la", "#lb")
	<-time.After(600 * time.Millisecond)
	entries := log.snapshot()
	if started := checkpointCount(entries, "start:la") + checkpointCount(entries, "start:lb"); started != 1 {
		t.Fatalf("latest scope let %d requests start at once: %v", started, entries)
	}
	test.ClickNow(t, page, "#lb")
	<-time.After(300 * time.Millisecond)
	test.TestClassNot(t, page, "#la", "busy")
	test.TestClassNow(t, page, "#lb", "busy")
}

func TestScopeDoubleEmpty(t *testing.T) {
	log := &checkpointLog{}
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &doubleEmptyFragment{log: log}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.ClickNow(t, page, "#z")
	waitCheckpointLog(t, log, time.Second, "start:z")
	test.ClickNow(t, page, "#a")
	waitCheckpointLog(t, log, time.Second, "start:a")
	test.ClickNow(t, page, "#c")
	<-time.After(50 * time.Millisecond)
	test.ClickNow(t, page, "#j")
	waitCheckpointLog(t, log, 2*time.Second, "start:j")
	<-time.After(100 * time.Millisecond)
	test.ClickNow(t, page, "#c2")
	entries := waitCheckpointLog(t, log, 4*time.Second, "end:c2")
	if n := checkpointCount(entries, "start:c"); n != 0 {
		t.Fatalf("blocking should reject the promoted checkpoint c: %v", entries)
	}
	checkpointBefore(t, entries, "end:j", "start:c2")
}

func badIndicatorPage(t *testing.T) (*test.Bro, *rod.Page, *checkpointLog) {
	t.Helper()
	log := &checkpointLog{}
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &badIndicatorFragment{log: log}
	})
	return bro, bro.Page(t, "/"), log
}

func TestScopeBadIndicatorSerial(t *testing.T) {
	bro, page, log := badIndicatorPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#sx")
	waitCheckpointLog(t, log, time.Second, "start:sx")
	test.ClickNow(t, page, "#sy")
	waitCheckpointLog(t, log, 2*time.Second, "start:sy")
	test.ClickNow(t, page, "#sz")
	waitCheckpointLog(t, log, 2*time.Second, "start:sz")
}

func TestScopeBadIndicatorBlocking(t *testing.T) {
	bro, page, log := badIndicatorPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#by")
	waitCheckpointLog(t, log, time.Second, "start:by")
	<-time.After(200 * time.Millisecond)
	test.ClickNow(t, page, "#bz")
	waitCheckpointLog(t, log, time.Second, "start:bz")
}

func debounceReentryPage(t *testing.T) (*test.Bro, *rod.Page, *checkpointLog) {
	t.Helper()
	log := &checkpointLog{}
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &debounceReentryFragment{log: log}
	})
	page := bro.Page(t, "/")
	page.MustEval(`() => {
		window.__errs = []
		window.addEventListener("error", e => window.__errs.push(String(e.message)))
		window.addEventListener("unhandledrejection", e => window.__errs.push("unhandled: " + String(e.reason)))
	}`)
	return bro, page, log
}

func debounceErrs(page *rod.Page) []string {
	var out []string
	for _, v := range page.MustEval(`() => window.__errs`).Arr() {
		out = append(out, v.Str())
	}
	return out
}

func TestScopeDebounceReentry(t *testing.T) {
	bro, page, log := debounceReentryPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickBurst(t, page, "#x", "#y", "#z")
	<-time.After(800 * time.Millisecond)
	entries := log.snapshot()
	errs := debounceErrs(page)
	if len(errs) != 0 {
		t.Fatalf("uncaught errors after debounce re-entry: %v (log %v)", errs, entries)
	}
}

func TestScopeDebounceReentryLimit(t *testing.T) {
	bro, page, log := debounceReentryPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#p")
	waitCheckpointLog(t, log, 2*time.Second, "start:p")
	t1 := time.Now()
	test.ClickBurst(t, page, "#lx", "#ly", "#lz")
	<-time.After(time.Until(t1.Add(250 * time.Millisecond)))
	test.ClickBurst(t, page, "#w")
	<-time.After(time.Until(t1.Add(500 * time.Millisecond)))
	w2 := time.Now()
	test.ClickBurst(t, page, "#w")
	entries := waitCheckpointLog(t, log, 2*time.Second, "start:w")
	waited := log.at("start:w").Sub(w2)
	errs := debounceErrs(page)
	if waited < 250*time.Millisecond {
		t.Fatalf("w launched %v after its last click, before its 300ms duration and 600ms limit: %v", waited, entries)
	}
	if len(errs) != 0 {
		t.Fatalf("uncaught errors: %v", errs)
	}
}

func TestScopeBadIndicatorQueued(t *testing.T) {
	bro, page, log := badIndicatorPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#ea")
	waitCheckpointLog(t, log, time.Second, "start:ea")
	test.ClickNow(t, page, "#eb")
	waitCheckpointLog(t, log, time.Second, "start:eb")
	test.ClickNow(t, page, "#ec")
	waitCheckpointLog(t, log, 2*time.Second, "start:ec")
	waitCheckpointLog(t, log, 2*time.Second, "end:eb")
	<-time.After(100 * time.Millisecond)
	test.TestClassNot(t, page, "#e", "busy")
}

func TestScopeBadIndicatorCleanup(t *testing.T) {
	bro, page, log := badIndicatorPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#fd")
	waitCheckpointLog(t, log, time.Second, "start:fd")
	test.TestAttrNo(t, page, "#f", "data-x")
	test.TestAttrNow(t, page, "#f", "data-keep", "orig")
	test.TestClassNot(t, page, "#f", "busy")
	test.TestClassNot(t, page, "#f", "done")
	waitCheckpointLog(t, log, time.Second, "end:fd")
}

func TestIndicatorClassWhitespace(t *testing.T) {
	bro, page, log := badIndicatorPage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#gd")
	waitCheckpointLog(t, log, time.Second, "start:gd")
	test.TestClassNow(t, page, "#g", "busy")
	test.TestClassNow(t, page, "#g", "done")
	test.TestClassNow(t, page, "#g", "now")
	waitCheckpointLog(t, log, time.Second, "end:gd")
}

func doubleSettlePage(t *testing.T) (*test.Bro, *rod.Page, *checkpointLog) {
	t.Helper()
	log := &checkpointLog{}
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &doubleSettleFragment{log: log}
	})
	page := bro.Page(t, "/")
	test.ClickNow(t, page, "#p")
	waitCheckpointLog(t, log, time.Second, "start:p")
	test.ClickBurst(t, page, "#q", "#r")
	waitCheckpointLog(t, log, 2*time.Second, "end:p")
	waitCheckpointLog(t, log, 2*time.Second, "start:r")
	<-time.After(200 * time.Millisecond)
	if n := checkpointCount(log.snapshot(), "start:q"); n != 0 {
		t.Fatalf("blocking should reject q: %v", log.snapshot())
	}
	return bro, page, log
}

func TestScopeDoubleSettleSerial(t *testing.T) {
	bro, page, log := doubleSettlePage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#s")
	entries := waitCheckpointLog(t, log, 4*time.Second, "end:s")
	checkpointBefore(t, entries, "end:r", "start:s")
}

func TestScopeDoubleSettleBlocking(t *testing.T) {
	bro, page, log := doubleSettlePage(t)
	defer bro.Close()
	defer page.Close()
	test.ClickNow(t, page, "#t")
	waitCheckpointLog(t, log, 4*time.Second, "end:r")
	<-time.After(300 * time.Millisecond)
	if n := checkpointCount(log.snapshot(), "start:t"); n != 0 {
		t.Fatalf("blocking should reject t while r runs: %v", log.snapshot())
	}
}
