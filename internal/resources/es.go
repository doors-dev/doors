// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package resources

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/evanw/esbuild/pkg/api"
)


type BuildErrors []api.Message

func (b BuildErrors) Error() string {
	var errs []error
	for _, m := range b {
		var formatted string
		if m.Location != nil {
			formatted = fmt.Sprintf("%s:%d:%d: %s",
				m.Location.File,
				m.Location.Line,
				m.Location.Column,
				m.Text)
		} else {
			formatted = m.Text
		}
		errs = append(errs, errors.New(formatted))
	}
	return errors.Join(errs...).Error()
}

func build(options *api.BuildOptions) ([]byte, error) {
	options.Write = false
	options.Platform = api.PlatformBrowser
	result := api.Build(*options)
	if len(result.Errors) != 0 {
		for _, m := range result.Errors {
			slog.Error("esbuild error", slog.String("text", m.Text))
		}
		return nil, BuildErrors(result.Errors)

	}
	if len(result.OutputFiles) == 0 {
		return nil, BuildErrors([]api.Message{{
			Text: "no output produced",
		}})
	}
	data := result.OutputFiles[0].Contents
	return data, nil
}
