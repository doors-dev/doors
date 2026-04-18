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
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type helperDoor struct{}

type helperShutdown struct{}

func (helperShutdown) Shutdown() {}

func (helperDoor) Cinema() beam.Cinema {
	return nil
}

func (helperDoor) RegisterHook(func(context.Context, http.ResponseWriter, *http.Request) bool, func(context.Context)) (core.Hook, bool) {
	return core.Hook{}, false
}

func (helperDoor) ID() uint64 {
	return 1
}

func (helperDoor) Reload(context.Context) {}

func (helperDoor) XReload(context.Context) <-chan error {
	ch := make(chan error)
	close(ch)
	return ch
}

func (helperDoor) RootCore() core.Core {
	return nil
}

type helperDoorWithRoot struct {
	root core.Core
}

func (helperDoorWithRoot) Cinema() beam.Cinema {
	return nil
}

func (helperDoorWithRoot) RegisterHook(func(context.Context, http.ResponseWriter, *http.Request) bool, func(context.Context)) (core.Hook, bool) {
	return core.Hook{}, false
}

func (helperDoorWithRoot) ID() uint64 {
	return 1
}

func (helperDoorWithRoot) Reload(context.Context) {}

func (helperDoorWithRoot) XReload(context.Context) <-chan error {
	ch := make(chan error)
	close(ch)
	return ch
}

func (h helperDoorWithRoot) RootCore() core.Core {
	return h.root
}

type helperInstance struct {
	expire         time.Duration
	adapters       path.Adapters
	conf           common.SystemConf
	lastCallAction action.Action
	lastCallParams action.CallParams
	callCheckErr   error
	runtime        shredder.Runtime
}

func (h *helperInstance) CallCtx(_ context.Context, act action.Action, _ func(json.RawMessage, error), _ func(), params action.CallParams) context.CancelFunc {
	h.lastCallAction = act
	h.lastCallParams = params
	return func() {}
}

func (h *helperInstance) CallCheck(_ func() bool, act action.Action, onResult func(json.RawMessage, error), _ func(), params action.CallParams) {
	h.lastCallAction = act
	h.lastCallParams = params
	if onResult != nil && h.callCheckErr != nil {
		onResult(nil, h.callCheckErr)
	}
}

func (h *helperInstance) CSPCollector() *common.CSPCollector {
	return nil
}

func (h *helperInstance) ModuleRegistry() core.ModuleRegistry {
	return nil
}

func (h *helperInstance) ResourceRegistry() *resources.Registry {
	return nil
}

func (h *helperInstance) ID() string {
	return "instance-1"
}

func (h *helperInstance) RootID() uint64 {
	return 0
}

func (h *helperInstance) Conf() *common.SystemConf {
	return &h.conf
}

func (h *helperInstance) NewID() uint64 {
	return 2
}

func (h *helperInstance) NewLink(any) (core.Link, error) {
	return core.Link{}, nil
}

func (h *helperInstance) Runtime() shredder.Runtime {
	return h.runtime
}

func (h *helperInstance) SetStatus(int) {}

func (h *helperInstance) SessionExpire(d time.Duration) {
	h.expire = d
}

func (h *helperInstance) SessionEnd() {}

func (h *helperInstance) InstanceEnd() {}

func (h *helperInstance) SessionID() string {
	return "session-1"
}

func (h *helperInstance) Adapters() path.Adapters {
	return h.adapters
}

func (h *helperInstance) PathMaker() path.PathMaker {
	return path.NewPathMaker("")
}

func (h *helperInstance) UpdateTitle(string, gox.Attrs) {}

func (h *helperInstance) UpdateMeta(string, bool, gox.Attrs) {}

type helperLocation struct {
	Home bool   `path:"/home"`
	Tag  string `query:"tag"`
}

func helperContext(t *testing.T, adapters path.Adapters) (context.Context, *helperInstance) {
	t.Helper()
	inst := &helperInstance{adapters: adapters}
	return context.WithValue(context.Background(), ctex.KeyCore, core.NewCore(inst, helperDoor{})), inst
}

