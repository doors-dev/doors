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

import "sync"

type joinedFrame struct {
	mu        sync.Mutex
	callbacks []func(error)
	joinCount int
	baseFrame
}

func (f *joinedFrame) onComplete() {
	for _, callback := range f.callbacks {
		callback(nil)
	}
}

func (f *joinedFrame) register(callback func(error)) {
	f.mu.Lock()
	f.callbacks = append(f.callbacks, callback)
	ready := len(f.callbacks) == f.joinCount
	f.mu.Unlock()
	if !ready {
		return
	}
	f.activate()
}

func (j *joinedFrame) execute(callback func(error)) {
	j.register(callback)
}

func Join(release bool, frames ...AnyFrame) Frame {
	if len(frames) == 0 {
		panic("join must have frames")
	}
	joined := &joinedFrame{
		joinCount: len(frames),
	}
	joined.baseFrame.onComplete = joined.onComplete
	for _, frame := range frames {
		frame.schedule(joined)
		if release {
			if g, ok := frame.(Guard); ok {
				g.Release()
			}
		}
	}
	return joined
}
