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

func linkOpen(t *testing.T) (*rod.Page, *gate) {
	t.Helper()
	g := newGate()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &linkFragment{g: g}
	})
	page := bro.Page(t, "/")
	t.Cleanup(bro.Close)
	t.Cleanup(func() { page.Close() })
	return page, g
}

// A plain link shows the parts of the new view as they get ready: the fast one changes while the slow one still renders.
func TestLinkPlainStreams(t *testing.T) {
	page, g := linkOpen(t)
	waitContent(t, page, "#slow", "slow-")
	test.ClickNow(t, page, "#link-plain")
	waitEntered(t, g)
	waitContent(t, page, "#fast", "fast-plain")
	if got := content(t, page, "#slow"); got != "slow-" {
		t.Fatalf("expected the slow part still rendering, got %q", got)
	}
	close(g.proceed)
	waitContent(t, page, "#slow", "slow-plain")
	pathWait(t, page, "/s/plain")
}

// A link with Batch keeps the current view until the whole new one is ready, then swaps it and the URL at once.
func TestLinkBatchSwapsAtOnce(t *testing.T) {
	page, g := linkOpen(t)
	waitContent(t, page, "#slow", "slow-")
	test.ClickNow(t, page, "#link-batched")
	waitEntered(t, g)
	time.Sleep(quiet)
	if fast, slow := content(t, page, "#fast"), content(t, page, "#slow"); fast != "fast-" || slow != "slow-" {
		t.Fatalf("expected the current view intact while the new one renders, got %q and %q", fast, slow)
	}
	close(g.proceed)
	waitContent(t, page, "#fast", "fast-batched")
	waitContent(t, page, "#slow", "slow-batched")
	pathWait(t, page, "/s/batched")
}
