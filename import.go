package doors

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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

// Import represents a resource that can be imported into a web page.
// Imports can be JavaScript modules, CSS stylesheets, or external resources.
// They are processed during rendering to generate appropriate HTML tags and import maps.
type Import interface {
	print() string
	init(*importMap, *resources.Registry, *common.CSPCollector) error
}

// ImportModule imports a JavaScript/TypeScript file, processes it through esbuild,
// and makes it available as an ES module. The module can be added to the import map
// with a specifier name and/or loaded directly via script tag.
type ImportModule struct {
	Specifier string // Import map specifier name (optional)
	Path      string // File system path to the module
	Profile   string // Build profile for esbuild options
	Load      bool   // Whether to load the module immediately via script tag
	Name      string // Custom name for the generated file (optional)
}

func (m ImportModule) print() string {
	return "module " + m.Path
}

func (m ImportModule) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.print()))
		return nil
	}
	s, err := r.Module(m.Path, m.Profile)
	if err != nil {
		return err
	}
	path := importPath(s, m.Path, m.Name, "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

// ImportModuleBytes imports JavaScript/TypeScript content from a byte slice,
// processes it through esbuild, and makes it available as an ES module.
type ImportModuleBytes struct {
	Specifier string // Import map specifier name (optional)
	Content   []byte // Module source code
	Profile   string // Build profile for esbuild options
	Load      bool   // Whether to load the module immediately via script tag
	Name      string // Custom name for the generated file
}

func (m ImportModuleBytes) print() string {
	return "module bytes"
}

func (m ImportModuleBytes) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.print()))
		return nil
	}

	s, err := r.ModuleBytes(m.Content, m.Profile)
	if err != nil {
		return err
	}
	path := importPath(s, m.Name, "", "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

// ImportModuleRaw imports a JavaScript file without any processing or transformation.
// The file is served as-is without going through esbuild.
type ImportModuleRaw struct {
	Specifier string // Import map specifier name (optional)
	Path      string // File system path to the module
	Load      bool   // Whether to load the module immediately via script tag
	Name      string // Custom name for the generated file (optional)
}

func (m ImportModuleRaw) print() string {
	return "module raw " + m.Path
}

func (m ImportModuleRaw) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.print()))
		return nil
	}

	s, err := r.ModuleRaw(m.Path)
	if err != nil {
		return err
	}
	path := importPath(s, m.Path, m.Name, "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

// ImportModuleRawBytes imports JavaScript content from a byte slice without
// any processing or transformation. The content is served as-is.
type ImportModuleRawBytes struct {
	Specifier string // Import map specifier name (optional)
	Content   []byte // Raw JavaScript content
	Load      bool   // Whether to load the module immediately via script tag
	Name      string // Custom name for the generated file
}

func (m ImportModuleRawBytes) print() string {
	return "module raw bytes"
}

func (m ImportModuleRawBytes) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.print()))
		return nil
	}
	s, err := r.ModuleRawBytes(m.Content)
	if err != nil {
		return err
	}
	path := importPath(s, "", m.Name, "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

// ImportModuleBundle creates a bundled JavaScript module from an entry point,
// bundling all local imports and dependencies into a single file using esbuild.
type ImportModuleBundle struct {
	Specifier string // Import map specifier name (optional)
	Entry     string // Entry point file for the bundle
	Profile   string // Build profile for esbuild options
	Load      bool   // Whether to load the module immediately via script tag
	Name      string // Custom name for the generated file (optional)
}

func (m ImportModuleBundle) print() string {
	return "module bundle " + m.Entry
}

func (m ImportModuleBundle) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.print()))
		return nil
	}
	s, err := r.ModuleBundle(m.Entry, m.Profile)
	if err != nil {
		return err
	}
	path := importPath(s, "", m.Name, "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

// ImportModuleBundleFS creates a bundled JavaScript module from a file system entry point,
// bundling all local imports and dependencies into a single file using esbuild.
// This is useful for embedding assets or working with embed.FS.
type ImportModuleBundleFS struct {
	CacheKey  string // Unique cache key for this filesystem/bundle combination
	Specifier string // Import map specifier name (optional)
	FS        fs.FS  // File system to read from
	Entry     string // Entry point file within the filesystem
	Profile   string // Build profile for esbuild options
	Load      bool   // Whether to load the module immediately via script tag
	Name      string // Custom name for the generated file (optional)
}

func (m ImportModuleBundleFS) print() string {
	return "module bundle fs " + m.CacheKey
}

func (m ImportModuleBundleFS) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	if m.Specifier == "" && !m.Load {
		slog.Warn("imported resource skipped: load is false and no specifier provided", slog.String("import", m.print()))
		return nil
	}
	s, err := r.ModuleBundleFS(m.CacheKey, m.FS, m.Entry, m.Profile)
	if err != nil {
		return err
	}
	path := importPath(s, "", m.Name, "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

// ImportModuleHosted imports a JavaScript module that is hosted locally
// but not processed by the build system. The Src should be a full path
// starting from the application root.
type ImportModuleHosted struct {
	Specifier string // Import map specifier name (optional)
	Load      bool   // Whether to load the module immediately via script tag
	Src       string // Full path to the hosted module
}

func (m ImportModuleHosted) print() string {
	return "local module " + m.Src
}

func (m ImportModuleHosted) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	if m.Specifier != "" {
		im.addImport(m.Specifier, m.Src)
	}
	if m.Load {
		im.addRender(importScript(m.Src))
	}
	return nil
}

// ImportModuleExternal imports a JavaScript module from an external URL.
// This adds the URL to the Content Security Policy script sources.
type ImportModuleExternal struct {
	Specifier string // Import map specifier name (optional)
	Load      bool   // Whether to load the module immediately via script tag
	Src       string // External URL to the module
}

func (m ImportModuleExternal) print() string {
	return "external module " + m.Src
}

func (m ImportModuleExternal) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	c.ScriptSource(m.Src)
	if m.Specifier != "" {
		im.addImport(m.Specifier, m.Src)
	}
	if m.Load {
		im.addRender(importScript(m.Src))
	}
	return nil
}

