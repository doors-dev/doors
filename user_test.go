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
	"log/slog"
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

type helperDoor struct {
	inst *helperInstance
}

func (h helperDoor) UserCall(ctx context.Context, check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) {
	h.inst.UserCall(ctx, check, action, onResult, onCancel, params)
}

type helperShutdown struct{}

func (helperShutdown) Kill()                {}
func (helperShutdown) Logger() *slog.Logger { return slog.Default() }

func (h helperDoor) Instance() core.Instance {
	return h.inst
}

func (helperDoor) CleanFrame() shredder.SimpleFrame {
	return &shredder.ValveFrame{}
}

func (helperDoor) ReadyFrame() shredder.SimpleFrame {
	return &shredder.ValveFrame{}
}

func (helperDoor) Cinema() beam.Cinema {
	return nil
}

func (helperDoor) RegisterHook(func(context.Context, http.ResponseWriter, *http.Request) bool, func(context.Context)) (core.Hook, bool) {
	return core.Hook{}, false
}

func (helperDoor) ID() uint64 {
	return 1
}

func (helperDoor) Reload(context.Context) <-chan error {
	ch := make(chan error)
	close(ch)
	return ch
}

func (helperDoor) RootCore() core.Core {
	return nil
}

type helperDoorWithRoot struct {
	inst *helperInstance
	root core.Core
}

func (h helperDoorWithRoot) Instance() core.Instance {
	return h.inst
}

func (helperDoorWithRoot) CleanFrame() shredder.SimpleFrame {
	return &shredder.ValveFrame{}
}

func (helperDoorWithRoot) ReadyFrame() shredder.SimpleFrame {
	return &shredder.ValveFrame{}
}

