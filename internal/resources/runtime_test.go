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

package resources

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/doors-dev/doors/internal/common"
	"github.com/evanw/esbuild/pkg/api"
)

type resourceTestSettings struct {
	conf *common.Conf
}

func (s resourceTestSettings) Conf() *common.Conf {
	return s.conf
}

func (s resourceTestSettings) ESProfile(string) api.BuildOptions {
	return api.BuildOptions{}
}

func (s resourceTestSettings) Logger() *slog.Logger {
	return slog.Default()
}

type errStaticEntry struct{}

func (errStaticEntry) Read() ([]byte, error) { return nil, context.Canceled }
func (errStaticEntry) entryID(w idWriter)    { _, _ = w.WriteString("err-static") }

type errScriptEntry struct {
	readErr    error
	applyErr   error
	entryPoint string
}

func (e errScriptEntry) Read() ([]byte, error) {
	if e.readErr != nil {
		return nil, e.readErr
	}
	return []byte(`console.log("ok")`), nil
}

func (e errScriptEntry) Apply(opt *api.BuildOptions) error {
	if e.applyErr != nil {
		return e.applyErr
	}
	if e.entryPoint != "" {
		opt.EntryPoints = []string{e.entryPoint}
	}
	return nil
}
func (e errScriptEntry) entryID(w idWriter) { _, _ = w.WriteString("err-script") }

type errStyleEntry struct{}

func (errStyleEntry) Read() ([]byte, error) { return nil, context.Canceled }
func (errStyleEntry) entryID(w idWriter)    { _, _ = w.WriteString("err-style") }

func TestResourceContentAndServeCache(t *testing.T) {
	text := strings.Repeat("hello ", 100)
	res := NewResource([]byte(text), "text/plain", resourceSettings{
		cacheControl: "public, max-age=60",
	})
	if got := string(res.Content()); got != text {
		t.Fatalf("Content() = %q", got)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	res.ServeCache(rec, req, false)

	if got := rec.Header().Get("Content-Type"); got != "text/plain" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q", got)
	}
	reader, err := gzip.NewReader(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	unzipped, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if string(unzipped) != text {
		t.Fatalf("gzipped response = %q", string(unzipped))
	}

	plain := NewResource([]byte("world"), "text/plain", resourceSettings{
		cacheControl: "public, max-age=60",
		disableGzip:  true,
	})
	plainRec := httptest.NewRecorder()
	plain.Serve(plainRec, req)
	if got := plainRec.Header().Get("Cache-Control"); got != "public, max-age=60" {
		t.Fatalf("cached Cache-Control = %q", got)
	}
	if got := plainRec.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("unexpected gzip header for disabled gzip = %q", got)
	}
	if plainRec.Body.String() != "world" {
		t.Fatalf("plain response = %q", plainRec.Body.String())
	}
}

func TestResourceGzipOnlyWhenSmaller(t *testing.T) {
	rng := rand.New(rand.NewChaCha8([32]byte{}))
	noise := make([]byte, 4000)
	for i := range noise {
		noise[i] = byte(rng.Uint32())
	}
	marginal := append(append([]byte{}, noise...), noise[:280]...)
	gz, raw := len(common.Zip(marginal)), len(marginal)
	if gz >= raw || gz < raw-raw/20 {
		t.Fatalf("marginal content zips %d -> %d, want under 5%% saving", raw, gz)
	}
	for name, content := range map[string][]byte{"incompressible": noise, "marginal": marginal} {
		res := NewResource(content, "application/octet-stream", resourceSettings{cacheControl: "public"})
		for range 2 {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rec := httptest.NewRecorder()
			res.Serve(rec, req)
			if got := rec.Header().Get("Content-Encoding"); got != "" {
				t.Fatalf("%s: Content-Encoding = %q", name, got)
			}
			if got := rec.Header().Get("Vary"); got != "" {
				t.Fatalf("%s: Vary = %q", name, got)
			}
			if got := rec.Header().Get("Content-Length"); got != strconv.Itoa(len(content)) {
				t.Fatalf("%s: Content-Length = %q", name, got)
			}
			if !bytes.Equal(rec.Body.Bytes(), content) {
				t.Fatalf("%s: body differs", name)
			}
		}
	}
}

