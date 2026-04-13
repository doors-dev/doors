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

package printer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

type testComp struct{}

func (testComp) Main() gox.Elem { return nil }

type titleInstance struct {
	title      string
	titleAttrs gox.Attrs
	registry   *resources.Registry
	conf       common.SystemConf
	csp        *common.CSPCollector
	modules    *testModuleRegistry
	metas      []testMetaUpdate
}

func (t *titleInstance) CallCtx(context.Context, action.Action, func(json.RawMessage, error), func(), action.CallParams) context.CancelFunc {
	return func() {}
}
func (t *titleInstance) CallCheck(func() bool, action.Action, func(json.RawMessage, error), func(), action.CallParams) {
}
func (t *titleInstance) CSPCollector() *common.CSPCollector {
	if t.csp != nil {
		return t.csp
	}
	return (&common.CSP{}).NewCollector()
}
func (t *titleInstance) ModuleRegistry() core.ModuleRegistry   { return t.modules }
func (t *titleInstance) ResourceRegistry() *resources.Registry { return t.registry }
func (t *titleInstance) ID() string                            { return "instance" }
func (t *titleInstance) RootID() uint64                        { return 1 }
func (t *titleInstance) Conf() *common.SystemConf              { return &t.conf }
func (t *titleInstance) NewID() uint64                         { return 1 }
func (t *titleInstance) NewLink(any) (core.Link, error)        { return core.Link{}, nil }
func (t *titleInstance) Runtime() shredder.Runtime             { return nil }
func (t *titleInstance) SetStatus(int)                         {}
func (t *titleInstance) SessionExpire(time.Duration)           {}
func (t *titleInstance) SessionEnd()                           {}
func (t *titleInstance) InstanceEnd()                          {}
func (t *titleInstance) SessionID() string                     { return "session" }
func (t *titleInstance) Adapters() path.Adapters               { return nil }
func (t *titleInstance) PathMaker() path.PathMaker             { return path.NewPathMaker("srv") }
func (t *titleInstance) UpdateTitle(content string, attrs gox.Attrs) {
	t.title = content
	t.titleAttrs = attrs
}
func (t *titleInstance) UpdateMeta(name string, property bool, attrs gox.Attrs) {
	t.metas = append(t.metas, testMetaUpdate{name: name, property: property, attrs: attrs})
}

type titleDoor struct{}

func (titleDoor) Cinema() beam.Cinema { return nil }
func (titleDoor) RegisterHook(func(context.Context, http.ResponseWriter, *http.Request) bool, func(context.Context)) (core.Hook, bool) {
	return core.Hook{}, false
}
func (titleDoor) ID() uint64 { return 7 }
func (titleDoor) RootCore() core.Core {
	return nil
}

type testMetaUpdate struct {
	name     string
	property bool
	attrs    gox.Attrs
}

type testModuleRegistry struct {
	values map[string]string
}

func (m *testModuleRegistry) Add(specifier string, path string) {
	if m.values == nil {
		m.values = map[string]string{}
	}
	m.values[specifier] = path
}

type hookDoor struct {
	id        uint64
	allowHook bool
	nextHook  uint64
}

func (d *hookDoor) Cinema() beam.Cinema { return nil }

func (d *hookDoor) RegisterHook(func(context.Context, http.ResponseWriter, *http.Request) bool, func(context.Context)) (core.Hook, bool) {
	if !d.allowHook {
		return core.Hook{}, false
	}
	d.nextHook++
	return core.Hook{DoorID: d.id, HookID: d.nextHook}, true
}

func (d *hookDoor) ID() uint64 { return d.id }
func (d *hookDoor) RootCore() core.Core {
	return nil
}

func newPrinterCore(t *testing.T, allowHook bool) (context.Context, *titleInstance, *hookDoor, *testModuleRegistry) {
	t.Helper()
	conf := common.SystemConf{}
	common.InitDefaults(&conf)
	modules := &testModuleRegistry{}
	inst := &titleInstance{
		registry: resources.NewRegistry(pagePrinterSettings{conf: &conf}),
		conf:     conf,
		csp:      (&common.CSP{}).NewCollector(),
		modules:  modules,
	}
	door := &hookDoor{id: 7, allowHook: allowHook}
	ctx := context.WithValue(context.Background(), ctex.KeyCore, core.NewCore(inst, door))
	return ctx, inst, door, modules
}

