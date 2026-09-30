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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/doors-dev/gox"
)

// headBigText is well above the 4 KiB response buffer even when gzipped, so
// the GET response is streamed without a Content-Length.
var headBigText = func() []string {
	rnd := rand.New(rand.NewPCG(1, 1))
	out := make([]string, 1000)
	for i := range out {
		out[i] = fmt.Sprintf("%016x%016x", rnd.Uint64(), rnd.Uint64())
	}
	return out
}()

var headHookRegexp = regexp.MustCompile(`hook:(/~/doors/h/([0-9a-zA-Z]+)/\d+)`)

var headScriptRegexp = regexp.MustCompile(`src="(/~/doors/r/[^"]+)"`)

func headTag(cur gox.Cursor, name string, body func() error) error {
	if err := cur.Init(name); err != nil {
		return err
	}
	if err := cur.Submit(); err != nil {
		return err
	}
	if err := body(); err != nil {
		return err
	}
	return cur.Close()
}

// headPage renders a page with a hook that answers "hook". kind "status" sets
// status 404, "big" appends headBigText, and "err" fails the render.
func headPage(kind string) gox.Comp {
	return gox.Elem(func(cur gox.Cursor) error {
		if kind == "err" {
			return errors.New("boom")
		}
		hook, ok := NewHook(cur.Context(), ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hook"))
		}))
		if !ok {
			return errors.New("hook not registered")
		}
		return headTag(cur, "html", func() error {
			return headTag(cur, "body", func() error {
				if kind == "status" {
					if err := cur.Comp(Status(http.StatusNotFound)); err != nil {
						return err
					}
				}
				if err := cur.Text("hook:" + hook); err != nil {
					return err
				}
				if kind != "big" {
					return nil
				}
				for _, s := range headBigText {
					if err := headTag(cur, "p", func() error { return cur.Text(s) }); err != nil {
						return err
					}
				}
				return nil
			})
		})
	})
}

