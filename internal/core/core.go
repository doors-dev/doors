package core

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/doors-dev/doors/internal/beam2"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

type JobCore interface {
	gox.Job
	Apply(core Core)
}

type Hook struct {
	DoorID uint64
	HookID uint64
	Cancel context.CancelFunc
}

type Instance interface {
	CallCtx(ctx context.Context, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams) context.CancelFunc
	CallCheck(check func() bool, action action.Action, onResult func(json.RawMessage, error), onCancel func(), params action.CallParams)
	CSPCollector() *common.CSPCollector
	AddModuleImport(specifier string, path string)
	ImportRegistry() *resources.Registry
	ID() string
	RootID() uint64
	Conf() *common.SystemConf
	Detached() bool
}

type Door interface {
	Cinema() beam2.Cinema
	RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (Hook, bool)
}

func NewCore(inst Instance, door Door) Core {
	return &core{
		door: door,
		inst: inst,
	}
}

type Core = *core

var _ beam2.Core = &core{}

type core struct {
	door Door
	inst Instance
}

func (c Core) Cinema() beam2.Cinema {
	return c.door.Cinema()
}

type Done = bool

func (c Core) Detached() bool {
	return c.inst.Detached()
}

func (c Core) InstanceID() string {
	return c.inst.ID()
}

func (c Core) RootID() uint64 {
	return c.inst.RootID()
}

func (c Core) Conf() *common.SystemConf {
	return c.inst.Conf()
}

func (c Core) ImportRegistry() *resources.Registry {
	return c.inst.ImportRegistry()
}

func (c Core) AddModuleImport(specifier string, path string) {
	c.inst.AddModuleImport(specifier, path)
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
