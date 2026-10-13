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

package client

import (
	"os"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/test"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/go-rod/rod"
)

var browser *rod.Browser

func TestMain(m *testing.M) {
	browser = test.NewBrowser()
	code := m.Run()
	browser.MustClose()
	os.Exit(code)
}

// TestClient runs every registered client case in a blank page, one subtest per case.
func TestClient(t *testing.T) {
	result := api.Build(api.BuildOptions{
		EntryPoints: []string{"cases.ts"},
		Bundle:      true,
		Format:      api.FormatIIFE,
		GlobalName:  "_d0rCases",
		Target:      api.ES2022,
		Platform:    api.PlatformBrowser,
		MangleProps: "_$",
	})
	if len(result.Errors) != 0 {
		t.Fatal(api.FormatMessages(result.Errors, api.FormatMessagesOptions{}))
	}
	page := browser.MustPage("about:blank")
	defer page.MustClose()
	if _, err := page.Eval("(src) => { (0, eval)(src) }", string(result.OutputFiles[0].Contents)); err != nil {
		t.Fatal(err)
	}
	names, err := page.Eval("() => _d0rCases.names()")
	if err != nil {
		t.Fatal(err)
	}
	if len(names.Value.Arr()) == 0 {
		t.Fatal("no cases registered")
	}
	stuck := false
	for _, name := range names.Value.Arr() {
		if stuck {
			break
		}
		t.Run(name.Str(), func(t *testing.T) {
			res, err := page.Timeout(10*time.Second).Eval("(name) => _d0rCases.run(name)", name.Str())
			if err != nil {
				stuck = true
				t.Fatal(err)
			}
			if msg := res.Value.Str(); msg != "" {
				t.Fatal(msg)
			}
		})
	}
	if stuck {
		t.Fatal("a case did not return, the remaining cases were not run")
	}
}
