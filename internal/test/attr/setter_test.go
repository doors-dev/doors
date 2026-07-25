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

func TestSetter(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &setterFragment{
			r: test.NewReporter(2),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestAttrNo(t, page, "#t1", "data-test")
	test.TestAttrNo(t, page, "#t2", "data-test")
	test.TestAttrNo(t, page, "#t3", "data-test")

	test.Click(t, page, "#set-string")
	test.TestAttr(t, page, "#t1", "data-test", `a&b"c"<d>`)
	test.TestAttr(t, page, "#t2", "data-test", `a&b"c"<d>`)
	test.TestAttr(t, page, "#t3", "data-test", `a&b"c"<d>`)

	test.Click(t, page, "#set-bare")
	test.TestAttr(t, page, "#t1", "data-test", "")
	test.TestAttr(t, page, "#t2", "data-test", "")
	test.TestAttr(t, page, "#t3", "data-test", "")

	test.Click(t, page, "#set-int")
	test.TestAttr(t, page, "#t1", "data-test", "42")
	test.TestAttr(t, page, "#t2", "data-test", "42")
	test.TestAttr(t, page, "#t3", "data-test", "42")

	test.Click(t, page, "#remove-nil")
	test.TestAttrNo(t, page, "#t1", "data-test")
	test.TestAttrNo(t, page, "#t2", "data-test")
	test.TestAttrNo(t, page, "#t3", "data-test")

	test.Click(t, page, "#set-string")
	test.TestAttr(t, page, "#t1", "data-test", `a&b"c"<d>`)
	test.Click(t, page, "#remove-false")
	test.TestAttrNo(t, page, "#t1", "data-test")
	test.TestAttrNo(t, page, "#t2", "data-test")
	test.TestAttrNo(t, page, "#t3", "data-test")

	test.Click(t, page, "#xset")
	waitReport(t, page, 0, "3")
	test.TestAttr(t, page, "#t1", "data-count", "x")
	test.TestAttr(t, page, "#t2", "data-count", "x")
	test.TestAttr(t, page, "#t3", "data-count", "x")

	test.Click(t, page, "#set-string")
	test.TestAttr(t, page, "#t1", "data-test", `a&b"c"<d>`)
	test.Click(t, page, "#clear")
	test.Click(t, page, "#show")
	test.TestAttrNo(t, page, "#t1", "data-test")
	test.TestAttrNo(t, page, "#t2", "data-test")
	test.TestAttrNo(t, page, "#t3", "data-test")

	test.Click(t, page, "#set-int")
	test.TestAttr(t, page, "#t1", "data-test", "42")
	test.TestAttr(t, page, "#t2", "data-test", "42")
	test.TestAttr(t, page, "#t3", "data-test", "42")
}
