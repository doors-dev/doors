package door

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/printer"
	"github.com/doors-dev/doors/internal/shredder"
)

type tracker struct {
	mu          sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	outer       *outerTracker
	rw          shredder.ReadWriteThread
	renderCtx   context.Context
	printers    common.Set[*printer.PayloadPrinter]
	cinema      beam.Cinema
	cleanGuard  shredder.ValveFrame
	cleaned     atomic.Bool
	cleanThread shredder.ReadWriteThread
	children    common.Set[*outerTracker]
	hooks       common.Set[uint64]
	node        *node
}

func (t *tracker) Context() context.Context {
	return t.renderCtx
}

func (t *tracker) ReadFrame() shredder.ReleaseFrame {
	return t.rw.Read()
}

func (t *tracker) Runtime() shredder.Runtime {
	return t.outer.Runtime()
}

var _ beam.Door = (*tracker)(nil)

func (t *tracker) Cinema() beam.Cinema {
	return t.cinema
}

func (t *tracker) CleanFrame() shredder.Frame {
	return &t.cleanGuard
}

func (t *tracker) ID() uint64 {
	return t.outer.ID()
}

func (t *tracker) Instance() core.Instance {
	return t.outer.Instance()
}

func (t *tracker) ReadyFrame() shredder.ReleaseFrame {
	return shredder.JoinRelease(t.ctx, t.outer.callGuard, t.rw.Read(), t.cleanThread.Read())
}

func (t *tracker) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, race core.Race) (core.Hook, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx.Err() != nil {
		return core.Hook{}, false
	}
	if t.hooks == nil {
		t.hooks = common.NewSet[uint64]()
	}
	h := newHook(t.inst().NewID(), t, onTrigger, race)
	t.hooks.Add(h.id)
	t.outer.root.addHook(h)
	return core.Hook{
		HookID: h.id,
		Cancel: h.cancel,
	}, true
}

func (t *tracker) removeHook(id uint64) {
	t.outer.root.removeHook(id)
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.hooks == nil {
		return
	}
	t.hooks.Remove(id)
}

func (t *tracker) inst() Instance {
	return t.outer.root.inst
}

func (t *tracker) Reload(ctx context.Context) <-chan error {
	if t.node == nil {
		ch := make(chan error, 1)
		ch <- errors.New("root door cannot be reloaded")
		close(ch)
		return ch
	}
	return t.node.door.reloadSelf(ctx, t.node)
}

func (t *tracker) RootCore() core.Core {
	return t.outer.root.core
}

func (t *tracker) UserCall(ctx context.Context, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	callFrame := t.callFrame(ctx)
	defer callFrame.Release()
	callFrame.Run(ctx, t.Runtime(), func(b bool) {
		if !b {
			if onCancel != nil {
				onCancel()
			}
			return
		}
		t.inst().UserCall(ctx, action, onResult, onCancel, params)
	})
}

func (t *tracker) cinemaFrame() shredder.ReleaseFrame {
	return shredder.JoinRelease(t.ctx, t.cinema.ReadFrame(), t.outer.cinema.ReadFrame())
}

func (t *tracker) isEmpty() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.children.Len() == 0
}

func (t *tracker) callFrame(ctx context.Context) shredder.ReleaseFrame {
	frames := ctex.GetFrames(ctx)
	return shredder.JoinRelease(ctx, frames.Call(), t.outer.callGuard, t.rw.Read())
}

var _ core.Door = (*tracker)(nil)

func (t *tracker) newPrinter() (*printer.PayloadPrinter, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx.Err() != nil {
		return nil, false
	}
	if t.printers == nil {
		t.printers = common.NewSet[*printer.PayloadPrinter]()
	}
	p := printer.NewPayloadPrinter(t.Instance().Session().App().Conf().ServerDisableGzip, func(p *printer.PayloadPrinter) {
		t.mu.Lock()
		defer t.mu.Unlock()
		if t.printers == nil {
			return
		}
		t.printers.Remove(p)
	})
	t.printers.Add(p)
	return p, true
}

func (t *tracker) addChild(c *outerTracker) {
	t.mu.Lock()
	if t.cleaned.Load() {
		t.mu.Unlock()
		c.clean()
		return
	}
	if t.children == nil {
		t.children = common.NewSet[*outerTracker]()
	}
	t.children.Add(c)
	t.mu.Unlock()
}

