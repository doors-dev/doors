package components

import (
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/test"
	"github.com/go-rod/rod"
)

func TestTitle(t *testing.T) {
	p := common.RandId()
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
			return doors.ResponseComp(&test.Page{
				Source: r,
				F: &LinksFragment{
					Param: p,
				},
			})
		})
	})
	page := bro.Page(t, "/")
	<-time.After(200 * time.Millisecond)

	// Test initial state (home) - has description, keywords, og:title
	test.TestContent(t, page, "title", "home")
	test.TestAttr(t, page, `meta[name="description"]`, "content", "Welcome to the home page")
	test.TestAttr(t, page, `meta[name="keywords"]`, "content", "home, main, index")
	test.TestAttr(t, page, `meta[property="og:title"]`, "content", "Home Page")
	test.TestAttr(t, page, "#active-default", "aria-current", "page")

	// Test param state - has description, keywords, og:title, author
	test.Click(t, page, "#param")
	test.TestContent(t, page, "title", p)
	test.TestAttr(t, page, `meta[name="description"]`, "content", "Page for parameter: "+p)
	test.TestAttr(t, page, `meta[name="keywords"]`, "content", "param, "+p)
	test.TestAttr(t, page, `meta[property="og:title"]`, "content", "Param: "+p)
	test.TestAttr(t, page, `meta[name="author"]`, "content", "Parameter Author")
	test.TestAttr(t, page, "#active-starts", "data-active", "starts")
	test.TestAttr(t, page, "#active-segments", "data-active", "segments")

	// Test string state - updates title and emits string-specific meta
	test.Click(t, page, "#string")
	test.TestContent(t, page, "title", "s")
	test.TestAttr(t, page, `meta[name="description"]`, "content", "String page description")
	test.TestAttr(t, page, `meta[name="category"]`, "content", "text-content")

	// Go back to home state - verify restoration of the active title/meta values
	test.Click(t, page, "#home")
	<-time.After(200 * time.Millisecond)
	test.TestContent(t, page, "title", "home")
	test.TestAttr(t, page, `meta[name="description"]`, "content", "Welcome to the home page")
	test.TestAttr(t, page, `meta[name="keywords"]`, "content", "home, main, index")
	test.TestAttr(t, page, `meta[property="og:title"]`, "content", "Home Page")
	test.TestAttr(t, page, "#active-default", "aria-current", "page")
}

func TestActionHelpers(t *testing.T) {
	p := common.RandId()
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
			return doors.ResponseComp(&test.Page{
				Source: r,
				F: &LinksFragment{
					Param: p,
				},
			})
		})
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	<-time.After(200 * time.Millisecond)

	test.TestAttrNo(t, page, "#action-target", "data-indicated")
	test.ClickNow(t, page, "#action-indicate")
	<-time.After(50 * time.Millisecond)
	test.TestAttr(t, page, "#action-target", "data-indicated", "true")
	<-time.After(250 * time.Millisecond)
	test.TestAttrNo(t, page, "#action-target", "data-indicated")

	if page.MustEval("() => window.scrollY").Num() != 0 {
		t.Fatal("expected page to start at scrollY=0")
	}
	test.ClickNow(t, page, "#action-scroll")
	<-time.After(200 * time.Millisecond)
	if page.MustEval("() => window.scrollY").Num() <= 0 {
		t.Fatal("expected ActionOnlyScroll to move the page")
	}

	test.ClickNow(t, page, "#raw-assign")
	<-time.After(300 * time.Millisecond)
	test.TestContent(t, page, "title", "s")
}

func proxyPage(t *testing.T) *rod.Page {
	t.Helper()
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
			return doors.ResponseComp(&test.Page{
				Source: r,
				F: &ProxyFragment{
					r: test.NewReporter(1),
				},
			})
		})
	})
	t.Cleanup(func() {
		bro.Close()
	})

	page := bro.Page(t, "/")
	t.Cleanup(func() {
		page.Close()
	})
	return page
}

func TestProxyAttrModifierLiteral(t *testing.T) {
	page := proxyPage(t)
	test.Click(t, page, "#proxy-literal")
	test.TestReport(t, page, "literal")
}

func TestProxyAttrModifierContainerPenetration(t *testing.T) {
	page := proxyPage(t)
	test.Click(t, page, "#proxy-container")
	test.TestReport(t, page, "container")
}
