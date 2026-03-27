package imports

import (
	"io"
	"net/http"
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

func testValueAttr(t *testing.T, h func(doors.Source[test.Path]) gox.Elem, selector string, name string, check func(t *testing.T, value string)) {
	t.Helper()
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
	check(t, getAttr(t, page, selector, name))
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

func fetchText(t *testing.T, url string) string {
	t.Helper()
	if strings.HasPrefix(url, "/") {
		url = test.Host + url
	}
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d for %s", resp.StatusCode, url)
	}
	return string(body)
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
func TestScriptInlineNamedExt(t *testing.T) {
	testValueAttr(t, scriptInlineNamedExtHead, "#script-inline-ext", "src", func(t *testing.T, src string) {
		if !strings.Contains(src, "/r/") && !strings.Contains(src, "/h/") {
			t.Fatal("expected inline script to be served through resource or hook path, got ", src)
		}
		if !strings.HasSuffix(src, ".inline-script.js") {
			t.Fatal("expected named inline script path to end with .inline-script.js, got ", src)
		}
		if strings.HasSuffix(src, ".inline-script.js.js") {
			t.Fatal("expected named inline script path to avoid duplicate extension, got ", src)
		}
	})
}
func TestScriptInlineBytes(t *testing.T) {
	testValue(t, scriptInlineBytesHead)
}
func TestScriptString(t *testing.T) {
	testValue(t, scriptStringHead)
}
func TestScriptPrivate(t *testing.T) {
	testValueAttr(t, scriptPrivateHead, "#script-private", "src", func(t *testing.T, src string) {
		if !strings.Contains(src, "/h/") {
			t.Fatal("expected private script path to use hook route, got ", src)
		}
		if !strings.HasSuffix(src, "/private-script.js") {
			t.Fatal("expected private script path to end with /private-script.js, got ", src)
		}
	})
}
func TestScriptNoCache(t *testing.T) {
	testValueAttr(t, scriptNoCacheHead, "#script-nocache", "src", func(t *testing.T, src string) {
		if !strings.Contains(src, "/h/") {
			t.Fatal("expected nocache script path to use hook route, got ", src)
		}
		if !strings.HasSuffix(src, "/nocache-script.js") {
			t.Fatal("expected nocache script path to end with /nocache-script.js, got ", src)
		}
	})
}
func TestStyleHosted(t *testing.T) {
	testStyle(t, styleHostedHead)
}
func TestStyleHostedRaw(t *testing.T) {
	testStyleAttr(t, styleHostedRawHead, func(t *testing.T, href string) {
		if !strings.Contains(href, "/module/style.css") {
			t.Fatal("expected raw hosted stylesheet path to keep /module/style.css, got ", href)
		}
		if strings.Contains(href, "/h/") || strings.Contains(href, "/r/") {
			t.Fatal("expected raw hosted stylesheet path to stay direct, got ", href)
		}
	})
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
func TestStyleRaw(t *testing.T) {
	page := emptyPage(t, styleRawHead)
	<-time.After(100 * time.Millisecond)
	checkColor(t, page)
	test.TestMust(t, page, "head style")
}
func TestStyleMinify(t *testing.T) {
	testStyleAttr(t, styleMinifyHead, func(t *testing.T, href string) {
		css := fetchText(t, href)
		if strings.Contains(css, "\n") {
			t.Fatal("expected minified stylesheet to avoid newlines, got ", css)
		}
		if !strings.Contains(css, "h1{color:red}") && !strings.Contains(css, "h1{color:red;}") {
			t.Fatal("expected minified stylesheet content, got ", css)
		}
	})
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
func TestStylePrivate(t *testing.T) {
	testStyleAttr(t, stylePrivateHead, func(t *testing.T, href string) {
		if !strings.Contains(href, "/h/") {
			t.Fatal("expected private inline stylesheet path to use hook route, got ", href)
		}
		if !strings.HasSuffix(href, "/private-inline.css") {
			t.Fatal("expected private inline stylesheet path to end with /private-inline.css, got ", href)
		}
	})
}
func TestStylePrivateNamedExt(t *testing.T) {
	testStyleAttr(t, stylePrivateNamedExtHead, func(t *testing.T, href string) {
		if !strings.Contains(href, "/h/") {
			t.Fatal("expected private inline stylesheet path to use hook route, got ", href)
		}
		if !strings.HasSuffix(href, "/private-inline.css") {
			t.Fatal("expected private inline stylesheet path to end with /private-inline.css, got ", href)
		}
		if strings.HasSuffix(href, "/private-inline.css.css") {
			t.Fatal("expected private inline stylesheet path to avoid duplicate extension, got ", href)
		}
	})
}
func TestStyleNoCache(t *testing.T) {
	testStyleAttr(t, styleNoCacheHead, func(t *testing.T, href string) {
		if !strings.Contains(href, "/h/") {
			t.Fatal("expected nocache inline stylesheet path to use hook route, got ", href)
		}
		if !strings.HasSuffix(href, "/nocache-inline.css") {
			t.Fatal("expected nocache inline stylesheet path to end with /nocache-inline.css, got ", href)
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
	href = getAttr(t, page, "#private-href", "href")
	if !strings.Contains(href, "/h/") {
		t.Fatal("expected private href to use hook route, got ", href)
	}
	if !strings.Contains(href, "/private.txt") {
		t.Fatal("expected private href to contain /private.txt, got ", href)
	}
	href = getAttr(t, page, "#private-href-modify", "href")
	if !strings.Contains(href, "/h/") {
		t.Fatal("expected private href modify to use hook route, got ", href)
	}
	if !strings.Contains(href, "/private-modify.txt") {
		t.Fatal("expected private href modify to contain /private-modify.txt, got ", href)
	}
	src := getAttr(t, page, "#private-frame", "src")
	if !strings.Contains(src, "/h/") {
		t.Fatal("expected private frame src to use hook route, got ", src)
	}
	if !strings.Contains(src, "/frame.html") {
		t.Fatal("expected private frame src to contain /frame.html, got ", src)
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
