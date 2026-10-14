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

package doors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front/actions"
	"github.com/doors-dev/doors/internal/instance/utils"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/gox"
)

func linksUntil(h *lifecycleHarness, from int, done func([]linksCall) bool) []linksCall {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		links := h.client.issued()[from:]
		if h.client.idle() && done(links) {
			time.Sleep(50 * time.Millisecond)
			return h.client.issued()[from:]
		}
		if time.Now().After(deadline) {
			h.t.Fatalf("timed out waiting for calls, issued %v", linksNames(links))
		}
		time.Sleep(time.Millisecond)
	}
}

func linksNames(links []linksCall) []string {
	names := make([]string, len(links))
	for i, l := range links {
		names[i] = l.name
	}
	return names
}

func linksExpect(h *lifecycleHarness, from int, want ...string) []linksCall {
	h.t.Helper()
	slices.Sort(want)
	links := linksUntil(h, from, func(links []linksCall) bool {
		return len(links) >= len(want) && !slices.Contains(linksNames(links), "")
	})
	got := linksNames(links)
	slices.Sort(got)
	if !slices.Equal(got, want) {
		h.t.Fatalf("expected calls %v, got %v", want, linksNames(links))
	}
	return links
}

func linksNamed(t *testing.T, links []linksCall, name string, n int) []linksCall {
	t.Helper()
	var found []linksCall
	for _, l := range links {
		if l.name == name {
			found = append(found, l)
		}
	}
	if len(found) != n {
		t.Fatalf("expected %d %s calls, got %d in %v", n, name, len(found), linksNames(links))
	}
	return found
}

func linksScroll(t *testing.T, links []linksCall, selector string) linksCall {
	t.Helper()
	for _, l := range links {
		if a, ok := l.action.(actions.Scroll); ok && a.Selector == selector {
			return l
		}
	}
	t.Fatalf("expected a scroll call to %s in %v", selector, linksNames(links))
	return linksCall{}
}

func linksHook(h *lifecycleHarness, d core.Door, batch bool, track uint64, handler func(ctx context.Context)) {
	h.t.Helper()
	hook, ok := d.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		handler(ctx)
		return true
	}, RaceSerial, batch)
	if !ok {
		h.t.Fatal("expected hook registration to succeed")
	}
	if !h.root.TriggerHook(hook.HookID, httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", nil), track) {
		h.t.Fatal("expected hook trigger to be found")
	}
}

func linksPage(h *lifecycleHarness, doors ...*Door) (context.Context, core.Door, int) {
	h.t.Helper()
	pageCtx := h.renderPage(tagElem("body", "body", func(cur gox.Cursor) error {
		for _, d := range doors {
			if err := cur.Comp(d); err != nil {
				return err
			}
		}
		return nil
	}))
	h.waitIdle()
	return pageCtx, pageCtx.Value(common.KeyCore).(core.Core).Door(), len(h.client.issued())
}

func linksHead(title string, description string, keywords string) gox.Elem {
	return func(cur gox.Cursor) error {
		if err := cur.Init("title"); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		if err := cur.Text(title); err != nil {
			return err
		}
		if err := cur.Close(); err != nil {
			return err
		}
		for name, content := range map[string]string{"description": description, "keywords": keywords} {
			if err := cur.InitVoid("meta"); err != nil {
				return err
			}
			if err := cur.Set("name", name); err != nil {
				return err
			}
			if err := cur.Set("content", content); err != nil {
				return err
			}
			if err := cur.Submit(); err != nil {
				return err
			}
		}
		return nil
	}
}

