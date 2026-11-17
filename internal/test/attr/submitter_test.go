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

func submitterPage(t *testing.T) (*test.Bro, *rod.Page) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &submitterFragment{r: test.NewReporter(2)}
	})
	return bro, bro.Page(t, "/")
}

func expectSubmitter(t *testing.T, page *rod.Page, pressed string, other string, count string) {
	t.Helper()
	<-time.After(50 * time.Millisecond)
	test.TestAttrNow(t, page, pressed, "data-busy", "yes")
	test.TestAttrNo(t, page, other, "data-busy")
	test.TestAttrNo(t, page, "#sform", "data-busy")
	waitReportId(t, page, 1, count, 2*time.Second)
	waitAttrNo(t, page, pressed, "data-busy", time.Second)
}

func TestSubmitterClick(t *testing.T) {
	bro, page := submitterPage(t)
	defer bro.Close()
	defer page.Close()

	test.ClickNow(t, page, "#publish")
	expectSubmitter(t, page, "#publish", "#draft", "1")
	test.TestReportId(t, page, 0, "publish")

	test.ClickNow(t, page, "#draft")
	expectSubmitter(t, page, "#draft", "#publish", "2")
	test.TestReportId(t, page, 0, "draft")

	_, err := page.Eval(`() => document.querySelector("#publish").dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }))`)
	if err != nil {
		t.Fatal(err)
	}
	expectSubmitter(t, page, "#publish", "#draft", "3")
	test.TestReportId(t, page, 0, "publish")
}

func TestSubmitterEnter(t *testing.T) {
	bro, page := submitterPage(t)
	defer bro.Close()
	defer page.Close()

	test.Click(t, page, "#title")
	page.Keyboard.Press(input.Enter)
	page.Keyboard.Release(input.Enter)
	expectSubmitter(t, page, "#draft", "#publish", "1")
	test.TestReportId(t, page, 0, "draft")
}

func TestSubmitterNone(t *testing.T) {
	bro, page := submitterPage(t)
	defer bro.Close()
	defer page.Close()

	_, err := page.Eval(`() => document.querySelector("#sform").requestSubmit()`)
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(50 * time.Millisecond)
	test.TestAttrNo(t, page, "#draft", "data-busy")
	test.TestAttrNo(t, page, "#publish", "data-busy")
	test.TestAttrNo(t, page, "#sform", "data-busy")
	waitReportId(t, page, 1, "1", 2*time.Second)
	test.TestReportId(t, page, 0, "none")
}
