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
