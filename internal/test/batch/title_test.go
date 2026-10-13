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

package batch

import (
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func titleOpen(t *testing.T) (*rod.Page, *gate) {
	t.Helper()
	g := newGate()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &titleFragment{g: g}
	})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	return page, g
}

func titleOf(t *testing.T, page *rod.Page) string {
	t.Helper()
	return page.MustEval("() => document.title").String()
}

func titleWait(t *testing.T, page *rod.Page, want string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for {
		got := titleOf(t, page)
		if got == want {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("title: expected %q, got %q", want, got)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// Title updates are not batched: the title of content re-rendered by a Batch hook changes while the content is still held.
func TestTitleNotBatched(t *testing.T) {
	page, g := titleOpen(t)
	titleWait(t, page, "first")
	test.ClickNow(t, page, "#title-batched")
	waitEntered(t, g)
	titleWait(t, page, "second")
	if got := content(t, page, "#a"); got != "a0" {
		t.Fatalf("expected the content held, got %q", got)
	}
	close(g.proceed)
	waitContent(t, page, "#a", "a1")
	waitContent(t, page, "#b", "b1")
	titleWait(t, page, "second")
}