func newHeadApp(kind string, conf Conf, with ...With) App {
	conf.ServerSessionCookieNoSecure = true
	if conf.InstanceConnectTimeout == 0 {
		conf.InstanceConnectTimeout = time.Minute
	}
	with = append([]With{
		WithConf(conf),
		WithCSP(CSP{}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	}, with...)
	return NewApp(func(context.Context, Request) gox.Comp {
		return headPage(kind)
	}, with...)
}

func headProtos(t *testing.T, f func(t *testing.T, h2 bool)) {
	for _, h2 := range []bool{false, true} {
		name := "h1"
		if h2 {
			name = "h2"
		}
		t.Run(name, func(t *testing.T) { f(t, h2) })
	}
}

func headServer(t *testing.T, h http.Handler, h2 bool) *httptest.Server {
	t.Helper()
	srv := httptest.NewUnstartedServer(h)
	if h2 {
		srv.EnableHTTP2 = true
		srv.StartTLS()
	} else {
		srv.Start()
	}
	t.Cleanup(srv.Close)
	return srv
}

func headClient(t *testing.T, srv *httptest.Server, withJar bool) *http.Client {
	t.Helper()
	c := &http.Client{
		Transport: srv.Client().Transport,
		Timeout:   5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if withJar {
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		c.Jar = jar
	}
	return c
}

func headDo(t *testing.T, c *http.Client, method string, url string, gz bool) (*http.Response, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if gz {
		req.Header.Set("Accept-Encoding", "gzip")
	} else {
		req.Header.Set("Accept-Encoding", "identity")
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s %s: read body: %v", method, url, err)
	}
	return resp, body
}

func headCheckProto(t *testing.T, resp *http.Response, h2 bool) {
	t.Helper()
	want := 1
	if h2 {
		want = 2
	}
	if resp.ProtoMajor != want {
		t.Fatalf("expected HTTP/%d, got %s", want, resp.Proto)
	}
}

// cspDirectives returns the CSP directives sorted, since they are generated
// from maps in no stable order.
func cspDirectives(h http.Header) []string {
	parts := strings.Split(h.Get("Content-Security-Policy"), ";")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	slices.Sort(parts)
	return parts
}

func headHook(t *testing.T, body []byte) (hookURL string, instanceID string) {
	t.Helper()
	m := headHookRegexp.FindSubmatch(body)
	if m == nil {
		t.Fatalf("no hook URL in page body")
	}
	return string(m[1]), string(m[2])
}

func TestHeadPageMatchesGet(t *testing.T) {
	headProtos(t, func(t *testing.T, h2 bool) {
		for _, tc := range []struct {
			kind   string
			status int
		}{
			{"plain", http.StatusOK},
			{"status", http.StatusNotFound},
			{"big", http.StatusOK},
			{"err", http.StatusInternalServerError},
		} {
			kind := tc.kind
			t.Run(kind, func(t *testing.T) {
				app := newHeadApp(kind, Conf{})
				srv := headServer(t, app, h2)
				c := headClient(t, srv, false)
				for _, gz := range []bool{false, true} {
					before := app.InstanceCount()
					head, headBody := headDo(t, c, http.MethodHead, srv.URL+"/", gz)
					headCheckProto(t, head, h2)
					// The HEAD instance must be gone before the response
					// completes, not after the connect timeout.
					if n := app.InstanceCount(); n != before {
						t.Fatalf("gz=%v: instance count %d after HEAD, want %d", gz, n, before)
					}
					get, getBody := headDo(t, c, http.MethodGet, srv.URL+"/", gz)
					if head.StatusCode != tc.status || get.StatusCode != tc.status {
						t.Fatalf("gz=%v: HEAD status %d, GET status %d, want %d", gz, head.StatusCode, get.StatusCode, tc.status)
					}
					if len(headBody) != 0 {
						t.Fatalf("gz=%v: HEAD body has %d bytes", gz, len(headBody))
					}
					if len(getBody) == 0 {
						t.Fatalf("gz=%v: GET body is empty", gz)
					}
					for _, name := range []string{"Content-Type", "Cache-Control", "Content-Encoding", "X-Content-Type-Options"} {
						if head.Header.Get(name) != get.Header.Get(name) {
							t.Fatalf("gz=%v: %s HEAD %q, GET %q", gz, name, head.Header.Get(name), get.Header.Get(name))
						}
					}
					if !slices.Equal(cspDirectives(head.Header), cspDirectives(get.Header)) {
						t.Fatalf("gz=%v: CSP HEAD %q, GET %q", gz, head.Header.Get("Content-Security-Policy"), get.Header.Get("Content-Security-Policy"))
					}
					if head.Header.Get("Set-Cookie") == "" || get.Header.Get("Set-Cookie") == "" {
						t.Fatalf("gz=%v: expected a session cookie on both HEAD and GET", gz)
					}
					if kind == "err" {
						if head.Header.Get("Content-Length") != get.Header.Get("Content-Length") {
							t.Fatalf("gz=%v: error Content-Length HEAD %q, GET %q", gz, head.Header.Get("Content-Length"), get.Header.Get("Content-Length"))
						}
						continue
					}
					if cl := head.Header.Get("Content-Length"); cl != "" {
						t.Fatalf("gz=%v: page HEAD has Content-Length %s", gz, cl)
					}
				}
			})
		}
	})
}

func TestHeadPageWritesNoBody(t *testing.T) {
	// A recorder sees every byte the handler writes, which the HTTP server
	// would silently drop for HEAD. A gzip writer closed with no input still
	// writes its header and footer.
	for _, kind := range []string{"plain", "status", "big"} {
		for _, gz := range []bool{false, true} {
			app := newHeadApp(kind, Conf{})
			req := httptest.NewRequest(http.MethodHead, "/", nil)
			if gz {
				req.Header.Set("Accept-Encoding", "gzip")
			}
			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)
			if rec.Body.Len() != 0 {
				t.Fatalf("%s gz=%v: HEAD handler wrote %d body bytes", kind, gz, rec.Body.Len())
			}
			if got := rec.Header().Get("Content-Encoding") == "gzip"; got != gz {
				t.Fatalf("%s gz=%v: Content-Encoding %q", kind, gz, rec.Header().Get("Content-Encoding"))
			}
			if app.InstanceCount() != 0 {
				t.Fatalf("%s gz=%v: %d instances left after HEAD", kind, gz, app.InstanceCount())
			}
		}
	}
}

func TestHeadKeepsTabs(t *testing.T) {
	headProtos(t, func(t *testing.T, h2 bool) {
		app := newHeadApp("plain", Conf{SessionInstanceLimit: 2})
		srv := headServer(t, app, h2)
		c := headClient(t, srv, true)
		var hooks []string
		for range 2 {
			resp, body := headDo(t, c, http.MethodGet, srv.URL+"/", false)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("GET status %d", resp.StatusCode)
			}
			hook, _ := headHook(t, body)
			hooks = append(hooks, hook)
		}
		sessions, instances := app.SessionCount(), app.InstanceCount()
		for range 5 {
			resp, _ := headDo(t, c, http.MethodHead, srv.URL+"/", true)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("HEAD status %d", resp.StatusCode)
			}
		}
		if n := app.SessionCount(); n != sessions {
			t.Fatalf("session count %d after HEADs, want %d", n, sessions)
		}
		if n := app.InstanceCount(); n != instances {
			t.Fatalf("instance count %d after HEADs, want %d", n, instances)
		}
		// Both tabs are still live: a suspended instance answers its hooks
		// with 410 Gone.
		for _, hook := range hooks {
			resp, body := headDo(t, c, http.MethodGet, srv.URL+hook, true)
			if resp.StatusCode != http.StatusOK || string(body) != "hook" {
				t.Fatalf("hook %s after HEADs: status %d body %q", hook, resp.StatusCode, body)
			}
		}
	})
}

