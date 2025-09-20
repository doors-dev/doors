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
	"sync"
)

type task interface {
	spawn(*frame, bool) (*Thread, func())
}

type multitask struct {
	mu     sync.Mutex
	heads  []threadHead
	count  int
	queue  []*JoinedThread
	thread *Thread
	killed bool
}

func (m *multitask) next() {
	m.mu.Lock()
	toStart := make([]*JoinedThread, 0)
	newQueue := make([]*JoinedThread, 0)
	m.queue[0].instant = true
	saveToQueue := false
	for _, thread := range m.queue {
		if saveToQueue || !thread.instant {
			saveToQueue = true
			newQueue = append(newQueue, thread)
			continue
		}
		toStart = append(toStart, thread)
	}
	m.queue = newQueue
	m.count = len(toStart)
	m.mu.Unlock()
	for _, thread := range toStart {
		ok := thread.start(m)
		if !ok {
			m.mu.Lock()
			if m.killed {
				m.mu.Unlock()
				return
			}
			m.killed = true
			m.mu.Unlock()
			m.thread.abort()
			return
		}
	}
}

func (m *multitask) spawn(f *frame, killed bool) (*Thread, func()) {
	m.mu.Lock() // here
	defer m.mu.Unlock()
	if m.killed {
		return m.thread, func() {
			f.threadDone(m.thread)
		}
	}
	m.thread.addHead(f)
	if killed {
		m.killed = true
		return m.thread, func() {
			m.thread.abort()
		}
	}
	m.count -= 1
	if m.count != 0 {
		return m.thread, func() {}
	}
	if len(m.queue) != 0 {
		return m.thread, func() {
			m.next()
		}
	}
	ok := m.thread.spawn()
	if !ok {
		return m.thread, func() {
			m.thread.abort()
		}
	}
	return m.thread, func() {}
}
