// doors
// Copyright (c) 2026 doors dev LLC
//
// Dual-licensed: AGPL-3.0-only (see LICENSE) OR a commercial license.
// Commercial inquiries: sales@doors.dev
//
// SPDX-License-Identifier: AGPL-3.0-only OR LicenseRef-doors-commercial

package resources

import (
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/doors-dev/doors/internal/common"
	"github.com/evanw/esbuild/pkg/api"
)

type idWriter interface {
	io.Writer
	io.StringWriter
}

type StaticEntry interface {
	Read() ([]byte, error)
	entryID(h idWriter)
}

type StaticFS struct {
	FS   fs.FS
	Path string
	Name string
}

func (e StaticFS) Read() ([]byte, error) {
	return fs.ReadFile(e.FS, e.Path)
}

func (e StaticFS) entryID(w idWriter) {
	w.WriteString("fs")
	w.WriteString(e.Path)
	if e.Name == "" {
		b, _ := e.Read()
		w.Write(b)
		return
	}
	w.WriteString(e.Name)
}

type StaticPath struct {
	Path string
}

func (e StaticPath) Read() ([]byte, error) {
	return os.ReadFile(e.Path)
}

func (e StaticPath) entryID(w idWriter) {
	w.WriteString("path")
	w.WriteString(e.Path)
}

type StaticBytes struct {
	Content []byte
}

func (e StaticBytes) Read() ([]byte, error) {
	return e.Content, nil
}

func (e StaticBytes) entryID(w idWriter) {
	w.WriteString("content")
	w.Write(e.Content)
}

type StaticString struct {
	Content string
}

func (e StaticString) Read() ([]byte, error) {
	return common.AsBytes(e.Content), nil
}

func (e StaticString) entryID(w idWriter) {
	w.WriteString("content")
	w.WriteString(e.Content)
}

type ScriptEntry interface {
	Read() ([]byte, error)
	Apply(*api.BuildOptions) error
	entryID(h idWriter)
}

type ScriptFS struct {
	FS   fs.FS
	Path string
	Name string
}

func (e ScriptFS) Read() ([]byte, error) {
	return fs.ReadFile(e.FS, e.Path)
}

func (e ScriptFS) Apply(opt *api.BuildOptions) error {
	opt.EntryPoints = []string{e.Path}
	opt.Plugins = append(opt.Plugins, fsPlugin(e.FS))
	return nil
}

func (e ScriptFS) entryID(w idWriter) {
	w.WriteString("fs")
	w.WriteString(e.Path)
	if e.Name == "" {
		b, _ := e.Read()
		w.Write(b)
		return
	}
	w.WriteString(e.Name)
}

type ScriptInlineFS struct {
	FS   fs.FS
	Path string
	Name string
}

func inlineScriptArgs(kind Kind) string {
	if kind != KindTS {
		return "$on, $data, $hook, $fetch, $G, $sys, HookErr"
	}
	return "$on: (name: string, handler: (arg: any, err?: Error) => any) => void, " +
		"$data: <T = any>(name: string) => T | Promise<ArrayBuffer>, " +
		"$hook: (name: string, arg?: any) => Promise<any>, " +
		"$fetch: (name: string, arg?: any) => Promise<Response>, " +
		"$G: {[key: string]: any}, " +
		"$sys: {ready:() => Promise<void>, clean: (handler: () => void | Promise<void>) => void, activateLinks: () => void}, " +
		"HookErr: new (...args: any[]) => Error"
}

func (e ScriptInlineFS) Read() ([]byte, error) {
	return fs.ReadFile(e.FS, e.Path)
}

func (e ScriptInlineFS) Apply(opt *api.BuildOptions) error {
	data, err := e.Read()
	if err != nil {
		return err
	}
	kind := KindJS
	if strings.HasSuffix(strings.ToLower(e.Path), ".ts") {
		kind = KindTS
	}
	opt.Stdin = &api.StdinOptions{
		Contents:   "_d0r(document.currentScript, async (" + inlineScriptArgs(kind) + ") => {\n" + string(data) + "\n})",
		Sourcefile: "index." + kind.String(),
		Loader:     kind.Loader(),
	}
	return nil
}

func (e ScriptInlineFS) entryID(w idWriter) {
	w.WriteString("inline_fs")
	w.WriteString(e.Path)
	if e.Name == "" {
		b, _ := e.Read()
		w.Write(b)
		return
	}
	w.WriteString(e.Name)
}

type ScriptPath struct {
	Path string
}

func (e ScriptPath) Read() ([]byte, error) {
	return os.ReadFile(e.Path)
}

func (e ScriptPath) Apply(opt *api.BuildOptions) error {
	opt.EntryPoints = []string{e.Path}
	return nil
}

func (e ScriptPath) entryID(w idWriter) {
	w.WriteString("path")
	w.WriteString(e.Path)
}

type Kind int

const (
	KindJS Kind = iota
	KindTS
)

func (e Kind) Loader() api.Loader {
	switch e {
	case KindJS:
		return api.LoaderJS
	case KindTS:
		return api.LoaderTS
	default:
		return api.LoaderJS
	}
}

func (e Kind) String() string {
	switch e {
	case KindJS:
		return "js"
	case KindTS:
		return "ts"
	default:
		return "unknown"
	}
}

type ScriptBytes struct {
	Content []byte
	Kind    Kind
}

