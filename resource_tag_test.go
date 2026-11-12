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
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/gox"
)

type tagAttr struct {
	name  string
	value any
}

func stageTag(cur gox.Cursor, tag string, void bool, mod gox.Modify, attrs ...tagAttr) error {
	init := cur.Init
	if void {
		init = cur.InitVoid
	}
	if err := init(tag); err != nil {
		return err
	}
	for _, a := range attrs {
		if err := cur.Set(a.name, a.value); err != nil {
			return err
		}
	}
	if mod != nil {
		if err := cur.Modify(mod); err != nil {
			return err
		}
	}
	if err := cur.Submit(); err != nil {
		return err
	}
	if void {
		return nil
	}
	return cur.Close()
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func stageApp(logs *syncBuffer, body func(cur gox.Cursor) error) App {
	return NewApp(func(context.Context, Request) gox.Comp {
		return gox.Elem(func(cur gox.Cursor) error {
			return headTag(cur, "html", func() error {
				return headTag(cur, "body", func() error {
					return body(cur)
				})
			})
		})
	},
		WithConf(Conf{ServerSessionCookieNoSecure: true, InstanceConnectTimeout: time.Minute}),
		WithLogger(slog.New(slog.NewTextHandler(logs, nil))),
	)
}

func stagePage(t *testing.T, body func(cur gox.Cursor) error) (*http.Client, string, string, string) {
	t.Helper()
	logs := &syncBuffer{}
	srv := headServer(t, stageApp(logs, body), false)
	c := headClient(t, srv, true)
	_, page := headDo(t, c, http.MethodGet, srv.URL+"/", false)
	return c, srv.URL, string(page), logs.String()
}

func stageAttr(t *testing.T, page string, id string, name string) string {
	t.Helper()
	re := regexp.MustCompile(`<[^>]*id="` + id + `"[^>]*>`)
	tag := re.FindString(page)
	if tag == "" {
		t.Fatalf("no tag %q in page %q", id, page)
	}
	m := regexp.MustCompile(`\s` + name + `="([^"]*)"`).FindStringSubmatch(tag)
	if m == nil {
		t.Fatalf("tag %q has no %s: %q", id, name, tag)
	}
	return m[1]
}

func stageWant(t *testing.T, page string, id string) string {
	t.Helper()
	m := regexp.MustCompile(`want:` + id + `=([^;]*);`).FindStringSubmatch(page)
	if m == nil {
		t.Fatalf("no expected URL for %q in page %q", id, page)
	}
	return m[1]
}

func stageWrite(cur gox.Cursor, id string, r ServedResource) error {
	url, err := r.URL(cur.Context())
	if err != nil {
		return err
	}
	return cur.Text(fmt.Sprintf("want:%s=%s;", id, url))
}

// TestResourceStageTags renders configured resources as src, href and
// modifiers and checks each tag gets the URL that URL(ctx) gives.
func TestResourceStageTags(t *testing.T) {
	js := ResourceString("window.a = 1 + 1").Script()
	mod := ResourceString("export const a = 1").Script().Module()
	css := ResourceString("body {  color: red;  }").Style()
	logo := ResourceString("logo").ContentType("image/svg+xml").Cache()
	pre := ResourceString("export const b = 2").Script().Module()
	classic := ResourceString("export const c = 3").Script()
	stages := []struct {
		id string
		r  ServedResource
	}{{"js", js}, {"mod", mod}, {"css", css}, {"logo", logo}, {"pre", pre}, {"named", js.Name("app.js")}, {"typed", classic.Module()}}
	c, base, page, logs := stagePage(t, func(cur gox.Cursor) error {
		for _, s := range stages {
			if err := stageWrite(cur, s.id, s.r); err != nil {
				return err
			}
		}
		if err := stageTag(cur, "script", false, nil, tagAttr{"id", "js"}, tagAttr{"src", js}); err != nil {
			return err
		}
		if err := stageTag(cur, "script", false, mod, tagAttr{"id", "mod"}, tagAttr{"specifier", "app"}); err != nil {
			return err
		}
		if err := stageTag(cur, "link", true, nil, tagAttr{"id", "css"}, tagAttr{"rel", "stylesheet"}, tagAttr{"href", css}); err != nil {
			return err
		}
		if err := stageTag(cur, "link", true, nil, tagAttr{"id", "pre"}, tagAttr{"rel", "modulepreload"}, tagAttr{"href", pre}); err != nil {
			return err
		}
		if err := stageTag(cur, "link", true, nil, tagAttr{"id", "typed"}, tagAttr{"rel", "modulepreload"}, tagAttr{"href", classic}, tagAttr{"type", "module"}); err != nil {
			return err
		}
		if err := stageTag(cur, "script", false, nil, tagAttr{"id", "named"}, tagAttr{"src", js}, tagAttr{"name", "app.js"}); err != nil {
			return err
		}
		return stageTag(cur, "img", true, nil, tagAttr{"id", "logo"}, tagAttr{"src", logo})
	})
	served := map[string]struct{ attr, body, ct string }{
		"pre":   {"href", "export", "javascript"},
		"named": {"src", "window.a=2", "javascript"},
		"typed": {"href", "export", "javascript"},
		"js":    {"src", "window.a=2", "javascript"},
		"mod":   {"src", "export", "javascript"},
		"css":   {"href", "color", "text/css"},
		"logo":  {"src", "logo", "image/svg+xml"},
	}
	for _, s := range stages {
		want := served[s.id]
		got := stageAttr(t, page, s.id, want.attr)
		if expected := stageWant(t, page, s.id); got != expected {
			t.Fatalf("%s: %s %q, URL(ctx) %q; logs %s", s.id, want.attr, got, expected, logs)
		}
		resp, body := headDo(t, c, http.MethodGet, base+got, false)
		if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), want.body) {
			t.Fatalf("%s: status %d, body %q", s.id, resp.StatusCode, body)
		}
		if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, want.ct) {
			t.Fatalf("%s: content type %q", s.id, ct)
		}
	}
	if typ := stageAttr(t, page, "mod", "type"); typ != "module" {
		t.Fatalf("module script type %q", typ)
	}
	if !strings.Contains(page, `"app":"`+stageWant(t, page, "mod")+`"`) {
		t.Fatalf("specifier not mapped to the module URL in page %q", page)
	}
	if named := stageAttr(t, page, "named", "src"); !strings.HasSuffix(named, "app.js") {
		t.Fatalf("named src %q", named)
	}
}

