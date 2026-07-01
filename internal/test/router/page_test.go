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

package router

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
	"github.com/go-rod/rod"
)

type PathParallel struct {
	Path bool `path:"/parallel"`
}

func routeBro(routes ...doors.RouteSource[doors.Location]) *test.Bro {
	return test.NewBro(browser, func(ctx context.Context, r doors.Request) gox.Comp {
		return doors.Route(routes...)
	})
}

func routeBroOptions(options []doors.With, routes ...doors.RouteSource[doors.Location]) *test.Bro {
	return test.NewBroOptions(browser, func(ctx context.Context, r doors.Request) gox.Comp {
		return doors.Route(routes...)
	}, options)
}

func routerConf(conf doors.Conf) doors.Conf {
	if test.LimitMode() {
		conf.InstanceGoroutineLimit = 1
	}
	return conf
}

func locationBro[C gox.Comp](render func(doors.Source[doors.Location]) C) *test.Bro {
	return test.NewBro(browser, func(ctx context.Context, r doors.Request) gox.Comp {
		return render(doors.Router(ctx))
	})
}

func testPath(t *testing.T, page *rod.Page, path string) {
	t.Helper()
	url := strings.Split(strings.Trim(page.MustInfo().URL, "/"), "/")
	last := url[len(url)-1]
	if last != path {
		t.Fatal("path expected " + path + " actual " + last)
	}
}

func testQueryValue(t *testing.T, page *rod.Page, key string, value string) {
	t.Helper()
	info, err := url.Parse(page.MustInfo().URL)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Query().Get(key); got != value {
		t.Fatalf("query %s expected %q actual %q", key, value, got)
	}
}

func testNoQueryValue(t *testing.T, page *rod.Page, key string) {
	t.Helper()
	info, err := url.Parse(page.MustInfo().URL)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Query().Get(key); got != "" {
		t.Fatalf("query %s expected empty actual %q", key, got)
	}
}

