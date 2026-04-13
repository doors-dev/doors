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
)

func TestCaptureError(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &errorFragment{}
	})
	page := bro.Page(t, "/")
	// initial
	test.TestContent(t, page, "#report", "initial")

	// root error handler
	test.Click(t, page, "#err_1")
	test.TestContent(t, page, "#report", "root_error/err_1")

	// root normal handler
	test.Click(t, page, "#err_2")
	test.TestContent(t, page, "#report", "root/err_2")

	// n1 scope: error routed to n1 expectation
	test.Click(t, page, "#err_3")
	test.TestContent(t, page, "#report", "n1_error/err_3")

	// n1 scope: normal event
	test.Click(t, page, "#err_4")
	test.TestContent(t, page, "#report", "n1/err_4")

	// n2 scope: normal event
	test.TestContent(t, page, "#indicator", "init")
	test.TestAttrNo(t, page, "#indicator", "data-indicator")
	test.Click(t, page, "#err_5")
	test.TestContent(t, page, "#report", "n2/err_5")
	test.TestContent(t, page, "#indicator", "indicator")
	test.TestAttr(t, page, "#indicator", "data-indicator", "true")
	<-time.After(500 * time.Millisecond)
	test.TestContent(t, page, "#indicator", "init")
	test.TestAttrNo(t, page, "#indicator", "data-indicator")

	// n2 scope: error routed to n1 expectation
	test.Click(t, page, "#err_6")
	test.TestContent(t, page, "#report", "n1_error/err_6")
}
