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

package tabstate

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func tabStateBro(mode string) *test.Bro {
	return test.NewFragmentBro(browser, func() test.Fragment {
		return &tabStateFragment{mode: mode}
	})
}

func waitContent(t *testing.T, page *rod.Page, selector string, content string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if test.GetContent(t, page, selector) == content {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	test.TestContent(t, page, selector, content)
}

func waitPath(t *testing.T, page *rod.Page, path string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		u, err := url.Parse(page.MustInfo().URL)
		if err == nil && u.Path == path {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected path %q, got %q", path, page.MustInfo().URL)
}

func reload(t *testing.T, page *rod.Page) {
	t.Helper()
	before := test.GetContent(t, page, "#instance-id")
	if err := page.Reload(); err != nil {
		t.Fatal(err)
	}
	page.MustWaitLoad()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if test.GetContent(t, page, "#instance-id") != before {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected a new instance after reload")
}

func initialHTML(t *testing.T, path string) string {
	t.Helper()
	res, err := http.Get(test.Host + path)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func initialValue(t *testing.T, path string) string {
	t.Helper()
	html := initialHTML(t, path)
	_, after, ok := strings.Cut(html, `id="value"`)
	if !ok {
		t.Fatal("value element not found in initial html")
	}
	_, after, _ = strings.Cut(after, ">")
	value, _, _ := strings.Cut(after, "<")
	return value
}

func TestTabStateFreshIsNilThenZero(t *testing.T) {
	bro := tabStateBro(modePlain)
	defer bro.Close()
	if got := initialValue(t, "/"); got != "nil" {
		t.Fatalf("expected nil before sync, got %q", got)
	}
	page := bro.Page(t, "/")
	defer page.Close()
	waitContent(t, page, "#value", "0")
}

func TestTabStateSurvivesReload(t *testing.T) {
	bro := tabStateBro(modePlain)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	waitContent(t, page, "#value", "0")
	test.Click(t, page, "#inc")
	waitContent(t, page, "#value", "1")
	test.Click(t, page, "#inc")
	waitContent(t, page, "#value", "2")
	reload(t, page)
	waitContent(t, page, "#value", "2")
	test.Click(t, page, "#inc")
	waitContent(t, page, "#value", "3")
	reload(t, page)
	waitContent(t, page, "#value", "3")
}

func TestTabStateClearIsZeroAfterInit(t *testing.T) {
	bro := tabStateBro(modePlain)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	waitContent(t, page, "#value", "0")
	test.Click(t, page, "#inc")
	waitContent(t, page, "#value", "1")
	test.Click(t, page, "#clear")
	waitContent(t, page, "#value", "0")
	reload(t, page)
	waitContent(t, page, "#value", "0")
}

func TestTabStateLatestAfterBack(t *testing.T) {
	bro := tabStateBro(modePlain)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	waitContent(t, page, "#value", "0")
	test.Click(t, page, "#inc")
	waitContent(t, page, "#value", "1")
	instance := test.GetContent(t, page, "#instance-id")

	test.Click(t, page, "#go-s")
	waitPath(t, page, "/s")
	test.Click(t, page, "#inc")
	waitContent(t, page, "#value", "2")

	if err := page.NavigateBack(); err != nil {
		t.Fatal(err)
	}
	waitPath(t, page, "/")
	test.TestContent(t, page, "#instance-id", instance)
	test.TestContent(t, page, "#value", "2")

	reload(t, page)
	waitContent(t, page, "#value", "2")
	waitPath(t, page, "/")
}

func TestTabStateLatestAfterForward(t *testing.T) {
	bro := tabStateBro(modePlain)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	waitContent(t, page, "#value", "0")
	test.Click(t, page, "#go-s")
	waitPath(t, page, "/s")
	if err := page.NavigateBack(); err != nil {
		t.Fatal(err)
	}
	waitPath(t, page, "/")
	test.Click(t, page, "#inc")
	waitContent(t, page, "#value", "1")
	if err := page.NavigateForward(); err != nil {
		t.Fatal(err)
	}
	waitPath(t, page, "/s")
	test.TestContent(t, page, "#value", "1")
	reload(t, page)
	waitContent(t, page, "#value", "1")
}

func TestTabStatePreInitUpdate(t *testing.T) {
	bro := tabStateBro(modePre)
	defer bro.Close()
	if got := initialValue(t, "/"); got != "5" {
		t.Fatalf("expected 5 before sync, got %q", got)
	}
	page := bro.Page(t, "/")
	defer page.Close()
	waitContent(t, page, "#value", "5")
	time.Sleep(100 * time.Millisecond)
	test.TestContent(t, page, "#value", "5")
	reload(t, page)
	waitContent(t, page, "#value", "5")
	reload(t, page)
	waitContent(t, page, "#value", "5")
}

func TestTabStatePreInitUpdateThenNil(t *testing.T) {
	bro := tabStateBro(modePreNil)
	defer bro.Close()
	if got := initialValue(t, "/"); got != "nil" {
		t.Fatalf("expected nil before sync, got %q", got)
	}
	page := bro.Page(t, "/")
	defer page.Close()
	waitContent(t, page, "#value", "0")
}
