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

package core

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type Hook struct {
	HookID uint64
	Cancel context.CancelFunc
}

type ModuleRegistry interface {
	Add(specifier string, path string)
}

type App interface {
	PathMaker() path.PathMaker
	ResourceRegistry() resources.Registry
	Conf() *common.Conf
}

type Session interface {
	App() App
	ID() string
	Expire(time.Duration)
	Context() context.Context
	Kill()
}

type TitleMeta interface {
	gox.Editor
	UpdateTitle(value string, attrs gox.Attrs) context.CancelFunc
	UpdateMeta(prop bool, name string, attrs gox.Attrs) context.CancelFunc
}

type Instance interface {
	Session() Session
	Store() ctex.Store
	UserCall(ctx context.Context, check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams)
	CSPCollector() common.CSPCollector
	ModuleRegistry() ModuleRegistry
	ID() string
	RootID() uint64
	NewID() uint64
	Runtime() shredder.Runtime
	SetStatus(int)
	Location() beam.Source[path.Location]
	Kill()
	TitleMeta() TitleMeta
}

type Door interface {
	Instance() Instance
	Cinema() beam.Cinema
	ID() uint64
	RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (Hook, bool)
	Reload(ctx context.Context)
	XReload(ctx context.Context) <-chan error
	RootCore() Core
	UserCall(ctx context.Context, check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams)
	Clean(func())
}

func NewCore(door Door) Core {
	return &core{
		door: door,
	}
}

type Core = *core

var _ beam.Core = &core{}

type core struct {
	door Door
}

func (c Core) Location() beam.Source[path.Location] {
	return c.Instance().Location()
}

func (c Core) Reload(ctx context.Context) {
	c.door.Reload(ctx)
}

func (c Core) XReload(ctx context.Context) <-chan error {
	return c.Door().XReload(ctx)
}

func (c Core) RootCore() Core {
	return c.door.RootCore()
}

func (c Core) Clean(f func()) {
	c.door.Clean(f)
}

func (c Core) SessionContext() context.Context {
	return c.Instance().Session().Context()
}

func (c Core) TitleMeta() TitleMeta {
	return c.Instance().TitleMeta()
}

func (c Core) PathMaker() path.PathMaker {
	return c.Instance().Session().App().PathMaker()
}

func (c Core) Door() Door {
	return c.door
}

func (c Core) DoorID() uint64 {
	return c.door.ID()
}

func (c Core) Instance() Instance {
	return c.door.Instance()
}

func (c Core) SessionExpire(d time.Duration) {
	c.Instance().Session().Expire(d)
}

func (c Core) SessionEnd() {
	c.Instance().Session().Kill()
}

func (c Core) InstanceEnd() {
	c.Instance().Kill()
}

func (c Core) SessionID() string {
	return c.Instance().Session().ID()
}

func (c Core) Runtime() shredder.Runtime {
	return c.Instance().Runtime()
}

func (c Core) Cinema() beam.Cinema {
	return c.door.Cinema()
}

func (c Core) SetStatus(status int) {
	c.Instance().SetStatus(status)
}

func (c Core) InstanceID() string {
	return c.Instance().ID()
}

func (c Core) RootID() uint64 {
	return c.Instance().RootID()
}

func (c Core) NewID() uint64 {
	return c.Instance().NewID()
}

func (c Core) Conf() *common.Conf {
	return c.Instance().Session().App().Conf()
}

func (c Core) ResourceRegistry() resources.Registry {
	return c.Instance().Session().App().ResourceRegistry()
}

func (c Core) ModuleRegistry() ModuleRegistry {
	return c.Instance().ModuleRegistry()
}

func (c Core) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (Hook, bool) {
	return c.door.RegisterHook(onTrigger, onCancel)
}

func (c Core) Call(ctx context.Context, check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) {
	c.door.UserCall(ctx, check, action, onResult, onCancel, params)
}

func (c Core) CSPCollector() common.CSPCollector {
	return c.Instance().CSPCollector()
}
