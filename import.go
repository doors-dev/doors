// doors
// Copyright (c) 2025 doors dev LLC
//
// Licensed under the Business Source License 1.1 (BUSL-1.1).
// See LICENSE.txt for details.
//
// For commercial use, see LICENSE-COMMERCIAL.txt and COMMERCIAL-EULA.md.
// To purchase a license, visit https://doors.dev or contact sales@doors.dev.

package doors

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/doors-dev/doors/internal/common"
	"github.com/doors-dev/doors/internal/instance"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/router"
)

func importPath(r *resources.Resource, path string, name string, ext string) string {
	fileName := ext
	if name != "" {
		fileName = name + "." + fileName
	} else if path != "" {
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		fileName = name + "." + fileName
	}
	return router.ResourcePath(r, fileName)
}

// ImportModule imports a JS/TS file, processes it with esbuild, and exposes it as an ES module.
type ImportModule struct {
	// Import map specifier name.
	// Optional.
	Specifier string
	// File system path to the module.
	// Required.
	Path string
	// Build profile for esbuild options.
	// Optional.
	Profile string
	// Load the module immediately via a script tag.
	// Optional.
	Load bool
	// Custom name for the generated file.
	// Optional.
	Name string
	// Additional HTML attributes for the script tag.
	// Optional.
	Attrs templ.Attributes
}

func (m ImportModule) info() string {
	return "module " + m.Path
}
func (m ImportModule) Render(ctx context.Context, w io.Writer) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.info()))
		return nil
	}
	man := newImportManager(ctx)
	s, err := man.registry.Module(m.Path, m.Profile)
	if err != nil {
		return err
	}
	path := importPath(s, m.Path, m.Name, "js")
	if m.Specifier != "" {
		man.rm.AddImport(m.Specifier, path)
	}
	if m.Load {
		return importScript(path, m.Attrs).Render(ctx, w)
	}
	return nil
}

// ImportModuleBytes imports JS/TS content from bytes, processes it with esbuild, and exposes it as an ES module.
type ImportModuleBytes struct {
	// Import map specifier name.
	// Optional.
	Specifier string
	// Module source code.
	// Required.
	Content []byte
	// Build profile for esbuild options.
	// Optional.
	Profile string
	// Load the module immediately via a script tag.
	// Optional.
	Load bool
	// Custom name for the generated file.
	// Required.
	Name string
	// Additional HTML attributes for the script tag.
	// Optional.
	Attrs templ.Attributes
}


func (m ImportModuleBytes) info() string {
	return "module bytes"
}

func (m ImportModuleBytes) Render(ctx context.Context, w io.Writer) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.info()))
		return nil
	}
	man := newImportManager(ctx)
	s, err := man.registry.ModuleBytes(m.Content, m.Profile)
	if err != nil {
		return err
	}
	path := importPath(s, m.Name, "", "js")
	if m.Specifier != "" {
		man.rm.AddImport(m.Specifier, path)
	}
	if m.Load {
		return importScript(path, m.Attrs).Render(ctx, w)
	}
	return nil
}

// ImportModuleRaw imports a JS file as-is without processing.
type ImportModuleRaw struct {
	// Import map specifier name.
	// Optional.
	Specifier string
	// File system path to the module.
	// Required.
	Path string
	// Load the module immediately via a script tag.
	// Optional.
	Load bool
	// Custom name for the generated file.
	// Optional.
	Name string
	// Additional HTML attributes for the script tag.
	// Optional.
	Attrs templ.Attributes
}

func (m ImportModuleRaw) info() string {
	return "module raw " + m.Path
}

func (m ImportModuleRaw) Render(ctx context.Context, w io.Writer) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.info()))
		return nil
	}
	man := newImportManager(ctx)
	s, err := man.registry.ModuleRaw(m.Path)
	if err != nil {
		return err
	}
	path := importPath(s, m.Path, m.Name, "js")
	if m.Specifier != "" {
		man.rm.AddImport(m.Specifier, path)
	}
	if m.Load {
		return importScript(path, m.Attrs).Render(ctx, w)
	}
	return nil
}

// ImportModuleRawBytes serves raw JS content from bytes without processing.
type ImportModuleRawBytes struct {
	// Import map specifier name.
	// Optional.
	Specifier string
	// Raw JavaScript content.
	// Required.
	Content []byte
	// Load the module immediately via a script tag.
	// Optional.
	Load bool
	// Custom name for the generated file.
	// Required.
	Name string
	// Additional HTML attributes for the script tag.
	// Optional.
	Attrs templ.Attributes
}

