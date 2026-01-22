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

type Panicer interface {
	OnPanic(error)
}

func NewSpawner(ctx context.Context, limit int, panicer Panicer) Spawner {
	s := &spawner{
		limit:   limit,
		ctx:     ctx,
		panicer: panicer,
	}
	Go(s.run)
	return s
}

type spawner struct {
	mu      sync.Mutex
	ctx     context.Context
	killed  bool
	queue   deque.Deque[func()]
	panicer Panicer
	ch      chan struct{}
	limit   int
}

func (s *spawner) Spawn(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ctx.Err() != nil {
		err := common.Catch(f)
		if err != nil {
			s.panicer.OnPanic(err)
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
			<-backpressure
			if err != nil {
				s.panicer.OnPanic(err)
			}
		})
	}
}

