package instance

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/path"
)

type AnyInstance interface {
	Id() string
	Serve(http.ResponseWriter, int) error
	UpdatePath(m any, adapter path.AnyAdapter) bool
	TriggerHook(uint64, uint64, http.ResponseWriter, *http.Request) bool
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

func NewInstance[M comparable](sess *Session, page Page[M], adapter *path.Adapter[M], m *M, opt *Options) (AnyInstance, bool) {
	inst := &Instance[M]{
		id:      common.RandId(),
		beam:    beam.NewSourceBeam(*m),
		adapter: adapter,
		page:    page,
		opt:     opt,
		session: sess,
	}
	inst.resetKillTimer()
	return inst, sess.AddInstance(inst)
}

type Instance[M any] struct {
	mu       sync.RWMutex
	killed   bool
	id       string
	beam     beam.SourceBeam[M]
	adapter  *path.Adapter[M]
	page     Page[M]
	opt      *Options
	core     *core[M]
	session  *Session
	ctx      context.Context
	included atomic.Bool
	killTimer  *time.Timer
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
	inst.killTimer = time.AfterFunc(inst.conf().InstanceTTL, func() {
		inst.end(causeKilled)
	})
	return true
}



func (inst *Instance[M]) conf() *common.SystemConf {
	return inst.session.router.Conf()
}

func (inst *Instance[M]) include() bool {
	return !inst.included.Swap(true)
}

func (inst *Instance[M]) getSession() coreSession {
	return inst.session
}

func (inst *Instance[M]) TriggerHook(nodeId uint64, hookId uint64, w http.ResponseWriter, r *http.Request) bool {
	inst.mu.RLock()
	if inst.killed || inst.core == nil {
		inst.mu.RUnlock()
		return false
	}
	inst.mu.RUnlock()
	return inst.core.TriggerHook(nodeId, hookId, w, r)
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
	inst.end(causeSyncError)
}
func (inst *Instance[M]) Serve(w http.ResponseWriter, code int) error {
	inst.mu.Lock()
	if inst.killed {
		return errors.New("Instance killed before render")
	}
	if inst.core != nil {
		inst.mu.Unlock()
		log.Fatal("Instance rendered twice")
	}
	spawner := inst.session.router.Spawner()
	inst.core = newCore[M](inst, newSolitaire(inst, common.GetSolitaireConf(inst.conf())), spawner)
	inst.mu.Unlock()
	err := inst.core.serve(w, inst.page.Render(inst.beam), code)
	if err != nil {
		defer inst.end(causeKilled)
	}
	return err
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
	inst.core.end(cause)
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

