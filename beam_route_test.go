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
	"errors"
	"fmt"
	"html"
	"maps"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"weak"

	"github.com/doors-dev/doors/internal/beam"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/path"
	"github.com/doors-dev/gox"
)

// rtHarness renders a page and rebuilds the client DOM from the applied door actions.
type rtHarness struct {
	*lifecycleHarness
	rec     *rtRec
	page    context.Context
	initial string
	baseIDs uint64
	doors0  map[string]uint64
	strict  bool
}

func newRT(t *testing.T) *rtHarness {
	t.Helper()
	return newRTN(t, 8)
}

func newRTN(t *testing.T, workers int) *rtHarness {
	t.Helper()
	return &rtHarness{
		lifecycleHarness: newLifecycleHarness(t, workers),
		rec:              newRTRec(),
		strict:           true,
	}
}

// at points the instance location at uri; call it before render.
func (h *rtHarness) at(uri string) Source[Location] {
	h.inst.location = beam.NewSource(rtLoc(uri), path.EqualLocation, false)
	return source[Location]{Source: h.inst.location}
}

func rtSlot(name string, el gox.Elem) gox.Elem {
	return tagElem("div", "rt-"+name, el)
}

var rtSlotRe = regexp.MustCompile(` id="rt-([^"]+)"`)

func (h *rtHarness) render(content ...gox.Elem) {
	h.t.Helper()
	ctx, page, err := h.renderPageHTML(func(cur gox.Cursor) error {
		for _, el := range content {
			if err := el(cur); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		h.t.Fatalf("page render failed: %v", err)
	}
	h.page = DetachedContext(ctx)
	h.initial = page
	h.baseIDs = h.inst.ids.Load()
	h.doors0 = map[string]uint64{}
	for _, m := range rtSlotRe.FindAllStringSubmatch(page, -1) {
		if inner, ok := rtSlotInner(page, m[1]); ok {
			if ids := rtDoorIDs(inner); len(ids) != 0 {
				h.doors0[m[1]] = ids[0]
			}
		}
	}
}

var rtVoid = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true, "img": true,
	"input": true, "link": true, "meta": true, "source": true, "track": true, "wbr": true,
}

func rtIndexTag(s, tag string) int {
	off := 0
	for {
		i := strings.Index(s[off:], "<"+tag)
		if i < 0 {
			return -1
		}
		j := off + i + 1 + len(tag)
		if j < len(s) && (s[j] == ' ' || s[j] == '>') {
			return off + i
		}
		off = j
	}
}

func rtElement(dom, marker string) (start, openEnd, closeStart, end int, ok bool) {
	i := strings.Index(dom, marker)
	if i < 0 {
		return 0, 0, 0, 0, false
	}
	start = strings.LastIndex(dom[:i], "<")
	if start < 0 {
		return 0, 0, 0, 0, false
	}
	n := strings.IndexAny(dom[start+1:], " >")
	if n < 0 {
		return 0, 0, 0, 0, false
	}
	tag := dom[start+1 : start+1+n]
	g := strings.Index(dom[i:], ">")
	if g < 0 {
		return 0, 0, 0, 0, false
	}
	openEnd = i + g + 1
	if rtVoid[tag] {
		return start, openEnd, openEnd, openEnd, true
	}
	closeTag := "</" + tag + ">"
	depth, pos := 1, openEnd
	for {
		c := strings.Index(dom[pos:], closeTag)
		if c < 0 {
			return 0, 0, 0, 0, false
		}
		if o := rtIndexTag(dom[pos:], tag); o >= 0 && o < c {
			depth++
			pos += o + 1 + len(tag)
			continue
		}
		depth--
		if depth == 0 {
			return start, openEnd, pos + c, pos + c + len(closeTag), true
		}
		pos += c + len(closeTag)
	}
}

func rtIsDoorAction(a lifecycleAction) bool {
	return a.name == "door_replace" || a.name == "door_update"
}

func rtDOM(initial string, acts []lifecycleAction) string {
	dom := initial
	for _, a := range acts {
		if !rtIsDoorAction(a) {
			continue
		}
		for _, m := range []string{fmt.Sprintf(` id="d0r%d"`, a.id), fmt.Sprintf(` data-d0r="%d"`, a.id)} {
			if s, oe, cs, e, ok := rtElement(dom, m); ok {
				if a.name == "door_replace" {
					dom = dom[:s] + a.body + dom[e:]
				} else {
					dom = dom[:oe] + a.body + dom[cs:]
				}
				break
			}
		}
	}
	return dom
}

var (
	rtDoorRe = regexp.MustCompile(` id="d0r(\d+)"| data-d0r="(\d+)"`)
	rtTagRe  = regexp.MustCompile(`<[^>]*>`)
)

func rtDoorIDs(s string) []uint64 {
	var ids []uint64
	for _, m := range rtDoorRe.FindAllStringSubmatch(s, -1) {
		v := m[1]
		if v == "" {
			v = m[2]
		}
		n, _ := strconv.ParseUint(v, 10, 64)
		ids = append(ids, n)
	}
	return ids
}

func rtSlotInner(dom, name string) (string, bool) {
	_, oe, cs, _, ok := rtElement(dom, ` id="rt-`+name+`"`)
	if !ok {
		return "", false
	}
	return dom[oe:cs], true
}

func rtText(s string) string {
	return html.UnescapeString(rtTagRe.ReplaceAllString(s, ""))
}

func rtIsBlank(body string, id uint64) bool {
	return strings.HasPrefix(body, "<d0-r ") && strings.HasSuffix(body, "></d0-r>") &&
		strings.Count(body, "<") == 2 && slices.Equal(rtDoorIDs(body), []uint64{id})
}

func (h *rtHarness) dom() string {
	return rtDOM(h.initial, h.client.actions())
}

func (h *rtHarness) slot(name string) (string, bool) {
	return rtSlotInner(h.dom(), name)
}

func (h *rtHarness) html(name string) string {
	h.t.Helper()
	dom := h.dom()
	inner, ok := rtSlotInner(dom, name)
	if !ok {
		h.t.Fatalf("slot %q not found in dom:\n%s", name, dom)
	}
	return inner
}

func (h *rtHarness) door(name string) uint64 {
	h.t.Helper()
	ids := rtDoorIDs(h.html(name))
	if len(ids) == 0 {
		h.t.Fatalf("no door in slot %q, dom:\n%s", name, h.dom())
	}
	return ids[0]
}

func (h *rtHarness) doorCount(name string) int {
	h.t.Helper()
	return len(rtDoorIDs(h.html(name)))
}

func (h *rtHarness) doorActions(id uint64) []lifecycleAction {
	var out []lifecycleAction
	for _, a := range h.client.actions() {
		if rtIsDoorAction(a) && a.id == id {
			out = append(out, a)
		}
	}
	return out
}

// replaces returns the texts of the slot's route door replaces; blank ones must be the bare container.
func (h *rtHarness) replaces(name string) []string {
	h.t.Helper()
	id := h.doors0[name]
	var out []string
	for _, a := range h.doorActions(id) {
		text := rtText(a.body)
		if a.name != "door_replace" || !slices.Equal(rtDoorIDs(a.body), []uint64{id}) || text == "" && !rtIsBlank(a.body, id) {
			h.t.Fatalf("slot %q: want a replace of door %d, got %s %q", name, id, a.name, a.body)
		}
		out = append(out, text)
	}
	return out
}

func (h *rtHarness) settle(op func(ctx context.Context)) {
	h.t.Helper()
	done := make(chan struct{})
	OnSettle(h.page, func(context.Context) { close(done) }, op)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		h.t.Fatalf("settle timed out, dom:\n%s", h.dom())
	}
	h.waitIdle()
}

func rtSet[T any](h *rtHarness, s Source[T], v T) []error {
	h.t.Helper()
	var ch <-chan error
	h.settle(func(ctx context.Context) { ch = s.Update(ctx, v) })
	return drainOp(h.t, ch)
}

func rtMutate[T any](h *rtHarness, s Source[T], f func(T) T) []error {
	h.t.Helper()
	var ch <-chan error
	h.settle(func(ctx context.Context) { ch = s.Mutate(ctx, f) })
	return drainOp(h.t, ch)
}

func rtSameErrs(got, want []error) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if want[i] == nil && got[i] != nil || want[i] != nil && !errors.Is(got[i], want[i]) {
			return false
		}
	}
	return true
}

func (h *rtHarness) eventually(name, want string) {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if h.client.idle() {
			if inner, ok := h.slot(name); ok && rtText(inner) == want {
				return
			}
		}
		if time.Now().After(deadline) {
			inner, ok := h.slot(name)
			h.t.Fatalf("slot %q: want %q, got %q (found=%v), dom:\n%s", name, want, rtText(inner), ok, h.dom())
		}
		time.Sleep(time.Millisecond)
	}
}

func (h *rtHarness) expect(name, want string) {
	h.t.Helper()
	h.expectAt("", name, want)
}

func (h *rtHarness) expectAt(at, name, want string) {
	h.t.Helper()
	dom := h.dom()
	inner, ok := rtSlotInner(dom, name)
	if !ok {
		h.t.Fatalf("%sslot %q not found, dom:\n%s", at, name, dom)
	}
	if got := rtText(inner); got != want {
		h.t.Fatalf("%sslot %q: want %q, got %q, dom:\n%s", at, name, want, got, dom)
	}
}

func (h *rtHarness) expectAll(at string, want map[string]string) {
	h.t.Helper()
	for _, name := range slices.Sorted(maps.Keys(want)) {
		h.expectAt(at, name, want[name])
	}
}

func (h *rtHarness) quiet(name string, step func()) {
	h.t.Helper()
	h.quietAll("", []string{name}, step)
}

// quietAll fails when step sends any action to the route doors of slots.
func (h *rtHarness) quietAll(at string, slots []string, step func()) {
	h.t.Helper()
	before := make([]int, len(slots))
	for i, s := range slots {
		before[i] = len(h.doorActions(h.door(s)))
	}
	step()
	for i, s := range slots {
		if got := h.doorActions(h.door(s)); len(got) != before[i] {
			h.t.Fatalf("%sslot %q: expected no new route door actions, got %+v", at, s, got[before[i]:])
		}
	}
}

func (h *rtHarness) oneDoor(name string) {
	h.t.Helper()
	h.oneDoorAt("", name)
}

// oneDoorAt checks that the slot keeps only its render-time door, which took only replaces, and no new ids.
func (h *rtHarness) oneDoorAt(at, name string) {
	h.t.Helper()
	if n := h.doorCount(name); n != 1 {
		h.t.Fatalf("%sslot %q: want 1 door, got %d, dom:\n%s", at, name, n, h.dom())
	}
	h.routeDoorAt(at, name)
	if ids := h.inst.ids.Load(); ids != h.baseIDs {
		h.t.Fatalf("%sids moved from %d to %d: something created a door, tracker or hook", at, h.baseIDs, ids)
	}
	allowed := map[uint64]bool{}
	for _, id := range h.doors0 {
		allowed[id] = true
	}
	for _, a := range h.client.actions() {
		if rtIsDoorAction(a) && !allowed[a.id] {
			h.t.Fatalf("%saction %s for door %d, which is not a render-time slot door %v", at, a.name, a.id, h.doors0)
		}
	}
}

// routeDoorAt checks that the slot's first door is its render-time door and took only replaces.
func (h *rtHarness) routeDoorAt(at, name string) {
	h.t.Helper()
	want, ok := h.doors0[name]
	if got := h.door(name); !ok || got != want {
		h.t.Fatalf("%sslot %q: door %d, want render-time door %d (found=%v)", at, name, got, want, ok)
	}
	for _, a := range h.doorActions(want) {
		if a.name != "door_replace" {
			h.t.Fatalf("%sslot %q route door %d got %s", at, name, want, a.name)
		}
	}
}

func (h *rtHarness) counts(want map[string]int) {
	h.t.Helper()
	h.countsAt("", want)
}

func (h *rtHarness) countsAt(at string, want map[string]int) {
	h.t.Helper()
	var bad []string
	for _, k := range slices.Sorted(maps.Keys(want)) {
		if got := h.rec.count(k); got != want[k] {
			bad = append(bad, fmt.Sprintf("%s=%d (want %d)", k, got, want[k]))
		}
	}
	if len(bad) != 0 {
		h.t.Fatalf("%scounts: %s\nlog: %v", at, strings.Join(bad, ", "), h.rec.all())
	}
}

// waitCounts waits for counters that settle after the instance ends.
func (h *rtHarness) waitCounts(want map[string]int) {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		ok := true
		for k, n := range want {
			ok = ok && h.rec.count(k) == n
		}
		if ok {
			return
		}
		if time.Now().After(deadline) {
			h.countsAt("eventually: ", want)
		}
		time.Sleep(time.Millisecond)
	}
}

// balanced waits until every render of the labels is cleaned.
func (h *rtHarness) balanced(labels ...string) {
	h.t.Helper()
	want := map[string]int{}
	for _, l := range labels {
		want["clean:"+l] = h.rec.count("render:" + l)
	}
	h.waitCounts(want)
}

// rtStep sets a value: quiet means no route door action, otherwise exactly one replace.
type rtStep[T any] struct {
	set    T
	want   string
	quiet  bool
	ret    []error
	counts map[string]int
	preds  string
}

type (
	rtI = rtStep[int]
	rtL = rtStep[Location]
	rtN = map[string]int
)

func rtRun[T any](h *rtHarness, slot string, src Source[T], initial string, steps ...rtStep[T]) {
	h.t.Helper()
	rtRunSteps(h, slot, src, initial, false, "", steps)
}

// rtRunP is rtRun that also checks the predicate log, preds being the log before the first step.
func rtRunP[T any](h *rtHarness, slot string, src Source[T], initial, preds string, steps ...rtStep[T]) {
	h.t.Helper()
	rtRunSteps(h, slot, src, initial, true, preds, steps)
}

func rtRunSteps[T any](h *rtHarness, slot string, src Source[T], initial string, track bool, preds string, steps []rtStep[T]) {
	h.t.Helper()
	seen := 0
	checkPreds := func(at, want string) {
		h.t.Helper()
		all := h.rec.vals("pred:")
		if got := strings.Join(all[seen:], " "); track && got != want {
			h.t.Fatalf("%spreds %q, want %q", at, got, want)
		}
		seen = len(all)
	}
	h.expectAt("initial: ", slot, initial)
	checkPreds("initial: ", preds)
	if h.strict {
		h.oneDoorAt("initial: ", slot)
	}
	for i, st := range steps {
		at := fmt.Sprintf("step %d (set %v): ", i, st.set)
		id := h.door(slot)
		n := len(h.doorActions(id))
		ret := rtSet(h, src, st.set)
		want := 1
		if st.quiet {
			want = 0
		}
		if got := len(h.doorActions(id)) - n; got != want {
			h.t.Fatalf("%swant %d route door actions, got %d, dom:\n%s", at, want, got, h.dom())
		}
		if st.ret != nil && !rtSameErrs(ret, st.ret) {
			h.t.Fatalf("%supdate returned %v, want %v", at, ret, st.ret)
		}
		h.expectAt(at, slot, st.want)
		checkPreds(at, st.preds)
		if st.counts != nil {
			h.countsAt(at, st.counts)
		}
		if h.strict {
			h.oneDoorAt(at, slot)
		}
	}
}

// rtRec counts and logs labelled events from route content.
type rtRec struct {
	mu  sync.Mutex
	n   map[string]int
	log []string
}

func newRTRec() *rtRec {
	return &rtRec{n: map[string]int{}}
}

func (r *rtRec) hit(key string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.n[key]++
	r.log = append(r.log, key)
}

func (r *rtRec) count(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.n[key]
}

func (r *rtRec) vals(prefix string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []string
	for _, k := range r.log {
		if v, ok := strings.CutPrefix(k, prefix); ok {
			out = append(out, v)
		}
	}
	return out
}

func (r *rtRec) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.log)
}

type rtLeaf struct {
	rec   *rtRec
	label string
}

func (l rtLeaf) Main() gox.Elem {
	return func(cur gox.Cursor) error {
		l.rec.hit("render:" + l.label)
		OnClean(cur.Context(), func() { l.rec.hit("clean:" + l.label) })
		return cur.Text(l.label)
	}
}

func rtComp[T any](r *rtRec, label string) BeamRoute[T] {
	return RouteComp[T](rtLeaf{r, label})
}

func rtBind[T any](r *rtRec, label string) BeamRoute[T] {
	return rtBindIf[T](r, label, nil)
}

// rtBindIf renders nothing for values show rejects.
func rtBindIf[T any](r *rtRec, label string, show func(T) bool) BeamRoute[T] {
	return RouteBind(func(v T) gox.Elem {
		text := label + "=" + fmt.Sprint(v)
		r.hit("render:" + label)
		r.hit("bind:" + text)
		if show != nil && !show(v) {
			return nil
		}
		return func(cur gox.Cursor) error {
			OnClean(cur.Context(), func() { r.hit("clean:" + label) })
			return cur.Text(text)
		}
	})
}

type rtBox[T any] struct {
	mu  sync.Mutex
	all []T
}

func (b *rtBox[T]) put(v T) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.all = append(b.all, v)
}

func (b *rtBox[T]) len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.all)
}

func (b *rtBox[T]) at(i int) T {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.all[i]
}

func (b *rtBox[T]) last() T {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.all[len(b.all)-1]
}

// same reports whether every captured value is the first one.
func (b *rtBox[T]) same() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, v := range b.all {
		if any(v) != any(b.all[0]) {
			return false
		}
	}
	return len(b.all) != 0
}

func rtSource[T any](r *rtRec, label string, box *rtBox[Source[T]]) SourceRoute[T] {
	return RouteSource(func(s Source[T]) gox.Elem {
		box.put(s)
		return rtLeaf{r, label}.Main()
	})
}

func rtBeam[T any](r *rtRec, label string, box *rtBox[Beam[T]]) BeamRoute[T] {
	return RouteBeam(func(b Beam[T]) gox.Elem {
		box.put(b)
		return rtLeaf{r, label}.Main()
	})
}

func rtPred[T any](r *rtRec, label string, p func(T) bool) func(T) bool {
	return func(v T) bool {
		r.hit("pred:" + label + "=" + fmt.Sprint(v))
		return p(v)
	}
}

func rtDerive[T1, T2 any](r *rtRec, label string, f func(Source[T1]) Source[T2]) func(Source[T1]) Source[T2] {
	return func(s Source[T1]) Source[T2] {
		r.hit("derive:" + label)
		return f(s)
	}
}

func rtDeriveBeam[T1, T2 any](r *rtRec, label string, f func(Beam[T1]) Beam[T2]) func(Beam[T1]) Beam[T2] {
	return func(b Beam[T1]) Beam[T2] {
		r.hit("derive:" + label)
		return f(b)
	}
}

func rtBeams[T any](rs []SourceRoute[T]) []BeamRoute[T] {
	out := make([]BeamRoute[T], len(rs))
	for i, r := range rs {
		out[i] = r.(BeamRoute[T])
	}
	return out
}

// The route builders take b: true builds Beam routes, false Source routes.

func rtMatch[T any](b bool, p func(T) bool, rs ...SourceRoute[T]) SourceRoute[T] {
	if b {
		return RouteBeamMatch(p, rtBeams(rs)...)
	}
	return RouteMatch(p, rs...)
}

func rtValue[T any](b bool, v T, rs ...SourceRoute[T]) SourceRoute[T] {
	if b {
		return RouteBeamValue(v, rtBeams(rs)...)
	}
	return RouteValue(v, rs...)
}

func rtModel[M any](b bool, rs ...SourceRoute[M]) SourceRoute[Location] {
	if b {
		return RouteBeamModel[M](rtBeams(rs)...)
	}
	return RouteModel[M](rs...)
}

// rtDeriveTo derives with get and eq; set writes back on the Source side.
func rtDeriveTo[T, U any](b bool, r *rtRec, label string, get func(T) U, set func(T, U) T, eq func(a, b U) bool, rs ...SourceRoute[U]) SourceRoute[T] {
	if b {
		return RouteBeamDerive(rtDeriveBeam(r, label, func(x Beam[T]) Beam[U] { return DeriveBeamEqual(x, get, eq) }), rtBeams(rs)...)
	}
	return RouteDerive(rtDerive(r, label, func(x Source[T]) Source[U] { return DeriveSourceEqual(x, get, set, eq) }), rs...)
}

// rtD derives an int with get, nil meaning identity.
func rtD(b bool, r *rtRec, label string, get func(int) int, rs ...SourceRoute[int]) SourceRoute[int] {
	if get == nil {
		get = rtID
	}
	return rtDeriveTo(b, r, label, get, rtIDSet, rtEq, rs...)
}

// rtOnce is a leaf that renders once per activation.
func rtOnce[T any](b bool, r *rtRec, label string) SourceRoute[T] {
	if b {
		return rtBeam[T](r, label, nil)
	}
	return rtSource[T](r, label, nil)
}

func rtMount[T any](b bool, src Source[T], rs ...SourceRoute[T]) gox.Elem {
	if b {
		return src.RouteBeam(rtBeams(rs)...)
	}
	return src.Route(rs...)
}

// rtEach runs f for the Source routes and for the Beam routes.
func rtEach(t *testing.T, f func(t *testing.T, b bool)) {
	t.Helper()
	for _, b := range []bool{false, true} {
		t.Run(map[bool]string{false: "source", true: "beam"}[b], func(t *testing.T) { f(t, b) })
	}
}

func rtEq[T comparable](a, b T) bool { return a == b }

func rtID(v int) int { return v }

func rtIDSet(_, v int) int { return v }

func rtTens(v int) int { return v / 10 }

func rtTensSet(a, b int) int { return b*10 + a%10 }

func rtTensSource(s Source[int]) Source[int] { return DeriveSource(s, rtTens, rtTensSet) }

func rtTensBeam(b Beam[int]) Beam[int] { return DeriveBeam(b, rtTens) }

func rtEven(v int) bool { return v%2 == 0 }

func rtOdd(v int) bool { return v%2 != 0 }

func rtPos(v int) bool { return v > 0 }

func rtSmall(v int) bool { return v < 10 }

func rtLt(n int) func(int) bool { return func(v int) bool { return v < n } }

func rtGt(n int) func(int) bool { return func(v int) bool { return v > n } }

func rtGe(n int) func(int) bool { return func(v int) bool { return v >= n } }

func rtBit(m int) func(int) bool { return func(v int) bool { return v&m != 0 } }

func rtIn(lo, hi int) func(int) bool { return func(v int) bool { return v >= lo && v < hi } }

type rtBig struct {
	kind int
	data []byte
}

func newRTBig(kind int) *rtBig {
	return &rtBig{kind: kind, data: make([]byte, 1<<20)}
}

func (b *rtBig) String() string {
	return "big" + strconv.Itoa(b.kind)
}

type rtPathA struct {
	A bool `path:"/a"`
}

type rtPathB struct {
	B bool `path:"/b"`
}

type rtPathItem struct {
	Item bool `path:"/item/:ID"`
	ID   int
	Tab  string `query:"tab"`
}

type rtPathName struct {
	Item bool `path:"/item/:Name"`
	Name string
}

type rtPathSettings struct {
	Settings bool `path:"/settings"`
}

type rtPathBad struct {
	X string
}

type rtPathQuery struct {
	Q    bool    `path:"/q"`
	Tag  *string `query:"tag"`
	Page *int    `query:"page"`
}

func (m rtPathQuery) String() string {
	tag, page := "", ""
	if m.Tag != nil {
		tag = *m.Tag
	}
	if m.Page != nil {
		page = strconv.Itoa(*m.Page)
	}
	return tag + "|" + page
}

var rtItem1 = rtPathItem{Item: true, ID: 1}

// rtCustom is a hand-written location type for /item/:ID?tab=...
type rtCustom struct {
	ID, Tab string
}

func (c rtCustom) Encode() (Location, error) {
	if c.ID == "" {
		return Location{}, errors.New("rtCustom: empty ID")
	}
	loc := Location{Segments: []string{"item", c.ID}}
	if c.Tab != "" {
		loc.Query = url.Values{"tab": {c.Tab}}
	}
	return loc, nil
}

func rtCustomGet(l Location) rtCustom {
	if len(l.Segments) != 2 || l.Segments[0] != "item" {
		return rtCustom{}
	}
	return rtCustom{ID: l.Segments[1], Tab: l.Query.Get("tab")}
}

func rtCustomSet(prev Location, c rtCustom) Location {
	next, err := c.Encode()
	if err != nil {
		return prev
	}
	return next
}

func rtLoc(uri string) Location {
	return path.NewLocationFromEscapedURI(uri)
}

func rtLocSource(uri string) Source[Location] {
	return NewSourceEqual(rtLoc(uri), path.EqualLocation)
}

func rtFirst(seg string) func(Location) bool {
	return func(l Location) bool { return len(l.Segments) != 0 && l.Segments[0] == seg }
}

func rtStripGet(l Location) Location {
	l = l.Clone()
	if len(l.Segments) != 0 {
		l.Segments = l.Segments[1:]
	}
	return l
}

// rtStrip routes the location without its first segment; writes put prefix back.
func rtStrip(b bool, r *rtRec, label, prefix string, rs ...SourceRoute[Location]) SourceRoute[Location] {
	set := func(_ Location, l Location) Location {
		l = l.Clone()
		l.Segments = append([]string{prefix}, l.Segments...)
		return l
	}
	return rtDeriveTo(b, r, label, rtStripGet, set, path.EqualLocation, rs...)
}

// rtWantPaths waits for the client to settle and checks its set_path calls.
func rtWantPaths(h *rtHarness, want ...string) {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var got []string
		for _, a := range h.client.actions() {
			if a.name == "set_path" {
				got = append(got, a.body)
			}
		}
		if len(got) >= len(want) && h.client.idle() {
			if !slices.Equal(got, want) {
				h.t.Fatalf("set_path calls %q, want %q", got, want)
			}
			return
		}
		if time.Now().After(deadline) {
			h.t.Fatalf("timed out waiting for set_path calls %q, have %q", want, got)
		}
		time.Sleep(time.Millisecond)
	}
}

func rtSP(p string, replace bool) string {
	return fmt.Sprintf("&{%s %t}", p, replace)
}

func rtLocIs(h *rtHarness, loc Beam[Location], uri string) {
	h.t.Helper()
	if got, want := loc.Get(), rtLoc(uri); !path.EqualLocation(got, want) {
		h.t.Fatalf("location %q, want %q", got.String(), want.String())
	}
}

var (
	_ func(gox.Comp) BeamRoute[int]                                                   = RouteComp[int]
	_ func(func(int) gox.Elem) BeamRoute[int]                                         = RouteBind[int]
	_ func(func(Beam[int]) gox.Elem) BeamRoute[int]                                   = RouteBeam[int]
	_ func(func(int) bool, ...BeamRoute[int]) BeamRoute[int]                          = RouteBeamMatch[int]
	_ func(int, ...BeamRoute[int]) BeamRoute[int]                                     = RouteBeamValue[int]
	_ func(func(Beam[int]) Beam[string], ...BeamRoute[string]) BeamRoute[int]         = RouteBeamDerive[int, string]
	_ func(...BeamRoute[rtPathA]) BeamRoute[Location]                                 = RouteBeamModel[rtPathA]
	_ func(func(Source[int]) gox.Elem) SourceRoute[int]                               = RouteSource[int]
	_ func(func(int) bool, ...SourceRoute[int]) SourceRoute[int]                      = RouteMatch[int]
	_ func(int, ...SourceRoute[int]) SourceRoute[int]                                 = RouteValue[int]
	_ func(func(Source[int]) Source[string], ...SourceRoute[string]) SourceRoute[int] = RouteDerive[int, string]
	_ func(...SourceRoute[rtPathA]) SourceRoute[Location]                             = RouteModel[rtPathA]
	_ func(Source[int], ...SourceRoute[int]) gox.Elem                                 = Source[int].Route
	_ func(Source[int], ...BeamRoute[int]) gox.Elem                                   = Source[int].RouteBeam
	_ func(Beam[int], ...BeamRoute[int]) gox.Elem                                     = Beam[int].RouteBeam
)

