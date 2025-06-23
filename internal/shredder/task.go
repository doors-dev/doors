package shredder

import (
	"sync"
	"sync/atomic"
)

type task interface {
	spawn(*frame, bool) *Thread
}

type multiTask struct {
	size      int32
	heads     []*JoinedThread
	thread    *Thread
	cursor    atomic.Int32
	countDown atomic.Int32
}

func runTask(s *Spawner, f func(*Thread), threads []*JoinedThread) {
	task := &multiTask{
		size:      int32(len(threads)),
		cursor:    atomic.Int32{},
		countDown: atomic.Int32{},
		heads:     threads,
		thread: &Thread{
			mu:      sync.Mutex{},
			main:    f,
			heads:   make([]threadHead, len(threads)),
			spawner: s,
			writing: false,
			tail:    nil,
		},
	}
	task.cursor.Store(-1)
	task.countDown.Store(task.size)
	_, ok := task.shift()
    if !ok {
        f(nil)
    }
}

func (m *multiTask) shift() (int32, bool) {
	cursor := m.cursor.Add(1)
	ready := cursor == m.size
	ok := true
	if !ready {
		thread := m.heads[cursor]
		if thread.write {
			ok = thread.thread.writeTask(m)
		} else {
			ok = thread.thread.readTask(m)
		}
	}
	return cursor, ok
}

func (m *multiTask) spawn(f *frame, killed bool) *Thread {
	var cursor int32
	var ok bool
	if killed {
		cursor = m.cursor.Load() + 1
		ok = false
	} else {
		cursor, ok = m.shift()
	}
	index := int32(-1)
	var countDown int32
	if ok {
		index = m.size - cursor
		m.thread.mu.Lock()
		m.thread.heads[index] = f
		m.thread.mu.Unlock()
		countDown = m.countDown.Add(-1)
	} else {
		countDown = m.countDown.Add(-m.size + cursor - 1)
	}
	if countDown == 0 {
		ok = m.cursor.Load() == m.size
		if ok {
			ok = m.thread.spawn()
		}
	}
	if !ok {
		if index != -1 {
			m.thread.mu.Lock()
			m.thread.heads[index] = nil
			m.thread.mu.Unlock()
		}
		if countDown == 0 {
			m.thread.abort()
		}
		return nil
	}
	return m.thread
}
