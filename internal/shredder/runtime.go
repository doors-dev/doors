package shredder

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gammazero/deque"
)

type Shutdown interface {
	Shutdown()
}

type Runtime = *runtime

func NewRuntime(ctx context.Context, workerLimit int, shutdown Shutdown) Runtime {
	ctx, cancel := context.WithCancel(ctx)
	s := &runtime{
		ctx:         ctx,
		cancel:      cancel,
		workerLimit: workerLimit,
		pool:        make([]chan task, workerLimit),
		cold:        make(chan task),
		hot:         make(chan int, workerLimit),
	}
	go s.loop()
	return s
}

type runtime struct {
	ctx         context.Context
	cancel      context.CancelFunc
	shutdown    Shutdown
	workerLimit int
	pool        []chan task
	cold        chan task
	hot         chan int
}

func (r Runtime) Context() context.Context {
	return r.ctx
}

func (r Runtime) Cancel() {
	r.cancel()
}

type task struct {
	ctx      context.Context
	fun      func(bool)
	callback func(error)
}

func (t task) runUnsafe() {
	canceled := t.ctx.Err() != nil
	if canceled {
		t.cancel()
		return
	}
	defer t.callback(nil)
	t.fun(true)
}


func (t task) run(r *runtime) {
	canceled := t.ctx.Err() != nil || r.ctx.Err() != nil
	if canceled {
		t.cancel()
		return
	}
	err := catch(t.fun, true)
	if err != nil {
		r.onPanic(err)
	}
	if t.callback == nil {
		return
	}
	t.callback(err)
}

func (t task) cancel() {
	t.fun(false)
	if t.callback == nil {
		return
	}
	t.callback(context.Canceled)
}

func (r *runtime) onPanic(err error) {
	slog.Error(err.Error())
	r.Cancel()
}


func (r Runtime) SafeCtxFun(ctx context.Context, fun func(context.Context)) {
	err := catch(fun, ctx)
	if err == nil {
		return
	}
	r.onPanic(err)
}

func (r Runtime) SafeHook(ctx context.Context, w http.ResponseWriter, req *http.Request, handler func(context.Context, http.ResponseWriter, *http.Request) bool) (bool, error) {
	done, err := catchHook(ctx, w, req, handler)
	if err != nil {
		defer r.onPanic(err)
		return false, err
	}
	return done, err
}

func (r Runtime) Go(ctx context.Context, fun func(ctx context.Context)) {
	go func() {
		err := catch(fun, ctx)
		if err == nil {
			return
		}
		r.onPanic(err)
	}()

}

func (r Runtime) Run(ctx context.Context, fun func(bool), callback func(error)) {
	if ctx == nil {
		ctx = context.Background()
	}
	t := task{ctx, fun, callback}
	if r == nil {
		t.runUnsafe()
		return
	}
	t.run(r)
}

func (r Runtime) Submit(ctx context.Context, fun func(bool), callback func(error)) {
	if ctx == nil {
		ctx = context.Background()
	}
	if r == nil {
		panic("submit expects runtime")
	}
	t := task{ctx, fun, callback}
	if r.ctx.Err() != nil {
		t.cancel()
		return
	}
	ok := r.submitHot(t)
	if ok {
		return
	}
	r.submitCold(t)
}

func (r *runtime) submitCold(t task) {
	select {
	case <-t.ctx.Done():
		t.cancel()
	case <-r.ctx.Done():
		t.cancel()
	case r.cold <- t:
	}
}

func (r *runtime) submitHot(t task) bool {
	num, ok := r.getHotNum()
	if !ok {
		return false
	}
	r.pool[num] <- t
	return true
}

func (r *runtime) getHotNum() (int, bool) {
	select {
	case num := <-r.hot:
		return num, true
	default:
		return 0, false
	}
}

func (r *runtime) loop() {
	done := false
	workerCount := 0
	var queue deque.Deque[task]
	for {
		if queue.Len() == 0 {
			if done {
				break
			}
			select {
			case <-r.ctx.Done():
				r.shutdown.Shutdown()
				done = true
			case f := <-r.cold:
				ok := r.submitHot(f)
				if ok {
					continue
				}
				if workerCount == r.workerLimit {
					queue.PushBack(f)
					continue
				}
				ch := make(chan task)
				id := workerCount
				workerCount += 1
				r.pool[id] = ch
				go func() {
					for t := range ch {
						t.run(r)
						r.hot <- id
					}
				}()
				ch <- f
			}
			continue
		}
		if done {
			num := <-r.hot
			r.pool[num] <- queue.PopFront()
			continue
		}
		num, ok := r.getHotNum()
		if ok {
			r.pool[num] <- queue.PopFront()
			continue
		}
		select {
		case <-r.ctx.Done():
			r.shutdown.Shutdown()
			done = true
		case f := <-r.cold:
			queue.PushBack(f)
		case num := <-r.hot:
			r.pool[num] <- queue.PopFront()
		}
	}
	for range len(r.pool) {
		num := <-r.hot
		close(r.pool[num])
	}
}

func catch[T any](f func(T), arg T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("instance runtime panic: %v\n%s", r, debug.Stack())
		}
	}()
	f(arg)
	return
}

func catchHook(ctx context.Context, w http.ResponseWriter, req *http.Request, handler func(context.Context, http.ResponseWriter, *http.Request) bool) (done bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("instance runtime panic: %v\n%s", r, debug.Stack())
		}
	}()
	done = handler(ctx, w, req)
	return
}
