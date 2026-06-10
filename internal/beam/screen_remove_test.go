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

package beam

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/shredder"
)

type screenTestShutdown struct{}

func (screenTestShutdown) Kill()                {}
func (screenTestShutdown) Logger() *slog.Logger { return slog.Default() }

type screenTestDoor struct {
	rt  shredder.Runtime
	ctx context.Context
}

func (d *screenTestDoor) Runtime() shredder.Runtime { return d.rt }
func (d *screenTestDoor) ReadFrame() shredder.Frame { return shredder.FreeFrame{} }
func (d *screenTestDoor) Context() context.Context  { return d.ctx }

type screenTestCore struct{ c Cinema }

func (r screenTestCore) Cinema() Cinema { return r.c }

func screenTestCtx(t *testing.T) context.Context {
	t.Helper()
	rt := shredder.NewRuntime(context.Background(), 8, screenTestShutdown{})
	t.Cleanup(rt.Cancel)
	door := &screenTestDoor{rt: rt, ctx: context.Background()}
	cin := NewCinema(nil, door)
	return context.WithValue(context.Background(), common.KeyCore, screenTestCore{cin})
}

func expectValue(t *testing.T, got <-chan int, want int) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case v := <-got:
			if v == want {
				return
			}
		case <-deadline:
			t.Fatalf("subscriber never received %d", want)
		}
	}
}

// Regression for screen removal (scheduleRemove drain): a transient Read
// leaves the screen empty and schedules removal; a subscription added right
// after must still receive subsequent updates.
func TestScreenRemoveThenResubscribe(t *testing.T) {
	ctx := screenTestCtx(t)
	src := NewSource(0, DefaultEqual[int], false)

	v, ok := src.Read(ctx)
	if !ok || v != 0 {
		t.Fatalf("read got %d ok=%v", v, ok)
	}

	got := make(chan int, 10)
	if !src.Sub(ctx, func(_ context.Context, v int) bool {
		got <- v
		return false
	}) {
		t.Fatal("sub failed")
	}

	src.Update(ctx, 1)
	expectValue(t, got, 1)
}

// Same, but the removal has had time to fully run before re-subscribing —
// the screen must be re-created and re-attached to the source.
func TestScreenRemoveSettleThenResubscribe(t *testing.T) {
	ctx := screenTestCtx(t)
	src := NewSource(0, DefaultEqual[int], false)

	if _, ok := src.Read(ctx); !ok {
		t.Fatal("read failed")
	}
	time.Sleep(100 * time.Millisecond)

	got := make(chan int, 10)
	if !src.Sub(ctx, func(_ context.Context, v int) bool {
		got <- v
		return false
	}) {
		t.Fatal("sub failed")
	}

	src.Update(ctx, 1)
	expectValue(t, got, 1)
}
