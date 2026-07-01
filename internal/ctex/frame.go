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
	"github.com/doors-dev/doors/internal/shredder"
)

type Frames struct {
	after *shredder.AfterFrame
	sync  shredder.Frame
	init  shredder.AnyFrame
}

func (f Frames) Call() shredder.SimpleFrame {
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

func (f Frames) InitFrame(ctx context.Context) shredder.Frame {
	var after shredder.AnyFrame = nil
	var sync shredder.Frame = nil
	var init shredder.AnyFrame = nil
	if f.after == nil {
		after = shredder.FreeFrame{}
	} else {
		after = f.after
	}
	if f.sync == nil {
		sync = shredder.FreeFrame{}
	} else {
		sync = f.sync
	}
	if f.init == nil {
		init = shredder.FreeFrame{}
	} else {
		init = f.init
	}
	return shredder.Join(ctx, false, init, sync, after)
}

func AfterFrameInsert(ctx context.Context) (context.Context, *shredder.AfterFrame) {
	fs := Frames{
		after: &shredder.AfterFrame{},
	}
	return context.WithValue(ctx, common.KeyFrame, fs), fs.after
}

func SyncFrameInsert(ctx context.Context, sync shredder.Frame, init shredder.Frame) context.Context {
	fs, ok := ctx.Value(common.KeyFrame).(Frames)
	if ok {
		fs.sync = sync
		fs.init = init
	} else {
		fs = Frames{
			sync: sync,
			init: init,
		}
	}
	return context.WithValue(ctx, common.KeyFrame, fs)
}

func AfterFrame(ctx context.Context) (*shredder.AfterFrame, bool) {
	fs, ok := ctx.Value(common.KeyFrame).(Frames)
	if !ok {
		return nil, false
	}
	return fs.after, fs.after != nil
}

func GetFrames(ctx context.Context) Frames {
	f, _ := ctx.Value(common.KeyFrame).(Frames)
	return f
}

func FrameInfect(source context.Context, target context.Context) context.Context {
	return infectedContext{
		source: source,
		target: target,
	}
}

type infectedContext struct {
	source context.Context
	target context.Context
}

func (c infectedContext) Deadline() (deadline time.Time, ok bool) {
	return c.target.Deadline()
}

func (c infectedContext) Done() <-chan struct{} {
	return c.target.Done()
}

func (c infectedContext) Err() error {
	return c.target.Err()
}

func (c infectedContext) Value(key any) any {
	switch key {
	case common.KeyCore, common.KeySession:
		return c.target.Value(key)
	default:
		return c.source.Value(key)
	}
}

var _ context.Context = infectedContext{}
