// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package printer

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

func NewResourcePrinter(printer gox.Printer) gox.Printer {
	return &resourcePrinter{
		printer: printer,
	}
}

type resourcePrinter struct {
	printer  gox.Printer
	resource any
}

func (r *resourcePrinter) Send(job gox.Job) error {
	if res, ok := r.resource.(*resource); ok {
		return r.processRes(job, res)
	}
	if title, ok := r.resource.(*title); ok {
		return r.processTitle(job, title)
	}
	return r.scan(job)
}

func (p *resourcePrinter) scan(job gox.Job) error {
	openJob, ok := job.(*gox.JobHeadOpen)
	if !ok {
		return p.printer.Send(job)
	}
	switch true {
	case strings.EqualFold(openJob.Tag, "title"):
		p.resource = &title{openJob: openJob}
		return nil
	case strings.EqualFold(openJob.Tag, "meta"):
		return p.processMeta(openJob)
	case strings.EqualFold(openJob.Tag, "script"):
		return p.prepareScript(openJob)
	case strings.EqualFold(openJob.Tag, "link"):
		attr, ok := openJob.Attrs.Find("rel")
		if !ok {
			return p.scanGenericSrc(openJob)
		}
		str, ok := attr.Value().(string)
		if !ok {
			return p.scanGenericSrc(openJob)
		}
		if strings.EqualFold(str, "stylesheet") {
			return p.prepareLinkStyle(openJob)
		}
		if strings.EqualFold(str, "modulepreload") {
			return p.prepareLinkModule(openJob)
		}
		return p.scanGenericSrc(openJob)
	case strings.EqualFold(openJob.Tag, "style"):
		return p.prepareStyle(openJob)
	default:
		return p.scanGenericSrc(openJob)
	}

}

func (p *resourcePrinter) scanGenericSrc(openJob *gox.JobHeadOpen) error {
	if openJob.Kind == gox.KindContainer {
		return nil
	}
	attr, ok := openJob.Attrs.Find("src")
	if !ok {
		attr, ok = openJob.Attrs.Find("href")
		if !ok {
			return p.printer.Send(openJob)
		}
	}
	if !attr.IsSet() {
		return p.printer.Send(openJob)
	}
	cache := false
	if a, ok := openJob.Attrs.Find("cache"); ok && a.IsSet() {
		a.Unset()
		cache = true
	}
	typ := ""
	if a, ok := openJob.Attrs.Find("content-type"); ok && a.IsSet() {
		typ, _ = a.Value().(string)
		a.Unset()
	}
	src, ok := p.getSource(attr).(Source)
	if !ok {
		if cache {
			return errors.New("cache attr requires ResourceFS, ResourceLocalFS, ResourceBytes, or ResourceString")
		}
		return p.printer.Send(openJob)
	}
	name := ""
	if nameAttr, ok := openJob.Attrs.Find("name"); ok && nameAttr.IsSet() {
		name, _ = nameAttr.Value().(string)
		nameAttr.Unset()
	}
	if cache {
		static, ok := src.(SourceStatic)
		if !ok {
			return errors.New("cache attr requires ResourceFS, ResourceLocalFS, ResourceBytes, or ResourceString")
		}
		core := openJob.Context().Value(ctex.KeyCore).(core.Core)
		res, err := core.ResourceRegistry().Static(static.StaticEntry(), typ)
		if err != nil {
			return err
		}
		path := core.PathMaker().Resource(res, name)
		attr.Set(path)
		return p.printer.Send(openJob)
	}
	handler := src.Handler()
	if handler == nil {
		return errors.New("source does not provide a handler")
	}
	core := openJob.Context().Value(ctex.KeyCore).(core.Core)
	hook, ok := core.RegisterHook(handler, nil)
	if !ok {
		return context.Canceled
	}
	path := core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.HookID, name)
	attr.Set(path)
	return p.printer.Send(openJob)
}