// UserCall implements [core.Door].
func (h helperDoorWithRoot) UserCall(ctx context.Context, check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) {
	panic("unimplemented")
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

func (helperDoorWithRoot) Reload(context.Context) <-chan error {
	ch := make(chan error)
	close(ch)
	return ch
}

func (h helperDoorWithRoot) RootCore() core.Core {
	return h.root
}

type helperInstance struct {
	expire         time.Duration
	conf           common.Conf
	lastCallAction action.Action
	lastCallParams action.CallParams
	callCheckErr   error
	runtime        shredder.Runtime
	session        *helperSession
	location       beam.Source[path.Location]
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

func (h *helperInstance) UserCall(_ context.Context, _ func() bool, act action.Action, onResult func(json.RawMessage, error), _ func(), params action.CallParams) {
	h.lastCallAction = act
	h.lastCallParams = params
	if onResult != nil && h.callCheckErr != nil {
		onResult(nil, h.callCheckErr)
	}
}

func (h *helperInstance) CSPCollector() common.CSPCollector {
	return (&common.CSP{}).NewCollector()
}

func (h *helperInstance) ModuleRegistry() core.ModuleRegistry {
	return nil
}

func (h *helperInstance) LastSeen() time.Time {
	return time.Time{}
}

func (h *helperInstance) ID() string {
	return "instance-1"
}

func (h *helperInstance) RootID() uint64 {
	return 0
}

func (h *helperInstance) Conf() *common.Conf {
	return &h.conf
}

func (h *helperInstance) NewID() uint64 {
	return 2
}

func (h *helperInstance) Runtime() shredder.Runtime {
	return h.runtime
}

func (h *helperInstance) SetStatus(int) {}

func (h *helperInstance) Session() core.Session {
	return h.session
}

func (h *helperInstance) Store() ctex.Store {
	return ctex.NewStore()
}

func (h *helperInstance) Location() beam.Source[path.Location] {
	return h.location
}

func (h *helperInstance) Kill() {}

func (h *helperInstance) Logger() *slog.Logger {
	return slog.Default()
}

func (h *helperInstance) TitleMeta() core.TitleMeta {
	return nil
}

type helperApp struct {
	conf *common.Conf
}

func (h *helperApp) PathMaker() path.PathMaker {
	return path.NewPathMaker("__Host-", "", "")
}

func (h *helperApp) ResourceRegistry() resources.Registry {
	return nil
}

func (h *helperApp) Conf() *common.Conf {
	return h.conf
}

func (h *helperApp) Logger() *slog.Logger {
	return slog.Default()
}

func (h *helperApp) Draining() bool {
	return false
}

func (h *helperApp) PrinterMiddleware() func(next gox.Printer) gox.Printer {
	return func(next gox.Printer) gox.Printer { return next }
}

type helperSession struct {
	inst   *helperInstance
	app    *helperApp
	ctx    context.Context
	cancel context.CancelFunc
}

func (h *helperSession) Logger() *slog.Logger {
	return h.app.Logger()
}

func (h *helperSession) App() core.App {
	return h.app
}

func (h *helperSession) Context() context.Context {
	return h.ctx
}

func (h *helperSession) Store() ctex.Store {
	return ctex.NewStore()
}

func (h *helperSession) LastSeen() time.Time {
	return time.Time{}
}

func (h *helperSession) ID() string {
	return "session-1"
}

func (h *helperSession) Expire(d time.Duration) {
	h.inst.expire = d
}

func (h *helperSession) Kill() {
	h.cancel()
}

type helperLocation struct {
	Home bool   `path:"/home"`
	Tag  string `query:"tag"`
}

func helperContext(t *testing.T) (context.Context, *helperInstance) {
	t.Helper()
	conf := common.Conf{}
	common.InitDefaults(&conf)
	inst := &helperInstance{
		conf:     conf,
		location: beam.NewSource(path.Location{}, path.EqualLocation, false),
	}
	ctx, cancel := context.WithCancel(context.Background())
	inst.session = &helperSession{inst: inst, app: &helperApp{conf: &inst.conf}, ctx: ctx, cancel: cancel}
	ctx = context.WithValue(ctx, common.KeySession, inst.session)
	inst.session.ctx = ctx
	return context.WithValue(ctx, common.KeyCore, core.NewCore(helperDoor{inst: inst})), inst
}

func helperContextWithRoot(t *testing.T) (context.Context, *helperInstance, core.Core) {
	t.Helper()
	ctx, inst := helperContext(t)
	inst.runtime = shredder.NewRuntime(context.Background(), 1, helperShutdown{})
	t.Cleanup(func() {
		inst.runtime.Cancel()
	})
	root := core.NewCore(helperDoor{inst: inst})
	ctx = context.WithValue(ctx, common.KeyCore, core.NewCore(helperDoorWithRoot{inst: inst, root: root}))
	return ctx, inst, root
}

func TestUserHelpers(t *testing.T) {
	ctx, inst := helperContext(t)
	SessionExpire(ctx, time.Hour)
	if inst.expire != time.Hour {
		t.Fatalf("unexpected session expire duration: %v", inst.expire)
	}

	location, err := NewLocation(helperLocation{Home: true, Tag: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if location.String() != "/home?tag=x" {
		t.Fatalf("unexpected location: %q", location.String())
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

func TestSessionContextCanceledBySessionEnd(t *testing.T) {
	ctx, _ := helperContext(t)
	sessionCtx := SessionContext(ctx)
	SessionEnd(ctx)

	select {
	case <-sessionCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("expected SessionEnd to cancel SessionContext")
	}
}

func TestDetachedContextKeepsOwnerAndClearsFrame(t *testing.T) {
	ctx, _ := helperContext(t)
	ctx = context.WithValue(ctx, "value", "kept")
	ctx, _ = ctex.AfterFrameInsert(ctx)
	owner, _ := ctx.Value(common.KeyCore).(core.Core)
	base, cancel := context.WithCancel(ctx)
	free := DetachedContext(base)

	if !ctex.IsFreeCtx(free) {
		t.Fatal("expected DetachedContext to mark context as free")
	}
	if free.Value("value") != "kept" {
		t.Fatal("expected DetachedContext to preserve context values")
	}
	if got, _ := free.Value(common.KeyCore).(core.Core); got != owner {
		t.Fatal("expected DetachedContext to keep the current Doors owner")
	}
	if _, ok := ctex.AfterFrame(free); ok {
		t.Fatal("expected DetachedContext to clear current frame binding")
	}

	cancel()
	select {
	case <-free.Done():
	default:
		t.Fatal("expected Free to keep current context lifecycle")
	}
}

func TestInstanceContextSwitchesToRootCoreAndRuntime(t *testing.T) {
	ctx, inst, root := helperContextWithRoot(t)
	ctx = context.WithValue(ctx, "value", "kept")
	ctx, _ = ctex.AfterFrameInsert(ctx)
	base, cancel := context.WithCancel(ctx)
	free := InstanceContext(base)

	if !ctex.IsFreeCtx(free) {
		t.Fatal("expected InstanceContext to mark context as free")
	}
	if free.Value("value") != "kept" {
		t.Fatal("expected InstanceContext to preserve context values")
	}
	if got, _ := free.Value(common.KeyCore).(core.Core); got != root {
		t.Fatal("expected InstanceContext to switch to root Doors context")
	}
	if _, ok := ctex.AfterFrame(free); ok {
		t.Fatal("expected InstanceContext to clear current frame binding")
	}

	cancel()
	select {
	case <-free.Done():
		t.Fatal("expected InstanceContext to use instance runtime lifecycle instead of current owner lifecycle")
	default:
	}

	inst.runtime.Cancel()
	select {
	case <-free.Done():
	default:
		t.Fatal("expected InstanceContext to follow instance runtime cancellation")
	}
}

func TestCallUsesSolitaireDisableGzip(t *testing.T) {
	ctx, inst := helperContext(t)

	Call[json.RawMessage](ctx, ActionEmit{Name: "plain", Arg: "hello"})
	emit, ok := inst.lastCallAction.(action.Emit)
	if !ok {
		t.Fatalf("expected emit action, got %T", inst.lastCallAction)
	}
	if emit.Payload.Type() != action.PayloadTextGZ {
		t.Fatalf("expected gzip text payload by default, got %v", emit.Payload.Type())
	}

	inst.conf.SolitaireDisableGzip = true
	Call[json.RawMessage](ctx, ActionEmit{Name: "plain", Arg: "hello"})
	emit, ok = inst.lastCallAction.(action.Emit)
	if !ok {
		t.Fatalf("expected emit action, got %T", inst.lastCallAction)
	}
	if emit.Payload.Type() != action.PayloadText {
		t.Fatalf("expected plain text payload when solitaire gzip is disabled, got %v", emit.Payload.Type())
	}
}

func TestCallUsesCanceledContext(t *testing.T) {
	ctx, inst := helperContext(t)
	canceled, cancel := context.WithCancel(ctx)
	cancel()

	Call[json.RawMessage](canceled, ActionEmit{Name: "still-runs", Arg: "hello"})

	emit, ok := inst.lastCallAction.(action.Emit)
	if !ok {
		t.Fatalf("expected emit action from canceled context, got %T", inst.lastCallAction)
	}
	if emit.Name != "still-runs" {
		t.Fatalf("expected canceled-context call to dispatch action %q, got %q", "still-runs", emit.Name)
	}
}

func TestSharedAttrRestoreOnUpdateError(t *testing.T) {
	ctx, inst := helperContext(t)
	shared := NewAShared("data-shared", "start")
	attrs := gox.NewAttrs()
	if err := shared.Modify(ctx, "div", attrs); err != nil {
		t.Fatal(err)
	}
	inst.callCheckErr = errors.New("boom")

	shared.Update(ctx, "next")

	set, ok := inst.lastCallAction.(action.AttrSet)
	if !ok {
		t.Fatalf("expected AttrSet action, got %T", inst.lastCallAction)
	}
	if set.Name != "data-shared" {
		t.Fatalf("expected attr name %q, got %q", "data-shared", set.Name)
	}
	if set.Value == nil || *set.Value != "next" {
		t.Fatalf("expected attempted update value %q, got %v", "next", set.Value)
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
	ctx, inst := helperContext(t)
	shared := NewAShared("data-shared", "start")
	attrs := gox.NewAttrs()
	if err := shared.Modify(ctx, "div", attrs); err != nil {
		t.Fatal(err)
	}
	inst.callCheckErr = errors.New("boom")

	shared.Disable(ctx)

	remove, ok := inst.lastCallAction.(action.AttrSet)
	if !ok {
		t.Fatalf("expected AttrSet action, got %T", inst.lastCallAction)
	}
	if remove.ID == 0 {
		t.Fatal("expected dynamic attr id on remove action")
	}
	if remove.Value != nil {
		t.Fatalf("expected nil value on remove action, got %q", *remove.Value)
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

func TestSetterSetValueSemantics(t *testing.T) {
	ctx, inst := helperContext(t)
	setter := &Setter{}
	str := func(s string) *string { return &s }
	cases := []struct {
		name  string
		value any
		want  *string
	}{
		{"string", "v", str("v")},
		{"int", 42, str("42")},
		{"true bare", true, str("")},
		{"false removes", false, nil},
		{"nil removes", nil, nil},
		{"output", Class("a").Add("b"), str("a b")},
	}
	var id uint64
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Call[any](ctx, setter.Set("data-x", tc.value))
			act, ok := inst.lastCallAction.(action.AttrSet)
			if !ok {
				t.Fatalf("expected AttrSet action, got %T", inst.lastCallAction)
			}
			if act.Name != "data-x" {
				t.Fatalf("expected attr name %q, got %q", "data-x", act.Name)
			}
			if id == 0 {
				id = act.ID
			} else if act.ID != id {
				t.Fatalf("expected stable setter id %d, got %d", id, act.ID)
			}
			if tc.want == nil {
				if act.Value != nil {
					t.Fatalf("expected remove (nil value), got %q", *act.Value)
				}
				return
			}
			if act.Value == nil {
				t.Fatalf("expected value %q, got remove (nil value)", *tc.want)
			}
			if *act.Value != *tc.want {
				t.Fatalf("expected value %q, got %q", *tc.want, *act.Value)
			}
		})
	}
}
