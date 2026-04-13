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
	"net/url"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
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

func TestCapturePipelineOptions(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &captureFragment{
			r: test.NewReporter(5),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

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
	_, err := page.Eval(`() => {
		const el = document.querySelector("#filter-input")
		if (!(el instanceof HTMLInputElement)) {
			throw new Error("filter-input not found")
		}
		el.dispatchEvent(new KeyboardEvent("keydown", { key: "c", bubbles: true }))
	}`)
	if err != nil {
		t.Fatal(err)
	}
	test.TestReportId(t, page, 4, "")
	_, err = page.Eval(`() => {
		const el = document.querySelector("#filter-input")
		if (!(el instanceof HTMLInputElement)) {
			throw new Error("filter-input not found")
		}
		el.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true }))
	}`)
	if err != nil {
		t.Fatal(err)
	}
	waitReportValue(t, func() string { return page.MustInfo().URL }, func() string {
		return test.GetReportContent(t, page, 4)
	}, "Enter")
}
