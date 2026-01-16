// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package shredder2

import (
	"errors"
	"sync"

	"github.com/doors-dev/doors/internal/common"
)

type threadHead interface {
	threadDone(*Thread)
}

type Thread struct {
	mu      sync.Mutex
	main    func(*Thread)
	counter int
	heads   []threadHead
	killed  bool
	running bool
	spawner *Spawner
	tail    *frame
	after   func()
}

func (t *Thread) Spawner() *Spawner {
	return t.spawner
}

func (t *Thread) addHead(h threadHead) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.counter += 1
	t.heads[len(t.heads)-t.counter] = h
}

func (t *Thread) IsDone() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.spawner == nil
}

func (t *Thread) done() {
	t.killed = true
	for i, head := range t.heads {
		if head == nil {
			continue
		}
		head.threadDone(t)
		t.heads[i] = nil
	}
	if t.after != nil {
		t.after()
		t.after = nil
	}
}

func (t *Thread) Killed() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.killed
}

func (t *Thread) abort() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.main(nil)
	t.done()
}

func (t *Thread) Kill(after func()) bool {
	t.mu.Lock()
	if t.killed {
		t.mu.Unlock()
		return false
	}
	t.killed = true
	if t.tail == nil {
		t.mu.Unlock()
		if after != nil {
			after()
		}
		return true
	}
	threads := t.tail.listThreads()
	t.after = after
	t.mu.Unlock()
	for _, threads := range threads {
		threads.Kill(nil)
	}
	return true
}

func (t *Thread) spawn() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.killed {
		return false
	}
	t.running = true
	return t.spawner.Spawn(func() {
		t.main(t)
		t.main = nil
	}, func(error) {
		t.mu.Lock()
		t.running = false
		defer t.mu.Unlock()
		if t.tail == nil {
			t.done()
		}
	})
}

func (t *Thread) frameDone(frame *frame) {
	t.mu.Lock()
	if !frame.isDone() {
		t.mu.Unlock()
		return
	}
	if frame.next != nil {
		if t.tail == frame {
			t.tail = frame.next
		}
		f := frame.next.start(t.killed)
		defer f()
		t.mu.Unlock()
		return
	}
	t.tail = nil
	defer t.mu.Unlock()
	if len(t.heads) == 0 || t.running {
		return
	}
	t.done()
}

func (th *Thread) readTask(t task) bool {
	th.mu.Lock()
	if th.killed {
		th.mu.Unlock()
		return false
	}
	if th.tail != nil && !th.tail.writing {
		f := th.tail.add(t, false)
		defer f()
		th.mu.Unlock()
		return true
	}
	frame := &frame{
		next:    nil,
		parent:  th,
		threads: common.NewSet[*Thread](),
		tasks:   []task{t},
	}
	if th.tail == nil {
		f := frame.start(false)
		defer f()
	} else {
		th.tail.setNext(frame)
	}
	th.tail = frame
	th.mu.Unlock()
	return true
}

func (th *Thread) writeTask(t task, tryStarve bool) bool {
	th.mu.Lock()
	if th.killed {
		th.mu.Unlock()
		return false
	}
	frame := &frame{
		mu:      sync.Mutex{},
		next:    nil,
		parent:  th,
		threads: common.NewSet[*Thread](),
		tasks:   []task{t},
		writing: true,
	}
	if th.tail == nil {
		f := frame.start(false)
		th.tail = frame
		defer f()
		th.mu.Unlock()
		return true
	}
	defer th.mu.Unlock()
	th.tail.setNext(frame)
	if !th.tail.writing && tryStarve {
		return true
	}
	th.tail = frame
	return true
}

func (t *Thread) R() *JoinedThread {
	return &JoinedThread{
		mode:   joinRead,
		thread: t,
	}
}

func (t *Thread) W() *JoinedThread {
	return &JoinedThread{
		mode:   joinWrite,
		thread: t,
	}
}

func (t *Thread) Ws() *JoinedThread {
	return &JoinedThread{
		mode:   joinWriteStarve,
		thread: t,
	}
}

func (t *Thread) Ri() *JoinedThread {
	return &JoinedThread{
		mode:    joinRead,
		thread:  t,
		instant: true,
	}
}

func (t *Thread) Wi() *JoinedThread {
	return &JoinedThread{
		mode:    joinWrite,
		thread:  t,
		instant: true,
	}
}

func (t *Thread) Wsi() *JoinedThread {
	return &JoinedThread{
		mode:    joinWriteStarve,
		thread:  t,
		instant: true,
	}
}

func Run(f func(*Thread), threads ...*JoinedThread) {
	if len(threads) == 0 {
		panic(errors.New("Threads to run on are not provided"))
	}
	task := &multitask{
		queue: threads,
		thread: &Thread{
			mu:      sync.Mutex{},
			main:    f,
			heads:   make([]threadHead, len(threads)),
			spawner: threads[0].thread.spawner,
			tail:    nil,
		},
	}
	task.next()
}
