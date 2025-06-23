package shredder

import (
	"sync"

	"github.com/doors-dev/doors/internal/common"
)

type frame struct {
	mu      sync.Mutex
	next    *frame
	parent  *Thread
	tasks   []task
	threads common.Set[*Thread]
}

func (f *frame) listThreads() []*Thread {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.threads.Slice()
}

func (f *frame) isDone() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.threads.Len() == 0
}

func (f *frame) threadDone(t *Thread) {
	f.mu.Lock()
	defer func() {
		done := f.threads.Len() == 0
		f.mu.Unlock()
		if done {
			f.parent.frameDone(f)
		}
	}()
	f.threads.Remove(t)
}

func (f *frame) run(task task, killed bool) {
	thread := task.spawn(f, killed)
	if thread == nil {
		return
	}
	f.threads.Add(thread)
}

func (f *frame) start(killed bool) func() {
	f.mu.Lock()
	return func() {
		defer func() {
			done := f.threads.Len() == 0
			f.mu.Unlock()
			if done {
				f.parent.frameDone(f)
			}
		}()
		for _, task := range f.tasks {
			f.run(task, killed)
		}
		f.tasks = nil
	}
}

func (f *frame) add(task task, killed bool) func() {
	f.mu.Lock()
	return func() {
		defer f.mu.Unlock()
		if f.tasks == nil {
			f.run(task, killed)
			return
		}
		f.tasks = append(f.tasks, task)
	}
}
