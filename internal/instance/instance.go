package instance

import (
	"context"
	"errors"
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
	Connect(common.EventSender)
	kill(suspend bool)
	CallResponse(*CallResponse) bool
}

type Page[M any] interface {
	Render(s beam.SourceBeam[M]) templ.Component
}

type Options struct {
	Detached bool
	Rerouted bool
	TTL      time.Duration
}

func NewInstance[M any](sess *Session, page Page[M], adapter *path.Adapter[M], m *M, opt *Options) (AnyInstance, bool) {
	inst := &Instance[M]{
		id:      common.RandId(),
		beam:    beam.NewSourceBeam(*m, true),
		adapter: adapter,
		page:    page,
		opt:     opt,
		session: sess,
	}
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
}

func (inst *Instance[M]) include() bool {
	return !inst.included.Swap(true)
}

func (inst *Instance[M]) getSession() coreSession {
	return inst.session
}

func (inst *Instance[M]) CallResponse(r *CallResponse) bool {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	if inst.killed || inst.core == nil {
		return false
	}
	return inst.core.connector.CallResponse(r)
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

func (inst *Instance[M]) Connect(conn common.EventSender) {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	if inst.killed || inst.core == nil {
		conn.Tx(common.GoneEvent{})
		conn.Close()
		return
	}
	inst.core.connector.Connect(conn)
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
	inst.core = newCore[M](inst, newConnector(inst, inst.opt.TTL), spawner)
	inst.mu.Unlock()
	err := inst.core.serve(w, inst.page.Render(inst.beam), code)
	if err != nil {
		defer inst.kill(true)
	}
	return err
}

func (inst *Instance[M]) Id() string {
	return inst.id
}

func (inst *Instance[M]) kill(suspend bool) {
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
	inst.core.kill(suspend)
}

func (inst *Instance[M]) end() {
	inst.kill(true)
}
