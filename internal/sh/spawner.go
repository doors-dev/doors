package sh

import (
	"context"
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
	go s.run()
	return s
}

type spawner struct {
	mu      sync.Mutex
	ctx     context.Context
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
	backpressure := make(chan chan<- task, s.limit)
	issued := 0
	for {
		s.mu.Lock()
		if s.queue.Len() == 0 {
			if s.ctx.Err() != nil {
				s.mu.Unlock()
				break
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
		var ch chan<- task
		if issued == s.limit {
			ch = <-backpressure
		} else {
			select {
			case ch = <-backpressure:
			default:
				issued += 1
				ch = spawn()
			}
		}
		ch <- task{
			run:     next,
			panicer: s.panicer,
			ch:      backpressure,
		}
	}
	for issued > 0 {
		ch := <-backpressure
		close(ch)
		issued -= 1
	}
}

type task struct {
	run     func()
	panicer Panicer
	ch      chan chan<- task
}

func spawn() chan<- task {
	ch := make(chan task, 0)
	go func() {
		for t := range ch {
			err := common.Catch(t.run)
			if err != nil {
				t.panicer.OnPanic(err)
			}
			t.ch <- ch
		}
	}()
	return ch
}
