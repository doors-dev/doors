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

type r2Harness struct {
	*lifecycleHarness
	rec     *r2Rec
	page    context.Context
	initial string
	baseIDs uint64
	doors0  map[string]uint64
	strict  bool
}

func newR2Harness(t *testing.T) *r2Harness {
	t.Helper()
	return newR2HarnessN(t, 8)
}

func newR2HarnessN(t *testing.T, workers int) *r2Harness {
	t.Helper()
	return &r2Harness{
		lifecycleHarness: newLifecycleHarness(t, workers),
		rec:              newR2Rec(),
		doors0:           map[string]uint64{},
		strict:           true,
	}
}

func r2Slot(name string, el gox.Elem) gox.Elem {
	return tagElem("div", "r2-"+name, el)
}

var r2SlotRe = regexp.MustCompile(` id="r2-([^"]+)"`)

func (h *r2Harness) render(content ...gox.Elem) {
	h.t.Helper()
	ctx, page, err := h.renderPageHTML(func(cur gox.Cursor) error {
		for _, el := range content {
			if el == nil {
				continue
			}
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
	for _, m := range r2SlotRe.FindAllStringSubmatch(page, -1) {
		inner, ok := r2SlotInner(page, m[1])
		if !ok {
			continue
		}
		if ids := r2DoorIDs(inner); len(ids) != 0 {
			h.doors0[m[1]] = ids[0]
		}
	}
}

var r2Void = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true, "img": true,
	"input": true, "link": true, "meta": true, "source": true, "track": true, "wbr": true,
}

func r2IndexTag(s, tag string) int {
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

func r2Element(dom, marker string) (start, openEnd, closeStart, end int, ok bool) {
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
	if r2Void[tag] {
		return start, openEnd, openEnd, openEnd, true
	}
	closeTag := "</" + tag + ">"
	depth, pos := 1, openEnd
	for {
		c := strings.Index(dom[pos:], closeTag)
		if c < 0 {
			return 0, 0, 0, 0, false
		}
		if o := r2IndexTag(dom[pos:], tag); o >= 0 && o < c {
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

func r2DoorElement(dom string, id uint64) (start, openEnd, closeStart, end int, ok bool) {
	for _, m := range []string{fmt.Sprintf(` id="d0r%d"`, id), fmt.Sprintf(` data-d0r="%d"`, id)} {
		if start, openEnd, closeStart, end, ok = r2Element(dom, m); ok {
			return
		}
	}
	return 0, 0, 0, 0, false
}

func r2IsDoorAction(a lifecycleAction) bool {
	return a.name == "door_replace" || a.name == "door_update"
}

func r2DOM(initial string, acts []lifecycleAction) string {
	dom := initial
	for _, a := range acts {
		if !r2IsDoorAction(a) {
			continue
		}
		s, oe, cs, e, ok := r2DoorElement(dom, a.id)
		if !ok {
			continue
		}
		if a.name == "door_replace" {
			dom = dom[:s] + a.body + dom[e:]
		} else {
			dom = dom[:oe] + a.body + dom[cs:]
		}
	}
	return dom
}

var (
	r2DoorRe = regexp.MustCompile(` id="d0r(\d+)"| data-d0r="(\d+)"`)
	r2TagRe  = regexp.MustCompile(`<[^>]*>`)
)

func r2DoorIDs(s string) []uint64 {
	var ids []uint64
	for _, m := range r2DoorRe.FindAllStringSubmatch(s, -1) {
		v := m[1]
		if v == "" {
			v = m[2]
		}
		n, _ := strconv.ParseUint(v, 10, 64)
		ids = append(ids, n)
	}
	return ids
}

func r2SlotInner(dom, name string) (string, bool) {
	_, oe, cs, _, ok := r2Element(dom, ` id="r2-`+name+`"`)
	if !ok {
		return "", false
	}
	return dom[oe:cs], true
}

func r2Text(s string) string {
	return html.UnescapeString(r2TagRe.ReplaceAllString(s, ""))
}

func (h *r2Harness) dom() string {
	return r2DOM(h.initial, h.client.actions())
}

func (h *r2Harness) slot(name string) (string, bool) {
	return r2SlotInner(h.dom(), name)
}

func (h *r2Harness) html(name string) string {
	h.t.Helper()
	dom := h.dom()
	inner, ok := r2SlotInner(dom, name)
	if !ok {
		h.t.Fatalf("slot %q not found in dom:\n%s", name, dom)
	}
	return inner
}

func (h *r2Harness) text(name string) string {
	h.t.Helper()
	return r2Text(h.html(name))
}

func (h *r2Harness) door(name string) uint64 {
	h.t.Helper()
	ids := r2DoorIDs(h.html(name))
	if len(ids) == 0 {
		h.t.Fatalf("no door in slot %q, dom:\n%s", name, h.dom())
	}
	return ids[0]
}

func (h *r2Harness) doorCount(name string) int {
	h.t.Helper()
	return len(r2DoorIDs(h.html(name)))
}

func (h *r2Harness) doorActions(id uint64) []lifecycleAction {
	var out []lifecycleAction
	for _, a := range h.client.actions() {
		if r2IsDoorAction(a) && a.id == id {
			out = append(out, a)
		}
	}
	return out
}

func (h *r2Harness) settle(op func(ctx context.Context)) {
	h.t.Helper()
	if h.page == nil {
		h.t.Fatal("settle before render")
	}
	done := make(chan struct{})
	OnSettle(h.page, func(context.Context) { close(done) }, op)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		h.t.Fatalf("settle timed out, dom:\n%s", h.dom())
	}
	h.waitIdle()
}

func r2Set[T any](h *r2Harness, s Source[T], v T) []error {
	h.t.Helper()
	var ch <-chan error
	h.settle(func(ctx context.Context) { ch = s.Update(ctx, v) })
	return drainOp(h.t, ch)
}

func r2Mutate[T any](h *r2Harness, s Source[T], f func(T) T) []error {
	h.t.Helper()
	var ch <-chan error
	h.settle(func(ctx context.Context) { ch = s.Mutate(ctx, f) })
	return drainOp(h.t, ch)
}

func (h *r2Harness) eventually(name, want string) {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if h.client.idle() {
			if inner, ok := h.slot(name); ok && r2Text(inner) == want {
				return
			}
		}
		if time.Now().After(deadline) {
			inner, ok := h.slot(name)
			h.t.Fatalf("slot %q: want %q, got %q (found=%v), dom:\n%s", name, want, r2Text(inner), ok, h.dom())
		}
		time.Sleep(time.Millisecond)
	}
}

func (h *r2Harness) expect(name, want string) {
	h.t.Helper()
	h.expectAt("", name, want)
}

func (h *r2Harness) expectAt(at, name, want string) {
	h.t.Helper()
	dom := h.dom()
	inner, ok := r2SlotInner(dom, name)
	if !ok {
		h.t.Fatalf("%sslot %q not found, dom:\n%s", at, name, dom)
	}
	if got := r2Text(inner); got != want {
		h.t.Fatalf("%sslot %q: want %q, got %q, dom:\n%s", at, name, want, got, dom)
	}
}

func (h *r2Harness) quiet(name string, step func()) {
	h.t.Helper()
	h.quietAt("", name, step)
}

func (h *r2Harness) quietAt(at, name string, step func()) {
	h.t.Helper()
	id := h.door(name)
	n := len(h.doorActions(id))
	step()
	if got := h.doorActions(id); len(got) != n {
		h.t.Fatalf("%sslot %q door %d: expected no new actions, got %+v", at, name, id, got[n:])
	}
}

func (h *r2Harness) oneDoor(name string) {
	h.t.Helper()
	h.oneDoorAt("", name)
}

func (h *r2Harness) oneDoorAt(at, name string) {
	h.t.Helper()
	if n := h.doorCount(name); n != 1 {
		h.t.Fatalf("%sslot %q: want 1 door, got %d, dom:\n%s", at, name, n, h.dom())
	}
	want, ok := h.doors0[name]
	if !ok {
		h.t.Fatalf("%sslot %q had no door at render", at, name)
	}
	if got := h.door(name); got != want {
		h.t.Fatalf("%sslot %q: door %d, want render-time door %d", at, name, got, want)
	}
	if ids := h.inst.ids.Load(); ids != h.baseIDs {
		h.t.Fatalf("%sids moved from %d to %d: something created a door, tracker or hook", at, h.baseIDs, ids)
	}
	allowed := map[uint64]bool{}
	for _, id := range h.doors0 {
		allowed[id] = true
	}
	for _, a := range h.client.actions() {
		if !r2IsDoorAction(a) {
			continue
		}
		if !allowed[a.id] {
			h.t.Fatalf("%saction %s for door %d, which is not a render-time slot door %v", at, a.name, a.id, h.doors0)
		}
		if a.id == want && a.name != "door_replace" {
			h.t.Fatalf("%sslot %q door %d got %s: route doors only take door_replace", at, name, want, a.name)
		}
	}
}

func (h *r2Harness) counts(want map[string]int) {
	h.t.Helper()
	h.countsAt("", want)
}

func (h *r2Harness) countsAt(at string, want map[string]int) {
	h.t.Helper()
	var bad []string
	for _, k := range slices.Sorted(func(yield func(string) bool) {
		for k := range want {
			if !yield(k) {
				return
			}
		}
	}) {
		if got := h.rec.count(k); got != want[k] {
			bad = append(bad, fmt.Sprintf("%s=%d (want %d)", k, got, want[k]))
		}
	}
	if len(bad) != 0 {
		h.t.Fatalf("%scounts: %s\nlog: %v", at, strings.Join(bad, ", "), h.rec.all())
	}
}

type r2Step[T any] struct {
	set    T
	want   string
	quiet  bool
	ret    []error
	counts map[string]int
}

func r2SameErrs(got, want []error) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if want[i] == nil {
			if got[i] != nil {
				return false
			}
			continue
		}
		if !errors.Is(got[i], want[i]) {
			return false
		}
	}
	return true
}

func r2Run[T any](h *r2Harness, slot string, src Source[T], initial string, steps ...r2Step[T]) {
	h.t.Helper()
	h.expectAt("initial: ", slot, initial)
	if h.strict {
		h.oneDoorAt("initial: ", slot)
	}
	for i, st := range steps {
		at := fmt.Sprintf("step %d (set %v): ", i, st.set)
		var ret []error
		do := func() { ret = r2Set(h, src, st.set) }
		if st.quiet {
			h.quietAt(at, slot, do)
		} else {
			do()
		}
		if st.ret != nil && !r2SameErrs(ret, st.ret) {
			h.t.Fatalf("%supdate returned %v, want %v", at, ret, st.ret)
		}
		h.expectAt(at, slot, st.want)
		if st.counts != nil {
			h.countsAt(at, st.counts)
		}
		if h.strict {
			h.oneDoorAt(at, slot)
		}
	}
}

func r2Same(h *r2Harness, a, b string) {
	h.t.Helper()
	if ta, tb := h.text(a), h.text(b); ta != tb {
		h.t.Fatalf("slot %q has %q, slot %q has %q, dom:\n%s", a, ta, b, tb, h.dom())
	}
}

type r2Rec struct {
	mu  sync.Mutex
	n   map[string]int
	log []string
}

func newR2Rec() *r2Rec {
	return &r2Rec{n: map[string]int{}}
}

func (r *r2Rec) hit(key string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.n[key]++
	r.log = append(r.log, key)
}

func (r *r2Rec) count(key string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.n[key]
}

func (r *r2Rec) seq(prefix string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []string
	for _, k := range r.log {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	return out
}

func (r *r2Rec) vals(prefix string) []string {
	out := r.seq(prefix)
	for i, k := range out {
		out[i] = k[len(prefix):]
	}
	return out
}

func (r *r2Rec) all() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.log)
}

type r2Leaf struct {
	rec   *r2Rec
	label string
}

func (l r2Leaf) Main() gox.Elem {
	return func(cur gox.Cursor) error {
		l.rec.hit("render:" + l.label)
		OnClean(cur.Context(), func() { l.rec.hit("clean:" + l.label) })
		return cur.Text(l.label)
	}
}

func r2Comp[T any](r *r2Rec, label string) BeamRoute[T] {
	return RouteComp[T](r2Leaf{r, label})
}

func r2Bind[T any](r *r2Rec, label string) BeamRoute[T] {
	return r2BindIf[T](r, label, nil)
}

func r2BindIf[T any](r *r2Rec, label string, show func(T) bool) BeamRoute[T] {
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

type r2Box[T any] struct {
	mu  sync.Mutex
	all []T
}

func (b *r2Box[T]) put(v T) {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.all = append(b.all, v)
}

func (b *r2Box[T]) len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.all)
}

func (b *r2Box[T]) at(i int) T {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.all[i]
}

func (b *r2Box[T]) last() T {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.all[len(b.all)-1]
}

func r2Source[T any](r *r2Rec, label string, box *r2Box[Source[T]]) SourceRoute[T] {
	return RouteSource(func(s Source[T]) gox.Elem {
		box.put(s)
		return r2Leaf{r, label}.Main()
	})
}

func r2Beam[T any](r *r2Rec, label string, box *r2Box[Beam[T]]) BeamRoute[T] {
	return RouteBeam(func(b Beam[T]) gox.Elem {
		box.put(b)
		return r2Leaf{r, label}.Main()
	})
}

func r2Pred[T any](r *r2Rec, label string, p func(T) bool) func(T) bool {
	return func(v T) bool {
		r.hit("pred:" + label + "=" + fmt.Sprint(v))
		return p(v)
	}
}

func r2Derive[T1, T2 any](r *r2Rec, label string, f func(Source[T1]) Source[T2]) func(Source[T1]) Source[T2] {
	return func(s Source[T1]) Source[T2] {
		r.hit("derive:" + label)
		return f(s)
	}
}

func r2DeriveBeam[T1, T2 any](r *r2Rec, label string, f func(Beam[T1]) Beam[T2]) func(Beam[T1]) Beam[T2] {
	return func(b Beam[T1]) Beam[T2] {
		r.hit("derive:" + label)
		return f(b)
	}
}

func r2Div(x int) func(Source[int]) Source[int] {
	return func(s Source[int]) Source[int] {
		return DeriveSource(s, func(v int) int { return v / x }, func(a, b int) int { return b*x + a%x })
	}
}

type r2Big struct {
	kind int
	data []byte
}

func newR2Big(kind int) *r2Big {
	return &r2Big{kind: kind, data: make([]byte, 1<<20)}
}

func (b *r2Big) String() string {
	return "big" + strconv.Itoa(b.kind)
}

func r2Gone[T any](p weak.Pointer[T]) bool {
	for range 20 {
		runtime.GC()
		if p.Value() == nil {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

type r2Kit struct {
	name   string
	beam   bool
	match  func(p func(int) bool, rs ...SourceRoute[int]) SourceRoute[int]
	value  func(v int, rs ...SourceRoute[int]) SourceRoute[int]
	derive func(r *r2Rec, label string, get func(int) int, set func(a, b int) int, rs ...SourceRoute[int]) SourceRoute[int]
}

func r2Lens(get func(int) int, set func(a, b int) int) (func(int) int, func(a, b int) int) {
	if get == nil {
		get = func(v int) int { return v }
	}
	if set == nil {
		set = func(_, b int) int { return b }
	}
	return get, set
}

var r2Kits = []r2Kit{
	{
		name: "source",
		match: func(p func(int) bool, rs ...SourceRoute[int]) SourceRoute[int] {
			return RouteMatch(p, rs...)
		},
		value: func(v int, rs ...SourceRoute[int]) SourceRoute[int] {
			return RouteValue(v, rs...)
		},
		derive: func(r *r2Rec, label string, get func(int) int, set func(a, b int) int, rs ...SourceRoute[int]) SourceRoute[int] {
			get, set = r2Lens(get, set)
			return RouteDerive(r2Derive(r, label, func(s Source[int]) Source[int] {
				return DeriveSource(s, get, set)
			}), rs...)
		},
	},
	{
		name: "beam",
		beam: true,
		match: func(p func(int) bool, rs ...SourceRoute[int]) SourceRoute[int] {
			return RouteBeamMatch(p, r2Beams(rs)...)
		},
		value: func(v int, rs ...SourceRoute[int]) SourceRoute[int] {
			return RouteBeamValue(v, r2Beams(rs)...)
		},
		derive: func(r *r2Rec, label string, get func(int) int, _ func(a, b int) int, rs ...SourceRoute[int]) SourceRoute[int] {
			get, _ = r2Lens(get, nil)
			return RouteBeamDerive(r2DeriveBeam(r, label, func(b Beam[int]) Beam[int] {
				return DeriveBeam(b, get)
			}), r2Beams(rs)...)
		},
	},
}

func r2Beams[T any](rs []SourceRoute[T]) []BeamRoute[T] {
	out := make([]BeamRoute[T], len(rs))
	for i, r := range rs {
		out[i] = r.(BeamRoute[T])
	}
	return out
}

type r2Entry struct {
	name  string
	beam  bool
	mount func(src Source[int], rs ...SourceRoute[int]) gox.Elem
}

func r2ID(v int) int { return v }

func r2IDSet(_, v int) int { return v }

var r2Entries = []r2Entry{
	{"source.Route", false, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return src.Route(rs...)
	}},
	{"source.RouteBeam", true, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return src.RouteBeam(r2Beams(rs)...)
	}},
	{"derivedSource.Route", false, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return DeriveSource(src, r2ID, r2IDSet).Route(rs...)
	}},
	{"derivedSource.RouteBeam", true, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return DeriveSource(src, r2ID, r2IDSet).RouteBeam(r2Beams(rs)...)
	}},
	{"derivedBeam.RouteBeam", true, func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return DeriveBeam(src, r2ID).RouteBeam(r2Beams(rs)...)
	}},
}

func r2Matrix(t *testing.T, f func(t *testing.T, k r2Kit, e r2Entry)) {
	t.Helper()
	for _, k := range r2Kits {
		for _, e := range r2Entries {
			if e.beam && !k.beam {
				continue
			}
			t.Run(k.name+"/"+e.name, func(t *testing.T) { f(t, k, e) })
		}
	}
}

func r2Parallel(t *testing.T, n int, f func(t *testing.T, i int)) {
	t.Helper()
	t.Run("parallel", func(t *testing.T) {
		for i := range n {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				t.Parallel()
				f(t, i)
			})
		}
	})
}

type r2PathA struct {
	A bool `path:"/a"`
}

type r2PathB struct {
	B bool `path:"/b"`
}

type r2PathItem struct {
	Item bool `path:"/item/:ID"`
	ID   int
	Tab  string `query:"tab"`
}

type r2PathName struct {
	Item bool `path:"/item/:Name"`
	Name string
}

type r2PathBad struct {
	X string
}

type r2Custom struct {
	ID, Tab string
}

func (c r2Custom) Encode() (Location, error) {
	if c.ID == "" {
		return Location{}, errors.New("r2Custom: empty ID")
	}
	loc := Location{Segments: []string{"item", c.ID}}
	if c.Tab != "" {
		loc.Query = url.Values{"tab": {c.Tab}}
	}
	return loc, nil
}

func r2Loc(uri string) Location {
	return path.NewLocationFromEscapedURI(uri)
}

func r2LocSource(uri string) Source[Location] {
	return NewSourceEqual(r2Loc(uri), path.EqualLocation)
}

func r2Strip(prefix string) func(Source[Location]) Source[Location] {
	return func(s Source[Location]) Source[Location] {
		return DeriveSourceEqual(s, func(l Location) Location {
			l = l.Clone()
			if len(l.Segments) != 0 {
				l.Segments = l.Segments[1:]
			}
			return l
		}, func(_ Location, l Location) Location {
			l = l.Clone()
			l.Segments = append([]string{prefix}, l.Segments...)
			return l
		}, path.EqualLocation)
	}
}

var (
	_ func(gox.Comp) BeamRoute[int]                                                   = RouteComp[int]
	_ func(func(int) gox.Elem) BeamRoute[int]                                         = RouteBind[int]
	_ func(func(Beam[int]) gox.Elem) BeamRoute[int]                                   = RouteBeam[int]
	_ func(func(int) bool, ...BeamRoute[int]) BeamRoute[int]                          = RouteBeamMatch[int]
	_ func(int, ...BeamRoute[int]) BeamRoute[int]                                     = RouteBeamValue[int]
	_ func(func(Beam[int]) Beam[string], ...BeamRoute[string]) BeamRoute[int]         = RouteBeamDerive[int, string]
	_ func(...BeamRoute[r2PathA]) BeamRoute[Location]                                 = RouteBeamModel[r2PathA]
	_ func(func(Source[int]) gox.Elem) SourceRoute[int]                               = RouteSource[int]
	_ func(func(int) bool, ...SourceRoute[int]) SourceRoute[int]                      = RouteMatch[int]
	_ func(int, ...SourceRoute[int]) SourceRoute[int]                                 = RouteValue[int]
	_ func(func(Source[int]) Source[string], ...SourceRoute[string]) SourceRoute[int] = RouteDerive[int, string]
	_ func(...SourceRoute[r2PathA]) SourceRoute[Location]                             = RouteModel[r2PathA]
	_ func(Source[int], ...SourceRoute[int]) gox.Elem                                 = Source[int].Route
	_ func(Source[int], ...BeamRoute[int]) gox.Elem                                   = Source[int].RouteBeam
	_ func(Beam[int], ...BeamRoute[int]) gox.Elem                                     = Beam[int].RouteBeam
)

// The first seek runs inside the page render: no client action, one door
// per container, and a blank container when nothing matches at the top.
func TestRoute2HarnessInitial(t *testing.T) {
	h := newR2Harness(t)
	a, b, c := NewSource(1), NewSource(0), NewSource(0)
	h.render(
		r2Slot("match", a.Route(RouteValue(1, r2Comp[int](h.rec, "ONE")), r2Comp[int](h.rec, "D"))),
		r2Slot("none", b.Route(RouteValue(1, r2Comp[int](h.rec, "NEVER")))),
		r2Slot("elem", c.Route(RouteComp[int](tagElem("section", "sec", textElem("S<&>"))))),
	)
	h.expect("match", "ONE")
	h.expect("none", "")
	h.expect("elem", "S<&>")
	if got := h.html("none"); !strings.HasPrefix(got, "<d0-r ") || !strings.HasSuffix(got, "></d0-r>") {
		t.Fatalf("expected a blank route container, got %q", got)
	}
	if got := h.html("elem"); !strings.HasPrefix(got, "<section ") || !strings.Contains(got, fmt.Sprintf(` data-d0r="%d"`, h.door("elem"))) {
		t.Fatalf("expected the section adopted as the container, got %q", got)
	}
	seen := map[uint64]bool{}
	for _, name := range []string{"match", "none", "elem"} {
		h.oneDoor(name)
		seen[h.doors0[name]] = true
	}
	if len(seen) != 3 {
		t.Fatalf("expected three distinct route doors, got %v", h.doors0)
	}
	if acts := h.client.actions(); len(acts) != 0 {
		t.Fatalf("expected no client actions, got %+v", acts)
	}
	h.counts(map[string]int{"render:ONE": 1, "clean:ONE": 0, "render:D": 0, "render:NEVER": 0})
}

// An active non-default branch is sticky while it still matches; an active
// default yields to the first earlier sibling that matches.
func TestRoute2HarnessSwitch(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(1)
	h.render(r2Slot("a", src.Route(
		RouteMatch(func(v int) bool { return v%2 == 0 }, r2Comp[int](h.rec, "EVEN")),
		RouteMatch(func(v int) bool { return v < 10 }, r2Comp[int](h.rec, "SMALL")),
		r2Comp[int](h.rec, "D"),
	)))
	r2Run(h, "a", src, "SMALL",
		r2Step[int]{set: 2, want: "SMALL", quiet: true, ret: []error{nil}},
		r2Step[int]{set: 12, want: "EVEN"},
		r2Step[int]{set: 3, want: "SMALL", counts: map[string]int{"render:SMALL": 2, "clean:SMALL": 1}},
		r2Step[int]{set: 13, want: "D"},
		r2Step[int]{set: 11, want: "D", quiet: true},
		r2Step[int]{set: 4, want: "EVEN", counts: map[string]int{"render:D": 1, "clean:D": 1, "render:EVEN": 2}},
		r2Step[int]{set: 4, want: "EVEN", quiet: true, ret: []error{}},
	)
	if n := len(h.doorActions(h.door("a"))); n != 4 {
		t.Fatalf("expected 4 door replaces, got %d", n)
	}
}

// keepEnd at the top level re-selects in the same update, re-trying the
// ended route; with no match the container goes blank. A match whose
// children match nothing does not match: selection falls through to the
// later default, initially and on a nested path break.
func TestRoute2HarnessKeepEnd(t *testing.T) {
	h := newR2Harness(t)
	a, b := NewSource(1), NewSource(1)
	h.render(
		r2Slot("a", a.Route(RouteMatch(r2Pred(h.rec, "a", func(v int) bool { return v > 0 }), r2Bind[int](h.rec, "B")))),
		r2Slot("b", b.Route(
			RouteMatch(func(v int) bool { return v > 0 }, RouteValue(5, r2Comp[int](h.rec, "X"))),
			r2Comp[int](h.rec, "D"),
		)),
	)
	r2Run(h, "a", a, "B=1",
		r2Step[int]{set: 2, want: "B=2"},
		r2Step[int]{set: 0, want: ""},
		r2Step[int]{set: -1, want: "", quiet: true},
		r2Step[int]{set: 3, want: "B=3", counts: map[string]int{"render:B": 3, "clean:B": 2}},
	)
	if got, want := h.rec.vals("pred:a="), []string{"1", "2", "0", "0", "-1", "3"}; !slices.Equal(got, want) {
		t.Fatalf("predicate log %v, want %v", got, want)
	}
	r2Run(h, "b", b, "D",
		r2Step[int]{set: 5, want: "X"},
		r2Step[int]{set: 2, want: "D"},
		r2Step[int]{set: 3, want: "D", quiet: true},
		r2Step[int]{set: 0, want: "D", quiet: true},
		r2Step[int]{set: 5, want: "X", counts: map[string]int{"render:X": 2, "clean:X": 1, "render:D": 2, "clean:D": 2}},
	)
}

// Deep switches across nested match, value and derive containers stay on the
// render-time door, for every kit and entry point.
func TestRoute2HarnessDeepSwitch(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		for _, workers := range []int{8, 1} {
			h := newR2HarnessN(t, workers)
			src := NewSource(0)
			tens := func(v int) int { return v / 10 }
			tensSet := func(a, b int) int { return b*10 + a%10 }
			h.render(r2Slot("deep", e.mount(src,
				k.match(func(v int) bool { return v < 1000 },
					k.derive(h.rec, "tens", tens, tensSet,
						k.value(1, r2Comp[int](h.rec, "T1")),
						k.match(func(t int) bool { return t >= 5 },
							k.derive(h.rec, "hund", tens, tensSet,
								k.value(0, r2Comp[int](h.rec, "H0")),
								r2Bind[int](h.rec, "HB"),
							),
						),
						r2Bind[int](h.rec, "TB"),
					),
				),
				r2Comp[int](h.rec, "BIG"),
			)))
			r2Run(h, "deep", src, "TB=0",
				r2Step[int]{set: 15, want: "T1"},
				r2Step[int]{set: 17, want: "T1", quiet: true},
				r2Step[int]{set: 55, want: "H0"},
				r2Step[int]{set: 250, want: "HB=2"},
				r2Step[int]{set: 299, want: "HB=2", quiet: true},
				r2Step[int]{set: 310, want: "HB=3"},
				r2Step[int]{set: 1500, want: "BIG"},
				r2Step[int]{set: 1501, want: "BIG", quiet: true},
				r2Step[int]{set: 3, want: "TB=0"},
				r2Step[int]{set: 7, want: "TB=0", quiet: true, counts: map[string]int{
					"derive:tens": 1, "derive:hund": 1, "render:T1": 1, "render:H0": 1, "render:BIG": 1, "render:TB": 2,
				}},
			)
		}
	})
}

// Inside a parent Door the route door id changes only on parent reload, and
// the slot disappears on unmount.
func TestRoute2HarnessParentDoor(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	src := NewSource(0)
	p := &Door{}
	p.Inner(context.Background(), r2Slot("in", src.Route(
		RouteValue(1, RouteComp[int](tagElem("section", "sec", textElem("ONE")))),
		r2Bind[int](h.rec, "B"),
	)))
	h.render(r2Slot("outer", mountDoor(p)))
	h.expect("in", "B=0")
	d0 := h.door("in")
	if h.doors0["in"] != d0 || h.doors0["outer"] == d0 {
		t.Fatalf("unexpected render-time doors %v", h.doors0)
	}
	r2Set(h, src, 1)
	h.expect("in", "ONE")
	if h.door("in") != d0 || h.doorCount("in") != 1 {
		t.Fatalf("expected the section to carry door %d, got %q", d0, h.html("in"))
	}
	r2Set(h, src, 2)
	h.expect("in", "B=2")
	h.settle(func(ctx context.Context) { p.Reload(ctx) })
	h.expect("in", "B=2")
	d1 := h.door("in")
	if d1 == d0 {
		t.Fatalf("expected a new route door after reload, still %d", d0)
	}
	r2Set(h, src, 1)
	h.expect("in", "ONE")
	for id, n := range map[uint64]int{d0: 2, d1: 1} {
		acts := h.doorActions(id)
		if len(acts) != n {
			t.Fatalf("door %d: want %d actions, got %+v", id, n, acts)
		}
		for _, a := range acts {
			if a.name != "door_replace" {
				t.Fatalf("door %d got %s", id, a.name)
			}
		}
	}
	h.settle(func(ctx context.Context) { p.Unmount(ctx) })
	if _, ok := h.slot("in"); ok {
		t.Fatalf("expected the inner slot gone, dom:\n%s", h.dom())
	}
	h.expect("outer", "")
	h.counts(map[string]int{"render:B": 3, "clean:B": 3})
}

type (
	r2fI = r2Step[int]
	r2fS = r2Step[string]
	r2fL = r2Step[Location]
	r2fN = map[string]int
)

type r2fState struct {
	Int int
	Str string
}

func r2fStr(s r2fState) string { return s.Str }

func r2fSetStr(s r2fState, v string) r2fState {
	s.Str = v
	return s
}

func r2fStrSet(s r2fState) bool { return s.Str != "" }

func r2fList(s r2fState) []string {
	if s.Str == "" {
		return nil
	}
	return []string{s.Str, strconv.Itoa(s.Int)}
}

func r2fSetList(s r2fState, v []string) r2fState {
	if len(v) == 0 {
		s.Str = ""
		return s
	}
	s.Str = v[0]
	return s
}

func r2fStrSource(s Source[r2fState]) Source[string] { return DeriveSource(s, r2fStr, r2fSetStr) }

func r2fStrBeam(b Beam[r2fState]) Beam[string] { return DeriveBeam(b, r2fStr) }

func r2fListSource(s Source[r2fState]) Source[[]string] {
	return DeriveSourceEqual(s, r2fList, r2fSetList, slices.Equal[[]string])
}

func r2fListBeam(b Beam[r2fState]) Beam[[]string] {
	return DeriveBeamEqual(b, r2fList, slices.Equal[[]string])
}

func r2fEven(v int) bool { return v%2 == 0 }

func r2fPos(v int) bool { return v > 0 }

func r2fSmall(v int) bool { return v < 10 }

type r2fCustomRoute struct{ ID, Tab string }

func r2fIsCustom(l Location) bool { return len(l.Segments) == 2 && l.Segments[0] == "custom" }

func r2fGetCustom(l Location) r2fCustomRoute {
	if !r2fIsCustom(l) {
		return r2fCustomRoute{}
	}
	return r2fCustomRoute{ID: l.Segments[1], Tab: l.Query.Get("tab")}
}

func r2fSetCustom(_ Location, r r2fCustomRoute) Location {
	q := url.Values{}
	if r.Tab != "" {
		q.Set("tab", r.Tab)
	}
	return Location{Segments: []string{"custom", r.ID}, Query: q}
}

type r2fSection string

func r2fRouteDoor(h *r2Harness, name string) {
	h.t.Helper()
	id := h.door(name)
	if want, ok := h.doors0[name]; !ok || id != want {
		h.t.Fatalf("slot %q: route door %d, want render-time door %d (found=%v)", name, id, want, ok)
	}
	for _, a := range h.doorActions(id) {
		if a.name != "door_replace" {
			h.t.Fatalf("slot %q route door %d got %s", name, id, a.name)
		}
	}
}

func r2fSingle(h *r2Harness, name string) {
	h.t.Helper()
	if n := h.doorCount(name); n != 1 {
		h.t.Fatalf("slot %q: want 1 door, got %d, dom:\n%s", name, n, h.dom())
	}
	r2fRouteDoor(h, name)
}

func r2fExpect(h *r2Harness, at string, want map[string]string) {
	h.t.Helper()
	for _, k := range slices.Sorted(func(yield func(string) bool) {
		for k := range want {
			if !yield(k) {
				return
			}
		}
	}) {
		h.expectAt(at, k, want[k])
	}
}

func r2fQuiet(h *r2Harness, names []string, step func()) {
	h.t.Helper()
	before := map[string]int{}
	for _, n := range names {
		before[n] = len(h.doorActions(h.door(n)))
	}
	step()
	for _, n := range names {
		if got := h.doorActions(h.door(n)); len(got) != before[n] {
			h.t.Fatalf("slot %q: expected no new route door actions, got %+v", n, got[before[n]:])
		}
	}
}

func r2fPredDelta(h *r2Harness, seen *int) []string {
	all := h.rec.vals("pred:")
	out := all[*seen:]
	*seen = len(all)
	return out
}

func r2fCombos() [][2]int {
	var out [][2]int
	for ki, k := range r2Kits {
		for ei, e := range r2Entries {
			if e.beam && !k.beam {
				continue
			}
			out = append(out, [2]int{ki, ei})
		}
	}
	return out
}

func r2fNoMatch(t *testing.T, k r2Kit, e r2Entry, workers int) {
	h := newR2HarnessN(t, workers)
	src := NewSource(0)
	h.render(r2Slot("a", e.mount(src, k.value(1, r2Comp[int](h.rec, "ONE")))))
	if got := h.html("a"); !strings.HasPrefix(got, "<d0-r ") || !strings.HasSuffix(got, "></d0-r>") || h.doorCount("a") != 1 {
		t.Fatalf("expected one blank route container, got %q", got)
	}
	id := h.door("a")
	r2Run(h, "a", src, "",
		r2fI{set: 2, want: "", quiet: true, ret: []error{nil}, counts: r2fN{"render:ONE": 0}},
		r2fI{set: 1, want: "ONE", ret: []error{nil}, counts: r2fN{"render:ONE": 1, "clean:ONE": 0}},
		r2fI{set: 3, want: "", counts: r2fN{"render:ONE": 1, "clean:ONE": 1}},
		r2fI{set: 1, want: "ONE", counts: r2fN{"render:ONE": 2, "clean:ONE": 1}},
	)
	acts := h.doorActions(id)
	if len(acts) != 3 {
		t.Fatalf("want 3 replaces, got %+v", acts)
	}
	if b := acts[1].body; acts[1].name != "door_replace" || !strings.HasPrefix(b, "<d0-r ") || r2Text(b) != "" || !slices.Equal(r2DoorIDs(b), []uint64{id}) {
		t.Fatalf("keepEnd should replace door %d with a blank container, got %+v", id, acts[1])
	}
}

func r2fSticky(t *testing.T, k r2Kit, e r2Entry, workers int) {
	h := newR2HarnessN(t, workers)
	src := NewSource(1)
	h.render(r2Slot("a", e.mount(src,
		k.match(r2Pred(h.rec, "even", r2fEven), r2Comp[int](h.rec, "EVEN")),
		k.match(r2Pred(h.rec, "small", r2fSmall), r2Comp[int](h.rec, "SMALL")),
		r2Comp[int](h.rec, "D"),
	)))
	r2Run(h, "a", src, "SMALL",
		r2fI{set: 2, want: "SMALL", quiet: true, ret: []error{nil}},
		r2fI{set: 12, want: "EVEN"},
		r2fI{set: 3, want: "SMALL", counts: r2fN{"render:SMALL": 2, "clean:SMALL": 1}},
		r2fI{set: 13, want: "D"},
		r2fI{set: 11, want: "D", quiet: true},
		r2fI{set: 4, want: "EVEN", counts: r2fN{
			"render:EVEN": 2, "clean:EVEN": 1, "render:SMALL": 2, "clean:SMALL": 2, "render:D": 1, "clean:D": 1,
		}},
	)
	if got := h.rec.vals("pred:"); !slices.Equal(got, r2fStickyPreds) {
		t.Fatalf("predicate log\n got %v\nwant %v", got, r2fStickyPreds)
	}
}

var r2fStickyPreds = []string{
	"even=1", "small=1",
	"small=2",
	"small=12", "even=12",
	"even=3", "even=3", "small=3",
	"small=13", "even=13", "small=13",
	"even=11", "small=11",
	"even=4",
}

func r2fDefault(t *testing.T, k r2Kit, e r2Entry, workers int) {
	h := newR2HarnessN(t, workers)
	src := NewSource(0)
	h.render(r2Slot("a", e.mount(src,
		k.value(1, r2Comp[int](h.rec, "A")),
		k.match(r2Pred(h.rec, "b", func(v int) bool { return v == 2 }), r2Comp[int](h.rec, "B")),
		r2Bind[int](h.rec, "D"),
	)))
	r2Run(h, "a", src, "D=0",
		r2fI{set: 3, want: "D=3", counts: r2fN{"render:D": 2, "clean:D": 1}},
		r2fI{set: 2, want: "B", counts: r2fN{"render:D": 2, "clean:D": 2, "render:B": 1}},
		r2fI{set: 0, want: "D=0", counts: r2fN{"render:D": 3, "clean:B": 1}},
		r2fI{set: 1, want: "A", counts: r2fN{"render:D": 3, "clean:D": 3, "render:A": 1, "render:B": 1}},
	)
	if got, want := h.rec.vals("pred:b="), []string{"0", "3", "2", "0", "0"}; !slices.Equal(got, want) {
		t.Fatalf("pred:b log %v, want %v", got, want)
	}
}

func r2fBind(t *testing.T, k r2Kit, e r2Entry, workers int) {
	h := newR2HarnessN(t, workers)
	src := NewSource(0)
	h.render(r2Slot("a", e.mount(src,
		k.match(func(v int) bool { return v >= 0 }, r2Bind[int](h.rec, "B")),
		r2Comp[int](h.rec, "NEG"),
	)))
	h.expect("a", "B=0")
	id := h.door("a")
	steps := []struct {
		set           int
		want          string
		acts          int
		render, clean int
	}{
		{1, "B=1", 1, 2, 1},
		{2, "B=2", 1, 3, 2},
		{5, "B=5", 1, 4, 3},
		{-1, "NEG", 1, 4, 4},
		{-3, "NEG", 0, 4, 4},
		{3, "B=3", 1, 5, 4},
		{4, "B=4", 1, 6, 5},
	}
	for i, st := range steps {
		n := len(h.doorActions(id))
		r2Set(h, src, st.set)
		acts := h.doorActions(id)
		if len(acts) != n+st.acts {
			t.Fatalf("step %d (set %d): want %d new actions, got %+v", i, st.set, st.acts, acts[n:])
		}
		if st.acts == 1 && r2Text(acts[n].body) != st.want {
			t.Fatalf("step %d (set %d): replace body %q, want text %q", i, st.set, acts[n].body, st.want)
		}
		h.expect("a", st.want)
		h.counts(r2fN{"render:B": st.render, "clean:B": st.clean})
		h.oneDoor("a")
	}
}

// [F01] The first matching route in order renders during the page render.
func TestRoute2FlatFirstMatchInitial(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(2)
		h.render(r2Slot("a", e.mount(src,
			k.value(3, r2Comp[int](h.rec, "A")),
			k.match(r2fPos, r2Comp[int](h.rec, "B")),
			k.match(func(v int) bool { return v > 1 }, r2Comp[int](h.rec, "C")),
			r2Comp[int](h.rec, "D"),
		)))
		inner, ok := r2SlotInner(h.initial, "a")
		if !ok || r2Text(inner) != "B" {
			t.Fatalf("printed page should show B, got %q (found=%v)", inner, ok)
		}
		if acts := h.client.actions(); len(acts) != 0 {
			t.Fatalf("expected no client actions, got %+v", acts)
		}
		h.expect("a", "B")
		h.counts(r2fN{"render:A": 0, "render:B": 1, "render:C": 0, "render:D": 0})
		h.oneDoor("a")
	})
}

// [F02][F20][F21] No match, match, keepEnd, re-match on every kit and entry.
func TestRoute2FlatNoMatchRematch(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		for _, w := range []int{8, 1} {
			r2fNoMatch(t, k, e, w)
		}
	})
}

// [F03] old: source.Route() with no routes. An empty route list renders a
// blank container and ignores updates.
func TestRoute2FlatEmptyRoutes(t *testing.T) {
	for _, e := range r2Entries {
		t.Run(e.name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(0)
			h.render(r2Slot("a", e.mount(src)))
			if got := h.html("a"); !strings.HasPrefix(got, "<d0-r ") || !strings.HasSuffix(got, "></d0-r>") {
				t.Fatalf("expected a blank container, got %q", got)
			}
			r2Run(h, "a", src, "",
				r2fI{set: 1, want: "", quiet: true, ret: []error{nil}},
				r2fI{set: 2, want: "", quiet: true},
				r2fI{set: 2, want: "", quiet: true, ret: []error{}},
				r2fI{set: 0, want: "", quiet: true},
			)
			if n := len(h.doorActions(h.door("a"))); n != 0 {
				t.Fatalf("expected no actions, got %d", n)
			}
		})
	}
}

// [F04][F20][F21] old: sticky active branch. Sticky non-default branches; when
// one ends, the list is re-selected from the top, the ended route included.
func TestRoute2FlatSticky(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		for _, w := range []int{8, 1} {
			r2fSticky(t, k, e, w)
		}
	})
}

// [F05][F20][F21] An active default re-seeks only its earlier siblings; an
// ended non-default branch re-selects the whole list.
func TestRoute2FlatDefaultPriority(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		for _, w := range []int{8, 1} {
			r2fDefault(t, k, e, w)
		}
	})
}

// [F08][F20][F21] RouteBind re-renders on every propagated change.
func TestRoute2FlatBindRerender(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		for _, w := range []int{8, 1} {
			r2fBind(t, k, e, w)
		}
	})
}

// [F04][F05] The key tables on concurrent instances.
func TestRoute2FlatParallel(t *testing.T) {
	combos := r2fCombos()
	r2Parallel(t, 2*len(combos), func(t *testing.T, i int) {
		c := combos[i%len(combos)]
		if i < len(combos) {
			r2fSticky(t, r2Kits[c[0]], r2Entries[c[1]], 4)
		} else {
			r2fDefault(t, r2Kits[c[0]], r2Entries[c[1]], 4)
		}
	})
}

// [F05] Several defaults and positions: the first default shadows everything
// after it, and an active default yields to the first earlier match. An
// ended branch re-selects from the top, so its own predicate runs twice.
func TestRoute2FlatSeveralDefaults(t *testing.T) {
	h := newR2Harness(t)
	a, b, c := NewSource(0), NewSource(0), NewSource(1)
	h.render(
		r2Slot("a", a.Route(
			RouteValue(1, r2Comp[int](h.rec, "A")),
			r2Bind[int](h.rec, "D1"),
			RouteValue(2, r2Comp[int](h.rec, "NEVER")),
			r2Comp[int](h.rec, "D2"),
		)),
		r2Slot("b", b.Route(
			r2Comp[int](h.rec, "D0"),
			RouteValue(1, r2Comp[int](h.rec, "NEVER0")),
			r2Bind[int](h.rec, "NEVERB"),
		)),
		r2Slot("c", c.Route(
			RouteMatch(r2Pred(h.rec, "three", func(v int) bool { return v%3 == 0 }), r2Comp[int](h.rec, "T")),
			RouteMatch(r2Pred(h.rec, "five", func(v int) bool { return v%5 == 0 }), r2Comp[int](h.rec, "F")),
			r2Bind[int](h.rec, "D"),
		)),
	)
	r2Run(h, "a", a, "D1=0",
		r2fI{set: 1, want: "A"},
		r2fI{set: 2, want: "D1=2"},
		r2fI{set: 1, want: "A"},
		r2fI{set: 3, want: "D1=3", counts: r2fN{"render:NEVER": 0, "render:D2": 0, "render:A": 2, "render:D1": 3}},
	)
	r2Run(h, "b", b, "D0",
		r2fI{set: 1, want: "D0", quiet: true},
		r2fI{set: 2, want: "D0", quiet: true, counts: r2fN{"render:D0": 1, "render:NEVER0": 0, "render:NEVERB": 0}},
	)
	seen := len(h.rec.vals("pred:"))
	steps := []struct {
		set   int
		want  string
		preds []string
	}{
		{7, "D=7", []string{"three=7", "five=7"}},
		{15, "T", []string{"three=15"}},
		{10, "F", []string{"three=10", "three=10", "five=10"}},
		{30, "F", []string{"five=30"}},
		{8, "D=8", []string{"five=8", "three=8", "five=8"}},
		{25, "F", []string{"three=25", "five=25"}},
	}
	for i, st := range steps {
		r2Set(h, c, st.set)
		h.expect("c", st.want)
		if got := r2fPredDelta(h, &seen); !slices.Equal(got, st.preds) {
			t.Fatalf("step %d (set %d): preds %v, want %v", i, st.set, got, st.preds)
		}
		h.oneDoor("c")
	}
}

// [F05] A derive branch counts as a default for its parent whenever its
// active child is a default, at any index: the parent re-tries earlier
// siblings on every update. Once the branch switches to a non-default child
// it is sticky again.
func TestRoute2FlatDeriveDefaultIndex(t *testing.T) {
	h := newR2Harness(t)
	a, b := NewSource(0), NewSource(0)
	id := func(s Source[int]) Source[int] { return DeriveSource(s, r2ID, r2IDSet) }
	odd := func(v int) bool { return v%2 != 0 }
	h.render(
		r2Slot("first", a.Route(
			RouteValue(1, r2Comp[int](h.rec, "A")),
			RouteDerive(id, r2Bind[int](h.rec, "DD"), RouteValue(5, r2Comp[int](h.rec, "F"))),
		)),
		r2Slot("later", b.Route(
			RouteMatch(odd, r2Comp[int](h.rec, "A2")),
			RouteDerive(id, RouteMatch(func(v int) bool { return v >= 10 }, r2Comp[int](h.rec, "F2")), r2Bind[int](h.rec, "DL")),
		)),
	)
	r2Run(h, "first", a, "DD=0",
		r2fI{set: 5, want: "DD=5"},
		r2fI{set: 1, want: "A"},
		r2fI{set: 2, want: "DD=2", counts: r2fN{"render:F": 0, "render:A": 1}},
	)
	r2Run(h, "later", b, "DL=0",
		r2fI{set: 1, want: "A2"},
		r2fI{set: 2, want: "DL=2"},
		r2fI{set: 10, want: "F2"},
		r2fI{set: 11, want: "F2", quiet: true, counts: r2fN{"render:A2": 1}},
		r2fI{set: 4, want: "DL=4"},
		r2fI{set: 3, want: "A2", counts: r2fN{"render:A2": 2, "render:F2": 1, "render:DL": 3, "clean:DL": 3}},
	)
}

// [F06] old: routes after an always-matching leaf are never tried. A default
// shadows later siblings and never deactivates itself; routes before it can
// still take over.
func TestRoute2FlatDefaultShadows(t *testing.T) {
	h := newR2Harness(t)
	sb, sc, sm := NewSource(0), NewSource(0), NewSource(0)
	h.render(
		r2Slot("bind", sb.Route(r2Bind[int](h.rec, "D"), RouteValue(1, r2Comp[int](h.rec, "NEVER")))),
		r2Slot("comp", sc.Route(r2Comp[int](h.rec, "DC"), RouteMatch(func(int) bool { return true }, r2Comp[int](h.rec, "NEVER2")))),
		r2Slot("mid", sm.Route(
			RouteMatch(func(v int) bool { return v == 1 }, r2Comp[int](h.rec, "A")),
			r2Comp[int](h.rec, "DM"),
			RouteMatch(func(v int) bool { return v >= 0 }, r2Comp[int](h.rec, "B")),
		)),
	)
	r2Run(h, "bind", sb, "D=0",
		r2fI{set: 1, want: "D=1", counts: r2fN{"render:D": 2}},
		r2fI{set: 2, want: "D=2", counts: r2fN{"render:D": 3}},
		r2fI{set: 1, want: "D=1", counts: r2fN{"render:D": 4, "render:NEVER": 0}},
	)
	r2Run(h, "comp", sc, "DC",
		r2fI{set: 1, want: "DC", quiet: true},
		r2fI{set: 2, want: "DC", quiet: true},
		r2fI{set: 1, want: "DC", quiet: true, counts: r2fN{"render:DC": 1, "render:NEVER2": 0}},
	)
	r2Run(h, "mid", sm, "DM",
		r2fI{set: 1, want: "A"},
		r2fI{set: 2, want: "DM"},
		r2fI{set: 5, want: "DM", quiet: true},
		r2fI{set: 1, want: "A", counts: r2fN{"render:A": 2, "render:DM": 2, "render:B": 0}},
	)
}

// [F07] RouteComp, RouteSource and RouteBeam render once per activation.
func TestRoute2FlatRenderOncePerActivation(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		for _, kind := range []string{"comp", "source", "beam"} {
			if kind == "source" && (k.beam || e.beam) {
				continue
			}
			t.Run(kind, func(t *testing.T) {
				h := newR2Harness(t)
				src := NewSource(0)
				sbox := &r2Box[Source[int]]{}
				bbox := &r2Box[Beam[int]]{}
				var leaf SourceRoute[int]
				switch kind {
				case "comp":
					leaf = r2Comp[int](h.rec, "L")
				case "source":
					leaf = r2Source(h.rec, "L", sbox)
				case "beam":
					leaf = r2Beam(h.rec, "L", bbox)
				}
				h.render(r2Slot("a", e.mount(src, k.match(r2fPos, leaf), r2Comp[int](h.rec, "D"))))
				r2Run(h, "a", src, "D",
					r2fI{set: 1, want: "L", counts: r2fN{"render:L": 1, "clean:D": 1}},
					r2fI{set: 2, want: "L", quiet: true, counts: r2fN{"render:L": 1}},
					r2fI{set: 3, want: "L", quiet: true, counts: r2fN{"render:L": 1, "clean:L": 0}},
					r2fI{set: 0, want: "D", counts: r2fN{"render:L": 1, "clean:L": 1, "render:D": 2}},
					r2fI{set: 5, want: "L", counts: r2fN{"render:L": 2, "clean:L": 1, "clean:D": 2}},
					r2fI{set: 6, want: "L", quiet: true, counts: r2fN{"render:L": 2}},
				)
				switch kind {
				case "source":
					if sbox.len() != 2 || sbox.at(0) != sbox.at(1) {
						t.Fatalf("want the same Source on both activations, got %d captures", sbox.len())
					}
					if v := sbox.last().Get(); v != 6 {
						t.Fatalf("captured source holds %d, want 6", v)
					}
				case "beam":
					if bbox.len() != 2 || bbox.at(0) != bbox.at(1) {
						t.Fatalf("want the same Beam on both activations, got %d captures", bbox.len())
					}
					if v := bbox.last().Get(); v != 6 {
						t.Fatalf("captured beam holds %d, want 6", v)
					}
				}
			})
		}
	})
}

// [F09] old: RouteDerive(f).Bind uses the derived comparator. RouteBind
// compares with the routed beam's comparator, bound at init.
func TestRoute2FlatBindDerivedComparator(t *testing.T) {
	tens := func(v int) int { return v / 10 }
	tensSet := func(a, b int) int { return b*10 + a%10 }
	parity := func(a, b int) bool { return (a-b)%2 == 0 }
	cases := []struct {
		name  string
		src   func(Source[int]) Source[int]
		beam  func(Beam[int]) Beam[int]
		steps []r2fI
	}{
		{
			name: "eq",
			src:  func(s Source[int]) Source[int] { return DeriveSource(s, tens, tensSet) },
			beam: func(b Beam[int]) Beam[int] { return DeriveBeam(b, tens) },
			steps: []r2fI{
				{set: 1, want: "B=0", quiet: true, counts: r2fN{"render:B": 1}},
				{set: 5, want: "B=0", quiet: true, counts: r2fN{"render:B": 1}},
				{set: 12, want: "B=1", counts: r2fN{"render:B": 2}},
				{set: 19, want: "B=1", quiet: true, counts: r2fN{"render:B": 2}},
				{set: 25, want: "B=2", counts: r2fN{"render:B": 3, "clean:B": 2}},
			},
		},
		{
			name: "custom",
			src:  func(s Source[int]) Source[int] { return DeriveSourceEqual(s, r2ID, r2IDSet, parity) },
			beam: func(b Beam[int]) Beam[int] { return DeriveBeamEqual(b, r2ID, parity) },
			steps: []r2fI{
				{set: 2, want: "B=0", quiet: true, counts: r2fN{"render:B": 1}},
				{set: 3, want: "B=3", counts: r2fN{"render:B": 2}},
				{set: 5, want: "B=3", quiet: true, counts: r2fN{"render:B": 2}},
				{set: 4, want: "B=4", counts: r2fN{"render:B": 3}},
				{set: 10, want: "B=4", quiet: true, counts: r2fN{"render:B": 3}},
			},
		},
		{
			name: "nil",
			src:  func(s Source[int]) Source[int] { return DeriveSourceEqual(s, tens, tensSet, nil) },
			beam: func(b Beam[int]) Beam[int] { return DeriveBeamEqual(b, tens, nil) },
			steps: []r2fI{
				{set: 1, want: "B=0", counts: r2fN{"render:B": 2}},
				{set: 5, want: "B=0", counts: r2fN{"render:B": 3}},
				{set: 12, want: "B=1", counts: r2fN{"render:B": 4}},
				{set: 13, want: "B=1", counts: r2fN{"render:B": 5, "clean:B": 4}},
			},
		},
	}
	for _, c := range cases {
		for _, beam := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/beam=%v", c.name, beam), func(t *testing.T) {
				h := newR2Harness(t)
				src := NewSource(0)
				if beam {
					h.render(r2Slot("a", src.RouteBeam(RouteBeamDerive(r2DeriveBeam(h.rec, "d", c.beam), r2Bind[int](h.rec, "B")))))
				} else {
					h.render(r2Slot("a", src.Route(RouteDerive(r2Derive(h.rec, "d", c.src), r2Bind[int](h.rec, "B")))))
				}
				r2Run(h, "a", src, "B=0", c.steps...)
				loud := 0
				for _, st := range c.steps {
					if !st.quiet {
						loud++
					}
				}
				if n := len(h.doorActions(h.door("a"))); n != loud {
					t.Fatalf("want exactly one replace per loud step (%d), got %d", loud, n)
				}
				h.counts(r2fN{"derive:d": 1})
			})
		}
	}
}

// [F10] old: RouteDeriveEqual(derive, nil) propagates every update. A root
// with a nil comparator propagates equal values: Bind re-renders, Comp stays
// quiet and value routes never match.
func TestRoute2FlatNilComparatorRoot(t *testing.T) {
	h := newR2Harness(t)
	src := NewSourceEqual(1, nil)
	h.render(
		r2Slot("bind", src.Route(r2Bind[int](h.rec, "B"))),
		r2Slot("comp", src.Route(r2Comp[int](h.rec, "C"))),
		r2Slot("value", src.Route(RouteValue(1, r2Comp[int](h.rec, "X")), r2Comp[int](h.rec, "D"))),
		r2Slot("bvalue", src.RouteBeam(RouteBeamValue(1, r2Comp[int](h.rec, "XB")), r2Bind[int](h.rec, "DB"))),
		r2Slot("none", src.Route(RouteValue(1, r2Comp[int](h.rec, "XN")))),
	)
	r2fExpect(h, "initial: ", map[string]string{"bind": "B=1", "comp": "C", "value": "D", "bvalue": "DB=1", "none": ""})
	bindID := h.door("bind")
	for i, v := range []int{1, 1, 2, 2, 1} {
		at := fmt.Sprintf("step %d (set %d): ", i, v)
		n := len(h.doorActions(bindID))
		var ret []error
		r2fQuiet(h, []string{"comp", "value", "none"}, func() { ret = r2Set(h, src, v) })
		if !r2SameErrs(ret, []error{nil}) {
			t.Fatalf("%supdate returned %v, want [nil]", at, ret)
		}
		if got := len(h.doorActions(bindID)); got != n+1 {
			t.Fatalf("%sbind: want one new action, got %d", at, got-n)
		}
		want := strconv.Itoa(v)
		r2fExpect(h, at, map[string]string{"bind": "B=" + want, "comp": "C", "value": "D", "bvalue": "DB=" + want, "none": ""})
		h.countsAt(at, r2fN{"render:B": i + 2, "render:DB": i + 2, "render:C": 1, "render:D": 1, "render:X": 0, "render:XB": 0, "render:XN": 0})
		for _, s := range []string{"bind", "comp", "value", "bvalue", "none"} {
			h.oneDoorAt(at, s)
		}
	}
}

// [F11] old: RouteValue(v) matches with ==. Value routes on int and string.
func TestRoute2FlatValueComparable(t *testing.T) {
	h := newR2Harness(t)
	n, s := NewSource(0), NewSource("a")
	h.render(
		r2Slot("int", n.Route(RouteValue(1, r2Comp[int](h.rec, "ONE")), RouteValue(2, r2Comp[int](h.rec, "TWO")))),
		r2Slot("str", s.RouteBeam(RouteBeamValue("x", r2Comp[string](h.rec, "X")), RouteBeamValue("y", r2Comp[string](h.rec, "Y")), r2Bind[string](h.rec, "S"))),
	)
	r2Run(h, "int", n, "",
		r2fI{set: 1, want: "ONE"},
		r2fI{set: 2, want: "TWO"},
		r2fI{set: 3, want: ""},
		r2fI{set: -2, want: "", quiet: true},
		r2fI{set: 2, want: "TWO", counts: r2fN{"render:ONE": 1, "render:TWO": 2, "clean:TWO": 1}},
	)
	r2Run(h, "str", s, "S=a",
		r2fS{set: "x", want: "X"},
		r2fS{set: "y", want: "Y"},
		r2fS{set: "X", want: "S=X"},
		r2fS{set: "z", want: "S=z"},
		r2fS{set: "x", want: "X", counts: r2fN{"render:X": 2, "render:Y": 1, "render:S": 3}},
	)
}

// [F12] Value routes with a custom comparator, including a non-comparable T
// and a comparator of a derived source.
func TestRoute2FlatValueCustomComparator(t *testing.T) {
	t.Run("fold", func(t *testing.T) {
		h := newR2Harness(t)
		src := NewSourceEqual("abc", strings.EqualFold)
		h.render(r2Slot("a", src.Route(RouteValue("ABC", r2Comp[string](h.rec, "X")), r2Bind[string](h.rec, "S"))))
		r2Run(h, "a", src, "X",
			r2fS{set: "ABC", want: "X", quiet: true, ret: []error{}},
			r2fS{set: "abd", want: "S=abd", ret: []error{nil}},
			r2fS{set: "aBc", want: "X", counts: r2fN{"render:X": 2, "render:S": 1}},
		)
	})
	t.Run("location", func(t *testing.T) {
		h := newR2Harness(t)
		src := r2LocSource("/b")
		h.render(r2Slot("a", src.Route(RouteValue(r2Loc("/a?x=1"), r2Comp[Location](h.rec, "X")), r2Bind[Location](h.rec, "L"))))
		r2Run(h, "a", src, "L=/b",
			r2fL{set: r2Loc("/a?x=1"), want: "X"},
			r2fL{set: r2Loc("/a?x=1"), want: "X", quiet: true, ret: []error{}},
			r2fL{set: r2Loc("/a?x=2"), want: "L=/a?x=2"},
			r2fL{set: r2Loc("/a?x=1"), want: "X", counts: r2fN{"render:X": 2, "render:L": 2}},
		)
	})
	mod := func(a, b int) bool { return a%10 == b%10 }
	for _, beam := range []bool{false, true} {
		t.Run(fmt.Sprintf("derived/beam=%v", beam), func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(0)
			if beam {
				h.render(r2Slot("a", src.RouteBeam(RouteBeamDerive(func(b Beam[int]) Beam[int] {
					return DeriveBeamEqual(b, r2ID, mod)
				}, RouteBeamValue(3, r2Comp[int](h.rec, "X")), r2Bind[int](h.rec, "B")))))
			} else {
				h.render(r2Slot("a", src.Route(RouteDerive(func(s Source[int]) Source[int] {
					return DeriveSourceEqual(s, r2ID, r2IDSet, mod)
				}, RouteValue(3, r2Comp[int](h.rec, "X")), r2Bind[int](h.rec, "B")))))
			}
			r2Run(h, "a", src, "B=0",
				r2fI{set: 13, want: "X"},
				r2fI{set: 23, want: "X", quiet: true},
				r2fI{set: 4, want: "B=4"},
				r2fI{set: 3, want: "X", counts: r2fN{"render:X": 2, "render:B": 2}},
			)
		})
	}
}

// [F13] old: nil comparator. Value routes over a nil comparator never match.
func TestRoute2FlatValueNilComparator(t *testing.T) {
	h := newR2Harness(t)
	root, ds, db, none := NewSourceEqual(1, nil), NewSource(1), NewSource(1), NewSourceEqual(1, nil)
	h.render(
		r2Slot("root", root.Route(RouteValue(1, r2Comp[int](h.rec, "X1")), r2Bind[int](h.rec, "B1"))),
		r2Slot("ds", ds.Route(RouteDerive(func(s Source[int]) Source[int] {
			return DeriveSourceEqual(s, r2ID, r2IDSet, nil)
		}, RouteValue(1, r2Comp[int](h.rec, "X2")), r2Bind[int](h.rec, "B2")))),
		r2Slot("db", db.RouteBeam(RouteBeamDerive(func(b Beam[int]) Beam[int] {
			return DeriveBeamEqual(b, r2ID, nil)
		}, RouteBeamValue(1, r2Comp[int](h.rec, "X3")), r2Bind[int](h.rec, "B3")))),
		r2Slot("none", none.Route(RouteValue(1, r2Comp[int](h.rec, "X4")))),
	)
	r2Run(h, "root", root, "B1=1",
		r2fI{set: 1, want: "B1=1", ret: []error{nil}, counts: r2fN{"render:B1": 2}},
		r2fI{set: 2, want: "B1=2"},
		r2fI{set: 1, want: "B1=1", counts: r2fN{"render:X1": 0, "render:B1": 4}},
	)
	r2Run(h, "ds", ds, "B2=1",
		r2fI{set: 2, want: "B2=2"},
		r2fI{set: 1, want: "B2=1", counts: r2fN{"render:X2": 0}},
	)
	r2Run(h, "db", db, "B3=1",
		r2fI{set: 2, want: "B3=2"},
		r2fI{set: 1, want: "B3=1", counts: r2fN{"render:X3": 0}},
	)
	r2Run(h, "none", none, "",
		r2fI{set: 1, want: "", quiet: true, ret: []error{nil}},
		r2fI{set: 2, want: "", quiet: true},
		r2fI{set: 1, want: "", quiet: true, counts: r2fN{"render:X4": 0}},
	)
}

// [F14] old: nil render results. A nil Elem renders a blank container, an
// active nil non-default branch stays sticky and a nil default still yields.
func TestRoute2FlatNilElem(t *testing.T) {
	h := newR2Harness(t)
	sBind, sSrc, sBeam, sComp, sNilBind, sKit := NewSource(0), NewSource(0), NewSource(0), NewSource(0), NewSource(0), NewSource(0)
	nilSource := func(label string) SourceRoute[int] {
		return RouteSource(func(Source[int]) gox.Elem { h.rec.hit("render:" + label); return nil })
	}
	nilBeam := func(label string) BeamRoute[int] {
		return RouteBeam(func(Beam[int]) gox.Elem { h.rec.hit("render:" + label); return nil })
	}
	h.render(
		r2Slot("bindif", sBind.Route(RouteMatch(r2fPos, r2BindIf[int](h.rec, "B", r2fEven)), r2Comp[int](h.rec, "D1"))),
		r2Slot("src", sSrc.Route(RouteMatch(r2fPos, nilSource("S")), r2Comp[int](h.rec, "D2"))),
		r2Slot("beam", sBeam.Route(RouteMatch(r2fPos, nilBeam("M")), r2Comp[int](h.rec, "D3"))),
		r2Slot("comp", sComp.Route(RouteValue(1, r2Comp[int](h.rec, "A")), RouteComp[int](nil))),
		r2Slot("nilbind", sNilBind.Route(RouteValue(1, r2Comp[int](h.rec, "A2")), RouteBind(func(int) gox.Elem { h.rec.hit("render:N"); return nil }))),
		r2Slot("kit", sKit.RouteBeam(RouteBeamMatch(r2fPos, nilBeam("MK")), RouteComp[int](nil))),
	)
	r2Run(h, "bindif", sBind, "D1",
		r2fI{set: 1, want: "", counts: r2fN{"render:B": 1, "clean:D1": 1, "clean:B": 0}},
		r2fI{set: 2, want: "B=2", counts: r2fN{"render:B": 2}},
		r2fI{set: 3, want: "", counts: r2fN{"render:B": 3, "clean:B": 1}},
		r2fI{set: 0, want: "D1", counts: r2fN{"render:D1": 2, "clean:B": 1}},
	)
	for _, c := range []struct {
		slot, leaf, def string
		src             Source[int]
	}{{"src", "S", "D2", sSrc}, {"beam", "M", "D3", sBeam}} {
		r2Run(h, c.slot, c.src, c.def,
			r2fI{set: 1, want: "", counts: r2fN{"render:" + c.leaf: 1, "clean:" + c.def: 1}},
			r2fI{set: 2, want: "", quiet: true, counts: r2fN{"render:" + c.leaf: 1, "render:" + c.def: 1}},
			r2fI{set: 0, want: c.def, counts: r2fN{"render:" + c.def: 2}},
			r2fI{set: 5, want: "", counts: r2fN{"render:" + c.leaf: 2}},
		)
	}
	r2Run(h, "comp", sComp, "",
		r2fI{set: 2, want: "", quiet: true},
		r2fI{set: 1, want: "A"},
		r2fI{set: 0, want: ""},
		r2fI{set: 1, want: "A", counts: r2fN{"render:A": 2, "clean:A": 1}},
	)
	r2Run(h, "nilbind", sNilBind, "",
		r2fI{set: 2, want: "", counts: r2fN{"render:N": 2}},
		r2fI{set: 1, want: "A2", counts: r2fN{"render:A2": 1}},
		r2fI{set: 3, want: "", counts: r2fN{"render:N": 3, "clean:A2": 1}},
	)
	r2Run(h, "kit", sKit, "",
		r2fI{set: 1, want: "", counts: r2fN{"render:MK": 1}},
		r2fI{set: 2, want: "", quiet: true},
		r2fI{set: 0, want: ""},
		r2fI{set: -1, want: "", quiet: true},
		r2fI{set: 4, want: "", counts: r2fN{"render:MK": 2}},
	)
}

// [F15] A single-element branch is adopted as the route container.
func TestRoute2FlatAdoptedElement(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(1)
	h.render(r2Slot("a", src.Route(
		RouteValue(1, RouteComp[int](tagElem("section", "sec", textElem("S")))),
		r2Bind[int](h.rec, "B"),
	)))
	id := h.door("a")
	got := h.html("a")
	if !strings.HasPrefix(got, `<section data-d0p="`) || !strings.Contains(got, fmt.Sprintf(` data-d0r="%d"`, id)) || !strings.Contains(got, ` id="sec"`) {
		t.Fatalf("expected an adopted section, got %q", got)
	}
	r2Run(h, "a", src, "S",
		r2fI{set: 2, want: "B=2"},
		r2fI{set: 3, want: "B=3"},
		r2fI{set: 1, want: "S"},
		r2fI{set: 4, want: "B=4"},
	)
	acts := h.doorActions(id)
	if len(acts) != 4 {
		t.Fatalf("want 4 replaces, got %+v", acts)
	}
	if b := acts[0].body; !strings.HasPrefix(b, "<d0-r ") || !strings.Contains(b, fmt.Sprintf(` id="d0r%d"`, id)) {
		t.Fatalf("switch to text should restore <d0-r id=d0r%d>, got %q", id, b)
	}
	if b := acts[2].body; !strings.HasPrefix(b, "<section ") || !strings.Contains(b, fmt.Sprintf(` data-d0r="%d"`, id)) {
		t.Fatalf("switch back should adopt the section as door %d, got %q", id, b)
	}
}

// [F16] Suppressed updates do not reach routing.
func TestRoute2FlatSuppressed(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(1)
	ci := NewSourceEqual("a", strings.EqualFold)
	h.render(
		r2Slot("a", src.Route(RouteMatch(r2Pred(h.rec, "p", r2fPos), r2Bind[int](h.rec, "B")), r2Comp[int](h.rec, "D"))),
		r2Slot("ci", ci.Route(RouteMatch(r2Pred(h.rec, "ci", func(string) bool { return true }), r2Bind[string](h.rec, "C")))),
	)
	r2Set(h, src, 22)
	h.expect("a", "B=22")
	log, acts := h.rec.all(), len(h.client.actions())
	tens := DeriveSource(src, func(v int) int { return v / 10 }, func(a, b int) int { return b*10 + a%10 })
	for i, op := range []func() []error{
		func() []error { return r2Set(h, src, 22) },
		func() []error { return r2Mutate(h, src, func(v int) int { return v }) },
		func() []error { return r2Set(h, tens, 2) },
		func() []error { return r2Set(h, ci, "A") },
		func() []error { return r2Mutate(h, ci, strings.ToUpper) },
	} {
		if ret := op(); len(ret) != 0 {
			t.Fatalf("op %d: want a suppressed update, got %v", i, ret)
		}
		if got := h.rec.all(); !slices.Equal(got, log) {
			t.Fatalf("op %d: recorder moved: %v -> %v", i, log, got)
		}
		if got := len(h.client.actions()); got != acts {
			t.Fatalf("op %d: actions moved from %d to %d", i, acts, got)
		}
	}
	r2fExpect(h, "", map[string]string{"a": "B=22", "ci": "C=a"})
	h.oneDoor("a")
	h.oneDoor("ci")
}

// [F17] Leaves receive a live Source or Beam that tracks later updates.
func TestRoute2FlatLiveLeafValue(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(0)
	sbox, dbox := &r2Box[Source[int]]{}, &r2Box[Source[int]]{}
	bbox, kbox := &r2Box[Beam[int]]{}, &r2Box[Beam[int]]{}
	h.render(
		r2Slot("s", src.Route(RouteMatch(r2fPos, r2Source(h.rec, "S", sbox)), r2Comp[int](h.rec, "D1"))),
		r2Slot("b", src.Route(RouteMatch(r2fPos, r2Beam(h.rec, "B", bbox)), r2Comp[int](h.rec, "D2"))),
		r2Slot("k", src.RouteBeam(RouteBeamMatch(r2fPos, r2Beam(h.rec, "K", kbox)), r2Comp[int](h.rec, "D3"))),
		r2Slot("d", src.Route(RouteDerive(r2Div(10), RouteMatch(r2fPos, r2Source(h.rec, "T", dbox))), r2Comp[int](h.rec, "D4"))),
	)
	live := func(at string, want, tens int) {
		t.Helper()
		for name, b := range map[string]Beam[int]{"s": sbox.last(), "b": bbox.last(), "k": kbox.last()} {
			if v := b.Get(); v != want {
				t.Fatalf("%s%s: Get %d, want %d", at, name, v, want)
			}
			if v, ok := b.Read(h.page); !ok || v != want {
				t.Fatalf("%s%s: Read %d %v, want %d", at, name, v, ok, want)
			}
		}
		if v, ok := dbox.last().Read(h.page); !ok || v != tens || dbox.last().Get() != tens {
			t.Fatalf("%sderived: Read %d %v, want %d", at, v, ok, tens)
		}
	}
	active := map[string]string{"s": "S", "b": "B", "k": "K", "d": "T"}
	idle := map[string]string{"s": "D1", "b": "D2", "k": "D3", "d": "D4"}
	for i, st := range []struct {
		set, tens, renders, dRenders int
		on                           bool
	}{
		{12, 1, 1, 1, true},
		{15, 1, 1, 1, true},
		{27, 2, 1, 1, true},
		{0, 0, 1, 1, false},
		{31, 3, 2, 2, true},
		{38, 3, 2, 2, true},
	} {
		at := fmt.Sprintf("step %d (set %d): ", i, st.set)
		r2Set(h, src, st.set)
		if st.on {
			r2fExpect(h, at, active)
			live(at, st.set, st.tens)
		} else {
			r2fExpect(h, at, idle)
		}
		h.countsAt(at, r2fN{"render:S": st.renders, "render:B": st.renders, "render:K": st.renders, "render:T": st.dRenders})
	}
	for name, n := range map[string]int{"s": sbox.len(), "b": bbox.len(), "k": kbox.len(), "d": dbox.len()} {
		if n != 2 {
			t.Fatalf("%s: want 2 captures, got %d", name, n)
		}
	}
	if sbox.at(0) != sbox.at(1) || bbox.at(0) != bbox.at(1) || kbox.at(0) != kbox.at(1) || dbox.at(0) != dbox.at(1) {
		t.Fatal("captured Source/Beam should be the same object across activations")
	}
	if sbox.last() != src {
		t.Fatal("a RouteSource directly under RouteMatch should receive the entry source")
	}
}

// [F18] A leaf render error leaves the old content on the client and the
// route keeps working.
func TestRoute2FlatLeafRenderError(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	src := NewSource(0)
	boom := RouteComp[int](gox.Elem(func(gox.Cursor) error {
		h.rec.hit("render:ERR")
		return errors.New("r2f boom")
	}))
	h.render(r2Slot("a", src.Route(RouteValue(1, boom), r2Bind[int](h.rec, "D"))))
	id := h.door("a")
	r2Set(h, src, 1)
	h.expect("a", "D=0")
	if n := len(h.doorActions(id)); n != 0 {
		t.Fatalf("failed render should send nothing, got %d actions", n)
	}
	if n := h.waitLogged("door rendering error", 1); n != 1 {
		t.Fatalf("want one render error log, got %d: %v", n, h.logs.all())
	}
	r2Set(h, src, 2)
	h.expect("a", "D=2")
	select {
	case <-h.inst.ended:
		t.Fatal("instance ended after a branch render error")
	default:
	}
	r2fSingle(h, "a")
	if h.door("a") != id {
		t.Fatalf("route door changed from %d to %d", id, h.door("a"))
	}
	h.counts(r2fN{"render:ERR": 1, "render:D": 2})
}

// [F20] A route on a non-identity derived entry only runs when the derived
// value changes. When ODD ends, the re-selection checks it again before T.
func TestRoute2FlatDerivedEntryNonIdentity(t *testing.T) {
	tens := func(v int) int { return v / 10 }
	tensSet := func(a, b int) int { return b*10 + a%10 }
	odd := func(v int) bool { return v%2 != 0 }
	for _, c := range []struct {
		name  string
		mount func(h *r2Harness, src Source[int]) gox.Elem
	}{
		{"derivedSource.Route", func(h *r2Harness, src Source[int]) gox.Elem {
			return DeriveSource(src, tens, tensSet).Route(RouteMatch(r2Pred(h.rec, "odd", odd), r2Comp[int](h.rec, "ODD")), r2Bind[int](h.rec, "T"))
		}},
		{"derivedSource.RouteBeam", func(h *r2Harness, src Source[int]) gox.Elem {
			return DeriveSource(src, tens, tensSet).RouteBeam(RouteBeamMatch(r2Pred(h.rec, "odd", odd), r2Comp[int](h.rec, "ODD")), r2Bind[int](h.rec, "T"))
		}},
		{"derivedBeam.RouteBeam", func(h *r2Harness, src Source[int]) gox.Elem {
			return DeriveBeam(src, tens).RouteBeam(RouteBeamMatch(r2Pred(h.rec, "odd", odd), r2Comp[int](h.rec, "ODD")), r2Bind[int](h.rec, "T"))
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(0)
			h.render(r2Slot("a", c.mount(h, src)))
			seen := 0
			for i, st := range []struct {
				set   int
				want  string
				quiet bool
				preds []string
			}{
				{0, "T=0", false, []string{"0"}},
				{5, "T=0", true, nil},
				{12, "ODD", false, []string{"1"}},
				{19, "ODD", true, nil},
				{25, "T=2", false, []string{"2", "2"}},
				{29, "T=2", true, nil},
				{31, "ODD", false, []string{"3"}},
			} {
				at := fmt.Sprintf("step %d (set %d): ", i, st.set)
				if i > 0 {
					do := func() { r2Set(h, src, st.set) }
					if st.quiet {
						h.quietAt(at, "a", do)
					} else {
						do()
					}
				}
				h.expectAt(at, "a", st.want)
				all := h.rec.vals("pred:odd=")
				if got := all[seen:]; !slices.Equal(got, st.preds) {
					t.Fatalf("%spreds %v, want %v", at, got, st.preds)
				}
				seen = len(all)
				h.oneDoorAt(at, "a")
			}
		})
	}
}

// RouteMatch with no nested routes never matches: selection falls through to
// later siblings, and with nothing else matching the container is blank. A
// blank container stays untouched while it remains blank; content to blank and
// blank to content each take exactly one replace.
func TestRoute2FlatMatchNoRoutes(t *testing.T) {
	h := newR2Harness(t)
	a, b, c := NewSource(0), NewSource(0), NewSource(-1)
	h.render(
		r2Slot("a", a.Route(RouteMatch[int](r2fPos), r2Comp[int](h.rec, "D"))),
		r2Slot("b", b.RouteBeam(RouteBeamMatch[int](r2fPos), RouteBeamMatch(func(v int) bool { return v > 5 }, r2Comp[int](h.rec, "LATE")))),
		r2Slot("c", c.Route(RouteMatch[int](r2fEven), RouteMatch(r2fPos, RouteValue(5, r2Comp[int](h.rec, "X"))))),
	)
	replaces := func(slot string, want ...string) {
		t.Helper()
		id := h.door(slot)
		acts := h.doorActions(id)
		got := make([]string, len(acts))
		for i, a := range acts {
			if a.name != "door_replace" || !slices.Equal(r2DoorIDs(a.body), []uint64{id}) {
				t.Fatalf("slot %q: want a replace of door %d, got %+v", slot, id, a)
			}
			got[i] = r2Text(a.body)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("slot %q: replaces %q, want %q", slot, got, want)
		}
	}
	r2Run(h, "a", a, "D",
		r2fI{set: 1, want: "D", quiet: true},
		r2fI{set: 2, want: "D", quiet: true},
		r2fI{set: -1, want: "D", quiet: true},
		r2fI{set: 3, want: "D", quiet: true, counts: r2fN{"render:D": 1, "clean:D": 0}},
	)
	replaces("a")
	r2Run(h, "b", b, "",
		r2fI{set: 7, want: "LATE"},
		r2fI{set: 9, want: "LATE", quiet: true},
		r2fI{set: 0, want: "", counts: r2fN{"clean:LATE": 1}},
		r2fI{set: -1, want: "", quiet: true},
		r2fI{set: 8, want: "LATE", counts: r2fN{"render:LATE": 2, "clean:LATE": 1}},
	)
	replaces("b", "LATE", "", "LATE")
	r2Run(h, "c", c, "",
		r2fI{set: 2, want: "", quiet: true},
		r2fI{set: 3, want: "", quiet: true},
		r2fI{set: 5, want: "X"},
		r2fI{set: 7, want: "", counts: r2fN{"clean:X": 1}},
		r2fI{set: -1, want: "", quiet: true},
		r2fI{set: 5, want: "X"},
		r2fI{set: -3, want: "", counts: r2fN{"render:X": 2, "clean:X": 2}},
		r2fI{set: 4, want: "", quiet: true},
	)
	replaces("c", "X", "", "X", "")
}

// RouteMatch(p, RouteComp[T](nil)) is the explicit way to render nothing for
// a range: it matches, so it blocks later siblings and blanks the container,
// and it is sticky like any non-default branch. A blank container stays
// untouched while it remains blank.
func TestRoute2FlatNilMatchClaims(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(1)
		h.render(r2Slot("a", e.mount(src,
			k.match(r2Pred(h.rec, "even", r2fEven), r2Comp[int](h.rec, "EVEN")),
			k.match(r2Pred(h.rec, "pos", r2fPos), RouteComp[int](nil)),
			k.match(r2Pred(h.rec, "late", func(v int) bool { return v > 5 }), r2Comp[int](h.rec, "LATE")),
			r2Bind[int](h.rec, "D"),
		)))
		r2fPredRun(h, "a", src, "", []string{"even=1", "pos=1"},
			r2fPI{set: 2, want: "", preds: []string{"pos=2"}},
			r2fPI{set: 7, want: "", preds: []string{"pos=7"}},
			r2fPI{set: 0, want: "EVEN", acts: 1, preds: []string{"pos=0", "even=0"}},
			r2fPI{set: 3, want: "", acts: 1, preds: []string{"even=3", "even=3", "pos=3"}},
			r2fPI{set: 9, want: "", preds: []string{"pos=9"}},
			r2fPI{set: -1, want: "D=-1", acts: 1, preds: []string{"pos=-1", "even=-1", "pos=-1", "late=-1"}},
			r2fPI{set: 5, want: "", acts: 1, preds: []string{"even=5", "pos=5"}},
			r2fPI{set: -3, want: "D=-3", acts: 1, preds: []string{"pos=-3", "even=-3", "pos=-3", "late=-3"}},
			r2fPI{set: -5, want: "D=-5", acts: 1, preds: []string{"even=-5", "pos=-5", "late=-5"}},
		)
		h.counts(r2fN{"render:EVEN": 1, "clean:EVEN": 1, "render:LATE": 0, "render:D": 3, "clean:D": 2})
	})
}

// A match whose nested routes match nothing does not match: selection falls
// through to the later siblings, both on the first render and when the
// container re-selects after a break. A branch selected by falling through is
// as sticky as any other: the earlier match is not re-tried while it holds.
func TestRoute2FlatUnmatchedChildrenFallThrough(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(1)
		h.render(r2Slot("a", e.mount(src,
			k.match(r2Pred(h.rec, "outer", r2fPos), k.value(5, r2Comp[int](h.rec, "X"))),
			k.match(r2Pred(h.rec, "small", r2fSmall), r2Comp[int](h.rec, "S")),
			r2Comp[int](h.rec, "D"),
		)))
		r2fPredRun(h, "a", src, "S", []string{"outer=1", "small=1"},
			r2fPI{set: 5, want: "S", preds: []string{"small=5"}},
			r2fPI{set: 12, want: "D", acts: 1, preds: []string{"small=12", "outer=12", "small=12"}},
			r2fPI{set: 5, want: "X", acts: 1, preds: []string{"outer=5"}},
			r2fPI{set: 6, want: "S", acts: 1, preds: []string{"outer=6", "outer=6", "small=6"}},
			r2fPI{set: 15, want: "D", acts: 1, preds: []string{"small=15", "outer=15", "small=15"}},
			r2fPI{set: 0, want: "S", acts: 1, preds: []string{"outer=0", "small=0"}},
			r2fPI{set: -5, want: "S", preds: []string{"small=-5"}},
			r2fPI{set: 5, want: "S", preds: []string{"small=5"}},
		)
		h.counts(r2fN{"render:X": 1, "clean:X": 1, "render:S": 3, "clean:S": 2, "render:D": 2, "clean:D": 2})
	})
}

// Derive runs once per route instance, even for branches that never match,
// and a route list is a template: each render gets its own state.
func TestRoute2FlatTemplateReuse(t *testing.T) {
	h := newR2Harness(t)
	a, b := NewSource(0), NewSource(0)
	routes := []SourceRoute[int]{
		RouteValue(1, r2Comp[int](h.rec, "ONE")),
		RouteDerive(r2Derive(h.rec, "never", r2Div(10)), RouteValue(99, r2Comp[int](h.rec, "NEVER"))),
		RouteMatch(r2fPos, RouteDerive(r2Derive(h.rec, "tens", r2Div(10)), r2Bind[int](h.rec, "T"))),
		r2Bind[int](h.rec, "D"),
	}
	h.render(r2Slot("a", a.Route(routes...)), r2Slot("b", b.Route(routes...)))
	h.counts(r2fN{"derive:never": 2, "derive:tens": 2})
	h.quiet("b", func() {
		r2Run(h, "a", a, "D=0",
			r2fI{set: 1, want: "ONE"},
			r2fI{set: 25, want: "T=2"},
			r2fI{set: 29, want: "T=2", quiet: true},
			r2fI{set: 0, want: "D=0"},
		)
	})
	h.expect("b", "D=0")
	r2Run(h, "b", b, "D=0",
		r2fI{set: 31, want: "T=3"},
		r2fI{set: 1, want: "T=0"},
	)
	h.expect("a", "D=0")
	h.counts(r2fN{"derive:never": 2, "derive:tens": 2, "render:NEVER": 0})
}

// Bursts end in the right branch; a NoSkip source sees every value.
func TestRoute2FlatBurst(t *testing.T) {
	burst := func(h *r2Harness, src Source[int], vals ...int) {
		h.t.Helper()
		chs := make([]<-chan error, 0, len(vals))
		h.settle(func(ctx context.Context) {
			for _, v := range vals {
				chs = append(chs, src.Update(ctx, v))
			}
		})
		for i, ch := range chs {
			ret := drainOp(h.t, ch)
			if i == len(chs)-1 {
				if !r2SameErrs(ret, []error{nil}) {
					h.t.Fatalf("last update returned %v", ret)
				}
				continue
			}
			if !r2SameErrs(ret, []error{nil}) && !r2SameErrs(ret, []error{context.Canceled}) {
				h.t.Fatalf("update %d returned %v", i, ret)
			}
		}
	}
	tree := func(h *r2Harness, src Source[int]) gox.Elem {
		return src.Route(
			RouteMatch(r2Pred(h.rec, "even", r2fEven), r2Comp[int](h.rec, "EVEN")),
			RouteMatch(r2Pred(h.rec, "small", r2fSmall), r2Comp[int](h.rec, "SMALL")),
			r2Comp[int](h.rec, "D"),
		)
	}
	for _, w := range []int{8, 1} {
		t.Run(fmt.Sprintf("skip/workers=%d", w), func(t *testing.T) {
			h := newR2HarnessN(t, w)
			src, bsrc := NewSource(1), NewSource(0)
			h.render(r2Slot("a", tree(h, src)), r2Slot("b", bsrc.Route(RouteMatch(r2fPos, r2Bind[int](h.rec, "B")))))
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
	}
	t.Run("goroutine", func(t *testing.T) {
		h := newR2Harness(t)
		src := NewSource(1)
		h.render(r2Slot("a", tree(h, src)))
		r2Set(h, src, 12)
		h.expect("a", "EVEN")
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
	t.Run("noskip", func(t *testing.T) {
		h := newR2Harness(t)
		src := NewSourceNoSkip(1)
		h.render(r2Slot("a", tree(h, src)))
		chs := []<-chan error{}
		h.settle(func(ctx context.Context) {
			for _, v := range []int{2, 12, 3, 13, 11, 4} {
				chs = append(chs, src.Update(ctx, v))
			}
		})
		for i, ch := range chs {
			if ret := drainOp(t, ch); !r2SameErrs(ret, []error{nil}) {
				t.Fatalf("update %d returned %v", i, ret)
			}
		}
		h.expect("a", "EVEN")
		if got := h.rec.vals("pred:"); !slices.Equal(got, r2fStickyPreds) {
			t.Fatalf("predicate log\n got %v\nwant %v", got, r2fStickyPreds)
		}
		h.counts(r2fN{"render:EVEN": 2, "render:SMALL": 2, "render:D": 1, "clean:SMALL": 2, "clean:D": 1})
		h.oneDoor("a")
	})
}

// RouteSource gets a writable Source; writes from the branch propagate and
// can switch branches.
func TestRoute2FlatSourceLeafWrites(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(1)
	box := &r2Box[Source[int]]{}
	h.render(r2Slot("a", src.Route(
		RouteMatch(r2fSmall, r2Source(h.rec, "SMALL", box)),
		RouteMatch(func(v int) bool { return v >= 10 }, r2Comp[int](h.rec, "BIG")),
	)))
	h.expect("a", "SMALL")
	s := box.last()
	if s != src {
		t.Fatal("the leaf should receive the entry source")
	}
	h.quiet("a", func() {
		if ret := r2Set(h, s, 5); !r2SameErrs(ret, []error{nil}) {
			t.Fatalf("write returned %v", ret)
		}
	})
	if v := src.Get(); v != 5 {
		t.Fatalf("entry source holds %d, want 5", v)
	}
	r2Set(h, s, 15)
	h.expect("a", "BIG")
	if ret := r2Set(h, s, 15); len(ret) != 0 {
		t.Fatalf("equal write should be suppressed, got %v", ret)
	}
	r2Mutate(h, s, func(v int) int { return v - 10 })
	h.expect("a", "SMALL")
	h.counts(r2fN{"render:SMALL": 2, "clean:SMALL": 1, "render:BIG": 1, "clean:BIG": 1})
	if box.len() != 2 || box.at(1) != s {
		t.Fatal("reactivation should pass the same source")
	}
	h.oneDoor("a")
}

// [F23] old: RouteBeam accepted wherever RouteSource is. Beam-flavored routes
// mix with source routes in Route.
func TestRoute2FlatOldMixedFlavors(t *testing.T) {
	var (
		_ SourceRoute[int]      = RouteBeamValue[int](0)
		_ SourceRoute[int]      = RouteBeamMatch[int](r2fPos)
		_ SourceRoute[int]      = RouteBind(func(int) gox.Elem { return nil })
		_ SourceRoute[int]      = RouteComp[int](nil)
		_ SourceRoute[int]      = RouteBeamDerive(func(b Beam[int]) Beam[int] { return b })
		_ SourceRoute[Location] = RouteBeamModel[r2PathA]()
	)
	h := newR2Harness(t)
	src := NewSource(0)
	loc := r2LocSource("/c")
	sbox, bbox := &r2Box[Source[int]]{}, &r2Box[Beam[int]]{}
	h.render(
		r2Slot("a", src.Route(
			RouteMatch(func(v int) bool { return v < 0 }, r2Source(h.rec, "NEG", sbox)),
			RouteBeamMatch(func(v int) bool { return v > 100 }, r2Beam(h.rec, "HUGE", bbox)),
			RouteBeamValue(50, r2Comp[int](h.rec, "FIFTY")),
			RouteBeamDerive(func(b Beam[int]) Beam[int] { return DeriveBeam(b, func(v int) int { return v / 10 }) },
				RouteBeamValue(3, r2Comp[int](h.rec, "THIRTIES"))),
			r2Bind[int](h.rec, "B"),
		)),
		r2Slot("loc", loc.Route(
			RouteBeamModel[r2PathA](r2Comp[r2PathA](h.rec, "A")),
			RouteModel[r2PathB](r2Comp[r2PathB](h.rec, "PB")),
			r2Bind[Location](h.rec, "L"),
		)),
	)
	r2Run(h, "a", src, "B=0",
		r2fI{set: -1, want: "NEG"},
		r2fI{set: 101, want: "HUGE"},
		r2fI{set: 50, want: "FIFTY"},
		r2fI{set: 35, want: "THIRTIES"},
		r2fI{set: 39, want: "THIRTIES", quiet: true},
		r2fI{set: 7, want: "B=7"},
		r2fI{set: -5, want: "NEG", counts: r2fN{"render:NEG": 2, "render:HUGE": 1, "render:FIFTY": 1, "render:THIRTIES": 1, "render:B": 2}},
	)
	if sbox.last() != src || bbox.last().Get() != -5 {
		t.Fatal("leaves should receive the entry source")
	}
	r2Run(h, "loc", loc, "L=/c",
		r2fL{set: r2Loc("/a"), want: "A"},
		r2fL{set: r2Loc("/b"), want: "PB"},
		r2fL{set: r2Loc("/c?x=1"), want: "L=/c?x=1"},
		r2fL{set: r2Loc("/a"), want: "A", counts: r2fN{"render:A": 2, "render:PB": 1}},
	)
}

// old F01-F08, F14-F16: TestRouteSourceAndBeamState (BeamRouteStateFragment)
// on Route: defaults, derive branches, list branches, writes through the
// branch Source and re-activation.
func TestRoute2FlatOldRouteSourceAndBeamState(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	src := NewSource(r2fState{Int: 1})
	lensBox := &r2Box[Source[string]]{}
	listBox := &r2Box[Source[[]string]]{}
	lensRoute := func(l Source[string]) gox.Elem {
		h.rec.hit("render:lens")
		lensBox.put(l)
		return l.Bind(func(v string) gox.Elem { return textElem("lens:" + v) })
	}
	lensListRoute := func(l Source[[]string]) gox.Elem {
		h.rec.hit("render:lens-list")
		listBox.put(l)
		return l.Bind(func(v []string) gox.Elem { return textElem(strings.Join(v, ",")) })
	}
	beamRoute := func(b Beam[string]) gox.Elem {
		h.rec.hit("render:beam")
		return b.Bind(func(v string) gox.Elem { return textElem("beam:" + v) })
	}
	beamListRoute := func(b Beam[[]string]) gox.Elem {
		h.rec.hit("render:beam-list")
		return b.Bind(func(v []string) gox.Elem { return textElem(strings.Join(v, ",")) })
	}
	stateDefault := func(label string) func(Beam[r2fState]) gox.Elem {
		return func(b Beam[r2fState]) gox.Elem {
			h.rec.hit("render:" + label)
			return b.Bind(func(v r2fState) gox.Elem { return textElem(fmt.Sprintf("%s:%d:%s", label, v.Int, v.Str)) })
		}
	}
	h.render(
		r2Slot("lens", src.Route(
			RouteMatch(r2fStrSet, RouteDerive(r2Derive(h.rec, "lens", r2fStrSource), RouteSource(lensRoute))),
			RouteBeam(stateDefault("default")),
		)),
		r2Slot("beam", src.RouteBeam(
			RouteBeamMatch(r2fStrSet, RouteBeamDerive(r2DeriveBeam(h.rec, "beam", r2fStrBeam), RouteBeam(beamRoute))),
			RouteBeam(stateDefault("beam-default")),
		)),
		r2Slot("lens-list", src.Route(
			RouteMatch(r2fStrSet, RouteDerive(r2Derive(h.rec, "lens-list", r2fListSource), RouteSource(lensListRoute))),
		)),
		r2Slot("beam-list", src.RouteBeam(
			RouteBeamMatch(r2fStrSet, RouteBeamDerive(r2DeriveBeam(h.rec, "beam-list", r2fListBeam), RouteBeam(beamListRoute))),
		)),
	)
	slots := []string{"lens", "beam", "lens-list", "beam-list"}
	routeDoors := func(at string) {
		t.Helper()
		for _, s := range slots {
			r2fRouteDoor(h, s)
		}
	}
	derives := r2fN{"derive:lens": 1, "derive:beam": 1, "derive:lens-list": 1, "derive:beam-list": 1}
	with := func(a, b r2fN) r2fN {
		out := r2fN{}
		for k, v := range a {
			out[k] = v
		}
		for k, v := range b {
			out[k] = v
		}
		return out
	}

	r2fExpect(h, "initial: ", map[string]string{"lens": "default:1:", "beam": "beam-default:1:", "lens-list": "", "beam-list": ""})
	h.countsAt("initial: ", with(derives, r2fN{"render:default": 1, "render:beam-default": 1, "render:lens": 0, "render:beam": 0, "render:lens-list": 0, "render:beam-list": 0}))
	for _, s := range []string{"lens-list", "beam-list"} {
		if n := h.doorCount(s); n != 1 {
			t.Fatalf("%s: an unmatched container should be one blank door, got %d", s, n)
		}
	}
	routeDoors("initial: ")

	r2Set(h, src, r2fState{Int: 2, Str: "doc"})
	r2fExpect(h, "doc: ", map[string]string{"lens": "lens:doc", "beam": "beam:doc", "lens-list": "doc,2", "beam-list": "doc,2"})
	h.countsAt("doc: ", with(derives, r2fN{"render:default": 1, "render:beam-default": 1, "render:lens": 1, "render:beam": 1, "render:lens-list": 1, "render:beam-list": 1}))

	r2fQuiet(h, slots, func() { r2Set(h, src, r2fState{Int: 2, Str: "next"}) })
	r2fExpect(h, "next: ", map[string]string{"lens": "lens:next", "beam": "beam:next", "lens-list": "next,2", "beam-list": "next,2"})
	h.countsAt("next: ", with(derives, r2fN{"render:lens": 1, "render:beam": 1, "render:lens-list": 1, "render:beam-list": 1}))

	r2fQuiet(h, slots, func() {
		if ret := r2Set(h, lensBox.last(), "child"); !r2SameErrs(ret, []error{nil}) {
			t.Fatalf("branch write returned %v", ret)
		}
	})
	if got := src.Get(); got != (r2fState{Int: 2, Str: "child"}) {
		t.Fatalf("entry source holds %+v", got)
	}
	r2fExpect(h, "child: ", map[string]string{"lens": "lens:child", "beam": "beam:child", "lens-list": "child,2", "beam-list": "child,2"})
	h.countsAt("child: ", with(derives, r2fN{"render:lens": 1, "render:beam": 1, "render:lens-list": 1, "render:beam-list": 1}))

	r2Set(h, lensBox.last(), "")
	if got := src.Get(); got != (r2fState{Int: 2}) {
		t.Fatalf("entry source holds %+v", got)
	}
	r2fExpect(h, "clear: ", map[string]string{"lens": "default:2:", "beam": "beam-default:2:", "lens-list": "", "beam-list": ""})
	h.countsAt("clear: ", with(derives, r2fN{"render:default": 2, "render:beam-default": 2, "render:lens": 1}))

	r2fQuiet(h, slots, func() { r2Set(h, src, r2fState{Int: 4}) })
	r2fExpect(h, "default-int: ", map[string]string{"lens": "default:4:", "beam": "beam-default:4:", "lens-list": "", "beam-list": ""})
	h.countsAt("default-int: ", with(derives, r2fN{"render:default": 2, "render:beam-default": 2}))

	r2Set(h, src, r2fState{Int: 5, Str: "again"})
	r2fExpect(h, "again: ", map[string]string{"lens": "lens:again", "beam": "beam:again", "lens-list": "again,5", "beam-list": "again,5"})
	h.countsAt("again: ", with(derives, r2fN{"render:default": 2, "render:lens": 2, "render:beam": 2, "render:lens-list": 2, "render:beam-list": 2}))
	if lensBox.len() != 2 || lensBox.at(0) != lensBox.at(1) || listBox.len() != 2 || listBox.at(0) != listBox.at(1) {
		t.Fatal("the derived branch Source should live for the whole route instance")
	}
	if got := listBox.last().Get(); !slices.Equal(got, []string{"again", "5"}) {
		t.Fatalf("list source holds %v", got)
	}
	routeDoors("again: ")
}

// old F09: TestRouteNoDefaultHandlesRapidUnmatch. A match then an unmatch in
// one burst leaves the container blank.
func TestRoute2FlatOldRapidUnmatch(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	src := NewSource(r2fState{})
	lens := func(l Source[string]) gox.Elem {
		return l.Bind(func(v string) gox.Elem { return textElem("lens:" + v) })
	}
	beam := func(b Beam[string]) gox.Elem {
		return b.Bind(func(v string) gox.Elem { return textElem("beam:" + v) })
	}
	h.render(
		r2Slot("lens", src.Route(RouteMatch(r2fStrSet, RouteDerive(r2fStrSource, RouteSource(lens))))),
		r2Slot("beam", src.RouteBeam(RouteBeamMatch(r2fStrSet, RouteBeamDerive(r2fStrBeam, RouteBeam(beam))))),
	)
	r2fExpect(h, "initial: ", map[string]string{"lens": "", "beam": ""})
	var a, b <-chan error
	h.settle(func(ctx context.Context) {
		a = src.Update(ctx, r2fState{Int: 1, Str: "queued"})
		b = src.Update(ctx, r2fState{Int: 2})
	})
	if ret := drainOp(t, a); !r2SameErrs(ret, []error{nil}) && !r2SameErrs(ret, []error{context.Canceled}) {
		t.Fatalf("first update returned %v", ret)
	}
	if ret := drainOp(t, b); !r2SameErrs(ret, []error{nil}) {
		t.Fatalf("second update returned %v", ret)
	}
	r2fExpect(h, "none: ", map[string]string{"lens": "", "beam": ""})
	r2fSingle(h, "lens")
	r2fSingle(h, "beam")
	r2Set(h, src, r2fState{Int: 3, Str: "after"})
	r2fExpect(h, "hit: ", map[string]string{"lens": "lens:after", "beam": "beam:after"})
	r2fRouteDoor(h, "lens")
	r2fRouteDoor(h, "beam")
}

// old F10: TestDeferredRouteInner/Outer RouteFragment. Fixed comps render on
// every activation and never on in-branch updates.
func TestRoute2FlatOldDeferredRouteCompActivations(t *testing.T) {
	h := newR2Harness(t)
	src := r2LocSource("/")
	to := func(name string) func(Location) bool {
		return func(l Location) bool { return len(l.Segments) > 0 && l.Segments[0] == name }
	}
	h.render(r2Slot("a", src.Route(
		RouteMatch(to("inner"), RouteComp[Location](r2Leaf{h.rec, "inner"})),
		RouteMatch(to("outer"), RouteComp[Location](r2Leaf{h.rec, "outer"})),
		RouteComp[Location](r2Leaf{h.rec, "home"}),
	)))
	r2Run(h, "a", src, "home",
		r2fL{set: r2Loc("/inner"), want: "inner", counts: r2fN{"render:inner": 1, "clean:home": 1}},
		r2fL{set: r2Loc("/inner?x=1"), want: "inner", quiet: true},
		r2fL{set: r2Loc("/inner/deep"), want: "inner", quiet: true},
		r2fL{set: r2Loc("/"), want: "home", counts: r2fN{"render:home": 2, "clean:inner": 1}},
		r2fL{set: r2Loc("/inner"), want: "inner", counts: r2fN{"render:inner": 2, "clean:home": 2}},
		r2fL{set: r2Loc("/outer"), want: "outer", counts: r2fN{"render:outer": 1, "clean:inner": 2}},
		r2fL{set: r2Loc("/home?y=2"), want: "home", counts: r2fN{"render:home": 3}},
		r2fL{set: r2Loc("/other"), want: "home", quiet: true, counts: r2fN{"render:home": 3, "render:inner": 2, "render:outer": 1}},
	)
}

// old F11: TestRouteBindMethods. Bind-flavored derive, match and default
// branches render the current value.
func TestRoute2FlatOldRouteBindMethods(t *testing.T) {
	h := newR2Harness(t)
	src := r2LocSource("/custom/bind-test?tab=active")
	isCrossA := func(l Location) bool { return len(l.Segments) == 1 && l.Segments[0] == "cross-a" }
	h.render(r2Slot("a", src.Route(
		RouteMatch(r2fIsCustom, RouteDerive(func(s Source[Location]) Source[r2fCustomRoute] {
			return DeriveSourceEqual(s, r2fGetCustom, r2fSetCustom, func(a, b r2fCustomRoute) bool { return a == b })
		}, RouteBind(func(r r2fCustomRoute) gox.Elem {
			h.rec.hit("render:derive")
			return textElem("derive-bind:" + r.ID + ":" + r.Tab)
		}))),
		RouteMatch(isCrossA, RouteBind(func(l Location) gox.Elem {
			h.rec.hit("render:match")
			return textElem("match-bind:" + l.Path())
		})),
		RouteBind(func(l Location) gox.Elem {
			h.rec.hit("render:default")
			return textElem("default-bind:" + l.Path())
		}),
	)))
	r2Run(h, "a", src, "derive-bind:bind-test:active",
		r2fL{set: r2Loc("/cross-a"), want: "match-bind:/cross-a"},
		r2fL{set: r2Loc("/custom/bind-test?tab=active"), want: "derive-bind:bind-test:active"},
		r2fL{set: r2Loc("/custom/bind-test?tab=active&x=1"), want: "derive-bind:bind-test:active", quiet: true},
		r2fL{set: r2Loc("/custom/other?tab=x"), want: "derive-bind:other:x"},
		r2fL{set: r2Loc("/unknown"), want: "default-bind:/unknown"},
		r2fL{set: r2Loc("/unknown2"), want: "default-bind:/unknown2"},
		r2fL{set: r2Loc("/cross-a"), want: "match-bind:/cross-a", counts: r2fN{"render:derive": 3, "render:match": 2, "render:default": 2}},
	)
}

// old F12: Bind re-renders on change and not when the comparator says equal.
func TestRoute2FlatOldBindComparator(t *testing.T) {
	h := newR2Harness(t)
	st := NewSource(r2fState{Int: 1, Str: "a"})
	ci := NewSourceEqual("abc", strings.EqualFold)
	raw := NewSource("abc")
	h.render(
		r2Slot("derive", st.Route(RouteMatch(r2fStrSet, RouteDerive(r2fStrSource, r2Bind[string](h.rec, "D"))))),
		r2Slot("ci", ci.Route(RouteMatch(func(v string) bool { return v != "" }, r2Bind[string](h.rec, "C")))),
		r2Slot("fold", raw.Route(RouteDerive(func(s Source[string]) Source[string] {
			return DeriveSourceEqual(s, func(v string) string { return v }, func(_, v string) string { return v }, strings.EqualFold)
		}, r2Bind[string](h.rec, "R")))),
	)
	r2Run(h, "derive", st, "D=a",
		r2Step[r2fState]{set: r2fState{2, "a"}, want: "D=a", quiet: true, counts: r2fN{"render:D": 1}},
		r2Step[r2fState]{set: r2fState{2, "b"}, want: "D=b", counts: r2fN{"render:D": 2}},
		r2Step[r2fState]{set: r2fState{3, "b"}, want: "D=b", quiet: true},
		r2Step[r2fState]{set: r2fState{3, ""}, want: ""},
		r2Step[r2fState]{set: r2fState{4, "b"}, want: "D=b", counts: r2fN{"render:D": 3}},
	)
	r2Run(h, "ci", ci, "C=abc",
		r2fS{set: "ABC", want: "C=abc", quiet: true, ret: []error{}},
		r2fS{set: "abd", want: "C=abd", ret: []error{nil}},
		r2fS{set: "ABD", want: "C=abd", quiet: true, ret: []error{}, counts: r2fN{"render:C": 2}},
	)
	r2Run(h, "fold", raw, "R=abc",
		r2fS{set: "ABC", want: "R=abc", quiet: true, ret: []error{nil}},
		r2fS{set: "Abd", want: "R=Abd"},
		r2fS{set: "abd", want: "R=Abd", quiet: true, counts: r2fN{"render:R": 2}},
	)
}

// old F17: TestRouteSourceAndDerivedBeamEntrypoints. Routes on a derived entry
// get the entry source itself; writes reach the root.
func TestRoute2FlatOldEntrypoints(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	root := NewSource(r2fState{Int: 1, Str: "a"})
	str := DeriveSource(root, r2fStr, r2fSetStr)
	parity := DeriveBeam(root, func(s r2fState) string {
		if s.Int%2 == 0 {
			return "even"
		}
		return "odd"
	})
	lensBox, simpleBox := &r2Box[Source[string]]{}, &r2Box[Source[string]]{}
	bindLeaf := func(label string, box *r2Box[Source[string]]) func(Source[string]) gox.Elem {
		return func(l Source[string]) gox.Elem {
			h.rec.hit("render:" + label)
			box.put(l)
			return l.Bind(func(v string) gox.Elem { return textElem(label + ":" + v) })
		}
	}
	beamLeaf := func(label string) func(Beam[string]) gox.Elem {
		return func(b Beam[string]) gox.Elem {
			h.rec.hit("render:" + label)
			return b.Bind(func(v string) gox.Elem { return textElem(label + ":" + v) })
		}
	}
	nonEmpty := func(v string) bool { return v != "" }
	simple := func(v string) bool { return strings.HasPrefix(v, "simple") }
	h.render(
		r2Slot("lens", str.Route(RouteMatch(nonEmpty, RouteSource(bindLeaf("lens", lensBox))))),
		r2Slot("lens-beam", str.RouteBeam(RouteBeamMatch(nonEmpty, RouteBeam(beamLeaf("beam"))))),
		r2Slot("derived", parity.RouteBeam(
			RouteBeamMatch(func(v string) bool { return v == "odd" }, RouteBeam(beamLeaf("derived"))),
			RouteBeam(beamLeaf("derived-default")),
		)),
		r2Slot("simple", str.Route(RouteMatch(simple, RouteSource(bindLeaf("simple-lens", simpleBox))))),
		r2Slot("simple-beam", str.RouteBeam(RouteBeamMatch(simple, RouteBeam(beamLeaf("simple-beam"))))),
	)
	all := []string{"lens", "lens-beam", "derived", "simple", "simple-beam"}
	r2fExpect(h, "initial: ", map[string]string{"lens": "lens:a", "lens-beam": "beam:a", "derived": "derived:odd", "simple": "", "simple-beam": ""})
	if lensBox.last() != str {
		t.Fatal("a match leaf on a derived entry should receive the entry source")
	}

	r2fQuiet(h, all, func() { r2Set(h, lensBox.last(), "b") })
	if got := root.Get(); got != (r2fState{1, "b"}) {
		t.Fatalf("root holds %+v", got)
	}
	r2fExpect(h, "lens-update: ", map[string]string{"lens": "lens:b", "lens-beam": "beam:b", "derived": "derived:odd", "simple": "", "simple-beam": ""})

	r2Set(h, str, "simple")
	r2fExpect(h, "set-simple: ", map[string]string{"lens": "lens:simple", "lens-beam": "beam:simple", "simple": "simple-lens:simple", "simple-beam": "simple-beam:simple", "derived": "derived:odd"})

	r2fQuiet(h, all, func() { r2Set(h, simpleBox.last(), "simple-lens") })
	if got := root.Get(); got != (r2fState{1, "simple-lens"}) || str.Get() != "simple-lens" {
		t.Fatalf("root holds %+v, entry %q", got, str.Get())
	}
	r2fExpect(h, "simple-update: ", map[string]string{"lens": "lens:simple-lens", "lens-beam": "beam:simple-lens", "simple": "simple-lens:simple-lens", "simple-beam": "simple-beam:simple-lens"})

	r2fQuiet(h, []string{"lens", "lens-beam", "simple", "simple-beam"}, func() { r2Set(h, root, r2fState{2, "simple-lens"}) })
	r2fExpect(h, "even: ", map[string]string{"derived": "derived-default:even", "simple": "simple-lens:simple-lens", "lens": "lens:simple-lens"})
	h.counts(r2fN{"render:lens": 1, "render:beam": 1, "render:simple-lens": 1, "render:simple-beam": 1, "render:derived": 1, "render:derived-default": 1})
	if simpleBox.last() != str {
		t.Fatal("simple leaf should receive the entry source")
	}
	for _, s := range all {
		r2fRouteDoor(h, s)
	}
}

// old F18: docs Section example. Value routes swap branches and an unknown
// value falls to the default, even when it renders the same comp.
func TestRoute2FlatOldSectionValue(t *testing.T) {
	for _, beam := range []bool{true, false} {
		t.Run(fmt.Sprintf("beam=%v", beam), func(t *testing.T) {
			h := newR2Harness(t)
			sec := NewSource(r2fSection("profile"))
			profile := RouteComp[r2fSection](r2Leaf{h.rec, "PROFILE"})
			billing := RouteComp[r2fSection](r2Leaf{h.rec, "BILLING"})
			if beam {
				h.render(r2Slot("a", sec.RouteBeam(
					RouteBeamValue[r2fSection]("profile", profile),
					RouteBeamValue[r2fSection]("billing", billing),
					profile,
				)))
			} else {
				h.render(r2Slot("a", sec.Route(
					RouteValue[r2fSection]("profile", profile),
					RouteValue[r2fSection]("billing", billing),
					profile,
				)))
			}
			r2Run(h, "a", sec, "PROFILE",
				r2Step[r2fSection]{set: "billing", want: "BILLING"},
				r2Step[r2fSection]{set: "settings", want: "PROFILE", counts: r2fN{"render:PROFILE": 2}},
				r2Step[r2fSection]{set: "other", want: "PROFILE", quiet: true},
				r2Step[r2fSection]{set: "profile", want: "PROFILE", counts: r2fN{"render:PROFILE": 3, "clean:PROFILE": 2}},
				r2Step[r2fSection]{set: "billing", want: "BILLING", counts: r2fN{"render:BILLING": 2}},
			)
		})
	}
}

// old F20, F21: the active branch is re-checked first; when it ends, the whole
// list is re-tried in order, the ended one included (old skipped it). On a
// flat list its predicate just failed, so only the log differs.
func TestRoute2FlatOldReseekOrder(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(2)
	bit := func(label string, m int) func(int) bool {
		return r2Pred(h.rec, label, func(v int) bool { return v&m != 0 })
	}
	h.render(r2Slot("a", src.Route(
		RouteMatch(bit("A", 1), r2Comp[int](h.rec, "A")),
		RouteMatch(bit("B", 2), r2Comp[int](h.rec, "B")),
		RouteMatch(bit("C", 4), r2Comp[int](h.rec, "C")),
	)))
	h.expect("a", "B")
	seen := 0
	if got := r2fPredDelta(h, &seen); !slices.Equal(got, []string{"A=2", "B=2"}) {
		t.Fatalf("initial preds %v", got)
	}
	for i, st := range []struct {
		set   int
		want  string
		quiet bool
		preds []string
	}{
		{3, "B", true, []string{"B=3"}},
		{1, "A", false, []string{"B=1", "A=1"}},
		{5, "A", true, []string{"A=5"}},
		{4, "C", false, []string{"A=4", "A=4", "B=4", "C=4"}},
		{6, "C", true, []string{"C=6"}},
		{2, "B", false, []string{"C=2", "A=2", "B=2"}},
		{0, "", false, []string{"B=0", "A=0", "B=0", "C=0"}},
		{7, "A", false, []string{"A=7"}},
	} {
		at := fmt.Sprintf("step %d (set %d): ", i, st.set)
		do := func() { r2Set(h, src, st.set) }
		if st.quiet {
			h.quietAt(at, "a", do)
		} else {
			do()
		}
		h.expectAt(at, "a", st.want)
		if got := r2fPredDelta(h, &seen); !slices.Equal(got, st.preds) {
			t.Fatalf("%spreds %v, want %v", at, got, st.preds)
		}
		h.oneDoorAt(at, "a")
	}
}

// old F24: RouteDefault gets a writable Source. A write from the default that
// makes an earlier route match hands over to it.
func TestRoute2FlatOldWritableDefault(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(0)
	box := &r2Box[Source[int]]{}
	h.render(r2Slot("a", src.Route(
		RouteMatch(func(v int) bool { return v == 7 }, r2Comp[int](h.rec, "A")),
		r2Source(h.rec, "D", box),
	)))
	h.expect("a", "D")
	h.quiet("a", func() { r2Set(h, box.last(), 3) })
	if src.Get() != 3 {
		t.Fatalf("entry source holds %d", src.Get())
	}
	r2Set(h, box.last(), 7)
	h.expect("a", "A")
	r2Set(h, src, 0)
	h.expect("a", "D")
	h.counts(r2fN{"render:D": 2, "clean:D": 1, "render:A": 1, "clean:A": 1})
	if box.len() != 2 || box.at(0) != box.at(1) || box.last() != src {
		t.Fatal("the default should get the entry source on every activation")
	}
	h.oneDoor("a")
}

// old F25: the old router ignored derived writes while unmatched. Route has
// no built-in guard: a retained branch Source calls the user's set.
func TestRoute2FlatOldDeriveWriteAfterDeactivate(t *testing.T) {
	for _, guarded := range []bool{false, true} {
		t.Run(fmt.Sprintf("guarded=%v", guarded), func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(r2fState{Int: 1, Str: "a"})
			box := &r2Box[Source[string]]{}
			set := func(s r2fState, v string) r2fState {
				h.rec.hit("set:" + v)
				if guarded && s.Str == "" {
					return s
				}
				s.Str = v
				return s
			}
			h.render(r2Slot("a", src.Route(
				RouteMatch(r2fStrSet, RouteDerive(func(s Source[r2fState]) Source[string] {
					return DeriveSourceEqual(s, r2fStr, set, func(a, b string) bool { return a == b })
				}, r2Source(h.rec, "L", box))),
				r2Bind[r2fState](h.rec, "D"),
			)))
			h.expect("a", "L")
			l := box.last()
			r2Set(h, src, r2fState{Int: 1})
			h.expect("a", "D={1 }")
			ret := r2Set(h, l, "late")
			h.counts(r2fN{"set:late": 1})
			if guarded {
				if len(ret) != 0 || src.Get() != (r2fState{Int: 1}) {
					t.Fatalf("guarded set: ret %v, root %+v", ret, src.Get())
				}
				h.expect("a", "D={1 }")
				return
			}
			if !r2SameErrs(ret, []error{nil}) || src.Get() != (r2fState{1, "late"}) {
				t.Fatalf("raw set: ret %v, root %+v", ret, src.Get())
			}
			h.expect("a", "L")
			h.counts(r2fN{"render:L": 2})
			if box.last() != l {
				t.Fatal("the derived Source should outlive the deactivation")
			}
		})
	}
}

// Clean resets bind's stored value: a deactivated RouteBind does not keep the
// old value alive.
func TestRoute2FlatBindCleanReleasesValue(t *testing.T) {
	h := newR2Harness(t)
	big := newR2Big(1)
	wp := weak.Make(big)
	src := NewSource(big)
	big = nil
	h.render(r2Slot("a", src.Route(
		RouteMatch(func(b *r2Big) bool { return b.kind == 1 }, r2Bind[*r2Big](h.rec, "B")),
		r2Comp[*r2Big](h.rec, "D"),
	)))
	h.expect("a", "B=big1")
	r2Set(h, src, newR2Big(2))
	h.expect("a", "D")
	r2Set(h, src, newR2Big(3))
	h.expect("a", "D")
	if !r2Gone(wp) {
		t.Fatal("the value of a deactivated bind branch is still reachable")
	}
	h.oneDoor("a")
	runtime.KeepAlive(src)
	runtime.KeepAlive(h)
}

type r2fP[T any] struct {
	set   T
	want  string
	acts  int
	preds []string
}

type r2fPI = r2fP[int]

func r2fPredRun[T any](h *r2Harness, slot string, src Source[T], initial string, preds []string, steps ...r2fP[T]) {
	h.t.Helper()
	seen := 0
	h.expectAt("initial: ", slot, initial)
	if got := r2fPredDelta(h, &seen); !slices.Equal(got, preds) {
		h.t.Fatalf("initial: preds %v, want %v", got, preds)
	}
	h.oneDoorAt("initial: ", slot)
	id := h.door(slot)
	for i, st := range steps {
		at := fmt.Sprintf("step %d (set %v): ", i, st.set)
		n := len(h.doorActions(id))
		r2Set(h, src, st.set)
		if got := len(h.doorActions(id)) - n; got != st.acts {
			h.t.Fatalf("%swant %d new door actions, got %d", at, st.acts, got)
		}
		h.expectAt(at, slot, st.want)
		if got := r2fPredDelta(h, &seen); !slices.Equal(got, st.preds) {
			h.t.Fatalf("%spreds %v, want %v", at, got, st.preds)
		}
		h.oneDoorAt(at, slot)
	}
}

type r2fItem struct {
	ID  int
	Tab string
}

func (i r2fItem) String() string { return strconv.Itoa(i.ID) + "/" + i.Tab }

func r2fItemID(a, b r2fItem) bool { return a.ID == b.ID }

func r2fTens(v int) int { return v / 10 }

func r2fTensSet(a, b int) int { return b*10 + a%10 }

func r2fOdd(v int) bool { return v%2 != 0 }

// After an end the whole list is re-selected in order, the ended route
// included, in the same update. An intact path stays sticky even when an
// earlier sibling matches; once anything below breaks, the ancestor that still
// matches loses its stickiness and an earlier sibling can win.
func TestRoute2FlatEndReselectsAll(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(2)
		h.render(r2Slot("a", e.mount(src,
			k.match(r2Pred(h.rec, "big", func(v int) bool { return v >= 100 }), r2Comp[int](h.rec, "BIG")),
			k.match(r2Pred(h.rec, "pos", r2fPos),
				k.match(r2Pred(h.rec, "even", r2fEven), r2Comp[int](h.rec, "EVEN")),
				k.match(r2Pred(h.rec, "small", r2fSmall), r2Comp[int](h.rec, "SMALL")),
			),
			r2Bind[int](h.rec, "D"),
		)))
		r2fPredRun(h, "a", src, "EVEN", []string{"big=2", "pos=2", "even=2"},
			r2fPI{set: 4, want: "EVEN", preds: []string{"pos=4", "even=4"}},
			r2fPI{set: 104, want: "EVEN", preds: []string{"pos=104", "even=104"}},
			r2fPI{set: 105, want: "BIG", acts: 1, preds: []string{"pos=105", "even=105", "big=105"}},
			r2fPI{set: 3, want: "SMALL", acts: 1, preds: []string{"big=3", "big=3", "pos=3", "even=3", "small=3"}},
			r2fPI{set: 6, want: "SMALL", preds: []string{"pos=6", "small=6"}},
			r2fPI{set: 12, want: "EVEN", acts: 1, preds: []string{"pos=12", "small=12", "big=12", "pos=12", "even=12"}},
			r2fPI{set: -1, want: "D=-1", acts: 1, preds: []string{"pos=-1", "big=-1", "pos=-1"}},
			r2fPI{set: -2, want: "D=-2", acts: 1, preds: []string{"big=-2", "pos=-2"}},
			r2fPI{set: 50, want: "EVEN", acts: 1, preds: []string{"big=50", "pos=50", "even=50"}},
		)
		h.counts(r2fN{
			"render:BIG": 1, "clean:BIG": 1, "render:EVEN": 3, "clean:EVEN": 2,
			"render:SMALL": 1, "clean:SMALL": 1, "render:D": 2, "clean:D": 2,
		})
	})
}

// A match whose nested routes match nothing falls through to the default.
// The default re-seeks its earlier siblings in order on every update and the
// first that matches takes over: an earlier match wins even when the skipped
// match's nested routes would now match too; once they match on their own,
// that match takes over from the default.
func TestRoute2FlatFallThroughDefaultYield(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(1)
		h.render(r2Slot("a", e.mount(src,
			k.match(r2Pred(h.rec, "e", func(v int) bool { return v >= 100 }), r2Comp[int](h.rec, "E")),
			k.match(r2Pred(h.rec, "m", r2fPos),
				k.match(r2Pred(h.rec, "five", func(v int) bool { return v%5 == 0 }), r2Comp[int](h.rec, "FIVE")),
			),
			r2Comp[int](h.rec, "D"),
		)))
		r2fPredRun(h, "a", src, "D", []string{"e=1", "m=1", "five=1"},
			r2fPI{set: 2, want: "D", preds: []string{"e=2", "m=2", "five=2"}},
			r2fPI{set: 105, want: "E", acts: 1, preds: []string{"e=105"}},
			r2fPI{set: 3, want: "D", acts: 1, preds: []string{"e=3", "e=3", "m=3", "five=3"}},
			r2fPI{set: 10, want: "FIVE", acts: 1, preds: []string{"e=10", "m=10", "five=10"}},
			r2fPI{set: 20, want: "FIVE", preds: []string{"m=20", "five=20"}},
			r2fPI{set: 21, want: "D", acts: 1, preds: []string{"m=21", "five=21", "e=21", "m=21", "five=21"}},
			r2fPI{set: 0, want: "D", preds: []string{"e=0", "m=0"}},
			r2fPI{set: 150, want: "E", acts: 1, preds: []string{"e=150"}},
		)
		h.counts(r2fN{"render:E": 2, "clean:E": 1, "render:FIVE": 1, "clean:FIVE": 1, "render:D": 3, "clean:D": 3})
	})
}

// A derive level compares the new derived value with the last one it walked,
// using the derived comparator. Equal means the subtree is untouched: no
// predicate below runs and the door stays quiet. A nil comparator re-walks
// every update; a coarse comparator ignores changes it calls equal, even when
// a predicate below would now switch.
func TestRoute2FlatDeriveSkip(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
			h := newR2Harness(t)
			src := NewSource(1)
			h.render(r2Slot("a", e.mount(src,
				k.match(r2Pred(h.rec, "pos", r2fPos), k.derive(h.rec, "tens", r2fTens, r2fTensSet,
					k.match(r2Pred(h.rec, "odd", r2fOdd), r2Comp[int](h.rec, "ODD")),
					r2Bind[int](h.rec, "T"),
				)),
				r2Comp[int](h.rec, "D"),
			)))
			r2fPredRun(h, "a", src, "T=0", []string{"pos=1", "odd=0"},
				r2fPI{set: 5, want: "T=0", preds: []string{"pos=5"}},
				r2fPI{set: 9, want: "T=0", preds: []string{"pos=9"}},
				r2fPI{set: 12, want: "ODD", acts: 1, preds: []string{"pos=12", "odd=1"}},
				r2fPI{set: 19, want: "ODD", preds: []string{"pos=19"}},
				r2fPI{set: 25, want: "T=2", acts: 1, preds: []string{"pos=25", "odd=2", "pos=25", "odd=2"}},
				r2fPI{set: 28, want: "T=2", preds: []string{"pos=28"}},
				r2fPI{set: -3, want: "D", acts: 1, preds: []string{"pos=-3", "pos=-3"}},
				r2fPI{set: 7, want: "T=0", acts: 1, preds: []string{"pos=7", "odd=0"}},
			)
			h.counts(r2fN{"derive:tens": 1, "render:T": 3, "render:ODD": 1, "render:D": 1})
		})
	})
	for _, beam := range []bool{false, true} {
		t.Run(fmt.Sprintf("nil/beam=%v", beam), func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(1)
			if beam {
				h.render(r2Slot("a", src.RouteBeam(RouteBeamDerive(r2DeriveBeam(h.rec, "tens", func(b Beam[int]) Beam[int] {
					return DeriveBeamEqual(b, r2fTens, nil)
				}), RouteBeamMatch(r2Pred(h.rec, "odd", r2fOdd), r2Comp[int](h.rec, "ODD")), r2Comp[int](h.rec, "Z")))))
			} else {
				h.render(r2Slot("a", src.Route(RouteDerive(r2Derive(h.rec, "tens", func(s Source[int]) Source[int] {
					return DeriveSourceEqual(s, r2fTens, r2fTensSet, nil)
				}), RouteMatch(r2Pred(h.rec, "odd", r2fOdd), r2Comp[int](h.rec, "ODD")), r2Comp[int](h.rec, "Z")))))
			}
			r2fPredRun(h, "a", src, "Z", []string{"odd=0"},
				r2fPI{set: 5, want: "Z", preds: []string{"odd=0"}},
				r2fPI{set: 9, want: "Z", preds: []string{"odd=0"}},
				r2fPI{set: 12, want: "ODD", acts: 1, preds: []string{"odd=1"}},
				r2fPI{set: 19, want: "ODD", preds: []string{"odd=1"}},
				r2fPI{set: 25, want: "Z", acts: 1, preds: []string{"odd=2", "odd=2"}},
				r2fPI{set: 28, want: "Z", preds: []string{"odd=2"}},
			)
			h.counts(r2fN{"derive:tens": 1, "render:Z": 2, "render:ODD": 1})
		})
	}
	for _, beam := range []bool{false, true} {
		t.Run(fmt.Sprintf("coarse/beam=%v", beam), func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(r2fItem{ID: 1})
			tab := r2Pred(h.rec, "tab", func(i r2fItem) bool { return i.Tab == "x" })
			ident := func(i r2fItem) r2fItem { return i }
			if beam {
				h.render(r2Slot("a", src.RouteBeam(RouteBeamDerive(r2DeriveBeam(h.rec, "id", func(b Beam[r2fItem]) Beam[r2fItem] {
					return DeriveBeamEqual(b, ident, r2fItemID)
				}), RouteBeamMatch(tab, r2Comp[r2fItem](h.rec, "TABX")), r2Bind[r2fItem](h.rec, "I")))))
			} else {
				h.render(r2Slot("a", src.Route(RouteDerive(r2Derive(h.rec, "id", func(s Source[r2fItem]) Source[r2fItem] {
					return DeriveSourceEqual(s, ident, func(_, i r2fItem) r2fItem { return i }, r2fItemID)
				}), RouteMatch(tab, r2Comp[r2fItem](h.rec, "TABX")), r2Bind[r2fItem](h.rec, "I")))))
			}
			r2fPredRun(h, "a", src, "I=1/", []string{"tab=1/"},
				r2fP[r2fItem]{set: r2fItem{1, "x"}, want: "I=1/"},
				r2fP[r2fItem]{set: r2fItem{1, "y"}, want: "I=1/"},
				r2fP[r2fItem]{set: r2fItem{2, "x"}, want: "TABX", acts: 1, preds: []string{"tab=2/x"}},
				r2fP[r2fItem]{set: r2fItem{2, ""}, want: "TABX"},
				r2fP[r2fItem]{set: r2fItem{3, ""}, want: "I=3/", acts: 1, preds: []string{"tab=3/", "tab=3/"}},
				r2fP[r2fItem]{set: r2fItem{3, "x"}, want: "I=3/"},
				r2fP[r2fItem]{set: r2fItem{4, "x"}, want: "TABX", acts: 1, preds: []string{"tab=4/x"}},
			)
			h.counts(r2fN{"derive:id": 1, "render:I": 2, "render:TABX": 2})
		})
	}
}

// A derive caches a no-match for its derived value: re-sought by a later
// default, or by the container re-selection after ODD ends, it does not walk
// its subtree again until the value changes. The re-selection cleans only the
// active branch, so the cache of an inactive derive survives it.
func TestRoute2FlatDeriveNoMatchCache(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(0)
		h.render(r2Slot("a", e.mount(src,
			k.derive(h.rec, "tens", r2fTens, r2fTensSet,
				k.match(r2Pred(h.rec, "t", func(v int) bool { return v == 5 }), r2Comp[int](h.rec, "FIFTY")),
			),
			k.match(r2Pred(h.rec, "odd", r2fOdd), r2Comp[int](h.rec, "ODD")),
			r2Bind[int](h.rec, "D"),
		)))
		r2fPredRun(h, "a", src, "D=0", []string{"t=0", "odd=0"},
			r2fPI{set: 2, want: "D=2", acts: 1, preds: []string{"odd=2"}},
			r2fPI{set: 4, want: "D=4", acts: 1, preds: []string{"odd=4"}},
			r2fPI{set: 12, want: "D=12", acts: 1, preds: []string{"t=1", "odd=12"}},
			r2fPI{set: 14, want: "D=14", acts: 1, preds: []string{"odd=14"}},
			r2fPI{set: 15, want: "ODD", acts: 1, preds: []string{"odd=15"}},
			r2fPI{set: 17, want: "ODD", preds: []string{"odd=17"}},
			r2fPI{set: 16, want: "D=16", acts: 1, preds: []string{"odd=16", "odd=16"}},
			r2fPI{set: 18, want: "D=18", acts: 1, preds: []string{"odd=18"}},
			r2fPI{set: 52, want: "FIFTY", acts: 1, preds: []string{"t=5"}},
			r2fPI{set: 58, want: "FIFTY"},
			r2fPI{set: 63, want: "ODD", acts: 1, preds: []string{"t=6", "t=6", "odd=63"}},
		)
		h.counts(r2fN{"derive:tens": 1, "render:FIFTY": 1, "clean:FIFTY": 1, "render:ODD": 2, "clean:ODD": 1})
	})
}

type r2nKit struct {
	r2Kit
	leaf func(r *r2Rec, label string) SourceRoute[int]
}

func (k r2nKit) d(r *r2Rec, label string, get func(int) int, rs ...SourceRoute[int]) SourceRoute[int] {
	return k.derive(r, label, get, nil, rs...)
}

func r2nCompLeaf(r *r2Rec, label string) SourceRoute[int] {
	return r2Comp[int](r, label)
}

var r2nKits = []r2nKit{
	{r2Kits[0], r2nCompLeaf},
	{r2Kits[1], r2nCompLeaf},
	{r2Kit{name: "mixed", match: r2Kits[0].match, value: r2Kits[1].value, derive: r2Kits[0].derive},
		func(r *r2Rec, label string) SourceRoute[int] { return r2Source[int](r, label, nil) }},
}

func r2nMatrix(t *testing.T, f func(t *testing.T, k r2nKit, e r2Entry)) {
	t.Helper()
	for _, k := range r2nKits {
		for _, e := range r2Entries {
			if e.beam && !k.beam {
				continue
			}
			t.Run(k.name+"/"+e.name, func(t *testing.T) { f(t, k, e) })
		}
	}
}

func r2nGt(x int) func(int) bool { return func(v int) bool { return v > x } }

func r2nLt(x int) func(int) bool { return func(v int) bool { return v < x } }

func r2nGe(x int) func(int) bool { return func(v int) bool { return v >= x } }

func r2nMod(m, r int) func(int) bool { return func(v int) bool { return v%m == r } }

func r2nIn(lo, hi int) func(int) bool { return func(v int) bool { return v >= lo && v < hi } }

func r2nAlways[T any](T) bool { return true }

func r2nNever[T any](T) bool { return false }

func r2nIDSource(s Source[int]) Source[int] { return DeriveSource(s, r2ID, r2IDSet) }

func r2nIDBeam(b Beam[int]) Beam[int] { return DeriveBeam(b, r2ID) }

func r2nTens(s Source[int]) Source[int] {
	return DeriveSource(s, func(v int) int { return v / 10 }, func(a, b int) int { return b*10 + a%10 })
}

func r2nBlankReplaces(h *r2Harness, slot string) int {
	h.t.Helper()
	id := h.door(slot)
	marker := fmt.Sprintf(` id="d0r%d"`, id)
	n := 0
	for _, a := range h.doorActions(id) {
		if r2Text(a.body) != "" {
			continue
		}
		if a.name != "door_replace" || !strings.HasPrefix(a.body, "<d0-r ") || !strings.HasSuffix(a.body, "></d0-r>") ||
			strings.Count(a.body, "<") != 2 || !strings.Contains(a.body, marker) {
			h.t.Fatalf("slot %q: expected a blank container replace, got %s %q", slot, a.name, a.body)
		}
		n++
	}
	return n
}

func r2nCompact(s []string) []string {
	return slices.Compact(slices.Clone(s))
}

func r2nPaths(h *r2Harness) []string {
	var out []string
	for _, a := range h.client.actions() {
		if a.name == "set_path" {
			out = append(out, a.body)
		}
	}
	return out
}

func r2nRouterHarness(t *testing.T, uri string) *r2Harness {
	t.Helper()
	h := newR2Harness(t)
	h.inst.location = beam.NewSource(r2Loc(uri), path.EqualLocation, false)
	return h
}

func r2nUpdateReplace[T any](h *r2Harness, s Source[T], v T) []error {
	h.t.Helper()
	var ch <-chan error
	h.settle(func(ctx context.Context) { ch = s.Update(HistoryReplaceContext(ctx), v) })
	return drainOp(h.t, ch)
}

type r2nCase struct {
	name    string
	start   int
	tree    func(r *r2Rec, k r2nKit) []SourceRoute[int]
	initial string
	steps   []r2Step[int]
	after   func(h *r2Harness)
}

func r2nRunCase(t *testing.T, c r2nCase) {
	t.Helper()
	r2nMatrix(t, func(t *testing.T, k r2nKit, e r2Entry) {
		for _, w := range []int{8, 1} {
			h := newR2HarnessN(t, w)
			src := NewSource(c.start)
			h.render(r2Slot("n", e.mount(src, c.tree(h.rec, k)...)))
			r2Run(h, "n", src, c.initial, c.steps...)
			if c.after != nil {
				c.after(h)
			}
		}
	})
}

func r2nWantBlanks(n int) func(h *r2Harness) {
	return func(h *r2Harness) {
		h.t.Helper()
		if got := r2nBlankReplaces(h, "n"); got != n {
			h.t.Fatalf("want %d blank container replaces, got %d", n, got)
		}
	}
}

func r2nAll(fs ...func(h *r2Harness)) func(h *r2Harness) {
	return func(h *r2Harness) {
		h.t.Helper()
		for _, f := range fs {
			f(h)
		}
	}
}

func r2nWantActions(n int) func(h *r2Harness) {
	return func(h *r2Harness) {
		h.t.Helper()
		if got := len(h.doorActions(h.door("n"))); got != n {
			h.t.Fatalf("want %d door replaces, got %d", n, got)
		}
	}
}

var r2nN01 = r2nCase{
	name:  "N01",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nGt(0), k.match(r2nGt(10), k.leaf(r, "A"))),
			k.leaf(r, "D"),
		}
	},
	initial: "D",
	steps: []r2Step[int]{
		{set: 5, want: "D", quiet: true},
		{set: 15, want: "A"},
		{set: 7, want: "D"},
		{set: 0, want: "D", quiet: true},
		{set: 3, want: "D", quiet: true},
		{set: 20, want: "A", counts: map[string]int{"render:A": 2, "clean:A": 1, "render:D": 2, "clean:D": 2}},
	},
	after: r2nAll(r2nWantActions(3), r2nWantBlanks(0)),
}

var r2nN02 = r2nCase{
	name:  "N02",
	start: 1,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nGt(0), k.match(r2Pred(r, "in", r2nGt(10)), k.leaf(r, "A"))),
		}
	},
	initial: "",
	steps: []r2Step[int]{
		{set: 2, want: "", quiet: true},
		{set: 3, want: "", quiet: true},
		{set: 15, want: "A"},
		{set: 16, want: "A", quiet: true, counts: map[string]int{"render:A": 1}},
		{set: 17, want: "A", quiet: true, counts: map[string]int{"render:A": 1}},
		{set: 4, want: "", counts: map[string]int{"render:A": 1, "clean:A": 1}},
		{set: 5, want: "", quiet: true},
	},
	after: func(h *r2Harness) {
		h.t.Helper()
		want := []string{"1", "2", "3", "15", "16", "17", "4", "4", "5"}
		if got := h.rec.vals("pred:in="); !slices.Equal(got, want) {
			h.t.Fatalf("pred:in log %v, want %v", got, want)
		}
		r2nWantBlanks(1)(h)
	},
}

var r2nN03 = r2nCase{
	name:  "N03",
	start: 5,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nLt(100), k.match(r2nGt(10), k.leaf(r, "A"))),
			k.match(r2nGt(0), k.leaf(r, "B")),
		}
	},
	initial: "B",
	steps: []r2Step[int]{
		{set: 150, want: "B", quiet: true},
		{set: 50, want: "B", quiet: true},
		{set: 0, want: ""},
		{set: 50, want: "A"},
		{set: 5, want: "B", counts: map[string]int{"render:A": 1, "clean:A": 1, "render:B": 2, "clean:B": 1}},
	},
	after: r2nWantBlanks(1),
}

var r2nN04 = r2nCase{
	name:  "N04",
	start: 1,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nGt(0), k.value(1, r2Comp[int](r, "X")), k.value(2, r2Comp[int](r, "Y"))),
			k.leaf(r, "D"),
		}
	},
	initial: "X",
	steps: []r2Step[int]{
		{set: 2, want: "Y"},
		{set: 3, want: "D"},
		{set: 0, want: "D", quiet: true},
		{set: 1, want: "X", counts: map[string]int{"render:X": 2, "clean:X": 1, "clean:Y": 1, "render:D": 1, "clean:D": 1}},
	},
	after: r2nWantActions(3),
}

var r2nN05 = r2nCase{
	name:  "N05",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.value(5, r2Comp[int](r, "P")),
			k.d(r, "d", nil, r2Bind[int](r, "Y")),
		}
	},
	initial: "Y=0",
	steps: []r2Step[int]{
		{set: 2, want: "Y=2"},
		{set: 5, want: "P"},
		{set: 6, want: "Y=6", counts: map[string]int{"derive:d": 1, "render:P": 1, "render:Y": 3}},
	},
}

var r2nN06 = r2nCase{
	name:  "N06",
	start: 20,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.value(5, r2Comp[int](r, "P")),
			k.d(r, "d", nil, k.match(r2nIn(1, 10), r2Comp[int](r, "X")), r2Bind[int](r, "Y")),
		}
	},
	initial: "Y=20",
	steps: []r2Step[int]{
		{set: 30, want: "Y=30"},
		{set: 5, want: "P", counts: map[string]int{"render:Y": 2, "clean:Y": 2}},
		{set: 3, want: "X"},
		{set: 5, want: "X", quiet: true, counts: map[string]int{"render:P": 1}},
		{set: 40, want: "Y=40"},
		{set: 5, want: "P", counts: map[string]int{"derive:d": 1, "render:P": 2, "render:X": 1, "clean:X": 1, "render:Y": 3}},
	},
}

var r2nN07 = r2nCase{
	name:  "N07",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.d(r, "d", func(v int) int { return v / 10 }, k.value(1, r2Comp[int](r, "X"))),
			k.leaf(r, "D"),
		}
	},
	initial: "D",
	steps: []r2Step[int]{
		{set: 12, want: "X"},
		{set: 25, want: "D"},
		{set: 13, want: "X", counts: map[string]int{"derive:d": 1, "render:X": 2, "render:D": 2}},
	},
}

var r2nN08 = r2nCase{
	name:  "N08",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.value(5, r2Comp[int](r, "P")),
			k.match(r2nGe(0), r2Bind[int](r, "Y")),
		}
	},
	initial: "Y=0",
	steps: []r2Step[int]{
		{set: 2, want: "Y=2"},
		{set: 5, want: "Y=5"},
		{set: 5, want: "Y=5", quiet: true, ret: []error{}, counts: map[string]int{"render:P": 0}},
	},
}

var r2nN09 = r2nCase{
	name:  "N09",
	start: 2,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.value(7, r2Comp[int](r, "P")),
			k.match(r2nGt(0), k.value(1, r2Comp[int](r, "X")), r2Bind[int](r, "Y")),
		}
	},
	initial: "Y=2",
	steps: []r2Step[int]{
		{set: 1, want: "X"},
		{set: 3, want: "Y=3"},
		{set: 7, want: "Y=7", counts: map[string]int{"render:P": 0, "render:X": 1, "render:Y": 3}},
	},
}

// A non-default branch deep in the tree is sticky against earlier siblings at
// every level above it while its path is intact (50, 60). A break re-selects
// from scratch: DEEP is tried again, its inner match falls through and the
// derive's default DD is selected (150, 5). The derive is then a default for
// its parent, so MIDV at the parent's level takes over (60), and inside the
// derive DEEP takes over from DD (50, 20), while the matching TOPV never does:
// the outer match above is sticky.
var r2nSticky = r2nCase{
	name:  "sticky-deep",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.value(50, r2Comp[int](r, "TOPV")),
			k.match(r2nGt(0),
				k.value(60, r2Comp[int](r, "MIDV")),
				k.d(r, "d", nil,
					k.match(r2nGt(10), k.match(r2nLt(100), k.leaf(r, "DEEP"))),
					k.leaf(r, "DD"),
				),
			),
			k.leaf(r, "D"),
		}
	},
	initial: "D",
	steps: []r2Step[int]{
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
		{set: 60, want: "MIDV", counts: map[string]int{
			"render:TOPV": 1, "render:MIDV": 2, "render:DEEP": 3, "render:DD": 3, "render:D": 2, "derive:d": 1,
		}},
	},
	after: r2nAll(r2nWantActions(10), r2nWantBlanks(0)),
}

// Defaults at depth chain up through derives, so a top-level earlier sibling
// takes over.
var r2nDepthT1 = r2nCase{
	name:  "default-depth-t1",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.value(5, r2Comp[int](r, "P1")),
			k.d(r, "x", nil, k.d(r, "y", nil, r2Bind[int](r, "DEF"))),
		}
	},
	initial: "DEF=0",
	steps: []r2Step[int]{
		{set: 5, want: "P1"},
		{set: 6, want: "DEF=6"},
		{set: 7, want: "DEF=7"},
		{set: 5, want: "P1", counts: map[string]int{"render:DEF": 3, "render:P1": 2, "derive:x": 1, "derive:y": 1}},
	},
}

// The same default chain under a match: only the match level re-checks.
var r2nDepthT2 = r2nCase{
	name:  "default-depth-t2",
	start: 1,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.value(5, r2Comp[int](r, "P1")),
			k.match(r2nGt(0),
				k.value(6, r2Comp[int](r, "P2")),
				k.d(r, "x", nil, k.d(r, "y", nil, r2Bind[int](r, "DEF"))),
			),
		}
	},
	initial: "DEF=1",
	steps: []r2Step[int]{
		{set: 5, want: "DEF=5"},
		{set: 6, want: "P2"},
		{set: 7, want: "DEF=7"},
		{set: 0, want: ""},
		{set: 5, want: "P1", counts: map[string]int{"render:P1": 1, "render:P2": 1, "render:DEF": 3}},
	},
}

// A default chain at index 1 inside a derive makes the derive a default too.
var r2nDepthT3 = r2nCase{
	name:  "default-depth-t3",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.value(5, r2Comp[int](r, "P1")),
			k.d(r, "x", nil, k.value(6, r2Comp[int](r, "P2")), k.d(r, "y", nil, r2Bind[int](r, "DEF"))),
		}
	},
	initial: "DEF=0",
	steps: []r2Step[int]{
		{set: 5, want: "P1"},
		{set: 6, want: "P2"},
		{set: 7, want: "DEF=7"},
		{set: 5, want: "P1", counts: map[string]int{"render:P1": 2, "render:P2": 1, "render:DEF": 2, "derive:x": 1, "derive:y": 1}},
	},
}

// A default at index 1 of an inner derive at index 0 makes both derives
// defaults; a non-default inner branch makes both sticky again.
var r2nDepthT4 = r2nCase{
	name:  "default-depth-t4",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.value(5, r2Comp[int](r, "P1")),
			k.d(r, "x", nil, k.d(r, "y", nil, k.match(r2nIn(5, 7), r2Comp[int](r, "P2")), r2Bind[int](r, "DEF"))),
		}
	},
	initial: "DEF=0",
	steps: []r2Step[int]{
		{set: 5, want: "P1"},
		{set: 6, want: "P2"},
		{set: 5, want: "P2", quiet: true, counts: map[string]int{"render:P1": 1}},
		{set: 7, want: "DEF=7"},
		{set: 5, want: "P1", counts: map[string]int{"render:P1": 2, "render:P2": 1, "render:DEF": 2, "derive:x": 1, "derive:y": 1}},
	},
}

// Old: TestRouteBindMethods to-default -> to-match, projected to depth 2.
var r2nPort01Depth2 = r2nCase{
	name:  "port01-depth2",
	start: 1,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nGt(0), k.match(r2nMod(2, 0), k.leaf(r, "A")), k.leaf(r, "D")),
			k.leaf(r, "TOP"),
		}
	},
	initial: "D",
	steps: []r2Step[int]{
		{set: 3, want: "D", quiet: true},
		{set: 4, want: "A", counts: map[string]int{"render:A": 1, "clean:D": 1, "render:TOP": 0}},
		{set: 5, want: "D", counts: map[string]int{"render:D": 2}},
		{set: 0, want: "TOP"},
		{set: 2, want: "A"},
	},
	after: r2nWantActions(4),
}

// Old: TestRouteSourceAndBeamState route-source-lens, projected to depth 3.
// The inner range match ending breaks the path, and the outer match whose
// derive now matches nothing falls through to TOP (1002).
var r2nPort01Depth3 = r2nCase{
	name:  "port01-depth3",
	start: 1,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nGt(0), k.d(r, "x", nil,
				k.match(r2nLt(1000), k.match(r2nMod(2, 0), k.leaf(r, "A")), k.leaf(r, "D")),
			)),
			k.leaf(r, "TOP"),
		}
	},
	initial: "D",
	steps: []r2Step[int]{
		{set: 3, want: "D", quiet: true},
		{set: 4, want: "A", counts: map[string]int{"render:A": 1, "clean:D": 1, "render:TOP": 0}},
		{set: 5, want: "D", counts: map[string]int{"render:D": 2}},
		{set: 0, want: "TOP"},
		{set: 2, want: "A"},
		{set: 1002, want: "TOP", counts: map[string]int{"render:TOP": 2, "derive:x": 1}},
	},
	after: r2nWantActions(5),
}

// Old: the beam_route.go sticky rule (F20), projected into a nested group.
var r2nPort02 = r2nCase{
	name:  "port02",
	start: 2,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nGt(0), k.match(r2nMod(3, 0), k.leaf(r, "A")), k.match(r2nMod(2, 0), k.leaf(r, "B"))),
			k.leaf(r, "D"),
		}
	},
	initial: "B",
	steps: []r2Step[int]{
		{set: 6, want: "B", quiet: true},
		{set: 9, want: "A"},
		{set: 12, want: "A", quiet: true},
		{set: 5, want: "D"},
		{set: 4, want: "B"},
		{set: 0, want: "D", counts: map[string]int{"render:A": 1, "render:B": 2, "clean:B": 2, "render:D": 2, "clean:D": 1}},
	},
}

// Old: TestRouteSourceAndBeamState list routes / no default, projected: with
// no default at the top, an outer match whose nested match fails matches
// nothing and the container is blank; blank to blank is quiet.
var r2nPort04 = r2nCase{
	name:  "port04",
	start: 1,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nGt(0), k.match(r2nMod(5, 0), k.leaf(r, "A"))),
			k.match(r2nLt(0), k.leaf(r, "N")),
		}
	},
	initial: "",
	steps: []r2Step[int]{
		{set: 2, want: "", quiet: true},
		{set: 5, want: "A"},
		{set: 6, want: ""},
		{set: 7, want: "", quiet: true},
		{set: -1, want: "N"},
		{set: 10, want: "A"},
		{set: -2, want: "N"},
		{set: 3, want: "", counts: map[string]int{"render:A": 2, "clean:A": 2, "render:N": 2, "clean:N": 2}},
	},
	after: r2nAll(r2nWantActions(6), r2nWantBlanks(2)),
}

// Old: TestRouteSourceAndBeamState route-lens-clear and TestRouterSourceCombined
// transitions, projected: an outer switch cleans the whole subtree, and
// re-entering seeks from index 0 and renders the bind again for an equal value.
var r2nPort05 = r2nCase{
	name:  "port05",
	start: 5,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nIn(1, 100), k.match(r2nLt(50), r2Bind[int](r, "a")), k.match(r2nGe(40), k.leaf(r, "b"))),
			k.leaf(r, "D"),
		}
	},
	initial: "a=5",
	steps: []r2Step[int]{
		{set: 45, want: "a=45"},
		{set: 60, want: "b"},
		{set: 45, want: "b", quiet: true},
		{set: 200, want: "D", counts: map[string]int{"clean:b": 1, "clean:a": 2}},
		{set: 45, want: "a=45", counts: map[string]int{"render:a": 3}},
		{set: 5, want: "a=5"},
		{set: 200, want: "D"},
		{set: 70, want: "b", counts: map[string]int{
			"render:a": 4, "clean:a": 4, "render:b": 2, "clean:b": 1, "render:D": 2, "clean:D": 2,
		}},
	},
}

func r2nBit(b int) func(int) bool { return func(v int) bool { return v&b != 0 } }

// A deep end breaks the whole path and re-selects from scratch: the earlier A
// wins over the inner Y (11) and over the still matching q-match (3). With
// only q holding, the q-match matches nothing, so the selection falls through
// it to a blank container (2) and the next update selects afresh (3).
// Bits: p=1, q=2, r=4, s=8.
var r2nDeepEnd = r2nCase{
	name:  "deep-end",
	start: 6,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nBit(1), k.leaf(r, "A")),
			k.match(r2nBit(2), k.match(r2nBit(4), k.leaf(r, "X")), k.match(r2nBit(8), k.leaf(r, "Y"))),
		}
	},
	initial: "X",
	steps: []r2Step[int]{
		{set: 11, want: "A"},
		{set: 6, want: "X"},
		{set: 3, want: "A"},
		{set: 6, want: "X"},
		{set: 10, want: "Y"},
		{set: 6, want: "X"},
		{set: 2, want: ""},
		{set: 3, want: "A"},
		{set: 7, want: "A", quiet: true},
		{set: 14, want: "X", counts: map[string]int{"render:A": 3, "render:X": 5, "clean:X": 4, "render:Y": 1, "clean:Y": 1}},
	},
	after: r2nAll(r2nWantActions(9), r2nWantBlanks(1)),
}

// A default yields locally: under a sticky ancestor, the matching TOPA at the
// ancestor's level never takes over (3, 7), while INB at the default's own
// level does (7). The switched-in INB is sticky against INC (15, 14) until its
// path breaks; the re-selection then starts from scratch (10, 3).
// Bits: a=1, m=2, b=4, c=8.
var r2nLocalYield = r2nCase{
	name:  "local-yield",
	start: 2,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nBit(1), k.leaf(r, "TOPA")),
			k.match(r2nBit(2),
				k.match(r2nBit(8), k.leaf(r, "INC")),
				k.match(r2nBit(4), k.leaf(r, "INB")),
				r2Bind[int](r, "D"),
			),
		}
	},
	initial: "D=2",
	steps: []r2Step[int]{
		{set: 3, want: "D=3"},
		{set: 7, want: "INB"},
		{set: 15, want: "INB", quiet: true},
		{set: 14, want: "INB", quiet: true},
		{set: 10, want: "INC"},
		{set: 3, want: "TOPA", counts: map[string]int{"render:TOPA": 1, "render:D": 2, "render:INB": 1, "render:INC": 1}},
	},
}

// Under a default derive the re-checks run top-down and the first level that
// switches wins: INB switches in locally (4) and makes the derive sticky (5
// quiet), but from D=0 the top TOPA takes over before INB (5).
// Bits: a=1, b=4.
var r2nTopDownYield = r2nCase{
	name:  "top-down-yield",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nBit(1), k.leaf(r, "TOPA")),
			k.d(r, "d", nil, k.match(r2nBit(4), k.leaf(r, "INB")), r2Bind[int](r, "D")),
		}
	},
	initial: "D=0",
	steps: []r2Step[int]{
		{set: 4, want: "INB"},
		{set: 5, want: "INB", quiet: true},
		{set: 1, want: "TOPA"},
		{set: 0, want: "D=0"},
		{set: 5, want: "TOPA", counts: map[string]int{"render:TOPA": 2, "render:INB": 1, "render:D": 2, "derive:d": 1}},
	},
}

// No late re-check: keeping a path never switches into a default, so a default
// is never left shown while an earlier sibling matches. When O ends, the path
// breaks and the re-selection from scratch goes straight to the earlier X; the
// derive's default D is never rendered on the way.
var r2nNoLate = r2nCase{
	name:  "no-late-recheck",
	start: 1,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nMod(2, 0), k.leaf(r, "X")),
			k.d(r, "d", nil, k.match(r2nMod(2, 1), k.leaf(r, "O")), r2Bind[int](r, "D")),
		}
	},
	initial: "O",
	steps: []r2Step[int]{
		{set: 2, want: "X", counts: map[string]int{"render:D": 0}},
		{set: 3, want: "O"},
		{set: 5, want: "O", quiet: true},
		{set: 4, want: "X", counts: map[string]int{"render:X": 2, "render:O": 2, "render:D": 0, "derive:d": 1}},
	},
}

// A local switch below two default derives clears the default at every level
// above it: after INB switches in, neither MIDC at the outer derive's level
// nor TOPA at the top re-checks (13, 12); after MIDC switches in at the outer
// derive's level, TOPA does not re-check (9). Bits: a=1, b=4, c=8.
var r2nTwoDerives = r2nCase{
	name:  "switch-two-derives",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nBit(1), k.leaf(r, "TOPA")),
			k.d(r, "d1", nil,
				k.match(r2nBit(8), k.leaf(r, "MIDC")),
				k.d(r, "d2", nil, k.match(r2nBit(4), k.leaf(r, "INB")), r2Bind[int](r, "D")),
			),
		}
	},
	initial: "D=0",
	steps: []r2Step[int]{
		{set: 4, want: "INB"},
		{set: 13, want: "INB", quiet: true},
		{set: 12, want: "INB", quiet: true},
		{set: 8, want: "MIDC"},
		{set: 0, want: "D=0"},
		{set: 8, want: "MIDC"},
		{set: 9, want: "MIDC", quiet: true},
		{set: 1, want: "TOPA"},
		{set: 0, want: "D=0", counts: map[string]int{
			"render:D": 3, "clean:D": 2, "render:INB": 1, "render:MIDC": 2, "render:TOPA": 1, "derive:d1": 1, "derive:d2": 1,
		}},
	},
}

// A match whose nested derive children match nothing does not match: the
// selection falls through it to the later D (35). While the derived value
// stays equal, the derive keeps its no-match and does not walk its children
// again (36); a changed derived value walks them (12). Predicate logs hold
// the derived (tens) values.
var r2nMatchDeriveFall = r2nCase{
	name:  "match-derive-fallthrough",
	start: 0,
	tree: func(r *r2Rec, k r2nKit) []SourceRoute[int] {
		return []SourceRoute[int]{
			k.match(r2nGt(0), k.d(r, "d", func(v int) int { return v / 10 },
				k.match(r2Pred(r, "t1", func(t int) bool { return t == 1 }), k.leaf(r, "X")),
				k.match(r2Pred(r, "t2", func(t int) bool { return t == 2 }), k.leaf(r, "Y")),
			)),
			r2Bind[int](r, "D"),
		}
	},
	initial: "D=0",
	steps: []r2Step[int]{
		{set: 15, want: "X"},
		{set: 25, want: "Y"},
		{set: 35, want: "D=35"},
		{set: 36, want: "D=36"},
		{set: 12, want: "X"},
		{set: 0, want: "D=0", counts: map[string]int{"render:X": 2, "render:Y": 1, "render:D": 4, "clean:D": 3, "derive:d": 1}},
	},
	after: func(h *r2Harness) {
		h.t.Helper()
		for _, c := range []struct {
			name string
			want []string
		}{
			{"t1", []string{"1", "2", "2", "3", "1"}},
			{"t2", []string{"2", "3", "3"}},
		} {
			if got := h.rec.vals("pred:" + c.name + "="); !slices.Equal(got, c.want) {
				h.t.Fatalf("pred:%s log %v, want %v", c.name, got, c.want)
			}
		}
	},
}

var r2nCases = []r2nCase{
	r2nN01, r2nN02, r2nN03, r2nN04, r2nN05, r2nN06, r2nN07, r2nN08, r2nN09,
	r2nSticky, r2nDepthT1, r2nDepthT2, r2nDepthT3, r2nDepthT4,
	r2nPort01Depth2, r2nPort01Depth3, r2nPort02, r2nPort04, r2nPort05,
	r2nDeepEnd, r2nLocalYield, r2nTopDownYield, r2nNoLate, r2nTwoDerives, r2nMatchDeriveFall,
}

// N01: a match whose predicate holds but whose children match nothing does not
// match: the selection falls through it to the later default, which keeps
// yielding to it only once its children match.
func TestRoute2NestedFallThrough(t *testing.T) { r2nRunCase(t, r2nN01) }

// N02: with nothing matching at the top the container is blank, and every
// update seeks from the top again, quietly while nothing matches. A break that
// leaves nothing matching blanks the container, re-trying the ended branch.
func TestRoute2NestedNoMatchReseek(t *testing.T) { r2nRunCase(t, r2nN02) }

// N03: the selection falls through a non-matching match to a later
// non-default sibling, which is then sticky against the earlier match.
func TestRoute2NestedFallThroughSticky(t *testing.T) { r2nRunCase(t, r2nN03) }

// N04: an inner end re-selects from scratch; the outer match still holds but
// its children match nothing, so the selection falls through to the top
// default.
func TestRoute2NestedInnerEndFallsThrough(t *testing.T) { r2nRunCase(t, r2nN04) }

// N05: a derive whose active default is at index 0 is a default itself.
func TestRoute2NestedGroupDefaultIndexZero(t *testing.T) { r2nRunCase(t, r2nN05) }

// N06: a derive whose active default is at index > 0 is a default too and
// yields to an earlier sibling; once it switches to a non-default child it is
// sticky again.
func TestRoute2NestedGroupDefaultIndexNonZero(t *testing.T) { r2nRunCase(t, r2nN06) }

// N07: a derive whose children match nothing does not match, so the selection
// falls through to the later default.
func TestRoute2NestedDeriveNoMatch(t *testing.T) { r2nRunCase(t, r2nN07) }

// N08: a match swallows the default-ness of its content.
func TestRoute2NestedMatchSwallowsDefault(t *testing.T) { r2nRunCase(t, r2nN08) }

// N09: inner default re-check works under a sticky outer match.
func TestRoute2NestedInnerDefaultRecheck(t *testing.T) { r2nRunCase(t, r2nN09) }

func TestRoute2NestedStickyDeepVsEarlierSiblings(t *testing.T) { r2nRunCase(t, r2nSticky) }

func TestRoute2NestedDefaultAtDepth(t *testing.T) {
	for _, c := range []r2nCase{r2nDepthT1, r2nDepthT2, r2nDepthT3, r2nDepthT4} {
		t.Run(c.name, func(t *testing.T) { r2nRunCase(t, c) })
	}
}

func TestRoute2NestedDeepEndReselects(t *testing.T) { r2nRunCase(t, r2nDeepEnd) }

func TestRoute2NestedDefaultYieldLocal(t *testing.T) {
	for _, c := range []r2nCase{r2nLocalYield, r2nTopDownYield} {
		t.Run(c.name, func(t *testing.T) { r2nRunCase(t, c) })
	}
}

func TestRoute2NestedSwitchClearsDefaultTwoDerives(t *testing.T) { r2nRunCase(t, r2nTwoDerives) }

func TestRoute2NestedMatchDeriveFallThrough(t *testing.T) { r2nRunCase(t, r2nMatchDeriveFall) }

type r2nDocs struct {
	Docs bool `path:"/docs"`
}

type r2nSettings struct {
	Settings bool `path:"/settings"`
}

// The model-route form goes from /docs straight to /settings; the section's
// not-found default is rendered only for a path no model decodes.
func TestRoute2NestedNoLateRecheck(t *testing.T) {
	t.Run("int", func(t *testing.T) { r2nRunCase(t, r2nNoLate) })
	t.Run("model", func(t *testing.T) {
		h := newR2Harness(t)
		loc := r2LocSource("/docs")
		section := func(s Source[Location]) Source[Location] {
			return DeriveSourceEqual(s, Location.Clone, func(_, l Location) Location { return l.Clone() }, path.EqualLocation)
		}
		h.render(r2Slot("m", loc.Route(
			RouteModel[r2nSettings](r2Comp[r2nSettings](h.rec, "settings")),
			RouteDerive(r2Derive(h.rec, "section", section),
				RouteModel[r2nDocs](r2Comp[r2nDocs](h.rec, "docs")),
				r2Comp[Location](h.rec, "NF"),
			),
		)))
		r2Run(h, "m", loc, "docs",
			r2Step[Location]{set: r2Loc("/settings"), want: "settings", counts: map[string]int{"render:NF": 0}},
			r2Step[Location]{set: r2Loc("/docs"), want: "docs"},
			r2Step[Location]{set: r2Loc("/zzz"), want: "NF"},
			r2Step[Location]{set: r2Loc("/settings"), want: "settings"},
			r2Step[Location]{set: r2Loc("/docs"), want: "docs", counts: map[string]int{
				"render:settings": 2, "render:docs": 3, "render:NF": 1, "derive:section": 1,
			}},
		)
	})
}

// A derive skip at depth: while the inner derive's value is equal by its own
// same-decade comparator, nothing below it runs, so the in predicate is not
// called, X does not take over although the value is odd (30, 130) and B does
// not re-render; the outer levels still process every update (out) and every
// tens change (mid).
func TestRoute2NestedDeriveSkipAtDepth(t *testing.T) {
	decade := func(a, b int) bool { return a/10 == b/10 }
	for _, beam := range []bool{false, true} {
		t.Run(map[bool]string{false: "source", true: "beam"}[beam], func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(20)
			out := r2Pred(h.rec, "out", r2nGe(0))
			mid := r2Pred(h.rec, "mid", r2nGe(0))
			in := r2Pred(h.rec, "in", r2nMod(2, 1))
			el := src.Route(RouteMatch(out, RouteDerive(r2Derive(h.rec, "tens", r2nTens),
				RouteMatch(mid, RouteDerive(r2Derive(h.rec, "dec", func(s Source[int]) Source[int] {
					return DeriveSourceEqual(s, r2ID, r2IDSet, decade)
				}), RouteMatch(in, r2Comp[int](h.rec, "X")), r2Bind[int](h.rec, "B"))),
			)))
			if beam {
				el = src.RouteBeam(RouteBeamMatch(out, RouteBeamDerive(r2DeriveBeam(h.rec, "tens", func(b Beam[int]) Beam[int] {
					return DeriveBeam(b, func(v int) int { return v / 10 })
				}), RouteBeamMatch(mid, RouteBeamDerive(r2DeriveBeam(h.rec, "dec", func(b Beam[int]) Beam[int] {
					return DeriveBeamEqual(b, r2ID, decade)
				}), RouteBeamMatch(in, r2Comp[int](h.rec, "X")), r2Bind[int](h.rec, "B"))),
				)))
			}
			h.render(r2Slot("k", el))
			r2Run(h, "k", src, "B=2",
				r2Step[int]{set: 25, want: "B=2", quiet: true},
				r2Step[int]{set: 30, want: "B=2", quiet: true},
				r2Step[int]{set: 120, want: "B=12"},
				r2Step[int]{set: 130, want: "B=12", quiet: true},
				r2Step[int]{set: 210, want: "X", counts: map[string]int{"render:B": 2, "render:X": 1, "derive:tens": 1, "derive:dec": 1}},
			)
			for _, c := range []struct {
				name string
				want []string
			}{
				{"out", []string{"20", "25", "30", "120", "130", "210"}},
				{"mid", []string{"2", "3", "12", "13", "21"}},
				{"in", []string{"2", "12", "21"}},
			} {
				if got := h.rec.vals("pred:" + c.name + "="); !slices.Equal(got, c.want) {
					t.Fatalf("pred:%s log %v, want %v", c.name, got, c.want)
				}
			}
		})
	}
}

// N11: every table above gives the same texts and door op counts on every
// kit (all source, all beam, hand-mixed) and every entry point.
func TestRoute2NestedKitParity(t *testing.T) {
	for _, c := range r2nCases {
		t.Run(c.name, func(t *testing.T) {
			var ref []string
			var refName string
			r2nMatrix(t, func(t *testing.T, k r2nKit, e r2Entry) {
				h := newR2Harness(t)
				src := NewSource(c.start)
				h.render(r2Slot("n", e.mount(src, c.tree(h.rec, k)...)))
				id := h.door("n")
				got := []string{h.text("n") + "#0"}
				for _, st := range c.steps {
					n := len(h.doorActions(id))
					r2Set(h, src, st.set)
					got = append(got, fmt.Sprintf("%s#%d", h.text("n"), len(h.doorActions(id))-n))
				}
				if ref == nil {
					ref, refName = got, k.name+"/"+e.name
					return
				}
				if !slices.Equal(got, ref) {
					t.Fatalf("%s/%s gives %q, %s gives %q", k.name, e.name, got, refName, ref)
				}
			})
		})
	}
}

// N10: a 5-level tree mixing source and beam containers stays on one door. The
// never matching branch at the bottom makes the value-7 branch above it match
// nothing, so on 107 the default HIGH yields to the later LOWODD instead.
func TestRoute2NestedDeepFiveLevel(t *testing.T) {
	for _, w := range []int{8, 1} {
		h := newR2HarnessN(t, w)
		src := NewSource(0)
		mod100 := func(s Source[int]) Source[int] {
			return DeriveSource(s, func(v int) int { return v % 100 }, func(a, b int) int { return a/100*100 + b })
		}
		h.render(r2Slot("deep", src.Route(
			RouteMatch(r2nGe(100),
				RouteDerive(r2Derive(h.rec, "mod", mod100),
					RouteBeamValue(7, RouteBeamMatch(r2nAlways[int],
						RouteBeamDerive(r2DeriveBeam(h.rec, "id", r2nIDBeam),
							RouteBeamMatch(r2nNever[int], r2Comp[int](h.rec, "NEVER"))))),
					RouteMatch(r2nLt(50),
						RouteMatch(r2nMod(2, 0), r2Bind[int](h.rec, "LOWEVEN")),
						r2Bind[int](h.rec, "LOWODD"),
					),
					r2Bind[int](h.rec, "HIGH"),
				),
			),
			r2Bind[int](h.rec, "ROOT"),
		)))
		r2Run(h, "deep", src, "ROOT=0",
			r2Step[int]{set: 100, want: "LOWEVEN=0"},
			r2Step[int]{set: 101, want: "LOWODD=1"},
			r2Step[int]{set: 150, want: "HIGH=50"},
			r2Step[int]{set: 107, want: "LOWODD=7"},
			r2Step[int]{set: 108, want: "LOWEVEN=8"},
			r2Step[int]{set: 2, want: "ROOT=2"},
			r2Step[int]{set: 103, want: "LOWODD=3"},
			r2Step[int]{set: 104, want: "LOWEVEN=4", counts: map[string]int{
				"render:NEVER": 0, "derive:mod": 1, "derive:id": 1, "render:ROOT": 2, "render:LOWEVEN": 3,
			}},
		)
	}
}

// N12: switches between leaves in different subtrees; subtree C never
// matches but its derives still run once.
func TestRoute2NestedSubtreeSwitch(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(0)
	h.render(r2Slot("s", src.Route(
		RouteMatch(r2nIn(100, 200),
			RouteDerive(r2Derive(h.rec, "dA", func(s Source[int]) Source[int] {
				return DeriveSource(s, func(v int) int { return v - 100 }, func(_, b int) int { return b + 100 })
			}),
				RouteMatch(r2nLt(50),
					RouteValue(10, r2Comp[int](h.rec, "AXY")),
					r2Comp[int](h.rec, "AXW"),
				),
				r2Comp[int](h.rec, "AV"),
			),
		),
		RouteMatch(r2nIn(200, 300),
			RouteValue(201, r2Comp[int](h.rec, "BZ")),
			RouteBeamDerive(r2DeriveBeam(h.rec, "dB", func(b Beam[int]) Beam[int] {
				return DeriveBeam(b, func(v int) int { return v % 10 })
			}), RouteBeamValue(5, r2Comp[int](h.rec, "BQ"))),
			r2Comp[int](h.rec, "BD"),
		),
		RouteMatch(r2nLt(0),
			RouteDerive(r2Derive(h.rec, "dC1", func(s Source[int]) Source[int] {
				return DeriveSource(s, func(v int) int { return -v }, func(_, b int) int { return -b })
			}),
				RouteDerive(r2Derive(h.rec, "dC2", func(s Source[int]) Source[string] {
					return DeriveSource(s, strconv.Itoa, func(a int, _ string) int { return a })
				}), RouteValue("1", r2Comp[string](h.rec, "CL"))),
				RouteBeamDerive(r2DeriveBeam(h.rec, "dC3", r2nIDBeam), r2Bind[int](h.rec, "CB")),
			),
		),
		r2Comp[int](h.rec, "TOP"),
	)))
	r2Run(h, "s", src, "TOP",
		r2Step[int]{set: 110, want: "AXY"},
		r2Step[int]{set: 201, want: "BZ"},
		r2Step[int]{set: 120, want: "AXW"},
		r2Step[int]{set: 170, want: "AV"},
		r2Step[int]{set: 201, want: "BZ", counts: map[string]int{"render:BZ": 2, "clean:BZ": 1, "clean:AV": 1}},
		r2Step[int]{set: 205, want: "BQ"},
		r2Step[int]{set: 207, want: "BD"},
		r2Step[int]{set: 110, want: "AXY"},
		r2Step[int]{set: 0, want: "TOP"},
	)
	h.counts(map[string]int{
		"render:TOP": 2, "clean:TOP": 1,
		"render:AXY": 2, "clean:AXY": 2,
		"render:BZ": 2, "clean:BZ": 2,
		"render:AXW": 1, "clean:AXW": 1,
		"render:AV": 1, "clean:AV": 1,
		"render:BQ": 1, "clean:BQ": 1,
		"render:BD": 1, "clean:BD": 1,
		"render:CL": 0, "render:CB": 0,
		"derive:dA": 1, "derive:dB": 1, "derive:dC1": 1, "derive:dC2": 1, "derive:dC3": 1,
	})
	if n := len(h.doorActions(h.door("s"))); n != 9 {
		t.Fatalf("want 9 door replaces, got %d", n)
	}
}

// N13: no-match chains at levels 2 to 5 (including contradictory nested
// values) fall through to the default D2 inside the outer match (slot e) or,
// with no defaults at all, to a blank container (slot b). Each fill or end of
// the deep branch either switches locally or breaks the path and re-selects
// from scratch. Only replacing shown content with nothing touches the door; a
// non-matching update while the container is already blank is quiet.
func TestRoute2NestedDeepFallThrough(t *testing.T) {
	chain := func(h *r2Harness, p string) SourceRoute[int] {
		return RouteMatch(r2nGt(10),
			RouteMatch(r2nGt(20),
				RouteValue(1, RouteValue(2, r2Comp[int](h.rec, p+"X"))),
				RouteBeamMatch(r2nGt(30),
					RouteBeamValue(31, RouteBeamValue(32, r2Comp[int](h.rec, p+"Y"))),
					RouteBeamMatch(r2nMod(7, 0), r2Comp[int](h.rec, p+"DEEP")),
				),
			),
		)
	}
	h := newR2Harness(t)
	src, bare := NewSource(0), NewSource(0)
	h.render(
		r2Slot("e", src.Route(
			RouteMatch(r2nGt(0), chain(h, "E"), r2Comp[int](h.rec, "D2")),
			r2Comp[int](h.rec, "D1"),
		)),
		r2Slot("b", bare.Route(chain(h, "B"))),
	)
	r2Run(h, "e", src, "D1",
		r2Step[int]{set: 5, want: "D2"},
		r2Step[int]{set: 15, want: "D2", quiet: true},
		r2Step[int]{set: 25, want: "D2", quiet: true},
		r2Step[int]{set: 27, want: "D2", quiet: true},
		r2Step[int]{set: 31, want: "D2", quiet: true},
		r2Step[int]{set: 32, want: "D2", quiet: true},
		r2Step[int]{set: 35, want: "EDEEP"},
		r2Step[int]{set: 36, want: "D2"},
		r2Step[int]{set: 38, want: "D2", quiet: true},
		r2Step[int]{set: 42, want: "EDEEP"},
		r2Step[int]{set: 22, want: "D2"},
		r2Step[int]{set: 12, want: "D2", quiet: true},
		r2Step[int]{set: 3, want: "D2", quiet: true},
		r2Step[int]{set: 0, want: "D1", counts: map[string]int{
			"render:D1": 2, "render:D2": 3, "render:EDEEP": 2, "clean:EDEEP": 2, "render:EX": 0, "render:EY": 0,
		}},
	)
	if n := r2nBlankReplaces(h, "e"); n != 0 {
		t.Fatalf("slot e: want no blank container replaces, got %d", n)
	}
	if n := len(h.doorActions(h.door("e"))); n != 6 {
		t.Fatalf("slot e: want 6 door replaces, got %d", n)
	}
	r2Run(h, "b", bare, "",
		r2Step[int]{set: 5, want: "", quiet: true},
		r2Step[int]{set: 15, want: "", quiet: true},
		r2Step[int]{set: 25, want: "", quiet: true},
		r2Step[int]{set: 31, want: "", quiet: true},
		r2Step[int]{set: 32, want: "", quiet: true},
		r2Step[int]{set: 35, want: "BDEEP"},
		r2Step[int]{set: 36, want: ""},
		r2Step[int]{set: 38, want: "", quiet: true},
		r2Step[int]{set: 42, want: "BDEEP"},
		r2Step[int]{set: 22, want: ""},
		r2Step[int]{set: 12, want: "", quiet: true},
		r2Step[int]{set: 0, want: "", quiet: true, counts: map[string]int{
			"render:BDEEP": 2, "clean:BDEEP": 2, "render:BX": 0, "render:BY": 0,
		}},
	)
	if n := r2nBlankReplaces(h, "b"); n != 2 {
		t.Fatalf("slot b: want 2 blank container replaces, got %d", n)
	}
	if n := len(h.doorActions(h.door("b"))); n != 4 {
		t.Fatalf("slot b: want 4 door replaces, got %d", n)
	}
}

type r2nCard struct {
	Name string
	Len  int
}

var r2nNames = []string{"zero", "One", "TWO", "three", "four", "Four", "six", "Seven", "eight", "nine"}

func r2nNameIndex(s string) int {
	for i, n := range r2nNames {
		if strings.EqualFold(n, s) {
			return i
		}
	}
	return -1
}

// N14: int -> int -> string (EqualFold) -> struct, with value routes on each
// level's comparator and a bottom RouteSource writing through every lens.
func TestRoute2NestedDeriveTypeChain(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(90)
	var bottom r2Box[Source[r2nCard]]
	d2 := func(s Source[int]) Source[string] {
		return DeriveSourceEqual(s, func(v int) string { return r2nNames[v] }, func(_ int, n string) int { return r2nNameIndex(n) }, strings.EqualFold)
	}
	d3 := func(s Source[string]) Source[r2nCard] {
		return DeriveSource(s, func(n string) r2nCard { return r2nCard{strings.ToLower(n), len(n)} }, func(_ string, c r2nCard) string { return c.Name })
	}
	h.render(r2Slot("c", src.Route(
		RouteDerive(r2Derive(h.rec, "d1", r2nTens),
			RouteValue(9, r2Comp[int](h.rec, "NINE")),
			RouteDerive(r2Derive(h.rec, "d2", d2),
				RouteValue("one", r2Comp[string](h.rec, "ONE")),
				RouteDerive(r2Derive(h.rec, "d3", d3),
					RouteValue(r2nCard{"two", 3}, r2Source[r2nCard](h.rec, "BOTTOM", &bottom)),
					RouteMatch(func(c r2nCard) bool { return c.Len > 4 }, r2Bind[r2nCard](h.rec, "CARD")),
				),
				r2Bind[string](h.rec, "S"),
			),
		),
	)))
	r2Run(h, "c", src, "NINE",
		r2Step[int]{set: 20, want: "BOTTOM"},
	)
	if ret := r2Set(h, bottom.last(), r2nCard{Name: "three"}); !r2SameErrs(ret, []error{nil}) {
		t.Fatalf("bottom write returned %v", ret)
	}
	if got := src.Get(); got != 30 {
		t.Fatalf("bottom write mapped to %d, want 30", got)
	}
	h.expect("c", "CARD={three 5}")
	h.oneDoor("c")
	r2Run(h, "c", src, "CARD={three 5}",
		r2Step[int]{set: 40, want: "S=four"},
		r2Step[int]{set: 50, want: "S=four", quiet: true},
		r2Step[int]{set: 10, want: "ONE"},
		r2Step[int]{set: 70, want: "CARD={seven 5}"},
		r2Step[int]{set: 20, want: "BOTTOM"},
	)
	if bottom.len() != 2 {
		t.Fatalf("want a fresh bottom render, got %d", bottom.len())
	}
	r2Mutate(h, bottom.last(), func(c r2nCard) r2nCard {
		if c != (r2nCard{"two", 3}) {
			t.Errorf("bottom source holds %v", c)
		}
		return r2nCard{Name: "EIGHT"}
	})
	if got := src.Get(); got != 80 {
		t.Fatalf("bottom write mapped to %d, want 80", got)
	}
	h.expect("c", "CARD={eight 5}")
	h.oneDoor("c")
	h.counts(map[string]int{
		"derive:d1": 1, "derive:d2": 1, "derive:d3": 1,
		"render:NINE": 1, "render:ONE": 1, "render:S": 1, "render:BOTTOM": 2, "render:CARD": 3,
	})
}

func r2nLadder(r *r2Rec, l, depth int) SourceRoute[int] {
	leaf := r2Comp[int](r, fmt.Sprintf("V%d", l))
	var kids []SourceRoute[int]
	if l >= 3 {
		kids = append(kids, RouteBeamValue(l*100+1, leaf))
	} else {
		kids = append(kids, RouteValue(l*100+1, leaf))
	}
	if l < depth {
		kids = append(kids, r2nLadder(r, l+1, depth))
	}
	kids = append(kids, r2Bind[int](r, fmt.Sprintf("D%d", l)))
	switch l {
	case 1:
		return RouteMatch(r2nGe(l*100), kids...)
	case 2:
		return RouteDerive(r2Derive(r, "d2", r2nIDSource), kids...)
	case 4:
		return RouteBeamDerive(r2DeriveBeam(r, "d4", r2nIDBeam), r2Beams(kids)...)
	}
	return RouteBeamMatch(r2nGe(l*100), r2Beams(kids)...)
}

// N15: one dynamic container at every depth from 1 to 5, alternating
// RouteMatch, RouteDerive, RouteBeamMatch and RouteBeamDerive.
func TestRoute2NestedOneDoorAnyDepth(t *testing.T) {
	for depth := 1; depth <= 5; depth++ {
		t.Run(strconv.Itoa(depth), func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(0)
			h.render(r2Slot("l", src.Route(r2nLadder(h.rec, 1, depth), r2Bind[int](h.rec, "TOP"))))
			var steps []r2Step[int]
			for l := 1; l <= depth; l++ {
				steps = append(steps, r2Step[int]{set: l*100 + 1, want: fmt.Sprintf("V%d", l)})
				d := l
				if l < depth && (l+1 == 2 || l+1 == 4) {
					d = l + 1
				}
				steps = append(steps, r2Step[int]{set: l*100 + 50, want: fmt.Sprintf("D%d=%d", d, l*100+50)})
			}
			steps = append(steps,
				r2Step[int]{set: 0, want: "TOP=0"},
				r2Step[int]{set: 101, want: "V1"},
			)
			if h.baseIDs != 2 {
				t.Fatalf("render created %d ids, want the root and one route door", h.baseIDs)
			}
			r2Run(h, "l", src, "TOP=0", steps...)
			if n := len(h.doorActions(h.door("l"))); n != len(steps) {
				t.Fatalf("want %d door replaces, got %d", len(steps), n)
			}
		})
	}
}

// N16: an explicit Route inside a RouteSource leaf is a second door with its
// own state, re-initialized on every activation of the outer branch.
func TestRoute2NestedExplicitInnerRoute(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	src := NewSource(0)
	h.render(r2Slot("o", src.Route(
		RouteMatch(r2nLt(100), RouteSource(func(s Source[int]) gox.Elem {
			h.rec.hit("outer-leaf")
			return s.Route(
				RouteValue(1, r2Comp[int](h.rec, "IA")),
				RouteDerive(r2Derive(h.rec, "inner", r2nIDSource), r2Bind[int](h.rec, "IB")),
			)
		})),
		r2Comp[int](h.rec, "OUT"),
	)))
	h.expect("o", "IB=0")
	if n := h.doorCount("o"); n != 2 {
		t.Fatalf("want 2 doors, got %d", n)
	}
	outer := h.door("o")
	inner := r2DoorIDs(h.html("o"))[1]
	if outer != h.doors0["o"] || inner == outer {
		t.Fatalf("unexpected doors outer=%d inner=%d", outer, inner)
	}
	for _, st := range []struct {
		v    int
		want string
	}{{1, "IA"}, {2, "IB=2"}, {3, "IB=3"}, {1, "IA"}} {
		n := len(h.doorActions(inner))
		h.quiet("o", func() { r2Set(h, src, st.v) })
		h.expect("o", st.want)
		if got := len(h.doorActions(inner)); got != n+1 {
			t.Fatalf("set %d: want one inner action, got %d", st.v, got-n)
		}
	}
	r2Set(h, src, 150)
	h.expect("o", "OUT")
	if n := h.doorCount("o"); n != 1 {
		t.Fatalf("want 1 door after the switch, got %d", n)
	}
	r2Set(h, src, 4)
	h.expect("o", "IB=4")
	ids := r2DoorIDs(h.html("o"))
	if len(ids) != 2 || ids[0] != outer || ids[1] == inner {
		t.Fatalf("want the outer door and a new inner door, got %v (outer=%d old inner=%d)", ids, outer, inner)
	}
	h.counts(map[string]int{"outer-leaf": 2, "derive:inner": 2, "render:OUT": 1, "clean:OUT": 1, "render:IA": 2})
	if n := len(h.doorActions(outer)); n != 2 {
		t.Fatalf("want 2 outer replaces, got %d", n)
	}
}

// N17: one route template placed at two positions inits two independent
// instances; the same value in a second tree inits a third. An end inside the
// second instance re-selects from scratch and picks the earlier X (15).
func TestRoute2NestedTemplateReuse(t *testing.T) {
	shared := func(h *r2Harness) SourceRoute[int] {
		return RouteDerive(r2Derive(h.rec, "d", r2nIDSource),
			RouteValue(7, r2Comp[int](h.rec, "SEVEN")),
			r2Bind[int](h.rec, "S"),
		)
	}
	tree := func(h *r2Harness, src Source[int], tmpl SourceRoute[int]) gox.Elem {
		return src.Route(
			RouteMatch(r2nLt(10), tmpl),
			RouteMatch(r2nLt(20), RouteValue(15, r2Comp[int](h.rec, "X")), tmpl),
			r2Comp[int](h.rec, "D"),
		)
	}
	t.Run("one-tree", func(t *testing.T) {
		h := newR2Harness(t)
		src := NewSource(1)
		h.render(r2Slot("t", tree(h, src, shared(h))))
		h.counts(map[string]int{"derive:d": 2, "render:S": 1})
		r2Run(h, "t", src, "S=1",
			r2Step[int]{set: 12, want: "S=12", counts: map[string]int{"render:S": 2, "clean:S": 1}},
			r2Step[int]{set: 7, want: "SEVEN"},
			r2Step[int]{set: 15, want: "X"},
			r2Step[int]{set: 25, want: "D"},
			r2Step[int]{set: 15, want: "X"},
			r2Step[int]{set: 16, want: "S=16"},
			r2Step[int]{set: 3, want: "S=3", counts: map[string]int{"render:S": 4}},
			r2Step[int]{set: 25, want: "D"},
			r2Step[int]{set: 3, want: "S=3", counts: map[string]int{"render:S": 5, "render:SEVEN": 1, "render:X": 2, "derive:d": 2}},
			r2Step[int]{set: 7, want: "SEVEN", counts: map[string]int{"render:SEVEN": 2}},
		)
	})
	t.Run("two-trees", func(t *testing.T) {
		h := newR2Harness(t)
		tmpl := shared(h)
		src, other := NewSource(1), NewSource(1)
		h.render(r2Slot("t", tree(h, src, tmpl)), r2Slot("u", other.Route(tmpl)))
		h.counts(map[string]int{"derive:d": 3, "render:S": 2})
		u := h.door("u")
		r2Run(h, "t", src, "S=1",
			r2Step[int]{set: 12, want: "S=12"},
			r2Step[int]{set: 7, want: "SEVEN"},
		)
		h.expect("u", "S=1")
		if n := len(h.doorActions(u)); n != 0 {
			t.Fatalf("second tree got %d actions from the first tree's source", n)
		}
		n := len(h.doorActions(h.door("t")))
		r2Run(h, "u", other, "S=1",
			r2Step[int]{set: 7, want: "SEVEN"},
			r2Step[int]{set: 8, want: "S=8"},
		)
		h.expect("t", "SEVEN")
		if got := len(h.doorActions(h.door("t"))); got != n {
			t.Fatalf("first tree got %d actions from the second tree's source", got-n)
		}
		h.counts(map[string]int{"derive:d": 3})
	})
}

const (
	r2nInt = iota
	r2nStr
	r2nPair
)

type r2nKV struct {
	K, V int
}

const (
	r2nOpComp = iota
	r2nOpBind
	r2nOpMatch
	r2nOpValue
	r2nOpDerive
)

const (
	r2nLeafComp = iota
	r2nLeafSource
	r2nLeafBeam
	r2nLeafNil
)

type r2nFn struct {
	name    string
	in, out int
	f       func(any) any
}

var r2nFns = []r2nFn{
	{"/10", r2nInt, r2nInt, func(v any) any { return v.(int) / 10 }},
	{"%10", r2nInt, r2nInt, func(v any) any { return v.(int) % 10 }},
	{"%4", r2nInt, r2nInt, func(v any) any { return v.(int) % 4 }},
	{"/7", r2nInt, r2nInt, func(v any) any { return v.(int) / 7 }},
	{"/2", r2nInt, r2nInt, func(v any) any { return v.(int) / 2 }},
	{"parity", r2nInt, r2nStr, func(v any) any {
		if v.(int)%2 == 0 {
			return "even"
		}
		return "odd"
	}},
	{"digit", r2nInt, r2nStr, func(v any) any { return strconv.Itoa(v.(int) % 10) }},
	{"band", r2nInt, r2nStr, func(v any) any {
		switch x := v.(int); {
		case x < 20:
			return "lo"
		case x < 40:
			return "mid"
		}
		return "hi"
	}},
	{"itoa", r2nInt, r2nStr, func(v any) any { return strconv.Itoa(v.(int)) }},
	{"kv/10", r2nInt, r2nPair, func(v any) any { return r2nKV{v.(int) / 10, v.(int) % 10} }},
	{"kv%3", r2nInt, r2nPair, func(v any) any { return r2nKV{v.(int) % 3, v.(int) / 3 % 5} }},
	{"len", r2nStr, r2nInt, func(v any) any { return len(v.(string)) }},
	{"atoi", r2nStr, r2nInt, func(v any) any {
		n, err := strconv.Atoi(v.(string))
		if err != nil {
			return -1
		}
		return n
	}},
	{"c0", r2nStr, r2nInt, func(v any) any { return int(v.(string)[0]) % 5 }},
	{"first", r2nStr, r2nStr, func(v any) any { return v.(string)[:1] }},
	{"upper", r2nStr, r2nStr, func(v any) any { return strings.ToUpper(v.(string)) }},
	{"rev", r2nStr, r2nStr, func(v any) any {
		b := []byte(v.(string))
		slices.Reverse(b)
		return string(b)
	}},
	{"K", r2nPair, r2nInt, func(v any) any { return v.(r2nKV).K }},
	{"V", r2nPair, r2nInt, func(v any) any { return v.(r2nKV).V }},
	{"K+V", r2nPair, r2nInt, func(v any) any { return v.(r2nKV).K + v.(r2nKV).V }},
	{"swap", r2nPair, r2nPair, func(v any) any { return r2nKV{v.(r2nKV).V, v.(r2nKV).K} }},
	{"kv", r2nPair, r2nStr, func(v any) any { return fmt.Sprintf("%d:%d", v.(r2nKV).K, v.(r2nKV).V) }},
}

type r2nEq struct {
	name string
	f    func(a, b any) bool
}

func (e r2nEq) equal(a, b any) bool { return e.f != nil && e.f(a, b) }

var r2nExact = r2nEq{"==", func(a, b any) bool { return a == b }}

var r2nCoarse = [...][]r2nEq{
	r2nInt: {
		{"~%2", func(a, b any) bool { return a.(int)%2 == b.(int)%2 }},
		{"~/3", func(a, b any) bool { return a.(int)/3 == b.(int)/3 }},
	},
	r2nStr: {
		{"~len", func(a, b any) bool { return len(a.(string)) == len(b.(string)) }},
		{"~c0", func(a, b any) bool { return a.(string)[0] == b.(string)[0] }},
	},
	r2nPair: {
		{"~K", func(a, b any) bool { return a.(r2nKV).K == b.(r2nKV).K }},
		{"~V", func(a, b any) bool { return a.(r2nKV).V == b.(r2nKV).V }},
	},
}

type r2nNode struct {
	op    int
	label string
	kind  int
	beam  bool
	pname string
	pred  func(any) bool
	sname string
	show  func(any) bool
	val   any
	fn    r2nFn
	eq    r2nEq
	out   r2nEq
	kids  []*r2nNode

	cur      int
	dflt     bool
	stored   any
	last     any
	none     bool
	renders  int
	contents int
}

type r2nStats struct {
	switches, breaks, skips, cached, falls, pairSkips int
}

func (n *r2nNode) isLeaf() bool { return n.op == r2nOpComp || n.op == r2nOpBind }

func (n *r2nNode) isNil() bool { return n.op == r2nOpComp && n.kind == r2nLeafNil }

func (n *r2nNode) content() bool {
	return n.op == r2nOpComp && !n.isNil() || n.op == r2nOpBind && (n.show == nil || n.show(n.stored))
}

func (n *r2nNode) text() string {
	switch {
	case !n.content():
		return ""
	case n.op == r2nOpBind:
		return n.label + "=" + fmt.Sprint(n.stored)
	}
	return n.label
}

func (n *r2nNode) holds(v any) bool {
	if n.op == r2nOpValue {
		return n.eq.equal(n.val, v)
	}
	return n.pred(v)
}

// seek reports whether the node matches v and, if it does, whether it is a
// default: leaves are, match and value groups never are, a derive is when its
// selected child is.
func (n *r2nNode) seek(v any, st *r2nStats) (ok, dflt bool) {
	switch n.op {
	case r2nOpComp, r2nOpBind:
		n.stored = v
		return true, true
	case r2nOpMatch, r2nOpValue:
		if !n.holds(v) {
			return false, false
		}
		if n.cur, n.dflt = r2nSeek(n.kids, v, st); n.cur < 0 {
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
	n.cur, n.dflt = r2nSeek(n.kids, w, st)
	n.last, n.none = w, n.cur < 0
	return !n.none, n.dflt
}

func r2nSeek(kids []*r2nNode, v any, st *r2nStats) (int, bool) {
	for i, k := range kids {
		if ok, dflt := k.seek(v, st); ok {
			return i, dflt
		}
	}
	return -1, false
}

func (n *r2nNode) leave() {
	if n.cur >= 0 {
		n.kids[n.cur].leave()
	}
	n.cur, n.dflt, n.stored, n.last, n.none = -1, false, nil, nil, false
}

const (
	r2nIntact = iota
	r2nShown
	r2nBroken
)

// r2nKeep walks the active path top-down. A level whose active child is a
// default first re-seeks its earlier siblings, and the first match takes over
// there; a failed predicate breaks the path; a derive whose new value equals
// the last one it acted on leaves its subtree alone.
func r2nKeep(kids []*r2nNode, cur *int, dflt *bool, v any, st *r2nStats) int {
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
					return r2nShown
				}
			}
		}
		a := kids[*cur]
		switch a.op {
		case r2nOpComp:
			return r2nIntact
		case r2nOpBind:
			if a.eq.equal(a.stored, v) {
				return r2nIntact
			}
			a.stored = v
			return r2nShown
		case r2nOpMatch, r2nOpValue:
			if !a.holds(v) {
				return r2nBroken
			}
			up = up[:0]
		default:
			w := a.fn.f(v)
			if a.out.equal(a.last, w) {
				if w != a.last {
					st.skips++
					if _, ok := w.(r2nKV); ok {
						st.pairSkips++
					}
				}
				return r2nIntact
			}
			a.last, v = w, w
			up = append(up, dflt)
		}
		kids, cur, dflt = a.kids, &a.cur, &a.dflt
	}
}

func (n *r2nNode) leaf() *r2nNode {
	for !n.isLeaf() {
		n = n.kids[n.cur]
	}
	return n
}

func (n *r2nNode) walk(f func(*r2nNode)) {
	f(n)
	for _, k := range n.kids {
		k.walk(f)
	}
}

func (n *r2nNode) describe(b *strings.Builder) {
	switch n.op {
	case r2nOpComp:
		fmt.Fprintf(b, "%s(%s)", n.label, [...]string{"comp", "source", "beam", "nil"}[n.kind])
		return
	case r2nOpBind:
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
	case r2nOpMatch:
		fmt.Fprintf(b, "%s{%s}", n.label, n.pname)
	case r2nOpValue:
		fmt.Fprintf(b, "%s{=%#v}", n.label, n.val)
	default:
		fmt.Fprintf(b, "%s{%s", n.label, n.fn.name)
		if n.out.name != r2nExact.name {
			fmt.Fprintf(b, " %s", n.out.name)
		}
		b.WriteString("}")
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

type r2nTree struct {
	slot  string
	entry int
	top   []*r2nNode
	on    bool
	cur   int
	dflt  bool
	blank bool
	acts  int
	views map[string]bool
	st    r2nStats
}

// apply runs one update: the container keeps an intact path, re-renders on a
// local switch or a changed bind, and on a break cleans the tree and selects
// from scratch in the same update, the broken branch included.
func (tr *r2nTree) apply(v any, first bool) {
	if tr.on {
		switch r2nKeep(tr.top, &tr.cur, &tr.dflt, v, &tr.st) {
		case r2nIntact:
			tr.note()
			return
		case r2nBroken:
			tr.st.breaks++
			tr.top[tr.cur].leave()
			tr.on = false
		}
	}
	if !tr.on {
		tr.cur, tr.dflt = r2nSeek(tr.top, v, &tr.st)
		tr.on = tr.cur >= 0
	}
	tr.show(first)
	tr.note()
}

func (tr *r2nTree) show(first bool) {
	var leaf *r2nNode
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

func (tr *r2nTree) note() {
	_, key, _ := tr.view()
	tr.views[key] = true
}

func (tr *r2nTree) view() (text, key string, depth int) {
	if !tr.on {
		return "", "-", 0
	}
	n := tr.top[tr.cur]
	for !n.isLeaf() {
		key += n.label + "/"
		depth++
		n = n.kids[n.cur]
	}
	if n.isNil() {
		return "", key + n.label + "(nil)", depth
	}
	return n.text(), key + n.label + ":" + n.text(), depth
}

func (tr *r2nTree) describe() string {
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

type r2nGen struct {
	rng   *rand.Rand
	tree  int
	n     int
	depth int
}

func (g *r2nGen) label(p string) string {
	g.n++
	return fmt.Sprintf("%s%d_%d", p, g.tree, g.n)
}

func (g *r2nGen) pick(dom []any) any {
	return dom[g.rng.IntN(len(dom))]
}

func (g *r2nGen) predicate(typ int, dom []any) (string, func(any) bool) {
	switch r := g.rng.IntN(100); {
	case r < 5:
		return "true", func(any) bool { return true }
	case r < 9:
		return "false", func(any) bool { return false }
	}
	switch typ {
	case r2nInt:
		k := g.pick(dom).(int)
		switch g.rng.IntN(8) {
		case 0:
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
		case 6:
			return fmt.Sprintf("<%d", k), func(v any) bool { return v.(int) < k }
		}
		return fmt.Sprintf("==%d", k), func(v any) bool { return v.(int) == k }
	case r2nPair:
		p := g.pick(dom).(r2nKV)
		switch g.rng.IntN(4) {
		case 0:
			return fmt.Sprintf("K==%d", p.K), func(v any) bool { return v.(r2nKV).K == p.K }
		case 1:
			return fmt.Sprintf("V<%d", p.V), func(v any) bool { return v.(r2nKV).V < p.V }
		case 2:
			m := 2 + g.rng.IntN(2)
			rem := g.rng.IntN(m)
			return fmt.Sprintf("V%%%d==%d", m, rem), func(v any) bool { return v.(r2nKV).V%m == rem }
		}
		return fmt.Sprintf("==%v", p), func(v any) bool { return v.(r2nKV) == p }
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

func r2nMapDom(dom []any, f func(any) any) []any {
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

func (g *r2nGen) group(depth, typ int, dom []any, beam bool, eq r2nEq, lo, hi int) []*r2nNode {
	kids := make([]*r2nNode, lo+g.rng.IntN(hi-lo+1))
	for i := range kids {
		kids[i] = g.node(depth, typ, dom, beam, eq)
	}
	if len(kids) > 1 && kids[0].isLeaf() && g.rng.IntN(4) != 0 {
		kids = append(kids[1:], kids[0])
	}
	return kids
}

func (g *r2nGen) comparator(typ int) r2nEq {
	exact := 5
	if typ == r2nPair {
		exact = 2
	}
	switch r := g.rng.IntN(10); {
	case r < exact:
		return r2nExact
	case r < 8:
		return r2nCoarse[typ][g.rng.IntN(len(r2nCoarse[typ]))]
	}
	return r2nEq{name: "nil"}
}

func (g *r2nGen) leaf(n *r2nNode, typ int, dom []any, beam bool) *r2nNode {
	switch r := g.rng.IntN(20); {
	case r < 9:
		n.op, n.label = r2nOpBind, g.label("B")
		if g.rng.IntN(4) == 0 {
			n.sname, n.show = g.predicate(typ, dom)
		}
	case r < 18:
		n.op, n.label = r2nOpComp, g.label("L")
		if beam {
			n.kind = 2 * g.rng.IntN(2)
		} else {
			n.kind = g.rng.IntN(3)
		}
	default:
		n.op, n.label, n.kind = r2nOpComp, g.label("N"), r2nLeafNil
	}
	return n
}

func (g *r2nGen) node(depth, typ int, dom []any, beam bool, eq r2nEq) *r2nNode {
	n := &r2nNode{cur: -1, eq: eq}
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
		n.op, n.label = r2nOpMatch, g.label("M")
		n.pname, n.pred = g.predicate(typ, dom)
		if g.rng.IntN(12) != 0 {
			n.kids = g.group(depth+1, typ, dom, beam, eq, 1, 3)
		}
	case r < 72:
		n.op, n.label = r2nOpValue, g.label("V")
		n.val = g.pick(dom)
		if g.rng.IntN(10) == 0 {
			n.val = [...]any{1000, "zz", r2nKV{-1, -1}}[typ]
		}
		n.kids = g.group(depth+1, typ, dom, beam, eq, 1, 2)
	default:
		n.op, n.label = r2nOpDerive, g.label("D")
		var fns []r2nFn
		for _, fn := range r2nFns {
			if fn.in == typ {
				fns = append(fns, fn)
			}
		}
		n.fn = fns[g.rng.IntN(len(fns))]
		n.out = g.comparator(n.fn.out)
		if g.rng.IntN(20) != 0 {
			n.kids = g.group(depth+1, n.fn.out, r2nMapDom(dom, n.fn.f), beam, n.out, 1, 3)
		}
	}
	return n
}

func (g *r2nGen) gen(i int, dom []any) *r2nTree {
	g.tree, g.n = i, 0
	g.depth = 2 + g.rng.IntN(4)
	tr := &r2nTree{slot: fmt.Sprintf("t%d", i), entry: g.rng.IntN(5), cur: -1, blank: true, views: map[string]bool{}}
	tr.top = g.group(0, r2nInt, dom, tr.entry >= 2, r2nExact, 2, 4)
	return tr
}

func r2nSrcRoutes[T comparable](r *r2Rec, ns []*r2nNode) []SourceRoute[T] {
	out := make([]SourceRoute[T], len(ns))
	for i, n := range ns {
		out[i] = r2nSrcRoute[T](r, n)
	}
	return out
}

func r2nBeamRoutes[T comparable](r *r2Rec, ns []*r2nNode) []BeamRoute[T] {
	out := make([]BeamRoute[T], len(ns))
	for i, n := range ns {
		out[i] = r2nBeamRoute[T](r, n)
	}
	return out
}

func r2nSrcRoute[T comparable](r *r2Rec, n *r2nNode) SourceRoute[T] {
	switch {
	case n.op == r2nOpComp && n.kind == r2nLeafSource:
		return r2Source[T](r, n.label, nil)
	case n.beam || n.isLeaf():
		return r2nBeamRoute[T](r, n)
	case n.op == r2nOpMatch:
		return RouteMatch(func(v T) bool { return n.pred(v) }, r2nSrcRoutes[T](r, n.kids)...)
	case n.op == r2nOpValue:
		return RouteValue(n.val.(T), r2nSrcRoutes[T](r, n.kids)...)
	}
	switch n.fn.out {
	case r2nInt:
		return RouteDerive(r2Derive(r, n.label, r2nDeriveSource[T, int](n)), r2nSrcRoutes[int](r, n.kids)...)
	case r2nStr:
		return RouteDerive(r2Derive(r, n.label, r2nDeriveSource[T, string](n)), r2nSrcRoutes[string](r, n.kids)...)
	}
	return RouteDerive(r2Derive(r, n.label, r2nDeriveSource[T, r2nKV](n)), r2nSrcRoutes[r2nKV](r, n.kids)...)
}

func r2nBeamRoute[T comparable](r *r2Rec, n *r2nNode) BeamRoute[T] {
	switch {
	case n.op == r2nOpComp && n.kind == r2nLeafSource:
		panic("r2n: RouteSource leaf under a beam container")
	case n.op == r2nOpComp && n.kind == r2nLeafBeam:
		return r2Beam[T](r, n.label, nil)
	case n.op == r2nOpComp && n.kind == r2nLeafNil:
		return RouteComp[T](nil)
	case n.op == r2nOpComp:
		return r2Comp[T](r, n.label)
	case n.op == r2nOpBind && n.show != nil:
		return r2BindIf(r, n.label, func(v T) bool { return n.show(v) })
	case n.op == r2nOpBind:
		return r2Bind[T](r, n.label)
	case n.op == r2nOpMatch:
		return RouteBeamMatch(func(v T) bool { return n.pred(v) }, r2nBeamRoutes[T](r, n.kids)...)
	case n.op == r2nOpValue:
		return RouteBeamValue(n.val.(T), r2nBeamRoutes[T](r, n.kids)...)
	}
	switch n.fn.out {
	case r2nInt:
		return RouteBeamDerive(r2DeriveBeam(r, n.label, r2nDeriveBeam[T, int](n)), r2nBeamRoutes[int](r, n.kids)...)
	case r2nStr:
		return RouteBeamDerive(r2DeriveBeam(r, n.label, r2nDeriveBeam[T, string](n)), r2nBeamRoutes[string](r, n.kids)...)
	}
	return RouteBeamDerive(r2DeriveBeam(r, n.label, r2nDeriveBeam[T, r2nKV](n)), r2nBeamRoutes[r2nKV](r, n.kids)...)
}

func r2nTyped[U any](e r2nEq) func(a, b U) bool {
	if e.f == nil {
		return nil
	}
	return func(a, b U) bool { return e.f(a, b) }
}

func r2nDeriveSource[T, U comparable](n *r2nNode) func(Source[T]) Source[U] {
	get := func(v T) U { return n.fn.f(v).(U) }
	set := func(a T, _ U) T { return a }
	return func(s Source[T]) Source[U] {
		if n.out.name == r2nExact.name {
			return DeriveSource(s, get, set)
		}
		return DeriveSourceEqual(s, get, set, r2nTyped[U](n.out))
	}
}

func r2nDeriveBeam[T, U comparable](n *r2nNode) func(Beam[T]) Beam[U] {
	get := func(v T) U { return n.fn.f(v).(U) }
	return func(b Beam[T]) Beam[U] {
		if n.out.name == r2nExact.name {
			return DeriveBeam(b, get)
		}
		return DeriveBeamEqual(b, get, r2nTyped[U](n.out))
	}
}

func r2nMount(r *r2Rec, src Source[int], tr *r2nTree) gox.Elem {
	switch tr.entry {
	case 0:
		return src.Route(r2nSrcRoutes[int](r, tr.top)...)
	case 1:
		return r2nIDSource(src).Route(r2nSrcRoutes[int](r, tr.top)...)
	case 2:
		return src.RouteBeam(r2nBeamRoutes[int](r, tr.top)...)
	case 3:
		return r2nIDBeam(src).RouteBeam(r2nBeamRoutes[int](r, tr.top)...)
	}
	return r2nIDSource(src).RouteBeam(r2nBeamRoutes[int](r, tr.top)...)
}

type r2nPage struct {
	h       *r2Harness
	trees   []*r2nTree
	dom     string
	seen    int
	acts    map[uint64]int
	allowed map[uint64]bool
	hist    []int
}

func (pg *r2nPage) check(at string) {
	h := pg.h
	h.t.Helper()
	acts := h.client.actions()
	fresh := acts[pg.seen:]
	for _, a := range fresh {
		if !r2IsDoorAction(a) {
			continue
		}
		if !pg.allowed[a.id] || a.name != "door_replace" {
			h.t.Fatalf("%s: %s for door %d, outside the render-time route doors", at, a.name, a.id)
		}
		pg.acts[a.id]++
	}
	pg.dom = r2DOM(pg.dom, fresh)
	pg.seen = len(acts)
	if ids := h.inst.ids.Load(); ids != h.baseIDs {
		h.t.Fatalf("%s: ids moved from %d to %d", at, h.baseIDs, ids)
	}
	for _, tr := range pg.trees {
		id := h.doors0[tr.slot]
		inner, ok := r2SlotInner(pg.dom, tr.slot)
		if !ok {
			h.t.Fatalf("%s: slot %s missing, dom:\n%s", at, tr.slot, pg.dom)
		}
		if ids := r2DoorIDs(inner); len(ids) != 1 || ids[0] != id {
			h.t.Fatalf("%s: slot %s has doors %v, want only %d\ntree: %s", at, tr.slot, ids, id, tr.describe())
		}
		want, key, _ := tr.view()
		if got := r2Text(inner); got != want || pg.acts[id] != tr.acts {
			h.t.Fatalf("%s: slot %s: text %q, want %q (path %s); door replaces %d, want %d\ntree: %s\nvalues: %v",
				at, tr.slot, got, want, key, pg.acts[id], tr.acts, tr.describe(), pg.hist)
		}
	}
}

func (pg *r2nPage) final() {
	pg.h.t.Helper()
	want := map[string]int{}
	for _, tr := range pg.trees {
		var shown *r2nNode
		if tr.on && !tr.blank {
			shown = tr.top[tr.cur].leaf()
		}
		for _, top := range tr.top {
			top.walk(func(n *r2nNode) {
				switch {
				case n.isNil():
				case n.isLeaf():
					want["render:"+n.label] = n.renders
					want["clean:"+n.label] = n.contents
					if n == shown && n.content() {
						want["clean:"+n.label]--
					}
				case n.op == r2nOpDerive:
					want["derive:"+n.label] = 1
				}
			})
		}
	}
	pg.h.counts(want)
}

// An independent model of the routing rules checked against 240 generated
// trees. Leaves always match and are defaults; a match or value group matches
// when its own condition holds and one of its children matches, a derive when
// one of its children matches the derived value, so a group whose children
// match nothing falls through to its later siblings. A derive is a default
// when its selected child is. Selection is the first match in order, depth
// first. An update walks the intact active path top-down: a level whose
// active child is a default first re-seeks its earlier siblings, and the
// first match takes over there and clears the default flags above; a failed
// predicate on the path cleans the tree and re-selects from scratch in the
// same update; a derive whose derived value equals the last one it acted on
// (by its own comparator: exact, coarse such as one struct field, or nil)
// leaves everything below untouched, and in selection keeps a no-match for an
// equal value; binds re-render on a change by their level's comparator.
// Leaves may render nothing: RouteComp(nil) blanks the container, a bind
// returning nil does not. No door replace between two blank views. Checked:
// text, door replace count and the single door after every update, then
// exact render, clean and derive counters.

func TestRoute2NestedOracle(t *testing.T) {
	const pages, perPage, updates = 10, 24, 64
	dom := make([]any, 60)
	for i := range dom {
		dom[i] = i
	}
	var trees, lively, blanks, nils, deepest int
	var st r2nStats
	for p := range pages {
		t.Run(fmt.Sprintf("page%d", p), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(20261012, uint64(p)))
			g := &r2nGen{rng: rng}
			workers := 8
			if p%3 == 2 {
				workers = 1
			}
			h := newR2HarnessN(t, workers)
			start := rng.IntN(len(dom))
			src := NewSource(start)
			pg := &r2nPage{h: h, acts: map[uint64]int{}, allowed: map[uint64]bool{}, hist: []int{start}}
			var els []gox.Elem
			for i := range perPage {
				tr := g.gen(i, dom)
				pg.trees = append(pg.trees, tr)
				els = append(els, r2Slot(tr.slot, r2nMount(h.rec, src, tr)))
			}
			h.render(els...)
			pg.dom = h.initial
			for _, id := range h.doors0 {
				pg.allowed[id] = true
			}
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
				ret := r2Set(h, src, v)
				want := []error{}
				if v != prev {
					want = []error{nil}
					for _, tr := range pg.trees {
						tr.apply(v, false)
					}
				}
				if !r2SameErrs(ret, want) {
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
	t.Logf("%d trees, %d with 3+ distinct views, %d with a blank view (no match), %d with a nil leaf view, deepest active path %d; %d fall-throughs, %d local switches, %d breaks, %d coarse skips (%d on struct fields), %d coarse cached no-matches",
		trees, lively, blanks, nils, deepest, st.falls, st.switches, st.breaks, st.skips, st.pairSkips, st.cached)
	if trees == pages*perPage && (lively*2 < trees || blanks == 0 || nils == 0 || deepest < 4 ||
		st.falls == 0 || st.switches == 0 || st.breaks == 0 || st.skips == 0 || st.pairSkips == 0 || st.cached == 0) {
		t.Fatalf("generator too tame: %d/%d lively, %d blank, %d nil, depth %d, %+v", lively, trees, blanks, nils, deepest, st)
	}
}

// Old: TestRouteBindMethods (to-default -> to-match) and TestRouteSourceAndBeamState
// (route-source-lens): an inner default yields to an earlier inner sibling.
func TestRoute2NestedPortInnerDefaultYields(t *testing.T) {
	for _, c := range []r2nCase{r2nPort01Depth2, r2nPort01Depth3} {
		t.Run(c.name, func(t *testing.T) { r2nRunCase(t, c) })
	}
}

// Old: beam_route.go routeRender sticky rule (F20), inside a nested group.
func TestRoute2NestedPortInnerSticky(t *testing.T) { r2nRunCase(t, r2nPort02) }

// Old: TestRouteSourceAndBeamState (route-lens-clear): when the inner
// non-default ends, the re-selection renders the inner default fresh under the
// still matching outer match; the top default is never rendered.
func TestRoute2NestedPortInnerDefaultFresh(t *testing.T) {
	for _, e := range r2Entries {
		if e.beam {
			continue
		}
		t.Run(e.name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(5)
			var srcs r2Box[Source[int]]
			var beams r2Box[Beam[int]]
			h.render(r2Slot("n", e.mount(src,
				RouteMatch(r2nGt(0),
					RouteMatch(r2nGt(10), r2Source[int](h.rec, "A", &srcs)),
					r2Beam[int](h.rec, "ID", &beams),
				),
				r2Comp[int](h.rec, "TOP"),
			)))
			r2Run(h, "n", src, "ID",
				r2Step[int]{set: 7, want: "ID", quiet: true},
				r2Step[int]{set: 15, want: "A"},
				r2Step[int]{set: 3, want: "ID", counts: map[string]int{"render:ID": 2, "clean:ID": 1, "clean:A": 1, "render:TOP": 0}},
				r2Step[int]{set: 16, want: "A", counts: map[string]int{"render:A": 2, "render:TOP": 0}},
				r2Step[int]{set: 0, want: "TOP"},
			)
			if srcs.len() != 2 || beams.len() != 2 || beams.last().Get() != 0 {
				t.Fatalf("captured %d sources and %d beams (last beam %d)", srcs.len(), beams.len(), beams.last().Get())
			}
			r2Set(h, srcs.at(0), 12)
			if src.Get() != 12 {
				t.Fatalf("leaf source wrote %d to the root", src.Get())
			}
			h.expect("n", "A")
			h.oneDoor("n")
		})
	}
}

// Old: TestRouteSourceAndBeamState list routes and TestRouteNoDefaultHandlesRapidUnmatch:
// an outer match with no nested match does not match, and with no default the
// container is blank.
func TestRoute2NestedPortNoMatchBlank(t *testing.T) { r2nRunCase(t, r2nPort04) }

// Old: TestRouteSourceAndBeamState (route-lens-clear) and TestRouterSourceCombined.
func TestRoute2NestedPortOuterSwitchCleansSubtree(t *testing.T) { r2nRunCase(t, r2nPort05) }

func r2nSection(s Source[r2PathItem]) Source[string] {
	return DeriveSource(s, func(p r2PathItem) string { return p.Tab }, func(p r2PathItem, tab string) r2PathItem {
		p.Tab = tab
		return p
	})
}

// Old: TestRouteSourceAndBeamState initial (F01/F02) and
// TestRouterBeamCombinedCustomEncoderRoute initial (M07): the first seek
// descends the whole tree inside the page render, falling through the model
// route to DEF when the decoded model's tab matches nothing.
func TestRoute2NestedPortInitialDeepInline(t *testing.T) {
	tree := func(h *r2Harness, box *r2Box[Source[string]]) gox.Elem {
		return Route(
			RouteModel[r2PathItem](
				RouteDerive(r2Derive(h.rec, "sec", r2nSection),
					RouteValue("info", RouteMatch(func(s string) bool { return s != "" },
						r2Source[string](h.rec, "LEAF", box))),
				),
			),
			r2Comp[Location](h.rec, "DEF"),
		)
	}
	t.Run("leaf", func(t *testing.T) {
		h := r2nRouterHarness(t, "/item/5?tab=info")
		var box r2Box[Source[string]]
		h.render(r2Slot("p", tree(h, &box)))
		if want := fmt.Sprintf(`<d0-r data-d0p="1" id="d0r%d">LEAF</d0-r>`, h.doors0["p"]); !strings.Contains(h.initial, want) {
			t.Fatalf("leaf not inline in the page:\n%s", h.initial)
		}
		h.expect("p", "LEAF")
		h.oneDoor("p")
		if acts := h.client.actions(); len(acts) != 0 {
			t.Fatalf("expected no client actions, got %+v", acts)
		}
		if box.len() != 1 || box.last().Get() != "info" {
			t.Fatalf("leaf source: %d captured", box.len())
		}
		h.counts(map[string]int{"render:LEAF": 1, "render:DEF": 0, "derive:sec": 1})
	})
	t.Run("fallthrough", func(t *testing.T) {
		h := r2nRouterHarness(t, "/item/5?tab=other")
		h.render(r2Slot("p", tree(h, nil)))
		if want := fmt.Sprintf(`<d0-r data-d0p="1" id="d0r%d">DEF</d0-r>`, h.doors0["p"]); !strings.Contains(h.initial, want) {
			t.Fatalf("default not inline in the page:\n%s", h.initial)
		}
		h.expect("p", "DEF")
		h.oneDoor("p")
		if acts := h.client.actions(); len(acts) != 0 {
			t.Fatalf("expected no client actions, got %+v", acts)
		}
		r2Run(h, "p", Router(h.page), "DEF",
			r2Step[Location]{set: r2Loc("/item/6?tab=other"), want: "DEF", quiet: true},
			r2Step[Location]{set: r2Loc("/item/6?tab=info"), want: "LEAF"},
			r2Step[Location]{set: r2Loc("/nope"), want: "DEF", counts: map[string]int{"render:LEAF": 1, "render:DEF": 2, "derive:sec": 1}},
		)
	})
}

// Old: TestRouteSourceAndBeamState (route-source-next), M05/M08: updates on
// the same path are quiet for render-once leaves; a deep bind re-renders only
// on a changed value.
func TestRoute2NestedPortSamePathQuiet(t *testing.T) {
	leaves := map[string]func(r *r2Rec) SourceRoute[int]{
		"comp":   func(r *r2Rec) SourceRoute[int] { return r2Comp[int](r, "LEAF") },
		"source": func(r *r2Rec) SourceRoute[int] { return r2Source[int](r, "LEAF", nil) },
		"beam":   func(r *r2Rec) SourceRoute[int] { return r2Beam[int](r, "LEAF", nil) },
	}
	for _, name := range []string{"comp", "source", "beam"} {
		t.Run(name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(1)
			h.render(r2Slot("q", src.Route(
				RouteMatch(r2nGt(0), RouteDerive(r2Derive(h.rec, "tens", r2nTens),
					RouteMatch(r2nGe(5), leaves[name](h.rec)),
					r2Bind[int](h.rec, "B"),
				)),
			)))
			r2Run(h, "q", src, "B=0",
				r2Step[int]{set: 2, want: "B=0", quiet: true},
				r2Step[int]{set: 9, want: "B=0", quiet: true},
				r2Step[int]{set: 15, want: "B=1"},
				r2Step[int]{set: 50, want: "LEAF"},
				r2Step[int]{set: 55, want: "LEAF", quiet: true},
				r2Step[int]{set: 60, want: "LEAF", quiet: true},
				r2Step[int]{set: 99, want: "LEAF", quiet: true},
				r2Step[int]{set: 41, want: "B=4"},
				r2Step[int]{set: 0, want: ""},
				r2Step[int]{set: 7, want: "B=0", counts: map[string]int{
					"render:LEAF": 1, "clean:LEAF": 1, "render:B": 4, "clean:B": 3, "derive:tens": 1,
				}},
			)
		})
	}
}

// Old analogue: RouteDefault* (default) vs RouteDerive/RouteModel (never a
// default) in routeRender. A derive is a default whenever its active child is a
// default, at any index, and sticky while it is not; wrapped in a match it is
// always sticky.
func TestRoute2NestedPortDeriveDefaultIndex(t *testing.T) {
	h := newR2Harness(t)
	a, b, c := NewSource(1), NewSource(1), NewSource(1)
	mod10 := func(s Source[int]) Source[int] {
		return DeriveSource(s, func(v int) int { return v % 10 }, func(a, b int) int { return a/10*10 + b })
	}
	h.render(
		r2Slot("a", a.Route(
			RouteMatch(r2nGe(100), r2Comp[int](h.rec, "A1")),
			RouteDerive(mod10, r2Bind[int](h.rec, "R1")),
		)),
		r2Slot("b", b.Route(
			RouteMatch(r2nGe(100), r2Comp[int](h.rec, "A2")),
			RouteDerive(mod10, RouteMatch(func(v int) bool { return v == 9 }, r2Comp[int](h.rec, "B2")), r2Bind[int](h.rec, "R2")),
		)),
		r2Slot("c", c.Route(
			RouteMatch(r2nGe(100), r2Comp[int](h.rec, "A3")),
			RouteMatch(r2nAlways[int], RouteDerive(mod10, r2Bind[int](h.rec, "R3"))),
		)),
	)
	r2Run(h, "a", a, "R1=1",
		r2Step[int]{set: 5, want: "R1=5"},
		r2Step[int]{set: 100, want: "A1"},
		r2Step[int]{set: 101, want: "A1", quiet: true},
		r2Step[int]{set: 7, want: "R1=7"},
		r2Step[int]{set: 200, want: "A1"},
	)
	r2Run(h, "b", b, "R2=1",
		r2Step[int]{set: 100, want: "A2"},
		r2Step[int]{set: 101, want: "A2", quiet: true},
		r2Step[int]{set: 9, want: "B2"},
		r2Step[int]{set: 109, want: "B2", quiet: true},
		r2Step[int]{set: 5, want: "R2=5"},
		r2Step[int]{set: 150, want: "A2", counts: map[string]int{"render:A2": 2, "render:B2": 1, "render:R2": 2}},
	)
	r2Run(h, "c", c, "R3=1",
		r2Step[int]{set: 100, want: "R3=0"},
		r2Step[int]{set: 105, want: "R3=5", counts: map[string]int{"render:A3": 0}},
	)
}

// Old: TestRouteSourceAndBeamState (route-lens-update),
// TestRouteSourceAndDerivedBeamEntrypoints (two-level write) and
// TestHistoryReplaceContextPathModel: a deep write maps back through every
// lens to the Router location.
func TestRoute2NestedPortDeepWriteBack(t *testing.T) {
	h := r2nRouterHarness(t, "/item/5?tab=info")
	var boxA, boxB r2Box[Source[string]]
	h.render(r2Slot("w", Route(
		RouteModel[r2PathItem](
			RouteDerive(r2Derive(h.rec, "sec", r2nSection),
				RouteMatch(func(s string) bool { return strings.HasPrefix(s, "info") }, r2Source[string](h.rec, "A", &boxA)),
				RouteValue("edit", r2Source[string](h.rec, "B", &boxB)),
			),
		),
		r2Comp[Location](h.rec, "DEF"),
	)))
	loc := func() string { return h.inst.location.Get().String() }
	h.expect("w", "A")
	h.oneDoor("w")
	h.quiet("w", func() {
		if ret := r2Set(h, boxA.last(), "info2"); !r2SameErrs(ret, []error{nil}) {
			t.Fatalf("write returned %v", ret)
		}
	})
	h.expect("w", "A")
	if loc() != "/item/5?tab=info2" {
		t.Fatalf("location %s", loc())
	}
	if ret := r2Set(h, boxA.last(), "info2"); !r2SameErrs(ret, []error{}) {
		t.Fatalf("equal write returned %v", ret)
	}
	if ret := r2Set(h, boxA.last(), "edit"); !r2SameErrs(ret, []error{nil}) {
		t.Fatalf("write returned %v", ret)
	}
	h.expect("w", "B")
	if loc() != "/item/5?tab=edit" {
		t.Fatalf("location %s", loc())
	}
	if ret := r2nUpdateReplace(h, boxB.last(), "info"); !r2SameErrs(ret, []error{nil}) {
		t.Fatalf("write returned %v", ret)
	}
	h.expect("w", "A")
	if loc() != "/item/5?tab=info" {
		t.Fatalf("location %s", loc())
	}
	h.oneDoor("w")
	want := []string{"&{/item/5?tab=info2 false}", "&{/item/5?tab=edit false}", "&{/item/5?tab=info true}"}
	if got := r2nPaths(h); !slices.Equal(got, want) {
		t.Fatalf("set_path %q, want %q", got, want)
	}
	h.counts(map[string]int{"render:A": 2, "clean:A": 1, "render:B": 1, "clean:B": 1, "render:DEF": 0, "derive:sec": 1})
}

// Old: RouteModelBeam with DeriveBeam(...).RouteBeam (docs/05-routing.md,
// "Dispatch on a derived value"): only a variant change swaps.
func TestRoute2NestedPortModelVariantBeam(t *testing.T) {
	h := r2nRouterHarness(t, "/item/5?tab=info")
	h.render(r2Slot("v", func(cur gox.Cursor) error {
		return cur.Comp(Router(cur.Context()).RouteBeam(
			RouteBeamModel[r2PathItem](
				RouteBeamDerive(r2DeriveBeam(h.rec, "variant", func(b Beam[r2PathItem]) Beam[string] {
					return DeriveBeam(b, func(p r2PathItem) string { return p.Tab })
				}),
					RouteBeamValue("info", r2Comp[string](h.rec, "V1")),
					RouteBeamValue("edit", r2Comp[string](h.rec, "V2")),
					r2Comp[string](h.rec, "O"),
				),
			),
			r2Comp[Location](h.rec, "DEF"),
		))
	}))
	r2Run(h, "v", Router(h.page), "V1",
		r2Step[Location]{set: r2Loc("/item/6?tab=info"), want: "V1", quiet: true},
		r2Step[Location]{set: r2Loc("/item/7?tab=info"), want: "V1", quiet: true},
		r2Step[Location]{set: r2Loc("/item/7?tab=edit"), want: "V2"},
		r2Step[Location]{set: r2Loc("/item/8?tab=edit"), want: "V2", quiet: true},
		r2Step[Location]{set: r2Loc("/item/8"), want: "O"},
		r2Step[Location]{set: r2Loc("/item/9"), want: "O", quiet: true},
		r2Step[Location]{set: r2Loc("/item/9?tab=info"), want: "V1"},
		r2Step[Location]{set: r2Loc("/a"), want: "DEF"},
		r2Step[Location]{set: r2Loc("/item/1?tab=edit"), want: "V2", counts: map[string]int{
			"render:V1": 2, "render:V2": 2, "render:O": 1, "render:DEF": 1, "derive:variant": 1,
		}},
	)
}

type r2nAdminUsers struct {
	Users bool `path:"/admin/users"`
}

type r2nAdminPosts struct {
	Posts bool `path:"/admin/posts/:ID"`
	ID    int
}

type r2nHome struct {
	Home bool `path:"/"`
}

// Old: TestRouterSourceCombinedPathModelCustomAndRawRoutes and
// TestRouterBeamModelTransitionsAndDefault404, composed two levels deep
// (docs/05-routing.md, "Compose levels").
func TestRoute2NestedPortSectionModels(t *testing.T) {
	h := r2nRouterHarness(t, "/")
	isAdmin := func(l Location) bool { return len(l.Segments) > 0 && l.Segments[0] == "admin" }
	h.render(r2Slot("s", Route(
		RouteMatch(isAdmin,
			RouteModel[r2nAdminUsers](r2Source[r2nAdminUsers](h.rec, "U", nil)),
			RouteModel[r2nAdminPosts](r2Bind[r2nAdminPosts](h.rec, "P")),
			r2Comp[Location](h.rec, "ANF"),
		),
		RouteModel[r2nHome](r2Source[r2nHome](h.rec, "H", nil)),
		r2Comp[Location](h.rec, "NF"),
	)))
	r2Run(h, "s", Router(h.page), "H",
		r2Step[Location]{set: r2Loc("/admin/users"), want: "U"},
		r2Step[Location]{set: r2Loc("/admin/posts/3"), want: "P={true 3}"},
		r2Step[Location]{set: r2Loc("/admin/posts/4"), want: "P={true 4}"},
		r2Step[Location]{set: r2Loc("/admin/zzz"), want: "ANF"},
		r2Step[Location]{set: r2Loc("/admin"), want: "ANF", quiet: true},
		r2Step[Location]{set: r2Loc("/admin/users"), want: "U"},
		r2Step[Location]{set: r2Loc("/zzz"), want: "NF"},
		r2Step[Location]{set: r2Loc("/"), want: "H"},
		r2Step[Location]{set: r2Loc("/admin/posts/9"), want: "P={true 9}", counts: map[string]int{
			"render:U": 2, "render:P": 3, "render:ANF": 1, "render:H": 2, "render:NF": 1,
		}},
	)
}

// Old: TestRouteNoDefaultHandlesRapidUnmatch, projected: back-to-back updates
// through nested matches settle on the final value's branch. The outer
// predicate runs once per value the route sees, and once more when that value
// breaks the active A or B path, since the re-selection tries the ended branch
// again.
func TestRoute2NestedPortRapidNested(t *testing.T) {
	final := func(v int) string {
		switch {
		case v <= 0 || v >= 100:
			return "D"
		case v < 50:
			return "A"
		}
		return "B"
	}
	seqs := [][]int{{10, 200, 60}, {10, 60, 10, 60, 5}, {60, 0, 60, 300}, {5, 55, 5, 300, 55}, {200, 10, 0}, {70, 20, 0, 30}}
	for _, noSkip := range []bool{false, true} {
		for _, held := range []bool{false, true} {
			t.Run(fmt.Sprintf("noskip=%v/held=%v", noSkip, held), func(t *testing.T) {
				h := newR2Harness(t)
				src := NewSource(0)
				if noSkip {
					src = NewSourceNoSkip(0)
				}
				h.render(r2Slot("r", src.Route(
					RouteMatch(r2Pred(h.rec, "p", r2nIn(1, 100)),
						RouteMatch(r2nLt(50), r2Comp[int](h.rec, "A")),
						RouteMatch(r2nGe(50), r2Comp[int](h.rec, "B")),
					),
					r2Comp[int](h.rec, "D"),
				)))
				h.expect("r", "D")
				all := []string{"0"}
				for _, seq := range seqs {
					var chs []<-chan error
					if held {
						h.client.hold()
						for _, v := range seq {
							chs = append(chs, src.Update(h.page, v))
						}
						h.waitQueued(1)
						h.client.release()
					} else {
						h.settle(func(ctx context.Context) {
							for _, v := range seq {
								chs = append(chs, src.Update(ctx, v))
							}
						})
					}
					var last []error
					for _, ch := range chs {
						last = drainOp(t, ch)
					}
					if !r2SameErrs(last, []error{nil}) {
						t.Fatalf("seq %v: last update returned %v", seq, last)
					}
					want := final(seq[len(seq)-1])
					h.eventually("r", want)
					n := len(h.client.actions())
					time.Sleep(20 * time.Millisecond)
					if got := len(h.client.actions()); got != n {
						t.Fatalf("seq %v: %d late actions after settling on %q", seq, got-n, want)
					}
					h.expect("r", want)
					h.oneDoor("r")
					for _, v := range seq {
						all = append(all, strconv.Itoa(v))
					}
				}
				got := h.rec.vals("pred:p=")
				var seen []string
				state := ""
				for i := 0; i < len(got); {
					v := got[i]
					n, _ := strconv.Atoi(v)
					c := 1
					if (state == "A" || state == "B") && final(n) != state {
						c = 2
					}
					if i+c > len(got) || got[i+c-1] != v {
						t.Fatalf("predicate log %v: value %s at %d should run the predicate %d time(s) from %q", got, v, i, c, state)
					}
					seen = append(seen, v)
					state = final(n)
					i += c
				}
				if noSkip && !slices.Equal(seen, all) {
					t.Fatalf("NoSkip predicate log %v saw %v, want every value %v", got, seen, all)
				}
				i := 0
				for _, v := range seen {
					for i < len(all) && all[i] != v {
						i++
					}
					if i == len(all) {
						t.Fatalf("predicate log %v saw %v, not in update order %v", got, seen, all)
					}
					i++
				}
			})
		}
	}
}

// Old: TestRouteSourceAndDerivedBeamEntrypoints (derivedBeam and derivedSource
// entries), to depth: beam containers nest under RouteBeam and behave like
// their source counterparts. When the inner value route ends on "100", the
// re-selection from scratch picks the earlier HUNDRED, not the inner LAST. On
// "101" the length-3 match falls through to the inner LAST. A model route
// whose children do not match the decoded model falls through to NF.
func TestRoute2NestedPortBeamEntries(t *testing.T) {
	lastDigit := func(b Beam[string]) Beam[int] {
		return DeriveBeam(b, func(s string) int { return int(s[len(s)-1] - '0') })
	}
	tree := func(r *r2Rec, box *r2Box[Beam[int]]) []BeamRoute[string] {
		return []BeamRoute[string]{
			RouteBeamMatch(func(s string) bool { return len(s) == 3 }, RouteBeamValue("100", r2Comp[string](r, "HUNDRED"))),
			RouteBeamMatch(func(s string) bool { return len(s) >= 2 },
				RouteBeamDerive(r2DeriveBeam(r, "last", lastDigit),
					RouteBeamValue(7, r2Bind[int](r, "SEVEN")),
					r2Beam[int](r, "LAST", box),
				),
			),
			r2Comp[string](r, "X"),
		}
	}
	atoi := func(a int, s string) int {
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
		return a
	}
	for _, e := range []struct {
		name  string
		mount func(src Source[int], rs []BeamRoute[string]) gox.Elem
	}{
		{"derivedBeam", func(src Source[int], rs []BeamRoute[string]) gox.Elem {
			return DeriveBeam(src, strconv.Itoa).RouteBeam(rs...)
		}},
		{"derivedSource", func(src Source[int], rs []BeamRoute[string]) gox.Elem {
			return DeriveSource(src, strconv.Itoa, atoi).RouteBeam(rs...)
		}},
	} {
		t.Run(e.name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(5)
			var box r2Box[Beam[int]]
			h.render(r2Slot("b", e.mount(src, tree(h.rec, &box))))
			r2Run(h, "b", src, "X",
				r2Step[int]{set: 17, want: "SEVEN=7"},
				r2Step[int]{set: 27, want: "SEVEN=7", quiet: true},
				r2Step[int]{set: 23, want: "LAST"},
				r2Step[int]{set: 24, want: "LAST", quiet: true},
				r2Step[int]{set: 37, want: "SEVEN=7"},
				r2Step[int]{set: 100, want: "HUNDRED"},
				r2Step[int]{set: 5, want: "X"},
				r2Step[int]{set: 100, want: "HUNDRED"},
				r2Step[int]{set: 101, want: "LAST"},
				r2Step[int]{set: 102, want: "LAST", quiet: true},
				r2Step[int]{set: 55, want: "LAST", quiet: true, counts: map[string]int{
					"render:LAST": 2, "render:SEVEN": 2, "render:HUNDRED": 2, "render:X": 2, "derive:last": 1,
				}},
			)
			if got := box.last().Get(); box.len() != 2 || got != 5 {
				t.Fatalf("last beam holds %d (%d captured)", got, box.len())
			}
		})
	}
	strip := func(l Location) Location {
		l = l.Clone()
		if len(l.Segments) != 0 {
			l.Segments = l.Segments[1:]
		}
		return l
	}
	for _, e := range []struct {
		name  string
		mount func(loc Source[Location], rs ...BeamRoute[Location]) gox.Elem
	}{
		{"model/derivedBeam", func(loc Source[Location], rs ...BeamRoute[Location]) gox.Elem {
			return DeriveBeamEqual(loc, strip, path.EqualLocation).RouteBeam(rs...)
		}},
		{"model/derivedSource", func(loc Source[Location], rs ...BeamRoute[Location]) gox.Elem {
			return r2Strip("x")(loc).RouteBeam(rs...)
		}},
	} {
		t.Run(e.name, func(t *testing.T) {
			h := newR2Harness(t)
			loc := r2LocSource("/x/item/5")
			h.render(r2Slot("m", e.mount(loc,
				RouteBeamModel[r2PathItem](RouteBeamMatch(func(p r2PathItem) bool { return p.ID > 3 }, r2Bind[r2PathItem](h.rec, "ITEM"))),
				r2Comp[Location](h.rec, "NF"),
			)))
			r2Run(h, "m", loc, "ITEM={true 5 }",
				r2Step[Location]{set: r2Loc("/x/item/2"), want: "NF"},
				r2Step[Location]{set: r2Loc("/x/item/3?tab=a"), want: "NF", quiet: true},
				r2Step[Location]{set: r2Loc("/x/zz"), want: "NF", quiet: true},
				r2Step[Location]{set: r2Loc("/x/item/9"), want: "ITEM={true 9 }"},
				r2Step[Location]{set: r2Loc("/x/item/9?tab=t"), want: "ITEM={true 9 t}"},
				r2Step[Location]{set: r2Loc("/y/item/9?tab=t"), want: "ITEM={true 9 t}", quiet: true, counts: map[string]int{
					"render:ITEM": 3, "render:NF": 1,
				}},
			)
		})
	}
}

// Old: BeamRouteStateFragment (RouteDefaultBeam inside Route), to depth: beam
// leaves and containers under source containers get the beam or source of
// their own level.
func TestRoute2NestedPortBeamInSource(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		h := newR2Harness(t)
		src := NewSource(0)
		var boxA, boxD r2Box[Beam[int]]
		var boxS r2Box[Source[int]]
		h.render(r2Slot("i", src.Route(
			RouteMatch(r2nGt(0),
				RouteDerive(r2Derive(h.rec, "tens", r2nTens),
					RouteBeamMatch(r2nGe(5), r2Beam[int](h.rec, "A", &boxA)),
					RouteMatch(func(t int) bool { return t == 3 }, r2Source[int](h.rec, "S", &boxS)),
					r2Bind[int](h.rec, "B"),
				),
			),
			r2Beam[int](h.rec, "D", &boxD),
		)))
		if boxD.len() != 1 || boxD.last().Get() != 0 {
			t.Fatalf("default beam not captured")
		}
		r2Run(h, "i", src, "D",
			r2Step[int]{set: 12, want: "B=1"},
			r2Step[int]{set: 55, want: "A"},
		)
		if got := boxA.last().Get(); got != 5 {
			t.Fatalf("A beam holds %d, want the tens 5", got)
		}
		r2Run(h, "i", src, "A", r2Step[int]{set: 67, want: "A", quiet: true})
		if got := boxA.last().Get(); got != 6 || boxA.len() != 1 {
			t.Fatalf("A beam holds %d (%d captured), want 6 from one render", got, boxA.len())
		}
		r2Run(h, "i", src, "A", r2Step[int]{set: 34, want: "S"})
		if got := boxS.last().Get(); got != 3 {
			t.Fatalf("S source holds %d, want the tens 3", got)
		}
		if ret := r2Set(h, boxS.last(), 4); !r2SameErrs(ret, []error{nil}) {
			t.Fatalf("write returned %v", ret)
		}
		if got := src.Get(); got != 44 {
			t.Fatalf("tens write mapped to %d, want 44", got)
		}
		h.expect("i", "B=4")
		h.oneDoor("i")
		r2Run(h, "i", src, "B=4", r2Step[int]{set: 0, want: "D"})
		if boxD.len() != 2 || boxD.last().Get() != 0 {
			t.Fatalf("default beam: %d captured", boxD.len())
		}
	})
	t.Run("model", func(t *testing.T) {
		h := newR2Harness(t)
		loc := r2LocSource("/item/5?tab=x")
		var boxM r2Box[Beam[r2PathItem]]
		var boxT r2Box[Beam[string]]
		var boxMS r2Box[Source[r2PathItem]]
		var boxL r2Box[Beam[Location]]
		h.render(r2Slot("m", loc.Route(
			RouteModel[r2PathItem](
				RouteBeamMatch(func(p r2PathItem) bool { return p.ID > 100 }, r2Beam[r2PathItem](h.rec, "MBIG", &boxM)),
				RouteBeamDerive(r2DeriveBeam(h.rec, "tab", func(b Beam[r2PathItem]) Beam[string] {
					return DeriveBeam(b, func(p r2PathItem) string { return p.Tab })
				}), RouteBeamValue("x", r2Beam[string](h.rec, "TX", &boxT))),
				r2Source[r2PathItem](h.rec, "MS", &boxMS),
			),
			r2Beam[Location](h.rec, "LD", &boxL),
		)))
		if got := boxT.last().Get(); got != "x" {
			t.Fatalf("tab beam holds %q", got)
		}
		r2Run(h, "m", loc, "TX",
			r2Step[Location]{set: r2Loc("/item/6?tab=x"), want: "TX", quiet: true},
			r2Step[Location]{set: r2Loc("/item/6?tab=y"), want: "MS"},
		)
		if got := boxMS.last().Get(); got != (r2PathItem{Item: true, ID: 6, Tab: "y"}) {
			t.Fatalf("model source holds %+v", got)
		}
		if ret := r2Set(h, boxMS.last(), r2PathItem{Item: true, ID: 150, Tab: "y"}); !r2SameErrs(ret, []error{nil}) {
			t.Fatalf("model write returned %v", ret)
		}
		if got := loc.Get().String(); got != "/item/150?tab=y" {
			t.Fatalf("model write encoded %s", got)
		}
		h.expect("m", "MBIG")
		if got := boxM.last().Get(); got.ID != 150 {
			t.Fatalf("model beam holds %+v", got)
		}
		h.oneDoor("m")
		r2Run(h, "m", loc, "MBIG", r2Step[Location]{set: r2Loc("/zz"), want: "LD"})
		if got := boxL.last().Get().String(); got != "/zz" {
			t.Fatalf("location beam holds %s", got)
		}
		h.counts(map[string]int{"derive:tab": 1, "render:TX": 1, "render:MS": 1, "render:MBIG": 1, "render:LD": 1})
	})
}

// A decoded model whose nested routes match nothing makes the model route fall
// through to the later 404.
func TestRoute2NestedPortModelFallThrough404(t *testing.T) {
	h := r2nRouterHarness(t, "/item/0")
	h.strict = false
	h.render(
		r2Slot("new", Route(
			RouteModel[r2PathItem](RouteMatch(func(p r2PathItem) bool { return p.ID > 0 }, r2Source[r2PathItem](h.rec, "X", nil))),
			r2Comp[Location](h.rec, "NF"),
		)),
	)
	router := Router(h.page)
	for i, st := range []struct {
		uri, nw, old string
		quiet        bool
	}{
		{"/item/0", "NF", "OLD", false},
		{"/item/0?tab=a", "NF", "OLD", true},
		{"/item/3", "X", "OLD", false},
		{"/item/0", "NF", "OLD", false},
		{"/zzz", "NF", "ONF", true},
		{"/item/0", "NF", "OLD", true},
	} {
		if i > 0 {
			do := func() { r2Set(h, router, r2Loc(st.uri)) }
			if st.quiet {
				h.quiet("new", do)
			} else {
				do()
			}
		}
		h.expect("new", st.nw)
		if n := h.doorCount("new"); n != 1 || h.door("new") != h.doors0["new"] {
			t.Fatalf("step %d: new router left its door", i)
		}
	}
	h.counts(map[string]int{"render:X": 1, "clean:X": 1, "render:NF": 2, "clean:NF": 1})
}

// Value routes and binds at depth use their own level's comparator: a nil
// comparator never matches a value and always re-renders a bind; a custom one
// decides both.
func TestRoute2NestedLevelComparators(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(0)
	nilTens := func(s Source[int]) Source[int] {
		return DeriveSourceEqual(s, func(v int) int { return v / 10 }, func(a, b int) int { return b*10 + a%10 }, nil)
	}
	parity := func(b Beam[int]) Beam[int] {
		return DeriveBeamEqual(b, r2ID, func(x, y int) bool { return x%2 == y%2 })
	}
	raw := NewSourceEqual(0, nil)
	h.render(
		r2Slot("c", src.Route(
			RouteMatch(r2nLt(100),
				RouteDerive(nilTens,
					RouteValue(1, r2Comp[int](h.rec, "NILV")),
					RouteMatch(r2nGe(5), RouteBeamDerive(parity,
						RouteBeamValue(6, r2Comp[int](h.rec, "EVENV")),
						r2Bind[int](h.rec, "PB"),
					)),
					r2Bind[int](h.rec, "NB"),
				),
			),
			r2Comp[int](h.rec, "D"),
		)),
		r2Slot("raw", raw.Route(
			RouteMatch(r2nAlways[int], RouteValue(0, r2Comp[int](h.rec, "ZERO")), r2Bind[int](h.rec, "RB")),
		)),
	)
	r2Run(h, "c", src, "NB=0",
		r2Step[int]{set: 15, want: "NB=1"},
		r2Step[int]{set: 16, want: "NB=1", counts: map[string]int{"render:NB": 3, "render:NILV": 0}},
		r2Step[int]{set: 55, want: "PB=5"},
		r2Step[int]{set: 75, want: "PB=5", quiet: true},
		r2Step[int]{set: 80, want: "EVENV"},
		r2Step[int]{set: 90, want: "PB=9", counts: map[string]int{"render:PB": 2, "render:EVENV": 1, "render:NILV": 0}},
	)
	r2Run(h, "raw", raw, "RB=0",
		r2Step[int]{set: 0, want: "RB=0", ret: []error{nil}},
		r2Step[int]{set: 0, want: "RB=0", counts: map[string]int{"render:RB": 3, "render:ZERO": 0}},
	)
}

func r2sEven(v int) bool { return v%2 == 0 }

func r2sLt(n int) func(int) bool { return func(v int) bool { return v < n } }

func r2sGt(n int) func(int) bool { return func(v int) bool { return v > n } }

func r2sTens(v int) int { return v / 10 }

func r2sTensSet(a, b int) int { return b*10 + a%10 }

func r2sTree(k r2Kit, r *r2Rec) []SourceRoute[int] {
	return []SourceRoute[int]{
		k.match(func(v int) bool { return v >= 100 }, k.derive(r, "tens", r2sTens, r2sTensSet,
			k.value(10, r2Comp[int](r, "TEN")),
			k.match(r2sEven, r2Bind[int](r, "EB")),
		)),
		k.match(r2sEven, r2Comp[int](r, "EVEN")),
		k.match(r2sLt(10), r2Comp[int](r, "SMALL")),
		r2Bind[int](r, "D"),
	}
}

const r2sTreeInitial = "SMALL"

// r2sScript runs r2sTree from 1. At 110 the derive's children match nothing,
// so the >=100 group does not match and EVEN is selected; at 111 the derive,
// inactive, returns its cached no-match without walking.
func r2sScript() []r2Step[int] {
	return []r2Step[int]{
		{set: 2, want: "SMALL", quiet: true, ret: []error{nil}},
		{set: 12, want: "EVEN"},
		{set: 13, want: "D=13"},
		{set: 15, want: "D=15"},
		{set: 100, want: "TEN"},
		{set: 101, want: "TEN", quiet: true},
		{set: 110, want: "EVEN"},
		{set: 111, want: "D=111"},
		{set: 120, want: "EB=12"},
		{set: 140, want: "EB=14"},
		{set: 5, want: "SMALL"},
		{set: 0, want: "SMALL", quiet: true},
		{set: 0, want: "SMALL", quiet: true, ret: []error{}},
	}
}

var r2sScriptCounts = map[string]int{
	"render:SMALL": 2, "clean:SMALL": 1,
	"render:EVEN": 2, "clean:EVEN": 2,
	"render:D": 3, "clean:D": 3,
	"render:TEN": 1, "clean:TEN": 1,
	"render:EB": 2, "clean:EB": 2,
	"derive:tens": 1,
}

func r2sParent(content gox.Elem) *Door {
	p := &Door{}
	p.Inner(context.Background(), content)
	return p
}

func r2sWait(h *r2Harness, want map[string]int) {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var bad []string
		for k, n := range want {
			if got := h.rec.count(k); got != n {
				bad = append(bad, fmt.Sprintf("%s=%d (want %d)", k, got, n))
			}
		}
		if len(bad) == 0 {
			return
		}
		if time.Now().After(deadline) {
			slices.Sort(bad)
			h.t.Fatalf("counts: %s\nlog: %v", strings.Join(bad, ", "), h.rec.all())
		}
		time.Sleep(time.Millisecond)
	}
}

func r2sBalanced(h *r2Harness, labels ...string) {
	h.t.Helper()
	want := map[string]int{}
	for _, l := range labels {
		want["clean:"+l] = h.rec.count("render:" + l)
	}
	r2sWait(h, want)
}

func r2sDrainAll(t *testing.T, chs []<-chan error) [][]error {
	t.Helper()
	out := make([][]error, len(chs))
	for i, ch := range chs {
		out[i] = drainOp(t, ch)
	}
	return out
}

// [S01] Branch content is cleaned exactly once on every deactivation path
// that stays inside one container, and RouteBind cleans on each re-render.
// Instance end cleans whatever is still active.
func TestRoute2StateCleanOnDeactivation(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		for _, workers := range []int{8, 1} {
			h := newR2HarnessN(t, workers)
			r := h.rec
			s1, s2, s3, s4, s5, s6 := NewSource(1), NewSource(1), NewSource(2), NewSource(1), NewSource(7), NewSource(1)
			h.render(
				r2Slot("disp", e.mount(s1, k.match(r2sEven, r2Comp[int](r, "E1")), r2Comp[int](r, "D1"))),
				r2Slot("end", e.mount(s2,
					k.match(r2sLt(10), r2Comp[int](r, "S2")),
					k.match(r2sLt(100), r2Comp[int](r, "M2")),
					r2Comp[int](r, "D2"),
				)),
				r2Slot("fall", e.mount(s3, k.match(r2sGt(0), k.match(r2sEven, r2Comp[int](r, "N3"))), r2Comp[int](r, "D3"))),
				r2Slot("top", e.mount(s4, k.match(r2sGt(0), r2Bind[int](r, "T4")))),
				r2Slot("deep", e.mount(s5,
					k.match(r2sLt(1000), k.derive(r, "dd", nil, nil,
						k.match(r2sGt(0), k.value(7, r2Comp[int](r, "V5")), r2Bind[int](r, "B5")),
					)),
					r2Comp[int](r, "D5"),
				)),
				r2Slot("bind", e.mount(s6, r2Bind[int](r, "B6"))),
			)
			h.counts(map[string]int{"render:D1": 1, "render:S2": 1, "render:N3": 1, "render:T4": 1, "render:V5": 1, "render:B6": 1, "derive:dd": 1})
			// a default displaced by an earlier sibling, and back
			r2Run(h, "disp", s1, "D1",
				r2Step[int]{set: 2, want: "E1", counts: map[string]int{"render:D1": 1, "clean:D1": 1, "render:E1": 1, "clean:E1": 0}},
				r2Step[int]{set: 3, want: "D1", counts: map[string]int{"render:E1": 1, "clean:E1": 1, "render:D1": 2, "clean:D1": 1}},
				r2Step[int]{set: 4, want: "E1", counts: map[string]int{"render:E1": 2, "clean:E1": 1, "render:D1": 2, "clean:D1": 2}},
			)
			// an active non-default ends and the re-selection picks another sibling
			r2Run(h, "end", s2, "S2",
				r2Step[int]{set: 50, want: "M2", counts: map[string]int{"clean:S2": 1, "render:M2": 1, "clean:M2": 0}},
				r2Step[int]{set: 500, want: "D2", counts: map[string]int{"clean:M2": 1, "render:D2": 1, "clean:D2": 0}},
				r2Step[int]{set: 5, want: "S2", counts: map[string]int{"clean:D2": 1, "render:S2": 2, "clean:S2": 1}},
			)
			// a nested end re-selects past the outer match, whose children match
			// nothing, onto the default; the default stays until the earlier
			// match finds a child again
			r2Run(h, "fall", s3, "N3",
				r2Step[int]{set: 3, want: "D3", counts: map[string]int{"clean:N3": 1, "render:D3": 1}},
				r2Step[int]{set: 5, want: "D3", quiet: true, counts: map[string]int{"clean:N3": 1, "render:N3": 1, "clean:D3": 0}},
				r2Step[int]{set: 0, want: "D3", quiet: true, counts: map[string]int{"render:D3": 1, "clean:D3": 0}},
				r2Step[int]{set: 4, want: "N3", counts: map[string]int{"clean:D3": 1, "render:N3": 2, "clean:N3": 1}},
			)
			// a top-level end re-selects nothing and blanks the container
			r2Run(h, "top", s4, "T4=1",
				r2Step[int]{set: 0, want: "", counts: map[string]int{"clean:T4": 1}},
				r2Step[int]{set: -1, want: "", quiet: true, counts: map[string]int{"clean:T4": 1, "render:T4": 1}},
				r2Step[int]{set: 2, want: "T4=2", counts: map[string]int{"clean:T4": 1, "render:T4": 2}},
				r2Step[int]{set: 3, want: "T4=3", counts: map[string]int{"clean:T4": 2, "render:T4": 3}},
			)
			// deep leaves cleaned on every path break, including one whose
			// re-selection falls through the outer match onto the default
			r2Run(h, "deep", s5, "V5",
				r2Step[int]{set: 8, want: "B5=8", counts: map[string]int{"clean:V5": 1}},
				r2Step[int]{set: 9, want: "B5=9", counts: map[string]int{"clean:B5": 1}},
				r2Step[int]{set: 2000, want: "D5", counts: map[string]int{"clean:B5": 2, "render:D5": 1}},
				r2Step[int]{set: 7, want: "V5", counts: map[string]int{"clean:D5": 1, "render:V5": 2}},
				r2Step[int]{set: 0, want: "D5", counts: map[string]int{"clean:V5": 2, "render:D5": 2}},
				r2Step[int]{set: 7, want: "V5", counts: map[string]int{"clean:D5": 2, "clean:V5": 2, "render:V5": 3, "derive:dd": 1}},
			)
			// RouteBind cleans the previous render on each re-render
			r2Run(h, "bind", s6, "B6=1",
				r2Step[int]{set: 2, want: "B6=2", counts: map[string]int{"clean:B6": 1}},
				r2Step[int]{set: 3, want: "B6=3", counts: map[string]int{"clean:B6": 2}},
				r2Step[int]{set: 3, want: "B6=3", quiet: true, ret: []error{}, counts: map[string]int{"clean:B6": 2, "render:B6": 3}},
				r2Step[int]{set: 4, want: "B6=4", counts: map[string]int{"clean:B6": 3, "render:B6": 4}},
			)
			labels := []string{"E1", "D1", "S2", "M2", "D2", "N3", "D3", "T4", "V5", "B5", "D5", "B6"}
			h.inst.Kill()
			r2sBalanced(h, labels...)
		}
	})
}

// [S02] Clean releases RouteBind's stored value on every deactivation path,
// so the old value can be collected.
func TestRoute2StateBindValueReleased(t *testing.T) {
	kind := func(n int) func(*r2Big) bool {
		return func(b *r2Big) bool { return b.kind < n }
	}
	pass := func(s Source[*r2Big]) Source[*r2Big] { return s }
	cases := []struct {
		name   string
		routes func(r *r2Rec) []SourceRoute[*r2Big]
		after  string
	}{
		{"match ends", func(r *r2Rec) []SourceRoute[*r2Big] {
			return []SourceRoute[*r2Big]{RouteMatch(kind(10), r2Bind[*r2Big](r, "B")), r2Comp[*r2Big](r, "D")}
		}, "D"},
		{"default displaced", func(r *r2Rec) []SourceRoute[*r2Big] {
			return []SourceRoute[*r2Big]{RouteMatch(func(b *r2Big) bool { return b.kind >= 10 }, r2Comp[*r2Big](r, "X")), r2Bind[*r2Big](r, "B")}
		}, "X"},
		{"inner end falls through to default", func(r *r2Rec) []SourceRoute[*r2Big] {
			return []SourceRoute[*r2Big]{RouteMatch(kind(100), RouteMatch(kind(10), r2Bind[*r2Big](r, "B"))), r2Comp[*r2Big](r, "D")}
		}, "D"},
		{"top keepEnd", func(r *r2Rec) []SourceRoute[*r2Big] {
			return []SourceRoute[*r2Big]{RouteMatch(kind(10), r2Bind[*r2Big](r, "B"))}
		}, ""},
		{"group default in derive displaced", func(r *r2Rec) []SourceRoute[*r2Big] {
			return []SourceRoute[*r2Big]{RouteDerive(pass,
				RouteMatch(func(b *r2Big) bool { return b.kind >= 10 }, r2Comp[*r2Big](r, "X")),
				r2Bind[*r2Big](r, "B"),
			)}
		}, "X"},
		{"derive default displaced by sibling", func(r *r2Rec) []SourceRoute[*r2Big] {
			return []SourceRoute[*r2Big]{
				RouteMatch(func(b *r2Big) bool { return b.kind >= 10 }, r2Comp[*r2Big](r, "X")),
				RouteDerive(pass, r2Bind[*r2Big](r, "B")),
			}
		}, "X"},
		{"beam match ends", func(r *r2Rec) []SourceRoute[*r2Big] {
			return []SourceRoute[*r2Big]{RouteBeamMatch(kind(10), r2Bind[*r2Big](r, "B")), r2Comp[*r2Big](r, "D")}
		}, "D"},
		{"beam derive default displaced", func(r *r2Rec) []SourceRoute[*r2Big] {
			return []SourceRoute[*r2Big]{RouteBeamDerive(func(b Beam[*r2Big]) Beam[*r2Big] { return b },
				RouteBeamMatch(func(b *r2Big) bool { return b.kind >= 10 }, r2Comp[*r2Big](r, "X")),
				r2Bind[*r2Big](r, "B"),
			)}
		}, "X"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newR2Harness(t)
			b1 := newR2Big(1)
			w := weak.Make(b1)
			src := NewSource(b1)
			b1 = nil
			h.render(r2Slot("a", src.Route(c.routes(h.rec)...)))
			h.expect("a", "B=big1")
			r2Set(h, src, newR2Big(20))
			h.expect("a", c.after)
			h.oneDoor("a")
			if !r2Gone(w) {
				t.Fatalf("the old bind value is still reachable after the switch")
			}
			runtime.KeepAlive(src)
			runtime.KeepAlive(h)
		})
	}
}

// [S03] Re-activation starts fresh: the nested group seeks from scratch
// instead of restoring the old sticky branch.
func TestRoute2StateReactivationFresh(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(1)
		r := h.rec
		h.render(r2Slot("a", e.mount(src,
			k.match(r2sGt(0),
				k.match(r2sEven, r2Comp[int](r, "E")),
				k.match(r2sLt(10), r2Comp[int](r, "S")),
			),
			r2Comp[int](r, "D"),
		)))
		r2Run(h, "a", src, "S",
			r2Step[int]{set: 2, want: "S", quiet: true},
			r2Step[int]{set: 0, want: "D"},
			r2Step[int]{set: 2, want: "E", counts: map[string]int{
				"render:S": 1, "clean:S": 1, "render:D": 1, "clean:D": 1, "render:E": 1, "clean:E": 0,
			}},
		)
	})
}

// [S04] Every derive position runs once per route instance at init, even
// unreachable and deeply nested ones, and never again.
func TestRoute2StateDeriveOncePerInstance(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(0)
		r := h.rec
		h.render(r2Slot("a", e.mount(src,
			k.match(r2sLt(0), k.derive(r, "never", nil, nil, r2Comp[int](r, "NEG"))),
			k.match(r2sLt(1000), k.derive(r, "tens", r2sTens, r2sTensSet,
				k.value(1, r2Comp[int](r, "T1")),
				k.match(func(t int) bool { return t >= 5 }, k.derive(r, "hund", r2sTens, r2sTensSet,
					k.value(0, r2Comp[int](r, "H0")),
					r2Bind[int](r, "HB"),
				)),
				r2Bind[int](r, "TB"),
			)),
			r2Comp[int](r, "BIG"),
		)))
		once := map[string]int{"derive:never": 1, "derive:tens": 1, "derive:hund": 1}
		h.counts(once)
		r2Run(h, "a", src, "TB=0",
			r2Step[int]{set: 15, want: "T1", counts: once},
			r2Step[int]{set: 55, want: "H0", counts: once},
			r2Step[int]{set: 250, want: "HB=2", counts: once},
			r2Step[int]{set: 1500, want: "BIG", counts: once},
			r2Step[int]{set: 60, want: "H0", counts: once},
			r2Step[int]{set: 3, want: "TB=0", counts: once},
			r2Step[int]{set: 60, want: "H0", counts: once},
			r2Step[int]{set: 2000, want: "BIG", counts: once},
			r2Step[int]{set: 17, want: "T1", counts: once},
		)
		h.counts(map[string]int{"render:NEG": 0, "render:H0": 3, "render:BIG": 2, "render:T1": 2})
	})
}

// [S04] A RouteSource under a derive receives the same derived Source on
// every activation.
func TestRoute2StateDeriveSourceIdentity(t *testing.T) {
	for _, e := range r2Entries {
		if e.beam {
			continue
		}
		t.Run(e.name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(15)
			box := &r2Box[Source[int]]{}
			h.render(r2Slot("a", e.mount(src,
				RouteMatch(r2sLt(100), RouteDerive(r2Derive(h.rec, "div", r2Div(10)), r2Source(h.rec, "L", box))),
				r2Comp[int](h.rec, "D"),
			)))
			r2Run(h, "a", src, "L",
				r2Step[int]{set: 200, want: "D"},
				r2Step[int]{set: 25, want: "L"},
				r2Step[int]{set: 300, want: "D"},
				r2Step[int]{set: 35, want: "L", counts: map[string]int{"derive:div": 1, "render:L": 3, "clean:L": 2}},
			)
			if box.len() != 3 {
				t.Fatalf("want 3 captures, got %d", box.len())
			}
			for i := 1; i < box.len(); i++ {
				if box.at(i) != box.at(0) {
					t.Fatalf("activation %d got a different derived source", i)
				}
			}
			if got := box.last().Get(); got != 3 {
				t.Fatalf("derived source reads %d, want 3", got)
			}
		})
	}
}

// [S05] One routes slice in two Route calls with different sources keeps
// independent state per render: active branch, default flag, bind value and
// derive cache.
func TestRoute2StateIsolationSameRoutes(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		a, b := NewSource(1), NewSource(1)
		routes := r2sTree(k, h.rec)
		h.render(r2Slot("a", e.mount(a, routes...)), r2Slot("b", e.mount(b, routes...)))
		if h.doors0["a"] == h.doors0["b"] {
			t.Fatalf("both containers share door %d", h.doors0["a"])
		}
		h.counts(map[string]int{"derive:tens": 2})
		steps := []struct {
			src  Source[int]
			v    int
			a, b string
		}{
			{a, 2, "SMALL", "SMALL"},
			{b, 100, "SMALL", "TEN"},
			{a, 12, "EVEN", "TEN"},
			{b, 110, "EVEN", "EVEN"},
			{a, 13, "D=13", "EVEN"},
			{b, 111, "D=13", "D=111"},
			{a, 100, "TEN", "D=111"},
			{b, 120, "TEN", "EB=12"},
			{a, 110, "EVEN", "EB=12"},
			{b, 5, "EVEN", "SMALL"},
			{a, 5, "SMALL", "SMALL"},
		}
		for i, st := range steps {
			other := "b"
			if st.src == b {
				other = "a"
			}
			h.quietAt(fmt.Sprintf("step %d: ", i), other, func() { r2Set(h, st.src, st.v) })
			h.expectAt(fmt.Sprintf("step %d: ", i), "a", st.a)
			h.expectAt(fmt.Sprintf("step %d: ", i), "b", st.b)
			h.oneDoorAt(fmt.Sprintf("step %d: ", i), "a")
			h.oneDoorAt(fmt.Sprintf("step %d: ", i), "b")
		}
		h.counts(map[string]int{
			"derive:tens": 2, "render:SMALL": 4, "render:TEN": 2, "render:EVEN": 3, "render:D": 2, "render:EB": 1,
		})
	})
}

// [S06] The same routes slice and the same source in two Route calls: both
// containers follow the same script, each on its own door.
func TestRoute2StateIsolationSameSource(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(1)
		routes := r2sTree(k, h.rec)
		h.render(r2Slot("a", e.mount(src, routes...)), r2Slot("b", e.mount(src, routes...)))
		h.counts(map[string]int{"derive:tens": 2})
		da, db := h.doors0["a"], h.doors0["b"]
		if da == db {
			t.Fatalf("both containers share door %d", da)
		}
		for i, st := range r2sScript() {
			na, nb := len(h.doorActions(da)), len(h.doorActions(db))
			r2Set(h, src, st.set)
			at := fmt.Sprintf("step %d (set %d): ", i, st.set)
			h.expectAt(at, "a", st.want)
			r2Same(h, "a", "b")
			h.oneDoorAt(at, "a")
			h.oneDoorAt(at, "b")
			ga, gb := len(h.doorActions(da))-na, len(h.doorActions(db))-nb
			want := 1
			if st.quiet {
				want = 0
			}
			if ga != want || gb != want {
				t.Fatalf("%sdoor actions a=%d b=%d, want %d each", at, ga, gb, want)
			}
		}
		want := map[string]int{"derive:tens": 2}
		for k, n := range r2sScriptCounts {
			if k != "derive:tens" {
				want[k] = 2 * n
			}
		}
		h.counts(want)
	})
}

func r2sXY(k r2Kit, r *r2Rec) []SourceRoute[int] {
	return []SourceRoute[int]{
		k.match(r2Pred(r, "x", r2sEven), k.derive(r, "d", nil, nil, r2Comp[int](r, "X"))),
		k.match(r2Pred(r, "y", r2sLt(10)), r2Comp[int](r, "Y")),
		r2Comp[int](r, "D"),
	}
}

// [S07] A parent reload renders a fresh route instance: new door, one more
// derive call, a fresh seek, the old leaf cleaned and the old instance
// silent.
func TestRoute2StateParentReload(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		h.strict = false
		src := NewSource(1)
		p := r2sParent(r2Slot("in", e.mount(src, r2sXY(k, h.rec)...)))
		h.render(r2Slot("outer", mountDoor(p)))
		r2Run(h, "in", src, "Y", r2Step[int]{set: 2, want: "Y", quiet: true})
		d0 := h.door("in")
		if d0 != h.doors0["in"] {
			t.Fatalf("route door moved before reload")
		}
		n0 := len(h.doorActions(d0))
		h.counts(map[string]int{"derive:d": 1, "render:Y": 1, "clean:Y": 0})
		h.settle(func(ctx context.Context) { p.Reload(ctx) })
		h.expect("in", "X")
		d1 := h.door("in")
		if d1 == d0 {
			t.Fatalf("reload kept route door %d", d0)
		}
		if h.doorCount("in") != 1 {
			t.Fatalf("want one door in the reloaded slot, dom:\n%s", h.dom())
		}
		h.counts(map[string]int{"derive:d": 2, "render:Y": 1, "clean:Y": 1, "render:X": 1})
		ys := len(h.rec.seq("pred:y="))
		r2Run(h, "in", src, "X",
			r2Step[int]{set: 4, want: "X", quiet: true},
			r2Step[int]{set: 6, want: "X", quiet: true},
		)
		if got := len(h.rec.seq("pred:y=")); got != ys {
			t.Fatalf("the old instance still evaluates predicates: %v", h.rec.seq("pred:"))
		}
		r2Run(h, "in", src, "X", r2Step[int]{set: 3, want: "Y"})
		if got := len(h.doorActions(d0)); got != n0 {
			t.Fatalf("old route door %d got %d new actions", d0, got-n0)
		}
		if acts := h.doorActions(d1); len(acts) != 1 || acts[0].name != "door_replace" {
			t.Fatalf("new route door: want one door_replace, got %+v", acts)
		}
		// 3 twice: the ended X branch is re-tried by the re-selection
		if got, want := h.rec.vals("pred:x="), []string{"1", "2", "4", "6", "3", "3"}; !slices.Equal(got, want) {
			t.Fatalf("pred x log %v, want %v", got, want)
		}
		h.counts(map[string]int{"derive:d": 2, "clean:X": 1, "render:Y": 2})
	})
}

// [S08] The parent replacing its content drops the route: its leaf is
// cleaned and later updates reach nothing.
func TestRoute2StateParentReplace(t *testing.T) {
	ops := []struct {
		name string
		op   func(p *Door, ctx context.Context)
	}{
		{"inner", func(p *Door, ctx context.Context) { p.Inner(ctx, textElem("OTHER")) }},
		{"outer", func(p *Door, ctx context.Context) { p.Outer(ctx, tagElem("section", "other", textElem("OTHER"))) }},
	}
	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
				h := newR2Harness(t)
				h.strict = false
				src := NewSource(1)
				p := r2sParent(r2Slot("in", e.mount(src, r2sXY(k, h.rec)...)))
				h.render(r2Slot("outer", mountDoor(p)))
				r2Run(h, "in", src, "Y", r2Step[int]{set: 2, want: "Y", quiet: true})
				h.settle(func(ctx context.Context) { op.op(p, ctx) })
				if _, ok := h.slot("in"); ok {
					t.Fatalf("route slot survived the parent update, dom:\n%s", h.dom())
				}
				h.expect("outer", "OTHER")
				h.counts(map[string]int{"render:Y": 1, "clean:Y": 1})
				preds, acts := len(h.rec.seq("pred:")), len(h.client.actions())
				for _, v := range []int{4, 3, 12, 0} {
					r2Set(h, src, v)
				}
				if got := h.rec.seq("pred:"); len(got) != preds {
					t.Fatalf("dropped routes still evaluate predicates: %v", got[preds:])
				}
				if got := h.client.actions(); len(got) != acts {
					t.Fatalf("dropped routes still produce actions: %+v", got[acts:])
				}
				h.counts(map[string]int{"render:X": 0, "render:Y": 1, "render:D": 0, "derive:d": 1})
			})
		})
	}
}

func r2sNoSubs(h *r2Harness, src Source[int], from int, set func(v int) []error) {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for v := from; ; v++ {
		if ret := set(v); len(ret) == 0 {
			return
		}
		if time.Now().After(deadline) {
			h.t.Fatalf("updates still find subscribers")
		}
		time.Sleep(time.Millisecond)
	}
}

// [S09] Unmount removes the container and stops route callbacks; the source
// eventually has no subscribers.
func TestRoute2StateUnmount(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		h.strict = false
		src := NewSource(1)
		p := r2sParent(r2Slot("in", e.mount(src, r2sXY(k, h.rec)...)))
		h.render(r2Slot("outer", mountDoor(p)))
		r2Run(h, "in", src, "Y", r2Step[int]{set: 2, want: "Y", quiet: true})
		h.settle(func(ctx context.Context) { p.Unmount(ctx) })
		if _, ok := h.slot("in"); ok {
			t.Fatalf("route slot survived unmount, dom:\n%s", h.dom())
		}
		h.counts(map[string]int{"render:Y": 1, "clean:Y": 1})
		preds, acts := len(h.rec.seq("pred:")), len(h.client.actions())
		r2sNoSubs(h, src, 100, func(v int) []error { return r2Set(h, src, v) })
		if got := h.rec.seq("pred:"); len(got) != preds {
			t.Fatalf("unmounted routes still evaluate predicates: %v", got[preds:])
		}
		if got := h.client.actions(); len(got) != acts {
			t.Fatalf("unmounted routes still produce actions: %+v", got[acts:])
		}
	})
}

// [S10] Instance end cleans the active leaf; afterwards updates find no
// subscribers and no predicate runs.
func TestRoute2StateInstanceEnd(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(1)
		h.render(r2Slot("a", e.mount(src, r2sXY(k, h.rec)...)))
		r2Run(h, "a", src, "Y", r2Step[int]{set: 2, want: "Y", quiet: true})
		h.inst.Kill()
		r2sWait(h, map[string]int{"render:Y": 1, "clean:Y": 1})
		preds := len(h.rec.seq("pred:"))
		r2sNoSubs(h, src, 100, func(v int) []error { return drainOp(t, src.Update(context.Background(), v)) })
		for v := 200; v < 205; v++ {
			if ret := drainOp(t, src.Update(context.Background(), v)); len(ret) != 0 {
				t.Fatalf("update after instance end returned %v", ret)
			}
		}
		if got := h.rec.seq("pred:"); len(got) != preds {
			t.Fatalf("predicates ran after instance end: %v", got[preds:])
		}
		h.counts(map[string]int{"render:X": 0, "render:D": 0})
	})
}

// [S11] Concurrent instances share one routes slice; each runs the full
// script independently and inits its own derives exactly once.
func TestRoute2StateParallelInstances(t *testing.T) {
	for _, k := range r2Kits {
		t.Run(k.name, func(t *testing.T) {
			shared := newR2Rec()
			var mu sync.Mutex
			derived := map[any]int{}
			count := func(key any) {
				mu.Lock()
				defer mu.Unlock()
				derived[key]++
			}
			derive := func(get func(int) int, set func(a, b int) int, rs ...SourceRoute[int]) SourceRoute[int] {
				if k.beam {
					return RouteBeamDerive(func(b Beam[int]) Beam[int] {
						count(b)
						return DeriveBeam(b, get)
					}, r2Beams(rs)...)
				}
				return RouteDerive(func(s Source[int]) Source[int] {
					count(s)
					return DeriveSource(s, get, set)
				}, rs...)
			}
			routes := []SourceRoute[int]{
				k.match(func(v int) bool { return v >= 100 }, derive(r2sTens, r2sTensSet,
					k.value(10, r2Comp[int](shared, "TEN")),
					k.match(r2sEven, derive(r2ID, r2IDSet, r2Bind[int](shared, "EB"))),
				)),
				k.match(r2sEven, r2Comp[int](shared, "EVEN")),
				k.match(r2sLt(10), r2Comp[int](shared, "SMALL")),
				r2Bind[int](shared, "D"),
			}
			var entries []r2Entry
			for _, e := range r2Entries {
				if !e.beam || k.beam {
					entries = append(entries, e)
				}
			}
			const n = 6
			r2Parallel(t, n, func(t *testing.T, i int) {
				h := newR2Harness(t)
				src := NewSource(1)
				h.render(r2Slot("a", entries[i%len(entries)].mount(src, routes...)))
				r2Run(h, "a", src, r2sTreeInitial, r2sScript()...)
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
			// every instance ended in its subtest cleanup, so all content is cleaned
			deadline := time.Now().Add(5 * time.Second)
			for _, l := range []string{"SMALL", "EVEN", "D", "TEN", "EB"} {
				if got, want := shared.count("render:"+l), n*r2sScriptCounts["render:"+l]; got != want {
					t.Fatalf("render:%s=%d, want %d", l, got, want)
				}
				for shared.count("clean:"+l) != shared.count("render:"+l) {
					if time.Now().After(deadline) {
						t.Fatalf("clean:%s=%d, want %d", l, shared.count("clean:"+l), shared.count("render:"+l))
					}
					time.Sleep(time.Millisecond)
				}
			}
		})
	}
}

// [S12] Two instances route one global source from different sticky states
// without cross-talk.
func TestRoute2StateSharedSourceInstances(t *testing.T) {
	src := NewSource(1)
	shared := newR2Rec()
	beams := []BeamRoute[int]{
		RouteBeamMatch(r2sEven, r2Comp[int](shared, "X")),
		RouteBeamMatch(r2sLt(10), r2Comp[int](shared, "Y")),
		r2Comp[int](shared, "D"),
	}
	routes := make([]SourceRoute[int], len(beams))
	for i, b := range beams {
		routes[i] = b
	}
	ha, hb := newR2Harness(t), newR2Harness(t)
	ha.render(r2Slot("a", src.Route(routes...)))
	ha.expect("a", "Y")
	if ret := drainOp(t, src.Update(ha.page, 2)); !r2SameErrs(ret, []error{nil}) {
		t.Fatalf("update returned %v", ret)
	}
	ha.eventually("a", "Y")
	hb.render(r2Slot("a", src.RouteBeam(beams...)))
	hb.expect("a", "X")
	steps := []struct {
		v    int
		a, b string
	}{
		{4, "Y", "X"},
		{6, "Y", "X"},
		{3, "Y", "Y"},
		{20, "X", "X"},
		{13, "D", "D"},
		{8, "X", "X"},
		{2, "X", "X"},
	}
	for i, st := range steps {
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
}

// r2sRapidLog replays the RapidUpdates routes over the values the container
// saw and returns the p predicate log the spec prescribes: a keep checks p
// once, and a path break re-selects from scratch, checking p again.
func r2sRapidLog(seen []int) []string {
	var log []string
	p := func(v int) bool {
		log = append(log, fmt.Sprint(v))
		return v%10 != 0
	}
	state := ""
	sel := func(v int) {
		switch {
		case !p(v):
			state = "Z"
		case r2sEven(v):
			state = "E"
		default:
			state = "B"
		}
	}
	for i, v := range seen {
		if i == 0 {
			sel(v)
			continue
		}
		switch state {
		case "Z":
			if p(v) {
				state = "B"
				if r2sEven(v) {
					state = "E"
				}
			}
		case "E":
			if !p(v) || !r2sEven(v) {
				sel(v)
			}
		case "B":
			if !p(v) {
				sel(v)
			} else if r2sEven(v) {
				state = "E"
			}
		}
	}
	return log
}

// [S13] 200 back-to-back updates converge on the final value with one door.
// At page root and on a NoSkip source the container sees every value, inside
// a child door a strictly increasing subsequence; either way the predicate
// log is exactly one keep check per value seen plus a re-selection check on
// each path break.
func TestRoute2StateRapidUpdates(t *testing.T) {
	cases := []struct {
		name   string
		parent bool
		noSkip bool
	}{
		{"root", false, false},
		{"parent", true, false},
		{"parent noskip", true, true},
	}
	for _, c := range cases {
		for _, workers := range []int{8, 1} {
			t.Run(fmt.Sprintf("%s/w%d", c.name, workers), func(t *testing.T) {
				h := newR2HarnessN(t, workers)
				src := NewSource(0)
				if c.noSkip {
					src = NewSourceNoSkip(0)
				}
				r := h.rec
				route := r2Slot("a", src.Route(
					RouteMatch(r2Pred(r, "p", func(v int) bool { return v%10 != 0 }),
						RouteMatch(r2sEven, r2Comp[int](r, "E")),
						r2Bind[int](r, "B"),
					),
					r2Comp[int](r, "Z"),
				))
				if c.parent {
					h.render(r2Slot("outer", mountDoor(r2sParent(route))))
				} else {
					h.render(route)
				}
				h.expect("a", "Z")
				const last = 201
				var chs []<-chan error
				h.settle(func(ctx context.Context) {
					for v := 1; v <= last; v++ {
						chs = append(chs, src.Update(ctx, v))
					}
				})
				rets := r2sDrainAll(t, chs)
				for i, ret := range rets {
					ok := r2SameErrs(ret, []error{nil}) || !c.noSkip && i != len(rets)-1 && r2SameErrs(ret, []error{context.Canceled})
					if !ok {
						t.Fatalf("update %d returned %v", i+1, ret)
					}
				}
				h.expect("a", fmt.Sprintf("B=%d", last))
				h.oneDoor("a")
				got := r.vals("pred:p=")
				var seen []int
				for _, s := range got {
					var v int
					fmt.Sscan(s, &v)
					if len(seen) == 0 || seen[len(seen)-1] != v {
						seen = append(seen, v)
					}
				}
				if c.parent && !c.noSkip {
					for i := 1; i < len(seen); i++ {
						if seen[i] <= seen[i-1] {
							t.Fatalf("seen values not strictly increasing: %v", got)
						}
					}
					if len(seen) == 0 || seen[0] != 0 || seen[len(seen)-1] != last {
						t.Fatalf("pred log %v must start at 0 and end at %d", got, last)
					}
				} else {
					want := make([]int, 0, last+1)
					for v := 0; v <= last; v++ {
						want = append(want, v)
					}
					if !slices.Equal(seen, want) {
						t.Fatalf("container saw %v, want every value 0..%d", seen, last)
					}
				}
				if want := r2sRapidLog(seen); !slices.Equal(got, want) {
					t.Fatalf("pred log %v\nwant %v", got, want)
				}
			})
		}
	}
}

// [S14] Writes from inside a branch: through the captured Source, through a
// RouteDerive lens, an equal write, and writes through stale captures.
func TestRoute2StateWriteFromBranch(t *testing.T) {
	t.Run("source", func(t *testing.T) {
		h := newR2Harness(t)
		src := NewSource(1)
		box := &r2Box[Source[int]]{}
		h.render(r2Slot("a", src.Route(RouteMatch(r2sLt(10), r2Source(h.rec, "W", box)), r2Bind[int](h.rec, "D"))))
		h.expect("a", "W")
		s := box.last()
		if ret := r2Set(h, s, 20); !r2SameErrs(ret, []error{nil}) {
			t.Fatalf("write returned %v", ret)
		}
		h.expect("a", "D=20")
		h.counts(map[string]int{"render:W": 1, "clean:W": 1})
		if src.Get() != 20 {
			t.Fatalf("root is %d", src.Get())
		}
		h.quiet("a", func() {
			if ret := r2Set(h, s, 20); len(ret) != 0 {
				t.Fatalf("equal write returned %v", ret)
			}
		})
		// documented: a stale capture still writes the root
		r2Set(h, s, 3)
		h.expect("a", "W")
		h.oneDoor("a")
		h.counts(map[string]int{"render:W": 2, "clean:W": 1, "clean:D": 1})
	})
	t.Run("derive", func(t *testing.T) {
		h := newR2Harness(t)
		src := NewSource(15)
		box := &r2Box[Source[int]]{}
		h.render(r2Slot("a", src.Route(
			RouteMatch(r2sLt(100), RouteDerive(r2Derive(h.rec, "div", r2Div(10)),
				RouteMatch(r2sLt(3), r2Source(h.rec, "L", box)),
				r2Bind[int](h.rec, "TB"),
			)),
			r2Comp[int](h.rec, "BIG"),
		)))
		h.expect("a", "L")
		s := box.last()
		if ret := r2Set(h, s, 4); !r2SameErrs(ret, []error{nil}) {
			t.Fatalf("derived write returned %v", ret)
		}
		h.expect("a", "TB=4")
		if src.Get() != 45 {
			t.Fatalf("root is %d, want 45", src.Get())
		}
		h.counts(map[string]int{"clean:L": 1})
		h.quiet("a", func() {
			if ret := r2Set(h, s, 4); len(ret) != 0 {
				t.Fatalf("equal derived write returned %v", ret)
			}
		})
		// documented: a stale capture still writes the root through its lens
		r2Set(h, s, 1)
		h.expect("a", "L")
		if src.Get() != 15 {
			t.Fatalf("root is %d, want 15", src.Get())
		}
		r2Set(h, src, 500)
		h.expect("a", "BIG")
		r2Set(h, s, 2)
		h.expect("a", "L")
		if src.Get() != 20 {
			t.Fatalf("root is %d, want 20", src.Get())
		}
		r2Mutate(h, box.last(), func(v int) int { return v + 5 })
		h.expect("a", "TB=7")
		if src.Get() != 70 {
			t.Fatalf("root is %d, want 70", src.Get())
		}
		h.oneDoor("a")
		h.counts(map[string]int{"derive:div": 1, "render:L": 3, "clean:L": 3, "render:BIG": 1, "clean:BIG": 1})
		for i := 1; i < box.len(); i++ {
			if box.at(i) != box.at(0) {
				t.Fatalf("capture %d is a different derived source", i)
			}
		}
	})
}

func r2sRedirect(r *r2Rec, next int, now bool) SourceRoute[int] {
	return RouteSource(func(s Source[int]) gox.Elem {
		return func(cur gox.Cursor) error {
			r.hit("render:redirect")
			if now {
				s.Update(cur.Context(), next)
			} else {
				OnReady(cur.Context(), func(ctx context.Context) { s.Update(ctx, next) })
			}
			return cur.Text("redirect")
		}
	})
}

// [S15] A branch redirecting on activation reaches the target exactly once,
// from an update or from the initial render, with OnReady or straight from
// the render.
func TestRoute2StateRedirect(t *testing.T) {
	for _, now := range []bool{false, true} {
		for _, initial := range []bool{false, true} {
			t.Run(fmt.Sprintf("render=%v/initial=%v", now, initial), func(t *testing.T) {
				h := newR2Harness(t)
				start := 0
				if initial {
					start = 1
				}
				src := NewSource(start)
				h.render(r2Slot("a", src.Route(
					RouteValue(1, r2sRedirect(h.rec, 2, now)),
					RouteValue(2, r2Comp[int](h.rec, "target")),
					r2Comp[int](h.rec, "home"),
				)))
				if !initial {
					h.expect("a", "home")
					r2Set(h, src, 1)
				}
				h.eventually("a", "target")
				time.Sleep(20 * time.Millisecond)
				h.waitIdle()
				h.expect("a", "target")
				h.oneDoor("a")
				h.counts(map[string]int{"render:redirect": 1, "render:target": 1})
				if src.Get() != 2 {
					t.Fatalf("source is %d", src.Get())
				}
			})
		}
	}
}

// [S15] A hook handler registered by branch content writes the source with
// its handler ctx; the branch switches and its hook dies with it.
func TestRoute2StateHandlerWrite(t *testing.T) {
	for _, e := range r2Entries {
		if e.beam {
			continue
		}
		t.Run(e.name, func(t *testing.T) { r2sHandlerWrite(t, e) })
	}
}

func r2sHandlerWrite(t *testing.T, e r2Entry) {
	h := newR2Harness(t)
	h.strict = false
	src := NewSource(1)
	hooks := &r2Box[uint64]{}
	leaf := RouteSource(func(s Source[int]) gox.Elem {
		return func(cur gox.Cursor) error {
			h.rec.hit("render:H")
			c := cur.Context().Value(common.KeyCore).(core.Core)
			hook, ok := c.Door().RegisterHook(func(ctx context.Context, _ http.ResponseWriter, _ *http.Request) bool {
				h.rec.hit("handler")
				s.Update(ctx, s.Get()+20)
				return true
			}, RaceSerial)
			if ok {
				hooks.put(hook.HookID)
			}
			return cur.Text("H")
		}
	})
	h.render(r2Slot("a", e.mount(src, RouteMatch(r2sLt(10), leaf), r2Bind[int](h.rec, "D"))))
	trigger := func(id uint64) bool {
		return h.root.TriggerHook(id, httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), 0)
	}
	h.expect("a", "H")
	if !trigger(hooks.last()) {
		t.Fatal("hook trigger failed")
	}
	h.eventually("a", "D=21")
	h.counts(map[string]int{"handler": 1, "render:H": 1, "clean:D": 0})
	stale := hooks.last()
	r2Set(h, src, 3)
	h.expect("a", "H")
	if trigger(stale) {
		t.Fatal("the deactivated branch's hook still triggers")
	}
	if !trigger(hooks.last()) {
		t.Fatal("hook trigger failed")
	}
	h.eventually("a", "D=23")
	h.counts(map[string]int{"handler": 2, "render:H": 2, "clean:D": 1})
}

// [S16] Concurrent writers while routes switch converge on the final value.
func TestRoute2StateConcurrentWriters(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(1)
		r := h.rec
		h.render(r2Slot("a", e.mount(src,
			k.match(func(v int) bool { return v >= 0 && v%2 == 0 },
				k.match(r2sLt(50), r2Comp[int](r, "EL")),
				k.derive(r, "div", r2sTens, r2sTensSet, r2Bind[int](r, "ET")),
			),
			k.match(func(v int) bool { return v >= 0 && v < 50 }, r2Comp[int](r, "SM")),
			k.match(r2sLt(0), r2Bind[int](r, "NEG")),
			r2Bind[int](r, "D"),
		)))
		h.expect("a", "SM")
		var mu sync.Mutex
		var chs []<-chan error
		var wg sync.WaitGroup
		for g := range 8 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := range 30 {
					ch := src.Update(h.page, (g*37+j*53)%200)
					mu.Lock()
					chs = append(chs, ch)
					mu.Unlock()
				}
			}()
		}
		wg.Wait()
		r2sDrainAll(t, chs)
		if ret := r2Set(h, src, -1); !r2SameErrs(ret, []error{nil}) {
			t.Fatalf("final update returned %v", ret)
		}
		h.expect("a", "NEG=-1")
		h.oneDoor("a")
		h.counts(map[string]int{"derive:div": 1})
	})
}

// [S17] A Bind door inside an active branch takes the value changes; the
// route door stays quiet and switching away drops the inner door.
func TestRoute2StateNestedDoor(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		h.strict = false
		src := NewSource(1)
		r := h.rec
		leaf := RouteBeam(func(b Beam[int]) gox.Elem {
			r.hit("render:leaf")
			return b.Bind(func(v int) gox.Elem {
				return func(cur gox.Cursor) error {
					r.hit("render:inner")
					OnClean(cur.Context(), func() { r.hit("clean:inner") })
					return cur.Text(fmt.Sprint("in=", v))
				}
			})
		})
		h.render(r2Slot("a", e.mount(src, k.match(r2sLt(10), leaf), r2Comp[int](r, "D"))))
		h.expect("a", "in=1")
		ids := r2DoorIDs(h.html("a"))
		if len(ids) != 2 || ids[0] != h.doors0["a"] {
			t.Fatalf("want the route door and one inner door, got %v", ids)
		}
		inner := ids[1]
		r2Run(h, "a", src, "in=1",
			r2Step[int]{set: 2, want: "in=2", quiet: true},
			r2Step[int]{set: 3, want: "in=3", quiet: true},
		)
		acts := h.doorActions(inner)
		if len(acts) != 2 || acts[0].name != "door_replace" || acts[1].name != "door_replace" {
			t.Fatalf("inner door: want two door_replace, got %+v", acts)
		}
		h.counts(map[string]int{"render:leaf": 1, "render:inner": 3, "clean:inner": 2})
		r2Set(h, src, 20)
		h.expect("a", "D")
		if slices.Contains(r2DoorIDs(h.html("a")), inner) {
			t.Fatalf("inner door %d survived the switch", inner)
		}
		h.counts(map[string]int{"render:leaf": 1, "render:inner": 3, "clean:inner": 3})
		quiet := len(h.doorActions(inner))
		r2Set(h, src, 5)
		h.expect("a", "in=5")
		if ids := r2DoorIDs(h.html("a")); len(ids) != 2 || ids[0] != h.doors0["a"] || ids[1] == inner {
			t.Fatalf("want the route door and a fresh inner door, got %v", ids)
		}
		if len(h.doorActions(inner)) != quiet {
			t.Fatalf("dropped inner door got actions")
		}
		if acts := h.doorActions(h.doors0["a"]); len(acts) != 2 {
			t.Fatalf("route door: want 2 door_replace, got %+v", acts)
		}
		h.counts(map[string]int{"render:leaf": 2, "render:inner": 4, "clean:inner": 3, "clean:D": 1})
	})
}

type r2sPredStep struct {
	set    int
	want   string
	preds  string
	quiet  bool
	counts map[string]int
}

func r2sRunPreds(h *r2Harness, slot string, src Source[int], initial, preds string, steps ...r2sPredStep) {
	h.t.Helper()
	if got := strings.Join(h.rec.vals("pred:"), " "); got != preds {
		h.t.Fatalf("initial: preds %q, want %q", got, preds)
	}
	cur := initial
	for i, st := range steps {
		n := len(h.rec.seq("pred:"))
		r2Run(h, slot, src, cur, r2Step[int]{set: st.set, want: st.want, quiet: st.quiet, counts: st.counts})
		if got := strings.Join(h.rec.vals("pred:")[n:], " "); got != st.preds {
			h.t.Fatalf("step %d (set %d): preds %q, want %q", i, st.set, got, st.preds)
		}
		cur = st.want
	}
}

func r2sOdd(v int) bool { return v%2 != 0 }

// r2sSpy wraps a route node and counts its clean calls as "reset:<label>".
type r2sSpy[T any] struct {
	rec   *r2Rec
	label string
	inner SourceRoute[T]
}

func (s r2sSpy[T]) init(src Source[T]) route[T] {
	return &r2sSpyRoute[T]{rec: s.rec, label: s.label, inner: s.inner.init(src)}
}

func (s r2sSpy[T]) initBeam(b Beam[T]) route[T] {
	return &r2sSpyRoute[T]{rec: s.rec, label: s.label, inner: s.inner.(BeamRoute[T]).initBeam(b)}
}

type r2sSpyRoute[T any] struct {
	rec   *r2Rec
	label string
	inner route[T]
}

func (s *r2sSpyRoute[T]) clean() {
	s.rec.hit("reset:" + s.label)
	s.inner.clean()
}

func (s *r2sSpyRoute[T]) seek(ctx context.Context, v T) (gox.Comp, seekResult) {
	return s.inner.seek(ctx, v)
}

func (s *r2sSpyRoute[T]) keep(ctx context.Context, v T) (gox.Comp, keepResult) {
	return s.inner.keep(ctx, v)
}

func r2sSpyOn[T any](r *r2Rec, label string, inner SourceRoute[T]) SourceRoute[T] {
	return r2sSpy[T]{rec: r, label: label, inner: inner}
}

// [S18] A path break ends the whole old path and re-selects from scratch in
// the same update: every node of the old active chain is cleaned exactly
// once and no other node is, an ancestor that still matches re-seeks from its
// first child, the old leaf content is cleaned exactly once, and a
// re-activated bind renders again even for the value it held before the
// break. A match whose children no longer match falls through to the
// default, and a default yield cleans only the displaced default.
func TestRoute2StateCleanOnReselection(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		r := h.rec
		spy := func(label string, inner SourceRoute[int]) SourceRoute[int] { return r2sSpyOn(r, label, inner) }
		src := NewSource(1)
		h.render(r2Slot("a", e.mount(src,
			spy("O", k.match(r2Pred(r, "o", r2sGt(0)),
				spy("MA", k.match(r2Pred(r, "a", func(v int) bool { return v%4 == 0 }), spy("A", r2Bind[int](r, "A")))),
				spy("MB", k.match(r2Pred(r, "b", r2sLt(50)), spy("DV", k.derive(r, "d", nil, nil,
					spy("MC", k.match(r2Pred(r, "c", r2sOdd), spy("C", r2Bind[int](r, "C")))),
				)))),
			)),
			spy("D", r2Bind[int](r, "D")),
		)))
		resets := func(chainA, chainC, d int, more map[string]int) map[string]int {
			m := map[string]int{
				"reset:O": chainA + chainC, "reset:MA": chainA, "reset:A": chainA,
				"reset:MB": chainC, "reset:DV": chainC, "reset:MC": chainC, "reset:C": chainC,
				"reset:D": d,
			}
			maps.Copy(m, more)
			return m
		}
		h.counts(resets(0, 0, 0, nil))
		r2sRunPreds(h, "a", src, "C=1", "o=1 a=1 b=1 c=1",
			r2sPredStep{set: 3, want: "C=3", preds: "o=3 b=3 c=3", counts: resets(0, 0, 0, map[string]int{"render:C": 2, "clean:C": 1})},
			r2sPredStep{set: 4, want: "A=4", preds: "o=4 b=4 c=4 o=4 a=4", counts: resets(0, 1, 0, map[string]int{"render:C": 2, "clean:C": 2, "render:A": 1})},
			r2sPredStep{set: 8, want: "A=8", preds: "o=8 a=8", counts: resets(0, 1, 0, map[string]int{"render:A": 2, "clean:A": 1})},
			r2sPredStep{set: 3, want: "C=3", preds: "o=3 a=3 o=3 a=3 b=3 c=3", counts: resets(1, 1, 0, map[string]int{
				"bind:C=3": 2, "render:C": 3, "clean:C": 2, "clean:A": 2,
			})},
			r2sPredStep{set: 5, want: "C=5", preds: "o=5 b=5 c=5", counts: resets(1, 1, 0, map[string]int{"render:C": 4, "clean:C": 3})},
			r2sPredStep{set: 62, want: "D=62", preds: "o=62 b=62 o=62 a=62 b=62", counts: resets(1, 2, 0, map[string]int{"render:C": 4, "clean:C": 4, "render:D": 1})},
			r2sPredStep{set: 64, want: "A=64", preds: "o=64 a=64", counts: resets(1, 2, 1, map[string]int{"render:A": 3, "clean:A": 2, "clean:D": 1})},
			r2sPredStep{set: 0, want: "D=0", preds: "o=0 o=0", counts: resets(2, 2, 1, map[string]int{"clean:A": 3, "render:D": 2})},
			r2sPredStep{set: 7, want: "C=7", preds: "o=7 a=7 b=7 c=7", counts: resets(2, 2, 2, map[string]int{"render:C": 5, "clean:C": 4, "clean:D": 2})},
		)
		h.counts(map[string]int{"derive:d": 1})
		h.inst.Kill()
		r2sBalanced(h, "A", "C", "D")
	})
}

// [S18] A full re-selection that lands elsewhere, here falling through the
// outer match onto the default, resets every node of the old path: the
// derive's last value and the bind's stored value are both released.
func TestRoute2StateCleanReleasesOldPath(t *testing.T) {
	kind := func(f func(int) bool) func(*r2Big) bool {
		return func(b *r2Big) bool { return f(b.kind) }
	}
	four := func(v int) bool { return v%4 == 0 }
	cases := []struct {
		name  string
		mount func(src Source[*r2Big], r *r2Rec) gox.Elem
	}{
		{"source", func(src Source[*r2Big], r *r2Rec) gox.Elem {
			return src.Route(
				RouteMatch(kind(r2sGt(0)),
					RouteMatch(kind(four), r2Comp[*r2Big](r, "A")),
					RouteMatch(kind(r2sLt(50)), RouteDerive(func(s Source[*r2Big]) Source[*r2Big] { return s },
						RouteMatch(kind(r2sOdd), r2Bind[*r2Big](r, "C")),
					)),
				),
				r2Comp[*r2Big](r, "D"),
			)
		}},
		{"beam", func(src Source[*r2Big], r *r2Rec) gox.Elem {
			return src.RouteBeam(
				RouteBeamMatch(kind(r2sGt(0)),
					RouteBeamMatch(kind(four), r2Comp[*r2Big](r, "A")),
					RouteBeamMatch(kind(r2sLt(50)), RouteBeamDerive(func(b Beam[*r2Big]) Beam[*r2Big] { return b },
						RouteBeamMatch(kind(r2sOdd), r2Bind[*r2Big](r, "C")),
					)),
				),
				r2Comp[*r2Big](r, "D"),
			)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newR2Harness(t)
			b1 := newR2Big(1)
			w := weak.Make(b1)
			src := NewSource(b1)
			b1 = nil
			h.render(r2Slot("a", c.mount(src, h.rec)))
			h.expect("a", "C=big1")
			r2Set(h, src, newR2Big(62))
			h.expect("a", "D")
			h.oneDoor("a")
			h.counts(map[string]int{"render:C": 1, "clean:C": 1, "render:A": 0, "render:D": 1})
			if !r2Gone(w) {
				t.Fatalf("the old path still holds its value after the re-selection")
			}
			runtime.KeepAlive(src)
			runtime.KeepAlive(h)
		})
	}
}

// [S19] A derive cleaned with the active chain walks its subtree again on the
// next seek, even for the value it last matched. A no-match found by a seek
// is cached until the derived value changes: clean only walks the active
// chain, so an inactive derive keeps its cache across a full re-selection and
// across values where its ancestor does not match.
func TestRoute2StateDeriveCacheAfterDeactivation(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		r := h.rec
		src := NewSource(15)
		h.render(r2Slot("a", e.mount(src,
			k.match(r2Pred(r, "g", r2sGt(0)), r2sSpyOn(r, "DV", k.derive(r, "d", r2sTens, r2sTensSet,
				k.match(r2Pred(r, "t", r2sOdd), r2Bind[int](r, "T")),
			))),
			k.match(r2Pred(r, "e", r2sEven), r2Bind[int](r, "E")),
			r2Bind[int](r, "D"),
		)))
		r2sRunPreds(h, "a", src, "T=1", "g=15 t=1",
			r2sPredStep{set: 17, want: "T=1", quiet: true, preds: "g=17", counts: map[string]int{"reset:DV": 0}},
			r2sPredStep{set: -17, want: "D=-17", preds: "g=-17 g=-17 e=-17", counts: map[string]int{"reset:DV": 1}},
			r2sPredStep{set: 13, want: "T=1", preds: "g=13 t=1", counts: map[string]int{"render:T": 2, "clean:T": 1, "reset:DV": 1}},
			r2sPredStep{set: 20, want: "E=20", preds: "g=20 t=2 g=20 t=2 e=20", counts: map[string]int{"reset:DV": 2}},
			r2sPredStep{set: 22, want: "E=22", preds: "e=22"},
			r2sPredStep{set: 23, want: "D=23", preds: "e=23 g=23 e=23", counts: map[string]int{"reset:DV": 2, "clean:E": 2}},
			r2sPredStep{set: 25, want: "D=25", preds: "g=25 e=25"},
			r2sPredStep{set: -1, want: "D=-1", preds: "g=-1 e=-1"},
			r2sPredStep{set: 27, want: "D=27", preds: "g=27 e=27"},
			r2sPredStep{set: 31, want: "T=3", preds: "g=31 t=3", counts: map[string]int{"reset:DV": 2}},
			r2sPredStep{set: 33, want: "T=3", quiet: true, preds: "g=33"},
			r2sPredStep{set: 51, want: "T=5", preds: "g=51 t=5"},
		)
		h.counts(map[string]int{"derive:d": 1, "render:T": 4, "render:E": 2, "render:D": 5, "reset:DV": 2})
	})
}

func r2sNear(a, b int) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= 5
}

// [S20] The derive skip compares with the last value the derive acted on,
// not the last value it saw: under a non-transitive comparator 0, 4, 8, 12,
// 14 re-render at 0, 8 and 14 only, the same as a RouteBind on the derived
// entry, which compares with its stored value.
func TestRoute2StateDeriveSkipLast(t *testing.T) {
	cases := []struct {
		name  string
		mount func(src Source[int], r *r2Rec) (gox.Elem, gox.Elem)
	}{
		{"source", func(src Source[int], r *r2Rec) (gox.Elem, gox.Elem) {
			near := func(s Source[int]) Source[int] { return DeriveSourceEqual(s, r2ID, r2IDSet, r2sNear) }
			return src.Route(RouteDerive(r2Derive(r, "near", near),
					RouteMatch(r2Pred(r, "n", func(int) bool { return true }), r2Bind[int](r, "B")),
				)),
				near(src).Route(r2Bind[int](r, "E"))
		}},
		{"beam", func(src Source[int], r *r2Rec) (gox.Elem, gox.Elem) {
			near := func(b Beam[int]) Beam[int] { return DeriveBeamEqual(b, r2ID, r2sNear) }
			return src.RouteBeam(RouteBeamDerive(r2DeriveBeam(r, "near", near),
					RouteBeamMatch(r2Pred(r, "n", func(int) bool { return true }), r2Bind[int](r, "B")),
				)),
				near(src).RouteBeam(r2Bind[int](r, "E"))
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(0)
			d, b := c.mount(src, h.rec)
			h.render(r2Slot("d", d), r2Slot("e", b))
			steps := []struct {
				set   int
				shown int
				quiet bool
			}{
				{4, 0, true},
				{8, 8, false},
				{12, 8, true},
				{14, 14, false},
			}
			h.expect("d", "B=0")
			h.expect("e", "E=0")
			for i, st := range steps {
				at := fmt.Sprintf("step %d (set %d): ", i, st.set)
				dd, de := h.door("d"), h.door("e")
				nd, ne := len(h.doorActions(dd)), len(h.doorActions(de))
				r2Set(h, src, st.set)
				h.expectAt(at, "d", fmt.Sprintf("B=%d", st.shown))
				h.expectAt(at, "e", fmt.Sprintf("E=%d", st.shown))
				want := 1
				if st.quiet {
					want = 0
				}
				if gd, ge := len(h.doorActions(dd))-nd, len(h.doorActions(de))-ne; gd != want || ge != want {
					t.Fatalf("%sdoor actions d=%d e=%d, want %d each", at, gd, ge, want)
				}
				h.oneDoorAt(at, "d")
				h.oneDoorAt(at, "e")
			}
			if got, want := h.rec.vals("pred:n="), []string{"0", "8", "14"}; !slices.Equal(got, want) {
				t.Fatalf("pred n log %v, want %v", got, want)
			}
			h.counts(map[string]int{
				"derive:near": 1, "render:B": 3, "render:E": 3,
				"bind:B=4": 0, "bind:B=12": 0, "bind:E=4": 0, "bind:E=12": 0,
			})
		})
	}
}

// [S21] Many root re-selections in a row all land on the render-time door.
// With a default: through matches whose children no longer match, a cached
// derive and default yields. Without one: through top-level no-match and a
// leaf rendering nothing, where blank to blank sends nothing and content to
// blank or back is one replace.
func TestRoute2StateOneDoorManyReselections(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		r := h.rec
		sa, sb := NewSource(1), NewSource(1)
		within := func(lo, hi int) func(int) bool { return func(v int) bool { return v >= lo && v < hi } }
		h.render(
			r2Slot("a", e.mount(sa,
				k.match(within(100, 200), k.value(150, r2Comp[int](r, "X"))),
				k.match(within(200, 300), k.value(250, r2Comp[int](r, "Y"))),
				k.value(1, r2Comp[int](r, "A")),
				k.value(2, r2Comp[int](r, "B")),
				k.value(3, r2Comp[int](r, "C")),
				k.derive(r, "d", r2sTens, r2sTensSet, k.value(5, r2Bind[int](r, "T"))),
				r2Bind[int](r, "D"),
			)),
			r2Slot("b", e.mount(sb,
				k.match(within(100, 200), k.value(150, r2Comp[int](r, "X2"))),
				k.value(1, r2Comp[int](r, "A2")),
				k.value(4, RouteComp[int](nil)),
				k.derive(r, "e", r2sTens, r2sTensSet, k.value(5, r2Comp[int](r, "T2"))),
			)),
		)
		withDefault := []r2Step[int]{
			{set: 2, want: "B"},
			{set: 3, want: "C"},
			{set: 150, want: "X"},
			{set: 160, want: "D=160"},
			{set: 250, want: "Y"},
			{set: 260, want: "D=260"},
			{set: 170, want: "D=170"},
			{set: 55, want: "T=5"},
			{set: 57, want: "T=5", quiet: true},
			{set: 65, want: "D=65"},
			{set: 66, want: "D=66"},
			{set: 1, want: "A"},
		}
		blank := []r2Step[int]{
			{set: 150, want: "X2"},
			{set: 160, want: ""},
			{set: 170, want: "", quiet: true},
			{set: 171, want: "", quiet: true},
			{set: 4, want: "", quiet: true},
			{set: 400, want: "", quiet: true},
			{set: 55, want: "T2"},
			{set: 57, want: "T2", quiet: true},
			{set: 4, want: ""},
			{set: 1, want: "A2"},
		}
		const cycles = 8
		var stepsA, stepsB []r2Step[int]
		for range cycles {
			stepsA = append(stepsA, withDefault...)
			stepsB = append(stepsB, blank...)
		}
		r2Run(h, "a", sa, "A", stepsA...)
		r2Run(h, "b", sb, "A2", stepsB...)
		if got := len(h.doorActions(h.doors0["a"])); got != 11*cycles {
			t.Fatalf("route door a: %d replaces, want %d", got, 11*cycles)
		}
		if got := len(h.doorActions(h.doors0["b"])); got != 5*cycles {
			t.Fatalf("route door b: %d replaces, want %d", got, 5*cycles)
		}
		h.counts(map[string]int{
			"derive:d": 1, "render:A": 1 + cycles, "render:B": cycles, "render:X": cycles, "render:Y": cycles,
			"render:T": cycles, "render:D": 5 * cycles, "clean:A": cycles, "clean:D": 5 * cycles,
			"derive:e": 1, "render:A2": 1 + cycles, "clean:A2": cycles, "render:X2": cycles, "clean:X2": cycles,
			"render:T2": cycles, "clean:T2": cycles,
		})
	})
}

func r2sAt(h *r2Harness, uri string) {
	h.t.Helper()
	drainOp(h.t, h.inst.location.Update(context.Background(), r2Loc(uri)))
}

func r2sLocSteps(steps ...string) []r2Step[Location] {
	out := make([]r2Step[Location], 0, len(steps)/2)
	for i := 0; i+1 < len(steps); i += 2 {
		out = append(out, r2Step[Location]{set: r2Loc(steps[i]), want: steps[i+1]})
	}
	return out
}

// Old S01 (router page tests: TestRouterBeamModelTransitionsAndDefault404,
// TestRouterSourceCrossRouteKeepsInstance, TestRouterSourceCombined...,
// TestRouteBindMethods, TestBrowserBackRestoresQueryWithoutReload): location
// changes switch branches in place as replaces of the single route door; the
// instance lives on and content outside the route door is untouched.
func TestRoute2StateOldInPlaceSwitch(t *testing.T) {
	h := newR2Harness(t)
	r := h.rec
	r2sAt(h, "/a")
	h.render(
		r2Slot("before", textElem("BEFORE")),
		r2Slot("r", Route(
			RouteModel[r2PathA](r2Comp[r2PathA](r, "A")),
			RouteModel[r2PathItem](r2Bind[r2PathItem](r, "ITEM")),
			RouteModel[r2PathB](RouteMatch(func(r2PathB) bool { return true }, r2Comp[r2PathB](r, "B"))),
			RouteComp[Location](r2Leaf{r, "404"}),
		)),
		r2Slot("after", textElem("AFTER")),
	)
	router := Router(h.page)
	r2Run(h, "r", router, "A", r2sLocSteps(
		"/b", "B",
		"/item/5?tab=x", "ITEM={true 5 x}",
		"/item/6", "ITEM={true 6 }",
		"/zzz", "404",
		"/item/6", "ITEM={true 6 }",
		"/a", "A",
		"/b?q=1", "B",
	)...)
	h.expect("before", "BEFORE")
	h.expect("after", "AFTER")
	select {
	case <-h.inst.ended:
		t.Fatal("the instance ended")
	default:
	}
	dom, d := h.dom(), h.doors0["r"]
	s0, _, _, e0, ok0 := r2DoorElement(h.initial, d)
	s1, _, _, e1, ok1 := r2DoorElement(dom, d)
	if !ok0 || !ok1 || h.initial[:s0] != dom[:s1] || h.initial[e0:] != dom[e1:] {
		t.Fatalf("content outside the route door changed:\n%s\n---\n%s", h.initial, dom)
	}
	for _, a := range h.client.actions() {
		if r2IsDoorAction(a) && a.id != d {
			t.Fatalf("unexpected door action %+v", a)
		}
	}
	h.counts(map[string]int{"render:A": 2, "render:B": 2, "render:ITEM": 3, "render:404": 1})
}

type r2sState struct {
	Int int
	Str string
}

// Old S02 (internal/test/beam TestRouteSourceAndBeamState): four containers
// on one source keep their own branch and render counts while values flow
// through the branch's own Bind, writes go through the captured lens.
func TestRoute2StateOldContainersOnOneSource(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	r := h.rec
	src := NewSource(r2sState{Int: 1})
	hasStr := func(v r2sState) bool { return v.Str != "" }
	getStr := func(v r2sState) string { return v.Str }
	setStr := func(v r2sState, x string) r2sState { v.Str = x; return v }
	list := func(v r2sState) []string {
		if v.Str == "" {
			return nil
		}
		return []string{v.Str, fmt.Sprint(v.Int)}
	}
	setList := func(v r2sState, x []string) r2sState {
		v.Str = ""
		if len(x) != 0 {
			v.Str = x[0]
		}
		return v
	}
	join := func(v []string) gox.Elem { return textElem(strings.Join(v, ",")) }
	strLens := func(s Source[r2sState]) Source[string] { return DeriveSource(s, getStr, setStr) }
	strBeam := func(b Beam[r2sState]) Beam[string] { return DeriveBeam(b, getStr) }
	listLens := func(s Source[r2sState]) Source[[]string] {
		return DeriveSourceEqual(s, list, setList, slices.Equal[[]string])
	}
	listBeam := func(b Beam[r2sState]) Beam[[]string] { return DeriveBeamEqual(b, list, slices.Equal[[]string]) }
	lenses := &r2Box[Source[string]]{}
	defaultFor := func(label, prefix string) BeamRoute[r2sState] {
		return RouteBeam(func(b Beam[r2sState]) gox.Elem {
			r.hit("render:" + label)
			return b.Bind(func(v r2sState) gox.Elem { return textElem(fmt.Sprintf("%s:%d:%s", prefix, v.Int, v.Str)) })
		})
	}
	h.render(
		r2Slot("src", src.Bind(func(v r2sState) gox.Elem { return textElem(fmt.Sprintf("%d:%s", v.Int, v.Str)) })),
		r2Slot("lens", src.Route(
			RouteMatch(hasStr, RouteDerive(strLens,
				RouteSource(func(l Source[string]) gox.Elem {
					r.hit("render:lens")
					lenses.put(l)
					return l.Bind(func(v string) gox.Elem { return textElem("lens:" + v) })
				}),
			)),
			defaultFor("lensdef", "default"),
		)),
		r2Slot("beam", src.RouteBeam(
			RouteBeamMatch(hasStr, RouteBeamDerive(strBeam,
				RouteBeam(func(b Beam[string]) gox.Elem {
					r.hit("render:beam")
					return b.Bind(func(v string) gox.Elem { return textElem("beam:" + v) })
				}),
			)),
			defaultFor("beamdef", "beam-default"),
		)),
		r2Slot("lenslist", src.Route(
			RouteMatch(hasStr, RouteDerive(listLens,
				RouteSource(func(l Source[[]string]) gox.Elem {
					r.hit("render:lenslist")
					return l.Bind(join)
				}),
			)),
		)),
		r2Slot("beamlist", src.RouteBeam(
			RouteBeamMatch(hasStr, RouteBeamDerive(listBeam,
				RouteBeam(func(b Beam[[]string]) gox.Elem {
					r.hit("render:beamlist")
					return b.Bind(join)
				}),
			)),
		)),
	)
	check := func(at string, texts [6]string, counts map[string]int) {
		t.Helper()
		for i, name := range []string{"src", "lens", "beam", "lenslist", "beamlist"} {
			h.expectAt(at, name, texts[i])
		}
		for _, name := range []string{"lens", "beam", "lenslist", "beamlist"} {
			if h.door(name) != h.doors0[name] {
				t.Fatalf("%s: route door of %q moved", at, name)
			}
		}
		h.countsAt(at, counts)
	}
	check("initial: ", [6]string{"1:", "default:1:", "beam-default:1:", "", ""}, map[string]int{
		"render:lensdef": 1, "render:beamdef": 1, "render:lens": 0, "render:beam": 0, "render:lenslist": 0, "render:beamlist": 0,
	})
	active := map[string]int{"render:lensdef": 1, "render:beamdef": 1, "render:lens": 1, "render:beam": 1, "render:lenslist": 1, "render:beamlist": 1}
	r2Set(h, src, r2sState{Int: 2, Str: "doc"})
	check("doc: ", [6]string{"2:doc", "lens:doc", "beam:doc", "doc,2", "doc,2"}, active)
	r2Set(h, src, r2sState{Int: 2, Str: "next"})
	check("next: ", [6]string{"2:next", "lens:next", "beam:next", "next,2", "next,2"}, active)
	r2Set(h, lenses.last(), "child")
	check("lens update: ", [6]string{"2:child", "lens:child", "beam:child", "child,2", "child,2"}, active)
	r2Set(h, lenses.last(), "")
	cleared := map[string]int{"render:lensdef": 2, "render:beamdef": 2, "render:lens": 1, "render:beam": 1, "render:lenslist": 1, "render:beamlist": 1}
	check("lens clear: ", [6]string{"2:", "default:2:", "beam-default:2:", "", ""}, cleared)
	r2Set(h, src, r2sState{Int: 4})
	check("default int: ", [6]string{"4:", "default:4:", "beam-default:4:", "", ""}, cleared)
}

type r2sLoaders struct {
	r *r2Rec
	g *heavyGate
}

func (l r2sLoaders) heavy(label string) gox.Elem {
	return tagElem("b", "heavy-"+label, l.g.elem("heavy "+label))
}

func (l r2sLoaders) skel(label string) gox.Elem {
	return tagElem("i", "skel-"+label, textElem("skel "+label))
}

func (l r2sLoaders) section(label string) gox.Elem {
	return tagElem("section", "wrap-"+label, func(cur gox.Cursor) error {
		if IsDocument(cur.Context()) {
			return l.heavy(label)(cur)
		}
		l.r.hit("deferred " + label)
		DeferredInner(cur.Context(), l.heavy(label))
		return l.skel(label)(cur)
	})
}

func (l r2sLoaders) swap(label string) gox.Elem {
	return func(cur gox.Cursor) error {
		if IsDocument(cur.Context()) {
			return l.heavy(label)(cur)
		}
		l.r.hit("deferred " + label)
		DeferredOuter(cur.Context(), l.heavy(label))
		return l.skel(label)(cur)
	}
}

func (l r2sLoaders) routes() []SourceRoute[string] {
	to := func(label string) func(string) bool { return func(v string) bool { return v == label } }
	return []SourceRoute[string]{
		RouteMatch(to("inner"), RouteComp[string](l.section("inner"))),
		RouteMatch(to("outer"), RouteComp[string](l.swap("outer"))),
		RouteComp[string](tagElem("p", "home", textElem("home"))),
	}
}

func r2sSameDoor(h *r2Harness, name string) {
	h.t.Helper()
	if n := h.doorCount(name); n != 1 {
		h.t.Fatalf("slot %q: want 1 door, got %d, dom:\n%s", name, n, h.dom())
	}
	if h.door(name) != h.doors0[name] {
		h.t.Fatalf("slot %q: route door moved, dom:\n%s", name, h.dom())
	}
}

// Old S03 (internal/test/deferred TestDeferredRouteDocument): the initially
// matched branch renders in the page with IsDocument, so loaders do not
// defer and no door op is issued; later switches are door ops.
func TestRoute2StateOldDocumentInitial(t *testing.T) {
	for _, start := range []string{"inner", "outer"} {
		t.Run(start, func(t *testing.T) {
			h := newR2Harness(t)
			h.strict = false
			l := r2sLoaders{h.rec, newHeavyGate(t)}
			src := NewSource(start)
			h.render(r2Slot("r", src.Route(l.routes()...)))
			h.expect("r", "heavy "+start)
			if strings.Contains(h.initial, "skel") {
				t.Fatalf("page html has a skeleton:\n%s", h.initial)
			}
			if start == "inner" && !strings.HasPrefix(h.html("r"), "<section ") {
				t.Fatalf("want the section adopted as the route container, got %q", h.html("r"))
			}
			if acts := h.client.actions(); len(acts) != 0 {
				t.Fatalf("initial selection issued actions %+v", acts)
			}
			h.counts(map[string]int{"deferred inner": 0, "deferred outer": 0})
			for _, next := range []string{"outer", "inner", "outer"} {
				if next == start {
					continue
				}
				r2Set(h, src, next)
				h.eventually("r", "heavy "+next)
				r2sSameDoor(h, "r")
			}
			if len(h.doorActions(h.doors0["r"])) == 0 {
				t.Fatalf("later switches issued no door ops")
			}
		})
	}
}

// Old S04 (internal/test/deferred TestDeferredRouteInner): switching into a
// section loader shows the skeleton inside the section, then heavy replaces
// it there, on every activation.
func TestRoute2StateOldDeferredInner(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	g := newHeavyGate(t)
	l := r2sLoaders{h.rec, g}
	src := NewSource("home")
	h.render(r2Slot("r", src.Route(l.routes()...)))
	h.expect("r", "home")
	d := h.doors0["r"]
	g.shut()
	r2Set(h, src, "inner")
	g.waitEntered(t)
	h.waitIdle()
	h.expect("r", "skel inner")
	if html := h.html("r"); !strings.HasPrefix(html, "<section ") || !strings.Contains(html, `id="skel-inner"`) {
		t.Fatalf("want the skeleton inside the section, got %q", html)
	}
	g.open()
	h.eventually("r", "heavy inner")
	if html := h.html("r"); !strings.HasPrefix(html, "<section ") || !strings.Contains(html, `id="heavy-inner"`) || strings.Contains(html, "skel") {
		t.Fatalf("want heavy inside the section, got %q", html)
	}
	r2Set(h, src, "home")
	h.expect("r", "home")
	r2Set(h, src, "inner")
	h.eventually("r", "heavy inner")
	if html := h.html("r"); !strings.HasPrefix(html, "<section ") || strings.Contains(html, "skel") {
		t.Fatalf("want heavy inside the section, got %q", html)
	}
	r2sSameDoor(h, "r")
	var names []string
	for _, a := range h.doorActions(d) {
		names = append(names, a.name)
	}
	if want := []string{"door_replace", "door_update", "door_replace", "door_replace", "door_update"}; !slices.Equal(names, want) {
		t.Fatalf("route door actions %v, want %v", names, want)
	}
	h.counts(map[string]int{"deferred inner": 2})
}

// Old S05 (internal/test/deferred TestDeferredRouteOuter): a branch rooted
// at the skeleton swaps the route container for heavy with DeferredOuter,
// and the route door stays live for later switches.
func TestRoute2StateOldDeferredOuter(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	g := newHeavyGate(t)
	l := r2sLoaders{h.rec, g}
	src := NewSource("home")
	h.render(r2Slot("r", src.Route(l.routes()...)))
	d := h.doors0["r"]
	g.shut()
	r2Set(h, src, "outer")
	g.waitEntered(t)
	h.waitIdle()
	h.expect("r", "skel outer")
	if html := h.html("r"); !strings.HasPrefix(html, "<i ") || !strings.Contains(html, fmt.Sprintf(`data-d0r="%d"`, d)) {
		t.Fatalf("want the skeleton as the route container, got %q", html)
	}
	g.open()
	h.eventually("r", "heavy outer")
	if html := h.html("r"); !strings.HasPrefix(html, "<b ") || !strings.Contains(html, fmt.Sprintf(`data-d0r="%d"`, d)) {
		t.Fatalf("want heavy as the live route container, got %q", html)
	}
	r2Set(h, src, "inner")
	h.eventually("r", "heavy inner")
	if html := h.html("r"); !strings.HasPrefix(html, "<section ") {
		t.Fatalf("want the section, got %q", html)
	}
	r2Set(h, src, "outer")
	h.eventually("r", "heavy outer")
	if html := h.html("r"); !strings.HasPrefix(html, "<b ") || strings.Contains(html, "section") || strings.Contains(html, "skel") {
		t.Fatalf("want only heavy, got %q", html)
	}
	r2sSameDoor(h, "r")
	h.counts(map[string]int{"deferred outer": 2, "deferred inner": 1})
}

// Old S06 (TestRouteSourceAndBeamState default re-render, TestRouteBindMethods):
// a re-activated branch renders anew, even for the value it last saw.
func TestRoute2StateOldReactivationRenders(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(1)
		r := h.rec
		h.render(
			r2Slot("bind", e.mount(src, k.match(r2sLt(10), r2Bind[int](r, "B")), r2Comp[int](r, "D"))),
			r2Slot("def", e.mount(src, k.match(r2sGt(10), r2Comp[int](r, "X")), r2Comp[int](r, "DEF"))),
		)
		r2Run(h, "bind", src, "B=1",
			r2Step[int]{set: 20, want: "D", counts: map[string]int{"render:B": 1, "render:D": 1, "render:X": 1, "render:DEF": 1}},
			r2Step[int]{set: 1, want: "B=1", counts: map[string]int{"render:B": 2, "render:DEF": 2}},
			r2Step[int]{set: 1, want: "B=1", quiet: true, ret: []error{}},
			r2Step[int]{set: 2, want: "B=2", counts: map[string]int{"render:B": 3, "render:DEF": 2}},
			r2Step[int]{set: 30, want: "D", counts: map[string]int{"render:D": 2, "render:X": 2}},
			r2Step[int]{set: 2, want: "B=2", counts: map[string]int{"render:B": 4, "render:DEF": 3}},
		)
		h.expect("def", "DEF")
		h.oneDoor("def")
	})
}

// Old S07 (router page tests: routeBro(...) route values reused for every
// page): one routes slice serves two instances and two containers in one
// page with independent branch, bind value, derive cache and derived source.
// An admin path whose model does not decode falls through to the 404.
func TestRoute2StateOldSharedRoutes(t *testing.T) {
	shared := newR2Rec()
	routes := []SourceRoute[Location]{
		RouteModel[r2PathA](r2Comp[r2PathA](shared, "A")),
		RouteModel[r2PathItem](
			RouteMatch(func(m r2PathItem) bool { return m.ID < 10 }, r2Bind[r2PathItem](shared, "SB")),
			r2Comp[r2PathItem](shared, "BIGITEM"),
		),
		RouteMatch(func(l Location) bool { return len(l.Segments) != 0 && l.Segments[0] == "admin" },
			RouteDerive(r2Strip("admin"), RouteModel[r2PathItem](r2Bind[r2PathItem](shared, "ADMIN"))),
		),
		RouteComp[Location](r2Leaf{shared, "404"}),
	}
	h1, h2 := newR2Harness(t), newR2Harness(t)
	r2sAt(h1, "/a")
	r2sAt(h2, "/a")
	h1.render(r2Slot("r", Route(routes...)), r2Slot("twin", Route(routes...)))
	h2.render(r2Slot("r", Route(routes...)))
	steps := []struct {
		h    *r2Harness
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
		{h2, "/item/4", "SB={true 4 }", "SB={true 4 }"},
		{h2, "/item/5", "SB={true 4 }", "SB={true 5 }"},
		{h1, "/item/5", "SB={true 5 }", "SB={true 5 }"},
	}
	for i, st := range steps {
		at := fmt.Sprintf("step %d (%s): ", i, st.uri)
		other := h2
		if st.h == h2 {
			other = h1
		}
		other.quietAt(at, "r", func() { r2Set(st.h, Router(st.h.page), r2Loc(st.uri)) })
		h1.expectAt(at, "r", st.a)
		h2.expectAt(at, "r", st.b)
		r2Same(h1, "r", "twin")
		h1.oneDoorAt(at, "r")
		h1.oneDoorAt(at, "twin")
		h2.oneDoorAt(at, "r")
	}
	if got := shared.count("render:SB"); got != 11 {
		t.Fatalf("render:SB=%d: the bind value leaked between containers, log %v", got, shared.all())
	}
}

// Old S08 (beam_route.go deriveRouteSource: derived source rebuilt per
// activation): the RouteDerive derive runs once per Route render.
func TestRoute2StateOldDeriveOnce(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	src := NewSource(1)
	box := &r2Box[Source[int]]{}
	routes := []SourceRoute[int]{
		RouteMatch(r2sLt(10), RouteDerive(r2Derive(h.rec, "c", func(s Source[int]) Source[int] { return DeriveSource(s, r2ID, r2IDSet) }),
			r2Source(h.rec, "R", box),
		)),
		RouteMatch(r2sLt(0), RouteDerive(r2Derive(h.rec, "c", func(s Source[int]) Source[int] { return s }), r2Comp[int](h.rec, "NEVER"))),
		r2Comp[int](h.rec, "D"),
	}
	p := r2sParent(r2Slot("b", src.Route(routes...)))
	h.render(r2Slot("a", src.Route(routes...)), r2Slot("outer", mountDoor(p)))
	h.counts(map[string]int{"derive:c": 4})
	for _, v := range []int{2, 20, 3, 30, 4, 4, 40, 5} {
		r2Set(h, src, v)
	}
	h.expect("a", "R")
	h.expect("b", "R")
	h.counts(map[string]int{"derive:c": 4, "render:R": 8, "render:NEVER": 0})
	h.settle(func(ctx context.Context) { p.Reload(ctx) })
	h.expect("b", "R")
	h.counts(map[string]int{"derive:c": 6})
}

// Old S09 (lens-route re-activation in TestRouteSourceAndBeamState): after
// re-activation the leaf gets the same derived Source, writes through it map
// back, and the derived comparator still governs RouteBind.
func TestRoute2StateOldDerivedSourceKept(t *testing.T) {
	h := newR2Harness(t)
	r := h.rec
	a, b := NewSource(15), NewSource(1)
	box := &r2Box[Source[int]]{}
	parity := func(s Source[int]) Source[int] {
		return DeriveSourceEqual(s, r2ID, r2IDSet, func(x, y int) bool { return x%2 == y%2 })
	}
	h.render(
		r2Slot("a", a.Route(
			RouteMatch(r2sLt(100), RouteDerive(r2Derive(r, "div", r2Div(10)), r2Source(r, "L", box))),
			r2Comp[int](r, "D"),
		)),
		r2Slot("b", b.Route(
			RouteMatch(r2sLt(100), RouteDerive(r2Derive(r, "par", parity), r2Bind[int](r, "PB"))),
			r2Comp[int](r, "BD"),
		)),
	)
	r2Run(h, "a", a, "L",
		r2Step[int]{set: 200, want: "D"},
		r2Step[int]{set: 25, want: "L"},
	)
	if box.len() != 2 || box.at(0) != box.at(1) {
		t.Fatalf("re-activation got a new derived source")
	}
	h.quiet("a", func() { r2Set(h, box.last(), 7) })
	if a.Get() != 75 {
		t.Fatalf("root is %d, want 75", a.Get())
	}
	r2Run(h, "b", b, "PB=1",
		r2Step[int]{set: 3, want: "PB=1", quiet: true},
		r2Step[int]{set: 200, want: "BD"},
		r2Step[int]{set: 5, want: "PB=5"},
		r2Step[int]{set: 7, want: "PB=5", quiet: true},
		r2Step[int]{set: 6, want: "PB=6"},
	)
	h.counts(map[string]int{"derive:div": 1, "derive:par": 1, "render:PB": 3})
}

// Old S10 (internal/test/beam TestRouteSourceAndDerivedBeamEntrypoints): a
// root change that leaves a derived entry equal triggers no route callback
// for containers on that entry, while the parity entry switches.
func TestRoute2StateOldDerivedEntryEqual(t *testing.T) {
	h := newR2Harness(t)
	h.strict = false
	r := h.rec
	source := NewSource(r2sState{Int: 1, Str: "a"})
	str := DeriveSource(source, func(s r2sState) string { return s.Str }, func(s r2sState, v string) r2sState { s.Str = v; return s })
	parity := DeriveBeam(source, func(s r2sState) string {
		if s.Int%2 == 0 {
			return "even"
		}
		return "odd"
	})
	nonEmpty := func(v string) bool { return v != "" }
	simple := func(v string) bool { return strings.HasPrefix(v, "simple") }
	lenses := &r2Box[Source[string]]{}
	simples := &r2Box[Source[string]]{}
	srcLeaf := func(box *r2Box[Source[string]], prefix string) SourceRoute[string] {
		return RouteSource(func(l Source[string]) gox.Elem {
			box.put(l)
			return l.Bind(func(v string) gox.Elem { return textElem(prefix + v) })
		})
	}
	beamLeaf := func(prefix string) BeamRoute[string] {
		return RouteBeam(func(b Beam[string]) gox.Elem {
			return b.Bind(func(v string) gox.Elem { return textElem(prefix + v) })
		})
	}
	h.render(
		r2Slot("src", source.Bind(func(v r2sState) gox.Elem { return textElem(fmt.Sprintf("%d:%s", v.Int, v.Str)) })),
		r2Slot("lens", str.Route(RouteMatch(r2Pred(r, "lens", nonEmpty), srcLeaf(lenses, "lens:")))),
		r2Slot("lensbeam", str.RouteBeam(RouteBeamMatch(r2Pred(r, "lensbeam", nonEmpty), beamLeaf("beam:")))),
		r2Slot("parity", parity.RouteBeam(
			RouteBeamMatch(r2Pred(r, "parity", func(v string) bool { return v == "odd" }), beamLeaf("derived:")),
			beamLeaf("derived-default:"),
		)),
		r2Slot("simple", str.Route(RouteMatch(r2Pred(r, "simple", simple), srcLeaf(simples, "simple-lens:")))),
		r2Slot("simplebeam", str.RouteBeam(RouteBeamMatch(r2Pred(r, "simplebeam", simple), beamLeaf("simple-beam:")))),
	)
	names := []string{"src", "lens", "lensbeam", "parity", "simple", "simplebeam"}
	check := func(at string, want ...string) {
		t.Helper()
		for i, name := range names {
			h.expectAt(at, name, want[i])
		}
	}
	check("initial: ", "1:a", "lens:a", "beam:a", "derived:odd", "", "")
	r2Set(h, lenses.last(), "b")
	check("lens update: ", "1:b", "lens:b", "beam:b", "derived:odd", "", "")
	r2Set(h, str, "simple")
	check("set simple: ", "1:simple", "lens:simple", "beam:simple", "derived:odd", "simple-lens:simple", "simple-beam:simple")
	r2Set(h, simples.last(), "simple-lens")
	check("simple lens update: ", "1:simple-lens", "lens:simple-lens", "beam:simple-lens", "derived:odd", "simple-lens:simple-lens", "simple-beam:simple-lens")
	strSlots := []string{"lens", "lensbeam", "simple", "simplebeam"}
	before := map[string]int{}
	ids := map[uint64]int{}
	for _, name := range strSlots {
		before[name] = len(r.seq("pred:" + name + "="))
		for _, id := range r2DoorIDs(h.html(name)) {
			ids[id] = len(h.doorActions(id))
		}
	}
	r2Set(h, source, r2sState{Int: 2, Str: str.Get()})
	check("even: ", "2:simple-lens", "lens:simple-lens", "beam:simple-lens", "derived-default:even", "simple-lens:simple-lens", "simple-beam:simple-lens")
	for _, name := range strSlots {
		if got := len(r.seq("pred:" + name + "=")); got != before[name] {
			t.Fatalf("%s: route callback ran for an equal derived value: %v", name, r.seq("pred:"+name+"="))
		}
	}
	for id, n := range ids {
		if got := len(h.doorActions(id)); got != n {
			t.Fatalf("door %d on the str entry got %d new actions", id, got-n)
		}
	}
	// even twice: the ended match is re-tried by the re-selection
	if got := r.vals("pred:parity="); !slices.Equal(got, []string{"odd", "even", "even"}) {
		t.Fatalf("parity predicate log %v", got)
	}
}

// Nested levels under a model route resolve on one door.
func TestRoute2StateOldOneDoorNested(t *testing.T) {
	hn := newR2Harness(t)
	r2sAt(hn, "/item/5")
	five := r2PathItem{Item: true, ID: 5}
	small := func(m r2PathItem) bool { return m.ID < 100 }
	hn.render(r2Slot("r", Route(
		RouteModel[r2PathItem](
			RouteMatch(small, RouteValue(five, r2Comp[r2PathItem](hn.rec, "A")), r2Comp[r2PathItem](hn.rec, "B")),
			r2Comp[r2PathItem](hn.rec, "C"),
		),
		RouteComp[Location](r2Leaf{hn.rec, "DEF"}),
	)))
	steps := []struct{ uri, want string }{
		{"/item/6", "B"},
		{"/item/5", "A"},
		{"/item/5?tab=x", "B"},
		{"/item/200", "C"},
		{"/item/7", "B"},
		{"/zzz", "DEF"},
		{"/item/5", "A"},
		{"/item/300", "C"},
	}
	hn.expect("r", "A")
	for i, st := range steps {
		at := fmt.Sprintf("step %d (%s): ", i, st.uri)
		r2Set(hn, Router(hn.page), r2Loc(st.uri))
		hn.expectAt(at, "r", st.want)
		hn.oneDoorAt(at, "r")
	}
}

// Old S13 (router page tests: TestParallelComponentRender): Parallel children
// inside a route branch render with the page, and again on re-activation.
func TestRoute2StateOldParallelBranch(t *testing.T) {
	h := newR2Harness(t)
	r := h.rec
	r2sAt(h, "/a")
	par := func(s Source[r2PathA]) gox.Elem {
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
	}
	h.render(r2Slot("r", Route(
		RouteModel[r2PathA](RouteSource(par)),
		RouteModel[r2PathB](r2Comp[r2PathB](r, "B")),
		RouteComp[Location](r2Leaf{r, "404"}),
	)))
	h.expect("r", "P0P1P2P3")
	h.counts(map[string]int{"par": 4})
	r2Run(h, "r", Router(h.page), "P0P1P2P3", r2sLocSteps(
		"/b", "B",
		"/a", "P0P1P2P3",
		"/zzz", "404",
		"/a?x=1", "P0P1P2P3",
	)...)
	h.counts(map[string]int{"par": 12})
}

type r2mCrossA struct {
	Path bool `path:"/cross-a"`
}

type r2mCrossB struct {
	Path bool `path:"/cross-b"`
}

type r2mQuery struct {
	Path bool    `path:"/q"`
	Tag  *string `query:"tag"`
	Page *int    `query:"page"`
}

type r2mEscaped struct {
	Path bool `path:"/escaped/:Name"`
	Name string
}

type r2mPathC struct {
	C1 bool `path:"/c1"`
	C2 bool `path:"/c2"`
}

type r2mReplaceStep int

type r2mReplace struct {
	Step r2mReplaceStep `/:"rep-a | rep-b | rep-c"`
}

type r2mCustomRoute struct {
	ID  string
	Tab string
}

func (r r2mCustomRoute) Encode() (Location, error) {
	query := url.Values{}
	if r.Tab != "" {
		query.Set("tab", r.Tab)
	}
	return Location{Segments: []string{"custom", r.ID}, Query: query}, nil
}

type r2mState struct {
	Int int
	Str string
}

func r2mIsCustom(l Location) bool {
	return len(l.Segments) == 2 && l.Segments[0] == "custom"
}

func r2mDeriveCustom(l Location) (r2mCustomRoute, bool) {
	if !r2mIsCustom(l) {
		return r2mCustomRoute{}, false
	}
	return r2mCustomRoute{ID: l.Segments[1], Tab: l.Query.Get("tab")}, true
}

func r2mGetCustom(l Location) r2mCustomRoute {
	c, _ := r2mDeriveCustom(l)
	return c
}

func r2mSetCustom(l Location, r r2mCustomRoute) Location {
	next, err := r.Encode()
	if err != nil {
		return l
	}
	return next
}

func r2mEqCustom(a, b r2mCustomRoute) bool { return a == b }

func r2mCustomLens(s Source[Location]) Source[r2mCustomRoute] {
	return DeriveSourceEqual(s, r2mGetCustom, r2mSetCustom, r2mEqCustom)
}

func r2mCustomBeamLens(b Beam[Location]) Beam[r2mCustomRoute] {
	return DeriveBeamEqual(b, r2mGetCustom, r2mEqCustom)
}

func r2mIsRaw(l Location) bool {
	return len(l.Segments) == 1 && l.Segments[0] == "raw"
}

func r2mIsCrossA(l Location) bool {
	return len(l.Segments) == 1 && l.Segments[0] == "cross-a"
}

func r2mIsItem(l Location) bool {
	return len(l.Segments) == 2 && l.Segments[0] == "item"
}

func r2mCustomItemGet(l Location) r2Custom {
	if !r2mIsItem(l) {
		return r2Custom{}
	}
	return r2Custom{ID: l.Segments[1], Tab: l.Query.Get("tab")}
}

func r2mCustomItemLens(s Source[Location]) Source[r2Custom] {
	return DeriveSourceEqual(s, r2mCustomItemGet, func(prev Location, c r2Custom) Location {
		next, err := c.Encode()
		if err != nil {
			return prev
		}
		return next
	}, func(a, b r2Custom) bool { return a == b })
}

func r2mQueryText(m r2mQuery) string {
	tag, page := "", ""
	if m.Tag != nil {
		tag = *m.Tag
	}
	if m.Page != nil {
		page = strconv.Itoa(*m.Page)
	}
	return "query|" + tag + "|" + page
}

func r2mCustomText(name string) func(r2mCustomRoute) string {
	return func(c r2mCustomRoute) string { return name + "|" + c.ID + "|" + c.Tab }
}

func r2mLocID(l Location) Location { return l }

func r2mLocIDSet(_ Location, l Location) Location { return l }

func r2mStripBeam(b Beam[Location]) Beam[Location] {
	return DeriveBeamEqual(b, func(l Location) Location {
		l = l.Clone()
		if len(l.Segments) != 0 {
			l.Segments = l.Segments[1:]
		}
		return l
	}, path.EqualLocation)
}

func r2mBindF[T any](r *r2Rec, label string, f func(T) string) BeamRoute[T] {
	return RouteBind(func(v T) gox.Elem {
		text := f(v)
		r.hit("render:" + label)
		return func(cur gox.Cursor) error {
			OnClean(cur.Context(), func() { r.hit("clean:" + label) })
			return cur.Text(text)
		}
	})
}

func r2mCapture[T any](r *r2Rec, label string, box *r2Box[Source[T]]) func(Source[T]) Source[T] {
	return func(s Source[T]) Source[T] {
		r.hit("derive:" + label)
		box.put(s)
		return s
	}
}

func r2mRouterAt(h *r2Harness, uri string) Source[Location] {
	h.inst.helperInstance.location = beam.NewSource(r2Loc(uri), path.EqualLocation, false)
	return source[Location]{Source: h.inst.helperInstance.location}
}

func r2mRouterElem(f func(l Source[Location]) gox.Elem) gox.Elem {
	return func(cur gox.Cursor) error {
		return cur.Comp(f(Router(cur.Context())))
	}
}

func r2mSP(p string, replace bool) string {
	return fmt.Sprintf("&{%s %t}", p, replace)
}

func r2mPaths(h *r2Harness) []string {
	var out []string
	for _, a := range h.client.actions() {
		if a.name == "set_path" {
			out = append(out, a.body)
		}
	}
	return out
}

func r2mWantPaths(h *r2Harness, want ...string) {
	h.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		got := r2mPaths(h)
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

func r2mGo(s Source[Location], uri string) func(ctx context.Context) <-chan error {
	return func(ctx context.Context) <-chan error { return s.Update(ctx, r2Loc(uri)) }
}

func r2mPut[T any](s Source[T], v T) func(ctx context.Context) <-chan error {
	return func(ctx context.Context) <-chan error { return s.Update(ctx, v) }
}

func r2mPutBox[T any](b *r2Box[Source[T]], v T) func(ctx context.Context) <-chan error {
	return func(ctx context.Context) <-chan error { return b.last().Update(ctx, v) }
}

func r2mReplaceCtx(op func(ctx context.Context) <-chan error) func(ctx context.Context) <-chan error {
	return func(ctx context.Context) <-chan error { return op(HistoryReplaceContext(ctx)) }
}

func r2mLocIs(h *r2Harness, loc Beam[Location], uri string) func(at string) {
	return func(at string) {
		h.t.Helper()
		if got, want := loc.Get(), r2Loc(uri); !path.EqualLocation(got, want) {
			h.t.Fatalf("%slocation %q, want %q", at, got.String(), want.String())
		}
	}
}

type r2mStep struct {
	at     string
	op     func(ctx context.Context) <-chan error
	ret    []error
	want   map[string]string
	quiet  []string
	counts map[string]int
	check  func(at string)
}

func r2mQuietAt(h *r2Harness, at string, slots []string, step func()) {
	h.t.Helper()
	ids := make([]uint64, len(slots))
	ns := make([]int, len(slots))
	for i, s := range slots {
		ids[i] = h.door(s)
		ns[i] = len(h.doorActions(ids[i]))
	}
	step()
	for i, s := range slots {
		if got := h.doorActions(ids[i]); len(got) != ns[i] {
			h.t.Fatalf("%sslot %q door %d: expected no new actions, got %+v", at, s, ids[i], got[ns[i]:])
		}
	}
}

func r2mRun(h *r2Harness, steps ...r2mStep) {
	h.t.Helper()
	for i, st := range steps {
		at := fmt.Sprintf("step %d (%s): ", i, st.at)
		if st.op != nil {
			var ret []error
			r2mQuietAt(h, at, st.quiet, func() {
				var ch <-chan error
				h.settle(func(ctx context.Context) { ch = st.op(ctx) })
				if ch != nil {
					ret = drainOp(h.t, ch)
				}
			})
			if st.ret != nil && !r2SameErrs(ret, st.ret) {
				h.t.Fatalf("%supdate returned %v, want %v", at, ret, st.ret)
			}
		}
		names := slices.Sorted(maps.Keys(st.want))
		for _, name := range names {
			h.expectAt(at, name, st.want[name])
		}
		if st.counts != nil {
			h.countsAt(at, st.counts)
		}
		if st.check != nil {
			st.check(at)
		}
		if h.strict {
			for _, name := range names {
				h.oneDoorAt(at, name)
			}
		}
	}
}

type r2mFam struct {
	name string
	beam bool
}

var r2mFams = []r2mFam{{"source", false}, {"beam", true}}

func r2mModel[M any](f r2mFam, rs ...SourceRoute[M]) SourceRoute[Location] {
	if f.beam {
		return RouteBeamModel[M](r2Beams(rs)...)
	}
	return RouteModel[M](rs...)
}

func r2mValue[T any](f r2mFam, v T, rs ...SourceRoute[T]) SourceRoute[T] {
	if f.beam {
		return RouteBeamValue(v, r2Beams(rs)...)
	}
	return RouteValue(v, rs...)
}

func r2mMatch[T any](f r2mFam, p func(T) bool, rs ...SourceRoute[T]) SourceRoute[T] {
	if f.beam {
		return RouteBeamMatch(p, r2Beams(rs)...)
	}
	return RouteMatch(p, rs...)
}

func r2mFirst(seg string) func(Location) bool {
	return func(l Location) bool { return len(l.Segments) != 0 && l.Segments[0] == seg }
}

func r2mStrip(f r2mFam, r *r2Rec, label, prefix string, rs ...SourceRoute[Location]) SourceRoute[Location] {
	if f.beam {
		return RouteBeamDerive(r2DeriveBeam(r, label, r2mStripBeam), r2Beams(rs)...)
	}
	return RouteDerive(r2Derive(r, label, r2Strip(prefix)), rs...)
}

func r2mItemCustom(f r2mFam, rs ...SourceRoute[r2Custom]) SourceRoute[Location] {
	if f.beam {
		return RouteBeamDerive(func(b Beam[Location]) Beam[r2Custom] {
			return DeriveBeamEqual(b, r2mCustomItemGet, func(a, b r2Custom) bool { return a == b })
		}, r2Beams(rs)...)
	}
	return RouteDerive(r2mCustomItemLens, rs...)
}

func r2mOnce[T any](f r2mFam, r *r2Rec, label string) SourceRoute[T] {
	if f.beam {
		return r2Beam[T](r, label, nil)
	}
	return r2Source[T](r, label, nil)
}

func r2mSame[T any](f r2mFam, r *r2Rec, label string, rs ...SourceRoute[T]) SourceRoute[T] {
	if f.beam {
		return RouteBeamDerive(r2DeriveBeam(r, label, func(b Beam[T]) Beam[T] { return b }), r2Beams(rs)...)
	}
	return RouteDerive(r2Derive(r, label, func(s Source[T]) Source[T] { return s }), rs...)
}

type r2mOrigin struct {
	name   string
	router bool
	make   func(h *r2Harness, uri string) Source[Location]
}

var r2mOrigins = []r2mOrigin{
	{"NewSourceEqual", false, func(_ *r2Harness, uri string) Source[Location] { return r2LocSource(uri) }},
	{"NewSourceEqualNoSkip", false, func(_ *r2Harness, uri string) Source[Location] {
		return NewSourceEqualNoSkip(r2Loc(uri), path.EqualLocation)
	}},
	{"Router", true, r2mRouterAt},
}

type r2mEntry struct {
	name   string
	beam   bool
	router bool
	mount  func(loc Source[Location], rs ...SourceRoute[Location]) gox.Elem
}

var r2mEntries = []r2mEntry{
	{"Route", false, false, func(loc Source[Location], rs ...SourceRoute[Location]) gox.Elem {
		return loc.Route(rs...)
	}},
	{"RouteBeam", true, false, func(loc Source[Location], rs ...SourceRoute[Location]) gox.Elem {
		return loc.RouteBeam(r2Beams(rs)...)
	}},
	{"DeriveSourceEqual.Route", false, false, func(loc Source[Location], rs ...SourceRoute[Location]) gox.Elem {
		return DeriveSourceEqual(loc, r2mLocID, r2mLocIDSet, path.EqualLocation).Route(rs...)
	}},
	{"DeriveSourceEqual.RouteBeam", true, false, func(loc Source[Location], rs ...SourceRoute[Location]) gox.Elem {
		return DeriveSourceEqual(loc, r2mLocID, r2mLocIDSet, path.EqualLocation).RouteBeam(r2Beams(rs)...)
	}},
	{"DeriveBeamEqual.RouteBeam", true, false, func(loc Source[Location], rs ...SourceRoute[Location]) gox.Elem {
		return DeriveBeamEqual(loc, r2mLocID, path.EqualLocation).RouteBeam(r2Beams(rs)...)
	}},
	{"package.Route", false, true, func(_ Source[Location], rs ...SourceRoute[Location]) gox.Elem {
		return Route(rs...)
	}},
	{"Router.RouteBeam", true, true, func(_ Source[Location], rs ...SourceRoute[Location]) gox.Elem {
		return r2mRouterElem(func(l Source[Location]) gox.Elem { return l.RouteBeam(r2Beams(rs)...) })
	}},
}

func r2mMatrix(t *testing.T, f func(t *testing.T, fam r2mFam, e r2mEntry, o r2mOrigin)) {
	t.Helper()
	for _, fam := range r2mFams {
		for _, e := range r2mEntries {
			if e.beam && !fam.beam {
				continue
			}
			for _, o := range r2mOrigins {
				if e.router && !o.router {
					continue
				}
				t.Run(fam.name+"/"+e.name+"/"+o.name, func(t *testing.T) { f(t, fam, e, o) })
			}
		}
	}
}

type r2mLocStep struct {
	uri    string
	want   string
	quiet  bool
	same   bool
	counts map[string]int
}

type r2mCase struct {
	name    string
	uri     string
	routes  func(r *r2Rec, f r2mFam) []SourceRoute[Location]
	initial string
	steps   []r2mLocStep
}

func r2mFamily(t *testing.T, cases ...r2mCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r2mMatrix(t, func(t *testing.T, fam r2mFam, e r2mEntry, o r2mOrigin) {
				workers := []int{8}
				if o.name == "NewSourceEqual" {
					workers = []int{8, 1}
				}
				for _, w := range workers {
					h := newR2HarnessN(t, w)
					loc := o.make(h, c.uri)
					h.render(r2Slot("m", e.mount(loc, c.routes(h.rec, fam)...)))
					steps := make([]r2Step[Location], len(c.steps))
					for i, s := range c.steps {
						ret := []error{nil}
						if s.same {
							ret = []error{}
						}
						steps[i] = r2Step[Location]{set: r2Loc(s.uri), want: s.want, quiet: s.quiet, ret: ret, counts: s.counts}
					}
					r2Run(h, "m", loc, c.initial, steps...)
				}
			})
		})
	}
}

var r2mItem1 = r2PathItem{Item: true, ID: 1}

// M01: the initial Location decodes into the model and the bind follows
// decoded changes.
func TestRoute2ModelDecode(t *testing.T) {
	r2mFamily(t, r2mCase{
		name: "item",
		uri:  "/item/5?tab=x",
		routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				r2mModel[r2PathItem](f, r2mBindF(r, "item", func(m r2PathItem) string {
					return fmt.Sprintf("item=%d/%s", m.ID, m.Tab)
				})),
			}
		},
		initial: "item=5/x",
		steps: []r2mLocStep{
			{uri: "/item/6", want: "item=6/"},
			{uri: "/zzz", want: "", counts: map[string]int{"render:item": 2, "clean:item": 2}},
			{uri: "/item/7?tab=q", want: "item=7/q"},
			{uri: "/item/7?other=1&tab=q", want: "item=7/q", quiet: true, counts: map[string]int{"render:item": 3, "clean:item": 2}},
			{uri: "/item/7?tab=q&other=1", want: "item=7/q", quiet: true, same: true},
		},
	})
}

// M02: A to B to A in one door; each model branch renders once per
// activation, the default takes unmatched locations and yields again.
func TestRoute2ModelTransitions(t *testing.T) {
	r2mFamily(t, r2mCase{
		name: "a-b-a",
		uri:  "/a",
		routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				r2mModel[r2PathA](f, r2mOnce[r2PathA](f, r, "A")),
				r2mModel[r2PathB](f, r2mSame(f, r, "dB", r2mOnce[r2PathB](f, r, "B"))),
				r2Comp[Location](r, "404"),
			}
		},
		initial: "A",
		steps: []r2mLocStep{
			{uri: "/b", want: "B", counts: map[string]int{"render:A": 1, "clean:A": 1, "render:B": 1}},
			{uri: "/a", want: "A", counts: map[string]int{"render:A": 2, "clean:B": 1}},
			{uri: "/b?x=1", want: "B"},
			{uri: "/b?x=2", want: "B", quiet: true},
			{uri: "/zzz", want: "404"},
			{uri: "/zzz?y=1", want: "404", quiet: true},
			{uri: "/a", want: "A", counts: map[string]int{
				"render:A": 3, "clean:A": 2, "render:B": 2, "clean:B": 2, "render:404": 1, "clean:404": 1, "derive:dB": 1,
			}},
		},
	})
}

// M03: a change within one model keeps the branch: a bind re-renders, a
// once-rendered leaf stays quiet, and a value route selects by DeepEqual.
func TestRoute2ModelWithin(t *testing.T) {
	r2mFamily(t,
		r2mCase{
			name: "bind",
			uri:  "/item/1",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{r2mModel[r2PathItem](f, r2Bind[r2PathItem](r, "B"))}
			},
			initial: "B={true 1 }",
			steps: []r2mLocStep{
				{uri: "/item/2", want: "B={true 2 }", counts: map[string]int{"render:B": 2, "clean:B": 1}},
				{uri: "/item/2?tab=t", want: "B={true 2 t}", counts: map[string]int{"render:B": 3, "clean:B": 2}},
			},
		},
		r2mCase{
			name: "once",
			uri:  "/item/1",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{r2mModel[r2PathItem](f, r2mOnce[r2PathItem](f, r, "S"))}
			},
			initial: "S",
			steps: []r2mLocStep{
				{uri: "/item/2", want: "S", quiet: true},
				{uri: "/item/3?tab=z", want: "S", quiet: true, counts: map[string]int{"render:S": 1, "clean:S": 0}},
			},
		},
		r2mCase{
			name: "value",
			uri:  "/item/1",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathItem](f,
						r2mValue(f, r2mItem1, r2mOnce[r2PathItem](f, r, "ONE")),
						r2Bind[r2PathItem](r, "OTHER"),
					),
				}
			},
			initial: "ONE",
			steps: []r2mLocStep{
				{uri: "/item/1?tab=x", want: "OTHER={true 1 x}", counts: map[string]int{"render:ONE": 1, "clean:ONE": 1, "render:OTHER": 1}},
				{uri: "/item/2", want: "OTHER={true 2 }", counts: map[string]int{"render:OTHER": 2, "clean:OTHER": 1}},
				{uri: "/item/1", want: "ONE", counts: map[string]int{"render:ONE": 2, "clean:OTHER": 2}},
				{uri: "/item/1?zz=9", want: "ONE", quiet: true, counts: map[string]int{"render:ONE": 2, "render:OTHER": 2}},
			},
		},
	)
}

// M04: a location change that decodes into a DeepEqual model is quiet even
// for a bind.
func TestRoute2ModelEqualDecode(t *testing.T) {
	r2mFamily(t, r2mCase{
		name: "unknown-query",
		uri:  "/item/1?tab=x",
		routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
			return []SourceRoute[Location]{r2mModel[r2PathItem](f, r2Bind[r2PathItem](r, "B"))}
		},
		initial: "B={true 1 x}",
		steps: []r2mLocStep{
			{uri: "/item/1?tab=x&zz=1", want: "B={true 1 x}", quiet: true, counts: map[string]int{"render:B": 1}},
			{uri: "/item/1?zz=2&tab=x", want: "B={true 1 x}", quiet: true, counts: map[string]int{"render:B": 1}},
			{uri: "/item/1", want: "B={true 1 }", counts: map[string]int{"render:B": 2, "clean:B": 1}},
		},
	})
}

// M07: the default renders unmatched locations and gives way to a model as
// soon as one matches.
func TestRoute2ModelDefault404(t *testing.T) {
	r2mFamily(t, r2mCase{
		name: "recover",
		uri:  "/zzz",
		routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				r2mModel[r2PathA](f, r2mOnce[r2PathA](f, r, "A")),
				r2Comp[Location](r, "404"),
			}
		},
		initial: "404",
		steps: []r2mLocStep{
			{uri: "/a", want: "A", counts: map[string]int{"render:A": 1, "clean:404": 1}},
			{uri: "/zzz", want: "404", counts: map[string]int{"render:404": 2, "clean:A": 1}},
			{uri: "/yyy", want: "404", quiet: true},
			{uri: "/a", want: "A", counts: map[string]int{"render:A": 2, "render:404": 2, "clean:404": 2}},
		},
	})
}

// M08: a Location that decodes into a model none of the model's children
// match does not match the model route: selection falls through to the later
// default, or leaves the container blank when nothing at the top matches.
// Rendering nothing for a decoded model takes an explicit nil leaf.
func TestRoute2ModelFallThrough(t *testing.T) {
	r2mFamily(t,
		r2mCase{
			name: "to-404",
			uri:  "/x",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathItem](f, r2mValue(f, r2mItem1, r2mOnce[r2PathItem](f, r, "ONE"))),
					r2Comp[Location](r, "404"),
				}
			},
			initial: "404",
			steps: []r2mLocStep{
				{uri: "/item/2", want: "404", quiet: true, counts: map[string]int{"render:404": 1, "clean:404": 0}},
				{uri: "/item/3", want: "404", quiet: true},
				{uri: "/item/1", want: "ONE", counts: map[string]int{"render:ONE": 1, "clean:404": 1}},
				{uri: "/item/2", want: "404", counts: map[string]int{"render:404": 2, "clean:ONE": 1}},
				{uri: "/x", want: "404", quiet: true, counts: map[string]int{"render:404": 2, "clean:404": 1, "render:ONE": 1}},
			},
		},
		r2mCase{
			name: "first-render",
			uri:  "/item/2",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathItem](f, r2mValue(f, r2mItem1, r2mOnce[r2PathItem](f, r, "ONE"))),
					r2Comp[Location](r, "404"),
				}
			},
			initial: "404",
			steps: []r2mLocStep{
				{uri: "/item/1", want: "ONE", counts: map[string]int{"render:ONE": 1, "clean:404": 1}},
				{uri: "/item/5", want: "404", counts: map[string]int{"render:404": 2, "clean:ONE": 1}},
			},
		},
		r2mCase{
			name: "blank",
			uri:  "/item/2",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathItem](f, r2mValue(f, r2mItem1, r2mOnce[r2PathItem](f, r, "ONE"))),
				}
			},
			initial: "",
			steps: []r2mLocStep{
				{uri: "/item/3", want: "", quiet: true},
				{uri: "/zzz", want: "", quiet: true},
				{uri: "/item/1", want: "ONE"},
				{uri: "/item/1?tab=t", want: "", counts: map[string]int{"render:ONE": 1, "clean:ONE": 1}},
				{uri: "/item/4", want: "", quiet: true},
				{uri: "/zzz", want: "", quiet: true},
				{uri: "/item/1", want: "ONE", counts: map[string]int{"render:ONE": 2, "clean:ONE": 1}},
				{uri: "/zzz", want: "", counts: map[string]int{"render:ONE": 2, "clean:ONE": 2}},
			},
		},
		r2mCase{
			name: "explicit-nil",
			uri:  "/x",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathItem](f,
						r2mValue(f, r2mItem1, r2mOnce[r2PathItem](f, r, "ONE")),
						RouteComp[r2PathItem](nil),
					),
					r2Comp[Location](r, "404"),
				}
			},
			initial: "404",
			steps: []r2mLocStep{
				{uri: "/item/2", want: "", counts: map[string]int{"render:404": 1, "clean:404": 1}},
				{uri: "/item/3", want: "", quiet: true},
				{uri: "/item/1", want: "ONE"},
				{uri: "/item/2", want: "", counts: map[string]int{"render:ONE": 1, "clean:ONE": 1, "render:404": 1}},
				{uri: "/x", want: "404", counts: map[string]int{"render:404": 2}},
				{uri: "/zzz", want: "404", quiet: true},
			},
		},
	)
}

// M08b: a model route with no children never matches, whatever the Location
// decodes into.
func TestRoute2ModelNoChildren(t *testing.T) {
	r2mFamily(t,
		r2mCase{
			name: "to-404",
			uri:  "/item/1",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathItem](f),
					r2mModel[r2PathA](f),
					r2Comp[Location](r, "404"),
				}
			},
			initial: "404",
			steps: []r2mLocStep{
				{uri: "/item/2", want: "404", quiet: true},
				{uri: "/a", want: "404", quiet: true},
				{uri: "/zzz", want: "404", quiet: true, counts: map[string]int{"render:404": 1, "clean:404": 0}},
			},
		},
		r2mCase{
			name: "blank",
			uri:  "/item/1",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{r2mModel[r2PathItem](f)}
			},
			initial: "",
			steps: []r2mLocStep{
				{uri: "/item/2", want: "", quiet: true},
				{uri: "/zzz", want: "", quiet: true},
			},
		},
	)
}

type r2mSettings struct {
	Path bool `path:"/settings"`
}

func r2mSectionOf(l Location) string {
	if len(l.Segments) == 0 {
		return ""
	}
	return l.Segments[0]
}

func r2mSectionSet(l Location, s string) Location {
	l = l.Clone()
	if len(l.Segments) == 0 {
		l.Segments = []string{s}
		return l
	}
	l.Segments[0] = s
	return l
}

func r2mSection(f r2mFam, r *r2Rec, rs ...SourceRoute[string]) SourceRoute[Location] {
	if f.beam {
		return RouteBeamDerive(r2DeriveBeam(r, "section", func(b Beam[Location]) Beam[string] {
			return DeriveBeam(b, r2mSectionOf)
		}), r2Beams(rs)...)
	}
	return RouteDerive(r2Derive(r, "section", func(s Source[Location]) Source[string] {
		return DeriveSource(s, r2mSectionOf, r2mSectionSet)
	}), rs...)
}

// M08c: a model before a section derive with a not-found default: leaving
// /docs for /settings breaks the active path, and the same update re-selects
// from scratch, so settings renders directly and notFound never shows.
func TestRoute2ModelSettingsSection(t *testing.T) {
	r2mFamily(t, r2mCase{
		name: "settings",
		uri:  "/docs",
		routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				r2mModel[r2mSettings](f, r2mOnce[r2mSettings](f, r, "settings")),
				r2mSection(f, r,
					r2mValue(f, "docs", r2mOnce[string](f, r, "docs")),
					r2Comp[string](r, "notFound"),
				),
			}
		},
		initial: "docs",
		steps: []r2mLocStep{
			{uri: "/settings", want: "settings", counts: map[string]int{"render:settings": 1, "clean:docs": 1, "render:notFound": 0}},
			{uri: "/docs", want: "docs", counts: map[string]int{"render:docs": 2, "clean:settings": 1, "render:notFound": 0}},
			{uri: "/docs?x=1", want: "docs", quiet: true},
			{uri: "/settings?x=1", want: "settings", counts: map[string]int{"render:settings": 2, "clean:docs": 2, "render:notFound": 0}},
			{uri: "/other", want: "notFound", counts: map[string]int{"render:notFound": 1, "clean:settings": 2}},
			{uri: "/settings", want: "settings", counts: map[string]int{"render:settings": 3, "clean:notFound": 1}},
			{uri: "/other", want: "notFound", counts: map[string]int{"render:notFound": 2}},
			{uri: "/docs", want: "docs", counts: map[string]int{
				"render:docs": 3, "render:settings": 3, "render:notFound": 2, "clean:notFound": 2, "derive:section": 1,
			}},
		},
	})
}

// M08d: the model derive skips an equal model: a query-only change that
// decodes into the same model re-runs no predicate under the model route,
// active or not; a changed model walks the subtree again.
func TestRoute2ModelDeriveSkip(t *testing.T) {
	r2mFamily(t, r2mCase{
		name: "query-only",
		uri:  "/item/2",
		routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				r2mModel[r2PathItem](f,
					r2mMatch(f, r2Pred(r, "p", func(m r2PathItem) bool { return m.ID == 1 }), r2mOnce[r2PathItem](f, r, "ONE")),
				),
				r2Comp[Location](r, "404"),
			}
		},
		initial: "404",
		steps: []r2mLocStep{
			{uri: "/item/2?zz=1", want: "404", quiet: true, counts: map[string]int{"pred:p={true 2 }": 1}},
			{uri: "/item/2?zz=2", want: "404", quiet: true, counts: map[string]int{"pred:p={true 2 }": 1}},
			{uri: "/item/1", want: "ONE", counts: map[string]int{"pred:p={true 1 }": 1, "render:ONE": 1}},
			{uri: "/item/1?zz=1", want: "ONE", quiet: true, counts: map[string]int{"pred:p={true 1 }": 1}},
			{uri: "/item/1?zz=2&tab=", want: "ONE", quiet: true, counts: map[string]int{"pred:p={true 1 }": 1}},
			{uri: "/item/1?tab=t", want: "ONE", quiet: true, counts: map[string]int{"pred:p={true 1 t}": 1, "render:ONE": 1}},
			{uri: "/item/1?tab=t&zz=1", want: "ONE", quiet: true, counts: map[string]int{"pred:p={true 1 t}": 1}},
			{uri: "/item/2", want: "404", counts: map[string]int{"pred:p={true 2 }": 3, "render:404": 2, "clean:ONE": 1}},
			{uri: "/item/2?zz=3", want: "404", quiet: true, counts: map[string]int{"pred:p={true 2 }": 3}},
		},
	})
}

// M09: a model type without a valid adapter never matches; later siblings
// render instead. The adapter error goes to slog.Default.
func TestRoute2ModelAdapterError(t *testing.T) {
	r2mFamily(t, r2mCase{
		name: "bad-first",
		uri:  "/a",
		routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
			return []SourceRoute[Location]{
				r2mModel[r2PathBad](f, r2mOnce[r2PathBad](f, r, "BAD")),
				r2mModel[int](f, r2mOnce[int](f, r, "INT")),
				r2mModel[r2PathA](f, r2mOnce[r2PathA](f, r, "A")),
				r2Comp[Location](r, "404"),
			}
		},
		initial: "A",
		steps: []r2mLocStep{
			{uri: "/zzz", want: "404"},
			{uri: "/", want: "404", quiet: true},
			{uri: "/a", want: "A", counts: map[string]int{"render:BAD": 0, "render:INT": 0, "render:A": 2, "render:404": 1}},
		},
	})
}

// M10: overlapping models: the first in order wins on seek, and an active
// model stays while it still decodes, even if an earlier one matches.
func TestRoute2ModelOverlap(t *testing.T) {
	r2mFamily(t,
		r2mCase{
			name: "once",
			uri:  "/item/abc",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathItem](f, r2mOnce[r2PathItem](f, r, "INT")),
					r2mModel[r2PathName](f, r2mOnce[r2PathName](f, r, "NAME")),
					r2Comp[Location](r, "404"),
				}
			},
			initial: "NAME",
			steps: []r2mLocStep{
				{uri: "/item/5", want: "NAME", quiet: true},
				{uri: "/other", want: "404"},
				{uri: "/item/5", want: "INT"},
				{uri: "/item/abc", want: "NAME", counts: map[string]int{"render:NAME": 2, "render:INT": 1, "clean:INT": 1, "render:404": 1}},
			},
		},
		r2mCase{
			name: "bind",
			uri:  "/item/abc",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathItem](f, r2Bind[r2PathItem](r, "INT")),
					r2mModel[r2PathName](f, r2Bind[r2PathName](r, "NAME")),
					r2Comp[Location](r, "404"),
				}
			},
			initial: "NAME={true abc}",
			steps: []r2mLocStep{
				{uri: "/item/5", want: "NAME={true 5}"},
				{uri: "/other", want: "404"},
				{uri: "/item/5", want: "INT={true 5 }", counts: map[string]int{"render:NAME": 2, "render:INT": 1}},
			},
		},
	)
}

// M11: RouteBeamModel hands out a read-only Beam that follows the decoded
// model without re-rendering the leaf.
func TestRoute2ModelBeamReadOnly(t *testing.T) {
	h := newR2Harness(t)
	loc := r2LocSource("/item/5?tab=x")
	box, dbox := &r2Box[Beam[r2PathItem]]{}, &r2Box[Beam[r2PathItem]]{}
	h.render(
		r2Slot("a", loc.RouteBeam(RouteBeamModel[r2PathItem](r2Beam(h.rec, "I", box)))),
		r2Slot("b", DeriveBeamEqual(loc, r2mLocID, path.EqualLocation).RouteBeam(RouteBeamModel[r2PathItem](r2Beam(h.rec, "J", dbox)))),
	)
	for _, b := range []*r2Box[Beam[r2PathItem]]{box, dbox} {
		if b.len() != 1 {
			t.Fatalf("expected one beam capture, got %d", b.len())
		}
		if _, ok := b.last().(Source[r2PathItem]); ok {
			t.Fatal("RouteBeamModel leaf got a writable Source")
		}
	}
	model := func(want r2PathItem) func(string) {
		return func(at string) {
			for _, b := range []*r2Box[Beam[r2PathItem]]{box, dbox} {
				if got := b.last().Get(); got != want {
					t.Fatalf("%sbeam holds %+v, want %+v", at, got, want)
				}
			}
		}
	}
	r2mRun(h,
		r2mStep{at: "initial", want: map[string]string{"a": "I", "b": "J"}, check: model(r2PathItem{Item: true, ID: 5, Tab: "x"})},
		r2mStep{at: "/item/6", op: r2mGo(loc, "/item/6"), quiet: []string{"a", "b"}, want: map[string]string{"a": "I", "b": "J"}, check: model(r2PathItem{Item: true, ID: 6})},
		r2mStep{at: "/zzz", op: r2mGo(loc, "/zzz"), want: map[string]string{"a": "", "b": ""}},
		r2mStep{at: "/item/7", op: r2mGo(loc, "/item/7"), want: map[string]string{"a": "I", "b": "J"}, check: model(r2PathItem{Item: true, ID: 7}),
			counts: map[string]int{"render:I": 2, "render:J": 2, "clean:I": 1, "clean:J": 1}},
	)
	if box.len() != 2 || dbox.len() != 2 {
		t.Fatalf("expected one capture per activation, got %d and %d", box.len(), dbox.len())
	}
}

// M05: writing the branch's model Source re-encodes the Location; the branch
// stays, nested routes follow, and an equal model is suppressed.
func TestRoute2ModelWriteBack(t *testing.T) {
	for _, o := range r2mOrigins {
		t.Run(o.name, func(t *testing.T) {
			h := newR2Harness(t)
			loc := o.make(h, "/item/3?tab=x")
			box := &r2Box[Source[r2PathItem]]{}
			sbox := &r2Box[Source[r2PathItem]]{}
			h.render(
				r2Slot("m", loc.Route(
					RouteModel[r2PathItem](RouteDerive(r2mCapture(h.rec, "cap", box),
						RouteValue(r2mItem1, r2Comp[r2PathItem](h.rec, "ONE")),
						r2Bind[r2PathItem](h.rec, "B"),
					)),
					r2Comp[Location](h.rec, "404"),
				)),
				r2Slot("s", loc.Route(RouteModel[r2PathItem](r2Source(h.rec, "S", sbox)))),
			)
			r2mRun(h,
				r2mStep{at: "initial", want: map[string]string{"m": "B={true 3 x}", "s": "S"}, counts: map[string]int{"derive:cap": 1}},
				r2mStep{at: "write 7/y", op: r2mPutBox(box, r2PathItem{Item: true, ID: 7, Tab: "y"}), ret: []error{nil}, quiet: []string{"s"},
					want: map[string]string{"m": "B={true 7 y}", "s": "S"}, check: r2mLocIs(h, loc, "/item/7?tab=y")},
				r2mStep{at: "write equal", op: r2mPutBox(box, r2PathItem{Item: true, ID: 7, Tab: "y"}), ret: []error{}, quiet: []string{"m", "s"},
					want: map[string]string{"m": "B={true 7 y}", "s": "S"}, check: r2mLocIs(h, loc, "/item/7?tab=y")},
				r2mStep{at: "write 1", op: r2mPutBox(box, r2mItem1), ret: []error{nil}, quiet: []string{"s"},
					want: map[string]string{"m": "ONE", "s": "S"}, check: r2mLocIs(h, loc, "/item/1?tab=")},
				r2mStep{at: "write via leaf", op: r2mPutBox(sbox, r2PathItem{Item: true, ID: 2}), ret: []error{nil}, quiet: []string{"s"},
					want: map[string]string{"m": "B={true 2 }", "s": "S"}, check: r2mLocIs(h, loc, "/item/2?tab=")},
				r2mStep{at: "mutate", op: func(ctx context.Context) <-chan error {
					return box.last().Mutate(ctx, func(m r2PathItem) r2PathItem { m.Tab = "z"; return m })
				}, ret: []error{nil}, quiet: []string{"s"}, want: map[string]string{"m": "B={true 2 z}", "s": "S"}, check: r2mLocIs(h, loc, "/item/2?tab=z"),
					counts: map[string]int{"derive:cap": 1, "render:S": 1, "render:ONE": 1, "clean:ONE": 1, "render:B": 4, "clean:B": 3}},
				r2mStep{at: "leave", op: r2mGo(loc, "/zzz"), want: map[string]string{"m": "404", "s": ""}},
				r2mStep{at: "write while inactive", op: r2mPutBox(box, r2PathItem{Item: true, ID: 5}), ret: []error{nil},
					want: map[string]string{"m": "B={true 5 }", "s": "S"}, check: r2mLocIs(h, loc, "/item/5?tab=")},
			)
			if box.len() != 1 {
				t.Fatalf("expected the derive to run once, got %d", box.len())
			}
			if o.router {
				r2mWantPaths(h, r2mSP("/item/7?tab=y", false), r2mSP("/item/1?tab=", false), r2mSP("/item/2?tab=", false),
					r2mSP("/item/2?tab=z", false), r2mSP("/zzz", false), r2mSP("/item/5?tab=", false))
			}
		})
	}
}

// M06: a model or custom value that cannot be encoded leaves the Location
// unchanged: the write is suppressed and nothing reaches the client.
func TestRoute2ModelEncodeError(t *testing.T) {
	for _, o := range r2mOrigins {
		t.Run(o.name, func(t *testing.T) {
			h := newR2Harness(t)
			loc := o.make(h, "/item/abc")
			nbox := &r2Box[Source[r2PathName]]{}
			cbox := &r2Box[Source[r2Custom]]{}
			h.render(
				r2Slot("n", loc.Route(RouteModel[r2PathName](RouteDerive(r2mCapture(h.rec, "n", nbox), r2Bind[r2PathName](h.rec, "N"))))),
				r2Slot("c", loc.Route(RouteMatch(r2mIsItem, RouteDerive(r2mCustomItemLens,
					RouteDerive(r2mCapture(h.rec, "c", cbox), r2Bind[r2Custom](h.rec, "C")))))),
			)
			both := []string{"n", "c"}
			r2mRun(h,
				r2mStep{at: "initial", want: map[string]string{"n": "N={true abc}", "c": "C={abc }"}},
				r2mStep{at: "empty name", op: r2mPutBox(nbox, r2PathName{Item: true}), ret: []error{}, quiet: both,
					want: map[string]string{"n": "N={true abc}", "c": "C={abc }"}, check: r2mLocIs(h, loc, "/item/abc")},
				r2mStep{at: "zero model", op: r2mPutBox(nbox, r2PathName{}), ret: []error{}, quiet: both,
					want: map[string]string{"n": "N={true abc}", "c": "C={abc }"}, check: r2mLocIs(h, loc, "/item/abc")},
				r2mStep{at: "empty custom", op: r2mPutBox(cbox, r2Custom{Tab: "t"}), ret: []error{}, quiet: both,
					want: map[string]string{"n": "N={true abc}", "c": "C={abc }"}, check: r2mLocIs(h, loc, "/item/abc")},
				r2mStep{at: "custom", op: r2mPutBox(cbox, r2Custom{ID: "z", Tab: "t"}), ret: []error{nil},
					want: map[string]string{"n": "N={true z}", "c": "C={z t}"}, check: r2mLocIs(h, loc, "/item/z?tab=t")},
				r2mStep{at: "name", op: r2mPutBox(nbox, r2PathName{Item: true, Name: "w"}), ret: []error{nil},
					want: map[string]string{"n": "N={true w}", "c": "C={w }"}, check: r2mLocIs(h, loc, "/item/w"),
					counts: map[string]int{"render:N": 3, "render:C": 3}},
			)
			if o.router {
				r2mWantPaths(h, r2mSP("/item/z?tab=t", false), r2mSP("/item/w", false))
			}
		})
	}
}

// M12: a prefix lens inside a match routes the rest of the Location through
// nested models, with inner and outer defaults, and writes restore the prefix.
func TestRoute2ModelNestedLens(t *testing.T) {
	isApp := func(l Location) bool { return len(l.Segments) != 0 && l.Segments[0] == "app" }
	for _, o := range r2mOrigins {
		t.Run(o.name, func(t *testing.T) {
			h := newR2Harness(t)
			loc := o.make(h, "/app/item/3")
			box := &r2Box[Source[r2PathItem]]{}
			h.render(r2Slot("m", loc.Route(
				RouteMatch(isApp, RouteDerive(r2Derive(h.rec, "strip", r2Strip("app")),
					RouteModel[r2PathA](r2Source[r2PathA](h.rec, "A", nil)),
					RouteModel[r2PathItem](RouteDerive(r2mCapture(h.rec, "cap", box), r2Bind[r2PathItem](h.rec, "ITEM"))),
					r2Comp[Location](h.rec, "inner404"),
				)),
				r2Comp[Location](h.rec, "outer404"),
			)))
			m := func(s string) map[string]string { return map[string]string{"m": s} }
			r2mRun(h,
				r2mStep{at: "initial", want: m("ITEM={true 3 }")},
				r2mStep{at: "write 4", op: r2mPutBox(box, r2PathItem{Item: true, ID: 4}), ret: []error{nil}, want: m("ITEM={true 4 }"), check: r2mLocIs(h, loc, "/app/item/4?tab=")},
				r2mStep{at: "write tab", op: r2mPutBox(box, r2PathItem{Item: true, ID: 4, Tab: "t"}), ret: []error{nil}, want: m("ITEM={true 4 t}"), check: r2mLocIs(h, loc, "/app/item/4?tab=t")},
				r2mStep{at: "/app/zzz", op: r2mGo(loc, "/app/zzz"), want: m("inner404")},
				r2mStep{at: "/app/a", op: r2mGo(loc, "/app/a"), want: m("A")},
				r2mStep{at: "/app/a?q=1", op: r2mGo(loc, "/app/a?q=1"), quiet: []string{"m"}, want: m("A")},
				r2mStep{at: "/zzz", op: r2mGo(loc, "/zzz"), want: m("outer404")},
				r2mStep{at: "/app/item/8", op: r2mGo(loc, "/app/item/8"), want: m("ITEM={true 8 }")},
				r2mStep{at: "/app", op: r2mGo(loc, "/app"), want: m("inner404")},
				r2mStep{at: "/app/zz", op: r2mGo(loc, "/app/zz"), quiet: []string{"m"}, want: m("inner404"),
					counts: map[string]int{"derive:strip": 1, "derive:cap": 1, "render:A": 1, "render:inner404": 2, "render:outer404": 1, "render:ITEM": 4}},
			)
			if o.router {
				r2mWantPaths(h, r2mSP("/app/item/4?tab=", false), r2mSP("/app/item/4?tab=t", false), r2mSP("/app/zzz", false),
					r2mSP("/app/a", false), r2mSP("/app/a?q=1", false), r2mSP("/zzz", false), r2mSP("/app/item/8", false),
					r2mSP("/app", false), r2mSP("/app/zz", false))
			}
		})
	}
	t.Run("beam", func(t *testing.T) {
		h := newR2Harness(t)
		loc := r2LocSource("/app/item/3")
		h.render(r2Slot("m", loc.RouteBeam(
			RouteBeamMatch(isApp, RouteBeamDerive(r2DeriveBeam(h.rec, "strip", r2mStripBeam),
				RouteBeamModel[r2PathA](r2Beam[r2PathA](h.rec, "A", nil)),
				RouteBeamModel[r2PathItem](r2Bind[r2PathItem](h.rec, "ITEM")),
				r2Comp[Location](h.rec, "inner404"),
			)),
			r2Comp[Location](h.rec, "outer404"),
		)))
		r2Run(h, "m", loc, "ITEM={true 3 }",
			r2Step[Location]{set: r2Loc("/app/item/4?tab=t"), want: "ITEM={true 4 t}"},
			r2Step[Location]{set: r2Loc("/app/zzz"), want: "inner404"},
			r2Step[Location]{set: r2Loc("/app/a"), want: "A"},
			r2Step[Location]{set: r2Loc("/zzz"), want: "outer404"},
			r2Step[Location]{set: r2Loc("/app/item/8"), want: "ITEM={true 8 }"},
			r2Step[Location]{set: r2Loc("/app"), want: "inner404", counts: map[string]int{"derive:strip": 1}},
		)
	})
}

// M13: raw Location routes compare with the source comparator; a write
// through the raw Source switches the branch. A nil comparator never matches
// a value route and re-renders binds on every update.
func TestRoute2ModelRawLocation(t *testing.T) {
	isSeg := func(l Location) bool { return len(l.Segments) == 2 && l.Segments[0] == "seg" }
	for _, o := range r2mOrigins {
		t.Run(o.name, func(t *testing.T) {
			h := newR2Harness(t)
			loc := o.make(h, "/raw")
			box := &r2Box[Source[Location]]{}
			h.render(r2Slot("m", loc.Route(
				RouteValue(r2Loc("/raw"), r2Source(h.rec, "RAW", box)),
				RouteMatch(isSeg, r2Bind[Location](h.rec, "SEG")),
				r2Comp[Location](h.rec, "404"),
			)))
			m := func(s string) map[string]string { return map[string]string{"m": s} }
			r2mRun(h,
				r2mStep{at: "initial", want: m("RAW")},
				r2mStep{at: "write /seg/x", op: r2mPutBox(box, r2Loc("/seg/x")), ret: []error{nil}, want: m("SEG=/seg/x"), check: r2mLocIs(h, loc, "/seg/x")},
				r2mStep{at: "/seg/y", op: r2mGo(loc, "/seg/y"), want: m("SEG=/seg/y")},
				r2mStep{at: "/seg/y?b=2&a=1", op: r2mGo(loc, "/seg/y?b=2&a=1"), want: m("SEG=/seg/y?a=1&b=2")},
				r2mStep{at: "/seg/y?a=1&b=2", op: r2mGo(loc, "/seg/y?a=1&b=2"), ret: []error{}, quiet: []string{"m"}, want: m("SEG=/seg/y?a=1&b=2")},
				r2mStep{at: "/raw?x=1", op: r2mGo(loc, "/raw?x=1"), want: m("404")},
				r2mStep{at: "/raw", op: r2mGo(loc, "/raw"), want: m("RAW")},
				r2mStep{at: "/raw/", op: r2mGo(loc, "/raw/"), ret: []error{}, quiet: []string{"m"}, want: m("RAW"),
					counts: map[string]int{"render:RAW": 2, "render:SEG": 3, "render:404": 1}},
			)
			if o.router {
				r2mWantPaths(h, r2mSP("/seg/x", false), r2mSP("/seg/y", false), r2mSP("/seg/y?a=1&b=2", false),
					r2mSP("/raw?x=1", false), r2mSP("/raw", false))
			}
		})
	}
	t.Run("nil comparator", func(t *testing.T) {
		h := newR2Harness(t)
		loc := NewSourceEqual(r2Loc("/raw"), nil)
		h.render(r2Slot("m", loc.Route(
			RouteValue(r2Loc("/raw"), r2Comp[Location](h.rec, "RAW")),
			r2Bind[Location](h.rec, "ANY"),
		)))
		r2Run(h, "m", loc, "ANY=/raw",
			r2Step[Location]{set: r2Loc("/raw"), want: "ANY=/raw", ret: []error{nil}, counts: map[string]int{"render:ANY": 2, "render:RAW": 0}},
			r2Step[Location]{set: r2Loc("/other"), want: "ANY=/other", ret: []error{nil}, counts: map[string]int{"render:ANY": 3, "render:RAW": 0}},
		)
	})
}

func r2mCombinedSrc(r *r2Rec, box *r2Box[Source[r2mCustomRoute]]) []SourceRoute[Location] {
	return []SourceRoute[Location]{
		RouteModel[r2mCrossA](r2Source[r2mCrossA](r, "model-a", nil)),
		RouteMatch(r2mIsCustom, RouteDerive(r2mCustomLens, r2Source(r, "custom", box))),
		RouteModel[r2mQuery](r2Source[r2mQuery](r, "query", nil)),
		RouteMatch(r2mIsRaw, r2Source[Location](r, "raw", nil)),
		r2Comp[Location](r, "default"),
	}
}

func r2mCombinedBind(r *r2Rec) []SourceRoute[Location] {
	return []SourceRoute[Location]{
		RouteModel[r2mCrossA](r2mBindF(r, "model-a", func(r2mCrossA) string { return "model-a" })),
		RouteMatch(r2mIsCustom, RouteDerive(r2mCustomLens, r2mBindF(r, "custom", r2mCustomText("custom")))),
		RouteModel[r2mQuery](r2mBindF(r, "query", r2mQueryText)),
		RouteMatch(r2mIsRaw, r2mBindF(r, "raw", func(l Location) string { return "raw|" + l.Query.Get("from") })),
		r2Comp[Location](r, "default"),
	}
}

// M14a: routerCombinedLensDocument rebuilt with Route in one door: model,
// custom lens, query model, raw match and default.
func TestRoute2ModelPortCombinedLens(t *testing.T) {
	for _, o := range r2mOrigins {
		t.Run(o.name, func(t *testing.T) {
			h := newR2Harness(t)
			loc := o.make(h, "/cross-a")
			rb := newR2Rec()
			box := &r2Box[Source[r2mCustomRoute]]{}
			h.render(
				r2Slot("s", loc.Route(r2mCombinedSrc(h.rec, box)...)),
				r2Slot("b", loc.Route(r2mCombinedBind(rb)...)),
			)
			w := func(s, b string) map[string]string { return map[string]string{"s": s, "b": b} }
			r2mRun(h,
				r2mStep{at: "initial", want: w("model-a", "model-a")},
				r2mStep{at: "model-to-custom", op: r2mGo(loc, "/custom/hello%20world%2Fone?tab=details"),
					want: w("custom", "custom|hello world/one|details"), check: func(at string) {
						if got := loc.Get().Path(); got != "/custom/hello%20world%2Fone" {
							t.Fatalf("%sescaped path %q", at, got)
						}
					}},
				r2mStep{at: "custom-next", op: r2mGo(loc, "/custom/next%2Fchild?tab=again"), quiet: []string{"s"},
					want: w("custom", "custom|next/child|again")},
				r2mStep{at: "custom-lens-write", op: r2mPutBox(box, r2mCustomRoute{ID: "lens write/value", Tab: "written"}), ret: []error{nil},
					quiet: []string{"s"}, want: w("custom", "custom|lens write/value|written"), check: func(at string) {
						if got := loc.Get().String(); got != "/custom/lens%20write%2Fvalue?tab=written" {
							t.Fatalf("%slocation %q", at, got)
						}
					}},
				r2mStep{at: "custom-to-query", op: r2mGo(loc, "/q?tag=from-custom&page=5"), want: w("query", "query|from-custom|5")},
				r2mStep{at: "query-to-raw", op: r2mGo(loc, "/raw?from=query"), want: w("raw", "raw|query")},
				r2mStep{at: "raw-to-model", op: r2mGo(loc, "/cross-a"), want: w("model-a", "model-a")},
				r2mStep{at: "missing", op: r2mGo(loc, "/missing"), want: w("default", "default")},
				r2mStep{at: "back", op: r2mGo(loc, "/cross-a"), want: w("model-a", "model-a"),
					counts: map[string]int{"render:model-a": 3, "render:custom": 1, "render:query": 1, "render:raw": 1, "render:default": 1}},
			)
			if got := rb.count("render:custom"); got != 3 {
				t.Fatalf("custom bind rendered %d times, want 3", got)
			}
			if o.router {
				r2mWantPaths(h, r2mSP("/custom/hello%20world%2Fone?tab=details", false), r2mSP("/custom/next%2Fchild?tab=again", false),
					r2mSP("/custom/lens%20write%2Fvalue?tab=written", false), r2mSP("/q?page=5&tag=from-custom", false),
					r2mSP("/raw?from=query", false), r2mSP("/cross-a", false), r2mSP("/missing", false), r2mSP("/cross-a", false))
			}
		})
	}
}

// M14b: routerCombinedBeamDocument rebuilt with RouteBeam and
// RouteBeamDerive: the custom branch wins over the earlier model on load.
func TestRoute2ModelPortCombinedBeam(t *testing.T) {
	for _, o := range r2mOrigins {
		t.Run(o.name, func(t *testing.T) {
			h := newR2Harness(t)
			loc := o.make(h, "/custom/beam%2Fid?tab=start")
			rb := newR2Rec()
			box := &r2Box[Beam[r2mCustomRoute]]{}
			h.render(
				r2Slot("o", loc.RouteBeam(
					RouteBeamModel[r2mCrossA](r2Beam[r2mCrossA](h.rec, "cross-a", nil)),
					RouteBeamMatch(r2mIsCustom, RouteBeamDerive(r2mCustomBeamLens, r2Beam(h.rec, "custom-beam", box))),
					RouteBeamModel[r2mQuery](r2Beam[r2mQuery](h.rec, "query-beam", nil)),
					r2Comp[Location](h.rec, "default"),
				)),
				r2Slot("b", DeriveBeamEqual(loc, r2mLocID, path.EqualLocation).RouteBeam(
					RouteBeamModel[r2mCrossA](r2mBindF(rb, "cross-a", func(m r2mCrossA) string { return "cross-a|" + fmt.Sprint(m.Path) })),
					RouteBeamMatch(r2mIsCustom, RouteBeamDerive(r2mCustomBeamLens, r2mBindF(rb, "custom-beam", r2mCustomText("custom-beam")))),
					RouteBeamModel[r2mQuery](r2mBindF(rb, "query-beam", func(m r2mQuery) string {
						if m.Tag == nil {
							return "query-beam|"
						}
						return "query-beam|" + *m.Tag
					})),
					r2Comp[Location](rb, "default"),
				)),
			)
			held := func(want r2mCustomRoute) func(string) {
				return func(at string) {
					if got := box.last().Get(); got != want {
						t.Fatalf("%scustom beam holds %+v, want %+v", at, got, want)
					}
				}
			}
			w := func(o, b string) map[string]string { return map[string]string{"o": o, "b": b} }
			r2mRun(h,
				r2mStep{at: "initial", want: w("custom-beam", "custom-beam|beam/id|start"), check: held(r2mCustomRoute{ID: "beam/id", Tab: "start"})},
				r2mStep{at: "beam-custom-next", op: r2mGo(loc, "/custom/beam%20next?tab=read"), quiet: []string{"o"},
					want: w("custom-beam", "custom-beam|beam next|read"), check: held(r2mCustomRoute{ID: "beam next", Tab: "read"})},
				r2mStep{at: "beam-custom-to-model", op: r2mGo(loc, "/cross-a"), want: w("cross-a", "cross-a|true")},
				r2mStep{at: "query", op: r2mGo(loc, "/q?tag=t"), want: w("query-beam", "query-beam|t")},
				r2mStep{at: "missing", op: r2mGo(loc, "/missing"), want: w("default", "default")},
				r2mStep{at: "custom again", op: r2mGo(loc, "/custom/x"), want: w("custom-beam", "custom-beam|x|"),
					counts: map[string]int{"render:custom-beam": 2, "render:cross-a": 1, "render:query-beam": 1, "render:default": 1}},
			)
		})
	}
}

// M14c: routeBindDocument rebuilt with binds: derive-bind, match-bind and a
// default bind.
func TestRoute2ModelPortRouteBind(t *testing.T) {
	for _, o := range r2mOrigins {
		t.Run(o.name, func(t *testing.T) {
			h := newR2Harness(t)
			loc := o.make(h, "/custom/bind-test?tab=active")
			h.render(r2Slot("m", loc.Route(
				RouteMatch(r2mIsCustom, RouteDerive(r2mCustomLens, r2mBindF(h.rec, "derive-bind", r2mCustomText("derive-bind")))),
				RouteMatch(r2mIsCrossA, r2mBindF(h.rec, "match-bind", func(l Location) string { return "match-bind|" + l.Path() })),
				r2mBindF(h.rec, "default-bind", func(l Location) string { return "default-bind|" + l.Path() }),
			)))
			r2Run(h, "m", loc, "derive-bind|bind-test|active",
				r2Step[Location]{set: r2Loc("/cross-a"), want: "match-bind|/cross-a"},
				r2Step[Location]{set: r2Loc("/custom/bind-test?tab=active"), want: "derive-bind|bind-test|active"},
				r2Step[Location]{set: r2Loc("/unknown"), want: "default-bind|/unknown"},
				r2Step[Location]{set: r2Loc("/cross-a"), want: "match-bind|/cross-a"},
				r2Step[Location]{set: r2Loc("/other"), want: "default-bind|/other"},
				r2Step[Location]{set: r2Loc("/custom/a"), want: "derive-bind|a|", counts: map[string]int{
					"render:derive-bind": 3, "render:match-bind": 2, "render:default-bind": 2, "clean:default-bind": 2,
				}},
			)
		})
	}
}

// M15: model routing from derived entry points matches routing on a root
// source, and writes through the derived entry restore the prefix.
func TestRoute2ModelDerivedEntry(t *testing.T) {
	h := newR2Harness(t)
	locP := r2LocSource("/app/item/3")
	locR := r2LocSource("/item/3")
	recP, recR := newR2Rec(), newR2Rec()
	box := &r2Box[Source[r2PathItem]]{}
	tree := func(r *r2Rec, box *r2Box[Source[r2PathItem]]) []SourceRoute[Location] {
		return []SourceRoute[Location]{
			RouteModel[r2PathA](r2Source[r2PathA](r, "A", nil)),
			RouteModel[r2PathItem](RouteDerive(r2mCapture(r, "cap", box), r2Bind[r2PathItem](r, "ITEM"))),
			r2Comp[Location](r, "404"),
		}
	}
	beamTree := func(r *r2Rec) []BeamRoute[Location] {
		return []BeamRoute[Location]{
			RouteBeamModel[r2PathA](r2Beam[r2PathA](r, "A", nil)),
			RouteBeamModel[r2PathItem](r2Bind[r2PathItem](r, "ITEM")),
			r2Comp[Location](r, "404"),
		}
	}
	h.render(
		r2Slot("strip", r2Strip("app")(locP).Route(tree(recP, box)...)),
		r2Slot("root", locR.Route(tree(recR, nil)...)),
		r2Slot("stripbeam", r2mStripBeam(locP).RouteBeam(beamTree(newR2Rec())...)),
		r2Slot("stripsrcbeam", r2Strip("app")(locP).RouteBeam(beamTree(newR2Rec())...)),
		r2Slot("rootbeam", locR.RouteBeam(beamTree(newR2Rec())...)),
		r2Slot("idbeam", DeriveBeamEqual(locR, r2mLocID, path.EqualLocation).RouteBeam(beamTree(newR2Rec())...)),
	)
	slots := []string{"strip", "root", "stripbeam", "stripsrcbeam", "rootbeam", "idbeam"}
	check := func(at, want string) {
		t.Helper()
		for _, s := range slots {
			h.expectAt(at, s, want)
			h.oneDoorAt(at, s)
		}
	}
	both := func(rest string) {
		t.Helper()
		h.settle(func(ctx context.Context) {
			locP.Update(ctx, r2Loc("/app"+rest))
			locR.Update(ctx, r2Loc(rest))
		})
	}
	check("initial", "ITEM={true 3 }")
	for _, st := range []struct{ rest, want string }{
		{"/a", "A"},
		{"/item/5?tab=t", "ITEM={true 5 t}"},
		{"/zzz", "404"},
		{"/item/6", "ITEM={true 6 }"},
	} {
		both(st.rest)
		check(st.rest, st.want)
	}
	if errs := r2Set(h, box.last(), r2PathItem{Item: true, ID: 9}); !r2SameErrs(errs, []error{nil}) {
		t.Fatalf("write through the derived entry returned %v", errs)
	}
	r2mLocIs(h, locP, "/app/item/9?tab=")("write: ")
	h.expect("strip", "ITEM={true 9 }")
	h.expect("stripbeam", "ITEM={true 9 }")
	h.expect("root", "ITEM={true 6 }")
	r2Set(h, locR, r2Loc("/item/9"))
	check("synced", "ITEM={true 9 }")
	both("/a")
	check("/a again", "A")
	if recP.count("derive:cap") != 1 || recR.count("derive:cap") != 1 {
		t.Fatalf("expected one derive per instance, got %d and %d", recP.count("derive:cap"), recR.count("derive:cap"))
	}
	if recP.count("render:A") != 2 || recR.count("render:A") != 2 {
		t.Fatalf("A rendered %d and %d times, want 2", recP.count("render:A"), recR.count("render:A"))
	}
}

// Route values are templates: one routes list mounted twice, or in several
// instances, gets independent state.
func TestRoute2ModelTemplateShared(t *testing.T) {
	shared := newR2Rec()
	rs := []SourceRoute[Location]{
		RouteModel[r2PathItem](RouteDerive(r2Derive(shared, "cap", func(s Source[r2PathItem]) Source[r2PathItem] { return s }),
			r2Bind[r2PathItem](shared, "B"))),
		r2Comp[Location](shared, "404"),
	}
	t.Run("two slots", func(t *testing.T) {
		h := newR2Harness(t)
		locA, locB := r2LocSource("/item/1"), r2LocSource("/zzz")
		before := shared.count("derive:cap")
		h.render(r2Slot("a", locA.Route(rs...)), r2Slot("b", locB.Route(rs...)))
		if got := shared.count("derive:cap") - before; got != 2 {
			t.Fatalf("expected one derive per mounted instance, got %d", got)
		}
		r2mRun(h,
			r2mStep{at: "initial", want: map[string]string{"a": "B={true 1 }", "b": "404"}},
			r2mStep{at: "a", op: r2mGo(locA, "/item/2"), quiet: []string{"b"}, want: map[string]string{"a": "B={true 2 }", "b": "404"}},
			r2mStep{at: "b", op: r2mGo(locB, "/item/7"), quiet: []string{"a"}, want: map[string]string{"a": "B={true 2 }", "b": "B={true 7 }"}},
			r2mStep{at: "a leaves", op: r2mGo(locA, "/x"), quiet: []string{"b"}, want: map[string]string{"a": "404", "b": "B={true 7 }"}},
		)
	})
	before := shared.count("derive:cap")
	r2Parallel(t, 6, func(t *testing.T, i int) {
		h := newR2Harness(t)
		loc := r2LocSource("/item/" + strconv.Itoa(i))
		h.render(r2Slot("m", loc.Route(rs...)))
		r2Run(h, "m", loc, fmt.Sprintf("B={true %d }", i),
			r2Step[Location]{set: r2Loc(fmt.Sprintf("/item/%d?tab=t", i)), want: fmt.Sprintf("B={true %d t}", i)},
			r2Step[Location]{set: r2Loc("/none"), want: "404"},
			r2Step[Location]{set: r2Loc(fmt.Sprintf("/item/%d", i+10)), want: fmt.Sprintf("B={true %d }", i+10)},
		)
	})
	if got := shared.count("derive:cap") - before; got != 6 {
		t.Fatalf("expected one derive per instance, got %d", got)
	}
}

// Several location updates inside one settle batch end on the last value in
// one door, with skipping and with NoSkip sources.
func TestRoute2ModelBurst(t *testing.T) {
	scripts := []struct {
		uris []string
		want string
	}{
		{[]string{"/a", "/item/1", "/zzz", "/item/2?tab=q", "/a", "/item/9"}, "ITEM={true 9 }"},
		{[]string{"/item/1", "/item/2", "/item/3", "/zzz"}, "404"},
		{[]string{"/zzz", "/item/4", "/a"}, "A"},
		{[]string{"/a", "/b", "/a", "/b"}, "B"},
		{[]string{"/item/5", "/item/5?tab=x", "/item/5"}, "ITEM={true 5 }"},
	}
	for _, o := range r2mOrigins {
		for _, workers := range []int{8, 1} {
			t.Run(fmt.Sprintf("%s/%d", o.name, workers), func(t *testing.T) {
				h := newR2HarnessN(t, workers)
				loc := o.make(h, "/a")
				h.render(r2Slot("m", loc.Route(
					RouteModel[r2PathA](r2Source[r2PathA](h.rec, "A", nil)),
					RouteModel[r2PathB](r2Source[r2PathB](h.rec, "B", nil)),
					RouteModel[r2PathItem](r2Bind[r2PathItem](h.rec, "ITEM")),
					r2Comp[Location](h.rec, "404"),
				)))
				for i, sc := range scripts {
					h.settle(func(ctx context.Context) {
						for _, uri := range sc.uris {
							loc.Update(ctx, r2Loc(uri))
						}
					})
					at := fmt.Sprintf("script %d: ", i)
					h.expectAt(at, "m", sc.want)
					h.oneDoorAt(at, "m")
					r2mLocIs(h, loc, sc.uris[len(sc.uris)-1])(at)
				}
			})
		}
	}
}

// Old M01 (TestPageStatic): a model route renders its branch when the
// initial Router location decodes into the model.
func TestRoute2ModelPortPageStatic(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/a")
	static := RouteSource(func(Source[r2PathA]) gox.Elem { return textElem("a") })
	h.render(
		r2Slot("router", r2mRouterElem(func(l Source[Location]) gox.Elem { return l.Route(RouteModel[r2PathA](static)) })),
		r2Slot("pkg", Route(RouteModel[r2PathA](static))),
	)
	r2mRun(h,
		r2mStep{at: "initial", want: map[string]string{"router": "a", "pkg": "a"}},
		r2mStep{at: "/b", op: r2mGo(Router(h.page), "/b"), want: map[string]string{"router": "", "pkg": ""}},
		r2mStep{at: "/a", op: r2mGo(Router(h.page), "/a"), want: map[string]string{"router": "a", "pkg": "a"}},
	)
}

// Old M02 (TestRouterBeamModelTransitionsAndDefault404): read-only model
// routes switch A to B to A, each leaf's Beam holds the decoded model, and an
// unmatched location selects the default.
func TestRoute2ModelPortBeamTransitions(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/cross-a")
	abox, bbox := &r2Box[Beam[r2mCrossA]]{}, &r2Box[Beam[r2mCrossB]]{}
	routes := func(r *r2Rec, abox *r2Box[Beam[r2mCrossA]], bbox *r2Box[Beam[r2mCrossB]]) []BeamRoute[Location] {
		return []BeamRoute[Location]{
			RouteBeamModel[r2mCrossA](r2Beam(r, "cross-a", abox)),
			RouteBeamModel[r2mCrossB](r2Beam(r, "cross-b", bbox)),
			RouteComp[Location](r2Leaf{r, "default"}),
		}
	}
	rb := newR2Rec()
	h.render(
		r2Slot("route2", r2mRouterElem(func(l Source[Location]) gox.Elem {
			rs := routes(h.rec, abox, bbox)
			out := make([]SourceRoute[Location], len(rs))
			for i, r := range rs {
				out[i] = r
			}
			return l.Route(out...)
		})),
		r2Slot("beam2", r2mRouterElem(func(l Source[Location]) gox.Elem { return l.RouteBeam(routes(rb, nil, nil)...) })),
	)
	router := Router(h.page)
	w := func(s string) map[string]string { return map[string]string{"route2": s, "beam2": s} }
	r2mRun(h,
		r2mStep{at: "initial", want: w("cross-a"), check: func(at string) {
			if !abox.last().Get().Path {
				t.Fatalf("%scross-a beam does not hold the decoded model", at)
			}
		}},
		r2mStep{at: "cross-next", op: r2mGo(router, "/cross-b"), want: w("cross-b"), check: func(at string) {
			if !bbox.last().Get().Path {
				t.Fatalf("%scross-b beam does not hold the decoded model", at)
			}
		}},
		r2mStep{at: "cross-prev", op: r2mGo(router, "/cross-a"), want: w("cross-a")},
		r2mStep{at: "missing", op: r2mGo(router, "/missing"), want: w("default"),
			counts: map[string]int{"render:cross-a": 2, "render:cross-b": 1, "render:default": 1}},
	)
	r2mWantPaths(h, r2mSP("/cross-b", false), r2mSP("/cross-a", false), r2mSP("/missing", false))
}

// Old M03 (TestRouterSourceCrossRouteKeepsInstance): writable model routes
// switch A to B, back to A renders A again, and /missing selects the default.
func TestRoute2ModelPortSourceCross(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/cross-a")
	h.render(r2Slot("m", Route(
		RouteModel[r2mCrossA](r2Source[r2mCrossA](h.rec, "cross-a", nil)),
		RouteModel[r2mCrossB](r2Source[r2mCrossB](h.rec, "cross-b", nil)),
		RouteComp[Location](r2Leaf{h.rec, "default"}),
	)))
	r2Run(h, "m", Router(h.page), "cross-a",
		r2Step[Location]{set: r2Loc("/cross-b"), want: "cross-b"},
		r2Step[Location]{set: r2Loc("/cross-a"), want: "cross-a", counts: map[string]int{"render:cross-a": 2, "clean:cross-a": 1, "clean:cross-b": 1}},
		r2Step[Location]{set: r2Loc("/missing"), want: "default"},
	)
}

// Old M08 (TestBrowserBackRestoresQueryWithoutReload): query-only changes
// keep the model branch while the optional pointer fields go nil, set, nil.
func TestRoute2ModelPortQueryOnly(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/q")
	box := &r2Box[Source[r2mQuery]]{}
	h.render(
		r2Slot("s", Route(RouteModel[r2mQuery](r2Source(h.rec, "Q", box)))),
		r2Slot("b", Route(RouteModel[r2mQuery](r2mBindF(h.rec, "query", r2mQueryText)))),
	)
	fields := func(tag string, page int) func(string) {
		return func(at string) {
			m := box.last().Get()
			if tag == "" {
				if m.Tag != nil || m.Page != nil {
					t.Fatalf("%sexpected nil query fields, got %+v", at, m)
				}
				return
			}
			if m.Tag == nil || *m.Tag != tag || m.Page == nil || *m.Page != page {
				t.Fatalf("%sexpected tag %q page %d, got %+v", at, tag, page, m)
			}
		}
	}
	router := Router(h.page)
	r2mRun(h,
		r2mStep{at: "initial", want: map[string]string{"s": "Q", "b": "query||"}, check: fields("", 0)},
		r2mStep{at: "query-next", op: r2mGo(router, "/q?tag=next&page=2"), quiet: []string{"s"},
			want: map[string]string{"s": "Q", "b": "query|next|2"}, check: fields("next", 2)},
		r2mStep{at: "unknown param", op: r2mGo(router, "/q?tag=next&page=2&x=1"), quiet: []string{"s", "b"},
			want: map[string]string{"s": "Q", "b": "query|next|2"}, check: fields("next", 2)},
		r2mStep{at: "back", op: r2mGo(router, "/q"), quiet: []string{"s"},
			want: map[string]string{"s": "Q", "b": "query||"}, check: fields("", 0),
			counts: map[string]int{"render:Q": 1, "render:query": 3}},
	)
}

// Old M09 (TestPathModelEscapedSegmentDecodeAndEncode): an escaped segment
// decodes into the model and a written model encodes with escapes.
func TestRoute2ModelPortEscaped(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/escaped/hello%20world%2Fagain")
	box := &r2Box[Source[r2mEscaped]]{}
	h.render(r2Slot("m", Route(RouteModel[r2mEscaped](RouteDerive(r2mCapture(h.rec, "cap", box),
		r2mBindF(h.rec, "name", func(m r2mEscaped) string { return m.Name }))))))
	next := r2mEscaped{Path: true, Name: "next value/again"}
	if l, err := NewLocation(next); err != nil || l.String() != "/escaped/next%20value%2Fagain" {
		t.Fatalf("NewLocation gave %q, %v", l.String(), err)
	}
	r2mRun(h,
		r2mStep{at: "initial", want: map[string]string{"m": "hello world/again"}},
		r2mStep{at: "write", op: r2mPutBox(box, next), ret: []error{nil}, want: map[string]string{"m": "next value/again"}, check: func(at string) {
			if got := Router(h.page).Get().Path(); got != "/escaped/next%20value%2Fagain" {
				t.Fatalf("%spath %q", at, got)
			}
		}},
	)
	r2mWantPaths(h, r2mSP("/escaped/next%20value%2Fagain", false))
}

// Old M10 (TestLocations): a list of model routes selects the decoded model;
// a multi-variant model keeps its branch between variants while the flags
// update.
func TestRoute2ModelPortLocations(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/a")
	box := &r2Box[Source[r2mPathC]]{}
	cText := func(m r2mPathC) string {
		if m.C1 {
			return "c1"
		}
		return "c2"
	}
	h.render(
		r2Slot("s", Route(
			RouteModel[r2PathA](r2Source[r2PathA](h.rec, "A", nil)),
			RouteModel[r2mPathC](r2Source(h.rec, "C", box)),
			RouteModel[r2PathB](r2Source[r2PathB](h.rec, "B", nil)),
		)),
		r2Slot("b", Route(
			RouteModel[r2PathA](r2mBindF(h.rec, "a", func(r2PathA) string { return "a" })),
			RouteModel[r2mPathC](r2mBindF(h.rec, "c", cText)),
			RouteModel[r2PathB](r2mBindF(h.rec, "b", func(r2PathB) string { return "b" })),
		)),
	)
	router := Router(h.page)
	w := func(s, b string) map[string]string { return map[string]string{"s": s, "b": b} }
	r2mRun(h,
		r2mStep{at: "initial", want: w("A", "a")},
		r2mStep{at: "assign c1", op: r2mGo(router, "/c1"), want: w("C", "c1")},
		r2mStep{at: "c2 via model", op: r2mPutBox(box, r2mPathC{C2: true}), ret: []error{nil}, quiet: []string{"s"}, want: w("C", "c2"),
			check: r2mLocIs(h, router, "/c2")},
		r2mStep{at: "c1", op: r2mGo(router, "/c1"), quiet: []string{"s"}, want: w("C", "c1")},
		r2mStep{at: "b", op: r2mGo(router, "/b"), want: w("B", "b")},
		r2mStep{at: "back a", op: r2mGo(router, "/a"), want: w("A", "a")},
		r2mStep{at: "none", op: r2mGo(router, "/zzz"), want: w("", "")},
		r2mStep{at: "c2", op: r2mGo(router, "/c2"), want: w("C", "c2"), counts: map[string]int{"render:C": 2, "render:A": 2, "render:B": 1}},
	)
	r2mWantPaths(h, r2mSP("/c1", false), r2mSP("/c2", false), r2mSP("/c1", false), r2mSP("/b", false),
		r2mSP("/a", false), r2mSP("/zzz", false), r2mSP("/c2", false))
}

func r2mReplaceDoc(h *r2Harness, box *r2Box[Source[r2mReplace]]) {
	h.render(
		r2Slot("s", Route(RouteModel[r2mReplace](r2Source(h.rec, "rep", box)))),
		r2Slot("b", Route(RouteModel[r2mReplace](r2mBindF(h.rec, "step", func(p r2mReplace) string {
			return "step=" + strconv.Itoa(int(p.Step))
		})))),
	)
}

// Old M11 (TestHistoryReplaceContextPathModel): model Source updates push,
// HistoryReplaceContext updates replace; the enum-path branch stays.
func TestRoute2ModelPortHistoryReplaceModel(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/rep-a")
	box := &r2Box[Source[r2mReplace]]{}
	r2mReplaceDoc(h, box)
	w := func(step string) map[string]string { return map[string]string{"s": "rep", "b": "step=" + step} }
	r2mRun(h,
		r2mStep{at: "initial", want: w("0")},
		r2mStep{at: "pm-push", op: r2mPutBox(box, r2mReplace{Step: 1}), ret: []error{nil}, quiet: []string{"s"}, want: w("1"),
			check: r2mLocIs(h, Router(h.page), "/rep-b")},
		r2mStep{at: "pm-replace", op: r2mReplaceCtx(r2mPutBox(box, r2mReplace{Step: 2})), ret: []error{nil}, quiet: []string{"s"}, want: w("2"),
			check: r2mLocIs(h, Router(h.page), "/rep-c")},
		r2mStep{at: "back", op: r2mGo(Router(h.page), "/rep-a"), quiet: []string{"s"}, want: w("0"), counts: map[string]int{"render:rep": 1}},
	)
	r2mWantPaths(h, r2mSP("/rep-b", false), r2mSP("/rep-c", true), r2mSP("/rep-a", false))
}

// Old M12 (TestHistoryReplaceContextMutateAndNoLeak): Mutate with
// HistoryReplaceContext replaces and the next plain update pushes again.
func TestRoute2ModelPortHistoryReplaceMutate(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/rep-a")
	box := &r2Box[Source[r2mReplace]]{}
	r2mReplaceDoc(h, box)
	w := func(step string) map[string]string { return map[string]string{"s": "rep", "b": "step=" + step} }
	r2mRun(h,
		r2mStep{at: "pm-push", op: r2mPutBox(box, r2mReplace{Step: 1}), ret: []error{nil}, quiet: []string{"s"}, want: w("1")},
		r2mStep{at: "pm-replace-mutate", op: func(ctx context.Context) <-chan error {
			return box.last().Mutate(HistoryReplaceContext(ctx), func(p r2mReplace) r2mReplace { p.Step = 2; return p })
		}, ret: []error{nil}, quiet: []string{"s"}, want: w("2"), check: r2mLocIs(h, Router(h.page), "/rep-c")},
		r2mStep{at: "pm-push again", op: r2mPutBox(box, r2mReplace{Step: 1}), ret: []error{nil}, quiet: []string{"s"}, want: w("1")},
	)
	r2mWantPaths(h, r2mSP("/rep-b", false), r2mSP("/rep-c", true), r2mSP("/rep-b", false))
}

// Old M13 (TestHistoryReplaceContext): plain and HistoryReplaceContext Router
// updates re-render a raw Location bind, and set_path carries the flag.
func TestRoute2ModelPortHistoryReplaceRaw(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/a")
	h.render(r2Slot("m", Route(r2mBindF(h.rec, "path", func(l Location) string { return l.String() }))))
	locC1, err1 := NewLocation(r2mPathC{C1: true})
	locC2, err2 := NewLocation(r2mPathC{C2: true})
	if err1 != nil || err2 != nil {
		t.Fatalf("NewLocation: %v %v", err1, err2)
	}
	router := Router(h.page)
	r2mRun(h,
		r2mStep{at: "initial", want: map[string]string{"m": "/a"}},
		r2mStep{at: "soft-assign", op: r2mPut(router, locC1), ret: []error{nil}, want: map[string]string{"m": "/c1"}},
		r2mStep{at: "soft-replace", op: r2mReplaceCtx(r2mPut(router, locC2)), ret: []error{nil}, want: map[string]string{"m": "/c2"}},
		r2mStep{at: "back", op: r2mGo(router, "/a"), ret: []error{nil}, want: map[string]string{"m": "/a"}, counts: map[string]int{"render:path": 4}},
	)
	r2mWantPaths(h, r2mSP("/c1", false), r2mSP("/c2", true), r2mSP("/a", false))
}

// Old M14 (TestLocationModelMatchesAnyURL): a default leaf over Location
// sees any URL with its path and query.
func TestRoute2ModelPortAnyURL(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/x?tag=a&page=3")
	box := &r2Box[Source[Location]]{}
	h.render(
		r2Slot("b", Route(r2mBindF(h.rec, "loc", func(l Location) string {
			return l.String() + "|" + l.Path() + "|" + l.Query.Get("tag") + "|" + l.Query.Get("page")
		}))),
		r2Slot("s", Route(r2Source(h.rec, "S", box))),
	)
	router := Router(h.page)
	steps := []r2mStep{{at: "initial", want: map[string]string{"b": "/x?page=3&tag=a|/x|a|3", "s": "S"}}}
	for _, st := range []struct{ uri, want string }{
		{"/some/deep/path?tag=t&page=9", "/some/deep/path?page=9&tag=t|/some/deep/path|t|9"},
		{"/hello%20world", "/hello%20world|/hello%20world||"},
		{"/", "/|/||"},
		{"/q?tag=only", "/q?tag=only|/q|only|"},
	} {
		uri := st.uri
		steps = append(steps, r2mStep{at: uri, op: r2mGo(router, uri), quiet: []string{"s"}, want: map[string]string{"b": st.want, "s": "S"},
			check: func(at string) { r2mLocIs(h, box.last(), uri)(at) }})
	}
	r2mRun(h, steps...)
	h.counts(map[string]int{"render:S": 1, "render:loc": 5})
}

// Old M15 (doors.Route): the package-level Route behaves like
// Router(ctx).Route inside an Elem.
func TestRoute2ModelPortPackageRoute2(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/item/1")
	rp, rr := newR2Rec(), newR2Rec()
	tree := func(r *r2Rec) []SourceRoute[Location] {
		return []SourceRoute[Location]{
			RouteModel[r2PathA](r2Source[r2PathA](r, "A", nil)),
			RouteModel[r2PathItem](r2Bind[r2PathItem](r, "I")),
			r2Comp[Location](r, "404"),
		}
	}
	h.render(
		r2Slot("pkg", Route(tree(rp)...)),
		r2Slot("elem", r2mRouterElem(func(l Source[Location]) gox.Elem { return l.Route(tree(rr)...) })),
	)
	router := Router(h.page)
	steps := []r2mStep{{at: "initial", want: map[string]string{"pkg": "I={true 1 }", "elem": "I={true 1 }"}}}
	for _, st := range []struct{ uri, want string }{
		{"/a", "A"}, {"/item/2?tab=q", "I={true 2 q}"}, {"/zzz", "404"}, {"/a?x=1", "A"}, {"/item/3", "I={true 3 }"},
	} {
		steps = append(steps, r2mStep{at: st.uri, op: r2mGo(router, st.uri), want: map[string]string{"pkg": st.want, "elem": st.want}})
	}
	r2mRun(h, steps...)
	for _, k := range []string{"render:A", "render:I", "render:404", "clean:A", "clean:I", "clean:404"} {
		if rp.count(k) != rr.count(k) {
			t.Fatalf("%s: package %d, elem %d", k, rp.count(k), rr.count(k))
		}
	}
}

// Old M16-M18 (TestRouteSourceAndDerivedBeamEntrypoints): routes on a
// derived Source and a derived Beam; branch writes flow back to the root.
func TestRoute2ModelPortEntrypoints(t *testing.T) {
	h := newR2Harness(t)
	root := NewSource(r2mState{Int: 1, Str: "a"})
	str := DeriveSource(root, func(s r2mState) string { return s.Str }, func(s r2mState, v string) r2mState { s.Str = v; return s })
	parity := DeriveBeam(root, func(s r2mState) string {
		if s.Int%2 == 0 {
			return "even"
		}
		return "odd"
	})
	nonEmpty := func(v string) bool { return v != "" }
	simple := func(v string) bool { return strings.HasPrefix(v, "simple") }
	idLens := func(s Source[string]) Source[string] {
		return DeriveSource(s, func(v string) string { return v }, func(_ string, v string) string { return v })
	}
	lensBox, simpleBox := &r2Box[Source[string]]{}, &r2Box[Source[string]]{}
	h.render(
		r2Slot("lens", str.Route(RouteMatch(nonEmpty, RouteDerive(r2Derive(h.rec, "lens", idLens), r2Source(h.rec, "lens", lensBox))))),
		r2Slot("lensv", str.Route(RouteMatch(nonEmpty, RouteDerive(idLens, r2mBindF(h.rec, "lensv", func(v string) string { return "lens:" + v }))))),
		r2Slot("lensbeam", str.RouteBeam(RouteBeamMatch(nonEmpty, RouteBeamDerive(func(b Beam[string]) Beam[string] {
			return DeriveBeam(b, func(v string) string { return v })
		}, r2mBindF(h.rec, "lensbeam", func(v string) string { return "beam:" + v }))))),
		r2Slot("derived", parity.RouteBeam(
			RouteBeamValue("odd", r2mBindF(h.rec, "derived", func(v string) string { return "derived:" + v })),
			r2mBindF(h.rec, "derived-default", func(v string) string { return "derived-default:" + v }),
		)),
		r2Slot("derivedm", parity.RouteBeam(
			RouteBeamMatch(func(v string) bool { return v == "odd" }, r2Beam[string](h.rec, "derived-match", nil)),
			r2Beam[string](h.rec, "derived-match-default", nil),
		)),
		r2Slot("simplelens", str.Route(RouteMatch(simple, RouteDerive(r2mCapture(h.rec, "simple", simpleBox),
			r2mBindF(h.rec, "simplelens", func(v string) string { return "simple-lens:" + v }))))),
		r2Slot("simplebeam", str.RouteBeam(RouteBeamMatch(simple, r2mBindF(h.rec, "simplebeam", func(v string) string { return "simple-beam:" + v })))),
	)
	rootIs := func(want r2mState) func(string) {
		return func(at string) {
			if got := root.Get(); got != want {
				t.Fatalf("%sroot holds %+v, want %+v", at, got, want)
			}
		}
	}
	strs := []string{"lens", "lensv", "lensbeam", "simplelens", "simplebeam"}
	r2mRun(h,
		r2mStep{at: "initial", want: map[string]string{
			"lens": "lens", "lensv": "lens:a", "lensbeam": "beam:a", "derived": "derived:odd", "derivedm": "derived-match",
			"simplelens": "", "simplebeam": "",
		}},
		r2mStep{at: "route-entry-lens-update", op: r2mPutBox(lensBox, "b"), ret: []error{nil}, quiet: []string{"lens", "derived", "derivedm"},
			want:  map[string]string{"lens": "lens", "lensv": "lens:b", "lensbeam": "beam:b", "simplelens": "", "simplebeam": ""},
			check: rootIs(r2mState{Int: 1, Str: "b"})},
		r2mStep{at: "route-entry-set-simple", op: r2mPut(str, "simple"), ret: []error{nil}, quiet: []string{"lens", "derived"},
			want:  map[string]string{"lensv": "lens:simple", "lensbeam": "beam:simple", "simplelens": "simple-lens:simple", "simplebeam": "simple-beam:simple"},
			check: rootIs(r2mState{Int: 1, Str: "simple"})},
		r2mStep{at: "route-entry-simple-lens-update", op: r2mPutBox(simpleBox, "simple-lens"), ret: []error{nil}, quiet: []string{"lens"},
			want:  map[string]string{"lensv": "lens:simple-lens", "simplelens": "simple-lens:simple-lens", "simplebeam": "simple-beam:simple-lens"},
			check: rootIs(r2mState{Int: 1, Str: "simple-lens"})},
		r2mStep{at: "route-entry-even", op: r2mPut(root, r2mState{Int: 2, Str: "simple-lens"}), ret: []error{nil}, quiet: strs,
			want: map[string]string{"derived": "derived-default:even", "derivedm": "derived-match-default", "lensv": "lens:simple-lens"}},
		r2mStep{at: "odd again", op: r2mPut(root, r2mState{Int: 3, Str: "simple-lens"}), ret: []error{nil}, quiet: strs,
			want: map[string]string{"derived": "derived:odd", "derivedm": "derived-match"}},
		r2mStep{at: "clear", op: r2mPut(str, ""), ret: []error{nil}, quiet: []string{"derived", "derivedm"},
			want: map[string]string{"lens": "", "lensv": "", "lensbeam": "", "simplelens": "", "simplebeam": ""},
			counts: map[string]int{"derive:lens": 1, "derive:simple": 1, "render:lens": 1, "render:derived": 2, "render:derived-default": 1,
				"render:derived-match": 2, "render:derived-match-default": 1}},
	)
}

// Old M19 (RouteModel adapter error): a model type that is not a valid path
// model never matches and the default renders instead.
func TestRoute2ModelPortInvalidModel(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/a")
	h.render(
		r2Slot("s", Route(
			RouteModel[int](r2Source[int](h.rec, "INT", nil)),
			RouteModel[r2PathBad](r2Source[r2PathBad](h.rec, "BAD", nil)),
			r2Comp[Location](h.rec, "def"),
		)),
		r2Slot("b", r2mRouterElem(func(l Source[Location]) gox.Elem {
			return l.RouteBeam(
				RouteBeamModel[r2PathBad](r2Beam[r2PathBad](h.rec, "BBAD", nil)),
				RouteBeamModel[int](r2Beam[int](h.rec, "BINT", nil)),
				r2Comp[Location](h.rec, "bdef"),
			)
		})),
	)
	r2mRun(h,
		r2mStep{at: "initial", want: map[string]string{"s": "def", "b": "bdef"}},
		r2mStep{at: "/b", op: r2mGo(Router(h.page), "/b"), quiet: []string{"s", "b"}, want: map[string]string{"s": "def", "b": "bdef"},
			counts: map[string]int{"render:INT": 0, "render:BAD": 0, "render:BBAD": 0, "render:BINT": 0, "render:def": 1, "render:bdef": 1}},
	)
}

// Old M20 (RouteModel encode error): writing PathC{} with no variant
// selected leaves the Location unchanged, sends no set_path, keeps the branch.
func TestRoute2ModelPortEncodeErrorNoSetPath(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/c1")
	box := &r2Box[Source[r2mPathC]]{}
	h.render(
		r2Slot("s", Route(RouteModel[r2mPathC](r2Source(h.rec, "C", box)))),
		r2Slot("b", Route(RouteModel[r2mPathC](r2Bind[r2mPathC](h.rec, "c")))),
	)
	router := Router(h.page)
	r2mRun(h,
		r2mStep{at: "initial", want: map[string]string{"s": "C", "b": "c={true false}"}},
		r2mStep{at: "zero", op: r2mPutBox(box, r2mPathC{}), ret: []error{}, quiet: []string{"s", "b"},
			want: map[string]string{"s": "C", "b": "c={true false}"}, check: r2mLocIs(h, router, "/c1")},
		r2mStep{at: "c2", op: r2mPutBox(box, r2mPathC{C2: true}), ret: []error{nil}, quiet: []string{"s"},
			want: map[string]string{"s": "C", "b": "c={false true}"}, check: r2mLocIs(h, router, "/c2")},
	)
	r2mWantPaths(h, r2mSP("/c2", false))
}

// Old M21 (Router godoc): an equal Location is suppressed: no route
// callback, no re-render, no set_path.
func TestRoute2ModelPortEqualLocation(t *testing.T) {
	h := newR2Harness(t)
	r2mRouterAt(h, "/a?x=1&y=2")
	isA := func(l Location) bool { return len(l.Segments) == 1 && l.Segments[0] == "a" }
	h.render(r2Slot("m", Route(
		RouteMatch(r2Pred(h.rec, "p", isA), r2Bind[Location](h.rec, "B")),
		r2Comp[Location](h.rec, "D"),
	)))
	router := Router(h.page)
	r2mRun(h,
		r2mStep{at: "initial", want: map[string]string{"m": "B=/a?x=1&y=2"}},
		r2mStep{at: "equal", op: r2mGo(router, "/a?y=2&x=1"), ret: []error{}, quiet: []string{"m"},
			want: map[string]string{"m": "B=/a?x=1&y=2"}, counts: map[string]int{"render:B": 1}},
		r2mStep{at: "equal slash", op: r2mGo(router, "/a/?x=1&y=2"), ret: []error{}, quiet: []string{"m"},
			want: map[string]string{"m": "B=/a?x=1&y=2"}, counts: map[string]int{"render:B": 1}},
		r2mStep{at: "change", op: r2mGo(router, "/a?x=1"), ret: []error{nil}, want: map[string]string{"m": "B=/a?x=1"}, counts: map[string]int{"render:B": 2}},
	)
	if got, want := h.rec.vals("pred:p="), []string{"/a?x=1&y=2", "/a?x=1"}; !slices.Equal(got, want) {
		t.Fatalf("predicate log %v, want %v", got, want)
	}
	r2mWantPaths(h, r2mSP("/a?x=1", false))
}

// Model routing on comparable sources made with NewSource, DeriveSource and
// DeriveBeam over a Location: value, match and bind routes over the model,
// and writes through the derived entry re-encode the Location.
func TestRoute2ModelComparableEntries(t *testing.T) {
	adapter, err := path.GetModelAdapter[r2PathItem]()
	if err != nil {
		t.Fatal(err)
	}
	decode := func(l Location) r2PathItem {
		m, ok := adapter.Decode(l)
		if !ok {
			return r2PathItem{}
		}
		return *m
	}
	encode := func(prev Location, m r2PathItem) Location {
		l, err := adapter.Encode(&m)
		if err != nil {
			return prev
		}
		return l
	}
	tree := func(r *r2Rec, box *r2Box[Source[r2PathItem]]) []SourceRoute[r2PathItem] {
		return []SourceRoute[r2PathItem]{
			RouteValue(r2PathItem{}, r2Comp[r2PathItem](r, "NONE")),
			RouteValue(r2mItem1, r2Source(r, "ONE", box)),
			RouteMatch(func(m r2PathItem) bool { return m.ID >= 10 }, r2Bind[r2PathItem](r, "BIG")),
			r2Bind[r2PathItem](r, "ITEM"),
		}
	}
	h := newR2Harness(t)
	loc := r2LocSource("/item/1")
	model := NewSource(r2mItem1)
	derived := DeriveSource(loc, decode, encode)
	box, dbox := &r2Box[Source[r2PathItem]]{}, &r2Box[Source[r2PathItem]]{}
	h.render(
		r2Slot("new", model.Route(tree(newR2Rec(), box)...)),
		r2Slot("newbeam", model.RouteBeam(
			RouteBeamValue(r2PathItem{}, r2Comp[r2PathItem](newR2Rec(), "NONE")),
			RouteBeamValue(r2mItem1, r2Comp[r2PathItem](newR2Rec(), "ONE")),
			RouteBeamMatch(func(m r2PathItem) bool { return m.ID >= 10 }, r2Bind[r2PathItem](newR2Rec(), "BIG")),
			r2Bind[r2PathItem](newR2Rec(), "ITEM"),
		)),
		r2Slot("derived", derived.Route(tree(h.rec, dbox)...)),
		r2Slot("derivedbeam", DeriveBeam(loc, decode).RouteBeam(
			RouteBeamValue(r2PathItem{}, r2Comp[r2PathItem](newR2Rec(), "NONE")),
			RouteBeamValue(r2mItem1, r2Comp[r2PathItem](newR2Rec(), "ONE")),
			RouteBeamMatch(func(m r2PathItem) bool { return m.ID >= 10 }, r2Bind[r2PathItem](newR2Rec(), "BIG")),
			r2Bind[r2PathItem](newR2Rec(), "ITEM"),
		)),
		r2Slot("derivedsrcbeam", derived.RouteBeam(
			RouteBeamValue(r2PathItem{}, r2Comp[r2PathItem](newR2Rec(), "NONE")),
			RouteBeamValue(r2mItem1, r2Comp[r2PathItem](newR2Rec(), "ONE")),
			RouteBeamMatch(func(m r2PathItem) bool { return m.ID >= 10 }, r2Bind[r2PathItem](newR2Rec(), "BIG")),
			r2Bind[r2PathItem](newR2Rec(), "ITEM"),
		)),
	)
	slots := []string{"new", "newbeam", "derived", "derivedbeam", "derivedsrcbeam"}
	all := func(want string) map[string]string {
		out := map[string]string{}
		for _, s := range slots {
			out[s] = want
		}
		return out
	}
	both := func(uri string, m r2PathItem) func(ctx context.Context) <-chan error {
		return func(ctx context.Context) <-chan error {
			model.Update(ctx, m)
			return loc.Update(ctx, r2Loc(uri))
		}
	}
	r2mRun(h,
		r2mStep{at: "initial", want: all("ONE")},
		r2mStep{at: "2", op: both("/item/2", r2PathItem{Item: true, ID: 2}), want: all("ITEM={true 2 }")},
		r2mStep{at: "12", op: both("/item/12", r2PathItem{Item: true, ID: 12}), want: all("BIG={true 12 }")},
		r2mStep{at: "1 after big", op: both("/item/1", r2mItem1), want: all("ONE")},
		r2mStep{at: "none", op: both("/zzz", r2PathItem{}), want: all("NONE")},
		r2mStep{at: "1", op: both("/item/1", r2mItem1), want: all("ONE")},
		r2mStep{at: "write derived", op: r2mPutBox(dbox, r2PathItem{Item: true, ID: 3, Tab: "w"}), ret: []error{nil},
			want:  map[string]string{"new": "ONE", "newbeam": "ONE", "derived": "ITEM={true 3 w}", "derivedbeam": "ITEM={true 3 w}", "derivedsrcbeam": "ITEM={true 3 w}"},
			check: r2mLocIs(h, loc, "/item/3?tab=w")},
		r2mStep{at: "write model", op: r2mPutBox(box, r2PathItem{Item: true, ID: 3, Tab: "w"}), ret: []error{nil},
			want: all("ITEM={true 3 w}"), counts: map[string]int{"render:ONE": 3, "render:NONE": 1, "render:BIG": 1}},
	)
}

func r2mDeepTree(f r2mFam, r *r2Rec, top ...SourceRoute[r2PathItem]) []SourceRoute[Location] {
	adminItem := []SourceRoute[r2PathItem]{r2Bind[r2PathItem](r, "ADMIN-ITEM")}
	if len(top) != 0 {
		adminItem = top
	}
	return []SourceRoute[Location]{
		r2mMatch(f, r2mFirst("app"), r2mStrip(f, r, "s1", "app",
			r2mMatch(f, r2mFirst("admin"), r2mStrip(f, r, "s2", "admin",
				r2mModel[r2PathItem](f, adminItem...),
				r2mModel[r2PathA](f, r2mOnce[r2PathA](f, r, "ADMIN-A")),
				r2Comp[Location](r, "admin404"),
			)),
			r2mModel[r2PathItem](f,
				r2mValue(f, r2mItem1, r2mOnce[r2PathItem](f, r, "APP-ONE")),
				r2Bind[r2PathItem](r, "APP-ITEM"),
			),
			r2mModel[r2PathA](f, r2mOnce[r2PathA](f, r, "APP-A")),
			r2Comp[Location](r, "app404"),
		)),
		r2mModel[r2PathA](f, r2mOnce[r2PathA](f, r, "A")),
		r2Comp[Location](r, "404"),
	}
}

// Two prefix lenses deep, with models, values, defaults and stickiness at
// every level, all applied to the one route door, for every entry point.
func TestRoute2ModelDeepLens(t *testing.T) {
	r2mFamily(t, r2mCase{
		name: "app-admin",
		uri:  "/app/item/1",
		routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
			return r2mDeepTree(f, r)
		},
		initial: "APP-ONE",
		steps: []r2mLocStep{
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
			{uri: "/app/item/7?q=1&tab=x", want: "APP-ITEM={true 7 x}", quiet: true, counts: map[string]int{
				"derive:s1": 1, "derive:s2": 1, "render:APP-ONE": 2, "render:ADMIN-A": 1, "render:admin404": 2, "render:app404": 1,
				"render:A": 1, "render:404": 1, "render:ADMIN-ITEM": 2, "render:APP-ITEM": 3, "render:APP-A": 1,
			}},
		},
	})
}

// Writes from two lens levels deep restore both prefixes.
func TestRoute2ModelDeepLensWriteBack(t *testing.T) {
	for _, o := range r2mOrigins {
		t.Run(o.name, func(t *testing.T) {
			h := newR2Harness(t)
			loc := o.make(h, "/app/admin/item/3")
			box := &r2Box[Source[r2PathItem]]{}
			src := r2mFam{"source", false}
			h.render(r2Slot("m", loc.Route(r2mDeepTree(src, h.rec,
				RouteDerive(r2mCapture(h.rec, "cap", box), r2Bind[r2PathItem](h.rec, "ADMIN-ITEM")))...)))
			m := func(s string) map[string]string { return map[string]string{"m": s} }
			r2mRun(h,
				r2mStep{at: "initial", want: m("ADMIN-ITEM={true 3 }")},
				r2mStep{at: "write", op: r2mPutBox(box, r2PathItem{Item: true, ID: 9, Tab: "t"}), ret: []error{nil},
					want: m("ADMIN-ITEM={true 9 t}"), check: r2mLocIs(h, loc, "/app/admin/item/9?tab=t")},
				r2mStep{at: "equal", op: r2mPutBox(box, r2PathItem{Item: true, ID: 9, Tab: "t"}), ret: []error{}, quiet: []string{"m"},
					want: m("ADMIN-ITEM={true 9 t}")},
				r2mStep{at: "leave", op: r2mGo(loc, "/app/item/1"), want: m("APP-ONE")},
				r2mStep{at: "write back in", op: r2mPutBox(box, r2PathItem{Item: true, ID: 2, Tab: "u"}), ret: []error{nil},
					want: m("ADMIN-ITEM={true 2 u}"), check: r2mLocIs(h, loc, "/app/admin/item/2?tab=u"),
					counts: map[string]int{"derive:cap": 1, "derive:s1": 1, "derive:s2": 1, "render:ADMIN-ITEM": 3}},
			)
			if o.router {
				r2mWantPaths(h, r2mSP("/app/admin/item/9?tab=t", false), r2mSP("/app/item/1", false), r2mSP("/app/admin/item/2?tab=u", false))
			}
		})
	}
}

// A derive branch over a Location is a default for its parent while its
// active child is a default, at index 0 or later, so earlier models take
// over; once it switches to a non-default child it is sticky again.
func TestRoute2ModelDeriveDefaultIndex(t *testing.T) {
	r2mFamily(t,
		r2mCase{
			name: "index0",
			uri:  "/item/x",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathA](f, r2mOnce[r2PathA](f, r, "A")),
					r2mItemCustom(f, r2Bind[r2Custom](r, "C")),
					r2Comp[Location](r, "404"),
				}
			},
			initial: "C={x }",
			steps: []r2mLocStep{
				{uri: "/a", want: "A"},
				{uri: "/item/y", want: "C={y }"},
				{uri: "/zzz", want: "C={ }"},
				{uri: "/a", want: "A", counts: map[string]int{"render:A": 2, "render:C": 3, "render:404": 0}},
			},
		},
		r2mCase{
			name: "index1",
			uri:  "/item/x",
			routes: func(r *r2Rec, f r2mFam) []SourceRoute[Location] {
				return []SourceRoute[Location]{
					r2mModel[r2PathItem](f, r2mOnce[r2PathItem](f, r, "I")),
					r2mItemCustom(f,
						r2mMatch(f, func(c r2Custom) bool { return c.Tab == "v" }, r2mOnce[r2Custom](f, r, "V")),
						r2Bind[r2Custom](r, "C"),
					),
					r2Comp[Location](r, "404"),
				}
			},
			initial: "C={x }",
			steps: []r2mLocStep{
				{uri: "/item/1", want: "I"},
				{uri: "/item/y", want: "C={y }"},
				{uri: "/item/y?tab=v", want: "V"},
				{uri: "/item/2?tab=v", want: "V", quiet: true},
				{uri: "/item/z", want: "C={z }"},
				{uri: "/item/3", want: "I", counts: map[string]int{"render:I": 2, "render:V": 1, "render:C": 3, "render:404": 0}},
			},
		},
	)
}

func r2xRise(old, new int) bool {
	return new <= old
}

type r2xMount struct {
	name  string
	mount func(src Source[int], rs ...SourceRoute[int]) gox.Elem
	value func(v int, rs ...SourceRoute[int]) SourceRoute[int]
}

var r2xMounts = []r2xMount{
	{"source", func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return src.Route(rs...)
	}, func(v int, rs ...SourceRoute[int]) SourceRoute[int] {
		return RouteValue(v, rs...)
	}},
	{"beam", func(src Source[int], rs ...SourceRoute[int]) gox.Elem {
		return src.RouteBeam(r2Beams(rs)...)
	}, func(v int, rs ...SourceRoute[int]) SourceRoute[int] {
		return RouteBeamValue(v, r2Beams(rs)...)
	}},
}

// RouteBind calls the comparator as (stored, incoming): with a rise-only
// comparator every propagated rise re-renders.
func TestRoute2MutBindComparatorOrder(t *testing.T) {
	for _, m := range r2xMounts {
		t.Run(m.name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSourceEqual(1, r2xRise)
			h.render(r2Slot("a", m.mount(src, r2Bind[int](h.rec, "B"))))
			r2Run(h, "a", src, "B=1",
				r2Step[int]{set: 2, want: "B=2"},
				r2Step[int]{set: 0, want: "B=2", quiet: true},
				r2Step[int]{set: 5, want: "B=5", counts: map[string]int{"render:B": 3}},
			)
		})
	}
}

// A bind below a derive uses the derived comparator in the same order.
func TestRoute2MutBindDerivedComparatorOrder(t *testing.T) {
	h := newR2Harness(t)
	src := NewSource(10)
	h.render(r2Slot("a", src.Route(RouteDerive(func(s Source[int]) Source[int] {
		return DeriveSourceEqual(s, func(v int) int { return v / 10 }, func(_, b int) int { return b * 10 }, r2xRise)
	}, r2Bind[int](h.rec, "T")))))
	r2Run(h, "a", src, "T=1",
		r2Step[int]{set: 25, want: "T=2"},
		r2Step[int]{set: 3, want: "T=2", quiet: true},
		r2Step[int]{set: 40, want: "T=4", counts: map[string]int{"render:T": 3}},
	)
}

// Value routes call the comparator as (route value, current value): with a
// rise-only comparator RouteValue(5) matches every value up to 5.
func TestRoute2MutValueComparatorOrder(t *testing.T) {
	for _, m := range r2xMounts {
		t.Run(m.name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSourceEqual(0, r2xRise)
			h.render(r2Slot("a", m.mount(src,
				m.value(5, r2Comp[int](h.rec, "UPTO5")),
				r2Comp[int](h.rec, "D"),
			)))
			r2Run(h, "a", src, "UPTO5",
				r2Step[int]{set: 5, want: "UPTO5", quiet: true},
				r2Step[int]{set: 6, want: "D"},
			)
		})
	}
}

func r2xTree(k r2Kit, r *r2Rec) []SourceRoute[int] {
	return []SourceRoute[int]{
		k.match(func(v int) bool { return v%2 == 0 }, r2Comp[int](r, "EVEN")),
		k.match(func(v int) bool { return v > 0 && v < 10 }, r2Comp[int](r, "SMALL")),
		k.match(func(v int) bool { return v >= 10 },
			k.derive(r, "tens", func(v int) int { return v / 10 }, func(a, b int) int { return b*10 + a%10 },
				k.value(5, r2Comp[int](r, "FIFTY")),
				r2Bind[int](r, "T"),
			),
		),
		r2Bind[int](r, "D"),
	}
}

var r2xSteps = []struct {
	v    int
	want string
}{
	{2, "SMALL"}, {12, "EVEN"}, {13, "T=1"}, {15, "T=1"}, {33, "T=3"}, {51, "FIFTY"}, {4, "EVEN"},
	{7, "SMALL"}, {9, "SMALL"}, {11, "T=1"}, {-3, "D=-3"}, {-5, "D=-5"}, {-4, "EVEN"},
}

// One Route element value rendered in two slots keeps two route states.
func TestRoute2MutSameElemTwoSlots(t *testing.T) {
	r2Matrix(t, func(t *testing.T, k r2Kit, e r2Entry) {
		h := newR2Harness(t)
		src := NewSource(1)
		el := e.mount(src, r2xTree(k, h.rec)...)
		h.render(r2Slot("a", el), r2Slot("b", el))
		if h.doors0["a"] == h.doors0["b"] {
			t.Fatalf("both slots share door %d", h.doors0["a"])
		}
		h.expect("a", "SMALL")
		h.expect("b", "SMALL")
		for i, st := range r2xSteps {
			at := fmt.Sprintf("step %d (set %d): ", i, st.v)
			r2Set(h, src, st.v)
			h.expectAt(at, "a", st.want)
			h.expectAt(at, "b", st.want)
			h.oneDoorAt(at, "a")
			h.oneDoorAt(at, "b")
		}
		h.counts(map[string]int{"derive:tens": 2})
	})
}

// One Route element value rendered by several page instances on one source
// keeps one route state per instance.
func TestRoute2MutSameElemInstances(t *testing.T) {
	for _, k := range r2Kits {
		t.Run(k.name, func(t *testing.T) {
			rec := newR2Rec()
			src := NewSource(1)
			var el gox.Elem
			if k.beam {
				el = src.RouteBeam(r2Beams(r2xTree(k, rec))...)
			} else {
				el = src.Route(r2xTree(k, rec)...)
			}
			hs := make([]*r2Harness, 4)
			for i := range hs {
				hs[i] = newR2Harness(t)
				hs[i].render(r2Slot("a", el))
				hs[i].expect("a", "SMALL")
			}
			for i, st := range r2xSteps {
				if ret := drainOp(t, src.Update(hs[i%len(hs)].page, st.v)); !r2SameErrs(ret, []error{nil}) {
					t.Fatalf("step %d: update returned %v", i, ret)
				}
				for _, h := range hs {
					h.eventually("a", st.want)
				}
			}
			if n := rec.count("derive:tens"); n != len(hs) {
				t.Fatalf("derive ran %d times, want %d", n, len(hs))
			}
		})
	}
}

// RouteBind renders inside the door render, after the subscription callback
// returns: a slow render does not hold back the update's propagation.
func TestRoute2MutBindRenderInDoor(t *testing.T) {
	for _, m := range r2xMounts {
		t.Run(m.name, func(t *testing.T) {
			h := newR2Harness(t)
			src := NewSource(1)
			gate := make(chan struct{})
			var once sync.Once
			release := func() { once.Do(func() { close(gate) }) }
			defer release()
			h.render(r2Slot("a", m.mount(src, RouteBind(func(v int) gox.Elem {
				if v == 2 {
					<-gate
				}
				return textElem(fmt.Sprint("B=", v))
			}))))
			h.expect("a", "B=1")
			ch := src.Update(h.page, 2)
			select {
			case err := <-ch:
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
}