func waitQueryValue(t *testing.T, page *rod.Page, key string, value string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		info, err := url.Parse(page.MustInfo().URL)
		if err == nil && info.Query().Get(key) == value {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	testQueryValue(t, page, key, value)
}

func waitNoQueryValue(t *testing.T, page *rod.Page, key string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		info, err := url.Parse(page.MustInfo().URL)
		if err == nil && info.Query().Get(key) == "" {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	testNoQueryValue(t, page, key)
}

func waitEscapedPath(t *testing.T, page *rod.Page, path string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		info, err := url.Parse(page.MustInfo().URL)
		if err == nil && info.EscapedPath() == path {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	info, err := url.Parse(page.MustInfo().URL)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.EscapedPath(); got != path {
		t.Fatalf("escaped path expected %q actual %q", path, got)
	}
}

func waitPath(t *testing.T, page *rod.Page, path string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		segments := strings.Split(strings.Trim(page.MustInfo().URL, "/"), "/")
		if segments[len(segments)-1] == path {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	testPath(t, page, path)
}

func waitContent(t *testing.T, page *rod.Page, selector string, content string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if test.GetContent(t, page, selector) == content {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	test.TestContent(t, page, selector, content)
}

func hasClass(page *rod.Page, selector string, className string) bool {
	el, err := page.Timeout(200 * time.Millisecond).Element(selector)
	if err != nil {
		return false
	}
	classAttr, err := el.Attribute("class")
	if err != nil || classAttr == nil {
		return false
	}
	return strings.Contains(" "+*classAttr+" ", " "+className+" ")
}

func waitClass(t *testing.T, page *rod.Page, selector string, className string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if hasClass(page, selector, className) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	test.TestClass(t, page, selector, className)
}

func waitClassNot(t *testing.T, page *rod.Page, selector string, className string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if !hasClass(page, selector, className) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	test.TestClassNot(t, page, selector, className)
}

func TestPageStatic(t *testing.T) {
	bro := routeBro(doors.RouteModel(func(_ doors.Source[PathA]) gox.Elem {
		return static("a", 0)
	}))
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "a")
}

func TestPageStaticCode(t *testing.T) {
	bro := routeBro(doors.RouteModel(func(_ doors.Source[PathA]) gox.Elem {
		return static("a", 404)
	}))
	defer bro.Close()
	page := bro.PageStatus(t, "/a", http.StatusNotFound)
	defer page.Close()
	test.TestContent(t, page, "#path", "a")
}

func TestRouterBeamModelTransitionsAndDefault404(t *testing.T) {
	bro := test.NewBro(browser, func(ctx context.Context, r doors.Request) gox.Comp {
		return routerBeamDocument()
	})
	defer bro.Close()

	page := bro.Page(t, "/cross-a")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	test.TestContent(t, page, "#route-name", "cross-a")
	test.TestContent(t, page, "#route-model", "true")

	test.Click(t, page, "#beam-cross-next")
	waitContent(t, page, "#page-name", "cross-b")
	testPath(t, page, "cross-b")
	test.TestContent(t, page, "#route-model", "true")

	nextInstance := test.GetContent(t, page, "#instance-id")
	if nextInstance != initialInstance {
		t.Fatalf("expected RouterBeam ALink route switch to keep instance, got %q then %q", initialInstance, nextInstance)
	}

	test.Click(t, page, "#beam-cross-prev")
	waitContent(t, page, "#route-name", "cross-a")
	testPath(t, page, "cross-a")
	waitContent(t, page, "#instance-id", initialInstance)

	defaultPage := bro.PageStatus(t, "/missing", http.StatusNotFound)
	defer defaultPage.Close()

	test.TestContent(t, defaultPage, "#route-name", "default")
}

func TestRouterSourceCrossRouteKeepsInstance(t *testing.T) {
	bro := test.NewBro(browser, func(ctx context.Context, r doors.Request) gox.Comp {
		return routerLensCrossDocument()
	})
	defer bro.Close()

	page := bro.Page(t, "/cross-a")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	test.Click(t, page, "#cross-next")
	waitContent(t, page, "#page-name", "cross-b")
	testPath(t, page, "cross-b")

	nextInstance := test.GetContent(t, page, "#instance-id")
	if nextInstance != initialInstance {
		t.Fatalf("expected RouterSource ALink route switch to keep instance, got %q then %q", initialInstance, nextInstance)
	}

	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "cross-a")
	waitContent(t, page, "#instance-id", initialInstance)
	if _, err := page.Timeout(time.Second).Element("#cross-next"); err != nil {
		t.Fatalf("expected browser back to restore cross-a route content: %v", err)
	}

	defaultPage := bro.PageStatus(t, "/missing", http.StatusNotFound)
	defer defaultPage.Close()

	test.TestContent(t, defaultPage, "#route-name", "default")
}

func TestRouterSourceCombinedPathModelCustomAndRawRoutes(t *testing.T) {
	bro := test.NewBro(browser, func(ctx context.Context, r doors.Request) gox.Comp {
		return routerCombinedLensDocument()
	})
	defer bro.Close()

	page := bro.Page(t, "/cross-a")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	test.TestContent(t, page, "#route-name", "model-a")

	href := page.MustElement("#model-to-custom").MustAttribute("href")
	if href == nil || *href != "/custom/hello%20world%2Fone?tab=details" {
		t.Fatalf("expected custom encoder href %q actual %v", "/custom/hello%20world%2Fone?tab=details", href)
	}

	test.Click(t, page, "#model-to-custom")
	waitContent(t, page, "#route-name", "custom")
	waitContent(t, page, "#custom-id", "hello world/one")
	waitContent(t, page, "#custom-tab", "details")
	waitEscapedPath(t, page, "/custom/hello%20world%2Fone")
	waitQueryValue(t, page, "tab", "details")
	waitContent(t, page, "#instance-id", initialInstance)

	href = page.MustElement("#custom-next").MustAttribute("href")
	if href == nil || *href != "/custom/next%2Fchild?tab=again" {
		t.Fatalf("expected same custom branch href %q actual %v", "/custom/next%2Fchild?tab=again", href)
	}

	test.Click(t, page, "#custom-next")
	waitContent(t, page, "#custom-id", "next/child")
	waitContent(t, page, "#custom-tab", "again")
	waitEscapedPath(t, page, "/custom/next%2Fchild")
	waitQueryValue(t, page, "tab", "again")
	waitContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#custom-lens-write")
	waitContent(t, page, "#custom-id", "lens write/value")
	waitContent(t, page, "#custom-tab", "written")
	waitEscapedPath(t, page, "/custom/lens%20write%2Fvalue")
	waitQueryValue(t, page, "tab", "written")
	waitContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#custom-to-query")
	waitContent(t, page, "#route-name", "query")
	waitContent(t, page, "#tag", "from-custom")
	waitContent(t, page, "#page-value", "5")
	waitEscapedPath(t, page, "/q")
	waitContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#query-to-raw")
	waitContent(t, page, "#route-name", "raw")
	waitContent(t, page, "#raw-from", "query")
	waitEscapedPath(t, page, "/raw")
	waitContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#raw-to-model")
	waitContent(t, page, "#route-name", "model-a")
	testPath(t, page, "cross-a")
	waitContent(t, page, "#instance-id", initialInstance)

	defaultPage := bro.PageStatus(t, "/missing", http.StatusNotFound)
	defer defaultPage.Close()
	test.TestContent(t, defaultPage, "#route-name", "default")
}

func TestRouterBeamCombinedCustomEncoderRoute(t *testing.T) {
	bro := test.NewBro(browser, func(ctx context.Context, r doors.Request) gox.Comp {
		return routerCombinedBeamDocument()
	})
	defer bro.Close()

	page := bro.Page(t, "/custom/beam%2Fid?tab=start")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	test.TestContent(t, page, "#route-name", "custom-beam")
	test.TestContent(t, page, "#custom-id", "beam/id")
	test.TestContent(t, page, "#custom-tab", "start")

	href := page.MustElement("#beam-custom-next").MustAttribute("href")
	if href == nil || *href != "/custom/beam%20next?tab=read" {
		t.Fatalf("expected beam custom href %q actual %v", "/custom/beam%20next?tab=read", href)
	}

	test.Click(t, page, "#beam-custom-next")
	waitContent(t, page, "#custom-id", "beam next")
	waitContent(t, page, "#custom-tab", "read")
	waitEscapedPath(t, page, "/custom/beam%20next")
	waitQueryValue(t, page, "tab", "read")
	waitContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#beam-custom-to-model")
	waitContent(t, page, "#route-name", "cross-a")
	testPath(t, page, "cross-a")
	waitContent(t, page, "#instance-id", initialInstance)

	defaultPage := bro.PageStatus(t, "/missing", http.StatusNotFound)
	defer defaultPage.Close()
	test.TestContent(t, defaultPage, "#route-name", "default")
}

func TestLocations(t *testing.T) {
	bro := routeBro(
		doors.RouteModel(pageA),
		doors.RouteModel(pageC),
		doors.RouteModel(func(_ doors.Source[PathB]) gox.Elem {
			return static("b", 0)
		}),
	)
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()

	testPath(t, page, "a")
	test.Click(t, page, "#assign")
	testPath(t, page, "c1")
	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "a")

	test.Click(t, page, "#assign")
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

func TestBrowserBackRestoresQueryWithoutReload(t *testing.T) {
	bro := routeBro(doors.RouteModel(pageQuery))
	defer bro.Close()

	page := bro.Page(t, "/q")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	testNoQueryValue(t, page, "tag")
	testNoQueryValue(t, page, "page")
	test.TestContent(t, page, "#tag", "")
	test.TestContent(t, page, "#page-value", "")

	test.Click(t, page, "#query-next")
	waitQueryValue(t, page, "tag", "next")
	waitQueryValue(t, page, "page", "2")
	waitContent(t, page, "#tag", "next")
	waitContent(t, page, "#page-value", "2")

	nextInstance := test.GetContent(t, page, "#instance-id")
	if nextInstance != initialInstance {
		t.Fatalf("expected same instance after location navigation, got %q then %q", initialInstance, nextInstance)
	}

	page.NavigateBack()
	waitNoQueryValue(t, page, "tag")
	waitNoQueryValue(t, page, "page")
	waitContent(t, page, "#tag", "")
	waitContent(t, page, "#page-value", "")

	restoredInstance := test.GetContent(t, page, "#instance-id")
	if restoredInstance != initialInstance {
		t.Fatalf("expected browser back restore to keep same instance, got %q then %q", initialInstance, restoredInstance)
	}
}

func TestBrowserBackRestoresQueryWithDrainReload(t *testing.T) {
	bro := routeBro(doors.RouteModel(pageQuery))
	defer bro.Close()

	page := bro.Page(t, "/q")
	defer page.Close()
	bro.App().Drain(func() {})

	initialInstance := test.GetContent(t, page, "#instance-id")
	test.Click(t, page, "#query-next")
	waitQueryValue(t, page, "tag", "next")
	waitQueryValue(t, page, "page", "2")
	waitContent(t, page, "#tag", "next")
	waitContent(t, page, "#page-value", "2")

	nextInstance := test.GetContent(t, page, "#instance-id")
	if nextInstance == initialInstance {
		t.Fatalf("expected drain navigation to full-reload, got same instance %q", nextInstance)
	}

	page.NavigateBack()
	waitNoQueryValue(t, page, "tag")
	waitNoQueryValue(t, page, "page")
	waitContent(t, page, "#tag", "")
	waitContent(t, page, "#page-value", "")

	restoredInstance := test.GetContent(t, page, "#instance-id")
	if restoredInstance == nextInstance {
		t.Fatalf("expected drain browser back to full-reload, got same instance %q", restoredInstance)
	}
}

func TestPageLoadTimeout(t *testing.T) {
	bro := routeBroOptions([]doors.With{
		doors.WithConf(routerConf(doors.Conf{RequestTimeout: time.Second})),
		doors.WithErrorPage(func(_ *http.Request, err error) gox.Elem {
			return plainErrorPage(err)
		}),
	}, doors.RouteModel(func(_ doors.Source[PathSlow]) gox.Elem {
		return pageSlow()
	}))
	defer bro.Close()

	resp, err := http.Get(test.Host + "/slow")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyText := string(body)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected timeout status %d actual %d body %q", http.StatusInternalServerError, resp.StatusCode, bodyText)
	}
	if !strings.Contains(bodyText, "context deadline exceeded") {
		t.Fatalf("expected custom error page body, got %q", bodyText)
	}
	if strings.Contains(bodyText, "slow-page") {
		t.Fatalf("expected timeout to prevent slow page render, got %q", bodyText)
	}
}

func TestParallelComponentRender(t *testing.T) {
	bro := routeBroOptions([]doors.With{
		doors.WithConf(routerConf(doors.Conf{RequestTimeout: 2 * time.Second})),
	}, doors.RouteModel(func(_ doors.Source[PathParallel]) gox.Elem {
		return pageParallel()
	}))
	defer bro.Close()

	start := time.Now()
	resp, err := http.Get(test.Host + "/parallel")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d actual %d body %q", http.StatusOK, resp.StatusCode, string(body))
	}
	bodyText := string(body)
	if !strings.Contains(bodyText, "part-a") || !strings.Contains(bodyText, "part-b") || !strings.Contains(bodyText, "part-c") {
		t.Fatalf("expected all slow parts in body, got %q", bodyText)
	}
	if test.LimitMode() {
		if elapsed < 1300*time.Millisecond {
			t.Fatalf("expected serialized render in limit mode, got %s", elapsed)
		}
		return
	}
	if elapsed >= 900*time.Millisecond {
		t.Fatalf("expected explicit parallel render under %s, got %s", 900*time.Millisecond, elapsed)
	}
}

func TestLocationModelMatchesAnyURL(t *testing.T) {
	bro := locationBro(pageLocation)
	defer bro.Close()

	page := bro.Page(t, "/any/deep/path?tag=hello&page=7")
	defer page.Close()

	test.TestContent(t, page, "#location-string", "/any/deep/path?page=7&tag=hello")
	test.TestContent(t, page, "#location-path", "/any/deep/path")
	test.TestContent(t, page, "#tag-value", "hello")
	test.TestContent(t, page, "#page-query-value", "7")
}

func TestActiveLinkMatchersOnLoad(t *testing.T) {
	bro := locationBro(pageLocationActive)
	defer bro.Close()

	page := bro.Page(t, "/active?mode=view&optional=yes&page=9#details")
	defer page.Close()

	waitContent(t, page, "#location-string", "/active?mode=view&optional=yes&page=9")
	waitClass(t, page, "#active-query", "active")
	waitClass(t, page, "#active-only-ignore-some", "active")
	waitClass(t, page, "#active-only-some", "active")
	waitClass(t, page, "#active-only-if-present", "active")
	waitClass(t, page, "#active-fragment", "active")
	waitClassNot(t, page, "#active-full", "active")
	waitClassNot(t, page, "#active-starts", "active")
}

func TestActiveLinkMatchersByClick(t *testing.T) {
	bro := locationBro(pageLocationActive)
	defer bro.Close()

	page := bro.Page(t, "/active")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")

	waitClass(t, page, "#active-full", "active")
	waitClass(t, page, "#active-ignore-all", "active")
	waitClass(t, page, "#active-only-if-present", "active")
	waitClassNot(t, page, "#active-fragment", "active")
	waitClassNot(t, page, "#active-starts", "active")

	test.Click(t, page, "#nav-starts")
	waitContent(t, page, "#location-string", "/active/section/child")
	waitClass(t, page, "#active-starts", "active")
	waitClass(t, page, "#active-segments", "active")
	waitClassNot(t, page, "#active-full", "active")

	test.Click(t, page, "#nav-segments")
	waitContent(t, page, "#location-string", "/active/other")
	waitClass(t, page, "#active-segments", "active")
	waitClassNot(t, page, "#active-starts", "active")

	test.Click(t, page, "#nav-fragment")
	waitClass(t, page, "#active-fragment", "active")
	waitClass(t, page, "#active-full", "active")

	test.Click(t, page, "#nav-query")
	waitQueryValue(t, page, "mode", "view")
	waitQueryValue(t, page, "page", "9")
	waitClass(t, page, "#active-ignore-all", "active")
	waitClass(t, page, "#active-query", "active")
	waitClass(t, page, "#active-only-some", "active")
	waitClass(t, page, "#active-only-if-present", "active")
	waitClassNot(t, page, "#active-only-ignore-some", "active")
	waitClassNot(t, page, "#active-fragment", "active")

	test.Click(t, page, "#nav-query-optional")
	waitQueryValue(t, page, "optional", "yes")
	waitClass(t, page, "#active-query", "active")
	waitClass(t, page, "#active-only-ignore-some", "active")
	waitClass(t, page, "#active-only-some", "active")
	waitClass(t, page, "#active-only-if-present", "active")

	test.Click(t, page, "#nav-query-optional-miss")
	waitQueryValue(t, page, "optional", "no")
	waitClass(t, page, "#active-ignore-all", "active")
	waitClass(t, page, "#active-only-some", "active")
	waitClassNot(t, page, "#active-query", "active")
	waitClassNot(t, page, "#active-only-ignore-some", "active")
	waitClassNot(t, page, "#active-only-if-present", "active")

	finalInstance := test.GetContent(t, page, "#instance-id")
	if finalInstance != initialInstance {
		t.Fatalf("expected same instance for location navigation, got %q then %q", initialInstance, finalInstance)
	}
}

func TestPathModelEscapedSegmentDecodeAndEncode(t *testing.T) {
	bro := routeBro(doors.RouteModel(pageEscaped))
	defer bro.Close()

	page := bro.Page(t, "/escaped/hello%20world%2Fagain")
	defer page.Close()

	test.TestContent(t, page, "#name-value", "hello world/again")

	href := page.MustElement("#next-escaped").MustAttribute("href")
	if href == nil {
		t.Fatal("expected escaped path link href")
	}
	if *href != "/escaped/next%20value%2Fagain" {
		t.Fatalf("expected escaped href %q actual %q", "/escaped/next%20value%2Fagain", *href)
	}
}

func TestAfterAssign(t *testing.T) {
	bro := routeBro(
		doors.RouteModel(pageA),
		doors.RouteModel(pageC),
		doors.RouteModel(func(_ doors.Source[PathB]) gox.Elem {
			return static("b", 0)
		}),
	)
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
	bro := routeBro(
		doors.RouteModel(pageA),
		doors.RouteModel(pageC),
		doors.RouteModel(func(_ doors.Source[PathB]) gox.Elem {
			return static("b", 0)
		}),
	)
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
	bro := routeBro(
		doors.RouteModel(pageA),
		doors.RouteModel(pageC),
		doors.RouteModel(func(_ doors.Source[PathB]) gox.Elem {
			return static("b", 0)
		}),
	)
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

func TestHistoryReplaceContext(t *testing.T) {
	bro := locationBro(pageHistoryReplace)
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")

	test.Click(t, page, "#soft-assign")
	testPath(t, page, "c1")
	waitContent(t, page, "#path", "/c1")
	test.TestContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#soft-replace")
	testPath(t, page, "c2")
	waitContent(t, page, "#path", "/c2")
	test.TestContent(t, page, "#instance-id", initialInstance)

	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "a")
	waitContent(t, page, "#path", "/a")
	test.TestContent(t, page, "#instance-id", initialInstance)
}

func TestHistoryReplaceContextPathModel(t *testing.T) {
	bro := routeBro(doors.RouteModel(pageReplaceModel))
	defer bro.Close()
	page := bro.Page(t, "/rep-a")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	testPath(t, page, "rep-a")

	test.Click(t, page, "#pm-push")
	testPath(t, page, "rep-b")
	waitContent(t, page, "#step", "1")
	test.TestContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#pm-replace")
	testPath(t, page, "rep-c")
	waitContent(t, page, "#step", "2")
	test.TestContent(t, page, "#instance-id", initialInstance)

	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "rep-a")
	test.TestContent(t, page, "#instance-id", initialInstance)
}

