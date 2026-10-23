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

package path

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/doors-dev/doors/internal/common"
)

// TestPortableURL builds /w/ URLs with an escaped name.
func TestPortableURL(t *testing.T) {
	pm := NewPathMaker("__Host-", "blue", "")
	for _, c := range []struct {
		name string
		want string
	}{
		{"", "/~/blue/w/tok"},
		{"app.css", "/~/blue/w/tok/app.css"},
		{"my file.css", "/~/blue/w/tok/my%20file.css"},
		{"a/b", "/~/blue/w/tok/a%2Fb"},
	} {
		if got := pm.Portable("tok", c.name); got != c.want {
			t.Fatalf("Portable(%q): got %q, want %q", c.name, got, c.want)
		}
	}
}

// TestPortableRewrite extracts the token and the escaped tail from /w/ URLs.
func TestPortableRewrite(t *testing.T) {
	pm := NewPathMaker("__Host-", "blue", "")
	for _, c := range []struct {
		name   string
		target string
		token  string
		tail   string
	}{
		{"no tail", "/~/blue/w/tok", "tok", ""},
		{"query only", "/~/blue/w/tok?x=1", "tok", ""},
		{"trailing slash", "/~/blue/w/tok/", "tok", "/"},
		{"nested tail", "/~/blue/w/tok/a/b", "tok", "/a/b"},
		{"tail with query", "/~/blue/w/tok/a/b?x=1&y=2", "tok", "/a/b"},
		{"escaped segment", "/~/blue/w/tok/a%2Fb/c", "tok", "/a%2Fb/c"},
		{"double slash tail", "/~/blue/w/tok//users/42", "tok", "//users/42"},
		{"double slash only", "/~/blue/w/tok//", "tok", "//"},
		{"other server segment", "/~/other/w/tok/a", "tok", "/a"},
		{"named", pm.Portable("tok", "my file.css"), "tok", "/my%20file.css"},
	} {
		t.Run(c.name, func(t *testing.T) {
			token, tail, ok := pm.Rewrite(httptest.NewRequest(http.MethodGet, c.target, nil))
			if !ok || token != c.token || tail != c.tail {
				t.Fatalf("Rewrite(%q): got (%q, %q, %v), want (%q, %q, true)", c.target, token, tail, ok, c.token, c.tail)
			}
		})
	}
	for _, target := range []string{
		"/~/blue/h/inst1/20",
		"/~/blue/r/abc",
		"/~/blue",
		"/~/blue/w",
		"/~/blue/w/",
		"/~/blue/w/to-k",
		"/~/blue/wx/tok",
		"/w/tok",
		"/not-system/w/tok",
	} {
		if _, _, ok := pm.Rewrite(httptest.NewRequest(http.MethodGet, target, nil)); ok {
			t.Fatalf("Rewrite(%q): expected no match", target)
		}
	}
	bad := &http.Request{Method: http.MethodGet, URL: &url.URL{Opaque: "/~/blue/w/tok/%zz"}}
	if _, _, ok := pm.Rewrite(bad); ok {
		t.Fatal("expected a bad tail escape to fail rewriting")
	}
}

// TestHookTail joins a hook path with a tail, keeping escapes in RawPath.
func TestHookTail(t *testing.T) {
	pm := NewPathMaker("__Host-", "blue", "")
	for _, c := range []struct {
		tail string
		path string
		raw  string
	}{
		{"", "/~/blue/h/inst1/20", ""},
		{"/", "/~/blue/h/inst1/20/", ""},
		{"/a/b", "/~/blue/h/inst1/20/a/b", ""},
		{"/a%2Fb", "/~/blue/h/inst1/20/a/b", "/~/blue/h/inst1/20/a%2Fb"},
		{"//users", "/~/blue/h/inst1/20//users", ""},
		{"/my%20file.css", "/~/blue/h/inst1/20/my file.css", ""},
		{"/%zz", "/~/blue/h/inst1/20", ""},
	} {
		p, raw := pm.HookTail("inst1", 20, c.tail)
		if p != c.path || raw != c.raw {
			t.Fatalf("HookTail(%q): got (%q, %q), want (%q, %q)", c.tail, p, raw, c.path, c.raw)
		}
	}
}

// TestHookMatchPrefixOverride reports the context hook prefix instead of the /h/ URL.
func TestHookMatchPrefixOverride(t *testing.T) {
	pm := NewPathMaker("__Host-", "blue", "")
	prefix := pm.Portable("tok", "")
	for _, c := range []struct {
		name    string
		tail    string
		query   string
		path    string
		rawPath string
	}{
		{"no tail", "", "", "/", ""},
		{"query", "/a", "x=1", "/a", ""},
		{"escaped tail", "/a%2Fb/c", "", "/a/b/c", "/a%2Fb/c"},
		{"double slash", "//users/42", "", "//users/42", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, prefix+c.tail, nil)
			req.URL.Path, req.URL.RawPath = pm.HookTail("inst1", 20, c.tail)
			req.URL.RawQuery = c.query
			req = req.WithContext(context.WithValue(req.Context(), common.KeyHookPrefix, prefix))
			match, ok := pm.Match(req)
			if !ok {
				t.Fatalf("expected %q to match", req.URL.RequestURI())
			}
			hook, ok := match.Hook()
			if !ok || hook.Instance != "inst1" || hook.Hook != 20 {
				t.Fatalf("unexpected hook match: %#v", hook)
			}
			if hook.Prefix != prefix {
				t.Fatalf("got prefix %q, want %q", hook.Prefix, prefix)
			}
			if hook.Path != c.path || hook.RawPath != c.rawPath {
				t.Fatalf("got path (%q, %q), want (%q, %q)", hook.Path, hook.RawPath, c.path, c.rawPath)
			}
		})
	}
}