type failPrinter struct {
	calls int
	fail  int
}

func (p *failPrinter) Send(gox.Job) error {
	p.calls++
	if p.calls == p.fail {
		return context.Canceled
	}
	return nil
}

type recordedOpen struct {
	tag   string
	attrs gox.Attrs
}

type recordingPrinter struct {
	opens  []recordedOpen
	closes int
	raw    []string
	bytes  [][]byte
}

func (p *recordingPrinter) Send(job gox.Job) error {
	if open, ok := job.(*gox.JobHeadOpen); ok {
		p.opens = append(p.opens, recordedOpen{
			tag:   open.Tag,
			attrs: open.Attrs.Clone(),
		})
	}
	if _, ok := job.(*gox.JobHeadClose); ok {
		p.closes++
	}
	if raw, ok := job.(*gox.JobRaw); ok {
		p.raw = append(p.raw, raw.Text)
	}
	if bytesJob, ok := job.(*gox.JobBytes); ok {
		p.bytes = append(p.bytes, append([]byte(nil), bytesJob.Bytes...))
	}
	return nil
}

func TestResourceEntryHelpers(t *testing.T) {
	t.Run("script empty", func(t *testing.T) {
		r := &embeddedResource{}
		if r.scriptEntry() != nil {
			t.Fatal("empty script entry should be nil")
		}
	})

	t.Run("script string and bytes", func(t *testing.T) {
		r := &embeddedResource{}
		r.appendString("const a = 1;")
		entry := r.scriptEntry()
		script, ok := entry.(resources.ScriptInlineString)
		if !ok || script.Content != "const a = 1;" {
			t.Fatalf("single string script entry = %#v", entry)
		}

		r = &embeddedResource{}
		r.appendBytes([]byte("const b = 2;"))
		entry = r.scriptEntry()
		scriptBytes, ok := entry.(resources.ScriptInlineBytes)
		if !ok || string(scriptBytes.Content) != "const b = 2;" {
			t.Fatalf("single bytes script entry = %#v", entry)
		}

		r = &embeddedResource{}
		r.appendString("const ")
		r.appendBytes([]byte("c = 3;"))
		entry = r.scriptEntry()
		merged, ok := entry.(resources.ScriptInlineString)
		if !ok || merged.Content != "const c = 3;" {
			t.Fatalf("merged script entry = %#v", entry)
		}
	})

	t.Run("style empty", func(t *testing.T) {
		r := &embeddedResource{}
		if r.styleEntry() != nil {
			t.Fatal("empty style entry should be nil")
		}
	})

	t.Run("style string and bytes", func(t *testing.T) {
		r := &embeddedResource{}
		r.appendString("h1 { color: red; }")
		entry := r.styleEntry()
		style, ok := entry.(resources.StyleString)
		if !ok || style.Content != "h1 { color: red; }" {
			t.Fatalf("single string style entry = %#v", entry)
		}

		r = &embeddedResource{}
		r.appendBytes([]byte("h2 { color: blue; }"))
		entry = r.styleEntry()
		styleBytes, ok := entry.(resources.StyleBytes)
		if !ok || string(styleBytes.Content) != "h2 { color: blue; }" {
			t.Fatalf("single bytes style entry = %#v", entry)
		}

		r = &embeddedResource{}
		r.appendString("h3")
		r.appendBytes([]byte(" { color: green; }"))
		entry = r.styleEntry()
		merged, ok := entry.(resources.StyleBytes)
		if !ok || string(merged.Content) != "h3 { color: green; }" {
			t.Fatalf("merged style entry = %#v", entry)
		}
	})
}

func TestResourceDump(t *testing.T) {
	var out bytes.Buffer
	r := &embeddedResource{
		openJob:  gox.NewJobHeadOpen(context.Background(), 1, gox.KindRegular, "script", gox.NewAttrs()),
		closeJob: gox.NewJobHeadClose(context.Background(), 1, gox.KindRegular, "script"),
	}
	if err := r.dump(defaultPrinter{&out}); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "<script></script>" {
		t.Fatalf("dump output = %q", got)
	}
}

