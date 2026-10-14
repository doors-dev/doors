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
// The returned channel is optional to use. It sends nil once the call is
// scheduled, or an error if the action cannot be built. On a detached context
// ([DetachedContext], [Go]) a second value reports the result: nil, or an
// error if the client fails or the result cannot be decoded; otherwise the
// channel closes once the call is scheduled. If you need the result, wait for
// the second value; it only arrives on a detached context. If the call is
// canceled, the channel sends context.Canceled and closes. Canceling ctx
// requests best-effort cancellation. Never wait on the channel during
// rendering or in a subscription.
//
// To capture the client result, arm the action with its Into method before
// dispatch; see [ActionInto]. dst is written only on a detached context,
// before the second value.
//
// ctx must belong to a Doors render or handler; otherwise Call panics.
func Call(ctx context.Context, a Action) <-chan error {
	core := ctx.Value(common.KeyCore).(core.Core)
	ch := make(chan error, 2)
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
	if !ctex.GetFrames(ctx).Detached() {
		scheduled := false
		core.Door().UserCall(ctx, prep.action, func() {
			scheduled = true
			ch <- nil
			close(ch)
		}, nil, func() {
			if !scheduled {
				ch <- context.Canceled
				close(ch)
			}
		}, prep.params)
		return ch
	}
	core.Door().UserCall(ctx, prep.action, func() {
		ch <- nil
	}, func(rm json.RawMessage, err error) {
		if err == nil && prep.decode != nil {
			err = prep.decode(rm)
		}
		ch <- err
		close(ch)
	}, func() {
		ch <- context.Canceled
		close(ch)
	}, prep.params)
	return ch
}
