package shredder

import (
	"sync"

	"github.com/doors-dev/doors/internal/common"
)

type threadHead interface {
	threadDone(*Thread)
}

type Thread struct {
	mu      sync.Mutex
	main    func(*Thread)
	heads   []threadHead
	killed  bool
	running bool
	spawner *Spawner
	writing bool
	tail    *frame
	after   func()
}

func (t *Thread) addHead(head threadHead) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.spawner == nil {
		return false
	}
	t.heads = append([]threadHead{head}, t.heads...)
	return true
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
	return t.spawner.Go(func() {
		t.main(t)
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
		return
	}
	if frame.next != nil {
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
	if th.tail != nil && !th.writing {
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
	th.writing = false
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
	}
	if th.tail == nil {
		f := frame.start(false)
		defer f()
	} else {
		th.tail.setNext(frame)
	}
	if !th.writing && tryStarve {
		th.mu.Unlock()
		return true
	}
	th.writing = true
	th.tail = frame
	th.mu.Unlock()
	return true
}/*
func (th *Thread) writeTask(t task) bool {
	th.mu.Lock()
	if th.killed {
		th.mu.Unlock()
		return false
	}
	th.writing = true
	frame := &frame{
		mu:      sync.Mutex{},
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
} */

func (th *Thread) Read(f func(*Thread), joined ...*JoinedThread) {
	th.executeMulti(f, R(th), joined)
}

func (th *Thread) WriteStarving(f func(*Thread), joined ...*JoinedThread) {
	th.executeMulti(f, WS(th), joined)
}
func (th *Thread) Write(f func(*Thread), joined ...*JoinedThread) {
	th.executeMulti(f, W(th), joined)
}
func (th *Thread) ReadInstant(f func(*Thread), joined ...*JoinedThread) {
	th.executeInstant(f, R(th), joined)
}

func (th *Thread) WriteInstant(f func(*Thread), joined ...*JoinedThread) {
	th.executeInstant(f, W(th), joined)
}
func (th *Thread) WriteInstantStarving(f func(*Thread), joined ...*JoinedThread) {
	th.executeInstant(f, WS(th), joined)
}

func (th *Thread) executeMulti(f func(*Thread), self *JoinedThread, joined []*JoinedThread) {
	runMultiTask(th.spawner, f, append([]*JoinedThread{self}, joined...))
}
func (th *Thread) executeInstant(f func(*Thread), self *JoinedThread, joined []*JoinedThread) {
	runInstantTask(th.spawner, f, append([]*JoinedThread{self}, joined...))
}
