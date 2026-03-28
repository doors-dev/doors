// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package core

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type Hook struct {
	DoorID uint64
	HookID uint64
	Cancel context.CancelFunc
}

type Link struct {
	Location path.Location
	On       func(context.Context)
}

/*
func (h *Link) ClickHandler() (func(context.Context), bool) {
	if h.On == nil {
		return nil, false
	}
	return h.On, true
} */

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
	NewID() uint64
	NewLink(any) (Link, error)
	Runtime() shredder.Runtime
	License() string
	SetStatus(int)
	SessionExpire(time.Duration)
	SessionEnd()
	InstanceEnd()
	SessionID() string
	Adapters() path.Adapters
	PathMaker() path.PathMaker
	UpdateTitle(content string, attrs gox.Attrs)
	UpdateMeta(name string, property bool, attrs gox.Attrs)
}

type Door interface {
	Cinema() beam.Cinema
	RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, onCancel func(ctx context.Context)) (Hook, bool)
	ID() uint64
	RootCore() Core
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

func (c Core) RootCore() Core {
	return c.door.RootCore()
}

func (c Core) UpdateMeta(name string, property bool, attrs gox.Attrs) {
	c.inst.UpdateMeta(name, property, attrs)
}

func (c Core) UpdateTitle(content string, attrs gox.Attrs) {
	c.inst.UpdateTitle(content, attrs)
}

func (c Core) PathMaker() path.PathMaker {
	return c.inst.PathMaker()
}

func (c Core) Door() Door {
	return c.door
}

func (c Core) Instance() Instance {
	return c.inst
}

func (c Core) Adapters() path.Adapters {
	return c.inst.Adapters()
}

func (c Core) SessionExpire(d time.Duration) {
	c.inst.SessionExpire(d)
}

func (c Core) SessionEnd() {
	c.inst.SessionEnd()
}

func (c Core) InstanceEnd() {
	c.inst.InstanceEnd()
}

func (c Core) SessionID() string {
	return c.inst.SessionID()
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

func (c Core) License() string {
	return c.inst.License()
}

func (c Core) DoorID() uint64 {
	return c.door.ID()
}

func (c Core) NewLink(m any) (Link, error) {
	return c.inst.NewLink(m)
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
