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
	"github.com/doors-dev/doors/internal/resources"
	"github.com/evanw/esbuild/pkg/api"
)

// ESConf provides named esbuild profiles for script and style processing.
type ESConf = resources.BuildProfiles

// JSX configures how esbuild should transform JSX input.
type JSX struct {
	JSX          api.JSX
	Factory      string
	ImportSource string
	Fragment     string
	SideEffects  bool
	Dev          bool
}

// JSXPreact returns JSX settings suitable for classic Preact transforms.
func JSXPreact() JSX {
	return JSX{
		Factory:  "h",
		Fragment: "Fragment",
	}
}

// JSXReact returns JSX settings suitable for React's automatic runtime.
func JSXReact() JSX {
	return JSX{
		JSX: api.JSXAutomatic,
	}
}

// ESOptions is a simple [ESConf] implementation for one build profile.
type ESOptions struct {
	External []string
	Minify   bool
	JSX      JSX
}

// Options implements [ESConf].
func (opt ESOptions) Options(_profile string) api.BuildOptions {
	return api.BuildOptions{
		Target:            api.ES2022,
		External:          opt.External,
		MinifySyntax:      opt.Minify,
		MinifyWhitespace:  opt.Minify,
		MinifyIdentifiers: opt.Minify,
		JSX:               opt.JSX.JSX,
		JSXFactory:        opt.JSX.Factory,
		JSXDev:            opt.JSX.Dev,
		JSXSideEffects:    opt.JSX.SideEffects,
		JSXFragment:       opt.JSX.Fragment,
		JSXImportSource:   opt.JSX.ImportSource,
	}
}