// rtCase is a script on one Route call, run with Source and Beam routes; named preds are checked.
type rtCase struct {
	name    string
	start   int
	tree    func(b bool, r *rtRec) []SourceRoute[int]
	initial string
	preds   string
	steps   []rtI
	after   func(h *rtHarness)
}

func rtCases(t *testing.T, cases ...rtCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rtEach(t, func(t *testing.T, b bool) {
				h := newRT(t)
				src := NewSource(c.start)
				h.render(rtSlot("n", rtMount(b, src, c.tree(b, h.rec)...)))
				track := c.preds != ""
				for _, st := range c.steps {
					track = track || st.preds != ""
				}
				if track {
					rtRunP(h, "n", src, c.initial, c.preds, c.steps...)
				} else {
					rtRun(h, "n", src, c.initial, c.steps...)
				}
				if c.after != nil {
					c.after(h)
				}
			})
		})
	}
}

func rtBlanks(n int) func(h *rtHarness) {
	return func(h *rtHarness) {
		h.t.Helper()
		got := h.replaces("n")
		blanks := 0
		for _, s := range got {
			if s == "" {
				blanks++
			}
		}
		if blanks != n {
			h.t.Fatalf("replaces %q: want %d blank", got, n)
		}
	}
}

// The first selection renders in the page: first match wins, no match is blank, a lone element is the container.
func TestRouteInitialRender(t *testing.T) {
	h := newRT(t)
	a, b, c := NewSource(2), NewSource(0), NewSource(1)
	h.render(
		rtSlot("match", a.Route(
			RouteValue(3, rtComp[int](h.rec, "A")),
			RouteMatch(rtPos, rtComp[int](h.rec, "B")),
			RouteMatch(rtGt(1), rtComp[int](h.rec, "C")),
			rtComp[int](h.rec, "D"),
		)),
		rtSlot("none", b.Route(RouteValue(1, rtComp[int](h.rec, "NEVER")))),
		rtSlot("elem", c.Route(
			RouteValue(1, RouteComp[int](tagElem("section", "sec", textElem("S<&>")))),
			rtBind[int](h.rec, "E"),
		)),
	)
	h.expectAll("", map[string]string{"match": "B", "none": "", "elem": "S<&>"})
	if acts := h.client.actions(); len(acts) != 0 {
		t.Fatalf("expected no client actions, got %+v", acts)
	}
	if got := h.html("none"); !rtIsBlank(got, h.doors0["none"]) {
		t.Fatalf("expected a blank route container, got %q", got)
	}
	id := h.doors0["elem"]
	if got := h.html("elem"); !strings.HasPrefix(got, `<section data-d0p="`) || !strings.Contains(got, fmt.Sprintf(` data-d0r="%d"`, id)) {
		t.Fatalf("expected the section to be the container, got %q", got)
	}
	if d := h.doors0; len(d) != 3 || d["match"] == d["none"] || d["none"] == d["elem"] || d["match"] == d["elem"] {
		t.Fatalf("expected three distinct route doors, got %v", d)
	}
	h.oneDoor("match")
	h.oneDoor("none")
	h.counts(rtN{"render:A": 0, "render:B": 1, "render:C": 0, "render:D": 0, "render:NEVER": 0})
	rtRun(h, "elem", c, "S<&>",
		rtI{set: 2, want: "E=2"},
		rtI{set: 3, want: "E=3"},
		rtI{set: 1, want: "S<&>"},
		rtI{set: 4, want: "E=4"},
	)
	acts := h.doorActions(id)
	if b := acts[0].body; !strings.HasPrefix(b, "<d0-r ") || !strings.Contains(b, fmt.Sprintf(` id="d0r%d"`, id)) {
		t.Fatalf("leaving the section should restore <d0-r id=d0r%d>, got %q", id, b)
	}
	if b := acts[2].body; !strings.HasPrefix(b, "<section ") || !strings.Contains(b, fmt.Sprintf(` data-d0r="%d"`, id)) {
		t.Fatalf("coming back should make the section door %d again, got %q", id, b)
	}
}

var rtEntries = []struct {
	name  string
	beam  bool
	mount func(src Source[int], rs ...SourceRoute[int]) gox.Elem
}{
	{"source.Route", false, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return src.Route(rs...)
	}},
	{"source.RouteBeam", true, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return src.RouteBeam(rtBeams(rs)...)
	}},
	{"derivedSource.Route", false, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return DeriveSource(src, rtID, rtIDSet).Route(rs...)
	}},
	{"derivedSource.RouteBeam", true, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return DeriveSource(src, rtID, rtIDSet).RouteBeam(rtBeams(rs)...)
	}},
	{"derivedBeam.RouteBeam", true, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return DeriveBeam(src, rtID).RouteBeam(rtBeams(rs)...)
	}},
}

// On every entry point: matching branches stay, leaves give way, an end re-selects the list, ended route included.
func TestRouteEntryPoints(t *testing.T) {
	for _, b := range []bool{false, true} {
		for _, e := range rtEntries {
			if e.beam && !b {
				continue
			}
			t.Run(fmt.Sprintf("beam=%v/%s", b, e.name), func(t *testing.T) {
				h := newRT(t)
				src := NewSource(1)
				h.render(rtSlot("a", e.mount(src,
					rtMatch(b, rtPred(h.rec, "even", rtEven), rtComp[int](h.rec, "EVEN")),
					rtMatch(b, rtPred(h.rec, "small", rtSmall), rtComp[int](h.rec, "SMALL")),
					rtComp[int](h.rec, "D"),
				)))
				rtRunP(h, "a", src, "SMALL", "even=1 small=1",
					rtI{set: 2, want: "SMALL", quiet: true, ret: []error{nil}, preds: "small=2"},
					rtI{set: 12, want: "EVEN", preds: "small=12 even=12"},
					rtI{set: 3, want: "SMALL", preds: "even=3 even=3 small=3"},
					rtI{set: 13, want: "D", preds: "small=13 even=13 small=13"},
					rtI{set: 11, want: "D", quiet: true, preds: "even=11 small=11"},
					rtI{set: 4, want: "EVEN", preds: "even=4"},
					rtI{set: 4, want: "EVEN", quiet: true, ret: []error{}},
				)
				h.counts(rtN{"render:EVEN": 2, "clean:EVEN": 1, "render:SMALL": 2, "clean:SMALL": 2, "render:D": 1, "clean:D": 1})

				h = newRT(t)
				src = NewSource(0)
				h.render(rtSlot("a", e.mount(src,
					rtValue(b, 1, rtComp[int](h.rec, "A")),
					rtMatch(b, rtPred(h.rec, "b", func(v int) bool { return v == 2 }), rtComp[int](h.rec, "B")),
					rtBind[int](h.rec, "D"),
				)))
				rtRunP(h, "a", src, "D=0", "b=0",
					rtI{set: 3, want: "D=3", preds: "b=3", counts: rtN{"render:D": 2, "clean:D": 1}},
					rtI{set: 2, want: "B", preds: "b=2", counts: rtN{"render:D": 2, "clean:D": 2, "render:B": 1}},
					rtI{set: 0, want: "D=0", preds: "b=0 b=0", counts: rtN{"render:D": 3, "clean:B": 1}},
					rtI{set: 1, want: "A", counts: rtN{"render:D": 3, "clean:D": 3, "render:A": 1, "render:B": 1}},
				)
			})
		}
	}
}

// Routes are tried in order: a leaf shadows the routes after it and gives way to the first earlier match.
func TestRouteSelectionOrder(t *testing.T) {
	h := newRT(t)
	a, b, m := NewSource(0), NewSource(0), NewSource(0)
	h.render(
		rtSlot("a", a.Route(
			RouteValue(1, rtComp[int](h.rec, "A")),
			rtBind[int](h.rec, "D1"),
			RouteValue(2, rtComp[int](h.rec, "NEVER")),
			rtComp[int](h.rec, "D2"),
		)),
		rtSlot("b", b.Route(
			rtComp[int](h.rec, "D0"),
			RouteValue(1, rtComp[int](h.rec, "NEVER0")),
			rtBind[int](h.rec, "NEVERB"),
		)),
		rtSlot("m", m.Route(
			RouteMatch(func(v int) bool { return v == 1 }, rtComp[int](h.rec, "MA")),
			rtComp[int](h.rec, "DM"),
			RouteMatch(rtGe(0), rtComp[int](h.rec, "MB")),
		)),
	)
	rtRun(h, "a", a, "D1=0",
		rtI{set: 1, want: "A"},
		rtI{set: 2, want: "D1=2"},
		rtI{set: 1, want: "A"},
		rtI{set: 3, want: "D1=3", counts: rtN{"render:NEVER": 0, "render:D2": 0, "render:A": 2, "render:D1": 3}},
	)
	rtRun(h, "b", b, "D0",
		rtI{set: 1, want: "D0", quiet: true},
		rtI{set: 2, want: "D0", quiet: true, counts: rtN{"render:D0": 1, "render:NEVER0": 0, "render:NEVERB": 0}},
	)
	rtRun(h, "m", m, "DM",
		rtI{set: 1, want: "MA"},
		rtI{set: 2, want: "DM"},
		rtI{set: 5, want: "DM", quiet: true},
		rtI{set: 1, want: "MA", counts: rtN{"render:MA": 2, "render:DM": 2, "render:MB": 0}},
	)
}

// An intact path stays against earlier siblings; a break re-selects the whole list, the ended route included.
func TestRouteEndReselects(t *testing.T) {
	rtCases(t, rtCase{
		name:  "nested",
		start: 2,
		tree: func(b bool, r *rtRec) []SourceRoute[int] {
			return []SourceRoute[int]{
				rtMatch(b, rtPred(r, "big", rtGe(100)), rtComp[int](r, "BIG")),
				rtMatch(b, rtPred(r, "pos", rtPos),
					rtMatch(b, rtPred(r, "even", rtEven), rtComp[int](r, "EVEN")),
					rtMatch(b, rtPred(r, "small", rtSmall), rtComp[int](r, "SMALL")),
				),
				rtBind[int](r, "D"),
			}
		},
		initial: "EVEN",
		preds:   "big=2 pos=2 even=2",
		steps: []rtI{
			{set: 4, want: "EVEN", quiet: true, preds: "pos=4 even=4"},
			{set: 104, want: "EVEN", quiet: true, preds: "pos=104 even=104"},
			{set: 105, want: "BIG", preds: "pos=105 even=105 big=105"},
			{set: 3, want: "SMALL", preds: "big=3 big=3 pos=3 even=3 small=3"},
			{set: 6, want: "SMALL", quiet: true, preds: "pos=6 small=6"},
			{set: 12, want: "EVEN", preds: "pos=12 small=12 big=12 pos=12 even=12"},
			{set: -1, want: "D=-1", preds: "pos=-1 big=-1 pos=-1"},
			{set: -2, want: "D=-2", preds: "big=-2 pos=-2"},
			{set: 50, want: "EVEN", preds: "big=50 pos=50 even=50", counts: rtN{
				"render:BIG": 1, "clean:BIG": 1, "render:EVEN": 3, "clean:EVEN": 2,
				"render:SMALL": 1, "clean:SMALL": 1, "render:D": 2, "clean:D": 2,
			}},
		},
	})
}

// A group matches only when one of its children does; otherwise selection moves to its next sibling.
func TestRouteFallThrough(t *testing.T) {
	rtCases(t,
		// a branch selected by falling through is sticky like any other
		rtCase{
			name:  "match",
			start: 1,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtPred(r, "outer", rtPos), rtValue(b, 5, rtComp[int](r, "X"))),
					rtMatch(b, rtPred(r, "small", rtSmall), rtComp[int](r, "S")),
					rtComp[int](r, "D"),
				}
			},
			initial: "S",
			preds:   "outer=1 small=1",
			steps: []rtI{
				{set: 5, want: "S", quiet: true, preds: "small=5"},
				{set: 12, want: "D", preds: "small=12 outer=12 small=12"},
				{set: 5, want: "X", preds: "outer=5"},
				{set: 6, want: "S", preds: "outer=6 outer=6 small=6"},
				{set: 15, want: "D", preds: "small=15 outer=15 small=15"},
				{set: 0, want: "S", preds: "outer=0 small=0"},
				{set: -5, want: "S", quiet: true, preds: "small=-5"},
				{set: 5, want: "S", quiet: true, preds: "small=5", counts: rtN{
					"render:X": 1, "clean:X": 1, "render:S": 3, "clean:S": 2, "render:D": 2, "clean:D": 2,
				}},
			},
		},
		// the leaf re-checks a group whose children did not match
		rtCase{
			name:  "yield",
			start: 1,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtPred(r, "e", rtGe(100)), rtComp[int](r, "E")),
					rtMatch(b, rtPred(r, "m", rtPos),
						rtMatch(b, rtPred(r, "five", func(v int) bool { return v%5 == 0 }), rtComp[int](r, "FIVE")),
					),
					rtComp[int](r, "D"),
				}
			},
			initial: "D",
			preds:   "e=1 m=1 five=1",
			steps: []rtI{
				{set: 2, want: "D", quiet: true, preds: "e=2 m=2 five=2"},
				{set: 105, want: "E", preds: "e=105"},
				{set: 3, want: "D", preds: "e=3 e=3 m=3 five=3"},
				{set: 10, want: "FIVE", preds: "e=10 m=10 five=10"},
				{set: 20, want: "FIVE", quiet: true, preds: "m=20 five=20"},
				{set: 21, want: "D", preds: "m=21 five=21 e=21 m=21 five=21"},
				{set: 0, want: "D", quiet: true, preds: "e=0 m=0"},
				{set: 150, want: "E", preds: "e=150", counts: rtN{
					"render:E": 2, "clean:E": 1, "render:FIVE": 1, "clean:FIVE": 1, "render:D": 3, "clean:D": 3,
				}},
			},
		},
		rtCase{
			name:  "derive",
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtD(b, r, "d", rtTens, rtValue(b, 1, rtComp[int](r, "X"))),
					rtComp[int](r, "D"),
				}
			},
			initial: "D",
			steps: []rtI{
				{set: 12, want: "X"},
				{set: 25, want: "D"},
				{set: 13, want: "X", counts: rtN{"derive:d": 1, "render:X": 2, "render:D": 2}},
			},
		},
		// with nothing left to fall to the container goes blank
		rtCase{
			name:  "blank",
			start: 1,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtPred(r, "p", rtPos), rtMatch(b, func(v int) bool { return v%5 == 0 }, rtComp[int](r, "A"))),
					rtMatch(b, rtLt(0), rtComp[int](r, "N")),
				}
			},
			initial: "",
			preds:   "p=1",
			steps: []rtI{
				{set: 2, want: "", quiet: true, preds: "p=2"},
				{set: 5, want: "A", preds: "p=5"},
				{set: 6, want: "", preds: "p=6 p=6"},
				{set: 7, want: "", quiet: true, preds: "p=7"},
				{set: -1, want: "N", preds: "p=-1"},
				{set: 10, want: "A", preds: "p=10"},
				{set: -2, want: "N", preds: "p=-2 p=-2"},
				{set: 3, want: "", preds: "p=3", counts: rtN{"render:A": 2, "clean:A": 2, "render:N": 2, "clean:N": 2}},
			},
			after: rtBlanks(2),
		},
	)
}

// A RouteMatch without children, like an empty route list, never matches.
func TestRouteMatchNoChildren(t *testing.T) {
	h := newRT(t)
	a, b, c, e, f := NewSource(0), NewSource(0), NewSource(-1), NewSource(0), NewSource(0)
	h.render(
		rtSlot("a", a.Route(RouteMatch[int](rtPos), rtComp[int](h.rec, "D"))),
		rtSlot("b", b.RouteBeam(RouteBeamMatch[int](rtPos), RouteBeamMatch(rtGt(5), rtComp[int](h.rec, "LATE")))),
		rtSlot("c", c.Route(RouteMatch[int](rtEven), RouteMatch(rtPos, RouteValue(5, rtComp[int](h.rec, "X"))))),
		rtSlot("e", e.Route()),
		rtSlot("f", f.RouteBeam()),
	)
	rtRun(h, "a", a, "D",
		rtI{set: 1, want: "D", quiet: true},
		rtI{set: -1, want: "D", quiet: true},
		rtI{set: 3, want: "D", quiet: true, counts: rtN{"render:D": 1, "clean:D": 0}},
	)
	rtRun(h, "b", b, "",
		rtI{set: 7, want: "LATE"},
		rtI{set: 9, want: "LATE", quiet: true},
		rtI{set: 0, want: "", counts: rtN{"clean:LATE": 1}},
		rtI{set: -1, want: "", quiet: true},
		rtI{set: 8, want: "LATE", counts: rtN{"render:LATE": 2, "clean:LATE": 1}},
	)
	rtRun(h, "c", c, "",
		rtI{set: 2, want: "", quiet: true},
		rtI{set: 3, want: "", quiet: true},
		rtI{set: 5, want: "X"},
		rtI{set: 7, want: "", counts: rtN{"clean:X": 1}},
		rtI{set: -1, want: "", quiet: true},
		rtI{set: 5, want: "X"},
		rtI{set: 4, want: "", counts: rtN{"render:X": 2, "clean:X": 2}},
	)
	for _, s := range []string{"e", "f"} {
		src := map[string]Source[int]{"e": e, "f": f}[s]
		rtRun(h, s, src, "",
			rtI{set: 1, want: "", quiet: true, ret: []error{nil}},
			rtI{set: 1, want: "", quiet: true, ret: []error{}},
		)
	}
	if got := h.replaces("c"); !slices.Equal(got, []string{"X", "", "X", ""}) {
		t.Fatalf("slot c replaces %q", got)
	}
}

// RouteComp(nil) matches like any leaf, so its group claims a range of values and renders nothing.
func TestRouteNilClaims(t *testing.T) {
	rtCases(t, rtCase{
		name:  "claims",
		start: 1,
		tree: func(b bool, r *rtRec) []SourceRoute[int] {
			return []SourceRoute[int]{
				rtMatch(b, rtPred(r, "even", rtEven), rtComp[int](r, "EVEN")),
				rtMatch(b, rtPred(r, "pos", rtPos), RouteComp[int](nil)),
				rtMatch(b, rtPred(r, "late", rtGt(5)), rtComp[int](r, "LATE")),
				rtBind[int](r, "D"),
			}
		},
		initial: "",
		preds:   "even=1 pos=1",
		steps: []rtI{
			{set: 2, want: "", quiet: true, preds: "pos=2"},
			{set: 7, want: "", quiet: true, preds: "pos=7"},
			{set: 0, want: "EVEN", preds: "pos=0 even=0"},
			{set: 3, want: "", preds: "even=3 even=3 pos=3"},
			{set: 9, want: "", quiet: true, preds: "pos=9"},
			{set: -1, want: "D=-1", preds: "pos=-1 even=-1 pos=-1 late=-1"},
			{set: 5, want: "", preds: "even=5 pos=5"},
			{set: -3, want: "D=-3", preds: "pos=-3 even=-3 pos=-3 late=-3"},
			{set: -5, want: "D=-5", preds: "even=-5 pos=-5 late=-5", counts: rtN{
				"render:EVEN": 1, "clean:EVEN": 1, "render:LATE": 0, "render:D": 3, "clean:D": 2,
			}},
		},
	})
}

// Blank to blank sends nothing, content to blank or back one replace; a leaf rendering nothing is content.
func TestRouteBlankContainer(t *testing.T) {
	h := newRT(t)
	top, nl, def, nb := NewSource(0), NewSource(1), NewSource(0), NewSource(0)
	h.render(
		rtSlot("top", top.Route(RouteValue(1, rtComp[int](h.rec, "ONE")))),
		rtSlot("nil", nl.Route(
			RouteMatch(rtIn(100, 200), RouteValue(150, rtComp[int](h.rec, "X"))),
			RouteValue(1, rtComp[int](h.rec, "A")),
			RouteValue(4, RouteComp[int](nil)),
			RouteDerive(rtTensSource, RouteValue(5, rtComp[int](h.rec, "T"))),
		)),
		rtSlot("def", def.RouteBeam(RouteBeamValue(1, rtComp[int](h.rec, "B")), RouteComp[int](nil))),
		rtSlot("nb", nb.Route(RouteValue(1, rtComp[int](h.rec, "C")), RouteBind(func(int) gox.Elem { h.rec.hit("render:N"); return nil }))),
	)
	for _, s := range []string{"top", "def"} {
		if got := h.html(s); !rtIsBlank(got, h.doors0[s]) {
			t.Fatalf("slot %q: expected a blank container, got %q", s, got)
		}
	}
	rtRun(h, "top", top, "",
		rtI{set: 2, want: "", quiet: true, ret: []error{nil}, counts: rtN{"render:ONE": 0}},
		rtI{set: 1, want: "ONE"},
		rtI{set: 3, want: "", counts: rtN{"render:ONE": 1, "clean:ONE": 1}},
		rtI{set: 1, want: "ONE"},
	)
	rtRun(h, "nil", nl, "A",
		rtI{set: 150, want: "X"},
		rtI{set: 160, want: ""},
		rtI{set: 170, want: "", quiet: true},
		rtI{set: 4, want: "", quiet: true},
		rtI{set: 400, want: "", quiet: true},
		rtI{set: 55, want: "T"},
		rtI{set: 57, want: "T", quiet: true},
		rtI{set: 4, want: ""},
		rtI{set: 1, want: "A"},
	)
	rtRun(h, "def", def, "",
		rtI{set: 2, want: "", quiet: true},
		rtI{set: 1, want: "B"},
		rtI{set: 0, want: ""},
		rtI{set: 1, want: "B", counts: rtN{"render:B": 2, "clean:B": 1}},
	)
	rtRun(h, "nb", nb, "",
		rtI{set: 2, want: "", counts: rtN{"render:N": 2}},
		rtI{set: 1, want: "C"},
		rtI{set: 3, want: "", counts: rtN{"render:N": 3, "clean:C": 1}},
	)
	for slot, want := range map[string][]string{
		"top": {"ONE", "", "ONE"},
		"nil": {"X", "", "T", "", "A"},
		"def": {"B", "", "B"},
	} {
		if got := h.replaces(slot); !slices.Equal(got, want) {
			t.Fatalf("slot %q replaces %q, want %q", slot, got, want)
		}
	}
}

// A derive showing a leaf is a leaf for its parent, at any index and through derives; a match is not.
func TestRouteDeriveDefault(t *testing.T) {
	rtCases(t,
		// sticky again once it shows a non-leaf child
		rtCase{
			name:  "index1",
			start: 20,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtValue(b, 5, rtComp[int](r, "P")),
					rtD(b, r, "d", nil, rtMatch(b, rtIn(1, 10), rtComp[int](r, "X")), rtBind[int](r, "Y")),
				}
			},
			initial: "Y=20",
			steps: []rtI{
				{set: 30, want: "Y=30"},
				{set: 5, want: "P", counts: rtN{"render:Y": 2, "clean:Y": 2}},
				{set: 3, want: "X"},
				{set: 5, want: "X", quiet: true, counts: rtN{"render:P": 1}},
				{set: 40, want: "Y=40"},
				{set: 5, want: "P", counts: rtN{"derive:d": 1, "render:P": 2, "render:X": 1, "clean:X": 1, "render:Y": 3}},
			},
		},
		rtCase{
			name:  "chain",
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtValue(b, 5, rtComp[int](r, "P1")),
					rtD(b, r, "x", nil, rtValue(b, 6, rtComp[int](r, "P2")), rtD(b, r, "y", nil, rtBind[int](r, "DEF"))),
				}
			},
			initial: "DEF=0",
			steps: []rtI{
				{set: 5, want: "P1"},
				{set: 6, want: "P2"},
				{set: 7, want: "DEF=7"},
				{set: 5, want: "P1", counts: rtN{"render:P1": 2, "render:P2": 1, "render:DEF": 2, "derive:x": 1, "derive:y": 1}},
			},
		},
		rtCase{
			name:  "chain-sticky",
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtValue(b, 5, rtComp[int](r, "P1")),
					rtD(b, r, "x", nil, rtD(b, r, "y", nil, rtMatch(b, rtIn(5, 7), rtComp[int](r, "P2")), rtBind[int](r, "DEF"))),
				}
			},
			initial: "DEF=0",
			steps: []rtI{
				{set: 5, want: "P1"},
				{set: 6, want: "P2"},
				{set: 5, want: "P2", quiet: true, counts: rtN{"render:P1": 1}},
				{set: 7, want: "DEF=7"},
				{set: 5, want: "P1", counts: rtN{"render:P1": 2, "render:P2": 1, "render:DEF": 2, "derive:x": 1, "derive:y": 1}},
			},
		},
		// only the match's own level re-checks
		rtCase{
			name:  "under-match",
			start: 1,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtValue(b, 5, rtComp[int](r, "P1")),
					rtMatch(b, rtPos,
						rtValue(b, 6, rtComp[int](r, "P2")),
						rtD(b, r, "x", nil, rtD(b, r, "y", nil, rtBind[int](r, "DEF"))),
					),
				}
			},
			initial: "DEF=1",
			steps: []rtI{
				{set: 5, want: "DEF=5"},
				{set: 6, want: "P2"},
				{set: 7, want: "DEF=7"},
				{set: 0, want: ""},
				{set: 5, want: "P1", counts: rtN{"render:P1": 1, "render:P2": 1, "render:DEF": 3}},
			},
		},
	)
}

// A leaf gives way locally and top-down, and the takeover clears the leaf state of the derives above.
func TestRouteLocalYield(t *testing.T) {
	rtCases(t,
		rtCase{
			name:  "under-match",
			start: 2,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtBit(1), rtComp[int](r, "TOPA")),
					rtMatch(b, rtBit(2),
						rtMatch(b, rtBit(8), rtComp[int](r, "INC")),
						rtMatch(b, rtBit(4), rtComp[int](r, "INB")),
						rtBind[int](r, "D"),
					),
				}
			},
			initial: "D=2",
			steps: []rtI{
				{set: 3, want: "D=3"},
				{set: 7, want: "INB"},
				{set: 15, want: "INB", quiet: true},
				{set: 14, want: "INB", quiet: true},
				{set: 10, want: "INC"},
				{set: 3, want: "TOPA", counts: rtN{"render:TOPA": 1, "render:D": 2, "render:INB": 1, "render:INC": 1}},
			},
		},
		rtCase{
			name:  "top-down",
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtBit(1), rtComp[int](r, "TOPA")),
					rtD(b, r, "d", nil, rtMatch(b, rtBit(4), rtComp[int](r, "INB")), rtBind[int](r, "D")),
				}
			},
			initial: "D=0",
			steps: []rtI{
				{set: 4, want: "INB"},
				{set: 5, want: "INB", quiet: true},
				{set: 1, want: "TOPA"},
				{set: 0, want: "D=0"},
				{set: 5, want: "TOPA", counts: rtN{"render:TOPA": 2, "render:INB": 1, "render:D": 2, "derive:d": 1}},
			},
		},
		rtCase{
			name:  "two-derives",
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtBit(1), rtComp[int](r, "TOPA")),
					rtD(b, r, "d1", nil,
						rtMatch(b, rtBit(8), rtComp[int](r, "MIDC")),
						rtD(b, r, "d2", nil, rtMatch(b, rtBit(4), rtComp[int](r, "INB")), rtBind[int](r, "D")),
					),
				}
			},
			initial: "D=0",
			steps: []rtI{
				{set: 4, want: "INB"},
				{set: 13, want: "INB", quiet: true},
				{set: 12, want: "INB", quiet: true},
				{set: 8, want: "MIDC"},
				{set: 0, want: "D=0"},
				{set: 8, want: "MIDC"},
				{set: 9, want: "MIDC", quiet: true},
				{set: 1, want: "TOPA"},
				{set: 0, want: "D=0", counts: rtN{
					"render:D": 3, "clean:D": 2, "render:INB": 1, "render:MIDC": 2, "render:TOPA": 1, "derive:d1": 1, "derive:d2": 1,
				}},
			},
		},
	)
}

