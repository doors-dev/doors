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
	"net/http"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

func newStyleProps(rel bool) props {
	return &styleProps{
		output: styleDefault,
		resourceProps: resourceProps{
			mode: resources.ResourceMode(-1),
		},
		rel: rel,
	}
}

type styleProps struct {
	output styleOutput
	resourceProps
	rel bool
}

func (s *styleProps) Submit(job *gox.JobHeadOpen, p *resourcePrinter) error {
	s.cleanAttrs()
	s.setDefaultMode(resources.ModeHost)
	if !s.rel && s.output == styleRaw {
		return p.printer.Send(job)
	}
	if !s.rel {
		p.resource = &embeddedResource{
			openJob: job,
			kind:    embeddedStyle,
			props:   &s.resourceProps,
		}
		return nil
	}
	core := job.Ctx.Value(ctex.KeyCore).(core.Core)
	switch src := s.source.(type) {
	case string:
		return p.printer.Send(job)
	case SourceExternal:
		core.CSPCollector().StyleSource(string(src))
		return p.printer.Send(job)
	case SourceStatic:
		entry := src.styleEntry()
		res, err := core.ResourceRegistry().Style(entry, s.output == styleDefault, s.mode)
		if err != nil {
			return err
		}
		path, err := s.resourceURL(core, res)
		if err != nil {
			return err
		}
		s.sourceAttr.Set(path)
		return p.printer.Send(job)
	case SourceHandler:
		hander := src.Handler()
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			return hander(ctx, w, r)
		}, nil)
		if !ok {
			return context.Canceled
		}
		path := core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.HookID, s.name)
		s.sourceAttr.Set(path)
		return p.printer.Send(job)
	default:
		panic("internal error: unknown style source kind should have been filtered earlier")
	}

}

func (s *styleProps) Validate() error {
	if err := s.resourceProps.Validate(); err != nil {
		return err
	}
	return nil
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
			s.output = styleRaw
			continue
		}
		if s.readName(attr) {
			continue
		}
	}
	if s.sourceKind == sourceUnknown {
		return false, nil
	}
	if s.sourceKind == sourceLink && s.output == styleDefault {
		s.output = styleRaw
	}
	if s.sourceKind == sourceHandler && s.output == styleDefault {
		s.output = styleRaw
	}
	if s.name == "" {
		s.name = "style.css"
	}
	return true, nil
}

type styleOutput string

const (
	styleRaw     styleOutput = "raw"
	styleDefault styleOutput = "default"
)

func (m styleOutput) String() string {
	return string(m)
}
