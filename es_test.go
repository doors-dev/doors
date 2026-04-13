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
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

func TestESOptionsAndJSXPresets(t *testing.T) {
	preact := JSXPreact()
	if preact.Factory != "h" || preact.Fragment != "Fragment" {
		t.Fatalf("unexpected preact preset: %+v", preact)
	}

	react := JSXReact()
	if react.JSX != api.JSXAutomatic {
		t.Fatalf("unexpected react preset: %+v", react)
	}

	opt := ESOptions{
		External: []string{"react"},
		Minify:   true,
		JSX: JSX{
			JSX:          api.JSXAutomatic,
			Factory:      "h",
			ImportSource: "preact",
			Fragment:     "Fragment",
			SideEffects:  true,
			Dev:          true,
		},
	}

	build := opt.Options("ignored")
	if build.Target != api.ES2022 {
		t.Fatalf("unexpected target: %v", build.Target)
	}
	if len(build.External) != 1 || build.External[0] != "react" {
		t.Fatalf("unexpected externals: %#v", build.External)
	}
	if !build.MinifySyntax || !build.MinifyWhitespace || !build.MinifyIdentifiers {
		t.Fatal("expected minify flags to be enabled")
	}
	if build.JSX != api.JSXAutomatic {
		t.Fatalf("unexpected jsx mode: %v", build.JSX)
	}
	if build.JSXFactory != "h" || build.JSXFragment != "Fragment" || build.JSXImportSource != "preact" {
		t.Fatalf("unexpected jsx settings: %+v", build)
	}
	if !build.JSXDev || !build.JSXSideEffects {
		t.Fatal("expected jsx dev and side effects flags to be enabled")
	}
}
