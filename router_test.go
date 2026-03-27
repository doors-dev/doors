package doors

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

type testSessionCallback struct{}

func (testSessionCallback) Create(string, http.Header) {}

func (testSessionCallback) Delete(string) {}

func readURL(t *testing.T, server *httptest.Server, path string) (int, http.Header, string) {
	t.Helper()
	resp, err := http.Get(server.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, resp.Header, string(body)
}

func TestRouteFileServing(t *testing.T) {
	dir := t.TempDir()
	dirFile := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(dirFile, []byte("dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	singleFile := filepath.Join(dir, "robots.txt")
	if err := os.WriteFile(singleFile, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}

	router := NewRouter()
	UseRoute(router, RouteFS{
		Prefix:       "assets",
		FS:           fstest.MapFS{"hello.txt": &fstest.MapFile{Data: []byte("asset")}},
		CacheControl: "public, max-age=60",
	})
	UseRoute(router, RouteDir{
		Prefix:       "public",
		DirPath:      dir,
		CacheControl: "no-cache",
	})
	UseRoute(router, RouteFile{
		Path:         "robots.txt",
		FilePath:     singleFile,
		CacheControl: "max-age=120",
	})

	server := httptest.NewServer(router)
	defer server.Close()

	status, headers, body := readURL(t, server, "/assets/hello.txt")
	if status != http.StatusOK {
		t.Fatalf("unexpected assets status: %d", status)
	}
	if body != "asset" {
		t.Fatalf("unexpected assets body: %q", body)
	}
	if headers.Get("Cache-Control") != "public, max-age=60" {
		t.Fatalf("unexpected assets cache-control: %q", headers.Get("Cache-Control"))
	}

	status, headers, body = readURL(t, server, "/public/hello.txt")
	if status != http.StatusOK {
		t.Fatalf("unexpected dir status: %d", status)
	}
	if body != "dir" {
		t.Fatalf("unexpected dir body: %q", body)
	}
	if headers.Get("Cache-Control") != "no-cache" {
		t.Fatalf("unexpected dir cache-control: %q", headers.Get("Cache-Control"))
	}

	status, headers, body = readURL(t, server, "/robots.txt")
	if status != http.StatusOK {
		t.Fatalf("unexpected file status: %d", status)
	}
	if body != "file" {
		t.Fatalf("unexpected file body: %q", body)
	}
	if headers.Get("Cache-Control") != "max-age=120" {
		t.Fatalf("unexpected file cache-control: %q", headers.Get("Cache-Control"))
	}
}

func TestRouteResourceAndFallback(t *testing.T) {
	router := NewRouter()
	UseRoute(router, RouteResource{
		Path:        "/hello.txt",
		Resource:    ResourceBytes([]byte("hello")),
		ContentType: "text/plain",
	})
	UseFallback(router, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("fallback"))
	}))
	UseSessionCallback(router, testSessionCallback{})
	UseESConf(router, ESOptions{JSX: JSXReact(), Minify: false})
	UseSystemConf(router, SystemConf{})
	UseCSP(router, CSP{})
	UseLicense(router, "invalid-cert")
	UseServerID(router, "blue")

	server := httptest.NewServer(router)
	defer server.Close()

	status, headers, body := readURL(t, server, "/hello.txt")
	if status != http.StatusOK {
		t.Fatalf("unexpected resource status: %d", status)
	}
	if body != "hello" {
		t.Fatalf("unexpected resource body: %q", body)
	}
	if !strings.HasPrefix(headers.Get("Content-Type"), "text/plain") {
		t.Fatalf("unexpected resource content-type: %q", headers.Get("Content-Type"))
	}

	status, _, body = readURL(t, server, "/missing")
	if status != http.StatusTeapot {
		t.Fatalf("unexpected fallback status: %d", status)
	}
	if body != "fallback" {
		t.Fatalf("unexpected fallback body: %q", body)
	}
}

func TestRouteMatchRejectsRoot(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if (RouteResource{Path: "/"}).Match(req) {
		t.Fatal("expected resource route to reject root path")
	}
	if (RouteFS{Prefix: "/"}).Match(req) {
		t.Fatal("expected fs route to reject root prefix")
	}
	if (RouteDir{Prefix: ""}).Match(req) {
		t.Fatal("expected dir route to reject empty prefix")
	}
	if (RouteFile{Path: ""}).Match(req) {
		t.Fatal("expected file route to reject empty path")
	}
}