func TestHeadStaticContentLength(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dir.txt"), []byte(strings.Repeat("d", 5000)), 0o644); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte(strings.Repeat("f", 3000)), 0o644); err != nil {
		t.Fatal(err)
	}
	headProtos(t, func(t *testing.T, h2 bool) {
		app := newHeadApp("plain", Conf{})
		app.Use(
			UseFS("fs", fstest.MapFS{"fs.txt": &fstest.MapFile{Data: []byte(strings.Repeat("s", 7000))}}, CacheControlStatic),
			UseDir("dir", dir, CacheControlStatic),
			UseFile("file.txt", file, CacheControlStatic),
			UseResource("res.js", ResourceString(strings.Repeat("console.log(1);", 400)), "application/javascript"),
		)
		srv := headServer(t, app, h2)
		c := headClient(t, srv, false)
		_, page := headDo(t, c, http.MethodGet, srv.URL+"/", false)
		script := headScriptRegexp.FindSubmatch(page)
		if script == nil {
			t.Fatal("no registry resource in page body")
		}
		for _, path := range []string{"/fs/fs.txt", "/dir/dir.txt", "/file.txt", "/res.js", string(script[1])} {
			for _, gz := range []bool{false, true} {
				sessions := app.SessionCount()
				head, headBody := headDo(t, c, http.MethodHead, srv.URL+path, gz)
				headCheckProto(t, head, h2)
				get, getBody := headDo(t, c, http.MethodGet, srv.URL+path, gz)
				if head.StatusCode != http.StatusOK || get.StatusCode != http.StatusOK {
					t.Fatalf("%s gz=%v: HEAD status %d, GET status %d", path, gz, head.StatusCode, get.StatusCode)
				}
				if len(headBody) != 0 {
					t.Fatalf("%s gz=%v: HEAD body has %d bytes", path, gz, len(headBody))
				}
				if head.Header.Get("Content-Length") != strconv.Itoa(len(getBody)) {
					t.Fatalf("%s gz=%v: HEAD Content-Length %q, GET body %d bytes", path, gz, head.Header.Get("Content-Length"), len(getBody))
				}
				for _, name := range []string{"Content-Type", "Cache-Control", "Content-Encoding"} {
					if head.Header.Get(name) != get.Header.Get(name) {
						t.Fatalf("%s gz=%v: %s HEAD %q, GET %q", path, gz, name, head.Header.Get(name), get.Header.Get(name))
					}
				}
				// Falling through to the page would render it and store a
				// new session.
				if n := app.SessionCount(); n != sessions {
					t.Fatalf("%s gz=%v: session count %d, want %d", path, gz, n, sessions)
				}
			}
		}
	})
}

