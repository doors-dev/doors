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
	license    string
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
func (t *titleInstance) License() string                       { return t.license }
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
		license:  "licensed",
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

func TestResourceEntryHelpers(t *testing.T) {
	t.Run("script empty", func(t *testing.T) {
		r := &resource{}
		if r.scriptEntry() != nil {
			t.Fatal("empty script entry should be nil")
		}
	})

	t.Run("script string and bytes", func(t *testing.T) {
		r := &resource{}
		r.appendString("const a = 1;")
		entry := r.scriptEntry()
		script, ok := entry.(resources.ScriptInlineString)
		if !ok || script.Content != "const a = 1;" {
			t.Fatalf("single string script entry = %#v", entry)
		}

		r = &resource{}
		r.appendBytes([]byte("const b = 2;"))
		entry = r.scriptEntry()
		scriptBytes, ok := entry.(resources.ScriptInlineBytes)
		if !ok || string(scriptBytes.Content) != "const b = 2;" {
			t.Fatalf("single bytes script entry = %#v", entry)
		}

		r = &resource{}
		r.appendString("const ")
		r.appendBytes([]byte("c = 3;"))
		entry = r.scriptEntry()
		merged, ok := entry.(resources.ScriptInlineString)
		if !ok || merged.Content != "const c = 3;" {
			t.Fatalf("merged script entry = %#v", entry)
		}
	})

	t.Run("style empty", func(t *testing.T) {
		r := &resource{}
		if r.styleEntry() != nil {
			t.Fatal("empty style entry should be nil")
		}
	})

	t.Run("style string and bytes", func(t *testing.T) {
		r := &resource{}
		r.appendString("h1 { color: red; }")
		entry := r.styleEntry()
		style, ok := entry.(resources.StyleString)
		if !ok || style.Content != "h1 { color: red; }" {
			t.Fatalf("single string style entry = %#v", entry)
		}

		r = &resource{}
		r.appendBytes([]byte("h2 { color: blue; }"))
		entry = r.styleEntry()
		styleBytes, ok := entry.(resources.StyleBytes)
		if !ok || string(styleBytes.Content) != "h2 { color: blue; }" {
			t.Fatalf("single bytes style entry = %#v", entry)
		}

		r = &resource{}
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
	r := &resource{
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

	if err := rp.processTitle(gox.NewJobHeadOpen(context.Background(), 11, gox.KindRegular, "span", gox.NewAttrs()), tit); err == nil || !strings.Contains(err.Error(), "title can't contain other tags") {
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
	if err == nil || !strings.Contains(err.Error(), "unexpected close job") {
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
	if err == nil || !strings.Contains(err.Error(), "non-void meta") {
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

	t.Run("rejects cache on plain string", func(t *testing.T) {
		rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}
		attrs := gox.NewAttrs()
		attrs.Get("src").Set("/plain.js")
		attrs.Get("cache").Set(true)
		err := rp.scanGenericSrc(gox.NewJobHeadOpen(ctx, 2, gox.KindVoid, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "cache attr requires") {
			t.Fatalf("unexpected cache error: %v", err)
		}
	})

	t.Run("rejects source without handler", func(t *testing.T) {
		rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}
		attrs := gox.NewAttrs()
		attrs.Get("href").Set(SourceExternal("https://cdn.example/app.css"))
		err := rp.scanGenericSrc(gox.NewJobHeadOpen(ctx, 3, gox.KindVoid, "link", attrs))
		if err == nil || !strings.Contains(err.Error(), "source does not provide a handler") {
			t.Fatalf("unexpected handler error: %v", err)
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
		attrs.Get("content-type").Set("text/plain")
		attrs.Get("name").Set("asset.txt")
		job := gox.NewJobHeadOpen(ctx, 5, gox.KindVoid, "img", attrs)
		if err := rp.scanGenericSrc(job); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, `/r/`) || !strings.Contains(got, `.asset.txt`) {
			t.Fatalf("expected cached resource path, got %q", got)
		}
		if strings.Contains(got, "content-type") {
			t.Fatalf("expected content-type attr to be removed, got %q", got)
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

	t.Run("rejects non-void", func(t *testing.T) {
		rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}
		err := rp.prepareLinkStyle(gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "link", gox.NewAttrs()))
		if err == nil || !strings.Contains(err.Error(), "non-void link stylesheet tag") {
			t.Fatalf("unexpected non-void error: %v", err)
		}
	})

	t.Run("passes through raw styles", func(t *testing.T) {
		var out bytes.Buffer
		rp := &resourcePrinter{printer: defaultPrinter{&out}}
		attrs := gox.NewAttrs()
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set("/app.css")
		attrs.Get("output").Set("raw")
		if err := rp.prepareLinkStyle(gox.NewJobHeadOpen(ctx, 2, gox.KindVoid, "link", attrs)); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `/app.css`) {
			t.Fatalf("expected raw stylesheet passthrough, got %q", out.String())
		}
	})

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

	t.Run("creates private style hook path from source", func(t *testing.T) {
		var out bytes.Buffer
		rp := &resourcePrinter{printer: defaultPrinter{&out}}
		attrs := gox.NewAttrs()
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(SourceString("body{color:red}"))
		attrs.Get("private").Set(true)
		attrs.Get("name").Set("private.css")
		if err := rp.prepareLinkStyle(gox.NewJobHeadOpen(ctx, 4, gox.KindVoid, "link", attrs)); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, `/h/`) || !strings.Contains(got, `private.css`) {
			t.Fatalf("expected private stylesheet hook path, got %q", got)
		}
	})
}

func TestPrepareScriptAndSendBranches(t *testing.T) {
	t.Run("inline script becomes hosted resource", func(t *testing.T) {
		ctx, _, _, _ := newPrinterCore(t, true)
		var out bytes.Buffer
		rp := NewResourcePrinter(defaultPrinter{&out})
		open := gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "script", gox.NewAttrs())
		if err := rp.Send(open); err != nil {
			t.Fatal(err)
		}
		if err := rp.Send(gox.NewJobRaw(ctx, "console.log('hi')")); err != nil {
			t.Fatal(err)
		}
		if err := rp.Send(gox.NewJobHeadClose(ctx, 1, gox.KindRegular, "script")); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, `<script src="/~/srv/r/`) || !strings.Contains(got, `.inline.js`) {
			t.Fatalf("expected hosted inline script, got %q", got)
		}
	})

	t.Run("inline nocache style becomes hook-backed link", func(t *testing.T) {
		ctx, _, _, _ := newPrinterCore(t, true)
		var out bytes.Buffer
		rp := NewResourcePrinter(defaultPrinter{&out})
		attrs := gox.NewAttrs()
		attrs.Get("nocache").Set(true)
		open := gox.NewJobHeadOpen(ctx, 2, gox.KindRegular, "style", attrs)
		if err := rp.Send(open); err != nil {
			t.Fatal(err)
		}
		if err := rp.Send(gox.NewJobBytes(ctx, []byte("body{color:red}"))); err != nil {
			t.Fatal(err)
		}
		if err := rp.Send(gox.NewJobHeadClose(ctx, 2, gox.KindRegular, "style")); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, `<link`) ||
			!strings.Contains(got, `/~/srv/h/instance/7/1/inline.css`) ||
			!strings.Contains(got, `rel="stylesheet"`) {
			t.Fatalf("expected nocache style link output, got %q", got)
		}
	})

	t.Run("specifier only script is skipped and registered", func(t *testing.T) {
		ctx, _, _, modules := newPrinterCore(t, true)
		var out bytes.Buffer
		rp := NewResourcePrinter(defaultPrinter{&out})
		attrs := gox.NewAttrs()
		attrs.Get("src").Set("/assets/app.js")
		attrs.Get("specifier").Set("app")
		open := gox.NewJobHeadOpen(ctx, 3, gox.KindRegular, "script", attrs)
		if err := rp.Send(open); err != nil {
			t.Fatal(err)
		}
		if err := rp.Send(gox.NewJobHeadClose(ctx, 3, gox.KindRegular, "script")); err != nil {
			t.Fatal(err)
		}
		if out.Len() != 0 {
			t.Fatalf("expected specifier-only script to be skipped, got %q", out.String())
		}
		if modules.values["app"] != "/assets/app.js" {
			t.Fatalf("expected module registry add, got %#v", modules.values)
		}
	})

	t.Run("supports proxy script source", func(t *testing.T) {
		ctx, _, _, _ := newPrinterCore(t, true)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("proxied"))
		}))
		defer server.Close()

		var out bytes.Buffer
		rp := NewResourcePrinter(defaultPrinter{&out})
		attrs := gox.NewAttrs()
		attrs.Get("src").Set(SourceProxy(server.URL))
		open := gox.NewJobHeadOpen(ctx, 4, gox.KindRegular, "script", attrs)
		if err := rp.Send(open); err != nil {
			t.Fatal(err)
		}
		if err := rp.Send(gox.NewJobHeadClose(ctx, 4, gox.KindRegular, "script")); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); !strings.Contains(got, `/h/instance/7/1/`) {
			t.Fatalf("expected proxy-backed hook path, got %q", got)
		}
	})
}

