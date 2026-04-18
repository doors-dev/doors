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
	"slices"
	"sync"
)

type AfterFrame struct {
	mu        sync.Mutex
	counter   int
	fired     bool
	activated bool
	after     []executable
}

func (f *AfterFrame) Activate() {
	f.mu.Lock()
	if f.activated {
		f.mu.Unlock()
		return
	}
	f.activated = true
	if f.counter != 0 {
		f.mu.Unlock()
		return
	}
	after := f.after
	f.after = nil
	f.fired = true
	f.mu.Unlock()
	for _, e := range slices.Backward(after) {
		e.execute(func(err error) {})
	}
}

func (f *AfterFrame) RunAfter(ctx context.Context, r Runtime, fun func(bool)) {
	e := run{runtime: r, ctx: ctx, fun: fun}
	f.mu.Lock()
	if f.fired {
		f.mu.Unlock()
		e.execute(func(err error) {})
		return
	}
	f.after = append(f.after, e)
	f.mu.Unlock()
}

func (f *AfterFrame) schedule(e executable) {
	f.mu.Lock()
	if f.fired {
		f.mu.Unlock()
		e.execute(func(err error) {})
		return
	}
	f.counter += 1
	f.mu.Unlock()
	e.execute(f.report)
}

func (f *AfterFrame) report(error) {
	f.mu.Lock()
	if f.fired {
		f.mu.Unlock()
		panic("can't report after fire")
	}
	f.counter -= 1
	if f.counter != 0 || !f.activated {
		f.mu.Unlock()
		return
	}
	after := f.after
	f.after = nil
	f.fired = true
	f.mu.Unlock()
	for _, e := range slices.Backward(after) {
		e.execute(func(err error) {})
	}
}

func (f *AfterFrame) Run(ctx context.Context, r Runtime, fun func(bool)) {
	f.schedule(run{runtime: r, ctx: ctx, fun: fun})

}

func (f *AfterFrame) Submit(ctx context.Context, r Runtime, fun func(bool)) {
	f.schedule(spawn{runtime: r, ctx: ctx, fun: fun})
}

var _ SimpleFrame = &AfterFrame{}
