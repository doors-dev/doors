package imports

import (
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
	"github.com/go-rod/rod"
)

func checkColor(t *testing.T, page *rod.Page) {
	element := page.MustElement("h1")
	styleValue, err := element.Eval(`() => getComputedStyle(this).color`)
	if err != nil {
		t.Fatalf("")
	}
	if styleValue.Value.Str() != "rgb(255, 0, 0)" {
		t.Fatal("h1 expected color", " red ", "got ", styleValue.Value.Str())
	}
}

func testStyle(t *testing.T, h func(doors.Source[test.Path]) gox.Elem) {
	bro := test.NewBro(browser,
		func(r doors.Router) {
			doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
				return doors.ResponseComp(&test.Page{
					Source: r,
					Header: "Testing Imports",
					H:      h,
					F:      &ModuleFragment{},
				})
			})
			doors.UseRoute(r, doors.RouteDir{Prefix: "module", DirPath: modulePath})
		},
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	<-time.After(100 * time.Millisecond)
	checkColor(t, page)
}
func testModule(t *testing.T, h func(doors.Source[test.Path]) gox.Elem) {
	page := modulePage(t, h)
	defer page.Close()

	<-time.After(100 * time.Millisecond)
	test.TestReport(t, page, "hello")
}

func modulePage(t *testing.T, h func(doors.Source[test.Path]) gox.Elem) *rod.Page {
	t.Helper()
	bro := test.NewBro(browser,
		func(r doors.Router) {
			doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
				return doors.ResponseComp(&test.Page{
					Source: r,
					Header: "Testing Imports",
					H:      h,
					F:      &ModuleFragment{},
				})
			})
			doors.UseRoute(r, doors.RouteDir{Prefix: "module", DirPath: modulePath})
		},
	)
	t.Cleanup(func() {
		bro.Close()
	})

	<-time.After(100 * time.Millisecond)
	page := bro.Page(t, "/")
	t.Cleanup(func() {
		page.Close()
	})
	return page
}

func testValue(t *testing.T, h func(doors.Source[test.Path]) gox.Elem) {
	bro := test.NewBro(browser,
		func(r doors.Router) {
			doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
				return doors.ResponseComp(&test.Page{
					Source: r,
					Header: "Testing Imports",
					H:      h,
					F:      &ValueFragment{},
				})
			})
			doors.UseRoute(r, doors.RouteDir{Prefix: "module", DirPath: modulePath})
		},
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	<-time.After(100 * time.Millisecond)
	test.TestReport(t, page, "hello")
}

func emptyPage(t *testing.T, h func(doors.Source[test.Path]) gox.Elem) *rod.Page {
	t.Helper()
	bro := test.NewBro(browser,
		func(r doors.Router) {
			doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
				return doors.ResponseComp(&test.Page{
					Source: r,
					Header: "Testing Imports",
					H:      h,
					F:      &Empty{},
				})
			})
			doors.UseRoute(r, doors.RouteDir{Prefix: "module", DirPath: modulePath})
		},
	)
	t.Cleanup(func() {
		bro.Close()
	})
	page := bro.Page(t, "/")
	t.Cleanup(func() {
		page.Close()
	})
	return page
}

func getAttr(t *testing.T, page *rod.Page, selector string, name string) string {
	t.Helper()
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("attr: element ", selector, " not found")
	}
	attr, err := el.Attribute(name)
	if err != nil {
		t.Fatal("attr: element ", selector, " attribute ", name, " not found")
	}
	if attr == nil {
		t.Fatal("attr: element ", selector, " attribute ", name, " is nil")
	}
	return *attr
}

func testStyleAttr(t *testing.T, h func(doors.Source[test.Path]) gox.Elem, check func(t *testing.T, href string)) {
	t.Helper()
	bro := test.NewBro(browser,
		func(r doors.Router) {
			doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
				return doors.ResponseComp(&test.Page{
					Source: r,
					Header: "Testing Imports",
					H:      h,
					F:      &ModuleFragment{},
				})
			})
			doors.UseRoute(r, doors.RouteDir{Prefix: "module", DirPath: modulePath})
		},
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	<-time.After(100 * time.Millisecond)
	checkColor(t, page)
	check(t, getAttr(t, page, `head link[rel="stylesheet"]`, "href"))
}

func TestModule(t *testing.T) {
	testModule(t, moduleHead)
}
func TestModuleVisible(t *testing.T) {
	page := modulePage(t, moduleVisibleHead)
	<-time.After(100 * time.Millisecond)
	test.TestReport(t, page, "hello")
	test.TestAttr(t, page, "#module-tag", "type", "module")
}

func TestModuleBytes(t *testing.T) {
	testModule(t, moduleBytesHead)
}
func TestModuleString(t *testing.T) {
	testModule(t, moduleStringHead)
}
func TestModuleRaw(t *testing.T) {
	testModule(t, moduleRawHead)
}
func TestModuleRawBytes(t *testing.T) {
	testModule(t, moduleRawBytesHead)
}
func TestModuleRawBytesShort(t *testing.T) {
	testModule(t, moduleRawBytesShortHead)
}
func TestModuleRawBytesModify(t *testing.T) {
	testModule(t, moduleRawBytesModifyHead)
}
func TestModulePreloadBytes(t *testing.T) {
	page := emptyPage(t, modulePreloadBytesHead)
	test.TestAttr(t, page, `head link[rel="modulepreload"]`, "rel", "modulepreload")
}
func TestModulePreloadNamed(t *testing.T) {
	page := emptyPage(t, modulePreloadNamedHead)
	href := getAttr(t, page, `head link[rel="modulepreload"]`, "href")
	if !strings.Contains(href, ".module-preload.js") {
		t.Fatal("expected modulepreload href to contain .module-preload.js, got ", href)
	}
}
func TestModuleFS(t *testing.T) {
	testModule(t, moduleBundleFSHead)
}

