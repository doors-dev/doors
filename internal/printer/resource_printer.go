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
	"errors"
	"path/filepath"
	"strings"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
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
	if res, ok := r.resource.(*embeddedResource); ok {
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
		props := newScriptProps(false)
		return p.processProps(openJob, props)
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
			props := newStyleProps(true)
			return p.processProps(openJob, props)
		}
		if strings.EqualFold(str, "modulepreload") {
			props := newScriptProps(true)
			return p.processProps(openJob, props)
		}
		return p.scanGenericSrc(openJob)
	case strings.EqualFold(openJob.Tag, "style"):
		props := newStyleProps(false)
		return p.processProps(openJob, props)
	default:
		if openJob.Kind == gox.KindContainer {
			return p.printer.Send(openJob)
		}
		return p.scanGenericSrc(openJob)
	}
}

func (p *resourcePrinter) processProps(open *gox.JobHeadOpen, props props) error {
	match, err := props.Read(open.Attrs)
	if err != nil {
		return err
	}
	if !match {
		return p.printer.Send(open)
	}
	if err := props.Validate(); err != nil {
		return err
	}
	return props.Submit(open, p)
}

func (p *resourcePrinter) scanGenericSrc(openJob *gox.JobHeadOpen) error {
	props := newResourceProps()
	return p.processProps(openJob, props)
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

func (p *resourcePrinter) processRes(job gox.Job, res *embeddedResource) error {
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

type embeddedKind int

const (
	embeddedScript embeddedKind = iota
	embeddedStyle
)

type embeddedResource struct {
	openJob  *gox.JobHeadOpen
	closeJob *gox.JobHeadClose
	kind     embeddedKind
	content  []any
	props    *resourceProps
}

func (r *embeddedResource) appendString(s string) {
	r.content = append(r.content, s)
}

func (r *embeddedResource) appendBytes(b []byte) {
	r.content = append(r.content, b)
}

func (r *embeddedResource) render(p gox.Printer) error {
	core := r.openJob.Context().Value(ctex.KeyCore).(core.Core)
	if r.props.name == "" {
		r.props.name = "inline"
	}
	switch r.kind {
	case embeddedScript:
		r.props.name = resourceFileName(r.props.name, "js")
		return r.renderScript(core, p)
	case embeddedStyle:
		r.props.name = resourceFileName(r.props.name, "css")
		return r.renderStyle(core, p)
	default:
		panic("unknown resource kind")
	}
}

func (r *embeddedResource) renderScript(core core.Core, p gox.Printer) error {
	entry := r.scriptEntry()
	if entry == nil {
		return r.dump(p)
	}
	res, err := core.ResourceRegistry().Script(entry, resources.FormatDefault{}, "", r.props.mode)
	if err != nil {
		return err
	}
	src, err := r.props.resourceURL(core, res)
	if err != nil {
		return err
	}
	r.openJob.Attrs.Get("src").Set(src)
	return r.dump(p)
}

func (r *embeddedResource) renderStyle(core core.Core, p gox.Printer) error {
	entry := r.styleEntry()
	if entry == nil {
		return r.dump(p)
	}
	res, err := core.ResourceRegistry().Style(entry, true, r.props.mode)
	if err != nil {
		return err
	}
	src, err := r.props.resourceURL(core, res)
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

func (r *embeddedResource) scriptEntry() resources.ScriptEntry {
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

func (r *embeddedResource) styleEntry() resources.StyleEntry {
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

func (r *embeddedResource) dump(p gox.Printer) error {
	if err := p.Send(r.openJob); err != nil {
		return err
	}
	if err := p.Send(r.closeJob); err != nil {
		return err
	}
	return nil
}

func resourceFileName(name string, ext string) string {
	if filepath.Ext(name) != "" {
		return name
	}
	return name + "." + ext
}
