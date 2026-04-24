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

func newScriptProps(rel bool) props {
	return &scriptProps{
		resourceProps: resourceProps{
			mode: resources.ResourceMode(-1),
		},
		module: rel,
		rel:    rel,
		output: scriptDefault,
	}
}

type scriptProps struct {
	output    scriptOutput
	module    bool
	rel       bool
	profile   string
	specifier string
	ts        bool
	resourceProps
}

func (s *scriptProps) Submit(job *gox.JobHeadOpen, p *resourcePrinter) error {
	s.cleanAttrs()
	s.setDefaultMode(resources.ModeHost)
	if s.sourceKind == sourceUnset && s.output == scriptRaw {
		return p.printer.Send(job)
	}
	if s.sourceKind == sourceUnset && s.output == scriptInline {
		p.resource = &embeddedResource{
			openJob: job,
			kind:    embeddedScript,
			props:   &s.resourceProps,
		}
		return nil
	}
	core := job.Ctx.Value(ctex.KeyCore).(core.Core)
	switch src := s.source.(type) {
	case string:
		if s.specifier != "" {
			core.ModuleRegistry().Add(s.specifier, src)
		}
		return p.printer.Send(job)
	case SourceExternal:
		if s.specifier != "" {
			core.ModuleRegistry().Add(s.specifier, string(src))
		}
		core.CSPCollector().ScriptSource(string(src))
		return p.printer.Send(job)
	case SourceStatic:
		entry := src.scriptEntry(s.output == scriptInline, s.ts)
		format, err := s.output.format(s.module)
		if err != nil {
			return err
		}
		res, err := core.ResourceRegistry().Script(entry, format, s.profile, s.mode)
		if err != nil {
			return err
		}
		path, err := s.resourceURL(core, res)
		if err != nil {
			return err
		}
		if s.specifier != "" {
			core.ModuleRegistry().Add(s.specifier, path)
		}
		s.sourceAttr.Set(path)
		return p.printer.Send(job)
	case SourceHandler:
		handler := src.Handler()
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			return handler(ctx, w, r)
		}, nil)
		if !ok {
			return context.Canceled
		}
		path := core.PathMaker().Hook(core.InstanceID(), hook.HookID, s.name)
		if s.specifier != "" {
			core.ModuleRegistry().Add(s.specifier, path)
		}
		s.sourceAttr.Set(path)
		return p.printer.Send(job)
	default:
		panic("internal error: unknown script source kind should have been filtered earlier")
	}

}

func (s *scriptProps) Validate() error {
	if err := s.resourceProps.Validate(); err != nil {
		return err
	}
	switch true {
	case s.sourceKind == sourceUnset && s.output == scriptBundle:
		return errors.New("inline scripts cannot be bundled; provide src")
	case s.specifier != "" && !s.module:
		return errors.New("specifier is only supported on modules")
	case s.sourceKind == sourceUnset && s.ts:
		return errors.New("embedded scripts do not support TypeScript")
	case s.output == scriptInline && s.module:
		return errors.New("inline scripts do not support modules")
	case s.sourceKind == sourceUnset && s.output == scriptInline && s.ts:
		return errors.New("embedded scripts cannot be compiled from TypeScript")
	case s.rel && s.output == scriptInline:
		return errors.New("modulepreload links do not support inline output")
	case s.output == scriptRaw && s.ts:
		return errors.New("raw output does not support TypeScript sources")
	case s.sourceKind == sourceLink && s.output != scriptRaw:
		return errors.New("direct URL scripts only support raw output")
	case s.sourceKind == sourceHandler && s.output != scriptRaw:
		return errors.New("handler-backed scripts only support raw output")
	}
	return nil
}

