package printer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

func writePrinterTempFile(t *testing.T, name string, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSourceFSBranches(t *testing.T) {
	fsys := fstest.MapFS{
		"asset.txt": {Data: []byte("fs-content")},
	}
	src := SourceFS{FS: fsys, Entry: "asset.txt"}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/asset.txt", nil)
	if src.Handler()(context.Background(), rec, req) {
		t.Fatal("SourceFS handler should return false")
	}
	if rec.Body.String() != "fs-content" {
		t.Fatalf("SourceFS handler body = %q", rec.Body.String())
	}

	entry := src.StaticEntry()
	static, ok := entry.(resources.StaticFS)
	if !ok || static.Path != "asset.txt" {
		t.Fatalf("SourceFS.StaticEntry = %#v", entry)
	}
	if got := src.scriptEntry(true, false); func() bool { _, ok := got.(resources.ScriptInlineFS); return ok }() == false {
		t.Fatalf("SourceFS inline script entry = %#v", got)
	}
}

func TestSourceLocalFSBranches(t *testing.T) {
	path := writePrinterTempFile(t, "asset.txt", "path-content")
	src := SourceLocalFS(path)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/asset.txt", nil)
	if src.Handler()(context.Background(), rec, req) {
		t.Fatal("SourceLocalFS handler should return false")
	}
	if rec.Body.String() != "path-content" {
		t.Fatalf("SourceLocalFS handler body = %q", rec.Body.String())
	}

	entry := src.StaticEntry()
	static, ok := entry.(resources.StaticPath)
	if !ok || static.Path != path {
		t.Fatalf("SourceLocalFS.StaticEntry = %#v", entry)
	}
	if got := src.scriptEntry(true, false); func() bool { _, ok := got.(resources.ScriptInlinePath); return ok }() == false {
		t.Fatalf("SourceLocalFS inline script entry = %#v", got)
	}
}

func TestSourceHookAndExternalBranches(t *testing.T) {
	hook := SourceHook(func(_ context.Context, w http.ResponseWriter, _ *http.Request) bool {
		_, _ = w.Write([]byte("hook"))
		return true
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if !hook.Handler()(context.Background(), rec, req) {
		t.Fatal("SourceHook handler should preserve return value")
	}
	if rec.Body.String() != "hook" {
		t.Fatalf("SourceHook body = %q", rec.Body.String())
	}
	if hook.scriptEntry(false, false) != nil {
		t.Fatal("SourceHook scriptEntry should be nil")
	}

	external := SourceExternal("https://example.com/app.js")
	if external.Handler() != nil {
		t.Fatal("SourceExternal handler should be nil")
	}
	if external.scriptEntry(false, false) != nil {
		t.Fatal("SourceExternal scriptEntry should be nil")
	}
	if external.styleEntry() != nil {
		t.Fatal("SourceExternal styleEntry should be nil")
	}
	attrs := gox.NewAttrs()
	if err := external.Modify(context.Background(), "link", attrs); err != nil {
		t.Fatal(err)
	}
	if got, _ := attrs.Find("href"); got.Value() != external {
		t.Fatalf("SourceExternal.Modify href = %#v", got.Value())
	}
}

func TestSourceProxyInvalidURLAndStringStatic(t *testing.T) {
	proxy := SourceProxy("://bad url")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if proxy.Handler()(context.Background(), rec, req) {
		t.Fatal("SourceProxy invalid handler should return false")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("SourceProxy invalid status = %d", rec.Code)
	}
	if proxy.scriptEntry(false, false) != nil {
		t.Fatal("SourceProxy scriptEntry should be nil")
	}

	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("proxied"))
	}))
	defer targetServer.Close()
	valid := SourceProxy(targetServer.URL)
	if valid.Handler() == nil {
		t.Fatal("SourceProxy valid handler should not be nil")
	}

	str := SourceString("hello")
	entry := str.StaticEntry()
	static, ok := entry.(resources.StaticString)
	if !ok || static.Content != "hello" {
		t.Fatalf("SourceString.StaticEntry = %#v", entry)
	}
}

