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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/doors-dev/gox"
)

var resourceURLRegexp = regexp.MustCompile(`res:([^;]+);same:(\w+);err:(\w+)`)

func newResourceURLApp() App {
	return NewApp(func(context.Context, Request) gox.Comp {
		return gox.Elem(func(cur gox.Cursor) error {
			ctx := cur.Context()
			url, err := NewResource(ctx, ResourceBytes([]byte("export const a = 1")), "text/javascript")
			if err != nil {
				return err
			}
			again, err := NewResource(ctx, ResourceBytes([]byte("export const a = 1")), "text/javascript")
			if err != nil {
				return err
			}
			_, missing := NewResource(ctx, ResourceLocalFS("no/such/file.js"), "text/javascript")
			return headTag(cur, "html", func() error {
				return headTag(cur, "body", func() error {
					return cur.Text(fmt.Sprintf("res:%s;same:%t;err:%t", url, url == again, missing != nil))
				})
			})
		})
	},
		WithConf(Conf{ServerSessionCookieNoSecure: true, InstanceConnectTimeout: time.Minute, ServerCacheControl: "public, max-age=60"}),
		WithLogger(slog.New(slog.NewTextHandler(io.Discard, nil))),
	)
}

// TestNewResource serves a static resource from a cached URL that any client can load.
func TestNewResource(t *testing.T) {
	srv := headServer(t, newResourceURLApp(), false)
	_, body := headDo(t, headClient(t, srv, true), http.MethodGet, srv.URL+"/", false)
	m := resourceURLRegexp.FindSubmatch(body)
	if m == nil {
		t.Fatalf("no resource URL in page body %q", body)
	}
	url := string(m[1])
	if string(m[2]) != "true" {
		t.Fatal("expected the same content to give the same URL")
	}
	if string(m[3]) != "true" {
		t.Fatal("expected an error for a missing file")
	}
	c := headClient(t, srv, false)
	for _, uri := range []string{url, url + ".app.js"} {
		resp, got := headDo(t, c, http.MethodGet, srv.URL+uri, false)
		if resp.StatusCode != http.StatusOK || string(got) != "export const a = 1" {
			t.Fatalf("%s: status %d, body %q", uri, resp.StatusCode, got)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "text/javascript" {
			t.Fatalf("%s: content type %q", uri, ct)
		}
		if cc := resp.Header.Get("Cache-Control"); cc != "public, max-age=60" {
			t.Fatalf("%s: cache control %q", uri, cc)
		}
	}
}