func TestProcessTitleErrors(t *testing.T) {
	rp := &resourcePrinter{}
	tit := &title{
		openJob: gox.NewJobHeadOpen(context.Background(), 10, gox.KindRegular, "title", gox.NewAttrs()),
	}

	if err := rp.processTitle(gox.NewJobHeadOpen(context.Background(), 11, gox.KindRegular, "span", gox.NewAttrs()), tit); err == nil || !strings.Contains(err.Error(), "cannot contain nested tags") {
		t.Fatalf("unexpected nested-title error: %v", err)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("processTitle should panic on components")
		}
	}()
	_ = rp.processTitle(gox.NewJobComp(context.Background(), testComp{}), tit)
}

func TestProcessTitleWrongClose(t *testing.T) {
	rp := &resourcePrinter{}
	tit := &title{
		openJob: gox.NewJobHeadOpen(context.Background(), 10, gox.KindRegular, "title", gox.NewAttrs()),
	}
	err := rp.processTitle(gox.NewJobHeadClose(context.Background(), 11, gox.KindRegular, "title"), tit)
	if err == nil || !strings.Contains(err.Error(), "does not match the open tag") {
		t.Fatalf("wrong close error = %v", err)
	}
}

func TestProcessTitleSuccess(t *testing.T) {
	inst := &titleInstance{}
	ctx := context.WithValue(context.Background(), ctex.KeyCore, core.NewCore(inst, titleDoor{}))
	open := gox.NewJobHeadOpen(ctx, 10, gox.KindRegular, "title", gox.NewAttrs())
	open.Attrs.Get("data-id").Set("hero")

	rp := &resourcePrinter{}
	tit := &title{openJob: open}

	if err := rp.processTitle(gox.NewJobText(ctx, "Hello"), tit); err != nil {
		t.Fatal(err)
	}
	if err := rp.processTitle(gox.NewJobHeadClose(ctx, 10, gox.KindRegular, "title"), tit); err != nil {
		t.Fatal(err)
	}
	if inst.title != "Hello" {
		t.Fatalf("title content = %q", inst.title)
	}
	if inst.titleAttrs == nil {
		t.Fatal("expected title attrs clone")
	}
	attr, ok := inst.titleAttrs.Find("data-id")
	if !ok || attr.Value() != "hero" {
		t.Fatalf("title attrs = %#v", inst.titleAttrs)
	}
	if rp.resource != nil {
		t.Fatal("resource printer should clear active title resource")
	}
}

func TestProcessMetaBranches(t *testing.T) {
	ctx, inst, _, _ := newPrinterCore(t, true)
	rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}

	err := rp.processMeta(gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "meta", gox.NewAttrs()))
	if err == nil || !strings.Contains(err.Error(), "must be a void element") {
		t.Fatalf("non-void meta error = %v", err)
	}

	nameAttrs := gox.NewAttrs()
	nameAttrs.Get("name").Set("description")
	nameAttrs.Get("content").Set("hello")
	if err := rp.processMeta(gox.NewJobHeadOpen(ctx, 2, gox.KindVoid, "meta", nameAttrs)); err != nil {
		t.Fatal(err)
	}

	propertyAttrs := gox.NewAttrs()
	propertyAttrs.Get("property").Set("og:title")
	propertyAttrs.Get("content").Set("hero")
	if err := rp.processMeta(gox.NewJobHeadOpen(ctx, 3, gox.KindVoid, "meta", propertyAttrs)); err != nil {
		t.Fatal(err)
	}

	if len(inst.metas) != 2 {
		t.Fatalf("expected 2 meta updates, got %d", len(inst.metas))
	}
	if inst.metas[0].name != "description" || inst.metas[0].property {
		t.Fatalf("unexpected name meta update = %#v", inst.metas[0])
	}
	if inst.metas[1].name != "og:title" || !inst.metas[1].property {
		t.Fatalf("unexpected property meta update = %#v", inst.metas[1])
	}
	if _, ok := inst.metas[0].attrs.Find("name"); ok {
		t.Fatalf("expected name attr to be unset in stored meta")
	}

	var out bytes.Buffer
	rp = &resourcePrinter{printer: defaultPrinter{&out}}
	passAttrs := gox.NewAttrs()
	passAttrs.Get("charset").Set("utf-8")
	if err := rp.processMeta(gox.NewJobHeadOpen(ctx, 4, gox.KindVoid, "meta", passAttrs)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `charset="utf-8"`) {
		t.Fatalf("expected passthrough meta output, got %q", out.String())
	}
}

