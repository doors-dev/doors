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

type ScriptProvider interface {
	name(ext string) string
	scriptEntry() resources.ScriptEntry
}

type StyleProvider interface {
	name(ext string) string
	styleEntry() resources.StyleEntry
}

type ProviderFS struct {
	FS   fs.FS
	Path string
	Name string
}

func (s ProviderFS) name(ext string) string {

	return s.Name + "." + ext
}

func (s ProviderFS) styleEntry() resources.StyleEntry {
	return resources.StyleFS{
		FS:   s.FS,
		Name: s.Name,
		Path: s.Path,
	}
}

func (s ProviderFS) scriptEntry() resources.ScriptEntry {
	return resources.ScriptFS{
		FS:   s.FS,
		Name: s.Name,
		Path: s.Path,
	}
}

type ProviderPath string

func (s ProviderPath) name(ext string) string {
	base := filepath.Base(string(s))
	fileExt := filepath.Ext(base)
	name := strings.TrimSuffix(base, fileExt)
	return name + "." + ext
}

func (s ProviderPath) scriptEntry() resources.ScriptEntry {
	return resources.ScriptPath{
		Path: string(s),
	}
}

func (s ProviderPath) styleEntry() resources.StyleEntry {
	return resources.StylePath{
		Path: string(s),
	}
}

type ProviderStyleString string

func (s ProviderStyleString) name(ext string) string {
	return ext
}

func (s ProviderStyleString) styleEntry() resources.StyleEntry {
	return resources.StyleString{
		Content: string(s),
	}
}

type ProviderStyleBytes []byte

func (s ProviderStyleBytes) name(ext string) string {
	return ext
}

func (s ProviderStyleBytes) styleEntry() resources.StyleEntry {
	return resources.StyleBytes{
		Content: s,
	}
}

type ProviderTsString string

func (s ProviderTsString) name(ext string) string {
	return ext
}

func (s ProviderTsString) scriptEntry() resources.ScriptEntry {
	return resources.ScriptString{
		Content: string(s),
		Kind:    resources.KindTS,
	}
}

type ProviderTsBytes []byte

func (s ProviderTsBytes) name(ext string) string {
	return ext
}

func (s ProviderTsBytes) scriptEntry() resources.ScriptEntry {
	return resources.ScriptBytes{
		Content: s,
		Kind:    resources.KindTS,
	}
}

type ProviderJsString string

func (s ProviderJsString) name(ext string) string {
	return ext
}

func (s ProviderJsString) scriptEntry() resources.ScriptEntry {
	return resources.ScriptString{
		Content: string(s),
		Kind:    resources.KindJS,
	}
}

type ProviderJsBytes []byte

func (s ProviderJsBytes) name(ext string) string {
	return ext
}

func (s ProviderJsBytes) scriptEntry() resources.ScriptEntry {
	return resources.ScriptBytes{
		Content: s,
		Kind:    resources.KindJS,
	}
}

type ProviderExternal string

func (s ProviderExternal) name(ext string) string {
	return ""
}

func (s ProviderExternal) scriptEntry() resources.ScriptEntry {
	panic("exteral source can't provide script entry")
}

func (s ProviderExternal) styleEntry() resources.StyleEntry {
	panic("exteral source can't provide style entry")
}

type ProviderLocal string

func (s ProviderLocal) name(ext string) string {
	return ""
}

func (s ProviderLocal) scriptEntry() resources.ScriptEntry {
	panic("local source can't provide script entry")
}

func (s ProviderLocal) styleEntry() resources.StyleEntry {
	panic("local source can't provide style entry")
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

type ImportModule struct {
	Provider  ScriptProvider
	Format    ScriptFormat
	Specifier string
	Profile   string
}

func (m ImportModule) build(core core.Core) (string, error) {
	if loc, ok := m.Provider.(ProviderLocal); ok {
		return string(loc), nil
	}
	if ext, ok := m.Provider.(ProviderExternal); ok {
		core.CSPCollector().ScriptSource(string(ext))
		return string(ext), nil
	}
	res, err := core.ResourceRegistry().Script(
		m.Provider.scriptEntry(),
		m.Format.scriptFormat(true),
		m.Profile,
		resources.ModeHost,
	)
	if err != nil {
		return "", err
	}
	path := router.ResourcePath(res, m.Provider.name("js"))
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
	if m.Specifier != "" {
		core.ModuleRegistry().Add(m.Specifier, path)
	}
	return nil
}

type ImportCommonJS struct {
	Provider ScriptProvider
	Format   ScriptFormat
	Profile  string
}

func (m ImportCommonJS) build(core core.Core) (string, error) {
	if loc, ok := m.Provider.(ProviderLocal); ok {
		return string(loc), nil
	}
	if ext, ok := m.Provider.(ProviderExternal); ok {
		core.CSPCollector().ScriptSource(string(ext))
		return string(ext), nil
	}
	res, err := core.ResourceRegistry().Script(
		m.Provider.scriptEntry(),
		m.Format.scriptFormat(false),
		m.Profile,
		resources.ModeHost,
	)
	if err != nil {
		return "", err
	}
	path := router.ResourcePath(res, m.Provider.name("js"))
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
	Provider StyleProvider
	Minify   bool
}

func (m ImportStyle) build(core core.Core) (string, error) {
	if loc, ok := m.Provider.(ProviderLocal); ok {
		return string(loc), nil
	}
	if ext, ok := m.Provider.(ProviderExternal); ok {
		core.CSPCollector().StyleSource(string(ext))
		return string(ext), nil
	}
	res, err := core.ResourceRegistry().Style(
		m.Provider.styleEntry(),
		m.Minify,
		resources.ModeHost,
	)
	if err != nil {
		return "", err
	}
	path := router.ResourcePath(res, m.Provider.name("css"))
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
