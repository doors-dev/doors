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
	"fmt"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/gox"
)

func newStyleProps(rel bool) props {
	s := &styleProps{rel: rel}
	s.target = &s.res
	return s
}

type styleProps struct {
	Style
	tagAttrs
	rel bool
}

func (s *styleProps) Submit(job *gox.JobOpen, p *resourcePrinter) error {
	s.cleanAttrs()
	if !s.rel && s.output == styleRaw {
		return p.printer.Send(job)
	}
	if !s.rel {
		p.resource = &embeddedResource{
			openJob: job,
			body:    &s.Style,
		}
		return nil
	}
	c := job.Ctx.Value(common.KeyCore).(core.Core)
	switch src := s.source.(type) {
	case string:
		return p.printer.Send(job)
	case SourceExternal:
		if document, _ := job.Ctx.Value(common.KeyDocument).(bool); document {
			c.Instance().CSPCollector().StyleSource(string(src))
		}
		return p.printer.Send(job)
	case SourceHandler:
		path, err := s.Resolve(c)
		if err != nil {
			return err
		}
		s.sourceAttr.Set(path)
		return p.printer.Send(job)
	default:
		common.Logger(job.Ctx).Error(
			"non-compatible src/href value",
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

func (s *styleProps) Read(attrs gox.Attrs) (bool, error) {
	for _, attr := range attrs.List() {
		match, err := s.readMode(attr)
		if err != nil {
			return true, err
		}
		if match {
			continue
		}
		if s.rel && s.readHref(attr) {
			continue
		}
		if attr.Name() == styleRaw.String() && isTrue(attr) {
			s.reg(attr)
			s.SetRaw()
			continue
		}
		if s.readName(attr) {
			continue
		}
	}
	if s.srcKind == sourceUnknown || s.rel && s.srcKind == sourceUnset {
		return false, nil
	}
	return true, nil
}
