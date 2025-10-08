package router

import (
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func testPath(t *testing.T, page *rod.Page, path string) {
	url := strings.Split(strings.Trim(page.MustInfo().URL, "/"), "/")
	last := url[len(url)-1]
	if last != path {
		t.Fatal("path expected " + path + " actual " + last)
	}
}

func TestPageStatic(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.StaticPage(static("a"), 0)
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "a")
}

func TestPageStaticCode(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.StaticPage(static("a"), 404)
	}))
	defer bro.Close()
	page := bro.PageStatus(t, "/a", 404)
	defer page.Close()
	test.TestContent(t, page, "#path", "a")
}

func TestPageRedirect(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.Redirect(PathB{}, 0)
	}), doors.UsePage(func(p doors.PageRouter[PathB], r doors.RModel[PathB]) doors.ModelRoute {
		return p.StaticPage(static("b"), 0)
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "b")
}

func TestPageReroute(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.Reroute(PathC{PathC1: true}, false)
	}), doors.UsePage(func(p doors.PageRouter[PathC], r doors.RModel[PathC]) doors.ModelRoute {
		return p.PageFunc(pageC)
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "c1")
	testPath(t, page, "c1")
}

func TestPageRerouteDetached(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.Reroute(PathC{PathC1: true}, true)
	}), doors.UsePage(func(p doors.PageRouter[PathC], r doors.RModel[PathC]) doors.ModelRoute {
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

func TestPageError(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.Redirect(PathC{}, 0)
	}), doors.UsePage(func(p doors.PageRouter[PathC], r doors.RModel[PathC]) doors.ModelRoute {
		return p.PageFunc(pageC)
	}), doors.UseErrorPage(func(message string) templ.Component {
		return static("error")
	}))
	defer bro.Close()
	page := bro.PageStatus(t, "/a", 500)
	defer page.Close()
	test.TestContent(t, page, "#path", "error")
}

func TestPageInfiniteReroute(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.Reroute(PathC{}, true)
	}), doors.UsePage(func(p doors.PageRouter[PathC], r doors.RModel[PathC]) doors.ModelRoute {
		return p.Reroute(PathA{}, true)
	}), doors.UseErrorPage(func(message string) templ.Component {
		return static("error")
	}))
	defer bro.Close()
	page := bro.PageStatus(t, "/a", 500)
	defer page.Close()
	test.TestContent(t, page, "#path", "error")
}

func TestLocations(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.PageFunc(pageA)
	}), doors.UsePage(func(p doors.PageRouter[PathC], r doors.RModel[PathC]) doors.ModelRoute {
		return p.PageFunc(pageC)
	}), doors.UseErrorPage(func(message string) templ.Component {
		return static("error")
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	testPath(t, page, "a")
	defer page.Close()
	test.Click(t, page, "#assign")
	testPath(t, page, "c1")
	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "a")
	test.Click(t, page, "#assign")
	testPath(t, page, "c1")
	test.Click(t, page, "#replace")
	testPath(t, page, "c2")
	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "a")
	test.Click(t, page, "#assign")
	marker := test.GetContent(t, page, "#marker")
	test.Click(t, page, "#reload")
	marker2 := test.GetContent(t, page, "#marker")
	if marker == marker2 {
		t.Fatalf("reload did not work")
	}
	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "a")
}
func TestAfterAssign(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.PageFunc(pageA)
	}), doors.UsePage(func(p doors.PageRouter[PathC], r doors.RModel[PathC]) doors.ModelRoute {
		return p.PageFunc(pageC)
	}), doors.UseErrorPage(func(message string) templ.Component {
		return static("error")
	}), doors.UsePage(func(p doors.PageRouter[PathB], r doors.RModel[PathB]) doors.ModelRoute {
		return p.StaticPage(static("b"), 0)
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.Click(t, page, "#assign")
	test.Click(t, page, "#assign_after")
	testPath(t, page, "b")
	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "c1")
}

func TestAfterReplace(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.PageFunc(pageA)
	}), doors.UsePage(func(p doors.PageRouter[PathC], r doors.RModel[PathC]) doors.ModelRoute {
		return p.PageFunc(pageC)
	}), doors.UseErrorPage(func(message string) templ.Component {
		return static("error")
	}), doors.UsePage(func(p doors.PageRouter[PathB], r doors.RModel[PathB]) doors.ModelRoute {
		return p.StaticPage(static("b"), 0)
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.Click(t, page, "#assign")
	test.Click(t, page, "#replace_after")
	testPath(t, page, "b")
	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "a")
}
func TestAfterReload(t *testing.T) {
	bro := test.NewBro(browser, doors.UsePage(func(p doors.PageRouter[PathA], r doors.RModel[PathA]) doors.ModelRoute {
		return p.PageFunc(pageA)
	}), doors.UsePage(func(p doors.PageRouter[PathC], r doors.RModel[PathC]) doors.ModelRoute {
		return p.PageFunc(pageC)
	}), doors.UseErrorPage(func(message string) templ.Component {
		return static("error")
	}), doors.UsePage(func(p doors.PageRouter[PathB], r doors.RModel[PathB]) doors.ModelRoute {
		return p.StaticPage(static("b"), 0)
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.Click(t, page, "#assign")
	marker := test.GetContent(t, page, "#marker")
	test.Click(t, page, "#reload_after")
	marker2 := test.GetContent(t, page, "#marker")
	if marker == marker2 {
		t.Fatalf("reload did not work")
	}
	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "a")

}