func TestScanGenericSrcBranches(t *testing.T) {
	ctx, _, _, _ := newPrinterCore(t, true)

	t.Run("passes through missing attrs", func(t *testing.T) {
		var out bytes.Buffer
		rp := &resourcePrinter{printer: defaultPrinter{&out}}
		job := gox.NewJobHeadOpen(ctx, 1, gox.KindVoid, "img", gox.NewAttrs())
		if err := rp.scanGenericSrc(job); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); !strings.Contains(got, "<img") {
			t.Fatalf("expected passthrough img, got %q", got)
		}
	})

	t.Run("registers hook source", func(t *testing.T) {
		var out bytes.Buffer
		rp := &resourcePrinter{printer: defaultPrinter{&out}}
		attrs := gox.NewAttrs()
		attrs.Get("src").Set(HandlerSimpleFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}))
		job := gox.NewJobHeadOpen(ctx, 4, gox.KindVoid, "script", attrs)
		if err := rp.scanGenericSrc(job); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); !strings.Contains(got, `/h/instance/7/1`) {
			t.Fatalf("expected hook path, got %q", got)
		}
	})

	t.Run("caches static source", func(t *testing.T) {
		var out bytes.Buffer
		rp := &resourcePrinter{printer: defaultPrinter{&out}}
		attrs := gox.NewAttrs()
		attrs.Get("src").Set(SourceString("hello"))
		attrs.Get("cache").Set(true)
		attrs.Get("type").Set("text/plain")
		attrs.Get("name").Set("asset.txt")
		job := gox.NewJobHeadOpen(ctx, 5, gox.KindVoid, "img", attrs)
		if err := rp.scanGenericSrc(job); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, `/r/`) || !strings.Contains(got, `.asset.txt`) {
			t.Fatalf("expected cached resource path, got %q", got)
		}
	})

	t.Run("returns canceled when hook registration fails", func(t *testing.T) {
		failCtx, _, _, _ := newPrinterCore(t, false)
		rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}
		attrs := gox.NewAttrs()
		attrs.Get("src").Set(SourceString("hello"))
		err := rp.scanGenericSrc(gox.NewJobHeadOpen(failCtx, 6, gox.KindVoid, "iframe", attrs))
		if err != context.Canceled {
			t.Fatalf("expected context canceled, got %v", err)
		}
	})
}

func TestPrepareLinkStyleBranches(t *testing.T) {
	ctx, inst, _, _ := newPrinterCore(t, true)

	t.Run("tracks external styles in csp", func(t *testing.T) {
		var out bytes.Buffer
		rp := &resourcePrinter{printer: defaultPrinter{&out}}
		attrs := gox.NewAttrs()
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(SourceExternal("https://cdn.example/app.css"))
		if err := rp.prepareLinkStyle(gox.NewJobHeadOpen(ctx, 3, gox.KindVoid, "link", attrs)); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `https://cdn.example/app.css`) {
			t.Fatalf("expected external stylesheet passthrough, got %q", out.String())
		}
		if !strings.Contains(inst.CSPCollector().Generate(), "https://cdn.example/app.css") {
			t.Fatalf("expected csp style source to be recorded, got %q", inst.CSPCollector().Generate())
		}
	})
}

