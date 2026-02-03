package door

import (
	"context"
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type Instance interface {
	Conf() *common.SystemConf
	Call(call action.Call)
	core.Instance
}

type Root = *root

func NewRoot(inst Instance) Root {
	r := &root{
		inst:    inst,
		prime:   common.NewPrime(),
		tackers: make(map[uint64]*tracker),
	}
	r.tracker = newRootTracker(r)
	return r
}

type root struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	prime   *common.Prime
	inst    Instance
	tackers map[uint64]*tracker
	tracker *tracker
}

func (r *root) runtime() shredder.Runtime {
	return r.inst.Runtime()
}

func (r *root) resourceRegistry() *resources.Registry {
	return r.inst.ResourceRegistry()
}

func (r Root) ID() uint64 {
	return r.tracker.id
}

func (r Root) Context() context.Context {
	return r.tracker.ctx
}

func (r Root) Kill() {
	r.tracker.kill()
}

func (r Root) Render(el gox.Elem) (Pipe, shredder.SimpleFrame) {
	thread := shredder.Thread{}
	renderFrame := thread.Frame()
	readyFrame := thread.Frame()
	pipe := newPipe(shredder.FreeFrame{})
	pipe.tracker = r.tracker
	pipe.renderFrame = shredder.Join(true, renderFrame, r.tracker.newRenderFrame())
	pipe.renderFrame.Run(r.tracker.ctx, r.runtime(), func(ok bool) {
		defer pipe.close()
		if !ok {
			return
		}
		err := el.Print(pipe.tracker.ctx, pipe)
		if err != nil {
			pipe.Send(gox.NewJobError(pipe.tracker.ctx, err))
		}
	})
	return pipe, readyFrame
}

func (i Root) TriggerHook(doorID uint64, hookId uint64, w http.ResponseWriter, r *http.Request, track uint64) bool {
	if i.tracker.id == doorID {
		return i.tracker.trigger(hookId, w, r)
	}
	i.mu.Lock()
	tracker, ok := i.tackers[doorID]
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

func (r *root) addTracker(t *tracker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tackers[t.id] = t
}

func (r *root) removeTracker(t *tracker) {
	if r.tracker.isKilled() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.tackers[t.id]
	if !ok || existing != t {
		return
	}
	delete(r.tackers, t.id)
}

func (r *root) NewID() uint64 {
	return r.prime.Gen()
}
