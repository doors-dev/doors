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

package door

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

type testApp struct {
	conf common.Conf
}

func (a *testApp) Logger() *slog.Logger                 { return testLogger }
func (a *testApp) PathMaker() path.PathMaker            { return path.NewPathMaker("__Host-", "", "") }
func (a *testApp) ResourceRegistry() resources.Registry { return nil }
func (a *testApp) Conf() *common.Conf                   { return &a.conf }
func (a *testApp) Migrating() bool                      { return false }
func (a *testApp) PrinterMiddleware() func(next gox.Printer) gox.Printer {
	return func(next gox.Printer) gox.Printer { return next }
}

type testSession struct {
	app *testApp
	ctx context.Context
}

func (s *testSession) Logger() *slog.Logger     { return testLogger }
func (s *testSession) App() core.App            { return s.app }
func (s *testSession) ID() string               { return "session" }
func (s *testSession) Expire(time.Duration)     {}
func (s *testSession) LastSeen() time.Time      { return time.Time{} }
func (s *testSession) Context() context.Context { return s.ctx }
func (s *testSession) Store() ctex.Store        { return ctex.NewStore() }
func (s *testSession) Kill()                    {}

// testInstance mirrors the production instance around a real Root: calls are
// applied asynchronously in the order they are issued, and ending the
// instance cancels the runtime and pending calls and kills the root.
type testInstance struct {
	session *testSession
	runtime shredder.Runtime
	root    Root
	ids     atomic.Uint64
	mu      sync.Mutex
	ended   bool
	calls   chan actions.Call
	endOnce sync.Once
}

func newTestInstance(t *testing.T) *testInstance {
	t.Helper()
	inst := &testInstance{calls: make(chan actions.Call, 1024)}
	inst.session = &testSession{app: &testApp{}}
	common.InitDefaults(&inst.session.app.conf)
	inst.session.ctx = context.WithValue(context.Background(), common.KeySession, inst.session)
	inst.runtime = shredder.NewRuntime(inst.session.ctx, 8, inst)
	inst.root = NewRoot(inst)
	go inst.client()
	t.Cleanup(inst.Kill)
	return inst
}

func (i *testInstance) client() {
	for c := range i.calls {
		i.mu.Lock()
		ended := i.ended
		i.mu.Unlock()
		if ended {
			c.Cancel()
			continue
		}
		_, free, ok := c.Action()
		if !ok {
			c.Cancel()
			continue
		}
		free()
		c.Result(nil, nil)
	}
}

func (i *testInstance) Call(c actions.Call) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.ended {
		c.Cancel()
		return
	}
	i.calls <- c
}

func (i *testInstance) Kill() {
	i.endOnce.Do(func() {
		i.runtime.Cancel()
		i.mu.Lock()
		i.ended = true
		close(i.calls)
		i.mu.Unlock()
		i.root.Kill()
	})
}

func (i *testInstance) UserCall(context.Context, actions.Action, func(json.RawMessage, error), func(), actions.CallParams) {
}

func (i *testInstance) UserCallCheck(func() bool, actions.Action, func(json.RawMessage, error), func(), actions.CallParams) {
}

func (i *testInstance) Session() core.Session                { return i.session }
func (i *testInstance) Logger() *slog.Logger                 { return testLogger }
func (i *testInstance) Store() ctex.Store                    { return ctex.NewStore() }
func (i *testInstance) CSPCollector() common.CSPCollector    { return (&common.CSP{}).NewCollector() }
func (i *testInstance) ModuleRegistry() core.ModuleRegistry  { return nil }
func (i *testInstance) ID() string                           { return "instance" }
func (i *testInstance) LastSeen() time.Time                  { return time.Time{} }
func (i *testInstance) RootID() uint64                       { return i.root.ID() }
func (i *testInstance) NewID() uint64                        { return i.ids.Add(1) }
func (i *testInstance) Runtime() shredder.Runtime            { return i.runtime }
func (i *testInstance) SetStatus(int)                        {}
func (i *testInstance) Location() beam.Source[path.Location] { return nil }
func (i *testInstance) TitleMeta() core.TitleMeta            { return nil }
func (i *testInstance) TabState(core.TabStateDeriver)        {}

var _ Instance = (*testInstance)(nil)

type discardPrinter struct{}

func (discardPrinter) Send(j gox.Job) error {
	return j.Output(io.Discard)
}

func renderTestPage(t *testing.T, inst *testInstance, content gox.Elem) {
	t.Helper()
	stack, err := inst.root.Render(context.Background(), content)
	if err != nil {
		t.Fatal(err)
	}
	if err := stack.Print(discardPrinter{}); err != nil {
		t.Fatal(err)
	}
}

func waitOps(t *testing.T, chs []<-chan error) {
	t.Helper()
	timeout := time.After(10 * time.Second)
	for _, ch := range chs {
		for done := false; !done; {
			select {
			case err, ok := <-ch:
				if !ok {
					done = true
				} else if err != nil {
					t.Fatalf("operation failed: %v", err)
				}
			case <-timeout:
				t.Fatal("operation channel did not close")
			}
		}
	}
}

func text(s string) gox.Elem {
	return func(cur gox.Cursor) error {
		return cur.Text(s)
	}
}

func (t *tracker) printerCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.printers.Len()
}

// Door.Static renders with the parent's printer, which the parent's content
// tracker drops once the call is applied.
func TestStaticReleasesParentPrinter(t *testing.T) {
	inst := newTestInstance(t)
	doors := make([]*Door, 100)
	for i := range doors {
		doors[i] = &Door{}
		doors[i].Inner(context.Background(), text("x"))
	}
	renderTestPage(t, inst, func(cur gox.Cursor) error {
		for _, d := range doors {
			if err := cur.Comp(d); err != nil {
				return err
			}
		}
		return nil
	})
	var chs []<-chan error
	for _, d := range doors {
		chs = append(chs, d.Static(context.Background(), text("s")))
	}
	waitOps(t, chs)
	if n := inst.root.tracker.printerCount(); n != 0 {
		t.Fatalf("root content tracker retains %d printers", n)
	}
}

// The same for a chain of Door.Static appends, which all render into the
// same parent.
func TestStaticChainReleasesParentPrinter(t *testing.T) {
	inst := newTestInstance(t)
	tail := &Door{}
	renderTestPage(t, inst, func(cur gox.Cursor) error {
		return cur.Comp(tail)
	})
	for range 100 {
		next := &Door{}
		waitOps(t, []<-chan error{tail.Static(context.Background(), func(cur gox.Cursor) error {
			if err := cur.Text("item"); err != nil {
				return err
			}
			return cur.Comp(next)
		})})
		tail = next
	}
	if n := inst.root.tracker.printerCount(); n != 0 {
		t.Fatalf("root content tracker retains %d printers", n)
	}
}