// No late re-check: when O ends, the re-selection goes straight to X and D never renders.
func TestRouteNoLateRecheck(t *testing.T) {
	rtCases(t, rtCase{
		name:  "derive",
		start: 1,
		tree: func(b bool, r *rtRec) []SourceRoute[int] {
			return []SourceRoute[int]{
				rtMatch(b, rtEven, rtComp[int](r, "X")),
				rtD(b, r, "d", nil, rtMatch(b, rtOdd, rtComp[int](r, "O")), rtBind[int](r, "D")),
			}
		},
		initial: "O",
		steps: []rtI{
			{set: 2, want: "X", counts: rtN{"render:D": 0}},
			{set: 3, want: "O"},
			{set: 5, want: "O", quiet: true},
			{set: 4, want: "X", counts: rtN{"render:X": 2, "render:O": 2, "render:D": 0, "derive:d": 1}},
		},
	})
}

// Deep paths stay against earlier siblings at every level; a deep break re-selects from the top.
func TestRouteDeepPaths(t *testing.T) {
	rtCases(t,
		// DD, a leaf in the derive, gives way to MIDV and DEEP, but TOPV never beats the sticky outer match
		rtCase{
			name:  "sticky",
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtValue(b, 50, rtComp[int](r, "TOPV")),
					rtMatch(b, rtPos,
						rtValue(b, 60, rtComp[int](r, "MIDV")),
						rtD(b, r, "d", nil,
							rtMatch(b, rtGt(10), rtMatch(b, rtLt(100), rtComp[int](r, "DEEP"))),
							rtComp[int](r, "DD"),
						),
					),
					rtComp[int](r, "D"),
				}
			},
			initial: "D",
			steps: []rtI{
				{set: 20, want: "DEEP"},
				{set: 50, want: "DEEP", quiet: true},
				{set: 60, want: "DEEP", quiet: true},
				{set: 150, want: "DD"},
				{set: 50, want: "DEEP"},
				{set: 150, want: "DD"},
				{set: 60, want: "MIDV"},
				{set: 5, want: "DD"},
				{set: 20, want: "DEEP"},
				{set: 60, want: "DEEP", quiet: true},
				{set: 0, want: "D"},
				{set: 50, want: "TOPV"},
				{set: 60, want: "MIDV", counts: rtN{
					"render:TOPV": 1, "render:MIDV": 2, "render:DEEP": 3, "render:DD": 3, "render:D": 2, "derive:d": 1,
				}},
			},
		},
		// bits p=1 q=2 r=4 s=8: the earlier A beats the still matching q-match, which falls through when only q holds
		rtCase{
			name:  "end",
			start: 6,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtBit(1), rtComp[int](r, "A")),
					rtMatch(b, rtBit(2), rtMatch(b, rtBit(4), rtComp[int](r, "X")), rtMatch(b, rtBit(8), rtComp[int](r, "Y"))),
				}
			},
			initial: "X",
			steps: []rtI{
				{set: 11, want: "A"},
				{set: 6, want: "X"},
				{set: 3, want: "A"},
				{set: 6, want: "X"},
				{set: 10, want: "Y"},
				{set: 6, want: "X"},
				{set: 2, want: ""},
				{set: 3, want: "A"},
				{set: 7, want: "A", quiet: true},
				{set: 14, want: "X", counts: rtN{"render:A": 3, "render:X": 5, "clean:X": 4, "render:Y": 1, "clean:Y": 1}},
			},
			after: rtBlanks(1),
		},
	)
}

// Re-entering a group seeks it afresh: the old sticky child is not restored and a bind renders again.
func TestRouteReactivation(t *testing.T) {
	rtCases(t,
		rtCase{
			name:  "subtree",
			start: 5,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtIn(1, 100), rtMatch(b, rtLt(50), rtBind[int](r, "a")), rtMatch(b, rtGe(40), rtComp[int](r, "b"))),
					rtComp[int](r, "D"),
				}
			},
			initial: "a=5",
			steps: []rtI{
				{set: 45, want: "a=45"},
				{set: 60, want: "b"},
				{set: 45, want: "b", quiet: true},
				{set: 200, want: "D", counts: rtN{"clean:b": 1, "clean:a": 2}},
				{set: 45, want: "a=45", counts: rtN{"render:a": 3}},
				{set: 5, want: "a=5"},
				{set: 200, want: "D"},
				{set: 70, want: "b", counts: rtN{"render:a": 4, "clean:a": 4, "render:b": 2, "clean:b": 1, "render:D": 2, "clean:D": 2}},
			},
		},
	)
}

// RouteValue uses the routed source's comparator as (route value, value): ==, custom, nil never matches.
func TestRouteValueComparator(t *testing.T) {
	h := newRT(t)
	r := h.rec
	rise := func(old, new int) bool { return new <= old }
	mod := func(a, b int) bool { return a%10 == b%10 }
	parity := func(a, b int) bool { return a%2 == b%2 }
	n, s, fold, loc := NewSource(0), NewSource("a"), NewSourceEqual("abc", strings.EqualFold), rtLocSource("/b")
	ds, db, dep := NewSource(0), NewSource(0), NewSource(0)
	nl, nds, ndb, none := NewSourceEqual(1, nil), NewSource(1), NewSource(1), NewSourceEqual(1, nil)
	rs, rb := NewSourceEqual(0, rise), NewSourceEqual(0, rise)
	h.render(
		rtSlot("int", n.Route(RouteValue(1, rtComp[int](r, "ONE")), RouteValue(2, rtComp[int](r, "TWO")))),
		rtSlot("str", s.RouteBeam(RouteBeamValue("x", rtComp[string](r, "X")), RouteBeamValue("y", rtComp[string](r, "Y")), rtBind[string](r, "S"))),
		rtSlot("fold", fold.Route(RouteValue("ABC", rtComp[string](r, "FX")), rtBind[string](r, "F"))),
		rtSlot("loc", loc.Route(RouteValue(rtLoc("/a?x=1"), rtComp[Location](r, "LX")), rtBind[Location](r, "L"))),
		rtSlot("ds", ds.Route(RouteDerive(func(s Source[int]) Source[int] { return DeriveSourceEqual(s, rtID, rtIDSet, mod) },
			RouteValue(3, rtComp[int](r, "DX")), rtBind[int](r, "DB")))),
		rtSlot("db", db.RouteBeam(RouteBeamDerive(func(b Beam[int]) Beam[int] { return DeriveBeamEqual(b, rtID, mod) },
			RouteBeamValue(3, rtComp[int](r, "BX")), rtBind[int](r, "BB")))),
		rtSlot("nil", nl.Route(RouteValue(1, rtComp[int](r, "X1")), rtBind[int](r, "B1"))),
		rtSlot("nds", nds.Route(RouteDerive(func(s Source[int]) Source[int] { return DeriveSourceEqual(s, rtID, rtIDSet, nil) },
			RouteValue(1, rtComp[int](r, "X2")), rtBind[int](r, "B2")))),
		rtSlot("ndb", ndb.RouteBeam(RouteBeamDerive(func(b Beam[int]) Beam[int] { return DeriveBeamEqual(b, rtID, nil) },
			RouteBeamValue(1, rtComp[int](r, "X3")), rtBind[int](r, "B3")))),
		rtSlot("none", none.Route(RouteValue(1, rtComp[int](r, "X4")))),
		rtSlot("rs", rs.Route(RouteValue(5, rtComp[int](r, "UPTO5")), rtComp[int](r, "RD"))),
		rtSlot("rb", rb.RouteBeam(RouteBeamValue(5, rtComp[int](r, "UPTO5B")), rtComp[int](r, "RDB"))),
		rtSlot("dep", dep.Route(
			RouteMatch(rtLt(100), RouteDerive(func(s Source[int]) Source[int] { return DeriveSourceEqual(s, rtTens, rtTensSet, nil) },
				RouteValue(1, rtComp[int](r, "NILV")),
				RouteMatch(rtGe(5), RouteBeamDerive(func(b Beam[int]) Beam[int] { return DeriveBeamEqual(b, rtID, parity) },
					RouteBeamValue(6, rtComp[int](r, "EVENV")),
					rtBind[int](r, "PB"),
				)),
				rtBind[int](r, "NB"),
			)),
			rtComp[int](r, "DEP"),
		)),
	)
	rtRun(h, "int", n, "",
		rtI{set: 1, want: "ONE"},
		rtI{set: 2, want: "TWO"},
		rtI{set: 3, want: ""},
		rtI{set: -2, want: "", quiet: true},
		rtI{set: 2, want: "TWO", counts: rtN{"render:ONE": 1, "render:TWO": 2, "clean:TWO": 1}},
	)
	rtRun(h, "str", s, "S=a",
		rtStep[string]{set: "x", want: "X"},
		rtStep[string]{set: "y", want: "Y"},
		rtStep[string]{set: "X", want: "S=X"},
		rtStep[string]{set: "z", want: "S=z"},
		rtStep[string]{set: "x", want: "X", counts: rtN{"render:X": 2, "render:Y": 1, "render:S": 3}},
	)
	rtRun(h, "fold", fold, "FX",
		rtStep[string]{set: "ABC", want: "FX", quiet: true, ret: []error{}},
		rtStep[string]{set: "abd", want: "F=abd", ret: []error{nil}},
		rtStep[string]{set: "aBc", want: "FX", counts: rtN{"render:FX": 2, "render:F": 1}},
	)
	rtRun(h, "loc", loc, "L=/b",
		rtL{set: rtLoc("/a?x=1"), want: "LX"},
		rtL{set: rtLoc("/a?x=1"), want: "LX", quiet: true, ret: []error{}},
		rtL{set: rtLoc("/a?x=2"), want: "L=/a?x=2"},
		rtL{set: rtLoc("/a?x=1"), want: "LX", counts: rtN{"render:LX": 2, "render:L": 2}},
	)
	for _, c := range []struct{ slot, x, b string }{{"ds", "DX", "DB"}, {"db", "BX", "BB"}} {
		rtRun(h, c.slot, map[string]Source[int]{"ds": ds, "db": db}[c.slot], c.b+"=0",
			rtI{set: 13, want: c.x},
			rtI{set: 23, want: c.x, quiet: true},
			rtI{set: 4, want: c.b + "=4"},
			rtI{set: 3, want: c.x, counts: rtN{"render:" + c.x: 2, "render:" + c.b: 2}},
		)
	}
	rtRun(h, "nil", nl, "B1=1",
		rtI{set: 1, want: "B1=1", ret: []error{nil}, counts: rtN{"render:B1": 2}},
		rtI{set: 2, want: "B1=2"},
		rtI{set: 1, want: "B1=1", counts: rtN{"render:X1": 0, "render:B1": 4}},
	)
	for _, c := range []struct{ slot, x, b string }{{"nds", "X2", "B2"}, {"ndb", "X3", "B3"}} {
		rtRun(h, c.slot, map[string]Source[int]{"nds": nds, "ndb": ndb}[c.slot], c.b+"=1",
			rtI{set: 2, want: c.b + "=2"},
			rtI{set: 1, want: c.b + "=1", counts: rtN{"render:" + c.x: 0}},
		)
	}
	rtRun(h, "none", none, "",
		rtI{set: 1, want: "", quiet: true, ret: []error{nil}},
		rtI{set: 2, want: "", quiet: true, counts: rtN{"render:X4": 0}},
	)
	rtRun(h, "rs", rs, "UPTO5", rtI{set: 5, want: "UPTO5", quiet: true}, rtI{set: 6, want: "RD"})
	rtRun(h, "rb", rb, "UPTO5B", rtI{set: 5, want: "UPTO5B", quiet: true}, rtI{set: 6, want: "RDB"})
	rtRun(h, "dep", dep, "NB=0",
		rtI{set: 15, want: "NB=1"},
		rtI{set: 16, want: "NB=1", counts: rtN{"render:NB": 3, "render:NILV": 0}},
		rtI{set: 55, want: "PB=5"},
		rtI{set: 75, want: "PB=5", quiet: true},
		rtI{set: 80, want: "EVENV"},
		rtI{set: 90, want: "PB=9", counts: rtN{"render:PB": 2, "render:EVENV": 1, "render:NILV": 0}},
	)
}

// RouteBind re-renders, cleaning the old content, when the value changes by its level's comparator.
func TestRouteBindComparator(t *testing.T) {
	derived := func(name string, get func(int) int, eq func(a, b int) bool, steps ...rtI) rtCase {
		return rtCase{
			name:  name,
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{rtDeriveTo(b, r, "d", get, rtIDSet, eq, rtBind[int](r, "B"))}
			},
			initial: "B=0",
			steps:   steps,
			after:   func(h *rtHarness) { h.counts(rtN{"derive:d": 1}) },
		}
	}
	rtCases(t,
		rtCase{
			name:  "change",
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{rtMatch(b, rtGe(0), rtBind[int](r, "B")), rtComp[int](r, "NEG")}
			},
			initial: "B=0",
			steps: []rtI{
				{set: 1, want: "B=1", counts: rtN{"render:B": 2, "clean:B": 1}},
				{set: 2, want: "B=2", counts: rtN{"render:B": 3, "clean:B": 2}},
				{set: -1, want: "NEG", counts: rtN{"render:B": 3, "clean:B": 3}},
				{set: -3, want: "NEG", quiet: true},
				{set: 3, want: "B=3", counts: rtN{"render:B": 4, "clean:B": 3}},
				{set: 4, want: "B=4", counts: rtN{"render:B": 5, "clean:B": 4}},
			},
		},
		derived("eq", rtTens, rtEq,
			rtI{set: 5, want: "B=0", quiet: true},
			rtI{set: 12, want: "B=1"},
			rtI{set: 19, want: "B=1", quiet: true},
			rtI{set: 25, want: "B=2", counts: rtN{"render:B": 3, "clean:B": 2}},
		),
		derived("custom", rtID, func(a, b int) bool { return a%2 == b%2 },
			rtI{set: 2, want: "B=0", quiet: true},
			rtI{set: 3, want: "B=3"},
			rtI{set: 5, want: "B=3", quiet: true},
			rtI{set: 4, want: "B=4", counts: rtN{"render:B": 3}},
		),
		derived("nil", rtTens, nil,
			rtI{set: 1, want: "B=0", counts: rtN{"render:B": 2}},
			rtI{set: 12, want: "B=1"},
			rtI{set: 13, want: "B=1", counts: rtN{"render:B": 4, "clean:B": 3}},
		),
	)
}

// A derive passes an update down only when its value changes by its comparator; nil passes everything.
func TestRouteDeriveSkip(t *testing.T) {
	decade := func(a, b int) bool { return a/10 == b/10 }
	rtCases(t,
		rtCase{
			name:  "equal",
			start: 1,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtPred(r, "pos", rtPos), rtD(b, r, "tens", rtTens,
						rtMatch(b, rtPred(r, "odd", rtOdd), rtComp[int](r, "ODD")),
						rtBind[int](r, "T"),
					)),
					rtComp[int](r, "D"),
				}
			},
			initial: "T=0",
			preds:   "pos=1 odd=0",
			steps: []rtI{
				{set: 5, want: "T=0", quiet: true, preds: "pos=5"},
				{set: 9, want: "T=0", quiet: true, preds: "pos=9"},
				{set: 12, want: "ODD", preds: "pos=12 odd=1"},
				{set: 19, want: "ODD", quiet: true, preds: "pos=19"},
				{set: 25, want: "T=2", preds: "pos=25 odd=2 pos=25 odd=2"},
				{set: 28, want: "T=2", quiet: true, preds: "pos=28"},
				{set: -3, want: "D", preds: "pos=-3 pos=-3"},
				{set: 7, want: "T=0", preds: "pos=7 odd=0", counts: rtN{"derive:tens": 1, "render:T": 3, "render:ODD": 1, "render:D": 1}},
			},
		},
		rtCase{
			name:  "nil",
			start: 1,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{rtDeriveTo(b, r, "tens", rtTens, rtTensSet, nil,
					rtMatch(b, rtPred(r, "odd", rtOdd), rtComp[int](r, "ODD")),
					rtComp[int](r, "Z"),
				)}
			},
			initial: "Z",
			preds:   "odd=0",
			steps: []rtI{
				{set: 5, want: "Z", quiet: true, preds: "odd=0"},
				{set: 12, want: "ODD", preds: "odd=1"},
				{set: 19, want: "ODD", quiet: true, preds: "odd=1"},
				{set: 25, want: "Z", preds: "odd=2 odd=2"},
				{set: 28, want: "Z", quiet: true, preds: "odd=2", counts: rtN{"derive:tens": 1, "render:Z": 2, "render:ODD": 1}},
			},
		},
		rtCase{
			name:  "coarse",
			start: 10,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{rtDeriveTo(b, r, "dec", rtID, rtIDSet, decade,
					rtMatch(b, rtPred(r, "five", func(v int) bool { return v%10 == 5 }), rtComp[int](r, "FIVE")),
					rtBind[int](r, "I"),
				)}
			},
			initial: "I=10",
			preds:   "five=10",
			steps: []rtI{
				{set: 15, want: "I=10", quiet: true},
				{set: 25, want: "FIVE", preds: "five=25"},
				{set: 21, want: "FIVE", quiet: true},
				{set: 31, want: "I=31", preds: "five=31 five=31"},
				{set: 35, want: "I=31", quiet: true},
				{set: 45, want: "FIVE", preds: "five=45", counts: rtN{"derive:dec": 1, "render:I": 2, "render:FIVE": 2}},
			},
		},
		// the outer levels see every update, the same-decade level skips
		rtCase{
			name:  "depth",
			start: 20,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtPred(r, "out", rtGe(0)), rtD(b, r, "tens", rtTens,
						rtMatch(b, rtPred(r, "mid", rtGe(0)), rtDeriveTo(b, r, "dec", rtID, rtIDSet, decade,
							rtMatch(b, rtPred(r, "in", rtOdd), rtComp[int](r, "X")),
							rtBind[int](r, "B"),
						)),
					)),
				}
			},
			initial: "B=2",
			preds:   "out=20 mid=2 in=2",
			steps: []rtI{
				{set: 25, want: "B=2", quiet: true, preds: "out=25"},
				{set: 30, want: "B=2", quiet: true, preds: "out=30 mid=3"},
				{set: 120, want: "B=12", preds: "out=120 mid=12 in=12"},
				{set: 130, want: "B=12", quiet: true, preds: "out=130 mid=13"},
				{set: 210, want: "X", preds: "out=210 mid=21 in=21", counts: rtN{"render:B": 2, "render:X": 1, "derive:tens": 1, "derive:dec": 1}},
			},
		},
	)
}

// The derive skip compares with the last value acted on, not the last seen, like a RouteBind on the entry.
func TestRouteDeriveSkipLast(t *testing.T) {
	near := func(a, b int) bool { return max(a-b, b-a) <= 5 }
	rtEach(t, func(t *testing.T, b bool) {
		h := newRT(t)
		src := NewSource(0)
		entry := DeriveSourceEqual(src, rtID, rtIDSet, near).Route(rtBind[int](h.rec, "E"))
		if b {
			entry = DeriveBeamEqual(src, rtID, near).RouteBeam(rtBind[int](h.rec, "E"))
		}
		h.render(
			rtSlot("d", rtMount(b, src, rtDeriveTo(b, h.rec, "near", rtID, rtIDSet, near,
				rtMatch(b, rtPred(h.rec, "n", func(int) bool { return true }), rtBind[int](h.rec, "B")),
			))),
			rtSlot("e", entry),
		)
		for i, st := range []struct {
			set, shown int
			quiet      bool
		}{{4, 0, true}, {8, 8, false}, {12, 8, true}, {14, 14, false}} {
			at := fmt.Sprintf("step %d (set %d): ", i, st.set)
			nd, ne := len(h.doorActions(h.doors0["d"])), len(h.doorActions(h.doors0["e"]))
			rtSet(h, src, st.set)
			h.expectAll(at, map[string]string{"d": fmt.Sprintf("B=%d", st.shown), "e": fmt.Sprintf("E=%d", st.shown)})
			want := 1
			if st.quiet {
				want = 0
			}
			if gd, ge := len(h.doorActions(h.doors0["d"]))-nd, len(h.doorActions(h.doors0["e"]))-ne; gd != want || ge != want {
				t.Fatalf("%sdoor actions d=%d e=%d, want %d each", at, gd, ge, want)
			}
		}
		if got := h.rec.vals("pred:n="); !slices.Equal(got, []string{"0", "8", "14"}) {
			t.Fatalf("pred n log %v", got)
		}
		h.counts(rtN{"derive:near": 1, "render:B": 3, "render:E": 3})
	})
}

// rtSpy counts clean calls of the wrapped route as "reset:<label>".
type rtSpy[T any] struct {
	rec   *rtRec
	label string
	inner SourceRoute[T]
}

func (s rtSpy[T]) init(src Source[T]) route[T] {
	return &rtSpyRoute[T]{s.rec, s.label, s.inner.init(src)}
}

func (s rtSpy[T]) initBeam(b Beam[T]) route[T] {
	return &rtSpyRoute[T]{s.rec, s.label, s.inner.(BeamRoute[T]).initBeam(b)}
}

type rtSpyRoute[T any] struct {
	rec   *rtRec
	label string
	route[T]
}

func (s *rtSpyRoute[T]) clean() {
	s.rec.hit("reset:" + s.label)
	s.route.clean()
}

// A derive caches a no-match until its value changes; clean resets only the active chain.
func TestRouteDeriveNoMatchCache(t *testing.T) {
	rtCases(t,
		rtCase{
			name:  "seek",
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtD(b, r, "tens", rtTens, rtMatch(b, rtPred(r, "t", func(v int) bool { return v == 5 }), rtComp[int](r, "FIFTY"))),
					rtMatch(b, rtPred(r, "odd", rtOdd), rtComp[int](r, "ODD")),
					rtBind[int](r, "D"),
				}
			},
			initial: "D=0",
			preds:   "t=0 odd=0",
			steps: []rtI{
				{set: 2, want: "D=2", preds: "odd=2"},
				{set: 12, want: "D=12", preds: "t=1 odd=12"},
				{set: 14, want: "D=14", preds: "odd=14"},
				{set: 15, want: "ODD", preds: "odd=15"},
				{set: 17, want: "ODD", quiet: true, preds: "odd=17"},
				{set: 16, want: "D=16", preds: "odd=16 odd=16"},
				{set: 18, want: "D=18", preds: "odd=18"},
				{set: 52, want: "FIFTY", preds: "t=5"},
				{set: 58, want: "FIFTY", quiet: true},
				{set: 63, want: "ODD", preds: "t=6 t=6 odd=63", counts: rtN{
					"derive:tens": 1, "render:FIFTY": 1, "clean:FIFTY": 1, "render:ODD": 2, "clean:ODD": 1,
				}},
			},
		},
		rtCase{
			name:  "match",
			start: 0,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtPos, rtD(b, r, "d", rtTens,
						rtMatch(b, rtPred(r, "t1", func(t int) bool { return t == 1 }), rtComp[int](r, "X")),
						rtMatch(b, rtPred(r, "t2", func(t int) bool { return t == 2 }), rtComp[int](r, "Y")),
					)),
					rtBind[int](r, "D"),
				}
			},
			initial: "D=0",
			steps: []rtI{
				{set: 15, want: "X", preds: "t1=1"},
				{set: 25, want: "Y", preds: "t1=2 t1=2 t2=2"},
				{set: 35, want: "D=35", preds: "t2=3 t1=3 t2=3"},
				{set: 36, want: "D=36"},
				{set: 12, want: "X", preds: "t1=1"},
				{set: 0, want: "D=0", counts: rtN{"render:X": 2, "render:Y": 1, "render:D": 4, "clean:D": 3, "derive:d": 1}},
			},
		},
		rtCase{
			name:  "clean",
			start: 15,
			tree: func(b bool, r *rtRec) []SourceRoute[int] {
				return []SourceRoute[int]{
					rtMatch(b, rtPred(r, "g", rtPos), rtSpy[int]{r, "DV", rtD(b, r, "d", rtTens,
						rtMatch(b, rtPred(r, "t", rtOdd), rtBind[int](r, "T")),
					)}),
					rtMatch(b, rtPred(r, "e", rtEven), rtBind[int](r, "E")),
					rtBind[int](r, "D"),
				}
			},
			initial: "T=1",
			preds:   "g=15 t=1",
			steps: []rtI{
				{set: 17, want: "T=1", quiet: true, preds: "g=17", counts: rtN{"reset:DV": 0}},
				{set: -17, want: "D=-17", preds: "g=-17 g=-17 e=-17", counts: rtN{"reset:DV": 1}},
				{set: 13, want: "T=1", preds: "g=13 t=1", counts: rtN{"render:T": 2, "clean:T": 1, "reset:DV": 1}},
				{set: 20, want: "E=20", preds: "g=20 t=2 g=20 t=2 e=20", counts: rtN{"reset:DV": 2}},
				{set: 22, want: "E=22", preds: "e=22"},
				{set: 23, want: "D=23", preds: "e=23 g=23 e=23", counts: rtN{"reset:DV": 2, "clean:E": 2}},
				{set: 25, want: "D=25", preds: "g=25 e=25"},
				{set: -1, want: "D=-1", preds: "g=-1 e=-1"},
				{set: 27, want: "D=27", preds: "g=27 e=27"},
				{set: 31, want: "T=3", preds: "g=31 t=3", counts: rtN{"reset:DV": 2}},
				{set: 33, want: "T=3", quiet: true, preds: "g=33"},
				{set: 51, want: "T=5", preds: "g=51 t=5", counts: rtN{"derive:d": 1, "render:T": 4, "render:E": 2, "render:D": 5, "reset:DV": 2}},
			},
		},
	)
}

func rtParent(content gox.Elem) *Door {
	p := &Door{}
	p.Inner(freeCtx(), content)
	return p
}

