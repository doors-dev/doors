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
	"sync"

	"github.com/doors-dev/doors/internal/common"
)

type afterFrameState int

const (
	afterWaiting = iota
	afterActivated
	afterFired
)

type AfterFrame struct {
	mu      sync.Mutex
	state   afterFrameState
	counter int
	valve   ValveFrame
}

func (f *AfterFrame) Activate() {
	f.mu.Lock()
	if f.state != afterWaiting {
		f.mu.Unlock()
		return
	}
	if f.counter != 0 {
		f.state = afterActivated
		f.mu.Unlock()
		return
	}
	f.state = afterFired
	f.valve.reverse = true
	f.mu.Unlock()
	f.valve.Activate()
}

func (f *AfterFrame) After() SimpleFrame {
	return &f.valve
}

func (f *AfterFrame) schedule(e executable, _ *slog.Logger) {
	f.mu.Lock()
	if f.state != afterFired {
		f.counter += 1
	}
	f.mu.Unlock()
	e.execute(f.report)
}

func (f *AfterFrame) report(error) {
	f.mu.Lock()
	if f.state == afterFired {
		f.mu.Unlock()
		return
	}
	f.counter -= 1
	if f.counter > 0 || f.state != afterActivated {
		f.mu.Unlock()
		return
	}
	f.state = afterFired
	f.valve.reverse = true
	f.mu.Unlock()
	f.valve.Activate()
}

func (f *AfterFrame) Run(ctx context.Context, r Runtime, fun func(bool)) {
	f.schedule(run{runtime: r, ctx: ctx, fun: fun}, common.Logger(ctx))

}

func (f *AfterFrame) Submit(ctx context.Context, r Runtime, fun func(bool)) {
	f.schedule(spawn{runtime: r, ctx: ctx, fun: fun}, common.Logger(ctx))
}

var _ SimpleFrame = (*AfterFrame)(nil)