func (t *tracker) removeChild(c *outerTracker) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.children == nil {
		return
	}
	t.children.Remove(c)
}

func (t *tracker) clean() {
	if t.cleaned.Swap(true) {
		return
	}
	t.cancel()
	t.cinema.Clean()
	t.mu.Lock()
	printers := t.printers
	children := t.children
	hooks := t.hooks
	t.printers = nil
	t.children = nil
	t.hooks = nil
	t.mu.Unlock()
	for p := range printers {
		p.Release()
	}
	for c := range children {
		c.clean()
	}
	for id := range hooks {
		t.outer.root.cancelHook(id)
	}
	t.outer.removeChild(t)
	write := t.cleanThread.Write()
	write.Run(nil, nil, func(bool) {
		t.cleanGuard.Activate()
	})
	write.Release()
}

func newRootOuterTracker(r *root) *outerTracker {
	ctx, cancel := context.WithCancel(r.runtime().Context())
	tracker := &outerTracker{
		id:        r.inst.NewID(),
		root:      r,
		ctx:       ctx,
		cancel:    cancel,
		callGuard: new(shredder.ValveFrame),
	}
	tracker.callGuard.Activate()
	tracker.placeGuard = &tracker.outerGuard
	tracker.cinema = beam.NewCinema(nil, tracker)
	tracker.renderCtx = common.NewRenderCtx(context.WithValue(tracker.ctx, common.KeyCore, core.NewCore(tracker)), tracker.user)
	return tracker
}

func newOuterTracker1(prev *outerTracker) *outerTracker {
	ctx, cancel := context.WithCancel(prev.parent.ctx)
	tracker := &outerTracker{
		id:         prev.id,
		root:       prev.root,
		parent:     prev.parent,
		ctx:        ctx,
		cancel:     cancel,
		callGuard:  prev.callGuard,
		placeGuard: prev.placeGuard,
		user:       prev.user,
	}
	tracker.cinema = beam.NewCinema(prev.parent.Cinema(), tracker)
	tracker.renderCtx = common.NewRenderCtx(context.WithValue(tracker.ctx, common.KeyCore, core.NewCore(tracker)), tracker.user)
	prev.parent.addChild(tracker)
	return tracker
}

func newOuterTracker2(parent *tracker, callGuard *shredder.ValveFrame, userCtx context.Context) *outerTracker {
	ctx, cancel := context.WithCancel(parent.ctx)
	tracker := &outerTracker{
		id:        parent.Instance().NewID(),
		root:      parent.outer.root,
		parent:    parent,
		ctx:       ctx,
		cancel:    cancel,
		callGuard: callGuard,
		user:      common.UserCtx(userCtx, parent.renderCtx),
	}
	tracker.cinema = beam.NewCinema(parent.Cinema(), tracker)
	tracker.placeGuard = &tracker.outerGuard
	tracker.renderCtx = common.NewRenderCtx(context.WithValue(tracker.ctx, common.KeyCore, core.NewCore(tracker)), tracker.user)
	parent.addChild(tracker)
	return tracker
}

type outerTracker struct {
	id             uint64
	root           *root
	parent         *tracker
	ctx            context.Context
	cancel         context.CancelFunc
	callGuard      *shredder.ValveFrame
	placeGuard     *shredder.ValveFrame
	outerGuard     shredder.ValveFrame
	cleanTrigger   shredder.ValveFrame
	cleanThread    shredder.ReadWriteThread
	printers       common.Set[*printer.PayloadPrinter]
	renderCtx      context.Context
	placementError error
	outerError     error
	cinema         beam.Cinema
	mu             sync.Mutex
	cleaned        atomic.Bool
	children       common.Set[*tracker]
	hooks          common.Set[uint64]
	current        *tracker
	user           context.Context
}

func (t *outerTracker) Context() context.Context {
	return t.renderCtx
}

func (t *outerTracker) ReadFrame() shredder.ReleaseFrame {
	return shredder.Join(t.ctx, &t.outerGuard)
}

func (t *outerTracker) Runtime() shredder.Runtime {
	return t.root.inst.Runtime()
}

var _ beam.Door = (*outerTracker)(nil)

func (t *outerTracker) Cinema() beam.Cinema {
	return t.cinema
}

func (t *outerTracker) CleanFrame() shredder.Frame {
	return &t.cleanTrigger
}

func (t *outerTracker) ID() uint64 {
	return t.id
}

