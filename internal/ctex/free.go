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

package ctex

import (
	"context"
	"time"

	"github.com/doors-dev/doors/internal/common"
)

func NewDetachedContext(ctx context.Context, runtime context.Context, preserveBatch bool) context.Context {
	return detachedContext{ctx, runtime, preserveBatch}
}

type detachedContext struct {
	ctx           context.Context
	runtime       context.Context
	preserveBatch bool
}

func (f detachedContext) Deadline() (deadline time.Time, ok bool) {
	return f.runtime.Deadline()
}

func (f detachedContext) Done() <-chan struct{} {
	return f.runtime.Done()
}

func (f detachedContext) Err() error {
	return f.runtime.Err()
}

func (f detachedContext) Value(key any) any {
	if key == common.KeyDocument {
		return nil
	}
	if key == common.KeyFrame {
		if !f.preserveBatch {
			return nil
		}
		fs := GetFrames(f.ctx)
		return Frames{batch: fs.batch, inst: fs.inst, batchFrame: fs.batchFrame}
	}
	return f.ctx.Value(key)
}

var _ context.Context = detachedContext{}

func IsDetachedCtx(ctx context.Context) bool {
	_, ok := ctx.(detachedContext)
	return ok
}