func (p *resourcePrinter) processMeta(openJob *gox.JobHeadOpen) error {
	if openJob.Kind != gox.KindVoid {
		return errors.New("encountered non-void meta tag")
	}
	property := false
	attr := openJob.Attrs.Get("name")
	if !attr.IsSet() {
		property = true
		attr = openJob.Attrs.Get("property")
		if !attr.IsSet() {
			return p.printer.Send(openJob)
		}
	}
	b := &bytes.Buffer{}
	if err := attr.OutputValue(b); err != nil {
		return err
	}
	attr.Unset()
	name := b.String()
	core := openJob.Context().Value(ctex.KeyCore).(core.Core)
	core.UpdateMeta(name, property, openJob.Attrs.Clone())
	gox.Release(openJob)
	return nil
}

func (p *resourcePrinter) prepareStyle(job *gox.JobHeadOpen) error {
	mode := resources.ModeHost
	name := ""
	output := styleDefault
	for _, attr := range job.Attrs.List() {
		if attr.Name() == "name" {
			name, _ = attr.Value().(string)
			attr.Unset()
			continue
		}
		if attr.Name() == "output" && attr.IsSet() {
			v, ok := p.parseStyleOutput(attr.Value())
			if !ok {
				return errors.New("unexpected style output kind")
			}
			attr.Unset()
			output = v
			continue
		}
		if attr.Name() == "private" && attr.IsSet() {
			attr.Unset()
			mode = resources.ModeCache
			continue
		}
		if attr.Name() == "nocache" && attr.IsSet() {
			attr.Unset()
			mode = resources.ModeNoCache
			continue
		}
	}
	if output == styleRaw {
		return p.printer.Send(job)
	}
	p.resource = &resource{
		openJob:     job,
		kind:        resourceStyle,
		mode:        mode,
		name:        name,
		styleMinify: output == styleMinify,
	}
	return nil
}

func (p *resourcePrinter) prepareLinkStyle(job *gox.JobHeadOpen) error {
	if job.Kind != gox.KindVoid {
		return errors.New("encountered non-void link stylesheet tag")
	}
	mode := resources.ModeHost
	name := ""
	styleMode := styleDefault
	var nameAttr gox.Attr
	var hrefAttr gox.Attr
	for _, attr := range job.Attrs.List() {
		if attr.Name() == "name" {
			nameAttr = attr
			name, _ = attr.Value().(string)
			continue
		}
		if attr.Name() == "output" && attr.IsSet() {
			v, ok := p.parseStyleOutput(attr.Value())
			if !ok {
				return errors.New("unexpected style output kind")
			}
			attr.Unset()
			styleMode = v
			continue
		}
		if attr.Name() == "private" && attr.IsSet() {
			attr.Unset()
			mode = resources.ModeCache
			continue
		}
		if attr.Name() == "nocache" && attr.IsSet() {
			attr.Unset()
			mode = resources.ModeNoCache
			continue
		}
		if attr.Name() == "href" && attr.IsSet() {
			hrefAttr = attr
			continue
		}
	}
	if hrefAttr == nil {
		return p.printer.Send(job)
	}
	if styleMode == styleRaw {
		return p.printer.Send(job)
	}
	if nameAttr != nil {
		nameAttr.Unset()
	}
	core := job.Context().Value(ctex.KeyCore).(core.Core)

	hrefValue := p.getSource(hrefAttr)
	if name == "" {
		if src, ok := hrefValue.(Source); ok {
			name = src.name("css")
		}
	}

	switch src := hrefValue.(type) {
	case string:
		hrefAttr.Set(src)
		return p.printer.Send(job)
	case SourceExternal:
		core.CSPCollector().StyleSource(string(src))
		hrefAttr.Set(string(src))
		return p.printer.Send(job)
	case Source:
		entry := src.styleEntry()
		if entry == nil {
			handler := src.Handler()
			if handler == nil {
				return p.printer.Send(job)
			}
			hook, ok := core.RegisterHook(handler, nil)
			if !ok {
				return context.Canceled
			}
			path := core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.HookID, name)
			hrefAttr.Set(path)
			return p.printer.Send(job)
		}
		res, err := core.ResourceRegistry().Style(entry, styleMode == styleMinify, mode)
		if err != nil {
			return err
		}
		path, err := resourceURL(core, res, mode, name)
		if err != nil {
			return err
		}
		hrefAttr.Set(path)
		return p.printer.Send(job)
	default:
		return p.printer.Send(job)
	}
}

