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

	"github.com/doors-dev/doors/internal/test"
)

func TestEmitterPointer(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &emitterPointerFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestReportId(t, page, 0, "")

	test.Click(t, page, "#emit-click")
	waitReport(t, page, 0, "click")
	waitReport(t, page, 1, "2")
	waitReport(t, page, 2, "2")

	test.Click(t, page, "#emit-down")
	waitReport(t, page, 0, "pointerdown")

	test.Click(t, page, "#emit-up")
	waitReport(t, page, 0, "pointerup")
}

func TestEmitterKeyboard(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &emitterKeyFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.Click(t, page, "#emit-keydown")
	waitReport(t, page, 0, "keydown")
	waitReport(t, page, 1, "a")
	waitReport(t, page, 2, "KeyA")

	test.Click(t, page, "#emit-keyup")
	waitReport(t, page, 0, "keyup")
	waitReport(t, page, 1, "A")
	waitReport(t, page, 3, "true")
	waitReport(t, page, 4, "true")
}

func TestEmitterFocus(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &emitterFocusFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.Click(t, page, "#emit-focus")
	waitReport(t, page, 0, "focus")

	test.Click(t, page, "#emit-blur")
	waitReport(t, page, 0, "blur")

	test.Click(t, page, "#emit-focusin")
	waitReport(t, page, 1, "in")

	test.Click(t, page, "#emit-focusout")
	waitReport(t, page, 1, "out")
}

func TestEmitterForm(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &emitterFormFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.Click(t, page, "#emit-input")
	waitReport(t, page, 0, "input")
	waitReport(t, page, 1, "hey")

	test.Click(t, page, "#emit-change")
	waitReport(t, page, 2, "change")
	waitReport(t, page, 3, "field")

	test.Click(t, page, "#emit-submit")
	waitReport(t, page, 4, "submit")
}

func TestEmitterMulti(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &emitterMultiFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.Click(t, page, "#emit-count")
	waitReport(t, page, 0, "2")
	waitReport(t, page, 1, "hit")
	waitReport(t, page, 2, "hit")
}
