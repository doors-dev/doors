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

func newResourcePrinter(printer gox.Printer) gox.Printer {
	return &resourcePrinter{
		printer: printer,
	}
}

type resourcePrinter struct {
	printer  gox.Printer
	resource *resource
}

func (r *resourcePrinter) Send(job gox.Job) error {
	if r.resource == nil {
		return r.scan(job)
	}
	return r.process(job)
}

func (p *resourcePrinter) scan(job gox.Job) error {
	openJob, ok := job.(*gox.JobHeadOpen)
	if !ok {
		return p.printer.Send(job)
	}
	switch true {
	case strings.EqualFold(openJob.Tag, "script"):
		return p.prepareScript(openJob)
	case strings.EqualFold(openJob.Tag, "style"):
		return p.prepareStyle(openJob)
	default:
		return p.printer.Send(job)
	}
}

func (p *resourcePrinter) prepareStyle(job *gox.JobHeadOpen) error {
	mode := resources.ModeHost
	name := ""
	for _, attr := range job.Attrs.List() {
		if strings.EqualFold(attr.Name(), "escape") && attr.IsSet() {
			attr.Unset()
			return p.printer.Send(job)
		}
		if strings.EqualFold(attr.Name(), "name") {
			w := bytes.Buffer{}
			if err := attr.OutputValue(&w); err == nil {
				name = w.String()
			}
		}
		if strings.EqualFold(attr.Name(), "private") && attr.IsSet() {
			attr.Unset()
			mode = resources.ModeCache
		}
		if strings.EqualFold(attr.Name(), "nocache") && attr.IsSet() {
			attr.Unset()
			mode = resources.ModeNoCache
		}
	}
	p.resource = &resource{
		openJob: job,
		kind:    resourceStyle,
		mode:    mode,
		name:    name,
	}
	return nil
}

func (p *resourcePrinter) prepareScript(job *gox.JobHeadOpen) error {
	send := false
	mode := resources.ModeHost
	name := ""
	for _, attr := range job.Attrs.List() {
		if strings.EqualFold(attr.Name(), "escape") && attr.IsSet() {
			send = true
			attr.Unset()
		}
		if strings.EqualFold(attr.Name(), "name") {
			w := bytes.Buffer{}
			if err := attr.OutputValue(&w); err == nil {
				name = w.String()
			}
		}
		if strings.EqualFold(attr.Name(), "private") && attr.IsSet() {
			attr.Unset()
			mode = resources.ModeCache
		}
		if strings.EqualFold(attr.Name(), "nocache") && attr.IsSet() {
			attr.Unset()
			mode = resources.ModeNoCache
		}
		if !send && strings.EqualFold(attr.Name(), "src") && attr.IsSet() {
			send = true
		}
		if !send && strings.EqualFold(attr.Name(), "type") && attr.IsSet() {
			typ, _ := attr.Value().(string)
			if !strings.EqualFold(typ, "text/javascript") && !strings.EqualFold(typ, "application/javascript") {
				send = true
			}
		}
		if name, ok := strings.CutPrefix(attr.Name(), "data:"); ok {
			value := attr.Value()
			attr.Unset()
			front.AttrsSetData(job.Attrs, name, value)
		}
	}
	if send {
		return p.printer.Send(job)
	}
	p.resource = &resource{
		openJob: job,
		kind:    resourceScript,
		mode:    mode,
		name:    name,
	}
	return nil
}

func (p *resourcePrinter) process(job gox.Job) error {
	closeJob, ok := job.(*gox.JobHeadClose)
	if ok {
		if closeJob.ID != p.resource.openJob.ID {
			return errors.New("resource head close id missmatch")
		}
		p.resource.closeJob = closeJob
		r := p.resource
		p.resource = nil
		return r.render(p.printer)
	}
	r, ok := job.(*gox.JobRaw)
	if ok {
		p.resource.appendString(r.Text)
		gox.Release(r)
		return nil
	}
	b, ok := job.(*gox.JobBytes)
	if ok {
		p.resource.appendBytes(b.Bytes)
		gox.Release(b)
		return nil
	}
	return errors.New("style and script must contain only raw or byte jobs")
}

type resourceKind int

const (
	resourceScript resourceKind = iota
	resourceStyle
)

type resource struct {
	openJob  *gox.JobHeadOpen
	closeJob *gox.JobHeadClose
	kind     resourceKind
	mode     resources.ResourceMode
	content  []any
	name     string
}

func (r *resource) appendString(s string) {
	r.content = append(r.content, s)
}

func (r *resource) appendBytes(b []byte) {
	r.content = append(r.content, b)
}

func (r *resource) render(p gox.Printer) error {
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
	res, err := core.ResourceRegistry().Style(entry, true, r.mode)
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
	switch r.mode {
	case resources.ModeHost:
		return core.PathMaker().Resource(res, r.name), nil
	case resources.ModeCache, resources.ModeNoCache:
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			res.Serve(w, r)
			return false
		}, nil)
		if !ok {
			return "", context.Canceled
		}
		return core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.DoorID, r.name), nil
	default:
		panic("unexpected resource type")
	}
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
