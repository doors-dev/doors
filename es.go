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

type ESConf = resources.BuildProfiles

type JSX struct {
	JSX          api.JSX
	Factory      string
	ImportSource string
	Fragment     string
	SideEffects  bool
	Dev          bool
}

func JSXPreact() JSX {
	return JSX{
		Factory:  "h",
		Fragment: "Fragment",
	}
}

func JSXReact() JSX {
	return JSX{
		JSX: api.JSXAutomatic,
	}
}

type ESOptions struct {
	External []string
	Minify   bool
	JSX      JSX
}

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
