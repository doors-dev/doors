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
	"errors"
	"fmt"
	"strings"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/front"
	"github.com/doors-dev/gox"
)

func newScriptTag(rel bool) resourceTag {
	s := &scriptTag{rel: rel}
	s.tagBase = tagBase{target: &s.Script, tag: "script"}
	if rel {
		s.tag = "link rel=modulepreload"
	}
	s.SetModule(rel)
	return s
}

type scriptTag struct {
	Script
	tagBase
	rel       bool
	specifier string
	data      []gox.Attr
}

func (s *scriptTag) Submit(job *gox.JobOpen, p *resourcePrinter) error {
	c := job.Ctx.Value(common.KeyCore).(core.Core)
	for _, attr := range s.data {
		if err := s.readData(job.Attrs, attr, c); err != nil {
			return err
		}
	}
	s.cleanAttrs()
	if s.srcKind == sourceUnset && s.output == scriptRaw {
		return p.printer.Send(job)
	}
	if s.srcKind == sourceUnset && s.output == scriptInline {
		p.resource = &embeddedResource{
			openJob: job,
			body:    &s.Script,
		}
		return nil
	}
	document, _ := job.Ctx.Value(common.KeyDocument).(bool)
	switch src := s.source.(type) {
	case string:
		if document && s.specifier != "" {
			c.Instance().ModuleRegistry().Add(s.specifier, src)
		}
		return p.printer.Send(job)
	case SourceExternal:
		if document && s.specifier != "" {
			c.Instance().ModuleRegistry().Add(s.specifier, string(src))
		}
		if document {
			c.Instance().CSPCollector().ScriptSource(string(src))
		}
		return p.printer.Send(job)
	case SourceHandler:
		path, err := s.Resolve(c)
		if err != nil {
			return err
		}
		if document && s.specifier != "" {
			c.Instance().ModuleRegistry().Add(s.specifier, path)
		}
		s.sourceAttr.Set(path)
		return p.printer.Send(job)
	default:
		common.Logger(job.Ctx).Error(
			"internal error: unknown script source kind should have been filtered earlier",
			"source_kind",
			s.srcKind,
			"source_type",
			fmt.Sprintf("%T", s.source),
		)
		if s.sourceAttr != nil {
			s.sourceAttr.Unset()
		}
		return p.printer.Send(job)
	}

}

func (s *scriptTag) Validate() error {
	if err := s.Script.Validate(); err != nil {
		return err
	}
	switch true {
	case s.specifier != "" && !s.module:
		return errors.New("specifier is only supported on modules")
	case s.rel && s.output == scriptInline:
		return errors.New("modulepreload links do not support inline output")
	}
	return nil
}

func (s *scriptTag) Read(attrs gox.Attrs) (bool, error) {
	source := "src"
	if s.rel {
		source = "href"
	}
	stage := false
	if attr, ok := attrs.Find(source); ok {
		_, stage = attr.Value().(*Script)
		if err := s.setSource(attr); err != nil {
			return true, err
		}
		if stage && !s.rel && s.module && !attrs.Has("type") {
			attrs.Get("type").Set("module")
		}
	}
	for attr := range attrs.All() {
		if attr.Name() == source {
			continue
		}
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
		if s.readString(attr, "profile", s.SetProfile) {
			continue
		}
		if s.readString(attr, "specifier", func(v string) { s.specifier = v }) {
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
		if !s.rel && strings.HasPrefix(attr.Name(), "data:") {
			s.data = append(s.data, attr)
		}
	}
	if stage && s.rel && !s.module {
		return true, errors.New("modulepreload links require a module script resource")
	}
	if s.srcKind == sourceUnknown {
		return false, nil
	}
	if s.srcKind == sourceUnset && s.rel {
		return false, nil
	}
	return true, nil
}

func (s *scriptTag) readOutput(attr gox.Attr) (bool, error) {
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
	if err := s.SetOutput(output); err != nil {
		return true, err
	}
	s.reg(attr)
	return true, nil
}

func (s *scriptTag) readType(attr gox.Attr) (bool, bool) {
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
		s.SetModule(false)
	case strings.EqualFold(typ, "module"):
		if s.rel {
			s.reg(attr)
		}
		s.SetModule(true)
	case strings.EqualFold(typ, "module/javascript"):
		if s.rel {
			s.reg(attr)
		} else {
			attr.Set("module")
		}
		s.SetModule(true)
	case strings.EqualFold(typ, "text/typescript"), strings.EqualFold(typ, "application/typescript"),
		strings.EqualFold(typ, "typescript"):
		s.reg(attr)
		s.SetModule(false)
		s.SetTS(true)
	case strings.EqualFold(typ, "module/typescript"):
		if s.rel {
			s.reg(attr)
		} else {
			attr.Set("module")
		}
		s.SetModule(true)
		s.SetTS(true)
	default:
		return true, false
	}
	return true, true
}

func (s *scriptTag) readData(attrs gox.Attrs, attr gox.Attr, c core.Core) error {
	name := strings.TrimPrefix(attr.Name(), "data:")
	value := attr.Value()
	var r Resource
	switch v := value.(type) {
	case Resource:
		r = v
	case SourceStatic, SourceHandler:
		a := &Asset{}
		a.setSource(v)
		r = a
	}
	if r != nil {
		if err := r.Validate(); err != nil {
			return err
		}
		url, err := r.Resolve(c)
		if err != nil {
			return err
		}
		value = url
	}
	s.reg(attr)
	return front.AttrsSetData(attrs, name, value)
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