// TestResourceStageHandler renders handler-backed resources with settings on
// script, stylesheet and generic tags.
func TestResourceStageHandler(t *testing.T) {
	hook := func(body string) ResourceDynamic {
		return ResourceHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(body))
		})
	}
	tags := []struct {
		id, tag, attr string
		void          bool
		r             ServedResource
		rel           string
	}{
		{"hjs", "script", "src", false, hook("hook-js").Portable(), ""},
		{"hcss", "link", "href", true, hook("hook-css").Name("x.css"), "stylesheet"},
		{"hpre", "link", "href", true, hook("hook-pre").Name("x.js"), "modulepreload"},
		{"himg", "img", "src", true, hook("hook-img").Name("x.png").Portable(), ""},
	}
	c, base, page, logs := stagePage(t, func(cur gox.Cursor) error {
		for _, tg := range tags {
			attrs := []tagAttr{{"id", tg.id}, {tg.attr, tg.r}}
			if tg.rel != "" {
				attrs = append(attrs, tagAttr{"rel", tg.rel})
			}
			if err := stageTag(cur, tg.tag, tg.void, nil, attrs...); err != nil {
				return err
			}
		}
		return nil
	})
	names := map[string]string{"hcss": "/x.css", "hpre": "/x.js", "himg": "/x.png"}
	for _, tg := range tags {
		got := stageAttr(t, page, tg.id, tg.attr)
		if !strings.HasSuffix(got, names[tg.id]) {
			t.Fatalf("%s: %s %q; logs %s", tg.id, tg.attr, got, logs)
		}
		resp, body := headDo(t, c, http.MethodGet, base+got, false)
		if resp.StatusCode != http.StatusOK || string(body) != "hook-"+strings.TrimPrefix(tg.id, "h") {
			t.Fatalf("%s: status %d, body %q", tg.id, resp.StatusCode, body)
		}
	}
}

