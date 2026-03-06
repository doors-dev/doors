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
	"github.com/doors-dev/doors/internal/router"
	"github.com/doors-dev/gox"
)

type HostMode int

const (
	HostModePublic HostMode = iota
	HostModePrivate
	HostModeNoCache
)

func (h HostMode) src(core core.Core, res *resources.Resource, name string) (string, error) {
	switch h {
	case HostModePublic:
		return router.ResourcePath(res, name), nil
	case HostModePrivate, HostModeNoCache:
		hook, ok := core.RegisterHook(func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool {
			res.Serve(w, r)
			return false
		}, nil)
		if !ok {
			return "", context.Canceled
		}
		return fmt.Sprintf("/~0/%s/%d/%d/%s", core.InstanceID(), hook.DoorID, hook.HookID, name), nil
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
		}
		return resources.FormatCommon{Bundle: true}
	case ScriptOutputRaw:
		return resources.FormatRaw{}
	default:
		panic("unknown script format")
	}
}

type scriptKind int

const (
	scriptModule scriptKind = iota
	scriptImportModule
	scriptCommonJS
	scriptInline
)

type ScriptKind struct {
	kind      scriptKind
	specifier string
}

func ScriptKindModule() ScriptKind {
	return ScriptKind{kind: scriptModule}
}

func ScriptKindImportModule(specifier string) ScriptKind {
	return ScriptKind{kind: scriptImportModule, specifier: specifier}
}

func ScriptKindInline() ScriptKind {
	return ScriptKind{kind: scriptInline}
}

func ScriptKindCommon() ScriptKind {
	return ScriptKind{kind: scriptCommonJS}
}

func (k ScriptKind) module() bool {
	switch k.kind {
	case scriptModule, scriptImportModule:
		return true
	case scriptCommonJS, scriptInline:
		return false
	default:
		panic("unknown script kind")
	}
}

func (k ScriptKind) Edit(cur gox.Cursor, core core.Core, path string) error {
	switch k.kind {
	case scriptModule:
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
		return cur.Close()

	case scriptImportModule:
		core.ModuleRegistry().Add(k.specifier, path)
		return nil

	case scriptCommonJS, scriptInline:
		if err := cur.Init("script"); err != nil {
			return err
		}
		if err := cur.AttrSet("src", path); err != nil {
			return err
		}
		if err := cur.Submit(); err != nil {
			return err
		}
		return cur.Close()

	default:
		panic("unknown script kind")
	}
}

func (k ScriptKind) Modify(core core.Core, tag, path string, attrs gox.Attrs) error {
	switch k.kind {
	case scriptModule, scriptImportModule:
		switch {
		case strings.EqualFold(tag, "script"):
			attrs.Get("type").Set("module")
			attrs.Get("src").Set(path)
		case strings.EqualFold(tag, "link"):
			attrs.Get("rel").Set("modulepreload")
			attrs.Get("href").Set(path)
		default:
			return fmt.Errorf("unsupported tag %s", tag)
		}
		if k.kind == scriptImportModule {
			core.ModuleRegistry().Add(k.specifier, path)
		}
		return nil

	case scriptCommonJS, scriptInline:
		switch {
		case strings.EqualFold(tag, "script"):
			attrs.Get("type").Set(nil)
			attrs.Get("src").Set(path)
			return nil
		default:
			return fmt.Errorf("unsupported tag %s", tag)
		}

	default:
		panic("unknown script kind")
	}
}

type ScriptString struct {
	Kind    ScriptKind
	Output  ScriptOutput
	Content string
	Profile string
	Host    HostMode
}

func (s ScriptString) entry() resources.ScriptEntry {
	switch s.Kind.kind {
	case scriptModule, scriptImportModule, scriptCommonJS:
		return resources.ScriptString{
			Content: s.Content,
			Kind:    resources.KindJS,
		}
	case scriptInline:
		return resources.ScriptInlineString{
			Content: s.Content,
			Kind:    resources.KindJS,
		}
	default:
		panic("unknown script kind")
	}
}

func (s ScriptString) build(core core.Core) (string, error) {
	if s.Kind.kind == scriptInline && s.Output != ScriptOutputDefault {
		return "", errors.New("inline scripts support only ScriptOutputDefault")
	}
	var format resources.ScriptFormat = resources.FormatDefault{}
	if s.Kind.kind != scriptInline {
		format = s.Output.scriptFormat(s.Kind.module())
	}
	res, err := core.ResourceRegistry().Script(
		s.entry(),
		format,
		s.Profile,
		s.Host.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return s.Host.src(core, res, "script.js")
}

func (s ScriptString) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Edit(cur, core, path)
}

