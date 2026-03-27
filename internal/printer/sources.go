// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package printer

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/doors-dev/doors/internal/resources"
	"github.com/doors-dev/gox"
)

type Source interface {
	Handler() HandlerFunc
	name(ext string) string
	scriptEntry(inline bool, ts bool) resources.ScriptEntry
	styleEntry() resources.StyleEntry
	gox.Modify
}

type SourceStatic interface {
	Source
	StaticEntry() resources.StaticEntry
}

type SourceFS struct {
	FS    fs.FS
	Entry string
}

func (s SourceFS) Handler() HandlerFunc {
	return func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
		http.ServeFileFS(w, r, s.FS, s.Entry)
		return false
	}
}

func (s SourceFS) name(ext string) string {
	return sourceNameFromPath(s.Entry, ext)
}

func (s SourceFS) StaticEntry() resources.StaticEntry {
	return resources.StaticFS{
		FS:   s.FS,
		Path: s.Entry,
	}
}

func (s SourceFS) scriptEntry(inline bool, ts bool) resources.ScriptEntry {
	if inline {
		return resources.ScriptInlineFS{
			FS:   s.FS,
			Path: s.Entry,
		}
	}
	return resources.ScriptFS{
		FS:   s.FS,
		Path: s.Entry,
	}
}

func (s SourceFS) styleEntry() resources.StyleEntry {
	return resources.StyleFS{
		FS:   s.FS,
		Path: s.Entry,
	}
}

func (s SourceFS) Modify(_ context.Context, tag string, attrs gox.Attrs) error {
	return modifySource(tag, attrs, s)
}

var _ Source = SourceFS{}
var _ SourceStatic = SourceFS{}

type SourceLocalFS string

func (s SourceLocalFS) Handler() HandlerFunc {
	return func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
		http.ServeFile(w, r, string(s))
		return false
	}
}

func (s SourceLocalFS) name(ext string) string {
	return sourceNameFromPath(string(s), ext)
}

func (s SourceLocalFS) StaticEntry() resources.StaticEntry {
	return resources.StaticPath{
		Path: string(s),
	}
}

func (s SourceLocalFS) scriptEntry(inline bool, ts bool) resources.ScriptEntry {
	if inline {
		return resources.ScriptInlinePath{
			Path: string(s),
		}
	}
	return resources.ScriptPath{
		Path: string(s),
	}
}

func (s SourceLocalFS) styleEntry() resources.StyleEntry {
	return resources.StylePath{
		Path: string(s),
	}
}

func (s SourceLocalFS) Modify(_ context.Context, tag string, attrs gox.Attrs) error {
	return modifySource(tag, attrs, s)
}

var _ Source = SourceLocalFS("")
var _ SourceStatic = SourceLocalFS("")

type SourceBytes []byte

func (s SourceBytes) Handler() HandlerFunc {
	return func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
		_, _ = w.Write(s)
		return false
	}
}

func (s SourceBytes) name(ext string) string {
	return sourceDefaultName(ext)
}

func (s SourceBytes) StaticEntry() resources.StaticEntry {
	return resources.StaticBytes{
		Content: s,
	}
}

func (s SourceBytes) scriptEntry(inline bool, ts bool) resources.ScriptEntry {
	kind := resources.KindJS
	if ts {
		kind = resources.KindTS
	}
	if inline {
		return resources.ScriptInlineBytes{
			Content: s,
			Kind:    kind,
		}
	}
	return resources.ScriptBytes{
		Content: s,
		Kind:    kind,
	}
}

func (s SourceBytes) styleEntry() resources.StyleEntry {
	return resources.StyleBytes{
		Content: s,
	}
}

func (s SourceBytes) Modify(_ context.Context, tag string, attrs gox.Attrs) error {
	return modifySource(tag, attrs, s)
}

var _ Source = SourceBytes(nil)
var _ SourceStatic = SourceBytes(nil)

type SourceHook func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool

func (s SourceHook) Handler() HandlerFunc {
	return HandlerFunc(s)
}

func (s SourceHook) name(ext string) string {
	return sourceDefaultName(ext)
}

func (s SourceHook) scriptEntry(inline bool, ts bool) resources.ScriptEntry {
	return nil
}

func (s SourceHook) styleEntry() resources.StyleEntry {
	return nil
}

