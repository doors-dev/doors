package router

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/doors"
	introuter "github.com/doors-dev/doors/internal/router"
	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
	"github.com/go-rod/rod"
)

type PathParallel struct {
	Path bool `path:"/parallel"`
}

type slowPart struct {
	id    string
	text  string
	delay time.Duration
}

func (s slowPart) Main() gox.Elem {
	<-time.After(s.delay)
	return gox.Elem(func(cur gox.Cursor) error {
		if err := cur.Init("div"); err != nil {
			return err
		}
		if err := cur.AttrSet("id", s.id); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		if err := cur.Text(s.text); err != nil {
			return err
		}
		return cur.Close()
	})
}

func pageParallel() gox.Elem {
	return gox.Elem(func(cur gox.Cursor) error {
		if err := cur.Init("html"); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		if err := cur.Init("body"); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		if err := cur.Comp(slowPart{id: "part-a", text: "part-a", delay: 500 * time.Millisecond}); err != nil {
			return err
		}
		if err := cur.Comp(slowPart{id: "part-b", text: "part-b", delay: 500 * time.Millisecond}); err != nil {
			return err
		}
		if err := cur.Close(); err != nil {
			return err
		}
		return cur.Close()
	})
}

func testPath(t *testing.T, page *rod.Page, path string) {
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

func TestPageStatic(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(req doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseComp(static("a", 0))
		})
	})
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "a")
}

func TestPageStaticCode(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseComp(static("a", 404))
		})
	})
	defer bro.Close()
	page := bro.PageStatus(t, "/a", 404)
	defer page.Close()
	test.TestContent(t, page, "#path", "a")
}

func TestPageRedirect(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseRedirect(PathB{}, 0)
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathB]) doors.Response {
			return doors.ResponseComp(static("b", 0))
		})
	})
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "b")
}

func TestPageReroute(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseReroute(PathC{PathC1: true})
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathC]) doors.Response {
			return doors.ResponseComp(pageC(r))
		})
	})
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "c1")
	testPath(t, page, "c1")
}

/*ac // removed
func TestPageRerouteDetached(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.ReqModel, r doors.Source[PathA]) doors.Res {
			return doors.ResReroute(PathC{PathC1: true}, true)
		})
		doors.UseModel(r, func(p doors.ReqModel, r doors.Source[PathC]) doors.Res {
			return doors.ResPage(pageC(r))
		})
	})
	defer bro.Close()
	page := bro.Page(t, "/a")
	defer page.Close()
	test.TestContent(t, page, "#path", "c1")
	testPath(t, page, "a")
	test.Click(t, page, "#c2")
	test.TestContent(t, page, "#path", "c2")
	testPath(t, page, "a")
} */

func TestPageError(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseRedirect(PathC{}, 0)
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathC]) doors.Response {
			return doors.ResponseComp(pageC(r))
		})
		doors.UseErrorPage(r, func(l doors.Location, err error) gox.Comp {
			return static("error", -1)
		})
	})
	defer bro.Close()
	page := bro.PageStatus(t, "/a", 500)
	defer page.Close()
	test.TestContent(t, page, "#path", "error")
}

func TestPageInfiniteReroute(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseReroute(PathC{})
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathC]) doors.Response {
			return doors.ResponseReroute(PathA{})
		})
		doors.UseErrorPage(r, func(l doors.Location, err error) gox.Comp {
			return static("error", -1)
		})
	})
	defer bro.Close()
	page := bro.PageStatus(t, "/a", 500)
	defer page.Close()
	test.TestContent(t, page, "#path", "error")
}

func TestLocations(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseComp(pageA(r))
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathC]) doors.Response {
			return doors.ResponseComp(pageC(r))
		})
		doors.UseErrorPage(r, func(l doors.Location, err error) gox.Comp {
			return static("error", -1)
		})
	})
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

func TestBrowserBackRestoresQueryWithoutReload(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, s doors.Source[PathQuery]) doors.Response {
			return doors.ResponseComp(pageQuery(s))
		})
	})
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
		t.Fatalf("expected same instance after same-model navigation, got %q then %q", initialInstance, nextInstance)
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

func TestBrowserBackRestoresQueryWithZombieReload(t *testing.T) {
	bro := test.NewBroWrap(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, s doors.Source[PathQuery]) doors.Response {
			return doors.ResponseComp(pageQuery(s))
		})
	}, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set(introuter.ZombieHeader, "1")
			w.Header().Set(introuter.ZombieHeader, "1")
			next.ServeHTTP(w, r)
		})
	})
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
	if nextInstance == initialInstance {
		t.Fatalf("expected zombie same-model navigation to full-reload, got same instance %q", nextInstance)
	}

	page.NavigateBack()
	waitNoQueryValue(t, page, "tag")
	waitNoQueryValue(t, page, "page")
	waitContent(t, page, "#tag", "")
	waitContent(t, page, "#page-value", "")

	restoredInstance := test.GetContent(t, page, "#instance-id")
	if restoredInstance == nextInstance {
		t.Fatalf("expected zombie browser back to full-reload, got same instance %q", restoredInstance)
	}
	if restoredInstance == initialInstance {
		t.Fatalf("expected zombie browser back to create a new instance, got initial instance %q", restoredInstance)
	}
}

