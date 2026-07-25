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

func TestXDyna(t *testing.T) {
	v1 := doors.IDRand()
	v2 := doors.IDRand()
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &xdynaFragment{
			r:  test.NewReporter(5),
			v1: v1,
			v2: v2,
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestAttr(t, page, "#target", "data-x", v1)

	// XUpdate that changes the value: dispatched -> channel yields nil.
	test.Click(t, page, "#x-update")
	waitReport(t, page, 0, "ok")
	test.TestAttr(t, page, "#target", "data-x", v2)

	// XUpdate with the same value: no call issued -> channel closes empty.
	test.Click(t, page, "#x-update-same")
	waitReport(t, page, 0, "noop")

	// XDisable: dispatched -> channel yields nil, attribute removed.
	test.Click(t, page, "#x-disable")
	waitReport(t, page, 1, "ok")
	test.TestAttrNo(t, page, "#target", "data-x")

	// XDisable again while already disabled: no call -> channel closes empty.
	test.Click(t, page, "#x-disable-again")
	waitReport(t, page, 1, "noop")

	// XEnable: dispatched -> channel yields nil, attribute back.
	test.Click(t, page, "#x-enable")
	waitReport(t, page, 2, "ok")
	test.TestAttr(t, page, "#target", "data-x", v2)
}
