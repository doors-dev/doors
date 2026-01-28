package sh

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/gammazero/deque"
)

type Shutdown interface {
	Shutdown()
}

type Runtime = *runtime

func NewRuntime(workerLimit int, shutdown Shutdown) Runtime {
	ctx, cancel := context.WithCancel(context.Background())
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
	callback func()
}

func (t task) run(r *runtime) {
	canceled := t.ctx.Err() != nil || r.ctx.Err() != nil
	if canceled {
		t.cancel()
		return
	}
	err := catch(t.fun)
	if err != nil {
		r.Cancel()
	}
	if t.callback == nil {
		return
	}
	t.callback()
}

func (t task) cancel() {
	t.fun(false)
	if t.callback == nil {
		return
	}
	t.callback()
}


func (r *runtime) Run(ctx context.Context, fun func(bool), callback func()) {
	t := task{ctx, fun, callback}
	t.run(r)
}

func (r *runtime) Submit(ctx context.Context, fun func(bool), callback func()) {
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

func catch(f func(bool)) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
		}
	}()
	f(true)
	return
}
