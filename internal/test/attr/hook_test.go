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

package attr

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/doors-dev/gox"
)

// TestAttrHookNoArg checks that $hook without an argument reaches every hook kind as JSON null.
func TestAttrHookNoArg(t *testing.T) {
	bro := test.NewFragmentBro(browser, func() test.Fragment {
		return &nullHookFragment{
			r: test.NewReporter(4),
		}
	})
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	<-time.After(300 * time.Millisecond)
	test.TestContent(t, page, "#ptr", "ptr")
	test.TestContent(t, page, "#str", "str")
	test.TestContent(t, page, "#obj", "obj")
	test.TestContent(t, page, "#raw", "raw")
	test.TestReportId(t, page, 0, "called:true")
	test.TestReportId(t, page, 1, `called:""`)
	test.TestReportId(t, page, 2, `called:""`)
	test.TestReportId(t, page, 3, "null")
}

// TestAttrHookTrack checks that $hook sends its track in the D0-Track header and handlers see the path after the hook.
func TestAttrHookTrack(t *testing.T) {
	var mu sync.Mutex
	var queries, tracks []string
	record := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/h/") {
				mu.Lock()
				queries = append(queries, r.URL.RawQuery)
				tracks = append(tracks, r.Header.Get("D0-Track"))
				mu.Unlock()
			}
			next.ServeHTTP(w, r)
		})
	}
	bro := test.NewPathBro(browser, func(s test.PathLens) gox.Comp {
		return &test.Page{
			Source: s,
			F: &trackHookFragment{
				r: test.NewReporter(1),
			},
		}
	}, record)
	page := bro.Page(t, "/")
	defer bro.Close()
	defer page.Close()
	<-time.After(300 * time.Millisecond)
	test.TestContent(t, page, "#track", "second")
	test.TestContent(t, page, "#bare", "/||")
	test.TestContent(t, page, "#slash", "/||")
	test.TestContent(t, page, "#sub", "/a/b/c|/a/b%2Fc|q=1")
	test.TestContent(t, page, "#query", "/||q=2")
	mu.Lock()
	defer mu.Unlock()
	if len(tracks) != 2 {
		t.Fatal("hook requests: ", len(tracks))
	}
	for i, track := range tracks {
		if _, err := strconv.ParseUint(track, 10, 64); err != nil {
			t.Fatal("track header: ", track)
		}
		if queries[i] != "" {
			t.Fatal("hook query: ", queries[i])
		}
	}
	if tracks[0] == tracks[1] {
		t.Fatal("track repeated: ", tracks[0])
	}
	test.TestReportId(t, page, 0, tracks[1])
}
