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
	page := valuePage(t, h)
	<-time.After(100 * time.Millisecond)
	test.TestReport(t, page, "hello")
}

func valuePage(t *testing.T, h func(doors.Source[test.Path]) gox.Elem) *rod.Page {
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
	t.Cleanup(func() {
		bro.Close()
	})
	page := bro.Page(t, "/")
	t.Cleanup(func() {
		page.Close()
	})
	return page
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

func getTextContent(t *testing.T, page *rod.Page, selector string) string {
	t.Helper()
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		t.Fatal("text: element ", selector, " not found")
	}
	value, err := el.Eval(`() => this.textContent`)
	if err != nil {
		t.Fatal(err)
	}
	return value.Value.Str()
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
	body, _ := fetchTextAndHeaders(t, url)
	return body
}

func fetchTextAndHeaders(t *testing.T, url string) (string, http.Header) {
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
	return string(body), resp.Header
}

func compactText(s string) string {
	return strings.NewReplacer(" ", "", "\n", "", "\t", "", "\r", "").Replace(s)
}

func fetchTextFromPage(t *testing.T, page *rod.Page, url string) string {
	t.Helper()
	value, err := page.Eval(`async url => {
		const res = await fetch(url)
		return {
			status: res.status,
			body: await res.text(),
		}
	}`, url)
	if err != nil {
		t.Fatal(err)
	}
	status := int(value.Value.Get("status").Int())
	if status != http.StatusOK {
		t.Fatalf("unexpected page fetch status %d for %s", status, url)
	}
	return value.Value.Get("body").Str()
}

func fetchTextAndContentTypeFromPage(t *testing.T, page *rod.Page, url string) (string, string) {
	t.Helper()
	value, err := page.Eval(`async url => {
		const res = await fetch(url)
		return {
			status: res.status,
			contentType: res.headers.get("content-type") || "",
			body: await res.text(),
		}
	}`, url)
	if err != nil {
		t.Fatal(err)
	}
	status := int(value.Value.Get("status").Int())
	if status != http.StatusOK {
		t.Fatalf("unexpected page fetch status %d for %s", status, url)
	}
	return value.Value.Get("body").Str(), value.Value.Get("contentType").Str()
}

func testScriptFetchExact(t *testing.T, h func(doors.Source[test.Path]) gox.Elem, selector string, expected string) {
	t.Helper()
	page := modulePage(t, h)
	src := getAttr(t, page, selector, "src")
	body := fetchText(t, src)
	if body != expected {
		t.Fatal("expected script source to stay raw, got ", body)
	}
	if strings.Contains(body, `_d0r(document.currentScript`) {
		t.Fatal("expected raw script output to avoid inline wrapper, got ", body)
	}
}

func testRenderError(t *testing.T, h func(doors.Source[test.Path]) gox.Elem) {
	t.Helper()
	page := emptyPage(t, h)
	test.TestMust(t, page, `[data-fw="error"]`)
}

