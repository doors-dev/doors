// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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
