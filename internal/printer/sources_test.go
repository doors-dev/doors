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

package printer

import (
	"bytes"
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
	inline, ok := src.scriptEntry(true, false).(resources.ScriptInlineFS)
	if !ok || inline.Path != "asset.txt" {
		t.Fatalf("SourceFS inline script entry = %#v", inline)
	}
	if data, err := inline.Read(); err != nil || string(data) != "fs-content" {
		t.Fatalf("SourceFS inline script entry read = %q %v", string(data), err)
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
	inline, ok := src.scriptEntry(true, false).(resources.ScriptInlinePath)
	if !ok || inline.Path != path {
		t.Fatalf("SourceLocalFS inline script entry = %#v", inline)
	}
	if data, err := inline.Read(); err != nil || string(data) != "path-content" {
		t.Fatalf("SourceLocalFS inline script entry read = %q %v", string(data), err)
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

	external := SourceExternal("https://example.com/app.js")
	attrs := gox.NewAttrs()
	if err := external.Modify(context.Background(), "link", attrs); err != nil {
		t.Fatal(err)
	}
	if got, _ := attrs.Find("href"); got.Value() != external {
		t.Fatalf("SourceExternal.Modify href = %#v", got.Value())
	}
	var out bytes.Buffer
	if err := external.Output(&out); err != nil {
		t.Fatal(err)
	}
	if out.String() != string(external) {
		t.Fatalf("SourceExternal.Output = %q", out.String())
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

func TestModifySourceHelpers(t *testing.T) {
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
	if !ok || scriptBytes.Kind != resources.KindTS || string(scriptBytes.Content) != "console.log('ts')" {
		t.Fatalf("SourceBytes script entry = %#v", entry)
	}

	stringSrc := SourceString("console.log('inline-ts')")
	inline := stringSrc.scriptEntry(true, true)
	scriptInline, ok := inline.(resources.ScriptInlineString)
	if !ok || scriptInline.Kind != resources.KindTS || scriptInline.Content != "console.log('inline-ts')" {
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
}