// OnBatch from a detached page ctx puts one non-zero batch id and no group on its batch_init, its ops' door calls and its batch_flush, and leaves later calls untagged.
func TestLinksOnBatchDetached(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	a, b := &Door{}, &Door{}
	pageCtx, _, from := linksPage(h, a, b)
	detached := DetachedContext(pageCtx)
	OnBatch(detached, func(context.Context) { h.events <- "batch" }, func(ctx context.Context) {
		a.Inner(ctx, textElem("a"))
	}, func(ctx context.Context) {
		b.Inner(ctx, textElem("b"))
	})
	h.waitEvent("batch")
	links := linksExpect(h, from, "batch_init", "door_update", "door_update", "batch_flush")
	batch := links[0].batch
	if links[0].name != "batch_init" || links[3].name != "batch_flush" {
		t.Fatalf("expected batch_init first and batch_flush last, got %v", linksNames(links))
	}
	if batch == 0 {
		t.Fatal("expected a non-zero batch id")
	}
	for _, l := range links {
		if l.batch != batch || l.group != 0 {
			t.Fatalf("expected %s in batch %d without a group, got %+v", l.name, batch, l)
		}
	}
	if links[1].door == links[2].door || links[1].door == links[0].door {
		t.Fatalf("expected each op to target its own door, got %d and %d", links[1].door, links[2].door)
	}
	from += len(links)
	a.Inner(detached, textElem("later"))
	Call(detached, ActionScroll{Selector: "#later"})
	for _, l := range linksExpect(h, from, "door_update", "scroll") {
		if l.batch != 0 || l.group != 0 {
			t.Fatalf("expected %s after the batch untagged, got batch %d", l.name, l.batch)
		}
	}
}

// OnBatch in a plain tracked hook puts the hook's group and its own batch on its ops, batch_init and batch_flush, while the hook's own call and report_hook carry the group only.
func TestLinksOnBatchInPlainHook(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	own, op := &Door{}, &Door{}
	_, page, from := linksPage(h, own, op)
	linksHook(h, page, false, 7, func(ctx context.Context) {
		OnBatch(ctx, func(context.Context) { h.events <- "batch" }, func(ctx context.Context) {
			op.Inner(ctx, textElem("op"))
		})
		own.Inner(ctx, textElem("own"))
	})
	h.waitEvent("batch")
	links := linksExpect(h, from, "batch_init", "door_update", "door_update", "batch_flush", "report hook")
	init := linksNamed(t, links, "batch_init", 1)[0]
	flush := linksNamed(t, links, "batch_flush", 1)[0]
	report := linksNamed(t, links, "report hook", 1)[0]
	updates := linksNamed(t, links, "door_update", 2)
	group, batch := report.group, init.batch
	if group == 0 || batch == 0 || group == batch || report.batch != 0 {
		t.Fatalf("expected a group on the report and a distinct batch on batch_init, got %+v and %+v", report, init)
	}
	for _, l := range links {
		if l.group != group {
			t.Fatalf("expected %s in group %d, got %d", l.name, group, l.group)
		}
	}
	if flush.batch != batch {
		t.Fatalf("expected batch_flush in batch %d, got %d", batch, flush.batch)
	}
	var inBatch, inGroup []linksCall
	for _, l := range updates {
		switch l.batch {
		case batch:
			inBatch = append(inBatch, l)
		case 0:
			inGroup = append(inGroup, l)
		}
	}
	if len(inBatch) != 1 || len(inGroup) != 1 || inBatch[0].door == inGroup[0].door {
		t.Fatalf("expected the op in the batch and the hook's own call outside, got batches %d and %d for doors %d and %d", updates[0].batch, updates[1].batch, updates[0].door, updates[1].door)
	}
	if init.seq > inBatch[0].seq || inBatch[0].seq > flush.seq {
		t.Fatalf("expected the op between batch_init and batch_flush, got %v", linksNames(links))
	}
	if flush.seq > report.seq {
		t.Fatalf("expected batch_flush before report_hook, got %v", linksNames(links))
	}
	if inGroup[0].seq > report.seq {
		t.Fatalf("expected the hook's own call before report_hook, got %v", linksNames(links))
	}
}

// Two OnBatch calls in one plain hook handler open two batches in the hook's one group.
func TestLinksTwoOnBatchInPlainHook(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	a, b := &Door{}, &Door{}
	_, page, from := linksPage(h, a, b)
	linksHook(h, page, false, 7, func(ctx context.Context) {
		OnBatch(ctx, func(context.Context) { h.events <- "first" }, func(ctx context.Context) {
			a.Inner(ctx, textElem("a"))
		})
		OnBatch(ctx, func(context.Context) { h.events <- "second" }, func(ctx context.Context) {
			b.Inner(ctx, textElem("b"))
		})
	})
	h.waitEvent("first")
	h.waitEvent("second")
	links := linksExpect(h, from, "batch_init", "batch_init", "door_update", "door_update", "batch_flush", "batch_flush", "report hook")
	report := linksNamed(t, links, "report hook", 1)[0]
	if report.group == 0 || report.batch != 0 {
		t.Fatalf("expected the report in a group without a batch, got %+v", report)
	}
	members := map[uint64][]string{}
	for _, l := range links {
		if l.group != report.group {
			t.Fatalf("expected %s in group %d, got %d", l.name, report.group, l.group)
		}
		if l.batch != 0 {
			members[l.batch] = append(members[l.batch], l.name)
		}
	}
	if len(members) != 2 {
		t.Fatalf("expected two batches, got %v", members)
	}
	for batch, names := range members {
		if !slices.Equal(names, []string{"batch_init", "door_update", "batch_flush"}) {
			t.Fatalf("expected batch %d to hold its init, op and flush, got %v", batch, names)
		}
	}
}