func fetchPageCSPHeader(t *testing.T, h func(doors.Source[test.Path]) gox.Elem, csp doors.CSP) string {
	t.Helper()
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseCSP(r, csp)
		doors.UseModel(r, func(pr doors.RequestModel, r doors.Source[test.Path]) doors.Response {
			return doors.ResponseComp(&test.Page{
				Source: r,
				Header: "Testing Imports",
				H:      h,
				F:      &Empty{},
			})
		})
		doors.UseRoute(r, doors.RouteDir{Prefix: "module", DirPath: modulePath})
	})
	defer bro.Close()

	resp, err := http.Get(test.Host + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d", resp.StatusCode)
	}
	header := resp.Header.Get("Content-Security-Policy")
	if header == "" {
		t.Fatal("expected Content-Security-Policy header")
	}
	return header
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
	testScriptFetchExact(t, moduleRawBytesHead, "#module-raw-bytes", string(moduleRawBytes))
}
func TestModuleRawBytesShort(t *testing.T) {
	testModule(t, moduleRawBytesShortHead)
	testScriptFetchExact(t, moduleRawBytesShortHead, "#module-raw-bytes-short", string(moduleRawBytes))
}
func TestModuleRawBytesModify(t *testing.T) {
	testModule(t, moduleRawBytesModifyHead)
	testScriptFetchExact(t, moduleRawBytesModifyHead, "#module-raw-bytes-modify", string(moduleRawBytes))
}
func TestModulePreloadBytes(t *testing.T) {
	page := emptyPage(t, modulePreloadBytesHead)
	href := getAttr(t, page, `head link[rel="modulepreload"]`, "href")
	test.TestAttr(t, page, `head link[rel="modulepreload"]`, "rel", "modulepreload")
	body := fetchText(t, href)
	if strings.Contains(body, `_d0r(document.currentScript`) {
		t.Fatal("expected modulepreload bytes content without inline wrapper, got ", body)
	}
	if !strings.Contains(body, `export`) || !strings.Contains(body, `"hello"`) {
		t.Fatal("expected modulepreload bytes content to serve JS module, got ", body)
	}
}
func TestModulePreloadNamed(t *testing.T) {
	page := emptyPage(t, modulePreloadNamedHead)
	href := getAttr(t, page, `head link[rel="modulepreload"]`, "href")
	if !strings.Contains(href, ".module-preload.js") {
		t.Fatal("expected modulepreload href to contain .module-preload.js, got ", href)
	}
	body := fetchText(t, href)
	if strings.Contains(body, `_d0r(document.currentScript`) {
		t.Fatal("expected named modulepreload content without inline wrapper, got ", body)
	}
	if !strings.Contains(body, `export`) || !strings.Contains(body, `"hello"`) {
		t.Fatal("expected named modulepreload content to serve JS module, got ", body)
	}
}
func TestModuleFS(t *testing.T) {
	testModule(t, moduleBundleFSHead)
}

func TestModuleTypeTS(t *testing.T) {
	testModule(t, moduleTypeTSHead)
}

