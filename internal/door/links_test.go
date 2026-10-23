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
	"github.com/doors-dev/doors/internal/front/actions"
)

func linksIssued(t *testing.T, inst *testInstance, from int, n int) ([]testLink, []actions.Call) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		links, calls := inst.issuedCalls()
		if len(calls) >= from+n {
			time.Sleep(50 * time.Millisecond)
			links, calls = inst.issuedCalls()
			if len(calls) != from+n {
				t.Fatalf("expected %d calls, got %d", n, len(calls)-from)
			}
			return links[from:], calls[from:]
		}
		if time.Now().After(deadline) {
			t.Fatalf("expected %d calls, got %d", n, len(calls)-from)
		}
		time.Sleep(time.Millisecond)
	}
}

// A user call from a Batch hook carries the hook group and batch, the issuing door and the seq of the call that placed it; a detached one only loses group and batch.
func TestLinksUserCallTargetsIssuingDoor(t *testing.T) {
	inst := newTestInstance(t)
	p, c := &Door{}, &Door{}
	renderTestPage(t, inst, withDoor(p))
	waitOps(t, []<-chan error{p.Inner(freeCtx(), withDoor(c))})
	cid, place := doorIDOf(c), callSeqOf(c)
	if place != 1 || cid == doorIDOf(p) {
		t.Fatalf("expected the child door placed by call 1, got door %d placed by %d", cid, place)
	}
	tracker := c.node.Load().tracker
	hook, ok := tracker.RegisterHook(func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) bool {
		tracker.UserCall(ctx, actions.Scroll{Selector: "#member"}, nil, nil, nil, actions.CallParams{})
		tracker.UserCall(ctex.NewDetachedContext(ctx, ctx, false), actions.Scroll{Selector: "#detached"}, nil, nil, nil, actions.CallParams{})
		return true
	}, core.RaceSerial, true, false)
	if !ok {
		t.Fatal("expected hook registration to succeed")
	}
	if !inst.root.TriggerHook(hook.HookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), 7) {
		t.Fatal("expected hook trigger to be found")
	}
	links, calls := linksIssued(t, inst, 1, 4)
	if _, ok := calls[0].(batchInit); !ok || links[0].batch == 0 || links[0].group == 0 {
		t.Fatalf("expected the batch to open first, got %T with %+v", calls[0], links[0])
	}
	group, batch := links[0].group, links[0].batch
	found := map[string]testLink{}
	for i, call := range calls {
		if user, ok := call.(linksUserCall); ok {
			found[user.action.(actions.Scroll).Selector] = links[i]
		}
	}
	member, detached := found["#member"], found["#detached"]
	if want := (testLink{seq: member.seq, group: group, batch: batch, door: cid, depSeq: place}); member.seq == 0 || member != want {
		t.Fatalf("expected the user call linked as %+v, got %+v", want, member)
	}
	if want := (testLink{seq: detached.seq, door: cid, depSeq: place}); detached.seq == 0 || detached != want {
		t.Fatalf("expected the detached user call linked as %+v, got %+v", want, detached)
	}
	if report, ok := calls[len(calls)-1].(reportHook); !ok || report.track != 7 || member.seq > links[len(links)-1].seq {
		t.Fatalf("expected the hook report to close the batch after the user call, got %#v", calls[len(calls)-1])
	}
}