// ImportStyleHosted imports a CSS stylesheet that is hosted locally
// but not processed by the build system. The Href should be a full path
// starting from the application root.
type ImportStyleHosted struct {
	Href string // Full path to the hosted stylesheet
}

func (m ImportStyleHosted) print() string {
	return "local style " + m.Href
}

func (m ImportStyleHosted) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	im.addRender(importStyle(m.Href))
	return nil
}

// ImportStyleExternal imports a CSS stylesheet from an external URL.
// This adds the URL to the Content Security Policy style sources.
type ImportStyleExternal struct {
	Href string // External URL to the stylesheet
}

func (m ImportStyleExternal) print() string {
	return "external style " + m.Href
}

func (m ImportStyleExternal) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	c.StyleSource(m.Href)
	im.addRender(importStyle(m.Href))
	return nil
}

// ImportStyle imports a CSS file, processes it (minification), and makes it
// available as a stylesheet link in the HTML head.
type ImportStyle struct {
	Path string // File system path to the CSS file
	Name string // Custom name for the generated file (optional)
}

func (m ImportStyle) print() string {
	return "style " + m.Path
}

func (m ImportStyle) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	s, err := r.Style(m.Path)
	if err != nil {
		return err
	}
	path := importPath(s, m.Path, m.Name, "css")
	im.addRender(importStyle(path))
	return nil
}

// ImportStyleBytes imports CSS content from a byte slice, processes it
// (minification), and makes it available as a stylesheet link.
type ImportStyleBytes struct {
	Content []byte // CSS source code
	Name    string // Custom name for the generated file
}

func (m ImportStyleBytes) print() string {
	return "style bytes"
}

func (m ImportStyleBytes) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	s, err := r.StyleBytes(m.Content)
	if err != nil {
		return err
	}
	path := importPath(s, m.Name, "", "css")
	im.addRender(importStyle(path))
	return nil
}

type importMap struct {
	Imports map[string]string `json:"imports"`
	renders []templ.Component
}

func (i *importMap) render(c *common.CSPCollector, ctx context.Context, w io.Writer) error {
	json, err := json.Marshal(i)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(json)
	c.ScriptHash(hash[:])
	_, err = w.Write(common.AsBytes("<script type=\"importmap\">"))
	if err != nil {
		return err
	}
	_, err = w.Write(json)
	if err != nil {
		return err
	}
	_, err = w.Write(common.AsBytes("</script>"))
	if err != nil {
		return err
	}
	for _, component := range i.renders {
		err = component.Render(ctx, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *importMap) addImport(name string, path string) {
	i.Imports[name] = path
}

func (i *importMap) addRender(component templ.Component) {
	i.renders = append(i.renders, component)
}

// Imports generates the HTML import map and resource loading tags for the specified imports.
// This component should be placed in the HTML head section and can only be used once per page.
// It processes all import entries, builds an ES modules import map, and generates the necessary
// script and link tags for loading resources.
//
// The function handles:
//   - ES module import map generation
//   - JavaScript and CSS resource processing
//   - Content Security Policy header updates
//   - Automatic resource caching and optimization
//
// Import entries are processed in order, and any errors are logged while continuing
// with the remaining entries.
func Imports(entries ...Import) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		instance := ctx.Value(common.InstanceCtxKey).(instance.Core)
		registy := instance.ImportRegistry()
		collector, ok := instance.CSPCollector()
		if !ok {
			slog.Error("@Imports is allowed only in initial render and only once.")
			return nil
		}
		im := &importMap{
			Imports: make(map[string]string),
			renders: make([]templ.Component, 0),
		}
		for _, entry := range entries {
			err := entry.init(im, registy, collector)
			if err != nil {
				slog.Error(err.Error(), slog.String("import", entry.print()))
				continue
			}
		}
		return im.render(collector, ctx, w)
	})
}
