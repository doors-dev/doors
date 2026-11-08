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

package components

import (
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
	"github.com/go-rod/rod"
)

func TestTitle(t *testing.T) {
	p := common.RandId()
	bro := test.NewPathBro(browser, func(r test.PathLens) gox.Comp {
		return &test.Page{
			Source: r,
			F: &LinksFragment{
				Param: p,
			},
		}
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
	bro := test.NewPathBro(browser, func(r test.PathLens) gox.Comp {
		return &test.Page{
			Source: r,
			F: &LinksFragment{
				Param: p,
			},
		}
	})
	defer bro.Close()
	page := bro.Page(t, "/")
	defer page.Close()
	<-time.After(200 * time.Millisecond)

	test.TestAttrNo(t, page, "#action-target", "data-indicated")
	test.ClickNow(t, page, "#action-indicate")
	<-time.After(50 * time.Millisecond)
	test.TestAttrNow(t, page, "#action-target", "data-indicated", "true")
	<-time.After(250 * time.Millisecond)
	test.TestAttrNo(t, page, "#action-target", "data-indicated")

	if page.MustEval("() => window.scrollY").Num() != 0 {
		t.Fatal("expected page to start at scrollY=0")
	}
	test.ClickNow(t, page, "#action-scroll")
	<-time.After(200 * time.Millisecond)
	if page.MustEval("() => window.scrollY").Num() <= 0 {
		t.Fatal("expected ActionScroll to move the page")
	}

	test.ClickNow(t, page, "#raw-assign")
	<-time.After(300 * time.Millisecond)
	test.TestContent(t, page, "title", "s")

	page.NavigateBack()
	<-time.After(300 * time.Millisecond)
	histLen := page.MustEval("() => history.length").Num()
	test.ClickNow(t, page, "#raw-replace")
	<-time.After(300 * time.Millisecond)
	test.TestContent(t, page, "title", "s")
	if got := page.MustEval("() => history.length").Num(); got != histLen {
		t.Fatalf("expected RawReplace to keep history length %v, got %v", histLen, got)
	}
}

func proxyPage(t *testing.T) *rod.Page {
	t.Helper()
	bro := test.NewPathBro(browser, func(r test.PathLens) gox.Comp {
		return &test.Page{
			Source: r,
			F: &ProxyFragment{
				r: test.NewReporter(1),
			},
		}
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

func TestProxyModClassLiteral(t *testing.T) {
	page := proxyPage(t)
	test.TestClass(t, page, "#proxy-class-literal", "proxy-literal")
}

func TestProxyModDirect(t *testing.T) {
	page := proxyPage(t)
	test.TestAttr(t, page, "#proxy-direct-mod", "data-proxy-mod", "button:direct")
}

func TestProxyModClassContainerFirstElementOnly(t *testing.T) {
	page := proxyPage(t)
	test.TestClass(t, page, "#proxy-class-container", "proxy-container")
	test.TestClassNot(t, page, "#proxy-class-container", "proxy-skip")
	test.TestClassNot(t, page, "#proxy-class-sibling", "proxy-container")
}

func TestProxyModClassComponent(t *testing.T) {
	page := proxyPage(t)
	test.TestClass(t, page, "#proxy-class-component", "base-component")
	test.TestClass(t, page, "#proxy-class-component", "proxy-component")
}

func TestProxyModClassParallel(t *testing.T) {
	page := proxyPage(t)
	test.TestClass(t, page, "#proxy-class-parallel", "proxy-parallel")
}

func TestClassAttributeAndModifier(t *testing.T) {
	page := proxyPage(t)
	test.TestClass(t, page, "#class-attr", "class-attr")
	test.TestClass(t, page, "#class-attr", "class-added")
	test.TestClassNot(t, page, "#class-attr", "class-skip")
	test.TestClass(t, page, "#class-mod", "base-mod")
	test.TestClass(t, page, "#class-mod", "class-mod")
	test.TestClassNot(t, page, "#class-mod", "class-skip")
}

func proxyDoorPage(t *testing.T) *rod.Page {
	t.Helper()
	bro := test.NewPathBro(browser, func(r test.PathLens) gox.Comp {
		return &test.Page{
			Source: r,
			F: &ProxyDoorFragment{
				r: test.NewReporter(5),
			},
		}
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

func TestProxyModDoorFrame(t *testing.T) {
	page := proxyDoorPage(t)
	test.TestAttr(t, page, "d0-r.pd-class", "data-pd", "d0-r:wrap")
	test.TestContent(t, page, "d0-r.pd-class #pd-wrap-inner", "wrap-0")
	test.TestAttr(t, page, "#pd-elem", "data-pd", "div:elem")
	test.TestAttr(t, page, "#pd-proxy", "data-pd", "div:proxy")
	if n := test.Count(page, "#pd-elem[data-d0r], #pd-proxy[data-d0r]"); n != 2 {
		t.Fatalf("expected user containers to be door frames, got %d", n)
	}
	test.Click(t, page, "#pd-wrap-inner")
	test.TestReportId(t, page, 0, "wrap-click")
	test.Click(t, page, "#pd-elem-inner")
	test.TestReportId(t, page, 1, "elem-click")
	test.Click(t, page, "#pd-proxy-inner")
	test.TestReportId(t, page, 2, "proxy-click")
}

func TestProxyModDoorInner(t *testing.T) {
	page := proxyDoorPage(t)
	test.Click(t, page, "#pd-inner")
	test.TestContent(t, page, "d0-r.pd-class #pd-wrap-inner", "wrap-1")
	test.TestAttr(t, page, "d0-r.pd-class", "data-pd", "d0-r:wrap")
	test.TestContent(t, page, "#pd-elem #pd-elem-inner", "elem-1")
	test.TestAttr(t, page, "#pd-elem", "data-pd", "div:elem")
	test.TestContent(t, page, "#pd-proxy #pd-proxy-inner", "proxy-1")
	test.TestAttr(t, page, "#pd-proxy", "data-pd", "div:proxy")
	test.Click(t, page, "#pd-wrap-inner")
	test.TestReportId(t, page, 0, "wrap-click")
	test.Click(t, page, "#pd-elem-inner")
	test.TestReportId(t, page, 1, "elem-click")
	test.Click(t, page, "#pd-proxy-inner")
	test.TestReportId(t, page, 2, "proxy-click")
}

func TestProxyModDoorOuter(t *testing.T) {
	page := proxyDoorPage(t)
	test.Click(t, page, "#pd-outer")
	test.TestContent(t, page, "d0-r.pd-class #pd-wrap-inner", "wrap-2")
	test.TestAttr(t, page, "d0-r.pd-class", "data-pd", "d0-r:wrap")
	test.TestContent(t, page, "#pd-elem #pd-elem-inner", "elem-2")
	test.TestAttr(t, page, "#pd-elem", "data-pd", "div:elem")
	test.TestContent(t, page, "#pd-proxy #pd-proxy-inner", "proxy-2")
	test.TestAttr(t, page, "#pd-proxy", "data-pd", "div:proxy")
	test.Click(t, page, "#pd-wrap-inner")
	test.TestReportId(t, page, 0, "wrap-click")
	test.Click(t, page, "#pd-elem-inner")
	test.TestReportId(t, page, 1, "elem-click")
	test.Click(t, page, "#pd-proxy-inner")
	test.TestReportId(t, page, 2, "proxy-click")
}

func TestProxyModDoorStatic(t *testing.T) {
	page := proxyDoorPage(t)
	test.TestAttr(t, page, "#pd-static", "data-pd", "p:static")
	test.TestAttrNo(t, page, "#pd-static-sibling", "data-pd")
	if n := test.Count(page, "#pd-root > d0-r #pd-static"); n != 0 {
		t.Fatalf("static door content should not be framed, got %d", n)
	}
	test.TestAttr(t, page, "d0-r[data-pd='d0-r:mounted']", "data-pd", "d0-r:mounted")
	test.Click(t, page, "#pd-mounted-static")
	test.TestAttr(t, page, "#pd-mounted-static", "data-pd", "p:mounted")
	if n := test.Count(page, "[data-pd='d0-r:mounted']"); n != 0 {
		t.Fatalf("mounted door frame should be replaced by static content, got %d", n)
	}
}

func TestProxyModDoorID(t *testing.T) {
	page := proxyDoorPage(t)
	test.TestContent(t, page, "d0-r#pd-id #pd-id-inner", "id-0")
	test.Click(t, page, "#pd-id-inner-update")
	test.TestContent(t, page, "d0-r#pd-id #pd-id-inner", "id-1")
	test.Click(t, page, "#pd-id-outer-update")
	test.TestContent(t, page, "d0-r#pd-id #pd-id-inner", "id-2")
	test.Click(t, page, "#pd-id-inner-update")
	test.TestContent(t, page, "d0-r#pd-id #pd-id-inner", "id-1")
}

func TestProxyModDoorOwnership(t *testing.T) {
	page := proxyDoorPage(t)
	test.TestAttr(t, page, "#pd-host-a > d0-r > d0-r", "data-pd", "d0-r:a")
	test.Click(t, page, "#pd-move-inner")
	test.TestReportId(t, page, 3, "move-a")
	test.Click(t, page, "#pd-remod")
	test.TestAttr(t, page, "#pd-host-a > d0-r > d0-r", "data-pd", "d0-r:b")
	test.Click(t, page, "#pd-move-inner")
	test.TestReportId(t, page, 3, "move-b")
	test.Click(t, page, "#pd-move")
	test.TestReportId(t, page, 3, "moved")
	test.TestContent(t, page, "#pd-host-b > d0-r > d0-r #pd-move-inner", "move")
	test.TestMustNot(t, page, "#pd-host-a #pd-move-inner")
	test.TestAttrNo(t, page, "#pd-host-b > d0-r > d0-r", "data-pd")
	test.Click(t, page, "#pd-move-inner")
	test.Click(t, page, "#pd-ping")
	test.TestReportId(t, page, 4, "ping")
	test.TestReportId(t, page, 3, "moved")
}

func TestProxyModDoorStaticText(t *testing.T) {
	page := proxyDoorPage(t)
	test.TestAttr(t, page, "#pd-text-host > d0-r", "data-pd", "d0-r:text")
	test.Click(t, page, "#pd-text-static")
	test.TestContent(t, page, "#pd-text-host", "done")
	test.TestMustNot(t, page, "#pd-text-host > d0-r")
}
