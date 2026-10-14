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
	"errors"
	"log/slog"
	"testing"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/gox"
)

func callSeqOf(d *Door) uint64 {
	return d.node.Load().tracker.outer.place.Load()
}

func doorIDOf(d *Door) uint64 {
	return d.node.Load().tracker.outer.id
}

func withDoor(d *Door) gox.Elem {
	return func(cur gox.Cursor) error {
		return cur.Comp(d)
	}
}

func requireLink(t *testing.T, inst *testInstance, want testLink) {
	t.Helper()
	if got := inst.link(want.seq); got != want {
		t.Fatalf("call %d: expected %+v, got %+v", want.seq, want, got)
	}
}

func TestCallSeqSeedsDoorsFromDeliveringCall(t *testing.T) {
	inst := newTestInstance(t)
	ctx := freeCtx()
	p, c, g := &Door{}, &Door{}, &Door{}
	renderTestPage(t, inst, withDoor(p))
	waitOps(t, []<-chan error{p.Inner(ctx, withDoor(c))})
	waitOps(t, []<-chan error{c.Inner(ctx, withDoor(g))})
	waitOps(t, []<-chan error{g.Inner(ctx, text("g"))})
	if seq := callSeqOf(p); seq != 0 {
		t.Fatalf("page door seeded with %d, want 0", seq)
	}
	if seq := callSeqOf(c); seq != 1 {
		t.Fatalf("child seeded with %d, want 1", seq)
	}
	if seq := callSeqOf(g); seq != 2 {
		t.Fatalf("grandchild seeded with %d, want 2", seq)
	}
	waitOps(t, []<-chan error{c.Outer(ctx, text("outer"))})
	if seq := callSeqOf(c); seq != 1 {
		t.Fatalf("outer successor seeded with %d, want 1", seq)
	}
}

func TestCallSeqSeedsDoorsFromStaticCall(t *testing.T) {
	inst := newTestInstance(t)
	ctx := freeCtx()
	tail, next := &Door{}, &Door{}
	renderTestPage(t, inst, withDoor(tail))
	waitOps(t, []<-chan error{tail.Static(ctx, withDoor(next))})
	waitOps(t, []<-chan error{next.Inner(ctx, text("next"))})
	if seq := callSeqOf(next); seq != 1 {
		t.Fatalf("door in static content seeded with %d, want 1", seq)
	}
}

// Every door call targets its own door and depends on the call that placed it.
func TestCallLinkTargetsDoorAndPlacement(t *testing.T) {
	inst := newTestInstance(t)
	ctx := freeCtx()
	p, c := &Door{}, &Door{}
	renderTestPage(t, inst, withDoor(p))
	pid := doorIDOf(p)
	waitOps(t, []<-chan error{p.Inner(ctx, withDoor(c))})
	cid := doorIDOf(c)
	waitOps(t, []<-chan error{c.Inner(ctx, text("a"))})
	waitOps(t, []<-chan error{c.Outer(ctx, text("b"))})
	waitOps(t, []<-chan error{c.Reload(ctx)})
	waitOps(t, []<-chan error{c.Unmount(ctx)})
	requireLink(t, inst, testLink{seq: 1, door: pid})
	requireLink(t, inst, testLink{seq: 2, door: cid, depSeq: 1})
	requireLink(t, inst, testLink{seq: 3, door: cid, depSeq: 1})
	requireLink(t, inst, testLink{seq: 4, door: cid, depSeq: 1})
	requireLink(t, inst, testLink{seq: 5, door: cid, depSeq: 1})
}

// A static call targets the parent scope and depends on the placement and the calls of the replaced door.
func TestCallLinkStatic(t *testing.T) {
	inst := newTestInstance(t)
	ctx := freeCtx()
	p, x, y := &Door{}, &Door{}, &Door{}
	renderTestPage(t, inst, withDoor(p))
	pid := doorIDOf(p)
	waitOps(t, []<-chan error{p.Inner(ctx, withDoor(x))})
	xid := doorIDOf(x)
	waitOps(t, []<-chan error{x.Static(ctx, withDoor(y))})
	yid := doorIDOf(y)
	waitOps(t, []<-chan error{y.Inner(ctx, text("y"))})
	requireLink(t, inst, testLink{seq: 2, door: pid, depSeq: 1, depDoor: xid})
	requireLink(t, inst, testLink{seq: 3, door: yid, depSeq: 2})
}

// A door moved by a re-render gets a new id seeded by the placing call, and the removal targets the old id.
func TestCallLinkMove(t *testing.T) {
	inst := newTestInstance(t)
	ctx := freeCtx()
	a, b, x := &Door{}, &Door{}, &Door{}
	renderTestPage(t, inst, gox.Elem(func(cur gox.Cursor) error {
		if err := cur.Comp(a); err != nil {
			return err
		}
		return cur.Comp(b)
	}))
	waitOps(t, []<-chan error{a.Inner(ctx, withDoor(x))})
	old := doorIDOf(x)
	waitOps(t, []<-chan error{x.Inner(ctx, text("x"))})
	waitOps(t, []<-chan error{b.Inner(ctx, withDoor(x))})
	moved := doorIDOf(x)
	if moved == old {
		t.Fatal("expected the moved door to get a new id")
	}
	waitOps(t, []<-chan error{x.Inner(ctx, text("moved"))})
	requireLink(t, inst, testLink{seq: 2, door: old, depSeq: 1})
	requireLink(t, inst, testLink{seq: 3, door: doorIDOf(b)})
	requireLink(t, inst, testLink{seq: 4, door: old, depSeq: 1})
	requireLink(t, inst, testLink{seq: 5, door: moved, depSeq: 3})
}

// A client error for a call whose content is already gone reports the operation as superseded and logs nothing; for live content it stays an execution error.
func TestCallResultSuperseded(t *testing.T) {
	result := func(ctx context.Context) (error, string) {
		var log bytes.Buffer
		task, ch := newUserTask(context.Background())
		c := &call{ctx: ctx, task: task, payload: emptyPayload{}, logger: slog.New(slog.NewTextHandler(&log, nil))}
		c.Result(nil, errors.New("door 5 not found"))
		return <-ch, log.String()
	}
	gone, cancel := context.WithCancel(context.Background())
	cancel()
	if err, log := result(gone); err != context.Canceled || log != "" {
		t.Fatalf("expected context.Canceled and no log, got %v and %q", err, log)
	}
	if err, log := result(context.Background()); !errors.Is(err, common.ErrExecution) || log == "" {
		t.Fatalf("expected a logged execution error, got %v and %q", err, log)
	}
}
