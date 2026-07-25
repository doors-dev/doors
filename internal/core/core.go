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
	"log/slog"
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
	Logger() *slog.Logger
	PathMaker() path.PathMaker
	ResourceRegistry() resources.Registry
	Conf() *common.Conf
	Draining() bool
	PrinterMiddleware() func(next gox.Printer) gox.Printer
}

type Session interface {
	Logger() *slog.Logger
	App() App
	ID() string
	Expire(time.Duration)
	LastSeen() time.Time
	Context() context.Context
	Store() ctex.Store
	Kill()
}

type TitleMeta interface {
	gox.Editor
	UpdateTitle(value string, attrs gox.Attrs) context.CancelFunc
	UpdateMeta(prop bool, name string, attrs gox.Attrs) context.CancelFunc
}

type Instance interface {
	Session() Session
	Logger() *slog.Logger
	Store() ctex.Store
	UserCall(ctx context.Context, check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams)
	CSPCollector() common.CSPCollector
	ModuleRegistry() ModuleRegistry
	ID() string
	LastSeen() time.Time
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
	Reload(ctx context.Context) <-chan error
	RootCore() Core
	UserCall(ctx context.Context, check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams)
	CleanFrame() shredder.SimpleFrame
	ReadyFrame() shredder.SimpleFrame
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

func (c Core) App() App {
	return c.door.Instance().Session().App()
}

func (c Core) Door() Door {
	return c.door
}

func (c Core) Instance() Instance {
	return c.door.Instance()
}

func (c Core) Cinema() beam.Cinema {
	return c.door.Cinema()
}