func (m ImportModuleRawBytes) info() string {
	return "module raw bytes"
}

func (m ImportModuleRawBytes) Render(ctx context.Context, w io.Writer) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.info()))
		return nil
	}
	man := newImportManager(ctx)
	s, err := man.registry.ModuleRawBytes(m.Content)
	if err != nil {
		return err
	}
	path := importPath(s, "", m.Name, "js")
	if m.Specifier != "" {
		man.rm.AddImport(m.Specifier, path)
	}
	if m.Load {
		return importScript(path, m.Attrs).Render(ctx, w)
	}
	return nil
}


// ImportModuleBundle bundles a JS entry with its deps into one file using esbuild.
type ImportModuleBundle struct {
	// Import map specifier name.
	// Optional.
	Specifier string
	// Entry point file for the bundle.
	// Required.
	Entry string
	// Build profile for esbuild options.
	// Optional.
	Profile string
	// Load the module immediately via a script tag.
	// Optional.
	Load bool
	// Custom name for the generated file.
	// Optional.
	Name string
	// Additional HTML attributes for the script tag.
	// Optional.
	Attrs templ.Attributes
}

func (m ImportModuleBundle) info() string {
	return "module bundle " + m.Entry
}

func (m ImportModuleBundle) Render(ctx context.Context, w io.Writer) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.info()))
		return nil
	}
	man := newImportManager(ctx)
	s, err := man.registry.ModuleBundle(m.Entry, m.Profile)
	if err != nil {
		return err
	}
	path := importPath(s, "", m.Name, "js")
	if m.Specifier != "" {
		man.rm.AddImport(m.Specifier, path)
	}
	if m.Load {
		return importScript(path, m.Attrs).Render(ctx, w)
	}
	return nil
}


// ImportModuleBundleFS bundles a JS entry from an fs.FS using esbuild.
type ImportModuleBundleFS struct {
	// Unique cache key for this filesystem/bundle combination.
	// Required.
	CacheKey string
	// Import map specifier name.
	// Optional.
	Specifier string
	// File system to read from.
	// Required.
	FS fs.FS
	// Entry point file within the filesystem.
	// Required.
	Entry string
	// Build profile for esbuild options.
	// Optional.
	Profile string
	// Load the module immediately via a script tag.
	// Optional.
	Load bool
	// Custom name for the generated file.
	// Optional.
	Name string
	// Additional HTML attributes for the script tag.
	// Optional.
	Attrs templ.Attributes
}

func (m ImportModuleBundleFS) info() string {
	return "module bundle fs " + m.CacheKey
}

func (m ImportModuleBundleFS) Render(ctx context.Context, w io.Writer) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.info()))
		return nil
	}
	man := newImportManager(ctx)
	s, err := man.registry.ModuleBundleFS(m.CacheKey, m.FS, m.Entry, m.Profile)
	if err != nil {
		return err
	}
	path := importPath(s, "", m.Name, "js")
	if m.Specifier != "" {
		man.rm.AddImport(m.Specifier, path)
	}
	if m.Load {
		return importScript(path, m.Attrs).Render(ctx, w)
	}
	return nil
}


// ImportModuleHosted registers a locally hosted JS module without processing.
type ImportModuleHosted struct {
	// Import map specifier name.
	// Optional.
	Specifier string
	// Load the module immediately via a script tag.
	// Optional.
	Load bool
	// Full path to the hosted module (application root).
	// Required.
	Src string
	// Additional HTML attributes for the script tag.
	// Optional.
	Attrs templ.Attributes
}

func (m ImportModuleHosted) info() string {
	return "local module " + m.Src
}

func (m ImportModuleHosted) Render(ctx context.Context, w io.Writer) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.info()))
		return nil
	}
	man := newImportManager(ctx)
	if m.Specifier != "" {
		man.rm.AddImport(m.Specifier, m.Src)
	}
	if m.Load {
		return importScript(m.Src, m.Attrs).Render(ctx, w)
	}
	return nil
}