func TestModulHosted(t *testing.T) {
	testModule(t, moduleBundleHostHead)
}

func TestModuleExternal(t *testing.T) {
	testModule(t, moduleExternalHead)
}
func TestModuleProxy(t *testing.T) {
	testModule(t, moduleProxyHead)
}
func TestScriptInline(t *testing.T) {
	testValue(t, scriptInlineHead)
}
func TestScriptString(t *testing.T) {
	testValue(t, scriptStringHead)
}
func TestStyleHosted(t *testing.T) {
	testStyle(t, styleHostedHead)
}
func TestStyleExternal(t *testing.T) {
	testStyle(t, styleExternalHead)
}

func TestStyleBytes(t *testing.T) {
	testStyle(t, styleBytesHead)
}
func TestStyleInline(t *testing.T) {
	testStyle(t, styleInlineHead)
}
func TestStyleBytesShort(t *testing.T) {
	testStyle(t, styleBytesShortHead)
}
func TestStyleBytesModify(t *testing.T) {
	testStyle(t, styleBytesModifyHead)
}
func TestStyleString(t *testing.T) {
	testStyle(t, styleStringHead)
}
func TestStyleProxy(t *testing.T) {
	testStyle(t, styleProxyHead)
}
func TestStyle(t *testing.T) {
	testStyle(t, styleHead)
}
func TestStyleFS(t *testing.T) {
	testStyle(t, styleFSHead)
}
func TestStyleNamed(t *testing.T) {
	testStyleAttr(t, styleNamedHead, func(t *testing.T, href string) {
		if !strings.Contains(href, ".named.css") {
			t.Fatal("expected hosted stylesheet path to contain .named.css, got ", href)
		}
	})
}

func TestModuleNamed(t *testing.T) {
	page := modulePage(t, moduleHead)
	<-time.After(100 * time.Millisecond)
	test.TestReport(t, page, "hello")
	test.TestMustNot(t, page, `script[src*="module.js"]`)
}
func TestStylePrivateNamed(t *testing.T) {
	testStyleAttr(t, stylePrivateNamedHead, func(t *testing.T, href string) {
		if !strings.Contains(href, "/h/") {
			t.Fatal("expected private stylesheet path to use hook route, got ", href)
		}
		if !strings.HasSuffix(href, "/private.css") {
			t.Fatal("expected private stylesheet path to end with /private.css, got ", href)
		}
	})
}

func TestReact(t *testing.T) {
	bro := test.NewBro(browser,
		func(r doors.Router) {
			doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
				return doors.ResponseComp(&test.Page{
					Source: r,
					Header: "Testing Imports",
					H:      reactHead,
					F:      &ReactFragment{},
				})
			})
		},
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	<-time.After(100 * time.Millisecond)
	test.TestContent(t, page, "#h2", "React")
	test.TestContent(t, page, "#ph2", "Preact")
	test.Click(t, page, "#inc")
	test.Click(t, page, "#inc")
	test.Click(t, page, "#dec")
	test.TestReportId(t, page, 0, "1")
	test.Click(t, page, "#pinc")
	test.Click(t, page, "#pinc")
	test.Click(t, page, "#pdec")
	test.Click(t, page, "#pdec")
	test.Click(t, page, "#pdec")
	test.TestReportId(t, page, 1, "-1")
}

func TestFiles(t *testing.T) {
	bro := test.NewBro(browser,
		func(r doors.Router) {
			doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
				return doors.ResponseComp(&test.Page{
					Source: r,
					Header: "Testing Imports",
					H:      staticFiles,
					F:      &Empty{},
				})
			})
		},
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	href := getAttr(t, page, "#cached-href", "href")
	if !strings.Contains(href, "/r/") {
		t.Fatal("expected cached href to use resource route, got ", href)
	}
	if !strings.Contains(href, ".hello.txt") {
		t.Fatal("expected cached href to contain .hello.txt, got ", href)
	}
	href = getAttr(t, page, "#cached-href-modify", "href")
	if !strings.Contains(href, "/r/") {
		t.Fatal("expected cached href modify to use resource route, got ", href)
	}
	if !strings.Contains(href, ".hello-modify.txt") {
		t.Fatal("expected cached href modify to contain .hello-modify.txt, got ", href)
	}
}

func TestFileCachedBad(t *testing.T) {
	bro := test.NewBro(browser,
		func(r doors.Router) {
			doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
				return doors.ResponseComp(&test.Page{
					Source: r,
					Header: "Testing Imports",
					H:      fileCachedHrefBad,
					F:      &Empty{},
				})
			})
		},
	)
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	test.TestMust(t, page, `[data-fw="error"]`)
}