// TestResourceStageErrors renders configured resources on the wrong tag or
// with conflicting tag attributes.
func TestResourceStageErrors(t *testing.T) {
	cases := []struct {
		name string
		tag  string
		void bool
		attr []tagAttr
		want string
	}{
		{"script on img", "img", true, []tagAttr{{"src", ResourceString("a").Script()}}, "script resource used on <img>"},
		{"style on script", "script", false, []tagAttr{{"src", ResourceString("a").Style()}}, "style resource used on <script>"},
		{"mode conflict", "script", false, []tagAttr{{"src", ResourceString("a").Script().NoCache()}, {"cache", true}}, "only one of cache, private, or nocache may be set"},
		{"style on modulepreload", "link", true, []tagAttr{{"rel", "modulepreload"}, {"href", ResourceString("a").Style()}}, "style resource used on <link rel=modulepreload>"},
		{"script on stylesheet", "link", true, []tagAttr{{"rel", "stylesheet"}, {"href", ResourceString("a").Script()}}, "script resource used on <link rel=stylesheet>"},
		{"style on link", "link", true, []tagAttr{{"href", ResourceString("a").Style()}}, "style resource used on <link>"},
		{"typed asset on script", "script", false, []tagAttr{{"src", ResourceString("a").ContentType("text/plain")}}, "content type is not supported on <script>"},
		{"classic modulepreload", "link", true, []tagAttr{{"rel", "modulepreload"}, {"href", ResourceString("a").Script()}}, "modulepreload links require a module script resource"},
		{"output conflict", "script", false, []tagAttr{{"src", ResourceString("a").Script().Bundle()}, {"raw", true}}, "only one of raw, inline, or bundle may be set"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, page, logs := stagePage(t, func(cur gox.Cursor) error {
				return stageTag(cur, tc.tag, tc.void, nil, tc.attr...)
			})
			if !strings.Contains(logs, tc.want) && !strings.Contains(page, tc.want) {
				t.Fatalf("no %q in page %q or logs %q", tc.want, page, logs)
			}
		})
	}
}

// TestResourcePlainStageTags checks that Name, Cache and Portable on a plain
// static resource act on script and stylesheet tags like the attributes do.
func TestResourcePlainStageTags(t *testing.T) {
	js := ResourceString("window.a = 1 + 1")
	css := ResourceString("body {  color: red;  }")
	pairs := []struct {
		id, tag, attr, rel string
		stage              ServedResource
		plain              []tagAttr
	}{
		{"js-name", "script", "src", "", js.Name("app.js"), []tagAttr{{"name", "app.js"}}},
		{"js-cache", "script", "src", "", js.Cache(), []tagAttr{{"cache", true}}},
		{"css-name", "link", "href", "stylesheet", css.Name("site.css"), []tagAttr{{"name", "site.css"}}},
	}
	_, _, page, logs := stagePage(t, func(cur gox.Cursor) error {
		for _, p := range pairs {
			void := p.tag == "link"
			stage := []tagAttr{{"id", p.id + "-stage"}, {p.attr, p.stage}}
			plain := append([]tagAttr{{"id", p.id + "-plain"}, {p.attr, ResourceString(map[string]string{"script": "window.a = 1 + 1", "link": "body {  color: red;  }"}[p.tag])}}, p.plain...)
			if p.rel != "" {
				stage = append(stage, tagAttr{"rel", p.rel})
				plain = append(plain, tagAttr{"rel", p.rel})
			}
			if err := stageTag(cur, p.tag, void, nil, stage...); err != nil {
				return err
			}
			if err := stageTag(cur, p.tag, void, nil, plain...); err != nil {
				return err
			}
		}
		return nil
	})
	for _, p := range pairs {
		got, want := stageAttr(t, page, p.id+"-stage", p.attr), stageAttr(t, page, p.id+"-plain", p.attr)
		if got == "" || got != want {
			t.Fatalf("%s: stage %q, attributes %q; logs %s", p.id, got, want, logs)
		}
	}
}

