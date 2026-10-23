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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/gox"
)

type portableKey struct{}

var portableRegexp = regexp.MustCompile(`id="(\w+)" src="(/~/[^/"]+/w/[0-9a-zA-Z]+[^"]*)"`)

func portableImg(cur gox.Cursor, id string, src any, name string) error {
	if err := cur.InitVoid("img"); err != nil {
		return err
	}
	if err := cur.Set("id", id); err != nil {
		return err
	}
	if err := cur.Set("src", src); err != nil {
		return err
	}
	if err := cur.Set("portable", true); err != nil {
		return err
	}
	if name != "" {
		if err := cur.Set("name", name); err != nil {
			return err
		}
	}
	return cur.Submit()
}

// newPortableApp renders portable handler images "mux", "named" and "once"
// next to a plain hook.
func newPortableApp() App {
	return NewApp(func(context.Context, Request) gox.Comp {
		return gox.Elem(func(cur gox.Cursor) error {
			mux := hookPathMux()
			mux.HandleFunc("GET /value", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, r.Context().Value(portableKey{}))
			})
			hook, ok := NewHook(cur.Context(), ResourceHandler(mux.ServeHTTP))
			if !ok {
				return errors.New("hook not registered")
			}
			return headTag(cur, "html", func() error {
				return headTag(cur, "body", func() error {
					if err := portableImg(cur, "mux", ResourceHandler(mux.ServeHTTP), ""); err != nil {
						return err
					}
					if err := portableImg(cur, "named", ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprintf(w, "%s|%s", r.URL.Path, r.Header.Get("Doors-Hook-Prefix"))
					}), "pic.png"); err != nil {
						return err
					}
					if err := portableImg(cur, "once", ResourceHook(func(_ context.Context, w http.ResponseWriter, _ *http.Request) bool {
						w.Write([]byte("once"))
						return true
					}), ""); err != nil {
						return err
					}
					return cur.Text("hook:" + hook)
				})
			})
		})
	},
		WithConf(Conf{ServerSessionCookieNoSecure: true, InstanceConnectTimeout: time.Minute}),
		WithCSP(CSP{}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
		WithValue(portableKey{}, "app-value"),
	)
}

func portableSetup(t *testing.T) (App, *http.Client, string, map[string]string, string) {
	t.Helper()
	app := newPortableApp()
	srv := headServer(t, app, false)
	c := headClient(t, srv, false)
	resp, body := headDo(t, c, http.MethodGet, srv.URL+"/", false)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("page status %d", resp.StatusCode)
	}
	urls := map[string]string{}
	for _, m := range portableRegexp.FindAllSubmatch(body, -1) {
		urls[string(m[1])] = string(m[2])
	}
	for _, id := range []string{"mux", "named", "once"} {
		if urls[id] == "" {
			t.Fatalf("no portable URL for %q in page body %q", id, body)
		}
	}
	hook, _ := headHook(t, body)
	return app, c, srv.URL, urls, hook
}

func portableDo(t *testing.T, c *http.Client, method string, url string, header http.Header) (*http.Response, string) {
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
	if cookies := resp.Header.Values("Set-Cookie"); len(cookies) != 0 {
		t.Fatalf("%s %s: unexpected Set-Cookie %q", method, url, cookies)
	}
	return resp, string(body)
}

// TestPortableNoCookie serves portable handler routes without a session cookie.
func TestPortableNoCookie(t *testing.T) {
	app, c, base, urls, _ := portableSetup(t)
	prefix := urls["mux"]
	sessions := app.SessionCount()
	for _, tc := range []struct {
		method string
		uri    string
		want   string
	}{
		{http.MethodGet, "", "root|/|/||||"},
		{http.MethodGet, "/", "root|/|/||||"},
		{http.MethodGet, "/users/42?t=7&x=y", "user|/users/42|/users/42|42|7|y|"},
		{http.MethodGet, "/files/a%2Fb", "file|/files/a/b|/files/a%2Fb||||"},
		{http.MethodPost, "/users", "create|/users|/users||||"},
	} {
		uri := prefix + tc.uri
		resp, got := portableDo(t, c, tc.method, base+uri, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s %s: status %d, body %q", tc.method, uri, resp.StatusCode, got)
		}
		if want := tc.want + uri; got != want {
			t.Fatalf("%s %s: got %q, want %q", tc.method, uri, got, want)
		}
	}
	if n := app.SessionCount(); n != sessions {
		t.Fatalf("session count %d, want %d", n, sessions)
	}
}

// TestPortablePrefix gives the handler the /w/ URL in Doors-Hook-Prefix, replacing a client value.
func TestPortablePrefix(t *testing.T) {
	_, c, base, urls, _ := portableSetup(t)
	prefix := urls["mux"]
	resp, got := portableDo(t, c, http.MethodGet, base+prefix+"/prefix?x=1", http.Header{"Doors-Hook-Prefix": {"/spoofed"}})
	if resp.StatusCode != http.StatusOK || got != prefix {
		t.Fatalf("status %d, got prefix %q, want %q", resp.StatusCode, got, prefix)
	}
	named := urls["named"]
	if !strings.HasSuffix(named, "/pic.png") {
		t.Fatalf("named URL %q has no name tail", named)
	}
	resp, got = portableDo(t, c, http.MethodGet, base+named, nil)
	if want := "/pic.png|" + strings.TrimSuffix(named, "/pic.png"); resp.StatusCode != http.StatusOK || got != want {
		t.Fatalf("status %d, got %q, want %q", resp.StatusCode, got, want)
	}
}

// TestPortableValue passes WithValue values to the handler request context.
func TestPortableValue(t *testing.T) {
	_, c, base, urls, _ := portableSetup(t)
	resp, got := portableDo(t, c, http.MethodGet, base+urls["mux"]+"/value", nil)
	if resp.StatusCode != http.StatusOK || got != "app-value" {
		t.Fatalf("status %d, got %q", resp.StatusCode, got)
	}
}

// TestPortableGone answers 404 for an unknown token and 410 for a non-portable hook without a cookie.
func TestPortableGone(t *testing.T) {
	_, c, base, urls, hook := portableSetup(t)
	unknown := urls["mux"][:strings.LastIndex(urls["mux"], "/")+1] + "unknown"
	for _, uri := range []string{unknown, unknown + "/a/b"} {
		resp, got := portableDo(t, c, http.MethodGet, base+uri, nil)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("%s: status %d, body %q, want 404", uri, resp.StatusCode, got)
		}
	}
	resp, got := headDo(t, c, http.MethodGet, base+hook, false)
	if resp.StatusCode != http.StatusGone {
		t.Fatalf("%s: status %d, body %q, want 410", hook, resp.StatusCode, got)
	}
}

// TestPortableOnce answers 404 for a one-shot portable hook after its first request.
func TestPortableOnce(t *testing.T) {
	_, c, base, urls, _ := portableSetup(t)
	resp, got := portableDo(t, c, http.MethodGet, base+urls["once"], nil)
	if resp.StatusCode != http.StatusOK || got != "once" {
		t.Fatalf("first: status %d, body %q", resp.StatusCode, got)
	}
	resp, got = portableDo(t, c, http.MethodGet, base+urls["once"], nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("second: status %d, body %q, want 404", resp.StatusCode, got)
	}
}
