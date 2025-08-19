package router

import (
	"strings"
	"testing"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func TestPageStatic(t *testing.T) {
	bro := test.NewBro(browser, doors.ServePage(func(p doors.PageRouter[PathA], r doors.RPage[PathA]) doors.PageRoute {
		return p.StaticPage(static("a"))
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "a")
}

func TestPageStaticCode(t *testing.T) {
	bro := test.NewBro(browser, doors.ServePage(func(p doors.PageRouter[PathA], r doors.RPage[PathA]) doors.PageRoute {
		return p.StaticPageStatus(static("a"), 404)
	}))
	defer bro.Close()
	page := bro.PageStatus(t, "/a", 404)
	defer page.Close()
	test.TestContent(t, page, "#path", "a")
}

func TestPageRedirect(t *testing.T) {
	bro := test.NewBro(browser, doors.ServePage(func(p doors.PageRouter[PathA], r doors.RPage[PathA]) doors.PageRoute {
		return p.Redirect(PathB{})
	}), doors.ServePage(func(p doors.PageRouter[PathB], r doors.RPage[PathB]) doors.PageRoute {
		return p.StaticPage(static("b"))
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "b")
}

func TestPageReroute(t *testing.T) {
	bro := test.NewBro(browser, doors.ServePage(func(p doors.PageRouter[PathA], r doors.RPage[PathA]) doors.PageRoute {
		return p.Reroute(PathC{PathC1: true}, false)
	}), doors.ServePage(func(p doors.PageRouter[PathC], r doors.RPage[PathC]) doors.PageRoute {
		return p.PageFunc(pageC)
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "c1")
	testPath(t, page, "c1")
}

func TestPageRerouteDetached(t *testing.T) {
	bro := test.NewBro(browser, doors.ServePage(func(p doors.PageRouter[PathA], r doors.RPage[PathA]) doors.PageRoute {
		return p.Reroute(PathC{PathC1: true}, true)
	}), doors.ServePage(func(p doors.PageRouter[PathC], r doors.RPage[PathC]) doors.PageRoute {
		return p.PageFunc(pageC)
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "c1")
	testPath(t, page, "a")
	test.Click(t, page, "#c2")
	test.TestContent(t, page, "#path", "c2")
	testPath(t, page, "a")
}

func testPath(t *testing.T, page *rod.Page, path string) {
	url := strings.Split(strings.Trim(page.MustInfo().URL, "/"), "/")
	last := url[len(url)-1]
	if last != path {
		t.Fatal("path expected " + path + " actual " + last)
	}
}
