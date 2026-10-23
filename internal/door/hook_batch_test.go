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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/gox"
)

func runBatchHook(t *testing.T, track uint64) (member testLink, detached testLink, opener testLink, closer testLink, closerCall any) {
	t.Helper()
	inst := newTestInstance(t)
	h, p, q := &Door{}, &Door{}, &Door{}
	renderTestPage(t, inst, gox.Elem(func(cur gox.Cursor) error {
		if err := cur.Comp(h); err != nil {
			return err
		}
		if err := cur.Comp(p); err != nil {
			return err
		}
		return cur.Comp(q)
	}))
	pid, qid := doorIDOf(p), doorIDOf(q)
	hook, ok := h.node.Load().tracker.RegisterHook(func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) bool {
		p.Inner(ctx, text("member"))
		q.Inner(ctex.NewDetachedContext(ctx, ctx, false), text("detached"))
		return true
	}, core.RaceSerial, true, false)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	if !inst.root.TriggerHook(hook.HookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), track) {
		t.Fatal("expected hook trigger to be found")
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		links, calls := inst.issuedCalls()
		if len(calls) == 4 {
			if _, ok := calls[0].(batchInit); !ok {
				t.Fatalf("expected the batch to open first, got %T", calls[0])
			}
			for i, l := range links[1:] {
				switch l.door {
				case pid:
					member = l
				case qid:
					detached = l
				default:
					closer, closerCall = l, calls[i+1]
				}
			}
			return member, detached, links[0], closer, closerCall
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected 4 calls, got %d", len(calls))
		}
		time.Sleep(time.Millisecond)
	}
}

// A batch hook with a report puts group and batch on the opener, the calls made with its ctx and the report that closes it.
func TestHookBatchClosedByReport(t *testing.T) {
	member, detached, opener, closer, closerCall := runBatchHook(t, 7)
	if opener.batch == 0 || opener.group == 0 || opener.group == opener.batch {
		t.Fatalf("expected distinct non-zero group and batch ids, got %+v", opener)
	}
	if member.batch != opener.batch || member.group != opener.group {
		t.Fatalf("expected the handler ctx call in group %d and batch %d, got %+v", opener.group, opener.batch, member)
	}
	if member.seq > closer.seq {
		t.Fatalf("expected the batch call before the closer, got %d after %d", member.seq, closer.seq)
	}
	if detached.seq == 0 || detached.batch != 0 || detached.group != 0 {
		t.Fatalf("expected the detached call outside the group and the batch, got %+v", detached)
	}
	if report, ok := closerCall.(reportHook); !ok || report.track != 7 || report.batch == nil || closer.batch != opener.batch || closer.group != opener.group {
		t.Fatalf("expected the hook report to close batch %d in group %d, got %#v with %+v", opener.batch, opener.group, closerCall, closer)
	}
}

// Without a hook report the batch has no group and is closed by an explicit flush.
func TestHookBatchClosedByFlush(t *testing.T) {
	member, _, opener, closer, closerCall := runBatchHook(t, 0)
	if opener.batch == 0 || opener.group != 0 {
		t.Fatalf("expected a batch without a group, got %+v", opener)
	}
	if member.batch != opener.batch || member.group != 0 {
		t.Fatalf("expected the handler ctx call in batch %d without a group, got %+v", opener.batch, member)
	}
	if _, ok := closerCall.(batchFlush); !ok || closer.batch != opener.batch || closer.group != 0 {
		t.Fatalf("expected a flush of batch %d without a group, got %#v with %+v", opener.batch, closerCall, closer)
	}
}

func runPlainHook(t *testing.T, track uint64) ([]testLink, []any) {
	t.Helper()
	inst := newTestInstance(t)
	h, p := &Door{}, &Door{}
	renderTestPage(t, inst, gox.Elem(func(cur gox.Cursor) error {
		if err := cur.Comp(h); err != nil {
			return err
		}
		return cur.Comp(p)
	}))
	done := make(chan (<-chan error), 1)
	hook, ok := h.node.Load().tracker.RegisterHook(func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) bool {
		done <- p.Inner(ctx, text("plain"))
		return true
	}, core.RaceSerial, false, false)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	inst.root.TriggerHook(hook.HookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), track)
	ch := <-done
	for i, want := range []bool{true, false} {
		select {
		case err, ok := <-ch:
			if err != nil || ok != want {
				t.Fatalf("expected a plain hook op to report scheduled and close, got %v, %v at step %d", err, ok, i)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting on the completion channel")
		}
	}
	want := 1
	if track != 0 {
		want = 2
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		links, calls := inst.issuedCalls()
		if len(calls) == want {
			issued := make([]any, len(calls))
			for i, c := range calls {
				issued[i] = c
			}
			return links, issued
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected %d calls, got %d", want, len(calls))
		}
		time.Sleep(time.Millisecond)
	}
}