// Every derive runs once per Route render, even one that never matches.
func TestRouteDeriveOnce(t *testing.T) {
	rtCases(t, rtCase{
		name:  "tree",
		start: 0,
		tree: func(b bool, r *rtRec) []SourceRoute[int] {
			return []SourceRoute[int]{
				rtMatch(b, rtLt(0), rtD(b, r, "never", nil, rtComp[int](r, "NEG"))),
				rtMatch(b, rtLt(1000), rtD(b, r, "tens", rtTens,
					rtValue(b, 1, rtComp[int](r, "T1")),
					rtMatch(b, rtGe(5), rtD(b, r, "hund", rtTens,
						rtValue(b, 0, rtComp[int](r, "H0")),
						rtBind[int](r, "HB"),
					)),
					rtBind[int](r, "TB"),
				)),
				rtComp[int](r, "BIG"),
			}
		},
		initial: "TB=0",
		steps: []rtI{
			{set: 15, want: "T1"},
			{set: 55, want: "H0"},
			{set: 250, want: "HB=2"},
			{set: 1500, want: "BIG"},
			{set: 60, want: "H0"},
			{set: 3, want: "TB=0"},
			{set: 60, want: "H0"},
			{set: 2000, want: "BIG"},
			{set: 17, want: "T1", counts: rtN{
				"derive:never": 1, "derive:tens": 1, "derive:hund": 1, "render:NEG": 0, "render:H0": 3, "render:BIG": 2, "render:T1": 2,
			}},
		},
	})
}

// Derives chain across types with each level's comparator, and a bottom RouteSource writes back through all.
func TestRouteDeriveTypes(t *testing.T) {
	type card struct {
		Name string
		Len  int
	}
	names := []string{"zero", "One", "TWO", "three", "four", "Four", "six", "Seven", "eight", "nine"}
	h := newRT(t)
	src := NewSource(90)
	var bottom rtBox[Source[card]]
	toName := func(s Source[int]) Source[string] {
		return DeriveSourceEqual(s, func(v int) string { return names[v] }, func(_ int, n string) int {
			return slices.IndexFunc(names, func(x string) bool { return strings.EqualFold(x, n) })
		}, strings.EqualFold)
	}
	toCard := func(s Source[string]) Source[card] {
		return DeriveSource(s, func(n string) card { return card{strings.ToLower(n), len(n)} }, func(_ string, c card) string { return c.Name })
	}
	h.render(rtSlot("c", src.Route(
		RouteDerive(rtDerive(h.rec, "d1", rtTensSource),
			RouteValue(9, rtComp[int](h.rec, "NINE")),
			RouteDerive(rtDerive(h.rec, "d2", toName),
				RouteValue("one", rtComp[string](h.rec, "ONE")),
				RouteDerive(rtDerive(h.rec, "d3", toCard),
					RouteValue(card{"two", 3}, rtSource(h.rec, "BOTTOM", &bottom)),
					RouteMatch(func(c card) bool { return c.Len > 4 }, rtBind[card](h.rec, "CARD")),
				),
				rtBind[string](h.rec, "S"),
			),
		),
	)))
	rtRun(h, "c", src, "NINE", rtI{set: 20, want: "BOTTOM"})
	if ret := rtSet(h, bottom.last(), card{Name: "three"}); !rtSameErrs(ret, []error{nil}) || src.Get() != 30 {
		t.Fatalf("bottom write returned %v and mapped to %d, want 30", ret, src.Get())
	}
	rtRun(h, "c", src, "CARD={three 5}",
		rtI{set: 40, want: "S=four"},
		rtI{set: 50, want: "S=four", quiet: true},
		rtI{set: 10, want: "ONE"},
		rtI{set: 70, want: "CARD={seven 5}"},
		rtI{set: 20, want: "BOTTOM"},
	)
	if bottom.len() != 2 || bottom.last().Get() != (card{"two", 3}) {
		t.Fatalf("want a fresh bottom render, got %d", bottom.len())
	}
	rtMutate(h, bottom.last(), func(card) card { return card{Name: "EIGHT"} })
	if got := src.Get(); got != 80 {
		t.Fatalf("bottom write mapped to %d, want 80", got)
	}
	h.expect("c", "CARD={eight 5}")
	h.oneDoor("c")
	h.counts(rtN{
		"derive:d1": 1, "derive:d2": 1, "derive:d3": 1,
		"render:NINE": 1, "render:ONE": 1, "render:S": 1, "render:BOTTOM": 2, "render:CARD": 3,
	})
}

// Leaf content is cleaned once on every way out, a bind on each re-render, the shown leaf at instance end.
func TestRouteClean(t *testing.T) {
	rtEach(t, func(t *testing.T, b bool) {
		h := newRT(t)
		r := h.rec
		s1, s2, s3, s4, s5, s6 := NewSource(1), NewSource(1), NewSource(2), NewSource(1), NewSource(7), NewSource(1)
		h.render(
			rtSlot("disp", rtMount(b, s1, rtMatch(b, rtEven, rtComp[int](r, "E1")), rtComp[int](r, "D1"))),
			rtSlot("end", rtMount(b, s2,
				rtMatch(b, rtLt(10), rtComp[int](r, "S2")),
				rtMatch(b, rtLt(100), rtComp[int](r, "M2")),
				rtComp[int](r, "D2"),
			)),
			rtSlot("fall", rtMount(b, s3, rtMatch(b, rtPos, rtMatch(b, rtEven, rtComp[int](r, "N3"))), rtComp[int](r, "D3"))),
			rtSlot("top", rtMount(b, s4, rtMatch(b, rtPos, rtBind[int](r, "T4")))),
			rtSlot("deep", rtMount(b, s5,
				rtMatch(b, rtLt(1000), rtD(b, r, "dd", nil,
					rtMatch(b, rtPos, rtValue(b, 7, rtComp[int](r, "V5")), rtBind[int](r, "B5")),
				)),
				rtComp[int](r, "D5"),
			)),
			rtSlot("bind", rtMount(b, s6, rtBind[int](r, "B6"))),
		)
		h.counts(rtN{"render:D1": 1, "render:S2": 1, "render:N3": 1, "render:T4": 1, "render:V5": 1, "render:B6": 1, "derive:dd": 1})
		// a leaf displaced by an earlier sibling, and back
		rtRun(h, "disp", s1, "D1",
			rtI{set: 2, want: "E1", counts: rtN{"render:D1": 1, "clean:D1": 1, "render:E1": 1, "clean:E1": 0}},
			rtI{set: 3, want: "D1", counts: rtN{"render:E1": 1, "clean:E1": 1, "render:D1": 2, "clean:D1": 1}},
			rtI{set: 4, want: "E1", counts: rtN{"render:E1": 2, "clean:E1": 1, "render:D1": 2, "clean:D1": 2}},
		)
		// an ended branch and the one the re-selection picks
		rtRun(h, "end", s2, "S2",
			rtI{set: 50, want: "M2", counts: rtN{"clean:S2": 1, "render:M2": 1, "clean:M2": 0}},
			rtI{set: 500, want: "D2", counts: rtN{"clean:M2": 1, "render:D2": 1, "clean:D2": 0}},
			rtI{set: 5, want: "S2", counts: rtN{"clean:D2": 1, "render:S2": 2, "clean:S2": 1}},
		)
		// a nested end falling through to the leaf
		rtRun(h, "fall", s3, "N3",
			rtI{set: 3, want: "D3", counts: rtN{"clean:N3": 1, "render:D3": 1}},
			rtI{set: 5, want: "D3", quiet: true, counts: rtN{"clean:N3": 1, "render:N3": 1, "clean:D3": 0}},
			rtI{set: 0, want: "D3", quiet: true, counts: rtN{"render:D3": 1, "clean:D3": 0}},
			rtI{set: 4, want: "N3", counts: rtN{"clean:D3": 1, "render:N3": 2, "clean:N3": 1}},
		)
		// a top-level end that blanks the container
		rtRun(h, "top", s4, "T4=1",
			rtI{set: 0, want: "", counts: rtN{"clean:T4": 1}},
			rtI{set: -1, want: "", quiet: true, counts: rtN{"clean:T4": 1, "render:T4": 1}},
			rtI{set: 2, want: "T4=2", counts: rtN{"clean:T4": 1, "render:T4": 2}},
			rtI{set: 3, want: "T4=3", counts: rtN{"clean:T4": 2, "render:T4": 3}},
		)
		// deep leaves on every path break
		rtRun(h, "deep", s5, "V5",
			rtI{set: 8, want: "B5=8", counts: rtN{"clean:V5": 1}},
			rtI{set: 9, want: "B5=9", counts: rtN{"clean:B5": 1}},
			rtI{set: 2000, want: "D5", counts: rtN{"clean:B5": 2, "render:D5": 1}},
			rtI{set: 7, want: "V5", counts: rtN{"clean:D5": 1, "render:V5": 2}},
			rtI{set: 0, want: "D5", counts: rtN{"clean:V5": 2, "render:D5": 2}},
			rtI{set: 7, want: "V5", counts: rtN{"clean:D5": 2, "clean:V5": 2, "render:V5": 3, "derive:dd": 1}},
		)
		rtRun(h, "bind", s6, "B6=1",
			rtI{set: 2, want: "B6=2", counts: rtN{"clean:B6": 1}},
			rtI{set: 3, want: "B6=3", counts: rtN{"clean:B6": 2}},
			rtI{set: 3, want: "B6=3", quiet: true, ret: []error{}, counts: rtN{"clean:B6": 2, "render:B6": 3}},
			rtI{set: 4, want: "B6=4", counts: rtN{"clean:B6": 3, "render:B6": 4}},
		)
		h.inst.Kill()
		h.balanced("E1", "D1", "S2", "M2", "D2", "N3", "D3", "T4", "V5", "B5", "D5", "B6")
	})
}

// A break cleans every node of the old chain exactly once and no other node.
func TestRouteCleanOldPath(t *testing.T) {
	resets := func(chainA, chainC, d int, more rtN) rtN {
		m := rtN{
			"reset:O": chainA + chainC, "reset:MA": chainA, "reset:A": chainA,
			"reset:MB": chainC, "reset:DV": chainC, "reset:MC": chainC, "reset:C": chainC,
			"reset:D": d,
		}
		maps.Copy(m, more)
		return m
	}
	rtCases(t, rtCase{
		name:  "chain",
		start: 1,
		tree: func(b bool, r *rtRec) []SourceRoute[int] {
			spy := func(label string, inner SourceRoute[int]) SourceRoute[int] { return rtSpy[int]{r, label, inner} }
			return []SourceRoute[int]{
				spy("O", rtMatch(b, rtPred(r, "o", rtPos),
					spy("MA", rtMatch(b, rtPred(r, "a", func(v int) bool { return v%4 == 0 }), spy("A", rtBind[int](r, "A")))),
					spy("MB", rtMatch(b, rtPred(r, "b", rtLt(50)), spy("DV", rtD(b, r, "d", nil,
						spy("MC", rtMatch(b, rtPred(r, "c", rtOdd), spy("C", rtBind[int](r, "C")))),
					)))),
				)),
				spy("D", rtBind[int](r, "D")),
			}
		},
		initial: "C=1",
		preds:   "o=1 a=1 b=1 c=1",
		steps: []rtI{
			{set: 3, want: "C=3", preds: "o=3 b=3 c=3", counts: resets(0, 0, 0, rtN{"render:C": 2, "clean:C": 1})},
			{set: 4, want: "A=4", preds: "o=4 b=4 c=4 o=4 a=4", counts: resets(0, 1, 0, rtN{"render:C": 2, "clean:C": 2, "render:A": 1})},
			{set: 8, want: "A=8", preds: "o=8 a=8", counts: resets(0, 1, 0, rtN{"render:A": 2, "clean:A": 1})},
			{set: 3, want: "C=3", preds: "o=3 a=3 o=3 a=3 b=3 c=3", counts: resets(1, 1, 0, rtN{"bind:C=3": 2, "render:C": 3, "clean:C": 2, "clean:A": 2})},
			{set: 5, want: "C=5", preds: "o=5 b=5 c=5", counts: resets(1, 1, 0, rtN{"render:C": 4, "clean:C": 3})},
			{set: 62, want: "D=62", preds: "o=62 b=62 o=62 a=62 b=62", counts: resets(1, 2, 0, rtN{"render:C": 4, "clean:C": 4, "render:D": 1})},
			{set: 64, want: "A=64", preds: "o=64 a=64", counts: resets(1, 2, 1, rtN{"render:A": 3, "clean:A": 2, "clean:D": 1})},
			{set: 0, want: "D=0", preds: "o=0 o=0", counts: resets(2, 2, 1, rtN{"clean:A": 3, "render:D": 2})},
			{set: 7, want: "C=7", preds: "o=7 a=7 b=7 c=7", counts: resets(2, 2, 2, rtN{"render:C": 5, "clean:C": 4, "clean:D": 2})},
		},
		after: func(h *rtHarness) {
			h.counts(rtN{"derive:d": 1})
			h.inst.Kill()
			h.balanced("A", "C", "D")
		},
	})
}

// Leaving a branch drops a bind's stored value and a derive's last value.
func TestRouteReleasesValues(t *testing.T) {
	kind := func(f func(int) bool) func(*rtBig) bool { return func(b *rtBig) bool { return f(b.kind) } }
	same := func(s Source[*rtBig]) Source[*rtBig] { return s }
	for _, c := range []struct {
		name   string
		routes func(r *rtRec) []SourceRoute[*rtBig]
		after  string
	}{
		{"end", func(r *rtRec) []SourceRoute[*rtBig] {
			return []SourceRoute[*rtBig]{RouteMatch(kind(rtLt(10)), rtBind[*rtBig](r, "B")), rtComp[*rtBig](r, "D")}
		}, "D"},
		{"leaf", func(r *rtRec) []SourceRoute[*rtBig] {
			return []SourceRoute[*rtBig]{RouteMatch(kind(rtGe(10)), rtComp[*rtBig](r, "X")), rtBind[*rtBig](r, "B")}
		}, "X"},
		{"derive leaf", func(r *rtRec) []SourceRoute[*rtBig] {
			return []SourceRoute[*rtBig]{RouteMatch(kind(rtGe(10)), rtComp[*rtBig](r, "X")), RouteDerive(same, rtBind[*rtBig](r, "B"))}
		}, "X"},
		{"beam derive leaf", func(r *rtRec) []SourceRoute[*rtBig] {
			return []SourceRoute[*rtBig]{RouteBeamDerive(func(b Beam[*rtBig]) Beam[*rtBig] { return b },
				RouteBeamMatch(kind(rtGe(10)), rtComp[*rtBig](r, "X")),
				rtBind[*rtBig](r, "B"),
			)}
		}, "X"},
		{"path", func(r *rtRec) []SourceRoute[*rtBig] {
			return []SourceRoute[*rtBig]{
				RouteMatch(kind(rtPos),
					RouteMatch(kind(func(v int) bool { return v%4 == 0 }), rtComp[*rtBig](r, "A")),
					RouteMatch(kind(rtLt(50)), RouteDerive(same, RouteMatch(kind(rtOdd), rtBind[*rtBig](r, "B")))),
				),
				rtComp[*rtBig](r, "D"),
			}
		}, "D"},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newRT(t)
			b1 := newRTBig(1)
			w := weak.Make(b1)
			src := NewSource(b1)
			b1 = nil
			h.render(rtSlot("a", src.Route(c.routes(h.rec)...)))
			h.expect("a", "B=big1")
			rtSet(h, src, newRTBig(62))
			h.expect("a", c.after)
			h.oneDoor("a")
			for i := 0; w.Value() != nil; i++ {
				if i == 20 {
					t.Fatal("the old value is still reachable after leaving its branch")
				}
				runtime.GC()
				time.Sleep(5 * time.Millisecond)
			}
			runtime.KeepAlive(src)
			runtime.KeepAlive(h)
		})
	}
}

// Leaves render once per activation and get their level's Source or Beam: the same each time, live, writable.
func TestRouteLeaves(t *testing.T) {
	h := newRT(t)
	r := h.rec
	src := NewSource(0)
	sbox, dbox := &rtBox[Source[int]]{}, &rtBox[Source[int]]{}
	bbox, kbox, ebox := &rtBox[Beam[int]]{}, &rtBox[Beam[int]]{}, &rtBox[Beam[int]]{}
	h.render(
		rtSlot("c", src.Route(RouteMatch(rtPos, rtComp[int](r, "C")), rtComp[int](r, "DC"))),
		rtSlot("s", src.Route(RouteMatch(rtPos, rtSource(r, "S", sbox)), rtComp[int](r, "DS"))),
		rtSlot("b", src.Route(RouteMatch(rtPos, rtBeam(r, "B", bbox)), rtComp[int](r, "DB"))),
		rtSlot("k", src.RouteBeam(RouteBeamMatch(rtPos, rtBeam(r, "K", kbox)), rtComp[int](r, "DK"))),
		rtSlot("d", src.Route(RouteMatch(rtPos, RouteDerive(rtTensSource, rtSource(r, "T", dbox))), rtComp[int](r, "DD"))),
		rtSlot("e", src.Route(RouteMatch(rtPos, RouteBeamDerive(rtTensBeam, rtBeam(r, "U", ebox))), rtComp[int](r, "DE"))),
	)
	slots := []string{"c", "s", "b", "k", "d", "e"}
	on := map[string]string{"c": "C", "s": "S", "b": "B", "k": "K", "d": "T", "e": "U"}
	off := map[string]string{"c": "DC", "s": "DS", "b": "DB", "k": "DK", "d": "DD", "e": "DE"}
	live := func(at string, b Beam[int], want int) {
		t.Helper()
		if v, ok := b.Read(h.page); !ok || v != want || b.Get() != want {
			t.Fatalf("%sleaf reads %d %v, get %d, want %d", at, v, ok, b.Get(), want)
		}
	}
	for i, st := range []struct {
		set, renders int
		switched     bool
	}{{12, 1, true}, {15, 1, false}, {27, 1, false}, {0, 1, true}, {31, 2, true}, {38, 2, false}} {
		at := fmt.Sprintf("step %d (set %d): ", i, st.set)
		step := func() { rtSet(h, src, st.set) }
		if st.switched {
			step()
		} else {
			h.quietAll(at, slots, step)
		}
		if st.set == 0 {
			h.expectAll(at, off)
			continue
		}
		h.expectAll(at, on)
		for _, l := range on {
			h.countsAt(at, rtN{"render:" + l: st.renders})
		}
		for _, b := range []Beam[int]{sbox.last(), bbox.last(), kbox.last()} {
			live(at, b, st.set)
		}
		live(at, dbox.last(), st.set/10)
		live(at, ebox.last(), st.set/10)
	}
	if !sbox.same() || !bbox.same() || !kbox.same() || !dbox.same() || !ebox.same() || sbox.len() != 2 || dbox.len() != 2 || ebox.len() != 2 {
		t.Fatal("leaves should get the same Source or Beam on every activation")
	}
	if sbox.last() != src {
		t.Fatal("a RouteSource under RouteMatch should get the routed source itself")
	}
	h.quietAll("write: ", slots, func() { rtSet(h, dbox.last(), 7) })
	if got := src.Get(); got != 78 {
		t.Fatalf("derived write mapped to %d, want 78", got)
	}
	for _, s := range slots {
		h.oneDoor(s)
	}
}

// Writes from a leaf route like any update, and a capture kept after its branch ended still writes the root.
func TestRouteWrites(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		h := newRT(t)
		src := NewSource(1)
		box := &rtBox[Source[int]]{}
		h.render(rtSlot("a", src.Route(
			RouteMatch(rtSmall, rtSource(h.rec, "SMALL", box)),
			RouteMatch(rtGe(10), rtComp[int](h.rec, "BIG")),
		)))
		s := box.last()
		h.quiet("a", func() {
			if ret := rtSet(h, s, 5); !rtSameErrs(ret, []error{nil}) || src.Get() != 5 {
				t.Fatalf("write returned %v, root %d", ret, src.Get())
			}
		})
		rtSet(h, s, 15)
		h.expect("a", "BIG")
		if ret := rtSet(h, s, 15); len(ret) != 0 {
			t.Fatalf("equal write should be suppressed, got %v", ret)
		}
		rtMutate(h, s, func(v int) int { return v - 10 })
		h.expect("a", "SMALL")
		h.counts(rtN{"render:SMALL": 2, "clean:SMALL": 1, "render:BIG": 1, "clean:BIG": 1})
		if box.len() != 2 || !box.same() {
			t.Fatal("re-activation should pass the same source")
		}
		h.oneDoor("a")
	})
	t.Run("derive", func(t *testing.T) {
		h := newRT(t)
		src := NewSource(15)
		box := &rtBox[Source[int]]{}
		h.render(rtSlot("a", src.Route(
			RouteMatch(rtLt(100), RouteDerive(rtDerive(h.rec, "div", rtTensSource),
				RouteMatch(rtLt(3), rtSource(h.rec, "L", box)),
				rtBind[int](h.rec, "TB"),
			)),
			rtComp[int](h.rec, "BIG"),
		)))
		s := box.last()
		rootIs := func(want int) {
			t.Helper()
			if got := src.Get(); got != want {
				t.Fatalf("root is %d, want %d", got, want)
			}
		}
		if ret := rtSet(h, s, 4); !rtSameErrs(ret, []error{nil}) {
			t.Fatalf("derived write returned %v", ret)
		}
		h.expect("a", "TB=4")
		rootIs(45)
		h.quiet("a", func() {
			if ret := rtSet(h, s, 4); len(ret) != 0 {
				t.Fatalf("equal derived write returned %v", ret)
			}
		})
		rtSet(h, s, 1)
		h.expect("a", "L")
		rootIs(15)
		rtSet(h, src, 500)
		h.expect("a", "BIG")
		rtSet(h, s, 2)
		h.expect("a", "L")
		rootIs(20)
		rtMutate(h, s, func(v int) int { return v + 5 })
		h.expect("a", "TB=7")
		rootIs(70)
		h.oneDoor("a")
		h.counts(rtN{"derive:div": 1, "render:L": 3, "clean:L": 3, "render:BIG": 1, "clean:BIG": 1})
		if !box.same() {
			t.Fatal("captures should be the same derived source")
		}
	})
}

// One Route call is one dynamic container, however deep its routes nest.
func TestRouteOneDoor(t *testing.T) {
	var ladder func(r *rtRec, l, depth int) SourceRoute[int]
	ladder = func(r *rtRec, l, depth int) SourceRoute[int] {
		leaf := rtComp[int](r, fmt.Sprintf("V%d", l))
		var kids []SourceRoute[int]
		if l >= 3 {
			kids = append(kids, RouteBeamValue(l*100+1, leaf))
		} else {
			kids = append(kids, RouteValue(l*100+1, leaf))
		}
		if l < depth {
			kids = append(kids, ladder(r, l+1, depth))
		}
		kids = append(kids, rtBind[int](r, fmt.Sprintf("D%d", l)))
		switch l {
		case 1:
			return RouteMatch(rtGe(l*100), kids...)
		case 2:
			return RouteDerive(rtDerive(r, "d2", func(s Source[int]) Source[int] { return DeriveSource(s, rtID, rtIDSet) }), kids...)
		case 4:
			return RouteBeamDerive(rtDeriveBeam(r, "d4", func(b Beam[int]) Beam[int] { return DeriveBeam(b, rtID) }), rtBeams(kids)...)
		}
		return RouteBeamMatch(rtGe(l*100), rtBeams(kids)...)
	}
	for depth := 1; depth <= 5; depth++ {
		t.Run(strconv.Itoa(depth), func(t *testing.T) {
			h := newRT(t)
			src := NewSource(0)
			h.render(rtSlot("l", src.Route(ladder(h.rec, 1, depth), rtBind[int](h.rec, "TOP"))))
			if h.baseIDs != 3 {
				t.Fatalf("render created %d ids, want the root, the path target and one route door", h.baseIDs)
			}
			var steps []rtI
			for l := 1; l <= depth; l++ {
				steps = append(steps, rtI{set: l*100 + 1, want: fmt.Sprintf("V%d", l)})
				d := l
				if l < depth && (l+1 == 2 || l+1 == 4) {
					d = l + 1
				}
				steps = append(steps, rtI{set: l*100 + 50, want: fmt.Sprintf("D%d=%d", d, l*100+50)})
			}
			steps = append(steps, rtI{set: 0, want: "TOP=0"}, rtI{set: 101, want: "V1"})
			rtRun(h, "l", src, "TOP=0", steps...)
		})
	}
}

// Doors in branch content are separate from the route door, and an inner Route renews on every activation.
func TestRouteInnerDoors(t *testing.T) {
	t.Run("route", func(t *testing.T) {
		h := newRT(t)
		h.strict = false
		src := NewSource(0)
		h.render(rtSlot("o", src.Route(
			RouteMatch(rtLt(100), RouteSource(func(s Source[int]) gox.Elem {
				h.rec.hit("outer-leaf")
				return s.Route(
					RouteValue(1, rtComp[int](h.rec, "IA")),
					RouteDerive(rtDerive(h.rec, "inner", func(s Source[int]) Source[int] { return s }), rtBind[int](h.rec, "IB")),
				)
			})),
			rtComp[int](h.rec, "OUT"),
		)))
		h.expect("o", "IB=0")
		ids := rtDoorIDs(h.html("o"))
		if len(ids) != 2 || ids[0] != h.doors0["o"] {
			t.Fatalf("want the route door and an inner door, got %v", ids)
		}
		outer, inner := ids[0], ids[1]
		for _, st := range []struct {
			v    int
			want string
		}{{1, "IA"}, {2, "IB=2"}, {3, "IB=3"}, {1, "IA"}} {
			n := len(h.doorActions(inner))
			h.quiet("o", func() { rtSet(h, src, st.v) })
			h.expect("o", st.want)
			if got := len(h.doorActions(inner)); got != n+1 {
				t.Fatalf("set %d: want one inner action, got %d", st.v, got-n)
			}
		}
		rtSet(h, src, 150)
		h.expect("o", "OUT")
		rtSet(h, src, 4)
		h.expect("o", "IB=4")
		if ids := rtDoorIDs(h.html("o")); len(ids) != 2 || ids[0] != outer || ids[1] == inner {
			t.Fatalf("want the route door and a new inner door, got %v (old inner %d)", ids, inner)
		}
		h.counts(rtN{"outer-leaf": 2, "derive:inner": 2, "render:OUT": 1, "clean:OUT": 1, "render:IA": 2})
		if n := len(h.doorActions(outer)); n != 2 {
			t.Fatalf("want 2 route door replaces, got %d", n)
		}
	})
	t.Run("bind", func(t *testing.T) {
		h := newRT(t)
		h.strict = false
		src := NewSource(1)
		r := h.rec
		h.render(rtSlot("a", src.Route(RouteMatch(rtLt(10), RouteBeam(func(b Beam[int]) gox.Elem {
			r.hit("render:leaf")
			return b.Bind(func(v int) gox.Elem {
				return func(cur gox.Cursor) error {
					r.hit("render:inner")
					OnClean(cur.Context(), func() { r.hit("clean:inner") })
					return cur.Text(fmt.Sprint("in=", v))
				}
			})
		})), rtComp[int](r, "D"))))
		ids := rtDoorIDs(h.html("a"))
		if len(ids) != 2 || ids[0] != h.doors0["a"] {
			t.Fatalf("want the route door and one inner door, got %v", ids)
		}
		inner := ids[1]
		rtRun(h, "a", src, "in=1",
			rtI{set: 2, want: "in=2", quiet: true},
			rtI{set: 3, want: "in=3", quiet: true, counts: rtN{"render:leaf": 1, "render:inner": 3, "clean:inner": 2}},
		)
		if acts := h.doorActions(inner); len(acts) != 2 {
			t.Fatalf("inner door: want two replaces, got %+v", acts)
		}
		rtRun(h, "a", src, "in=3",
			rtI{set: 20, want: "D", counts: rtN{"render:leaf": 1, "clean:inner": 3}},
			rtI{set: 5, want: "in=5", counts: rtN{"render:leaf": 2, "render:inner": 4, "clean:inner": 3, "clean:D": 1}},
		)
		if ids := rtDoorIDs(h.html("a")); len(ids) != 2 || ids[1] == inner || len(h.doorActions(inner)) != 2 {
			t.Fatalf("want a fresh inner door and the old one silent, got %v", ids)
		}
	})
}

