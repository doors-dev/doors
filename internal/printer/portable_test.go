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
	"net/http"
	"strings"
	"testing"

	"github.com/doors-dev/gox"
)

func portableHandler() HandlerSimpleFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}
}

func portableAttrs(src any, flags ...string) gox.Attrs {
	attrs := gox.NewAttrs()
	attrs.Get("src").Set(src)
	for _, f := range flags {
		attrs.Get(f).Set(true)
	}
	return attrs
}

// TestPortableValidation rejects portable with cache and cache or nocache on handlers, and allows private on handlers.
func TestPortableValidation(t *testing.T) {
	ctx, _, _, _ := newPrinterCore(t, true)
	for _, c := range []struct {
		name  string
		attrs gox.Attrs
		err   string
	}{
		{"portable static with cache", portableAttrs(SourceString("hello"), "portable", "cache"), "portable cannot be combined with cache"},
		{"handler with cache", portableAttrs(portableHandler(), "cache"), "do not support cache or nocache"},
		{"handler with nocache", portableAttrs(portableHandler(), "nocache"), "do not support cache or nocache"},
		{"portable handler with nocache", portableAttrs(portableHandler(), "portable", "nocache"), "do not support cache or nocache"},
		{"handler with private", portableAttrs(portableHandler(), "private"), ""},
		{"portable handler with private", portableAttrs(portableHandler(), "portable", "private"), ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			rp := &resourcePrinter{printer: defaultPrinter{&out}}
			err := rp.scanGenericSrc(gox.NewJobOpen(ctx, 1, gox.KindVoid, "img", c.attrs))
			if c.err == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !strings.Contains(out.String(), "/h/instance/") && !strings.Contains(out.String(), "/w/tok") {
					t.Fatalf("expected a hook URL, got %q", out.String())
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.err) {
				t.Fatalf("got error %v, want %q", err, c.err)
			}
		})
	}
	var out bytes.Buffer
	rp := &resourcePrinter{printer: defaultPrinter{&out}}
	if err := rp.prepareScript(gox.NewJobOpen(ctx, 2, gox.KindRegular, "script", portableAttrs(SourceString("let a = 1"), "portable", "cache"))); err == nil || !strings.Contains(err.Error(), "portable cannot be combined with cache") {
		t.Fatalf("script: got error %v", err)
	}
	style := gox.NewAttrs()
	style.Get("rel").Set("stylesheet")
	style.Get("href").Set(SourceString("a{}"))
	style.Get("portable").Set(true)
	style.Get("cache").Set(true)
	if err := rp.prepareLinkStyle(gox.NewJobOpen(ctx, 3, gox.KindVoid, "link", style)); err == nil || !strings.Contains(err.Error(), "portable cannot be combined with cache") {
		t.Fatalf("style: got error %v", err)
	}
}

// TestPortableURLs gives portable resources a /w/ URL and leaves the rest on /h/ or /r/.
func TestPortableURLs(t *testing.T) {
	style := func(src any, flags ...string) gox.Attrs {
		attrs := gox.NewAttrs()
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(src)
		for _, f := range flags {
			attrs.Get(f).Set(true)
		}
		return attrs
	}
	named := func(attrs gox.Attrs, name string) gox.Attrs {
		attrs.Get("name").Set(name)
		return attrs
	}
	for _, c := range []struct {
		name  string
		tag   string
		kind  gox.HeadKind
		attrs gox.Attrs
		want  string
	}{
		{"script static", "script", gox.KindRegular, named(portableAttrs(SourceString("let a = 1"), "portable"), "app.js"), `src="/~/srv/w/tok1/app.js"`},
		{"script static private", "script", gox.KindRegular, portableAttrs(SourceString("let a = 1"), "portable", "private"), `src="/~/srv/w/tok1/script.js"`},
		{"script handler", "script", gox.KindRegular, portableAttrs(portableHandler(), "portable"), `src="/~/srv/w/tok1/script.js"`},
		{"script static not portable", "script", gox.KindRegular, portableAttrs(SourceString("let a = 1")), `/~/srv/r/`},
		{"style static", "link", gox.KindVoid, style(SourceString("a{}"), "portable"), `href="/~/srv/w/tok1/style.css"`},
		{"style static nocache", "link", gox.KindVoid, style(SourceString("a{}"), "portable", "nocache"), `href="/~/srv/w/tok1/style.css"`},
		{"style handler", "link", gox.KindVoid, style(portableHandler(), "portable"), `href="/~/srv/w/tok1/style.css"`},
		{"style static not portable", "link", gox.KindVoid, style(SourceString("a{}")), `/~/srv/r/`},
		{"generic static", "img", gox.KindVoid, portableAttrs(SourceString("hello"), "portable"), `src="/~/srv/w/tok1"`},
		{"generic handler", "img", gox.KindVoid, named(portableAttrs(portableHandler(), "portable"), "my pic.png"), `src="/~/srv/w/tok1/my%20pic.png"`},
		{"generic handler not portable", "img", gox.KindVoid, portableAttrs(portableHandler()), `src="/~/srv/h/instance/1"`},
	} {
		t.Run(c.name, func(t *testing.T) {
			ctx, _, _, _ := newPrinterCore(t, true)
			var out bytes.Buffer
			printer := NewResourcePrinter(defaultPrinter{&out})
			if err := printer.Send(gox.NewJobOpen(ctx, 1, c.kind, c.tag, c.attrs)); err != nil {
				t.Fatal(err)
			}
			if c.kind == gox.KindRegular {
				if err := printer.Send(gox.NewJobClose(ctx, 1, c.kind, c.tag)); err != nil {
					t.Fatal(err)
				}
			}
			got := out.String()
			if !strings.Contains(got, c.want) {
				t.Fatalf("got %q, want it to contain %q", got, c.want)
			}
			if strings.Contains(got, "portable") {
				t.Fatalf("portable attribute left in output %q", got)
			}
		})
	}
}

// TestPortableEmbedded gives portable inline scripts and styles a /w/ URL.
func TestPortableEmbedded(t *testing.T) {
	for _, c := range []struct {
		tag  string
		body string
		want string
	}{
		{"script", `window.a = 1`, `src="/~/srv/w/tok1/script.js"`},
		{"style", `a{color:red}`, `href="/~/srv/w/tok1/style.css"`},
	} {
		t.Run(c.tag, func(t *testing.T) {
			ctx, _, _, _ := newPrinterCore(t, true)
			var out bytes.Buffer
			printer := NewResourcePrinter(defaultPrinter{&out})
			attrs := gox.NewAttrs()
			attrs.Get("portable").Set(true)
			if err := printer.Send(gox.NewJobOpen(ctx, 1, gox.KindRegular, c.tag, attrs)); err != nil {
				t.Fatal(err)
			}
			if err := printer.Send(gox.NewJobRaw(ctx, c.body)); err != nil {
				t.Fatal(err)
			}
			if err := printer.Send(gox.NewJobClose(ctx, 1, gox.KindRegular, c.tag)); err != nil {
				t.Fatal(err)
			}
			if got := out.String(); !strings.Contains(got, c.want) || strings.Contains(got, "portable") {
				t.Fatalf("got %q, want it to contain %q", got, c.want)
			}
		})
	}
}
