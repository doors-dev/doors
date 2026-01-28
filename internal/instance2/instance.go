// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package instance2

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/door2"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/license"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/doors/internal/solitaire"
	"github.com/doors-dev/gox"
)

type AnyInstance interface {
	ID() string
	Serve(http.ResponseWriter, *http.Request) error
	RestorePath(*http.Request) bool
	TriggerHook(uint64, uint64, http.ResponseWriter, *http.Request, uint64) bool
	Connect(w http.ResponseWriter, r *http.Request)
	end(common.EndCause)
}

type App[M any] interface {
	Main(path beam.SourceBeam[M]) gox.Elem
}

type setup[M any] struct {
	adapter  *path.Adapter[M]
	model    *M
	app      App[M]
	detached bool
	rerouted bool
}

func NewInstance[M any](sess *Session, app App[M], adapter *path.Adapter[M], m *M, detached bool, rerouted bool) (AnyInstance, bool) {
	inst := &Instance[M]{
		id: common.RandId(),
		setup: &setup[M]{
			adapter:  adapter,
			model:    m,
			app:      app,
			detached: detached,
			rerouted: rerouted,
		},
		session: sess,
		store:   ctex.NewStore(),
	}
	inst.SetStatus(http.StatusOK)
	return inst, sess.AddInstance(inst)
}

const (
	initial int32 = iota
	active
	killed
)

type Instance[M any] struct {
	id         string
	state      atomic.Int32
	setup      *setup[M]
	session    *Session
	navigator  *navigator[M]
	runtime    sh.Runtime
	solitaire  solitaire.Solitaire
	root       door2.Root
	killTimer  *killTimer
	store      ctex.Store
	csp        *common.CSPCollector
	importMap  *moduleImportMap
	pageStatus atomic.Int32
}

func (d *Instance[M]) SetStatus(status int) {
	d.pageStatus.Store(int32(status))
}

func (d *Instance[M]) License() license.License {
	return d.session.router.License()
}

func (c *Instance[M]) NewLink(m any) (core.Link, error) {
	return c.navigator.newLink(m)
}

func (c *Instance[M]) NewID() uint64 {
	return c.root.NewID()
}

func (c *Instance[M]) Detached() bool {
	return c.navigator.isDetached()
}

func (c *Instance[M]) RootID() uint64 {
	return c.root.ID()
}

func (c *Instance[M]) ResourceRegistry() *resources.Registry {
	return c.session.router.ResourceRegistry()
}

func (c *Instance[M]) AddModuleImport(specifier string, path string) {
	c.importMap.add(specifier, path)
}

func (c *Instance[M]) CSPCollector() *common.CSPCollector {
	return c.csp
}

func (c *Instance[M]) Call(call action.Call) {
	panic("unimplemented")
}

func (inst *Instance[M]) Conf() *common.SystemConf {
	return inst.session.router.Conf()
}

func (inst *Instance[M]) Touch() {
	inst.session.limiter.touch(inst.id)
}

func (inst *Instance[M]) Runtime() sh.Runtime {
	return inst.runtime
}

func (inst *Instance[M]) init() error {
	ok := inst.state.CompareAndSwap(initial, active)
	if !ok {
		return errors.New("Instance already started or killed")
	}
	inst.runtime = sh.NewRuntime(inst.Conf().InstanceGoroutineLimit, inst)
	ctx := inst.session.store.Inject(inst.runtime.Context(), ctex.KeySessionStore)
	ctx = inst.store.Inject(ctx, ctex.KeyInstanceStore)
	inst.root = door2.NewRoot(ctx, inst)
	inst.solitaire = solitaire.NewSolitaire(inst, common.GetSolitaireConf(inst.Conf()))
	inst.navigator = newNavigator(
		inst,
		inst.setup.adapter,
		inst.session.router.Adapters(),
		inst.setup.model,
		inst.root.Context(),
		inst.setup.detached,
		inst.setup.rerouted,
	)
	inst.killTimer = &killTimer{
		initial: inst.Conf().InstanceConnectTimeout,
		regular: inst.Conf().InstanceTTL,
		inst:    inst,
	}
	inst.csp = inst.session.router.CSP().NewCollector()
	inst.importMap = newImportMap()
	inst.killTimer.keepAlive()
	return nil
}

func (inst *Instance[M]) Serve(w http.ResponseWriter, r *http.Request) error {
	if err := inst.init(); err != nil {
		return err
	}
	el := inst.setup.app.Main(inst.navigator.getBeam())
	inst.setup = nil
	sh := sh.Shread{}
	renderFrame := sh.Frame()
	js := inst.root.Render(renderFrame, el)
	renderFrame.Release()
	ch := make(chan struct{})
	sh.Frame().Run(nil, func() {
		close(ch)
	})
	<-ch
	if err := inst.render(w, r, js); err != nil {
		defer inst.end(common.EndCauseKilled)
		return err
	}
	return nil

}

func (inst *Instance[M]) TriggerHook(doorID uint64, hookID uint64, w http.ResponseWriter, r *http.Request, track uint64) (ok bool) {
	defer func() {
		if ok {
			inst.Touch()
		}
	}()
	if inst.state.Load() != active {
		return false
	}
	return inst.root.TriggerHook(doorID, hookID, w, r, track)
}

func (inst *Instance[M]) Connect(w http.ResponseWriter, r *http.Request) {
	if inst.state.Load() != active {
		w.WriteHeader(http.StatusGone)
		return
	}
	inst.killTimer.keepAlive()
	inst.solitaire.Connect(w, r)
}

func (inst *Instance[M]) SyncError(err error) {
	slog.Debug("Instance syncronization error", slog.String("error", err.Error()), slog.String("type", "error"), slog.String("instance_id", inst.id))
	inst.end(common.EndCauseSyncError)
}

func (inst *Instance[M]) Shutdown() {
	inst.end(common.EndCauseKilled)
}

func (inst *Instance[M]) ID() string {
	return inst.id
}

func (inst *Instance[M]) end(cause common.EndCause) {
	if !inst.state.CompareAndSwap(active, killed) {
		return
	}
	inst.session.removeInstance(inst.id)
	inst.root.Kill()
	inst.solitaire.End(cause)
}

func (inst *Instance[M]) RestorePath(r *http.Request) bool {
	return inst.navigator.restore(r)
}
