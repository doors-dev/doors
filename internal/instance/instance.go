// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package instance

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/path"
)

type AnyInstance interface {
	Id() string
	Serve(http.ResponseWriter, *http.Request) error
	RestorePath(*http.Request) bool
	TriggerHook(uint64, uint64, http.ResponseWriter, *http.Request, uint64) bool
	Connect(w http.ResponseWriter, r *http.Request)
	end(endCause)
}

type Page[M any] interface {
	Render(s beam.SourceBeam[M]) templ.Component
}

type Options struct {
	Detached bool
	Rerouted bool
}

func NewInstance[M any](sess *Session, page Page[M], adapter *path.Adapter[M], m *M, opt *Options) (AnyInstance, bool) {
	inst := &Instance[M]{
		id:        common.RandId(),
		navigator: newNavigator(adapter, sess.router.Adapters(), m, opt.Detached, opt.Rerouted),
		page:      page,
		opt:       opt,
		session:   sess,
	}
	inst.resetKillTimer()
	return inst, sess.AddInstance(inst)
}

type Instance[M any] struct {
	mu        sync.RWMutex
	killed    bool
	id        string
	navigator *navigator[M]
	page      Page[M]
	opt       *Options
	core      *core[M]
	session   *Session
	ctx       context.Context
	killTimer *time.Timer
}

func (inst *Instance[M]) resetKillTimer() bool {
	if inst.killTimer != nil {
		stopped := inst.killTimer.Stop()
		if !stopped {
			return false
		}
		inst.killTimer.Reset(inst.conf().InstanceTTL)
		return true
	}
	inst.killTimer = time.AfterFunc(inst.conf().InstanceConnectTimeout, func() {
		slog.Debug("Inactive instance killed by timeout", slog.String("type", "message"), slog.String("instance_id", inst.id))
		inst.end(causeKilled)
	})
	return true
}

func (inst *Instance[M]) conf() *common.SystemConf {
	return inst.session.router.Conf()
}

func (inst *Instance[M]) getSession() coreSession {
	return inst.session
}

func (inst *Instance[M]) TriggerHook(doorId uint64, hookId uint64, w http.ResponseWriter, r *http.Request, track uint64) bool {
	inst.mu.RLock()
	if inst.killed || inst.core == nil {
		inst.mu.RUnlock()
		return false
	}
	inst.mu.RUnlock()
	ok := inst.core.TriggerHook(doorId, hookId, w, r, track)
	if ok {
		inst.touch()
	}
	return ok
}

func (inst *Instance[M]) touch() {
	inst.session.limiter.touch(inst.id)
}

func (inst *Instance[M]) Connect(w http.ResponseWriter, r *http.Request) {
	inst.mu.RLock()
	if inst.killed || inst.core == nil {
		inst.mu.RUnlock()
		w.WriteHeader(http.StatusGone)
		return
	}
	inst.resetKillTimer()
	inst.mu.RUnlock()
	inst.core.solitaire.Connect(w, r)
}

func (inst *Instance[M]) syncError(err error) {
	slog.Debug("Instance syncronization error", slog.String("error", err.Error()), slog.String("type", "error"), slog.String("instance_id", inst.id))
	inst.end(causeSyncError)
}
func (inst *Instance[M]) Serve(w http.ResponseWriter, r *http.Request) error {
	inst.mu.Lock()
	if inst.killed {
		inst.mu.Unlock()
		return errors.New("Instance killed before render")
	}
	if inst.core != nil {
		inst.mu.Unlock()
		panic("Instance rendered twice")
	}
	spawner := inst.session.router.Spawner(inst)
	solitaire := newSolitaire(inst, common.GetSolitaireConf(inst.conf()))
	inst.core = newCore(inst, solitaire, spawner, inst.navigator)
	inst.mu.Unlock()
	err := inst.core.serve(w, r, inst.page)
	if err != nil {
		defer inst.end(causeKilled)
	}
	return err
}

func (inst *Instance[M]) OnPanic(err error) {
	slog.Error(err.Error(), slog.String("type", "panic"), slog.String("instance_id", inst.id))
	inst.end(causeKilled)
}

func (inst *Instance[M]) Id() string {
	return inst.id
}

func (inst *Instance[M]) end(cause endCause) {
	inst.mu.Lock()
	if inst.killed {
		inst.mu.Unlock()
		return
	}
	inst.session.removeInstance(inst.id)
	inst.killed = true
	if inst.core == nil {
		inst.mu.Unlock()
		return
	}
	inst.mu.Unlock()
	inst.core.end(cause)
}

func (inst *Instance[M]) RestorePath(r *http.Request) bool {
	return inst.navigator.restore(r)
}

type endCause int

const (
	causeKilled endCause = iota
	causeSuspend
	causeSyncError
)

func (c endCause) Error() string {
	return fmt.Sprint("cause: ", int(c))
}