func (p *resourcePrinter) prepareLinkModule(job *gox.JobHeadOpen) error {
	scan := scriptRefScan{
		p:      p,
		job:    job,
		kind:   scriptRefModulePreload,
		mode:   resources.ModeHost,
		output: scriptDefault,
		module: true,
	}
	if err := scan.scan(); err != nil {
		return err
	}
	return scan.build()
}

func (p *resourcePrinter) parseStyleOutput(v any) (styleOutput, bool) {
	str, _ := v.(string)
	switch true {
	case strings.EqualFold(str, styleRaw.String()):
		return styleRaw, true
	case strings.EqualFold(str, styleMinify.String()):
		return styleMinify, true
	case str == "", strings.EqualFold(str, styleDefault.String()):
		return styleDefault, true
	default:
		return "", false
	}
}

type scriptOutout string

const (
	scriptRaw     scriptOutout = "raw"
	scriptInline  scriptOutout = "inline"
	scriptDefault scriptOutout = "default"
	scriptBundle  scriptOutout = "bundle"
)

func (o scriptOutout) String() string {
	return string(o)
}

type styleOutput string

const (
	styleRaw     styleOutput = "raw"
	styleDefault styleOutput = "default"
	styleMinify  styleOutput = "minify"
)

func (m styleOutput) String() string {
	return string(m)
}

func (o scriptOutout) format(module bool) (resources.ScriptFormat, error) {
	switch o {
	case scriptRaw:
		return resources.FormatRaw{}, nil
	case scriptDefault:
		if module {
			return resources.FormatModule{}, nil
		}
		return resources.FormatDefault{}, nil
	case scriptBundle:
		if module {
			return resources.FormatModule{Bundle: true}, nil
		}
		return resources.FormatCommon{Bundle: true}, nil
	case scriptInline:
		if module {
			return nil, errors.New("inline script can't be module")
		}
		return resources.FormatDefault{}, nil
	default:
		panic("unknown format")
	}
}

func resourceURL(core core.Core, res *resources.Resource, mode resources.ResourceMode, name string) (string, error) {
	switch mode {
	case resources.ModeHost:
		return core.PathMaker().Resource(res, name), nil
	case resources.ModeCache, resources.ModeNoCache:
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			res.Serve(w, r)
			return false
		}, nil)
		if !ok {
			return "", context.Canceled
		}
		return core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.HookID, name), nil
	default:
		panic("unexpected resource type")
	}
}

func (p *resourcePrinter) parseScriptOutput(v any) (scriptOutout, bool) {
	str, ok := v.(string)
	if !ok {
		return "", false
	}
	switch true {
	case strings.EqualFold(str, scriptRaw.String()):
		return scriptRaw, true
	case strings.EqualFold(str, scriptInline.String()):
		return scriptInline, true
	case strings.EqualFold(str, scriptBundle.String()):
		return scriptBundle, true
	case str == "", strings.EqualFold(str, scriptDefault.String()):
		return scriptDefault, true
	default:
		return "", false
	}
}

func (p *resourcePrinter) getSource(attr gox.Attr) any {
	value := attr.Value()

	if handler, ok := value.(HandlerFunc); ok {
		return SourceHook(handler)
	}

	if hander, ok := value.(HandlerSimpleFunc); ok {
		return SourceHook(func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
			hander(w, r)
			return false
		})
	}

	if bytes, ok := value.([]byte); ok {
		return SourceBytes(bytes)
	}

	return value
}

func (p *resourcePrinter) prepareScript(job *gox.JobHeadOpen) error {
	scan := scriptRefScan{
		p:      p,
		job:    job,
		kind:   scriptRefScript,
		mode:   resources.ModeHost,
		output: scriptDefault,
	}
	if err := scan.scan(); err != nil {
		return err
	}
	return scan.build()
}

type scriptRefKind int

const (
	scriptRefScript scriptRefKind = iota
	scriptRefModulePreload
)

