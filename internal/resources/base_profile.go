// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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