func TestPrepareScriptErrorsAndHelpers(t *testing.T) {
	ctx, _, _, _ := newPrinterCore(t, true)
	rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}

	t.Run("invalid style output", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("output").Set("weird")
		err := rp.prepareStyle(gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "style", attrs))
		if err == nil || !strings.Contains(err.Error(), "unexpected style output kind") {
			t.Fatalf("unexpected style output error: %v", err)
		}
	})

	t.Run("invalid script output", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("output").Set("weird")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 2, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "unknown script output") {
			t.Fatalf("unexpected script output error: %v", err)
		}
	})

	t.Run("inline bundle script is rejected", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("output").Set("bundle")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 3, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "can't be bundeled") {
			t.Fatalf("unexpected inline bundle error: %v", err)
		}
	})

	t.Run("inline module is rejected", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("type").Set("module")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 4, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "inline modules are not supported") {
			t.Fatalf("unexpected inline module error: %v", err)
		}
	})

	t.Run("inline typescript is rejected", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("type").Set("text/typescript")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 5, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "inline typescript is not supported") {
			t.Fatalf("unexpected inline ts error: %v", err)
		}
	})

	t.Run("inline modulepreload is rejected", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("rel").Set("modulepreload")
		attrs.Get("href").Set(SourceString("console.log(1)"))
		attrs.Get("output").Set("inline")
		err := rp.prepareLinkModule(gox.NewJobHeadOpen(ctx, 6, gox.KindVoid, "link", attrs))
		if err == nil || !strings.Contains(err.Error(), "inline modulepreload is not supported") {
			t.Fatalf("unexpected modulepreload inline error: %v", err)
		}
	})

	t.Run("raw typescript source is rejected", func(t *testing.T) {
		attrs := gox.NewAttrs()
		attrs.Get("src").Set(SourceString("let x: number = 1"))
		attrs.Get("output").Set("raw")
		attrs.Get("type").Set("text/typescript")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 7, gox.KindRegular, "script", attrs))
		if err == nil || !strings.Contains(err.Error(), "raw typescript can't be served") {
			t.Fatalf("unexpected raw ts error: %v", err)
		}
	})

	t.Run("regular src can't bundle or inline", func(t *testing.T) {
		bundleAttrs := gox.NewAttrs()
		bundleAttrs.Get("src").Set("/plain.js")
		bundleAttrs.Get("output").Set("bundle")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 8, gox.KindRegular, "script", bundleAttrs))
		if err == nil || !strings.Contains(err.Error(), "can't bundle script with regular src") {
			t.Fatalf("unexpected regular src bundle error: %v", err)
		}

		inlineAttrs := gox.NewAttrs()
		inlineAttrs.Get("src").Set("/plain.js")
		inlineAttrs.Get("output").Set("inline")
		err = rp.prepareScript(gox.NewJobHeadOpen(ctx, 9, gox.KindRegular, "script", inlineAttrs))
		if err == nil || !strings.Contains(err.Error(), `can't prepare "inline" script with regular src`) {
			t.Fatalf("unexpected regular src inline error: %v", err)
		}
	})

	t.Run("external and unknown src errors", func(t *testing.T) {
		externalAttrs := gox.NewAttrs()
		externalAttrs.Get("src").Set(SourceExternal("https://cdn.example/app.js"))
		externalAttrs.Get("output").Set("inline")
		err := rp.prepareScript(gox.NewJobHeadOpen(ctx, 10, gox.KindRegular, "script", externalAttrs))
		if err == nil || !strings.Contains(err.Error(), `can't prepare "inline" script with extarnal src`) {
			t.Fatalf("unexpected external inline error: %v", err)
		}

		unknownAttrs := gox.NewAttrs()
		unknownAttrs.Get("src").Set(123)
		unknownAttrs.Get("output").Set("bundle")
		err = rp.prepareScript(gox.NewJobHeadOpen(ctx, 11, gox.KindRegular, "script", unknownAttrs))
		if err == nil || !strings.Contains(err.Error(), "unknown type of src attribute on script") {
			t.Fatalf("unexpected unknown src error: %v", err)
		}

		unknownHref := gox.NewAttrs()
		unknownHref.Get("rel").Set("modulepreload")
		unknownHref.Get("href").Set(123)
		unknownHref.Get("output").Set("bundle")
		err = rp.prepareLinkModule(gox.NewJobHeadOpen(ctx, 12, gox.KindVoid, "link", unknownHref))
		if err == nil || !strings.Contains(err.Error(), "unknown type of href attribute on modulepreload link") {
			t.Fatalf("unexpected unknown href error: %v", err)
		}
	})

	t.Run("raw inline script stays inline", func(t *testing.T) {
		var out bytes.Buffer
		rp := NewResourcePrinter(defaultPrinter{&out})
		attrs := gox.NewAttrs()
		attrs.Get("output").Set("raw")
		open := gox.NewJobHeadOpen(ctx, 13, gox.KindRegular, "script", attrs)
		if err := rp.Send(open); err != nil {
			t.Fatal(err)
		}
		if err := rp.Send(gox.NewJobRaw(ctx, `console.log("raw")`)); err != nil {
			t.Fatal(err)
		}
		if err := rp.Send(gox.NewJobHeadClose(ctx, 13, gox.KindRegular, "script")); err != nil {
			t.Fatal(err)
		}
		got := out.String()
		if !strings.Contains(got, `console.log("raw")`) || strings.Contains(got, `/r/`) || strings.Contains(got, `/h/`) {
			t.Fatalf("expected raw inline script to stay inline, got %q", got)
		}
	})

	t.Run("resource url cancellation", func(t *testing.T) {
		failCtx, _, _, _ := newPrinterCore(t, false)
		res, err := resources.NewRegistry(pagePrinterSettings{conf: (&common.SystemConf{})}).Static(resources.StaticString{Content: "x"}, "text/plain")
		if err != nil {
			t.Fatal(err)
		}
		_, err = resourceURL(failCtx.Value(ctex.KeyCore).(core.Core), res, resources.ModeCache, "x.txt")
		if err != context.Canceled {
			t.Fatalf("expected canceled resource url, got %v", err)
		}
	})

	t.Run("parse helpers and source conversion", func(t *testing.T) {
		if got, ok := rp.parseStyleOutput("minify"); !ok || got != styleMinify {
			t.Fatalf("parseStyleOutput(minify) = %v %v", got, ok)
		}
		if got, ok := rp.parseStyleOutput("nope"); ok || got != "" {
			t.Fatalf("parseStyleOutput(nope) = %v %v", got, ok)
		}
		if got, ok := rp.parseScriptOutput("bundle"); !ok || got != scriptBundle {
			t.Fatalf("parseScriptOutput(bundle) = %v %v", got, ok)
		}
		if got, ok := rp.parseScriptOutput(42); ok || got != "" {
			t.Fatalf("parseScriptOutput(42) = %v %v", got, ok)
		}
		if got, ok := rp.parseStyleOutput(""); !ok || got != styleDefault {
			t.Fatalf("parseStyleOutput(\"\") = %v %v", got, ok)
		}
		if _, ok := rp.getSource(plainAttrWithValue(HandlerSimpleFunc(func(http.ResponseWriter, *http.Request) {}))).(SourceHook); !ok {
			t.Fatal("expected simple handler to convert to SourceHook")
		}
		if _, ok := rp.getSource(plainAttrWithValue(HandlerFunc(func(context.Context, http.ResponseWriter, *http.Request) bool { return false }))).(SourceHook); !ok {
			t.Fatal("expected full handler to convert to SourceHook")
		}
		if _, ok := rp.getSource(plainAttrWithValue([]byte("x"))).(SourceBytes); !ok {
			t.Fatal("expected []byte attr to convert to SourceBytes")
		}
		if got := rp.getSource(plainAttrWithValue("plain")); got != "plain" {
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
	res := &resource{
		openJob: gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "style", gox.NewAttrs()),
		kind:    resourceStyle,
	}
	if err := rp.processRes(gox.NewJobHeadClose(ctx, 2, gox.KindRegular, "style"), res); err == nil || !strings.Contains(err.Error(), "missmatch") {
		t.Fatalf("unexpected mismatch error: %v", err)
	}
	if err := rp.processRes(gox.NewJobText(ctx, "bad"), res); err == nil || !strings.Contains(err.Error(), "only raw or byte jobs") {
		t.Fatalf("unexpected invalid content error: %v", err)
	}
}

func TestResourceRenderFallbackAndDumpError(t *testing.T) {
	t.Run("empty script and style dump original tags", func(t *testing.T) {
		ctx, _, _, _ := newPrinterCore(t, true)

		var scriptOut bytes.Buffer
		script := &resource{
			openJob:  gox.NewJobHeadOpen(ctx, 1, gox.KindRegular, "script", gox.NewAttrs()),
			closeJob: gox.NewJobHeadClose(ctx, 1, gox.KindRegular, "script"),
			kind:     resourceScript,
			mode:     resources.ModeHost,
		}
		if err := script.render(defaultPrinter{&scriptOut}); err != nil {
			t.Fatal(err)
		}
		if scriptOut.String() != "<script></script>" {
			t.Fatalf("empty script render = %q", scriptOut.String())
		}

		var styleOut bytes.Buffer
		style := &resource{
			openJob:  gox.NewJobHeadOpen(ctx, 2, gox.KindRegular, "style", gox.NewAttrs()),
			closeJob: gox.NewJobHeadClose(ctx, 2, gox.KindRegular, "style"),
			kind:     resourceStyle,
			mode:     resources.ModeHost,
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
		res := &resource{
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

	t.Run("modulepreload rejects non-void", func(t *testing.T) {
		rp := &resourcePrinter{printer: defaultPrinter{&bytes.Buffer{}}}
		attrs := gox.NewAttrs()
		attrs.Get("rel").Set("modulepreload")
		err := rp.prepareLinkModule(gox.NewJobHeadOpen(ctx, 3, gox.KindRegular, "link", attrs))
		if err == nil || !strings.Contains(err.Error(), "non-void modulepreload") {
			t.Fatalf("unexpected modulepreload non-void error: %v", err)
		}
	})

	t.Run("modulepreload specifier-only string is skipped and registered", func(t *testing.T) {
		var out bytes.Buffer
		rp := NewResourcePrinter(defaultPrinter{&out})
		attrs := gox.NewAttrs()
		attrs.Get("rel").Set("modulepreload")
		attrs.Get("href").Set("/modules/app.js")
		attrs.Get("specifier").Set("app")
		attrs.Get("name").Set("app.js")
		open := gox.NewJobHeadOpen(ctx, 4, gox.KindVoid, "link", attrs)
		if err := rp.Send(open); err != nil {
			t.Fatal(err)
		}
		if got := out.String(); !strings.Contains(got, `rel="modulepreload"`) || !strings.Contains(got, `/modules/app.js`) {
			t.Fatalf("expected modulepreload passthrough output, got %q", got)
		}
		if modules.values["app"] != "/modules/app.js" {
			t.Fatalf("expected modulepreload registry add, got %#v", modules.values)
		}
	})

	t.Run("script scan rewrites data attrs and keeps tag when needed", func(t *testing.T) {
		var out bytes.Buffer
		rp := NewResourcePrinter(defaultPrinter{&out})
		attrs := gox.NewAttrs()
		attrs.Get("src").Set("/modules/app.js")
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