type scriptRefScan struct {
	p          *resourcePrinter
	job        *gox.JobHeadOpen
	kind       scriptRefKind
	mode       resources.ResourceMode
	name       string
	nameAttr   gox.Attr
	sourceAttr gox.Attr
	ts         bool
	unknown    bool
	output     scriptOutout
	specifier  string
	module     bool
	profile    string
	keepTag    bool
}

func (s *scriptRefScan) scan() error {
	if s.kind == scriptRefModulePreload && s.job.Kind != gox.KindVoid {
		return errors.New("encountered non-void modulepreload link tag")
	}
	sourceName := "src"
	if s.kind == scriptRefModulePreload {
		sourceName = "href"
	}
	for _, attr := range s.job.Attrs.List() {
		if attr.Name() == "output" && attr.IsSet() {
			v, ok := s.p.parseScriptOutput(attr.Value())
			if !ok {
				return errors.New("unknown script output")
			}
			attr.Unset()
			s.output = v
			continue
		}
		if attr.Name() == "name" {
			s.name, _ = attr.Value().(string)
			if s.kind == scriptRefModulePreload {
				s.nameAttr = attr
			} else {
				attr.Unset()
			}
			continue
		}
		if attr.Name() == "profile" {
			s.profile, _ = attr.Value().(string)
			attr.Unset()
			continue
		}
		if attr.Name() == "specifier" {
			s.specifier, _ = attr.Value().(string)
			attr.Unset()
			continue
		}
		if attr.Name() == "private" && attr.IsSet() {
			attr.Unset()
			s.mode = resources.ModeCache
			continue
		}
		if attr.Name() == "nocache" && attr.IsSet() {
			attr.Unset()
			s.mode = resources.ModeNoCache
			continue
		}
		if attr.Name() == sourceName && attr.IsSet() {
			s.sourceAttr = attr
			continue
		}
		if attr.Name() == "type" && attr.IsSet() {
			typ, _ := attr.Value().(string)
			switch s.kind {
			case scriptRefScript:
				switch true {
				case typ == "", strings.EqualFold(typ, "text/javascript"), strings.EqualFold(typ, "application/javascript"),
					strings.EqualFold(typ, "javascript"):
					attr.Unset()
				case strings.EqualFold(typ, "module"):
					s.module = true
				case strings.EqualFold(typ, "module/javascript"):
					attr.Set("module")
					s.module = true
				case strings.EqualFold(typ, "text/typescript"), strings.EqualFold(typ, "application/typescript"),
					strings.EqualFold(typ, "typescript"):
					attr.Unset()
					s.ts = true
				case strings.EqualFold(typ, "module/typescript"):
					attr.Set("module")
					s.ts = true
					s.module = true
				default:
					s.unknown = true
				}
			case scriptRefModulePreload:
				switch true {
				case typ == "", strings.EqualFold(typ, "text/javascript"), strings.EqualFold(typ, "application/javascript"),
					strings.EqualFold(typ, "javascript"), strings.EqualFold(typ, "module"),
					strings.EqualFold(typ, "module/javascript"):
					attr.Unset()
				case strings.EqualFold(typ, "text/typescript"), strings.EqualFold(typ, "application/typescript"),
					strings.EqualFold(typ, "typescript"), strings.EqualFold(typ, "module/typescript"):
					attr.Unset()
					s.ts = true
				default:
					s.unknown = true
				}
			default:
				panic("unexpected script ref kind")
			}
			continue
		}
		if s.kind == scriptRefScript {
			if name, ok := strings.CutPrefix(attr.Name(), "data:"); ok {
				value := attr.Value()
				attr.Unset()
				front.AttrsSetData(s.job.Attrs, name, value)
				continue
			}
		}
		if s.kind == scriptRefModulePreload && attr.Name() == "rel" {
			continue
		}
		if strings.HasPrefix(attr.Name(), "data-d0") {
			continue
		}
		if attr.IsSet() {
			s.keepTag = true
		}
	}
	return nil
}

