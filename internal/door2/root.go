package door2

import (
	"context"
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/sh"
)

type Instance interface {
	Call(call action.Call)
	OnPanic(error)
}

type Root struct {
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	prime   *common.Prime
	spawner sh.Spawner
	inst    Instance
	tackers map[uint64]*tracker
}

func (r *Root) onPanic(err error) {
	r.inst.OnPanic(err)
}

func (r *Root) addTracker(t *tracker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tackers[t.id] = t
}

func (r *Root) removeTracker(t *tracker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.tackers[t.id]
	if !ok || existing != t {
		return
	}
	delete(r.tackers, t.id)
}

func (r *Root) getSpawner() sh.Spawner {
	return r.spawner
}

func (r *Root) getContext() context.Context {
	return r.ctx
}

func (r *Root) getRoot() *Root {
	return r
}

func (r *Root) newId() uint64 {
	return r.prime.Gen()
}

func (i *Root) TriggerHook(doorId uint64, hookId uint64, w http.ResponseWriter, r *http.Request, track uint64) bool {
	i.mu.Lock()
	tracker, ok := i.tackers[doorId]
	i.mu.Unlock()
	if !ok {
		return false
	}
	ok = tracker.trigger(hookId, w, r)
	if !ok {
		return false
	}
	if track != 0 {
		i.inst.Call(reportHook(track))
	}
	return true
}

