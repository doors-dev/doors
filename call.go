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

package doors

import (
	"context"
	"encoding/json"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
)

// Call dispatches action to the client.
//
// Canceling ctx requests best-effort cancellation of the call.
//
// The returned channel is optional to use: it receives nil on success or an
// error on client or decode failure, then closes; if the call is canceled, it
// closes without a value.
//
// To capture the client result, arm the action with its Into method before
// dispatch; see [ActionInto].
//
// Do not wait on the channel during rendering. If you need to wait, use [Go]
// or your own goroutine with [DetachedContext].
func Call(ctx context.Context, a Action) <-chan error {
	core := ctx.Value(common.KeyCore).(core.Core)
	ch := make(chan error, 1)
	prep, err := a.action(ctx, core, !core.App().Conf().SolitaireDisableGzip)
	if err != nil {
		core.App().Logger().Error("Action preparation error", "error", err)
		ch <- err
		close(ch)
		return ch
	}
	if ctx.Err() != nil {
		ctex.LogCanceled(ctx, "call "+prep.action.Log())
	}
	core.Door().UserCall(
		ctx,
		nil,
		prep.action,
		func(rm json.RawMessage, err error) {
			if err == nil && prep.decode != nil {
				err = prep.decode(rm)
			}
			ch <- err
			close(ch)
		},
		func() {
			close(ch)
		},
		prep.params,
	)
	return ch
}