func (s *scriptRefScan) build() error {
	attrName := "src"
	if s.kind == scriptRefModulePreload {
		attrName = "href"
	}
	if s.specifier != "" {
		if s.kind == scriptRefScript {
			s.module = true
			s.unknown = false
			s.job.Attrs.Get("type").Set("module")
		}
	}
	only := s.kind == scriptRefScript && s.specifier != "" && !s.keepTag
	if s.unknown {
		return s.p.printer.Send(s.job)
	}
	if s.sourceAttr == nil {
		if s.kind == scriptRefModulePreload {
			return s.p.printer.Send(s.job)
		}
		if s.output == scriptRaw {
			return s.p.printer.Send(s.job)
		}
		if s.output == scriptBundle {
			return errors.New("inline scripts can't be bundeled, provide src")
		}
		if s.module {
			return errors.New("inline modules are not supported, provide src")
		}
		if s.ts {
			return errors.New("inline typescript is not supported, provide src")
		}
		s.p.resource = &resource{
			openJob: s.job,
			kind:    resourceScript,
			mode:    s.mode,
			name:    s.name,
		}
		return nil
	}
	if s.kind == scriptRefModulePreload && s.output == scriptInline {
		return errors.New("inline modulepreload is not supported")
	}
	if s.output == scriptRaw && s.ts {
		return errors.New("raw typescript can't be served")
	}
	core := s.job.Context().Value(ctex.KeyCore).(core.Core)
	sourceValue := s.p.getSource(s.sourceAttr)
	if s.name == "" {
		if src, ok := sourceValue.(Source); ok {
			s.name = src.name("js")
		}
	}
	switch src := sourceValue.(type) {
	case string:
		if s.ts {
			return errors.New("can't compile typescript with regular " + attrName)
		}
		if s.output == scriptBundle {
			return errors.New("can't bundle script with regular " + attrName)
		}
		if s.kind == scriptRefScript && s.output == scriptInline {
			return errors.New("can't prepare \"inline\" script with regular src")
		}
		if s.specifier != "" {
			core.ModuleRegistry().Add(s.specifier, src)
		}
		if only {
			if s.kind == scriptRefModulePreload && s.nameAttr != nil {
				s.nameAttr.Unset()
			}
			s.p.resource = &resource{
				openJob: s.job,
				kind:    resourceSkip,
			}
			return nil
		}
		if s.kind == scriptRefModulePreload && s.nameAttr != nil {
			s.nameAttr.Unset()
		}
		s.sourceAttr.Set(src)
		return s.p.printer.Send(s.job)
	case SourceHook:
		if s.output == scriptBundle || s.ts || (s.kind == scriptRefScript && s.output == scriptInline) {
			return errors.New("can't bundle or compile ts from hook source")
		}
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			return src.Handler()(ctx, w, r)
		}, nil)
		if !ok {
			return context.Canceled
		}
		path := core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.HookID, s.name)
		if s.specifier != "" {
			core.ModuleRegistry().Add(s.specifier, path)
		}
		if only {
			if s.kind == scriptRefModulePreload && s.nameAttr != nil {
				s.nameAttr.Unset()
			}
			s.p.resource = &resource{
				openJob: s.job,
				kind:    resourceSkip,
			}
			return nil
		}
		if s.kind == scriptRefModulePreload && s.nameAttr != nil {
			s.nameAttr.Unset()
		}
		s.sourceAttr.Set(path)
		return s.p.printer.Send(s.job)
	case SourceProxy:
		if s.output == scriptBundle || s.ts || (s.kind == scriptRefScript && s.output == scriptInline) {
			return errors.New("can't bundle or compile ts from proxy source")
		}
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			return src.Handler()(ctx, w, r)
		}, nil)
		if !ok {
			return context.Canceled
		}
		path := core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.HookID, s.name)
		if s.specifier != "" {
			core.ModuleRegistry().Add(s.specifier, path)
		}
		if only {
			if s.kind == scriptRefModulePreload && s.nameAttr != nil {
				s.nameAttr.Unset()
			}
			s.p.resource = &resource{
				openJob: s.job,
				kind:    resourceSkip,
			}
			return nil
		}
		if s.kind == scriptRefModulePreload && s.nameAttr != nil {
			s.nameAttr.Unset()
		}
		s.sourceAttr.Set(path)
		return s.p.printer.Send(s.job)
	case SourceExternal:
		if s.ts {
			return errors.New("can't compile typescript with external " + attrName)
		}
		if s.output == scriptBundle {
			return errors.New("can't bundle script with external " + attrName)
		}
		if s.kind == scriptRefScript && s.output == scriptInline {
			return errors.New("can't prepare \"inline\" script with extarnal src")
		}
		if s.specifier != "" {
			core.ModuleRegistry().Add(s.specifier, string(src))
		}
		core.CSPCollector().ScriptSource(string(src))
		if only {
			if s.kind == scriptRefModulePreload && s.nameAttr != nil {
				s.nameAttr.Unset()
			}
			s.p.resource = &resource{
				openJob: s.job,
				kind:    resourceSkip,
			}
			return nil
		}
		if s.kind == scriptRefModulePreload && s.nameAttr != nil {
			s.nameAttr.Unset()
		}
		s.sourceAttr.Set(string(src))
		return s.p.printer.Send(s.job)
	case Source:
		entry := src.scriptEntry(s.kind == scriptRefScript && s.output == scriptInline, s.ts)
		format, err := s.output.format(s.module)
		if err != nil {
			return err
		}
		res, err := core.ResourceRegistry().Script(entry, format, s.profile, s.mode)
		if err != nil {
			return err
		}
		path, err := resourceURL(core, res, s.mode, s.name)
		if err != nil {
			return err
		}
		if s.specifier != "" {
			core.ModuleRegistry().Add(s.specifier, path)
		}
		if only {
			if s.kind == scriptRefModulePreload && s.nameAttr != nil {
				s.nameAttr.Unset()
			}
			s.p.resource = &resource{
				openJob: s.job,
				kind:    resourceSkip,
			}
			return nil
		}
		if s.kind == scriptRefModulePreload && s.nameAttr != nil {
			s.nameAttr.Unset()
		}
		s.sourceAttr.Set(path)
		return s.p.printer.Send(s.job)
	default:
		if s.ts || (s.output != scriptDefault && s.output != scriptRaw) {
			if s.kind == scriptRefScript {
				return errors.New("unknown type of src attribute on script")
			}
			return errors.New("unknown type of href attribute on modulepreload link")
		}
		return s.p.printer.Send(s.job)
	}
}

