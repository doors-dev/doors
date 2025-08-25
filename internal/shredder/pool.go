// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package shredder

import (
	"runtime"
	"sync"

	"github.com/doors-dev/doors/internal/common"
	"github.com/panjf2000/ants/v2"
)

type Spawner struct {
	mu     sync.Mutex
	pool   *Pool
	queue  []func()
	ch     chan struct{}
	killed bool
	op     OnPanic
}

func (s *Spawner) NewThead() *Thread {
	return &Thread{
		mu:      sync.Mutex{},
		main:    nil,
		heads:   make([]threadHead, 0),
		spawner: s,
		killed:  false,
		tail:    nil,
	}
}

func (s *Spawner) Kill() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.killed {
		return
	}
	s.killed = true
	s.notify()
}

func (s *Spawner) Go(f func()) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.killed {
		return false
	}
	s.queue = append(s.queue, f)
	s.notify()
	return true
}

func (s *Spawner) notify() {
	if s.ch != nil {
		close(s.ch)
		s.ch = nil
	}
}

func (s *Spawner) run() {
	counter := 0
	mu := &sync.Mutex{}
	var doneCh chan struct{}
	for {
		s.mu.Lock()
		if !s.killed && len(s.queue) == 0 {
			ch := make(chan struct{})
			s.ch = ch
			s.mu.Unlock()
			<-ch
			s.mu.Lock()
		}
		if s.killed || len(s.queue) == 0 {
			s.mu.Unlock()
			break
		}
		queue := s.queue
		s.queue = make([]func(), 0)
		s.mu.Unlock()
		for _, f := range queue {
			mu.Lock()
			if doneCh != nil {
				temp := doneCh
				mu.Unlock()
				<-temp
				mu.Lock()
			}
			counter += 1
			if counter == s.pool.limit {
				doneCh = make(chan struct{})
			}
			mu.Unlock()
			s.pool.ants.Submit(func() {
				defer func() {
					mu.Lock()
					defer mu.Unlock()
					counter -= 1
					if doneCh != nil {
						close(doneCh)
						doneCh = nil
					}
				}()
				err := common.Catch(f)
				if err != nil {
					s.op.OnPanic(err)
				}
			})
		}
	}
	s.pool.done()
}

func NewPool(limit int) *Pool {
	size := runtime.NumCPU()
	ants, err := ants.NewMultiPool(size, limit, ants.LeastTasks)
	if err != nil {
		panic(err)
	}
	return &Pool{
		size:  size,
		mu:    sync.Mutex{},
		ants:  ants,
		limit: limit,
	}
}

type Pool struct {
	size         int
	mu           sync.Mutex
	spawnerCount int
	ants         *ants.MultiPool
	limit        int
}

func (p *Pool) Tune(limit int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.limit = limit
	p.adjust()
}

func (p *Pool) adjust() {
	total := p.spawnerCount * p.limit
	min := p.limit * p.size
	if total < min {
		return
	}
	perPool := total / p.size
	p.ants.Tune(perPool)
}

func (p *Pool) done() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spawnerCount -= 1
	p.adjust()
}

type OnPanic interface {
	OnPanic(error)
}

func (p *Pool) Spawner(op OnPanic) *Spawner {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spawnerCount += 1
	p.adjust()
	s := &Spawner{
		mu:    sync.Mutex{},
		pool:  p,
		queue: make([]func(), 0),
		op:    op,
	}
	go s.run()
	return s
}
