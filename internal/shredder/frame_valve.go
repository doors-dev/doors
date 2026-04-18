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
	"sync/atomic"
)

type ValveFrame struct {
	mu     sync.Mutex
	buffer []executable
	active atomic.Bool
}

func (f *ValveFrame) Activate() {
	f.mu.Lock()
	if f.active.Load() {
		f.mu.Unlock()
		return
	}
	f.active.Store(true)
	buf := f.buffer
	f.buffer = nil
	f.mu.Unlock()
	for i, e := range buf {
		f.schedule(e)
		buf[i] = nil
	}
}

func (f *ValveFrame) schedule(e executable) {
	if f.active.Load() {
		e.execute(func(error) {})
		return
	}
	f.mu.Lock()
	if f.active.Load() {
		f.mu.Unlock()
		e.execute(func(error) {})
		return
	}
	f.buffer = append(f.buffer, e)
	f.mu.Unlock()
}

func (f *ValveFrame) Run(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(run{runtime: s, ctx: ctx, fun: fun})
}

func (f *ValveFrame) Submit(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(spawn{runtime: s, ctx: ctx, fun: fun})
}

var _ SimpleFrame = &ValveFrame{}
