//go:build doors_imports_disabled

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
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/doors-dev/doors/internal/core"
	"github.com/doors-dev/doors/internal/ctex"
	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

type ScriptSource interface {
	name(ext string) string
	scriptEntry(inline bool) resources.ScriptEntry
}

type StyleSource interface {
	name(ext string) string
	styleEntry() resources.StyleEntry
}

type HostMode int

const (
	HostModePublic HostMode = iota
	HostModePrivate
	HostModeNoCache
)

func (h HostMode) src(core core.Core, res *resources.Resource, name string) (string, error) {
	switch h {
	case HostModePublic:
		return core.PathMaker().Resource(res, name), nil
	case HostModePrivate, HostModeNoCache:
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			res.Serve(w, r)
			return false
		}, nil)
		if !ok {
			return "", context.Canceled
		}
		return core.PathMaker().Hook(core.InstanceID(), hook.DoorID, hook.HookID, name), nil
	default:
		panic("wrong host mode")
	}
}

func (h HostMode) resourceMode() resources.ResourceMode {
	switch h {
	case HostModePrivate:
		return resources.ModeCache
	case HostModePublic:
		return resources.ModeHost
	case HostModeNoCache:
		return resources.ModeNoCache
	default:
		panic("wrong host mode")
	}
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

func (s SourceFS) scriptEntry(inline bool) resources.ScriptEntry {
	if inline {
		return resources.ScriptInlineFS{
			FS:   s.FS,
			Name: s.Name,
			Path: s.Path,
		}
	}
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

func (s SourcePath) scriptEntry(inline bool) resources.ScriptEntry {
	if inline {
		return resources.ScriptInlinePath{
			Path: string(s),
		}
	}
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
	return "style." + ext
}

func (s SourceStyleString) styleEntry() resources.StyleEntry {
	return resources.StyleString{
		Content: string(s),
	}
}

type SourceStyleBytes []byte

func (s SourceStyleBytes) name(ext string) string {
	return "style." + ext
}

func (s SourceStyleBytes) styleEntry() resources.StyleEntry {
	return resources.StyleBytes{
		Content: s,
	}
}

type SourceScriptBytes struct {
	Content    []byte
	TypeScript bool
}

func (s SourceScriptBytes) name(ext string) string {
	return "script." + ext
}

func (s SourceScriptBytes) scriptEntry(inline bool) resources.ScriptEntry {
	kind := resources.KindJS
	if s.TypeScript {
		kind = resources.KindTS
	}
	if inline {
		return resources.ScriptInlineBytes{
			Content: s.Content,
			Kind:    kind,
		}
	}
	return resources.ScriptBytes{
		Content: s.Content,
		Kind:    kind,
	}
}

type SourceScriptString struct {
	Content    string
	TypeScript bool
}

func (s SourceScriptString) name(ext string) string {
	return "script." + ext
}

func (s SourceScriptString) scriptEntry(inline bool) resources.ScriptEntry {
	kind := resources.KindJS
	if s.TypeScript {
		kind = resources.KindTS
	}
	if inline {
		return resources.ScriptInlineString{
			Content: s.Content,
			Kind:    kind,
		}
	}
	return resources.ScriptString{
		Content: s.Content,
		Kind:    kind,
	}
}

type SourceExternal string

func (s SourceExternal) name(ext string) string {
	return ""
}

func (s SourceExternal) scriptEntry(inline bool) resources.ScriptEntry {
	panic("external source can't provide script entry")
}

func (s SourceExternal) styleEntry() resources.StyleEntry {
	panic("external source can't provide style entry")
}

type SourceLocal string

func (s SourceLocal) name(ext string) string {
	return ""
}

func (s SourceLocal) scriptEntry(inline bool) resources.ScriptEntry {
	panic("local source can't provide script entry")
}

func (s SourceLocal) styleEntry() resources.StyleEntry {
	panic("local source can't provide style entry")
}

type ScriptOutput int

const (
	ScriptOutputDefault ScriptOutput = iota
	ScriptOutputBundle
	ScriptOutputRaw
)

func (f ScriptOutput) scriptFormat(module bool) resources.ScriptFormat {
	switch f {
	case ScriptOutputDefault:
		return resources.FormatDefault{}
	case ScriptOutputBundle:
		if module {
			return resources.FormatModule{Bundle: true}
		} else {
			return resources.FormatCommon{Bundle: true}
		}
	case ScriptOutputRaw:
		return resources.FormatRaw{}
	default:
		panic("unknown script format")
	}
}

type ScriptModule struct {
	Source    ScriptSource
	Output    ScriptOutput
	HostMode  HostMode
	Specifier string
	Profile   string
}

func (m ScriptModule) build(core core.Core) (string, error) {
	if loc, ok := m.Source.(SourceLocal); ok {
		return string(loc), nil
	}
	if ext, ok := m.Source.(SourceExternal); ok {
		core.CSPCollector().ScriptSource(string(ext))
		return string(ext), nil
	}
	entry := m.Source.scriptEntry(false)
	res, err := core.ResourceRegistry().Script(
		entry,
		m.Output.scriptFormat(true),
		m.Profile,
		m.HostMode.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return m.HostMode.src(core, res, m.Source.name("js"))
}

func (m ScriptModule) Edit(cur gox.Cursor) error {
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

func (m ScriptModule) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
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

type ScriptInline struct {
	Source   ScriptSource
	HostMode HostMode
	Profile  string
}

func (m ScriptInline) build(core core.Core) (string, error) {
	if _, ok := m.Source.(SourceLocal); ok {
		return "", errors.New("Local source is not supported for \"inline\" scripts")
	}
	if _, ok := m.Source.(SourceExternal); ok {
		return "", errors.New("External source is not supported for \"inline\" scripts")
	}
	entry := m.Source.scriptEntry(true)
	res, err := core.ResourceRegistry().Script(
		entry,
		resources.FormatDefault{},
		m.Profile,
		m.HostMode.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return m.HostMode.src(core, res, m.Source.name("js"))
}

func (m ScriptInline) Edit(cur gox.Cursor) error {
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

func (m ScriptInline) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := m.build(core)
	if err != nil {
		return err
	}
	switch true {
	case strings.EqualFold(tag, "script"):
		attrs.Get("type").Set(nil)
		attrs.Get("src").Set(path)
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
	return nil
}

type ScriptCommon struct {
	Source   ScriptSource
	HostMode HostMode
	Output   ScriptOutput
	Profile  string
}

func (m ScriptCommon) build(core core.Core) (string, error) {
	if loc, ok := m.Source.(SourceLocal); ok {
		return string(loc), nil
	}
	if ext, ok := m.Source.(SourceExternal); ok {
		core.CSPCollector().ScriptSource(string(ext))
		return string(ext), nil
	}
	entry := m.Source.scriptEntry(false)
	res, err := core.ResourceRegistry().Script(
		entry,
		m.Output.scriptFormat(false),
		m.Profile,
		m.HostMode.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return m.HostMode.src(core, res, m.Source.name("js"))
}

func (m ScriptCommon) Edit(cur gox.Cursor) error {
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

func (m ScriptCommon) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := m.build(core)
	if err != nil {
		return err
	}
	switch true {
	case strings.EqualFold(tag, "script"):
		attrs.Get("type").Set(nil)
		attrs.Get("src").Set(path)
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
	return nil
}

type Style struct {
	Source   StyleSource
	HostMode HostMode
	Minify   bool
}

func (m Style) build(core core.Core) (string, error) {
	if loc, ok := m.Source.(SourceLocal); ok {
		return string(loc), nil
	}
	if ext, ok := m.Source.(SourceExternal); ok {
		core.CSPCollector().StyleSource(string(ext))
		return string(ext), nil
	}
	res, err := core.ResourceRegistry().Style(
		m.Source.styleEntry(),
		m.Minify,
		m.HostMode.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return m.HostMode.src(core, res, m.Source.name("css"))
}

func (m Style) Edit(cur gox.Cursor) error {
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

func (m Style) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
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