func TestModuleTypeJS(t *testing.T) {
	testModule(t, moduleTypeJSHead)
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
func TestScriptRawInline(t *testing.T) {
	page := valuePage(t, scriptRawHead)
	<-time.After(100 * time.Millisecond)
	test.TestReport(t, page, "hello")
	el, err := page.Timeout(200 * time.Millisecond).Element("#script-raw-inline")
	if err != nil {
		t.Fatal(err)
	}
	src, err := el.Attribute("src")
	if err != nil {
		t.Fatal(err)
	}
	if src != nil {
		t.Fatal("expected raw inline script to stay inline, got src ", *src)
	}
	body := compactText(getTextContent(t, page, "#script-raw-inline"))
	if !strings.Contains(body, `window.__importsValue="hello"`) {
		t.Fatal("expected raw inline script body to stay embedded, got ", body)
	}
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

func TestScriptInlineBytesOutputApplied(t *testing.T) {
	page := emptyPage(t, scriptInlineBytesHead)
	src := getAttr(t, page, "#script-inline-bytes", "src")
	body := fetchText(t, src)
	if !strings.Contains(body, `_d0r(document.currentScript`) {
		t.Fatal("expected inline script output to use inline wrapper, got ", body)
	}
	if !strings.Contains(compactText(body), `window.__importsValue="hello"`) {
		t.Fatal("expected inline script output to contain original body, got ", body)
	}
}

func TestScriptString(t *testing.T) {
	testValue(t, scriptStringHead)
}

func TestScriptRawOutputApplied(t *testing.T) {
	page := modulePage(t, moduleRawHead)
	src := getAttr(t, page, "#module-raw", "src")
	body := fetchText(t, src)
	if body != string(moduleRawBytes) {
		t.Fatal("expected raw script output to serve original source, got ", body)
	}
	if strings.Contains(body, `_d0r(document.currentScript`) {
		t.Fatal("expected raw script output to avoid inline wrapper, got ", body)
	}
}

func TestScriptBundleOutputApplied(t *testing.T) {
	page := emptyPage(t, reactHead)
	preactSrc := getAttr(t, page, "#preact-bundle", "src")
	preactBody := fetchText(t, preactSrc)
	if strings.Contains(preactBody, `from 'preact'`) {
		t.Fatal("expected bundled preact script to inline dependency import, got ", preactBody)
	}
	if !strings.Contains(preactBody, "Increment") {
		t.Fatal("expected bundled preact script to contain app code, got ", preactBody)
	}

	reactSrc := getAttr(t, page, "#react-bundle", "src")
	reactBody := fetchText(t, reactSrc)
	if strings.Contains(reactBody, `from 'react'`) {
		t.Fatal("expected bundled react script to inline dependency import, got ", reactBody)
	}
	if !strings.Contains(reactBody, "Increment") {
		t.Fatal("expected bundled react script to contain app code, got ", reactBody)
	}
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
	body := compactText(getTextContent(t, page, "head style"))
	if !strings.Contains(body, "h1{color:red;}") && !strings.Contains(body, "h1{color:red}") {
		t.Fatal("expected raw style body to stay embedded, got ", body)
	}
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
	test.TestMust(t, page, `script[src*="module.js"]`)
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

func TestCSPHeader(t *testing.T) {
	header := fetchPageCSPHeader(t, cspHead, doors.CSP{
		ConnectSources:      []string{"https://api.example.com"},
		ScriptStrictDynamic: true,
	})
	expect := []string{
		"default-src 'self'",
		"script-src 'self'",
		"'strict-dynamic'",
		test.Host + "/module/index.js",
		"style-src 'self'",
		test.Host + "/module/style.css",
		"connect-src 'self' https://api.example.com",
		"'sha256-",
	}
	for _, part := range expect {
		if !strings.Contains(header, part) {
			t.Fatalf("expected CSP header to contain %q, got %q", part, header)
		}
	}
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
	defer page.Close()
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

	href := getAttr(t, page, "#file-href", "href")
	if got := compactText(fetchTextFromPage(t, page, href)); got != "h1{color:red}" && got != "h1{color:red;}" {
		t.Fatal("expected file href to serve stylesheet bytes, got ", got)
	}

	href = getAttr(t, page, "#file-raw-href", "href")
	if got := compactText(fetchTextFromPage(t, page, href)); got != compactText(string(styleRawBytes)) {
		t.Fatal("expected raw file href to serve stylesheet bytes, got ", got)
	}

	src := getAttr(t, page, "#file-src", "src")
	if got := fetchTextFromPage(t, page, src); strings.Contains(got, `_d0r(document.currentScript`) || !strings.Contains(got, `export`) || !strings.Contains(got, `"hello"`) {
		t.Fatal("expected file src to serve script bytes, got ", got)
	}

	src = getAttr(t, page, "#file-raw-src", "src")
	if got := fetchTextFromPage(t, page, src); got != string(moduleBytes) {
		t.Fatal("expected raw file src to serve handler bytes, got ", got)
	}

	href = getAttr(t, page, "#file-href-modify", "href")
	if got := compactText(fetchTextFromPage(t, page, href)); got != "h1{color:red}" && got != "h1{color:red;}" {
		t.Fatal("expected file href modify to serve stylesheet bytes, got ", got)
	}

	href = getAttr(t, page, "#file-raw-href-modify", "href")
	if got := compactText(fetchTextFromPage(t, page, href)); got != compactText(string(styleRawBytes)) {
		t.Fatal("expected raw file href modify to serve handler bytes, got ", got)
	}

	src = getAttr(t, page, "#file-src-modify", "src")
	if got := fetchTextFromPage(t, page, src); strings.Contains(got, `_d0r(document.currentScript`) || !strings.Contains(got, `export`) || !strings.Contains(got, `"hello"`) {
		t.Fatal("expected file src modify to serve script bytes, got ", got)
	}

	src = getAttr(t, page, "#file-raw-src-modify", "src")
	if got := fetchTextFromPage(t, page, src); got != string(moduleBytes) {
		t.Fatal("expected raw file src modify to serve handler bytes, got ", got)
	}

	src = getAttr(t, page, "#file-img-fs", "src")
	body, contentType := fetchTextAndContentTypeFromPage(t, page, src)
	if !strings.Contains(body, "<svg") {
		t.Fatal("expected img ResourceFS route to serve svg content, got ", body)
	}
	if !strings.Contains(contentType, "image/svg+xml") {
		t.Fatal("expected img ResourceFS content type image/svg+xml, got ", contentType)
	}

	href = getAttr(t, page, "#cached-href", "href")
	if !strings.Contains(href, "/r/") {
		t.Fatal("expected cached href to use resource route, got ", href)
	}
	if !strings.Contains(href, ".hello.txt") {
		t.Fatal("expected cached href to contain .hello.txt, got ", href)
	}
	if got := fetchTextFromPage(t, page, href); got != "hello" {
		t.Fatal("expected cached href to serve hello, got ", got)
	}
	href = getAttr(t, page, "#cached-href-modify", "href")
	if !strings.Contains(href, "/r/") {
		t.Fatal("expected cached href modify to use resource route, got ", href)
	}
	if !strings.Contains(href, ".hello-modify.txt") {
		t.Fatal("expected cached href modify to contain .hello-modify.txt, got ", href)
	}
	if got := fetchTextFromPage(t, page, href); got != "hello" {
		t.Fatal("expected cached href modify to serve hello, got ", got)
	}
	href = getAttr(t, page, "#private-href", "href")
	if !strings.Contains(href, "/h/") {
		t.Fatal("expected private href to use hook route, got ", href)
	}
	if !strings.Contains(href, "/private.txt") {
		t.Fatal("expected private href to contain /private.txt, got ", href)
	}
	if got := fetchTextFromPage(t, page, href); got != "hello" {
		t.Fatal("expected private href to serve hello, got ", got)
	}
	href = getAttr(t, page, "#private-href-modify", "href")
	if !strings.Contains(href, "/h/") {
		t.Fatal("expected private href modify to use hook route, got ", href)
	}
	if !strings.Contains(href, "/private-modify.txt") {
		t.Fatal("expected private href modify to contain /private-modify.txt, got ", href)
	}
	if got := fetchTextFromPage(t, page, href); got != "hello" {
		t.Fatal("expected private href modify to serve hello, got ", got)
	}
	src = getAttr(t, page, "#private-frame", "src")
	if !strings.Contains(src, "/h/") {
		t.Fatal("expected private frame src to use hook route, got ", src)
	}
	if !strings.Contains(src, "/frame.html") {
		t.Fatal("expected private frame src to contain /frame.html, got ", src)
	}
	if got := fetchTextFromPage(t, page, src); got != `<html><body>frame</body></html>` {
		t.Fatal("expected private frame src to serve html, got ", got)
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

func TestFileHandlerTypeBad(t *testing.T) {
	testRenderError(t, fileHandlerTypeBad)
}

func TestScriptDuplicateOutputBad(t *testing.T) {
	testRenderError(t, scriptDuplicateOutputBad)
}

func TestScriptHandlerBundleBad(t *testing.T) {
	testRenderError(t, scriptHandlerBundleBad)
}

func TestScriptRawTSBad(t *testing.T) {
	testRenderError(t, scriptRawTSBad)
}

func TestScriptSpecifierNonModuleBad(t *testing.T) {
	testRenderError(t, scriptSpecifierNonModuleBad)
}

func TestScriptInlineModuleBad(t *testing.T) {
	testRenderError(t, scriptInlineModuleBad)
}

func TestScriptInlineTSBad(t *testing.T) {
	testRenderError(t, scriptInlineTSBad)
}

func TestScriptDirectBundleBad(t *testing.T) {
	testRenderError(t, scriptDirectBundleBad)
}

func TestScriptHandlerInlineBad(t *testing.T) {
	testRenderError(t, scriptHandlerInlineBad)
}

func TestModulePreloadInlineBad(t *testing.T) {
	testRenderError(t, modulePreloadInlineBad)
}

func TestStyleHandlerPrivateBad(t *testing.T) {
	testRenderError(t, styleHandlerPrivateBad)
}

func TestStyleDirectPrivateBad(t *testing.T) {
	testRenderError(t, styleDirectPrivateBad)
}