func TestPrepareScriptAndSendBranches(t *testing.T) {
	ctx, inst, _, modules := newPrinterCore(t, true)

	attrs := gox.NewAttrs()
	attrs.Get("src").Set(SourceString(`window.__prepared = "ok"`))
	attrs.Get("type").Set("module")
	attrs.Get("specifier").Set("prepared")
	attrs.Get("name").Set("prepared-module.js")
	prepareRecorder := &recordingPrinter{}
	rp := &resourcePrinter{printer: prepareRecorder}
	if err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "script", attrs)); err != nil {
		t.Fatal(err)
	}
	if len(prepareRecorder.opens) != 1 {
		t.Fatalf("expected prepareScript to send exactly one open tag, got %#v", prepareRecorder.opens)
	}
	if prepareRecorder.opens[0].tag != "script" {
		t.Fatalf("expected prepareScript to preserve the script tag, got %#v", prepareRecorder.opens[0])
	}
	srcAttr, ok := prepareRecorder.opens[0].attrs.Find("src")
	if !ok {
		t.Fatal("expected prepareScript to keep src attr")
	}
	src, ok := srcAttr.Value().(string)
	if !ok {
		t.Fatalf("expected prepareScript to rewrite src to a string path, got %#v", srcAttr.Value())
	}
	if !strings.Contains(src, "/r/") || !strings.HasSuffix(src, ".prepared-module.js") {
		t.Fatalf("expected prepared script path to target a hosted resource, got %q", src)
	}
	if got := modules.values["prepared"]; got != src {
		t.Fatalf("expected module registry to store rewritten src, got %#v", modules.values)
	}

	match, ok := inst.PathMaker().Match(httptest.NewRequest(http.MethodGet, src, nil))
	if !ok {
		t.Fatalf("expected prepared script path to match a resource route, got %q", src)
	}
	resourceID, ok := match.Resource()
	if !ok {
		t.Fatalf("expected resource id from prepared script path, got %#v", match)
	}
	rec := httptest.NewRecorder()
	inst.registry.Serve(resourceID, rec, httptest.NewRequest(http.MethodGet, src, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected prepared script resource to serve successfully, got %d", rec.Code)
	}
	body := strings.NewReplacer(" ", "", "\n", "", "\t", "", "\r", "").Replace(rec.Body.String())
	if !strings.Contains(body, `window.__prepared="ok"`) {
		t.Fatalf("expected prepared script resource to contain original source, got %q", rec.Body.String())
	}
	if strings.Contains(body, `_d0r(document.currentScript`) {
		t.Fatalf("expected prepared module resource to avoid the inline wrapper, got %q", rec.Body.String())
	}

	sendRecorder := &recordingPrinter{}
	printer := NewResourcePrinter(sendRecorder)
	inlineAttrs := gox.NewAttrs()
	inlineAttrs.Get("name").Set("embedded.js")
	if err := printer.Send(gox.NewJobHeadOpen(ctx, 2, gox.KindRegular, "script", inlineAttrs)); err != nil {
		t.Fatal(err)
	}
	if err := printer.Send(gox.NewJobRaw(ctx, `window.__embedded = "ok"`)); err != nil {
		t.Fatal(err)
	}
	if err := printer.Send(gox.NewJobHeadClose(ctx, 2, gox.KindRegular, "script")); err != nil {
		t.Fatal(err)
	}
	if len(sendRecorder.opens) != 1 || sendRecorder.closes != 1 {
		t.Fatalf("expected embedded script render to emit one open and one close tag, got opens=%#v closes=%d", sendRecorder.opens, sendRecorder.closes)
	}
	if len(sendRecorder.raw) != 0 || len(sendRecorder.bytes) != 0 {
		t.Fatalf("expected embedded script body to be externalized, got raw=%#v bytes=%#v", sendRecorder.raw, sendRecorder.bytes)
	}
	inlineSrcAttr, ok := sendRecorder.opens[0].attrs.Find("src")
	if !ok {
		t.Fatal("expected embedded script render to add src attr")
	}
	inlineSrc, ok := inlineSrcAttr.Value().(string)
	if !ok {
		t.Fatalf("expected embedded script src to be rewritten to a string path, got %#v", inlineSrcAttr.Value())
	}
	if !strings.Contains(inlineSrc, "/r/") || !strings.HasSuffix(inlineSrc, ".embedded.js") {
		t.Fatalf("expected embedded script render to use a hosted resource path, got %q", inlineSrc)
	}

	match, ok = inst.PathMaker().Match(httptest.NewRequest(http.MethodGet, inlineSrc, nil))
	if !ok {
		t.Fatalf("expected embedded script path to match a resource route, got %q", inlineSrc)
	}
	resourceID, ok = match.Resource()
	if !ok {
		t.Fatalf("expected resource id from embedded script path, got %#v", match)
	}
	rec = httptest.NewRecorder()
	inst.registry.Serve(resourceID, rec, httptest.NewRequest(http.MethodGet, inlineSrc, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected embedded script resource to serve successfully, got %d", rec.Code)
	}
	inlineBody := strings.NewReplacer(" ", "", "\n", "", "\t", "", "\r", "").Replace(rec.Body.String())
	if !strings.Contains(inlineBody, `_d0r(document.currentScript`) {
		t.Fatalf("expected embedded script resource to use the inline wrapper, got %q", rec.Body.String())
	}
	if !strings.Contains(inlineBody, `window.__embedded="ok"`) {
		t.Fatalf("expected embedded script resource to contain original source, got %q", rec.Body.String())
	}
}

func TestPrepareScriptErrorsAndHelpers(t *testing.T) {
	ctx, _, _, _ := newPrinterCore(t, true)
	rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}

	t.Run("invalid script output", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("bundle").Set(true)
		attrs.Get("inline").Set(true)
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 2, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "only one of raw, inline, or bundle") {
			t.Fatalf("unexpected script output error: %v", err)
		}
	})

	t.Run("inline bundle script is rejected", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("bundle").Set(true)
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 3, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "cannot be bundled") {
			t.Fatalf("unexpected inline bundle error: %v", err)
		}
	})

	t.Run("inline module is rejected", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("type").Set("module")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 4, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "do not support modules") {
			t.Fatalf("unexpected inline module error: %v", err)
		}
	})

	t.Run("inline typescript is rejected", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("type").Set("text/typescript")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 5, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "do not support TypeScript") {
			t.Fatalf("unexpected inline ts error: %v", err)
		}
	})

	t.Run("raw typescript source is rejected", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("src").Set(SourceString("let x: number = 1"))
		attrs.Get("raw").Set(true)
		attrs.Get("type").Set("text/typescript")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 7, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "raw output does not support TypeScript") {
			t.Fatalf("unexpected raw ts error: %v", err)
		}
	})

	t.Run("regular src can't bundle or inline", func(t *testing.T) {
		bundleAttrs := gox.NewAttrs()
		bundleAttrs.Get("src").Set("/plain.js")
		bundleAttrs.Get("bundle").Set(true)
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 8, gox.KindRegular, "script", bundleAttrs))
		if err == nil || !strings.Contains(err.Error(), "only support raw output") {
			t.Fatalf("unexpected regular src bundle error: %v", err)
		}

		inlineAttrs := gox.NewAttrs()
		inlineAttrs.Get("src").Set("/plain.js")
		inlineAttrs.Get("inline").Set(true)
		err = rp.prepareScript(gox.NewJobHeadOpen(ctx, 9, gox.KindRegular, "script", inlineAttrs))
		if err == nil || !strings.Contains(err.Error(), "only support raw output") {
			t.Fatalf("unexpected regular src inline error: %v", err)
		}
	})

	t.Run("external inline errors while unknown sources pass through", func(t *testing.T) {
		externalAttrs := gox.NewAttrs()
		externalAttrs.Get("src").Set(SourceExternal("https://cdn.example/app.js"))
		externalAttrs.Get("inline").Set(true)
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 10, gox.KindRegular, "script", externalAttrs))
		if err == nil || !strings.Contains(err.Error(), "only support raw output") {
			t.Fatalf("unexpected external inline error: %v", err)
		}

		unknownAttrs := gox.NewAttrs()
		unknownAttrs.Get("src").Set(123)
		unknownAttrs.Get("bundle").Set(true)
		var out bytes.Buffer
		unknownRP := &resourcePrinter{printer: defaultPrinter{&out}}
		err = unknownRP.prepareScript(gox.NewJobHeadOpen(ctx, 11, gox.KindRegular, "script", unknownAttrs))
		if err != nil {
			t.Fatalf("expected unknown script source to pass through, got %v", err)
		}
		unknownHref := gox.NewAttrs()
		unknownHref.Get("rel").Set("modulepreload")
		unknownHref.Get("href").Set(123)
		unknownHref.Get("bundle").Set(true)
		out.Reset()
		unknownRP = &resourcePrinter{printer: defaultPrinter{&out}}
		err = unknownRP.prepareLinkModule(gox.NewJobHeadOpen(ctx, 12, gox.KindVoid, "link", unknownHref))
		if err != nil {
			t.Fatalf("expected unknown modulepreload href to pass through, got %v", err)
		}
	})

	t.Run("resource url cancellation", func(t *testing.T) {
		failCtx, _, _, _ := newPrinterCore(t, false)
		res, err := resources.NewRegistry(pagePrinterSettings{conf: (&common.SystemConf{})}).Static(resources.StaticString{Content: "x"}, "text/plain")
		if err != nil {
			t.Fatal(err)
		}
		_, err = resourceURL(failCtx.Value(ctex.KeyCore).(core.Core), res, resources.ModeNoHost, "x.txt")
		if err != context.Canceled {
			t.Fatalf("expected canceled resource url, got %v", err)
		}
	})

	t.Run("source conversion and format helpers", func(t *testing.T) {
		if _, ok := getSource(plainAttrWithValue(HandlerSimpleFunc(func(http.ResponseWriter, *http.Request) {}))).(SourceHook); !ok {
			t.Fatal("expected simple handler to convert to SourceHook")
		}
		if _, ok := getSource(plainAttrWithValue(HandlerFunc(func(context.Context, http.ResponseWriter, *http.Request) bool { return false }))).(SourceHook); !ok {
			t.Fatal("expected full handler to convert to SourceHook")
		}
		if _, ok := getSource(plainAttrWithValue([]byte("x"))).(SourceBytes); !ok {
			t.Fatal("expected []byte attr to convert to SourceBytes")
		}
		if got := getSource(plainAttrWithValue("plain")); got != "plain" {
			t.Fatalf("expected passthrough source value, got %#v", got)
		}
		format, err := scriptBundle.format(false)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := format.(resources.FormatCommon); !ok {
			t.Fatalf("expected common bundle format, got %#v", format)
		}
		if format, err = scriptDefault.format(true); err != nil {
			t.Fatal(err)
		} else if _, ok := format.(resources.FormatModule); !ok {
			t.Fatalf("expected module default format, got %#v", format)
		}
		if format, err = scriptRaw.format(false); err != nil {
			t.Fatal(err)
		} else if _, ok := format.(resources.FormatRaw); !ok {
			t.Fatalf("expected raw format, got %#v", format)
		}
		if _, err := scriptInline.format(true); err == nil {
			t.Fatal("expected inline module format error")
		}
	})
}

