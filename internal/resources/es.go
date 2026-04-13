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
			slog.Error("esbuild error", "text", m.Text)
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
