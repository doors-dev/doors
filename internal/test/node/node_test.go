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

package door

import (
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
	"github.com/go-rod/rod"
	"strconv"
	"testing"
	"time"
)

func TestDoorLoadPage(t *testing.T) {
	bro := test.NewPathBro(browser, func(r test.PathLens) gox.Comp {
		return &test.Page{
			Source: r,
			Header: "Page Door",
		}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	h1Text := page.MustElement("h1").MustText()
	if h1Text != "Page Door" {
		t.Fatal("header missmatch")
	}
}

func TestDoorInitialContent(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &BeforeFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
}

func TestDoorProxyWrapsMultipleRoots(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentProxyWrappedSiblings{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()

	if c := test.Count(page, "d0-r"); c != 1 {
		t.Fatal("expected one d0-r wrapper, got", c)
	}
	test.TestMust(t, page, "d0-r > #proxy-wrap-first")
	test.TestMust(t, page, "d0-r > #proxy-wrap-second")
}

func TestDoorProxyWrapsLoopRoots(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentProxyWrappedLoop{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()

	if c := test.Count(page, "d0-r"); c != 1 {
		t.Fatal("expected one d0-r wrapper for loop proxy, got", c)
	}
	test.TestMust(t, page, "d0-r > #proxy-loop-0")
	test.TestMust(t, page, "d0-r > #proxy-loop-1")
}

func DoorUpdatedBefore(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &BeforeFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#updated")
}

func TestDoorRemovedBefore(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &BeforeFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMustNot(t, page, "#removed")
}

func TestDoorReplacedBefore(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &BeforeFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#initReplaced")

	// test.TestMust(t, page, "body > #replaced")

}

func TestDoorDynamic(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &DynamicFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
	test.Click(t, page, "#update")
	test.TestMustNot(t, page, "#init")
	test.TestMust(t, page, "#updated")
	test.Click(t, page, "#replace")
	test.TestMustNot(t, page, "#updated")
	test.TestMust(t, page, "#replaced")
	test.Click(t, page, "#remove")
	test.TestMustNot(t, page, "#replaced")
}

func TestDoorEmbedded(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &EmbeddedFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")

}

func TestDoorEmbeddedRemove(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &EmbeddedFragment{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
	test.Click(t, page, "#clear")
	test.TestMustNot(t, page, "#init")
	test.Click(t, page, "#remove")
	test.Click(t, page, "#replace")
	test.TestMustNot(t, page, "#temp")
	test.TestMust(t, page, "#replaced")
}

func TestDoorUpdateX(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentX{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
	test.Click(t, page, "#updatex")
	test.TestReport(t, page, "ok upd")
	test.TestMustNot(t, page, "#init")
	test.TestMust(t, page, "#updated")
	test.Click(t, page, "#removex")
	test.TestReport(t, page, "ok del")
	test.TestMustNot(t, page, "#updated")
	test.Click(t, page, "#updatex")
	test.TestReport(t, page, "channel closed")
}

func TestDoorXLifecycle(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentXDoor{}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()

	test.TestMust(t, page, "#x-init")

	test.Click(t, page, "#xreload")
	test.TestReport(t, page, "ok reload")
	test.TestMust(t, page, "#x-init")

	test.Click(t, page, "#xrebase")
	test.TestReport(t, page, "ok rebase")
	test.TestMust(t, page, "#x-rebased-root")
	test.TestMust(t, page, "#x-rebased")

	test.Click(t, page, "#xclear")
	test.TestReport(t, page, "ok clear")
	test.TestMustNot(t, page, "#x-rebased")

	test.Click(t, page, "#xupdate")
	test.TestReport(t, page, "ok update")
	test.TestMust(t, page, "#x-updated")

	test.Click(t, page, "#xunmount")
	test.TestReport(t, page, "ok unmount")
	test.TestMustNot(t, page, "#x-updated")

	test.Click(t, page, "#xremount")
	test.TestMust(t, page, "#x-updated")

	test.Click(t, page, "#xreplace")
	test.TestReport(t, page, "ok replace")
	test.TestMust(t, page, "#x-replaced")
}

func TestDoorMultiple(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentMany{}
	})

	defer bro.Close()
	page := bro.Page(t, "/")

	defer page.Close()
	<-time.After(100 * time.Millisecond)
	c := test.Count(page, ".sample")
	if c != 1 {
		println(page.MustHTML())
		t.Fatal("Counted before upated, need 1, got", c)
	}
	test.Click(t, page, "#replace")
	c = test.Count(page, ".sample")
	if c != 100 {
		t.Fatal("Counted after reaplce, need 100, got", c)
	}
}

func TestDoorLifeCycle(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &LifeCycleFragment{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.TestMust(t, page, "#presist")
	test.TestMustNot(t, page, "#new")
	test.Click(t, page, "#reload")
	test.TestMust(t, page, "#presist")
	test.Click(t, page, "#updateEmpty")
	test.TestMustNot(t, page, "#presist")
	test.TestMust(t, page, "#new")
	test.Click(t, page, "#updateInner")
	test.TestMust(t, page, "#new")
	test.TestMust(t, page, "#inner-maintained")
	test.Click(t, page, "#updateEditor")
	test.TestMust(t, page, "#new")
	test.TestMust(t, page, "#inner-maintained")
	test.Click(t, page, "#updateEmptyAlt")
	test.TestMustNot(t, page, "#new")
	test.TestMust(t, page, "#new-alt")
	test.TestMust(t, page, "#inner-maintained")
	test.Click(t, page, "#updateContent")
	test.TestMustNot(t, page, "#inner-maintained")
	test.TestMustNot(t, page, "#new-alt")
	test.TestMustNot(t, page, "#new")
	test.TestMustNot(t, page, "#presist")
	test.TestMust(t, page, "#presist2")
	test.TestMust(t, page, "#new2")

	test.Click(t, page, "#updateEditor")
	test.TestMustNot(t, page, "#inner-maintained")
	test.TestMust(t, page, "#presist2")
	test.TestMust(t, page, "#new2")

	test.Click(t, page, "#clear")
	test.TestMustNot(t, page, "#presist2")
	test.TestMust(t, page, "#new2")
	test.Click(t, page, "#updateEditor")
	test.TestMustNot(t, page, "#presist2")
	test.TestMust(t, page, "#new2")
	test.Click(t, page, "#updateEditor")
	test.TestMustNot(t, page, "#presist2")
	test.TestMust(t, page, "#new2")

	test.Click(t, page, "#updateContent")
	test.TestMust(t, page, "#new2")
	test.TestMust(t, page, "#presist2")

	test.Click(t, page, "#remove")
	test.TestMustNot(t, page, "#new2")
	test.TestMustNot(t, page, "#presist2")

	test.Click(t, page, "#updateEditor")
	test.TestMustNot(t, page, "#new2")
	test.TestMustNot(t, page, "#presist2")

	test.Click(t, page, "#updateContent")
	test.TestMust(t, page, "#new2")
	test.TestMust(t, page, "#presist2")

	test.Click(t, page, "#unmount")
	test.TestMustNot(t, page, "#new2")
	test.TestMustNot(t, page, "#presist2")

	test.Click(t, page, "#updateEditor")
	test.TestMust(t, page, "#new2")
	test.TestMust(t, page, "#presist2")

	test.Click(t, page, "#updateOuter")
	test.TestMustNot(t, page, "#new2")
	test.TestMustNot(t, page, "#presist2")
	test.TestMust(t, page, "#outer-root")
	test.TestMust(t, page, "#outer-presist")

	test.Click(t, page, "#updateEditor")
	test.TestMust(t, page, "#outer-root")
	test.TestMust(t, page, "#outer-presist")

	test.Click(t, page, "#replaceStatic")
	test.TestMustNot(t, page, "#outer-root")
	test.TestMust(t, page, "#static-presist")

	test.Click(t, page, "#updateEditor")
	test.TestMust(t, page, "#static-presist")

	test.Click(t, page, "#updateContent")
	test.TestMustNot(t, page, "#static-presist")
	test.TestMust(t, page, "#new2")
	test.TestMust(t, page, "#presist2")
}

func TestDoorProxyReloadPreservesUpdatedContent(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentProxyReloadContent{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestMust(t, page, "#proxy-redraw-root")
	test.TestMustNot(t, page, "#proxy-redraw-content")

	test.Click(t, page, "#proxy-redraw-update")
	test.TestMust(t, page, "#proxy-redraw-root")
	test.TestMust(t, page, "#proxy-redraw-content")

	test.Click(t, page, "#proxy-redraw-remount")
	test.TestMust(t, page, "#proxy-redraw-root")
	test.TestMust(t, page, "#proxy-redraw-content")

	test.Click(t, page, "#proxy-redraw-reload")
	test.TestMust(t, page, "#proxy-redraw-root")
	test.TestMust(t, page, "#proxy-redraw-content")
}

func TestDoorReloadUsesClosestDynamicParent(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentClosestReload{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#outer-count", "outer-1")
	test.TestContent(t, page, "#inner-count", "inner-1")

	test.Click(t, page, "#reload-nearest")
	test.TestContent(t, page, "#outer-count", "outer-1")
	test.TestContent(t, page, "#inner-count", "inner-2")

	test.Click(t, page, "#reload-nearest")
	test.TestContent(t, page, "#outer-count", "outer-1")
	test.TestContent(t, page, "#inner-count", "inner-3")
}

func TestDoorReloadUsesClosestDynamicParentInProxySyntax(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentClosestReloadProxy{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#proxy-outer-count", "outer-1")
	test.TestContent(t, page, "#proxy-inner-count", "inner-1")

	test.Click(t, page, "#reload-nearest-proxy")
	test.TestContent(t, page, "#proxy-outer-count", "outer-1")
	test.TestContent(t, page, "#proxy-inner-count", "inner-2")

	test.Click(t, page, "#reload-nearest-proxy")
	test.TestContent(t, page, "#proxy-outer-count", "outer-1")
	test.TestContent(t, page, "#proxy-inner-count", "inner-3")
}

func TestInlineDoorPointerProxy(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentInlineDoorPointerProxy{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestMust(t, page, "#inline-door-root")
	test.TestContent(t, page, "#inline-door-count", "inline-1")

	test.Click(t, page, "#inline-door-reload")
	test.TestContent(t, page, "#inline-door-count", "inline-2")
}

func TestDoorContainerHookSurvivesInner(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentContainerInnerLifecycle{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#container-inner-value", "initial")

	test.Click(t, page, "#container-inner-root")
	test.TestContent(t, page, "#container-inner-value", "click-1")

	test.Click(t, page, "#container-inner-root")
	test.TestContent(t, page, "#container-inner-value", "click-2")
}

func TestDoorContainerEffectSurvivesInner(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentContainerEffectLifecycle{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestAttr(t, page, "#container-effect-root", "data-container-effect", "0")
	test.TestContent(t, page, "#container-effect-value", "initial-0")

	test.Click(t, page, "#container-effect-inner")
	test.TestContent(t, page, "#container-effect-value", "inner-0")

	test.Click(t, page, "#container-effect-update")
	test.TestContent(t, page, "#container-effect-value", "inner-1")
}

func TestDoorContainerOuterCleansContainerTracker(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentContainerOuterLifecycle{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#container-outer-value", "initial")

	test.Click(t, page, "#container-outer-root")
	test.TestMustNot(t, page, "#container-outer-root")
	test.TestMust(t, page, "#container-outer-new-root")
	test.TestContent(t, page, "#container-outer-value", "outer")

	test.Click(t, page, "#container-outer-report")
	test.TestReport(t, page, "cancels-1 watches-1")

	test.Click(t, page, "#container-outer-new-root")
	test.TestContent(t, page, "#container-outer-value", "outer-click-1")
}

func TestDoorContainerReloadCleansAndRebindsContainer(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentContainerReloadLifecycle{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#container-reload-value", "initial")

	test.Click(t, page, "#container-reload-root")
	test.TestContent(t, page, "#container-reload-value", "click-1")

	test.Click(t, page, "#container-reload-report")
	test.TestReport(t, page, "cancels-1 watches-2")

	test.Click(t, page, "#container-reload-root")
	test.TestContent(t, page, "#container-reload-value", "click-2")

	test.Click(t, page, "#container-reload-report")
	test.TestReport(t, page, "cancels-2 watches-3")
}

func TestDoorContainerHookStateSurvivesInner(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentContainerHookStateLifecycle{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#container-hook-state-value", "initial-0")

	test.Click(t, page, "#container-hook-state-root")
	test.TestContent(t, page, "#container-hook-state-value", "registered-0")

	test.Click(t, page, "#container-hook-state-report")
	test.TestReport(t, page, "state read-0-true derived-derived-0-true initial-0-true watch-true value-0 sub-0 watches-1 cancels-0")

	test.Click(t, page, "#container-hook-state-update")
	test.Click(t, page, "#container-hook-state-report")
	test.TestReport(t, page, "state read-0-true derived-derived-0-true initial-0-true watch-true value-1 sub-1 watches-2 cancels-0")

	test.Click(t, page, "#container-hook-state-root")
	test.TestContent(t, page, "#container-hook-state-value", "mutated-2")

	test.Click(t, page, "#container-hook-state-report")
	test.TestReport(t, page, "state read-0-true derived-derived-0-true initial-0-true watch-true value-2 sub-2 watches-3 cancels-0")
}

func TestDoorContainerHookStateCanceledByOuter(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentContainerHookStateOuterLifecycle{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#container-hook-outer-value", "initial")

	test.Click(t, page, "#container-hook-outer-root")
	test.TestMustNot(t, page, "#container-hook-outer-root")
	test.TestContent(t, page, "#container-hook-outer-value", "outer")

	test.Click(t, page, "#container-hook-outer-report")
	test.TestReport(t, page, "state read-0-true derived-derived-0-true initial-0-true watch-true value-0 sub-0 watches-1 cancels-1")

	test.Click(t, page, "#container-hook-outer-update")
	test.Click(t, page, "#container-hook-outer-report")
	test.TestReport(t, page, "state read-0-true derived-derived-0-true initial-0-true watch-true value-1 sub-0 watches-1 cancels-1")

	test.Click(t, page, "#container-hook-outer-new-root")
	test.TestContent(t, page, "#container-hook-outer-value", "outer-click-1")

	test.Click(t, page, "#container-hook-outer-report")
	test.TestReport(t, page, "state read-1-true derived-derived-1-true initial-1-true watch-true value-2 sub-1 watches-3 cancels-1")
}

func triggerBrowserKey(t *testing.T, page *rod.Page, selector string) {
	t.Helper()
	page.MustEval(`() => {
		const el = document.querySelector(` + strconv.Quote(selector) + `)
		if (!(el instanceof HTMLElement)) {
			throw new Error("event target not found: " + ` + strconv.Quote(selector) + `)
		}
		el.dispatchEvent(new KeyboardEvent("keydown", {
			key: "ContainerEffect",
			code: "KeyE",
			bubbles: true,
			cancelable: true,
		}))
	}`)
	<-time.After(200 * time.Millisecond)
}

func TestDoorContainerHookEffectReloadsContainer(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentContainerHookEffectLifecycle{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#container-hook-effect-renders", "1")
	test.TestContent(t, page, "#container-hook-effect-value", "outer-0")

	triggerBrowserKey(t, page, "#container-hook-effect-root")
	test.TestContent(t, page, "#container-hook-effect-renders", "2")
	test.TestContent(t, page, "#container-hook-effect-value", "registered-0")

	test.Click(t, page, "#container-hook-effect-report")
	test.TestReport(t, page, "state effect-0-true value-0 renders-2 registrations-1")

	test.Click(t, page, "#container-hook-effect-update")
	test.TestContent(t, page, "#container-hook-effect-renders", "3")
	test.TestContent(t, page, "#container-hook-effect-value", "registered-1")

	test.Click(t, page, "#container-hook-effect-report")
	test.TestReport(t, page, "state effect-0-true value-1 renders-3 registrations-1")

	test.Click(t, page, "#container-hook-effect-update")
	test.TestContent(t, page, "#container-hook-effect-renders", "3")
	test.TestContent(t, page, "#container-hook-effect-value", "registered-1")

	test.Click(t, page, "#container-hook-effect-report")
	test.TestReport(t, page, "state effect-0-true value-2 renders-3 registrations-1")

	triggerBrowserKey(t, page, "#container-hook-effect-root")
	test.TestContent(t, page, "#container-hook-effect-renders", "4")
	test.TestContent(t, page, "#container-hook-effect-value", "registered-2")

	test.Click(t, page, "#container-hook-effect-update")
	test.TestContent(t, page, "#container-hook-effect-renders", "5")
	test.TestContent(t, page, "#container-hook-effect-value", "registered-3")
}

func TestDoorContainerHookStateCanceledAndReboundByReload(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentContainerHookStateReloadLifecycle{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#container-hook-reload-value", "initial")

	test.Click(t, page, "#container-hook-reload-root")
	test.TestContent(t, page, "#container-hook-reload-value", "reload-1")

	test.Click(t, page, "#container-hook-reload-report")
	test.TestReport(t, page, "state read-0-true derived-derived-0-true initial-0-true watch-true value-0 sub-0 watches-1 cancels-1")

	test.Click(t, page, "#container-hook-reload-update")
	test.Click(t, page, "#container-hook-reload-report")
	test.TestReport(t, page, "state read-0-true derived-derived-0-true initial-0-true watch-true value-1 sub-0 watches-1 cancels-1")

	test.Click(t, page, "#container-hook-reload-root")
	test.TestContent(t, page, "#container-hook-reload-value", "reload-2")

	test.Click(t, page, "#container-hook-reload-report")
	test.TestReport(t, page, "state read-1-true derived-derived-1-true initial-1-true watch-true value-1 sub-0 watches-2 cancels-2")
}

func TestDoorTrackedReloadUsesClosestDynamicParent(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentClosestTrackedReload{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestContent(t, page, "#x-outer-count", "outer-1")
	test.TestContent(t, page, "#x-inner-count", "inner-1")

	test.Click(t, page, "#xreload-nearest")
	test.TestReport(t, page, "ok xreload")
	test.TestContent(t, page, "#x-outer-count", "outer-1")
	test.TestContent(t, page, "#x-inner-count", "inner-2")

	test.Click(t, page, "#xreload-nearest")
	test.TestReport(t, page, "ok xreload")
	test.TestContent(t, page, "#x-outer-count", "outer-1")
	test.TestContent(t, page, "#x-inner-count", "inner-3")
}

func TestDoorTrackedReloadFromRootReturnsError(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentRootTrackedReload{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.Click(t, page, "#root-xreload")
	test.TestReport(t, page, "channel err: root door cannot be reloaded")
}

func TestDoorDetachedReplaceTransitions(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentDetachedReplace{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestMust(t, page, "#replace-base")

	test.Click(t, page, "#replace-detached")
	test.TestReport(t, page, "ok replace")
	test.TestMust(t, page, "#replace-detached")

	test.Click(t, page, "#reload-after-replace")
	test.TestReport(t, page, "channel closed")

	test.Click(t, page, "#update-after-replace")
	test.TestReport(t, page, "channel closed")
	test.TestMustNot(t, page, "#replace-updated")

	test.Click(t, page, "#remount-after-replace")
	test.TestMust(t, page, "#replace-updated")
}

func TestDoorDetachedUnmountRebaseTransitions(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentDetachedRebase{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestMust(t, page, "#rebase-base")

	test.Click(t, page, "#unmount-detached")
	test.TestReport(t, page, "ok unmount")
	test.TestMustNot(t, page, "#rebase-base")

	test.Click(t, page, "#reload-after-unmount")
	test.TestReport(t, page, "channel closed")

	test.Click(t, page, "#rebase-after-unmount")
	test.TestReport(t, page, "channel closed")
	test.TestMustNot(t, page, "#rebased-detached-root")

	test.Click(t, page, "#remount-after-rebase")
	test.TestMust(t, page, "#rebased-detached-root")
	test.TestMust(t, page, "#rebased-detached")
}

func TestDoorProxyMoveBetweenParents(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentProxyMove{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestMust(t, page, "#frame1 #proxy-base")
	test.TestMustNot(t, page, "#frame2 #proxy-base")

	test.Click(t, page, "#rebase-proxy-move")
	test.TestReport(t, page, "ok rebase")
	test.TestMust(t, page, "#frame1 #proxy-moved-root")
	test.TestMustNot(t, page, "#frame2 #proxy-moved-root")

	test.Click(t, page, "#move-proxy")
	test.TestMustNot(t, page, "#frame1 #proxy-moved-root")
	test.TestMust(t, page, "#frame2 #proxy-moved-root")
	test.TestMust(t, page, "#frame2 #proxy-moved")
}

func TestDoorHierarchyCascade(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentHierarchy{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestMust(t, page, "#host1 #child-body")
	test.TestMust(t, page, "#host1 #grand-init")
	test.TestMustNot(t, page, "#host2 #child-body")

	test.Click(t, page, "#move-child")
	test.TestMustNot(t, page, "#host1 #child-body")
	test.TestMust(t, page, "#host2 #child-body")
	test.TestMust(t, page, "#host2 #grand-init")

	test.Click(t, page, "#grand-update")
	test.TestReport(t, page, "ok grand")
	test.TestMust(t, page, "#host2 #grand-updated")

	test.Click(t, page, "#remove-host2")
	test.TestMustNot(t, page, "#host2")
	test.TestMustNot(t, page, "#grand-updated")

	test.Click(t, page, "#grand-update")
	test.TestReport(t, page, "channel closed")
}

func TestDoorUpdateErrorTransition(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentErrorTransitions{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestMust(t, page, "#error-base")
	test.Click(t, page, "#update-error")
	test.TestReport(t, page, "channel err: update boom")
}

func TestDoorReplaceErrorTransition(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentErrorTransitions{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestMust(t, page, "#error-base")
	test.Click(t, page, "#replace-error")
	test.TestReport(t, page, "channel err: replace boom")
}

func TestDoorRebaseErrorTransition(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &FragmentErrorTransitions{}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()

	test.TestMust(t, page, "#error-base")
	test.Click(t, page, "#rebase-error")
	test.TestReport(t, page, "channel err: rebase boom")
}
