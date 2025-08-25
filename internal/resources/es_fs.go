// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package resources

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/doors-dev/doors/internal/common"
	"github.com/evanw/esbuild/pkg/api"
)

func fsPlugin(dir fs.FS) api.Plugin {
	return api.Plugin{
		Name: "fs",
		Setup: func(build api.PluginBuild) {
			build.OnResolve(api.OnResolveOptions{Filter: ".*"}, func(args api.OnResolveArgs) (api.OnResolveResult, error) {
				resolvePath := path.Clean(path.Join(path.Dir(args.Importer), args.Path))
				if filepath.Ext(resolvePath) == "" {
					resolvePath += ".ts"
				}
				return api.OnResolveResult{
					Path:      resolvePath,
					Namespace: "fs",
				}, nil
			})

			build.OnLoad(api.OnLoadOptions{Filter: ".*", Namespace: "fs"}, func(args api.OnLoadArgs) (api.OnLoadResult, error) {
				data, err := fs.ReadFile(dir, args.Path)
				if err != nil {
					return api.OnLoadResult{}, err
				}
				loader := detectLoader(args.Path)
				str := common.AsString(&data)
				return api.OnLoadResult{
					Contents:   &str,
					Loader:     loader,
					ResolveDir: path.Dir(args.Path),
				}, nil
			})
		},
	}
}

func detectLoader(filePath string) api.Loader {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".js", ".mjs", ".cjs":
		return api.LoaderJS
	case ".ts", ".mts", ".cts":
		return api.LoaderTS
	case ".tsx":
		return api.LoaderTSX
	case ".jsx":
		return api.LoaderJSX
	case ".json":
		return api.LoaderJSON
	case ".css":
		return api.LoaderCSS
	case ".txt":
		return api.LoaderText
	case ".wasm":
		return api.LoaderBinary
	default:
		return api.LoaderJS
	}
}