func helperContextWithRoot(t *testing.T, adapters path.Adapters) (context.Context, *helperInstance, core.Core) {
	t.Helper()
	inst := &helperInstance{adapters: adapters}
	inst.runtime = shredder.NewRuntime(context.Background(), 1, helperShutdown{})
	t.Cleanup(func() {
		inst.runtime.Cancel()
	})
	root := core.NewCore(inst, helperDoor{})
	ctx := context.WithValue(context.Background(), ctex.KeyCore, core.NewCore(inst, helperDoorWithRoot{root: root}))
	return ctx, inst, root
}

func TestUserHelpers(t *testing.T) {
	adapter, err := path.NewAdapter[helperLocation]()
	if err != nil {
		t.Fatal(err)
	}
	var adapters path.Adapters
	adapters.Add(adapter)

	ctx, inst := helperContext(t, adapters)
	SessionExpire(ctx, time.Hour)
	if inst.expire != time.Hour {
		t.Fatalf("unexpected session expire duration: %v", inst.expire)
	}

	location, err := NewLocation(ctx, helperLocation{Home: true, Tag: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if location.String() != "/home?tag=x" {
		t.Fatalf("unexpected location: %q", location.String())
	}

	emptyCtx, _ := helperContext(t, nil)
	if _, err := NewLocation(emptyCtx, helperLocation{Home: true}); err == nil {
		t.Fatal("expected missing adapter to fail")
	}

	if IDRand() == "" {
		t.Fatal("expected random id to be non-empty")
	}
	if IDRand() == IDRand() {
		t.Fatal("expected random ids to differ")
	}
	if IDString("hello") != IDString("hello") {
		t.Fatal("expected deterministic string id")
	}
	if IDString("hello") == IDString("world") {
		t.Fatal("expected distinct string ids")
	}
	if IDBytes([]byte("hello")) != IDBytes([]byte("hello")) {
		t.Fatal("expected deterministic byte id")
	}
	if IDBytes([]byte("hello")) == IDBytes([]byte("world")) {
		t.Fatal("expected distinct byte ids")
	}

	if ctex.IsFreeCtx(context.Background()) {
		t.Fatal("unexpected free context by default")
	}
}

func TestFreeKeepsOwnerAndClearsFrame(t *testing.T) {
	ctx, _ := helperContext(t, nil)
	ctx = context.WithValue(ctx, "value", "kept")
	ctx, _ = ctex.AfterFrameInsert(ctx)
	owner, _ := ctx.Value(ctex.KeyCore).(core.Core)
	base, cancel := context.WithCancel(ctx)
	free := Free(base)

	if !ctex.IsFreeCtx(free) {
		t.Fatal("expected Free to mark context as free")
	}
	if free.Value("value") != "kept" {
		t.Fatal("expected Free to preserve context values")
	}
	if got, _ := free.Value(ctex.KeyCore).(core.Core); got != owner {
		t.Fatal("expected Free to keep the current Doors owner")
	}
	if _, ok := ctex.AfterFrame(free); ok {
		t.Fatal("expected Free to clear current frame binding")
	}

	cancel()
	select {
	case <-free.Done():
	default:
		t.Fatal("expected Free to keep current context lifecycle")
	}
}

func TestFreeRootSwitchesToRootCoreAndRuntime(t *testing.T) {
	ctx, inst, root := helperContextWithRoot(t, nil)
	ctx = context.WithValue(ctx, "value", "kept")
	ctx, _ = ctex.AfterFrameInsert(ctx)
	base, cancel := context.WithCancel(ctx)
	free := FreeRoot(base)

	if !ctex.IsFreeCtx(free) {
		t.Fatal("expected FreeRoot to mark context as free")
	}
	if free.Value("value") != "kept" {
		t.Fatal("expected FreeRoot to preserve context values")
	}
	if got, _ := free.Value(ctex.KeyCore).(core.Core); got != root {
		t.Fatal("expected FreeRoot to switch to root Doors context")
	}
	if _, ok := ctex.AfterFrame(free); ok {
		t.Fatal("expected FreeRoot to clear current frame binding")
	}

	cancel()
	select {
	case <-free.Done():
		t.Fatal("expected FreeRoot to use instance runtime lifecycle instead of current owner lifecycle")
	default:
	}

	inst.runtime.Cancel()
	select {
	case <-free.Done():
	default:
		t.Fatal("expected FreeRoot to follow instance runtime cancellation")
	}
}

func TestCallUsesSolitaireDisableGzip(t *testing.T) {
	ctx, inst := helperContext(t, nil)

	Call(ctx, ActionEmit{Name: "plain", Arg: "hello"})
	emit, ok := inst.lastCallAction.(action.Emit)
	if !ok {
		t.Fatalf("expected emit action, got %T", inst.lastCallAction)
	}
	if emit.Payload.Type() != action.PayloadTextGZ {
		t.Fatalf("expected gzip text payload by default, got %v", emit.Payload.Type())
	}

	inst.conf.SolitaireDisableGzip = true
	Call(ctx, ActionEmit{Name: "plain", Arg: "hello"})
	emit, ok = inst.lastCallAction.(action.Emit)
	if !ok {
		t.Fatalf("expected emit action, got %T", inst.lastCallAction)
	}
	if emit.Payload.Type() != action.PayloadText {
		t.Fatalf("expected plain text payload when solitaire gzip is disabled, got %v", emit.Payload.Type())
	}
}

func TestCallUsesCanceledContext(t *testing.T) {
	ctx, inst := helperContext(t, nil)
	canceled, cancel := context.WithCancel(ctx)
	cancel()

	Call(canceled, ActionEmit{Name: "still-runs", Arg: "hello"})

	emit, ok := inst.lastCallAction.(action.Emit)
	if !ok {
		t.Fatalf("expected emit action from canceled context, got %T", inst.lastCallAction)
	}
	if emit.Name != "still-runs" {
		t.Fatalf("expected canceled-context call to dispatch action %q, got %q", "still-runs", emit.Name)
	}
}

func TestSharedAttrRestoreOnUpdateError(t *testing.T) {
	ctx, inst := helperContext(t, nil)
	shared := NewAShared("data-shared", "start")
	attrs := gox.NewAttrs()
	if err := shared.Modify(ctx, "div", attrs); err != nil {
		t.Fatal(err)
	}
	inst.callCheckErr = errors.New("boom")

	shared.Update(ctx, "next")

	set, ok := inst.lastCallAction.(*action.DynaSet)
	if !ok {
		t.Fatalf("expected DynaSet action, got %T", inst.lastCallAction)
	}
	if set.Value != "next" {
		t.Fatalf("expected attempted update value %q, got %q", "next", set.Value)
	}
	if shared.value != "start" {
		t.Fatalf("expected shared value restored to %q, got %q", "start", shared.value)
	}
	if !shared.enable {
		t.Fatal("expected shared attr to stay enabled after restore")
	}
	if shared.seq != 0 {
		t.Fatalf("expected restore to rewind seq to 0, got %d", shared.seq)
	}
}

func TestSharedAttrRestoreOnDisableError(t *testing.T) {
	ctx, inst := helperContext(t, nil)
	shared := NewAShared("data-shared", "start")
	attrs := gox.NewAttrs()
	if err := shared.Modify(ctx, "div", attrs); err != nil {
		t.Fatal(err)
	}
	inst.callCheckErr = errors.New("boom")

	shared.Disable(ctx)

	remove, ok := inst.lastCallAction.(*action.DynaRemove)
	if !ok {
		t.Fatalf("expected DynaRemove action, got %T", inst.lastCallAction)
	}
	if remove.ID == 0 {
		t.Fatal("expected dynamic attr id on remove action")
	}
	if !shared.enable {
		t.Fatal("expected shared attr enable flag restored after failed disable")
	}
	if shared.value != "start" {
		t.Fatalf("expected shared value preserved as %q, got %q", "start", shared.value)
	}
	if shared.seq != 0 {
		t.Fatalf("expected restore to rewind seq to 0, got %d", shared.seq)
	}
}
