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
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/doors-dev/doors/internal/common"
)

type bufferedExecutable struct {
	e executable
	l *slog.Logger
}

type ValveFrame struct {
	mu      sync.Mutex
	buffer  []bufferedExecutable
	active  atomic.Bool
	reverse bool
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
	if f.reverse {
		for _, e := range slices.Backward(buf) {
			f.schedule(e.e, e.l)
		}
		return
	}
	for _, e := range buf {
		f.schedule(e.e, e.l)
	}
}

func (f *ValveFrame) schedule(e executable, l *slog.Logger) {
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
	f.buffer = append(f.buffer, bufferedExecutable{e: e, l: l})
	f.mu.Unlock()
}

func (f *ValveFrame) Run(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(run{runtime: s, ctx: ctx, fun: fun}, common.Logger(ctx))
}

func (f *ValveFrame) Submit(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(spawn{runtime: s, ctx: ctx, fun: fun}, common.Logger(ctx))
}

var _ SimpleFrame = &ValveFrame{}
