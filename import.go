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

// ImportModule imports a JavaScript/TypeScript file, processes it through esbuild,
// and makes it available as an ES module. The module can be added to the import map
// with a specifier name and/or loaded directly via script tag.
type ImportModule struct {
	Specifier string           // Import map specifier name (optional)
	Path      string           // File system path to the module
	Profile   string           // Build profile for esbuild options
	Load      bool             // Whether to load the module immediately via script tag
	Name      string           // Custom name for the generated file (optional)
	Attrs     templ.Attributes // Additional html attributes
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

// ImportModuleBytes imports JavaScript/TypeScript content from a byte slice,
// processes it through esbuild, and makes it available as an ES module.
type ImportModuleBytes struct {
	Specifier string           // Import map specifier name (optional)
	Content   []byte           // Module source code
	Profile   string           // Build profile for esbuild options
	Load      bool             // Whether to load the module immediately via script tag
	Name      string           // Custom name for the generated file
	Attrs     templ.Attributes // Additional html attributes
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

// ImportModuleRaw imports a JavaScript file without any processing or transformation.
// The file is served as-is without going through esbuild.
type ImportModuleRaw struct {
	Specifier string           // Import map specifier name (optional)
	Path      string           // File system path to the module
	Load      bool             // Whether to load the module immediately via script tag
	Name      string           // Custom name for the generated file (optional)
	Attrs     templ.Attributes // Additional html attributes
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

// ImportModuleRawBytes imports JavaScript content from a byte slice without
// any processing or transformation. The content is served as-is.
type ImportModuleRawBytes struct {
	Specifier string           // Import map specifier name (optional)
	Content   []byte           // Raw JavaScript content
	Load      bool             // Whether to load the module immediately via script tag
	Name      string           // Custom name for the generated file
	Attrs     templ.Attributes // Additional html attributes
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

// ImportModuleBundle creates a bundled JavaScript module from an entry point,
// bundling all local imports and dependencies into a single file using esbuild.
type ImportModuleBundle struct {
	Specifier string           // Import map specifier name (optional)
	Entry     string           // Entry point file for the bundle
	Profile   string           // Build profile for esbuild options
	Load      bool             // Whether to load the module immediately via script tag
	Name      string           // Custom name for the generated file (optional)
	Attrs     templ.Attributes // Additional html attributes
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

// ImportModuleBundleFS creates a bundled JavaScript module from a file system entry point,
// bundling all local imports and dependencies into a single file using esbuild.
// This is useful for embedding assets or working with embed.FS.
type ImportModuleBundleFS struct {
	CacheKey  string           // Unique cache key for this filesystem/bundle combination
	Specifier string           // Import map specifier name (optional)
	FS        fs.FS            // File system to read from
	Entry     string           // Entry point file within the filesystem
	Profile   string           // Build profile for esbuild options
	Load      bool             // Whether to load the module immediately via script tag
	Name      string           // Custom name for the generated file (optional)
	Attrs     templ.Attributes // Additional html attributes
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

// ImportModuleHosted imports a JavaScript module that is hosted locally
// but not processed by the build system. The Src should be a full path
// starting from the application root.
type ImportModuleHosted struct {
	Specifier string           // Import map specifier name (optional)
	Load      bool             // Whether to load the module immediately via script tag
	Src       string           // Full path to the hosted module
	Attrs     templ.Attributes // Additional html attributes
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

// ImportModuleExternal imports a JavaScript module from an external URL.
// This adds the URL to the Content Security Policy script sources.
type ImportModuleExternal struct {
	Specifier string           // Import map specifier name (optional)
	Load      bool             // Whether to load the module immediately via script tag
	Src       string           // External URL to the module
	Attrs     templ.Attributes // Additional html attributes
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

// ImportStyleHosted imports a CSS stylesheet that is hosted locally
// but not processed by the build system. The Href should be a full path
// starting from the application root.
type ImportStyleHosted struct {
	Href  string           // Full path to the hosted stylesheet
	Attrs templ.Attributes // Additional html attributes
}

func (m ImportStyleHosted) info() string {
	return "local style " + m.Href
}

func (m ImportStyleHosted) Render(ctx context.Context, w io.Writer) error {
	return importStyle(m.Href, m.Attrs).Render(ctx, w)
}

// ImportStyleExternal imports a CSS stylesheet from an external URL.
// This adds the URL to the Content Security Policy style sources.
type ImportStyleExternal struct {
	Href  string           // External URL to the stylesheet
	Attrs templ.Attributes // Additional html attributes
}

func (m ImportStyleExternal) info() string {
	return "external style " + m.Href
}

func (m ImportStyleExternal) Render(ctx context.Context, w io.Writer) error {
	man := newImportManager(ctx)
	man.collector.StyleSource(m.Href)
	return importStyle(m.Href, m.Attrs).Render(ctx, w)
}

// ImportStyle imports a CSS file, processes it (minification), and makes it
// available as a stylesheet link in the HTML head.
type ImportStyle struct {
	Path  string           // File system path to the CSS file
	Name  string           // Custom name for the generated file (optional)
	Attrs templ.Attributes // Additional html attributes
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

// ImportStyleBytes imports CSS content from a byte slice, processes it
// (minification), and makes it available as a stylesheet link.
type ImportStyleBytes struct {
	Content []byte           // CSS source code
	Name    string           // Custom name for the generated file
	Attrs   templ.Attributes // Additional html attributes
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
	instance := ctx.Value(common.InstanceCtxKey).(instance.Core)
	registy := instance.ImportRegistry()
	collector := instance.CSPCollector()
	rm := ctx.Value(common.RenderMapCtxKey).(*common.RenderMap)
	return &importManager{
		registry:  registy,
		collector: collector,
		rm:        rm,
	}
}

// Deprecated: Direct import render supported
func Imports(content ...templ.Component) templ.Component {
	return Components(content...)
}