func (t *outerTracker) Instance() core.Instance {
	return t.root.inst
}

func (t *outerTracker) ReadyFrame() shredder.ReleaseFrame {
	return shredder.JoinRelease(t.ctx, t.callGuard, &t.outerGuard, t.cleanThread.Read())
}

func (t *outerTracker) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, race core.Race) (core.Hook, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx.Err() != nil {
		return core.Hook{}, false
	}
	if t.hooks == nil {
		t.hooks = common.NewSet[uint64]()
	}
	h := newHook(t.inst().NewID(), t, onTrigger, race)
	t.hooks.Add(h.id)
	t.root.addHook(h)
	return core.Hook{
		HookID: h.id,
		Cancel: h.cancel,
	}, true
}

func (t *outerTracker) removeHook(id uint64) {
	t.root.removeHook(id)
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.hooks == nil {
		return
	}
	t.hooks.Remove(id)
}

func (t *outerTracker) inst() Instance {
	return t.root.inst
}

func (t *outerTracker) Reload(ctx context.Context) <-chan error {
	t.mu.Lock()
	current := t.current
	t.mu.Unlock()
	return current.Reload(ctx)
}

func (t *outerTracker) RootCore() core.Core {
	return t.root.core
}

func (t *outerTracker) UserCall(ctx context.Context, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	callFrame := t.callFrame(ctx)
	defer callFrame.Release()
	callFrame.Run(ctx, t.Runtime(), func(b bool) {
		if !b {
			if onCancel != nil {
				onCancel()
			}
			return
		}
		t.inst().UserCall(ctx, action, onResult, onCancel, params)
	})
}

func (t *outerTracker) callFrame(ctx context.Context) shredder.ReleaseFrame {
	frames := ctex.GetFrames(ctx)
	return shredder.JoinRelease(ctx, frames.Call(), t.callGuard, &t.outerGuard)
}

var _ core.Door = (*outerTracker)(nil)

func (t *outerTracker) newPrinter() (*printer.PayloadPrinter, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ctx.Err() != nil {
		return nil, false
	}
	if t.printers == nil {
		t.printers = common.NewSet[*printer.PayloadPrinter]()
	}
	p := printer.NewPayloadPrinter(t.Instance().Session().App().Conf().ServerDisableGzip, func(p *printer.PayloadPrinter) {
		t.mu.Lock()
		defer t.mu.Unlock()
		if t.printers == nil {
			return
		}
		t.printers.Remove(p)
	})
	t.printers.Add(p)
	return p, true
}

func (t *outerTracker) addChild(c *tracker) {
	t.mu.Lock()
	if t.cleaned.Load() {
		t.mu.Unlock()
		c.clean()
		return
	}
	defer t.mu.Unlock()
	if t.children == nil {
		t.children = common.NewSet[*tracker]()
	}
	t.children.Add(c)
}

func (t *outerTracker) removeChild(c *tracker) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.children == nil {
		return
	}
	t.children.Remove(c)
}

func (t *outerTracker) clean() {
	if t.cleaned.Swap(true) {
		return
	}
	t.cancel()
	t.cinema.Clean()
	t.mu.Lock()
	printers := t.printers
	children := t.children
	hooks := t.hooks
	t.printers = nil
	t.children = nil
	t.hooks = nil
	t.mu.Unlock()
	for p := range printers {
		p.Release()
	}
	for c := range children {
		c.clean()
	}
	for id := range hooks {
		t.root.cancelHook(id)
	}
	if t.parent != nil {
		t.parent.removeChild(t)
	}
	write := t.cleanThread.Write()
	defer write.Release()
	write.Run(nil, nil, func(bool) {
		t.cleanTrigger.Activate()
	})
}

func (t *outerTracker) newTracker(n *node) *tracker {
	ctx, cancel := context.WithCancel(t.ctx)
	tracker := &tracker{
		ctx:    ctx,
		cancel: cancel,
		outer:  t,
		node:   n,
	}
	var parentCinema beam.Cinema
	if t.parent != nil {
		parentCinema = t.parent.Cinema()
	}
	tracker.cinema = beam.NewCinema(parentCinema, tracker)
	tracker.renderCtx = common.NewRenderCtx(context.WithValue(tracker.ctx, common.KeyCore, core.NewCore(tracker)), t.user)
	t.mu.Lock()
	t.current = tracker
	t.mu.Unlock()
	t.addChild(tracker)
	return tracker
}
