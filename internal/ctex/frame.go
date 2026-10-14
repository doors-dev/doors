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
	sync     shredder.ReleaseFrame
	init     shredder.Frame
	after    *shredder.AfterFrame
	inst     string
	group    uint64
	batch    uint64
	valve    *shredder.ValveFrame
	detached bool
}

func (f Frames) Owner() bool {
	return f.after != nil
}

func (f Frames) Detached() bool {
	return f.detached
}

func (f Frames) Group() uint64 {
	return f.group
}

func (f Frames) Batch() uint64 {
	return f.batch
}

func (f Frames) Inst() string {
	return f.inst
}

func (f Frames) BatchFrame() (*shredder.ValveFrame, bool) {
	return f.valve, f.valve != nil
}

func (f Frames) AfterFrame() (*shredder.AfterFrame, bool) {
	return f.after, f.after != nil
}

func (f Frames) Settle() Frames {
	return Frames{inst: f.inst, group: f.group, batch: f.batch}
}

func (f Frames) Call() shredder.Frame {
	if f.after == nil {
		return shredder.FreeFrame{}
	}
	return f.after
}

func (f Frames) Render() shredder.ReleaseFrame {
	if f.sync == nil {
		return shredder.FreeFrame{}
	}
	return f.sync
}

func (f Frames) InitFrame(ctx context.Context) shredder.ReleaseFrame {
	var after shredder.Frame = nil
	var sync shredder.ReleaseFrame = nil
	var init shredder.Frame = nil
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
	return shredder.Join(ctx, init, sync, after)
}

func NewAfterFrame(ctx context.Context, inst string, group uint64) (context.Context, *shredder.AfterFrame) {
	prev := GetFrames(ctx)
	fs := Frames{
		sync:  prev.sync,
		init:  prev.init,
		after: &shredder.AfterFrame{},
		inst:  inst,
		group: group,
	}
	return context.WithValue(ctx, common.KeyFrame, fs), fs.after
}

func NewBatch(ctx context.Context, id uint64) (context.Context, *shredder.ValveFrame) {
	fs := GetFrames(ctx)
	fs.batch = id
	fs.valve = &shredder.ValveFrame{}
	return context.WithValue(ctx, common.KeyFrame, fs), fs.valve
}

func SyncFrameInsert(ctx context.Context, sync shredder.ReleaseFrame, init shredder.ReleaseFrame) context.Context {
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
	case common.KeyCore, common.KeySession, common.KeyDocument:
		return c.target.Value(key)
	case common.KeyFrame:
		fs, ok := c.source.Value(key).(Frames)
		if !ok {
			return nil
		}
		owner, _ := c.target.Value(common.KeyCore).(interface{ InstanceID() string })
		if owner == nil {
			return Frames{sync: fs.sync, init: fs.init}
		}
		inst := fs.inst
		if fs.detached {
			if source, ok := c.source.Value(common.KeyCore).(interface{ InstanceID() string }); ok {
				inst = source.InstanceID()
			}
		}
		if inst == owner.InstanceID() {
			return fs
		}
		return Frames{sync: fs.sync, init: fs.init}
	default:
		return c.source.Value(key)
	}
}

var _ context.Context = infectedContext{}