func TestBuildErrorsAndBuild(t *testing.T) {
	errs := BuildErrors{
		{Text: "plain error"},
		{
			Text: "located error",
			Location: &api.Location{
				File:   "app.ts",
				Line:   3,
				Column: 4,
			},
		},
	}
	got := errs.Error()
	if !strings.Contains(got, "plain error") || !strings.Contains(got, "app.ts:3:4: located error") {
		t.Fatalf("BuildErrors.Error() = %q", got)
	}

	opt := api.BuildOptions{
		Stdin: &api.StdinOptions{
			Contents:   `console.log("ok")`,
			Sourcefile: "index.js",
			Loader:     api.LoaderJS,
		},
	}
	data, err := build(&opt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`console.log("ok")`)) && !bytes.Contains(data, []byte(`console.log("ok");`)) {
		t.Fatalf("build output = %q", string(data))
	}

	_, err = build(&api.BuildOptions{})
	if err == nil {
		t.Fatal("expected build error for empty options")
	}
}

func TestDetectLoader(t *testing.T) {
	cases := map[string]api.Loader{
		"app.js":   api.LoaderJS,
		"app.mjs":  api.LoaderJS,
		"app.cjs":  api.LoaderJS,
		"app.ts":   api.LoaderTS,
		"app.mts":  api.LoaderTS,
		"app.cts":  api.LoaderTS,
		"app.tsx":  api.LoaderTSX,
		"app.jsx":  api.LoaderJSX,
		"app.json": api.LoaderJSON,
		"app.css":  api.LoaderCSS,
		"app.txt":  api.LoaderText,
		"app.wasm": api.LoaderBinary,
		"app.bin":  api.LoaderJS,
	}
	for file, expected := range cases {
		if got := detectLoader(file); got != expected {
			t.Fatalf("detectLoader(%q) = %v, want %v", file, got, expected)
		}
	}
}

func TestRegistryServeAndScriptModes(t *testing.T) {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	rg := NewRegistry(resourceTestSettings{conf: &conf})

	if rg.MainScript() == nil {
		t.Fatal("expected main resources to be initialized")
	}

	hostRes, err := rg.Script(ScriptString{Content: `console.log("host")`, Kind: KindJS}, FormatRaw{}, "", ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	hostAgain, err := rg.Script(ScriptString{Content: `console.log("host")`, Kind: KindJS}, FormatRaw{}, "", ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	if hostRes != hostAgain {
		t.Fatal("expected host script resource to be cached")
	}

	noCacheA, err := rg.Script(ScriptString{Content: `console.log("nocache")`, Kind: KindJS}, FormatRaw{}, "", ModeNoCache)
	if err != nil {
		t.Fatal(err)
	}
	noCacheB, err := rg.Script(ScriptString{Content: `console.log("nocache")`, Kind: KindJS}, FormatRaw{}, "", ModeNoCache)
	if err != nil {
		t.Fatal(err)
	}
	if noCacheA == noCacheB {
		t.Fatal("expected nocache script resource to bypass registry cache")
	}

	okRec := httptest.NewRecorder()
	rg.Serve(hostRes.ID(), okRec, httptest.NewRequest(http.MethodGet, "/", nil))
	if okRec.Code != http.StatusOK {
		t.Fatalf("serve existing status = %d", okRec.Code)
	}
	if got := okRec.Header().Get("Content-Type"); got != "application/javascript" {
		t.Fatalf("serve existing content type = %q", got)
	}
	if got := okRec.Body.String(); got != string(hostRes.Content()) {
		t.Fatalf("serve existing body = %q", got)
	}

	missRec := httptest.NewRecorder()
	rg.Serve("missing", missRec, httptest.NewRequest(http.MethodGet, "/", nil))
	if missRec.Code != http.StatusNotFound {
		t.Fatalf("serve missing status = %d", missRec.Code)
	}
}

func TestRegistryErrorBranches(t *testing.T) {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	rg := NewRegistry(resourceTestSettings{conf: &conf})

	if _, err := rg.Static(errStaticEntry{}, "text/plain"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected static read error, got %v", err)
	}
	if _, err := rg.Script(errScriptEntry{readErr: context.Canceled}, FormatRaw{}, "", ModeHost); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected raw script read error, got %v", err)
	}
	if _, err := rg.Script(errScriptEntry{applyErr: context.Canceled}, FormatDefault{}, "", ModeHost); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected script apply error, got %v", err)
	}
	if _, err := rg.Style(errStyleEntry{}, false, ModeHost); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected style read error, got %v", err)
	}
}