// TestResourceChainForms renders every kind of chain value both as a src or
// href value and as a modifier, and checks how each is built and served.
func TestResourceChainForms(t *testing.T) {
	js := ResourceString("window.a = 1 + 1")
	css := ResourceString("body {  color: red;  }")
	img := ResourceString("<svg/>")
	handler := func(body string) ResourceDynamic {
		return ResourceHandler(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) })
	}
	const built, raw, minified, rawCSS = "window.a=2;", "window.a = 1 + 1", "body{color:red}", "body {  color: red;  }"
	cases := []struct {
		id, tag, attr, rel string
		r                  ServedResource
		body, url          string
	}{
		{"s-plain", "script", "src", "", js, built, "/r/"},
		{"s-name", "script", "src", "", js.Name("a.js"), built, "/r/"},
		{"s-cache", "script", "src", "", js.Cache(), built, "/r/"},
		{"s-portable", "script", "src", "", js.Portable(), built, "/w/"},
		{"s-script", "script", "src", "", js.Script(), built, "/r/"},
		{"s-bundle", "script", "src", "", js.Script().Bundle(), built, "/r/"},
		{"s-module", "script", "src", "", js.Script().Module(), built, "/r/"},
		{"s-raw", "script", "src", "", js.Script().Raw(), raw, "/r/"},
		{"s-raw-nocache", "script", "src", "", js.Script().Raw().NoCache(), raw, "/h/"},
		{"s-private", "script", "src", "", js.Script().Private(), built, "/h/"},
		{"s-nocache-portable", "script", "src", "", js.Script().NoCache().Portable(), built, "/w/"},
		{"s-handler", "script", "src", "", handler("hook-js"), "hook-js", "/h/"},
		{"s-handler-portable", "script", "src", "", handler("hook-js").Name("x.js").Portable(), "hook-js", "/w/"},
		{"c-plain", "link", "href", "stylesheet", css, minified, "/r/"},
		{"c-style", "link", "href", "stylesheet", css.Style(), minified, "/r/"},
		{"c-raw", "link", "href", "stylesheet", css.Style().Raw(), rawCSS, "/r/"},
		{"c-private", "link", "href", "stylesheet", css.Style().Private(), minified, "/h/"},
		{"c-portable", "link", "href", "stylesheet", css.Style().Portable(), minified, "/w/"},
		{"c-name", "link", "href", "stylesheet", css.Name("s.css"), minified, "/r/"},
		{"c-handler", "link", "href", "stylesheet", handler("hook-css"), "hook-css", "/h/"},
		{"i-plain", "img", "src", "", img, "<svg/>", "/h/"},
		{"i-cache", "img", "src", "", img.Name("i.svg").ContentType("image/svg+xml").Cache(), "<svg/>", "/r/"},
		{"i-portable", "img", "src", "", img.Portable(), "<svg/>", "/w/"},
		{"i-handler", "img", "src", "", handler("hook-img"), "hook-img", "/h/"},
	}
	c, base, page, logs := stagePage(t, func(cur gox.Cursor) error {
		for _, tc := range cases {
			void := tc.tag != "script"
			asAttr := []tagAttr{{"id", tc.id + "-attr"}, {tc.attr, tc.r}}
			asMod := []tagAttr{{"id", tc.id + "-mod"}}
			if tc.rel != "" {
				asAttr = append(asAttr, tagAttr{"rel", tc.rel})
				asMod = append(asMod, tagAttr{"rel", tc.rel})
			}
			if err := stageTag(cur, tc.tag, void, nil, asAttr...); err != nil {
				return err
			}
			if err := stageTag(cur, tc.tag, void, tc.r, asMod...); err != nil {
				return err
			}
		}
		return nil
	})
	for _, tc := range cases {
		for _, form := range []string{"-attr", "-mod"} {
			url := stageAttr(t, page, tc.id+form, tc.attr)
			if !strings.Contains(url, tc.url) {
				t.Fatalf("%s%s: %s %q, want %s URL; logs %s", tc.id, form, tc.attr, url, tc.url, logs)
			}
			client := c
			if tc.url == "/w/" {
				client = &http.Client{}
			}
			resp, body := headDo(t, client, http.MethodGet, base+url, false)
			if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), tc.body) {
				t.Fatalf("%s%s: status %d, body %q, want %q", tc.id, form, resp.StatusCode, body, tc.body)
			}
		}
	}
}

// TestResourceStoreHelper covers the recommended pattern for app-wide resource
// URLs: a helper builds the URL once into a Store and panics on error, which
// fails the render and stores nothing, so a later call retries.
func TestResourceStoreHelper(t *testing.T) {
	file := filepath.Join(t.TempDir(), "mod.js")
	store := NewStore()
	type moduleKey struct{}
	var builds atomic.Int32
	module := func(ctx context.Context) string {
		return store.Init(moduleKey{}, func() any {
			url, err := ResourceLocalFS(file).Script().Module().Bundle().URL(ctx)
			if err != nil {
				panic(err)
			}
			builds.Add(1)
			return url
		}).(string)
	}
	logs := &syncBuffer{}
	app := stageApp(logs, func(cur gox.Cursor) error {
		return stageTag(cur, "script", false, nil, tagAttr{"id", "mod"}, tagAttr{"data:module", module(cur.Context())})
	})
	srv := headServer(t, app, false)
	load := func() (int, string) {
		resp, page := headDo(t, headClient(t, srv, true), http.MethodGet, srv.URL+"/", false)
		return resp.StatusCode, string(page)
	}
	if status, _ := load(); status == http.StatusOK {
		t.Fatal("expected the render to fail while the file is missing")
	}
	if builds.Load() != 0 {
		t.Fatal("a failed build was stored")
	}
	if err := os.WriteFile(file, []byte("export const a = 1"), 0o644); err != nil {
		t.Fatal(err)
	}
	var url string
	for i := 0; i < 2; i++ {
		status, page := load()
		if status != http.StatusOK {
			t.Fatalf("render %d: status %d; logs %s", i, status, logs.String())
		}
		data := stageAttr(t, page, "mod", "data-d0d-module")
		if url == "" {
			url = data
		} else if data != url {
			t.Fatalf("render %d: data %q, first %q", i, data, url)
		}
	}
	if n := builds.Load(); n != 1 {
		t.Fatalf("built %d times, want 1", n)
	}
	got := module(context.Background())
	resp, body := headDo(t, headClient(t, srv, false), http.MethodGet, srv.URL+got, false)
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "export") {
		t.Fatalf("stored URL %q: status %d, body %q", got, resp.StatusCode, body)
	}
}

