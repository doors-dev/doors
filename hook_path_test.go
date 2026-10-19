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
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/gox"
)

// hookPathMux answers each route with the route name and what the handler
// sees of the request.
func hookPathMux() *http.ServeMux {
	mux := http.NewServeMux()
	reply := func(route string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%s|%s|%s|%s|%s|%s|%s", route, r.URL.Path, r.URL.EscapedPath(), r.PathValue("id"), r.URL.Query().Get("t"), r.URL.Query().Get("x"), r.RequestURI)
		}
	}
	mux.HandleFunc("GET /{$}", reply("root"))
	mux.HandleFunc("GET /users/{id}", reply("user"))
	mux.HandleFunc("POST /users", reply("create"))
	mux.HandleFunc("GET /files/{name}", reply("file"))
	mux.HandleFunc("GET /prefix", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(r.Header.Get("Doors-Hook-Prefix")))
	})
	return mux
}

func newHookPathApp(with ...With) App {
	with = append([]With{
		WithConf(Conf{ServerSessionCookieNoSecure: true, InstanceConnectTimeout: time.Minute}),
		WithCSP(CSP{}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	}, with...)
	return NewApp(func(context.Context, Request) gox.Comp {
		return gox.Elem(func(cur gox.Cursor) error {
			hook, ok := NewHook(cur.Context(), ResourceHandler(hookPathMux().ServeHTTP))
			if !ok {
				return errors.New("hook not registered")
			}
			return headTag(cur, "html", func() error {
				return headTag(cur, "body", func() error {
					return cur.Text("hook:" + hook)
				})
			})
		})
	}, with...)
}

func hookPathDo(t *testing.T, c *http.Client, method string, url string, header http.Header) string {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range header {
		req.Header[k] = v
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s %s: status %d, body %q", method, url, resp.StatusCode, body)
	}
	return string(body)
}

func hookPathSetup(t *testing.T, app App) (*http.Client, string, string) {
	t.Helper()
	srv := headServer(t, app, false)
	c := headClient(t, srv, true)
	_, body := headDo(t, c, http.MethodGet, srv.URL+"/", false)
	hook, _ := headHook(t, body)
	return c, srv.URL, hook
}

// TestHookPathRoutes serves sub-paths of a hook with an http.ServeMux.
func TestHookPathRoutes(t *testing.T) {
	c, base, hook := hookPathSetup(t, newHookPathApp())
	for _, tc := range []struct {
		method string
		uri    string
		want   string
	}{
		{http.MethodGet, hook, "root|/|/||||"},
		{http.MethodGet, hook + "/", "root|/|/||||"},
		{http.MethodGet, hook + "/users/42", "user|/users/42|/users/42|42|||"},
		{http.MethodGet, hook + "/users/42?t=7&x=y", "user|/users/42|/users/42|42|7|y|"},
		{http.MethodGet, hook + "?t=3", "root|/|/||3||"},
		{http.MethodPost, hook + "/users", "create|/users|/users||||"},
	} {
		got := hookPathDo(t, c, tc.method, base+tc.uri, nil)
		if want := tc.want + tc.uri; got != want {
			t.Fatalf("%s %s: got %q, want %q", tc.method, tc.uri, got, want)
		}
	}
}

// TestHookPathEscaped keeps an escaped slash in the hook tail as RawPath.
func TestHookPathEscaped(t *testing.T) {
	c, base, hook := hookPathSetup(t, newHookPathApp())
	uri := hook + "/files/a%2Fb"
	got := hookPathDo(t, c, http.MethodGet, base+uri, nil)
	if want := "file|/files/a/b|/files/a%2Fb||||" + uri; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestHookPathCallerRequest leaves the request seen by App.Use middleware unchanged.
func TestHookPathCallerRequest(t *testing.T) {
	paths := make(chan string, 16)
	app := newHookPathApp()
	app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			if strings.Contains(r.URL.Path, "/h/") {
				paths <- r.URL.Path
			}
		})
	})
	c, base, hook := hookPathSetup(t, app)
	hookPathDo(t, c, http.MethodGet, base+hook+"/users/42", nil)
	select {
	case got := <-paths:
		if want := hook + "/users/42"; got != want {
			t.Fatalf("middleware saw %q, want %q", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("middleware did not see the hook request")
	}
}

// TestHookPathPrefix gives the handler the hook URL in Doors-Hook-Prefix, replacing a client value.
func TestHookPathPrefix(t *testing.T) {
	c, base, hook := hookPathSetup(t, newHookPathApp())
	got := hookPathDo(t, c, http.MethodGet, base+hook+"/prefix?x=1", http.Header{"Doors-Hook-Prefix": {"/spoofed"}})
	if got != hook {
		t.Fatalf("got prefix %q, want %q", got, hook)
	}
}
