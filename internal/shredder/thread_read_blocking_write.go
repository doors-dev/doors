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
)

type ReadBlockingWriteThread struct {
	mu       sync.Mutex
	read     *baseFrame
	nextRead *baseFrame
	write    *baseFrame
}

func (f *ReadBlockingWriteThread) init() {
	if f.read != nil {
		return
	}
	f.read = &baseFrame{
		onComplete: f.complete,
	}
	f.read.activate()
}

func (f *ReadBlockingWriteThread) Read() Frame {
	f.mu.Lock()
	f.init()
	defer f.mu.Unlock()
	return Join(false, f.read)
}

func (f *ReadBlockingWriteThread) complete() {
	f.mu.Lock()
	f.read = f.nextRead
	f.nextRead = nil
	f.write.activate()
}

func (f *ReadBlockingWriteThread) Write() (write Frame, read Frame) {
	f.mu.Lock()
	f.init()
	if f.write != nil {
		f.mu.Unlock()
		panic("blocking frame contract violation: blocking frame is already issued")
	}
	f.nextRead = &baseFrame{
		onComplete: f.complete,
	}
	f.write = &baseFrame{
		onComplete: func() {
			f.write = nil
			f.mu.Unlock()
			f.read.activate()
		},
	}
	read = Join(false, f.nextRead)
	write = f.write
	f.mu.Unlock()
	f.read.Release()
	return
}