// OnBatch in a Batch hook joins it: one group and one batch on every call and exactly one batch_init.
func TestLinksOnBatchInBatchHook(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	own, op := &Door{}, &Door{}
	_, page, from := linksPage(h, own, op)
	linksHook(h, page, true, 7, func(ctx context.Context) {
		OnBatch(ctx, func(context.Context) { h.events <- "batch" }, func(ctx context.Context) {
			op.Inner(ctx, textElem("op"))
		})
		own.Inner(ctx, textElem("own"))
	})
	h.waitEvent("batch")
	links := linksExpect(h, from, "batch_init", "door_update", "door_update", "report hook")
	if links[0].name != "batch_init" || links[3].name != "report hook" {
		t.Fatalf("expected batch_init first and report_hook last, got %v", linksNames(links))
	}
	if links[0].batch == 0 || links[0].group == 0 || links[0].group == links[0].batch {
		t.Fatalf("expected distinct non-zero group and batch ids, got %+v", links[0])
	}
	for _, l := range links {
		if l.batch != links[0].batch || l.group != links[0].group {
			t.Fatalf("expected %s in group %d and batch %d, got %+v", l.name, links[0].group, links[0].batch, l)
		}
	}
}

// A user call in a Batch hook is tagged and targets the issuing door with its placement seq; with DetachedContext it is untagged.
func TestLinksUserCallInBatchHook(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	parent, child := &Door{}, &Door{}
	pageCtx, _, from := linksPage(h, parent)
	childCtx := make(chan context.Context, 1)
	child.Inner(context.Background(), gox.Elem(func(cur gox.Cursor) error {
		childCtx <- cur.Context()
		return nil
	}))
	parent.Inner(DetachedContext(pageCtx), mountDoor(child))
	placed := linksExpect(h, from, "door_update")[0]
	from++
	issuer := (<-childCtx).Value(common.KeyCore).(core.Core).Door()
	linksHook(h, issuer, true, 7, func(ctx context.Context) {
		Call(ctx, ActionScroll{Selector: "#member"})
		Call(DetachedContext(ctx), ActionScroll{Selector: "#detached"})
	})
	links := linksExpect(h, from, "batch_init", "scroll", "scroll", "report hook")
	init := linksNamed(t, links, "batch_init", 1)[0]
	group, batch := init.group, init.batch
	member, detached := linksScroll(t, links, "#member"), linksScroll(t, links, "#detached")
	if issuer.ID() == placed.door || issuer.ID() == h.root.ID() || placed.seq == 0 {
		t.Fatalf("expected a door placed by a call, got door %d placed by %+v", issuer.ID(), placed)
	}
	member.action, detached.action = nil, nil
	want := linksCall{seq: member.seq, name: "scroll", group: group, batch: batch, door: issuer.ID(), depSeq: placed.seq}
	if batch == 0 || group == 0 || member != want {
		t.Fatalf("expected the user call linked as %+v, got %+v", want, member)
	}
	want = linksCall{seq: detached.seq, name: "scroll", door: issuer.ID(), depSeq: placed.seq}
	if detached != want {
		t.Fatalf("expected the detached user call linked as %+v, got %+v", want, detached)
	}
}