func (p *resourcePrinter) processRes(job gox.Job, res *resource) error {
	closeJob, ok := job.(*gox.JobHeadClose)
	if ok {
		if closeJob.ID != res.openJob.ID {
			return errors.New("resource head close id missmatch")
		}
		res.closeJob = closeJob
		p.resource = nil
		return res.render(p.printer)
	}
	r, ok := job.(*gox.JobRaw)
	if ok {
		res.appendString(r.Text)
		gox.Release(r)
		return nil
	}
	b, ok := job.(*gox.JobBytes)
	if ok {
		res.appendBytes(b.Bytes)
		gox.Release(b)
		return nil
	}
	return errors.New("style and script must contain only raw or byte jobs")
}

type resourceKind int

const (
	resourceScript resourceKind = iota
	resourceStyle
	resourceSkip
)

type resource struct {
	openJob     *gox.JobHeadOpen
	closeJob    *gox.JobHeadClose
	kind        resourceKind
	mode        resources.ResourceMode
	content     []any
	name        string
	styleMinify bool
}

func (r *resource) appendString(s string) {
	r.content = append(r.content, s)
}

func (r *resource) appendBytes(b []byte) {
	r.content = append(r.content, b)
}

func (r *resource) render(p gox.Printer) error {
	if r.kind == resourceSkip {
		gox.Release(r.openJob)
		gox.Release(r.closeJob)
		return nil
	}
	core := r.openJob.Context().Value(ctex.KeyCore).(core.Core)
	if r.name == "" {
		r.name = "inline"
	}
	switch r.kind {
	case resourceScript:
		r.name += ".js"
		return r.renderScript(core, p)
	case resourceStyle:
		r.name += ".css"
		return r.renderStyle(core, p)
	default:
		panic("unknown resource kind")
	}
}

