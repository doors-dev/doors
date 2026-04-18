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

package beam

import (
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
)

func TestBeamBasics(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamUpdateFragment{
				r: test.NewReporter(1),
				b: doors.NewSource(state{}),
			}
		})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestReportId(t, page, 0, "0")
	test.Click(t, page, "#update")
	test.TestReportId(t, page, 0, "1")
	test.Click(t, page, "#mutate")
	test.TestReportId(t, page, 0, "2")
	test.Click(t, page, "#mutate-cancel")
	test.TestReportId(t, page, 0, "2")

}

func testConsistency(t *testing.T, f func() test.Fragment) {
	bro := test.NewFragmentBro(browser,
		f,
	)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestReportId(t, page, 1, "0")
	test.TestReportId(t, page, 2, "0")
	test.TestReportId(t, page, 3, "0")
	<-time.After(50 * time.Millisecond)
	test.TestReportId(t, page, 0, "3")
	test.TestReportId(t, page, 4, "3")
	test.Click(t, page, "#reload")
	test.TestReportId(t, page, 1, "3")
	test.TestReportId(t, page, 2, "3")
	test.TestReportId(t, page, 3, "3")
	test.TestReportId(t, page, 0, "6")
	<-time.After(50 * time.Millisecond)
	test.TestReportId(t, page, 4, "6")
}

func TestConsistent(t *testing.T) {
	testConsistency(t, func() test.Fragment {
		return &BeamConsistentFragment{
			r: test.NewReporter(1),
			b: doors.NewSource(state{}),
		}
	})
}
func TestDerive(t *testing.T) {
	testConsistency(t, func() test.Fragment {
		return &BeamDeriveFragment{
			r: test.NewReporter(1),
			b: doors.NewSource(state{}),
		}
	})
}

func TestSkip(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamSkipFragment{
				r: test.NewReporter(1),
				b: doors.NewSource(state{}),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.ClickNow(t, page, "#update1")
	test.ClickNow(t, page, "#update2")
	<-time.After(500 * time.Millisecond)
	test.TestReport(t, page, "init")
}

func TestNoSkip(t *testing.T) {
	b := doors.NewSource(state{})
	b.DisableSkipping()
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamSkipFragment{
				r: test.NewReporter(1),
				b: b,
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.ClickNow(t, page, "#update1")
	<-time.After(500 * time.Millisecond)
	test.TestReport(t, page, "propagated")
}

func TestEqualSubAndGo(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamEqualFragment{
				r: test.NewReporter(3),
				b: doors.NewSourceEqual(state{}, func(new state, old state) bool {
					return new.Int == old.Int
				}),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestReportId(t, page, 0, "0")
	test.TestContent(t, page, "#parity", "even")
	<-time.After(150 * time.Millisecond)
	test.TestReportId(t, page, 2, "go")

	test.Click(t, page, "#same")
	test.TestReportId(t, page, 0, "0")
	test.TestContent(t, page, "#parity", "even")

	test.Click(t, page, "#one")
	test.TestReportId(t, page, 0, "1")
	test.TestContent(t, page, "#parity", "odd")

	test.Click(t, page, "#three")
	test.TestReportId(t, page, 0, "3")
	test.TestContent(t, page, "#parity", "odd")

	test.Click(t, page, "#get")
	test.TestReportId(t, page, 1, "3")
}

func TestRenderBranchUpdateHoldsPropagation(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamRenderBranchUpdateFrameFragment{
				b: doors.NewSource(0),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	<-time.After(200 * time.Millisecond)
	test.TestContent(t, page, "#watcher-i", "1")
	test.TestContent(t, page, "#watcher-newi", "1")
}

func TestRenderBranchInitHoldsPropagation(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamRenderBranchInitFrameFragment{
				b: doors.NewSource(0),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	<-time.After(200 * time.Millisecond)
	test.TestContent(t, page, "#watcher-i", "0")
	test.TestContent(t, page, "#watcher-newi", "0")
}

func TestRenderUpdateWarningRepro(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamRenderUpdateWarningFragment{
				b: doors.NewSource(0),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.Click(t, page, "#warning-reload")
	<-time.After(200 * time.Millisecond)
}

func TestEffectSourceRerendersClosestDynamicParent(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamEffectSourceFragment{
				b: doors.NewSource(0),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#effect-source-value", "0")
	test.TestContent(t, page, "#effect-source-outer-renders", "1")
	test.TestContent(t, page, "#effect-source-inner-renders", "1")

	test.Click(t, page, "#effect-source-update-1")
	test.TestContent(t, page, "#effect-source-value", "1")
	test.TestContent(t, page, "#effect-source-outer-renders", "1")
	test.TestContent(t, page, "#effect-source-inner-renders", "2")

	test.Click(t, page, "#effect-source-update-2")
	test.TestContent(t, page, "#effect-source-value", "2")
	test.TestContent(t, page, "#effect-source-outer-renders", "1")
	test.TestContent(t, page, "#effect-source-inner-renders", "3")
}

func TestEffectDerivedBeamRerendersClosestDynamicParent(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamEffectDerivedFragment{
				b: doors.NewSource(0),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#effect-derived-value", "v:0")
	test.TestContent(t, page, "#effect-derived-outer-renders", "1")
	test.TestContent(t, page, "#effect-derived-inner-renders", "1")

	test.Click(t, page, "#effect-derived-update-1")
	test.TestContent(t, page, "#effect-derived-value", "v:1")
	test.TestContent(t, page, "#effect-derived-outer-renders", "1")
	test.TestContent(t, page, "#effect-derived-inner-renders", "2")

	test.Click(t, page, "#effect-derived-update-2")
	test.TestContent(t, page, "#effect-derived-value", "v:2")
	test.TestContent(t, page, "#effect-derived-outer-renders", "1")
	test.TestContent(t, page, "#effect-derived-inner-renders", "3")
}

func TestEffectMultipleDependenciesRerenderSameDynamicParent(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamEffectMultiFragment{
				left:  doors.NewSource(0),
				right: doors.NewSource(0),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#effect-multi-left", "0")
	test.TestContent(t, page, "#effect-multi-right", "0")
	test.TestContent(t, page, "#effect-multi-outer-renders", "1")
	test.TestContent(t, page, "#effect-multi-inner-renders", "1")

	test.Click(t, page, "#effect-multi-left-update")
	test.TestContent(t, page, "#effect-multi-left", "1")
	test.TestContent(t, page, "#effect-multi-right", "0")
	test.TestContent(t, page, "#effect-multi-outer-renders", "1")
	test.TestContent(t, page, "#effect-multi-inner-renders", "2")

	test.Click(t, page, "#effect-multi-right-update")
	test.TestContent(t, page, "#effect-multi-left", "1")
	test.TestContent(t, page, "#effect-multi-right", "1")
	test.TestContent(t, page, "#effect-multi-outer-renders", "1")
	test.TestContent(t, page, "#effect-multi-inner-renders", "3")
}

func TestEffectDuplicateDependencyRerendersOnce(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamEffectDuplicateFragment{
				b: doors.NewSource(0),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#effect-dup-first", "0")
	test.TestContent(t, page, "#effect-dup-second", "0")
	test.TestContent(t, page, "#effect-dup-outer-renders", "1")
	test.TestContent(t, page, "#effect-dup-inner-renders", "1")

	test.Click(t, page, "#effect-dup-update")
	test.TestContent(t, page, "#effect-dup-first", "1")
	test.TestContent(t, page, "#effect-dup-second", "1")
	test.TestContent(t, page, "#effect-dup-outer-renders", "1")
	test.TestContent(t, page, "#effect-dup-inner-renders", "2")
}