// Title, each meta name, path and tab state use their own stable target id; title and meta are never tagged, set_path carries the group and batch of its ctx.
func TestLinksStreamTargets(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	d := &Door{}
	pageCtx, page, from := linksPage(h, d)
	tab := utils.NewTabStateManager(h.inst, pageCtx)
	state := utils.DeriveTabState(tab, "n", func(a int, b int) bool { return a == b })
	tab.Initialize(nil)
	round := func(ctx context.Context, n int, title string) {
		d.Inner(ctx, linksHead(title, title+" description", title+" keywords"))
		h.inst.Location().Update(ctx, path.Location{Segments: []string{title}})
		state.Update(ctx, &n)
	}
	count := func(links []linksCall, name string) int {
		n := 0
		for _, l := range links {
			if l.name == name {
				n++
			}
		}
		return n
	}
	linksHook(h, page, true, 7, func(ctx context.Context) {
		round(ctx, 1, "first")
	})
	first := linksUntil(h, from, func(links []linksCall) bool {
		return count(links, "report hook") == 1 && count(links, "set_path") == 1 && count(links, "update_state") == 1 &&
			count(links, "update_title") >= 1 && count(links, "update_meta") >= 2
	})
	batch := linksNamed(t, first, "batch_init", 1)[0].batch
	report := linksNamed(t, first, "report hook", 1)[0]
	setPath := linksNamed(t, first, "set_path", 1)[0]
	if batch == 0 || setPath.batch != batch {
		t.Fatalf("expected set_path in batch %d, got %d", batch, setPath.batch)
	}
	if report.group == 0 || setPath.group != report.group {
		t.Fatalf("expected set_path in group %d, got %d", report.group, setPath.group)
	}
	if setPath.seq > report.seq {
		t.Fatalf("expected set_path before the batch closer, got %v", linksNames(first))
	}
	round(DetachedContext(pageCtx), 2, "second")
	second := linksUntil(h, from+len(first), func(links []linksCall) bool {
		return count(links, "set_path") == 1 && count(links, "update_state") == 1 &&
			count(links, "update_title") >= 1 && count(links, "update_meta") >= 2
	})
	if got := linksNamed(t, second, "set_path", 1)[0]; got.batch != 0 || got.group != 0 {
		t.Fatalf("expected the detached set_path untagged, got batch %d", got.batch)
	}
	targets := map[string]uint64{}
	updates := map[string]int{}
	for _, l := range slices.Concat(first, second) {
		var stream string
		switch a := l.action.(type) {
		case actions.UpdateTitle:
			stream = "title"
		case actions.UpdateMeta:
			stream = "meta " + a.Name
		case actions.RemoveMeta:
			stream = "meta " + a.Name
		case *actions.SetPath:
			stream = "path"
		case actions.UpdateState:
			stream = "state"
		default:
			continue
		}
		if (l.batch != 0 || l.group != 0) && stream != "path" && stream != "state" {
			t.Fatalf("expected %s untagged, got batch %d", stream, l.batch)
		}
		if l.depSeq != 0 || l.depDoor != 0 {
			t.Fatalf("expected %s without deps, got %+v", stream, l)
		}
		if target, ok := targets[stream]; ok && target != l.door {
			t.Fatalf("expected %s to keep target %d, got %d", stream, target, l.door)
		}
		targets[stream] = l.door
		updates[stream]++
	}
	doorTargets := map[uint64]bool{page.ID(): true, h.root.ID(): true}
	for _, l := range slices.Concat(first, second) {
		if l.name == "door_update" {
			doorTargets[l.door] = true
		}
	}
	seen := map[uint64]string{}
	for _, stream := range []string{"title", "meta description", "meta keywords", "path", "state"} {
		target := targets[stream]
		if updates[stream] < 2 {
			t.Fatalf("expected two %s updates, got %d", stream, updates[stream])
		}
		if target == 0 || doorTargets[target] {
			t.Fatalf("expected %s to have its own target, got %d", stream, target)
		}
		if other, ok := seen[target]; ok {
			t.Fatalf("expected %s and %s to have distinct targets, both use %d", stream, other, target)
		}
		seen[target] = stream
	}
}

