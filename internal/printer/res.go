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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/resources"
)

type sourceKind int

const (
	sourceUnset sourceKind = iota
	sourceLink
	sourceHandler
	sourceStatic
	sourceUnknown
)

type Resource interface {
	SetSource(any) error
	SetMode(resources.ResourceMode) error
	SetPortable(bool)
	SetName(string)
	Validate() error
	Prepare(core.App) (*resources.Resource, error)
	Resolve(core.Core) (string, error)
}

var (
	_ Resource = (*Asset)(nil)
	_ Resource = (*Script)(nil)
	_ Resource = (*Style)(nil)
)

type res struct {
	source   any
	srcKind  sourceKind
	mode     resources.ResourceMode
	modeSet  bool
	portable bool
	name     string
}

func (r *res) setSource(v any) {
	switch value := v.(type) {
	case HandlerFunc:
		r.source = SourceHook(value)
		r.srcKind = sourceHandler
	case HandlerSimpleFunc:
		r.source = SourceHook(func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
			value(w, r)
			return false
		})
		r.srcKind = sourceHandler
	case string:
		r.source = value
		r.srcKind = sourceLink
	case SourceExternal:
		r.source = value
		r.srcKind = sourceLink
	case []byte:
		r.source = SourceBytes(value)
		r.srcKind = sourceStatic
	case SourceStatic:
		r.source = value
		r.srcKind = sourceStatic
	case SourceHandler:
		r.source = value
		r.srcKind = sourceHandler
	default:
		r.srcKind = sourceUnknown
	}
}

func errKind(kind string) error {
	return errors.New(kind + " resource used")
}

func (r *res) SetMode(mode resources.ResourceMode) error {
	if r.modeSet && r.mode != mode {
		return errors.New("only one of cache, private, or nocache may be set")
	}
	r.mode = mode
	r.modeSet = true
	return nil
}

func (r *res) SetPortable(portable bool) {
	r.portable = portable
}

func (r *res) SetName(name string) {
	r.name = name
}

func (r *res) hasMode(mode resources.ResourceMode) bool {
	return r.modeSet && r.mode == mode
}

func (r *res) defaultMode(mode resources.ResourceMode) {
	if r.modeSet {
		return
	}
	r.mode = mode
	r.modeSet = true
}

func (r *res) defaultHosting() {
	if r.portable {
		r.defaultMode(resources.ModeNoHost)
	}
	r.defaultMode(resources.ModeHost)
}

func (r *res) validate() error {
	switch true {
	case r.srcKind == sourceUnknown:
		return r.unsupported()
	case r.srcKind == sourceLink && (r.modeSet || r.portable):
		return errors.New("direct URL sources do not support cache, private, nocache, or portable")
	case r.srcKind == sourceHandler && (r.hasMode(resources.ModeHost) || r.hasMode(resources.ModeNoCache)):
		return errors.New("handler-backed sources do not support cache or nocache")
	case r.portable && r.hasMode(resources.ModeHost):
		return errors.New("portable cannot be combined with cache")
	}
	return nil
}

func (r *res) link() (string, bool) {
	switch src := r.source.(type) {
	case string:
		return src, true
	case SourceExternal:
		return string(src), true
	}
	return "", false
}

func (r *res) handler() (HandlerFunc, error) {
	src, ok := r.source.(SourceHandler)
	if !ok {
		return nil, r.unsupported()
	}
	return src.Handler(), nil
}

func (r *res) unsupported() error {
	return fmt.Errorf("unsupported resource source %T", r.source)
}

func (r *res) hook(c core.Core, handler HandlerFunc) (string, error) {
	hook, ok := c.Door().RegisterHook(handler, core.RaceParallel, false, r.portable)
	if !ok {
		return "", context.Canceled
	}
	if hook.Portable != "" {
		return c.App().PathMaker().Portable(hook.Portable, r.name), nil
	}
	return c.App().PathMaker().Hook(c.Instance().ID(), hook.HookID, r.name), nil
}

func (r *res) handlerURL(c core.Core) (string, error) {
	handler, err := r.handler()
	if err != nil {
		return "", err
	}
	return r.hook(c, handler)
}