func (s SourceHook) Modify(_ context.Context, tag string, attrs gox.Attrs) error {
	return modifySource(tag, attrs, s)
}

var _ Source = SourceHook(nil)

type SourceProxy string

func (s SourceProxy) Handler() HandlerFunc {
	target, err := url.Parse(string(s))
	if err != nil {
		return func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
			http.Error(w, "invalid proxy source", http.StatusInternalServerError)
			return false
		}
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			outURL := *target
			r.Out.URL = &outURL
			r.Out.Host = target.Host
		},
	}
	return func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
		proxy.ServeHTTP(w, r)
		return false
	}
}

func (s SourceProxy) name(ext string) string {
	return sourceDefaultName(ext)
}

func (s SourceProxy) scriptEntry(inline bool, ts bool) resources.ScriptEntry {
	return nil
}

func (s SourceProxy) styleEntry() resources.StyleEntry {
	return nil
}

func (s SourceProxy) Modify(_ context.Context, tag string, attrs gox.Attrs) error {
	return modifySource(tag, attrs, s)
}

var _ Source = SourceProxy("")

type SourceString string

func (s SourceString) Handler() HandlerFunc {
	return func(_ context.Context, w http.ResponseWriter, r *http.Request) bool {
		_, _ = w.Write([]byte(s))
		return false
	}
}

func (s SourceString) name(ext string) string {
	return sourceDefaultName(ext)
}

func (s SourceString) StaticEntry() resources.StaticEntry {
	return resources.StaticString{
		Content: string(s),
	}
}

func (s SourceString) scriptEntry(inline bool, ts bool) resources.ScriptEntry {
	kind := resources.KindJS
	if ts {
		kind = resources.KindTS
	}
	if inline {
		return resources.ScriptInlineString{
			Content: string(s),
			Kind:    kind,
		}
	}
	return resources.ScriptString{
		Content: string(s),
		Kind:    kind,
	}
}

func (s SourceString) styleEntry() resources.StyleEntry {
	return resources.StyleString{
		Content: string(s),
	}
}

func (s SourceString) Modify(_ context.Context, tag string, attrs gox.Attrs) error {
	return modifySource(tag, attrs, s)
}

var _ Source = SourceString("")
var _ SourceStatic = SourceString("")

type SourceExternal string

func (s SourceExternal) Handler() HandlerFunc {
	return nil
}

func (s SourceExternal) name(ext string) string {
	return ""
}

func (s SourceExternal) scriptEntry(inline bool, ts bool) resources.ScriptEntry {
	return nil
}

func (s SourceExternal) styleEntry() resources.StyleEntry {
	return nil
}

func (s SourceExternal) Modify(_ context.Context, tag string, attrs gox.Attrs) error {
	return modifySource(tag, attrs, s)
}

var _ Source = SourceExternal("")

type HandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) bool
type HandlerSimpleFunc = func(w http.ResponseWriter, r *http.Request)

func modifySource(tag string, attrs gox.Attrs, src any) error {
	switch true {
	case strings.EqualFold(tag, "a"),
		strings.EqualFold(tag, "area"),
		strings.EqualFold(tag, "base"),
		strings.EqualFold(tag, "link"):
		attrs.Get("href").Set(src)
		return nil
	case strings.EqualFold(tag, "audio"),
		strings.EqualFold(tag, "embed"),
		strings.EqualFold(tag, "iframe"),
		strings.EqualFold(tag, "img"),
		strings.EqualFold(tag, "input"),
		strings.EqualFold(tag, "script"),
		strings.EqualFold(tag, "source"),
		strings.EqualFold(tag, "track"),
		strings.EqualFold(tag, "video"):
		attrs.Get("src").Set(src)
		return nil
	default:
		return fmt.Errorf("unsupported tag %s", tag)
	}
}

func sourceDefaultName(ext string) string {
	if strings.EqualFold(ext, "css") {
		return "style.css"
	}
	return "script." + ext
}

func sourceNameFromPath(path string, ext string) string {
	base := filepath.Base(path)
	if base == "." || base == "" {
		return sourceDefaultName(ext)
	}
	oldExt := filepath.Ext(base)
	if oldExt != "" {
		base = strings.TrimSuffix(base, oldExt)
	}
	if base == "" {
		return sourceDefaultName(ext)
	}
	return base + "." + ext
}
