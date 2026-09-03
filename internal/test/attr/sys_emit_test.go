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