func TestHeadSystemPaths(t *testing.T) {
	headProtos(t, func(t *testing.T, h2 bool) {
		app := newHeadApp("plain", Conf{})
		srv := headServer(t, app, h2)
		tab := headClient(t, srv, true)
		anon := headClient(t, srv, false)
		_, body := headDo(t, tab, http.MethodGet, srv.URL+"/", false)
		hook, instanceID := headHook(t, body)
		sessions := app.SessionCount()
		for _, tc := range []struct {
			path  string
			allow string
		}{
			{"/~/doors/s/" + instanceID + "?b=other", "GET, POST"},
			{"/~/doors/u/" + instanceID + "/other", "GET"},
		} {
			for _, c := range []*http.Client{tab, anon} {
				for _, method := range []string{http.MethodHead, http.MethodPut} {
					resp, _ := headDo(t, c, method, srv.URL+tc.path, true)
					if resp.StatusCode != http.StatusMethodNotAllowed {
						t.Fatalf("%s %s: status %d, want 405", method, tc.path, resp.StatusCode)
					}
					if resp.Header.Get("Allow") != tc.allow {
						t.Fatalf("%s %s: Allow %q, want %q", method, tc.path, resp.Header.Get("Allow"), tc.allow)
					}
				}
			}
		}
		// The method guard runs before the session is resolved, so the
		// anonymous requests create none.
		if n := app.SessionCount(); n != sessions {
			t.Fatalf("session count %d after rejected requests, want %d", n, sessions)
		}
		resp, _ := headDo(t, tab, http.MethodGet, srv.URL+hook, true)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("hook after rejected requests: status %d", resp.StatusCode)
		}
		resp, _ = headDo(t, tab, http.MethodPost, srv.URL+"/", true)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("POST page: status %d, want 404", resp.StatusCode)
		}
	})
}

func TestHeadErrorPage(t *testing.T) {
	errorPage := WithErrorPage(func(r *http.Request, err error) gox.Elem {
		return func(cur gox.Cursor) error { return cur.Text("error page") }
	})
	headProtos(t, func(t *testing.T, h2 bool) {
		app := newHeadApp("err", Conf{}, errorPage)
		srv := headServer(t, app, h2)
		c := headClient(t, srv, false)
		head, headBody := headDo(t, c, http.MethodHead, srv.URL+"/", true)
		headCheckProto(t, head, h2)
		get, getBody := headDo(t, c, http.MethodGet, srv.URL+"/", true)
		if head.StatusCode != http.StatusInternalServerError || get.StatusCode != http.StatusInternalServerError {
			t.Fatalf("HEAD status %d, GET status %d, want 500", head.StatusCode, get.StatusCode)
		}
		if len(headBody) != 0 {
			t.Fatalf("HEAD body has %d bytes", len(headBody))
		}
		if string(getBody) != "error page" {
			t.Fatalf("GET body %q, want the error page", getBody)
		}
	})
	app := newHeadApp("err", Conf{}, errorPage)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodHead, "/", nil))
	if rec.Code != http.StatusInternalServerError || rec.Body.Len() != 0 {
		t.Fatalf("HEAD handler: status %d, %d body bytes", rec.Code, rec.Body.Len())
	}
}
