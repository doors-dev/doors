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

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

type props interface {
	Read(gox.Attrs) (bool, error)
	Validate() error
	Submit(*gox.JobHeadOpen, *resourcePrinter) error
}

func newResourceProps() props {
	return &resourceProps{
		mode: resources.ResourceMode(-1),
	}
}

type sourceKind int

const (
	sourceUnset sourceKind = iota
	sourceLink
	sourceHandler
	sourceStatic
	sourceUnknown
)

type resourceProps struct {
	mode         resources.ResourceMode
	attrsToClean []gox.Attr
	source       any
	sourceAttr   gox.Attr
	sourceKind   sourceKind
	name         string
	contentType  string
}

func (r *resourceProps) Read(attrs gox.Attrs) (bool, error) {
	attr, ok := attrs.Find("src")
	if !ok {
		attr, ok = attrs.Find("href")
		if !ok {
			return false, nil
		}
	}
	r.readSource(attr)
	if r.sourceKind == sourceUnset {
		panic("internal error: source kind is unset after reading a resource source")
	}
	if r.sourceKind == sourceUnknown {
		return false, nil
	}
	if r.sourceKind == sourceLink {
		return false, nil
	}
	for _, attr := range attrs.List() {
		match, err := r.readMode(attr)
		if err != nil {
			return true, err
		}
		if match {
			continue
		}
		if r.readName(attr) {
			continue
		}
		if attr.Name() == "type" && attr.IsSet() {
			r.contentType, _ = attr.Value().(string)
		}
	}
	if r.mode == resources.ModeNoHost {
		r.mode = resources.ModeNoCache
	}
	return true, nil
}

func (r *resourceProps) Validate() error {
	switch true {
	case r.sourceKind == sourceLink && isModeSet(r.mode):
		return errors.New("direct URL sources do not support cache, private, or nocache")
	case r.sourceKind == sourceHandler && isModeSet(r.mode):
		return errors.New("handler-backed sources do not support cache, private, or nocache")
	case r.sourceKind == sourceHandler && r.contentType != "":
		return errors.New("handler-backed sources do not support the type attribute")
	}
	return nil
}

func (r *resourceProps) Submit(openJob *gox.JobHeadOpen, p *resourcePrinter) error {
	r.setDefaultMode(resources.ModeNoCache)
	sourceHandler := r.source.(SourceHandler)
	sourceStatic, isStatic := r.source.(SourceStatic)
	core := openJob.Context().Value(ctex.KeyCore).(core.Core)
	if r.mode == resources.ModeNoCache || !isStatic {
		handler := sourceHandler.Handler()
		contentType := r.contentType
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			if contentType != "" {
				w.Header().Set("Content-Type", contentType)
			}
			return handler(ctx, w, r)
		}, nil)
		if !ok {
			return context.Canceled
		}
		path := core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.HookID, r.name)
		r.sourceAttr.Set(path)
		return p.printer.Send(openJob)
	}
	res, err := core.ResourceRegistry().Static(sourceStatic.StaticEntry(), r.contentType)
	if err != nil {
		return err
	}
	path := core.PathMaker().Resource(res, r.name)
	r.sourceAttr.Set(path)
	return p.printer.Send(openJob)
}

func (r *resourceProps) setDefaultMode(defaultMode resources.ResourceMode) {
	if isModeSet(r.mode) {
		return
	}
	r.mode = defaultMode
}

func (r *resourceProps) reg(a gox.Attr) {
	r.attrsToClean = append(r.attrsToClean, a)
}

func (r *resourceProps) cleanAttrs() {
	for _, attr := range r.attrsToClean {
		attr.Unset()
	}
	r.attrsToClean = nil
}

func (r *resourceProps) readSource(attr gox.Attr) {
	if !attr.IsSet() {
		return
	}
	r.sourceAttr = attr
	switch value := attr.Value().(type) {
	case HandlerFunc:
		r.source = SourceHook(value)
		r.sourceKind = sourceHandler
	case HandlerSimpleFunc:
		r.source = SourceHook(func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
			value(w, r)
			return false
		})
		r.sourceKind = sourceHandler
	case string:
		r.source = value
		r.sourceKind = sourceLink
	case SourceExternal:
		r.source = value
		r.sourceKind = sourceLink
	case []byte:
		r.source = SourceBytes(value)
		r.sourceKind = sourceStatic
	case SourceStatic:
		r.source = value
		r.sourceKind = sourceStatic
	case SourceHandler:
		r.source = value
		r.sourceKind = sourceHandler
	default:
		r.sourceKind = sourceUnknown
	}
}

func (r *resourceProps) resourceURL(core core.Core, res *resources.Resource) (string, error) {
	mode := r.mode
	switch mode {
	case resources.ModeHost:
		return core.PathMaker().Resource(res, r.name), nil
	case resources.ModeNoHost, resources.ModeNoCache:
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			res.Serve(w, r)
			return false
		}, nil)
		if !ok {
			return "", context.Canceled
		}
		return core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.HookID, r.name), nil
	default:
		panic("internal error: unexpected resource mode")
	}
}

func (r *resourceProps) readString(attr gox.Attr, name string, target *string) bool {
	if attr.Name() != name {
		return false
	}
	value, ok := attr.Value().(string)
	if !ok {
		return true
	}
	*target = value
	r.reg(attr)
	return true
}

func (r *resourceProps) readSrc(attr gox.Attr) bool {
	if attr.Name() != "src" {
		return false
	}
	r.readSource(attr)
	return true
}
func (r *resourceProps) readHref(attr gox.Attr) bool {
	if attr.Name() != "href" {
		return false
	}
	r.readSource(attr)
	return true
}

func (r *resourceProps) readName(attr gox.Attr) bool {
	return r.readString(attr, "name", &r.name)
}

func (r *resourceProps) readMode(attr gox.Attr) (bool, error) {
	mode := resources.ResourceMode(-1)
	switch true {
	case attr.Name() == "private" && isTrue(attr):
		mode = resources.ModeNoHost
	case attr.Name() == "nocache" && isTrue(attr):
		mode = resources.ModeNoCache
	case attr.Name() == "cache" && isTrue(attr):
		mode = resources.ModeHost
	}
	if !isModeSet(mode) {
		return false, nil
	}
	if isModeSet(r.mode) {
		return true, errors.New("only one of cache, private, or nocache may be set")
	}
	r.mode = mode
	r.reg(attr)
	return true, nil
}

func isModeSet(mode resources.ResourceMode) bool {
	return int(mode) != -1
}