// ImportModuleExternal registers an external JS module URL and adds it to CSP.
type ImportModuleExternal struct {
	// Import map specifier name.
	// Optional.
	Specifier string
	// Load the module immediately via a script tag.
	// Optional.
	Load bool
	// External URL to the module.
	// Required.
	Src string
	// Additional HTML attributes for the script tag.
	// Optional.
	Attrs templ.Attributes
}


func (m ImportModuleExternal) info() string {
	return "external module " + m.Src
}

func (m ImportModuleExternal) Render(ctx context.Context, w io.Writer) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.info()))
		return nil
	}
	man := newImportManager(ctx)
	man.collector.ScriptSource(m.Src)
	if m.Specifier != "" {
		man.rm.AddImport(m.Specifier, m.Src)
	}
	if m.Load {
		return importScript(m.Src, m.Attrs).Render(ctx, w)
	}
	return nil
}

// ImportStyleHosted links a locally hosted CSS file without processing.
type ImportStyleHosted struct {
	// Path to the hosted stylesheet.
	// Required.
	Href string
	// Additional HTML attributes for the link tag.
	// Optional.
	Attrs templ.Attributes
}

func (m ImportStyleHosted) info() string {
	return "local style " + m.Href
}

func (m ImportStyleHosted) Render(ctx context.Context, w io.Writer) error {
	return importStyle(m.Href, m.Attrs).Render(ctx, w)
}

// ImportStyleExternal links a CSS stylesheet from an external URL and adds it to CSP.
type ImportStyleExternal struct {
	// External URL to the stylesheet.
	// Required.
	Href string
	// Additional HTML attributes for the link tag.
	// Optional.
	Attrs templ.Attributes
}

func (m ImportStyleExternal) info() string {
	return "external style " + m.Href
}

func (m ImportStyleExternal) Render(ctx context.Context, w io.Writer) error {
	man := newImportManager(ctx)
	man.collector.StyleSource(m.Href)
	return importStyle(m.Href, m.Attrs).Render(ctx, w)
}


// ImportStyle processes a CSS file (e.g., minify) and links it.
type ImportStyle struct {
	// File system path to the CSS file.
	// Required.
	Path string
	// Custom name for the generated file.
	// Optional.
	Name string
	// Additional HTML attributes for the link tag.
	// Optional.
	Attrs templ.Attributes
}


func (m ImportStyle) info() string {
	return "style " + m.Path
}

func (m ImportStyle) Render(ctx context.Context, w io.Writer) error {
	man := newImportManager(ctx)
	s, err := man.registry.Style(m.Path)
	if err != nil {
		return err
	}
	path := importPath(s, m.Path, m.Name, "css")
	return importStyle(path, m.Attrs).Render(ctx, w)
}


// ImportStyleBytes processes CSS content from bytes (e.g., minify) and links it.
type ImportStyleBytes struct {
	// CSS source code.
	// Required.
	Content []byte
	// Custom name for the generated file.
	// Required.
	Name string
	// Additional HTML attributes for the link tag.
	// Optional.
	Attrs templ.Attributes
}


func (m ImportStyleBytes) info() string {
	return "style bytes"
}

func (m ImportStyleBytes) Render(ctx context.Context, w io.Writer) error {
	man := newImportManager(ctx)
	s, err := man.registry.StyleBytes(m.Content)
	if err != nil {
		return err
	}
	path := importPath(s, m.Name, "", "css")
	return importStyle(path, m.Attrs).Render(ctx, w)
}

type importManager struct {
	registry  *resources.Registry
	collector *common.CSPCollector
	rm        *common.RenderMap
}

func newImportManager(ctx context.Context) *importManager {
	instance := ctx.Value(common.CtxKeyInstance).(instance.Core)
	registy := instance.ImportRegistry()
	collector := instance.CSPCollector()
	rm := ctx.Value(common.CtxKeyRenderMap).(*common.RenderMap)
	return &importManager{
		registry:  registy,
		collector: collector,
		rm:        rm,
	}
}

func importScript(src string, attrs templ.Attributes) templ.Component {
	if attrs == nil {
		attrs = make(templ.Attributes, 2)
	}
    attrs["type"] = "module"
    attrs["src"] = src
    return renderRaw("script", attrs, nil, true)
}

func importStyle(href string, attrs templ.Attributes) templ.Component {
	if attrs == nil {
		attrs = make(templ.Attributes, 2)
	}
    attrs["rel"] = "stylesheet"
    attrs["href"] = href
    return renderRaw("link", attrs, nil, false)
}
