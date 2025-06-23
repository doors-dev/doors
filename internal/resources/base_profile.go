package resources

import "github.com/evanw/esbuild/pkg/api"

type BaseProfile struct {
}

const minify bool = true

func (b BaseProfile) Options(profile string) api.BuildOptions {
	return api.BuildOptions{
		Target:            api.ES2017,
		MinifySyntax:      minify,
		MinifyWhitespace:  minify,
		MinifyIdentifiers: minify,
	}
}