// A route list is a template: each render and position gets its own branch, bind value and derive cache.
func TestRouteIsolation(t *testing.T) {
	tree := func(b bool, r *rtRec) []SourceRoute[int] {
		return []SourceRoute[int]{
			rtMatch(b, rtGe(100), rtD(b, r, "tens", rtTens,
				rtValue(b, 10, rtComp[int](r, "TEN")),
				rtMatch(b, rtEven, rtBind[int](r, "EB")),
			)),
			rtMatch(b, rtEven, rtComp[int](r, "EVEN")),
			rtMatch(b, rtLt(10), rtComp[int](r, "SMALL")),
			rtBind[int](r, "D"),
		}
	}
	t.Run("elem", func(t *testing.T) {
		rtEach(t, func(t *testing.T, b bool) {
			h := newRT(t)
			src := NewSource(1)
			el := rtMount(b, src, tree(b, h.rec)...)
			h.render(rtSlot("a", el), rtSlot("b", el))
			if h.doors0["a"] == h.doors0["b"] {
				t.Fatalf("both slots share door %d", h.doors0["a"])
			}
			for i, st := range []rtI{
				{set: 2, want: "SMALL", quiet: true},
				{set: 12, want: "EVEN"},
				{set: 13, want: "D=13"},
				{set: 100, want: "TEN"},
				{set: 101, want: "TEN", quiet: true},
				{set: 110, want: "EVEN"},
				{set: 111, want: "D=111"},
				{set: 120, want: "EB=12"},
				{set: 140, want: "EB=14"},
				{set: 5, want: "SMALL"},
			} {
				at := fmt.Sprintf("step %d (set %d): ", i, st.set)
				na, nb := len(h.doorActions(h.doors0["a"])), len(h.doorActions(h.doors0["b"]))
				rtSet(h, src, st.set)
				h.expectAll(at, map[string]string{"a": st.want, "b": st.want})
				want := 1
				if st.quiet {
					want = 0
				}
				if ga, gb := len(h.doorActions(h.doors0["a"]))-na, len(h.doorActions(h.doors0["b"]))-nb; ga != want || gb != want {
					t.Fatalf("%sdoor actions a=%d b=%d, want %d each", at, ga, gb, want)
				}
				h.oneDoorAt(at, "a")
				h.oneDoorAt(at, "b")
			}
			h.counts(rtN{"derive:tens": 2, "render:SMALL": 4, "render:EVEN": 4, "render:D": 4, "render:TEN": 2, "render:EB": 4})
		})
	})
	t.Run("positions", func(t *testing.T) {
		shared := func(h *rtHarness) SourceRoute[int] {
			return RouteDerive(rtDerive(h.rec, "d", func(s Source[int]) Source[int] { return s }),
				RouteValue(7, rtComp[int](h.rec, "SEVEN")),
				rtBind[int](h.rec, "S"),
			)
		}
		h := newRT(t)
		tmpl := shared(h)
		src, other := NewSource(1), NewSource(1)
		h.render(
			rtSlot("t", src.Route(
				RouteMatch(rtLt(10), tmpl),
				RouteMatch(rtLt(20), RouteValue(15, rtComp[int](h.rec, "X")), tmpl),
				rtComp[int](h.rec, "D"),
			)),
			rtSlot("u", other.Route(tmpl)),
		)
		h.counts(rtN{"derive:d": 3, "render:S": 2})
		h.quiet("u", func() {
			rtRun(h, "t", src, "S=1",
				rtI{set: 12, want: "S=12", counts: rtN{"render:S": 3, "clean:S": 1}},
				rtI{set: 7, want: "SEVEN"},
				rtI{set: 15, want: "X"},
				rtI{set: 25, want: "D"},
				rtI{set: 16, want: "S=16"},
				rtI{set: 3, want: "S=3"},
				rtI{set: 7, want: "SEVEN", counts: rtN{"render:SEVEN": 2, "render:X": 1}},
			)
		})
		h.expect("u", "S=1")
		h.quiet("t", func() {
			rtRun(h, "u", other, "S=1", rtI{set: 7, want: "SEVEN"}, rtI{set: 8, want: "S=8"})
		})
		h.counts(rtN{"derive:d": 3})
	})
	t.Run("locations", func(t *testing.T) {
		shared := newRTRec()
		routes := []SourceRoute[Location]{
			RouteModel[rtPathA](rtComp[rtPathA](shared, "A")),
			RouteModel[rtPathItem](
				RouteMatch(func(m rtPathItem) bool { return m.ID < 10 }, rtBind[rtPathItem](shared, "SB")),
				rtComp[rtPathItem](shared, "BIGITEM"),
			),
			RouteMatch(rtFirst("admin"), rtStrip(false, nil, "", "admin", RouteModel[rtPathItem](rtBind[rtPathItem](shared, "ADMIN")))),
			rtComp[Location](shared, "404"),
		}
		h1, h2 := newRT(t), newRT(t)
		h1.at("/a")
		h2.at("/a")
		h1.render(rtSlot("r", Route(routes...)), rtSlot("twin", Route(routes...)))
		h2.render(rtSlot("r", Route(routes...)))
		for i, st := range []struct {
			h    *rtHarness
			uri  string
			a, b string
		}{
			{h1, "/item/3", "SB={true 3 }", "A"},
			{h2, "/item/30", "SB={true 3 }", "BIGITEM"},
			{h1, "/item/40", "BIGITEM", "BIGITEM"},
			{h2, "/item/4", "BIGITEM", "SB={true 4 }"},
			{h1, "/item/4", "SB={true 4 }", "SB={true 4 }"},
			{h1, "/admin/zzz", "404", "SB={true 4 }"},
			{h2, "/admin/item/3", "404", "ADMIN={true 3 }"},
			{h1, "/admin/item/7", "ADMIN={true 7 }", "ADMIN={true 3 }"},
			{h2, "/zzz", "ADMIN={true 7 }", "404"},
			{h1, "/item/4", "SB={true 4 }", "404"},
			{h2, "/item/5", "SB={true 4 }", "SB={true 5 }"},
		} {
			at := fmt.Sprintf("step %d (%s): ", i, st.uri)
			other := map[*rtHarness]*rtHarness{h1: h2, h2: h1}[st.h]
			other.quietAll(at, []string{"r"}, func() { rtSet(st.h, Router(st.h.page), rtLoc(st.uri)) })
			h1.expectAll(at, map[string]string{"r": st.a, "twin": st.a})
			h2.expectAt(at, "r", st.b)
			h1.oneDoorAt(at, "r")
			h1.oneDoorAt(at, "twin")
			h2.oneDoorAt(at, "r")
		}
		if got := shared.count("render:SB"); got != 8 {
			t.Fatalf("render:SB=%d: bind values leaked between containers, log %v", got, shared.all())
		}
	})
}

// Concurrent instances share route values without sharing state.
func TestRouteInstances(t *testing.T) {
	t.Run("parallel", func(t *testing.T) {
		rtEach(t, func(t *testing.T, b bool) {
			shared := newRTRec()
			var mu sync.Mutex
			derived := map[any]int{}
			count := func(key any) {
				mu.Lock()
				defer mu.Unlock()
				derived[key]++
			}
			derive := func(get func(int) int, rs ...SourceRoute[int]) SourceRoute[int] {
				if b {
					return RouteBeamDerive(func(x Beam[int]) Beam[int] { count(x); return DeriveBeam(x, get) }, rtBeams(rs)...)
				}
				return RouteDerive(func(s Source[int]) Source[int] { count(s); return DeriveSource(s, get, rtIDSet) }, rs...)
			}
			routes := []SourceRoute[int]{
				rtMatch(b, rtGe(100), derive(rtTens,
					rtValue(b, 10, rtComp[int](shared, "TEN")),
					rtMatch(b, rtEven, derive(rtID, rtBind[int](shared, "EB"))),
				)),
				rtMatch(b, rtEven, rtComp[int](shared, "EVEN")),
				rtMatch(b, rtLt(10), rtComp[int](shared, "SMALL")),
				rtBind[int](shared, "D"),
			}
			const n = 6
			t.Run("instances", func(t *testing.T) {
				for i := range n {
					t.Run(strconv.Itoa(i), func(t *testing.T) {
						t.Parallel()
						h := newRT(t)
						src := NewSource(1)
						h.render(rtSlot("a", rtMount(b, src, routes...)))
						rtRun(h, "a", src, "SMALL",
							rtI{set: 12, want: "EVEN"},
							rtI{set: 13, want: "D=13"},
							rtI{set: 100, want: "TEN"},
							rtI{set: 110, want: "EVEN"},
							rtI{set: 111, want: "D=111"},
							rtI{set: 120, want: "EB=12"},
							rtI{set: 5, want: "SMALL"},
						)
					})
				}
			})
			mu.Lock()
			defer mu.Unlock()
			if len(derived) != 2*n {
				t.Fatalf("want %d derived sources, got %d", 2*n, len(derived))
			}
			for key, c := range derived {
				if c != 1 {
					t.Fatalf("derive ran %d times for %v", c, key)
				}
			}
			deadline := time.Now().Add(5 * time.Second)
			for _, l := range []string{"SMALL", "EVEN", "D", "TEN", "EB"} {
				for shared.count("clean:"+l) != shared.count("render:"+l) {
					if time.Now().After(deadline) {
						t.Fatalf("clean:%s=%d, want %d", l, shared.count("clean:"+l), shared.count("render:"+l))
					}
					time.Sleep(time.Millisecond)
				}
			}
			if got := shared.count("render:EVEN"); got != 2*n {
				t.Fatalf("render:EVEN=%d, want %d", got, 2*n)
			}
		})
	})
	t.Run("shared source", func(t *testing.T) {
		src := NewSource(1)
		shared := newRTRec()
		el := src.RouteBeam(
			RouteBeamMatch(rtEven, rtComp[int](shared, "X")),
			RouteBeamMatch(rtLt(10), rtComp[int](shared, "Y")),
			rtComp[int](shared, "D"),
		)
		ha, hb := newRT(t), newRT(t)
		ha.render(rtSlot("a", el))
		ha.expect("a", "Y")
		if ret := drainOp(t, src.Update(ha.page, 2)); !rtSameErrs(ret, []error{nil}) {
			t.Fatalf("update returned %v", ret)
		}
		ha.eventually("a", "Y")
		hb.render(rtSlot("a", el))
		hb.expect("a", "X")
		for i, st := range []struct {
			v    int
			a, b string
		}{{4, "Y", "X"}, {6, "Y", "X"}, {3, "Y", "Y"}, {20, "X", "X"}, {13, "D", "D"}, {8, "X", "X"}} {
			ctx := ha.page
			if i%2 == 1 {
				ctx = hb.page
			}
			drainOp(t, src.Update(ctx, st.v))
			ha.eventually("a", st.a)
			hb.eventually("a", st.b)
			ha.oneDoor("a")
			hb.oneDoor("a")
		}
	})
}

// A parent reload renders a new route instance; replacing or unmounting the parent drops it.
func TestRouteParentDoor(t *testing.T) {
	xy := func(r *rtRec) []SourceRoute[int] {
		return []SourceRoute[int]{
			RouteMatch(rtPred(r, "x", rtEven), rtD(false, r, "d", nil, rtComp[int](r, "X"))),
			RouteMatch(rtPred(r, "y", rtLt(10)), rtComp[int](r, "Y")),
			rtComp[int](r, "D"),
		}
	}
	t.Run("reload", func(t *testing.T) {
		h := newRT(t)
		h.strict = false
		src := NewSource(1)
		p := rtParent(rtSlot("in", src.Route(xy(h.rec)...)))
		h.render(rtSlot("outer", mountDoor(p)))
		rtRun(h, "in", src, "Y", rtI{set: 2, want: "Y", quiet: true})
		d0 := h.door("in")
		n0 := len(h.doorActions(d0))
		h.settle(func(ctx context.Context) { p.Reload(ctx) })
		h.expect("in", "X")
		d1 := h.door("in")
		if d1 == d0 || h.doorCount("in") != 1 {
			t.Fatalf("reload should render one new route door, got %v (old %d)", rtDoorIDs(h.html("in")), d0)
		}
		h.counts(rtN{"derive:d": 2, "render:Y": 1, "clean:Y": 1, "render:X": 1})
		rtRun(h, "in", src, "X",
			rtI{set: 4, want: "X", quiet: true},
			rtI{set: 6, want: "X", quiet: true},
			rtI{set: 3, want: "Y"},
		)
		if got := len(h.doorActions(d0)); got != n0 {
			t.Fatalf("old route door %d got %d new actions", d0, got-n0)
		}
		if got, want := h.rec.vals("pred:x="), []string{"1", "2", "4", "6", "3", "3"}; !slices.Equal(got, want) {
			t.Fatalf("pred x log %v, want %v: the old instance must stay silent", got, want)
		}
		h.counts(rtN{"derive:d": 2, "clean:X": 1, "render:Y": 2})
	})
	for _, c := range []struct {
		name  string
		op    func(p *Door, ctx context.Context)
		outer string
	}{
		{"inner", func(p *Door, ctx context.Context) { p.Inner(ctx, textElem("OTHER")) }, "OTHER"},
		{"outer", func(p *Door, ctx context.Context) { p.Outer(ctx, tagElem("section", "other", textElem("OTHER"))) }, "OTHER"},
		{"unmount", func(p *Door, ctx context.Context) { p.Unmount(ctx) }, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newRT(t)
			h.strict = false
			src := NewSource(1)
			p := rtParent(rtSlot("in", src.Route(xy(h.rec)...)))
			h.render(rtSlot("outer", mountDoor(p)))
			rtRun(h, "in", src, "Y", rtI{set: 2, want: "Y", quiet: true})
			h.settle(func(ctx context.Context) { c.op(p, ctx) })
			if _, ok := h.slot("in"); ok {
				t.Fatalf("route slot survived, dom:\n%s", h.dom())
			}
			h.expect("outer", c.outer)
			h.counts(rtN{"render:Y": 1, "clean:Y": 1})
			preds, acts := len(h.rec.vals("pred:")), len(h.client.actions())
			deadline := time.Now().Add(5 * time.Second)
			for v := 100; len(rtSet(h, src, v)) != 0; v++ {
				if time.Now().After(deadline) {
					t.Fatal("updates still find subscribers")
				}
			}
			if got := h.rec.vals("pred:"); len(got) != preds {
				t.Fatalf("dropped routes still evaluate predicates: %v", got[preds:])
			}
			if got := h.client.actions(); len(got) != acts {
				t.Fatalf("dropped routes still produce actions: %+v", got[acts:])
			}
			h.counts(rtN{"render:X": 0, "render:D": 0, "derive:d": 1})
		})
	}
}

// The instance end cleans the shown leaf and later updates find no subscribers.
func TestRouteInstanceEnd(t *testing.T) {
	h := newRT(t)
	src := NewSource(1)
	h.render(rtSlot("a", src.Route(
		RouteMatch(rtPred(h.rec, "x", rtEven), rtComp[int](h.rec, "X")),
		RouteMatch(rtPred(h.rec, "y", rtLt(10)), rtComp[int](h.rec, "Y")),
	)))
	rtRun(h, "a", src, "Y", rtI{set: 2, want: "Y", quiet: true})
	h.inst.Kill()
	h.waitCounts(rtN{"render:Y": 1, "clean:Y": 1})
	preds := len(h.rec.vals("pred:"))
	deadline := time.Now().Add(5 * time.Second)
	for v := 100; len(drainOp(t, src.Update(context.Background(), v))) != 0; v++ {
		if time.Now().After(deadline) {
			t.Fatal("updates still find subscribers")
		}
	}
	for v := 200; v < 205; v++ {
		if ret := drainOp(t, src.Update(context.Background(), v)); len(ret) != 0 {
			t.Fatalf("update after instance end returned %v", ret)
		}
	}
	if got := h.rec.vals("pred:"); len(got) != preds {
		t.Fatalf("predicates ran after instance end: %v", got[preds:])
	}
	h.counts(rtN{"render:X": 0})
}

// A leaf render error keeps the old content, is logged, and the route keeps working.
func TestRouteRenderError(t *testing.T) {
	h := newRT(t)
	h.strict = false
	src := NewSource(0)
	boom := RouteComp[int](gox.Elem(func(gox.Cursor) error {
		h.rec.hit("render:ERR")
		return errors.New("rt boom")
	}))
	h.render(rtSlot("a", src.Route(RouteValue(1, boom), rtBind[int](h.rec, "D"))))
	id := h.door("a")
	rtSet(h, src, 1)
	h.expect("a", "D=0")
	if n := len(h.doorActions(id)); n != 0 {
		t.Fatalf("failed render should send nothing, got %d actions", n)
	}
	if n := h.waitLogged("door rendering error", 1); n != 1 {
		t.Fatalf("want one render error log, got %d: %v", n, h.logs.all())
	}
	rtSet(h, src, 2)
	h.expect("a", "D=2")
	select {
	case <-h.inst.ended:
		t.Fatal("instance ended after a branch render error")
	default:
	}
	if h.doorCount("a") != 1 {
		t.Fatalf("want one door, dom:\n%s", h.dom())
	}
	h.routeDoorAt("", "a")
	h.counts(rtN{"render:ERR": 1, "render:D": 2})
}

// A branch redirecting on activation reaches the target exactly once.
func TestRouteRedirect(t *testing.T) {
	for _, now := range []bool{false, true} {
		for _, initial := range []bool{false, true} {
			t.Run(fmt.Sprintf("render=%v/initial=%v", now, initial), func(t *testing.T) {
				h := newRT(t)
				start := 0
				if initial {
					start = 1
				}
				src := NewSource(start)
				redirect := RouteSource(func(s Source[int]) gox.Elem {
					return func(cur gox.Cursor) error {
						h.rec.hit("render:redirect")
						if now {
							s.Update(cur.Context(), 2)
						} else {
							OnReady(cur.Context(), func(ctx context.Context) { s.Update(ctx, 2) })
						}
						return cur.Text("redirect")
					}
				})
				h.render(rtSlot("a", src.Route(
					RouteValue(1, redirect),
					RouteValue(2, rtComp[int](h.rec, "target")),
					rtComp[int](h.rec, "home"),
				)))
				if !initial {
					h.expect("a", "home")
					rtSet(h, src, 1)
				}
				h.eventually("a", "target")
				time.Sleep(20 * time.Millisecond)
				h.waitIdle()
				h.expect("a", "target")
				h.oneDoor("a")
				h.counts(rtN{"render:redirect": 1, "render:target": 1})
				if src.Get() != 2 {
					t.Fatalf("source is %d", src.Get())
				}
			})
		}
	}
}

// A hook of branch content writes the source and dies with its branch.
func TestRouteHookWrite(t *testing.T) {
	h := newRT(t)
	h.strict = false
	src := NewSource(1)
	hooks := &rtBox[uint64]{}
	leaf := RouteSource(func(s Source[int]) gox.Elem {
		return func(cur gox.Cursor) error {
			h.rec.hit("render:H")
			c := cur.Context().Value(common.KeyCore).(core.Core)
			hook, ok := c.Door().RegisterHook(func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) bool {
				h.rec.hit("handler")
				s.Update(ctx, s.Get()+20)
				return true
			}, RaceSerial, false)
			if ok {
				hooks.put(hook.HookID)
			}
			return cur.Text("H")
		}
	})
	h.render(rtSlot("a", src.Route(RouteMatch(rtLt(10), leaf), rtBind[int](h.rec, "D"))))
	trigger := func(id uint64) bool {
		return h.root.TriggerHook(id, httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), 0)
	}
	h.expect("a", "H")
	if !trigger(hooks.last()) {
		t.Fatal("hook trigger failed")
	}
	h.eventually("a", "D=21")
	h.counts(rtN{"handler": 1, "render:H": 1, "clean:D": 0})
	stale := hooks.last()
	rtSet(h, src, 3)
	h.expect("a", "H")
	if trigger(stale) {
		t.Fatal("the left branch's hook still triggers")
	}
	if !trigger(hooks.last()) {
		t.Fatal("hook trigger failed")
	}
	h.eventually("a", "D=23")
	h.counts(rtN{"handler": 2, "render:H": 2, "clean:D": 1})
}

// RouteBind renders in the door render, so a slow render does not hold back propagation.
func TestRouteBindRenderInDoor(t *testing.T) {
	rtEach(t, func(t *testing.T, b bool) {
		h := newRT(t)
		src := NewSource(1)
		gate := make(chan struct{})
		var once sync.Once
		release := func() { once.Do(func() { close(gate) }) }
		defer release()
		h.render(rtSlot("a", rtMount(b, src, RouteBind(func(v int) gox.Elem {
			if v == 2 {
				<-gate
			}
			return textElem(fmt.Sprint("B=", v))
		}))))
		h.expect("a", "B=1")
		select {
		case err := <-src.Update(h.page, 2):
			if err != nil {
				t.Fatalf("update failed: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("update propagation waited for the bind render")
		}
		release()
		h.eventually("a", "B=2")
	})
}

// Bursts of updates end on the last value's branch with one door; a NoSkip source sees every value.
func TestRouteBursts(t *testing.T) {
	sticky := func(h *rtHarness, src Source[int]) gox.Elem {
		return src.Route(
			RouteMatch(rtPred(h.rec, "even", rtEven), rtComp[int](h.rec, "EVEN")),
			RouteMatch(rtPred(h.rec, "small", rtSmall), rtComp[int](h.rec, "SMALL")),
			rtComp[int](h.rec, "D"),
		)
	}
	burst := func(h *rtHarness, src Source[int], vals ...int) {
		h.t.Helper()
		var chs []<-chan error
		h.settle(func(ctx context.Context) {
			for _, v := range vals {
				chs = append(chs, src.Update(ctx, v))
			}
		})
		for i, ch := range chs {
			ret := drainOp(h.t, ch)
			if !rtSameErrs(ret, []error{nil}) && (i == len(chs)-1 || !rtSameErrs(ret, []error{context.Canceled})) {
				h.t.Fatalf("update %d returned %v", i, ret)
			}
		}
	}
	t.Run("settle", func(t *testing.T) {
		h := newRT(t)
		src, bsrc := NewSource(1), NewSource(0)
		h.render(rtSlot("a", sticky(h, src)), rtSlot("b", bsrc.Route(RouteMatch(rtPos, rtBind[int](h.rec, "B")))))
		for i, c := range []struct {
			vals []int
			want string
		}{
			{[]int{2, 12, 3, 13, 4, 12}, "EVEN"},
			{[]int{3, 12, 4, 13}, "D"},
			{[]int{4, 3, 11, 12}, "EVEN"},
			{[]int{5, 8, 2, 15}, "D"},
		} {
			burst(h, src, c.vals...)
			h.expectAt(fmt.Sprintf("burst %d: ", i), "a", c.want)
			h.oneDoorAt(fmt.Sprintf("burst %d: ", i), "a")
		}
		burst(h, bsrc, 1, 2, 3, 0)
		h.expect("b", "")
		burst(h, bsrc, 5, 0, 6, 7)
		h.expect("b", "B=7")
		h.oneDoor("b")
	})
	t.Run("goroutine", func(t *testing.T) {
		h := newRT(t)
		src := NewSource(12)
		h.render(rtSlot("a", sticky(h, src)))
		done := make(chan struct{})
		go func() {
			defer close(done)
			for _, v := range []int{14, 16, 4, 2, 20, 22, 6, 13} {
				src.Update(h.page, v)
			}
		}()
		<-done
		h.eventually("a", "D")
		h.oneDoor("a")
	})
	t.Run("held", func(t *testing.T) {
		h := newRT(t)
		src := NewSource(0)
		h.render(rtSlot("r", src.Route(
			RouteMatch(rtIn(1, 100), RouteMatch(rtLt(50), rtComp[int](h.rec, "A")), RouteMatch(rtGe(50), rtComp[int](h.rec, "B"))),
			rtComp[int](h.rec, "D"),
		)))
		final := func(v int) string {
			switch {
			case v <= 0 || v >= 100:
				return "D"
			case v < 50:
				return "A"
			}
			return "B"
		}
		for _, seq := range [][]int{{10, 200, 60}, {10, 60, 10, 60, 5}, {60, 0, 60, 300}, {5, 55, 5, 300, 55}, {200, 10, 0}} {
			h.client.hold()
			var last <-chan error
			for _, v := range seq {
				last = src.Update(h.page, v)
			}
			h.waitQueued(1)
			h.client.release()
			if ret := drainOp(t, last); !rtSameErrs(ret, []error{nil}) {
				t.Fatalf("seq %v: last update returned %v", seq, ret)
			}
			h.eventually("r", final(seq[len(seq)-1]))
			n := len(h.client.actions())
			time.Sleep(20 * time.Millisecond)
			if got := len(h.client.actions()); got != n {
				t.Fatalf("seq %v: %d late actions", seq, got-n)
			}
			h.oneDoor("r")
		}
	})
	t.Run("noskip", func(t *testing.T) {
		h := newRT(t)
		src := NewSourceNoSkip(1)
		h.render(rtSlot("a", sticky(h, src)))
		var chs []<-chan error
		h.settle(func(ctx context.Context) {
			for _, v := range []int{2, 12, 3, 13, 11, 4} {
				chs = append(chs, src.Update(ctx, v))
			}
		})
		for i, ch := range chs {
			if ret := drainOp(t, ch); !rtSameErrs(ret, []error{nil}) {
				t.Fatalf("update %d returned %v", i, ret)
			}
		}
		h.expect("a", "EVEN")
		want := "even=1 small=1 small=2 small=12 even=12 even=3 even=3 small=3 small=13 even=13 small=13 even=11 small=11 even=4"
		if got := strings.Join(h.rec.vals("pred:"), " "); got != want {
			t.Fatalf("predicate log\n got %s\nwant %s", got, want)
		}
		h.counts(rtN{"render:EVEN": 2, "render:SMALL": 2, "render:D": 1, "clean:SMALL": 2, "clean:D": 1})
		h.oneDoor("a")
	})
	t.Run("concurrent", func(t *testing.T) {
		h := newRT(t)
		src := NewSource(1)
		r := h.rec
		h.render(rtSlot("a", src.Route(
			RouteMatch(func(v int) bool { return v >= 0 && v%2 == 0 },
				RouteMatch(rtLt(50), rtComp[int](r, "EL")),
				rtD(false, r, "div", rtTens, rtBind[int](r, "ET")),
			),
			RouteMatch(rtIn(0, 50), rtComp[int](r, "SM")),
			RouteMatch(rtLt(0), rtBind[int](r, "NEG")),
			rtBind[int](r, "D"),
		)))
		h.expect("a", "SM")
		var mu sync.Mutex
		var chs []<-chan error
		var wg sync.WaitGroup
		for g := range 8 {
			wg.Go(func() {
				for j := range 30 {
					ch := src.Update(h.page, (g*37+j*53)%200)
					mu.Lock()
					chs = append(chs, ch)
					mu.Unlock()
				}
			})
		}
		wg.Wait()
		for _, ch := range chs {
			drainOp(t, ch)
		}
		if ret := rtSet(h, src, -1); !rtSameErrs(ret, []error{nil}) {
			t.Fatalf("final update returned %v", ret)
		}
		h.expect("a", "NEG=-1")
		h.oneDoor("a")
		h.counts(rtN{"derive:div": 1})
	})
	t.Run("router", func(t *testing.T) {
		h := newRT(t)
		loc := h.at("/a")
		h.render(rtSlot("m", Route(
			RouteModel[rtPathA](rtSource[rtPathA](h.rec, "A", nil)),
			RouteModel[rtPathB](rtSource[rtPathB](h.rec, "B", nil)),
			RouteModel[rtPathItem](rtBind[rtPathItem](h.rec, "ITEM")),
			rtComp[Location](h.rec, "404"),
		)))
		for i, sc := range []struct {
			uris []string
			want string
		}{
			{[]string{"/a", "/item/1", "/zzz", "/item/2?tab=q", "/a", "/item/9"}, "ITEM={true 9 }"},
			{[]string{"/item/1", "/item/2", "/item/3", "/zzz"}, "404"},
			{[]string{"/zzz", "/item/4", "/a"}, "A"},
			{[]string{"/a", "/b", "/a", "/b"}, "B"},
			{[]string{"/item/5", "/item/5?tab=x", "/item/5"}, "ITEM={true 5 }"},
		} {
			h.settle(func(ctx context.Context) {
				for _, uri := range sc.uris {
					loc.Update(ctx, rtLoc(uri))
				}
			})
			at := fmt.Sprintf("script %d: ", i)
			h.expectAt(at, "m", sc.want)
			h.oneDoorAt(at, "m")
			rtLocIs(h, loc, sc.uris[len(sc.uris)-1])
		}
	})
}

// 200 rapid updates: the predicate runs once per value the container sees plus once per path break.
func TestRouteRapidUpdates(t *testing.T) {
	expected := func(seen []int) []string {
		var log []string
		p := func(v int) bool {
			log = append(log, strconv.Itoa(v))
			return v%10 != 0
		}
		state := ""
		sel := func(v int) {
			switch {
			case !p(v):
				state = "Z"
			case rtEven(v):
				state = "E"
			default:
				state = "B"
			}
		}
		for i, v := range seen {
			switch {
			case i == 0:
				sel(v)
			case state == "Z":
				if p(v) {
					state = map[bool]string{true: "E", false: "B"}[rtEven(v)]
				}
			case !p(v) || state == "E" && !rtEven(v):
				sel(v)
			case rtEven(v):
				state = "E"
			}
		}
		return log
	}
	for _, c := range []struct {
		name           string
		parent, noSkip bool
		workers        int
	}{
		{"root/w8", false, false, 8},
		{"root/w1", false, false, 1},
		{"parent", true, false, 8},
		{"parent/noskip", true, true, 8},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newRTN(t, c.workers)
			src := NewSource(0)
			if c.noSkip {
				src = NewSourceNoSkip(0)
			}
			r := h.rec
			route := rtSlot("a", src.Route(
				RouteMatch(rtPred(r, "p", func(v int) bool { return v%10 != 0 }),
					RouteMatch(rtEven, rtComp[int](r, "E")),
					rtBind[int](r, "B"),
				),
				rtComp[int](r, "Z"),
			))
			if c.parent {
				route = rtSlot("outer", mountDoor(rtParent(route)))
			}
			h.render(route)
			h.expect("a", "Z")
			const last = 201
			var chs []<-chan error
			h.settle(func(ctx context.Context) {
				for v := 1; v <= last; v++ {
					chs = append(chs, src.Update(ctx, v))
				}
			})
			for i, ch := range chs {
				ret := drainOp(t, ch)
				if !rtSameErrs(ret, []error{nil}) && (c.noSkip || i == len(chs)-1 || !rtSameErrs(ret, []error{context.Canceled})) {
					t.Fatalf("update %d returned %v", i+1, ret)
				}
			}
			h.expect("a", fmt.Sprintf("B=%d", last))
			h.oneDoor("a")
			got := r.vals("pred:p=")
			var seen []int
			for _, s := range got {
				v, _ := strconv.Atoi(s)
				if len(seen) == 0 || seen[len(seen)-1] != v {
					seen = append(seen, v)
				}
			}
			if c.parent && !c.noSkip {
				for i := 1; i < len(seen); i++ {
					if seen[i] <= seen[i-1] {
						t.Fatalf("seen values not rising: %v", got)
					}
				}
				if len(seen) == 0 || seen[0] != 0 || seen[len(seen)-1] != last {
					t.Fatalf("pred log %v must start at 0 and end at %d", got, last)
				}
			} else {
				for i, v := range seen {
					if v != i || len(seen) != last+1 {
						t.Fatalf("container saw %v, want every value 0..%d", seen, last)
					}
				}
			}
			if want := expected(seen); !slices.Equal(got, want) {
				t.Fatalf("pred log %v\nwant %v", got, want)
			}
		})
	}
}