func (e ScriptBytes) Read() ([]byte, error) {
	return e.Content, nil
}

func (e ScriptBytes) Apply(opt *api.BuildOptions) error {
	opt.Stdin = &api.StdinOptions{
		Contents:   common.AsString(&e.Content),
		Sourcefile: "index." + e.Kind.String(),
		Loader:     e.Kind.Loader(),
	}
	return nil
}

func (e ScriptBytes) entryID(w idWriter) {
	w.WriteString("content")
	w.WriteString(e.Kind.String())
	w.Write(e.Content)
}

type ScriptInlinePath struct {
	Path string
}

func (e ScriptInlinePath) Read() ([]byte, error) {
	return os.ReadFile(e.Path)
}

func (e ScriptInlinePath) Apply(opt *api.BuildOptions) error {
	data, err := e.Read()
	if err != nil {
		return err
	}
	kind := KindJS
	if strings.HasSuffix(strings.ToLower(e.Path), ".ts") {
		kind = KindTS
	}
	opt.Stdin = &api.StdinOptions{
		Contents:   "_d0r(document.currentScript, async (" + inlineScriptArgs(kind) + ") => {\n" + string(data) + "\n})",
		Sourcefile: "index." + kind.String(),
		Loader:     kind.Loader(),
	}
	return nil
}

func (e ScriptInlinePath) entryID(w idWriter) {
	w.WriteString("inline_path")
	w.WriteString(e.Path)
}

type ScriptInlineString struct {
	Content string
	Kind    Kind
}

func (e ScriptInlineString) Read() ([]byte, error) {
	return []byte(e.Content), nil
}

func (e ScriptInlineString) Apply(opt *api.BuildOptions) error {
	opt.Stdin = &api.StdinOptions{
		Contents:   "_d0r(document.currentScript, async (" + inlineScriptArgs(e.Kind) + ") => {\n" + e.Content + "\n})",
		Sourcefile: "index." + e.Kind.String(),
		Loader:     e.Kind.Loader(),
	}
	return nil
}

func (e ScriptInlineString) entryID(w idWriter) {
	w.WriteString("inline_string")
	w.WriteString(e.Kind.String())
	w.WriteString(e.Content)
}

type ScriptInlineBytes struct {
	Content []byte
	Kind    Kind
}

func (e ScriptInlineBytes) Read() ([]byte, error) {
	return e.Content, nil
}

func (e ScriptInlineBytes) Apply(opt *api.BuildOptions) error {
	opt.Stdin = &api.StdinOptions{
		Contents:   "_d0r(document.currentScript, async (" + inlineScriptArgs(e.Kind) + ") => {\n" + string(e.Content) + "\n})",
		Sourcefile: "index." + e.Kind.String(),
		Loader:     e.Kind.Loader(),
	}
	return nil
}

func (e ScriptInlineBytes) entryID(w idWriter) {
	w.WriteString("inline_bytes")
	w.WriteString(e.Kind.String())
	w.Write(e.Content)
}

type ScriptString struct {
	Content string
	Kind    Kind
}

func (e ScriptString) Read() ([]byte, error) {
	return common.AsBytes(e.Content), nil
}

func (e ScriptString) Apply(opt *api.BuildOptions) error {
	opt.Stdin = &api.StdinOptions{
		Contents:   e.Content,
		Sourcefile: "index." + e.Kind.String(),
		Loader:     e.Kind.Loader(),
	}
	return nil
}

func (e ScriptString) entryID(w idWriter) {
	w.WriteString("content")
	w.WriteString(e.Kind.String())
	w.WriteString(e.Content)
}

type ScriptFormat interface {
	Apply(*api.BuildOptions)
	formatID(w idWriter)
}

type FormatDefault struct{}

func (f FormatDefault) Apply(opt *api.BuildOptions) {
	_ = opt
}

func (f FormatDefault) formatID(w idWriter) {
	w.WriteString("auto")
}

type FormatModule struct {
	Bundle bool
}

func (f FormatModule) Apply(opt *api.BuildOptions) {
	opt.Format = api.FormatESModule
	opt.Bundle = f.Bundle
}

func (f FormatModule) formatID(w idWriter) {
	w.WriteString("module")
	if f.Bundle {
		w.WriteString("bundle")
	}
}

type FormatCommon struct {
	Bundle bool
}

func (f FormatCommon) Apply(opt *api.BuildOptions) {
	opt.Format = api.FormatCommonJS
	opt.Bundle = f.Bundle
}

func (f FormatCommon) formatID(w idWriter) {
	w.WriteString("common")
	if f.Bundle {
		w.WriteString("bundle")
	}
}

type FormatIIFE struct {
	Bundle     bool
	GlobalName string
}

func (f FormatIIFE) Apply(opt *api.BuildOptions) {
	opt.Format = api.FormatIIFE
	opt.Bundle = f.Bundle
	opt.GlobalName = f.GlobalName
}

func (f FormatIIFE) formatID(w idWriter) {
	w.WriteString("iife")
	if f.Bundle {
		w.WriteString("bundle")
	}
	if f.GlobalName != "" {
		w.WriteString(f.GlobalName)
	}
}

type FormatRaw struct{}

func (f FormatRaw) Apply(opt *api.BuildOptions) {
	panic("raw format is not for use")
}

func (f FormatRaw) formatID(w idWriter) {
	w.WriteString("raw")
}