func TestHistoryReplaceContextMutateAndNoLeak(t *testing.T) {
	bro := routeBro(doors.RouteModel(pageReplaceModel))
	defer bro.Close()
	page := bro.Page(t, "/rep-a")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	testPath(t, page, "rep-a")

	test.Click(t, page, "#pm-push")
	testPath(t, page, "rep-b")
	waitContent(t, page, "#step", "1")

	test.Click(t, page, "#pm-replace-mutate")
	testPath(t, page, "rep-c")
	waitContent(t, page, "#step", "2")

	test.Click(t, page, "#pm-push")
	testPath(t, page, "rep-b")
	waitContent(t, page, "#step", "1")
	test.TestContent(t, page, "#instance-id", initialInstance)

	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "rep-c")

	page.NavigateBack()
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "rep-a")
	test.TestContent(t, page, "#instance-id", initialInstance)
}

func historyLength(t *testing.T, page *rod.Page) int {
	t.Helper()
	obj, err := page.Eval("() => history.length")
	if err != nil {
		t.Fatal(err)
	}
	return obj.Value.Int()
}

func waitHistoryStateIdCleared(t *testing.T, page *rod.Page) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		obj, err := page.Eval("() => (history.state && history.state.id) ?? null")
		if err == nil && obj.Value.Nil() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("expected history.state.id to be cleared after replace confirm")
}

