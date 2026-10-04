package n

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/printer"
	"github.com/doors-dev/doors/internal/shredder"
)

type tracker struct {
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
	outer     *outerTracker
	rw        shredder.ReadWriteThread
	renderCtx context.Context
	printers  common.Set[*printer.PayloadPrinter]
	cinema    beam.Cinema
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

// CleanFrame implements [core.Door].
func (t *tracker) CleanFrame() shredder.Frame {
	panic("unimplemented")
}

func (t *tracker) ID() uint64 {
	return t.outer.ID()
}

func (t *tracker) Instance() core.Instance {
	return t.outer.Instance()
}

func (t *tracker) ReadyFrame() shredder.ReleaseFrame {
	return shredder.Join(t.ctx, true, t.outer.callGuard, t.rw.Read())
}

// RegisterHook implements [core.Door].
func (t *tracker) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, race core.Race) (core.Hook, bool) {
	panic("unimplemented")
}

// Reload implements [core.Door].
func (t *tracker) Reload(ctx context.Context) <-chan error {
	panic("unimplemented")
}

// RootCore implements [core.Door].
func (t *tracker) RootCore() core.Core {
	panic("unimplemented")
}

// UserCall implements [core.Door].
func (t *tracker) UserCall(ctx context.Context, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	panic("unimplemented")
}

func (t *tracker) callFrame(ctx context.Context) shredder.ReleaseFrame {
	frames := ctex.GetFrames(ctx)
	return shredder.Join(ctx, true, frames.Call(), t.outer.callGuard, t.rw.Read())
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

func (t *tracker) clean() {
	t.cancel()
	t.cinema.Clean()
	t.mu.Lock()
	printers := t.printers
	t.printers = nil
	t.mu.Unlock()
	for p := range printers {
		p.Release()
	}
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
	}
	tracker.cinema = beam.NewCinema(prev.parent.Cinema(), tracker)
	tracker.renderCtx = context.WithValue(tracker.ctx, common.KeyCore, tracker)
	return tracker
}

func newOuterTracker2(parent *tracker, callGuard *shredder.ValveFrame) *outerTracker {
	ctx, cancel := context.WithCancel(parent.ctx)
	tracker := &outerTracker{
		id:        parent.outer.root.ID(),
		root:      parent.outer.root,
		parent:    parent,
		ctx:       ctx,
		cancel:    cancel,
		callGuard: callGuard,
	}
	tracker.cinema = beam.NewCinema(parent.Cinema(), tracker)
	tracker.placeGuard = &tracker.outerGuard
	tracker.renderCtx = context.WithValue(tracker.ctx, common.KeyCore, tracker)
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
	renderCtx      context.Context
	placementError error
	outerError     error
	cinema         beam.Cinema
	mu             sync.Mutex
	printers       common.Set[*printer.PayloadPrinter]
}

func (t *outerTracker) Context() context.Context {
	return t.renderCtx
}

func (t *outerTracker) ReadFrame() shredder.ReleaseFrame {
	return shredder.Join(t.ctx, false, &t.outerGuard)
}

func (t *outerTracker) Runtime() shredder.Runtime {
	return t.root.inst.Runtime()
}

var _ beam.Door = (*outerTracker)(nil)

func (t *outerTracker) Cinema() beam.Cinema {
	return t.cinema
}

func (t *outerTracker) CleanFrame() shredder.Frame {
	panic("unimplemented")
}

func (t *outerTracker) ID() uint64 {
	return t.id
}

func (t *outerTracker) Instance() core.Instance {
	return t.root.inst
}

func (t *outerTracker) ReadyFrame() shredder.ReleaseFrame {
	return shredder.Join(t.ctx, false, t.callGuard, &t.outerGuard)
}

func (t *outerTracker) RegisterHook(onTrigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool, race core.Race) (core.Hook, bool) {
	panic("unimplemented")
}

func (t *outerTracker) Reload(ctx context.Context) <-chan error {
	panic("unimplemented")
}

func (t *outerTracker) RootCore() core.Core {
	panic("unimplemented")
}

func (t *outerTracker) UserCall(ctx context.Context, action actions.Action, onResult func(json.RawMessage, error), onCancel func(), params actions.CallParams) {
	panic("unimplemented")
}

func (t *outerTracker) callFrame(ctx context.Context) shredder.ReleaseFrame {
	frames := ctex.GetFrames(ctx)
	return shredder.Join(ctx, true, frames.Call(), t.callGuard, &t.outerGuard)
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

func (t *outerTracker) clean() {
	t.cancel()
	t.cinema.Clean()
	t.mu.Lock()
	printers := t.printers
	t.printers = nil
	t.mu.Unlock()
	for p := range printers {
		p.Release()
	}
}

func (t *outerTracker) newTracker() *tracker {
	ctx, cancel := context.WithCancel(t.ctx)
	tracker := &tracker{
		ctx:    ctx,
		cancel: cancel,
		outer:  t,
	}
	tracker.cinema = beam.NewCinema(t.parent.Cinema(), tracker)
	tracker.renderCtx = context.WithValue(tracker.ctx, common.KeyCore, tracker)
	return tracker
}
