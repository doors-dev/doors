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

type viewKind uint8

const (
	viewRender viewKind = iota
	viewDetached
	viewSettle
	viewInfected
	viewPlain
)

type viewSource uint8

const (
	fromValues viewSource = iota
	fromLife
	fromNone
	fromFrames
)

type WithValueKeys interface {
	IsWithValueKey(key any) bool
}

type view struct {
	life     context.Context
	values   context.Context
	keys     WithValueKeys
	frames   *Frames
	kind     viewKind
	document bool
	replace  bool
}

var _ context.Context = &view{}

func (v *view) Deadline() (time.Time, bool) {
	return v.life.Deadline()
}

func (v *view) Done() <-chan struct{} {
	return v.life.Done()
}

func (v *view) Err() error {
	return v.life.Err()
}

func (v *view) AfterFunc(f func()) func() bool {
	return context.AfterFunc(v.life, f)
}

func (v *view) Value(key any) any {
	switch {
	case v.frames != nil && key == common.KeyFrame:
		return v.frames
	case v.document && key == common.KeyDocument:
		return true
	case v.replace && key == common.KeyHistoryReplace:
		return true
	}
	switch v.route(key) {
	case fromLife:
		return v.life.Value(key)
	case fromNone:
		return nil
	case fromFrames:
		return v.computedFrames()
	default:
		val := v.values.Value(key)
		if val == nil && v.keys != nil && v.keys.IsWithValueKey(key) {
			return v.life.Value(key)
		}
		return val
	}
}

func (v *view) route(key any) viewSource {
	switch v.kind {
	case viewRender:
		if common.IsKey(key) {
			return fromLife
		}
		return fromValues
	case viewDetached, viewSettle:
		switch key {
		case common.KeyDocument:
			return fromNone
		case common.KeyFrame:
			return fromFrames
		}
		return fromValues
	case viewInfected:
		switch key {
		case common.KeyCore, common.KeySession, common.KeyDocument:
			return fromLife
		case common.KeyFrame:
			return fromFrames
		}
		return fromValues
	}
	return fromValues
}

func (v *view) bare() bool {
	return v.frames == nil && !v.document && !v.replace
}

func (v *view) computedFrames() any {
	switch v.kind {
	case viewDetached:
		return Frames{detached: true}
	case viewSettle:
		return GetFrames(v.values).Settle()
	case viewInfected:
		return v.infectedFrames()
	}
	return nil
}

func instanceID(ctx context.Context) (string, bool) {
	c, ok := ctx.Value(common.KeyCore).(interface{ InstanceID() string })
	if !ok {
		return "", false
	}
	return c.InstanceID(), true
}

func (v *view) infectedFrames() any {
	fs, ok := lookupFrames(v.values)
	if !ok {
		return nil
	}
	inst := fs.inst
	if fs.detached {
		inst, _ = instanceID(v.values)
	}
	if owner, ok := instanceID(v.life); ok && inst == owner {
		return fs
	}
	return Frames{sync: fs.sync, init: fs.init}
}

func asView(ctx context.Context) view {
	if v, ok := ctx.(*view); ok {
		return *v
	}
	return view{life: ctx, values: ctx, kind: viewPlain}
}

func withFrames(ctx context.Context, fs Frames) context.Context {
	v := asView(ctx)
	v.frames = &fs
	return &v
}

func NewRenderContext(system, user context.Context, keys WithValueKeys) context.Context {
	if v, ok := system.(*view); ok && v.kind == viewRender && v.bare() {
		system = v.life
	}
	if user == nil {
		return system
	}
	return &view{life: system, values: user, keys: keys, kind: viewRender}
}

func NewDocumentContext(ctx context.Context) context.Context {
	v := asView(ctx)
	if v.kind == viewRender {
		v.life = context.WithValue(v.life, common.KeyDocument, true)
		return &v
	}
	v.document = true
	return &v
}

func NewReplaceContext(ctx context.Context) context.Context {
	v := asView(ctx)
	v.replace = true
	return &v
}

func UserContext(ctx, parent context.Context) context.Context {
	if v, ok := ctx.(*view); ok && v.kind == viewRender && v.bare() {
		return v.values
	}
	if ctx == parent {
		return nil
	}
	return ctx
}

func NewDetachedContext(ctx, runtime context.Context, fake bool) context.Context {
	kind := viewDetached
	if fake {
		kind = viewSettle
	}
	return &view{life: runtime, values: ctx, kind: kind}
}

func NewInfectedContext(source, target context.Context) context.Context {
	return &view{life: target, values: source, kind: viewInfected}
}