func TestBrowserBackRestoresPreviousInstanceAcrossModels(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, s doors.Source[PathCrossA]) doors.Response {
			return doors.ResponseComp(pageCrossA())
		})
		doors.UseModel(r, func(p doors.RequestModel, s doors.Source[PathCrossB]) doors.Response {
			return doors.ResponseComp(pageCrossB())
		})
	})
	defer bro.Close()

	page := bro.Page(t, "/cross-a")
	defer page.Close()

	initialInstance := test.GetContent(t, page, "#instance-id")

	test.Click(t, page, "#cross-next")
	waitContent(t, page, "#page-name", "cross-b")

	nextInstance := test.GetContent(t, page, "#instance-id")
	if nextInstance == initialInstance {
		t.Fatalf("expected different instance after cross-model ALink navigation, got %q", nextInstance)
	}

	page.NavigateBack()
	waitContent(t, page, "#cross-next", "cross-next")

	restoredInstance := test.GetContent(t, page, "#instance-id")
	if restoredInstance != initialInstance {
		t.Fatalf("expected browser back to restore previous instance %q, got %q", initialInstance, restoredInstance)
	}
}

func TestPageLoadTimeout(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseSystemConf(r, doors.SystemConf{
			RequestTimeout: time.Second,
		})
		doors.UseModel(r, func(p doors.RequestModel, s doors.Source[PathSlow]) doors.Response {
			return doors.ResponseComp(pageSlow())
		})
		doors.UseErrorPage(r, func(l doors.Location, err error) gox.Comp {
			return pageError(err)
		})
	})
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
	if !strings.Contains(bodyText, "context deadline exceeded") {
		t.Fatalf("expected timeout error page body, got %q", bodyText)
	}
	if strings.Contains(bodyText, "slow-page") {
		t.Fatalf("expected timeout to prevent slow page render, got %q", bodyText)
	}
}

func TestParallelComponentRender(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseSystemConf(r, doors.SystemConf{
			RequestTimeout: 2 * time.Second,
		})
		doors.UseModel(r, func(p doors.RequestModel, s doors.Source[PathParallel]) doors.Response {
			return doors.ResponseComp(pageParallel())
		})
	})
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
	if !strings.Contains(bodyText, "part-a") || !strings.Contains(bodyText, "part-b") {
		t.Fatalf("expected both slow parts in body, got %q", bodyText)
	}
	if test.LimitMode() {
		if elapsed < 900*time.Millisecond {
			t.Fatalf("expected serialized render in limit mode, got %s", elapsed)
		}
		return
	}
	if elapsed >= 900*time.Millisecond {
		t.Fatalf("expected parallel render under %s, got %s", 900*time.Millisecond, elapsed)
	}
}

func TestLocationModelMatchesAnyURL(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, s doors.Source[doors.Location]) doors.Response {
			return doors.ResponseComp(pageLocation(s))
		})
	})
	defer bro.Close()

	page := bro.Page(t, "/any/deep/path?tag=hello&page=7")
	defer page.Close()

	test.TestContent(t, page, "#location-string", "/any/deep/path?page=7&tag=hello")
	test.TestContent(t, page, "#location-path", "/any/deep/path")
	test.TestContent(t, page, "#tag-value", "hello")
	test.TestContent(t, page, "#page-query-value", "7")
}

func TestPathModelEscapedSegmentDecodeAndEncode(t *testing.T) {
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, s doors.Source[PathEscaped]) doors.Response {
			return doors.ResponseComp(pageEscaped(s))
		})
	})
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
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseComp(pageA(r))
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathC]) doors.Response {
			return doors.ResponseComp(pageC(r))
		})
		doors.UseErrorPage(r, func(l doors.Location, err error) gox.Comp {
			return static("error", 0)
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathB]) doors.Response {
			return doors.ResponseComp(static("b", 0))
		})
	})
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
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseComp(pageA(r))
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathC]) doors.Response {
			return doors.ResponseComp(pageC(r))
		})
		doors.UseErrorPage(r, func(l doors.Location, err error) gox.Comp {
			return static("error", -1)
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathB]) doors.Response {
			return doors.ResponseComp(static("b", 0))
		})
	})
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
	bro := test.NewBro(browser, func(r doors.Router) {
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathA]) doors.Response {
			return doors.ResponseComp(pageA(r))
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathC]) doors.Response {
			return doors.ResponseComp(pageC(r))
		})
		doors.UseErrorPage(r, func(l doors.Location, err error) gox.Comp {
			return static("error", -1)
		})
		doors.UseModel(r, func(p doors.RequestModel, r doors.Source[PathB]) doors.Response {
			return doors.ResponseComp(static("b", 0))
		})
	})
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