func TestSourceNameAndHelpers(t *testing.T) {
	if got := sourceNameFromPath("", "js"); got != "script.js" {
		t.Fatalf("empty path name = %q", got)
	}
	if got := sourceNameFromPath(".", "css"); got != "style.css" {
		t.Fatalf("dot path name = %q", got)
	}
	if got := sourceNameFromPath("/tmp/app.bundle.ts", "js"); got != "app.bundle.js" {
		t.Fatalf("trimmed path name = %q", got)
	}

	attrs := gox.NewAttrs()
	if err := modifySource("script", attrs, "value"); err != nil {
		t.Fatal(err)
	}
	if got, _ := attrs.Find("src"); got.Value() != "value" {
		t.Fatalf("modifySource src = %#v", got.Value())
	}

	attrs = gox.NewAttrs()
	if err := modifySource("link", attrs, "value"); err != nil {
		t.Fatal(err)
	}
	if got, _ := attrs.Find("href"); got.Value() != "value" {
		t.Fatalf("modifySource href = %#v", got.Value())
	}

	if err := modifySource("div", gox.NewAttrs(), "value"); err == nil {
		t.Fatal("modifySource should reject unsupported tags")
	}

	if sourceDefaultName("css") != "style.css" || sourceDefaultName("js") != "script.js" {
		t.Fatal("sourceDefaultName mismatch")
	}
}

func TestSourceProxyHandlerRewritesHost(t *testing.T) {
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == "" {
			t.Fatal("proxy request host should be set")
		}
		_, _ = w.Write([]byte(r.Host + "|" + r.URL.Path))
	}))
	defer targetServer.Close()

	targetURL, err := url.Parse(targetServer.URL)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ignored", nil)
	if SourceProxy(targetServer.URL).Handler()(context.Background(), rec, req) {
		t.Fatal("SourceProxy handler should return false")
	}
	if rec.Body.String() != targetURL.Host+"|/" {
		t.Fatalf("SourceProxy body = %q", rec.Body.String())
	}

	attrs := gox.NewAttrs()
	src := SourceProxy(targetServer.URL)
	if err := src.Modify(context.Background(), "script", attrs); err != nil {
		t.Fatal(err)
	}
	if got, _ := attrs.Find("src"); got.Value() != src {
		t.Fatalf("SourceProxy.Modify src = %#v", got.Value())
	}
}

func TestSourceBytesAndStringScriptKinds(t *testing.T) {
	bytesSrc := SourceBytes([]byte("console.log('ts')"))
	entry := bytesSrc.scriptEntry(false, true)
	scriptBytes, ok := entry.(resources.ScriptBytes)
	if !ok || scriptBytes.Kind != resources.KindTS {
		t.Fatalf("SourceBytes script entry = %#v", entry)
	}

	stringSrc := SourceString("console.log('inline-ts')")
	inline := stringSrc.scriptEntry(true, true)
	scriptInline, ok := inline.(resources.ScriptInlineString)
	if !ok || scriptInline.Kind != resources.KindTS {
		t.Fatalf("SourceString inline script entry = %#v", inline)
	}

	attrs := gox.NewAttrs()
	if err := stringSrc.Modify(context.Background(), "a", attrs); err != nil {
		t.Fatal(err)
	}
	if got, _ := attrs.Find("href"); got.Value() != stringSrc {
		t.Fatalf("SourceString.Modify href = %#v", got.Value())
	}

	fsys := fstest.MapFS{"asset.txt": {Data: []byte("fs-content")}}
	fsSrc := SourceFS{FS: fsys, Entry: "asset.txt"}
	attrs = gox.NewAttrs()
	if err := fsSrc.Modify(context.Background(), "link", attrs); err != nil {
		t.Fatal(err)
	}
	gotAttr, _ := attrs.Find("href")
	gotSource, ok := gotAttr.Value().(SourceFS)
	if !ok || gotSource.Entry != fsSrc.Entry {
		t.Fatalf("SourceFS.Modify href = %#v", gotAttr.Value())
	}

	if got := sourceNameFromPath(".js", "js"); got != "script.js" {
		t.Fatalf("hidden file default name = %q", got)
	}
}
