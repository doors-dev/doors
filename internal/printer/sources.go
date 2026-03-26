// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package printer

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/doors-dev/doors/internal/resources"
)

type ScriptSource interface {
	name(ext string) string
	scriptEntry(inline bool) resources.ScriptEntry
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
