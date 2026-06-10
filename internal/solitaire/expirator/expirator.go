// Copyright 2026 doors dev LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package expirator

import (
	"container/heap"
	"sync"
	"time"
)

type expiration struct {
	id    uint64
	time  time.Time
	index int
}

type expirations []*expiration

func (h expirations) Head() *expiration {
	if h.Len() == 0 {
		return nil
	}
	return h[0]
}

func (h expirations) Len() int {
	return len(h)
}

func (h expirations) Less(i, j int) bool {
	return h[i].time.Before(h[j].time)
}

func (h expirations) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *expirations) Push(x any) {
	e := x.(*expiration)
	e.index = len(*h)
	*h = append(*h, e)
}

func (h *expirations) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	old[n-1] = nil
	*h = old[:n-1]
	return x
}

type ExpirationHandler interface {
	Expire()
}

func NewExpirator(handler ExpirationHandler) *Expirator {
	es := expirations(make([]*expiration, 0))
	return &Expirator{
		lookup:  make(map[uint64]*expiration),
		heap:    &es,
		handler: handler,
	}
}

type Expirator struct {
	mu      sync.Mutex
	expired bool
	lookup  map[uint64]*expiration
	handler ExpirationHandler
	timer   *time.Timer
	heap    *expirations
}

func (x *Expirator) Shutdown() {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.expired {
		return
	}
	x.expired = true
	if x.timer != nil {
		x.timer.Stop()
	}
}

func (x *Expirator) expire() {
	x.mu.Lock()
	if x.expired {
		x.mu.Unlock()
		return
	}
	x.expired = true
	x.mu.Unlock()
	x.handler.Expire()
}

func (x *Expirator) newHead(e *expiration) {
	if x.timer != nil && !x.timer.Stop() {
		return
	}
	if e == nil {
		x.timer = nil
		return
	}
	d := time.Until(e.time)
	if x.timer == nil {
		x.timer = time.AfterFunc(d, x.expire)
		return
	}
	x.timer.Reset(d)
}

func (x *Expirator) Report(id uint64) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.expired {
		return
	}
	e := x.lookup[id]
	delete(x.lookup, id)
	first := e.index == 0
	heap.Remove(x.heap, e.index)
	if !first {
		return
	}
	e = x.heap.Head()
	x.newHead(e)
}

func (x *Expirator) Track(id uint64, expire time.Time) {
	x.mu.Lock()
	defer x.mu.Unlock()
	if x.expired {
		return
	}
	e := &expiration{
		id:   id,
		time: expire,
	}
	x.lookup[e.id] = e
	heap.Push(x.heap, e)
	if e.index != 0 {
		return
	}
	x.newHead(e)
}
