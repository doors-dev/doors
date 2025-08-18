package door

import (
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"testing"
	"time"
)

func TestDoorLoadPage(t *testing.T) {
	bro := test.NewBro(browser,
		doors.ServePage(func(pr doors.PageRouter[test.Path], r doors.RPage[test.Path]) doors.PageRoute {
			return pr.Page(&test.Page{
				Header: "Page Door",
			})
		}),
	)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	h1Text := page.MustElement("h1").MustText()
	if h1Text != "Page Door" {
		t.Fatal("header missmatch")
	}
}

func TestDoorInitialContent(t *testing.T) {
	doorInitialContent(t,
		func() test.Fragment {
			return &BeforeFragment{}
		},
	)
}

func TestDoorInitialContentTag(t *testing.T) {
	doorInitialContent(t,
		func() test.Fragment {
			return &BeforeFragment{
				doorInit: doors.Door{
					Tag: "div",
				},
				doorUpdate: doors.Door{
					Tag: "div",
				},
				doorRemoved: doors.Door{
					Tag: "div",
				},
				doorReplaced: doors.Door{
					Tag: "div",
				},
			}
		},
	)
}

func doorInitialContent(t *testing.T, init func() test.Fragment) {
	bro := test.NewFragmentBro(browser, init)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
}

func TestDoorUpdatedBefore(t *testing.T) {
	doorUpdatedBefore(t,
		func() test.Fragment {
			return &BeforeFragment{}
		},
	)
}
func TestDoorUpdatedBeforeTag(t *testing.T) {
	doorUpdatedBefore(t,
		func() test.Fragment {
			return &BeforeFragment{
				doorInit: doors.Door{
					Tag: "div",
				},
				doorUpdate: doors.Door{
					Tag: "div",
				},
				doorRemoved: doors.Door{
					Tag: "div",
				},
				doorReplaced: doors.Door{
					Tag: "div",
				},
			}
		},
	)
}

func doorUpdatedBefore(t *testing.T, init func() test.Fragment) {
	bro := test.NewFragmentBro(browser, init)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#updated")
}

func TestDoorRemovedBefore(t *testing.T) {
	doorRemovedBefore(t,
		func() test.Fragment {
			return &BeforeFragment{}
		},
	)
}
func TestDoorRemovedBeforeTag(t *testing.T) {
	doorRemovedBefore(t,
		func() test.Fragment {
			return &BeforeFragment{
				doorInit: doors.Door{
					Tag: "div",
				},
				doorUpdate: doors.Door{
					Tag: "div",
				},
				doorRemoved: doors.Door{
					Tag: "div",
				},
				doorReplaced: doors.Door{
					Tag: "div",
				},
			}
		},
	)
}

func doorRemovedBefore(t *testing.T, init func() test.Fragment) {
	bro := test.NewFragmentBro(browser, init)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMustNot(t, page, "#removed")
}

func TestDoorReplacedBefore(t *testing.T) {
	doorReplacedBefore(t,
		func() test.Fragment {
			return &BeforeFragment{}
		},
	)
}
func TestDoorReplacedBeforeTag(t *testing.T) {
	doorReplacedBefore(t,
		func() test.Fragment {
			return &BeforeFragment{
				doorInit: doors.Door{
					Tag: "div",
				},
				doorUpdate: doors.Door{
					Tag: "div",
				},
				doorRemoved: doors.Door{
					Tag: "div",
				},
				doorReplaced: doors.Door{
					Tag: "div",
				},
			}
		},
	)
}

func doorReplacedBefore(t *testing.T, init func() test.Fragment) {
	bro := test.NewFragmentBro(browser, init)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#initReplaced")

	// test.TestMust(t, page, "body > #replaced")

}

func TestDoorDynamic(t *testing.T) {
	doorDynamic(t,
		func() test.Fragment {
			return &DynamicFragment{}
		},
	)
}

func TestDoorDynamicTag(t *testing.T) {
	doorDynamic(t,
		func() test.Fragment {
			return &DynamicFragment{
				n1: doors.Door{
					Tag: "div",
				},
				n2: doors.Door{
					Tag: "div",
				},
			}
		},
	)
}

func doorDynamic(t *testing.T, init func() test.Fragment) {
	bro := test.NewFragmentBro(browser, init)
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
	doorEmbedded(t,
		func() test.Fragment {
			return &EmbeddedFragment{}
		},
	)
}

func TestDoorEmbeddedTag(t *testing.T) {
	doorEmbedded(t,
		func() test.Fragment {
			return &EmbeddedFragment{
				n1: doors.Door{
					Tag: "div",
				},
				n2: doors.Door{
					Tag: "div",
				},
				n3: doors.Door{
					Tag: "div",
				},
			}
		},
	)
}
func doorEmbedded(t *testing.T, f func() test.Fragment) {
	bro := test.NewFragmentBro(browser, f)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
	test.Click(t, page, "#remove")
	test.TestMustNot(t, page, "#init")
	test.TestMust(t, page, "#static")
	test.Click(t, page, "#clear")
	test.TestMustNot(t, page, "#static")
	test.Click(t, page, "#replace")
	test.TestMustNot(t, page, "#temp")
	test.TestMust(t, page, "#replaced")
}

func TestEmbeddedRemove(t *testing.T) {
	doorEmbeddedRemove(t,
		func() test.Fragment {
			return &EmbeddedFragment{}
		},
	)
}
func TestEmbeddedRemoveTag(t *testing.T) {
	doorEmbeddedRemove(t,
		func() test.Fragment {
			return &EmbeddedFragment{
				n1: doors.Door{
					Tag: "div",
				},
				n2: doors.Door{
					Tag: "div",
				},
				n3: doors.Door{
					Tag: "div",
				},
			}
		},
	)
}
func doorEmbeddedRemove(t *testing.T, init func() test.Fragment) {
	bro := test.NewFragmentBro(browser, init)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	<-time.After(0 * time.Hour)
	test.TestMust(t, page, "#init")
	test.Click(t, page, "#clear")
	test.TestMustNot(t, page, "#init")
	test.Click(t, page, "#remove")
	test.Click(t, page, "#replace")
	test.TestMustNot(t, page, "#temp")
	test.TestMust(t, page, "#replaced")
}

func TestDoorUpdateX(t *testing.T) {
	doorUpdateX(t,
		func() test.Fragment {
			return &FragmentX{}
		},
	)
}

func doorUpdateX(t *testing.T, init func() test.Fragment) {
	bro := test.NewFragmentBro(browser, init)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	test.TestMust(t, page, "#init")
	test.Click(t, page, "#updatex")
	test.TestReport(t, page, "ok")
	test.TestMustNot(t, page, "#init")
	test.TestMust(t, page, "#updated")
	test.Click(t, page, "#removex")
	test.TestReport(t, page, "ok")
	test.TestMustNot(t, page, "#updated")
	test.Click(t, page, "#updatex")
	test.TestReport(t, page, "channel closed")
}

func TestDoorMultiple(t *testing.T) {
	doorMultiple(t, func() test.Fragment {
		return &FragmentMany{}
	})
}
func TestDoorMultipleTag(t *testing.T) {
	doorMultiple(t, func() test.Fragment {
		return &FragmentMany{
			n: doors.Door{
				Tag: "div",
			},
		}
	})
}
func doorMultiple(t *testing.T, init func() test.Fragment) {
	bro := test.NewFragmentBro(browser, init)
	page := bro.Page(t, "/")
	defer bro.Close()
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
