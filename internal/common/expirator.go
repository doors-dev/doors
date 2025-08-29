package common

import (
	"container/heap"
	"sync/atomic"
	"time"
)

type expiration struct {
	id    uint64
	time  time.Time
	index int
}

type expirations []*expiration

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

func NewExpirator(on func()) *Expirator {
	es := expirations(make([]*expiration, 0))
	return &Expirator{
		lookup: make(map[uint64]*expiration),
		heap:   &es,
		on:     on,
	}
}

type Expirator struct {
	expired atomic.Bool
	lookup  map[uint64]*expiration
	heap    *expirations
	on      func()
	head    *expiration
	timer   *time.Timer
}

func (x *Expirator) Shutdown() {
	if x.expired.Swap(true) {
		return
	}
	if x.timer != nil {
		x.timer.Stop()
	}
}

func (x *Expirator) expire() {
	if x.expired.Swap(true) {
		return
	}
	x.on()
}

func (x *Expirator) newHead(e *expiration) {
	if x.timer != nil {
		if !x.timer.Stop() {
			return
		}
	}
	x.head = e
	if e == nil {
		x.timer = nil
		return
	}
	d := time.Until(e.time)
	if x.timer == nil {
		x.timer = time.AfterFunc(d, x.expire)
	} else {
		x.timer.Reset(d)
	}
}

func (x *Expirator) Report(id uint64) {
	if x.expired.Load() {
		return
	}
	if x.head.id == id {
		var e *expiration
		if x.heap.Len() > 0 {
			e = heap.Pop(x.heap).(*expiration)
			delete(x.lookup, e.id)
		}
		x.newHead(e)
		return
	}
	e := x.lookup[id]
	delete(x.lookup, id)
	heap.Remove(x.heap, e.index)
}

func (x *Expirator) Track(id uint64, expire time.Time) {
	if x.expired.Load() {
		return
	}
	e := &expiration{
		id:   id,
		time: expire,
	}
	heap.Push(x.heap, e)
	if e.index == 0 && (x.head == nil || e.time.Before(x.head.time)) {
		heap.Pop(x.heap)
		x.newHead(e)
		return
	}
	x.lookup[e.id] = e
}
