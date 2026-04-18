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

type Frames struct {
	after *shredder.AfterFrame
	sync  shredder.Frame
}

func (f Frames) Send() shredder.SimpleFrame {
	if f.after == nil {
		return shredder.FreeFrame{}
	}
	return f.after
}

func (f Frames) Render() shredder.Frame {
	if f.sync == nil {
		return shredder.FreeFrame{}
	}
	return f.sync
}

func (f Frames) JoinedFrame() shredder.Frame {
	if f.after == nil && f.sync == nil {
		return shredder.FreeFrame{}
	}
	if f.sync == nil {
		return shredder.Join(false, f.after)
	}
	if f.after == nil {
		return f.sync
	}
	return shredder.Join(false, f.sync, f.after)
}

func AfterFrameInsert(ctx context.Context) (context.Context, *shredder.AfterFrame) {
	fs := Frames{
		after: &shredder.AfterFrame{},
	}
	return context.WithValue(ctx, keyFrame, fs), fs.after
}

func SyncFrameInsert(ctx context.Context, frame shredder.Frame) context.Context {
	fs, ok := ctx.Value(keyFrame).(Frames)
	if ok {
		fs.sync = frame
	} else {
		fs = Frames{
			sync: frame,
		}
	}
	return context.WithValue(ctx, keyFrame, fs)
}

func AfterFrame(ctx context.Context) (*shredder.AfterFrame, bool) {
	fs, ok := ctx.Value(keyFrame).(Frames)
	if !ok {
		return nil, false
	}
	return fs.after, fs.after != nil
}

func GetFrames(ctx context.Context) Frames {
	f, _ := ctx.Value(keyFrame).(Frames)
	return f
}

func FrameInfect(source context.Context, target context.Context) context.Context {
	fs, ok := source.Value(keyFrame).(Frames)
	if !ok {
		return target
	}
	return context.WithValue(target, keyFrame, fs)
}

func FrameRemove(ctx context.Context) context.Context {
	_, ok := ctx.Value(keyFrame).(Frames)
	if !ok {
		return ctx
	}
	return context.WithValue(ctx, keyFrame, nil)
}