type rtLoaders struct {
	r *rtRec
	g *heavyGate
}

func (l rtLoaders) heavy(label string) gox.Elem {
	return tagElem("b", "heavy-"+label, l.g.elem("heavy "+label))
}

func (l rtLoaders) skel(label string) gox.Elem {
	return tagElem("i", "skel-"+label, textElem("skel "+label))
}

func (l rtLoaders) routes() []SourceRoute[string] {
	to := func(label string) func(string) bool { return func(v string) bool { return v == label } }
	inner := tagElem("section", "wrap-inner", func(cur gox.Cursor) error {
		if IsDocument(cur.Context()) {
			return l.heavy("inner")(cur)
		}
		l.r.hit("deferred inner")
		DeferredInner(cur.Context(), l.heavy("inner"))
		return l.skel("inner")(cur)
	})
	outer := gox.Elem(func(cur gox.Cursor) error {
		if IsDocument(cur.Context()) {
			return l.heavy("outer")(cur)
		}
		l.r.hit("deferred outer")
		DeferredOuter(cur.Context(), l.heavy("outer"))
		return l.skel("outer")(cur)
	})
	return []SourceRoute[string]{
		RouteMatch(to("inner"), RouteComp[string](inner)),
		RouteMatch(to("outer"), RouteComp[string](outer)),
		RouteComp[string](tagElem("p", "home", textElem("home"))),
	}
}

// Deferred content in branches: none on the initial render, skeleton then content on later switches.
func TestRouteDeferred(t *testing.T) {
	sameDoor := func(h *rtHarness) {
		h.t.Helper()
		if h.doorCount("r") != 1 || h.door("r") != h.doors0["r"] {
			h.t.Fatalf("route door moved, dom:\n%s", h.dom())
		}
	}
	for _, start := range []string{"inner", "outer"} {
		t.Run("document/"+start, func(t *testing.T) {
			h := newRT(t)
			h.strict = false
			src := NewSource(start)
			h.render(rtSlot("r", src.Route(rtLoaders{h.rec, newHeavyGate(t)}.routes()...)))
			h.expect("r", "heavy "+start)
			if strings.Contains(h.initial, "skel") || len(h.client.actions()) != 0 {
				t.Fatalf("the initial branch should render in the page:\n%s", h.initial)
			}
			if start == "inner" && !strings.HasPrefix(h.html("r"), "<section ") {
				t.Fatalf("want the section as the route container, got %q", h.html("r"))
			}
			h.counts(rtN{"deferred inner": 0, "deferred outer": 0})
			for _, next := range []string{"outer", "inner", "outer"} {
				if next != start {
					rtSet(h, src, next)
					h.eventually("r", "heavy "+next)
					sameDoor(h)
				}
			}
		})
	}
	t.Run("inner", func(t *testing.T) {
		h := newRT(t)
		h.strict = false
		g := newHeavyGate(t)
		src := NewSource("home")
		h.render(rtSlot("r", src.Route(rtLoaders{h.rec, g}.routes()...)))
		g.shut()
		rtSet(h, src, "inner")
		g.waitEntered(t)
		h.waitIdle()
		h.expect("r", "skel inner")
		if html := h.html("r"); !strings.HasPrefix(html, "<section ") || !strings.Contains(html, `id="skel-inner"`) {
			t.Fatalf("want the skeleton inside the section, got %q", html)
		}
		g.open()
		h.eventually("r", "heavy inner")
		rtSet(h, src, "home")
		h.expect("r", "home")
		rtSet(h, src, "inner")
		h.eventually("r", "heavy inner")
		if html := h.html("r"); !strings.HasPrefix(html, "<section ") || strings.Contains(html, "skel") {
			t.Fatalf("want heavy inside the section, got %q", html)
		}
		var names []string
		for _, a := range h.doorActions(h.doors0["r"]) {
			names = append(names, a.name)
		}
		if want := []string{"door_replace", "door_update", "door_replace", "door_replace", "door_update"}; !slices.Equal(names, want) {
			t.Fatalf("route door actions %v, want %v", names, want)
		}
		h.counts(rtN{"deferred inner": 2})
	})
	t.Run("outer", func(t *testing.T) {
		h := newRT(t)
		h.strict = false
		g := newHeavyGate(t)
		src := NewSource("home")
		h.render(rtSlot("r", src.Route(rtLoaders{h.rec, g}.routes()...)))
		d := h.doors0["r"]
		container := func(tag string) {
			t.Helper()
			if html := h.html("r"); !strings.HasPrefix(html, "<"+tag+" ") || !strings.Contains(html, fmt.Sprintf(`data-d0r="%d"`, d)) {
				t.Fatalf("want <%s> as route door %d, got %q", tag, d, html)
			}
		}
		g.shut()
		rtSet(h, src, "outer")
		g.waitEntered(t)
		h.waitIdle()
		h.expect("r", "skel outer")
		container("i")
		g.open()
		h.eventually("r", "heavy outer")
		container("b")
		rtSet(h, src, "inner")
		h.eventually("r", "heavy inner")
		container("section")
		rtSet(h, src, "outer")
		h.eventually("r", "heavy outer")
		container("b")
		sameDoor(h)
		h.counts(rtN{"deferred outer": 2, "deferred inner": 1})
	})
}

// Parallel children inside a branch render with the page and on every activation.
func TestRouteParallelContent(t *testing.T) {
	h := newRT(t)
	r := h.rec
	h.at("/a")
	par := RouteSource(func(s Source[rtPathA]) gox.Elem {
		return func(cur gox.Cursor) error {
			for i := range 4 {
				err := Parallel().Proxy(cur, tagElem("p", fmt.Sprint("par-", i), func(cur gox.Cursor) error {
					time.Sleep(time.Duration(4-i) * time.Millisecond)
					r.hit("par")
					return cur.Text(fmt.Sprint("P", i))
				}))
				if err != nil {
					return err
				}
			}
			return nil
		}
	})
	h.render(rtSlot("r", Route(
		RouteModel[rtPathA](par),
		RouteModel[rtPathB](rtComp[rtPathB](r, "B")),
		rtComp[Location](r, "404"),
	)))
	h.counts(rtN{"par": 4})
	rtRun(h, "r", Router(h.page), "P0P1P2P3",
		rtL{set: rtLoc("/b"), want: "B"},
		rtL{set: rtLoc("/a"), want: "P0P1P2P3"},
		rtL{set: rtLoc("/zzz"), want: "404"},
		rtL{set: rtLoc("/a?x=1"), want: "P0P1P2P3"},
	)
	h.counts(rtN{"par": 12})
}

// Beam routes, model ones included, fit in Route next to Source routes.
func TestRouteMixedFlavors(t *testing.T) {
	var (
		_ SourceRoute[int]      = RouteBeamValue[int](0)
		_ SourceRoute[int]      = RouteBeamMatch[int](rtPos)
		_ SourceRoute[int]      = RouteBind(func(int) gox.Elem { return nil })
		_ SourceRoute[int]      = RouteComp[int](nil)
		_ SourceRoute[int]      = RouteBeamDerive(func(b Beam[int]) Beam[int] { return b })
		_ SourceRoute[Location] = RouteBeamModel[rtPathA]()
	)
	h := newRT(t)
	src := NewSource(0)
	loc := rtLocSource("/c")
	sbox, bbox := &rtBox[Source[int]]{}, &rtBox[Beam[int]]{}
	h.render(
		rtSlot("a", src.Route(
			RouteMatch(rtLt(0), rtSource(h.rec, "NEG", sbox)),
			RouteBeamMatch(rtGt(100), rtBeam(h.rec, "HUGE", bbox)),
			RouteBeamValue(50, rtComp[int](h.rec, "FIFTY")),
			RouteBeamDerive(rtTensBeam, RouteBeamValue(3, rtComp[int](h.rec, "THIRTIES"))),
			rtBind[int](h.rec, "B"),
		)),
		rtSlot("loc", loc.Route(
			RouteBeamModel[rtPathA](rtComp[rtPathA](h.rec, "A")),
			RouteModel[rtPathB](rtComp[rtPathB](h.rec, "PB")),
			rtBind[Location](h.rec, "L"),
		)),
	)
	rtRun(h, "a", src, "B=0",
		rtI{set: -1, want: "NEG"},
		rtI{set: 101, want: "HUGE"},
		rtI{set: 50, want: "FIFTY"},
		rtI{set: 35, want: "THIRTIES"},
		rtI{set: 39, want: "THIRTIES", quiet: true},
		rtI{set: 7, want: "B=7"},
		rtI{set: -5, want: "NEG", counts: rtN{"render:NEG": 2, "render:HUGE": 1, "render:FIFTY": 1, "render:THIRTIES": 1, "render:B": 2}},
	)
	if sbox.last() != src || bbox.last().Get() != -5 {
		t.Fatal("leaves should get the routed source")
	}
	rtRun(h, "loc", loc, "L=/c",
		rtL{set: rtLoc("/a"), want: "A"},
		rtL{set: rtLoc("/b"), want: "PB"},
		rtL{set: rtLoc("/c?x=1"), want: "L=/c?x=1"},
		rtL{set: rtLoc("/a"), want: "A", counts: rtN{"render:A": 2, "render:PB": 1}},
	)
}

type rtLocStep struct {
	uri    string
	want   string
	quiet  bool
	same   bool
	counts map[string]int
}

// rtModelCase is a location script run with Source and Beam routes, on a Location source and doors.Route.
type rtModelCase struct {
	name    string
	uri     string
	routes  func(b bool, r *rtRec) []SourceRoute[Location]
	initial string
	steps   []rtLocStep
}

func rtModelCases(t *testing.T, cases ...rtModelCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rtEach(t, func(t *testing.T, b bool) {
				for _, router := range []bool{false, true} {
					t.Run(map[bool]string{false: "source", true: "doors.Route"}[router], func(t *testing.T) {
						h := newRT(t)
						var loc Source[Location]
						var el gox.Elem
						if router {
							loc = h.at(c.uri)
							el = Route(c.routes(b, h.rec)...)
						} else {
							loc = rtLocSource(c.uri)
							el = rtMount(b, loc, c.routes(b, h.rec)...)
						}
						h.render(rtSlot("m", el))
						steps := make([]rtL, len(c.steps))
						for i, s := range c.steps {
							steps[i] = rtL{set: rtLoc(s.uri), want: s.want, quiet: s.quiet, counts: s.counts, ret: []error{nil}}
							if s.same {
								steps[i].ret = []error{}
							}
						}
						rtRun(h, "m", loc, c.initial, steps...)
					})
				}
			})
		})
	}
}

// RouteModel decodes the location: binds follow, equal models are quiet, values match by DeepEqual.
func TestRouteModelDecode(t *testing.T) {
	rtModelCases(t,
		rtModelCase{
			name: "bind",
			uri:  "/item/5?tab=x",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{rtModel[rtPathItem](b, rtBind[rtPathItem](r, "ITEM"))}
			},
			initial: "ITEM={true 5 x}",
			steps: []rtLocStep{
				{uri: "/item/6", want: "ITEM={true 6 }"},
				{uri: "/zzz", want: "", counts: rtN{"render:ITEM": 2, "clean:ITEM": 2}},
				{uri: "/item/7?tab=q", want: "ITEM={true 7 q}"},
				{uri: "/item/7?other=1&tab=q", want: "ITEM={true 7 q}", quiet: true},
				{uri: "/item/7?tab=q&other=1", want: "ITEM={true 7 q}", quiet: true, same: true},
				{uri: "/item/7", want: "ITEM={true 7 }", counts: rtN{"render:ITEM": 4, "clean:ITEM": 3}},
			},
		},
		rtModelCase{
			name: "value",
			uri:  "/item/1",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{rtModel[rtPathItem](b,
					rtValue(b, rtItem1, rtOnce[rtPathItem](b, r, "ONE")),
					rtBind[rtPathItem](r, "OTHER"),
				)}
			},
			initial: "ONE",
			steps: []rtLocStep{
				{uri: "/item/1?tab=x", want: "OTHER={true 1 x}", counts: rtN{"render:ONE": 1, "clean:ONE": 1, "render:OTHER": 1}},
				{uri: "/item/2", want: "OTHER={true 2 }", counts: rtN{"render:OTHER": 2, "clean:OTHER": 1}},
				{uri: "/item/1", want: "ONE", counts: rtN{"render:ONE": 2, "clean:OTHER": 2}},
				{uri: "/item/1?zz=9", want: "ONE", quiet: true, counts: rtN{"render:ONE": 2, "render:OTHER": 2}},
			},
		},
	)
}

// Model routes switch in one door and stay while their model decodes; an invalid model type never matches.
func TestRouteModelTransitions(t *testing.T) {
	rtModelCases(t,
		rtModelCase{
			name: "a-b-a",
			uri:  "/a",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					rtModel[rtPathA](b, rtOnce[rtPathA](b, r, "A")),
					rtModel[rtPathB](b, rtOnce[rtPathB](b, r, "B")),
					rtComp[Location](r, "404"),
				}
			},
			initial: "A",
			steps: []rtLocStep{
				{uri: "/b", want: "B", counts: rtN{"render:A": 1, "clean:A": 1, "render:B": 1}},
				{uri: "/a", want: "A", counts: rtN{"render:A": 2, "clean:B": 1}},
				{uri: "/b?x=1", want: "B"},
				{uri: "/b?x=2", want: "B", quiet: true},
				{uri: "/zzz", want: "404"},
				{uri: "/zzz?y=1", want: "404", quiet: true},
				{uri: "/a", want: "A", counts: rtN{"render:A": 3, "clean:A": 2, "render:B": 2, "clean:B": 2, "render:404": 1, "clean:404": 1}},
			},
		},
		rtModelCase{
			name: "overlap",
			uri:  "/item/abc",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					rtModel[rtPathItem](b, rtOnce[rtPathItem](b, r, "INT")),
					rtModel[rtPathName](b, rtOnce[rtPathName](b, r, "NAME")),
					rtComp[Location](r, "404"),
				}
			},
			initial: "NAME",
			steps: []rtLocStep{
				{uri: "/item/5", want: "NAME", quiet: true},
				{uri: "/other", want: "404"},
				{uri: "/item/5", want: "INT"},
				{uri: "/item/abc", want: "NAME", counts: rtN{"render:NAME": 2, "render:INT": 1, "clean:INT": 1, "render:404": 1}},
			},
		},
		rtModelCase{
			name: "adapter-error",
			uri:  "/a",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					rtModel[rtPathBad](b, rtOnce[rtPathBad](b, r, "BAD")),
					rtModel[int](b, rtOnce[int](b, r, "INT")),
					rtModel[rtPathA](b, rtOnce[rtPathA](b, r, "A")),
					rtComp[Location](r, "404"),
				}
			},
			initial: "A",
			steps: []rtLocStep{
				{uri: "/zzz", want: "404"},
				{uri: "/", want: "404", quiet: true},
				{uri: "/a", want: "A", counts: rtN{"render:BAD": 0, "render:INT": 0, "render:A": 2, "render:404": 1}},
			},
		},
	)
}

// A model route whose children do not match the decoded model, or that has none, falls through.
func TestRouteModelFallThrough(t *testing.T) {
	one := func(b bool, r *rtRec, rest ...SourceRoute[rtPathItem]) SourceRoute[Location] {
		return rtModel[rtPathItem](b, append([]SourceRoute[rtPathItem]{rtValue(b, rtItem1, rtOnce[rtPathItem](b, r, "ONE"))}, rest...)...)
	}
	rtModelCases(t,
		rtModelCase{
			name: "404",
			uri:  "/x",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{one(b, r), rtComp[Location](r, "404")}
			},
			initial: "404",
			steps: []rtLocStep{
				{uri: "/item/2", want: "404", quiet: true, counts: rtN{"render:404": 1, "clean:404": 0}},
				{uri: "/item/3", want: "404", quiet: true},
				{uri: "/item/1", want: "ONE", counts: rtN{"render:ONE": 1, "clean:404": 1}},
				{uri: "/item/2", want: "404", counts: rtN{"render:404": 2, "clean:ONE": 1}},
				{uri: "/x", want: "404", quiet: true, counts: rtN{"render:404": 2, "clean:404": 1, "render:ONE": 1}},
			},
		},
		rtModelCase{
			name: "first-render",
			uri:  "/item/2",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{one(b, r), rtComp[Location](r, "404")}
			},
			initial: "404",
			steps: []rtLocStep{
				{uri: "/item/1", want: "ONE", counts: rtN{"render:ONE": 1, "clean:404": 1}},
				{uri: "/item/5", want: "404", counts: rtN{"render:404": 2, "clean:ONE": 1}},
			},
		},
		rtModelCase{
			name: "blank",
			uri:  "/item/2",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{one(b, r)}
			},
			initial: "",
			steps: []rtLocStep{
				{uri: "/item/3", want: "", quiet: true},
				{uri: "/zzz", want: "", quiet: true},
				{uri: "/item/1", want: "ONE"},
				{uri: "/item/1?tab=t", want: "", counts: rtN{"render:ONE": 1, "clean:ONE": 1}},
				{uri: "/item/4", want: "", quiet: true},
				{uri: "/item/1", want: "ONE", counts: rtN{"render:ONE": 2, "clean:ONE": 1}},
				{uri: "/zzz", want: "", counts: rtN{"render:ONE": 2, "clean:ONE": 2}},
			},
		},
		rtModelCase{
			name: "nil",
			uri:  "/x",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{one(b, r, RouteComp[rtPathItem](nil)), rtComp[Location](r, "404")}
			},
			initial: "404",
			steps: []rtLocStep{
				{uri: "/item/2", want: "", counts: rtN{"render:404": 1, "clean:404": 1}},
				{uri: "/item/3", want: "", quiet: true},
				{uri: "/item/1", want: "ONE"},
				{uri: "/item/2", want: "", counts: rtN{"render:ONE": 1, "clean:ONE": 1, "render:404": 1}},
				{uri: "/x", want: "404", counts: rtN{"render:404": 2}},
				{uri: "/zzz", want: "404", quiet: true},
			},
		},
		rtModelCase{
			name: "no-children",
			uri:  "/item/1",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{rtModel[rtPathItem](b), rtModel[rtPathA](b), rtComp[Location](r, "404")}
			},
			initial: "404",
			steps: []rtLocStep{
				{uri: "/item/2", want: "404", quiet: true},
				{uri: "/a", want: "404", quiet: true, counts: rtN{"render:404": 1, "clean:404": 0}},
			},
		},
		rtModelCase{
			name: "no-children-blank",
			uri:  "/item/1",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{rtModel[rtPathItem](b)}
			},
			initial: "",
			steps: []rtLocStep{
				{uri: "/item/2", want: "", quiet: true},
				{uri: "/zzz", want: "", quiet: true},
			},
		},
	)
}

// Leaving /docs for /settings re-selects from the top in one update, so notFound never shows on the way.
func TestRouteModelSettingsSection(t *testing.T) {
	sectionOf := func(l Location) string {
		if len(l.Segments) == 0 {
			return ""
		}
		return l.Segments[0]
	}
	sectionSet := func(l Location, s string) Location {
		l = l.Clone()
		l.Segments = append([]string{s}, l.Segments[min(1, len(l.Segments)):]...)
		return l
	}
	rtModelCases(t, rtModelCase{
		name: "settings",
		uri:  "/docs",
		routes: func(b bool, r *rtRec) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				rtModel[rtPathSettings](b, rtOnce[rtPathSettings](b, r, "settings")),
				rtDeriveTo(b, r, "section", sectionOf, sectionSet, rtEq,
					rtValue(b, "docs", rtOnce[string](b, r, "docs")),
					rtComp[string](r, "notFound"),
				),
			}
		},
		initial: "docs",
		steps: []rtLocStep{
			{uri: "/settings", want: "settings", counts: rtN{"render:settings": 1, "clean:docs": 1, "render:notFound": 0}},
			{uri: "/docs", want: "docs", counts: rtN{"render:docs": 2, "clean:settings": 1, "render:notFound": 0}},
			{uri: "/docs?x=1", want: "docs", quiet: true},
			{uri: "/settings?x=1", want: "settings", counts: rtN{"render:settings": 2, "clean:docs": 2, "render:notFound": 0}},
			{uri: "/other", want: "notFound", counts: rtN{"render:notFound": 1, "clean:settings": 2}},
			{uri: "/settings", want: "settings", counts: rtN{"render:settings": 3, "clean:notFound": 1}},
			{uri: "/other", want: "notFound", counts: rtN{"render:notFound": 2}},
			{uri: "/docs", want: "docs", counts: rtN{
				"render:docs": 3, "render:settings": 3, "render:notFound": 2, "clean:notFound": 2, "derive:section": 1,
			}},
		},
	})
}

// A query-only change that decodes into the same model runs no predicate under the model route.
func TestRouteModelDeriveSkip(t *testing.T) {
	rtModelCases(t, rtModelCase{
		name: "query",
		uri:  "/item/2",
		routes: func(b bool, r *rtRec) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				rtModel[rtPathItem](b,
					rtMatch(b, rtPred(r, "p", func(m rtPathItem) bool { return m.ID == 1 }), rtOnce[rtPathItem](b, r, "ONE")),
				),
				rtComp[Location](r, "404"),
			}
		},
		initial: "404",
		steps: []rtLocStep{
			{uri: "/item/2?zz=1", want: "404", quiet: true, counts: rtN{"pred:p={true 2 }": 1}},
			{uri: "/item/2?zz=2", want: "404", quiet: true, counts: rtN{"pred:p={true 2 }": 1}},
			{uri: "/item/1", want: "ONE", counts: rtN{"pred:p={true 1 }": 1, "render:ONE": 1}},
			{uri: "/item/1?zz=1", want: "ONE", quiet: true, counts: rtN{"pred:p={true 1 }": 1}},
			{uri: "/item/1?zz=2&tab=", want: "ONE", quiet: true, counts: rtN{"pred:p={true 1 }": 1}},
			{uri: "/item/1?tab=t", want: "ONE", quiet: true, counts: rtN{"pred:p={true 1 t}": 1, "render:ONE": 1}},
			{uri: "/item/1?tab=t&zz=1", want: "ONE", quiet: true, counts: rtN{"pred:p={true 1 t}": 1}},
			{uri: "/item/2", want: "404", counts: rtN{"pred:p={true 2 }": 3, "render:404": 2, "clean:ONE": 1}},
			{uri: "/item/2?zz=3", want: "404", quiet: true, counts: rtN{"pred:p={true 2 }": 3}},
		},
	})
}

