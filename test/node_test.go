package test

import (
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/go-rod/rod"
)

func testMust(t *testing.T, page *rod.Page, selector string) {
	_, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("must: element ", selector, " not found")
	}
}
func testMustNot(t *testing.T, page *rod.Page, selector string) {
	_, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err == nil {
		t.Fatal("must not: element ", selector, " found")
	}
}

func click(t *testing.T, page *rod.Page, selector string) {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("click: element ", selector, " not found")
	}
	el.MustClick()
	<-time.After(100 * time.Millisecond)
}

func TestNodeLoadPage(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[PathNode], r doors.RPage[PathNode]) doors.PageRoute {
			return pr.Page(&PageNode{})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	h1Text := page.MustElement("h1").MustText()
	if h1Text != "Page Node" {
		t.Fatal("header missmatch")
	}
}

func TestNodeInitialContent(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[PathNode], r doors.RPage[PathNode]) doors.PageRoute {
			return pr.Page(&PageNode{
				f: &BeforeFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMust(t, page, "#init")
}

func TestNodeUpdatedBefore(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[PathNode], r doors.RPage[PathNode]) doors.PageRoute {
			return pr.Page(&PageNode{
				f: &BeforeFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMust(t, page, "#updated")
}
func TestNodeRemovedBefore(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[PathNode], r doors.RPage[PathNode]) doors.PageRoute {
			return pr.Page(&PageNode{
				f: &BeforeFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMustNot(t, page, "#removed")
}
func TestNodeReplacedBefore(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[PathNode], r doors.RPage[PathNode]) doors.PageRoute {
			return pr.Page(&PageNode{
				f: &BeforeFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMustNot(t, page, "#initReplaced")
	testMust(t, page, "body > #replaced")

}
func TestNodeDynamic(t *testing.T) {
	bro := newBro(
		doors.ServePage(func(pr doors.PageRouter[PathNode], r doors.RPage[PathNode]) doors.PageRoute {
			return pr.Page(&PageNode{
				f: &DynamicFragment{},
			})
		}),
	)
	page := bro.page(t, "/")
	defer bro.close()
	defer page.Close()
	testMust(t, page, "#init")
	click(t, page, "#update")
	testMustNot(t, page, "#init")
	testMust(t, page, "#updated")
	click(t, page, "#replace")
	testMustNot(t, page, "#updated")
	testMust(t, page, "#replaced")
	click(t, page, "#remove")
	testMustNot(t, page, "#replaced")
}