func waitHash(t *testing.T, page *rod.Page, hash string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		u, err := url.Parse(page.MustInfo().URL)
		if err == nil && u.Fragment == hash {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected url hash %q before timeout", hash)
}

func TestALinkHistoryReplace(t *testing.T) {
	bro := routeBro(doors.RouteModel(pageLinkNav))
	defer bro.Close()
	page := bro.Page(t, "/la")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	testPath(t, page, "la")
	lenStart := historyLength(t, page)

	test.Click(t, page, "#push-b")
	testPath(t, page, "lb")
	waitContent(t, page, "#step", "1")
	if got := historyLength(t, page); got != lenStart+1 {
		t.Fatalf("push should grow history by 1: %d -> %d", lenStart, got)
	}

	test.Click(t, page, "#replace-c")
	testPath(t, page, "lc")
	waitContent(t, page, "#step", "2")
	if got := historyLength(t, page); got != lenStart+1 {
		t.Fatalf("replace must not grow history: expected %d got %d", lenStart+1, got)
	}
	waitHistoryStateIdCleared(t, page)

	page.NavigateBack()
	<-time.After(150 * time.Millisecond)
	testPath(t, page, "la")
	waitContent(t, page, "#step", "0")
	test.TestContent(t, page, "#instance-id", initialInstance)
}

func TestALinkHistoryReplaceEdges(t *testing.T) {
	bro := routeBro(doors.RouteModel(pageLinkNav))
	defer bro.Close()
	page := bro.Page(t, "/la")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	testPath(t, page, "la")
	lenStart := historyLength(t, page)

	test.Click(t, page, "#replace-self")
	<-time.After(100 * time.Millisecond)
	testPath(t, page, "la")
	test.TestContent(t, page, "#step", "0")
	if got := historyLength(t, page); got != lenStart {
		t.Fatalf("self replace-link must not change history: %d -> %d", lenStart, got)
	}
	test.TestContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#replace-hash")
	waitHash(t, page, "sec")
	if got := historyLength(t, page); got != lenStart {
		t.Fatalf("hash-only replace-link must not change history: %d -> %d", lenStart, got)
	}
	test.TestContent(t, page, "#step", "0")
	test.TestContent(t, page, "#instance-id", initialInstance)
}

func TestALinkHistoryReplaceRevertsOnError(t *testing.T) {
	bro := routeBro(doors.RouteModel(pageLinkNav))
	defer bro.Close()
	page := bro.Page(t, "/la")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	testPath(t, page, "la")
	lenStart := historyLength(t, page)

	// Fail the next hook POST so the optimistic replace navigation gets a
	// network error. Only the hook path (/~/<server>/h/...) is intercepted, so
	// the SSE sync stream (/s/) and every other request pass through untouched.
	stop := test.HijackFailOnce(t, page, "*/h/*")
	defer stop()

	test.Click(t, page, "#replace-c")

	// navigator.replace optimistically replaceState-ed the URL to /lc; the failed
	// hook must trigger revert() which restores the prior href and state. Revert
	// is async (it runs after the failed fetch), so poll for the snap back.
	waitPath(t, page, "la")
	waitHistoryStateIdCleared(t, page)
	if got := historyLength(t, page); got != lenStart {
		t.Fatalf("revert must not leak a history entry: %d -> %d", lenStart, got)
	}
	test.TestContent(t, page, "#step", "0")
	test.TestContent(t, page, "#instance-id", initialInstance)

	// Drop the failure and prove the navigator state is clean post-revert: a
	// normal push nav must work, grow history by one and stay interactive.
	stop()
	test.Click(t, page, "#push-b")
	testPath(t, page, "lb")
	waitContent(t, page, "#step", "1")
	if got := historyLength(t, page); got != lenStart+1 {
		t.Fatalf("push after revert should grow history by 1: %d -> %d", lenStart, got)
	}

	// And a real browser back still works (no stray blockPop / corrupted state).
	page.NavigateBack()
	<-time.After(150 * time.Millisecond)
	testPath(t, page, "la")
	waitContent(t, page, "#step", "0")
	test.TestContent(t, page, "#instance-id", initialInstance)
}

func TestLastSeen(t *testing.T) {
	bro := locationBro(pageLastSeen)
	defer bro.Close()
	page := bro.Page(t, "/x")
	defer page.Close()

	test.Click(t, page, "#check")
	waitContent(t, page, "#result", "ok")
}

func TestRouteBindMethods(t *testing.T) {
	bro := test.NewBro(browser, func(ctx context.Context, r doors.Request) gox.Comp {
		return routeBindDocument()
	})
	defer bro.Close()

	page := bro.Page(t, "/custom/bind-test?tab=active")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")
	test.TestContent(t, page, "#route-name", "derive-bind")
	test.TestContent(t, page, "#custom-id", "bind-test")
	test.TestContent(t, page, "#custom-tab", "active")

	test.Click(t, page, "#to-match")
	waitContent(t, page, "#route-name", "match-bind")
	waitContent(t, page, "#match-path", "/cross-a")
	testPath(t, page, "cross-a")
	waitContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#to-derive")
	waitContent(t, page, "#route-name", "derive-bind")
	waitContent(t, page, "#custom-id", "bind-test")
	waitContent(t, page, "#custom-tab", "active")
	waitContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#to-default")
	waitContent(t, page, "#route-name", "default-bind")
	waitContent(t, page, "#default-path", "/unknown")
	waitContent(t, page, "#instance-id", initialInstance)

	test.Click(t, page, "#to-match")
	waitContent(t, page, "#route-name", "match-bind")
	waitContent(t, page, "#instance-id", initialInstance)
}