func (r *resource) renderScript(core core.Core, p gox.Printer) error {
	entry := r.scriptEntry()
	if entry == nil {
		return r.dump(p)
	}
	res, err := core.ResourceRegistry().Script(entry, resources.FormatDefault{}, "", r.mode)
	if err != nil {
		return err
	}
	src, err := r.src(core, res)
	if err != nil {
		return err
	}
	r.openJob.Attrs.Get("src").Set(src)
	return r.dump(p)
}

func (r *resource) renderStyle(core core.Core, p gox.Printer) error {
	entry := r.styleEntry()
	if entry == nil {
		return r.dump(p)
	}
	res, err := core.ResourceRegistry().Style(entry, r.styleMinify, r.mode)
	if err != nil {
		return err
	}
	src, err := r.src(core, res)
	if err != nil {
		return err
	}
	r.openJob.Kind = gox.KindVoid
	r.openJob.Tag = "link"
	r.openJob.Attrs.Get("rel").Set("stylesheet")
	r.openJob.Attrs.Get("href").Set(src)
	gox.Release(r.closeJob)
	return p.Send(r.openJob)
}

func (r *resource) scriptEntry() resources.ScriptEntry {
	if len(r.content) == 0 {
		return nil
	}
	if len(r.content) == 1 {
		switch c := r.content[0].(type) {
		case string:
			return resources.ScriptInlineString{
				Content: c,
				Kind:    resources.KindJS,
			}
		case []byte:
			return resources.ScriptInlineBytes{
				Content: c,
				Kind:    resources.KindJS,
			}
		default:
			panic("unexpected content kind")
		}
	}
	buf := strings.Builder{}
	for _, c := range r.content {
		switch c := c.(type) {
		case string:
			buf.WriteString(c)
		case []byte:
			buf.Write(c)
		default:
			panic("unexpected content kind")
		}
	}
	return resources.ScriptInlineString{
		Content: buf.String(),
		Kind:    resources.KindJS,
	}
}

func (r *resource) styleEntry() resources.StyleEntry {
	if len(r.content) == 0 {
		return nil
	}
	if len(r.content) == 1 {
		switch c := r.content[0].(type) {
		case string:
			return resources.StyleString{
				Content: c,
			}
		case []byte:
			return resources.StyleBytes{
				Content: c,
			}
		default:
			panic("unexpected content kind")
		}
	}
	buf := bytes.Buffer{}
	for _, c := range r.content {
		switch c := c.(type) {
		case string:
			buf.WriteString(c)
		case []byte:
			buf.Write(c)
		default:
			panic("unexpected content kind")
		}
	}
	return resources.StyleBytes{
		Content: buf.Bytes(),
	}
}

func (r *resource) src(core core.Core, res *resources.Resource) (string, error) {
	return resourceURL(core, res, r.mode, r.name)
}

func (r *resource) dump(p gox.Printer) error {
	if err := p.Send(r.openJob); err != nil {
		return err
	}
	if err := p.Send(r.closeJob); err != nil {
		return err
	}
	return nil
}

func (r *resourcePrinter) processTitle(j gox.Job, tit *title) error {
	if _, ok := j.(*gox.JobHeadOpen); ok {
		return errors.New("title can't contain other tags")
	}
	if _, ok := j.(*gox.JobComp); ok {
		panic("components are not expected here, must pe processed via pipe")
	}
	if closeJob, ok := j.(*gox.JobHeadClose); ok {
		if closeJob.ID != tit.openJob.ID {
			return errors.New("unexpected close job")
		}
		core := j.Context().Value(ctex.KeyCore).(core.Core)
		content := tit.buf.String()
		attrs := tit.openJob.Attrs.Clone()
		core.UpdateTitle(content, attrs)
		gox.Release(tit.openJob)
		gox.Release(closeJob)
		r.resource = nil
		return nil
	}
	return j.Output(&tit.buf)
}

type title struct {
	openJob *gox.JobHeadOpen
	buf     bytes.Buffer
}
