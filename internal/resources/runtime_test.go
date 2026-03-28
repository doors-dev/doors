package resources

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/doors-dev/doors/internal/common"
	"github.com/evanw/esbuild/pkg/api"
)

type resourceTestSettings struct {
	conf *common.SystemConf
}

func (s resourceTestSettings) Conf() *common.SystemConf {
	return s.conf
}

func (s resourceTestSettings) BuildProfiles() BuildProfiles {
	return BaseProfile{}
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
	res := NewResource([]byte("hello"), "text/plain", resourceSettings{
		cacheControl: "public, max-age=60",
	})
	if got := string(res.Content()); got != "hello" {
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
	if string(unzipped) != "hello" {
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
	conf := common.SystemConf{}
	common.InitDefaults(&conf)
	rg := NewRegistry(resourceTestSettings{conf: &conf})

	if rg.MainScript() == nil || rg.MainStyle() == nil {
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

	missRec := httptest.NewRecorder()
	rg.Serve("missing", missRec, httptest.NewRequest(http.MethodGet, "/", nil))
	if missRec.Code != http.StatusNotFound {
		t.Fatalf("serve missing status = %d", missRec.Code)
	}
}

func TestRegistryErrorBranches(t *testing.T) {
	conf := common.SystemConf{}
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
