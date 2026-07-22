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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/action"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/shredder"
	"github.com/doors-dev/gox"
)

type pipeTestApp struct {
	conf common.Conf
}

func (a *pipeTestApp) Logger() *slog.Logger                 { return slog.Default() }
func (a *pipeTestApp) PathMaker() path.PathMaker            { return path.NewPathMaker("__Host-", "test", "") }
func (a *pipeTestApp) ResourceRegistry() resources.Registry { return nil }
func (a *pipeTestApp) Conf() *common.Conf                   { return &a.conf }
func (a *pipeTestApp) Draining() bool                       { return false }
func (a *pipeTestApp) PrinterMiddleware() func(next gox.Printer) gox.Printer {
	return func(next gox.Printer) gox.Printer { return next }
}

type pipeTestSession struct {
	app *pipeTestApp
}

func (s *pipeTestSession) Logger() *slog.Logger     { return slog.Default() }
func (s *pipeTestSession) App() core.App            { return s.app }
func (s *pipeTestSession) ID() string               { return "session" }
func (s *pipeTestSession) Expire(time.Duration)     {}
func (s *pipeTestSession) LastSeen() time.Time      { return time.Time{} }
func (s *pipeTestSession) Context() context.Context { return context.Background() }
func (s *pipeTestSession) Store() ctex.Store        { return ctex.NewStore() }
func (s *pipeTestSession) Kill()                    {}

type pipeTestInstance struct {
	session *pipeTestSession
	runtime shredder.Runtime
	nextID  atomic.Uint64
}

func (i *pipeTestInstance) Call(action.Call)      {}
func (i *pipeTestInstance) Session() core.Session { return i.session }
func (i *pipeTestInstance) Logger() *slog.Logger  { return slog.Default() }
func (i *pipeTestInstance) Store() ctex.Store     { return ctex.NewStore() }
func (i *pipeTestInstance) UserCall(context.Context, func() bool, action.Action, func(json.RawMessage, error), func(), action.CallParams) {
}
func (i *pipeTestInstance) CSPCollector() common.CSPCollector {
	return (&common.CSP{}).NewCollector()
}
func (i *pipeTestInstance) ModuleRegistry() core.ModuleRegistry  { return nil }
func (i *pipeTestInstance) ID() string                           { return "instance" }
func (i *pipeTestInstance) LastSeen() time.Time                  { return time.Time{} }
func (i *pipeTestInstance) RootID() uint64                       { return 1 }
func (i *pipeTestInstance) NewID() uint64                        { return i.nextID.Add(1) }
func (i *pipeTestInstance) Runtime() shredder.Runtime            { return i.runtime }
func (i *pipeTestInstance) SetStatus(int)                        {}
func (i *pipeTestInstance) Location() beam.Source[path.Location] { return nil }
func (i *pipeTestInstance) Kill()                                {}
func (i *pipeTestInstance) TitleMeta() core.TitleMeta            { return nil }

type pipeTestShutdown struct{}

func (pipeTestShutdown) Kill()                {}
func (pipeTestShutdown) Logger() *slog.Logger { return slog.Default() }

func newRenderPipe(t *testing.T) *pipe {
	t.Helper()
	conf := common.Conf{}
	common.InitDefaults(&conf)
	rt := shredder.NewRuntime(context.Background(), 1, pipeTestShutdown{})
	t.Cleanup(rt.Cancel)
	inst := &pipeTestInstance{runtime: rt}
	inst.session = &pipeTestSession{app: &pipeTestApp{conf: conf}}
	r := NewRoot(inst)
	return newPipe(r.tracker, common.GetDequeBuffer(), nil, nil)
}

func fillFragment(p *pipe) {
	ctx := context.Background()
	p.buffer.PushBack(gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "div", gox.NewAttrs()))
	p.buffer.PushBack(gox.NewJobText(ctx, "fragment"))
	p.buffer.PushBack(gox.NewJobHeadClose(ctx, 1, gox.KindRegular, "div"))
}

type fragmentStamp struct {
	next    gox.Printer
	sendErr error
}

func (s *fragmentStamp) Send(j gox.Job) error {
	if s.sendErr != nil {
		gox.Release(j.(gox.Releaser))
		return s.sendErr
	}
	if open, ok := j.(*gox.JobHeadOpen); ok && open.Kind != gox.KindContainer && open.Attrs != nil {
		open.Attrs.Get("data-stamped").Set("yes")
	}
	return s.next.Send(j)
}

func TestPipeRenderPrinterMiddlewareObservesAndMutates(t *testing.T) {
	p := newRenderPipe(t)
	fillFragment(p)
	payload, err := p.Render(true, func(next gox.Printer) gox.Printer {
		return &fragmentStamp{next: next}
	})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := payload.Payload().Output(&out); err != nil {
		t.Fatal(err)
	}
	payload.Release()
	got := out.String()
	if !strings.Contains(got, `data-stamped="yes"`) {
		t.Fatalf("expected middleware attribute mutation in payload, got %q", got)
	}
	if !strings.Contains(got, "fragment") {
		t.Fatalf("expected fragment content in payload, got %q", got)
	}
}

func TestPipeRenderPrinterMiddlewareSendErrorFailsRender(t *testing.T) {
	sendErr := errors.New("send failed")
	p := newRenderPipe(t)
	fillFragment(p)
	payload, err := p.Render(true, func(next gox.Printer) gox.Printer {
		return &fragmentStamp{next: next, sendErr: sendErr}
	})
	if err != sendErr {
		t.Fatalf("expected send error to fail the render, got %v", err)
	}
	if payload != nil {
		t.Fatal("expected nil payload on send error")
	}
}

func TestPipeRenderWithoutMiddleware(t *testing.T) {
	p := newRenderPipe(t)
	fillFragment(p)
	payload, err := p.Render(true, func(next gox.Printer) gox.Printer { return next })
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := payload.Payload().Output(&out); err != nil {
		t.Fatal(err)
	}
	payload.Release()
	if !strings.Contains(out.String(), "fragment") {
		t.Fatalf("expected plain render to succeed, got %q", out.String())
	}
}