// A call issued with a ctx infected from another instance's Batch hook is untagged, unguarded and not awaited by that batch, while the same infection inside the instance keeps the batch.
func TestLinksForeignInstanceUntagged(t *testing.T) {
	ha := newLifecycleHarness(t, 8)
	hb := newLifecycleHarness(t, 8)
	hb.inst.linksID = "instance-2"
	da, db := &Door{}, &Door{}
	pageA, page, fromA := linksPage(ha, da)
	pageB, _, fromB := linksPage(hb, db)
	if a, b := pageA.Value(common.KeyCore).(core.Core).InstanceID(), pageB.Value(common.KeyCore).(core.Core).InstanceID(); a == b {
		t.Fatalf("expected two instance ids, got %s twice", a)
	}
	var foreignOp, foreignCall <-chan error
	gate := make(chan struct{})
	open := sync.OnceFunc(func() { close(gate) })
	defer open()
	linksHook(ha, page, true, 7, func(ctx context.Context) {
		foreign := ctex.FrameInfect(ctx, pageB)
		foreignOp = db.Inner(foreign, gox.Elem(func(cur gox.Cursor) error {
			<-gate
			return cur.Text("foreign")
		}))
		foreignCall = Call(foreign, ActionScroll{Selector: "#foreign"})
		same := ctex.FrameInfect(ctx, pageA)
		da.Inner(same, textElem("same"))
		Call(same, ActionScroll{Selector: "#same"})
	})
	linksA := linksExpect(ha, fromA, "batch_init", "door_update", "scroll", "report hook")
	batch, group := linksA[0].batch, linksA[0].group
	if batch == 0 || group == 0 {
		t.Fatal("expected non-zero group and batch ids")
	}
	for _, l := range linksA {
		if l.batch != batch || l.group != group {
			t.Fatalf("expected %s of the issuing instance in batch %d, got %d", l.name, batch, l.batch)
		}
	}
	if got := linksNames(linksExpect(hb, fromB, "scroll")); len(got) != 1 {
		t.Fatalf("expected the foreign door call to wait for its render, got %v", got)
	}
	open()
	for _, l := range linksExpect(hb, fromB, "door_update", "scroll") {
		if l.batch != 0 || l.group != 0 {
			t.Fatalf("expected %s of the foreign instance untagged, got batch %d", l.name, l.batch)
		}
	}
	if vals := drainOp(t, foreignOp); len(vals) != 2 || vals[0] != nil || vals[1] != nil {
		t.Fatalf("expected the foreign op to report scheduled and applied, got %v", vals)
	}
	if vals := drainOp(t, foreignCall); len(vals) != 1 || vals[0] != nil {
		t.Fatalf("expected the foreign call to report its result, got %v", vals)
	}
}

// OnSettle in a plain hook whose door is busy with a Static registers on at once, so an OnBatch in on issues batch_init before report_hook.
func TestLinksOnSettleBatchBeforeReportWithBusyDoor(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	a, b := &Door{}, &Door{}
	_, page, from := linksPage(h, a, b)
	linksHook(h, page, false, 7, func(ctx context.Context) {
		a.Static(ctx, textElem("static"))
		OnSettle(ctx, func(ctx context.Context) {
			OnBatch(ctx, func(context.Context) { h.events <- "batch" }, func(ctx context.Context) {
				b.Inner(ctx, textElem("b"))
			})
		})
	})
	h.waitEvent("batch")
	links := linksExpect(h, from, "door_replace", "report hook", "batch_init", "door_update", "batch_flush")
	init := linksNamed(t, links, "batch_init", 1)[0]
	report := linksNamed(t, links, "report hook", 1)[0]
	if init.seq > report.seq {
		t.Fatalf("expected batch_init before report_hook, got %v", linksNames(links))
	}
}

// OnSettle in OnBatch ops whose door is busy with a Static registers on at once, so a nested OnBatch in on issues batch_init before the outer batch_flush.
func TestLinksOnSettleInBatchOpsWithBusyDoor(t *testing.T) {
	h := newLifecycleHarness(t, 8)
	a, b := &Door{}, &Door{}
	_, page, from := linksPage(h, a, b)
	linksHook(h, page, false, 7, func(ctx context.Context) {
		OnBatch(ctx, func(context.Context) { h.events <- "outer" }, func(ctx context.Context) {
			a.Static(ctx, textElem("static"))
			OnSettle(ctx, func(ctx context.Context) {
				OnBatch(ctx, func(context.Context) { h.events <- "inner" }, func(ctx context.Context) {
					b.Inner(ctx, textElem("b"))
				})
			})
		})
	})
	h.waitEvents("outer", "inner")
	links := linksExpect(h, from, "batch_init", "door_replace", "batch_init", "door_update", "batch_flush", "batch_flush", "report hook")
	inits := linksNamed(t, links, "batch_init", 2)
	outer, inner := inits[0], inits[1]
	for _, flush := range linksNamed(t, links, "batch_flush", 2) {
		if flush.batch == outer.batch && inner.seq > flush.seq {
			t.Fatalf("expected the inner batch_init before the outer batch_flush, got %v", linksNames(links))
		}
	}
}
