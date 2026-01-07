// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

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
	writing bool
}

func (f *frame) setNext(next *frame) {
	if f.next == nil {
		f.next = next
		return
	}
	f.next.setNext(next)
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

func (f *frame) run(task task, killed bool) func() {
	thread, after := task.spawn(f, killed)
	f.threads.Add(thread)
	return after
}

func (f *frame) start(killed bool) func() {
	f.mu.Lock()
	defer f.mu.Unlock()
	after := make([]func(), len(f.tasks))
	for i, task := range f.tasks {
		after[i] = f.run(task, killed)
	}
	f.tasks = nil
	return func() {
		for _, af := range after {
			af()
		}
	}
}

func (f *frame) add(task task, killed bool) func() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.tasks == nil {
		return f.run(task, killed)
	}
	f.tasks = append(f.tasks, task)
	return func() {}
}
