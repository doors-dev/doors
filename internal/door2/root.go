package door2

import (
	"context"
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/sh"
	"github.com/doors-dev/gox"
)

type Instance interface {
	Conf() *common.SystemConf
	Call(call action.Call)
	sh.Panicer
	core.Instance
}

type Root = *root



func NewRoot(ctx context.Context, inst Instance) Root {
	r := &root{
		inst:    inst,
		prime:   common.NewPrime(),
		tackers: make(map[uint64]*tracker),
	}
	r.tracker = newRootTracker(ctx, r)
	r.spawner = sh.NewSpawner(r.tracker.ctx, r.inst.Conf().InstanceGoroutineLimit, inst)
	return r
}

type root struct {
	mu      sync.Mutex
	cancel  context.CancelFunc
	prime   *common.Prime
	spawner sh.Spawner
	inst    Instance
	tackers map[uint64]*tracker
	tracker *tracker
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

func (r Root) Render(instanceFrame sh.AnyFrame, el gox.Elem) JobStream {
	shread := sh.Shread{}
	renderFrame := shread.Frame()
	defer renderFrame.Release()
	r.tracker.initShread(&shread)
	pipe := newPipe()
	pipe.tracker = r.tracker
	pipe.frame = sh.Join(instanceFrame, renderFrame)
	defer pipe.frame.Release()
	js := newJobStream(pipe)
	pipe.frame.Run(r.spawner, func() {
		defer pipe.Close()
		err := el.Print(pipe.tracker.ctx, pipe)
		if err != nil {
			pipe.Send(gox.NewJobError(pipe.tracker.ctx, err))
		}
	})
	return js
}

func (i Root) TriggerHook(doorId uint64, hookId uint64, w http.ResponseWriter, r *http.Request, track uint64) bool {
	if i.tracker.id == doorId {
		return i.tracker.trigger(hookId, w, r)
	}
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

func (r *root) onPanic(err error) {
	r.inst.OnPanic(err)
}

func (r *root) addTracker(t *tracker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tackers[t.id] = t
}

func (r *root) removeTracker(t *tracker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.tackers[t.id]
	if !ok || existing != t {
		return
	}
	delete(r.tackers, t.id)
}

func (r *root) newId() uint64 {
	return r.prime.Gen()
}
