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

func newScriptProps(rel bool) props {
	s := &scriptProps{rel: rel}
	s.target = &s.res
	s.SetModule(rel)
	return s
}

type scriptProps struct {
	Script
	tagAttrs
	rel       bool
	specifier string
	data      []gox.Attr
}

func (s *scriptProps) Submit(job *gox.JobOpen, p *resourcePrinter) error {
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
	c := job.Ctx.Value(common.KeyCore).(core.Core)
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

func (s *scriptProps) Validate() error {
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
	if s.srcKind == sourceUnknown {
		return false, nil
	}
	if s.srcKind == sourceUnset && s.rel {
		return false, nil
	}
	for _, attr := range s.data {
		if err := s.readData(attrs, attr); err != nil {
			return true, err
		}
	}
	return true, nil
}

func (s *scriptProps) readOutput(attr gox.Attr) (bool, error) {
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

func (s *scriptProps) readData(attrs gox.Attrs, attr gox.Attr) error {
	name := strings.TrimPrefix(attr.Name(), "data:")
	value := attr.Value()
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
