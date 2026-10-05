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
	"testing"
	"time"
)

type valveOrder struct {
	mu    sync.Mutex
	order []string
}

func (o *valveOrder) entry(name string) func(bool) {
	return func(bool) {
		o.mu.Lock()
		o.order = append(o.order, name)
		o.mu.Unlock()
	}
}

func (o *valveOrder) get() []string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return slices.Clone(o.order)
}

func activateValve(t *testing.T, valve *ValveFrame) <-chan struct{} {
	t.Helper()
	done := make(chan struct{})
	go func() {
		valve.Activate()
		close(done)
	}()
	return done
}

func waitDone(t *testing.T, done <-chan struct{}, name string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", name)
	}
}

// Activate runs the buffered entries in order. An entry scheduled from
// another goroutine while the drain runs waits for the entries buffered
// before it; once the valve is open, entries run at once.
func TestValveFrameActivateKeepsOrder(t *testing.T) {
	ctx := context.Background()
	var valve ValveFrame
	var order valveOrder
	entered := make(chan struct{})
	release := make(chan struct{})
	valve.Run(ctx, nil, func(b bool) {
		order.entry("e1")(b)
		close(entered)
		<-release
	})
	valve.Run(ctx, nil, order.entry("e2"))
	done := activateValve(t, &valve)
	waitDone(t, entered, "the first entry")
	valve.Run(ctx, nil, order.entry("e3"))
	close(release)
	waitDone(t, done, "Activate")
	valve.Run(ctx, nil, order.entry("e4"))
	if got := order.get(); !slices.Equal(got, []string{"e1", "e2", "e3", "e4"}) {
		t.Fatalf("expected entries in schedule order, got %v", got)
	}
}

// An entry scheduled by a buffered entry while it runs in the drain waits for
// the rest of the buffer, and the drain still finishes.
func TestValveFrameActivateReentrantSchedule(t *testing.T) {
	ctx := context.Background()
	var valve ValveFrame
	var order valveOrder
	valve.Run(ctx, nil, func(b bool) {
		order.entry("e1")(b)
		valve.Run(ctx, nil, order.entry("e3"))
	})
	valve.Run(ctx, nil, order.entry("e2"))
	waitDone(t, activateValve(t, &valve), "Activate")
	if got := order.get(); !slices.Equal(got, []string{"e1", "e2", "e3"}) {
		t.Fatalf("expected entries in schedule order, got %v", got)
	}
}
