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
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"testing"
	"time"
)

func TestDoorLoadPage(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
			return doors.ResponseComp(&test.Page{
				Source: r,
				Header: "Page Door",
			})
		})
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
	test.Click(t, page, "#updateContent")
	test.TestMustNot(t, page, "#new")
	test.TestMustNot(t, page, "#presist")
	test.TestMust(t, page, "#presist2")
	test.TestMust(t, page, "#new2")

	test.Click(t, page, "#updateEditor")
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
	test.TestReport(t, page, "channel err: replaced door can't be reloaded")

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
	test.TestReport(t, page, "channel err: unmounted door can't be reloaded")

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