func TestRegistryStyleMinifyKey(t *testing.T) {
	entry := StyleString{Content: "a {\n  color: red;\n}\n"}
	for _, minFirst := range []bool{false, true} {
		conf := common.Conf{}
		common.InitDefaults(&conf)
		rg := NewRegistry(resourceTestSettings{conf: &conf})
		order := []bool{false, true}
		if minFirst {
			order = []bool{true, false}
		}
		got := map[bool]*Resource{}
		for _, minify := range order {
			res, err := rg.Style(entry, minify, ModeHost)
			if err != nil {
				t.Fatal(err)
			}
			got[minify] = res
		}
		raw, min := got[false], got[true]
		if raw == min || raw.id == min.id {
			t.Fatalf("minFirst=%v: minified and raw style share a cache entry", minFirst)
		}
		if string(raw.content) != entry.Content {
			t.Fatalf("minFirst=%v: raw content changed: %q", minFirst, raw.content)
		}
		if string(min.content) == entry.Content {
			t.Fatalf("minFirst=%v: minified content not minified: %q", minFirst, min.content)
		}
		again, err := rg.Style(entry, true, ModeHost)
		if err != nil {
			t.Fatal(err)
		}
		if again != min {
			t.Fatalf("minFirst=%v: minified style not cached", minFirst)
		}
	}
}