func TestProcessResErrors(t *testing.T) {
	ctx, _, _, _ := newPrinterCore(t, true)
	rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}
	res := &embeddedResource{
		openJob: gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "style", gox.NewAttrs()),
		kind:    embeddedStyle,
	}
	if err := rp.processRes(gox.NewJobHeadClose(ctx, 2, gox.KindRegular, "style"), res); err == nil || !strings.Contains(err.Error(), "does not match the open tag") {
		t.Fatalf("unexpected mismatch error: %v", err)
	}
	if err := rp.processRes(gox.NewJobText(ctx, "bad"), res); err == nil || !strings.Contains(err.Error(), "only text or byte jobs") {
		t.Fatalf("unexpected invalid content error: %v", err)
	}
}

func TestResourceRenderFallbackAndDumpError(t *testing.T) {
	t.Run("empty script and style dump original tags", func(t *testing.T) {
		ctx, _, _, _ := newPrinterCore(t, true)

		var scriptOut bytes.Buffer
		script := &embeddedResource{
			openJob:  gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "script", gox.NewAttrs()),
			closeJob: gox.NewJobHeadClose(ctx, 1, gox.KindRegular, "script"),
			kind:     embeddedScript,
			props:    &resourceProps{mode: resources.ModeHost},
		}
		if err := script.render(defaultPrinter{&scriptOut}); err != nil {
			t.Fatal(err)
		}
		if scriptOut.String() != "<script></script>" {
			t.Fatalf("empty script render = %q", scriptOut.String())
		}

		var styleOut bytes.Buffer
		style := &embeddedResource{
			openJob:  gox.NewJobHeadOpen(ctx, 2, gox.KindRegular, "style", gox.NewAttrs()),
			closeJob: gox.NewJobHeadClose(ctx, 2, gox.KindRegular, "style"),
			kind:     embeddedStyle,
			props:    &resourceProps{mode: resources.ModeHost},
		}
		if err := style.render(defaultPrinter{&styleOut}); err != nil {
			t.Fatal(err)
		}
		if styleOut.String() != "<style></style>" {
			t.Fatalf("empty style render = %q", styleOut.String())
		}
	})

	t.Run("dump returns printer error", func(t *testing.T) {
		ctx, _, _, _ := newPrinterCore(t, true)
		res := &embeddedResource{
			openJob:  gox.NewJobHeadOpen(ctx, 3, gox.KindRegular, "script", gox.NewAttrs()),
			closeJob: gox.NewJobHeadClose(ctx, 3, gox.KindRegular, "script"),
		}
		printer := &failPrinter{fail: 2}
		if err := res.dump(printer); err != context.Canceled {
			t.Fatalf("expected dump error, got %v", err)
		}
	})
}