func (s ScriptString) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Modify(core, tag, path, attrs)
}

type ScriptBytes struct {
	Kind    ScriptKind
	Content []byte
	Output  ScriptOutput
	Profile string
	Host    HostMode
}

func (s ScriptBytes) entry() resources.ScriptEntry {
	switch s.Kind.kind {
	case scriptModule, scriptImportModule, scriptCommonJS:
		return resources.ScriptBytes{
			Content: s.Content,
			Kind:    resources.KindJS,
		}
	case scriptInline:
		return resources.ScriptInlineBytes{
			Content: s.Content,
			Kind:    resources.KindJS,
		}
	default:
		panic("unknown script kind")
	}
}

func (s ScriptBytes) build(core core.Core) (string, error) {
	if s.Kind.kind == scriptInline && s.Output != ScriptOutputDefault {
		return "", errors.New("inline scripts support only ScriptOutputDefault")
	}
	var format resources.ScriptFormat = resources.FormatDefault{}
	if s.Kind.kind != scriptInline {
		format = s.Output.scriptFormat(s.Kind.module())
	}
	res, err := core.ResourceRegistry().Script(
		s.entry(),
		format,
		s.Profile,
		s.Host.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return s.Host.src(core, res, "script.js")
}

func (s ScriptBytes) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Edit(cur, core, path)
}

func (s ScriptBytes) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Modify(core, tag, path, attrs)
}

type ScriptPath struct {
	Kind    ScriptKind
	Path    string
	Output  ScriptOutput
	Profile string
	Host    HostMode
}

func (s ScriptPath) entry() resources.ScriptEntry {
	switch s.Kind.kind {
	case scriptModule, scriptImportModule, scriptCommonJS:
		return resources.ScriptPath{
			Path: s.Path,
		}
	case scriptInline:
		return resources.ScriptInlinePath{
			Path: s.Path,
		}
	default:
		panic("unknown script kind")
	}
}

func (s ScriptPath) build(core core.Core) (string, error) {
	if s.Kind.kind == scriptInline && s.Output != ScriptOutputDefault {
		return "", errors.New("inline scripts support only ScriptOutputDefault")
	}
	var format resources.ScriptFormat = resources.FormatDefault{}
	if s.Kind.kind != scriptInline {
		format = s.Output.scriptFormat(s.Kind.module())
	}
	res, err := core.ResourceRegistry().Script(
		s.entry(),
		format,
		s.Profile,
		s.Host.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	base := filepath.Base(s.Path)
	fileExt := filepath.Ext(base)
	name := strings.TrimSuffix(base, fileExt) + ".js"
	return s.Host.src(core, res, name)
}

func (s ScriptPath) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Edit(cur, core, path)
}

func (s ScriptPath) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Modify(core, tag, path, attrs)
}

type ScriptFS struct {
	Kind    ScriptKind
	Output  ScriptOutput
	FS      fs.FS
	Path    string
	Name    string
	Profile string
	Host    HostMode
}

func (s ScriptFS) fileName() string {
	if s.Name != "" {
		return s.Name + ".js"
	}
	base := filepath.Base(s.Path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext) + ".js"
}

func (s ScriptFS) entry() resources.ScriptEntry {
	switch s.Kind.kind {
	case scriptModule, scriptImportModule, scriptCommonJS:
		return resources.ScriptFS{
			FS:   s.FS,
			Path: s.Path,
			Name: s.Name,
		}
	case scriptInline:
		return resources.ScriptInlineFS{
			FS:   s.FS,
			Path: s.Path,
			Name: s.Name,
		}
	default:
		panic("unknown script kind")
	}
}

func (s ScriptFS) build(core core.Core) (string, error) {
	if s.Kind.kind == scriptInline && s.Output != ScriptOutputDefault {
		return "", errors.New("inline scripts support only ScriptOutputDefault")
	}
	var format resources.ScriptFormat = resources.FormatDefault{}
	if s.Kind.kind != scriptInline {
		format = s.Output.scriptFormat(s.Kind.module())
	}
	res, err := core.ResourceRegistry().Script(
		s.entry(),
		format,
		s.Profile,
		s.Host.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return s.Host.src(core, res, s.fileName())
}

func (s ScriptFS) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Edit(cur, core, path)
}

func (s ScriptFS) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Modify(core, tag, path, attrs)
}

type ScriptExternal struct {
	Kind ScriptKind
	Src  string
}

