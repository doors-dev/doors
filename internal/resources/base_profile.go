// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package resources

import "github.com/evanw/esbuild/pkg/api"

type BaseProfile struct {
}

const minify bool = true

func (b BaseProfile) Options(profile string) api.BuildOptions {
	return api.BuildOptions{
		Target:            api.ES2022,
		MinifySyntax:      minify,
		MinifyWhitespace:  minify,
		MinifyIdentifiers: minify,
	}
}