// A plain tracked hook puts one group on its calls and its report, opens no batch and reports only scheduling on its channels.
func TestHookGroup(t *testing.T) {
	links, calls := runPlainHook(t, 7)
	if _, ok := calls[0].(batchInit); ok {
		t.Fatal("expected a plain hook not to open a batch")
	}
	report, ok := calls[1].(reportHook)
	if !ok || report.track != 7 || report.batch != nil {
		t.Fatalf("expected a plain hook report without a batch frame, got %#v", calls[1])
	}
	if links[0].group == 0 || links[1].group != links[0].group {
		t.Fatalf("expected the call and the report to share a group, got %d and %d", links[0].group, links[1].group)
	}
	if links[0].batch != 0 || links[1].batch != 0 {
		t.Fatalf("expected no batch id on a plain hook, got %d and %d", links[0].batch, links[1].batch)
	}
}

// A hook without a report issues no closer and puts neither group nor batch on its calls.
func TestHookUntracked(t *testing.T) {
	links, _ := runPlainHook(t, 0)
	if links[0].batch != 0 || links[0].group != 0 {
		t.Fatalf("expected an untagged call, got %+v", links[0])
	}
}

// In a batch hook a door op reports scheduled and closes, a detached one still reports applied,
// and the batch frame opens when the client reports the closer.
func TestHookBatchChannels(t *testing.T) {
	inst := newTestInstance(t)
	h, p, q := &Door{}, &Door{}, &Door{}
	renderTestPage(t, inst, gox.Elem(func(cur gox.Cursor) error {
		if err := cur.Comp(h); err != nil {
			return err
		}
		if err := cur.Comp(p); err != nil {
			return err
		}
		return cur.Comp(q)
	}))
	member := make(chan (<-chan error), 1)
	detached := make(chan (<-chan error), 1)
	reported := make(chan struct{})
	hook, ok := h.node.Load().tracker.RegisterHook(func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) bool {
		member <- p.Inner(ctx, text("member"))
		detached <- q.Inner(ctex.NewDetachedContext(ctx, ctx, false), text("detached"))
		frame, _ := ctex.GetFrames(ctx).BatchFrame()
		frame.Run(nil, nil, func(bool) {
			close(reported)
		})
		return true
	}, core.RaceSerial, true, false)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	inst.root.TriggerHook(hook.HookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), 7)
	receive := func(ch <-chan error) (error, bool) {
		t.Helper()
		select {
		case err, ok := <-ch:
			return err, ok
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting on the completion channel")
			return nil, false
		}
	}
	ch := <-member
	if err, ok := receive(ch); err != nil || !ok {
		t.Fatalf("expected the batch op to report scheduled, got %v, %v", err, ok)
	}
	if _, ok := receive(ch); ok {
		t.Fatal("expected the batch op channel to close without reporting applied")
	}
	ch = <-detached
	for i := range 2 {
		if err, ok := receive(ch); err != nil || !ok {
			t.Fatalf("expected the detached op to report step %d, got %v, %v", i, err, ok)
		}
	}
	select {
	case <-reported:
	case <-time.After(2 * time.Second):
		t.Fatal("expected the batch frame to open when the closer is reported")
	}
}

// A Door of another instance operated with a batch handler ctx gets no ids, and its channel reports only scheduling.
func TestForeignDoorWithBatchCtx(t *testing.T) {
	inst := newTestInstance(t)
	other := newTestInstance(t)
	other.id = "other"
	h, q := &Door{}, &Door{}
	renderTestPage(t, inst, gox.Elem(func(cur gox.Cursor) error {
		return cur.Comp(h)
	}))
	renderTestPage(t, other, gox.Elem(func(cur gox.Cursor) error {
		return cur.Comp(q)
	}))
	_, before := other.issuedCalls()
	foreign := make(chan (<-chan error), 1)
	hook, ok := h.node.Load().tracker.RegisterHook(func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) bool {
		foreign <- q.Inner(ctx, text("foreign"))
		return true
	}, core.RaceSerial, true, false)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	inst.root.TriggerHook(hook.HookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), 7)
	ch := <-foreign
	for i, want := range []bool{true, false} {
		select {
		case err, ok := <-ch:
			if err != nil || ok != want {
				t.Fatalf("expected the foreign op to report scheduled and close, got %v, %v at step %d", err, ok, i)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for step %d of the foreign op", i)
		}
	}
	waitIssued(t, other, len(before)+1)
	links, _ := other.issuedCalls()
	if len(links) != len(before)+1 {
		t.Fatalf("expected one call on the other instance, got %d", len(links)-len(before))
	}
	if l := links[len(links)-1]; l.group != 0 || l.batch != 0 {
		t.Fatalf("expected the foreign door call without ids, got %+v", l)
	}
}

func waitIssued(t *testing.T, inst *testInstance, n int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if links, _ := inst.issuedCalls(); len(links) >= n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected %d issued calls", n)
		}
		time.Sleep(time.Millisecond)
	}
}
