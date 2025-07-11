package doors

import (
	"github.com/doors-dev/doors/internal/resources"
	"github.com/evanw/esbuild/pkg/api"
)

type ESProfiles = resources.BuildProfiles

type ESJSX struct {
	JSX          api.JSX
	Factory      string
	ImportSource string
	Fragment     string
	SideEffects  bool
	Dev          bool
}

type ESBuildOptions struct {
	External []string
	Minify   bool
	JSX      ESJSX
}

func (opt ESBuildOptions) Options(_profile string) api.BuildOptions {
	return api.BuildOptions{
		Target:            api.ES2017,
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
