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

package shredder

import (
	"sync"
	"sync/atomic"
)

type threadFrame struct {
	mu        sync.Mutex
	next      *threadFrame
	completed bool
	baseFrame
}

func (f *threadFrame) setNext(next *threadFrame) {
	f.mu.Lock()
	completed := f.completed
	f.next = next
	f.mu.Unlock()
	if !completed {
		return
	}
	next.activate()
}

func (f *threadFrame) onComplete() {
	f.mu.Lock()
	f.completed = true
	next := f.next
	f.mu.Unlock()
	if next == nil {
		return
	}
	next.activate()
}

type Thread struct {
	frame atomic.Pointer[threadFrame]
}

func (s *Thread) Guard() Guard {
	return s.Frame()
}

func (s *Thread) Frame() Frame {
	frame := &threadFrame{}
	frame.baseFrame.onComplete = frame.onComplete
	prev := s.frame.Swap(frame)
	if prev == nil {
		frame.activate()
	} else {
		prev.setNext(frame)
	}
	return frame
}
