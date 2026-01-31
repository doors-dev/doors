package resources

import (
	"io"
	"io/fs"
	"os"

	"github.com/doors-dev/doors/internal/common"
	"github.com/evanw/esbuild/pkg/api"
)

type idWriter interface {
	io.Writer
	io.StringWriter
}

type ScriptEntry interface {
	Read() ([]byte, error)
	Apply(*api.BuildOptions)
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

func (e ScriptFS) Apply(opt *api.BuildOptions) {
	opt.EntryPoints = []string{e.Path}
	opt.Plugins = append(opt.Plugins, fsPlugin(e.FS))
}

func (e ScriptFS) entryID(w idWriter) {
	w.WriteString("fs")
	w.WriteString(e.Path)
	w.WriteString(e.Name)
}

type ScriptPath struct {
	Path string
}

func (e ScriptPath) Read() ([]byte, error) {
	return os.ReadFile(e.Path)
}

func (e ScriptPath) Apply(opt *api.BuildOptions) {
	opt.EntryPoints = []string{e.Path}
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

func (e Kind) Load() api.Loader {
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

func (e ScriptBytes) Apply(opt *api.BuildOptions) {
	opt.Stdin = &api.StdinOptions{
		Contents:   common.AsString(&e.Content),
		Sourcefile: "index." + e.Kind.String(),
		Loader:     e.Kind.Load(),
	}
}

func (e ScriptBytes) entryID(w idWriter) {
	w.WriteString("content")
	w.WriteString(e.Kind.String())
	w.Write(e.Content)
}

type ScriptInline struct {
	Content string
}

func (e ScriptInline) Read() ([]byte, error) {
	return common.AsBytes(e.Content), nil
}

func (e ScriptInline) Apply(opt *api.BuildOptions) {
	opt.Stdin = &api.StdinOptions{
		Contents:   "_d00r(document.currentScript, async ($on, $data, $hook, $fetch, $G, $ready, $clean, HookErr) => {\n" + e.Content + "\n})",
		Sourcefile: "index.js",
		Loader:     api.LoaderJS,
	}
}

func (e ScriptInline) entryID(w idWriter) {
	w.WriteString("inline")
	w.WriteString(e.Content)
}

type ScriptString struct {
	Content string
	Kind    Kind
}

func (e ScriptString) Read() ([]byte, error) {
	return common.AsBytes(e.Content), nil
}

func (e ScriptString) Apply(opt *api.BuildOptions) {
	opt.Stdin = &api.StdinOptions{
		Contents:   e.Content,
		Sourcefile: "index." + e.Kind.String(),
		Loader:     api.LoaderJS,
	}
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

func (f FormatDefault) Apply(opt *api.BuildOptions) {}

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
