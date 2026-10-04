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
)

type ReadBlockingWriteThread struct {
	mu       sync.Mutex
	read     *ownedFrame
	nextRead *ownedFrame
	write    *baseFrame
}

func (t *ReadBlockingWriteThread) init() {
	if t.read != nil {
		return
	}
	read := t.newReadFrame()
	t.read = read
	t.mu.Unlock()
	read.activate()
	t.mu.Lock()
}

func (t *ReadBlockingWriteThread) newReadFrame() *ownedFrame {
	return &ownedFrame{owner: t}
}

func (t *ReadBlockingWriteThread) lock(op ownedOperation) {
	if op == scheduleOwned {
		return
	}
	t.mu.Lock()
}

func (t *ReadBlockingWriteThread) unlock(op ownedOperation, done bool) {
	if op == scheduleOwned {
		return
	}
	if !done {
		t.mu.Unlock()
		return
	}
	t.read = t.nextRead
	t.nextRead = nil
	t.write.activate()
}

func (t *ReadBlockingWriteThread) Read() ReleaseFrame {
	t.mu.Lock()
	t.init()
	defer t.mu.Unlock()
	return Join(context.Background(), false, t.read)
}

func (t *ReadBlockingWriteThread) Write() (write ReleaseFrame, read ReleaseFrame) {
	t.mu.Lock()
	t.init()
	if t.write != nil {
		t.mu.Unlock()
		panic("blocking frame contract violation: blocking frame is already issued")
	}
	t.nextRead = t.newReadFrame()
	t.write = &baseFrame{
		onComplete: func() {
			t.write = nil
			t.mu.Unlock()
			t.read.activate()
		},
	}
	read = Join(context.Background(), false, t.nextRead)
	write = t.write
	t.mu.Unlock()
	t.read.Release()
	return
}
