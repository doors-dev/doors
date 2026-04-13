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
	"github.com/go-rod/rod/lib/input"
)

func waitReport(t *testing.T, page *rod.Page, id int, expected string) {
	t.Helper()
	deadline := time.Now().Add(1500 * time.Millisecond)
	for {
		if got := test.GetReportContent(t, page, id); got == expected {
			return
		} else if time.Now().After(deadline) {
			t.Fatalf("report-%d expected %q before timeout, got %q", id, expected, got)
		}
		<-time.After(25 * time.Millisecond)
	}
}

func TestKey(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &keyFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.Click(t, page, "#input")
	page.Keyboard.Press(input.KeyC)
	waitReport(t, page, 0, "c")
	waitReport(t, page, 1, "down")
	test.TestReportId(t, page, 2, "")
	test.TestReportId(t, page, 3, "")
	page.Keyboard.Release(input.KeyC)
	waitReport(t, page, 2, "c")
	waitReport(t, page, 3, "up")

	test.TestReportId(t, page, 4, "")
	test.TestReportId(t, page, 5, "")
	test.TestReportId(t, page, 6, "")

	page.Keyboard.Press(input.AltLeft)
	page.Keyboard.Press(input.KeyE)
	page.Keyboard.Release(input.KeyE)
	page.Keyboard.Release(input.AltLeft)
	waitReport(t, page, 6, "true")

	page.Keyboard.Press(input.ShiftLeft)
	page.Keyboard.Press(input.KeyE)
	page.Keyboard.Release(input.KeyE)
	page.Keyboard.Release(input.ShiftLeft)
	waitReport(t, page, 4, "true")

	page.Keyboard.Press(input.ControlLeft)
	page.Keyboard.Press(input.KeyE)
	page.Keyboard.Release(input.KeyE)
	page.Keyboard.Release(input.ControlLeft)
	waitReport(t, page, 5, "true")

}
