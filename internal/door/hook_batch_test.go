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
	}, core.RaceSerial, true)
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

// A batch hook opens the batch, tags the calls made with its ctx and closes it with the hook report.
func TestHookBatchClosedByReport(t *testing.T) {
	member, detached, opener, closer, closerCall := runBatchHook(t, 7)
	if opener.batch == 0 {
		t.Fatal("expected a non-zero batch id")
	}
	if member.batch != opener.batch {
		t.Fatalf("expected the handler ctx call in batch %d, got %d", opener.batch, member.batch)
	}
	if member.seq > closer.seq {
		t.Fatalf("expected the batch call before the closer, got %d after %d", member.seq, closer.seq)
	}
	if detached.seq == 0 || detached.batch != 0 {
		t.Fatalf("expected the detached call outside the batch, got %+v", detached)
	}
	if report, ok := closerCall.(reportHook); !ok || report.track != 7 || closer.batch != opener.batch {
		t.Fatalf("expected the hook report to close batch %d, got %#v in batch %d", opener.batch, closerCall, closer.batch)
	}
}

// Without a hook report the batch is closed by an explicit flush.
func TestHookBatchClosedByFlush(t *testing.T) {
	member, _, opener, closer, closerCall := runBatchHook(t, 0)
	if member.batch != opener.batch {
		t.Fatalf("expected the handler ctx call in batch %d, got %d", opener.batch, member.batch)
	}
	if _, ok := closerCall.(batchFlush); !ok || closer.batch != opener.batch {
		t.Fatalf("expected a flush of batch %d, got %#v in batch %d", opener.batch, closerCall, closer.batch)
	}
}

// A hook without the batch option issues no batch calls and leaves its calls untagged.
func TestHookWithoutBatch(t *testing.T) {
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
	}, core.RaceSerial, false)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	inst.root.TriggerHook(hook.HookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), 7)
	waitOps(t, []<-chan error{<-done})
	deadline := time.Now().Add(2 * time.Second)
	for {
		links, calls := inst.issuedCalls()
		if len(calls) == 2 {
			if links[0].batch != 0 {
				t.Fatalf("expected an untagged call, got %+v", links[0])
			}
			if report, ok := calls[1].(reportHook); !ok || report.track != 7 || report.batch != nil || links[1].batch != 0 {
				t.Fatalf("expected a plain hook report, got %#v in batch %d", calls[1], links[1].batch)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected 2 calls, got %d", len(calls))
		}
		time.Sleep(time.Millisecond)
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
		ctex.GetFrames(ctx).BatchFrame(inst.ID()).Run(nil, nil, func(bool) {
			close(reported)
		})
		return true
	}, core.RaceSerial, true)
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
