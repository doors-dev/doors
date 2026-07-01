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
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func currentHash(t *testing.T, pageURL string) string {
	t.Helper()
	u, err := url.Parse(pageURL)
	if err != nil {
		t.Fatal(err)
	}
	return u.Fragment
}

func waitReportValue(t *testing.T, pageURL func() string, check func() string, expected string) {
	t.Helper()
	deadline := time.Now().Add(1500 * time.Millisecond)
	for {
		if got := check(); got == expected {
			return
		} else if time.Now().After(deadline) {
			t.Fatalf("expected %q before timeout, got %q (url %s)", expected, got, pageURL())
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func dispatchKey(t *testing.T, page *rod.Page, selector, init string) {
	t.Helper()
	_, err := page.Eval(fmt.Sprintf(`() => {
		const el = document.querySelector(%q)
		if (!el) {
			throw new Error("element not found: " + %q)
		}
		el.dispatchEvent(new KeyboardEvent("keydown", %s))
	}`, selector, selector, init))
	if err != nil {
		t.Fatal(err)
	}
}

func TestCapturePipelineOptions(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &captureFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	pageURL := func() string { return page.MustInfo().URL }
	report := func(id int) func() string {
		return func() string { return test.GetReportContent(t, page, id) }
	}

	test.Click(t, page, "#bubble-child")
	test.TestReportId(t, page, 0, "")
	test.TestReportId(t, page, 1, "child")

	test.Click(t, page, "#exact-child")
	test.TestReportId(t, page, 2, "")
	test.Click(t, page, "#exact-parent")
	test.TestReportId(t, page, 2, "exact")

	beforeURL := page.MustInfo().URL
	if hash := currentHash(t, beforeURL); hash != "" {
		t.Fatalf("expected empty hash before prevent-default click, got %q", hash)
	}
	test.Click(t, page, "#prevent-link")
	test.TestReportId(t, page, 3, "prevent")
	afterURL := page.MustInfo().URL
	if hash := currentHash(t, afterURL); hash != "" {
		t.Fatalf("expected preventDefault click to keep empty hash, got %q", hash)
	}

	test.Click(t, page, "#filter-input")
	dispatchKey(t, page, "#filter-input", `{ key: "c", bubbles: true }`)
	test.TestReportId(t, page, 4, "")
	dispatchKey(t, page, "#filter-input", `{ key: "Enter", bubbles: true }`)
	waitReportValue(t, pageURL, report(4), "1")
	dispatchKey(t, page, "#filter-input", `{ key: "Enter", ctrlKey: true, shiftKey: true, altKey: true, metaKey: true, bubbles: true }`)
	waitReportValue(t, pageURL, report(4), "2")

	dispatchKey(t, page, "#keys-ctrl", `{ key: "s", bubbles: true }`)
	test.TestReportId(t, page, 5, "")
	dispatchKey(t, page, "#keys-ctrl", `{ key: "s", ctrlKey: true, bubbles: true }`)
	waitReportValue(t, pageURL, report(5), "1")
	dispatchKey(t, page, "#keys-ctrl", `{ key: "s", ctrlKey: true, shiftKey: true, altKey: true, metaKey: true, bubbles: true }`)
	waitReportValue(t, pageURL, report(5), "2")

	dispatchKey(t, page, "#keys-ctrl-off", `{ key: "d", ctrlKey: true, bubbles: true }`)
	test.TestReportId(t, page, 6, "")
	dispatchKey(t, page, "#keys-ctrl-off", `{ key: "d", bubbles: true }`)
	waitReportValue(t, pageURL, report(6), "1")
	dispatchKey(t, page, "#keys-ctrl-off", `{ key: "d", shiftKey: true, altKey: true, metaKey: true, bubbles: true }`)
	waitReportValue(t, pageURL, report(6), "2")

	dispatchKey(t, page, "#keys-meta", `{ key: "e", bubbles: true }`)
	test.TestReportId(t, page, 7, "")
	dispatchKey(t, page, "#keys-meta", `{ key: "e", metaKey: true, bubbles: true }`)
	waitReportValue(t, pageURL, report(7), "1")

	dispatchKey(t, page, "#keys-multi", `{ key: "a", ctrlKey: true, bubbles: true }`)
	waitReportValue(t, pageURL, report(8), "1")
	dispatchKey(t, page, "#keys-multi", `{ key: "b", bubbles: true }`)
	test.TestReportId(t, page, 8, "1")
	dispatchKey(t, page, "#keys-multi", `{ key: "b", shiftKey: true, bubbles: true }`)
	waitReportValue(t, pageURL, report(8), "2")

	dispatchKey(t, page, "#keys-any", `{ key: "z", bubbles: true }`)
	test.TestReportId(t, page, 9, "")
	dispatchKey(t, page, "#keys-any", `{ key: "z", altKey: true, bubbles: true }`)
	waitReportValue(t, pageURL, report(9), "1")
	dispatchKey(t, page, "#keys-any", `{ key: "q", altKey: true, bubbles: true }`)
	waitReportValue(t, pageURL, report(9), "2")
}
