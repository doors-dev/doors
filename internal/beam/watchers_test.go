package beam

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/shredder"
)

type testCore struct {
	cinema Cinema
}

func (c testCore) Cinema() Cinema {
	return c.cinema
}

type testDoor struct {
	ctx    context.Context
	thread shredder.Thread
}

func (d *testDoor) NewFrame() shredder.Frame {
	return d.thread.Frame()
}

func (d *testDoor) Context() context.Context {
	return d.ctx
}

type noopShutdown struct{}

func (noopShutdown) Shutdown() {}

func newBeamContext(t *testing.T) context.Context {
	t.Helper()
	runtime := shredder.NewRuntime(context.Background(), 1, noopShutdown{})
	t.Cleanup(runtime.Cancel)

	door := &testDoor{ctx: context.Background()}
	cinema := NewCinema(nil, door, runtime)
	ctx := context.WithValue(context.Background(), ctex.KeyCore, testCore{cinema: cinema})
	door.ctx = ctx
	return ctx
}

func expectErr(t *testing.T, ch <-chan error) error {
	t.Helper()
	select {
	case err, ok := <-ch:
		if !ok {
			t.Fatal("expected channel result, got closed channel")
		}
		return err
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for error result")
		return nil
	}
}

func expectInt(t *testing.T, ch <-chan int) int {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for int value")
		return 0
	}
}

func expectString(t *testing.T, ch <-chan string) string {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for string value")
		return ""
	}
}

func expectSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for cancel signal")
	}
}

func TestSourceXUpdateAndXMutate(t *testing.T) {
	source := NewSourceEqual(0, nil)

	if err := expectErr(t, source.XUpdate(context.Background(), 1)); err != nil {
		t.Fatal(err)
	}
	if got := source.Get(); got != 1 {
		t.Fatal("unexpected source value after XUpdate:", got)
	}

	if err := expectErr(t, source.XMutate(context.Background(), func(v int) int {
		return v + 1
	})); err != nil {
		t.Fatal(err)
	}
	if got := source.Get(); got != 2 {
		t.Fatal("unexpected source value after XMutate:", got)
	}
}

