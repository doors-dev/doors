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

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
)

func TestDyna(t *testing.T) {
	v1 := doors.IDRand()
	v2 := doors.IDRand()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &dynaFragment{
			v1: v1,
			v2: v2,
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.TestAttr(t, page, "#t1", "data-test1", v1)
	test.TestAttr(t, page, "#t2", "data-test1", v1)
	test.TestAttr(t, page, "#t3", "data-test1", v1)
	test.TestAttr(t, page, "#t1", "data-test2", v2)
	test.TestAttr(t, page, "#t2", "data-test2", v2)
	test.TestAttr(t, page, "#t3", "data-test2", v2)

	test.Click(t, page, "#disable-2")
	test.TestAttrNo(t, page, "#t1", "data-test2")
	test.TestAttrNo(t, page, "#t2", "data-test2")
	test.TestAttrNo(t, page, "#t3", "data-test2")

	test.Click(t, page, "#update-2")
	test.TestAttrNo(t, page, "#t1", "data-test2")
	test.TestAttrNo(t, page, "#t2", "data-test2")
	test.TestAttrNo(t, page, "#t3", "data-test2")

	test.Click(t, page, "#enable-2")
	test.TestAttr(t, page, "#t1", "data-test2", v1)
	test.TestAttr(t, page, "#t2", "data-test2", v1)
	test.TestAttr(t, page, "#t3", "data-test2", v1)

	test.Click(t, page, "#disable-1")
	test.TestAttrNo(t, page, "#t1", "data-test1")
	test.TestAttrNo(t, page, "#t2", "data-test1")
	test.TestAttrNo(t, page, "#t3", "data-test1")

	test.Click(t, page, "#clear")
	test.Click(t, page, "#update-1")
	test.Click(t, page, "#disable-2")
	test.Click(t, page, "#enable-1")
	test.Click(t, page, "#show")
	test.TestAttr(t, page, "#t1", "data-test1", v2)
	test.TestAttr(t, page, "#t2", "data-test1", v2)
	test.TestAttr(t, page, "#t3", "data-test1", v2)
	test.TestAttrNo(t, page, "#t1", "data-test2")
	test.TestAttrNo(t, page, "#t2", "data-test2")
	test.TestAttrNo(t, page, "#t3", "data-test2")

	test.Click(t, page, "#enable-2")
	test.Click(t, page, "#replace")

	test.TestAttr(t, page, "#t1", "data-test1", v2)
	test.TestAttr(t, page, "#t2", "data-test1", v2)
	test.TestAttr(t, page, "#t3", "data-test1", v2)
	test.TestAttr(t, page, "#t1", "data-test2", v1)
	test.TestAttr(t, page, "#t2", "data-test2", v1)
	test.TestAttr(t, page, "#t3", "data-test2", v1)

}
