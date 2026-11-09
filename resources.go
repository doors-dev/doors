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

package doors

import (
	"context"
	"errors"
	"io/fs"
	"net/http"

	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/printer"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

// Resource is app-owned content that Doors serves under a generated src or
// href URL.
//
// Assign it to a src or href attribute, or attach it as a modifier to a tag
// that takes one of them. In code, URL gives the URL that serves it.
type Resource interface {
	printer.SourceHandler
	ServedResource
	// Portable serves it through a hook URL that works without the session
	// cookie, like the "portable" attribute.
	Portable() HookResource
}

// ResourceStatic is a [Resource] whose content Doors reads up front.
//
// In markup, the tag decides how it is built: <script src> builds a script,
// a stylesheet link minifies a stylesheet, other tags serve it as is. In code,
// it and its Name, ContentType, Cache and Portable stages are an asset; use
// Script or Style for the URL a script or stylesheet tag produces.
type ResourceStatic interface {
	printer.SourceStatic
	AssetResource
	Script() ScriptResource
	Style() StyleResource
}

// ResourceDynamic is a [Resource] served by a handler on every request.
type ResourceDynamic interface {
	Resource
	// Name sets the file name at the end of the URL, like the "name" attribute.
	Name(name string) HookResource
}

// ResourceExternal is a URL the browser loads directly from another host.
//
// Doors adds its exact URL to the generated Content-Security-Policy for script
// tags and stylesheet or modulepreload links rendered with the page.
type ResourceExternal = printer.SourceExternal

// ResourceFS returns a [ResourceStatic] that serves entry from fsys.
func ResourceFS(fsys fs.FS, entry string) ResourceStatic {
	return staticResource{printer.SourceFS{FS: fsys, Entry: entry}}
}

// ResourceLocalFS returns a [ResourceStatic] that serves the local file at
// path.
func ResourceLocalFS(path string) ResourceStatic {
	return staticResource{printer.SourceLocalFS(path)}
}

// ResourceBytes returns a [ResourceStatic] that serves content from memory.
func ResourceBytes(content []byte) ResourceStatic {
	return staticResource{printer.SourceBytes(content)}
}

// ResourceString returns a [ResourceStatic] that serves content from memory.
func ResourceString(content string) ResourceStatic {
	return staticResource{printer.SourceString(content)}
}

// ResourceHook returns a [ResourceDynamic] served by handler.
//
// Return true from handler to stop serving the resource after that request,
// false to keep it available.
//
// WARNING: the request body is not limited by the ServerRequestBodyLimit of
// [Conf].
func ResourceHook(handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool) ResourceDynamic {
	return handlerResource{printer.SourceHook(handler)}
}

// ResourceHandler returns a [ResourceDynamic] served by handler.
//
// Unlike [ResourceHook], the resource keeps serving every request.
//
// WARNING: the request body is not limited by the ServerRequestBodyLimit of
// [Conf].
func ResourceHandler(handler func(w http.ResponseWriter, r *http.Request)) ResourceDynamic {
	return handlerResource{printer.SourceHook(func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
		handler(w, r)
		return false
	})}
}

// ResourceProxy returns a [ResourceDynamic] that reverse-proxies requests to
// url.
//
// Unlike [ResourceExternal], the browser loads it from a Doors-generated URL.
// If url cannot be parsed, requests fail with status 500.
func ResourceProxy(url string) ResourceDynamic {
	return handlerResource{printer.SourceProxy(url)}
}

// ServedResource is any resource Doors can serve. Every stage implements it:
// use it in a src or href attribute, as a modifier, or get its URL. A chain
// ends as a [HookResource] or a [CacheableResource].
type ServedResource interface {
	gox.Modify
	// URL returns the URL path the matching tag attributes produce.
	//
	// It returns an error if the source cannot be read or built, and
	// context.Canceled if it is served through a hook and the content around
	// ctx has been cleared.
	//
	// ctx must belong to a Doors render or handler; otherwise URL panics.
	URL(ctx context.Context) (string, error)
	spec() resource
}

// CacheableResource is a resource that is cached in its current state.
// [UseResource] serves it from RAM at a fixed path.
type CacheableResource interface {
	ServedResource
	cacheable()
}

// HookResource is served through a hook on the closest dynamic parent. Its
// URL stops working once that content is cleared.
//
// The handler of a [Resource] sees the part after the hook URL as r.URL.Path,
// including the name, and the Doors-Hook-Prefix request header holds the hook
// URL without the name. Requests are handled concurrently.
//
// WARNING: the request body is not limited by the ServerRequestBodyLimit of
// [Conf].
type HookResource interface {
	ServedResource
	// Name sets the file name at the end of the URL, like the "name" attribute.
	Name(name string) HookResource
	// Portable makes the URL work without the session cookie, like the
	// "portable" attribute.
	Portable() HookResource
}

