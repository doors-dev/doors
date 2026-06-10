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
	"testing"
)

// Regression tests for AUDIT #13 (cinema read-frame leak in door/tracker
// writeFrame). ReadStarveWriteThread.Read() returns a Join wrapper around the
// internal read frame; the wrapper must be released for the read to ever
// complete. Joining it with release=false and dropping the reference pins the
// thread head forever and starves all queued writes.

// An unreleased read wrapper pins the thread: a write queued after the joined
// work completed never activates. This is the leaking pattern
// Join(ctx, false, ..., thread.Read()) with the wrapper reference dropped.
func TestReadStarveWriteThreadUnreleasedReadWrapperPinsWrites(t *testing.T) {
	runtime := NewRuntime(context.Background(), 1, testShutdown{})
	t.Cleanup(runtime.Cancel)
	ctx := context.Background()

	var thread ReadStarveWriteThread

	readWrapper := thread.Read()
	outer := Join(ctx, false, readWrapper) // wrapper never released

	workDone := make(chan bool, 1)
	outer.Run(ctx, runtime, func(b bool) { workDone <- b })
	outer.Release()
	expectBool(t, workDone, "joined work")

	w := thread.Write()
	fired := make(chan bool, 1)
	w.Run(ctx, runtime, func(b bool) { fired <- b })
	w.Release()

	// the write thread issued Release on the internal read frame, but the
	// dropped wrapper still holds its counter — the write must stay buffered
	expectNoSignal(t, fired, "write activation behind an unreleased read wrapper")
}

// The correct pattern (beam/screen.go): Join with release=true. The wrapper
// completes once the joined work completes, the read drains, and a queued
// write activates.
func TestReadStarveWriteThreadReleasedReadWrapperUnblocksWrites(t *testing.T) {
	runtime := NewRuntime(context.Background(), 1, testShutdown{})
	t.Cleanup(runtime.Cancel)
	ctx := context.Background()

	var thread ReadStarveWriteThread

	readWrapper := thread.Read()
	outer := Join(ctx, true, readWrapper)

	workDone := make(chan bool, 1)
	outer.Run(ctx, runtime, func(b bool) { workDone <- b })
	outer.Release()
	expectBool(t, workDone, "joined work")

	w := thread.Write()
	fired := make(chan bool, 1)
	w.Run(ctx, runtime, func(b bool) { fired <- b })
	w.Release()

	if !expectBool(t, fired, "write activation after read wrapper completion") {
		t.Fatal("write ran cancelled")
	}
}

// Ordering guarantee under release=true: a write issued while the joined work
// is still in flight must not run before that work completes — releasing the
// wrapper does not detach it from the outer frame's lifecycle.
func TestReadStarveWriteThreadWriteWaitsForInflightRead(t *testing.T) {
	runtime := NewRuntime(context.Background(), 1, testShutdown{})
	t.Cleanup(runtime.Cancel)
	ctx := context.Background()

	var thread ReadStarveWriteThread

	readWrapper := thread.Read()
	outer := Join(ctx, true, readWrapper) // joined work still open

	w := thread.Write()
	fired := make(chan bool, 1)
	w.Run(ctx, runtime, func(b bool) { fired <- b })
	w.Release()

	expectNoSignal(t, fired, "write activation before joined work completion")

	outer.Release() // joined work completes

	expectBool(t, fired, "write activation after joined work completion")
}