func (s ScriptExternal) build(core core.Core) (string, error) {
	if s.Kind.kind == scriptInline {
		return "", errors.New("inline script kind is not supported for external scripts")
	}
	core.CSPCollector().ScriptSource(s.Src)
	return s.Src, nil
}

func (s ScriptExternal) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Edit(cur, core, path)
}

func (s ScriptExternal) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Modify(core, tag, path, attrs)
}

type ScriptLocal struct {
	Kind ScriptKind
	Src  string
}

func (s ScriptLocal) build(core.Core) (string, error) {
	if s.Kind.kind == scriptInline {
		return "", errors.New("inline script kind is not supported for local scripts")
	}
	return s.Src, nil
}

func (s ScriptLocal) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Edit(cur, core, path)
}

func (s ScriptLocal) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	return s.Kind.Modify(core, tag, path, attrs)
}

type StyleString struct {
	Content string
	Minify  bool
	Host    HostMode
}

func (s StyleString) entry() resources.StyleEntry {
	return resources.StyleString{
		Content: s.Content,
	}
}

func (s StyleString) build(core core.Core) (string, error) {
	res, err := core.ResourceRegistry().Style(
		s.entry(),
		s.Minify,
		s.Host.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return s.Host.src(core, res, "style.css")
}

func (s StyleString) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
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

func (s StyleString) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(tag, "link"):
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(path)
		return nil
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
}

type StyleBytes struct {
	Content []byte
	Minify  bool
	Host    HostMode
}

func (s StyleBytes) entry() resources.StyleEntry {
	return resources.StyleBytes{
		Content: s.Content,
	}
}

func (s StyleBytes) build(core core.Core) (string, error) {
	res, err := core.ResourceRegistry().Style(
		s.entry(),
		s.Minify,
		s.Host.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return s.Host.src(core, res, "style.css")
}

func (s StyleBytes) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
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

func (s StyleBytes) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(tag, "link"):
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(path)
		return nil
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
}

type StylePath struct {
	Path   string
	Minify bool
	Host   HostMode
}

func (s StylePath) entry() resources.StyleEntry {
	return resources.StylePath{
		Path: s.Path,
	}
}

func (s StylePath) build(core core.Core) (string, error) {
	res, err := core.ResourceRegistry().Style(
		s.entry(),
		s.Minify,
		s.Host.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	base := filepath.Base(s.Path)
	fileExt := filepath.Ext(base)
	name := strings.TrimSuffix(base, fileExt) + ".css"
	return s.Host.src(core, res, name)
}

func (s StylePath) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
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

func (s StylePath) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(tag, "link"):
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(path)
		return nil
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
}

type StyleFS struct {
	FS     fs.FS
	Path   string
	Name   string
	Minify bool
	Host   HostMode
}

func (s StyleFS) fileName() string {
	if s.Name != "" {
		return s.Name + ".css"
	}
	base := filepath.Base(s.Path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext) + ".css"
}

func (s StyleFS) entry() resources.StyleEntry {
	return resources.StyleFS{
		FS:   s.FS,
		Path: s.Path,
		Name: s.Name,
	}
}

func (s StyleFS) build(core core.Core) (string, error) {
	res, err := core.ResourceRegistry().Style(
		s.entry(),
		s.Minify,
		s.Host.resourceMode(),
	)
	if err != nil {
		return "", err
	}
	return s.Host.src(core, res, s.fileName())
}

func (s StyleFS) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
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

func (s StyleFS) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(tag, "link"):
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(path)
		return nil
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
}

type StyleExternal struct {
	Src string
}

func (s StyleExternal) build(core core.Core) (string, error) {
	core.CSPCollector().StyleSource(s.Src)
	return s.Src, nil
}

func (s StyleExternal) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
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

func (s StyleExternal) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(tag, "link"):
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(path)
		return nil
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
}

type StyleLocal struct {
	Src string
}

func (s StyleLocal) build(core.Core) (string, error) {
	return s.Src, nil
}

func (s StyleLocal) Edit(cur gox.Cursor) error {
	core := cur.Context().Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
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

func (s StyleLocal) Modify(ctx context.Context, tag string, attrs gox.Attrs) error {
	core := ctx.Value(ctex.KeyCore).(core.Core)
	path, err := s.build(core)
	if err != nil {
		return err
	}
	switch {
	case strings.EqualFold(tag, "link"):
		attrs.Get("rel").Set("stylesheet")
		attrs.Get("href").Set(path)
		return nil
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
}