func TestRegistryStaticContentType(t *testing.T) {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	rg := NewRegistry(resourceTestSettings{conf: &conf})
	a, err := rg.Static(StaticString{Content: "{}"}, "application/json")
	if err != nil {
		t.Fatal(err)
	}
	b, err := rg.Static(StaticString{Content: "{}"}, "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	if a == b || a.ID() == b.ID() {
		t.Fatal("content types share a resource")
	}
	svg, err := rg.Static(StaticBytes{Content: []byte("x")}, "image/svg+xml")
	if err != nil {
		t.Fatal(err)
	}
	css, err := rg.Style(StyleString{Content: "x"}, false, ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	if svg.ID() == css.ID() {
		t.Fatal("static and style ids collide")
	}
	for res, want := range map[*Resource]string{a: "application/json", b: "text/plain", svg: "image/svg+xml", css: "text/css"} {
		rec := httptest.NewRecorder()
		rg.Serve(res.ID(), rec, httptest.NewRequest(http.MethodGet, "/", nil))
		if got := rec.Header().Get("Content-Type"); got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}

func TestRegistryStyleFSContent(t *testing.T) {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	rg := NewRegistry(resourceTestSettings{conf: &conf})
	a, err := rg.Style(StyleFS{FS: fstest.MapFS{"s.css": {Data: []byte("a{color:red}")}}, Path: "s.css"}, false, ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	b, err := rg.Style(StyleFS{FS: fstest.MapFS{"s.css": {Data: []byte("a{color:blue}")}}, Path: "s.css"}, false, ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	if a == b || string(b.content) != "a{color:blue}" {
		t.Fatalf("style fs entries collide: %q", b.content)
	}
}

func TestRegistryEntryIDFields(t *testing.T) {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	rg := NewRegistry(resourceTestSettings{conf: &conf})
	a, err := rg.Static(StaticFS{FS: fstest.MapFS{"a": {Data: []byte("b/c")}}, Path: "a"}, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := rg.Static(StaticFS{FS: fstest.MapFS{"a/b": {Data: []byte("c")}}, Path: "a/b"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if a == b || string(b.content) != "c" {
		t.Fatalf("path and content fields run together: %q", b.content)
	}
	named, err := rg.Static(StaticFS{FS: fstest.MapFS{"x": {Data: []byte("named")}}, Path: "x", Name: "n"}, "")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := rg.Static(StaticFS{FS: fstest.MapFS{"x": {Data: []byte("n")}}, Path: "x"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if named == plain || string(plain.content) != "n" {
		t.Fatalf("name and content fields collide: %q", plain.content)
	}
}

func TestRegistryPrivateThenHost(t *testing.T) {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	rg := NewRegistry(resourceTestSettings{conf: &conf})
	serve := func(res *Resource) int {
		rec := httptest.NewRecorder()
		rg.Serve(res.ID(), rec, httptest.NewRequest(http.MethodGet, "/", nil))
		return rec.Code
	}
	script := ScriptString{Content: "window.x = 1", Kind: KindJS}
	if _, err := rg.Script(script, FormatRaw{}, "", ModeNoHost); err != nil {
		t.Fatal(err)
	}
	res, err := rg.Script(script, FormatRaw{}, "", ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	if code := serve(res); code != http.StatusOK {
		t.Fatalf("script: got %d", code)
	}
	style := StyleString{Content: "a{color:red}"}
	if _, err := rg.Style(style, false, ModeNoHost); err != nil {
		t.Fatal(err)
	}
	res, err = rg.Style(style, false, ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	if code := serve(res); code != http.StatusOK {
		t.Fatalf("style: got %d", code)
	}
}

func TestRegistryFieldBoundary(t *testing.T) {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	rg := NewRegistry(resourceTestSettings{conf: &conf})
	a, err := rg.Static(StaticString{Content: "contentx"}, "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	b, err := rg.Static(StaticString{Content: "x"}, "text/plaincontent")
	if err != nil {
		t.Fatal(err)
	}
	if a == b || a.ID() == b.ID() {
		t.Fatal("field boundary collision")
	}
}

func TestKeysDistinct(t *testing.T) {
	js := ScriptString{Content: "x", Kind: KindJS}
	fsx := func(data string) fstest.MapFS {
		return fstest.MapFS{"x": {Data: []byte(data)}, "y": {Data: []byte(data)}}
	}
	keys := map[string][16]byte{
		"static":                       staticKey(StaticString{Content: "x"}, ""),
		"static typed":                 staticKey(StaticString{Content: "x"}, "text/plain"),
		"static boundary a":            staticKey(StaticString{Content: "contentx"}, "text/plain"),
		"static boundary b":            staticKey(StaticString{Content: "x"}, "text/plaincontent"),
		"static path":                  staticKey(StaticPath{Path: "x"}, ""),
		"static fs path a":             staticKey(StaticFS{FS: fstest.MapFS{"a": {Data: []byte("b/c")}}, Path: "a"}, ""),
		"static fs path a/b":           staticKey(StaticFS{FS: fstest.MapFS{"a/b": {Data: []byte("c")}}, Path: "a/b"}, ""),
		"static fs content":            staticKey(StaticFS{FS: fsx("n"), Path: "x"}, ""),
		"static fs zero content":       staticKey(StaticFS{FS: fsx("\x00n"), Path: "x"}, ""),
		"static fs named":              staticKey(StaticFS{FS: fsx("n"), Path: "x", Name: "n"}, ""),
		"static fs named other":        staticKey(StaticFS{FS: fsx("n"), Path: "x", Name: "m"}, ""),
		"static fs named other path":   staticKey(StaticFS{FS: fsx("n"), Path: "y", Name: "n"}, ""),
		"script default":               scriptKey(js, FormatDefault{}, ""),
		"script profile":               scriptKey(js, FormatDefault{}, "p"),
		"script raw":                   scriptKey(js, FormatRaw{}, ""),
		"script module":                scriptKey(js, FormatModule{}, ""),
		"script module bundle":         scriptKey(js, FormatModule{Bundle: true}, ""),
		"script common":                scriptKey(js, FormatCommon{}, ""),
		"script common bundle":         scriptKey(js, FormatCommon{Bundle: true}, ""),
		"script iife":                  scriptKey(js, FormatIIFE{}, ""),
		"script iife bundle":           scriptKey(js, FormatIIFE{Bundle: true}, ""),
		"script iife named":            scriptKey(js, FormatIIFE{GlobalName: "true"}, ""),
		"script iife bundle name":      scriptKey(js, FormatIIFE{GlobalName: "bundle"}, ""),
		"script ts":                    scriptKey(ScriptString{Content: "x", Kind: KindTS}, FormatDefault{}, ""),
		"script inline":                scriptKey(ScriptInlineString{Content: "x", Kind: KindJS}, FormatDefault{}, ""),
		"script inline ts":             scriptKey(ScriptInlineString{Content: "x", Kind: KindTS}, FormatDefault{}, ""),
		"script path":                  scriptKey(ScriptPath{Path: "x"}, FormatDefault{}, ""),
		"script inline path":           scriptKey(ScriptInlinePath{Path: "x"}, FormatDefault{}, ""),
		"script fs":                    scriptKey(ScriptFS{FS: fsx("x"), Path: "x"}, FormatDefault{}, ""),
		"script fs other content":      scriptKey(ScriptFS{FS: fsx("y"), Path: "x"}, FormatDefault{}, ""),
		"script fs other path":         scriptKey(ScriptFS{FS: fsx("x"), Path: "y"}, FormatDefault{}, ""),
		"script fs named":              scriptKey(ScriptFS{FS: fsx("x"), Path: "x", Name: "n"}, FormatDefault{}, ""),
		"script fs named other":        scriptKey(ScriptFS{FS: fsx("x"), Path: "x", Name: "m"}, FormatDefault{}, ""),
		"script inline fs":             scriptKey(ScriptInlineFS{FS: fsx("x"), Path: "x"}, FormatDefault{}, ""),
		"script inline fs other":       scriptKey(ScriptInlineFS{FS: fsx("y"), Path: "x"}, FormatDefault{}, ""),
		"script inline fs other path":  scriptKey(ScriptInlineFS{FS: fsx("x"), Path: "y"}, FormatDefault{}, ""),
		"script inline fs named":       scriptKey(ScriptInlineFS{FS: fsx("x"), Path: "x", Name: "n"}, FormatDefault{}, ""),
		"script inline fs named other": scriptKey(ScriptInlineFS{FS: fsx("x"), Path: "x", Name: "m"}, FormatDefault{}, ""),
		"style":                        styleKey(StyleString{Content: "x"}, false),
		"style minify":                 styleKey(StyleString{Content: "x"}, true),
		"style path":                   styleKey(StylePath{Path: "x"}, false),
		"style fs content":             styleKey(StyleFS{FS: fsx("n"), Path: "x"}, false),
		"style fs other content":       styleKey(StyleFS{FS: fsx("m"), Path: "x"}, false),
		"style fs other path":          styleKey(StyleFS{FS: fsx("n"), Path: "y"}, false),
		"style fs named":               styleKey(StyleFS{FS: fsx("n"), Path: "x", Name: "n"}, false),
		"style fs named other":         styleKey(StyleFS{FS: fsx("n"), Path: "x", Name: "m"}, false),
	}
	seen := map[[16]byte]string{}
	for name, key := range keys {
		if other, ok := seen[key]; ok {
			t.Errorf("%q and %q share a key", name, other)
		}
		seen[key] = name
	}
	same := [][2][16]byte{
		{staticKey(StaticBytes{Content: []byte("x")}, "t"), staticKey(StaticString{Content: "x"}, "t")},
		{scriptKey(ScriptBytes{Content: []byte("x"), Kind: KindJS}, FormatRaw{}, ""), scriptKey(js, FormatRaw{}, "")},
		{scriptKey(ScriptInlineBytes{Content: []byte("x"), Kind: KindJS}, FormatRaw{}, ""), scriptKey(ScriptInlineString{Content: "x", Kind: KindJS}, FormatRaw{}, "")},
		{styleKey(StyleBytes{Content: []byte("x")}, true), styleKey(StyleString{Content: "x"}, true)},
	}
	for i, pair := range same {
		if pair[0] != pair[1] {
			t.Errorf("equivalent inputs %d get different keys", i)
		}
	}
}

func TestRegistryRawScriptIgnoresProfile(t *testing.T) {
	conf := common.Conf{}
	common.InitDefaults(&conf)
	rg := NewRegistry(resourceTestSettings{conf: &conf})
	path := writeTempFile(t, "raw.js", `console.log("raw")`)
	a, err := rg.Script(ScriptPath{Path: path}, FormatRaw{}, "", ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	b, err := rg.Script(ScriptPath{Path: path}, FormatRaw{}, "p", ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("raw script cached once per profile")
	}
	c, err := rg.Script(ScriptPath{Path: path}, FormatDefault{}, "", ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	d, err := rg.Script(ScriptPath{Path: path}, FormatDefault{}, "p", ModeHost)
	if err != nil {
		t.Fatal(err)
	}
	if c == d {
		t.Fatal("built scripts with different profiles share a cache entry")
	}
}