func (s *scriptProps) Read(attrs gox.Attrs) (bool, error) {
	for _, attr := range attrs.List() {
		match, err := s.readMode(attr)
		if err != nil {
			return true, err
		}
		if match {
			continue
		}
		match, err = s.readOutput(attr)
		if err != nil {
			return true, err
		}
		if match {
			continue
		}
		if s.rel && s.readHref(attr) {
			continue
		}
		if !s.rel && s.readSrc(attr) {
			continue
		}
		if s.readString(attr, "profile", &s.profile) {
			continue
		}
		if s.readString(attr, "specifier", &s.specifier) {
			continue
		}
		if s.readName(attr) {
			continue
		}
		match, valid := s.readType(attr)
		if !valid {
			return false, nil
		}
		if match {
			continue
		}
		if !s.rel {
			s.readData(attrs, attr)
		}
	}
	if s.sourceKind == sourceUnknown {
		return false, nil
	}
	if s.sourceKind == sourceUnset && s.rel {
		return false, nil
	}
	if s.sourceKind == sourceUnset && s.output == scriptDefault {
		s.output = scriptInline
	}
	if s.sourceKind == sourceHandler && s.output == scriptDefault {
		s.output = scriptRaw
	}
	if s.sourceKind == sourceLink && s.output == scriptDefault {
		s.output = scriptRaw
	}
	if s.name == "" {
		s.name = "script.js"
	}
	return true, nil
}

func (r *scriptProps) readOutput(attr gox.Attr) (bool, error) {
	output := scriptDefault
	switch true {
	case attr.Name() == scriptBundle.String() && isTrue(attr):
		output = scriptBundle
	case attr.Name() == scriptInline.String() && isTrue(attr):
		output = scriptInline
	case attr.Name() == scriptRaw.String() && isTrue(attr):
		output = scriptRaw
	}
	if output == scriptDefault {
		return false, nil
	}
	if r.output != scriptDefault {
		return true, errors.New("only one of raw, inline, or bundle may be set")
	}
	r.output = output
	r.reg(attr)
	return true, nil
}

func (s *scriptProps) readType(attr gox.Attr) (bool, bool) {
	if attr.Name() != "type" {
		return false, true
	}
	if !attr.IsSet() {
		return true, true
	}
	typ, ok := attr.Value().(string)
	if !ok {
		return true, false
	}
	switch true {
	case typ == "", strings.EqualFold(typ, "text/javascript"), strings.EqualFold(typ, "application/javascript"), strings.EqualFold(typ, "javascript"):
		s.reg(attr)
		s.module = false
	case strings.EqualFold(typ, "module"):
		if s.rel {
			s.reg(attr)
		}
		s.module = true
	case strings.EqualFold(typ, "module/javascript"):
		if s.rel {
			s.reg(attr)
		} else {
			attr.Set("module")
		}
		s.module = true
	case strings.EqualFold(typ, "text/typescript"), strings.EqualFold(typ, "application/typescript"),
		strings.EqualFold(typ, "typescript"):
		s.reg(attr)
		s.module = false
		s.ts = true
	case strings.EqualFold(typ, "module/typescript"):
		if s.rel {
			s.reg(attr)
		} else {
			attr.Set("module")
		}
		s.module = true
		s.ts = true
	default:
		return true, false
	}
	return true, true
}

func (s *scriptProps) readData(attrs gox.Attrs, attr gox.Attr) {
	name, ok := strings.CutPrefix(attr.Name(), "data:")
	if !ok {
		return
	}
	value := attr.Value()
	s.reg(attr)
	front.AttrsSetData(attrs, name, value)
}

func isTrue(a gox.Attr) bool {
	if !a.IsSet() {
		return false
	}
	b, ok := a.Value().(bool)
	if !ok {
		return false
	}
	return b
}

type scriptOutput string

const (
	scriptRaw     scriptOutput = "raw"
	scriptInline  scriptOutput = "inline"
	scriptDefault scriptOutput = "default"
	scriptBundle  scriptOutput = "bundle"
)

func (o scriptOutput) String() string {
	return string(o)
}

func (o scriptOutput) format(module bool) (resources.ScriptFormat, error) {
	switch o {
	case scriptRaw:
		return resources.FormatRaw{}, nil
	case scriptDefault:
		if module {
			return resources.FormatModule{}, nil
		}
		return resources.FormatCommon{}, nil
	case scriptBundle:
		if module {
			return resources.FormatModule{Bundle: true}, nil
		}
		return resources.FormatCommon{Bundle: true}, nil
	case scriptInline:
		if module {
			return nil, errors.New("inline output does not support modules")
		}
		return resources.FormatCommon{}, nil
	default:
		panic("internal error: unknown script output")
	}
}
