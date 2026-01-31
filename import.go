// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package doors

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/doors/internal/router"
	"github.com/doors-dev/gox"
)

type ScriptSource interface {
	name(ext string) string
	scriptEntry() resources.ScriptEntry
}

type StyleSource interface {
	name(ext string) string
	styleEntry() resources.StyleEntry
}

type SourceFS struct {
	FS   fs.FS
	Path string
	Name string
}

func (s SourceFS) name(ext string) string {

	return s.Name + "." + ext
}

func (s SourceFS) styleEntry() resources.StyleEntry {
	return resources.StyleFS{
		FS:   s.FS,
		Name: s.Name,
		Path: s.Path,
	}
}

func (s SourceFS) scriptEntry() resources.ScriptEntry {
	return resources.ScriptFS{
		FS:   s.FS,
		Name: s.Name,
		Path: s.Path,
	}
}

type SourcePath string

func (s SourcePath) name(ext string) string {
	base := filepath.Base(string(s))
	fileExt := filepath.Ext(base)
	name := strings.TrimSuffix(base, fileExt)
	return name + "." + ext
}

func (s SourcePath) scriptEntry() resources.ScriptEntry {
	return resources.ScriptPath{
		Path: string(s),
	}
}

func (s SourcePath) styleEntry() resources.StyleEntry {
	return resources.StylePath{
		Path: string(s),
	}
}

type SourceStyleString string

func (s SourceStyleString) name(ext string) string {
	return ext
}

func (s SourceStyleString) styleEntry() resources.StyleEntry {
	return resources.StyleString{
		Content: string(s),
	}
}

type SourceStyleBytes []byte

func (s SourceStyleBytes) name(ext string) string {
	return ext
}

func (s SourceStyleBytes) styleEntry() resources.StyleEntry {
	return resources.StyleBytes{
		Content: s,
	}
}

type SourceTsString string

func (s SourceTsString) name(ext string) string {
	return ext
}

func (s SourceTsString) scriptEntry() resources.ScriptEntry {
	return resources.ScriptString{
		Content: string(s),
		Kind:    resources.KindTS,
	}
}

type SourceTsBytes []byte

func (s SourceTsBytes) name(ext string) string {
	return ext
}

func (s SourceTsBytes) scriptEntry() resources.ScriptEntry {
	return resources.ScriptBytes{
		Content: s,
		Kind:    resources.KindTS,
	}
}

type SourceJsString string

func (s SourceJsString) name(ext string) string {
	return ext
}

func (s SourceJsString) scriptEntry() resources.ScriptEntry {
	return resources.ScriptString{
		Content: string(s),
		Kind:    resources.KindJS,
	}
}

type SourceJsBytes []byte

func (s SourceJsBytes) name(ext string) string {
	return ext
}

func (s SourceJsBytes) scriptEntry() resources.ScriptEntry {
	return resources.ScriptBytes{
		Content: s,
		Kind:    resources.KindJS,
	}
}

type SourceExternal string

func (s SourceExternal) name(ext string) string {
	return ""
}

func (s SourceExternal) scriptEntry() resources.ScriptEntry {
	panic("exteral source can't provide script entry")
}

func (s SourceExternal) styleEntry() resources.StyleEntry {
	panic("exteral source can't provide style entry")
}

type ScriptFormat int

const (
	FormatDefault ScriptFormat = iota
	FormatBundle
	FormatRaw
)

func (f ScriptFormat) scriptFormat(module bool) resources.ScriptFormat {
	switch f {
	case FormatDefault:
		return resources.FormatDefault{}
	case FormatBundle:
		if module {
			return resources.FormatModule{Bundle: true}
		} else {
			return resources.FormatCommon{Bundle: true}
		}
	case FormatRaw:
		return resources.FormatRaw{}
	default:
		panic("unknown script format")
	}
}

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

type ImportModule struct {
	Source    ScriptSource
	Format    ScriptFormat
	Specifier string
	Profile   string
}

func (m ImportModule) build(core core.Core) (string, error) {
	if ext, ok := m.Source.(SourceExternal); ok {
		core.CSPCollector().ScriptSource(string(ext))
		return string(ext), nil
	}
	res, err := core.ResourceRegistry().Script(
		m.Source.scriptEntry(),
		m.Format.scriptFormat(true),
		m.Profile,
		resources.ModeHost,
	)
	if err != nil {
		return "", err
	}
	path := router.ResourcePath(res, m.Source.name("js"))
	return path, nil
}

func (m ImportModule) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := m.build(core)
	if err != nil {
		return err
	}
	if m.Specifier == "" {
		if err := cur.Init("script"); err != nil {
			return err
		}
		if err := cur.AttrSet("src", path); err != nil {
			return err
		}
		if err := cur.AttrSet("type", "module"); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		if err := cur.Close(); err != nil {
			return err
		}
		return nil
	}
	core.ModuleRegistry().Add(m.Specifier, path)
	return nil
}

func (m ImportModule) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := m.build(core)
	if err != nil {
		return err
	}
	switch true {
	case strings.EqualFold(tag, "script"):
		attrs.Get("type").Set("module")
		attrs.Get("src").Set(path)
	case strings.EqualFold(tag, "link"):
		attrs.Get("rel").Set("modulepreload")
		attrs.Get("href").Set(path)
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
	return nil
}

type ImportCommonJS struct {
	Source  ScriptSource
	Format  ScriptFormat
	Profile string
}

func (m ImportCommonJS) build(core core.Core) (string, error) {
	if ext, ok := m.Source.(SourceExternal); ok {
		core.CSPCollector().ScriptSource(string(ext))
		return string(ext), nil
	}
	res, err := core.ResourceRegistry().Script(
		m.Source.scriptEntry(),
		m.Format.scriptFormat(false),
		m.Profile,
		resources.ModeHost,
	)
	if err != nil {
		return "", err
	}
	path := router.ResourcePath(res, m.Source.name("js"))
	return path, nil
}

func (m ImportCommonJS) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := m.build(core)
	if err != nil {
		return err
	}
	if err := cur.Init("script"); err != nil {
		return err
	}
	if err := cur.AttrSet("src", path); err != nil {
		return err
	}
	if err := cur.Submit(); err != nil {
		return err
	}
	if err := cur.Close(); err != nil {
		return err
	}
	return nil
}

func (m ImportCommonJS) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := m.build(core)
	if err != nil {
		return err
	}
	switch true {
	case strings.EqualFold(tag, "script"):
		attrs.Get("type").SetBool(false)
		attrs.Get("src").Set(path)
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
	return nil
}

type ImportStyle struct {
	Source StyleSource
	Minify bool
}

func (m ImportStyle) build(core core.Core) (string, error) {
	if ext, ok := m.Source.(SourceExternal); ok {
		core.CSPCollector().StyleSource(string(ext))
		return string(ext), nil
	}
	res, err := core.ResourceRegistry().Style(
		m.Source.styleEntry(),
		m.Minify,
		resources.ModeHost,
	)
	if err != nil {
		return "", err
	}
	path := router.ResourcePath(res, m.Source.name("css"))
	return path, nil
}

func (m ImportStyle) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := m.build(core)
	if err != nil {
		return err
	}
	if err := cur.InitVoid("link"); err != nil {
		return err
	}
	if err := cur.AttrSet("rel", "stylesheet"); err != nil {
		return err
	}
	if err := cur.AttrSet("href", path); err != nil {
		return err
	}
	return cur.Submit()
}

func (m ImportStyle) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := m.build(core)
	if err != nil {
		return err
	}
	switch true {
	case strings.EqualFold(tag, "link"):
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(path)
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
	return nil
}
