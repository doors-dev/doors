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
	"github.com/doors-dev/doors/internal/app"
	"github.com/evanw/esbuild/pkg/api"
)

// JSX is the JSX transform configuration for [ESProfile].
type JSX struct {
	// JSX selects the esbuild JSX mode. Optional; the zero value is
	// [api.JSXTransform].
	JSX api.JSX
	// Factory is the function called for a JSX element outside the automatic
	// runtime. Optional; esbuild uses React.createElement.
	Factory string
	// ImportSource is the base package the automatic runtime imports from.
	// Optional; esbuild uses react.
	ImportSource string
	// Fragment is the expression used for a JSX fragment outside the automatic
	// runtime. Optional; esbuild uses React.Fragment.
	Fragment string
	// SideEffects tells esbuild that JSX elements may have side effects, so
	// unused ones are kept.
	SideEffects bool
	// Dev emits jsxDEV calls with source locations. Applies to the automatic
	// runtime only.
	Dev bool
}

// JSXPreact returns JSX settings that compile elements to h calls, with
// Fragment as the fragment expression.
func JSXPreact() JSX {
	return JSX{
		Factory:  "h",
		Fragment: "Fragment",
	}
}

// JSXReact returns JSX settings that use the automatic runtime and import from
// react/jsx-runtime.
func JSXReact() JSX {
	return JSX{
		JSX: api.JSXAutomatic,
	}
}

// ESProfile is a [With] option that applies one set of esbuild options to every
// script profile. Use [WithESProfiles] to vary options per profile.
type ESProfile struct {
	// External lists import paths esbuild leaves out of the bundle.
	External []string
	// Minify enables esbuild syntax, whitespace, and identifier minification.
	Minify bool
	// JSX configures the JSX transform.
	JSX JSX
}

func (opt ESProfile) apply(o *app.Options) {
	WithESProfiles(opt.Profile).apply(o)
}

var _ With = ESProfile{}

// Profile returns the same esbuild options for every profile name.
func (opt ESProfile) Profile(string) api.BuildOptions {
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