// A derive over the location showing a leaf gives way to earlier models, at any index.
func TestRouteModelDeriveDefault(t *testing.T) {
	custom := func(b bool, rs ...SourceRoute[rtCustom]) SourceRoute[Location] {
		return rtDeriveTo(b, nil, "", rtCustomGet, rtCustomSet, rtEq, rs...)
	}
	rtModelCases(t,
		rtModelCase{
			name: "index0",
			uri:  "/item/x",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					rtModel[rtPathA](b, rtOnce[rtPathA](b, r, "A")),
					custom(b, rtBind[rtCustom](r, "C")),
					rtComp[Location](r, "404"),
				}
			},
			initial: "C={x }",
			steps: []rtLocStep{
				{uri: "/a", want: "A"},
				{uri: "/item/y", want: "C={y }"},
				{uri: "/zzz", want: "C={ }"},
				{uri: "/a", want: "A", counts: rtN{"render:A": 2, "render:C": 3, "render:404": 0}},
			},
		},
		rtModelCase{
			name: "index1",
			uri:  "/item/x",
			routes: func(b bool, r *rtRec) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					rtModel[rtPathItem](b, rtOnce[rtPathItem](b, r, "I")),
					custom(b,
						rtMatch(b, func(c rtCustom) bool { return c.Tab == "v" }, rtOnce[rtCustom](b, r, "V")),
						rtBind[rtCustom](r, "C"),
					),
					rtComp[Location](r, "404"),
				}
			},
			initial: "C={x }",
			steps: []rtLocStep{
				{uri: "/item/1", want: "I"},
				{uri: "/item/y", want: "C={y }"},
				{uri: "/item/y?tab=v", want: "V"},
				{uri: "/item/2?tab=v", want: "V", quiet: true},
				{uri: "/item/z", want: "C={z }"},
				{uri: "/item/3", want: "I", counts: rtN{"render:I": 2, "render:V": 1, "render:C": 3, "render:404": 0}},
			},
		},
	)
}

// Two prefix lenses deep on one route door; writes from the bottom restore both prefixes.
func TestRouteModelDeepLens(t *testing.T) {
	tree := func(b bool, r *rtRec, adminItem ...SourceRoute[rtPathItem]) []SourceRoute[Location] {
		if len(adminItem) == 0 {
			adminItem = []SourceRoute[rtPathItem]{rtBind[rtPathItem](r, "ADMIN-ITEM")}
		}
		return []SourceRoute[Location]{
			rtMatch(b, rtFirst("app"), rtStrip(b, r, "s1", "app",
				rtMatch(b, rtFirst("admin"), rtStrip(b, r, "s2", "admin",
					rtModel[rtPathItem](b, adminItem...),
					rtModel[rtPathA](b, rtOnce[rtPathA](b, r, "ADMIN-A")),
					rtComp[Location](r, "admin404"),
				)),
				rtModel[rtPathItem](b,
					rtValue(b, rtItem1, rtOnce[rtPathItem](b, r, "APP-ONE")),
					rtBind[rtPathItem](r, "APP-ITEM"),
				),
				rtModel[rtPathA](b, rtOnce[rtPathA](b, r, "APP-A")),
				rtComp[Location](r, "app404"),
			)),
			rtModel[rtPathA](b, rtOnce[rtPathA](b, r, "A")),
			rtComp[Location](r, "404"),
		}
	}
	rtModelCases(t, rtModelCase{
		name: "app-admin",
		uri:  "/app/item/1",
		routes: func(b bool, r *rtRec) []SourceRoute[Location] {
			return tree(b, r)
		},
		initial: "APP-ONE",
		steps: []rtLocStep{
			{uri: "/app/item/2", want: "APP-ITEM={true 2 }"},
			{uri: "/app/admin/item/3", want: "ADMIN-ITEM={true 3 }"},
			{uri: "/app/admin/a", want: "ADMIN-A"},
			{uri: "/app/admin/a?x=1", want: "ADMIN-A", quiet: true},
			{uri: "/app/admin/zzz", want: "admin404"},
			{uri: "/app/admin/item/4", want: "ADMIN-ITEM={true 4 }"},
			{uri: "/app/a", want: "APP-A"},
			{uri: "/app/item/1", want: "APP-ONE"},
			{uri: "/app/zzz", want: "app404"},
			{uri: "/app/yyy", want: "app404", quiet: true},
			{uri: "/app/admin", want: "admin404"},
			{uri: "/a", want: "A"},
			{uri: "/zzz", want: "404"},
			{uri: "/app/item/7", want: "APP-ITEM={true 7 }"},
			{uri: "/app/item/7?tab=x", want: "APP-ITEM={true 7 x}"},
			{uri: "/app/item/7?q=1&tab=x", want: "APP-ITEM={true 7 x}", quiet: true, counts: rtN{
				"derive:s1": 1, "derive:s2": 1, "render:APP-ONE": 2, "render:ADMIN-A": 1, "render:admin404": 2, "render:app404": 1,
				"render:A": 1, "render:404": 1, "render:ADMIN-ITEM": 2, "render:APP-ITEM": 3, "render:APP-A": 1,
			}},
		},
	})
	t.Run("write", func(t *testing.T) {
		h := newRT(t)
		loc := h.at("/app/admin/item/3")
		box := &rtBox[Source[rtPathItem]]{}
		h.render(rtSlot("m", Route(tree(false, h.rec, RouteDerive(rtCapture(h.rec, "cap", box), rtBind[rtPathItem](h.rec, "ADMIN-ITEM")))...)))
		m := func(s string) map[string]string { return map[string]string{"m": s} }
		h.ops(loc,
			rtOp{at: "initial", want: m("ADMIN-ITEM={true 3 }")},
			rtOp{at: "write", do: rtPut(box, rtPathItem{Item: true, ID: 9, Tab: "t"}), ret: []error{nil}, want: m("ADMIN-ITEM={true 9 t}"), loc: "/app/admin/item/9?tab=t"},
			rtOp{at: "equal", do: rtPut(box, rtPathItem{Item: true, ID: 9, Tab: "t"}), ret: []error{}, quiet: []string{"m"}, want: m("ADMIN-ITEM={true 9 t}")},
			rtOp{at: "leave", do: rtGo(loc, "/app/item/1"), want: m("APP-ONE")},
			rtOp{at: "write back in", do: rtPut(box, rtPathItem{Item: true, ID: 2, Tab: "u"}), ret: []error{nil},
				want: m("ADMIN-ITEM={true 2 u}"), loc: "/app/admin/item/2?tab=u",
				counts: rtN{"derive:cap": 1, "derive:s1": 1, "derive:s2": 1, "render:ADMIN-ITEM": 3}},
		)
		rtWantPaths(h, rtSP("/app/admin/item/9?tab=t", false), rtSP("/app/item/1", false), rtSP("/app/admin/item/2?tab=u", false))
	})
}

// rtOp is one step over several slots; quiet slots must get no route door action.
type rtOp struct {
	at     string
	do     func(ctx context.Context) <-chan error
	ret    []error
	want   map[string]string
	quiet  []string
	counts map[string]int
	loc    string
}

func (h *rtHarness) ops(loc Beam[Location], steps ...rtOp) {
	h.t.Helper()
	for i, st := range steps {
		at := fmt.Sprintf("step %d (%s): ", i, st.at)
		if st.do != nil {
			var ret []error
			h.quietAll(at, st.quiet, func() {
				var ch <-chan error
				h.settle(func(ctx context.Context) { ch = st.do(ctx) })
				ret = drainOp(h.t, ch)
			})
			if st.ret != nil && !rtSameErrs(ret, st.ret) {
				h.t.Fatalf("%supdate returned %v, want %v", at, ret, st.ret)
			}
		}
		h.expectAll(at, st.want)
		if st.counts != nil {
			h.countsAt(at, st.counts)
		}
		if st.loc != "" {
			if got, want := loc.Get(), rtLoc(st.loc); !path.EqualLocation(got, want) {
				h.t.Fatalf("%slocation %q, want %q", at, got.String(), want.String())
			}
		}
		if h.strict {
			for name := range st.want {
				h.oneDoorAt(at, name)
			}
		}
	}
}

func rtGo(s Source[Location], uri string) func(ctx context.Context) <-chan error {
	return func(ctx context.Context) <-chan error { return s.Update(ctx, rtLoc(uri)) }
}

func rtPut[T any](box *rtBox[Source[T]], v T) func(ctx context.Context) <-chan error {
	return func(ctx context.Context) <-chan error { return box.last().Update(ctx, v) }
}

func rtCapture[T any](r *rtRec, label string, box *rtBox[Source[T]]) func(Source[T]) Source[T] {
	return func(s Source[T]) Source[T] {
		r.hit("derive:" + label)
		box.put(s)
		return s
	}
}

// Writing the model Source re-encodes the URL, keeps the branch and honors HistoryReplaceContext.
func TestRouteModelWriteBack(t *testing.T) {
	h := newRT(t)
	loc := h.at("/item/3?tab=x")
	box, sbox := &rtBox[Source[rtPathItem]]{}, &rtBox[Source[rtPathItem]]{}
	h.render(
		rtSlot("m", Route(
			RouteModel[rtPathItem](RouteDerive(rtCapture(h.rec, "cap", box),
				RouteValue(rtItem1, rtComp[rtPathItem](h.rec, "ONE")),
				rtBind[rtPathItem](h.rec, "B"),
			)),
			rtComp[Location](h.rec, "404"),
		)),
		rtSlot("s", Route(RouteModel[rtPathItem](rtSource(h.rec, "S", sbox)))),
	)
	item := func(id int, tab string) rtPathItem { return rtPathItem{Item: true, ID: id, Tab: tab} }
	w := func(m string) map[string]string { return map[string]string{"m": m, "s": "S"} }
	h.ops(loc,
		rtOp{at: "initial", want: w("B={true 3 x}"), counts: rtN{"derive:cap": 1}},
		rtOp{at: "write", do: rtPut(box, item(7, "y")), ret: []error{nil}, quiet: []string{"s"}, want: w("B={true 7 y}"), loc: "/item/7?tab=y"},
		rtOp{at: "equal", do: rtPut(box, item(7, "y")), ret: []error{}, quiet: []string{"m", "s"}, want: w("B={true 7 y}"), loc: "/item/7?tab=y"},
		rtOp{at: "value", do: rtPut(box, rtItem1), ret: []error{nil}, quiet: []string{"s"}, want: w("ONE"), loc: "/item/1?tab="},
		rtOp{at: "leaf", do: rtPut(sbox, item(2, "")), ret: []error{nil}, quiet: []string{"s"}, want: w("B={true 2 }"), loc: "/item/2?tab="},
		rtOp{at: "replace", do: func(ctx context.Context) <-chan error {
			return box.last().Update(HistoryReplaceContext(ctx), item(2, "r"))
		}, ret: []error{nil}, quiet: []string{"s"}, want: w("B={true 2 r}"), loc: "/item/2?tab=r"},
		rtOp{at: "mutate", do: func(ctx context.Context) <-chan error {
			return box.last().Mutate(HistoryReplaceContext(ctx), func(m rtPathItem) rtPathItem { m.Tab = "z"; return m })
		}, ret: []error{nil}, quiet: []string{"s"}, want: w("B={true 2 z}"), loc: "/item/2?tab=z",
			counts: rtN{"derive:cap": 1, "render:S": 1, "render:ONE": 1, "clean:ONE": 1, "render:B": 5, "clean:B": 4}},
		rtOp{at: "leave", do: rtGo(loc, "/zzz"), want: map[string]string{"m": "404", "s": ""}},
		rtOp{at: "inactive", do: rtPut(box, item(5, "")), ret: []error{nil}, want: w("B={true 5 }"), loc: "/item/5?tab="},
	)
	if box.len() != 1 {
		t.Fatalf("expected the derive to run once, got %d", box.len())
	}
	rtWantPaths(h, rtSP("/item/7?tab=y", false), rtSP("/item/1?tab=", false), rtSP("/item/2?tab=", false),
		rtSP("/item/2?tab=r", true), rtSP("/item/2?tab=z", true), rtSP("/zzz", false), rtSP("/item/5?tab=", false))
}

// A model that cannot be encoded leaves the URL alone; escaped segments decode and encode.
func TestRouteModelEncodeError(t *testing.T) {
	h := newRT(t)
	loc := h.at("/item/hello%20world%2Fagain")
	nbox, cbox := &rtBox[Source[rtPathName]]{}, &rtBox[Source[rtCustom]]{}
	h.render(
		rtSlot("n", Route(RouteModel[rtPathName](RouteDerive(rtCapture(h.rec, "n", nbox), rtBind[rtPathName](h.rec, "N"))))),
		rtSlot("c", Route(RouteMatch(func(l Location) bool { return rtCustomGet(l).ID != "" },
			RouteDerive(func(s Source[Location]) Source[rtCustom] { return DeriveSourceEqual(s, rtCustomGet, rtCustomSet, rtEq) },
				RouteDerive(rtCapture(h.rec, "c", cbox), rtBind[rtCustom](h.rec, "C")),
			),
		))),
	)
	both := []string{"n", "c"}
	same := map[string]string{"n": "N={true hello world/again}", "c": "C={hello world/again }"}
	h.ops(loc,
		rtOp{at: "initial", want: same},
		rtOp{at: "empty name", do: rtPut(nbox, rtPathName{Item: true}), ret: []error{}, quiet: both, want: same, loc: "/item/hello%20world%2Fagain"},
		rtOp{at: "zero model", do: rtPut(nbox, rtPathName{}), ret: []error{}, quiet: both, want: same},
		rtOp{at: "empty custom", do: rtPut(cbox, rtCustom{Tab: "t"}), ret: []error{}, quiet: both, want: same},
		rtOp{at: "custom", do: rtPut(cbox, rtCustom{ID: "z", Tab: "t"}), ret: []error{nil},
			want: map[string]string{"n": "N={true z}", "c": "C={z t}"}, loc: "/item/z?tab=t"},
		rtOp{at: "name", do: rtPut(nbox, rtPathName{Item: true, Name: "next value/again"}), ret: []error{nil},
			want: map[string]string{"n": "N={true next value/again}", "c": "C={next value/again }"}, loc: "/item/next%20value%2Fagain",
			counts: rtN{"render:N": 3, "render:C": 3}},
	)
	rtWantPaths(h, rtSP("/item/z?tab=t", false), rtSP("/item/next%20value%2Fagain", false))
}

// Location routes use the location comparator, so an equal URL runs no predicate.
func TestRouteModelRawLocation(t *testing.T) {
	isSeg := func(l Location) bool { return len(l.Segments) == 2 && l.Segments[0] == "seg" }
	h := newRT(t)
	loc := h.at("/raw")
	box := &rtBox[Source[Location]]{}
	h.render(rtSlot("m", Route(
		RouteValue(rtLoc("/raw"), rtSource(h.rec, "RAW", box)),
		RouteMatch(rtPred(h.rec, "seg", isSeg), rtBind[Location](h.rec, "SEG")),
		rtComp[Location](h.rec, "404"),
	)))
	m := func(s string) map[string]string { return map[string]string{"m": s} }
	h.ops(loc,
		rtOp{at: "initial", want: m("RAW")},
		rtOp{at: "write", do: rtPut(box, rtLoc("/seg/x")), ret: []error{nil}, want: m("SEG=/seg/x"), loc: "/seg/x"},
		rtOp{at: "/seg/y", do: rtGo(loc, "/seg/y"), want: m("SEG=/seg/y")},
		rtOp{at: "query", do: rtGo(loc, "/seg/y?b=2&a=1"), want: m("SEG=/seg/y?a=1&b=2")},
		rtOp{at: "equal", do: rtGo(loc, "/seg/y?a=1&b=2"), ret: []error{}, quiet: []string{"m"}, want: m("SEG=/seg/y?a=1&b=2")},
		rtOp{at: "/raw?x=1", do: rtGo(loc, "/raw?x=1"), want: m("404")},
		rtOp{at: "/raw", do: rtGo(loc, "/raw"), want: m("RAW")},
		rtOp{at: "/raw/", do: rtGo(loc, "/raw/"), ret: []error{}, quiet: []string{"m"}, want: m("RAW"),
			counts: rtN{"render:RAW": 2, "render:SEG": 3, "render:404": 1}},
	)
	if got, want := h.rec.vals("pred:seg="), []string{"/seg/x", "/seg/y", "/seg/y?a=1&b=2", "/raw?x=1", "/raw?x=1"}; !slices.Equal(got, want) {
		t.Fatalf("predicate log %v, want %v", got, want)
	}
	rtWantPaths(h, rtSP("/seg/x", false), rtSP("/seg/y", false), rtSP("/seg/y?a=1&b=2", false), rtSP("/raw?x=1", false), rtSP("/raw", false))
}

// Models, a custom location type, a query model and a raw match in one container, in every flavor.
func TestRouteModelCombined(t *testing.T) {
	isItem := func(l Location) bool { return rtCustomGet(l).ID != "" }
	isRaw := func(l Location) bool { return len(l.Segments) == 1 && l.Segments[0] == "raw" }
	custom := func(s Source[Location]) Source[rtCustom] { return DeriveSourceEqual(s, rtCustomGet, rtCustomSet, rtEq) }
	h := newRT(t)
	r := h.rec
	loc := h.at("/a")
	cbox, qbox := &rtBox[Source[rtCustom]]{}, &rtBox[Source[rtPathQuery]]{}
	bbox := &rtBox[Beam[rtCustom]]{}
	h.render(
		rtSlot("s", Route(
			RouteModel[rtPathA](rtSource[rtPathA](r, "A", nil)),
			RouteMatch(isItem, RouteDerive(custom, rtSource(r, "C", cbox))),
			RouteModel[rtPathQuery](rtSource(r, "Q", qbox)),
			RouteMatch(isRaw, rtSource[Location](r, "R", nil)),
			rtComp[Location](r, "D"),
		)),
		rtSlot("b", Route(
			RouteModel[rtPathA](rtBind[rtPathA](r, "a")),
			RouteMatch(isItem, RouteDerive(custom, rtBind[rtCustom](r, "c"))),
			RouteModel[rtPathQuery](rtBind[rtPathQuery](r, "q")),
			RouteMatch(isRaw, rtBind[Location](r, "r")),
			rtComp[Location](r, "d"),
		)),
		rtSlot("beam", func(cur gox.Cursor) error {
			return cur.Comp(Router(cur.Context()).RouteBeam(
				RouteBeamModel[rtPathA](rtBeam[rtPathA](r, "bA", nil)),
				RouteBeamMatch(isItem, RouteBeamDerive(func(b Beam[Location]) Beam[rtCustom] { return DeriveBeamEqual(b, rtCustomGet, rtEq) },
					rtBeam(r, "bC", bbox),
				)),
				RouteBeamModel[rtPathQuery](rtBind[rtPathQuery](r, "bq")),
				rtComp[Location](r, "bD"),
			))
		}),
	)
	w := func(s, b, beam string) map[string]string { return map[string]string{"s": s, "b": b, "beam": beam} }
	h.ops(loc,
		rtOp{at: "initial", want: w("A", "a={true}", "bA")},
		rtOp{at: "custom", do: rtGo(loc, "/item/hello%20world%2Fone?tab=details"),
			want: w("C", "c={hello world/one details}", "bC"), loc: "/item/hello%20world%2Fone?tab=details"},
		rtOp{at: "custom next", do: rtGo(loc, "/item/next%2Fchild?tab=again"), quiet: []string{"s", "beam"},
			want: w("C", "c={next/child again}", "bC")},
		rtOp{at: "custom write", do: rtPut(cbox, rtCustom{ID: "lens write/value", Tab: "written"}), ret: []error{nil}, quiet: []string{"s", "beam"},
			want: w("C", "c={lens write/value written}", "bC"), loc: "/item/lens%20write%2Fvalue?tab=written"},
	)
	if got := bbox.last().Get(); got != (rtCustom{ID: "lens write/value", Tab: "written"}) {
		t.Fatalf("custom beam holds %+v", got)
	}
	h.ops(loc,
		rtOp{at: "query", do: rtGo(loc, "/q?tag=next&page=2"), want: w("Q", "q=next|2", "bq=next|2")},
		rtOp{at: "query equal", do: rtGo(loc, "/q?tag=next&page=2&x=1"), quiet: []string{"s", "b", "beam"}, want: w("Q", "q=next|2", "bq=next|2")},
		rtOp{at: "query empty", do: rtGo(loc, "/q"), quiet: []string{"s"}, want: w("Q", "q=|", "bq=|")},
		rtOp{at: "raw", do: rtGo(loc, "/raw?from=query"), want: w("R", "r=/raw?from=query", "bD")},
		rtOp{at: "model", do: rtGo(loc, "/a"), want: w("A", "a={true}", "bA")},
		rtOp{at: "missing", do: rtGo(loc, "/missing"), want: w("D", "d", "bD")},
		rtOp{at: "back", do: rtGo(loc, "/a"), want: w("A", "a={true}", "bA"), counts: rtN{
			"render:A": 3, "render:C": 1, "render:Q": 1, "render:R": 1, "render:D": 1,
			"render:a": 3, "render:c": 3, "render:q": 2, "render:r": 1, "render:d": 1,
			"render:bA": 3, "render:bC": 1, "render:bq": 2, "render:bD": 2,
		}},
	)
	if m := qbox.last().Get(); m.Tag != nil || m.Page != nil {
		t.Fatalf("query model should have nil fields, got %+v", m)
	}
	rtWantPaths(h, rtSP("/item/hello%20world%2Fone?tab=details", false), rtSP("/item/next%2Fchild?tab=again", false),
		rtSP("/item/lens%20write%2Fvalue?tab=written", false), rtSP("/q?page=2&tag=next", false), rtSP("/q?page=2&tag=next&x=1", false),
		rtSP("/q", false), rtSP("/raw?from=query", false), rtSP("/a", false), rtSP("/missing", false), rtSP("/a", false))
}

// Model routes behave the same through doors.Route, Router(ctx).Route, derived entries and Beam leaves.
func TestRouteModelEntries(t *testing.T) {
	t.Run("router", func(t *testing.T) {
		h := newRT(t)
		loc := h.at("/item/1")
		tree := func(p string) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				RouteModel[rtPathA](rtSource[rtPathA](h.rec, p+"A", nil)),
				RouteModel[rtPathItem](rtBind[rtPathItem](h.rec, p+"I")),
				rtComp[Location](h.rec, p+"404"),
			}
		}
		h.render(
			rtSlot("before", textElem("BEFORE")),
			rtSlot("pkg", Route(tree("p")...)),
			rtSlot("elem", func(cur gox.Cursor) error { return cur.Comp(Router(cur.Context()).Route(tree("e")...)) }),
			rtSlot("after", textElem("AFTER")),
		)
		if acts := h.client.actions(); len(acts) != 0 {
			t.Fatalf("the first selection sent actions %+v", acts)
		}
		steps := []rtOp{{at: "initial", want: map[string]string{"pkg": "pI={true 1 }", "elem": "eI={true 1 }"}}}
		for _, st := range []struct{ uri, want string }{
			{"/a", "A"}, {"/item/2?tab=q", "I={true 2 q}"}, {"/zzz", "404"}, {"/a?x=1", "A"}, {"/item/3", "I={true 3 }"},
		} {
			steps = append(steps, rtOp{at: st.uri, do: rtGo(loc, st.uri), want: map[string]string{"pkg": "p" + st.want, "elem": "e" + st.want}})
		}
		h.ops(loc, steps...)
		h.expectAll("", map[string]string{"before": "BEFORE", "after": "AFTER"})
		for _, k := range []string{"A", "I", "404"} {
			for _, ev := range []string{"render:", "clean:"} {
				if p, e := h.rec.count(ev+"p"+k), h.rec.count(ev+"e"+k); p != e {
					t.Fatalf("%s%s: doors.Route %d, Router(ctx).Route %d", ev, k, p, e)
				}
			}
		}
	})
	t.Run("derived", func(t *testing.T) {
		h := newRT(t)
		locP, locR := rtLocSource("/app/item/3"), rtLocSource("/item/3")
		box := &rtBox[Source[rtPathItem]]{}
		strip := DeriveSourceEqual(locP, rtStripGet, func(_ Location, l Location) Location {
			l = l.Clone()
			l.Segments = append([]string{"app"}, l.Segments...)
			return l
		}, path.EqualLocation)
		tree := func(r *rtRec, box *rtBox[Source[rtPathItem]]) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				RouteModel[rtPathA](rtSource[rtPathA](r, "A", nil)),
				RouteModel[rtPathItem](RouteDerive(rtCapture(r, "cap", box), rtBind[rtPathItem](r, "ITEM"))),
				rtComp[Location](r, "404"),
			}
		}
		h.render(
			rtSlot("strip", strip.Route(tree(h.rec, box)...)),
			rtSlot("stripbeam", DeriveBeamEqual(locP, rtStripGet, path.EqualLocation).RouteBeam(
				RouteBeamModel[rtPathA](rtBeam[rtPathA](nil, "A", nil)),
				RouteBeamModel[rtPathItem](rtBind[rtPathItem](nil, "ITEM")),
				rtComp[Location](nil, "404"),
			)),
			rtSlot("root", locR.Route(tree(nil, nil)...)),
		)
		slots := []string{"strip", "stripbeam", "root"}
		for _, st := range []struct{ rest, want string }{
			{"", "ITEM={true 3 }"}, {"/a", "A"}, {"/item/5?tab=t", "ITEM={true 5 t}"}, {"/zzz", "404"}, {"/item/6", "ITEM={true 6 }"},
		} {
			if st.rest != "" {
				h.settle(func(ctx context.Context) {
					locP.Update(ctx, rtLoc("/app"+st.rest))
					locR.Update(ctx, rtLoc(st.rest))
				})
			}
			for _, s := range slots {
				h.expectAt(st.rest+": ", s, st.want)
				h.oneDoorAt(st.rest+": ", s)
			}
		}
		if ret := rtSet(h, box.last(), rtPathItem{Item: true, ID: 9}); !rtSameErrs(ret, []error{nil}) {
			t.Fatalf("write through the derived entry returned %v", ret)
		}
		rtLocIs(h, locP, "/app/item/9?tab=")
		h.expectAll("write: ", map[string]string{"strip": "ITEM={true 9 }", "stripbeam": "ITEM={true 9 }", "root": "ITEM={true 6 }"})
		h.counts(rtN{"derive:cap": 1, "render:A": 1})
	})
	t.Run("beam", func(t *testing.T) {
		h := newRT(t)
		loc := rtLocSource("/item/5?tab=x")
		box, dbox := &rtBox[Beam[rtPathItem]]{}, &rtBox[Beam[rtPathItem]]{}
		h.render(
			rtSlot("a", loc.RouteBeam(RouteBeamModel[rtPathItem](rtBeam(h.rec, "I", box)))),
			rtSlot("b", DeriveBeamEqual(loc, func(l Location) Location { return l }, path.EqualLocation).RouteBeam(
				RouteBeamModel[rtPathItem](rtBeam(h.rec, "J", dbox)),
			)),
		)
		holds := func(at string, want rtPathItem) {
			t.Helper()
			for _, b := range []*rtBox[Beam[rtPathItem]]{box, dbox} {
				if _, ok := b.last().(Source[rtPathItem]); ok {
					t.Fatalf("%sRouteBeamModel leaf got a writable Source", at)
				}
				if got := b.last().Get(); got != want {
					t.Fatalf("%sbeam holds %+v, want %+v", at, got, want)
				}
			}
		}
		ab := func(s string) map[string]string { return map[string]string{"a": s + "I", "b": s + "J"} }
		holds("initial: ", rtPathItem{Item: true, ID: 5, Tab: "x"})
		h.ops(loc,
			rtOp{at: "initial", want: ab("")},
			rtOp{at: "/item/6", do: rtGo(loc, "/item/6"), quiet: []string{"a", "b"}, want: ab("")},
		)
		holds("/item/6: ", rtPathItem{Item: true, ID: 6})
		h.ops(loc,
			rtOp{at: "/zzz", do: rtGo(loc, "/zzz"), want: map[string]string{"a": "", "b": ""}},
			rtOp{at: "/item/7", do: rtGo(loc, "/item/7"), want: ab(""), counts: rtN{"render:I": 2, "render:J": 2, "clean:I": 1, "clean:J": 1}},
		)
		holds("/item/7: ", rtPathItem{Item: true, ID: 7})
		if box.len() != 2 || dbox.len() != 2 {
			t.Fatalf("expected one capture per activation, got %d and %d", box.len(), dbox.len())
		}
	})
}