func (r *res) resourceURL(c core.Core, res *resources.Resource) (string, error) {
	switch r.mode {
	case resources.ModeHost:
		return c.App().PathMaker().Resource(res, r.name), nil
	case resources.ModeNoHost, resources.ModeNoCache:
		return r.hook(c, func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			res.Serve(w, r)
			return false
		})
	default:
		panic("internal error: unexpected resource mode")
	}
}

type body struct {
	str   string
	bytes []byte
	buf   *bytes.Buffer
	set   bool
}

func (b *body) appendString(s string) {
	if b.spill() {
		b.buf.WriteString(s)
		return
	}
	b.str = s
	b.set = true
}

func (b *body) appendBytes(p []byte) {
	if b.spill() {
		b.buf.Write(p)
		return
	}
	b.bytes = p
	b.set = true
}

func (b *body) spill() bool {
	if !b.set {
		return false
	}
	if b.buf == nil {
		b.buf = &bytes.Buffer{}
		b.buf.WriteString(b.str)
		b.buf.Write(b.bytes)
	}
	return true
}

func (b *body) empty() bool {
	return !b.set
}

func (b *body) content() (str string, data []byte, spilled bool) {
	if b.buf != nil {
		return "", b.buf.Bytes(), true
	}
	return b.str, b.bytes, false
}

type Asset struct {
	res
	contentType string
}

func (a *Asset) SetSource(v any) error {
	switch v := v.(type) {
	case *Asset:
		*a = *v
	case *Script:
		return errKind("script")
	case *Style:
		return errKind("style")
	default:
		a.setSource(v)
	}
	return nil
}

func (a *Asset) SetContentType(contentType string) {
	a.contentType = contentType
}

func (a *Asset) Validate() error {
	if a.hasMode(resources.ModeNoHost) && a.srcKind == sourceStatic {
		a.mode = resources.ModeNoCache
	}
	if err := a.validate(); err != nil {
		return err
	}
	if a.srcKind == sourceHandler && a.contentType != "" {
		return errors.New("handler-backed sources do not support the type attribute")
	}
	return nil
}

func (a *Asset) Resolve(c core.Core) (string, error) {
	a.defaultMode(resources.ModeNoCache)
	if url, ok := a.link(); ok {
		return url, nil
	}
	if a.srcKind == sourceStatic && a.mode != resources.ModeNoCache {
		res, err := a.Prepare(c.App())
		if err != nil {
			return "", err
		}
		return c.App().PathMaker().Resource(res, a.name), nil
	}
	handler, err := a.handler()
	if err != nil {
		return "", err
	}
	contentType := a.contentType
	return a.hook(c, func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		return handler(ctx, w, r)
	})
}

func (a *Asset) Prepare(app core.App) (*resources.Resource, error) {
	static, ok := a.source.(SourceStatic)
	if !ok {
		return nil, a.unsupported()
	}
	return app.ResourceRegistry().Static(static.StaticEntry(), a.contentType)
}

type scriptOutput string