func TestResourcePrinterScanAndModulePreloadBranches(t *testing.T) {
	ctx, _, _, modules := newPrinterCore(t, true)

	t.Run("scan captures title resource", func(t *testing.T) {
		rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}
		if err := rp.scan(gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "title", gox.NewAttrs())); err != nil {
			t.Fatal(err)
		}
		if _, ok := rp.resource.(*title); !ok {
			t.Fatalf("expected active title resource, got %#v", rp.resource)
		}
	})

	t.Run("scan passes generic source when rel is missing", func(t *testing.T) {
		var out bytes.Buffer
		rp := &resourcePrinter{printer: defaultPrinter{&out}}
		attrs := gox.NewAttrs()
		attrs.Get("href").Set(SourceString("hi"))
		if err := rp.scan(gox.NewJobHeadOpen(ctx, 2, gox.KindVoid, "link", attrs)); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `/h/instance/7/1`) {
			t.Fatalf("expected generic link hook path, got %q", out.String())
		}
	})

	t.Run("script scan rewrites data attrs and keeps tag when needed", func(t *testing.T) {
		var out bytes.Buffer
		rp := NewResourcePrinter(defaultPrinter{&out})
		attrs := gox.NewAttrs()
		attrs.Get("src").Set("/modules/app.js")
		attrs.Get("type").Set("module")
		attrs.Get("specifier").Set("kept")
		attrs.Get("data:mode").Set("fast")
		attrs.Get("async").Set(true)
		open := gox.NewJobHeadOpen(ctx, 5, gox.KindRegular, "script", attrs)
		if err := rp.Send(open); err != nil {
			t.Fatal(err)
		}
		if err := rp.Send(gox.NewJobHeadClose(ctx, 5, gox.KindRegular, "script")); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, `type="module"`) || !strings.Contains(got, `data-d0d-mode=`) {
			t.Fatalf("expected kept module tag with rewritten data attr, got %q", got)
		}
		if modules.values["kept"] != "/modules/app.js" {
			t.Fatalf("expected kept specifier registry add, got %#v", modules.values)
		}
	})
}

func plainAttrWithValue(value any) gox.Attr {
	attrs := gox.NewAttrs()
	attr := attrs.Get("value")
	attr.Set(value)
	return attr
}

func (rp *resourcePrinter) prepareLinkStyle(open *gox.JobHeadOpen) error {
	return rp.processProps(open, newStyleProps(true))
}

func (rp *resourcePrinter) prepareStyle(open *gox.JobHeadOpen) error {
	return rp.processProps(open, newStyleProps(false))
}

func (rp *resourcePrinter) prepareScript(open *gox.JobHeadOpen) error {
	return rp.processProps(open, newScriptProps(false))
}

func (rp *resourcePrinter) prepareLinkModule(open *gox.JobHeadOpen) error {
	return rp.processProps(open, newScriptProps(true))
}

func getSource(attr gox.Attr) any {
	props := &resourceProps{}
	props.readSource(attr)
	return props.source
}

func resourceURL(core core.Core, res *resources.Resource, mode resources.ResourceMode, name string) (string, error) {
	props := &resourceProps{
		mode: mode,
		name: name,
	}
	return props.resourceURL(core, res)
}
