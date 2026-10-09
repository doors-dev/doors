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
	"strings"
	"testing"
	"time"

	"github.com/doors-dev/gox"
)

func cspScript(cur gox.Cursor, host string) error {
	if err := cur.Init("script"); err != nil {
		return err
	}
	if err := cur.Set("src", ResourceExternal("https://"+host+"/a.js")); err != nil {
		return err
	}
	if err := cur.Submit(); err != nil {
		return err
	}
	return cur.Close()
}

func cspHost(i int) string {
	return fmt.Sprintf("h%d.example.com", i)
}

func TestCSPCollectsConcurrentDoorSources(t *testing.T) {
	const doors = 8
	page := gox.Elem(func(cur gox.Cursor) error {
		return headTag(cur, "html", func() error {
			return headTag(cur, "body", func() error {
				for i := range doors {
					host := cspHost(i)
					d := &Door{}
					d.Inner(cur.Context(), gox.Elem(func(cur gox.Cursor) error {
						return cspScript(cur, host)
					}))
					if err := cur.Comp(d); err != nil {
						return err
					}
				}
				return nil
			})
		})
	})
	conf := Conf{}
	conf.ServerSessionCookieNoSecure = true
	conf.InstanceConnectTimeout = time.Minute
	app := NewApp(func(context.Context, Request) gox.Comp { return page }, WithConf(conf), WithCSP(CSP{}))
	srv := headServer(t, app, false)
	c := headClient(t, srv, true)
	for range 20 {
		resp, _ := headDo(t, c, "GET", srv.URL+"/", false)
		header := resp.Header.Get("Content-Security-Policy")
		for i := range doors {
			if !strings.Contains(header, cspHost(i)) {
				t.Fatalf("CSP header misses %s: %s", cspHost(i), header)
			}
		}
	}
}
