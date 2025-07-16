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

type Import interface {
	print() string
	init(*importMap, *resources.Registry, *common.CSPCollector) error
}

type ImportModule struct {
	Specifier string
	Path      string
	Profile   string
	Load      bool
	Name      string
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

type ImportModuleBytes struct {
	Specifier string
	Content   []byte
	Profile   string
	Load      bool
	Name      string
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
	// c.ScriptHash(s.Hash())
	path := importPath(s, m.Name, "", "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

type ImportModuleRaw struct {
	Specifier string
	Path      string
	Load      bool
	Name      string
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
	// c.ScriptHash(s.Hash())
	path := importPath(s, m.Path, m.Name, "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

type ImportModuleRawBytes struct {
	Specifier string
	Content   []byte
	Load      bool
	Name      string
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
	// c.ScriptHash(s.Hash())
	path := importPath(s, "", m.Name, "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

type ImportModuleBundle struct {
	Specifier string
	Entry     string
	Profile   string
	Load      bool
	Name      string
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
	// c.ScriptHash(s.Hash())
	path := importPath(s, "", m.Name, "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

type ImportModuleBundleFS struct {
	CacheKey  string
	Specifier string
	FS        fs.FS
	Entry     string
	Profile   string
	Load      bool
	Name      string
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
	// c.ScriptHash(s.Hash())
	path := importPath(s, "", m.Name, "js")
	if m.Specifier != "" {
		im.addImport(m.Specifier, path)
	}
	if m.Load {
		im.addRender(importScript(path))
	}
	return nil
}

type ImportModuleHosted struct {
	Specifier string
	Load      bool
	Src       string
}

func (m ImportModuleHosted) print() string {
	return "local module " + m.Src
}

func (m ImportModuleHosted) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	//	c.ScriptSource(m.Src)
	if m.Specifier != "" {
		im.addImport(m.Specifier, m.Src)
	}
	if m.Load {
		im.addRender(importScript(m.Src))
	}
	return nil
}

type ImportModuleExternal struct {
	Specifier string
	Load      bool
	Src       string
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

type ImportStyleHosted struct {
	Href string
}

func (m ImportStyleHosted) print() string {
	return "local style " + m.Href
}

func (m ImportStyleHosted) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	// c.StyleSource(m.Href)
	im.addRender(importStyle(m.Href))
	return nil
}

type ImportStyleExternal struct {
	Href string
}

func (m ImportStyleExternal) print() string {
	return "external style " + m.Href
}

func (m ImportStyleExternal) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	c.StyleSource(m.Href)
	im.addRender(importStyle(m.Href))
	return nil
}

type ImportStyle struct {
	Path string
	Name string
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

type ImportStyleBytes struct {
	Content []byte
	Name    string
}

func (m ImportStyleBytes) print() string {
	return "style bytes"
}

func (m ImportStyleBytes) init(im *importMap, r *resources.Registry, c *common.CSPCollector) error {
	s, err := r.StyleBytes(m.Content)
	if err != nil {
		return err
	}
	// c.StyleHash(s.Hash())
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
