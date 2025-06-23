package node

import (
	"context"
	"net/http"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/common/ctxwg"
)

type Done = bool

type HookEntry struct {
	NodeId uint64
	HookId uint64
	inst   instance
}

func (h HookEntry) Cancel() {
	h.inst.CancelHook(h.NodeId, h.HookId, nil)
}

func (h HookEntry) cancel(err error) {
	h.inst.CancelHook(h.NodeId, h.HookId, err)
}

type CallHook struct {
	Trigger func(ctx context.Context, w http.ResponseWriter, r *http.Request)
	Cancel  func(ctx context.Context, err error)
}

func (s *CallHook) trigger(ctx context.Context, w http.ResponseWriter, r *http.Request) Done {
	if s.Trigger != nil {
		s.Trigger(ctx, w, r)
	}
	return true
}

func (c *CallHook) cancel(ctx context.Context, err error) {
	if c.Cancel != nil {
		c.cancel(ctx, err)
	}
}

type AttrHook struct {
	Trigger func(ctx context.Context, w http.ResponseWriter, r *http.Request) Done
	Cancel  func(ctx context.Context, err error)
}

func (s *AttrHook) trigger(ctx context.Context, w http.ResponseWriter, r *http.Request) Done {
	return s.Trigger(ctx, w, r)
}

func (s *AttrHook) cancel(ctx context.Context, err error) {
	if s.Cancel != nil {
		s.Cancel(ctx, err)
	}
}

type Hook interface {
	trigger(ctx context.Context, w http.ResponseWriter, r *http.Request) Done
	cancel(ctx context.Context, err error)
}

type doneErr struct {
	err error
}

type NodeHook struct {
	hook    Hook
	mu      sync.Mutex
	counter uint
	done    bool
	ch      chan struct{}
	err     error
	ctx     context.Context
}

func newHook(ctx context.Context, h Hook) *NodeHook {
	return &NodeHook{
		hook: h,
		ctx:  ctx,
	}
}

func (h *NodeHook) Cancel(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.done {
		return
	}
	h.done = true
	h.err = err
	if h.counter == 0 {
		h.hook.cancel(h.ctx, h.err)
	}
}

func (h *NodeHook) Trigger(w http.ResponseWriter, r *http.Request) (Done, bool) {
	h.mu.Lock()
	ch := make(chan struct{}, 0)
	prevCh := h.ch
	h.ch = ch
	h.mu.Unlock()
	if prevCh != nil {
		<-prevCh
	}
	h.mu.Lock()
	if h.done {
		h.mu.Unlock()
        close(ch)
		return false, false
	}
	h.counter += 1
	h.mu.Unlock()
	ctx := ctxwg.Insert(h.ctx)
	ctx = common.SetBlockingCtx(ctx)
	done := h.hook.trigger(ctx, w, r)
	if done {
		h.mu.Lock()
		h.done = done
		h.mu.Unlock()
	}
	close(ch)
	ctxwg.Wait(ctx)
	h.mu.Lock()
	h.counter -= 1
	last := !done && h.counter == 0 && h.done
	h.mu.Unlock()
	if last {
		h.hook.cancel(h.ctx, h.err)
	}
	return done, true
}