// RawResource is a script or style served without building. It is cached
// unless NoCache, Private or Portable is chosen.
type RawResource interface {
	CacheableResource
	// Name sets the file name at the end of the URL, like the "name" attribute.
	Name(name string) RawResource
	// NoCache serves it through a hook without keeping it in the shared
	// cache, like the "nocache" attribute. It is read on every render.
	NoCache() HookResource
	// Private keeps it in the shared cache but serves it through a hook
	// instead of a public URL, like the "private" attribute.
	Private() HookResource
	// Portable serves it through a hook URL that works without the session
	// cookie, like the "portable" attribute.
	Portable() HookResource
}

// ScriptResource is a [ResourceStatic] built like <script src>. It is cached
// unless NoCache, Private or Portable is chosen.
type ScriptResource interface {
	CacheableResource
	// Name sets the file name at the end of the URL, like the "name" attribute.
	Name(name string) ScriptResource
	// Raw serves the file without building it, like the "raw" attribute. It
	// drops Bundle, TypeScript and Profile set before it; Module is kept.
	Raw() RawResource
	// Bundle bundles imports, like the "bundle" attribute.
	Bundle() ScriptResource
	// Module builds an ES module, like type="module".
	Module() ScriptResource
	// TypeScript compiles in-memory content as TypeScript, like
	// type="typescript". Files go by their extension.
	TypeScript() ScriptResource
	// Profile selects the esbuild profile, like the "profile" attribute.
	Profile(profile string) ScriptResource
	// NoCache builds it on every render and serves it through a hook, like
	// the "nocache" attribute.
	NoCache() HookResource
	// Private keeps the build in the shared cache but serves it through a
	// hook instead of a public URL, like the "private" attribute.
	Private() HookResource
	// Portable serves it through a hook URL that works without the session
	// cookie, like the "portable" attribute.
	Portable() HookResource
}

// StyleResource is a [ResourceStatic] minified like a stylesheet link. It is
// cached unless NoCache, Private or Portable is chosen.
type StyleResource interface {
	CacheableResource
	// Name sets the file name at the end of the URL, like the "name" attribute.
	Name(name string) StyleResource
	// Raw serves the stylesheet without minifying it, like the "raw"
	// attribute.
	Raw() RawResource
	// NoCache minifies it on every render and serves it through a hook, like
	// the "nocache" attribute.
	NoCache() HookResource
	// Private keeps the result in the shared cache but serves it through a
	// hook instead of a public URL, like the "private" attribute.
	Private() HookResource
	// Portable serves it through a hook URL that works without the session
	// cookie, like the "portable" attribute.
	Portable() HookResource
}

// AssetResource is content served as is, like a generic src or href. It is
// the default state of a [ResourceStatic]. Without Cache, the URL is a hook
// that reads the content on every request.
type AssetResource interface {
	ServedResource
	// Name sets the file name at the end of the URL, like the "name" attribute.
	Name(name string) AssetResource
	// ContentType sets the Content-Type, like the "type" attribute.
	ContentType(contentType string) AssetResource
	// Cache keeps the content in RAM and serves it from a shared URL derived
	// from its hash, like the "cache" attribute.
	Cache() CacheableResource
	// Portable serves it through a hook URL that works without the session
	// cookie, like the "portable" attribute.
	Portable() HookResource
}

type staticResource struct {
	printer.SourceStatic
}

func (s staticResource) asset() resourceStage[AssetResource] {
	return resourceStage[AssetResource]{resource{src: s.SourceStatic, kind: kindAsset}}
}

func (s staticResource) Name(name string) AssetResource {
	return s.asset().Name(name)
}

func (s staticResource) ContentType(contentType string) AssetResource {
	return s.asset().ContentType(contentType)
}

func (s staticResource) Cache() CacheableResource {
	return s.asset().Cache()
}

func (s staticResource) Portable() HookResource {
	return s.asset().Portable()
}

func (s staticResource) URL(ctx context.Context) (string, error) {
	return s.asset().URL(ctx)
}

func (s staticResource) spec() resource {
	return s.asset().resource
}

func (s staticResource) Script() ScriptResource {
	return resourceStage[ScriptResource]{resource{src: s.SourceStatic, kind: kindScript}}
}

func (s staticResource) Style() StyleResource {
	return resourceStage[StyleResource]{resource{src: s.SourceStatic, kind: kindStyle}}
}

