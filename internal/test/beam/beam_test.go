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
	b := doors.NewSourceNoSkip(state{})
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

func TestBeamReadAndSubSequentialRegistration(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamReadAndSubFragment{
				source: doors.NewSource(1),
				r:      test.NewReporter(6),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestReportId(t, page, 0, "v:1")

	test.Click(t, page, "#beam-read-sub-update-2")
	test.TestReportId(t, page, 1, "v:2")

	test.Click(t, page, "#beam-read-sub-register-source")
	test.TestReportId(t, page, 2, "2")

	test.Click(t, page, "#beam-read-sub-update-3")
	test.TestReportId(t, page, 3, "3")

	test.Click(t, page, "#beam-read-sub-register-derived-2")
	test.TestReportId(t, page, 4, "v:3")

	test.Click(t, page, "#beam-read-sub-update-4")
	test.TestReportId(t, page, 5, "v:4")
}

func TestLensRoundTripPropagation(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamLensRoundTripFragment{
				r: test.NewReporter(2),
			}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#lens-source", "1:a:1")
	test.TestContent(t, page, "#lens-int", "1:1")
	test.TestContent(t, page, "#lens-str", "a:1")
	test.TestContent(t, page, "#lens-even", "false:1")
	test.TestContent(t, page, "#lens-label", "label:1:1")
	test.TestContent(t, page, "#lens-parity", "odd:1")

	test.Click(t, page, "#lens-source-update")
	test.TestContent(t, page, "#lens-source", "2:b:2")
	test.TestContent(t, page, "#lens-int", "2:2")
	test.TestContent(t, page, "#lens-str", "b:2")
	test.TestContent(t, page, "#lens-even", "true:2")
	test.TestContent(t, page, "#lens-label", "label:2:2")
	test.TestContent(t, page, "#lens-parity", "even:2")

	test.Click(t, page, "#lens-int-update")
	test.TestContent(t, page, "#lens-source", "5:b:3")
	test.TestContent(t, page, "#lens-int", "5:3")
	test.TestContent(t, page, "#lens-str", "b:2")
	test.TestContent(t, page, "#lens-even", "false:3")
	test.TestContent(t, page, "#lens-label", "label:5:3")
	test.TestContent(t, page, "#lens-parity", "odd:3")

	test.Click(t, page, "#lens-str-update")
	test.TestContent(t, page, "#lens-source", "5:lens:4")
	test.TestContent(t, page, "#lens-int", "5:3")
	test.TestContent(t, page, "#lens-str", "lens:3")
	test.TestContent(t, page, "#lens-even", "false:3")
	test.TestContent(t, page, "#lens-label", "label:5:3")
	test.TestContent(t, page, "#lens-parity", "odd:3")

	test.Click(t, page, "#lens-int-mutate")
	test.TestContent(t, page, "#lens-source", "6:lens:5")
	test.TestContent(t, page, "#lens-int", "6:4")
	test.TestContent(t, page, "#lens-str", "lens:3")
	test.TestContent(t, page, "#lens-even", "true:4")
	test.TestContent(t, page, "#lens-label", "label:6:4")
	test.TestContent(t, page, "#lens-parity", "even:4")

	test.Click(t, page, "#lens-even-false")
	test.TestContent(t, page, "#lens-source", "7:lens:6")
	test.TestContent(t, page, "#lens-int", "7:5")
	test.TestContent(t, page, "#lens-str", "lens:3")
	test.TestContent(t, page, "#lens-even", "false:5")
	test.TestContent(t, page, "#lens-label", "label:7:5")
	test.TestContent(t, page, "#lens-parity", "odd:5")

	test.Click(t, page, "#lens-even-true")
	test.TestContent(t, page, "#lens-source", "8:lens:7")
	test.TestContent(t, page, "#lens-int", "8:6")
	test.TestContent(t, page, "#lens-str", "lens:3")
	test.TestContent(t, page, "#lens-even", "true:6")
	test.TestContent(t, page, "#lens-label", "label:8:6")
	test.TestContent(t, page, "#lens-parity", "even:6")

	test.Click(t, page, "#lens-report")
	test.TestReportId(t, page, 0, "report source-8-lens-true int-8-true str-lens-true even-true-true label-label:8-true")

	test.Click(t, page, "#lens-xupdate")
	test.TestReportId(t, page, 1, "x-ok")
	test.TestContent(t, page, "#lens-source", "9:lens:8")
	test.TestContent(t, page, "#lens-int", "9:7")
	test.TestContent(t, page, "#lens-str", "lens:3")
	test.TestContent(t, page, "#lens-even", "false:7")
	test.TestContent(t, page, "#lens-label", "label:9:7")
	test.TestContent(t, page, "#lens-parity", "odd:7")
}

func TestRouteSourceAndBeamState(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamRouteStateFragment{}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#route-source", "1::1")
	test.TestContent(t, page, "#route-lens-default-render", "1")
	test.TestContent(t, page, "#route-lens-default-value", "default:1:")
	test.TestContent(t, page, "#route-beam-default-render", "1")
	test.TestContent(t, page, "#route-beam-default-value", "beam-default:1:")
	test.TestMustNot(t, page, "#route-lens-render")
	test.TestMustNot(t, page, "#route-beam-render")
	test.TestMustNot(t, page, "#route-lens-list-render")
	test.TestMustNot(t, page, "#route-beam-list-render")

	test.Click(t, page, "#route-source-lens")
	test.TestContent(t, page, "#route-source", "2:doc:2")
	test.TestContent(t, page, "#route-lens-render", "1")
	test.TestContent(t, page, "#route-lens-value", "lens:doc")
	test.TestContent(t, page, "#route-lens-list-render", "1")
	test.TestContent(t, page, "#route-lens-list-value", "doc,2")
	test.TestContent(t, page, "#route-beam-render", "1")
	test.TestContent(t, page, "#route-beam-value", "beam:doc")
	test.TestContent(t, page, "#route-beam-list-render", "1")
	test.TestContent(t, page, "#route-beam-list-value", "doc,2")
	test.TestMustNot(t, page, "#route-lens-default-value")
	test.TestMustNot(t, page, "#route-beam-default-value")

	test.Click(t, page, "#route-source-next")
	test.TestContent(t, page, "#route-source", "2:next:3")
	test.TestContent(t, page, "#route-lens-render", "1")
	test.TestContent(t, page, "#route-lens-value", "lens:next")
	test.TestContent(t, page, "#route-lens-list-render", "1")
	test.TestContent(t, page, "#route-lens-list-value", "next,2")
	test.TestContent(t, page, "#route-beam-render", "1")
	test.TestContent(t, page, "#route-beam-value", "beam:next")
	test.TestContent(t, page, "#route-beam-list-render", "1")
	test.TestContent(t, page, "#route-beam-list-value", "next,2")

	test.Click(t, page, "#route-lens-update")
	test.TestContent(t, page, "#route-source", "2:child:4")
	test.TestContent(t, page, "#route-lens-render", "1")
	test.TestContent(t, page, "#route-lens-value", "lens:child")
	test.TestContent(t, page, "#route-lens-list-render", "1")
	test.TestContent(t, page, "#route-lens-list-value", "child,2")
	test.TestContent(t, page, "#route-beam-render", "1")
	test.TestContent(t, page, "#route-beam-value", "beam:child")
	test.TestContent(t, page, "#route-beam-list-render", "1")
	test.TestContent(t, page, "#route-beam-list-value", "child,2")

	test.Click(t, page, "#route-lens-clear")
	test.TestContent(t, page, "#route-source", "2::5")
	test.TestContent(t, page, "#route-lens-default-render", "2")
	test.TestContent(t, page, "#route-lens-default-value", "default:2:")
	test.TestContent(t, page, "#route-beam-default-render", "2")
	test.TestContent(t, page, "#route-beam-default-value", "beam-default:2:")
	test.TestMustNot(t, page, "#route-lens-value")
	test.TestMustNot(t, page, "#route-beam-value")
	test.TestMustNot(t, page, "#route-lens-list-value")
	test.TestMustNot(t, page, "#route-beam-list-value")

	test.Click(t, page, "#route-source-default-int")
	test.TestContent(t, page, "#route-source", "4::6")
	test.TestContent(t, page, "#route-lens-default-render", "2")
	test.TestContent(t, page, "#route-lens-default-value", "default:4:")
	test.TestContent(t, page, "#route-beam-default-render", "2")
	test.TestContent(t, page, "#route-beam-default-value", "beam-default:4:")
	test.TestMustNot(t, page, "#route-lens-list-value")
	test.TestMustNot(t, page, "#route-beam-list-value")
}

func TestRouteNoDefaultHandlesRapidUnmatch(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamRouteNoDefaultBurstFragment{}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#route-burst-source", "0:")
	test.TestMustNot(t, page, "#route-burst-lens")
	test.TestMustNot(t, page, "#route-burst-beam")

	test.Click(t, page, "#route-burst-none")
	test.TestContent(t, page, "#route-burst-source", "2:")
	test.TestMustNot(t, page, "#route-burst-lens")
	test.TestMustNot(t, page, "#route-burst-beam")

	test.Click(t, page, "#route-burst-hit")
	test.TestContent(t, page, "#route-burst-source", "3:after")
	test.TestContent(t, page, "#route-burst-lens", "lens:after")
	test.TestContent(t, page, "#route-burst-beam", "beam:after")
}

func TestRouteSourceAndDerivedBeamEntrypoints(t *testing.T) {
	bro := test.NewFragmentBro(browser,
		func() test.Fragment {
			return &BeamRouteEntrypointsFragment{}
		})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#route-entry-source", "1:a")
	test.TestContent(t, page, "#route-entry-lens", "lens:a")
	test.TestContent(t, page, "#route-entry-lens-beam", "beam:a")
	test.TestContent(t, page, "#route-entry-derived", "derived:odd")
	test.TestMustNot(t, page, "#route-entry-derived-default")
	test.TestMustNot(t, page, "#route-entry-simple-lens")
	test.TestMustNot(t, page, "#route-entry-simple-beam")

	test.Click(t, page, "#route-entry-lens-update")
	test.TestContent(t, page, "#route-entry-source", "1:b")
	test.TestContent(t, page, "#route-entry-lens", "lens:b")
	test.TestContent(t, page, "#route-entry-lens-beam", "beam:b")
	test.TestContent(t, page, "#route-entry-derived", "derived:odd")
	test.TestMustNot(t, page, "#route-entry-simple-lens")
	test.TestMustNot(t, page, "#route-entry-simple-beam")

	test.Click(t, page, "#route-entry-set-simple")
	test.TestContent(t, page, "#route-entry-source", "1:simple")
	test.TestContent(t, page, "#route-entry-lens", "lens:simple")
	test.TestContent(t, page, "#route-entry-lens-beam", "beam:simple")
	test.TestContent(t, page, "#route-entry-simple-lens", "simple-lens:simple")
	test.TestContent(t, page, "#route-entry-simple-beam", "simple-beam:simple")
	test.TestContent(t, page, "#route-entry-derived", "derived:odd")

	test.Click(t, page, "#route-entry-simple-lens-update")
	test.TestContent(t, page, "#route-entry-source", "1:simple-lens")
	test.TestContent(t, page, "#route-entry-lens", "lens:simple-lens")
	test.TestContent(t, page, "#route-entry-lens-beam", "beam:simple-lens")
	test.TestContent(t, page, "#route-entry-simple-lens", "simple-lens:simple-lens")
	test.TestContent(t, page, "#route-entry-simple-beam", "simple-beam:simple-lens")
	test.TestContent(t, page, "#route-entry-derived", "derived:odd")

	test.Click(t, page, "#route-entry-even")
	test.TestContent(t, page, "#route-entry-source", "2:simple-lens")
	test.TestContent(t, page, "#route-entry-lens", "lens:simple-lens")
	test.TestContent(t, page, "#route-entry-lens-beam", "beam:simple-lens")
	test.TestContent(t, page, "#route-entry-simple-lens", "simple-lens:simple-lens")
	test.TestContent(t, page, "#route-entry-simple-beam", "simple-beam:simple-lens")
	test.TestMustNot(t, page, "#route-entry-derived")
	test.TestContent(t, page, "#route-entry-derived-default", "derived-default:even")
}