func TestBeamWatcherExtendedAPIs(t *testing.T) {
	ctx := newBeamContext(t)
	source := NewSource(1)
	derived := NewBeam(source, func(v int) string {
		return fmt.Sprintf("v:%d", v)
	})

	readAndSubUpdates := make(chan string, 1)
	initial, ok := derived.ReadAndSub(ctx, func(ctx context.Context, value string) bool {
		readAndSubUpdates <- value
		return true
	})
	if !ok {
		t.Fatal("expected ReadAndSub to subscribe")
	}
	if initial != "v:1" {
		t.Fatal("unexpected initial derived value:", initial)
	}

	source.Update(ctx, 2)
	if got := expectString(t, readAndSubUpdates); got != "v:2" {
		t.Fatal("unexpected derived update:", got)
	}

	sourceReadAndSubUpdates := make(chan int, 1)
	sourceReadAndSubCanceled := make(chan struct{}, 1)
	sourceInitial, sourceCancel, ok := source.XReadAndSub(ctx, func(ctx context.Context, value int) bool {
		sourceReadAndSubUpdates <- value
		return false
	}, func() {
		close(sourceReadAndSubCanceled)
	})
	if !ok {
		t.Fatal("expected source XReadAndSub to subscribe")
	}
	if sourceInitial != 2 {
		t.Fatal("unexpected initial source value:", sourceInitial)
	}

	source.Update(ctx, 3)
	if got := expectInt(t, sourceReadAndSubUpdates); got != 3 {
		t.Fatal("unexpected source XReadAndSub update:", got)
	}
	sourceCancel()
	expectSignal(t, sourceReadAndSubCanceled)

	derivedReadAndSubUpdates := make(chan string, 1)
	derivedReadAndSubCanceled := make(chan struct{}, 1)
	derivedInitial, derivedCancel, ok := derived.XReadAndSub(ctx, func(ctx context.Context, value string) bool {
		derivedReadAndSubUpdates <- value
		return false
	}, func() {
		close(derivedReadAndSubCanceled)
	})
	if !ok {
		t.Fatal("expected derived XReadAndSub to subscribe")
	}
	if derivedInitial != "v:3" {
		t.Fatal("unexpected initial derived XReadAndSub value:", derivedInitial)
	}

	source.Update(ctx, 4)
	if got := expectString(t, derivedReadAndSubUpdates); got != "v:4" {
		t.Fatal("unexpected derived XReadAndSub update:", got)
	}
	derivedCancel()
	expectSignal(t, derivedReadAndSubCanceled)

	sourceSubUpdates := make(chan int, 2)
	sourceSubCanceled := make(chan struct{}, 1)
	sourceSubCancel, ok := source.XSub(ctx, func(ctx context.Context, value int) bool {
		sourceSubUpdates <- value
		return false
	}, func() {
		close(sourceSubCanceled)
	})
	if !ok {
		t.Fatal("expected source XSub to subscribe")
	}
	if got := expectInt(t, sourceSubUpdates); got != 4 {
		t.Fatal("unexpected initial source XSub value:", got)
	}

	derivedSubUpdates := make(chan string, 2)
	derivedSubCanceled := make(chan struct{}, 1)
	derivedSubCancel, ok := derived.XSub(ctx, func(ctx context.Context, value string) bool {
		derivedSubUpdates <- value
		return false
	}, func() {
		close(derivedSubCanceled)
	})
	if !ok {
		t.Fatal("expected derived XSub to subscribe")
	}
	if got := expectString(t, derivedSubUpdates); got != "v:4" {
		t.Fatal("unexpected initial derived XSub value:", got)
	}

	source.Update(ctx, 5)
	if got := expectInt(t, sourceSubUpdates); got != 5 {
		t.Fatal("unexpected source XSub update:", got)
	}
	if got := expectString(t, derivedSubUpdates); got != "v:5" {
		t.Fatal("unexpected derived XSub update:", got)
	}

	sourceSubCancel()
	expectSignal(t, sourceSubCanceled)
	derivedSubCancel()
	expectSignal(t, derivedSubCanceled)

	noCoreCancel, ok := derived.XSub(context.Background(), func(context.Context, string) bool {
		return false
	}, nil)
	if ok {
		t.Fatal("expected XSub without a Doors context to fail")
	}
	noCoreCancel()
}

func TestBeamReadSubHelpersAndEquality(t *testing.T) {
	none()

	ctx := newBeamContext(t)
	source := NewSourceEqual(1, func(new int, old int) bool {
		return new == old
	})
	source.DisableSkipping()
	source.Mutate(ctx, func(v int) int {
		return v + 1
	})
	if got := source.Get(); got != 2 {
		t.Fatal("unexpected source value after Mutate:", got)
	}

	derived := NewBeamEqual(source, func(v int) int {
		return v % 2
	}, func(new int, old int) bool {
		return new == old
	})

	if got, ok := source.Read(ctx); !ok || got != 2 {
		t.Fatal("unexpected source Read result:", got, ok)
	}
	if got, ok := derived.Read(ctx); !ok || got != 0 {
		t.Fatal("unexpected beam Read result:", got, ok)
	}

	sourceUpdates := make(chan int, 1)
	sourceInitial, ok := source.ReadAndSub(ctx, func(ctx context.Context, value int) bool {
		sourceUpdates <- value
		return true
	})
	if !ok {
		t.Fatal("expected source ReadAndSub to subscribe")
	}
	if sourceInitial != 2 {
		t.Fatal("unexpected source initial value:", sourceInitial)
	}

	beamUpdates := make(chan int, 2)
	if !derived.Sub(ctx, func(ctx context.Context, value int) bool {
		beamUpdates <- value
		return value == 1
	}) {
		t.Fatal("expected beam Sub to subscribe")
	}
	if got := expectInt(t, beamUpdates); got != 0 {
		t.Fatal("unexpected initial beam Sub value:", got)
	}

	if !source.Sub(ctx, func(ctx context.Context, value int) bool {
		return value == 3
	}) {
		t.Fatal("expected source Sub to subscribe")
	}

	source.Update(ctx, 3)
	if got := expectInt(t, sourceUpdates); got != 3 {
		t.Fatal("unexpected source ReadAndSub update:", got)
	}
	if got := expectInt(t, beamUpdates); got != 1 {
		t.Fatal("unexpected beam Sub update:", got)
	}
}