const (
	scriptRaw     scriptOutput = "raw"
	scriptInline  scriptOutput = "inline"
	scriptDefault scriptOutput = ""
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

type Script struct {
	res
	body
	output  scriptOutput
	module  bool
	ts      bool
	profile string
}

func (s *Script) SetSource(v any) error {
	switch v := v.(type) {
	case *Script:
		*s = *v
	case *Asset:
		if v.contentType != "" {
			return errors.New("content type is not supported")
		}
		s.res = v.res
	case *Style:
		return errKind("style")
	default:
		s.setSource(v)
	}
	return nil
}

func (s *Script) SetOutput(output scriptOutput) error {
	if s.output != scriptDefault && s.output != output {
		return errors.New("only one of raw, inline, or bundle may be set")
	}
	s.output = output
	return nil
}

func (s *Script) SetModule(module bool) {
	s.module = module
}

func (s *Script) SetTS(ts bool) {
	s.ts = ts
}

func (s *Script) SetProfile(profile string) {
	s.profile = profile
}

func (s *Script) Validate() error {
	switch s.srcKind {
	case sourceUnset:
		if s.output == scriptDefault {
			s.output = scriptInline
		}
	case sourceHandler, sourceLink:
		if s.output == scriptDefault {
			s.output = scriptRaw
		}
	}
	if s.name == "" {
		s.name = "script.js"
	}
	if err := s.validate(); err != nil {
		return err
	}
	switch true {
	case s.srcKind == sourceUnset && s.output == scriptBundle:
		return errors.New("inline scripts cannot be bundled; provide src")
	case s.srcKind == sourceUnset && s.ts:
		return errors.New("embedded scripts do not support TypeScript")
	case s.output == scriptInline && s.module:
		return errors.New("inline scripts do not support modules")
	case s.output == scriptRaw && s.ts:
		return errors.New("raw output does not support TypeScript sources")
	case s.srcKind == sourceLink && s.output != scriptRaw:
		return errors.New("direct URL scripts only support raw output")
	case s.srcKind == sourceHandler && s.output != scriptRaw:
		return errors.New("handler-backed scripts only support raw output")
	}
	return nil
}

func (s *Script) Resolve(c core.Core) (string, error) {
	s.defaultHosting()
	if url, ok := s.link(); ok {
		return url, nil
	}
	if s.srcKind == sourceHandler {
		return s.handlerURL(c)
	}
	if s.srcKind == sourceUnset {
		s.name = resourceFileName(s.name, "js")
	}
	res, err := s.Prepare(c.App())
	if err != nil {
		return "", err
	}
	return s.resourceURL(c, res)
}

func (s *Script) Prepare(app core.App) (*resources.Resource, error) {
	entry := s.entry()
	if entry == nil {
		return nil, s.unsupported()
	}
	format := resources.ScriptFormat(resources.FormatDefault{})
	profile := ""
	if s.srcKind != sourceUnset {
		var err error
		if format, err = s.output.format(s.module); err != nil {
			return nil, err
		}
		profile = s.profile
	}
	return app.ResourceRegistry().Script(entry, format, profile, s.mode)
}

func (s *Script) entry() resources.ScriptEntry {
	if s.srcKind != sourceUnset {
		static, ok := s.source.(SourceStatic)
		if !ok {
			return nil
		}
		return static.scriptEntry(s.output == scriptInline, s.ts)
	}
	if s.empty() {
		return nil
	}
	str, data, spilled := s.content()
	switch true {
	case spilled:
		return resources.ScriptInlineString{Content: string(data), Kind: resources.KindJS}
	case data != nil:
		return resources.ScriptInlineBytes{Content: data, Kind: resources.KindJS}
	default:
		return resources.ScriptInlineString{Content: str, Kind: resources.KindJS}
	}
}

type styleOutput string

const (
	styleRaw     styleOutput = "raw"
	styleDefault styleOutput = ""
)

func (m styleOutput) String() string {
	return string(m)
}

type Style struct {
	res
	body
	output styleOutput
}

func (s *Style) SetSource(v any) error {
	switch v := v.(type) {
	case *Style:
		*s = *v
	case *Asset:
		if v.contentType != "" {
			return errors.New("content type is not supported")
		}
		s.res = v.res
	case *Script:
		return errKind("script")
	default:
		s.setSource(v)
	}
	return nil
}

func (s *Style) SetRaw() {
	s.output = styleRaw
}

func (s *Style) Validate() error {
	if (s.srcKind == sourceLink || s.srcKind == sourceHandler) && s.output == styleDefault {
		s.output = styleRaw
	}
	if s.name == "" {
		s.name = "style.css"
	}
	return s.validate()
}

func (s *Style) Resolve(c core.Core) (string, error) {
	s.defaultHosting()
	if url, ok := s.link(); ok {
		return url, nil
	}
	if s.srcKind == sourceHandler {
		return s.handlerURL(c)
	}
	if s.srcKind == sourceUnset {
		s.name = resourceFileName(s.name, "css")
	}
	res, err := s.Prepare(c.App())
	if err != nil {
		return "", err
	}
	return s.resourceURL(c, res)
}

func (s *Style) Prepare(app core.App) (*resources.Resource, error) {
	entry := s.entry()
	if entry == nil {
		return nil, s.unsupported()
	}
	minify := s.output == styleDefault || s.srcKind == sourceUnset
	return app.ResourceRegistry().Style(entry, minify, s.mode)
}

func (s *Style) entry() resources.StyleEntry {
	if s.srcKind != sourceUnset {
		static, ok := s.source.(SourceStatic)
		if !ok {
			return nil
		}
		return static.styleEntry()
	}
	if s.empty() {
		return nil
	}
	str, data, spilled := s.content()
	switch true {
	case spilled, data != nil:
		return resources.StyleBytes{Content: data}
	default:
		return resources.StyleString{Content: str}
	}
}
