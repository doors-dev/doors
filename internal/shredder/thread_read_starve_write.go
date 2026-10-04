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
	"context"
	"sync"

	"github.com/gammazero/deque"
)

type ReadStarveWriteThread struct {
	mu    sync.Mutex
	queue deque.Deque[Releaser]
}

func (t *ReadStarveWriteThread) Read() ReleaseFrame {
	t.mu.Lock()
	for frame := range t.queue.Iter() {
		if frame, ok := frame.(*ownedFrame); ok {
			read := Join(context.Background(), false, frame)
			t.mu.Unlock()
			return read
		}
	}
	frame := t.appendRead()
	frame.active = t.queue.Len() == 1
	read := Join(context.Background(), false, frame)
	t.mu.Unlock()
	return read
}

func (t *ReadStarveWriteThread) Write() ReleaseFrame {
	t.mu.Lock()
	var read *ownedFrame
	if t.queue.Len() != 0 {
		read, _ = t.queue.Back().(*ownedFrame)
	}
	frame := t.appendWrite()
	if t.queue.Len() == 1 {
		frame.activate()
	}
	t.mu.Unlock()
	if read != nil {
		read.Release()
	}
	return frame
}

func (t *ReadStarveWriteThread) appendWrite() *baseFrame {
	frame := &baseFrame{
		onComplete: func() {
			t.mu.Lock()
			t.queue.PopFront()
			if t.queue.Len() == 0 {
				t.mu.Unlock()
				return
			}
			next := t.queue.Front()
			t.mu.Unlock()
			switch f := next.(type) {
			case *baseFrame:
				f.activate()
			case *ownedFrame:
				f.activate()
			}
		},
	}
	t.queue.PushBack(frame)
	return frame
}

func (t *ReadStarveWriteThread) appendRead() *ownedFrame {
	frame := &ownedFrame{owner: t}
	t.queue.PushBack(frame)
	return frame
}

func (t *ReadStarveWriteThread) lock(op ownedOperation) {
	if op == scheduleOwned {
		return
	}
	t.mu.Lock()
}

func (t *ReadStarveWriteThread) unlock(op ownedOperation, done bool) {
	if op == scheduleOwned {
		return
	}
	if !done {
		t.mu.Unlock()
		return
	}
	t.queue.PopFront()
	if t.queue.Len() == 0 {
		t.mu.Unlock()
		return
	}
	next := t.queue.Front()
	t.mu.Unlock()
	next.(*baseFrame).activate()
}
