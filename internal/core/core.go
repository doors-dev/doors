package core

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/license"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
)

type Hook struct {
	DoorID uint64
	HookID uint64
	Cancel context.CancelFunc
}

type Link struct {
	Location *path.Location
	On       func(context.Context)
}


func (h *Link) Path() (string, bool) {
	if h.Location == nil {
		return "", false
	}
	return h.Location.String(), true
}

func (h *Link) ClickHandler() (func(context.Context), bool) {
	if h.On == nil {
		return nil, false
	}
	return h.On, true
}

type ModuleRegistry interface {
	Add(specifier string, path string)
}

type Instance interface {
	CallCtx(ctx context.Context, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) context.CancelFunc
	CallCheck(check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams)
	CSPCollector() *common.CSPCollector
	ModuleRegistry() ModuleRegistry
	ResourceRegistry() *resources.Registry
	ID() string
	RootID() uint64
	Conf() *common.SystemConf
	Detached() bool
	NewID() uint64
	NewLink(any) (Link, error)
	Runtime() shredder.Runtime
	License() license.License
	SetStatus(int)
}

type Door interface {
	Cinema() beam.Cinema
	RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (Hook, bool)
	ID() uint64
}

func NewCore(inst Instance, door Door) Core {
	return &core{
		door: door,
		inst: inst,
	}
}

type Core = *core

var _ beam.Core = &core{}

type core struct {
	door Door
	inst Instance
}

func (c Core) Runtime() shredder.Runtime {
	return c.inst.Runtime()
}

func (c Core) Cinema() beam.Cinema {
	return c.door.Cinema()
}

type Done = bool

func (c Core) SetStatus(status int) {
	c.inst.SetStatus(status)
}

func (c Core) License() license.License {
	return c.inst.License()
}

func (c Core) DoorID() uint64 {
	return c.door.ID()
}

func (c Core) NewLink(m any) (Link, error) {
	return c.inst.NewLink(m)
}

func (c Core) Detached() bool {
	return c.inst.Detached()
}

func (c Core) InstanceID() string {
	return c.inst.ID()
}

func (c Core) RootID() uint64 {
	return c.inst.RootID()
}

func (c Core) NewID() uint64 {
	return c.inst.NewID()
}

func (c Core) Conf() *common.SystemConf {
	return c.inst.Conf()
}

func (c Core) ResourceRegistry() *resources.Registry {
	return c.inst.ResourceRegistry()
}

func (c Core) ModuleRegistry() ModuleRegistry {
	return c.inst.ModuleRegistry()
}

func (c Core) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done, onCancel func(ctx context.Context)) (Hook, bool) {
	return c.door.RegisterHook(onTrigger, onCancel)
}

func (c Core) CallCtx(ctx context.Context, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) context.CancelFunc {
	return c.inst.CallCtx(ctx, action, onResult, onCancel, params)
}

func (c Core) CallCheck(check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) {
	c.inst.CallCheck(check, action, onResult, onCancel, params)
}

func (c Core) CSPCollector() *common.CSPCollector {
	return c.inst.CSPCollector()
}
