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

package door

import (
	"context"
	"errors"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
)

func logError(ctx context.Context, logger *slog.Logger, err error, caller common.Caller) {
	if errors.Is(err, context.Canceled) && ctx.Err() != nil {
		return
	}
	if source := caller.String(); source != "" {
		logger.Error("door rendering error", "error", err, "caller", source)
		return
	}
	logger.Error("door rendering error", "error", err)
}