func TestResourceScriptData(t *testing.T) {
	js := ResourceString("window.a = 1 + 1")
	css := ResourceString("body {  color: red;  }")
	handler := ResourceHandler(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("hook-data")) })
	cases := []struct {
		id   string
		r    any
		body string
		url  string
	}{
		{"plain", js, "window.a = 1 + 1", "/h/"},
		{"cache", js.ContentType("text/javascript").Cache(), "window.a = 1 + 1", "/r/"},
		{"script", js.Script(), "window.a=2;", "/r/"},
		{"module", js.Script().Module().Bundle(), "window.a=2;", "/r/"},
		{"private", js.Script().Private(), "window.a=2;", "/h/"},
		{"portable", js.Script().Portable(), "window.a=2;", "/w/"},
		{"style", css.Style(), "body{color:red}", "/r/"},
		{"handler", handler, "hook-data", "/h/"},
		{"handler-portable", handler.Portable(), "hook-data", "/w/"},
	}
	c, base, page, logs := stagePage(t, func(cur gox.Cursor) error {
		for _, tc := range cases {
			if err := stageTag(cur, "script", false, nil, tagAttr{"id", tc.id + "-inline"}, tagAttr{"data:x", tc.r}); err != nil {
				return err
			}
			if err := stageTag(cur, "script", false, nil, tagAttr{"id", tc.id + "-src"}, tagAttr{"src", "/app.js"}, tagAttr{"data:x", tc.r}); err != nil {
				return err
			}
		}
		return stageTag(cur, "script", false, nil, tagAttr{"id", "bytes"}, tagAttr{"data:x", []byte("x")})
	})
	payload := regexp.MustCompile(`^\[3,&#34;([^&]*)&#34;\]$`)
	for _, tc := range cases {
		for _, form := range []string{"-inline", "-src"} {
			m := payload.FindStringSubmatch(stageAttr(t, page, tc.id+form, "data-d0d-x"))
			if m == nil || !strings.Contains(m[1], tc.url) {
				t.Fatalf("%s%s: payload %v, want %s URL; logs %s", tc.id, form, m, tc.url, logs)
			}
			client := c
			if tc.url == "/w/" {
				client = &http.Client{}
			}
			resp, body := headDo(t, client, http.MethodGet, base+m[1], false)
			if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), tc.body) {
				t.Fatalf("%s%s: status %d, body %q, want %q", tc.id, form, resp.StatusCode, body, tc.body)
			}
		}
	}
	if got := stageAttr(t, page, "bytes", "data-d0d-x"); got != "[1,&#34;eA==&#34;]" {
		t.Fatalf("bytes payload %q", got)
	}
}

func TestResourceUnsupportedAttr(t *testing.T) {
	cases := []struct {
		name string
		tag  string
		attr tagAttr
		want string
	}{
		{"poster source", "video", tagAttr{"poster", ResourceString("x")}, "only src, href, and script data: attributes take a resource"},
		{"poster chain", "video", tagAttr{"poster", ResourceString("x").ContentType("image/png").Cache()}, "only src, href, and script data: attributes take a resource"},
		{"poster handler", "video", tagAttr{"poster", ResourceHandler(func(w http.ResponseWriter, r *http.Request) {})}, "only src, href, and script data: attributes take a resource"},
		{"data read error", "script", tagAttr{"data:x", ResourceLocalFS("./missing-data.js").Script()}, "missing-data.js"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, page, logs := stagePage(t, func(cur gox.Cursor) error {
				return stageTag(cur, tc.tag, false, nil, tagAttr{"id", "t"}, tc.attr)
			})
			if !strings.Contains(logs, tc.want) || strings.Contains(page, `id="t"`) {
				t.Fatalf("no render error %q; page %q, logs %q", tc.want, page, logs)
			}
		})
	}
}
