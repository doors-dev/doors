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

func TestSysEmit(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &sysEmitFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	waitReport(t, page, 0, "2")
	waitReport(t, page, 1, "2")
	waitReport(t, page, 2, "2")
	waitReport(t, page, 3, "2")
}

func TestSysEmitDerivedSubmit(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &sysEmitSubmitFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	waitReport(t, page, 0, "1")
	waitReport(t, page, 1, "1")
}

func TestSysEmitDerivedSubmitWithClick(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &sysEmitSubmitFragment{
			r:     test.NewReporter(10),
			click: true,
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	waitReport(t, page, 0, "2")
	waitReport(t, page, 1, "2")
}

func TestSysEmitDerivedSubmitError(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &sysEmitSubmitFragment{
			r:    test.NewReporter(10),
			fail: true,
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	waitReport(t, page, 0, "-1")
	waitReport(t, page, 1, "0")
}

func TestSysEmitUnrelated(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &sysEmitUnrelatedFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	waitReport(t, page, 0, "hit")
	waitReport(t, page, 1, "hit")
	waitReport(t, page, 2, "1")
}

func TestSysEmitDisabled(t *testing.T) {
	for _, mode := range []string{"disabled", "fieldset", "inert"} {
		t.Run(mode, func(t *testing.T) {
			bro := test.NewFragmentBro(browser, func() test.Fragment {
				return &sysEmitDisabledFragment{
					r:    test.NewReporter(10),
					mode: mode,
				}
			})
			defer bro.Close()
			page := bro.Page(t, "/")
			defer page.Close()

			waitReport(t, page, 1, "canceled: target is disabled")
			test.TestReportId(t, page, 0, "")
		})
	}
}

func TestSysEmitEnabled(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &sysEmitDisabledFragment{
			r: test.NewReporter(10),
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	waitReport(t, page, 1, "ok")
	waitReport(t, page, 0, "hit")
}