// The oracle is an independent model of the routing rules, run on generated trees.

const (
	omInt = iota
	omStr
	omPair
)

type omKV struct {
	K, V int
}

const (
	omOpComp = iota
	omOpBind
	omOpMatch
	omOpValue
	omOpDerive
)

const (
	omLeafComp = iota
	omLeafSource
	omLeafBeam
	omLeafNil
)

type omFn struct {
	name    string
	in, out int
	f       func(any) any
}

var omFns = []omFn{
	{"/10", omInt, omInt, func(v any) any { return v.(int) / 10 }},
	{"%10", omInt, omInt, func(v any) any { return v.(int) % 10 }},
	{"%4", omInt, omInt, func(v any) any { return v.(int) % 4 }},
	{"/7", omInt, omInt, func(v any) any { return v.(int) / 7 }},
	{"/2", omInt, omInt, func(v any) any { return v.(int) / 2 }},
	{"parity", omInt, omStr, func(v any) any {
		if v.(int)%2 == 0 {
			return "even"
		}
		return "odd"
	}},
	{"digit", omInt, omStr, func(v any) any { return strconv.Itoa(v.(int) % 10) }},
	{"band", omInt, omStr, func(v any) any {
		switch x := v.(int); {
		case x < 20:
			return "lo"
		case x < 40:
			return "mid"
		}
		return "hi"
	}},
	{"itoa", omInt, omStr, func(v any) any { return strconv.Itoa(v.(int)) }},
	{"kv/10", omInt, omPair, func(v any) any { return omKV{v.(int) / 10, v.(int) % 10} }},
	{"kv%3", omInt, omPair, func(v any) any { return omKV{v.(int) % 3, v.(int) / 3 % 5} }},
	{"len", omStr, omInt, func(v any) any { return len(v.(string)) }},
	{"atoi", omStr, omInt, func(v any) any {
		n, err := strconv.Atoi(v.(string))
		if err != nil {
			return -1
		}
		return n
	}},
	{"c0", omStr, omInt, func(v any) any { return int(v.(string)[0]) % 5 }},
	{"first", omStr, omStr, func(v any) any { return v.(string)[:1] }},
	{"upper", omStr, omStr, func(v any) any { return strings.ToUpper(v.(string)) }},
	{"rev", omStr, omStr, func(v any) any {
		b := []byte(v.(string))
		slices.Reverse(b)
		return string(b)
	}},
	{"K", omPair, omInt, func(v any) any { return v.(omKV).K }},
	{"V", omPair, omInt, func(v any) any { return v.(omKV).V }},
	{"K+V", omPair, omInt, func(v any) any { return v.(omKV).K + v.(omKV).V }},
	{"swap", omPair, omPair, func(v any) any { return omKV{v.(omKV).V, v.(omKV).K} }},
	{"kv", omPair, omStr, func(v any) any { return fmt.Sprintf("%d:%d", v.(omKV).K, v.(omKV).V) }},
}

type omEq struct {
	name string
	f    func(a, b any) bool
}

func (e omEq) equal(a, b any) bool { return e.f != nil && e.f(a, b) }

var omExact = omEq{"==", func(a, b any) bool { return a == b }}

var omCoarse = [...][]omEq{
	omInt: {
		{"~%2", func(a, b any) bool { return a.(int)%2 == b.(int)%2 }},
		{"~/3", func(a, b any) bool { return a.(int)/3 == b.(int)/3 }},
	},
	omStr: {
		{"~len", func(a, b any) bool { return len(a.(string)) == len(b.(string)) }},
		{"~c0", func(a, b any) bool { return a.(string)[0] == b.(string)[0] }},
	},
	omPair: {
		{"~K", func(a, b any) bool { return a.(omKV).K == b.(omKV).K }},
		{"~V", func(a, b any) bool { return a.(omKV).V == b.(omKV).V }},
	},
}

type omNode struct {
	op    int
	label string
	kind  int
	beam  bool
	pname string
	pred  func(any) bool
	sname string
	show  func(any) bool
	val   any
	fn    omFn
	eq    omEq
	out   omEq
	kids  []*omNode

	cur      int
	dflt     bool
	stored   any
	last     any
	none     bool
	renders  int
	contents int
}

type omStats struct {
	switches, breaks, skips, cached, falls, pairSkips int
}

func (n *omNode) isLeaf() bool { return n.op == omOpComp || n.op == omOpBind }

func (n *omNode) isNil() bool { return n.op == omOpComp && n.kind == omLeafNil }

func (n *omNode) content() bool {
	return n.op == omOpComp && !n.isNil() || n.op == omOpBind && (n.show == nil || n.show(n.stored))
}

func (n *omNode) text() string {
	switch {
	case !n.content():
		return ""
	case n.op == omOpBind:
		return n.label + "=" + fmt.Sprint(n.stored)
	}
	return n.label
}

func (n *omNode) holds(v any) bool {
	if n.op == omOpValue {
		return n.eq.equal(n.val, v)
	}
	return n.pred(v)
}

// seek reports whether n matches v and whether it then counts as a leaf for its parent.
func (n *omNode) seek(v any, st *omStats) (ok, dflt bool) {
	switch n.op {
	case omOpComp, omOpBind:
		n.stored = v
		return true, true
	case omOpMatch, omOpValue:
		if !n.holds(v) {
			return false, false
		}
		if n.cur, n.dflt = omSeek(n.kids, v, st); n.cur < 0 {
			st.falls++
			return false, false
		}
		return true, false
	}
	w := n.fn.f(v)
	if n.none && n.out.equal(n.last, w) {
		if w != n.last {
			st.cached++
		}
		return false, false
	}
	n.cur, n.dflt = omSeek(n.kids, w, st)
	n.last, n.none = w, n.cur < 0
	return !n.none, n.dflt
}

func omSeek(kids []*omNode, v any, st *omStats) (int, bool) {
	for i, k := range kids {
		if ok, dflt := k.seek(v, st); ok {
			return i, dflt
		}
	}
	return -1, false
}

func (n *omNode) leave() {
	if n.cur >= 0 {
		n.kids[n.cur].leave()
	}
	n.cur, n.dflt, n.stored, n.last, n.none = -1, false, nil, nil, false
}

const (
	omIntact = iota
	omShown
	omBroken
)

// omKeep walks the active path: leaves re-check earlier routes, failed conditions break, equal derives skip.
func omKeep(kids []*omNode, cur *int, dflt *bool, v any, st *omStats) int {
	var up []*bool
	for {
		if *dflt {
			for j := range *cur {
				if ok, _ := kids[j].seek(v, st); ok {
					kids[*cur].leave()
					*cur, *dflt = j, false
					for _, f := range up {
						*f = false
					}
					st.switches++
					return omShown
				}
			}
		}
		a := kids[*cur]
		switch a.op {
		case omOpComp:
			return omIntact
		case omOpBind:
			if a.eq.equal(a.stored, v) {
				return omIntact
			}
			a.stored = v
			return omShown
		case omOpMatch, omOpValue:
			if !a.holds(v) {
				return omBroken
			}
			up = up[:0]
		default:
			w := a.fn.f(v)
			if a.out.equal(a.last, w) {
				if w != a.last {
					st.skips++
					if _, ok := w.(omKV); ok {
						st.pairSkips++
					}
				}
				return omIntact
			}
			a.last, v = w, w
			up = append(up, dflt)
		}
		kids, cur, dflt = a.kids, &a.cur, &a.dflt
	}
}

func (n *omNode) leaf() *omNode {
	for !n.isLeaf() {
		n = n.kids[n.cur]
	}
	return n
}

func (n *omNode) walk(f func(*omNode)) {
	f(n)
	for _, k := range n.kids {
		k.walk(f)
	}
}

func (n *omNode) describe(b *strings.Builder) {
	switch n.op {
	case omOpComp:
		fmt.Fprintf(b, "%s(%s)", n.label, [...]string{"comp", "source", "beam", "nil"}[n.kind])
		return
	case omOpBind:
		b.WriteString(n.label)
		if n.show != nil {
			fmt.Fprintf(b, "(shown %s)", n.sname)
		}
		return
	}
	if n.beam {
		b.WriteString("b:")
	}
	switch n.op {
	case omOpMatch:
		fmt.Fprintf(b, "%s{%s}", n.label, n.pname)
	case omOpValue:
		fmt.Fprintf(b, "%s{=%#v}", n.label, n.val)
	default:
		fmt.Fprintf(b, "%s{%s %s}", n.label, n.fn.name, n.out.name)
	}
	b.WriteString("[")
	for i, k := range n.kids {
		if i > 0 {
			b.WriteString(", ")
		}
		k.describe(b)
	}
	b.WriteString("]")
}

type omTree struct {
	slot  string
	entry int
	top   []*omNode
	on    bool
	cur   int
	dflt  bool
	blank bool
	acts  int
	views map[string]bool
	st    omStats
}

// apply runs one update: keep the intact path, or clean and select from the top on a break.
func (tr *omTree) apply(v any, first bool) {
	if tr.on {
		switch omKeep(tr.top, &tr.cur, &tr.dflt, v, &tr.st) {
		case omIntact:
			tr.note()
			return
		case omBroken:
			tr.st.breaks++
			tr.top[tr.cur].leave()
			tr.on = false
		}
	}
	if !tr.on {
		tr.cur, tr.dflt = omSeek(tr.top, v, &tr.st)
		tr.on = tr.cur >= 0
	}
	tr.show(first)
	tr.note()
}

func (tr *omTree) show(first bool) {
	var leaf *omNode
	if tr.on {
		leaf = tr.top[tr.cur].leaf()
	}
	blank := leaf == nil || leaf.isNil()
	if blank && tr.blank {
		return
	}
	if !first {
		tr.acts++
	}
	tr.blank = blank
	if blank {
		return
	}
	leaf.renders++
	if leaf.content() {
		leaf.contents++
	}
}

func (tr *omTree) note() {
	_, key := tr.view()
	tr.views[key] = true
}

func (tr *omTree) view() (text, key string) {
	if !tr.on {
		return "", "-"
	}
	n := tr.top[tr.cur]
	for !n.isLeaf() {
		key += n.label + "/"
		n = n.kids[n.cur]
	}
	if n.isNil() {
		return "", key + n.label + "(nil)"
	}
	return n.text(), key + n.label + ":" + n.text()
}

func (tr *omTree) describe() string {
	var b strings.Builder
	fmt.Fprintf(&b, "entry %d: ", tr.entry)
	for i, n := range tr.top {
		if i > 0 {
			b.WriteString(", ")
		}
		n.describe(&b)
	}
	return b.String()
}

type omGen struct {
	rng   *rand.Rand
	tree  int
	n     int
	depth int
}

func (g *omGen) label(p string) string {
	g.n++
	return fmt.Sprintf("%s%d_%d", p, g.tree, g.n)
}

func (g *omGen) pick(dom []any) any {
	return dom[g.rng.IntN(len(dom))]
}

func (g *omGen) predicate(typ int, dom []any) (string, func(any) bool) {
	switch r := g.rng.IntN(100); {
	case r < 5:
		return "true", func(any) bool { return true }
	case r < 9:
		return "false", func(any) bool { return false }
	}
	switch typ {
	case omInt:
		k := g.pick(dom).(int)
		switch g.rng.IntN(8) {
		case 0, 6:
			return fmt.Sprintf("<%d", k), func(v any) bool { return v.(int) < k }
		case 1:
			return fmt.Sprintf(">=%d", k), func(v any) bool { return v.(int) >= k }
		case 2:
			j := g.pick(dom).(int)
			lo, hi := min(k, j), max(k, j)
			return fmt.Sprintf("%d..%d", lo, hi), func(v any) bool { return v.(int) >= lo && v.(int) <= hi }
		case 3, 4, 5:
			m := 2 + g.rng.IntN(3)
			rem := g.rng.IntN(m)
			return fmt.Sprintf("%%%d==%d", m, rem), func(v any) bool { return v.(int)%m == rem }
		}
		return fmt.Sprintf("==%d", k), func(v any) bool { return v.(int) == k }
	case omPair:
		p := g.pick(dom).(omKV)
		switch g.rng.IntN(4) {
		case 0:
			return fmt.Sprintf("K==%d", p.K), func(v any) bool { return v.(omKV).K == p.K }
		case 1:
			return fmt.Sprintf("V<%d", p.V), func(v any) bool { return v.(omKV).V < p.V }
		case 2:
			m := 2 + g.rng.IntN(2)
			rem := g.rng.IntN(m)
			return fmt.Sprintf("V%%%d==%d", m, rem), func(v any) bool { return v.(omKV).V%m == rem }
		}
		return fmt.Sprintf("==%v", p), func(v any) bool { return v.(omKV) == p }
	}
	s := g.pick(dom).(string)
	switch g.rng.IntN(4) {
	case 0:
		return fmt.Sprintf("==%q", s), func(v any) bool { return v.(string) == s }
	case 1:
		p := s[:1]
		return fmt.Sprintf("prefix%q", p), func(v any) bool { return strings.HasPrefix(v.(string), p) }
	case 2:
		l := len(s)
		return fmt.Sprintf("len==%d", l), func(v any) bool { return len(v.(string)) == l }
	}
	return fmt.Sprintf("<%q", s), func(v any) bool { return v.(string) < s }
}

func omMapDom(dom []any, f func(any) any) []any {
	seen := map[any]bool{}
	var out []any
	for _, v := range dom {
		if w := f(v); !seen[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	return out
}

func (g *omGen) group(depth, typ int, dom []any, beam bool, eq omEq, lo, hi int) []*omNode {
	kids := make([]*omNode, lo+g.rng.IntN(hi-lo+1))
	for i := range kids {
		kids[i] = g.node(depth, typ, dom, beam, eq)
	}
	if len(kids) > 1 && kids[0].isLeaf() && g.rng.IntN(4) != 0 {
		kids = append(kids[1:], kids[0])
	}
	return kids
}

func (g *omGen) comparator(typ int) omEq {
	exact := 5
	if typ == omPair {
		exact = 2
	}
	switch r := g.rng.IntN(10); {
	case r < exact:
		return omExact
	case r < 8:
		return omCoarse[typ][g.rng.IntN(len(omCoarse[typ]))]
	}
	return omEq{name: "nil"}
}

func (g *omGen) leaf(n *omNode, typ int, dom []any, beam bool) *omNode {
	switch r := g.rng.IntN(20); {
	case r < 9:
		n.op, n.label = omOpBind, g.label("B")
		if g.rng.IntN(4) == 0 {
			n.sname, n.show = g.predicate(typ, dom)
		}
	case r < 18:
		n.op, n.label = omOpComp, g.label("L")
		if beam {
			n.kind = 2 * g.rng.IntN(2)
		} else {
			n.kind = g.rng.IntN(3)
		}
	default:
		n.op, n.label, n.kind = omOpComp, g.label("N"), omLeafNil
	}
	return n
}

func (g *omGen) node(depth, typ int, dom []any, beam bool, eq omEq) *omNode {
	n := &omNode{cur: -1, eq: eq}
	r := g.rng.IntN(100)
	if depth >= g.depth || r < 28 && (depth > 0 || r < 4) {
		return g.leaf(n, typ, dom, beam)
	}
	if !beam && g.rng.IntN(5) == 0 {
		beam = true
	}
	n.beam = beam
	switch {
	case r < 55:
		n.op, n.label = omOpMatch, g.label("M")
		n.pname, n.pred = g.predicate(typ, dom)
		if g.rng.IntN(12) != 0 {
			n.kids = g.group(depth+1, typ, dom, beam, eq, 1, 3)
		}
	case r < 72:
		n.op, n.label = omOpValue, g.label("V")
		n.val = g.pick(dom)
		if g.rng.IntN(10) == 0 {
			n.val = [...]any{1000, "zz", omKV{-1, -1}}[typ]
		}
		n.kids = g.group(depth+1, typ, dom, beam, eq, 1, 2)
	default:
		n.op, n.label = omOpDerive, g.label("D")
		var fns []omFn
		for _, fn := range omFns {
			if fn.in == typ {
				fns = append(fns, fn)
			}
		}
		n.fn = fns[g.rng.IntN(len(fns))]
		n.out = g.comparator(n.fn.out)
		if g.rng.IntN(20) != 0 {
			n.kids = g.group(depth+1, n.fn.out, omMapDom(dom, n.fn.f), beam, n.out, 1, 3)
		}
	}
	return n
}

func (g *omGen) gen(i int, dom []any) *omTree {
	g.tree, g.n = i, 0
	g.depth = 2 + g.rng.IntN(4)
	tr := &omTree{slot: fmt.Sprintf("t%d", i), entry: g.rng.IntN(5), cur: -1, blank: true, views: map[string]bool{}}
	tr.top = g.group(0, omInt, dom, tr.entry >= 2, omExact, 2, 4)
	return tr
}

func omSrcRoutes[T comparable](r *rtRec, ns []*omNode) []SourceRoute[T] {
	out := make([]SourceRoute[T], len(ns))
	for i, n := range ns {
		out[i] = omSrcRoute[T](r, n)
	}
	return out
}

func omBeamRoutes[T comparable](r *rtRec, ns []*omNode) []BeamRoute[T] {
	out := make([]BeamRoute[T], len(ns))
	for i, n := range ns {
		out[i] = omBeamRoute[T](r, n)
	}
	return out
}

func omSrcRoute[T comparable](r *rtRec, n *omNode) SourceRoute[T] {
	switch {
	case n.op == omOpComp && n.kind == omLeafSource:
		return rtSource[T](r, n.label, nil)
	case n.beam || n.isLeaf():
		return omBeamRoute[T](r, n)
	case n.op == omOpMatch:
		return RouteMatch(func(v T) bool { return n.pred(v) }, omSrcRoutes[T](r, n.kids)...)
	case n.op == omOpValue:
		return RouteValue(n.val.(T), omSrcRoutes[T](r, n.kids)...)
	}
	switch n.fn.out {
	case omInt:
		return RouteDerive(rtDerive(r, n.label, omDeriveSource[T, int](n)), omSrcRoutes[int](r, n.kids)...)
	case omStr:
		return RouteDerive(rtDerive(r, n.label, omDeriveSource[T, string](n)), omSrcRoutes[string](r, n.kids)...)
	}
	return RouteDerive(rtDerive(r, n.label, omDeriveSource[T, omKV](n)), omSrcRoutes[omKV](r, n.kids)...)
}

func omBeamRoute[T comparable](r *rtRec, n *omNode) BeamRoute[T] {
	switch {
	case n.op == omOpComp && n.kind == omLeafBeam:
		return rtBeam[T](r, n.label, nil)
	case n.op == omOpComp && n.kind == omLeafNil:
		return RouteComp[T](nil)
	case n.op == omOpComp:
		return rtComp[T](r, n.label)
	case n.op == omOpBind && n.show != nil:
		return rtBindIf(r, n.label, func(v T) bool { return n.show(v) })
	case n.op == omOpBind:
		return rtBind[T](r, n.label)
	case n.op == omOpMatch:
		return RouteBeamMatch(func(v T) bool { return n.pred(v) }, omBeamRoutes[T](r, n.kids)...)
	case n.op == omOpValue:
		return RouteBeamValue(n.val.(T), omBeamRoutes[T](r, n.kids)...)
	}
	switch n.fn.out {
	case omInt:
		return RouteBeamDerive(rtDeriveBeam(r, n.label, omDeriveBeam[T, int](n)), omBeamRoutes[int](r, n.kids)...)
	case omStr:
		return RouteBeamDerive(rtDeriveBeam(r, n.label, omDeriveBeam[T, string](n)), omBeamRoutes[string](r, n.kids)...)
	}
	return RouteBeamDerive(rtDeriveBeam(r, n.label, omDeriveBeam[T, omKV](n)), omBeamRoutes[omKV](r, n.kids)...)
}

func omTyped[U any](e omEq) func(a, b U) bool {
	if e.f == nil {
		return nil
	}
	return func(a, b U) bool { return e.f(a, b) }
}

func omDeriveSource[T, U comparable](n *omNode) func(Source[T]) Source[U] {
	get := func(v T) U { return n.fn.f(v).(U) }
	set := func(a T, _ U) T { return a }
	return func(s Source[T]) Source[U] {
		if n.out.name == omExact.name {
			return DeriveSource(s, get, set)
		}
		return DeriveSourceEqual(s, get, set, omTyped[U](n.out))
	}
}

func omDeriveBeam[T, U comparable](n *omNode) func(Beam[T]) Beam[U] {
	get := func(v T) U { return n.fn.f(v).(U) }
	return func(b Beam[T]) Beam[U] {
		if n.out.name == omExact.name {
			return DeriveBeam(b, get)
		}
		return DeriveBeamEqual(b, get, omTyped[U](n.out))
	}
}

func omMount(r *rtRec, src Source[int], tr *omTree) gox.Elem {
	switch tr.entry {
	case 0:
		return src.Route(omSrcRoutes[int](r, tr.top)...)
	case 1:
		return DeriveSource(src, rtID, rtIDSet).Route(omSrcRoutes[int](r, tr.top)...)
	case 2:
		return src.RouteBeam(omBeamRoutes[int](r, tr.top)...)
	case 3:
		return DeriveBeam(src, rtID).RouteBeam(omBeamRoutes[int](r, tr.top)...)
	}
	return DeriveSource(src, rtID, rtIDSet).RouteBeam(omBeamRoutes[int](r, tr.top)...)
}

type omPage struct {
	h     *rtHarness
	trees []*omTree
	dom   string
	seen  int
	acts  map[uint64]int
	hist  []int
}

// check compares every slot's text and replace count with the model.
func (pg *omPage) check(at string) {
	h := pg.h
	h.t.Helper()
	acts := h.client.actions()
	fresh := acts[pg.seen:]
	for _, a := range fresh {
		if !rtIsDoorAction(a) {
			continue
		}
		if a.name != "door_replace" || !slices.Contains(slices.Collect(maps.Values(h.doors0)), a.id) {
			h.t.Fatalf("%s: %s for door %d, outside the render-time route doors", at, a.name, a.id)
		}
		pg.acts[a.id]++
	}
	pg.dom = rtDOM(pg.dom, fresh)
	pg.seen = len(acts)
	if ids := h.inst.ids.Load(); ids != h.baseIDs {
		h.t.Fatalf("%s: ids moved from %d to %d", at, h.baseIDs, ids)
	}
	for _, tr := range pg.trees {
		id := h.doors0[tr.slot]
		inner, ok := rtSlotInner(pg.dom, tr.slot)
		if !ok {
			h.t.Fatalf("%s: slot %s missing, dom:\n%s", at, tr.slot, pg.dom)
		}
		if ids := rtDoorIDs(inner); len(ids) != 1 || ids[0] != id {
			h.t.Fatalf("%s: slot %s has doors %v, want only %d\ntree: %s", at, tr.slot, ids, id, tr.describe())
		}
		want, key := tr.view()
		if got := rtText(inner); got != want || pg.acts[id] != tr.acts {
			h.t.Fatalf("%s: slot %s: text %q, want %q (path %s); door replaces %d, want %d\ntree: %s\nvalues: %v",
				at, tr.slot, got, want, key, pg.acts[id], tr.acts, tr.describe(), pg.hist)
		}
	}
}

// final checks exact render, clean and derive counters.
func (pg *omPage) final() {
	pg.h.t.Helper()
	want := map[string]int{}
	for _, tr := range pg.trees {
		var shown *omNode
		if tr.on && !tr.blank {
			shown = tr.top[tr.cur].leaf()
		}
		for _, top := range tr.top {
			top.walk(func(n *omNode) {
				switch {
				case n.isNil():
				case n.isLeaf():
					want["render:"+n.label] = n.renders
					want["clean:"+n.label] = n.contents
					if n == shown && n.content() {
						want["clean:"+n.label]--
					}
				case n.op == omOpDerive:
					want["derive:"+n.label] = 1
				}
			})
		}
	}
	pg.h.counts(want)
}

// 240 generated trees must match the oracle after every update; the generator must reach every rule.
func TestRouteOracle(t *testing.T) {
	const pages, perPage, updates = 10, 24, 64
	dom := make([]any, 60)
	for i := range dom {
		dom[i] = i
	}
	var trees, lively, blanks, nils, deepest int
	var st omStats
	for p := range pages {
		t.Run(fmt.Sprintf("page%d", p), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(20261012, uint64(p)))
			g := &omGen{rng: rng}
			workers := 8
			if p%3 == 2 {
				workers = 1
			}
			h := newRTN(t, workers)
			start := rng.IntN(len(dom))
			src := NewSource(start)
			pg := &omPage{h: h, acts: map[uint64]int{}, hist: []int{start}}
			var els []gox.Elem
			for i := range perPage {
				tr := g.gen(i, dom)
				pg.trees = append(pg.trees, tr)
				els = append(els, rtSlot(tr.slot, omMount(h.rec, src, tr)))
			}
			h.render(els...)
			pg.dom = h.initial
			for _, tr := range pg.trees {
				tr.apply(start, true)
			}
			pg.check("initial")
			prev := start
			for u := range updates {
				v := rng.IntN(len(dom))
				if rng.IntN(10) == 0 {
					v = prev
				}
				ret := rtSet(h, src, v)
				want := []error{}
				if v != prev {
					want = []error{nil}
					for _, tr := range pg.trees {
						tr.apply(v, false)
					}
				}
				if !rtSameErrs(ret, want) {
					t.Fatalf("update %d (set %d) returned %v, want %v", u, v, ret, want)
				}
				pg.hist = append(pg.hist, v)
				pg.check(fmt.Sprintf("update %d (set %d)", u, v))
				prev = v
			}
			pg.final()
			for _, tr := range pg.trees {
				trees++
				st.switches += tr.st.switches
				st.breaks += tr.st.breaks
				st.skips += tr.st.skips
				st.cached += tr.st.cached
				st.falls += tr.st.falls
				st.pairSkips += tr.st.pairSkips
				if len(tr.views) >= 3 {
					lively++
				}
				for k := range tr.views {
					switch {
					case k == "-":
						blanks++
					case strings.HasSuffix(k, "(nil)"):
						nils++
					}
					deepest = max(deepest, strings.Count(k, "/"))
				}
			}
		})
	}
	t.Logf("%d trees, %d with 3+ views, %d blank views, %d nil leaf views, deepest path %d; %+v", trees, lively, blanks, nils, deepest, st)
	if trees == pages*perPage && (lively*2 < trees || blanks == 0 || nils == 0 || deepest < 4 ||
		st.falls == 0 || st.switches == 0 || st.breaks == 0 || st.skips == 0 || st.pairSkips == 0 || st.cached == 0) {
		t.Fatalf("generator too tame: %d/%d lively, %d blank, %d nil, depth %d, %+v", lively, trees, blanks, nils, deepest, st)
	}
}
