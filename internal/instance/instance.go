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

package instance

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/door"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/doors/internal/solitaire"
	"github.com/doors-dev/gox"
)

type AnyInstance interface {
	ID() string
	Serve(http.ResponseWriter, *http.Request) error
	RestorePath(path.Location) bool
	TriggerHook(uint64, http.ResponseWriter, *http.Request, uint64) bool
	Connect(w http.ResponseWriter, r *http.Request)
	SetStatus(int)
	InstanceEnd()
	end(common.EndCause)
}

type App[M any] interface {
	Main(path *beam.SourceBeam[M]) gox.Elem
}

type setup[M any] struct {
	adapter  path.Adapter[M]
	beam     *beam.SourceBeam[M]
	comp     gox.Comp
	rerouted bool
}

type Options struct {
	Rerouted bool
}

func NewInstance[M any](sess *Session, adapter path.Adapter[M], beam *beam.SourceBeam[M], comp gox.Comp, opt Options) (AnyInstance, bool) {
	inst := &Instance[M]{
		id: common.RandId(),
		setup: &setup[M]{
			adapter:  adapter,
			beam:     beam,
			comp:     comp,
			rerouted: opt.Rerouted,
		},
		session: sess,
		store:   ctex.NewStore(),
	}
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
	runtime    shredder.Runtime
	solitaire  solitaire.Solitaire
	root       door.Root
	killTimer  *killTimer
	store      ctex.Store
	csp        *common.CSPCollector
	importMap  *importMap
	pageStatus atomic.Int32
	meta       *titleMeta
}

func (inst *Instance[M]) init() error {
	ok := inst.state.CompareAndSwap(initial, active)
	if !ok {
		return errors.New("instance has already started or stopped")
	}
	ctx := context.WithValue(context.Background(), ctex.KeySessionStore, inst.session.store)
	ctx = context.WithValue(ctx, ctex.KeyInstanceStore, inst.store)
	inst.runtime = shredder.NewRuntime(ctx, inst.Conf().InstanceGoroutineLimit, inst)
	inst.root = door.NewRoot(inst)
	inst.solitaire = solitaire.NewSolitaire(inst, common.GetSolitaireConf(inst.Conf()))
	inst.navigator = newNavigator(
		inst,
		inst.setup.adapter,
		inst.session.router.Adapters(),
		inst.setup.beam,
		inst.root.Context(),
		inst.setup.rerouted,
	)
	inst.killTimer = &killTimer{
		initial: inst.Conf().InstanceConnectTimeout,
		regular: inst.Conf().InstanceTTL,
		inst:    inst,
	}
	inst.csp = inst.session.router.CSP().NewCollector()
	inst.importMap = newImportMap()
	inst.meta = newTitleMeta(inst)
	inst.killTimer.keepAlive()
	return nil
}

func (i *Instance[M]) PathMaker() path.PathMaker {
	return i.session.router.PathMaker()
}

func (i *Instance[M]) Adapters() path.Adapters {
	return i.session.router.Adapters()
}

func (i *Instance[M]) SessionExpire(d time.Duration) {
	i.session.SetExpiration(d)
}

func (i *Instance[M]) SessionEnd() {
	i.session.Kill()
}

func (i *Instance[M]) InstanceEnd() {
	i.end(common.EndCauseKilled)
}

func (i *Instance[M]) SessionID() string {
	return i.session.ID()
}

func (d *Instance[M]) SetStatus(status int) {
	d.pageStatus.Store(int32(status))
}

func (inst *Instance[M]) getStatus() int {
	if s := inst.pageStatus.Load(); s > 0 {
		return int(s)
	}
	return http.StatusOK
}

func (c *Instance[M]) NewLink(m any) (core.Link, error) {
	return c.navigator.newLink(m)
}

func (c *Instance[M]) NewID() uint64 {
	return c.root.NewID()
}

func (c *Instance[M]) RootID() uint64 {
	return c.root.ID()
}

func (c *Instance[M]) ResourceRegistry() *resources.Registry {
	return c.session.router.ResourceRegistry()
}

func (c *Instance[M]) ModuleRegistry() core.ModuleRegistry {
	return c.importMap
}

func (c *Instance[M]) CSPCollector() *common.CSPCollector {
	return c.csp
}

func (c *Instance[M]) Call(call action.Call) {
	c.solitaire.Call(call)
}

func (inst *Instance[M]) Conf() *common.SystemConf {
	return inst.session.router.Conf()
}

func (inst *Instance[M]) Touch() {
	inst.session.limiter.touch(inst.id)
}

func (inst *Instance[M]) Runtime() shredder.Runtime {
	return inst.runtime
}

func (inst *Instance[M]) TriggerHook(hookID uint64, w http.ResponseWriter, r *http.Request, track uint64) bool {
	if inst.state.Load() != active {
		return false
	}
	ok := inst.root.TriggerHook(hookID, w, r, track)
	if ok {
		inst.Touch()
	}
	return ok

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
	slog.Debug("Instance synchronization error", "error", err, "type", "error", "instance_id", inst.id)
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
	inst.runtime.Cancel()
	inst.solitaire.End(cause)
	inst.root.Kill()
}

func (inst *Instance[M]) RestorePath(l path.Location) bool {
	return inst.navigator.restore(l)
}
