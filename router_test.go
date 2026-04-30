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

package doors

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/doors-dev/gox"
)

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

func testPage(label string) gox.Comp {
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
		if err := cur.Text(label); err != nil {
			return err
		}
		if err := cur.Close(); err != nil {
			return err
		}
		return cur.Close()
	})
}

func TestAppStaticMiddlewareServing(t *testing.T) {
	dir := t.TempDir()
	dirFile := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(dirFile, []byte("dir"), 0o644); err != nil {
		t.Fatal(err)
	}
	singleFile := filepath.Join(dir, "robots.txt")
	if err := os.WriteFile(singleFile, []byte("file"), 0o644); err != nil {
		t.Fatal(err)
	}

	app := NewApp(func(context.Context, Request) gox.Comp {
		return testPage("fallback")
	})
	app.Use(
		UseResource("asset.txt", ResourceBytes([]byte("asset")), "text/plain"),
		UseFS("fs", fstest.MapFS{"hello.txt": &fstest.MapFile{Data: []byte("fs")}}, "public, max-age=60"),
		UseDir("public", dir, "no-cache"),
		UseFile("robots.txt", singleFile, "max-age=120"),
	)

	server := httptest.NewServer(app)
	defer server.Close()

	status, headers, body := readURL(t, server, "/asset.txt")
	if status != http.StatusOK || body != "asset" {
		t.Fatalf("unexpected resource response: status=%d body=%q", status, body)
	}
	if !strings.HasPrefix(headers.Get("Content-Type"), "text/plain") {
		t.Fatalf("unexpected resource content-type: %q", headers.Get("Content-Type"))
	}

	status, headers, body = readURL(t, server, "/fs/hello.txt")
	if status != http.StatusOK || body != "fs" {
		t.Fatalf("unexpected fs response: status=%d body=%q", status, body)
	}
	if headers.Get("Cache-Control") != "public, max-age=60" {
		t.Fatalf("unexpected fs cache-control: %q", headers.Get("Cache-Control"))
	}

	status, headers, body = readURL(t, server, "/public/hello.txt")
	if status != http.StatusOK || body != "dir" {
		t.Fatalf("unexpected dir response: status=%d body=%q", status, body)
	}
	if headers.Get("Cache-Control") != "no-cache" {
		t.Fatalf("unexpected dir cache-control: %q", headers.Get("Cache-Control"))
	}

	status, headers, body = readURL(t, server, "/robots.txt")
	if status != http.StatusOK || body != "file" {
		t.Fatalf("unexpected file response: status=%d body=%q", status, body)
	}
	if headers.Get("Cache-Control") != "max-age=120" {
		t.Fatalf("unexpected file cache-control: %q", headers.Get("Cache-Control"))
	}
}
