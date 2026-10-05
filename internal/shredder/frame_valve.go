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

const (
	valveClosed int32 = iota
	valveDraining
	valveOpened
)

type ValveFrame struct {
	mu      sync.Mutex
	buffer  []bufferedExecutable
	state   atomic.Int32
	reverse bool
}

func (f *ValveFrame) Activate() {
	f.mu.Lock()
	if f.state.Load() > valveClosed {
		f.mu.Unlock()
		return
	}
	f.state.Store(valveDraining)
	buf := f.buffer
	f.buffer = nil
repeat:
	f.mu.Unlock()
	if f.reverse {
		for _, e := range slices.Backward(buf) {
			f.scheduleState(e.e, e.l, true)
		}
	} else {
		for _, e := range buf {
			f.scheduleState(e.e, e.l, true)
		}
	}
	f.mu.Lock()
	if len(f.buffer) > 0 {
		buf = f.buffer
		f.buffer = nil
		goto repeat
	}
	f.state.Store(valveOpened)
	f.mu.Unlock()
}

func (f *ValveFrame) scheduleState(e executable, l *slog.Logger, ingoreState bool) {
	if ingoreState {
		e.execute(func(error) {})
		return
	}
	if f.state.Load() == valveOpened {
		e.execute(func(error) {})
		return
	}
	f.mu.Lock()
	if f.state.Load() == valveOpened {
		f.mu.Unlock()
		e.execute(func(error) {})
		return
	}
	f.buffer = append(f.buffer, bufferedExecutable{e: e, l: l})
	f.mu.Unlock()
}

func (f *ValveFrame) schedule(e executable, l *slog.Logger) {
	f.scheduleState(e, l, false)
}

func (f *ValveFrame) Run(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(run{runtime: s, ctx: ctx, fun: fun}, common.Logger(ctx))
}

func (f *ValveFrame) Submit(ctx context.Context, s Runtime, fun func(bool)) {
	f.schedule(spawn{runtime: s, ctx: ctx, fun: fun}, common.Logger(ctx))
}

var _ Frame = &ValveFrame{}
