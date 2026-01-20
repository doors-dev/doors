package sh

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/gammazero/deque"
)

type Spawner = *spawner

func NewSpawner(ctx context.Context, limit int, onError func(error)) Spawner {
	s := &spawner{
		limit:   limit,
		ctx:     ctx,
		onError: onError,
	}
	Go(s.run)
	return s
}

type spawner struct {
	mu      sync.Mutex
	ctx     context.Context
	killed  bool
	queue   deque.Deque[func()]
	onError func(error)
	ch      chan struct{}
	limit   int
}

func (s *spawner) Spawn(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ctx.Err() != nil {
		err := common.Catch(f)
		if err != nil {
			s.onError(err)
		}
		return
	}
	s.queue.PushBack(f)
	if s.ch != nil {
		close(s.ch)
		s.ch = nil
	}
}

func (s *spawner) run() {
	backpressure := make(chan struct{}, s.limit)
	for {
		s.mu.Lock()
		if s.queue.Len() == 0 {
			if s.ctx.Err() != nil {
				s.mu.Unlock()
				return
			}
			ch := make(chan struct{})
			s.ch = ch
			s.mu.Unlock()
			select {
			case <-ch:
			case <-s.ctx.Done():
			}
			continue
		}
		next := s.queue.PopFront()
		s.mu.Unlock()
		backpressure <- struct{}{}
		Go(func() {
			err := common.Catch(next)
			s.onError(err)
			<-backpressure
		})
	}
}

func catch(f func(bool), v bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v\n%s", r, debug.Stack())
		}
	}()
	f(v)
	return
}
