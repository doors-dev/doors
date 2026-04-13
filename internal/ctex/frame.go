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

	"github.com/doors-dev/doors/internal/shredder"
)

func FrameInsert(ctx context.Context) (context.Context, *shredder.AfterFrame) {
	sh := &shredder.AfterFrame{}
	return context.WithValue(ctx, keyFrame, sh), sh
}

func AfterFrame(ctx context.Context) (*shredder.AfterFrame, bool) {
	f, ok := ctx.Value(keyFrame).(*shredder.AfterFrame)
	if !ok {
		return nil, false
	}
	return f, true
}

func Frame(ctx context.Context) shredder.SimpleFrame {
	f, ok := ctx.Value(keyFrame).(*shredder.AfterFrame)
	if !ok {
		return shredder.FreeFrame{}
	}
	return f
}

func FrameInfect(source context.Context, target context.Context) context.Context {
	f, ok := source.Value(keyFrame).(*shredder.AfterFrame)
	if !ok {
		return target
	}
	return context.WithValue(target, keyFrame, f)
}

func FrameRemove(ctx context.Context) context.Context {
	_, ok := ctx.Value(keyFrame).(*shredder.AfterFrame)
	if !ok {
		return ctx
	}
	return context.WithValue(ctx, keyFrame, nil)
}