type handlerResource struct {
	printer.SourceHandler
}

func (h handlerResource) hook() resourceStage[HookResource] {
	return resourceStage[HookResource]{resource{src: h.SourceHandler, kind: kindAsset}}
}

func (h handlerResource) Name(name string) HookResource {
	return h.hook().Name(name)
}

func (h handlerResource) Portable() HookResource {
	return h.hook().Portable()
}

func (h handlerResource) URL(ctx context.Context) (string, error) {
	return h.hook().URL(ctx)
}

func (h handlerResource) spec() resource {
	return h.hook().resource
}

type resourceKind int

const (
	kindAsset resourceKind = iota
	kindScript
	kindStyle
)

type resource struct {
	src         any
	kind        resourceKind
	mode        resources.ResourceMode
	modeSet     bool
	portable    bool
	name        string
	raw         bool
	bundle      bool
	module      bool
	ts          bool
	profile     string
	contentType string
}

func (r resource) spec() resource {
	return r
}

func (r resource) cacheable() {}

func (r resource) URL(ctx context.Context) (string, error) {
	c := ctx.Value(common.KeyCore).(core.Core)
	p := r.printer()
	if err := p.Validate(); err != nil {
		panic(errors.New("doors: invalid resource settings: " + err.Error()))
	}
	return p.Resolve(c)
}

func (r resource) Mutate(string, any) any {
	return r.printer()
}

func (r resource) Modify(_ context.Context, tag string, attrs gox.Attrs) error {
	return printer.ModifySource(tag, attrs, r)
}

type resourceStage[T any] struct {
	resource
}

func (s resourceStage[T]) Name(name string) T {
	s.name = name
	return any(s).(T)
}

func (s resourceStage[T]) Raw() RawResource {
	s.raw = true
	s.bundle = false
	s.ts = false
	s.profile = ""
	return resourceStage[RawResource](s)
}

func (s resourceStage[T]) Bundle() ScriptResource {
	s.bundle = true
	return resourceStage[ScriptResource](s)
}

func (s resourceStage[T]) Module() ScriptResource {
	s.module = true
	return resourceStage[ScriptResource](s)
}

func (s resourceStage[T]) TypeScript() ScriptResource {
	s.ts = true
	return resourceStage[ScriptResource](s)
}

func (s resourceStage[T]) Profile(profile string) ScriptResource {
	s.profile = profile
	return resourceStage[ScriptResource](s)
}

func (s resourceStage[T]) ContentType(contentType string) AssetResource {
	s.contentType = contentType
	return resourceStage[AssetResource](s)
}

func (s resourceStage[T]) Cache() CacheableResource {
	return resourceStage[CacheableResource]{s.withMode(resources.ModeHost)}
}

func (s resourceStage[T]) NoCache() HookResource {
	return resourceStage[HookResource]{s.withMode(resources.ModeNoCache)}
}

func (s resourceStage[T]) Private() HookResource {
	return resourceStage[HookResource]{s.withMode(resources.ModeNoHost)}
}

func (s resourceStage[T]) Portable() HookResource {
	s.portable = true
	return resourceStage[HookResource](s)
}

func (r resource) withMode(mode resources.ResourceMode) resource {
	r.mode = mode
	r.modeSet = true
	return r
}

func (r resource) printer() printer.Resource {
	var p printer.Resource
	switch r.kind {
	case kindScript:
		s := &printer.Script{}
		if r.raw {
			mustSet(s.SetOutput("raw"))
		}
		if r.bundle {
			mustSet(s.SetOutput("bundle"))
		}
		s.SetModule(r.module)
		s.SetTS(r.ts)
		s.SetProfile(r.profile)
		p = s
	case kindStyle:
		s := &printer.Style{}
		if r.raw {
			s.SetRaw()
		}
		p = s
	default:
		a := &printer.Asset{}
		a.SetContentType(r.contentType)
		p = a
	}
	mustSet(p.SetSource(r.src))
	if r.modeSet {
		mustSet(p.SetMode(r.mode))
	}
	p.SetPortable(r.portable)
	p.SetName(r.name)
	return p
}

func (r resource) static() bool {
	_, ok := r.src.(printer.SourceStatic)
	return ok
}

func (r resource) cached() printer.Resource {
	p := r.printer()
	if r.kind != kindAsset {
		mustSet(p.SetMode(resources.ModeNoHost))
	}
	return p
}

func mustSet(err error) {
	if err != nil {
		panic(errors.New("doors: invalid resource settings: " + err.Error()))
	}
}
