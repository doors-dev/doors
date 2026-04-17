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
	"time"
)

type testShutdown struct{}

func (testShutdown) Shutdown() {}

func expectBool(t *testing.T, ch <-chan bool, name string) bool {
	t.Helper()
	select {
	case got := <-ch:
		return got
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", name)
		return false
	}
}

func expectFrame(t *testing.T, ch <-chan Frame, name string) Frame {
	t.Helper()
	select {
	case got := <-ch:
		return got
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", name)
		return nil
	}
}

func expectNoSignal[T any](t *testing.T, ch <-chan T, name string) {
	t.Helper()
	select {
	case <-ch:
		t.Fatalf("unexpected %s", name)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestReadBlockingWriteThreadWriterLifecycle(t *testing.T) {
	runtime := NewRuntime(context.Background(), 1, testShutdown{})
	t.Cleanup(runtime.Cancel)

	var thread ReadBlockingWriteThread

	readOne := thread.Read()
	readTwo := thread.Read()
	t.Cleanup(readOne.Release)
	t.Cleanup(readTwo.Release)

	writeFrame, nextReadFrame := thread.Write()
	t.Cleanup(writeFrame.Release)
	t.Cleanup(nextReadFrame.Release)

	writerStarted := make(chan bool, 1)
	writerDone := make(chan struct{})
	writeFrame.Submit(context.Background(), runtime, func(ok bool) {
		writerStarted <- ok
		<-writerDone
	})

	nextReadStarted := make(chan bool, 1)
	nextReadFrame.Submit(context.Background(), runtime, func(ok bool) {
		nextReadStarted <- ok
	})

	starvingReadDone := make(chan Frame, 1)
	go func() {
		starvingReadDone <- thread.Read()
	}()
	starvingRead := expectFrame(t, starvingReadDone, "starving read frame")
	t.Cleanup(starvingRead.Release)

	readOne.Release()
	readTwo.Release()
	expectNoSignal(t, writerStarted, "writer activation before starving read release")

	starvingRead.Release()
	if !expectBool(t, writerStarted, "writer activation") {
		t.Fatal("expected writer task to run with ok=true")
	}
	expectNoSignal(t, nextReadStarted, "next read activation while writer is active")

	blockedReadDone := make(chan Frame, 1)
	go func() {
		blockedReadDone <- thread.Read()
	}()
	expectNoSignal(t, blockedReadDone, "read issuance during active writer")

	close(writerDone)
	expectNoSignal(t, blockedReadDone, "read issuance before write frame release")
	expectNoSignal(t, nextReadStarted, "next read activation before write frame release")

	writeFrame.Release()

	blockedRead := expectFrame(t, blockedReadDone, "read issuance after write completion")
	t.Cleanup(blockedRead.Release)
	if !expectBool(t, nextReadStarted, "next read activation after write completion") {
		t.Fatal("expected next read task to run with ok=true")
	}
}

func TestReadBlockingWriteThreadRejectsSecondPendingWrite(t *testing.T) {
	var thread ReadBlockingWriteThread

	readFrame := thread.Read()
	t.Cleanup(readFrame.Release)

	writeFrame, nextReadFrame := thread.Write()
	t.Cleanup(writeFrame.Release)
	t.Cleanup(nextReadFrame.Release)

	defer func() {
		if recover() == nil {
			t.Fatal("expected second pending write to panic")
		}
	}()

	thread.Write()
}

func TestReadWriteThreadSerializesWrites(t *testing.T) {
	runtime := NewRuntime(context.Background(), 1, testShutdown{})
	t.Cleanup(runtime.Cancel)

	var thread ReadWriteThread

	readOne := thread.Read()
	readTwo := thread.Read()
	t.Cleanup(readOne.Release)
	t.Cleanup(readTwo.Release)

	writeOne := thread.Write()
	t.Cleanup(writeOne.Release)

	midReadDone := make(chan Frame, 1)
	go func() {
		midReadDone <- thread.Read()
	}()
	midRead := expectFrame(t, midReadDone, "mid read frame")
	t.Cleanup(midRead.Release)

	writeTwoDone := make(chan Frame, 1)
	go func() {
		writeTwoDone <- thread.Write()
	}()
	writeTwo := expectFrame(t, writeTwoDone, "second write frame")
	t.Cleanup(writeTwo.Release)

	tailReadDone := make(chan Frame, 1)
	go func() {
		tailReadDone <- thread.Read()
	}()
	tailRead := expectFrame(t, tailReadDone, "tail read frame")
	t.Cleanup(tailRead.Release)

	writeOneStarted := make(chan bool, 1)
	writeOne.Submit(context.Background(), runtime, func(ok bool) {
		writeOneStarted <- ok
	})

	midReadStarted := make(chan bool, 1)
	midRead.Submit(context.Background(), runtime, func(ok bool) {
		midReadStarted <- ok
	})

	writeTwoStarted := make(chan bool, 1)
	writeTwo.Submit(context.Background(), runtime, func(ok bool) {
		writeTwoStarted <- ok
	})

	tailReadStarted := make(chan bool, 1)
	tailRead.Submit(context.Background(), runtime, func(ok bool) {
		tailReadStarted <- ok
	})

	readOne.Release()
	readTwo.Release()
	if !expectBool(t, writeOneStarted, "first write activation") {
		t.Fatal("expected first write to run with ok=true")
	}
	expectNoSignal(t, midReadStarted, "mid read activation before first write release")
	expectNoSignal(t, writeTwoStarted, "second write activation before first write release")
	expectNoSignal(t, tailReadStarted, "tail read activation before first write release")

	writeOne.Release()
	if !expectBool(t, midReadStarted, "mid read activation") {
		t.Fatal("expected mid read to run with ok=true")
	}
	expectNoSignal(t, writeTwoStarted, "second write activation before mid read release")
	expectNoSignal(t, tailReadStarted, "tail read activation before mid read release")

	midRead.Release()
	if !expectBool(t, writeTwoStarted, "second write activation") {
		t.Fatal("expected second write to run with ok=true")
	}
	expectNoSignal(t, tailReadStarted, "tail read activation before second write release")

	writeTwo.Release()
	if !expectBool(t, tailReadStarted, "tail read activation") {
		t.Fatal("expected tail read to run with ok=true")
	}
}

func TestReadWriteThreadQueuedWriteBlocksLaterReads(t *testing.T) {
	var thread ReadWriteThread

	initialRead := thread.Read()
	t.Cleanup(initialRead.Release)

	writeFrame := thread.Write()
	t.Cleanup(writeFrame.Release)

	lateReadOne := thread.Read()
	lateReadTwo := thread.Read()
	t.Cleanup(lateReadOne.Release)
	t.Cleanup(lateReadTwo.Release)

	writeStarted := make(chan bool, 1)
	writeFrame.Run(context.Background(), nil, func(ok bool) {
		writeStarted <- ok
	})

	lateReadOneStarted := make(chan bool, 1)
	lateReadOne.Run(context.Background(), nil, func(ok bool) {
		lateReadOneStarted <- ok
	})

	lateReadTwoStarted := make(chan bool, 1)
	lateReadTwo.Run(context.Background(), nil, func(ok bool) {
		lateReadTwoStarted <- ok
	})

	expectNoSignal(t, writeStarted, "write activation before initial read release")
	expectNoSignal(t, lateReadOneStarted, "first late read activation before initial read release")
	expectNoSignal(t, lateReadTwoStarted, "second late read activation before initial read release")

	initialRead.Release()
	if !expectBool(t, writeStarted, "write activation after initial read release") {
		t.Fatal("expected write to run with ok=true")
	}
	expectNoSignal(t, lateReadOneStarted, "first late read activation before write release")
	expectNoSignal(t, lateReadTwoStarted, "second late read activation before write release")

	writeFrame.Release()
	if !expectBool(t, lateReadOneStarted, "first late read activation after write release") {
		t.Fatal("expected first late read to run with ok=true")
	}
	if !expectBool(t, lateReadTwoStarted, "second late read activation after write release") {
		t.Fatal("expected second late read to run with ok=true")
	}
}

func TestReadStarveWriteThreadQueuedWriteIsStarvedByLaterReads(t *testing.T) {
	var thread ReadStarveWriteThread

	initialRead := thread.Read()
	t.Cleanup(initialRead.Release)

	writeFrame := thread.Write()
	t.Cleanup(writeFrame.Release)

	writerStarted := make(chan bool, 1)
	writeFrame.Run(context.Background(), nil, func(ok bool) {
		writerStarted <- ok
	})

	lateReadOneDone := make(chan Frame, 1)
	go func() {
		lateReadOneDone <- thread.Read()
	}()
	lateReadOne := expectFrame(t, lateReadOneDone, "first late read frame")
	t.Cleanup(lateReadOne.Release)

	lateReadTwoDone := make(chan Frame, 1)
	go func() {
		lateReadTwoDone <- thread.Read()
	}()
	lateReadTwo := expectFrame(t, lateReadTwoDone, "second late read frame")
	t.Cleanup(lateReadTwo.Release)

	lateReadOneStarted := make(chan bool, 1)
	lateReadOne.Run(context.Background(), nil, func(ok bool) {
		lateReadOneStarted <- ok
	})
	if !expectBool(t, lateReadOneStarted, "first late read activation") {
		t.Fatal("expected first late read to run with ok=true")
	}

	lateReadTwoStarted := make(chan bool, 1)
	lateReadTwo.Run(context.Background(), nil, func(ok bool) {
		lateReadTwoStarted <- ok
	})
	if !expectBool(t, lateReadTwoStarted, "second late read activation") {
		t.Fatal("expected second late read to run with ok=true")
	}

	expectNoSignal(t, writerStarted, "writer activation before reads release")

	initialRead.Release()
	expectNoSignal(t, writerStarted, "writer activation before first late read release")

	lateReadOne.Release()
	expectNoSignal(t, writerStarted, "writer activation before all late reads release")

	lateReadTwo.Release()
	if !expectBool(t, writerStarted, "writer activation after all starving reads release") {
		t.Fatal("expected writer to run with ok=true")
	}
}

func TestReadStarveWriteThreadStarvesAndSerializesPendingWrites(t *testing.T) {
	runtime := NewRuntime(context.Background(), 1, testShutdown{})
	t.Cleanup(runtime.Cancel)

	var thread ReadStarveWriteThread

	readOne := thread.Read()
	readTwo := thread.Read()
	t.Cleanup(readOne.Release)
	t.Cleanup(readTwo.Release)

	writeOne := thread.Write()
	t.Cleanup(writeOne.Release)

	writeOneStarted := make(chan bool, 1)
	writeOne.Submit(context.Background(), runtime, func(ok bool) {
		writeOneStarted <- ok
	})

	starvingReadDone := make(chan Frame, 1)
	go func() {
		starvingReadDone <- thread.Read()
	}()
	starvingRead := expectFrame(t, starvingReadDone, "starving read frame")
	t.Cleanup(starvingRead.Release)

	starvingReadStarted := make(chan bool, 1)
	starvingRead.Submit(context.Background(), runtime, func(ok bool) {
		starvingReadStarted <- ok
	})
	if !expectBool(t, starvingReadStarted, "starving read activation") {
		t.Fatal("expected starving read to run with ok=true")
	}

	readOne.Release()
	readTwo.Release()
	expectNoSignal(t, writeOneStarted, "first write activation before starving read release")

	starvingRead.Release()
	if !expectBool(t, writeOneStarted, "first write activation") {
		t.Fatal("expected first write to run with ok=true")
	}

	writeTwoDone := make(chan Frame, 1)
	go func() {
		writeTwoDone <- thread.Write()
	}()
	writeTwo := expectFrame(t, writeTwoDone, "second write frame")
	t.Cleanup(writeTwo.Release)

	futureReadDone := make(chan Frame, 1)
	go func() {
		futureReadDone <- thread.Read()
	}()
	futureRead := expectFrame(t, futureReadDone, "future read frame")
	t.Cleanup(futureRead.Release)

	writeThreeDone := make(chan Frame, 1)
	go func() {
		writeThreeDone <- thread.Write()
	}()
	writeThree := expectFrame(t, writeThreeDone, "third write frame")
	t.Cleanup(writeThree.Release)

	writeTwoStarted := make(chan bool, 1)
	writeTwo.Submit(context.Background(), runtime, func(ok bool) {
		writeTwoStarted <- ok
	})

	futureReadStarted := make(chan bool, 1)
	futureRead.Submit(context.Background(), runtime, func(ok bool) {
		futureReadStarted <- ok
	})

	writeThreeStarted := make(chan bool, 1)
	writeThree.Submit(context.Background(), runtime, func(ok bool) {
		writeThreeStarted <- ok
	})

	expectNoSignal(t, writeTwoStarted, "second write activation before first write release")
	expectNoSignal(t, futureReadStarted, "future read activation before first write release")
	expectNoSignal(t, writeThreeStarted, "third write activation before first write release")

	writeOne.Release()
	if !expectBool(t, writeTwoStarted, "second write activation") {
		t.Fatal("expected second write to run with ok=true")
	}
	expectNoSignal(t, futureReadStarted, "future read activation before second write release")
	expectNoSignal(t, writeThreeStarted, "third write activation before second write release")

	writeTwo.Release()
	if !expectBool(t, futureReadStarted, "future read activation") {
		t.Fatal("expected future read to run with ok=true")
	}
	expectNoSignal(t, writeThreeStarted, "third write activation before future read release")

	futureRead.Release()
	if !expectBool(t, writeThreeStarted, "third write activation") {
		t.Fatal("expected third write to run with ok=true")
	}
}
